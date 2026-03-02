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

// Package faasscheduler -
package functionscaler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/instancepool"
	"yuanrong.org/kernel/pkg/functionscaler/lease"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/rollout"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

type fakeLease struct{}

func (l *fakeLease) Extend() error {
	return nil
}
func (l *fakeLease) Release() error {
	return nil
}
func (l *fakeLease) GetInterval() time.Duration {
	return time.Second
}

var testFuncSpec = &types.FunctionSpecification{
	FuncCtx: context.TODO(),
	FuncKey: "TestFuncKey",
	FuncMetaData: commonTypes.FuncMetaData{
		Handler: "myHandler",
	},
	ResourceMetaData: commonTypes.ResourceMetaData{
		CPU:    500,
		Memory: 500,
	},
	InstanceMetaData: commonTypes.InstanceMetaData{
		MinInstance:   0,
		MaxInstance:   1000,
		ConcurrentNum: 100,
	},
	ExtendedMetaData: commonTypes.ExtendedMetaData{
		Initializer: commonTypes.Initializer{
			Handler: "myInitializer",
		},
	},
}

func TestNewFaaSScheduler(t *testing.T) {
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "SubscribeFuncSpec", func(_ *registry.Registry,
			subChan chan registry.SubEvent) {
		}),
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "SubscribeInsSpec", func(_ *registry.Registry,
			subChan chan registry.SubEvent) {
		}),
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "SubscribeInsConfig", func(_ *registry.Registry,
			subChan chan registry.SubEvent) {
		}),
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "SubscribeAliasSpec", func(_ *registry.Registry,
			subChan chan registry.SubEvent) {
		}),
		ApplyFunc((*etcd3.EtcdWatcher).StartList, func(_ *etcd3.EtcdWatcher) {
		}),
		ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, key string) string { return key }),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	time.Sleep(10 * time.Millisecond)
	convey.Convey("test New faasscheduler", t, func() {
		stopCh := make(chan struct{})
		faasScheduler := NewFaaSScheduler(stopCh)
		assert.Equal(t, true, faasScheduler != nil)
		close(stopCh)
	})
	convey.Convey("testfaasScheduler funcSpecCh", t, func() {
		stopCh := make(chan struct{})
		faasScheduler := NewFaaSScheduler(stopCh)
		assert.Equal(t, true, faasScheduler != nil)
		faasScheduler.funcSpecCh <- registry.SubEvent{EventMsg: struct{}{}}
		time.Sleep(5 * time.Microsecond)
		close(stopCh)
	})
	convey.Convey("testfaasScheduler insSpecCh", t, func() {
		stopCh := make(chan struct{})
		faasScheduler := NewFaaSScheduler(stopCh)
		assert.Equal(t, true, faasScheduler != nil)
		faasScheduler.insSpecCh <- registry.SubEvent{EventMsg: struct{}{}}
		time.Sleep(5 * time.Microsecond)
		close(stopCh)
	})
	convey.Convey("testfaasScheduler insConfigCh", t, func() {
		stopCh := make(chan struct{})
		faasScheduler := NewFaaSScheduler(stopCh)
		assert.Equal(t, true, faasScheduler != nil)
		faasScheduler.insConfigCh <- registry.SubEvent{EventMsg: struct{}{}}
		time.Sleep(5 * time.Microsecond)
		close(stopCh)
	})
}

func initRegistry() {
	patches := []*Patches{
		ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, key string) string { return key }),
		ApplyFunc((*registry.FaasSchedulerRegistry).WaitForETCDList, func() {}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	_ = registry.InitRegistry(make(chan struct{}))
	registry.GlobalRegistry.FaaSSchedulerRegistry = registry.NewFaasSchedulerRegistry(make(chan struct{}))
	selfregister.SelfInstanceID = "schedulerID-1"
	selfregister.GlobalSchedulerProxy.Add(&commonTypes.InstanceInfo{
		TenantID:     "123456789",
		FunctionName: "faasscheduler",
		Version:      "lastest",
		InstanceName: "schedulerID-1",
	}, "")
}

