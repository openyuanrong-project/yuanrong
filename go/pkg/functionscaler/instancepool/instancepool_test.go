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

// Package instancepool -
package instancepool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	wisecloudTypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/instancequeue"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/scaler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/concurrencyscheduler"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

var (
	pool InstancePool
)

func initRegistry() {
	defer ApplyFunc((*registry.FaasSchedulerRegistry).WaitForETCDList, func() {}).Reset()
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

func CreateTestInstancePool() InstancePool {
	funcSpec := &types.FunctionSpecification{
		FuncKey: "testFunction",
		FuncCtx: context.TODO(),
		FuncMetaData: commonTypes.FuncMetaData{
			Handler: "myHandler",
		},
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 500,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			MaxInstance:   100,
			MinInstance:   1,
			ConcurrentNum: 1,
		},
		ExtendedMetaData: commonTypes.ExtendedMetaData{
			Initializer: commonTypes.Initializer{
				Handler: "myInitializer",
			},
		},
	}
	initRegistry()
	insPool, _ := NewGenericInstancePool(funcSpec, faasManagerInfo{})
	return insPool
}

func CreateTestInstancePoolWithVPC() InstancePool {
	funcSpec := &types.FunctionSpecification{
		FuncKey: "testFunction",
		FuncCtx: context.TODO(),
		FuncMetaData: commonTypes.FuncMetaData{
			Handler: "myHandler",
		},
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 500,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			MaxInstance:   100,
			ConcurrentNum: 100,
		},
		ExtendedMetaData: commonTypes.ExtendedMetaData{
			Initializer: commonTypes.Initializer{
				Handler: "myInitializer",
			},
			VpcConfig: &commonTypes.VpcConfig{},
		},
	}
	initRegistry()
	insPool, _ := NewGenericInstancePool(funcSpec, faasManagerInfo{})
	insPool.HandleFaaSManagerUpdate(faasManagerInfo{
		funcKey:    "faasManager",
		instanceID: "faasManagerInstance",
	})
	return insPool
}

func TestNewGenericInstancePool(t *testing.T) {
	funcSpec := &types.FunctionSpecification{
		FuncKey: "testFunction",
		FuncCtx: context.TODO(),
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 500,
		},
	}
	initRegistry()
	pool, _ = NewGenericInstancePool(funcSpec, faasManagerInfo{})
	assert.Equal(t, true, pool != nil)
}

func TestHandleFuncEvent(t *testing.T) {
	config.GlobalConfig.Scenario = ""
	insPool := CreateTestInstancePool()
	funcSpecOld := &types.FunctionSpecification{
		FuncKey: "testFunction",
		FuncCtx: context.TODO(),
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 500,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			ConcurrentNum: 100,
			MaxInstance:   3,
			MinInstance:   1,
		},
	}
	insPool.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, &instanceconfig.Configuration{
		FuncKey:       "testFunction",
		InstanceLabel: "",
		InstanceMetaData: commonTypes.InstanceMetaData{
			ConcurrentNum: 100,
			MaxInstance:   3,
			MinInstance:   1,
		},
		NuwaRuntimeInfo: wisecloudTypes.NuwaRuntimeInfo{},
	})

	insPool.HandleFunctionEvent(registry.SubEventTypeUpdate, funcSpecOld)
	funcSpecNew := &types.FunctionSpecification{
		FuncKey: "testFunction",
		FuncCtx: context.TODO(),
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    1000,
			Memory: 1000,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			ConcurrentNum: 100,
			MaxInstance:   3,
			MinInstance:   1,
		},
	}
	insPool.HandleFunctionEvent(registry.SubEventTypeUpdate, funcSpecNew)
	genInsPool := insPool.(*GenericInstancePool)
	oldResKey := resspeckey.ConvertToResSpecKey(resspeckey.ConvertResourceMetaDataToResSpec(funcSpecOld.ResourceMetaData))
	newResKey := resspeckey.ConvertToResSpecKey(resspeckey.ConvertResourceMetaDataToResSpec(funcSpecNew.ResourceMetaData))
	assert.Equal(t, 1, len(genInsPool.reservedInstanceQueue))
	assert.Equal(t, 2, len(genInsPool.scaledInstanceQueue))
	assert.Equal(t, true, genInsPool.reservedInstanceQueue[oldResKey] == nil)
	assert.Equal(t, true, genInsPool.reservedInstanceQueue[newResKey] != nil)
}

func TestHandleInsConfigEvent(t *testing.T) {
	insPool := CreateTestInstancePool()
	insConfig := &instanceconfig.Configuration{
		FuncKey:       "testFunction",
		InstanceLabel: "aaaaa",
	}
	insPool.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, insConfig)
	insPool.HandleInstanceConfigEvent(registry.SubEventTypeDelete, insConfig)
	insAcqReq := &types.InstanceAcquireRequest{
		ResSpec: &resspeckey.ResourceSpecification{
			CPU:         600,
			Memory:      512,
			InvokeLabel: "aaaaa",
		},
	}

	_, err := insPool.AcquireInstance(insAcqReq)
	assert.NotNil(t, err)
}

func TestGetFuncSpec(t *testing.T) {
	insPool := CreateTestInstancePool()
	funSpec := insPool.GetFuncSpec()
	assert.Equal(t, "testFunction", funSpec.FuncKey)
}

