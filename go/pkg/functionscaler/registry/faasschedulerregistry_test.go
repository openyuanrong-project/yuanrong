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

// Package registry -
package registry

import (
	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
)

func TestFaasSchedulerRegistryWatcherHandler(t *testing.T) {
	fsr := &FaasSchedulerRegistry{
		FunctionScheduler:  make(map[string]*commonTypes.InstanceSpecification),
		ModuleScheduler:    make(map[string]*commonTypes.InstanceSpecification),
		functionListDoneCh: make(chan struct{}, 1),
		moduleListDoneCh:   make(chan struct{}, 1),
		stopDelayRemove:    make(map[string]chan struct{}),
	}

	event := &etcd3.Event{
		Type:      etcd3.PUT,
		Key:       "/sn/instance/business/yrk/tenant/12345678901234561234567890123456/function/0-system-faasExecutorJava8/version/$latest/defaultaz/task-b23aa1c4-2084-42b8-99b2-8907fa5ae6f4/f71875b1-3c20-4827-8600-0000000005d5",
		Value:     []byte("123"),
		PrevValue: []byte("123"),
		Rev:       1,
	}
	convey.Convey("test discoveryKeyType function", t, func() {
		fsr.discoveryKeyType = constant.SchedulerKeyTypeFunction
		convey.Convey("etcd opt error", func() {
			event.Type = 999
			fsr.functionSchedulerHandler(event)
		})
		convey.Convey("etcd put value success", func() {
			event.Type = etcd3.PUT
			event.Value = []byte(`{
    		"instanceID": "1f060613-68af-4a02-8000-000000e077ce",
    		"instanceStatus": {
    		    "code": 3,
    		    "msg": "running"
    		}}`)
			fsr.functionSchedulerHandler(event)
			_, ok := selfregister.GlobalSchedulerProxy.FaaSSchedulers.Load("f71875b1-3c20-4827-8600-0000000005d5")
			convey.So(ok, convey.ShouldBeTrue)
		})
		convey.Convey("etcd delete value success", func() {
			timer := time.NewTimer(1 * time.Second)
			defer gomonkey.ApplyFunc(time.NewTimer, func(d time.Duration) *time.Timer {
				return timer
			}).Reset()
			event.Type = etcd3.DELETE
			fsr.functionSchedulerHandler(event)
			time.Sleep(2 * time.Second)
			_, ok := selfregister.GlobalSchedulerProxy.FaaSSchedulers.Load("f71875b1-3c20-4827-8600-0000000005d5")
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("etcd put invalid funcKey", func() {
			event.Type = etcd3.PUT
			event.Key = "/sn/instance/business/yrk/tenant//function"
			fsr.functionSchedulerHandler(event)
			_, ok := selfregister.GlobalSchedulerProxy.FaaSSchedulers.Load("f71875b1-3c20-4827-8600-0000000005d5")
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("etcd SYNCED", func() {
			event.Type = etcd3.SYNCED
			fsr.functionSchedulerHandler(event)
			_, ok := selfregister.GlobalSchedulerProxy.FaaSSchedulers.Load("f71875b1-3c20-4827-8600-0000000005d5")
			convey.So(ok, convey.ShouldBeFalse)
		})
	})

	convey.Convey("test discoveryKeyType module", t, func() {
		defer func() {
			selfregister.GlobalSchedulerProxy.FaaSSchedulers.Delete("faas-scheduler-59ddbc4b75-8xdjf")
		}()
		fsr.discoveryKeyType = constant.SchedulerKeyTypeModule
		convey.Convey("etcd put valid funcKey for module scheduler", func() {
			event := &etcd3.Event{
				Type: etcd3.PUT,
				Key:  "/sn/faas-scheduler/instances/cluster001/x.x.x.x",
			}
			fsr.moduleSchedulerHandler(event)
			_, ok := selfregister.GlobalSchedulerProxy.FaaSSchedulers.Load("faas-scheduler-59ddbc4b75-8xdjf")
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("etcd SYNCED for module scheduler", func() {
			event.Type = etcd3.SYNCED
			fsr.moduleSchedulerHandler(event)
			_, ok := selfregister.GlobalSchedulerProxy.FaaSSchedulers.Load("faas-scheduler-59ddbc4b75-8xdjf")
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("etcd put invalid funcKey for module scheduler", func() {
			event := &etcd3.Event{
				Type: etcd3.PUT,
				Key:  "/sn/faas-scheduler/instances/cluster001/x.x.x.x/faas-scheduler-59ddbc4b75-8xdjf",
				Value: []byte(`{
    			"instanceID": "1f060613-68af-4a02-8000-000000e077ce",
    			"instanceStatus": {
    			    "code": 3,
    			    "msg": "running"
    			}}`),
				PrevValue: []byte("123"),
				Rev:       1,
			}
			fsr.moduleSchedulerHandler(event)
			_, ok := selfregister.GlobalSchedulerProxy.FaaSSchedulers.Load("faas-scheduler-59ddbc4b75-8xdjf")
			convey.So(ok, convey.ShouldBeTrue)
		})
	})
}

func TestFaaSSchedulerRegistryWatcherHandlerDelayDelete(t *testing.T) {
	fsr := &FaasSchedulerRegistry{
		FunctionScheduler:  make(map[string]*commonTypes.InstanceSpecification),
		ModuleScheduler:    make(map[string]*commonTypes.InstanceSpecification),
		functionListDoneCh: make(chan struct{}, 1),
		moduleListDoneCh:   make(chan struct{}, 1),
		stopDelayRemove:    make(map[string]chan struct{}),
	}
	event := &etcd3.Event{
		Type:      etcd3.PUT,
		PrevValue: []byte("123"),
		Rev:       1,
	}

	convey.Convey("test delete discoveryKeyType function", t, func() {
		fsr.discoveryKeyType = constant.SchedulerKeyTypeFunction
		instanceID := "1f060613-68af-4a02-8000-000000e077ce"
		functionSchedulerKey := "/sn/instance/business/yrk/tenant/12345678901234561234567890123456/function/0-system-faasExecutorJava8/version/$latest/defaultaz/task-b23aa1c4-2084-42b8-99b2-8907fa5ae6f4/1f060613-68af-4a02-8000-000000e077ce"
		event.Type = etcd3.PUT
		event.Key = functionSchedulerKey
		event.Value = []byte(`{
	    	"instanceID": "1f060613-68af-4a02-8000-000000e077ce",
	    	"instanceStatus": {
	    	    "code": 3,
	    	    "msg": "running"
	    	}}`)
		fsr.functionSchedulerHandler(event)
		convey.So(fsr.FunctionScheduler[instanceID], convey.ShouldNotBeNil)
		event.Type = etcd3.DELETE
		event.Key = "invalid key"
		fsr.functionSchedulerHandler(event)
		convey.So(fsr.FunctionScheduler[instanceID], convey.ShouldNotBeNil)
		event.Type = etcd3.DELETE
		event.Key = functionSchedulerKey
		fsr.functionSchedulerHandler(event)
		convey.So(fsr.FunctionScheduler[instanceID], convey.ShouldBeNil)
	})
	convey.Convey("test delete discoveryKeyType module", t, func() {
		fsr.discoveryKeyType = constant.SchedulerKeyTypeModule
		instanceID := "1f060613-68af-4a02-8000-000000e077ce"
		ModuleSchedulerKey := "/sn/faas-scheduler/instances/cluster001/node001/1f060613-68af-4a02-8000-000000e077ce"
		event.Type = etcd3.PUT
		event.Key = ModuleSchedulerKey
		event.Value = []byte(`{
	    	"instanceID": "1f060613-68af-4a02-8000-000000e077ce",
	    	"instanceStatus": {
	    	    "code": 3,
	    	    "msg": "running"
	    	}}`)
		fsr.moduleSchedulerHandler(event)
		convey.So(fsr.ModuleScheduler[instanceID], convey.ShouldNotBeNil)
		event.Type = etcd3.DELETE
		event.Key = "invalid key"
		fsr.moduleSchedulerHandler(event)
		convey.So(fsr.ModuleScheduler[instanceID], convey.ShouldNotBeNil)
		event.Type = etcd3.DELETE
		event.Key = ModuleSchedulerKey
		fsr.moduleSchedulerHandler(event)
		convey.So(fsr.ModuleScheduler[instanceID], convey.ShouldBeNil)
	})
}

func TestDelFunctionSchedulerFromProxy(t *testing.T) {
	fsr := &FaasSchedulerRegistry{
		FunctionScheduler:  make(map[string]*commonTypes.InstanceSpecification),
		ModuleScheduler:    make(map[string]*commonTypes.InstanceSpecification),
		functionListDoneCh: make(chan struct{}, 1),
		moduleListDoneCh:   make(chan struct{}, 1),
		stopDelayRemove:    make(map[string]chan struct{}),
	}
	insSpec := &commonTypes.InstanceSpecification{InstanceID: "instance1"}
	insInfo := &commonTypes.InstanceInfo{InstanceID: "instance1", InstanceName: "insName1"}
	config.GlobalConfig.AlarmConfig.EnableAlarm = true
	defer func() {
		selfregister.GlobalSchedulerProxy.FaaSSchedulers.Delete("insName1")
		config.GlobalConfig.AlarmConfig.EnableAlarm = false
	}()
	mockTimer := time.NewTimer(200 * time.Millisecond)
	defer gomonkey.ApplyFunc(time.NewTimer, func(d time.Duration) *time.Timer {
		return mockTimer
	}).Reset()
	convey.Convey("Test DelFunctionSchedulerFromProxy", t, func() {
		fsr.delFunctionSchedulerFromProxy(insSpec, insInfo)
		convey.So(fsr.stopDelayRemove["insName1"], convey.ShouldBeNil)
		selfregister.GlobalSchedulerProxy.FaaSSchedulers.Store("insName1", insInfo)
		fsr.delFunctionSchedulerFromProxy(insSpec, insInfo)
		convey.So(fsr.stopDelayRemove["insName1"], convey.ShouldNotBeNil)
		time.Sleep(500 * time.Millisecond)
		convey.So(fsr.stopDelayRemove["insName1"], convey.ShouldBeNil)
	})
}

func TestWaitForETCDList(t *testing.T) {
	fsr := &FaasSchedulerRegistry{
		functionListDoneCh: make(chan struct{}, 1),
		moduleListDoneCh:   make(chan struct{}, 1),
		stopCh:             make(chan struct{}, 1),
	}
	convey.Convey("Test WaitForETCDList", t, func() {
		fsr.discoveryKeyType = constant.SchedulerKeyTypeFunction
		go func() {
			time.Sleep(100 * time.Millisecond)
			close(fsr.functionListDoneCh)
		}()
		fsr.WaitForETCDList()
		_, ok := <-fsr.functionListDoneCh
		convey.So(ok, convey.ShouldBeFalse)

		fsr.discoveryKeyType = constant.SchedulerKeyTypeModule
		fsr.functionListDoneCh = make(chan struct{})
		go func() {
			time.Sleep(100 * time.Millisecond)
			close(fsr.functionListDoneCh)
			time.Sleep(100 * time.Millisecond)
			close(fsr.moduleListDoneCh)
		}()
		fsr.WaitForETCDList()
		_, ok = <-fsr.moduleListDoneCh
		convey.So(ok, convey.ShouldBeFalse)
	})
}
