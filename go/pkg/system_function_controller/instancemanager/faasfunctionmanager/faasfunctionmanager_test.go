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

package faasfunctionmanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"reflect"
	"sync"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	commonconstant "yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/sts"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
	commontype "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	faasmanagerconf "yuanrong.org/kernel/pkg/functionmanager/types"
	fcConfig "yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
	controllerutils "yuanrong.org/kernel/pkg/system_function_controller/utils"
)

type KvMock struct {
}

func (k *KvMock) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (k *KvMock) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	response := &clientv3.GetResponse{}
	response.Count = 2
	return response, nil
}

func (k *KvMock) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (k *KvMock) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse,
	error) {
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

func initConfig(configString string) {
	fcConfig.InitConfig([]byte(configString))
	routerEtcdConfig := etcd3.EtcdConfig{
		Servers:  []string{"1.2.3.4:1234"},
		User:     "tom",
		Password: "**",
	}
	metaEtcdConfig := etcd3.EtcdConfig{
		Servers:  []string{"1.2.3.4:5678"},
		User:     "tom",
		Password: "**",
	}
	managerConfig := types.ManagerConfig{
		ManagerInstanceNum: 10,
		CPU:                777,
		Memory:             777,
		RouterEtcd:         routerEtcdConfig,
		MetaEtcd:           metaEtcdConfig,
		NodeSelector:       map[string]string{"testkey": "testvalue"},
	}
	fcConfig.UpdateManagerConfig(&managerConfig)
}

func newFaaSFunctionManager(size int) (*FunctionManager, error) {
	stopCh := make(chan struct{})
	initConfig(`{
		"faasfunctionConfig": {
			"slaQuota": 1000,
			"functionCapability": 1,
			"authenticationEnable": false,
			"trafficLimitDisable": true,
			"http": {
                "resptimeout": 5,
                "workerInstanceReadTimeOut": 5,
                "maxRequestBodySize": 6
            }
		}
	}
	`)

	kv := &KvMock{}
	client := &clientv3.Client{KV: kv}
	etcdClient := &etcd3.EtcdClient{Client: client}
	once = sync.Once{}
	manager := NewFaaSFunctionManager(&mockUtils.FakeLibruntimeSdkClient{}, etcdClient, stopCh, size)
	time.Sleep(50 * time.Millisecond)
	manager.count = size
	manager.instanceCache = make(map[string]*types.InstanceSpecification)
	manager.terminalCache = map[string]*types.InstanceSpecification{}
	functionManager = manager
	return functionManager, nil
}

func newFaaSFunctionManagerWithRetry(size int) (*FunctionManager, error) {
	stopCh := make(chan struct{})
	initConfig(`{
		"faasfunctionConfig": {
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
		"enableRetry": true
	}
	`)

	kv := &KvMock{}
	client := &clientv3.Client{KV: kv}
	etcdClient := &etcd3.EtcdClient{Client: client}
	once = sync.Once{}
	manager := NewFaaSFunctionManager(&mockUtils.FakeLibruntimeSdkClient{}, etcdClient, stopCh, size)
	time.Sleep(50 * time.Millisecond)
	manager.count = size
	manager.instanceCache = make(map[string]*types.InstanceSpecification)
	manager.terminalCache = map[string]*types.InstanceSpecification{}
	functionManager = manager
	return functionManager, nil
}

func TestNewInstanceManager(t *testing.T) {
	Convey("Test NewInstanceManager", t, func() {
		Convey("Test NewInstanceManager with correct size", func() {
			got, err := newFaaSFunctionManager(16)
			So(err, ShouldBeNil)
			So(got, ShouldNotBeNil)
		})
	})
}

func Test_initInstanceCache(t *testing.T) {
	Convey("test initInstanceCache", t, func() {
		kv := &KvMock{}
		client := &clientv3.Client{KV: kv}
		etcdClient := &etcd3.EtcdClient{Client: client}
		functionManager = &FunctionManager{
			instanceCache:        make(map[string]*types.InstanceSpecification),
			terminalCache:        map[string]*types.InstanceSpecification{},
			etcdClient:           etcdClient,
			sdkClient:            &mockUtils.FakeLibruntimeSdkClient{},
			count:                1,
			stopCh:               make(chan struct{}),
			recreateInstanceIDCh: make(chan string, 1),
			ConfigChangeCh:       make(chan *types.ConfigChangeEvent, 100),
		}
		defer ApplyMethod(reflect.TypeOf(kv), "Get",
			func(k *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				bytes, _ := json.Marshal(fcConfig.GetFaaSManagerConfig())
				meta := types.InstanceSpecificationMeta{Function: "test-function", InstanceID: "123",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}
				marshal, _ := json.Marshal(meta)
				kvs := []*mvccpb.KeyValue{{Value: marshal}}
				response := &clientv3.GetResponse{Kvs: kvs}
				response.Count = 1
				return response, nil
			}).Reset()
		functionManager.initInstanceCache(etcdClient)
		cache := functionManager.GetInstanceCache()
		So(cache["123"], ShouldNotBeNil)
		close(functionManager.stopCh)
	})
}

func TestInstanceManager_CreateMultiInstances(t *testing.T) {
	Convey("Test CreateMultiInstances", t, func() {
		Convey("Test CreateMultiInstances with retry", func() {
			p := ApplyFunc((*mockUtils.FakeLibruntimeSdkClient).CreateInstance, func(_ *mockUtils.FakeLibruntimeSdkClient, _ api.FunctionMeta, _ []api.Arg, _ api.InvokeOptions) (string, error) {
				time.Sleep(100 * time.Millisecond)
				return "", api.ErrorInfo{
					Code:            10,
					Err:             fmt.Errorf("xxxxx"),
					StackTracesInfo: api.StackTracesInfo{},
				}
			})
			p2 := ApplyFunc(wait.ExponentialBackoffWithContext, func(ctx context.Context, backoff wait.Backoff, condition wait.ConditionWithContextFunc) error {
				_, err := condition(ctx)
				return err
			})
			manager, err := newFaaSFunctionManagerWithRetry(1)
			So(err, ShouldBeNil)
			ctx, cancel := context.WithCancel(context.TODO())
			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()
			err = manager.CreateMultiInstances(ctx, 1)
			So(err, ShouldBeError)

			err = manager.CreateMultiInstances(context.TODO(), 1)
			So(err, ShouldBeNil)
			p.Reset()
			p2.Reset()
		})

		instanceMgr, err := newFaaSFunctionManager(3)
		So(err, ShouldBeNil)
		defer ApplyMethod(reflect.TypeOf(instanceMgr), "CreateWithRetry",
			func(fm *FunctionManager, ctx context.Context, args []api.Arg, extraParams commontype.ExtraParams) error {
				if value := ctx.Value("err"); value != nil {
					if s := value.(string); s == "canceled" {
						return fmt.Errorf("create has been cancelled")
					}
				}
				return nil
			}).Reset()

		Convey("Test CreateMultiInstances when passed different count", func() {
			err = instanceMgr.CreateMultiInstances(context.TODO(), -1)
			So(err, ShouldBeNil)

			err = instanceMgr.CreateMultiInstances(context.TODO(), 1)
			So(err, ShouldBeNil)
		})

		Convey("Test CreateMultiInstances when failed to get scheduler config", func() {
			patches := ApplyFunc(json.Marshal, func(_ interface{}) ([]byte, error) {
				return nil, errors.New("json Marshal failed")
			})
			defer patches.Reset()
			err = instanceMgr.CreateMultiInstances(context.TODO(), 1)
			So(err, ShouldNotBeNil)
		})

		Convey("Test CreateMultiInstances when failed to create instance", func() {
			patches := ApplyMethod(reflect.TypeOf(instanceMgr), "CreateInstance",
				func(_ *FunctionManager, ctx context.Context, _ string, _ []api.Arg, _ *commontype.ExtraParams) string {
					return ""
				})
			defer patches.Reset()
			err = instanceMgr.CreateMultiInstances(context.TODO(), 1)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestInstanceManager_GetInstanceCountFromEtcd(t *testing.T) {
	Convey("GetInstanceCountFromEtcd", t, func() {
		Convey("failed", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)
			patches := []*Patches{
				ApplyMethod(reflect.TypeOf(&KvMock{}), "Get",
					func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse,
						error) {
						return nil, errors.New("get etcd error")
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			count := instanceMgr.GetInstanceCountFromEtcd()
			So(len(count), ShouldEqual, 0)
		})
	})

}

func TestInstanceManager_CreateExpectedInstanceCount(t *testing.T) {
	Convey("CreateExpectedInstanceCount", t, func() {
		Convey("success", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)
			patches := []*Patches{
				ApplyMethod(reflect.TypeOf(&KvMock{}), "Get",
					func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse,
						error) {
						response := &clientv3.GetResponse{}
						response.Count = 2
						response.Kvs = []*mvccpb.KeyValue{
							{
								Key: []byte("/sn/instance/business/yrk/tenant/default/function/0-system-faasmanager/version/$latest/defaultaz/task-66ccf050-50f6-4835-aa24-c1b15dddb00e/12996c08-0000-4000-8000-db6c3db0fcbf"),
							},
							{
								Key: []byte("/sn/instance/business/yrk/tenant/default/function/0-system-faasmanager/version/$latest/defaultaz/task-66ccf050-50f6-4835-aa24-c1b15dddb00e/12996c08-0000-4000-8000-db6c3db0fcb2"),
							},
						}
						return response, nil
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()
			count := instanceMgr.GetInstanceCountFromEtcd()
			So(len(count), ShouldEqual, 2)

			err = instanceMgr.CreateExpectedInstanceCount(context.TODO())
			So(err, ShouldBeNil)
		})
	})
}

func TestInstanceManager_SyncKillAllInstance(t *testing.T) {
	Convey("SyncKillAllInstance", t, func() {
		Convey("success", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)
			patches := []*Patches{
				ApplyMethod(reflect.TypeOf(&mockUtils.FakeLibruntimeSdkClient{}), "Kill",
					func(_ *mockUtils.FakeLibruntimeSdkClient, instanceID string, signal int, payload []byte) error {
						return nil
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()

			instanceMgr.instanceCache = map[string]*types.InstanceSpecification{
				"123": &types.InstanceSpecification{},
			}
			instanceMgr.SyncKillAllInstance()
			So(len(instanceMgr.instanceCache), ShouldEqual, 0)
			_, exist := instanceMgr.instanceCache["123"]
			So(exist, ShouldEqual, false)
		})
	})
}

func TestInstanceManager_RecoverInstance(t *testing.T) {
	Convey("RecoverInstance", t, func() {
		Convey("create failed", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)
			patches := []*Patches{
				ApplyFunc((*FunctionManager).KillInstance, func(_ *FunctionManager, _ string) error {
					return nil
				}),
				ApplyFunc((*FunctionManager).CreateMultiInstances,
					func(_ *FunctionManager, ctx context.Context, _ int) error {
						return errors.New("failed to create instances")
					}),
			}
			defer func() {
				for _, patch := range patches {
					time.Sleep(100 * time.Millisecond)
					patch.Reset()
				}
			}()

			instanceMgr.RecoverInstance(&types.InstanceSpecification{})
		})
	})
}

func TestKillExceptInstance(t *testing.T) {
	Convey("KillExceptInstance", t, func() {
		instanceMgr, err := newFaaSFunctionManager(1)
		So(err, ShouldBeNil)
		Convey("success", func() {
			instanceMgr.instanceCache["instanceID1"] = &types.InstanceSpecification{}
			instanceMgr.instanceCache["instanceID2"] = &types.InstanceSpecification{}
			err = instanceMgr.KillExceptInstance(2)
			So(err, ShouldBeNil)
		})
	})
}

func TestInstanceManager_HandleInstanceUpdate(t *testing.T) {
	Convey("HandleInstanceUpdate", t, func() {
		Convey("not exist", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)

			instanceMgr.instanceCache = map[string]*types.InstanceSpecification{
				"123": &types.InstanceSpecification{},
			}

			instanceMgr.HandleInstanceUpdate(&types.InstanceSpecification{
				InstanceID:                "456",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function"}})

			So(instanceMgr.instanceCache["456"], ShouldEqual, nil)
		})

		Convey("exist", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)

			instanceMgr.instanceCache = make(map[string]*types.InstanceSpecification)
			managerConfig := fcConfig.GetFaaSManagerConfig()
			bytes, _ := json.Marshal(managerConfig)
			instanceMgr.HandleInstanceUpdate(&types.InstanceSpecification{
				InstanceID: "123",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}},
			)

			So(instanceMgr.instanceCache["123"].InstanceSpecificationMeta.Function, ShouldEqual, "test-function")

			instanceMgr.HandleInstanceUpdate(&types.InstanceSpecification{
				InstanceID: "123",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					RequestID: "test-runtimeID",
					Args:      []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}},
			)
			So(instanceMgr.instanceCache["123"].InstanceSpecificationMeta.RequestID, ShouldEqual, "test-runtimeID")
		})

		Convey("delete extra instance", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)
			instanceMgr.instanceCache = make(map[string]*types.InstanceSpecification)

			defer ApplyMethod(reflect.TypeOf(instanceMgr.etcdClient.Client.KV), "Get",
				func(k *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse,
					error) {
					bytes, _ := json.Marshal(fcConfig.GetFaaSManagerConfig())
					meta := types.InstanceSpecificationMeta{Function: "test-function", InstanceID: "test-function",
						Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}
					marshal, _ := json.Marshal(meta)
					kvs := []*mvccpb.KeyValue{{Value: marshal}}
					response := &clientv3.GetResponse{Kvs: kvs}
					response.Count = 1
					return response, nil
				}).Reset()
			instanceMgr.initInstanceCache(instanceMgr.etcdClient)
			So(instanceMgr.instanceCache["test-function"], ShouldNotBeNil)
			managerConfig := fcConfig.GetFaaSManagerConfig()
			bytes, _ := json.Marshal(managerConfig)
			instanceMgr.HandleInstanceUpdate(&types.InstanceSpecification{
				InstanceID: "123",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}},
			)
			So(instanceMgr.instanceCache["123"], ShouldBeNil)
		})
	})
}

