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

package faasschedulermanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"yuanrong.org/kernel/runtime/libruntime/api"

	commonconstant "yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	commontype "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	stypes "yuanrong.org/kernel/pkg/functionscaler/types"
	fcConfig "yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/state"
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

	schedulerBasicConfig := types.SchedulerConfig{
		Configuration: stypes.Configuration{
			CPU:    999,
			Memory: 999,
			AutoScaleConfig: stypes.AutoScaleConfig{
				SLAQuota:      1000,
				ScaleDownTime: 60000,
				BurstScaleNum: 1000,
			},
			LeaseSpan:        600000,
			RouterETCDConfig: routerEtcdConfig,
			MetaETCDConfig:   metaEtcdConfig,
		},
		SchedulerNum: 10,
	}
	fcConfig.UpdateSchedulerConfig(&schedulerBasicConfig)
}

func newFaaSSchedulerManager(size int) (*SchedulerManager, error) {
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
	manager := NewFaaSSchedulerManager(&mockUtils.FakeLibruntimeSdkClient{}, etcdClient, stopCh, size, "")
	time.Sleep(50 * time.Millisecond)
	manager.count = size
	manager.instanceCache = make(map[string]*types.InstanceSpecification)
	manager.terminalCache = map[string]*types.InstanceSpecification{}
	return manager, nil
}

func newFaaSSchedulerManagerWithRetry(size int) (*SchedulerManager, error) {
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
	manager := NewFaaSSchedulerManager(&mockUtils.FakeLibruntimeSdkClient{}, etcdClient, stopCh, size, "")
	time.Sleep(50 * time.Millisecond)
	manager.count = size
	manager.instanceCache = make(map[string]*types.InstanceSpecification)
	manager.terminalCache = map[string]*types.InstanceSpecification{}
	return manager, nil
}

func TestNewInstanceManager(t *testing.T) {
	Convey("Test NewInstanceManager", t, func() {
		Convey("Test NewInstanceManager with correct size", func() {
			got, err := newFaaSSchedulerManager(16)
			So(err, ShouldBeNil)
			So(got, ShouldNotBeNil)
		})
	})
}

func Test_initInstanceCache(t *testing.T) {
	Convey("initInstanceCache", t, func() {
		manager, err := newFaaSSchedulerManager(1)
		So(err, ShouldBeNil)
		kv := &KvMock{}
		client := &clientv3.Client{KV: kv}
		etcdClient := &etcd3.EtcdClient{Client: client}
		defer ApplyMethod(reflect.TypeOf(kv), "Get",
			func(k *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				bytes, _ := json.Marshal(fcConfig.GetFaaSSchedulerConfig())
				meta := types.InstanceSpecificationMeta{Function: "test-function", InstanceID: "123",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}
				marshal, _ := json.Marshal(meta)
				kvs := []*mvccpb.KeyValue{{Value: marshal}}
				response := &clientv3.GetResponse{Kvs: kvs}
				response.Count = 1
				return response, nil
			}).Reset()
		manager.initInstanceCache(etcdClient)
		cache := manager.GetInstanceCache()
		So(cache["123"], ShouldNotBeNil)
		close(manager.stopCh)
	})
}

