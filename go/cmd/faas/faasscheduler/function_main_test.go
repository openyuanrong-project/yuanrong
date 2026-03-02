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

// Package main -
package main

import (
	"context"
	"errors"
	"reflect"
	"syscall"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestMain(m *testing.M) {
	patches := []*Patches{
		ApplyFunc((*etcd3.EtcdWatcher).StartList, func(ew *etcd3.EtcdWatcher) {
			ew.ResultChan <- &etcd3.Event{
				Type:      etcd3.SYNCED,
				Key:       "",
				Value:     nil,
				PrevValue: nil,
				Rev:       0,
				ETCDType:  "",
			}
		}),
		ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}),
	}
	defer func() {
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()
	registry.InitRegistry(stopCh)
	m.Run()
}

func TestInitHandler(t *testing.T) {
	var err error
	configJson := `{
		"cpu": 999,
		"memory": 999,
		"autoScaleConfig": {
			"slaQuota": 1000,
			"scaleDownTime": 20000,
			"burstScaleNum": 1000
		},
		"leaseSpan": 1000,
		"routerEtcd": {
			"servers": ["1.1.1.1:32379"],
			"username": "root",
			"password": ""
		},
		"metaEtcd": {
			"servers": ["1.1.1.1:32380"],
			"username": "root",
			"password": ""
		}
	}`
	testArgs := []api.Arg{
		{
			Type: api.Value,
			Data: []byte(configJson),
		},
	}
	patches := []*Patches{
		ApplyFunc(registry.InitRegistry, func(stopCh <-chan struct{}) error {
			return nil
		}),
		ApplyFunc(registry.StartRegistry, func() {
			return
		}),
		ApplyFunc((*registry.FunctionRegistry).WaitForETCDList, func() {}),
		ApplyFunc(functionscaler.NewFaaSScheduler,
			func(stopCh <-chan struct{}) *functionscaler.FaaSScheduler {
				return &functionscaler.FaaSScheduler{}
			}),
		ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "EtcdHeatBeat", func(e *etcd3.EtcdClient) error {
			return nil
		}),
		ApplyMethod(reflect.TypeOf(&etcd3.EtcdRegister{}), "Register", func(e *etcd3.EtcdRegister) error {
			return nil
		}),
	}
	defer func() {
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()
	selfregister.SetSelfInstanceSpec(&commonTypes.InstanceSpecification{})
	_, err = InitHandlerLibruntime(nil, &mockUtils.FakeLibruntimeSdkClient{})
	assert.Equal(t, "init args empty", err.Error())
	_, err = InitHandlerLibruntime(testArgs, &mockUtils.FakeLibruntimeSdkClient{})
	assert.Equal(t, true, err == nil)

	defer ApplyFunc(config.InitEtcd, func(stopCh <-chan struct{}) error {
		return errors.New("init etcd error")
	}).Reset()
	_, err = InitHandlerLibruntime(testArgs, &mockUtils.FakeLibruntimeSdkClient{})
	assert.Equal(t, true, err != nil)
}

func TestCallHandler(t *testing.T) {
	//	_, err := CallHandlerLibruntime(nil)
	configJson := `{
		"slaQuota": 1000,
		"scaleDownTime": 20000,
		"burstScaleNum": 1000,
		"leaseSpan": 1000,
		"routerEtcd": {
			"servers": ["1.1.1.1:32379"],
			"username": "root",
			"password": ""
		},
		"metaEtcd": {
			"servers": ["1.1.1.1:32380"],
			"username": "root",
			"password": ""
		}
	}`
	testFuncSpec := &types.FunctionSpecification{
		FuncCtx:      context.TODO(),
		FuncKey:      "TestFuncKey",
		FuncMetaData: commonTypes.FuncMetaData{},
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 500,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			MinInstance:   0,
			MaxInstance:   1000,
			ConcurrentNum: 100,
		},
	}
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "GetFuncSpec", func(_ *registry.Registry,
			funcKey string) *types.FunctionSpecification {
			return testFuncSpec
		}),
		ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch", func(_ *etcd3.EtcdWatcher) {
			return
		}),
	}
	defer func() {
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()

	config.InitConfig([]byte(configJson))
	functionscaler.InitGlobalScheduler(stopCh)
	functionscaler.GetGlobalScheduler().PoolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, testFuncSpec)
	testArgs := []api.Arg{
		{
			Type: api.Value,
			Data: []byte("acquire#TestFuncKey"),
		},
		{
			Type: api.Value,
			Data: []byte(""),
		},
	}
	rsp, err := CallHandlerLibruntime(testArgs)
	assert.Equal(t, true, err == nil)
	assert.Equal(t, true, rsp != nil)
}

func BenchmarkCallHandler(b *testing.B) {
	configJson := `{
		"slaQuota": 1000,
			"scaleDownTime": 20000,
			"burstScaleNum": 1000,
			"leaseSpan": 1000,
		"etcd": {
			"url": ["1.1.1.1:32379"],
			"username": "root",
			"password": ""
		}
	}`
	testFuncSpec := &types.FunctionSpecification{
		FuncCtx:      context.TODO(),
		FuncKey:      "TestFuncKey",
		FuncMetaData: commonTypes.FuncMetaData{},
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 500,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			MinInstance:   0,
			MaxInstance:   1000,
			ConcurrentNum: 100,
		},
	}
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "GetFuncSpec", func(_ *registry.Registry,
			funcKey string) *types.FunctionSpecification {
			return testFuncSpec
		}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	config.InitConfig([]byte(configJson))
	functionscaler.InitGlobalScheduler(make(chan struct{}))
	testArgs := []api.Arg{
		{
			Type: api.Value,
			Data: []byte("acquire#TestFuncKey"),
		},
		{
			Type: api.Value,
			Data: []byte(""),
		},
	}
	for i := 0; i < b.N; i++ {
		_, err := CallHandlerLibruntime(testArgs)
		if err != nil {
			b.Errorf("acquire instance thread error %s", err.Error())
		}
	}
}

func TestCheckpointHandler(t *testing.T) {
	convey.Convey("CheckpointHandler", t, func() {
		defer ApplyFunc(state.GetStateByte, func() ([]byte, error) {
			return []byte{}, nil
		}).Reset()
		handler, err := CheckpointHandlerLibruntime("checkpointID")
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(handler), convey.ShouldEqual, 0)
	})
}

func TestShutdownHandler(t *testing.T) {
	convey.Convey("ShutdownHandler", t, func() {
		stopCh = make(chan struct{})
		err := ShutdownHandlerLibruntime(uint64(30))
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestShutdownHandlerLibruntime(t *testing.T) {
	convey.Convey("ShutdownHandlerLibruntime", t, func() {
		stopCh = make(chan struct{})
		err := ShutdownHandlerLibruntime(uint64(30))
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSignalHandlerLibruntime(t *testing.T) {
	convey.Convey("SignalHandlerLibruntime", t, func() {
		err := SignalHandlerLibruntime(int(syscall.SIGTERM), nil)
		convey.So(err, convey.ShouldBeNil)
	})
}
