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
	"errors"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	fcConfig "yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
	"yuanrong.org/kernel/pkg/system_function_controller/utils"
)

var (
	validSchedulerEtcdKey = "/sn/instance/business/yrk/tenant/12/function/0-system-faasscheduler/version/$latest/defaultaz/requestID/123"
	validFrontendEtcdKey  = "/sn/instance/business/yrk/tenant/0/function/0-system-faasfrontend/version/$latest/defaultaz/requestID/123"
	validManagerEtcdKey   = "/sn/instance/business/yrk/tenant/0/function/0-system-faasmanager/version/$latest/defaultaz/requestID/123"
	invalidEtcdKey        = "/instance/business/yrk/tenant/12/function/0-system-faasscheduler/version/$latest/defaultaz/123"

	validEtcdValue = `{
		"runtimeID":"16444dbc",
		"deployedIP":"127.0.0.1",
		"deployedNode":"dggphis36581",
		"runtimeIP":"127.0.0.1",
		"runtimePort":"32065",
		"funcKey":"default/0-system-faasscheduler/$latest",
		"resource":{"cpu":"500","memory":"500","customresources":{}},
		"concurrency":1,
		"status":3,
		"labels":null}
	`

	configString = `{
		"frontendInstanceNum": 10,
		"schedulerInstanceNum": 10,
		"faasschedulerConfig": {
           "autoScaleConfig":{
              "slaQuota": 1000,
			  "scaleDownTime": 60000,
			  "burstScaleNum": 1000
            },
			"leaseSpan": 600000
		},
		"etcd": {
			"url": ["1.2.3.4:1234"],
			"username": "tom",
			"password": "**"
		}
	}
	`
)

func initConfig() {
	fcConfig.InitConfig([]byte(configString))
}

func generateETCDevent(eventType int, key string) *etcd3.Event {
	event := &etcd3.Event{
		Type:  eventType,
		Key:   key,
		Value: []byte(validEtcdValue),
	}
	return event
}

func initRegistry() {
	initConfig()

	patches := [...]*Patches{
		ApplyFunc(etcd3.GetSharedEtcdClient,
			func(_ *etcd3.EtcdConfig) (*clientv3.Client, error) {
				return nil, nil
			}),
		ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch",
			func(_ *etcd3.EtcdWatcher) {
				return
			}),
	}

	defer func() {
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()
	InitRegistry()
}