func TestMain(m *testing.M) {
	patches := []*Patches{
		ApplyFunc((*etcd3.EtcdWatcher).StartList, func(_ *etcd3.EtcdWatcher) {}),
		ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		ApplyFunc(etcd3.GetCAEMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		ApplyFunc((*etcd3.EtcdClient).AttachAZPrefix, func(_ *etcd3.EtcdClient, key string) string { return key }),
	}
	defer func() {
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()
	config.GlobalConfig = types.Configuration{
		AutoScaleConfig: types.AutoScaleConfig{
			SLAQuota:      500,
			ScaleDownTime: 1000,
			BurstScaleNum: 1000,
		},
		LeaseSpan: 500,
		LocalAuth: localauth.AuthConfig{
			AKey: "ENC(key=servicekek, value=6B6D73763030000101D615B6381ED56AF68123844D047428BDCCBF19957866" +
				"CD0D7F53C29438337667A93FB9A06C5ED4A3D925C87655E4C734)",
			SKey: "ENC(key=servicekek, value=6B6D73763030000101139308ABBC0C4120F949AC833416D5E6D8CA18D8C69E" +
				"4C5E03E553E18733B4119C4B716FF2C8265336BB2979545A24FDC07CDD6A6A02F412D0DE83BD43F2A07DDBC78EB2)",
			Duration: 0,
		},
	}
	initRegistry()
	instancepool.SetGlobalSdkClient(&mockUtils.FakeLibruntimeSdkClient{})
	m.Run()
}

func TestProcessSubscription(t *testing.T) {
	stopCh := make(chan struct{})
	defer close(stopCh)
	registry.GlobalRegistry = &registry.Registry{
		FaaSSchedulerRegistry:  registry.NewFaasSchedulerRegistry(stopCh),
		FunctionRegistry:       registry.NewFunctionRegistry(stopCh),
		InstanceRegistry:       registry.NewInstanceRegistry(stopCh),
		FaaSManagerRegistry:    registry.NewFaaSManagerRegistry(stopCh),
		InstanceConfigRegistry: registry.NewInstanceConfigRegistry(stopCh),
		AliasRegistry:          registry.NewAliasRegistry(stopCh),
		RolloutRegistry:        registry.NewRolloutRegistry(stopCh),
	}
	faasScheduler := NewFaaSScheduler(stopCh)
	convey.Convey("test processFunctionSubscription", t, func() {
		faasScheduler.funcSpecCh <- registry.SubEvent{
			EventType: registry.SubEventTypeUpdate,
			EventMsg: &types.FunctionSpecification{
				FuncKey: "testFunc",
			},
		}
	})
	convey.Convey("test processInstanceSubscription", t, func() {
		faasScheduler.insSpecCh <- registry.SubEvent{
			EventType: registry.SubEventTypeUpdate,
			EventMsg: &commonTypes.InstanceSpecification{
				InstanceID: "testIns",
			},
		}
	})
	convey.Convey("test processInstanceConfigSubscription", t, func() {
		faasScheduler.insConfigCh <- registry.SubEvent{
			EventType: registry.SubEventTypeUpdate,
			EventMsg: &instanceconfig.Configuration{
				FuncKey: "testFunc",
			},
		}
	})
}

func TestFaaSScheduler_processRolloutConfigSubscription(t *testing.T) {
	convey.Convey("test processRolloutConfigSubscription", t, func() {
		convey.Convey("baseline", func() {
			count := 0
			p := ApplyFunc((*instancepool.PoolManager).HandleRolloutRatioChange,
				func(_ *instancepool.PoolManager, ratio int) {
					count++
				})
			defer p.Reset()
			rolloutConfigCh := make(chan registry.SubEvent)
			faasScheduler := &FaaSScheduler{
				rolloutConfigCh: rolloutConfigCh,
				PoolManager:     &instancepool.PoolManager{},
			}
			go faasScheduler.processRolloutConfigSubscription()
			rolloutConfigCh <- registry.SubEvent{
				EventType: "aaa",
				EventMsg:  50,
			}
			time.Sleep(100 * time.Millisecond)
			convey.So(count, convey.ShouldEqual, 1)
			rolloutConfigCh <- registry.SubEvent{
				EventType: "aaa",
				EventMsg:  "123",
			}
			time.Sleep(100 * time.Millisecond)
			convey.So(count, convey.ShouldEqual, 1)
			close(rolloutConfigCh)
			time.Sleep(100 * time.Millisecond)
			convey.So(count, convey.ShouldEqual, 1)
		})
	})
}

func Test_processAliasSpecSubscription(t *testing.T) {
	convey.Convey("processAliasSpecSubscription", t, func() {
		stopCh := make(chan struct{})
		defer close(stopCh)
		registry.GlobalRegistry = &registry.Registry{
			FaaSSchedulerRegistry:  registry.NewFaasSchedulerRegistry(stopCh),
			FunctionRegistry:       registry.NewFunctionRegistry(stopCh),
			InstanceRegistry:       registry.NewInstanceRegistry(stopCh),
			FaaSManagerRegistry:    registry.NewFaaSManagerRegistry(stopCh),
			InstanceConfigRegistry: registry.NewInstanceConfigRegistry(stopCh),
			AliasRegistry:          registry.NewAliasRegistry(stopCh),
			RolloutRegistry:        registry.NewRolloutRegistry(stopCh),
		}
		faasScheduler := NewFaaSScheduler(stopCh)
		var checkAliasUrn string
		defer ApplyMethod(reflect.TypeOf(faasScheduler.PoolManager), "HandleAliasEvent",
			func(pm *instancepool.PoolManager, eventType registry.EventType, aliasUrn string) {
				checkAliasUrn = aliasUrn
			}).Reset()
		faasScheduler.aliasSpecCh <- registry.SubEvent{
			EventType: registry.SubEventTypeUpdate,
			EventMsg:  "aliasUrn",
		}
		time.Sleep(100 * time.Millisecond)
		convey.So(checkAliasUrn, convey.ShouldEqual, "aliasUrn")
		checkAliasUrn = ""
		faasScheduler.aliasSpecCh <- registry.SubEvent{
			EventType: registry.SubEventTypeUpdate,
			EventMsg:  123,
		}
		time.Sleep(100 * time.Millisecond)
		convey.So(checkAliasUrn, convey.ShouldEqual, "")
		close(faasScheduler.aliasSpecCh)
	})
}

func TestParseStateOperation(t *testing.T) {
	convey.Convey("success", t, func() {
		ops := "funcKey;stateId"
		targetName, stateID := parseStateOperation(ops)
		convey.So(stateID, convey.ShouldEqual, "stateId")
		convey.So(targetName, convey.ShouldEqual, "funcKey")
	})
}

func TestProcessInstanceRequest(t *testing.T) {
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "GetFuncSpec", func(_ *registry.Registry,
			funcKey string) *types.FunctionSpecification {
			return testFuncSpec
		}),
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "SubscribeFuncSpec", func(_ *registry.Registry,
			subChan chan registry.SubEvent) {
		}),
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "SubscribeInsSpec", func(_ *registry.Registry,
			subChan chan registry.SubEvent) {
		}),
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "SubscribeInsConfig", func(_ *registry.Registry,
			subChan chan registry.SubEvent) {
		}),
		ApplyMethod(reflect.TypeOf(&registry.Registry{}), "SubscribeAliasSpec", func(_ *registry.Registry,
			subChan chan registry.SubEvent) {
		}),
		ApplyFunc((*registry.FunctionAvailableRegistry).GeClusters, func(_ *registry.FunctionAvailableRegistry, _ string) []string {
			return []string{}
		}),
		ApplyFunc((*registry.FaaSFrontendRegistry).GetFrontends, func(_ *registry.FaaSFrontendRegistry, _ string) []string {
			return []string{}
		}),
		ApplyFunc((*instancepool.PoolManager).ReleaseAbnormalInstance, func(_ *instancepool.PoolManager,
			instance *types.Instance, logger api.FormatLogger) {
		}),
		ApplyFunc(state.Update, func(value interface{}, tags ...string) {
		}),
		ApplyFunc((*etcd3.EtcdWatcher).StartList, func(_ *etcd3.EtcdWatcher) {
		}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	stopCh := make(chan struct{})
	defer close(stopCh)
	faasScheduler := NewFaaSScheduler(stopCh)
	time.Sleep(1 * time.Second)
	faasScheduler.PoolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, testFuncSpec)
	faasScheduler.PoolManager.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, &instanceconfig.Configuration{
		FuncKey: "TestFuncKey",
		InstanceMetaData: commonTypes.InstanceMetaData{
			MaxInstance:   100,
			ConcurrentNum: 1,
		},
	})
	metrics := &types.InstanceThreadMetrics{}
	metricsData, _ := json.Marshal(metrics)
	releaseExtraData := &types.InstanceThreadMetrics{
		ProcReqNum:  11,
		AvgProcTime: 11,
		MaxProcTime: 11,
		IsAbnormal:  true,
	}
	releaseExtraRawData, _ := json.Marshal(releaseExtraData)
	acquireRsp := &commonTypes.InstanceResponse{}

	convey.Convey("acquire", t, func() {
		m := map[string][]byte{"resourcesData": []byte("")}
		bytes, _ := json.Marshal(m)
		acquireArgs := []api.Arg{
			{
				Type: api.Value,
				Data: []byte("acquire#TestFuncKey"),
			},
			{
				Type: api.Value,
				Data: bytes,
			},
			{
				Type: api.Value,
				Data: []byte(""),
			},
			{
				Type: api.Value,
				Data: []byte(""),
			},
		}
		convey.Convey("acquire success", func() {
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(acquireArgs, "")
			_ = json.Unmarshal(resData, acquireRsp)
			assert.Equal(t, constant.InsReqSuccessCode, acquireRsp.ErrorCode)
		})
		resourceRes := &resspeckey.ResourceSpecification{
			CPU:    300,
			Memory: 128,
		}
		resource, _ := json.Marshal(resourceRes)
		m[constant.InstanceRequirementResourcesKey] = resource
		convey.Convey("acquire set resource success", func() {
			acquireArgs[1].Data, _ = json.Marshal(m)
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(acquireArgs, "")
			_ = json.Unmarshal(resData, acquireRsp)
			assert.Equal(t, constant.InsReqSuccessCode, acquireRsp.ErrorCode)
		})
		convey.Convey("acquire set resource error", func() {
			m[constant.InstanceRequirementResourcesKey] = resource[1:1]
			acquireArgs[1].Data, _ = json.Marshal(m)
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(acquireArgs, "")
			_ = json.Unmarshal(resData, acquireRsp)
			assert.Equal(t, constant.InsReqSuccessCode, acquireRsp.ErrorCode)
		})
		convey.Convey("acquire metrics error", func() {
			defer ApplyMethod(reflect.TypeOf(registry.GlobalRegistry), "GetFuncSpec",
				func(_ *registry.Registry, funcKey string) *types.FunctionSpecification {
					return nil
				}).Reset()
			defer ApplyMethod(reflect.TypeOf(registry.GlobalRegistry), "FetchSilentFuncSpec",
				func(_ *registry.Registry, funcKey string) *types.FunctionSpecification {
					return nil
				}).Reset()
			releaseRsp := &commonTypes.InstanceResponse{}
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(acquireArgs, "")
			_ = json.Unmarshal(resData, releaseRsp)
			assert.Equal(t, statuscode.FuncMetaNotFoundErrCode, releaseRsp.ErrorCode)
		})
	})

	convey.Convey("retain", t, func() {
		retainArgs := []api.Arg{
			{
				Type: api.Value,
				Data: []byte(fmt.Sprintf("retain#%s", acquireRsp.ThreadID)),
			},
			{
				Type: api.Value,
				Data: []byte(""),
			},
			{
				Type: api.Value,
				Data: []byte(""),
			},
		}
		retainRsp := &commonTypes.InstanceResponse{}
		convey.Convey("retain success", func() {
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(retainArgs, "")
			_ = json.Unmarshal(resData, retainRsp)
			assert.Equal(t, constant.InsReqSuccessCode, retainRsp.ErrorCode)
		})
		convey.Convey("retain metrics error", func() {
			retainArgs[0].Data = []byte("retain#000thread111")
			retainArgs[1].Data = metricsData
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(retainArgs, "")
			_ = json.Unmarshal(resData, retainRsp)
			assert.Equal(t, statuscode.LeaseIDNotFoundCode, retainRsp.ErrorCode)
		})
		convey.Convey("retain stateThread error instance subHealth", func() {
			retainErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("retain#%s", "TestFuncKey1")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			p := &instancepool.PoolManager{}
			patch := ApplyMethod(reflect.TypeOf(p),
				"ReleaseStateThread", func(db *instancepool.PoolManager,
					insAlloc *types.InstanceAllocation) error {
					return errors.New("release state thread error")
				})
			defer patch.Reset()
			insAlloc := &types.InstanceAllocation{
				AllocationID: "testFunc-stateThread1",
				Instance: &types.Instance{
					FuncKey: "TestFuncKey1",
					InstanceStatus: commonTypes.InstanceStatus{
						Code: int32(constant.KernelInstanceStatusSubHealth),
					},
					ResKey: resspeckey.ResSpecKey{},
				},
			}
			faasScheduler.allocRecord.Store("TestFuncKey1", insAlloc)
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(retainErrArgs, "")
			_ = json.Unmarshal(resData, retainRsp)
			assert.Equal(t, statuscode.InstanceStatusAbnormalCode, retainRsp.ErrorCode)
		})
		convey.Convey("retain stateThread error", func() {
			retainErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("retain#%s", "TestFuncKey1")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			p := &instancepool.PoolManager{}
			patch1 := ApplyMethod(reflect.TypeOf(p), "ReleaseStateThread", func(db *instancepool.PoolManager,
				thread *types.InstanceAllocation) error {
				return nil
			})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(p), "RetainStateThread", func(db *instancepool.PoolManager,
				thread *types.InstanceAllocation) error {
				return errors.New("retain state thread error")
			})
			defer patch2.Reset()
			insAlloc := &types.InstanceAllocation{
				AllocationID: "testFunc-stateThread1",
				Instance: &types.Instance{
					FuncKey: "TestFuncKey1",
					InstanceStatus: commonTypes.InstanceStatus{
						Code: int32(constant.KernelInstanceStatusRunning),
					},
					ResKey: resspeckey.ResSpecKey{},
				},
				Lease: &lease.GenericInstanceLease{},
			}
			faasScheduler.allocRecord.Store("TestFuncKey1", insAlloc)
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(retainErrArgs, "")
			_ = json.Unmarshal(resData, retainRsp)
			assert.Equal(t, constant.LeaseExpireOrDeletedErrorCode, retainRsp.ErrorCode)
		})
		convey.Convey("retain stateThread success", func() {
			retainErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("retain#%s", "TestFuncKey1")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			p := &instancepool.PoolManager{}
			patch1 := ApplyMethod(reflect.TypeOf(p), "ReleaseStateThread", func(db *instancepool.PoolManager,
				thread *types.InstanceAllocation) error {
				return nil
			})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(p), "RetainStateThread", func(db *instancepool.PoolManager,
				thread *types.InstanceAllocation) error {
				return nil
			})
			defer patch2.Reset()
			insAlloc := &types.InstanceAllocation{
				AllocationID: "testFunc-stateThread1",
				Instance: &types.Instance{
					FuncKey: "TestFuncKey1",
					InstanceStatus: commonTypes.InstanceStatus{
						Code: int32(constant.KernelInstanceStatusRunning),
					},
					ResKey: resspeckey.ResSpecKey{},
				},
				Lease: &lease.GenericInstanceLease{},
			}
			faasScheduler.allocRecord.Store("TestFuncKey1", insAlloc)
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(retainErrArgs, "")
			_ = json.Unmarshal(resData, retainRsp)
			assert.Equal(t, constant.InsReqSuccessCode, retainRsp.ErrorCode)
		})
		convey.Convey("retain InsAlloc error release error", func() {
			retainErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("retain#%s", "TestFuncKey1")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			p := &instancepool.PoolManager{}
			patch1 := ApplyMethod(reflect.TypeOf(p), "ReleaseStateThread", func(db *instancepool.PoolManager,
				thread *types.InstanceAllocation) error {
				return nil
			})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(p), "RetainStateThread", func(db *instancepool.PoolManager,
				thread *types.InstanceAllocation) error {
				return nil
			})
			defer patch2.Reset()
			l := &fakeLease{}
			patch3 := ApplyMethod(reflect.TypeOf(l), "Release", func(l *fakeLease) error {
				return errors.New("release error")
			})
			defer patch3.Reset()
			insAlloc := &types.InstanceAllocation{
				AllocationID: "testFunc-Thread1",
				Instance: &types.Instance{
					FuncKey: "TestFuncKey1",
					InstanceStatus: commonTypes.InstanceStatus{
						Code: int32(constant.KernelInstanceStatusSubHealth),
					},
					ResKey: resspeckey.ResSpecKey{},
				},
				Lease: &lease.GenericInstanceLease{},
			}
			faasScheduler.allocRecord.Store("TestFuncKey1", insAlloc)
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(retainErrArgs, "")
			_ = json.Unmarshal(resData, retainRsp)
			assert.Equal(t, statuscode.InstanceStatusAbnormalCode, retainRsp.ErrorCode)
		})
		convey.Convey("retain InsAlloc error extend error", func() {
			retainErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("retain#%s", "TestFuncKey1")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			p := &instancepool.PoolManager{}
			patch1 := ApplyMethod(reflect.TypeOf(p),
				"ReleaseStateThread", func(db *instancepool.PoolManager,
					thread *types.InstanceAllocation) error {
					return nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(p),
				"RetainStateThread", func(db *instancepool.PoolManager,
					thread *types.InstanceAllocation) error {
					return nil
				})
			defer patch2.Reset()
			l := &fakeLease{}
			patch3 := ApplyMethod(reflect.TypeOf(l),
				"Extend", func(l *fakeLease) error {
					return errors.New("extend error")
				})
			defer patch3.Reset()
			insAlloc := &types.InstanceAllocation{
				AllocationID: "testFunc-Thread1",
				Instance: &types.Instance{
					FuncKey: "TestFuncKey1",
					InstanceStatus: commonTypes.InstanceStatus{
						Code: int32(constant.KernelInstanceStatusRunning),
					},
					ResKey: resspeckey.ResSpecKey{},
				},
				Lease: &lease.GenericInstanceLease{},
			}
			faasScheduler.allocRecord.Store("TestFuncKey1", insAlloc)
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(retainErrArgs, "")
			_ = json.Unmarshal(resData, retainRsp)
			assert.Equal(t, constant.LeaseExpireOrDeletedErrorCode, retainRsp.ErrorCode)
		})
		convey.Convey("retain insThdAlloc data error", func() {
			retainErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("retain#%s", "TestFuncKey1")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			faasScheduler.allocRecord.Store("TestFuncKey1", "")
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(retainErrArgs, "")
			_ = json.Unmarshal(resData, retainRsp)
			assert.Equal(t, statuscode.StatusInternalServerError, retainRsp.ErrorCode)
		})
	})

	convey.Convey("release", t, func() {
		releaseArgs := []api.Arg{
			{
				Type: api.Value,
				Data: []byte(fmt.Sprintf("release#%s", acquireRsp.ThreadID)),
			},
			{
				Type: api.Value,
				Data: []byte(""),
			},
			{
				Type: api.Value,
				Data: []byte(""),
			},
		}
		releaseRsp := &commonTypes.InstanceResponse{}
		convey.Convey("release success", func() {
			releaseArgs[1].Data = releaseExtraRawData
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(releaseArgs, "")
			_ = json.Unmarshal(resData, releaseRsp)
			assert.Equal(t, constant.InsReqSuccessCode, releaseRsp.ErrorCode)
		})
		convey.Convey("release metrics error", func() {
			releaseArgs[1].Data = releaseExtraRawData
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(releaseArgs, "")
			_ = json.Unmarshal(resData, releaseRsp)
			assert.Equal(t, statuscode.InstanceNotFoundErrCode, releaseRsp.ErrorCode)
		})
		convey.Convey("release state error func not exist", func() {
			releaseErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("release#%s", "TestFuncKeyE;stateID")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			patch := ApplyMethod(reflect.TypeOf(&registry.Registry{}), "GetFuncSpec", func(_ *registry.Registry,
				funcKey string) *types.FunctionSpecification {
				return nil
			})
			defer patch.Reset()
			releaseErrArgs[1].Data = releaseExtraRawData
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(releaseErrArgs, "")
			_ = json.Unmarshal(resData, releaseRsp)
			assert.Equal(t, statuscode.FuncMetaNotFoundErrCode, releaseRsp.ErrorCode)
		})
		convey.Convey("release state error delete error", func() {
			releaseErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("release#%s", "TestFuncKey;stateID")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			p := &instancepool.PoolManager{}
			patchGet := ApplyMethod(reflect.TypeOf(p), "GetAndDeleteState",
				func(db *instancepool.PoolManager, stateID string, funcKey string,
					funcSpec *types.FunctionSpecification, logger api.FormatLogger) bool {
					return false
				})
			defer patchGet.Reset()
			releaseErrArgs[1].Data = releaseExtraRawData
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(releaseErrArgs, "")
			_ = json.Unmarshal(resData, releaseRsp)
			assert.Equal(t, statuscode.StateNotExistedErrCode, releaseRsp.ErrorCode)
		})
		convey.Convey("release state success", func() {
			releaseErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("release#%s", "TestFuncKey;stateID")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			p := &instancepool.PoolManager{}
			patchGet := ApplyMethod(reflect.TypeOf(p), "GetAndDeleteState", func(db *instancepool.PoolManager,
				stateID string, funcKey string, funcSpec *types.FunctionSpecification, logger api.FormatLogger) bool {
				return true
			})
			defer patchGet.Reset()
			releaseErrArgs[1].Data = releaseExtraRawData
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(releaseErrArgs, "")
			_ = json.Unmarshal(resData, releaseRsp)
			assert.Equal(t, constant.InsReqSuccessCode, releaseRsp.ErrorCode)
		})
		convey.Convey("insThdAlloc data error", func() {
			releaseErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("release#%s", "TestFuncKeyE1")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			releaseErrArgs[1].Data = releaseExtraRawData
			faasScheduler.allocRecord.Store("TestFuncKeyE1", "")
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(releaseErrArgs, "")
			_ = json.Unmarshal(resData, releaseRsp)
			assert.Equal(t, statuscode.StatusInternalServerError, releaseRsp.ErrorCode)
		})
		convey.Convey("release state thread error", func() {
			releaseErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("release#%s", "TestFuncKey1")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			p := &instancepool.PoolManager{}
			patchGet := ApplyMethod(reflect.TypeOf(p), "ReleaseStateThread", func(db *instancepool.PoolManager,
				thread *types.InstanceAllocation) error {
				return errors.New("release state thread error")
			})
			defer patchGet.Reset()
			insAlloc := &types.InstanceAllocation{
				AllocationID: "testFunck-stateThread1",
				Instance: &types.Instance{
					FuncKey: "TestFuncKey1",
					ResKey:  resspeckey.ResSpecKey{},
				},
				Lease: &lease.GenericInstanceLease{},
			}
			faasScheduler.allocRecord.Store("TestFuncKey1", insAlloc)
			releaseErrArgs[1].Data = releaseExtraRawData
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(releaseErrArgs, "")
			_ = json.Unmarshal(resData, releaseRsp)
			assert.Equal(t, statuscode.StatusInternalServerError, releaseRsp.ErrorCode)
		})
		convey.Convey("release state thread success", func() {
			releaseErrArgs := []api.Arg{
				{
					Type: api.Value,
					Data: []byte(fmt.Sprintf("release#%s", "TestFuncKey1")),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			p := &instancepool.PoolManager{}
			patchGet := ApplyMethod(reflect.TypeOf(p),
				"ReleaseStateThread", func(db *instancepool.PoolManager,
					thread *types.InstanceAllocation) error {
					return nil
				})
			defer patchGet.Reset()
			insAlloc := &types.InstanceAllocation{
				AllocationID: "testFunc-stateThread1",
				Instance: &types.Instance{
					FuncKey: "TestFuncKey1",
					ResKey:  resspeckey.ResSpecKey{},
				},
			}
			faasScheduler.allocRecord.Store("TestFuncKey1", insAlloc)
			releaseErrArgs[1].Data = releaseExtraRawData
			resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(releaseErrArgs, "")
			_ = json.Unmarshal(resData, releaseRsp)
			assert.Equal(t, constant.InsReqSuccessCode, releaseRsp.ErrorCode)
		})
	})

	convey.Convey("error opt", t, func() {
		errorArgs := []api.Arg{
			{
				Type: api.Value,
				Data: []byte(fmt.Sprintf("xxxxx#")),
			},
			{
				Type: api.Value,
				Data: []byte(""),
			},
			{
				Type: api.Value,
				Data: []byte(""),
			},
		}
		errorRsp := &commonTypes.InstanceResponse{}
		resData, _ := faasScheduler.ProcessInstanceRequestLibruntime(errorArgs, "")
		_ = json.Unmarshal(resData, errorRsp)
		assert.Equal(t, constant.UnsupportedOperationErrorCode, errorRsp.ErrorCode)
	})
}

func Test_parseInstanceOperation(t *testing.T) {
	convey.Convey("test parseInstanceOperation", t, func() {
		convey.Convey("baseline", func() {
			op, name, data, _ := parseInstanceOperation([]api.Arg{
				{
					Type:            0,
					Data:            []byte("acquire#aaa"),
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
			}, "")
			convey.So(op, convey.ShouldNotBeNil)
			convey.So(name, convey.ShouldEqual, "aaa")
			convey.So(data, convey.ShouldBeNil)
		})
		convey.Convey("error args", func() {
			op, name, data, _ := parseInstanceOperation([]api.Arg{
				{
					Type:            0,
					Data:            []byte("acquire#aaa"),
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
			}, "")
			convey.So(op, convey.ShouldEqual, insOpAcquire)
			convey.So(name, convey.ShouldEqual, "aaa")
			convey.So(data, convey.ShouldBeNil)
		})
		convey.Convey("error types", func() {
			op, name, data, _ := parseInstanceOperation([]api.Arg{
				{
					Type:            1,
					Data:            []byte("acquire#aaa"),
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            1,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
			}, "")
			convey.So(op, convey.ShouldEqual, insOpUnknown)
			convey.So(name, convey.ShouldEqual, "")
			convey.So(data, convey.ShouldBeNil)
		})
		convey.Convey("error operationArg", func() {
			op, name, data, _ := parseInstanceOperation([]api.Arg{
				{
					Type:            0,
					Data:            []byte("acquire"),
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
			}, "")
			convey.So(op, convey.ShouldEqual, insOpUnknown)
			convey.So(name, convey.ShouldEqual, "")
			convey.So(data, convey.ShouldBeNil)
		})
	})
}

func TestFaaSScheduler_processSchedulerProxySubscription(t *testing.T) {
	convey.Convey("test processSchedulerProxySubscription", t, func() {
		convey.Convey("baseline", func() {
			count := 0
			p := ApplyFunc((*instancepool.PoolManager).HandleSchedulerManaged,
				func(_ *instancepool.PoolManager, eventType registry.EventType,
					insSpec *commonTypes.InstanceSpecification) {
					count++
				})
			defer p.Reset()
			schedulerCh := make(chan registry.SubEvent)
			faasScheduler := &FaaSScheduler{
				schedulerCh: schedulerCh,
				PoolManager: &instancepool.PoolManager{},
			}
			go faasScheduler.processSchedulerProxySubscription()
			schedulerCh <- registry.SubEvent{
				EventType: "aaa",
				EventMsg:  &commonTypes.InstanceSpecification{},
			}
			time.Sleep(100 * time.Millisecond)
			convey.So(count, convey.ShouldEqual, 1)
			schedulerCh <- registry.SubEvent{
				EventType: "aaa",
				EventMsg:  "123",
			}
			time.Sleep(100 * time.Millisecond)
			convey.So(count, convey.ShouldEqual, 1)
			close(schedulerCh)
			time.Sleep(100 * time.Millisecond)
			convey.So(count, convey.ShouldEqual, 1)
		})
	})
}

func TestFaaSScheduler_ProcessInstanceRequestLibruntime(t *testing.T) {
	faasScheduler := &FaaSScheduler{}
	convey.Convey("test ProcessInstanceRequestLibruntime", t, func() {
		convey.Convey("baseline", func() {
			defer ApplyFunc((*FaaSScheduler).handleInstanceAcquire, func(_ *FaaSScheduler,
				targetName string, extraData []byte,
				traceID string) *commonTypes.InstanceResponse {
				return &commonTypes.InstanceResponse{
					ErrorCode: 111,
				}
			}).Reset()
			resData, err := faasScheduler.ProcessInstanceRequestLibruntime([]api.Arg{
				{
					Type:            0,
					Data:            []byte("acquire#aaa"),
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
			}, "")
			response := &commonTypes.InstanceResponse{}
			_ = json.Unmarshal(resData, response)
			convey.So(err, convey.ShouldBeNil)
			convey.So(response.ErrorCode, convey.ShouldEqual, 111)
		})
	})
	convey.Convey("test create instance", t, func() {
		var createErr snerror.SNError
		defer ApplyFunc((*instancepool.PoolManager).CreateInstance, func(_ *instancepool.PoolManager,
			insCrtReq *types.InstanceCreateRequest) (*types.Instance, snerror.SNError) {
			if createErr != nil {
				return nil, createErr
			}
			return &types.Instance{}, nil
		}).Reset()
		defer ApplyFunc((*registry.Registry).GetFuncSpec, func(_ *registry.Registry, funcKey string) *types.FunctionSpecification {
			if funcKey == "testFunc" {
				return &types.FunctionSpecification{}
			}
			return nil
		}).Reset()
		resData, err := faasScheduler.ProcessInstanceRequestLibruntime([]api.Arg{
			{
				Type:            0,
				Data:            []byte("create#aaa"),
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            []byte(""),
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            nil,
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
		}, "")
		response := &commonTypes.InstanceResponse{}
		_ = json.Unmarshal(resData, response)
		convey.So(err, convey.ShouldBeNil)
		convey.So(response, convey.ShouldNotBeNil)
		convey.So(response.ErrorCode, convey.ShouldEqual, statuscode.FuncMetaNotFoundErrCode)
		resData, err = faasScheduler.ProcessInstanceRequestLibruntime([]api.Arg{
			{
				Type:            0,
				Data:            []byte("create#testFunc"),
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            []byte("wrong data"),
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            nil,
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
		}, "")
		_ = json.Unmarshal(resData, response)
		convey.So(err, convey.ShouldBeNil)
		convey.So(response, convey.ShouldNotBeNil)
		convey.So(response.ErrorCode, convey.ShouldEqual, statuscode.StatusInternalServerError)
		resData, err = faasScheduler.ProcessInstanceRequestLibruntime([]api.Arg{
			{
				Type:            0,
				Data:            []byte("create#testFunc"),
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            nil,
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            nil,
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
		}, "")
		_ = json.Unmarshal(resData, response)
		convey.So(err, convey.ShouldBeNil)
		convey.So(response, convey.ShouldNotBeNil)
		convey.So(response.ErrorCode, convey.ShouldEqual, constant.InsReqSuccessCode)
	})
	convey.Convey("test delete instance", t, func() {
		var deleteErr snerror.SNError
		defer ApplyFunc((*instancepool.PoolManager).DeleteInstance, func(_ *instancepool.PoolManager,
			instance *types.Instance) snerror.SNError {
			if deleteErr != nil {
				return deleteErr
			}
			return nil
		}).Reset()
		defer ApplyFunc((*registry.Registry).GetInstance, func(_ *registry.Registry, instanceID string) *types.Instance {
			if instanceID == "testIns" {
				return &types.Instance{}
			}
			return nil
		}).Reset()
		resData, err := faasScheduler.ProcessInstanceRequestLibruntime([]api.Arg{
			{
				Type:            0,
				Data:            []byte("delete#aaa"),
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            []byte(""),
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            nil,
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
		}, "")
		response := &commonTypes.InstanceResponse{}
		_ = json.Unmarshal(resData, response)
		convey.So(response, convey.ShouldNotBeNil)
		convey.So(response.ErrorCode, convey.ShouldEqual, statuscode.InstanceNotFoundErrCode)
		resData, err = faasScheduler.ProcessInstanceRequestLibruntime([]api.Arg{
			{
				Type:            0,
				Data:            []byte("delete#testIns"),
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            []byte(""),
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
			{
				Type:            0,
				Data:            nil,
				ObjectID:        "",
				NestedObjectIDs: nil,
			},
		}, "")
		_ = json.Unmarshal(resData, response)
		convey.So(err, convey.ShouldBeNil)
		convey.So(response, convey.ShouldNotBeNil)
		convey.So(response.ErrorCode, convey.ShouldEqual, constant.InsReqSuccessCode)
	})
}

func TestHandleInstanceCreate(t *testing.T) {
	tests := []struct {
		name         string
		funcKey      string
		extraData    []byte
		traceID      string
		mockFuncSpec *types.FunctionSpecification
		mockDataInfo *extraDataInfo
		mockResSpec  *resspeckey.ResourceSpecification
		mockInstance *types.Instance
		mockError    snerror.SNError
	}{
		{
			name:         "Function does not exist",
			funcKey:      "nonexistent-func",
			extraData:    []byte{},
			traceID:      "trace1",
			mockFuncSpec: nil,
		},
		{
			name:         "Failed to parse extra data",
			funcKey:      "test-func",
			extraData:    []byte("invalid-data"),
			traceID:      "trace2",
			mockFuncSpec: &types.FunctionSpecification{},
			mockError:    snerror.New(1, "parse error"),
		},
		{
			name:         "Failed to get resource specification",
			funcKey:      "test-func",
			extraData:    []byte("valid-data"),
			traceID:      "trace3",
			mockFuncSpec: &types.FunctionSpecification{},
			mockDataInfo: &extraDataInfo{resourceData: []byte("resourceData"), invokeLabel: []byte("invokeLabel")},
			mockError:    snerror.New(1, "resSpec error"),
		},
		{
			name:         "Failed to create instance",
			funcKey:      "test-func",
			extraData:    []byte("valid-data"),
			traceID:      "trace4",
			mockFuncSpec: &types.FunctionSpecification{},
			mockDataInfo: &extraDataInfo{resourceData: []byte("resourceData"), invokeLabel: []byte("invokeLabel")},
			mockError:    snerror.New(1, "create instance error"),
		},
		{
			name:         "Successfully created instance",
			funcKey:      "test-func",
			extraData:    []byte("valid-data"),
			traceID:      "trace5",
			mockFuncSpec: &types.FunctionSpecification{},
			mockDataInfo: &extraDataInfo{designateInstanceName: "instanceName", resourceData: []byte("resourceData"),
				invokeLabel: []byte("invokeLabel")},
			mockInstance: &types.Instance{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FaaSScheduler{
				PoolManager: &instancepool.PoolManager{},
			}

			patches := NewPatches()
			defer patches.Reset()

			patches.ApplyFunc(registry.GlobalRegistry.GetFuncSpec,
				func(funcKey string) *types.FunctionSpecification {
					return tt.mockFuncSpec
				})

			patches.ApplyMethod(reflect.TypeOf(fs.PoolManager), "CreateInstance",
				func(_ *instancepool.PoolManager, req *types.InstanceCreateRequest) (*types.Instance, snerror.SNError) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockInstance, nil
				})

			result := fs.handleInstanceCreate(tt.funcKey, tt.extraData, nil, tt.traceID)

			assert.NotNil(t, result)
		})
	}
}

func TestFaaSScheduler_parseExtraData(t *testing.T) {
	convey.Convey("test parseExtraData", t, func() {
		convey.Convey("baseline", func() {
			data := map[string][]byte{
				"instanceName":          []byte("testInstanceName"),
				"designateInstanceID":   []byte("testInstanceID"),
				"instanceCreateEvent":   []byte("testCreateEvent"),
				"resourcesData":         []byte("testResourceData"),
				"instanceCallerPodName": []byte("testPodName"),
				"poolLabel":             []byte("testPoolLabel"),
			}
			dataBytes, _ := json.Marshal(data)
			dataInfo, err := parseExtraData(dataBytes)
			convey.So(err, convey.ShouldBeNil)
			convey.So(dataInfo, convey.ShouldNotBeEmpty)
			convey.So(dataInfo.designateInstanceName, convey.ShouldEqual, "testInstanceName")
			convey.So(dataInfo.designateInstanceID, convey.ShouldEqual, "testInstanceID")
		})
		convey.Convey("invalid session config", func() {
			data := map[string][]byte{
				"instanceSessionConfig": []byte(`{"sessionID":"","sessionTTL":10}`),
			}
			dataBytes, _ := json.Marshal(data)
			_, err := parseExtraData(dataBytes)
			convey.So(err.Code(), convey.ShouldEqual, statuscode.InstanceSessionInvalidErrCode)
		})
		convey.Convey("invalid session concurrency", func() {
			data := map[string][]byte{
				"instanceSessionConfig": []byte(`{"sessionID":"test","sessionTTL":10,"concurrency":-1}`),
			}
			dataBytes, _ := json.Marshal(data)
			dataInfo, err := parseExtraData(dataBytes)
			convey.So(err, convey.ShouldBeNil)
			convey.So(dataInfo.instanceSession.Concurrency, convey.ShouldEqual, 1)
		})
	})
}

func TestGetResourceSpecification(t *testing.T) {
	defaultFuncSpec := &types.FunctionSpecification{
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:              300,
			Memory:           128,
			EphemeralStorage: 500,
		},
	}

	tests := []struct {
		name                   string
		resData                []byte
		labelData              []byte
		targetCPU              int64
		targetMem              int64
		targetEphemeralStorage int
		targetLabel            string
		targetErr              error
	}{
		{
			name:                   "normal",
			resData:                []byte("{\"CPU\": 500, \"Memory\": 512}"),
			labelData:              []byte("{\"X-Instance-Label\": \"aaaaa\"}"),
			targetCPU:              500,
			targetMem:              512,
			targetEphemeralStorage: 500,
			targetLabel:            "aaaaa",
			targetErr:              nil,
		},
		{
			name:                   "no resData",
			resData:                []byte("{}"),
			labelData:              []byte("{\"X-Instance-Label\": \"aaaaa\"}"),
			targetCPU:              300,
			targetMem:              128,
			targetEphemeralStorage: 500,
			targetLabel:            "aaaaa",
			targetErr:              nil,
		},
		{
			name:                   "unmarshal error",
			resData:                []byte("{\"CPU\": 500, \"Memory\": 512, \"test\": []}"),
			labelData:              []byte("{\"X-Instance-Label\": \"aaaaa\"}"),
			targetCPU:              500,
			targetMem:              512,
			targetEphemeralStorage: 500,
			targetLabel:            "aaaaa",
			targetErr:              fmt.Errorf("expected int or string, but got []"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resSpec, err := getResourceSpecification(tt.resData, tt.labelData, defaultFuncSpec)
			if err != nil {
				assert.Equal(t, tt.targetErr.Error(), err.Error())
				return
			}
			assert.Equal(t, tt.targetCPU, resSpec.CPU)
			assert.Equal(t, tt.targetMem, resSpec.Memory)
			assert.Equal(t, tt.targetEphemeralStorage, resSpec.EphemeralStorage)
			assert.Equal(t, tt.targetLabel, resSpec.InvokeLabel)
		})
	}
}

func TestFaaSScheduler_HandleRequestForward(t *testing.T) {
	faasScheduler := &FaaSScheduler{
		allocRecord: sync.Map{},
	}
	convey.Convey("Test HandleRequestForward", t, func() {
		convey.Convey("IsGaryUpdating is false", func() {
			rollout.GetGlobalRolloutHandler().IsGaryUpdating = false
			_, _, shouldReply := faasScheduler.HandleRequestForward(InstanceOperation("acquire"), []api.Arg{}, "")
			convey.So(shouldReply, convey.ShouldBeFalse)
		})
		convey.Convey("acquire should forward,invoke failed and reply", func() {
			args := []api.Arg{
				{
					Type:            0,
					Data:            []byte("acquire#aaa"),
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
			}
			rollout.GetGlobalRolloutHandler().IsGaryUpdating = true
			rollout.GetGlobalRolloutHandler().ForwardInstance = "instance"
			rolloutRatio := &rollout.RolloutRatio{
				RolloutRatio: "100%",
			}
			ratio, _ := json.Marshal(rolloutRatio)
			_ = rollout.GetGlobalRolloutHandler().ProcessRatioUpdate(ratio)

			defer ApplyFunc(rollout.InvokeByInstanceId, func(args []api.Arg, instanceID string,
				traceID string) ([]byte, error) {
				return nil, errors.New("invoke instance failed")
			}).Reset()
			_, err, shouldReply := faasScheduler.HandleRequestForward(InstanceOperation("acquire"), args, "")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invoke instance failed")
			convey.So(shouldReply, convey.ShouldBeFalse)
		})
		convey.Convey("acquire should forward,invoke success and reply", func() {
			args := []api.Arg{
				{
					Type:            0,
					Data:            []byte("acquire#aaa"),
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
			}
			rollout.GetGlobalRolloutHandler().IsGaryUpdating = true
			rolloutRatio := &rollout.RolloutRatio{
				RolloutRatio: "100%",
			}
			ratio, _ := json.Marshal(rolloutRatio)
			_ = rollout.GetGlobalRolloutHandler().ProcessRatioUpdate(ratio)

			defer ApplyFunc(rollout.InvokeByInstanceId, func(args []api.Arg, instanceID string,
				traceID string) ([]byte, error) {
				response := &commonTypes.InstanceResponse{}
				data, _ := json.Marshal(response)
				return data, nil
			}).Reset()
			_, err, shouldReply := faasScheduler.HandleRequestForward(InstanceOperation("acquire"), args, "")
			convey.So(err, convey.ShouldBeNil)
			convey.So(shouldReply, convey.ShouldBeTrue)
		})
		convey.Convey("acquire should forward,invoke success but no instance available not reply", func() {
			args := []api.Arg{
				{
					Type:            0,
					Data:            []byte("acquire#aaa"),
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
			}
			rollout.GetGlobalRolloutHandler().IsGaryUpdating = true
			rolloutRatio := &rollout.RolloutRatio{
				RolloutRatio: "100%",
			}
			ratio, _ := json.Marshal(rolloutRatio)
			_ = rollout.GetGlobalRolloutHandler().ProcessRatioUpdate(ratio)

			defer ApplyFunc(rollout.InvokeByInstanceId, func(args []api.Arg, instanceID string,
				traceID string) ([]byte, error) {
				response := &commonTypes.InstanceResponse{
					ErrorMessage: "no instance available",
					ErrorCode:    statuscode.NoInstanceAvailableErrCode,
				}
				data, _ := json.Marshal(response)
				return data, nil
			}).Reset()
			_, err, shouldReply := faasScheduler.HandleRequestForward(InstanceOperation("acquire"), args, "")
			convey.So(err, convey.ShouldBeNil)
			convey.So(shouldReply, convey.ShouldBeFalse)
		})
		convey.Convey("retain should forward and not reply", func() {
			args := []api.Arg{
				{
					Type:            0,
					Data:            []byte("retain#aaa"),
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
				{
					Type:            0,
					Data:            nil,
					ObjectID:        "",
					NestedObjectIDs: nil,
				},
			}
			defer ApplyFunc(rollout.InvokeByInstanceId, func(args []api.Arg, instanceID string,
				traceID string) ([]byte, error) {
				response := &commonTypes.InstanceResponse{}
				data, _ := json.Marshal(response)
				return data, nil
			}).Reset()
			rollout.GetGlobalRolloutHandler().IsGaryUpdating = true
			_, _, shouldReply := faasScheduler.HandleRequestForward(InstanceOperation("retain"), args, "")
			convey.So(shouldReply, convey.ShouldBeFalse)
		})
	})
}

func TestFaaSScheduler_handleRollout(t *testing.T) {
	faasScheduler := &FaaSScheduler{
		allocRecord: sync.Map{},
	}
	defer func() {
		selfregister.RegisterKey = ""
	}()
	convey.Convey("Test handleRollout", t, func() {
		alloc1 := &types.InstanceAllocation{
			Instance: &types.Instance{
				FuncKey: "funcA",
			},
		}
		alloc2 := &types.InstanceAllocation{
			Instance: &types.Instance{
				FuncKey: "funcA",
			},
		}
		alloc3 := &types.InstanceAllocation{
			Instance: &types.Instance{
				FuncKey: "funcB",
			},
		}
		faasScheduler.allocRecord.Store("lease1", alloc1)
		faasScheduler.allocRecord.Store("lease2", alloc2)
		faasScheduler.allocRecord.Store("lease3", alloc3)
		rsp := faasScheduler.handleRollout("instance1", "123")
		convey.So(rsp.ErrorCode, convey.ShouldEqual, statuscode.InternalErrorCode)
		config.GlobalConfig.EnableRollout = true
		rsp = faasScheduler.handleRollout("instance1", "123")
		convey.So(rsp.ErrorCode, convey.ShouldEqual, statuscode.InternalErrorCode)
		config.GlobalConfig.SchedulerDiscovery = &types.SchedulerDiscovery{RegisterMode: types.RegisterTypeContend}
		rsp = faasScheduler.handleRollout("instance1", "123")
		convey.So(rsp.ErrorCode, convey.ShouldEqual, statuscode.InternalErrorCode)
		selfregister.Registered = true
		selfregister.RegisterKey = "testKey"
		rsp = faasScheduler.handleRollout("instance1", "123")
		convey.So(rsp.ErrorCode, convey.ShouldEqual, constant.InsReqSuccessCode)
		convey.So(rsp.RegisterKey, convey.ShouldEqual, "testKey")
		convey.So(len(rsp.AllocRecord["funcA"]), convey.ShouldEqual, 2)
		convey.So(len(rsp.AllocRecord["funcB"]), convey.ShouldEqual, 1)
	})
}
func TestSyncAllocRecord(t *testing.T) {
	tests := []struct {
		name            string
		allocRecord     map[string][]string
		mockFuncSpec    *types.FunctionSpecification
		mockAcquireErr  snerror.SNError
		mockAllocResult *types.InstanceAllocation
		expectedCount   int
	}{
		{
			name: "Function does not exist",
			allocRecord: map[string][]string{
				"nonexistent-func": {"instance1-allocID"},
			},
			mockFuncSpec:  nil,
			expectedCount: 0,
		},
		{
			name: "AcquireInstance returns error",
			allocRecord: map[string][]string{
				"test-func": {"instance2-allocID"},
			},
			mockFuncSpec:   &types.FunctionSpecification{},
			mockAcquireErr: snerror.New(1, "acquire error"),
			expectedCount:  0,
		},
		{
			name: "Successfully acquire instance",
			allocRecord: map[string][]string{
				"test-func": {"instance3-allocID"},
			},
			mockFuncSpec:    &types.FunctionSpecification{},
			mockAllocResult: &types.InstanceAllocation{AllocationID: "alloc123"},
			expectedCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FaaSScheduler{
				allocRecord: sync.Map{},
				PoolManager: &instancepool.PoolManager{},
			}

			patches := NewPatches()
			defer patches.Reset()
			defer ApplyMethod(reflect.TypeOf(registry.GlobalRegistry), "GetFuncSpec",
				func(_ *registry.Registry, funcKey string) *types.FunctionSpecification {
					return tt.mockFuncSpec
				}).Reset()

			// Mock PoolManager.AcquireInstanceThread
			patches.ApplyMethod(
				reflect.TypeOf(fs.PoolManager),
				"AcquireInstanceThread",
				func(_ *instancepool.PoolManager, req *types.InstanceAcquireRequest) (*types.InstanceAllocation, snerror.SNError) {
					return tt.mockAllocResult, tt.mockAcquireErr
				},
			)

			fs.syncAllocRecord(tt.allocRecord)

			actualCount := 0
			fs.allocRecord.Range(func(key, value interface{}) bool {
				actualCount++
				return true
			})
			assert.Equal(t, tt.expectedCount, actualCount, "allocRecord count mismatch")
		})
	}
}

func TestReacquireLease(t *testing.T) {
	patches := NewPatches()
	defer patches.Reset()
	patches.ApplyFunc((*instancepool.PoolManager).AcquireInstanceThread, func(_ *instancepool.PoolManager, req *types.InstanceAcquireRequest) (*types.InstanceAllocation, snerror.SNError) {
		return &types.InstanceAllocation{Instance: &types.Instance{
			FuncKey: "",
		}}, nil
	})
	patches.ApplyMethod(reflect.TypeOf(registry.GlobalRegistry), "GetFuncSpec", func(_ *registry.Registry, funcKey string) *types.FunctionSpecification {
		return &types.FunctionSpecification{ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:                 0,
			Memory:              0,
			GpuMemory:           0,
			EnableDynamicMemory: false,
			CustomResources:     "",
			EnableTmpExpansion:  false,
			EphemeralStorage:    0,
			CustomResourcesSpec: "",
		}}
	})

	fs := &FaaSScheduler{
		allocRecord: sync.Map{},
		PoolManager: &instancepool.PoolManager{},
	}
	resp := fs.handleInstanceBatchRetain("e58bd817-1132-4b5b-8000-00000000009c-thread5f0d3377-59", []byte("{\"e58bd817-1132-4b5b-8000-00000000009c-thread5f0d3377-59\":{\"avgProcTime\":307,\"functionKey\":\"12345678901234561234567890123456/0@functest@functest/latest\",\"isAbnormal\":false,\"maxProcTime\":323,\"procReqNum\":217,\"reacquireData\":[123,34,114,101,115,111,117,114,99,101,115,68,97,116,97,34,58,91,49,50,51,44,51,52,44,54,55,44,56,48,44,56,53,44,51,52,44,53,56,44,53,52,44,52,56,44,52,56,44,52,52,44,51,52,44,55,55,44,49,48,49,44,49,48,57,44,49,49,49,44,49,49,52,44,49,50,49,44,51,52,44,53,56,44,53,51,44,52,57,44,53,48,44,49,50,53,93,125]}}"), "aaaaa")
	assert.Equal(t, len(resp.InstanceAllocSucceed), 1)
}

func TestSetupAgentCRsManager(t *testing.T) {
	faasScheduler := &FaaSScheduler{
		funcSpecCh:  make(chan registry.SubEvent, 1),
		insConfigCh: make(chan registry.SubEvent, 1),
	}
	oldAgentRegistry := registry.GlobalRegistry.AgentRegistry
	oldInstanceRegistry := registry.GlobalRegistry.InstanceRegistry
	oldFaaSSchedulerRegistry := registry.GlobalRegistry.FaaSSchedulerRegistry
	defer func() {
		registry.GlobalRegistry.AgentRegistry = oldAgentRegistry
		registry.GlobalRegistry.InstanceRegistry = oldInstanceRegistry
		registry.GlobalRegistry.FaaSSchedulerRegistry = oldFaaSSchedulerRegistry
	}()
	registry.GlobalRegistry.AgentRegistry = &registry.AgentRegistry{}
	registry.GlobalRegistry.InstanceRegistry = &registry.InstanceRegistry{}
	registry.GlobalRegistry.FaaSSchedulerRegistry = &registry.FaasSchedulerRegistry{}
	convey.Convey("Given the setupAgentCRsManager function", t, func() {

		setupAgentCRsManager(nil, faasScheduler)
		_ = os.Unsetenv(constant.EnableAgentCRDRegistry)

		convey.So(func() { setupAgentCRsManager(nil, faasScheduler) }, convey.ShouldNotPanic)
		os.Setenv(constant.EnableAgentCRDRegistry, "true")
		defer os.Unsetenv(constant.EnableAgentCRDRegistry)
		convey.So(func() { setupAgentCRsManager(nil, faasScheduler) }, convey.ShouldNotPanic)
	})
}

func TestAcquireNonOwnerSchedulerErrorCode(t *testing.T) {
	convey.Convey("Test NoOwnerSchedulerErrorCode", t, func() {
		faasScheduler := &FaaSScheduler{
			funcSpecCh:  make(chan registry.SubEvent, 1),
			insConfigCh: make(chan registry.SubEvent, 1),
		}
		expectSchedulerInstanceId := "test-scheduler-1"
		expectOk := false
		defer ApplyMethodFunc(selfregister.GlobalSchedulerProxy, "CheckFuncOwner", func(string) (string, bool) {
			return expectSchedulerInstanceId, expectOk
		}).Reset()
		targetName := "target1"
		testData := make(map[string]*types.InstanceThreadMetrics)
		testData[targetName] = &types.InstanceThreadMetrics{
			InsThdID:      "ins-thd-001",
			ProcNumPS:     120.5,
			ProcReqNum:    2400,
			AvgProcTime:   85,
			MaxProcTime:   210,
			IsAbnormal:    false,
			ReacquireData: []byte("reacq-001-data"),
			FunctionKey:   "ProcessUserRequest",
		}
		bytes, _ := json.Marshal(testData)

		resp := faasScheduler.handleInstanceAcquire(targetName, bytes, "traceId-123")
		convey.So(resp.ErrorCode, convey.ShouldEqual, statuscode.AcquireNonOwnerSchedulerErrorCode)
		convey.So(resp.ErrorMessage, convey.ShouldEqual, expectSchedulerInstanceId)
		resp1 := faasScheduler.handleInstanceBatchRetain(targetName, bytes, "traceId-123")
		convey.So(resp1.InstanceAllocFailed, convey.ShouldContainKey, targetName)
		convey.So(resp1.InstanceAllocFailed[targetName].ErrorCode, convey.ShouldEqual, statuscode.AcquireNonOwnerSchedulerErrorCode)

		expectOk = true
		resp = faasScheduler.handleInstanceAcquire(targetName, bytes, "traceId-123")
		convey.So(resp.ErrorCode, convey.ShouldNotEqual, statuscode.AcquireNonOwnerSchedulerErrorCode)
		resp1 = faasScheduler.handleInstanceBatchRetain(targetName, bytes, "traceId-123")
		_, ok := resp1.InstanceAllocFailed[targetName]
		if ok {
			convey.So(resp1.InstanceAllocFailed[targetName].ErrorCode, convey.ShouldNotEqual, statuscode.AcquireNonOwnerSchedulerErrorCode)
		}
	})
}
