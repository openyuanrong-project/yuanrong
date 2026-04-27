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

// Package instancequeue -
package instancequeue

import (
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commontypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
)

func TestMain(m *testing.M) {
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartList, func(_ *etcd3.EtcdWatcher) {}),
		gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		gomonkey.ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		gomonkey.ApplyFunc(etcd3.GetCAEMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		gomonkey.ApplyFunc((*registry.FaasSchedulerRegistry).WaitForETCDList, func() {}),
	}
	defer func() {
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()
	_ = registry.InitRegistry(make(chan struct{}))
	registry.GlobalRegistry.FaaSSchedulerRegistry = registry.NewFaasSchedulerRegistry(make(chan struct{}))
	selfregister.SelfInstanceID = "schedulerID-1"
	selfregister.GlobalSchedulerProxy.Add(&commontypes.InstanceInfo{
		TenantID:     "123456789",
		FunctionName: "faasscheduler",
		Version:      "lastest",
		InstanceName: "schedulerID-1",
	}, "")
	m.Run()
}

func TestBuildSnError(t *testing.T) {
	assert.Equal(t, statuscode.NoInstanceAvailableErrCode, buildSnError(scheduler.ErrNoInsAvailable).Code())
	assert.Equal(t, statuscode.InstanceNotFoundErrCode, buildSnError(scheduler.ErrInsNotExist).Code())
	assert.Equal(t, statuscode.InsThdReqTimeoutCode, buildSnError(scheduler.ErrInsReqTimeout).Code())
	assert.Equal(t, statuscode.NoInstanceAvailableErrCode, buildSnError(scheduler.ErrOverAcqLimitExceeded).Code())
	assert.Equal(t, statuscode.StatusInternalServerError, buildSnError(errors.New("some error")).Code())
}