func TestInitRegistryPublish(t *testing.T) {
	initRegistry()
	Convey("Test RegistryPublish", t, func() {
		Convey("test scheduler registry publish", func() {
			funCh := make(chan types.SubEvent, 16)
			stopCh := make(chan struct{})
			GlobalRegistry.functionRegistry[types.EventKindScheduler] = NewSchedulerRegistry(stopCh)
			GlobalRegistry.functionRegistry[types.EventKindScheduler].AddSubscriberChan(funCh)

			GlobalRegistry.functionRegistry[types.EventKindScheduler].publishEvent(types.SubEventTypeUpdate,
				&types.InstanceSpecification{InstanceID: "test123"})
			event := <-funCh
			functionSpec, ok := event.EventMsg.(*types.InstanceSpecification)
			if !ok {
				assert.Error(t, errors.New("event assert types.InstanceSpecification failed"))
			}
			So(functionSpec.InstanceID, ShouldEqual, "test123")
			So(event.EventKind, ShouldEqual, types.EventKindScheduler)
			So(event.EventType, ShouldEqual, types.SubEventTypeUpdate)
		})

		Convey("test manager registry publish", func() {
			funCh := make(chan types.SubEvent, 16)
			stopCh := make(chan struct{})
			GlobalRegistry.functionRegistry[types.EventKindManager] = NewManagerRegistry(stopCh)
			GlobalRegistry.functionRegistry[types.EventKindManager].AddSubscriberChan(funCh)

			GlobalRegistry.functionRegistry[types.EventKindManager].publishEvent(types.SubEventTypeUpdate,
				&types.InstanceSpecification{InstanceID: "test123"})
			event := <-funCh
			functionSpec, ok := event.EventMsg.(*types.InstanceSpecification)
			if !ok {
				assert.Error(t, errors.New("event assert types.InstanceSpecification failed"))
			}
			So(functionSpec.InstanceID, ShouldEqual, "test123")
			So(event.EventKind, ShouldEqual, types.EventKindManager)
			So(event.EventType, ShouldEqual, types.SubEventTypeUpdate)
		})

		Convey("test frontend registry publish", func() {
			funCh := make(chan types.SubEvent, 16)
			stopCh := make(chan struct{})
			GlobalRegistry.functionRegistry[types.EventKindFrontend] = NewFrontendRegistry(stopCh)
			GlobalRegistry.functionRegistry[types.EventKindFrontend].AddSubscriberChan(funCh)

			GlobalRegistry.functionRegistry[types.EventKindFrontend].publishEvent(types.SubEventTypeUpdate,
				&types.InstanceSpecification{InstanceID: "test123"})
			event := <-funCh
			functionSpec, ok := event.EventMsg.(*types.InstanceSpecification)
			if !ok {
				assert.Error(t, errors.New("event assert types.InstanceSpecification failed"))
			}
			So(functionSpec.InstanceID, ShouldEqual, "test123")
			So(event.EventKind, ShouldEqual, types.EventKindFrontend)
			So(event.EventType, ShouldEqual, types.SubEventTypeUpdate)
		})

		Convey("test all registry publish", func() {
			funCh := make(chan types.SubEvent, 16)
			GlobalRegistry.RegisterSubscriberChan(funCh)
			GlobalRegistry.functionRegistry[types.EventKindScheduler].publishEvent(types.SubEventTypeUpdate,
				&types.InstanceSpecification{InstanceID: "test123"})
			GlobalRegistry.functionRegistry[types.EventKindFrontend].publishEvent(types.SubEventTypeUpdate,
				&types.InstanceSpecification{InstanceID: "test456"})
			GlobalRegistry.functionRegistry[types.EventKindManager].publishEvent(types.SubEventTypeUpdate,
				&types.InstanceSpecification{InstanceID: "test789"})
			event := <-funCh
			functionSpec, ok := event.EventMsg.(*types.InstanceSpecification)
			if !ok {
				assert.Error(t, errors.New("event assert types.InstanceSpecification failed"))
			}
			So(functionSpec.InstanceID, ShouldEqual, "test123")
			So(event.EventKind, ShouldEqual, types.EventKindScheduler)
			So(event.EventType, ShouldEqual, types.SubEventTypeUpdate)

			event = <-funCh
			functionSpec, ok = event.EventMsg.(*types.InstanceSpecification)
			if !ok {
				assert.Error(t, errors.New("event assert types.InstanceSpecification failed"))
			}
			So(functionSpec.InstanceID, ShouldEqual, "test456")
			So(event.EventKind, ShouldEqual, types.EventKindFrontend)
			So(event.EventType, ShouldEqual, types.SubEventTypeUpdate)

			event = <-funCh
			functionSpec, ok = event.EventMsg.(*types.InstanceSpecification)
			if !ok {
				assert.Error(t, errors.New("event assert types.InstanceSpecification failed"))
			}
			So(functionSpec.InstanceID, ShouldEqual, "test789")
			So(event.EventKind, ShouldEqual, types.EventKindManager)
			So(event.EventType, ShouldEqual, types.SubEventTypeUpdate)
		})
	})
}