func TestInstanceManager_CreateMultiInstances(t *testing.T) {
	defer ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
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
			manager, err := newFaaSSchedulerManagerWithRetry(1)
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

		instanceMgr, err := newFaaSSchedulerManager(3)
		So(err, ShouldBeNil)
		defer ApplyMethod(reflect.TypeOf(instanceMgr), "CreateWithRetry",
			func(ffm *SchedulerManager, ctx context.Context, args []api.Arg,
				extraParams *commontype.ExtraParams) error {
				if value := ctx.Value("err"); value != nil {
					if s := value.(string); s == "canceled" {
						return fmt.Errorf("create has been cancelled")
					}
				}
				return nil
			}).Reset()
		Convey("Test CreateMultiInstances when passed different count", func() {
			err = instanceMgr.CreateMultiInstances(context.TODO(), 1)
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
				func(_ *SchedulerManager, ctx context.Context, function string, args []api.Arg,
					extraParams *commontype.ExtraParams) string {
					return ""
				})
			defer patches.Reset()
			err = instanceMgr.CreateMultiInstances(context.TODO(), 1)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestCreateWithRetry(t *testing.T) {
	Convey("CreateWithRetry", t, func() {
		Convey("retry", func() {
			instanceMgr, err := newFaaSSchedulerManager(1)
			So(err, ShouldBeNil)

			defer ApplyMethod(reflect.TypeOf(instanceMgr), "CreateInstance",
				func(s *SchedulerManager, ctx context.Context, function string, args []api.Arg,
					extraParams *commontype.ExtraParams) string {
					select {
					case <-ctx.Done():
						return "cancelled"
					default:
						return ""
					}
				}).Reset()
			args, extraParams, _ := genFunctionConfig("")
			ctx, cancelFunc := context.WithCancel(context.TODO())
			go func() {
				time.Sleep(2 * time.Second)
				cancelFunc()
			}()
			err = instanceMgr.CreateWithRetry(ctx, args, extraParams)
			So(err, ShouldBeError)
		})
	})
}

func TestInstanceManager_SyncKillAllInstance(t *testing.T) {
	Convey("KillAllInstance", t, func() {
		Convey("success", func() {
			instanceMgr, err := newFaaSSchedulerManager(1)
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

func TestInstanceManager_GetInstanceCountFromEtcd(t *testing.T) {
	Convey("GetInstanceCountFromEtcd", t, func() {
		Convey("failed", func() {
			instanceMgr, err := newFaaSSchedulerManager(1)
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
			instanceMgr, err := newFaaSSchedulerManager(2)
			So(err, ShouldBeNil)
			patches := []*Patches{
				ApplyMethod(reflect.TypeOf(&KvMock{}), "Get",
					func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse,
						error) {
						response := &clientv3.GetResponse{}
						response.Count = 2
						response.Kvs = []*mvccpb.KeyValue{
							{
								Key: []byte("/sn/instance/business/yrk/tenant/default/function/0-system-faasscheduler/version/$latest/defaultaz/task-66ccf050-50f6-4835-aa24-c1b15dddb00e/12996c08-0000-4000-8000-db6c3db0fcbf"),
							},
							{
								Key: []byte("/sn/instance/business/yrk/tenant/default/function/0-system-faasscheduler/version/$latest/defaultaz/task-66ccf050-50f6-4835-aa24-c1b15dddb00e/12996c08-0000-4000-8000-db6c3db0fcb2"),
							},
						}
						return response, nil
					}),
				ApplyFunc(state.GetState, func() state.ControllerState {
					controllerState := state.ControllerState{FaasInstance: make(map[string]map[string]struct{})}
					controllerState.FaasInstance[types.FaasSchedulerInstanceState] = map[string]struct{}{"test-2": {}}
					return controllerState
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

			schedulerConfig := fcConfig.GetFaaSSchedulerConfig()
			bytes, _ := json.Marshal(schedulerConfig)
			instanceMgr.instanceCache["test-1"] = &types.InstanceSpecification{
				FuncCtx:    nil,
				CancelFunc: nil,
				InstanceID: "test-1",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}},
				},
			}
			err = instanceMgr.CreateExpectedInstanceCount(context.TODO())
			So(err, ShouldBeNil)
		})
	})
}

func TestInstanceManager_RecoverInstance(t *testing.T) {
	Convey("RecoverInstance", t, func() {
		Convey("create failed", func() {
			instanceMgr, err := newFaaSSchedulerManager(1)
			So(err, ShouldBeNil)
			patches := []*Patches{
				ApplyFunc((*SchedulerManager).KillInstance, func(_ *SchedulerManager, _ string) error {
					return nil
				}),
				ApplyFunc((*SchedulerManager).CreateMultiInstances,
					func(_ *SchedulerManager, ctx context.Context, _ int) error {
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

func TestInstanceManager_HandleInstanceUpdate(t *testing.T) {
	defer ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	instanceID := "1"
	notFoundInstanceID := "2"
	instanceMgr, err := newFaaSSchedulerManager(2)
	assert.Nil(t, err)
	instanceMgr.addInstance(instanceID)
	schedulerConfig := fcConfig.GetFaaSSchedulerConfig()
	bytes, _ := json.Marshal(schedulerConfig)
	Convey("Test HandleInstanceEvent", t, func() {
		Convey("Test HandleInstanceEvent when instance exists", func() {
			specification := &types.InstanceSpecification{
				FuncCtx:    nil,
				CancelFunc: nil,
				InstanceID: instanceID,
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}},
				},
			}
			instanceMgr.HandleInstanceUpdate(specification)
			So(reflect.DeepEqual(instanceMgr.instanceCache[instanceID], specification), ShouldBeTrue)
			specification.InstanceSpecificationMeta.Function = "testFunction"
			instanceMgr.HandleInstanceUpdate(specification)
			So(reflect.DeepEqual(instanceMgr.instanceCache[instanceID], specification), ShouldBeTrue)
		})

		Convey("Test HandleInstanceEvent when instance not exists", func() {
			specification := &types.InstanceSpecification{
				FuncCtx:    nil,
				CancelFunc: nil,
				InstanceID: notFoundInstanceID,
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}},
				},
			}
			instanceMgr.HandleInstanceUpdate(specification)
			So(reflect.DeepEqual(instanceMgr.instanceCache[notFoundInstanceID], specification), ShouldBeTrue)
		})

		Convey("Test HandleInstanceDelete when create extra instance", func() {
			instanceMgr, err = newFaaSSchedulerManager(1)
			assert.Nil(t, err)
			defer ApplyMethod(reflect.TypeOf(instanceMgr.etcdClient.Client.KV), "Get",
				func(k *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse,
					error) {
					bytes, _ := json.Marshal(fcConfig.GetFaaSSchedulerConfig())
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
			specification := &types.InstanceSpecification{
				FuncCtx:    nil,
				CancelFunc: nil,
				InstanceID: instanceID,
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}},
				},
			}
			instanceMgr.HandleInstanceUpdate(specification)
			So(instanceMgr.instanceCache[instanceID], ShouldBeNil)
		})

		Convey("Test HandleInstanceUpdate when config change", func() {
			instanceMgr, err = newFaaSSchedulerManager(1)
			assert.Nil(t, err)
			cfg := &types.SchedulerConfig{}
			utils.DeepCopyObj(schedulerConfig, cfg)
			cfg.CPU = 1000
			marshal, _ := json.Marshal(cfg)
			specification := &types.InstanceSpecification{
				FuncCtx:    nil,
				CancelFunc: nil,
				InstanceID: instanceID,
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(marshal)}},
				},
			}
			instanceMgr.HandleInstanceUpdate(specification)
			So(instanceMgr.instanceCache[instanceID], ShouldBeNil)
		})
	})
}

func TestInstanceManager_HandleInstanceDelete(t *testing.T) {
	defer ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	Convey("Test HandleInstanceDelete", t, func() {
		instanceID := "1"
		notFoundInstanceID := "2"
		instanceMgr, err := newFaaSSchedulerManager(2)
		assert.Nil(t, err)
		schedulerConfig := fcConfig.GetFaaSSchedulerConfig()
		bytes, _ := json.Marshal(schedulerConfig)
		Convey("Test HandleInstanceDelete when instance exist", func() {
			instanceMgr.addInstance(instanceID)
			specification := &types.InstanceSpecification{
				FuncCtx:    nil,
				CancelFunc: nil,
				InstanceID: instanceID,
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}},
				},
			}
			instanceMgr.HandleInstanceDelete(specification)
			So(len(instanceMgr.instanceCache), ShouldEqual, 0)
		})

		Convey("Test HandleInstanceDelete when instance not exist", func() {
			specification := &types.InstanceSpecification{
				FuncCtx:                   nil,
				CancelFunc:                nil,
				InstanceID:                notFoundInstanceID,
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{},
			}
			instanceMgr.HandleInstanceDelete(specification)
			So(len(instanceMgr.instanceCache), ShouldEqual, 0)
		})

		Convey("Test HandleInstanceDelete when config change", func() {
			instanceMgr, err = newFaaSSchedulerManager(1)
			assert.Nil(t, err)
			cfg := &types.SchedulerConfig{}
			utils.DeepCopyObj(schedulerConfig, cfg)
			cfg.CPU = 1000
			marshal, _ := json.Marshal(cfg)
			specification := &types.InstanceSpecification{
				FuncCtx:    nil,
				CancelFunc: nil,
				InstanceID: instanceID,
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(marshal)}},
				},
			}
			instanceMgr.instanceCache[instanceID] = &types.InstanceSpecification{}
			ctx, cancelFunc := context.WithCancel(context.TODO())
			instanceMgr.recreateInstanceIDMap.Store(instanceID, cancelFunc)
			instanceMgr.HandleInstanceDelete(specification)
			<-ctx.Done()
		})
	})
}

