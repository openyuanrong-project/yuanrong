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

package logcollector

import (
	"context"
	"errors"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/collector/common"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	// wait in total 60s
	waitDashboardRegisterMaxTimes = 60
	waitDashboardRegisterInterval = 1 * time.Second

	dashboardRegisterKey = "/yr/dashboard"
	etcdGetTimeout       = 30 * time.Second
)

// GetManagerAddress from etcd
func GetManagerAddress() (string, error) {
	rsp, err := etcd3.GetRouterEtcdClient().Get(etcd3.CreateEtcdCtxInfoWithTimeout(context.TODO(), etcdGetTimeout),
		dashboardRegisterKey, clientv3.WithPrefix())
	if err != nil {
		return "", err
	}
	if len(rsp.Kvs) == 0 {
		return "", errors.New("failed to get dashboard address, get 0 responses")
	}
	return string(rsp.Kvs[0].Value), nil
}

// Register itself to manager
func Register() error {
	getDashboardAddrRetryLeftCnt := waitDashboardRegisterMaxTimes
	for ; getDashboardAddrRetryLeftCnt > 0; getDashboardAddrRetryLeftCnt -= 1 {
		addr, err := GetManagerAddress()
		if err != nil {
			log.GetLogger().Warnf("failed to get master address: %s, try again later", err.Error())
			time.Sleep(waitDashboardRegisterInterval)
			continue
		}
		common.CollectorConfigs.ManagerAddress = addr
		break
	}
	if getDashboardAddrRetryLeftCnt == 0 {
		return errors.New("failed to get dashboard address from etcd")
	}

	retryInterval := constant.GetRetryInterval()
	maxRetryTimes := constant.GetMaxRetryTimes()
	client := GetLogServiceClient()
	if client == nil {
		log.GetLogger().Errorf("failed to get log service client")
		return fmt.Errorf("failed to get log service client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), common.DefaultGrpcTimeoutS)
	defer cancel()

	for i := 0; i < maxRetryTimes; i++ {
		log.GetLogger().Infof("start to register %s to %s", common.CollectorConfigs.CollectorID,
			common.CollectorConfigs.ManagerAddress)
		response, err := client.Register(ctx, &logservice.RegisterRequest{
			CollectorID: common.CollectorConfigs.CollectorID,
			Address:     common.CollectorConfigs.Address,
		})
		if err != nil {
			log.GetLogger().Errorf("failed to send register, error: %v", err)
			time.Sleep(retryInterval)
			continue
		}
		if response.Code != 0 {
			log.GetLogger().Errorf("failed to register, error: %d, message: %s", response.Code, response.Message)
			time.Sleep(retryInterval)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to register: exceeds max retry time: %d", maxRetryTimes)
}