func TestRegistry_runWatcher(t *testing.T) {
	var event *etcd3.Event
	initConfig()
	InitRegistry()
	stopCh := make(chan struct{})
	GlobalRegistry.functionRegistry[types.EventKindFrontend] = NewFrontendRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindScheduler] = NewSchedulerRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindManager] = NewManagerRegistry(stopCh)
	patches := [...]*Patches{
		ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}),
		ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartList", func(_ *etcd3.EtcdWatcher) {
		}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	Convey("test registry runWatcher", t, func() {
		Convey("test scheduler registry runWatcher", func() {
			c := make(chan *etcd3.Event)
			patches := [...]*Patches{
				ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch",
					func(_ *etcd3.EtcdWatcher) {
						if !GlobalRegistry.functionRegistry[types.EventKindScheduler].watcherFilter(event) {
							GlobalRegistry.functionRegistry[types.EventKindScheduler].watcherHandler(event)
							c <- event
							return
						}
					}),
			}

			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()

			event = generateETCDevent(etcd3.HISTORYUPDATE, validSchedulerEtcdKey)
			GlobalRegistry.functionRegistry[types.EventKindScheduler].InitWatcher()
			GlobalRegistry.functionRegistry[types.EventKindScheduler].RunWatcher()
			So(<-c, ShouldEqual, event)
		})

		Convey("test frontend registry runWatcher", func() {
			c := make(chan *etcd3.Event)
			patches := [...]*Patches{
				ApplyFunc(etcd3.GetSharedEtcdClient,
					func(_ *etcd3.EtcdConfig) (*clientv3.Client, error) {
						return &clientv3.Client{}, nil
					}),
				ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch",
					func(_ *etcd3.EtcdWatcher) {
						if !GlobalRegistry.functionRegistry[types.EventKindFrontend].watcherFilter(event) {
							GlobalRegistry.functionRegistry[types.EventKindFrontend].watcherHandler(event)
							c <- event
							return
						}
					}),
			}

			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			event = generateETCDevent(etcd3.PUT, validFrontendEtcdKey)
			GlobalRegistry.functionRegistry[types.EventKindFrontend].InitWatcher()
			GlobalRegistry.functionRegistry[types.EventKindFrontend].RunWatcher()
			So(<-c, ShouldEqual, event)
		})

		Convey("test manager registry runWatcher", func() {
			c := make(chan *etcd3.Event)
			patches := [...]*Patches{
				ApplyFunc(etcd3.GetSharedEtcdClient,
					func(_ *etcd3.EtcdConfig) (*clientv3.Client, error) {
						return &clientv3.Client{}, nil
					}),
				ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch",
					func(_ *etcd3.EtcdWatcher) {
						if !GlobalRegistry.functionRegistry[types.EventKindManager].watcherFilter(event) {
							GlobalRegistry.functionRegistry[types.EventKindManager].watcherHandler(event)
							c <- event
							return
						}
					}),
			}

			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			event = generateETCDevent(etcd3.PUT, validManagerEtcdKey)
			GlobalRegistry.functionRegistry[types.EventKindManager].InitWatcher()
			GlobalRegistry.functionRegistry[types.EventKindManager].RunWatcher()
			So(<-c, ShouldEqual, event)
		})

		Convey("test all registry runWatcher", func() {
			eventCh := make(chan *etcd3.Event, 1)
			frontendEventCh := make(chan *etcd3.Event, 1)
			patches := [...]*Patches{
				ApplyFunc(etcd3.GetSharedEtcdClient,
					func(_ *etcd3.EtcdConfig) (*clientv3.Client, error) {
						return &clientv3.Client{}, nil
					}),
				ApplyMethod(reflect.TypeOf(&etcd3.EtcdWatcher{}), "StartWatch",
					func(_ *etcd3.EtcdWatcher) {
						event = generateETCDevent(etcd3.PUT, validSchedulerEtcdKey)
						event1 := generateETCDevent(etcd3.HISTORYUPDATE, validFrontendEtcdKey)
						event2 := generateETCDevent(etcd3.PUT, validManagerEtcdKey)
						if !GlobalRegistry.functionRegistry[types.EventKindScheduler].watcherFilter(event) {
							GlobalRegistry.functionRegistry[types.EventKindScheduler].watcherHandler(event)
							eventCh <- event
						}
						if !GlobalRegistry.functionRegistry[types.EventKindFrontend].watcherFilter(event1) {
							GlobalRegistry.functionRegistry[types.EventKindFrontend].watcherHandler(event1)
							frontendEventCh <- event1
						}
						if !GlobalRegistry.functionRegistry[types.EventKindManager].watcherFilter(event1) {
							GlobalRegistry.functionRegistry[types.EventKindManager].watcherHandler(event1)
							frontendEventCh <- event2
						}
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			GlobalRegistry.ProcessETCDList()
			GlobalRegistry.RunFunctionWatcher()
			e1, _ := <-eventCh
			So(e1.Key, ShouldEqual, validSchedulerEtcdKey)
			e2, _ := <-frontendEventCh
			So(e2.Key, ShouldEqual, validFrontendEtcdKey)
		})
	})

}

