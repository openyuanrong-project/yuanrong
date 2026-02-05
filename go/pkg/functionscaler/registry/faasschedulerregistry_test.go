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
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
)

func mockEtcdEvent(instanceName, instanceId string, isLease bool) *etcd3.Event {
	etcdPrefix := "/sn/faas-scheduler/instances///"
	etcdkey := etcdPrefix + instanceName

	if isLease {
		etcdkey += "/" + uuid.New().String()
	}
	instanceInfo := &commonTypes.InstanceSpecification{
		InstanceID: instanceId,
	}
	return &etcd3.Event{
		Type:      etcd3.PUT,
		Key:       etcdkey,
		Value:     getBytes(instanceInfo),
		PrevValue: nil,
		Rev:       0,
		ETCDType:  "",
	}
}

func getBytes(info *commonTypes.InstanceSpecification) []byte {
	bytes, _ := json.Marshal(info)
	return bytes
}

func TestFaasSchedulerRegistryWatcherInstanceHandler(t *testing.T) {
	fsr := &FaasSchedulerRegistry{
		functionScheduler: make(map[string]*commonTypes.InstanceSpecification),
		moduleScheduler: &ModuleSchedulerInfos{
			schedulerInsSpecInfos: make(map[string]*commonTypes.InstanceSpecification, 0),
			schedulerInsInfos:     make(map[string]*commonTypes.InstanceInfo, 0),
			leaseIds:              make(map[string]map[string]bool),
		},
		schedulerHashListDoneCh:     make(chan struct{}, 1),
		schedulerInstanceListDoneCh: make(chan struct{}, 1),
		instanceMap:                 make(map[string]*commonTypes.InstanceSpecification),
	}

	event := &etcd3.Event{
		Type:      etcd3.PUT,
		Key:       "/sn/instance/business/yrk/tenant/0/function/0-system-faasscheduler/version/$latest/defaultaz/3576f5675c4f758500/b4bfe40f-6acd-4bc7-a006-a88356f42391",
		Value:     []byte("123"),
		PrevValue: []byte("123"),
		Rev:       1,
	}
	convey.Convey("test instance event", t, func() {
		event.Type = 999
		fsr.schedulerInstanceHandler(event)
		convey.So(fsr.instanceMap, convey.ShouldNotContainKey, "b4bfe40f-6acd-4bc7-a006-a88356f42391")
		event.Type = etcd3.PUT
		event.Value = []byte(`{
			    		"instanceID": "1f060613-68af-4a02-8000-000000e077ce",
			    		"instanceStatus": {
			    		    "code": 3,
			    		    "msg": "running"
			    		}}`)
		fsr.schedulerInstanceHandler(event)
		convey.So(fsr.instanceMap, convey.ShouldContainKey, "b4bfe40f-6acd-4bc7-a006-a88356f42391")

		event.Type = etcd3.DELETE
		fsr.schedulerInstanceHandler(event)
		convey.So(fsr.instanceMap, convey.ShouldNotContainKey, "b4bfe40f-6acd-4bc7-a006-a88356f42391")
		event.Type = etcd3.PUT
		event.Key = "/sn/instance/business/yrk/tenant//function"
		fsr.schedulerInstanceHandler(event)
		convey.So(len(fsr.instanceMap), convey.ShouldEqual, 0)
		event.Type = etcd3.SYNCED
		fsr.schedulerInstanceHandler(event)
		convey.So(len(fsr.schedulerInstanceListDoneCh), convey.ShouldEqual, 1)
	})
}