func TestGenericInstancePool_CreateInstance(t *testing.T) {
	insPool := CreateTestInstancePool()
	convey.Convey("Test CreateInstance", t, func() {
		defer ApplyFunc((*instancequeue.OnDemandInstanceQueue).CreateInstance, func(_ *instancequeue.OnDemandInstanceQueue,
			insCrtReq *types.InstanceCreateRequest) (*types.Instance, snerror.SNError) {
			return &types.Instance{}, nil
		}).Reset()
		instance, err := insPool.CreateInstance(&types.InstanceCreateRequest{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(instance, convey.ShouldNotBeNil)
	})
}

func TestGenericInstancePool_DeleteInstance(t *testing.T) {
	insPool := CreateTestInstancePool().(*GenericInstancePool)
	convey.Convey("Test CreateInstance", t, func() {
		defer ApplyFunc((*instancequeue.OnDemandInstanceQueue).DeleteInstance, func(_ *instancequeue.OnDemandInstanceQueue,
			instance *types.Instance) snerror.SNError {
			return nil
		}).Reset()
		resKey := resspeckey.ResSpecKey{
			CPU:    100,
			Memory: 100,
		}
		instance := &types.Instance{
			ResKey: resKey,
		}
		err := insPool.DeleteInstance(instance)
		convey.So(err, convey.ShouldNotBeNil)
		insPool.onDemandInstanceQueue[resKey] = &instancequeue.OnDemandInstanceQueue{}
		err = insPool.DeleteInstance(instance)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestAcquireInstance(t *testing.T) {
	patches := mockInstanceOperation()
	defer unMockInstanceOperation(patches)
	convey.Convey("Test AcquireInstance", t, func() {
		convey.Convey("AcquireInstance success", func() {
			defer ApplyFunc(state.Update, func(value interface{}, tags ...string) {
			}).Reset()
			insPool := CreateTestInstancePool()
			insPool.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, &instanceconfig.Configuration{
				FuncKey: "test",
				InstanceMetaData: commonTypes.InstanceMetaData{
					MaxInstance:   100,
					ConcurrentNum: 1,
				},
			})
			insAcqReq := &types.InstanceAcquireRequest{
				ResSpec: &resspeckey.ResourceSpecification{
					CPU:    600,
					Memory: 512,
				},
			}
			insAlloc, err := insPool.AcquireInstance(insAcqReq)
			assert.Equal(t, nil, err)
			assert.Equal(t, true, insAlloc != nil)
		})
		convey.Convey("InstanceQueue AcquireInstance is not nil", func() {
			patchGet := ApplyMethod(reflect.TypeOf(&instancequeue.ScaledInstanceQueue{}),
				"AcquireInstance", func(_ *instancequeue.ScaledInstanceQueue) (thread *types.InstanceAllocation,
					acquireErr snerror.SNError) {
					return nil, snerror.New(0, "AcquireInstance error")
				})
			defer patchGet.Reset()
			insPool := CreateTestInstancePool()
			insThdApp := &types.InstanceAcquireRequest{}
			thd, err := insPool.AcquireInstance(insThdApp)
			convey.So(thd, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("acquire on-demand instance", func() {
			patchGet := ApplyMethod(reflect.TypeOf(&instancequeue.OnDemandInstanceQueue{}),
				"AcquireInstance", func(_ *instancequeue.OnDemandInstanceQueue) (thread *types.InstanceAllocation,
					acquireErr snerror.SNError) {
					return nil, snerror.New(0, "AcquireInstance error")
				})
			defer patchGet.Reset()
			insPool := CreateTestInstancePool()
			insThdApp := &types.InstanceAcquireRequest{InstanceName: "testInstance"}
			thd, err := insPool.AcquireInstance(insThdApp)
			convey.So(thd, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("acquire state instance", func() {
			patch := ApplyGlobalVar(&config.GlobalConfig, types.Configuration{
				AutoScaleConfig: types.AutoScaleConfig{
					SLAQuota:      1000,
					ScaleDownTime: 100,
					BurstScaleNum: 100000,
				},
			})
			defer patch.Reset()
			insPool := CreateTestInstancePool().(*GenericInstancePool)
			insPool.FuncSpec.FuncMetaData.IsStatefulFunction = true
			var createErr snerror.SNError
			insPool.stateRoute.createInstanceFunc = func(resSpec *resspeckey.ResourceSpecification, instanceType types.InstanceType,
				callerPodName string) (*types.Instance, error) {
				if createErr != nil {
					return nil, createErr
				}
				return &types.Instance{InstanceID: "instance1"}, nil
			}
			insThdApp := &types.InstanceAcquireRequest{StateID: "aaa"}
			createErr = snerror.New(statuscode.StatusInternalServerError, "internal error")
			insAlloc, err := insPool.AcquireInstance(insThdApp)
			convey.So(insAlloc, convey.ShouldBeNil)
			convey.So(err.Code(), convey.ShouldEqual, statuscode.NoInstanceAvailableErrCode)
			time.Sleep(200 * time.Millisecond)
			createErr = nil
			insAlloc, err = insPool.AcquireInstance(insThdApp)
			convey.So(err, convey.ShouldBeNil)
			convey.So(insAlloc.Instance.InstanceID, convey.ShouldEqual, "instance1")
			time.Sleep(200 * time.Millisecond)
			insAlloc, err = insPool.AcquireInstance(insThdApp)
			convey.So(err, convey.ShouldBeNil)
			convey.So(insAlloc.Instance.InstanceID, convey.ShouldEqual, "instance1")
		})
		convey.Convey("acquire session instance", func() {
			insPool := CreateTestInstancePool().(*GenericInstancePool)
			insPool.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, &instanceconfig.Configuration{
				InstanceMetaData: commonTypes.InstanceMetaData{
					MinInstance: 1,
					MaxInstance: 10,
				},
			})
			insAcqReq1 := &types.InstanceAcquireRequest{
				InstanceSession: commonTypes.InstanceSessionConfig{
					SessionID:   "123",
					SessionTTL:  10,
					Concurrency: 1,
				},
				ResSpec: &resspeckey.ResourceSpecification{
					CPU:    500,
					Memory: 500,
				},
			}
			insAlloc1, err := insPool.AcquireInstance(insAcqReq1)
			convey.So(err, convey.ShouldBeNil)
			record, exist := insPool.sessionRecordMap["123"]
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(record.instance, convey.ShouldEqual, insAlloc1.Instance)
			insAlloc2, err := insPool.AcquireInstance(insAcqReq1)
			convey.So(err, convey.ShouldBeNil)
			convey.So(insAlloc2.Instance.InstanceID, convey.ShouldEqual, insAlloc1.Instance.InstanceID)
		})
	})
}

func TestReleaseInstanceThread(t *testing.T) {
	patches := mockInstanceOperation()
	defer unMockInstanceOperation(patches)
	defer ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	initRegistry()
	insPool := CreateTestInstancePool()
	insPool.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, &instanceconfig.Configuration{
		FuncKey: "test",
		InstanceMetaData: commonTypes.InstanceMetaData{
			MaxInstance:   100,
			ConcurrentNum: 1,
		},
	})
	insThdApp := &types.InstanceAcquireRequest{
		ResSpec: &resspeckey.ResourceSpecification{
			CPU:    600,
			Memory: 512,
		},
	}
	thd, err := insPool.AcquireInstance(insThdApp)
	assert.Equal(t, nil, err)
	insPool.ReleaseInstance(thd)

	thd.Instance.InstanceType = types.InstanceTypeReserved
	insPool.ReleaseInstance(thd)

	thd.Instance.InstanceType = types.InstanceTypeScaled
	insPool.ReleaseInstance(thd)
}

func TestReleaseAbnormalInstance(t *testing.T) {
	patches := mockInstanceOperation()
	defer unMockInstanceOperation(patches)
	convey.Convey("test ReleaseAbnormalInstance", t, func() {
		patches := []*Patches{
			ApplyMethod(reflect.TypeOf(&instancequeue.ScaledInstanceQueue{}), "HandleFaultyInstance",
				func(_ *instancequeue.ScaledInstanceQueue, instance *types.Instance) {
				}),
		}
		defer func() {
			for _, patch := range patches {
				patch.Reset()
			}
		}()
		registry.GlobalRegistry = &registry.Registry{FaaSSchedulerRegistry: registry.NewFaasSchedulerRegistry(make(chan struct{}))}
		insPool := CreateTestInstancePool()
		instance := &types.Instance{
			InstanceType: "reserved",
			InstanceID:   "instanceID",
			ResKey: resspeckey.ResSpecKey{
				CPU:    888,
				Memory: 888,
			},
		}
		insPool.HandleInstanceEvent(registry.SubEventTypeRemove, instance)
		instance.InstanceType = "scaled"
		insPool.HandleInstanceEvent(registry.SubEventTypeRemove, instance)
		instance.InstanceType = "default"
		insPool.HandleInstanceEvent(registry.SubEventTypeRemove, instance)
	})
}

func BenchmarkAcquireInstanceThread(b *testing.B) {
	patches := mockInstanceOperation()
	defer unMockInstanceOperation(patches)
	defer ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	insPool := CreateTestInstancePool()
	insThdApp := &types.InstanceAcquireRequest{
		ResSpec: &resspeckey.ResourceSpecification{
			CPU:    600,
			Memory: 512,
		},
	}
	for i := 0; i < b.N; i++ {
		_, err := insPool.AcquireInstance(insThdApp)
		if err != nil {
			b.Errorf("acquire instance thread error %s", err.Error())
		}
	}
}

func TestAcquireInstanceThreadWithVPC(t *testing.T) {
	defer ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	insPool := CreateTestInstancePoolWithVPC().(*GenericInstancePool)
	insPool.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, &instanceconfig.Configuration{
		FuncKey: "test",
		InstanceMetaData: commonTypes.InstanceMetaData{
			MaxInstance:   100,
			ConcurrentNum: 1,
		},
	})
	patches := []*Patches{
		ApplyFunc(CreateInstance,
			func(request createInstanceRequest) (instance *types.Instance, createErr error) {
				createErr = errors.New("failed to create")
				return
			}),
	}

	// test create instance fail
	instance, err := insPool.createInstance("",
		"", types.InstanceTypeScaled, resspeckey.ResSpecKey{}, nil)
	time.Sleep(10 * time.Millisecond)
	assert.NotNil(t, err)

	// test create instance fail
	instance, err = insPool.createInstance("",
		"", types.InstanceTypeScaled, resspeckey.ResSpecKey{CustomResources: `{"npu":1}`}, nil)
	time.Sleep(10 * time.Millisecond)
	assert.NotNil(t, err)

	// test create instance success
	patches[0].Reset()
	patche := mockInstanceOperation()
	defer unMockInstanceOperation(patche)
	instance, err = insPool.createInstance("",
		"", types.InstanceTypeScaled, resspeckey.ResSpecKey{}, nil)
	time.Sleep(10 * time.Millisecond)
	assert.Nil(t, err)
	assert.NotNil(t, instance)

	// test createInstanceAndAddCallerPodName success
	instance, err = insPool.createInstanceAndAddCallerPodName(nil, types.InstanceTypeScaled, "callerPodName")
	time.Sleep(10 * time.Millisecond)
	assert.Nil(t, err)
	assert.NotNil(t, instance)
}

func TestHandleInsEvent(t *testing.T) {
	patches := mockInstanceOperation()
	defer unMockInstanceOperation(patches)
	insPool := CreateTestInstancePool()
	resKey := resspeckey.ResSpecKey{
		CPU:    500,
		Memory: 500,
	}
	instance := &types.Instance{
		InstanceID:    "instanceID",
		InstanceType:  types.InstanceTypeReserved,
		ResKey:        resKey,
		FuncKey:       "testFunction",
		ConcurrentNum: 1,
	}
	convey.Convey("HandleInsEvent", t, func() {
		convey.Convey("wait instance config", func() {
			start := time.Now()
			go func() {
				time.Sleep(2 * time.Second)
				insPool.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, &instanceconfig.Configuration{
					FuncKey: "testFunc",
					InstanceMetaData: commonTypes.InstanceMetaData{
						MaxInstance:   0,
						MinInstance:   0,
						ConcurrentNum: 0,
						InstanceType:  "reserved",
						IdleMode:      false,
					},
				})
			}()
			insPool.HandleInstanceEvent(registry.SubEventTypeUpdate, &types.Instance{
				ResKey:         resspeckey.ResSpecKey{},
				InstanceStatus: commonTypes.InstanceStatus{Code: 3},
				InstanceType:   "reserved",
				InstanceID:     "123456789",
				FuncKey:        "testFunc",
				FuncSig:        "",
				ConcurrentNum:  1,
			})
			insPool.HandleInstanceEvent(registry.SubEventTypeUpdate, &types.Instance{
				ResKey:         resspeckey.ResSpecKey{CPU: 500, Memory: 500},
				InstanceStatus: commonTypes.InstanceStatus{Code: 3},
				InstanceType:   "scaled",
				InstanceID:     "123456789-1",
				FuncKey:        "testFunc",
				FuncSig:        "",
				ConcurrentNum:  1,
			})
			end := time.Since(start)
			convey.So(end, convey.ShouldBeLessThan, 2*time.Second)
			insPool.HandleInstanceEvent(registry.SubEventTypeUpdate, &types.Instance{
				ResKey:         resspeckey.ResSpecKey{CPU: 500, Memory: 500},
				InstanceStatus: commonTypes.InstanceStatus{Code: 3},
				InstanceType:   "scaled",
				InstanceID:     "123456789-2",
				FuncKey:        "testFunc",
				FuncSig:        "",
				ConcurrentNum:  1,
			})

			insPool.HandleInstanceEvent(registry.SubEventTypeUpdate, instance)
			insPool.HandleInstanceEvent(registry.SubEventTypeDelete, instance)
			insPool.HandleInstanceEvent(registry.SubEventTypeUpdate, &types.Instance{
				ResKey:         resspeckey.ResSpecKey{CPU: 500, Memory: 500},
				InstanceStatus: commonTypes.InstanceStatus{Code: 7},
				InstanceType:   "scaled",
				InstanceID:     "123456789-1",
				FuncKey:        "testFunc",
				FuncSig:        "",
				ConcurrentNum:  1,
			})
			insPool.HandleInstanceEvent(registry.SubEventTypeUpdate, &types.Instance{
				ResKey:         resspeckey.ResSpecKey{CPU: 500, Memory: 500},
				InstanceStatus: commonTypes.InstanceStatus{Code: 10},
				InstanceType:   "scaled",
				InstanceID:     "123456789-2",
				FuncKey:        "testFunc",
				FuncSig:        "",
				ConcurrentNum:  1,
			})
		})
		convey.Convey("delete invalid instance", func() {
			var deleteTime int
			defer ApplyMethod(reflect.TypeOf(&instancequeue.ScaledInstanceQueue{}), "HandleFaultyInstance",
				func(_ *instancequeue.ScaledInstanceQueue, instance *types.Instance) {
					deleteTime++
				}).Reset()
			insPool.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, &instanceconfig.Configuration{
				FuncKey: "testFunc",
				InstanceMetaData: commonTypes.InstanceMetaData{
					MaxInstance:   10,
					MinInstance:   0,
					ConcurrentNum: 0,
					InstanceType:  "reserved",
					IdleMode:      false,
				},
			})
			insPool.HandleInstanceEvent(registry.SubEventTypeUpdate, &types.Instance{
				ResKey:         resspeckey.ResSpecKey{CPU: 500, Memory: 500},
				InstanceStatus: commonTypes.InstanceStatus{Code: 5},
				InstanceType:   types.InstanceTypeReserved,
				InstanceID:     "123456789-1",
				FuncKey:        "testFunc",
				FuncSig:        "",
				ConcurrentNum:  1,
			})
			convey.So(deleteTime, convey.ShouldEqual, 1)
		})
		convey.Convey("handle on-demand instance", func() {
			var (
				updateTime int
				deleteTime int
			)
			defer ApplyMethod(reflect.TypeOf(&instancequeue.OnDemandInstanceQueue{}), "HandleInstanceUpdate",
				func(_ *instancequeue.OnDemandInstanceQueue, instance *types.Instance) {
					updateTime++
				}).Reset()
			defer ApplyMethod(reflect.TypeOf(&instancequeue.OnDemandInstanceQueue{}), "HandleInstanceDelete",
				func(_ *instancequeue.OnDemandInstanceQueue, instance *types.Instance) {
					deleteTime++
				}).Reset()
			insPool.HandleInstanceEvent(registry.SubEventTypeUpdate, &types.Instance{
				ResKey:         resspeckey.ResSpecKey{CPU: 500, Memory: 500},
				InstanceStatus: commonTypes.InstanceStatus{Code: 3},
				InstanceType:   types.InstanceTypeOnDemand,
				InstanceID:     "instance1",
				FuncKey:        "function1",
				FuncSig:        "",
				ConcurrentNum:  1,
			})
			convey.So(updateTime, convey.ShouldEqual, 1)
			insPool.HandleInstanceEvent(registry.SubEventTypeDelete, &types.Instance{
				ResKey:         resspeckey.ResSpecKey{CPU: 500, Memory: 500},
				InstanceStatus: commonTypes.InstanceStatus{Code: 3},
				InstanceType:   types.InstanceTypeOnDemand,
				InstanceID:     "instance1",
				FuncKey:        "function1",
				FuncSig:        "",
				ConcurrentNum:  1,
			})
			convey.So(deleteTime, convey.ShouldEqual, 1)
		})
	})
}

func TestFilterInstanceIDMap(t *testing.T) {
	patches := mockInstanceOperation()
	defer unMockInstanceOperation(patches)
	filterMap := map[string]*types.Instance{}
	existsMap := map[string]map[string]*commonTypes.InstanceSpecification{}
	existsMap["321"] = map[string]*commonTypes.InstanceSpecification{"123": {CreateOptions: map[string]string{types.FunctionKeyNote: "321", types.InstanceTypeNote: "scaled"}}}
	gi := &GenericInstancePool{
		FuncSpec: &types.FunctionSpecification{FuncKey: "321"},
	}
	convey.Convey("map is nil", t, func() {
		gi.filterInstanceIDMap(filterMap, existsMap, "scaled")
		convey.So(len(existsMap), convey.ShouldEqual, 1)
	})
	filterMap["123"] = &types.Instance{FuncSig: "aaa"}
	filterMap["456"] = &types.Instance{FuncSig: "bbb"}
	convey.Convey("delete success", t, func() {
		gi.filterInstanceIDMap(filterMap, existsMap, "scaled")
		convey.So(filterMap, convey.ShouldContainKey, "123")
		convey.So(filterMap, convey.ShouldNotContainKey, "456")
	})
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
	config.GlobalConfig = types.Configuration{}
	config.GlobalConfig.AutoScaleConfig = types.AutoScaleConfig{
		SLAQuota:      1000,
		ScaleDownTime: 1000,
		BurstScaleNum: 100000,
	}
	config.GlobalConfig.LeaseSpan = 500
	config.GlobalConfig.LocalAuth = localauth.AuthConfig{
		AKey: "ENC(key=servicekek, value=6B6D73763030000101D615B6381ED56AF68123844D047428BDCCBF19957866" +
			"CD0D7F53C29438337667A93FB9A06C5ED4A3D925C87655E4C734)",
		SKey: "ENC(key=servicekek, value=6B6D73763030000101139308ABBC0C4120F949AC833416D5E6D8CA18D8C69E" +
			"4C5E03E553E18733B4119C4B716FF2C8265336BB2979545A24FDC07CDD6A6A02F412D0DE83BD43F2A07DDBC78EB2)",
		Duration: 0,
	}
	initRegistry()
	m.Run()
}

func mockInstanceOperation() []*Patches {
	SetGlobalSdkClient(&mockUtils.FakeLibruntimeSdkClient{})
	patches := []*Patches{
		ApplyFunc(CreateInstance, func(request createInstanceRequest) (instance *types.Instance, createErr error) {
			instance = &types.Instance{
				InstanceID:     uuid.New().String(),
				ConcurrentNum:  1,
				InstanceStatus: commonTypes.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
				ResKey:         resspeckey.ResSpecKey{InvokeLabel: ""},
			}
			return
		}),
		ApplyFunc(DeleteInstance, func(funcSpec *types.FunctionSpecification, faasManagerInfo faasManagerInfo,
			instance *types.Instance) error {
			return nil
		}),
		ApplyFunc(SignalInstance, func(instance *types.Instance, signal int) {
			return
		}),
	}
	return patches
}

func unMockInstanceOperation(patches []*Patches) {
	for _, patch := range patches {
		patch.Reset()
	}
}

type fakeInstanceQueue struct {
	instanceID map[string]struct{}
}

func (fq *fakeInstanceQueue) HandleFuncManagedChange() {
}

func (fq *fakeInstanceQueue) SetInstanceScheduler(instanceScheduler scheduler.InstanceScheduler) {
}

func (fq *fakeInstanceQueue) SetInstanceScaler(instanceScaler scaler.InstanceScaler) {
}

func (fq *fakeInstanceQueue) ScaleUpHandler(insNum int, callback scaler.ScaleUpCallback) {
}

func (fq *fakeInstanceQueue) ScaleDownHandler(insNum int, callback scaler.ScaleDownCallback) {
}

func (fq *fakeInstanceQueue) AcquireInstanceThread(DesignateInstanceID string, isLimiting bool) (*types.InstanceAllocation, snerror.SNError) {
	return nil, nil
}
func (fq *fakeInstanceQueue) ReleaseInstance(insThd *types.InstanceAllocation) snerror.SNError {
	return nil
}
func (fq *fakeInstanceQueue) HandleInstanceUpdate(instance *types.Instance)                 {}
func (fq *fakeInstanceQueue) HandleInstanceDelete(instance *types.Instance)                 {}
func (fq *fakeInstanceQueue) HandleFaultyInstance(instance *types.Instance)                 {}
func (fq *fakeInstanceQueue) HandleFuncSpecUpdate(funcSpec *types.FunctionSpecification)    {}
func (fq *fakeInstanceQueue) HandleInsConfigUpdate(insConfig *instanceconfig.Configuration) {}
func (fq *fakeInstanceQueue) Destroy()                                                      {}
func (fq *fakeInstanceQueue) RecoverInstance(input map[string]*types.Instance) {
	for key, _ := range input {
		if _, ok := fq.instanceID[key]; !ok {
			fq.instanceID[key] = struct{}{}
		}
	}
}
func (fq *fakeInstanceQueue) HandleAliasUpdate() {}

func (fq *fakeInstanceQueue) GetInstanceNumber(onlySelf bool) int { return 0 }
func (fq *fakeInstanceQueue) HandleInstanceCreating(instance *types.Instance) {
	return
}

func CreateScaledInstanceQueue(instanceType types.InstanceType) *instancequeue.ScaledInstanceQueue {
	basicInsQueConfig := &instancequeue.InsQueConfig{
		InstanceType: instanceType,
		FuncSpec:     &types.FunctionSpecification{},
		ResKey:       resspeckey.ResSpecKey{},
	}
	metricsCollector := &metrics.BucketCollector{}
	queue := instancequeue.NewScaledInstanceQueue(basicInsQueConfig, metricsCollector)
	InsThdReqQueue := requestqueue.NewInsAcqReqQueue("testFunction", 1000*time.Millisecond)
	queue.SetInstanceScheduler(concurrencyscheduler.NewScaledConcurrencyScheduler(&types.FunctionSpecification{
		FuncKey:          "testFunction",
		InstanceMetaData: commonTypes.InstanceMetaData{ConcurrentNum: 1},
	}, resspeckey.ResSpecKey{}, InsThdReqQueue))
	return queue
}

func TestRecoverInstance(t *testing.T) {
	reservedQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
		resspeckey.ResSpecKey{}: CreateScaledInstanceQueue(types.InstanceTypeReserved),
	}
	scaledQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
		resspeckey.ResSpecKey{CPU: 300, Memory: 128}: CreateScaledInstanceQueue(types.InstanceTypeScaled),
	}
	registry.GlobalRegistry = &registry.Registry{InstanceRegistry: registry.NewInstanceRegistry(make(chan struct{}))}
	gi := &GenericInstancePool{
		faasManagerInfo: faasManagerInfo{
			funcKey:    "faasManager-0123",
			instanceID: "01234567",
		},
		FuncSpec: &types.FunctionSpecification{
			FuncKey: "mock-funcKey-123",
		},
		reservedInstanceQueue: reservedQueue,
		scaledInstanceQueue:   scaledQueue,
		stateRoute:            StateRoute{stateRoute: map[string]*StateInstance{}, logger: log.GetLogger()},
	}
	funcSpec := &types.FunctionSpecification{
		FuncKey: "testFuncKey",
	}
	instancePoolState := &types.InstancePoolState{
		StateInstance: map[string]*types.Instance{
			"StateInstanceID123": &types.Instance{FuncSig: "22222"},
		},
	}
	var instanceRecord map[string]*types.Instance
	defer ApplyMethod(reflect.TypeOf(&instancequeue.ScaledInstanceQueue{}), "RecoverInstance",
		func(sq *instancequeue.ScaledInstanceQueue, instanceMap map[string]*types.Instance) {
			for k, v := range instanceMap {
				instanceRecord[k] = v
			}
		}).Reset()
	initRegistry()
	convey.Convey("recover success", t, func() {
		instanceRecord = make(map[string]*types.Instance)
		defer ApplyMethod(reflect.TypeOf(registry.GlobalRegistry.InstanceRegistry), "GetFunctionInstanceIDMap",
			func(ir *registry.InstanceRegistry) map[string]map[string]*commonTypes.InstanceSpecification {
				functionInstanceIDMap := make(map[string]map[string]*commonTypes.InstanceSpecification)
				functionInstanceIDMap["mock-funcKey-123"] = map[string]*commonTypes.InstanceSpecification{
					"ReservedInstanceID123": {CreateOptions: map[string]string{types.FunctionKeyNote: "22222"}},
					"ScaledInstanceID123":   {CreateOptions: map[string]string{types.FunctionKeyNote: "11111"}},
				}
				return functionInstanceIDMap
			}).Reset()
		var wg sync.WaitGroup
		wg.Add(1)
		gi.RecoverInstance(funcSpec, instancePoolState, false, &wg)
		wg.Wait()
		convey.So(gi.stateRoute.stateRoute["StateInstanceID123"], convey.ShouldNotBeNil)
	})
}

func Test_generateScheduleAffinity(t *testing.T) {
	type args struct {
		label string
	}
	tests := []struct {
		name string
		args args
		want []api.Affinity
	}{
		{
			name: "label1",
			args: args{"label1"},
			want: []api.Affinity{api.Affinity{
				PreferredPriority:        true,
				PreferredAntiOtherLabels: true,
				LabelOps: []api.LabelOperator{
					api.LabelOperator{
						Type:     2,
						LabelKey: "label1",
					},
				},
			}},
		},
		{
			name: "label1, unUseAntiOtherLabels",
			args: args{" label1,  unUseAntiOtherLabels"},
			want: []api.Affinity{api.Affinity{
				PreferredPriority:        true,
				PreferredAntiOtherLabels: false,
				LabelOps: []api.LabelOperator{
					api.LabelOperator{
						Type:     2,
						LabelKey: "label1",
					},
				},
			}},
		},
		{
			name: "label1,label2",
			args: args{label: " label1, label2 "},
			want: []api.Affinity{
				{
					PreferredPriority:        true,
					PreferredAntiOtherLabels: true,
					LabelOps: []api.LabelOperator{
						{
							Type:     2,
							LabelKey: "label1",
						},
					}},
				{
					PreferredPriority:        true,
					PreferredAntiOtherLabels: true,
					LabelOps: []api.LabelOperator{
						{
							Type:     2,
							LabelKey: "label2",
						},
					}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, generateScheduleAffinity([]api.Affinity{}, tt.args.label), "generateScheduleAffinity(%v)", tt.args.label)
		})
	}
}

func Test_filterStateInstanceMap(t *testing.T) {
	convey.Convey("filterStateInstanceMap", t, func() {
		convey.Convey("success", func() {
			stateMap := map[string]*types.Instance{
				"state-1": &types.Instance{
					ResKey: resspeckey.ResSpecKey{},
					InstanceStatus: commonTypes.InstanceStatus{
						Code: int32(constant.KernelInstanceStatusRunning),
					},
					InstanceType:  "",
					InstanceID:    "1",
					FuncKey:       "test-function",
					FuncSig:       "",
					ConcurrentNum: 2,
				},
				"state-2": &types.Instance{
					ResKey: resspeckey.ResSpecKey{},
					InstanceStatus: commonTypes.InstanceStatus{
						Code: int32(constant.KernelInstanceStatusExited),
					},
					InstanceType:  "",
					InstanceID:    "2",
					FuncKey:       "test-function",
					FuncSig:       "",
					ConcurrentNum: 2,
				},
			}
			etcdMap := map[string]map[string]*commonTypes.InstanceSpecification{}
			filterStateInstanceMap(stateMap, etcdMap, "test-function")
			convey.So(len(stateMap), convey.ShouldEqual, 2)
			convey.So(stateMap["state-2"].InstanceStatus.Code, convey.ShouldEqual, constant.KernelInstanceStatusExited)
		})
	})
}

func Test_makeRaspContainer(t *testing.T) {
	convey.Convey("test makeRaspContainer", t, func() {
		funcSpec := &types.FunctionSpecification{
			ExtendedMetaData: commonTypes.ExtendedMetaData{
				RaspConfig: commonTypes.RaspConfig{
					RaspImage: "test-image",
					Envs: []commonTypes.KV{
						{
							Name:  "1",
							Value: "11",
						}, {
							Name:  "2",
							Value: "22",
						},
					},
				},
			},
		}
		sideCardConfig := makeRaspContainer(funcSpec)
		convey.So(sideCardConfig.Image, convey.ShouldEqual, "test-image")
		env1 := false
		env2 := false

		for _, env := range sideCardConfig.Env {
			if env.Name == "1" && env.Value == "11" {
				env1 = true
			}
			if env.Name == "2" && env.Value == "22" {
				env2 = true
			}
		}
		convey.So(env1, convey.ShouldBeTrue)
		convey.So(env2, convey.ShouldBeTrue)
	})
}

func Test_initContainerAdd(t *testing.T) {
	convey.Convey("test initContainerAdd", t, func() {
		convey.Convey("add success", func() {
			funcSpec := &types.FunctionSpecification{
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					RaspConfig: commonTypes.RaspConfig{
						InitImage:      "test-initImage",
						RaspImage:      "test-image",
						RaspServerIP:   "127.0.0.1",
						RaspServerPort: "8080",
					},
				},
			}
			_, err := initContainerAdd(funcSpec)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("parse ip error ", func() {
			funcSpec := &types.FunctionSpecification{
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					RaspConfig: commonTypes.RaspConfig{
						InitImage:      "test-initImage",
						RaspImage:      "test-image",
						RaspServerIP:   "",
						RaspServerPort: "8080",
					},
				},
			}
			configData, err := initContainerAdd(funcSpec)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(configData), convey.ShouldEqual, "null")
		})
		convey.Convey("invalid port ", func() {
			funcSpec := &types.FunctionSpecification{
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					RaspConfig: commonTypes.RaspConfig{
						InitImage:      "test-initImage",
						RaspImage:      "test-image",
						RaspServerIP:   "127.0.0.1",
						RaspServerPort: "a",
					},
				},
			}
			configData, err := initContainerAdd(funcSpec)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(configData), convey.ShouldEqual, "null")
		})
		convey.Convey("json Marsha data error", func() {
			defer ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, fmt.Errorf("marshal error")
			}).Reset()
			funcSpec := &types.FunctionSpecification{
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					RaspConfig: commonTypes.RaspConfig{
						InitImage:      "test-initImage",
						RaspServerIP:   "127.0.0.1",
						RaspServerPort: "8080",
					},
				},
			}
			_, err := initContainerAdd(funcSpec)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func Test_SideCarAdd(t *testing.T) {
	convey.Convey("test SideCarAdd", t, func() {
		convey.Convey("add success", func() {
			funcSpec := &types.FunctionSpecification{
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					RaspConfig: commonTypes.RaspConfig{
						InitImage:      "test-initImage",
						RaspImage:      "test-image",
						RaspServerIP:   "127.0.0.1",
						RaspServerPort: "8080",
					},
					CustomFilebeatConfig: commonTypes.CustomFilebeatConfig{
						ImageAddress: "test-initImage",
					},
				},
			}
			_, err := sideCarAdd(funcSpec)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("json Marsha data error", func() {
			defer ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, fmt.Errorf("marshal error")
			}).Reset()
			funcSpec := &types.FunctionSpecification{
				ExtendedMetaData: commonTypes.ExtendedMetaData{
					RaspConfig: commonTypes.RaspConfig{
						InitImage:      "test-initImage",
						RaspImage:      "test-image",
						RaspServerIP:   "127.0.0.1",
						RaspServerPort: "8080",
					},
					CustomFilebeatConfig: commonTypes.CustomFilebeatConfig{
						ImageAddress: "test-initImage",
					},
				},
			}
			_, err := sideCarAdd(funcSpec)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestGenericInstancePool_handleManagedChange(t *testing.T) {
	convey.Convey("HandleSchedulerManaged", t, func() {
		defer ApplyFunc((*instancequeue.ScaledInstanceQueue).HandleInsConfigUpdate, func(_ *instancequeue.ScaledInstanceQueue,
			insConfig *instanceconfig.Configuration) {
			return
		}).Reset()
		defer ApplyFunc((*instancequeue.ScaledInstanceQueue).EnableInstanceScale, func(_ *instancequeue.ScaledInstanceQueue) {
		}).Reset()
		reservedQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
			resspeckey.ResSpecKey{}: CreateScaledInstanceQueue(types.InstanceTypeReserved),
		}
		scaledQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
			resspeckey.ResSpecKey{CPU: 300, Memory: 128}: CreateScaledInstanceQueue(types.InstanceTypeScaled),
		}
		gi := &GenericInstancePool{
			faasManagerInfo: faasManagerInfo{
				funcKey:    "faasManager-0123",
				instanceID: "01234567",
			},
			FuncSpec: &types.FunctionSpecification{
				FuncKey: "mock-funcKey-123",
			},
			reservedInstanceQueue: reservedQueue,
			scaledInstanceQueue:   scaledQueue,
			stateRoute:            StateRoute{stateRoute: map[string]*StateInstance{}, logger: log.GetLogger()},
		}
		gi.handleManagedChange()
	})
}

func TestGenericInstancePool_handleRatioChange(t *testing.T) {
	var patches []*Patches
	expectRatio := 0
	patches = append(patches, ApplyFunc(
		(*concurrencyscheduler.ScaledConcurrencyScheduler).ReassignInstanceWhenGray,
		func(s *concurrencyscheduler.ScaledConcurrencyScheduler, ratio int) {
			expectRatio = ratio
		},
	))
	defer func() {
		for _, p := range patches {
			p.Reset()
		}
	}()

	defer ApplyFunc((*instancequeue.ScaledInstanceQueue).HandleInsConfigUpdate, func(_ *instancequeue.ScaledInstanceQueue,
		insConfig *instanceconfig.Configuration) {
		return
	}).Reset()
	defer ApplyFunc((*instancequeue.ScaledInstanceQueue).EnableInstanceScale, func(_ *instancequeue.ScaledInstanceQueue) {
	}).Reset()
	queue := CreateScaledInstanceQueue(types.InstanceTypeReserved)
	reservedQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
		resspeckey.ResSpecKey{}: queue,
	}

	scaledQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
		resspeckey.ResSpecKey{CPU: 300, Memory: 128}: CreateScaledInstanceQueue(types.InstanceTypeScaled),
	}

	gi := &GenericInstancePool{
		faasManagerInfo: faasManagerInfo{
			funcKey:    "faasManager-0123",
			instanceID: "01234567",
		},
		FuncSpec: &types.FunctionSpecification{
			FuncKey: "mock-funcKey-123",
		},
		reservedInstanceQueue: reservedQueue,
		scaledInstanceQueue:   scaledQueue,
		stateRoute:            StateRoute{stateRoute: map[string]*StateInstance{}, logger: log.GetLogger()},
	}
	gi.handleRatioChange(50)
	assert.Equal(t, expectRatio, 50)
}

func TestGenericInstancePool_HandleInstanceConfigEvent(t *testing.T) {
	convey.Convey("HandleInstanceConfigEvent", t, func() {
		reservedQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
			resspeckey.ResSpecKey{CPU: 300, Memory: 128}:                        CreateScaledInstanceQueue(types.InstanceTypeReserved),
			resspeckey.ResSpecKey{CPU: 300, Memory: 128, InvokeLabel: "label1"}: CreateScaledInstanceQueue(types.InstanceTypeReserved),
		}
		scaledQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
			resspeckey.ResSpecKey{CPU: 300, Memory: 128}:                        CreateScaledInstanceQueue(types.InstanceTypeScaled),
			resspeckey.ResSpecKey{CPU: 300, Memory: 128, InvokeLabel: "label1"}: CreateScaledInstanceQueue(types.InstanceTypeScaled),
		}
		gi := &GenericInstancePool{
			defaultResSpec: &resspeckey.ResourceSpecification{CPU: 300, Memory: 128},
			faasManagerInfo: faasManagerInfo{
				funcKey:    "faasManager-0123",
				instanceID: "01234567",
			},
			FuncSpec: &types.FunctionSpecification{
				FuncKey: "mock-funcKey-123",
			},
			reservedInstanceQueue: reservedQueue,
			scaledInstanceQueue:   scaledQueue,
		}
		convey.Convey("delete insConfig without label", func() {
			defer ApplyFunc((*instancequeue.ScaledInstanceQueue).HandleInsConfigUpdate, func(_ *instancequeue.ScaledInstanceQueue,
				insConfig *instanceconfig.Configuration) {
			}).Reset()
			defer ApplyFunc((*instancequeue.ScaledInstanceQueue).EnableInstanceScale, func(_ *instancequeue.ScaledInstanceQueue) {
			}).Reset()
			callDestroy := 0
			defer ApplyFunc((*instancequeue.ScaledInstanceQueue).Destroy, func(_ *instancequeue.ScaledInstanceQueue) {
				callDestroy++
			}).Reset()
			gi.HandleInstanceConfigEvent(registry.SubEventTypeDelete, &instanceconfig.Configuration{})
			convey.So(callDestroy, convey.ShouldEqual, 0)
		})
	})
}