func TestSchedulerRegistry_watcherFilter(t *testing.T) {
	initConfig()
	InitRegistry()
	stopCh := make(chan struct{})
	GlobalRegistry.functionRegistry[types.EventKindFrontend] = NewFrontendRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindScheduler] = NewSchedulerRegistry(stopCh)
	tests := []struct {
		name    string
		etcdKey string
		want    bool
	}{
		{
			"test valid scheduler instance etcd key",
			validSchedulerEtcdKey,
			false,
		},
		{
			"test invalid length of scheduler instance etcd key",
			invalidEtcdKey,
			true,
		},
		{
			"test invalid content of scheduler instance etcd key",
			"/sn/instance/business/yrk/tenant/12/functionxxx/0-system-faasscheduler/version/$latest/defaultaz/requestID/123",
			true,
		},
		{
			"test invalid content of scheduler instance etcd key",
			"/sn/instance/business/yrk/tenant/12/function/0-system-faascontroller/version/$latest/defaultaz/requestID/123",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := generateETCDevent(etcd3.PUT, tt.etcdKey)
			assert.Equalf(t, tt.want, GlobalRegistry.functionRegistry[types.EventKindScheduler].watcherFilter(event), "watcherFilter(%v)", event)
		})
	}
}

func TestFrontendRegistry_watcherFilter(t *testing.T) {
	initConfig()
	InitRegistry()
	stopCh := make(chan struct{})
	GlobalRegistry.functionRegistry[types.EventKindFrontend] = NewFrontendRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindScheduler] = NewSchedulerRegistry(stopCh)
	tests := []struct {
		name    string
		etcdKey string
		want    bool
	}{
		{
			"test valid frontend instance etcd key",
			validFrontendEtcdKey,
			false,
		},
		{
			"test invalid length of frontend instance etcd key",
			invalidEtcdKey,
			true,
		},
		{
			"test invalid content of frontend instance etcd key",
			"/sn/instance/business/yrk/tenant/0/functionxxx/0-system-faasfrontend/version/$latest/defaultaz/requestID/123",
			true,
		},
		{
			"test invalid content of frontend instance etcd key",
			"/sn/instance/business/yrk/tenant/0/function/0-system-faascontroller/version/$latest/defaultaz/requestID/123",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := generateETCDevent(etcd3.PUT, tt.etcdKey)
			assert.Equalf(t, tt.want, GlobalRegistry.functionRegistry[types.EventKindFrontend].watcherFilter(event), "watcherFilter(%v)", event)
		})
	}
}

func TestManagerRegistry_watcherFilter(t *testing.T) {
	initConfig()
	InitRegistry()
	stopCh := make(chan struct{})
	GlobalRegistry.functionRegistry[types.EventKindFrontend] = NewFrontendRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindScheduler] = NewSchedulerRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindManager] = NewManagerRegistry(stopCh)
	tests := []struct {
		name    string
		etcdKey string
		want    bool
	}{
		{
			"test valid frontend instance etcd key",
			validFrontendEtcdKey,
			true,
		},
		{
			"test invalid length of frontend instance etcd key",
			invalidEtcdKey,
			true,
		},
		{
			"test invalid content of frontend instance etcd key",
			"/sn/instance/business/yrk/tenant/0/functionxxx/0-system-faasfrontend/version/$latest/defaultaz/requestID/123/xxx",
			true,
		},
		{
			"test invalid content of frontend instance etcd key",
			"/sn/instance/business/yrk/tenant/0/function/0-system-faascontroller/version/$latest/defaultaz/requestID/123/xxx",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := generateETCDevent(etcd3.PUT, tt.etcdKey)
			assert.Equalf(t, tt.want, GlobalRegistry.functionRegistry[types.EventKindManager].watcherFilter(event), "watcherFilter(%v)", event)
		})
	}
}

func TestSchedulerRegistry_watcherHandler(t *testing.T) {
	initConfig()
	InitRegistry()
	stopCh := make(chan struct{})
	GlobalRegistry.functionRegistry[types.EventKindFrontend] = NewFrontendRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindScheduler] = NewSchedulerRegistry(stopCh)
	tests := []struct {
		name      string
		etcdKey   string
		eventType int
		wantNil   bool
	}{
		{
			"test invalid instanceID",
			"",
			etcd3.PUT,
			true,
		},
		{
			"test watcherHandler with put event",
			validSchedulerEtcdKey,
			etcd3.PUT,
			false,
		},
		{
			"test watcherHandler with duplicated put event",
			validSchedulerEtcdKey,
			etcd3.PUT,
			false,
		},
		{
			"test watcherHandler with delete event",
			validSchedulerEtcdKey,
			etcd3.DELETE,
			true,
		},
		{
			"test watcherHandler with delete event",
			validSchedulerEtcdKey,
			etcd3.DELETE,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := generateETCDevent(tt.eventType, tt.etcdKey)
			GlobalRegistry.functionRegistry[types.EventKindScheduler].watcherHandler(event)

			instanceID := utils.ExtractInstanceIDFromEtcdKey(tt.etcdKey)
			ans := GlobalRegistry.functionRegistry[types.EventKindScheduler].getFunctionSpec(instanceID)
			if !tt.wantNil {
				assert.NotNil(t, ans)
			} else {
				assert.Nil(t, ans)
			}
		})
	}
}