func TestKillExceptInstance(t *testing.T) {
	Convey("KillExceptInstance", t, func() {
		instanceMgr, err := newFaaSSchedulerManager(1)
		So(err, ShouldBeNil)
		instanceMgr.instanceCache["testID"] = &types.InstanceSpecification{}
		err = instanceMgr.KillExceptInstance(1)
		So(err, ShouldBeNil)
	})
}

func TestRollingUpdate(t *testing.T) {
	Convey("test RollingUpdate", t, func() {
		manager, err := newFaaSSchedulerManager(2)
		So(err, ShouldBeNil)
		Convey("same config", func() {
			schedulerConfig := fcConfig.GetFaaSSchedulerConfig()
			bytes, _ := json.Marshal(schedulerConfig)
			manager.instanceCache["test-1"] = &types.InstanceSpecification{
				InstanceID: "test-1",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			manager.instanceCache["test-2"] = &types.InstanceSpecification{
				InstanceID: "test-2",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args: []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}}}}
			cfgEvent := &types.ConfigChangeEvent{
				SchedulerCfg: fcConfig.GetFaaSSchedulerConfig(),
				TraceID:      "traceID-123456789",
			}
			cfgEvent.SchedulerCfg.SchedulerNum = 2
			cfgEvent.Add(1)
			manager.ConfigChangeCh <- cfgEvent
			cfgEvent.Wait()
			So(len(manager.instanceCache), ShouldEqual, cfgEvent.SchedulerCfg.SchedulerNum)
			So(manager.instanceCache["test-1"], ShouldNotBeNil)
			So(manager.instanceCache["test-2"], ShouldNotBeNil)
		})

		Convey("different config", func() {
			schedulerConfig := fcConfig.GetFaaSSchedulerConfig()
			bytes, _ := json.Marshal(schedulerConfig)
			manager.instanceCache["test-1"] = &types.InstanceSpecification{
				InstanceID: "test-1",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args:           []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}},
					InstanceStatus: types.InstanceStatus{Code: int(commonconstant.KernelInstanceStatusRunning)}}}
			manager.instanceCache["test-2"] = &types.InstanceSpecification{
				InstanceID: "test-2",
				InstanceSpecificationMeta: types.InstanceSpecificationMeta{Function: "test-function",
					Args:           []map[string]string{{"value": base64.StdEncoding.EncodeToString(bytes)}},
					InstanceStatus: types.InstanceStatus{Code: int(commonconstant.KernelInstanceStatusRunning)},
				}}
			cfg := &types.SchedulerConfig{}
			utils.DeepCopyObj(fcConfig.GetFaaSSchedulerConfig(), cfg)
			cfgEvent := &types.ConfigChangeEvent{
				SchedulerCfg: cfg,
				TraceID:      "traceID-123456789",
			}
			cfgEvent.SchedulerCfg.CPU = 888
			cfgEvent.SchedulerCfg.SchedulerNum = 2
			cfgEvent.Add(1)
			manager.ConfigChangeCh <- cfgEvent
			cfgEvent.Wait()
			time.Sleep(2 * time.Second)
			close(manager.ConfigChangeCh)
			So(len(manager.terminalCache), ShouldEqual, 0)
			So(len(manager.instanceCache), ShouldEqual, cfgEvent.SchedulerCfg.SchedulerNum)
		})

		Convey("killAllTerminalInstance", func() {
			schedulerConfig := fcConfig.GetFaaSSchedulerConfig()
			bytes, _ := json.Marshal(schedulerConfig)
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
				SchedulerCfg: fcConfig.GetFaaSSchedulerConfig(),
				TraceID:      "traceID-123456789",
			}
			cfgEvent.SchedulerCfg.SchedulerNum = 2
			cfgEvent.Add(1)
			manager.RollingUpdate(context.TODO(), cfgEvent)
			cfgEvent.Wait()
			So(len(manager.instanceCache), ShouldEqual, 2)

			cfgEvent.SchedulerCfg.SchedulerNum = 2
			cfgEvent.Add(1)
			cancel, cancelFunc := context.WithCancel(context.TODO())
			cancelFunc()
			manager.RollingUpdate(cancel, cfgEvent)
			cfgEvent.Wait()
			So(len(manager.instanceCache), ShouldEqual, 0)

			cfgEvent.SchedulerCfg.SchedulerNum = 0
			cfgEvent.Add(1)
			manager.RollingUpdate(context.TODO(), cfgEvent)
			cfgEvent.Wait()
			So(len(manager.instanceCache), ShouldEqual, 0)
		})
	})
}

