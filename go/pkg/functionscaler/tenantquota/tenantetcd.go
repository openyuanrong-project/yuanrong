/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package tenantquota -
package tenantquota

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.etcd.io/etcd/client/v3/concurrency"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	defaultTTL                           = 1
	defaultUnlimitedInstanceNumPerTenant = -1 // define the default unlimited value per tenant
)

func max(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func getTenantInsInfoFromETCD(tenantID string) types.TenantInsInfo {
	tenantKey := fmt.Sprintf("/sn/functions/tenantinstancenumlimit/cluster/%s/tenant/%s",
		os.Getenv("CLUSTER_ID"), tenantID)
	tenantInsInfo := types.TenantInsInfo{}
	etcdValue, err := etcd3.GetValueFromEtcdWithRetry(tenantKey, etcd3.GetRouterEtcdClient())
	if err != nil {
		log.GetLogger().Warnf("failed to get tenant instance info, err: %s", err.Error())
		return tenantInsInfo
	}
	if err = json.Unmarshal(etcdValue, &tenantInsInfo); err != nil {
		log.GetLogger().Warnf("failed to Unmarshal tenant instance info, err: %s", err.Error())
		return tenantInsInfo
	}
	return tenantInsInfo
}

// IncreaseTenantInstanceNum 实例数未到上限，扩容先增加租户实例数
func IncreaseTenantInstanceNum(tenantID string, isReserved bool) (bool, bool) {
	var reachMaxOnDemandInsNum bool
	maxOnDemandInstance, maxReversedInstance := GetTenantCache().GetTenantQuotaNum(tenantID)
	// 若租户实例数quota在etcd中被置为-1，则跳过租户级流控(比如给某些VIP租户不限流)
	if maxOnDemandInstance == int64(defaultUnlimitedInstanceNumPerTenant) ||
		maxReversedInstance == int64(defaultUnlimitedInstanceNumPerTenant) {
		return false, false
	}

	lockKey := fmt.Sprintf("/lock/cluster/%s/tenant/%s", os.Getenv("CLUSTER_ID"), tenantID)
	routerEtcdClient := etcd3.GetRouterEtcdClient()

	onDemandInsNum, reversedInsNum := GetTenantCache().getTenantInstanceNum(tenantID)
	session, err := concurrency.NewSession(routerEtcdClient.Client, concurrency.WithTTL(defaultTTL)) // Generate lease
	if err != nil {
		log.GetLogger().Errorf("failed to new session: %s, determine based on cache", err.Error())
		if isReserved {
			reversedInsNum++
		} else {
			reachMaxOnDemandInsNum = maxOnDemandInstance < (onDemandInsNum + 1)
			if !reachMaxOnDemandInsNum {
				onDemandInsNum++
			}
		}
		GetTenantCache().updateTenantInstance(tenantID, onDemandInsNum, reversedInsNum)
		return reachMaxOnDemandInsNum, maxReversedInstance < reversedInsNum
	}
	defer session.Close()

	// Blocking, other requests will block waiting for the lock to be released
	locker := concurrency.NewLocker(session, lockKey)
	locker.Lock()

	// 1. 获取tenantID的函数实例数
	tenantInsInfo := getTenantInsInfoFromETCD(tenantID)
	tenantInsInfo.ReversedInsNum = max(tenantInsInfo.ReversedInsNum, reversedInsNum)
	tenantInsInfo.OnDemandInsNum = max(tenantInsInfo.OnDemandInsNum, onDemandInsNum)
	if isReserved {
		tenantInsInfo.ReversedInsNum++
		log.GetLogger().Debugf("tenantInsInfo.ReversedInsNum: %d", tenantInsInfo.ReversedInsNum)
	} else {
		// 2. Determine whether the limit is exceeded after increasing the number of instances
		if maxOnDemandInstance < tenantInsInfo.OnDemandInsNum+1 {
			onDemandInsNum = tenantInsInfo.OnDemandInsNum
			reversedInsNum = tenantInsInfo.ReversedInsNum
			GetTenantCache().updateTenantInstance(tenantID, onDemandInsNum, reversedInsNum)
			locker.Unlock()
			return true, maxReversedInstance < tenantInsInfo.ReversedInsNum
		}
		tenantInsInfo.OnDemandInsNum++
		log.GetLogger().Debugf("tenantInsInfo.OnDemandInsNum: %d", tenantInsInfo.OnDemandInsNum)
	}
	// 3. 弹性实例数不超限，更新（增加）实例数; 注意预留实例超限还是会创建实例，所以不管预留实例是否超限需要更新实例数
	updateTenantInstance(tenantID, tenantInsInfo.OnDemandInsNum, tenantInsInfo.ReversedInsNum)
	onDemandInsNum = tenantInsInfo.OnDemandInsNum
	reversedInsNum = tenantInsInfo.ReversedInsNum
	GetTenantCache().updateTenantInstance(tenantID, onDemandInsNum, reversedInsNum)
	locker.Unlock()

	return false, maxReversedInstance < tenantInsInfo.ReversedInsNum
}

// DecreaseTenantInstance Reduce the number of instances
func DecreaseTenantInstance(tenantID string, isReserved bool) {
	lockKey := fmt.Sprintf("/lock/cluster/%s/tenant/%s", os.Getenv("CLUSTER_ID"), tenantID)
	routerEtcdClient := etcd3.GetRouterEtcdClient()

	onDemandInsNum, reversedInsNum := GetTenantCache().getTenantInstanceNum(tenantID)
	session, err := concurrency.NewSession(routerEtcdClient.Client, concurrency.WithTTL(defaultTTL)) // Generate lease
	if err != nil {
		log.GetLogger().Errorf("failed to new session: %s", err.Error())
		if isReserved {
			reversedInsNum--
		} else {
			onDemandInsNum--
		}
		GetTenantCache().updateTenantInstance(tenantID, onDemandInsNum, reversedInsNum)
		return
	}
	defer session.Close()

	// Blocking, other requests will block waiting for the lock to be released
	locker := concurrency.NewLocker(session, lockKey)
	locker.Lock()

	// The number of instances needs to be reduced when creation fails, scales down, functions are deleted, etc.
	tenantInsInfo := getTenantInsInfoFromETCD(tenantID)
	tenantInsInfo.ReversedInsNum = max(tenantInsInfo.ReversedInsNum, reversedInsNum)
	tenantInsInfo.OnDemandInsNum = max(tenantInsInfo.OnDemandInsNum, onDemandInsNum)
	if isReserved {
		tenantInsInfo.ReversedInsNum--
		log.GetLogger().Debugf("tenantInsInfo.ReversedInsNum: %d", tenantInsInfo.ReversedInsNum)
	} else {
		tenantInsInfo.OnDemandInsNum--
		log.GetLogger().Debugf("tenantInsInfo.OnDemandInsNum: %d", tenantInsInfo.OnDemandInsNum)
	}
	updateTenantInstance(tenantID, tenantInsInfo.OnDemandInsNum, tenantInsInfo.ReversedInsNum)
	onDemandInsNum = tenantInsInfo.OnDemandInsNum
	reversedInsNum = tenantInsInfo.ReversedInsNum
	GetTenantCache().updateTenantInstance(tenantID, onDemandInsNum, reversedInsNum)
	locker.Unlock()
}

func updateTenantInstance(tenantID string, onDemandInsNum int64, reversedInsNum int64) {
	tenantInsInfo := types.TenantInsInfo{OnDemandInsNum: onDemandInsNum, ReversedInsNum: reversedInsNum}
	bytes, err := json.Marshal(tenantInsInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal tenantInsInfo, err: %s", err)
		return
	}
	ctx := etcd3.CreateEtcdCtxInfoWithTimeout(context.Background(), etcd3.DurationContextTimeout)
	routerEtcdClient := etcd3.GetRouterEtcdClient()
	tenantKey := fmt.Sprintf("/sn/functions/tenantinstancenumlimit/cluster/%s/tenant/%s",
		os.Getenv("CLUSTER_ID"), tenantID)
	err = routerEtcdClient.Put(ctx, tenantKey, string(bytes))
	if err != nil {
		log.GetLogger().Errorf("unable to put key: %s new value to router etcd, err:%s", tenantKey, err.Error())
	}
	return
}