func TestGenericInstancePool_CleanOrphansInstanceQueue(t *testing.T) {
	callDestroy := 0
	defer ApplyFunc((*instancequeue.ScaledInstanceQueue).Destroy, func(_ *instancequeue.ScaledInstanceQueue) {
		callDestroy++
	}).Reset()
	reservedQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
		resspeckey.ResSpecKey{}: CreateScaledInstanceQueue(types.InstanceTypeReserved),
	}

	scaledQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
		resspeckey.ResSpecKey{CPU: 300, Memory: 128}: CreateScaledInstanceQueue(types.InstanceTypeScaled),
	}
	onDemandInstanceQueue := map[resspeckey.ResSpecKey]*instancequeue.OnDemandInstanceQueue{
		resspeckey.ResSpecKey{CPU: 300, Memory: 128}: &instancequeue.OnDemandInstanceQueue{},
	}
	gi := &GenericInstancePool{
		faasManagerInfo: faasManagerInfo{
			funcKey:    "faasManager-0123",
			instanceID: "01234567",
		},
		FuncSpec: &types.FunctionSpecification{
			FuncKey: "mock-funcKey-123",
		},
		reservedInstanceQueue: reservedQueue,
		scaledInstanceQueue:   scaledQueue,
		onDemandInstanceQueue: onDemandInstanceQueue,
		defaultPoolLabel:      "",
		stateRoute:            StateRoute{stateRoute: map[string]*StateInstance{}, logger: log.GetLogger()},
	}
	gi.CleanOrphansInstanceQueue()
	assert.Equal(t, 1, callDestroy)
}

