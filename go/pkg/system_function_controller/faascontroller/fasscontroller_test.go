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

package faascontroller

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	ftypes "yuanrong.org/kernel/pkg/frontend/types"
	stypes "yuanrong.org/kernel/pkg/functionscaler/types"
	fcConfig "yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasfrontendmanager"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasfunctionmanager"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasschedulermanager"
	"yuanrong.org/kernel/pkg/system_function_controller/registry"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

var (
	configString = `{
		"frontendInstanceNum": 5,
		"schedulerInstanceNum": 5,
		"managerInstanceNum": 5,
		"faasschedulerConfig": {
			"slaQuota": 1000,
			"scaleDownTime": 60000,
			"burstScaleNum": 1000,
			"leaseSpan": 600000
		},
		"etcd": {
			"url": ["1.2.3.4:1234"],
			"username": "tom",
			"password": "**"
		},
		"enableRetry": true
	}
	`

	invalidConfigString = `{
		"frontendInstanceNum": 0,
		"schedulerInstanceNum": 0,
		"managerInstanceNum": 0,
		"faasschedulerConfig": {
			"slaQuota": 1000,
			"scaleDownTime": 60000,
			"burstScaleNum": 1000,
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

func newFaaSControllerHelper() (*FaaSController, error) {
	stopCh := make(chan struct{})
	fcConfig.InitConfig([]byte(configString))

	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&etcd3.EtcdInitParam{}), "InitClient", func(_ *etcd3.EtcdInitParam) error {
			return nil
		}),
		ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}),
		ApplyMethod(reflect.TypeOf(&faasschedulermanager.SchedulerManager{}), "GetInstanceCountFromEtcd",
			func(_ *faasschedulermanager.SchedulerManager) map[string]struct{} {
				return map[string]struct{}{}
			}),
		ApplyMethod(reflect.TypeOf(&faasfrontendmanager.FrontendManager{}), "GetInstanceCountFromEtcd",
			func(_ *faasfrontendmanager.FrontendManager) map[string]struct{} {
				return map[string]struct{}{}
			}),
		ApplyMethod(reflect.TypeOf(&faasfunctionmanager.FunctionManager{}), "GetInstanceCountFromEtcd",
			func(_ *faasfunctionmanager.FunctionManager) map[string]struct{} {
				return map[string]struct{}{}
			}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	controller, _ := NewFaaSController(&mockUtils.FakeLibruntimeSdkClient{}, stopCh)
	controller.instanceManager.FrontendManager = &faasfrontendmanager.FrontendManager{ConfigChangeCh: make(chan *types.ConfigChangeEvent,
		1)}
	controller.instanceManager.CommonSchedulerManager = &faasschedulermanager.SchedulerManager{ConfigChangeCh: make(chan *types.ConfigChangeEvent,
		1)}
	return controller, nil
}

// TestNewFaaSController -
func TestNewFaaSController(t *testing.T) {
	defer ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
		return &etcd3.EtcdClient{}
	}).Reset()
	defer ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "Put", func(_ *etcd3.EtcdClient,
		ctxInfo etcd3.EtcdCtxInfo, key string, value string, opts ...clientv3.OpOption) error {
		return nil
	}).Reset()
	// state.InitState()
	stopCh := make(chan struct{})
	Convey("Test NewFaaSController for failure", t, func() {

		Convey("Test NewFaaSController when passed invalid instance number", func() {
			fcConfig.InitConfig([]byte(invalidConfigString))
			controller, err := NewFaaSController(&mockUtils.FakeLibruntimeSdkClient{}, stopCh)
			So(controller, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
	time.Sleep(50 * time.Millisecond)
}
func Test_processFunctionSubscription(t *testing.T) {
	Convey("processFunctionSubscription", t, func() {
		defer ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}).Reset()
		defer ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "Put", func(_ *etcd3.EtcdClient,
			ctxInfo etcd3.EtcdCtxInfo, key string, value string, opts ...clientv3.OpOption) error {
			return nil
		}).Reset()
		controller, err := newFaaSControllerHelper()
		So(controller, ShouldNotBeNil)
		So(err, ShouldBeNil)

		Convey("assert failed and close ch", func() {
			go func() {
				controller.funcCh <- types.SubEvent{EventMsg: "invalid"}
				time.Sleep(500 * time.Millisecond)
				close(controller.funcCh)
			}()
			controller.processFunctionSubscription()
		})

		Convey("SubEventTypeUpdate", func() {
			controller.funcCh = make(chan types.SubEvent, 1)
			go func() {
				controller.funcCh <- types.SubEvent{
					EventType: types.SubEventTypeUpdate,
					EventMsg:  &types.InstanceSpecification{},
				}
				time.Sleep(500 * time.Millisecond)
				close(controller.funcCh)
			}()
			controller.processFunctionSubscription()
		})

		Convey("SubEventTypeDelete", func() {
			controller.funcCh = make(chan types.SubEvent, 1)
			go func() {
				controller.funcCh <- types.SubEvent{
					EventType: types.SubEventTypeDelete,
					EventMsg:  &types.InstanceSpecification{},
				}
				time.Sleep(500 * time.Millisecond)
				close(controller.funcCh)
			}()
			controller.processFunctionSubscription()
		})

		Convey("SubEventTypeRecover", func() {
			controller.funcCh = make(chan types.SubEvent, 1)
			go func() {
				controller.funcCh <- types.SubEvent{
					EventType: types.SubEventTypeRecover,
					EventMsg:  &types.InstanceSpecification{},
				}
				time.Sleep(500 * time.Millisecond)
				close(controller.funcCh)
			}()
			controller.processFunctionSubscription()
		})
		time.Sleep(50 * time.Millisecond)
	})
}

func TestFaaSController_FrontendSignalHandler(t *testing.T) {
	controller, err := newFaaSControllerHelper()
	assert.NotNil(t, controller)
	assert.Nil(t, err)
	Convey("FrontendSignalHandler", t, func() {
		Convey("config is emtpy", func() {
			err = controller.FrontendSignalHandler([]byte{})
			So(err, ShouldBeNil)
		})
		Convey("failed to decode frontend config", func() {
			err = controller.FrontendSignalHandler([]byte("123"))
			So(err, ShouldBeError)
		})
		Convey("success", func() {
			defer ApplyFunc(faasfrontendmanager.NewFaaSFrontendManager, func(libruntimeApi api.LibruntimeAPI,
				etcdClient *etcd3.EtcdClient, stopCh chan struct{}, size int, isDynamic bool) *faasfrontendmanager.FrontendManager {
				return &faasfrontendmanager.FrontendManager{ConfigChangeCh: make(chan *types.ConfigChangeEvent)}
			}).Reset()
			defer ApplyMethod(reflect.TypeOf(&registry.FaaSFrontendRegistry{}), "InitWatcher",
				func(fr *registry.FaaSFrontendRegistry) {}).Reset()
			defer ApplyMethod(reflect.TypeOf(&registry.FaaSFrontendRegistry{}), "RunWatcher",
				func(fr *registry.FaaSFrontendRegistry) {}).Reset()
			frontendConfig := types.FrontendConfig{
				Config: ftypes.Config{
					InstanceNum:          100,
					CPU:                  777,
					Memory:               777,
					SLAQuota:             1000,
					AuthenticationEnable: false,
					HTTPConfig: &ftypes.FrontendHTTP{
						MaxRequestBodySize: 6,
					},
				},
			}
			bytes, _ := json.Marshal(frontendConfig)
			encodeToString := base64.StdEncoding.EncodeToString(bytes)
			registry.InitRegistry()
			go func() {
				time.Sleep(50 * time.Millisecond)
				event := <-controller.instanceManager.FrontendManager.ConfigChangeCh
				event.Error = nil
				event.Done()
			}()
			err = controller.FrontendSignalHandler([]byte(encodeToString))
			So(err, ShouldBeNil)
		})
	})
}

func TestFaaSController_SchedulerSignalHandler(t *testing.T) {
	mockTenantID001 := "mock-tenant-001"
	mockTenantID002 := "mock-tenant-001"
	mockTenantSchedulerManager001 := &faasschedulermanager.SchedulerManager{ConfigChangeCh: make(chan *types.ConfigChangeEvent,
		1)}
	mockTenantSchedulerManager002 := &faasschedulermanager.SchedulerManager{ConfigChangeCh: make(chan *types.ConfigChangeEvent,
		1)}
	controller, err := newFaaSControllerHelper()
	assert.NotNil(t, controller)
	assert.Nil(t, err)
	Convey("SchedulerSignalHandler", t, func() {
		Convey("config is emtpy", func() {
			err = controller.SchedulerSignalHandler([]byte{})
			So(err, ShouldBeNil)
		})
		Convey("failed to decode scheduler config", func() {
			err = controller.SchedulerSignalHandler([]byte("123"))
			So(err, ShouldBeError)
		})
		Convey("success", func() {
			defer ApplyFunc(fcConfig.GetFaaSControllerConfig, func() types.Config {
				return types.Config{SchedulerExclusivity: []string{mockTenantID001, mockTenantID002}}
			}).Reset()
			defer ApplyFunc(faasschedulermanager.NewFaaSSchedulerManager, func(libruntimeAPI api.LibruntimeAPI,
				etcdClient *etcd3.EtcdClient, stopCh chan struct{}, size int,
				tenantID string) *faasschedulermanager.SchedulerManager {
				if tenantID == mockTenantID001 {
					return mockTenantSchedulerManager001
				}
				if tenantID == mockTenantID002 {
					return mockTenantSchedulerManager002
				}
				return &faasschedulermanager.SchedulerManager{ConfigChangeCh: make(chan *types.ConfigChangeEvent)}
			}).Reset()
			defer ApplyMethod(reflect.TypeOf(&registry.FaaSSchedulerRegistry{}), "InitWatcher",
				func(fr *registry.FaaSSchedulerRegistry) {}).Reset()
			defer ApplyMethod(reflect.TypeOf(&registry.FaaSSchedulerRegistry{}), "RunWatcher",
				func(fr *registry.FaaSSchedulerRegistry) {}).Reset()
			schedulerBasicConfig := types.SchedulerConfig{
				Configuration: stypes.Configuration{
					CPU:    999,
					Memory: 999,
					AutoScaleConfig: stypes.AutoScaleConfig{
						SLAQuota:      1000,
						ScaleDownTime: 60000,
						BurstScaleNum: 1000,
					},
					LeaseSpan: 600000,
				},
				SchedulerNum: 100,
			}
			bytes, _ := json.Marshal(schedulerBasicConfig)
			encodeToString := base64.StdEncoding.EncodeToString(bytes)
			registry.InitRegistry()
			go func() {
				time.Sleep(50 * time.Millisecond)
				event := <-controller.instanceManager.CommonSchedulerManager.ConfigChangeCh
				event.Error = nil
				event.Done()
			}()
			go func() {
				time.Sleep(50 * time.Millisecond)
				event := <-mockTenantSchedulerManager001.ConfigChangeCh
				event.Error = nil
				event.Done()
			}()
			go func() {
				time.Sleep(50 * time.Millisecond)
				event := <-mockTenantSchedulerManager002.ConfigChangeCh
				event.Error = nil
				event.Done()
			}()
			err = controller.SchedulerSignalHandler([]byte(encodeToString))
			So(err, ShouldBeNil)
		})
	})
}

func TestFaaSController_ManagerSignalHandler(t *testing.T) {
	controller, err := newFaaSControllerHelper()
	assert.NotNil(t, controller)
	assert.Nil(t, err)
	Convey("SchedulerSignalHandler", t, func() {
		Convey("config is emtpy", func() {
			err = controller.ManagerSignalHandler([]byte{})
			So(err, ShouldBeNil)
		})
		Convey("failed to decode scheduler config", func() {
			err = controller.ManagerSignalHandler([]byte("123"))
			So(err, ShouldBeError)
		})
		Convey("success", func() {
			defer ApplyFunc(faasfunctionmanager.NewFaaSFunctionManager, func(libruntimeAPI api.LibruntimeAPI,
				etcdClient *etcd3.EtcdClient, stopCh chan struct{}, size int) *faasfunctionmanager.FunctionManager {
				return &faasfunctionmanager.FunctionManager{ConfigChangeCh: make(chan *types.ConfigChangeEvent)}
			}).Reset()
			defer ApplyMethod(reflect.TypeOf(&registry.FaaSManagerRegistry{}), "InitWatcher",
				func(fr *registry.FaaSManagerRegistry) {}).Reset()
			defer ApplyMethod(reflect.TypeOf(&registry.FaaSManagerRegistry{}), "RunWatcher",
				func(fr *registry.FaaSManagerRegistry) {}).Reset()
			managerBasicConfig := types.ManagerConfig{
				CPU:    999,
				Memory: 999,
			}
			bytes, _ := json.Marshal(managerBasicConfig)
			encodeToString := base64.StdEncoding.EncodeToString(bytes)
			registry.InitRegistry()
			go func() {
				time.Sleep(50 * time.Millisecond)
				event := <-controller.instanceManager.FunctionManager.ConfigChangeCh
				event.Error = nil
				event.Done()
			}()
			err = controller.ManagerSignalHandler([]byte(encodeToString))
			So(err, ShouldBeNil)
		})
	})
}
