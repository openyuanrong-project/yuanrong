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

package instancemanager

import (
	"context"
	"reflect"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/constant"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasfrontendmanager"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasfunctionmanager"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasschedulermanager"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

type KvMock struct {
}

func (k *KvMock) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (k *KvMock) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	response := &clientv3.GetResponse{}
	response.Count = 10
	return response, nil
}

func (k *KvMock) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (k *KvMock) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (k *KvMock) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (k *KvMock) Txn(ctx context.Context) clientv3.Txn {
	// TODO implement me
	panic("implement me")
}

func newInstanceManagerHelper(sdkClient api.LibruntimeAPI, stopCh chan struct{},
	frontendNum, schedulerNum int) (*InstanceManager, error) {
	configString := `{
		"frontendInstanceNum": 100,
		"schedulerInstanceNum": 100,
		"faasschedulerConfig": {
            "autoScaleConfig":{
              "slaQuota": 1000,
			  "scaleDownTime": 60000,
			  "burstScaleNum": 1000
             },
			"leaseSpan": 600000
		},
		"faasfrontendConfig": {
			"slaQuota": 1000,
			"functionCapability": 1,
			"authenticationEnable": false,
			"trafficLimitDisable": true,
			"http": {
                "resptimeout": 5,
                "workerInstanceReadTimeOut": 5,
                "maxRequestBodySize": 6
            }
		},
		"routerEtcd": {
			"servers": ["1.2.3.4:1234"],
			"user": "tom",
			"password": "**"
		},
		"metaEtcd": {
			"servers": ["1.2.3.4:5678"],
			"user": "tom",
			"password": "**"
		}
	}
	`
	patches := []*gomonkey.Patches{
		gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdInitParam{}), "InitClient", func(_ *etcd3.EtcdInitParam) error {
			return nil
		}),
		gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}),
	}
	defer func() {
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()
	_ = config.InitConfig([]byte(configString))
	kv := &KvMock{}
	client := &clientv3.Client{KV: kv}
	etcdClient := &etcd3.EtcdClient{Client: client}

	return &InstanceManager{
		FrontendManager: faasfrontendmanager.NewFaaSFrontendManager(sdkClient, etcdClient, stopCh, 0, false),
		FunctionManager: faasfunctionmanager.NewFaaSFunctionManager(sdkClient, etcdClient, stopCh, 0),
	}, nil
}

func TestHandleEventUpdate(t *testing.T) {
	convey.Convey("HandleEventUpdate", t, func() {
		mockTenantID := "mock-tenant-001"
		manager, err := newInstanceManagerHelper(nil, make(chan struct{}), 1, 1)
		convey.So(manager, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)
		manager.ExclusivitySchedulerManagers = map[string]*faasschedulermanager.SchedulerManager{
			mockTenantID: {},
		}

		convey.Convey("EventKindFrontend", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&faasfrontendmanager.FrontendManager{}), "HandleInstanceUpdate",
					func(_ *faasfrontendmanager.FrontendManager, instanceSpec *types.InstanceSpecification) {
						return
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			spec := &types.InstanceSpecification{}
			manager.HandleEventUpdate(spec, types.EventKindFrontend)
		})

		convey.Convey("EventKindScheduler", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&faasschedulermanager.SchedulerManager{}), "HandleInstanceUpdate",
					func(_ *faasschedulermanager.SchedulerManager, instanceSpec *types.InstanceSpecification) {
						return
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			spec := &types.InstanceSpecification{}
			manager.HandleEventUpdate(spec, types.EventKindScheduler)
		})

		convey.Convey("EventKindSchedulerWithNotEmptyTenantID", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&faasschedulermanager.SchedulerManager{}), "HandleInstanceUpdate",
					func(_ *faasschedulermanager.SchedulerManager, instanceSpec *types.InstanceSpecification) {
						return
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			spec := &types.InstanceSpecification{
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					CreateOptions: map[string]string{
						constant.SchedulerExclusivity: mockTenantID,
					},
				},
			}
			manager.HandleEventUpdate(spec, types.EventKindScheduler)
		})

		convey.Convey("EventKindManager", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&faasfunctionmanager.FunctionManager{}), "HandleInstanceUpdate",
					func(_ *faasfunctionmanager.FunctionManager, instanceSpec *types.InstanceSpecification) {
						return
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			spec := &types.InstanceSpecification{}
			manager.HandleEventUpdate(spec, types.EventKindManager)
		})
	})
}

