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

package faasfrontendmanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	commonconstant "yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/redisclient"
	"yuanrong.org/kernel/pkg/common/faas_common/sts"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
	"yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	faasfrontendconf "yuanrong.org/kernel/pkg/frontend/types"
	ftypes "yuanrong.org/kernel/pkg/frontend/types"
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
	frontendConfig := types.FrontendConfig{
		Config: ftypes.Config{
			InstanceNum:          10,
			CPU:                  777,
			Memory:               777,
			SLAQuota:             1000,
			AuthenticationEnable: false,
			HTTPConfig: &ftypes.FrontendHTTP{
				MaxRequestBodySize: 6,
			},
			RouterEtcd:   routerEtcdConfig,
			MetaEtcd:     metaEtcdConfig,
			NodeSelector: map[string]string{"testkey": "testvalue"},
		},
	}
	fcConfig.UpdateFrontendConfig(&frontendConfig)
}

func newFaaSFrontendManager(size int) (*FrontendManager, error) {
	stopCh := make(chan struct{})
	initConfig(`{
		"routerEtcd": {
			"servers": ["1.2.3.4:1234"],
			"user": "tom",
			"password": "**"
		},
		"metaEtcd": {
			"servers": ["1.2.3.4:5678"],
			"user": "tom",
			"password": "**"
		},
		"alarmConfig":{"enableAlarm": true}
	}
	`)

	kv := &KvMock{}
	client := &clientv3.Client{KV: kv}
	etcdClient := &etcd3.EtcdClient{Client: client}
	manager := NewFaaSFrontendManager(&mockUtils.FakeLibruntimeSdkClient{}, etcdClient, stopCh, size, false)
	time.Sleep(50 * time.Millisecond)
	manager.count = size
	manager.instanceCache = make(map[string]*types.InstanceSpecification)
	manager.terminalCache = map[string]*types.InstanceSpecification{}
	return manager, nil
}

func newFaaSFrontendManagerWithRetry(size int) (*FrontendManager, error) {
	stopCh := make(chan struct{})
	initConfig(`{
		"routerEtcd": {
			"servers": ["1.2.3.4:1234"],
			"user": "tom",
			"password": "**"
		},
		"metaEtcd": {
			"servers": ["1.2.3.4:5678"],
			"user": "tom",
			"password": "**"
		},
		"alarmConfig":{"enableAlarm": true},
		"enableRetry": true
	}
	`)

	kv := &KvMock{}
	client := &clientv3.Client{KV: kv}
	etcdClient := &etcd3.EtcdClient{Client: client}
	manager := NewFaaSFrontendManager(&mockUtils.FakeLibruntimeSdkClient{}, etcdClient, stopCh, size, false)
	time.Sleep(50 * time.Millisecond)
	manager.count = size
	manager.instanceCache = make(map[string]*types.InstanceSpecification)
	manager.terminalCache = map[string]*types.InstanceSpecification{}
	return manager, nil
}

func Test_initInstanceCache(t *testing.T) {
	Convey("test initInstanceCache", t, func() {
		kv := &KvMock{}
		client := &clientv3.Client{KV: kv}
		etcdClient := &etcd3.EtcdClient{Client: client}
		frontendManager = &FrontendManager{
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
				bytes, _ := json.Marshal(fcConfig.GetFaaSFrontendConfig())
				meta := types.InstanceSpecificationMeta{Function: "test-function", InstanceID: "123",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}
				marshal, _ := json.Marshal(meta)
				kvs := []*mvccpb.KeyValue{{Value: marshal}}
				response := &clientv3.GetResponse{Kvs: kvs}
				response.Count = 1
				return response, nil
			}).Reset()
		frontendManager.initInstanceCache(etcdClient)
		cache := frontendManager.GetInstanceCache()
		So(cache["123"], ShouldNotBeNil)
		close(frontendManager.stopCh)
	})
}