func TestConfigDiff(t *testing.T) {
	Convey("ConfigDiff", t, func() {
		manager, err := newFaaSSchedulerManager(2)
		So(err, ShouldBeNil)
		Convey("same config ,different num", func() {
			cfg := &types.SchedulerConfig{}
			utils.DeepCopyObj(fcConfig.GetFaaSSchedulerConfig(), cfg)
			cfgEvent := &types.ConfigChangeEvent{
				SchedulerCfg: cfg,
				TraceID:      "traceID-123456789",
			}
			cfgEvent.SchedulerCfg.SchedulerNum = 100
			_, num := manager.ConfigDiff(cfgEvent)
			So(num, ShouldEqual, 100)
		})
	})
}

func TestSchedulerManager_KillInstance(t *testing.T) {
	Convey("kill instance test", t, func() {
		Convey("baseline", func() {
			manager, err := newFaaSSchedulerManager(2)
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

func Test_createExtraParams_for_InstanceLifeCycle(t *testing.T) {
	Convey("createExtraParams for InstanceLifeCycle", t, func() {
		conf := &types.SchedulerConfig{
			Configuration: stypes.Configuration{
				CPU:    1332,
				Memory: 324134,
			},
			SchedulerNum: 0,
		}
		newFaaSSchedulerManager(2)
		params, err := createExtraParams(conf, "")
		So(err, ShouldBeNil)
		So(params.CreateOpt[commonconstant.InstanceLifeCycle], ShouldEqual, commonconstant.InstanceLifeCycleDetached)
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
		conf := &types.SchedulerConfig{
			Configuration: stypes.Configuration{
				CPU:    1332,
				Memory: 324134,
			},
			SchedulerNum: 0,
		}
		for _, ttt := range tt {
			newFaaSSchedulerManager(2)
			conf.NodeAffinityPolicy = ttt.nodeAffinityPolicy
			conf.NodeAffinity = ttt.nodeAffinity
			params, err := createExtraParams(conf, "")
			So(err, ShouldBeNil)
			So(params.CreateOpt[commonconstant.InstanceLifeCycle], ShouldEqual,
				commonconstant.InstanceLifeCycleDetached)
			So(params.CreateOpt[commonconstant.DelegateNodeAffinityPolicy], ShouldEqual, ttt.nodeAffinityPolicy)
			So(params.CreateOpt[commonconstant.DelegateNodeAffinity], ShouldEqual, ttt.nodeAffinity)
		}
	})
}