func TestHandleEventDelete(t *testing.T) {
	convey.Convey("HandleEventDelete", t, func() {
		mockTenantID := "mock-tenant-001"
		manager, err := newInstanceManagerHelper(nil, make(chan struct{}), 1, 1)
		convey.So(manager, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)
		manager.ExclusivitySchedulerManagers = map[string]*faasschedulermanager.SchedulerManager{
			mockTenantID: {},
		}

		convey.Convey("EventKindFrontend", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&faasfrontendmanager.FrontendManager{}), "HandleInstanceDelete",
					func(_ *faasfrontendmanager.FrontendManager, instanceSpec *types.InstanceSpecification) {
						return
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			spec := &types.InstanceSpecification{}
			manager.HandleEventDelete(spec, types.EventKindFrontend)
		})

		convey.Convey("EventKindScheduler", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&faasschedulermanager.SchedulerManager{}), "HandleInstanceDelete",
					func(_ *faasschedulermanager.SchedulerManager, instanceSpec *types.InstanceSpecification) {
						return
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			spec := &types.InstanceSpecification{}
			manager.HandleEventDelete(spec, types.EventKindScheduler)
		})

		convey.Convey("EventKindSchedulerWithNotEmptyTenantID", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&faasschedulermanager.SchedulerManager{}), "HandleInstanceDelete",
					func(_ *faasschedulermanager.SchedulerManager, instanceSpec *types.InstanceSpecification) {
						return
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			spec := &types.InstanceSpecification{
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					CreateOptions: map[string]string{
						constant.SchedulerExclusivity: mockTenantID,
					},
				},
			}
			manager.HandleEventDelete(spec, types.EventKindScheduler)
		})

		convey.Convey("EventKindManager", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(&faasfunctionmanager.FunctionManager{}), "HandleInstanceDelete",
					func(_ *faasfunctionmanager.FunctionManager, instanceSpec *types.InstanceSpecification) {
						return
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			spec := &types.InstanceSpecification{}
			manager.HandleEventDelete(spec, types.EventKindManager)
		})
	})
}

func TestHandleEventRecover(t *testing.T) {
	convey.Convey("HandleEventRecover", t, func() {
		mockTenantID := "mock-tenant-001"
		manager, err := newInstanceManagerHelper(nil, make(chan struct{}), 1, 1)
		convey.So(manager, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)
		manager.ExclusivitySchedulerManagers = map[string]*faasschedulermanager.SchedulerManager{
			mockTenantID: {},
		}

		convey.Convey("EventKindFrontend", func() {
			manager.FrontendManager = &faasfrontendmanager.FrontendManager{}
			defer gomonkey.ApplyMethod(reflect.TypeOf(manager.FrontendManager), "KillInstance",
				func(ffm *faasfrontendmanager.FrontendManager, instanceID string) error {
					return nil
				}).Reset()
			spec := &types.InstanceSpecification{}
			manager.HandleEventRecover(spec, types.EventKindFrontend)
		})

		convey.Convey("EventKindScheduler", func() {
			manager.CommonSchedulerManager = &faasschedulermanager.SchedulerManager{}
			defer gomonkey.ApplyMethod(reflect.TypeOf(manager.CommonSchedulerManager), "KillInstance",
				func(ffm *faasschedulermanager.SchedulerManager, instanceID string) error {
					return nil
				}).Reset()
			spec := &types.InstanceSpecification{}
			manager.HandleEventRecover(spec, types.EventKindScheduler)
		})

		convey.Convey("EventKindSchedulerWithNotEmptyTenantID", func() {
			manager.CommonSchedulerManager = &faasschedulermanager.SchedulerManager{}
			defer gomonkey.ApplyMethod(reflect.TypeOf(manager.CommonSchedulerManager), "KillInstance",
				func(ffm *faasschedulermanager.SchedulerManager, instanceID string) error {
					return nil
				}).Reset()
			spec := &types.InstanceSpecification{
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					CreateOptions: map[string]string{
						constant.SchedulerExclusivity: mockTenantID,
					},
				},
			}
			manager.HandleEventRecover(spec, types.EventKindScheduler)
		})
	})
}