func TestGenericInstancePool_sessionOperations(t *testing.T) {
	convey.Convey("test session operations", t, func() {
		funcCtx, funcCancel := context.WithCancel(context.TODO())
		gi := &GenericInstancePool{
			FuncSpec: &types.FunctionSpecification{
				FuncKey: "mock-funcKey-123",
				FuncCtx: funcCtx,
			},
			sessionRecordMap:      make(map[string]sessionRecord),
			instanceSessionMap:    make(map[string]map[string]struct{}),
			sessionReaperInterval: 100 * time.Millisecond,
		}
		convey.Convey("record and process", func() {
			ctx, cancel := context.WithCancel(context.TODO())
			insAlloc := &types.InstanceAllocation{
				Instance:    &types.Instance{InstanceID: "instance1"},
				SessionInfo: types.SessionInfo{SessionID: "session1", SessionCtx: ctx},
			}
			gi.recordInstanceSession(insAlloc)
			convey.So(gi.sessionRecordMap["session1"], convey.ShouldNotBeEmpty)
			convey.So(len(gi.instanceSessionMap["instance1"]), convey.ShouldEqual, 1)
			insAcqReq := &types.InstanceAcquireRequest{
				InstanceSession: commonTypes.InstanceSessionConfig{SessionID: "session1"},
			}
			gi.processInstanceSession(insAcqReq)
			convey.So(insAcqReq.DesignateInstanceID, convey.ShouldEqual, "instance1")
			cancel()
			gi.processInstanceSession(insAcqReq)
			convey.So(gi.sessionRecordMap, convey.ShouldBeEmpty)
			convey.So(len(gi.instanceSessionMap["instance1"]), convey.ShouldEqual, 0)
		})
		convey.Convey("record and clean", func() {
			insAlloc1 := &types.InstanceAllocation{
				Instance:    &types.Instance{InstanceID: "instance1"},
				SessionInfo: types.SessionInfo{SessionID: "session1"},
			}
			insAlloc2 := &types.InstanceAllocation{
				Instance:    &types.Instance{InstanceID: "instance1"},
				SessionInfo: types.SessionInfo{SessionID: "session2"},
			}
			gi.recordInstanceSession(insAlloc1)
			gi.recordInstanceSession(insAlloc2)
			gi.cleanInstanceSession("instance1")
			convey.So(len(gi.instanceSessionMap["instance1"]), convey.ShouldEqual, 0)
		})
		convey.Convey("reaper", func() {
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			insAlloc := &types.InstanceAllocation{
				Instance:    &types.Instance{InstanceID: "instance1"},
				SessionInfo: types.SessionInfo{SessionID: "session1", SessionCtx: ctx},
			}
			gi.recordInstanceSession(insAlloc)
			go gi.cleanInstanceSession("instance1")
			time.Sleep(200 * time.Millisecond)
			funcCancel()
			convey.So(gi.sessionRecordMap, convey.ShouldBeEmpty)
			convey.So(len(gi.instanceSessionMap["instance1"]), convey.ShouldEqual, 0)
		})
	})
}