func TestInstanceManager_HandleInstanceDelete(t *testing.T) {
	Convey("HandleInstanceDelete", t, func() {
		Convey("not exist", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)

			instanceMgr.instanceCache = map[string]*types.InstanceSpecification{
				"123": &types.InstanceSpecification{},
			}
			instanceMgr.recreateInstanceIDMap.Store("456", "error type")
			instanceMgr.HandleInstanceDelete(&types.InstanceSpecification{
				InstanceID:                "456",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function"}})
			_, exist := instanceMgr.instanceCache["123"]
			So(exist, ShouldEqual, true)
		})

		Convey("not exist2", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)

			instanceMgr.instanceCache = map[string]*types.InstanceSpecification{
				"123": &types.InstanceSpecification{},
			}
			_, cancel := context.WithCancel(context.Background())
			instanceMgr.recreateInstanceIDMap.Store("456", cancel)
			instanceMgr.HandleInstanceDelete(&types.InstanceSpecification{
				InstanceID:                "456",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function"}})
			_, exist := instanceMgr.instanceCache["123"]
			So(exist, ShouldEqual, true)
		})

		Convey("exist", func() {
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)

			instanceMgr.instanceCache = map[string]*types.InstanceSpecification{
				"123": &types.InstanceSpecification{},
			}
			instanceMgr.HandleInstanceDelete(&types.InstanceSpecification{
				InstanceID:                "123",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function"}})
			_, exist := instanceMgr.instanceCache["123"]
			So(exist, ShouldEqual, false)
		})

		Convey("is not except instance", func() {
			defer ApplyFunc(isExceptInstance, func(meta *types.InstanceSpecificationMeta, targetSign string) bool {
				return true
			}).Reset()
			instanceMgr, err := newFaaSFunctionManager(1)
			So(err, ShouldBeNil)

			instanceMgr.instanceCache = map[string]*types.InstanceSpecification{
				"123": &types.InstanceSpecification{},
			}
			instanceMgr.HandleInstanceDelete(&types.InstanceSpecification{
				InstanceID:                "123",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function"}})
			_, exist := instanceMgr.instanceCache["123"]
			So(exist, ShouldEqual, false)
		})
	})
}

func TestFunctionManager_recreateInstance(t *testing.T) {
	Convey("test recreate instance", t, func() {
		instanceMgr, err := newFaaSFunctionManager(1)
		So(err, ShouldBeNil)
		defer ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
			return []byte{}, nil
		}).Reset()
		defer ApplyFunc(fcConfig.GetFaaSControllerConfig, func() *types.Config {
			return &types.Config{
				RawStsConfig: raw.StsConfig{StsEnable: true},
			}
		}).Reset()
		defer ApplyFunc(sts.GenerateSecretVolumeMounts, func(systemFunctionName string) ([]byte, error) {
			return []byte{}, nil
		}).Reset()
		instanceMgr.recreateInstanceIDCh <- "instanceID1"
		go instanceMgr.recreateInstance()
		time.Sleep(11 * time.Second)
		_, ok := instanceMgr.recreateInstanceIDMap.Load("instanceID1")
		So(ok, ShouldBeFalse)

		instanceMgr.recreateInstanceIDCh <- "instanceID1"
		So(len(instanceMgr.instanceCache), ShouldEqual, 1)
		close(instanceMgr.recreateInstanceIDCh)
	})
}

func TestSyncCreateInstance(t *testing.T) {
	Convey("test SyncCreateInstance", t, func() {
		manager, err := newFaaSFunctionManager(1)
		So(err, ShouldBeNil)
		err = manager.SyncCreateInstance(context.TODO())
		So(err, ShouldBeNil)

		Convey("context canceled", func() {
			ctx, cancelFunc := context.WithCancel(context.TODO())
			cancelFunc()
			err = manager.SyncCreateInstance(ctx)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestRollingUpdate(t *testing.T) {
	Convey("test RollingUpdate", t, func() {
		manager, err := newFaaSFunctionManager(2)
		So(err, ShouldBeNil)
		Convey("same config", func() {
			managerConfig := fcConfig.GetFaaSManagerConfig()
			bytes, _ := json.Marshal(managerConfig)
			manager.instanceCache["test-1"] = &types.InstanceSpecification{
				InstanceID: "test-1",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			manager.instanceCache["test-2"] = &types.InstanceSpecification{
				InstanceID: "test-2",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			cfgEvent := &types.ConfigChangeEvent{
				ManagerCfg: fcConfig.GetFaaSManagerConfig(),
				TraceID:    "traceID-123456789",
			}
			cfgEvent.ManagerCfg.ManagerInstanceNum = 2
			cfgEvent.Add(1)
			manager.ConfigChangeCh <- cfgEvent
			cfgEvent.Wait()
			So(len(manager.instanceCache), ShouldEqual, cfgEvent.ManagerCfg.ManagerInstanceNum)
			So(manager.instanceCache["test-1"], ShouldNotBeNil)
			So(manager.instanceCache["test-2"], ShouldNotBeNil)
		})

		Convey("different config", func() {
			managerConfig := fcConfig.GetFaaSManagerConfig()
			bytes, _ := json.Marshal(managerConfig)
			manager.instanceCache["test-1"] = &types.InstanceSpecification{
				InstanceID: "test-1",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			manager.instanceCache["test-2"] = &types.InstanceSpecification{
				InstanceID: "test-2",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			cfg := &types.ManagerConfig{}
			utils.DeepCopyObj(fcConfig.GetFaaSManagerConfig(), cfg)
			cfgEvent := &types.ConfigChangeEvent{
				ManagerCfg: cfg,
				TraceID:    "traceID-123456789",
			}
			cfgEvent.ManagerCfg.CPU = 888
			cfgEvent.ManagerCfg.ManagerInstanceNum = 10
			cfgEvent.Add(1)
			manager.ConfigChangeCh <- cfgEvent
			cfgEvent.Wait()
			close(manager.ConfigChangeCh)
			So(len(manager.terminalCache), ShouldEqual, 0)
			So(len(manager.instanceCache), ShouldEqual, cfgEvent.ManagerCfg.ManagerInstanceNum)
		})

		Convey("killAllTerminalInstance", func() {
			managerConfig := fcConfig.GetFaaSManagerConfig()
			bytes, _ := json.Marshal(managerConfig)
			manager.instanceCache = make(map[string]*types.InstanceSpecification)
			manager.instanceCache["test-1"] = &types.InstanceSpecification{
				InstanceID: "test-1",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			manager.instanceCache["test-2"] = &types.InstanceSpecification{
				InstanceID: "test-2",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			cfgEvent := &types.ConfigChangeEvent{
				ManagerCfg: fcConfig.GetFaaSManagerConfig(),
				TraceID:    "traceID-123456789",
			}
			cfgEvent.ManagerCfg.ManagerInstanceNum = 2
			cfgEvent.Add(1)
			manager.RollingUpdate(context.TODO(), cfgEvent)
			cfgEvent.Wait()
			So(len(manager.instanceCache), ShouldEqual, 2)
		})
	})
}

func Test_killAllTerminalInstance(t *testing.T) {
	Convey("killAllTerminalInstance", t, func() {
		manager, err := newFaaSFunctionManager(2)
		So(err, ShouldBeNil)
		managerConfig := fcConfig.GetFaaSManagerConfig()
		bytes, _ := json.Marshal(managerConfig)
		manager.terminalCache["test-1"] = &types.InstanceSpecification{
			InstanceID: "test-1",
			InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
				Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}

		defer ApplyMethod(reflect.TypeOf(manager), "KillInstance", func(fm *FunctionManager, instanceID string) error {
			return nil
		}).Reset()
		manager.killAllTerminalInstance()
		So(len(manager.terminalCache), ShouldEqual, 0)
	})
}

func TestConfigDiff(t *testing.T) {
	Convey("ConfigDiff", t, func() {
		manager, err := newFaaSFunctionManager(2)
		So(err, ShouldBeNil)
		Convey("same config ,different num", func() {
			cfg := &types.ManagerConfig{}
			utils.DeepCopyObj(fcConfig.GetFaaSManagerConfig(), cfg)
			cfgEvent := &types.ConfigChangeEvent{
				ManagerCfg: cfg,
				TraceID:    "traceID-123456789",
			}
			cfgEvent.ManagerCfg.ManagerInstanceNum = 100
			_, num := manager.ConfigDiff(cfgEvent)
			So(num, ShouldEqual, 100)
		})
	})
}

func TestFunctionManager_KillExceptInstance(t *testing.T) {
	Convey("KillExceptInstance", t, func() {
		instanceMgr, err := newFaaSFunctionManager(1)
		So(err, ShouldBeNil)
		err = instanceMgr.KillExceptInstance(1)
		So(err, ShouldBeNil)
	})
}

func TestFunctionManager_KillInstance(t *testing.T) {
	Convey("kill instance test", t, func() {
		Convey("baseline", func() {
			manager, err := newFaaSFunctionManager(2)
			So(err, ShouldBeNil)
			i := 0
			p := ApplyFunc((*mockUtils.FakeLibruntimeSdkClient).Kill,
				func(_ *mockUtils.FakeLibruntimeSdkClient, instanceID string, _ int, _ []byte) error {
					if i == 0 {
						i = 1
						return fmt.Errorf("error")
					}
					return nil
				})
			defer p.Reset()
			err = manager.KillInstance("aaa")
			So(err, ShouldBeNil)
		})
	})
}

func TestFunctionManager_CreateWithRetry(t *testing.T) {
	Convey("create with retry test", t, func() {
		Convey("baseline", func() {
			instanceMgr, _ := newFaaSFunctionManager(1)
			p := ApplyFunc((*FunctionManager).CreateInstance,
				func(_ *FunctionManager, ctx context.Context, function string,
					args []api.Arg, extraParams *commontype.ExtraParams) string {
					return "aaa"
				})
			defer p.Reset()
			err := instanceMgr.CreateWithRetry(context.TODO(), nil, commontype.ExtraParams{})
			So(err, ShouldBeNil)
		})
		Convey("retry success", func() {
			instanceMgr, _ := newFaaSFunctionManager(1)
			ins := ""
			p := ApplyFunc((*FunctionManager).CreateInstance,
				func(_ *FunctionManager, ctx context.Context, function string,
					args []api.Arg, extraParams *commontype.ExtraParams) string {
					if ins == "" {
						ins = "aaa"
						return ""
					}
					return ins
				})
			defer p.Reset()
			err := instanceMgr.CreateWithRetry(context.TODO(), nil, commontype.ExtraParams{})
			So(err, ShouldBeNil)
		})
	})
}

func Test_prepareCreateOptions(t *testing.T) {
	Convey("prepareCreateOptions test", t, func() {
		Convey("baseline", func() {
			p := ApplyFunc(fcConfig.GetFaaSManagerConfig, func() *types.ManagerConfig {
				return &types.ManagerConfig{
					Affinity: "aaaa",
				}
			})
			defer p.Reset()
			options, err := prepareCreateOptions(&types.ManagerConfig{
				Affinity: "aaaa",
			})
			So(err, ShouldBeNil)
			So(options[commonconstant.DelegateAffinity], ShouldEqual, "aaaa")
		})

		Convey("baseline1", func() {
			nodeAffinity := "{\"nodeAffinity\":{\"preferredDuringSchedulingIgnoredDuringExecution\":[{\"preference\":{\"matchExpressions\":[{\"key\":\"node-type\",\"operator\":\"In\",\"values\":[\"system\"]}]},\"weight\":1}]}}"
			nodeAffinityPolicy := "coverage"
			p := ApplyFunc(fcConfig.GetFaaSManagerConfig, func() *types.ManagerConfig {
				return &types.ManagerConfig{
					NodeAffinity:       nodeAffinity,
					NodeAffinityPolicy: nodeAffinityPolicy,
				}
			})
			defer p.Reset()
			options, err := prepareCreateOptions(&types.ManagerConfig{
				NodeAffinity:       nodeAffinity,
				NodeAffinityPolicy: nodeAffinityPolicy,
			})
			So(err, ShouldBeNil)
			So(options[commonconstant.DelegateNodeAffinity], ShouldEqual, nodeAffinity)
			So(options[commonconstant.DelegateNodeAffinityPolicy], ShouldEqual, nodeAffinityPolicy)
		})
		Convey("baseline2", func() {
			nodeAffinity := "{\"nodeAffinity\":{\"preferredDuringSchedulingIgnoredDuringExecution\":[{\"preference\":{\"matchExpressions\":[{\"key\":\"node-type\",\"operator\":\"In\",\"values\":[\"system\"]}]},\"weight\":1}]}}"
			nodeAffinityPolicy := "aggregation"
			p := ApplyFunc(fcConfig.GetFaaSManagerConfig, func() *types.ManagerConfig {
				return &types.ManagerConfig{
					NodeAffinity:       nodeAffinity,
					NodeAffinityPolicy: nodeAffinityPolicy,
				}
			})
			defer p.Reset()
			options, err := prepareCreateOptions(&types.ManagerConfig{
				NodeAffinity:       nodeAffinity,
				NodeAffinityPolicy: nodeAffinityPolicy,
			})
			So(err, ShouldBeNil)
			So(options[commonconstant.DelegateNodeAffinity], ShouldEqual, nodeAffinity)
			So(options[commonconstant.DelegateNodeAffinityPolicy], ShouldEqual, nodeAffinityPolicy)
		})
	})
}

func TestIsExceptInstance(t *testing.T) {
	Convey("Test isExceptInstance", t, func() {
		conf := &faasmanagerconf.ManagerConfig{
			MetaEtcd: etcd3.EtcdConfig{
				Servers:        []string{"127.0.0.1:32379"},
				User:           "",
				Password:       "",
				SslEnable:      false,
				AuthType:       "Noauth",
				UseSecret:      false,
				SecretName:     "etcd-client-secret",
				LimitRate:      0,
				LimitBurst:     0,
				LimitTimeout:   0,
				CaFile:         "",
				CertFile:       "",
				KeyFile:        "",
				PassphraseFile: "",
			},
			RouterEtcd:           etcd3.EtcdConfig{},
			FunctionCapability:   0,
			SccConfig:            crypto.SccConfig{},
			AuthenticationEnable: false,
			AlarmConfig:          alarm.Config{},
		}
		confStr, err := json.Marshal(conf)
		if err != nil {
			return
		}

		meta := &types.InstanceSpecificationMeta{
			Args: []map[string]string{},
		}
		targetSign := "testSign"
		if isExceptInstance(meta, targetSign) {
			t.Errorf("Expected isExceptInstance to return false when args is empty")
		}
		meta.Args = []map[string]string{
			{
				"value": "invalidBase64",
			},
		}
		if isExceptInstance(meta, targetSign) {
			t.Errorf("Expected isExceptInstance to return false when base64 decoding fails")
		}

		// 测试参数反序列化为前端配置失败的情况
		meta.Args[0]["value"] = base64.StdEncoding.EncodeToString([]byte("invalidJson"))
		if isExceptInstance(meta, targetSign) {
			t.Errorf("Expected isExceptInstance to return false when unmarshalling frontend config fails")
		}

		meta.Args[0]["value"] = base64.StdEncoding.EncodeToString([]byte(confStr))
		patches := ApplyFunc(controllerutils.GetManagerConfigSignature, func(managerCfg *types.ManagerConfig) string {
			return targetSign
		})
		defer patches.Reset()
		if !isExceptInstance(meta, targetSign) {
			t.Errorf("Expected isExceptInstance to return true")
		}
		newStr := make([]byte, len(confStr)+16)
		newStr = append([]byte("0000000000000000"), confStr[:]...)
		meta.Args[0]["value"] = base64.StdEncoding.EncodeToString(newStr)
		if !isExceptInstance(meta, targetSign) {
			t.Errorf("Expected isExceptInstance to return true")
		}

		newStr = append([]byte("0000000000000000"), newStr[:]...)
		meta.Args[0]["value"] = base64.StdEncoding.EncodeToString(newStr)
		if isExceptInstance(meta, targetSign) {
			t.Errorf("Expected isExceptInstance to return false")
		}
	})

}