func TestFaasSchedulerRegistryWatcherHashHandler(t *testing.T) {
	fsr := &FaasSchedulerRegistry{
		functionScheduler: make(map[string]*commonTypes.InstanceSpecification),
		moduleScheduler: &ModuleSchedulerInfos{
			schedulerInsSpecInfos: make(map[string]*commonTypes.InstanceSpecification, 0),
			schedulerInsInfos:     make(map[string]*commonTypes.InstanceInfo, 0),
			leaseIds:              make(map[string]map[string]bool),
		},
		schedulerHashListDoneCh:     make(chan struct{}, 1),
		schedulerInstanceListDoneCh: make(chan struct{}, 1),
	}
	proxyMap := make(map[string]*commonTypes.InstanceInfo)
	defer gomonkey.ApplyMethodFunc(selfregister.GlobalSchedulerProxy, "Add", func(faaSScheduler *commonTypes.InstanceInfo, exclusivity string) {
		proxyMap[faaSScheduler.InstanceName] = faaSScheduler
	}).Reset()
	defer gomonkey.ApplyMethodFunc(selfregister.GlobalSchedulerProxy, "Remove", func(faasScheduler *commonTypes.InstanceInfo) {
		delete(proxyMap, faasScheduler.InstanceName)
	}).Reset()
	hash1Event := mockEtcdEvent("schedulerName1", "schedulerInstanceId1", false)
	lease1Event := mockEtcdEvent("schedulerName1", "schedulerInstanceId1", true)
	hash2Event := mockEtcdEvent("schedulerName2", "schedulerInstanceId2", false)
	lease2Event := mockEtcdEvent("schedulerName2", "schedulerInstanceId2", true)
	convey.Convey("test discoveryKeyType function", t, func() {
		fsr.discoveryKeyType = constant.SchedulerKeyTypeFunction
		mockEvent := mockEtcdEvent("1", "1", false)
		mockEvent.Type = 999
		fsr.schedulerHashHandler(mockEvent)
		convey.So(len(proxyMap), convey.ShouldEqual, 0)
		convey.So(len(proxyMap), convey.ShouldEqual, 0)
		fsr.schedulerHashHandler(hash1Event)
		fsr.schedulerHashHandler(hash2Event)
		convey.So(len(proxyMap), convey.ShouldEqual, 2)

		hash1Event.Type = etcd3.DELETE
		fsr.schedulerHashHandler(hash1Event)
		convey.So(len(proxyMap), convey.ShouldEqual, 1)
		hash2Event.Type = etcd3.DELETE
		fsr.schedulerHashHandler(hash2Event)
		convey.So(len(proxyMap), convey.ShouldEqual, 0)

		hash1Event.Type = etcd3.SYNCED
		fsr.schedulerHashHandler(hash1Event)
		convey.So(len(fsr.schedulerHashListDoneCh), convey.ShouldEqual, 1)
		<-fsr.schedulerHashListDoneCh
	})

	convey.Convey("test discoveryKeyType module", t, func() {
		hash1Event.Type = etcd3.PUT
		hash2Event.Type = etcd3.PUT

		fsr.discoveryKeyType = constant.SchedulerKeyTypeModule
		mockEvent := mockEtcdEvent("1", "1", false)
		mockEvent.Type = 999
		fsr.schedulerHashHandler(mockEvent)
		convey.So(len(proxyMap), convey.ShouldEqual, 0)
		fsr.schedulerHashHandler(lease1Event)
		fsr.schedulerHashHandler(lease2Event)
		convey.So(len(proxyMap), convey.ShouldEqual, 0)
		fsr.schedulerHashHandler(hash1Event)
		fsr.schedulerHashHandler(hash2Event)
		convey.So(len(proxyMap), convey.ShouldEqual, 2)

		hash1Event.Type = etcd3.DELETE
		fsr.schedulerHashHandler(hash1Event)
		convey.So(len(proxyMap), convey.ShouldEqual, 1)
		lease2Event.Type = etcd3.DELETE
		fsr.schedulerHashHandler(lease2Event)
		convey.So(len(proxyMap), convey.ShouldEqual, 0)

		hash1Event.Type = etcd3.SYNCED
		fsr.schedulerHashHandler(hash1Event)
		convey.So(len(fsr.schedulerHashListDoneCh), convey.ShouldEqual, 1)
	})
}

func TestWaitForETCDList(t *testing.T) {
	fsr := &FaasSchedulerRegistry{
		schedulerHashListDoneCh:     make(chan struct{}, 1),
		schedulerInstanceListDoneCh: make(chan struct{}, 1),
	}
	hashListDone := atomic.Bool{}
	instanceListDone := atomic.Bool{}
	convey.Convey("Test WaitForETCDList", t, func() {
		go func() {
			time.Sleep(100 * time.Millisecond)
			hashListDone.Store(true)
			fsr.schedulerHashListDoneCh <- struct{}{}
			time.Sleep(100 * time.Millisecond)
			instanceListDone.Store(true)
			fsr.schedulerInstanceListDoneCh <- struct{}{}
		}()
		fsr.WaitForETCDList()
		convey.So(instanceListDone.Load(), convey.ShouldBeTrue)
		convey.So(hashListDone.Load(), convey.ShouldBeTrue)
	})
}

func TestFilter(t *testing.T) {
	convey.Convey("test filter", t, func() {
		fsr := &FaasSchedulerRegistry{}
		event := &etcd3.Event{}
		tests := []*struct {
			key    string
			expect bool
		}{
			{
				key:    "/sn/instance/business/yrk/tenant/0/function/0-system-faasscheduler/version/$latest/defaultaz/3576f5675c4f758500/b4bfe40f-6acd-4bc7-a006-a88356f42391",
				expect: false,
			}, {
				key:    "/sn/instance/business/yrk/tenant/0/function/0-system-faasscheduler/version/$latest/defaultaz/3576f5675c4f758500/b4bfe40f-6acd-4bc7-a006-a88356f42391/",
				expect: true,
			}, {
				key:    "/sn/instance/business/yrk/tenant/0/function/0-system-faasscheduler/version/$latest/defaultaz/3576f5675c4f758500",
				expect: true,
			}, {
				key:    "/sn/instance/business/yrk/tenant/0/function/0-system-faascontroller/version/$latest/defaultaz/3576f5675c4f758500/b4bfe40f-6acd-4bc7-a006-a88356f42391",
				expect: true,
			},
		}
		for _, tt := range tests {
			event.Key = tt.key
			convey.So(fsr.schedulerInstanceFilter(event), convey.ShouldEqual, tt.expect)
		}
		tests = []*struct {
			key    string
			expect bool
		}{
			{
				key:    "/sn/faas-scheduler/instances///1a2b3c4d-5678-4e9f-a012-34567890ab12",
				expect: false,
			}, {
				key:    "/sn/faas-scheduler/instances///1a2b3c4d-5678-4e9f-a012-34567890ab12/leaseid1",
				expect: false,
			}, {
				key:    "/sn/instance/business/yrk/tenant/0/function/0-system-faasscheduler/version/$latest/defaultaz/3576f5675c4f758500",
				expect: true,
			}, {
				key:    "/sn/faas-scheduler111/business/yrk/tenant/0/function/0-system-faascontroller/version/$latest/defaultaz/3576f5675c4f758500/b4bfe40f-6acd-4bc7-a006-a88356f42391",
				expect: true,
			},
		}
		for _, tt := range tests {
			event.Key = tt.key
			convey.So(fsr.schedulerHashFilter(event), convey.ShouldEqual, tt.expect)
		}
	})
}