func TestNewInstanceManager(t *testing.T) {
	Convey("Test NewInstanceManager", t, func() {
		Convey("Test NewInstanceManager with correct size", func() {
			got, err := newFaaSFrontendManager(16)
			So(err, ShouldBeNil)
			So(got, ShouldNotBeNil)
		})
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
			manager, err := newFaaSFrontendManagerWithRetry(1)
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

		instanceMgr, err := newFaaSFrontendManager(3)
		So(err, ShouldBeNil)
		defer ApplyMethod(reflect.TypeOf(instanceMgr), "CreateWithRetry",
			func(ffm *FrontendManager, ctx context.Context, args []api.Arg, extraParams *commonTypes.ExtraParams) error {
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
				func(_ *FrontendManager, ctx context.Context, _ string, _ []api.Arg, _ *commonTypes.ExtraParams) string {
					return ""
				})
			defer patches.Reset()
			err = instanceMgr.CreateMultiInstances(context.TODO(), 1)
			So(err, ShouldNotBeNil)
		})

		Convey("Test CreateMultiInstances when https enable", func() {
			fcConfig.GetFaaSFrontendConfig().HTTPSConfig = &tls.InternalHTTPSConfig{
				HTTPSEnable: true,
				TLSProtocol: "TLS",
				TLSCiphers:  "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			}
			defer func() { fcConfig.GetFaaSFrontendConfig().HTTPSConfig = nil }()

			err = instanceMgr.CreateMultiInstances(context.TODO(), 1)
			So(err, ShouldBeNil)
		})
	})
}