func TestGenericInstancePool_judgeExceedInstance(t *testing.T) {
	funcCtx, _ := context.WithCancel(context.TODO())
	logger := log.GetLogger()
	resKey := resspeckey.ResSpecKey{
		CPU:                 500,
		Memory:              600,
		EphemeralStorage:    0,
		CustomResources:     "",
		CustomResourcesSpec: "",
		InvokeLabel:         "",
	}
	gi := &GenericInstancePool{
		FuncSpec: &types.FunctionSpecification{
			FuncKey: "mock-funcKey-123",
			FuncCtx: funcCtx,
		},
		defaultResKey: resKey,
		insConfig: map[resspeckey.ResSpecKey]*instanceconfig.Configuration{resKey: &instanceconfig.Configuration{
			FuncKey:       "mock-funcKey-123",
			InstanceLabel: "",
			InstanceMetaData: commonTypes.InstanceMetaData{
				MaxInstance: 1,
				MinInstance: 0,
			},
		}},
		scaledInstanceQueue: map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{resKey: {}},
	}
	defer ApplyFunc((*GenericInstancePool).getCurrentInstanceNum, func(_ *GenericInstancePool, resKey resspeckey.ResSpecKey) (int, int, snerror.SNError) {
		return 2, 2, nil
	}).Reset()
	i := 0
	defer ApplyFunc((*instancequeue.ScaledInstanceQueue).ScaleDownHandler, func(_ *instancequeue.ScaledInstanceQueue, insNum int, callback scaler.ScaleDownCallback) {
		i++
	}).Reset()
	gi.judgeExceedInstance(resKey, logger)
	assert.Equal(t, i, 1)
	gi.judgeExceedInstance(resspeckey.ResSpecKey{
		CPU:                 300,
		Memory:              128,
		EphemeralStorage:    0,
		CustomResources:     "",
		CustomResourcesSpec: "",
		InvokeLabel:         "",
	}, logger)
	assert.Equal(t, i, 1)
}

