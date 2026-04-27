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
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

var (
	cache                         = &tenantCache{}
	defaultMaxOnDemandInstanceNum = int64(types.DefaultMaxOnDemandInstanceNumPerTenant)
	defaultMaxReversedInstanceNum = int64(types.DefaultMaxReversedInstanceNumPerTenant)
)

type tenantCache struct {
	// key: tenantID   value: *tenantQuota
	tenantQuotaList sync.Map
	// key: tenantID   value: *tenantInstance
	tenantInstanceList sync.Map
}

type tenantQuota struct {
	maxOnDemandInstance int64
	maxReversedInstance int64
	mux                 sync.RWMutex
}

// GetTenantCache is cache for tenant
func GetTenantCache() *tenantCache {
	return cache
}

func createDefaultTenantQuota() *tenantQuota {
	tenantIns := &tenantQuota{
		maxOnDemandInstance: defaultMaxOnDemandInstanceNum,
		maxReversedInstance: defaultMaxReversedInstanceNum,
		mux:                 sync.RWMutex{},
	}
	return tenantIns
}

type tenantInstance struct {
	onDemandInsNum int64
	reversedInsNum int64
	mux            sync.RWMutex
}

func createTenantInstance() *tenantInstance {
	tenantIns := &tenantInstance{
		onDemandInsNum: 0,
		reversedInsNum: 0,
		mux:            sync.RWMutex{},
	}
	return tenantIns
}

func (tc *tenantCache) getOrCreateTenantInstance(tenantID string) *tenantInstance {
	tenantInsIf, exist := tc.tenantInstanceList.LoadOrStore(tenantID, createTenantInstance())
	if !exist {
		log.GetLogger().Infof("tenantInstance cache has not been created, generate, tenant: %s", tenantID)
	}

	return tenantInsIf.(*tenantInstance)
}

func (tc *tenantCache) getTenantInstanceNum(tenantID string) (int64, int64) {
	tenantIns := tc.getOrCreateTenantInstance(tenantID)
	tenantIns.mux.RLock()
	defer tenantIns.mux.RUnlock()
	return tenantIns.onDemandInsNum, tenantIns.reversedInsNum
}

func (tc *tenantCache) updateTenantInstance(tenantID string, onDemandInsNum int64, reversedInsNum int64) {
	if reversedInsNum == 0 && onDemandInsNum == 0 {
		_, exist := tc.tenantInstanceList.Load(tenantID)
		if !exist {
			log.GetLogger().Infof("tenantInstance cache has not been created, tenant: %s", tenantID)
			return
		}
		tc.tenantInstanceList.Delete(tenantID)
		log.GetLogger().Infof("succeed to delete instance in tenant cache, tenant: %s", tenantID)
		return
	}
	tenantIns := tc.getOrCreateTenantInstance(tenantID)
	tenantIns.mux.Lock()
	tenantIns.onDemandInsNum = onDemandInsNum
	tenantIns.reversedInsNum = reversedInsNum
	tenantIns.mux.Unlock()
	log.GetLogger().Infof("succeed to update instance in tenant cache, tenant: %s", tenantID)
}

func (tc *tenantCache) getTenantQuota(tenantID string) *tenantQuota {
	tenantQuotaIf, exist := tc.tenantQuotaList.Load(tenantID)
	if !exist {
		log.GetLogger().Infof("tenant cache has not been created, tenant: %s", tenantID)
		return nil
	}

	return tenantQuotaIf.(*tenantQuota)
}

// GetTenantQuotaNum get max instance quota for tenant
func (tc *tenantCache) GetTenantQuotaNum(tenantID string) (int64, int64) {
	quota := tc.getTenantQuota(tenantID)
	if quota == nil {
		return defaultMaxOnDemandInstanceNum, defaultMaxReversedInstanceNum
	}
	quota.mux.RLock()
	defer quota.mux.RUnlock()
	return quota.maxOnDemandInstance, quota.maxReversedInstance
}

// DeleteTenantQuota trigger by tenant quota key deleted in etcd
func (tc *tenantCache) DeleteTenantQuota(tenantID string) {
	_, exist := tc.tenantQuotaList.Load(tenantID)
	if !exist {
		log.GetLogger().Infof("tenant cache has not been created, tenant: %s", tenantID)
		return
	}
	tc.tenantQuotaList.Delete(tenantID)
	log.GetLogger().Infof("succeed to delete quota in tenant cache trigger by etcd for tenant: %s", tenantID)
}

// UpdateOrAddTenantQuota trigger by tenant quota key add or update in etcd
func (tc *tenantCache) UpdateOrAddTenantQuota(tenantID string, tenantMetaInfo types.TenantMetaInfo) {
	quota := tc.getTenantQuota(tenantID)
	if quota == nil {
		quota = createDefaultTenantQuota()
		tc.tenantQuotaList.Store(tenantID, quota)
	}
	quota.mux.Lock()
	quota.maxOnDemandInstance = tenantMetaInfo.TenantInstanceMetaData.MaxOnDemandInstance
	quota.maxReversedInstance = tenantMetaInfo.TenantInstanceMetaData.MaxReversedInstance
	quota.mux.Unlock()
	log.GetLogger().Infof("succeed to update quota in tenant cache trigger by etcd for tenant: %s", tenantID)
}

// UpdateDefaultQuota trigger by default tenant quota key add or update in etcd
func (tc *tenantCache) UpdateDefaultQuota(tenantMetaInfo types.TenantMetaInfo) {
	// update global config
	defaultMaxOnDemandInstanceNum = tenantMetaInfo.TenantInstanceMetaData.MaxOnDemandInstance
	defaultMaxReversedInstanceNum = tenantMetaInfo.TenantInstanceMetaData.MaxReversedInstance
	log.GetLogger().Infof("succeed to reset default quota in tenant cache trigger by etcd, "+
		"defaultMaxOnDemandInstanceNum: %d, defaultMaxReversedInstanceNum: %d",
		defaultMaxOnDemandInstanceNum, defaultMaxReversedInstanceNum)
}