func TestInstanceManager_GetInstanceCountFromEtcd(t *testing.T) {
	Convey("GetInstanceCountFromEtcd", t, func() {
		Convey("failed", func() {
			instanceMgr, err := newFaaSFrontendManager(3)
			So(err, ShouldBeNil)
			patches := []*Patches{
				ApplyMethod(reflect.TypeOf(&KvMock{}), "Get",
					func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
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
			instanceMgr, err := newFaaSFrontendManager(1)
			So(err, ShouldBeNil)
			patches := []*Patches{
				ApplyMethod(reflect.TypeOf(&KvMock{}), "Get",
					func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
						response := &clientv3.GetResponse{}
						response.Count = 2
						response.Kvs = []*mvccpb.KeyValue{
							{
								Key: []byte("/sn/instance/business/yrk/tenant/default/function/0-system-faasfrontend/version/$latest/defaultaz/task-66ccf050-50f6-4835-aa24-c1b15dddb00e/12996c08-0000-4000-8000-db6c3db0fcbf"),
							},
							{
								Key: []byte("/sn/instance/business/yrk/tenant/default/function/0-system-faasfrontend/version/$latest/defaultaz/task-66ccf050-50f6-4835-aa24-c1b15dddb00e/12996c08-0000-4000-8000-db6c3db0fcb2"),
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

func TestInstanceManager_RecoverInstance(t *testing.T) {
	Convey("RecoverInstance", t, func() {
		instanceMgr, err := newFaaSFrontendManager(1)
		So(err, ShouldBeNil)
		Convey("create failed", func() {
			patches := []*Patches{
				ApplyFunc((*FrontendManager).KillInstance, func(_ *FrontendManager, _ string) error {
					return nil
				}),
				ApplyFunc((*FrontendManager).CreateMultiInstances, func(_ *FrontendManager, ctx context.Context, _ int) error {
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
		instanceMgr, err := newFaaSFrontendManager(1)
		So(err, ShouldBeNil)
		Convey("success", func() {
			instanceMgr.instanceCache["instanceID1"] = &types.InstanceSpecification{}
			instanceMgr.instanceCache["instanceID2"] = &types.InstanceSpecification{}
			err = instanceMgr.KillExceptInstance(0)
			So(err, ShouldBeNil)
			err = instanceMgr.KillExceptInstance(1)
			So(err, ShouldBeNil)
			defer ApplyMethod(reflect.TypeOf(instanceMgr), "KillInstance", func(ffm *FrontendManager, instanceID string) error {
				return fmt.Errorf("err")
			}).Reset()
			err = instanceMgr.KillExceptInstance(1)
			So(err, ShouldBeError)
		})
	})
}

func TestInstanceManager_HandleInstanceUpdate(t *testing.T) {
	Convey("HandleInstanceUpdate", t, func() {
		Convey("no need to update", func() {
			instanceMgr, err := newFaaSFrontendManager(1)
			So(err, ShouldBeNil)
			instanceMgr.HandleInstanceUpdate(&types.InstanceSpecification{
				InstanceID: "456",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function", InstanceStatus: types.InstanceStatus{
					Code: int(commonconstant.KernelInstanceStatusExiting),
				}}})
			So(instanceMgr.instanceCache["456"], ShouldEqual, nil)
		})

		Convey("not exist", func() {
			instanceMgr, err := newFaaSFrontendManager(1)
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
			instanceMgr, err := newFaaSFrontendManager(1)
			So(err, ShouldBeNil)

			instanceMgr.instanceCache = make(map[string]*types.InstanceSpecification)
			frontendConfig := fcConfig.GetFaaSFrontendConfig()
			bytes, _ := json.Marshal(frontendConfig)
			instanceMgr.HandleInstanceUpdate(&types.InstanceSpecification{
				InstanceID: "123",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}},
			)

			So(instanceMgr.instanceCache["123"].InstanceSpecificationMeta.Function, ShouldEqual, "test-function")

			instanceMgr.HandleInstanceUpdate(&types.InstanceSpecification{
				InstanceID: "123",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function", RequestID: "test-runtimeID",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}},
			)
			So(instanceMgr.instanceCache["123"].InstanceSpecificationMeta.RequestID, ShouldEqual, "test-runtimeID")
		})

		Convey("delete extra instance", func() {
			instanceMgr, err := newFaaSFrontendManager(1)
			So(err, ShouldBeNil)
			instanceMgr.instanceCache = make(map[string]*types.InstanceSpecification)

			defer ApplyMethod(reflect.TypeOf(instanceMgr.etcdClient.Client.KV), "Get",
				func(k *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
					bytes, _ := json.Marshal(fcConfig.GetFaaSFrontendConfig())
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
			frontendConfig := fcConfig.GetFaaSFrontendConfig()
			bytes, _ := json.Marshal(frontendConfig)
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
			instanceMgr, err := newFaaSFrontendManager(1)
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
			instanceMgr, err := newFaaSFrontendManager(1)
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
			size := 1
			kv := &KvMock{}
			client := &clientv3.Client{KV: kv}
			NewFaaSFrontendManager(&mockUtils.FakeLibruntimeSdkClient{}, &etcd3.EtcdClient{Client: client},
				make(chan struct{}), size, false)
			manager := GetFrontendManager()
			frontendConfig := fcConfig.GetFaaSFrontendConfig()
			bytes, _ := json.Marshal(frontendConfig)
			manager.instanceCache = map[string]*types.InstanceSpecification{
				"123": &types.InstanceSpecification{},
			}
			manager.HandleInstanceDelete(&types.InstanceSpecification{
				InstanceID: "123",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}})
			_, exist := manager.instanceCache["123"]
			So(exist, ShouldEqual, false)
		})
	})
}

func Test_recreateInstance(t *testing.T) {
	Convey("", t, func() {
		instanceMgr, err := newFaaSFrontendManager(1)
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
		So(len(instanceMgr.instanceCache) > 0, ShouldBeTrue)

		instanceMgr.recreateInstanceIDCh <- "instanceID1"
		So(len(instanceMgr.instanceCache), ShouldEqual, 1)
		close(instanceMgr.recreateInstanceIDCh)
	})
}

func TestSyncCreateInstance(t *testing.T) {
	Convey("test SyncCreateInstance", t, func() {
		manager, err := newFaaSFrontendManager(1)
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
		manager, err := newFaaSFrontendManager(2)
		So(err, ShouldBeNil)
		Convey("same config", func() {
			frontendConfig := fcConfig.GetFaaSFrontendConfig()
			bytes, _ := json.Marshal(frontendConfig)
			manager.instanceCache["test-1"] = &types.InstanceSpecification{
				InstanceID: "test-1",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			manager.instanceCache["test-2"] = &types.InstanceSpecification{
				InstanceID: "test-2",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			cfgEvent := &types.ConfigChangeEvent{
				FrontendCfg: fcConfig.GetFaaSFrontendConfig(),
				TraceID:     "traceID-123456789",
			}
			cfgEvent.FrontendCfg.InstanceNum = 2
			cfgEvent.Add(1)
			manager.ConfigChangeCh <- cfgEvent
			cfgEvent.Wait()
			So(len(manager.instanceCache), ShouldEqual, cfgEvent.FrontendCfg.InstanceNum)
			So(manager.instanceCache["test-1"], ShouldNotBeNil)
			So(manager.instanceCache["test-2"], ShouldNotBeNil)
		})

		Convey("different config", func() {
			frontendConfig := fcConfig.GetFaaSFrontendConfig()
			bytes, _ := json.Marshal(frontendConfig)
			manager.instanceCache["test-1"] = &types.InstanceSpecification{
				InstanceID: "test-1",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			manager.instanceCache["test-2"] = &types.InstanceSpecification{
				InstanceID: "test-2",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			cfg := &types.FrontendConfig{}
			utils.DeepCopyObj(fcConfig.GetFaaSFrontendConfig(), cfg)
			cfgEvent := &types.ConfigChangeEvent{
				FrontendCfg: cfg,
				TraceID:     "traceID-123456789",
			}
			cfgEvent.FrontendCfg.CPU = 888
			cfgEvent.FrontendCfg.InstanceNum = 10
			cfgEvent.Add(1)
			manager.ConfigChangeCh <- cfgEvent
			cfgEvent.Wait()
			close(manager.ConfigChangeCh)
			So(len(manager.terminalCache), ShouldEqual, 0)
			So(len(manager.instanceCache), ShouldEqual, cfgEvent.FrontendCfg.InstanceNum)
		})

		Convey("killAllTerminalInstance", func() {
			frontendConfig := fcConfig.GetFaaSFrontendConfig()
			bytes, _ := json.Marshal(frontendConfig)
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
				FrontendCfg: fcConfig.GetFaaSFrontendConfig(),
				TraceID:     "traceID-123456789",
			}
			cfgEvent.FrontendCfg.InstanceNum = 2
			cfgEvent.Add(1)
			manager.RollingUpdate(context.TODO(), cfgEvent)
			cfgEvent.Wait()
			So(len(manager.instanceCache), ShouldEqual, 2)
		})
	})
}

func Test_killAllTerminalInstance(t *testing.T) {
	Convey("killAllTerminalInstance", t, func() {
		manager, err := newFaaSFrontendManager(2)
		So(err, ShouldBeNil)
		frontendConfig := fcConfig.GetFaaSFrontendConfig()
		bytes, _ := json.Marshal(frontendConfig)
		manager.terminalCache["test-1"] = &types.InstanceSpecification{
			InstanceID: "test-1",
			InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
				Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}

		defer ApplyMethod(reflect.TypeOf(manager), "KillInstance", func(ffm *FrontendManager, instanceID string) error {
			return nil
		}).Reset()
		manager.killAllTerminalInstance()
		So(len(manager.terminalCache), ShouldEqual, 0)
	})
}

func TestConfigDiff(t *testing.T) {
	Convey("ConfigDiff", t, func() {
		manager, err := newFaaSFrontendManager(2)
		So(err, ShouldBeNil)
		Convey("same config ,different num", func() {
			cfg := &types.FrontendConfig{}
			utils.DeepCopyObj(fcConfig.GetFaaSFrontendConfig(), cfg)
			cfgEvent := &types.ConfigChangeEvent{
				FrontendCfg: cfg,
				TraceID:     "traceID-123456789",
			}
			cfgEvent.FrontendCfg.InstanceNum = 100
			_, num := manager.ConfigDiff(cfgEvent)
			So(num, ShouldEqual, 100)
		})
	})
}

func TestFrontendManager_KillInstance(t *testing.T) {
	Convey("kill instance test", t, func() {
		Convey("baseline", func() {
			manager, err := newFaaSFrontendManager(2)
			So(err, ShouldBeNil)
			i := 0
			p := ApplyFunc((*mockUtils.FakeLibruntimeSdkClient).Kill, func(_ *mockUtils.FakeLibruntimeSdkClient, instanceID string, _ int, _ []byte) error {
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

func TestFrontendManager_SyncKillAllInstance(t *testing.T) {
	Convey("kill instance test", t, func() {
		Convey("baseline", func() {
			manager, err := newFaaSFrontendManager(2)
			So(err, ShouldBeNil)
			p := ApplyFunc((*mockUtils.FakeLibruntimeSdkClient).Kill, func(_ *mockUtils.FakeLibruntimeSdkClient, instanceID string, _ int, _ []byte) error {
				return nil
			})
			defer p.Reset()
			manager.instanceCache = map[string]*types.InstanceSpecification{
				"123": &types.InstanceSpecification{},
			}
			manager.SyncKillAllInstance()
			So(len(manager.instanceCache), ShouldBeZeroValue)

		})
	})
}

func Test_prepareCreateOptions_for_InstanceLifeCycle(t *testing.T) {
	Convey("testInstanceLifeCycle", t, func() {
		conf := &types.FrontendConfig{
			Config: ftypes.Config{
				Image:        "",
				NodeSelector: map[string]string{},
			},
		}
		newFaaSFrontendManager(2)
		fcConfig.GetFaaSFrontendConfig()
		creatOpt, err := prepareCreateOptions(conf)
		So(err, ShouldBeNil)
		So(creatOpt[commonconstant.InstanceLifeCycle], ShouldEqual, commonconstant.InstanceLifeCycleDetached)
	})
}

func Test_prepareCreateOptions_for_NodeAffinity(t *testing.T) {
	Convey("testInstanceNodeAffinity", t, func() {
		tt := []struct {
			name               string
			nodeAffinity       string
			nodeAffinityPolicy string
		}{
			{
				name:               "case1",
				nodeAffinity:       "",
				nodeAffinityPolicy: "",
			},
			{
				name:               "case2",
				nodeAffinity:       "{\"requiredDuringSchedulingIgnoredDuringExecution\":{\"nodeSelectorTerms\":[{\"matchExpressions\":[{\"key\":\"node-role\",\"operator\":\"In\",\"values\":[\"edge\"]}]}]}}",
				nodeAffinityPolicy: commonconstant.DelegateNodeAffinityPolicyCoverage,
			},
			{
				name:               "case3",
				nodeAffinity:       "{\"requiredDuringSchedulingIgnoredDuringExecution\":{\"nodeSelectorTerms\":[{\"matchExpressions\":[{\"key\":\"node-role\",\"operator\":\"In\",\"values\":[\"edge-tagw\"]}]}]}}",
				nodeAffinityPolicy: commonconstant.DelegateNodeAffinityPolicyCoverage,
			},
		}
		conf := &types.FrontendConfig{
			Config: ftypes.Config{
				Image:        "",
				NodeSelector: map[string]string{},
			},
		}
		for _, ttt := range tt {
			newFaaSFrontendManager(2)
			conf.NodeAffinityPolicy = ttt.nodeAffinityPolicy
			conf.NodeAffinity = ttt.nodeAffinity
			fcConfig.GetFaaSFrontendConfig()
			creatOpt, err := prepareCreateOptions(conf)
			So(err, ShouldBeNil)
			So(creatOpt[commonconstant.InstanceLifeCycle], ShouldEqual, commonconstant.InstanceLifeCycleDetached)
			So(creatOpt[commonconstant.DelegateNodeAffinityPolicy], ShouldEqual, ttt.nodeAffinityPolicy)
			So(creatOpt[commonconstant.DelegateNodeAffinity], ShouldEqual, ttt.nodeAffinity)
		}
	})
}

func Test_createExtraParams(t *testing.T) {
	Convey("test createExtraParams", t, func() {
		Convey("baseline", func() {
			aff := "{\"nodeAffinity\":{\"preferredDuringSchedulingIgnoredDuringExecution\":[{\"preference\":{\"matchExpressions\":[{\"key\":\"node-type\",\"operator\":\"In\",\"values\":[\"system\"]}]},\"weight\":1}]}}"
			cfg := &types.FrontendConfig{
				Config: ftypes.Config{
					CPU:      100,
					Memory:   1024,
					Affinity: aff,
				},
			}
			p, err := createExtraParams(cfg)
			So(err, ShouldBeNil)
			So(p.CreateOpt[commonconstant.DelegateAffinity], ShouldEqual, aff)
		})
	})
}

func TestIsExceptInstance(t *testing.T) {
	Convey("Test isExceptInstance", t, func() {
		conf := &faasfrontendconf.Config{
			InstanceNum: 1,
			CPU:         500,
			Memory:      500,
			SLAQuota:    1000,
			Runtime:     faasfrontendconf.RuntimeConfig{},
			LocalAuth:   nil,
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
			RouterEtcd: etcd3.EtcdConfig{},
			RedisConfig: faasfrontendconf.RedisConfig{
				ClusterID:   "",
				ServerAddr:  "",
				ServerMode:  "",
				Password:    "",
				EnableTLS:   false,
				TimeoutConf: redisclient.TimeoutConf{},
			},
			HTTPConfig:              nil,
			HTTPSConfig:             nil,
			DataSystemConfig:        nil,
			BusinessType:            0,
			SccConfig:               crypto.SccConfig{},
			Image:                   "",
			MemoryControlConfig:     nil,
			MemoryEvaluatorConfig:   nil,
			DefaultTenantLimitQuota: 0,
			AuthenticationEnable:    false,
			RawStsConfig:            raw.StsConfig{},
			TrafficLimitParams:      nil,
			NodeSelector:            nil,
			AzID:                    "",
			ClusterID:               "",
			ClusterName:             "",
			AlarmConfig:             alarm.Config{},
			Version:                 "",
			AuthConfig:              faasfrontendconf.AuthConfig{},
			FunctionNameSeparator:   "",
			AlarmServerAddress:      "",
			InvokeMaxRetryTimes:     0,
			EtcdLeaseConfig:         nil,
			HeartbeatConfig:         nil,
			E2EMaxDelayTime:         0,
			RetryConfig:             nil,
			ShareKeys:               faasfrontendconf.ShareKeys{},
			Affinity:                "",
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
		patches := ApplyFunc(controllerutils.GetFrontendConfigSignature, func(frontendCfg *types.FrontendConfig) string {
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