func TestGenericInstancePool_ResetInstanceScheduler(t *testing.T) {
	convey.Convey("Given a GenericInstancePool", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		funcSpec := &types.FunctionSpecification{InstanceMetaData: commonTypes.InstanceMetaData{SchedulePolicy: types.InstanceSchedulePolicyConcurrency}, FuncKey: "test-func"}
		pool0 := &GenericInstancePool{
			FuncSpec:       funcSpec,
			insAcqReqQueue: make(map[resspeckey.ResSpecKey]*requestqueue.InsAcqReqQueue),
		}
		resKey := resspeckey.ResSpecKey{}
		queue := CreateScaledInstanceQueue(types.InstanceTypeScaled)

		err := pool0.resetInstanceScheduler(queue, resKey)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGenericInstancePool_notifyInfo(t *testing.T) {
	convey.Convey("test updateInstanceAliasData", t, func() {
		p, _ := NewGenericInstancePool(&types.FunctionSpecification{}, faasManagerInfo{})

		convey.So(func() {
			gp, _ := p.(*GenericInstancePool)
			gp.scaledInstanceQueue[resspeckey.ResSpecKey{}] = CreateScaledInstanceQueue(types.InstanceTypeScaled)
			gp.reservedInstanceQueue[resspeckey.ResSpecKey{}] = CreateScaledInstanceQueue(types.InstanceTypeReserved)
			gp.updateInstanceAliasData()
			gp.HandleFaaSSchedulerEvent()
			gp.HandleAliasEvent(registry.SubEventTypeUpdate, "")
		}, convey.ShouldNotPanic)
	})
}

func TestGenericInstancePool_handleInstanceSynced(t *testing.T) {
	cnt := 0
	defer ApplyFunc((*instancequeue.ScaledInstanceQueue).HandleInstanceSync,
		func(_ *instancequeue.ScaledInstanceQueue, fn utils.RecoverSessionCallback) {
			cnt++
		}).Reset()

	reservedQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
		resspeckey.ResSpecKey{}: CreateScaledInstanceQueue(types.InstanceTypeReserved),
	}
	scaledQueue := map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
		resspeckey.ResSpecKey{CPU: 300, Memory: 128}: CreateScaledInstanceQueue(types.InstanceTypeScaled),
	}
	gi := &GenericInstancePool{
		faasManagerInfo: faasManagerInfo{
			funcKey:    "faasManager-0123",
			instanceID: "01234567",
		},
		FuncSpec: &types.FunctionSpecification{
			FuncKey: "mock-funcKey-123",
		},
		reservedInstanceQueue: reservedQueue,
		scaledInstanceQueue:   scaledQueue,
		sessionRecordMap:      map[string]sessionRecord{},
		instanceSessionMap:    make(map[string]map[string]struct{}),
	}
	gi.handleInstanceSynced()
	assert.Equal(t, 2, cnt)
}