func TestFrontendRegistry_watcherHandler(t *testing.T) {
	initConfig()
	InitRegistry()
	stopCh := make(chan struct{})
	GlobalRegistry.functionRegistry[types.EventKindFrontend] = NewFrontendRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindScheduler] = NewSchedulerRegistry(stopCh)
	tests := []struct {
		name      string
		etcdKey   string
		eventType int
		wantNil   bool
	}{
		{
			"test invalid instanceID",
			"",
			etcd3.PUT,
			true,
		},
		{
			"test watcherHandler with put event",
			validFrontendEtcdKey,
			etcd3.PUT,
			false,
		},
		{
			"test watcherHandler with duplicated put event",
			validFrontendEtcdKey,
			etcd3.PUT,
			false,
		},
		{
			"test watcherHandler with delete event",
			validFrontendEtcdKey,
			etcd3.DELETE,
			true,
		},
		{
			"test watcherHandler with delete event",
			validFrontendEtcdKey,
			etcd3.DELETE,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := generateETCDevent(tt.eventType, tt.etcdKey)
			GlobalRegistry.functionRegistry[types.EventKindFrontend].watcherHandler(event)

			instanceID := utils.ExtractInstanceIDFromEtcdKey(tt.etcdKey)
			ans := GlobalRegistry.functionRegistry[types.EventKindFrontend].getFunctionSpec(instanceID)
			if !tt.wantNil {
				assert.NotNil(t, ans)
			} else {
				assert.Nil(t, ans)
			}
		})
	}
}

func TestManagerRegistry_watcherHandler(t *testing.T) {
	initConfig()
	InitRegistry()
	stopCh := make(chan struct{})
	GlobalRegistry.functionRegistry[types.EventKindFrontend] = NewFrontendRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindScheduler] = NewSchedulerRegistry(stopCh)
	GlobalRegistry.functionRegistry[types.EventKindManager] = NewManagerRegistry(stopCh)
	tests := []struct {
		name      string
		etcdKey   string
		eventType int
		wantNil   bool
	}{
		{
			"test invalid instanceID",
			"",
			etcd3.PUT,
			true,
		},
		{
			"test watcherHandler with put event",
			validFrontendEtcdKey,
			etcd3.PUT,
			false,
		},
		{
			"test watcherHandler with duplicated put event",
			validFrontendEtcdKey,
			etcd3.PUT,
			false,
		},
		{
			"test watcherHandler with delete event",
			validFrontendEtcdKey,
			etcd3.DELETE,
			true,
		},
		{
			"test watcherHandler with delete event",
			validFrontendEtcdKey,
			etcd3.DELETE,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := generateETCDevent(tt.eventType, tt.etcdKey)
			GlobalRegistry.functionRegistry[types.EventKindManager].watcherHandler(event)

			instanceID := utils.ExtractInstanceIDFromEtcdKey(tt.etcdKey)
			ans := GlobalRegistry.functionRegistry[types.EventKindManager].getFunctionSpec(instanceID)
			if !tt.wantNil {
				assert.NotNil(t, ans)
			} else {
				assert.Nil(t, ans)
			}
		})
	}
}

func Test_extractInstanceIDFromEtcdKey(t *testing.T) {
	type args struct {
		etcdKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test valid etcd key",
			args{etcdKey: validSchedulerEtcdKey},
			"123",
		},
		{
			"test invalid etcd key",
			args{etcdKey: invalidEtcdKey},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.ExtractInstanceIDFromEtcdKey(tt.args.etcdKey); got != tt.want {
				t.Errorf("extractInstanceIDFromEtcdKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSchedulerSpecFromEtcdValue(t *testing.T) {
	type args struct {
		etcdValue []byte
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			"test valid etcd value",
			args{etcdValue: []byte(validEtcdValue)},
			false,
		},
		{
			"test invalid etcd value",
			args{etcdValue: []byte("123")},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.GetInstanceSpecFromEtcdValue(tt.args.etcdValue)
			assert.Equal(t, got == nil, tt.wantNil)
		})
	}
}
