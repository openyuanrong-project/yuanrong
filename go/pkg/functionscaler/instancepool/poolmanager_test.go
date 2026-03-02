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
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	commonconstant "yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	commonstate "yuanrong.org/kernel/pkg/common/faas_common/state"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/instancequeue"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/stateinstance"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestNewPoolManager(t *testing.T) {
	stopCh := make(<-chan struct{})
	poolManager := NewPoolManager(stopCh)
	assert.Equal(t, stopCh, poolManager.stopCh)
}

func mockInsAcqReq() *types.InstanceAcquireRequest {
	insAcqReq := &types.InstanceAcquireRequest{}
	insAcqReq.FuncSpec = &types.FunctionSpecification{
		FuncKey: "testFunction",
		FuncCtx: context.TODO(),
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 500,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			ConcurrentNum: 100,
		},
	}
	return insAcqReq
}

func TestPoolManagerAcquireInstanceThread(t *testing.T) {
	initRegistry()
	convey.Convey("test PoolManagerAcquireInstanceThread", t, func() {
		convey.Convey("success", func() {
			patches := gomonkey.ApplyFunc((*GenericInstancePool).AcquireInstance, func(_ *GenericInstancePool,
				insThdApp *types.InstanceAcquireRequest) (*types.InstanceAllocation, snerror.SNError) {
				mockInstance := &types.Instance{
					ResKey: resspeckey.ResSpecKey{
						CPU:    1000,
						Memory: 1000,
					},
					InstanceType:  "",
					InstanceID:    "",
					FuncKey:       "",
					ConcurrentNum: 0,
				}
				instanceThread := &types.InstanceAllocation{
					Instance:     mockInstance,
					AllocationID: "mock-thread-id-123",
				}
				return instanceThread, nil
			})
			patches.ApplyFunc((*GenericInstancePool).ReleaseInstance, func(_ *GenericInstancePool,
				instance *types.InstanceAllocation) {
				return
			})
			defer patches.Reset()
			insAcqReq := mockInsAcqReq()
			poolManager := NewPoolManager(make(chan struct{}))
			poolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, &types.FunctionSpecification{
				FuncKey: "testFunction",
			})
			poolManager.instanceConfigRecord["testFunction"] = map[string]*instanceconfig.Configuration{
				DefaultInstanceLabel: {
					FuncKey:          "testFunction",
					InstanceMetaData: commonTypes.InstanceMetaData{},
				},
			}
			_, err := poolManager.AcquireInstanceThread(insAcqReq)
			assert.Equal(t, nil, err)
		})
		convey.Convey("state miss match", func() {
			insAcqReq := mockInsAcqReq()
			insAcqReq.StateID = "testStateID"
			poolManager := NewPoolManager(make(chan struct{}))
			poolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, &types.FunctionSpecification{
				FuncCtx: context.TODO(),
				FuncKey: "testFunction",
			})
			_, err := poolManager.AcquireInstanceThread(insAcqReq)
			convey.So(err.Code(), convey.ShouldEqual, statuscode.StateMismatch)
		})
		convey.Convey("newGenericInstancePool error", func() {
			patch := gomonkey.ApplyFunc(NewGenericInstancePool, func(funcSpec *types.FunctionSpecification,
				faasManagerInfo faasManagerInfo) (InstancePool, error) {
				return nil, errors.New("new pool error")
			})
			defer patch.Reset()
			insAcqReq := mockInsAcqReq()
			poolManager := NewPoolManager(make(chan struct{}))
			_, err := poolManager.AcquireInstanceThread(insAcqReq)
			convey.So(err.Code(), convey.ShouldEqual, statuscode.FuncMetaNotFoundErrCode)
		})
		convey.Convey("acquireInstanceThread error", func() {
			var mockInstancePool *GenericInstancePool
			patches := gomonkey.ApplyMethod(reflect.TypeOf(mockInstancePool), "AcquireInstance",
				func(_ *GenericInstancePool, insThdApp *types.InstanceAcquireRequest) (*types.InstanceAllocation,
					snerror.SNError) {
					return nil, snerror.New(statuscode.StatusInternalServerError, "acquire instance thread error")
				})
			defer patches.Reset()
			insAcqReq := mockInsAcqReq()
			poolManager := NewPoolManager(make(chan struct{}))
			poolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, &types.FunctionSpecification{
				FuncKey: "testFunction",
			})
			_, err := poolManager.AcquireInstanceThread(insAcqReq)
			convey.So(err.Code(), convey.ShouldEqual, statuscode.StatusInternalServerError)
		})
	})
}

func TestPoolManagerReleaseInstanceThread(t *testing.T) {
	initRegistry()
	var mockInstancePool *GenericInstancePool
	patches := gomonkey.ApplyMethod(reflect.TypeOf(mockInstancePool), "AcquireInstance",
		func(_ *GenericInstancePool, insThdApp *types.InstanceAcquireRequest) (*types.InstanceAllocation,
			snerror.SNError) {
			mockInstance := &types.Instance{
				ResKey: resspeckey.ResSpecKey{
					CPU:    1000,
					Memory: 1000,
				},
				InstanceType:  "",
				InstanceID:    "",
				FuncKey:       "",
				ConcurrentNum: 0,
			}
			instanceThread := &types.InstanceAllocation{
				Instance:     mockInstance,
				AllocationID: "mock-thread-id-123",
			}
			return instanceThread, nil
		})
	patches.ApplyMethod(reflect.TypeOf(mockInstancePool), "ReleaseInstance",
		func(_ *GenericInstancePool, instance *types.InstanceAllocation) {
			return
		})

	defer patches.Reset()
	poolManager := NewPoolManager(make(chan struct{}))
	poolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, &types.FunctionSpecification{
		FuncKey: "testFunction",
	})
	insAcqReq := mockInsAcqReq()
	poolManager.instanceConfigRecord["testFunction"] = map[string]*instanceconfig.Configuration{
		DefaultInstanceLabel: {
			FuncKey:          "testFunction",
			InstanceMetaData: commonTypes.InstanceMetaData{},
		},
	}
	insAlloc, err := poolManager.AcquireInstanceThread(insAcqReq)
	assert.Equal(t, nil, err)

	poolManager.ReleaseInstanceThread(insAlloc)
}

func TestPoolManagerReleaseAbnormalInstance(t *testing.T) {
	var mockInstancePool *GenericInstancePool
	patches := gomonkey.ApplyMethod(reflect.TypeOf(mockInstancePool), "AcquireInstance",
		func(_ *GenericInstancePool, insThdApp *types.InstanceAcquireRequest) (*types.InstanceAllocation,
			snerror.SNError) {
			mockInstance := &types.Instance{
				ResKey: resspeckey.ResSpecKey{
					CPU:    1000,
					Memory: 1000,
				},
				InstanceType:  "",
				InstanceID:    "",
				FuncKey:       "",
				ConcurrentNum: 0,
			}
			instanceThread := &types.InstanceAllocation{
				Instance:     mockInstance,
				AllocationID: "mock-thread-id-123",
			}
			return instanceThread, nil
		})
	patches.ApplyMethod(reflect.TypeOf(mockInstancePool), "HandleInstanceEvent",
		func(_ *GenericInstancePool, eventType registry.EventType, instance *types.Instance) {
			return
		})

	defer patches.Reset()
	initRegistry()
	poolManager := NewPoolManager(make(chan struct{}))
	poolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, &types.FunctionSpecification{
		FuncKey: "testFunction",
	})
	insAcqReq := mockInsAcqReq()
	poolManager.instanceConfigRecord["testFunction"] = map[string]*instanceconfig.Configuration{
		DefaultInstanceLabel: {
			FuncKey:          "testFunction",
			InstanceMetaData: commonTypes.InstanceMetaData{},
		},
	}
	insAlloc, err := poolManager.AcquireInstanceThread(insAcqReq)
	assert.Equal(t, nil, err)
	poolManager.ReleaseAbnormalInstance(insAlloc.Instance, log.GetLogger())
}

func TestHandleFunctionUpdate(t *testing.T) {
	var mockInstancePool *GenericInstancePool
	patches := gomonkey.ApplyMethod(reflect.TypeOf(mockInstancePool), "AcquireInstance",
		func(_ *GenericInstancePool, insThdApp *types.InstanceAcquireRequest) (*types.InstanceAllocation,
			snerror.SNError) {
			mockInstance := &types.Instance{
				ResKey: resspeckey.ResSpecKey{
					CPU:    1000,
					Memory: 1000,
				},
				InstanceType:  "",
				InstanceID:    "",
				FuncKey:       "",
				ConcurrentNum: 0,
			}
			instanceThread := &types.InstanceAllocation{
				Instance:     mockInstance,
				AllocationID: "mock-thread-id-123",
			}
			return instanceThread, nil
		})
	patches.ApplyMethod(reflect.TypeOf(mockInstancePool), "ReleaseInstance",
		func(_ *GenericInstancePool, instance *types.InstanceAllocation) {
			return
		})

	defer patches.Reset()
	initRegistry()
	poolManager := NewPoolManager(make(chan struct{}))
	poolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, &types.FunctionSpecification{
		FuncKey: "testFunction",
	})
	insAcqReq := mockInsAcqReq()
	poolManager.instanceConfigRecord["testFunction"] = map[string]*instanceconfig.Configuration{
		DefaultInstanceLabel: {
			FuncKey:          "testFunction",
			InstanceMetaData: commonTypes.InstanceMetaData{},
		},
	}
	_, err := poolManager.AcquireInstanceThread(insAcqReq)
	assert.Equal(t, err, nil)

	funcSpec1 := &types.FunctionSpecification{
		FuncKey: "testFunction123456",
		FuncCtx: context.TODO(),
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 700,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			ConcurrentNum: 100,
		},
	}
	poolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, funcSpec1)
	assert.Equal(t, funcSpec1, poolManager.instancePool[funcSpec1.FuncKey].(*GenericInstancePool).FuncSpec)
}

func TestHandleFunctionDelete(t *testing.T) {
	defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	var mockInstancePool *GenericInstancePool
	patches := gomonkey.ApplyMethod(reflect.TypeOf(mockInstancePool), "AcquireInstance",
		func(_ *GenericInstancePool, insThdApp *types.InstanceAcquireRequest) (*types.InstanceAllocation,
			snerror.SNError) {
			mockInstance := &types.Instance{
				ResKey: resspeckey.ResSpecKey{
					CPU:    1000,
					Memory: 1000,
				},
				InstanceType:  "scaled",
				InstanceID:    "test-instance",
				FuncKey:       "test-function",
				ConcurrentNum: 0,
			}
			instanceThread := &types.InstanceAllocation{
				Instance:     mockInstance,
				AllocationID: "mock-thread-id-123",
			}
			return instanceThread, nil
		})
	patches.ApplyMethod(reflect.TypeOf(mockInstancePool), "ReleaseInstance",
		func(_ *GenericInstancePool, instance *types.InstanceAllocation) {
			return
		})

	defer patches.Reset()
	initRegistry()
	poolManager := NewPoolManager(make(chan struct{}))
	insAcqReq := mockInsAcqReq()
	poolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, &types.FunctionSpecification{
		FuncKey: "testFunction",
	})
	poolManager.instanceConfigRecord["testFunction"] = map[string]*instanceconfig.Configuration{
		DefaultInstanceLabel: {
			FuncKey:          "testFunction",
			InstanceMetaData: commonTypes.InstanceMetaData{},
		},
	}
	_, err := poolManager.AcquireInstanceThread(insAcqReq)
	assert.Equal(t, err, nil)

	funcSpec1 := &types.FunctionSpecification{
		FuncKey: "testFunction",
		FuncCtx: context.TODO(),
		ResourceMetaData: commonTypes.ResourceMetaData{
			CPU:    500,
			Memory: 700,
		},
		InstanceMetaData: commonTypes.InstanceMetaData{
			ConcurrentNum: 100,
		},
	}
	assert.NotEqual(t, nil, poolManager.instancePool[funcSpec1.FuncKey])
	poolManager.HandleFunctionEvent(registry.SubEventTypeDelete, funcSpec1)
	assert.Equal(t, nil, poolManager.instancePool[funcSpec1.FuncKey])
}

func TestHandleInstanceEvent(t *testing.T) {
	patches := mockInstanceOperation()
	defer unMockInstanceOperation(patches)
	var mockInstancePool *GenericInstancePool = &GenericInstancePool{}

	pm := &PoolManager{
		instancePool: map[string]InstancePool{
			"instance1": mockInstancePool,
		},
	}
	convey.Convey("Func is error", t, func() {
		pm.HandleInstanceEvent(registry.SubEventTypeUpdate, &commonTypes.InstanceSpecification{
			Function: "123456/funcName",
		})
	})
	convey.Convey("update faasManager", t, func() {
		pm.HandleInstanceEvent(registry.SubEventTypeUpdate, &commonTypes.InstanceSpecification{
			Function:   "123456/0-system-faasmanager/$latest",
			InstanceID: "manager-faas-manager-5dcd75b58-74gt8",
		})
		convey.So(mockInstancePool.faasManagerInfo, convey.ShouldResemble, faasManagerInfo{
			funcKey:    "123456/0-system-faasmanager/$latest",
			instanceID: "manager-faas-manager-5dcd75b58-74gt8",
		})

		pm.HandleInstanceEvent(registry.SubEventTypeDelete, &commonTypes.InstanceSpecification{
			Function:   "123456/0-system-faasmanager/$latest",
			InstanceID: "manager-faas-manager-5dcd75b58-74gt9",
		})
		convey.So(mockInstancePool.faasManagerInfo, convey.ShouldResemble, faasManagerInfo{
			funcKey:    "123456/0-system-faasmanager/$latest",
			instanceID: "manager-faas-manager-5dcd75b58-74gt8",
		})

		pm.HandleInstanceEvent(registry.SubEventTypeUpdate, &commonTypes.InstanceSpecification{
			Function:   "123456/0-system-faasmanager/$latest",
			InstanceID: "manager-faas-manager-5dcd75b58-74gt9",
		})
		convey.So(mockInstancePool.faasManagerInfo, convey.ShouldResemble, faasManagerInfo{
			funcKey:    "123456/0-system-faasmanager/$latest",
			instanceID: "manager-faas-manager-5dcd75b58-74gt9",
		})

	})
	testFunc := &commonTypes.InstanceSpecification{
		Function:   "123456/0-system-testFunc/$latest",
		InstanceID: "testFunc123",
	}
	convey.Convey("instanceRecord not exist", t, func() {
		pm.HandleInstanceEvent(registry.SubEventTypeDelete, testFunc)
		convey.So(pm.instancePool, convey.ShouldNotContainKey, testFunc.InstanceID)
	})

	convey.Convey("delete instanceRecord", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&GenericInstancePool{}), "HandleInstanceEvent",
			func(_ *GenericInstancePool, eventType registry.EventType, instance *types.Instance) {
				return
			}).Reset()

		pm.HandleInstanceEvent(registry.SubEventTypeDelete, testFunc)
		convey.So(pm.instancePool, convey.ShouldNotContainKey, testFunc.InstanceID)
	})
	convey.Convey("delete leaseManager", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&GenericInstancePool{}), "HandleInstanceEvent",
			func(_ *GenericInstancePool, eventType registry.EventType, instance *types.Instance) {
				return
			}).Reset()
		pm.HandleInstanceEvent(registry.SubEventTypeDelete, testFunc)
		convey.So(pm.instancePool, convey.ShouldNotContainKey, testFunc.InstanceID)
	})

	convey.Convey("labels is not exist", t, func() {
		testFunc := &commonTypes.InstanceSpecification{
			Function:   "123456/0-system-testFunc/$latest",
			InstanceID: "testFunc123",
			Labels:     []string{"labels1", "labels2"},
		}
		pm := &PoolManager{
			instancePool: map[string]InstancePool{
				"instance1": mockInstancePool,
			},
		}
		pm.HandleInstanceEvent(registry.SubEventTypeDelete, testFunc)
		convey.So(pm.instancePool, convey.ShouldNotContainKey, testFunc.InstanceID)
	})

	convey.Convey("lease not exist", t, func() {
		mockInstancePool := &GenericInstancePool{
			FuncSpec: &types.FunctionSpecification{
				FuncKey: "funcKey",
			},
		}
		testFunc := &commonTypes.InstanceSpecification{
			Function:   "123456/0-system-testFunc/$latest",
			InstanceID: "testFunc123",
			Labels:     []string{"labels1", "instance1"},
		}
		pm := &PoolManager{
			instancePool: map[string]InstancePool{
				"instance1": mockInstancePool,
			},
		}
		pm.HandleInstanceEvent(registry.SubEventTypeDelete, testFunc)
		convey.So(pm.instancePool, convey.ShouldNotContainKey, testFunc.InstanceID)
	})

	convey.Convey("delete success", t, func() {
		mockInstancePool := &GenericInstancePool{
			FuncSpec: &types.FunctionSpecification{
				FuncKey: "funcKey",
			},
		}
		testFunc := &commonTypes.InstanceSpecification{
			Function:   "123456/0-system-testFunc/$latest",
			InstanceID: "testFunc123",
			Labels:     []string{"labels1", "instance1"},
		}
		pm := &PoolManager{
			instancePool: map[string]InstancePool{
				"instance1": mockInstancePool,
			},
		}
		pm.HandleInstanceEvent(registry.SubEventTypeDelete, testFunc)
		convey.So(pm.instancePool, convey.ShouldNotContainKey, testFunc.InstanceID)
	})

	convey.Convey("success HandleInstanceEvent", t, func() {
		insPool := &GenericInstancePool{
			FuncSpec: &types.FunctionSpecification{
				FuncKey: "123456/0-system-testFunc/$latest",
			},
			waitInsConfigChan: make(chan struct{}),
		}
		pm := &PoolManager{
			instancePool: map[string]InstancePool{
				"123456/0-system-testFunc/$latest": insPool,
			},
		}
		close(insPool.waitInsConfigChan)
		labels := map[string]string{podLabelInstanceType: string(types.InstanceTypeReserved)}
		b, _ := json.Marshal(labels)
		testInsSpec := &commonTypes.InstanceSpecification{
			Function:       "123456/0-system-testFunc/$latest",
			InstanceID:     "testIns123",
			Labels:         []string{"labels1", "instance1"},
			CreateOptions:  map[string]string{types.FunctionKeyNote: "123456/0-system-testFunc/$latest", commonconstant.DelegatePodLabels: string(b)},
			InstanceStatus: commonTypes.InstanceStatus{Code: int32(3)},
		}
		var getIns *types.Instance
		defer gomonkey.ApplyMethod(reflect.TypeOf(insPool), "HandleInstanceEvent",
			func(gi *GenericInstancePool, eventType registry.EventType, instance *types.Instance) {
				getIns = instance
			}).Reset()
		pm.HandleInstanceEvent(registry.SubEventTypeUpdate, testInsSpec)
		convey.So(getIns.InstanceID, convey.ShouldEqual, "testIns123")
	})
}

func TestReportMetrics(t *testing.T) {
	var mockInstancePool *GenericInstancePool
	patches := gomonkey.ApplyMethod(reflect.TypeOf(mockInstancePool), "AcquireInstance",
		func(_ *GenericInstancePool, insThdApp *types.InstanceAcquireRequest) (*types.InstanceAllocation,
			snerror.SNError) {
			mockInstance := &types.Instance{
				ResKey: resspeckey.ResSpecKey{
					CPU:    1000,
					Memory: 1000,
				},
				InstanceType:  "",
				InstanceID:    "",
				FuncKey:       "",
				ConcurrentNum: 0,
			}
			instanceThread := &types.InstanceAllocation{
				Instance:     mockInstance,
				AllocationID: "mock-thread-id-123",
			}
			return instanceThread, nil
		})
	patches.ApplyMethod(reflect.TypeOf(mockInstancePool), "ReleaseInstance",
		func(_ *GenericInstancePool, instance *types.InstanceAllocation) {
			return
		})

	defer patches.Reset()
	initRegistry()
	poolManager := NewPoolManager(make(chan struct{}))
	insAcqReq := mockInsAcqReq()
	poolManager.HandleFunctionEvent(registry.SubEventTypeUpdate, &types.FunctionSpecification{
		FuncKey: "testFunction",
	})
	poolManager.instanceConfigRecord["testFunction"] = map[string]*instanceconfig.Configuration{
		DefaultInstanceLabel: {
			FuncKey:          "testFunction",
			InstanceMetaData: commonTypes.InstanceMetaData{},
		},
	}
	_, err := poolManager.AcquireInstanceThread(insAcqReq)
	assert.Equal(t, err, nil)
	funcKey := insAcqReq.FuncSpec.FuncKey
	resKey := resspeckey.ResSpecKey{
		CPU:    10,
		Memory: 500,
	}
	insMetrics := &types.InstanceThreadMetrics{
		InsThdID:    "Instance-1",
		ProcNumPS:   0,
		ProcReqNum:  0,
		AvgProcTime: 0,
		MaxProcTime: 0,
	}

	poolManager.ReportMetrics(funcKey, resKey, insMetrics)

	// The instance pool does not exist
	delete(poolManager.instancePool, funcKey)
	poolManager.ReportMetrics(funcKey, resKey, insMetrics)
}

func Test_RecoverInstancePool(t *testing.T) {
	initRegistry()
	patches := []*gomonkey.Patches{
		gomonkey.ApplyMethod(reflect.TypeOf(&GenericInstancePool{}), "RecoverInstance",
			func(_ *GenericInstancePool, funcSpec *types.FunctionSpecification,
				instancePoolState *types.InstancePoolState, deleteFunc bool, wg *sync.WaitGroup) {
				wg.Done()
				return
			}),
		gomonkey.ApplyMethod(reflect.TypeOf(&registry.InstanceRegistry{}), "WaitForETCDList",
			func(ir *registry.InstanceRegistry) {
				return
			}),
		gomonkey.ApplyMethod(reflect.TypeOf(&commonstate.Queue{}), "SaveState",
			func(q *commonstate.Queue, state []byte, key string) error {
				return nil
			})}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	schedulerState := state.GetState()
	schedulerState.InstancePool["functionKey1"] = &types.InstancePoolState{
		StateInstance: map[string]*types.Instance{},
	}
	schedulerState.InstancePool["functionKey2"] = &types.InstancePoolState{
		StateInstance: map[string]*types.Instance{},
	}
	schedulerState.InstancePool["functionKey1"].StateInstance["InstanceID1"] = &types.Instance{FuncSig: "11111"}
	schedulerState.InstancePool["functionKey1"].StateInstance["InstanceID2"] = &types.Instance{FuncSig: "11111"}

	schedulerState.InstancePool["functionKey2"].StateInstance["InstanceID21"] = &types.Instance{FuncSig: "22222"}
	schedulerState.InstancePool["functionKey2"].StateInstance["InstanceID22"] = &types.Instance{FuncSig: "22222"}

	poolManager := NewPoolManager(make(chan struct{}))
	convey.Convey("RecoverInstancePool success", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&registry.Registry{}), "GetFuncSpec",
			func(_ *registry.Registry, funcKey string) *types.FunctionSpecification {
				return &types.FunctionSpecification{
					FuncKey: funcKey,
				}
			}).Reset()
		poolManager.RecoverInstancePool()
		_, exist := poolManager.instancePool["functionKey1"]
		convey.So(exist, convey.ShouldEqual, true)
		_, exist = poolManager.instancePool["functionKey2"]
		convey.So(exist, convey.ShouldEqual, true)
	})
	convey.Convey("failed to GetFuncSpec", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&registry.Registry{}), "GetFuncSpec",
			func(_ *registry.Registry, funcKey string) *types.FunctionSpecification {
				if funcKey == "functionKey1" {
					return nil
				}
				return &types.FunctionSpecification{
					FuncKey: funcKey,
				}
			}).Reset()
		poolManager.RecoverInstancePool()
		_, exist := poolManager.instancePool["functionKey1"]
		convey.So(exist, convey.ShouldEqual, false)
		_, exist = poolManager.instancePool["functionKey2"]
		convey.So(exist, convey.ShouldEqual, true)
	})
}

func TestHandleInstanceConfigEvent(t *testing.T) {
	var mockInstancePool *GenericInstancePool = &GenericInstancePool{}

	pm := &PoolManager{
		instancePool: map[string]InstancePool{
			"instance1": mockInstancePool,
		},
		instanceConfigRecord: map[string]map[string]*instanceconfig.Configuration{},
	}
	insConfig := &instanceconfig.Configuration{
		FuncKey: "testFunc",
	}
	convey.Convey("instance pool not exist", t, func() {
		pm.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, insConfig)
	})
	pm.HandleFunctionEvent(registry.SubEventTypeUpdate, &types.FunctionSpecification{
		FuncKey: "testFunc",
	})
	convey.Convey("instance pool exist", t, func() {
		pm.HandleInstanceConfigEvent(registry.SubEventTypeUpdate, insConfig)
	})
	convey.Convey("instance pool exist", t, func() {
		pm.HandleInstanceConfigEvent(registry.SubEventTypeDelete, insConfig)
	})
}

func TestHandleAliasEvent(t *testing.T) {
	convey.Convey("Test HandleAliasEvent", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(&GenericInstancePool{}), "HandleAliasEvent",
			func(_ *GenericInstancePool, eventType registry.EventType, aliasUrn string) {
			}).Reset()
		mockInstancePool := &GenericInstancePool{
			reservedInstanceQueue: map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{resspeckey.ResSpecKey{}: {}},
			FuncSpec:              &types.FunctionSpecification{FuncKey: "TenantID/FunctionName/Version"},
		}
		pm := &PoolManager{
			instancePool: map[string]InstancePool{
				"TenantID/FunctionName/Version": mockInstancePool,
			},
			instanceConfigRecord: map[string]map[string]*instanceconfig.Configuration{},
		}
		pm.HandleAliasEvent(registry.SubEventTypeUpdate, "sn:cn:yrk:TenantID:function:FunctionName:Version")
	})
}

func TestGetInstanceType(t *testing.T) {
	convey.Convey("Test getInstanceType", t, func() {
		convey.Convey("DelegatePodLabels is not exist", func() {
			createOptions := map[string]string{}
			instanceType := getInstanceType(createOptions)
			convey.So(instanceType, convey.ShouldEqual, types.InstanceTypeUnknown)
		})
		convey.Convey("unmarshal labels error", func() {
			createOptions := map[string]string{commonconstant.DelegatePodLabels: ""}
			instanceType := getInstanceType(createOptions)
			convey.So(instanceType, convey.ShouldEqual, types.InstanceTypeUnknown)
		})
		convey.Convey("podLabels is nil", func() {
			podLabels := map[string]string{}
			data, _ := json.Marshal(podLabels)
			createOptions := map[string]string{commonconstant.DelegatePodLabels: string(data)}
			instanceType := getInstanceType(createOptions)
			convey.So(instanceType, convey.ShouldEqual, types.InstanceTypeUnknown)
		})
		convey.Convey("podLabelInstanceType is not exist", func() {
			podLabels := map[string]string{"a": "b"}
			data, _ := json.Marshal(podLabels)
			createOptions := map[string]string{commonconstant.DelegatePodLabels: string(data)}
			instanceType := getInstanceType(createOptions)
			convey.So(instanceType, convey.ShouldEqual, types.InstanceTypeUnknown)
		})
		convey.Convey("get instanceType success", func() {
			podLabels := map[string]string{podLabelInstanceType: "reserved"}
			data, _ := json.Marshal(podLabels)
			createOptions := map[string]string{commonconstant.DelegatePodLabels: string(data)}
			instanceType := getInstanceType(createOptions)
			convey.So(instanceType, convey.ShouldEqual, types.InstanceTypeUnknown)
		})
	})
}

func TestCheckMinInsAndReport(t *testing.T) {
	convey.Convey("Test CheckMinInsAndReport", t, func() {
		rawGConfig := config.GlobalConfig
		config.GlobalConfig = types.Configuration{}
		stopCh := make(chan struct{})
		pm := &PoolManager{}
		pm.CheckMinInsAndReport(stopCh)
		time.Sleep(10 * time.Millisecond)
		close(stopCh)
		config.GlobalConfig = rawGConfig
		convey.So(pm, convey.ShouldNotBeNil)
	})
}

func TestJudgeAndReport(t *testing.T) {
	defer gomonkey.ApplyMethod(reflect.TypeOf(&instancequeue.ScaledInstanceQueue{}), "GetInstanceNumber",
		func(_ *instancequeue.ScaledInstanceQueue) int {
			return 0
		}).Reset()
	convey.Convey("Test JudgeAndReport", t, func() {
		convey.Convey("report success", func() {
			mockInstancePool := &GenericInstancePool{
				insConfig: map[resspeckey.ResSpecKey]*instanceconfig.Configuration{
					resspeckey.ResSpecKey{}: {
						InstanceMetaData: commonTypes.InstanceMetaData{
							MinInstance: 1,
						},
					},
				},
				reservedInstanceQueue: map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{
					resspeckey.ResSpecKey{}: &instancequeue.ScaledInstanceQueue{},
				},
				FuncSpec: &types.FunctionSpecification{
					FuncKey: "testFunc",
				},
				minScaleAlarmSign:  map[string]bool{},
				pendingInstanceNum: map[string]int{},
			}
			pm := &PoolManager{
				instancePool: map[string]InstancePool{
					"TenantID/FunctionName/Version": mockInstancePool,
				},
			}
			pm.judgeAndReport("clusterID")
			convey.So(mockInstancePool.minScaleAlarmSign[DefaultInstanceLabel], convey.ShouldBeTrue)
		})
		convey.Convey("minInstance is zero", func() {
			mockInstancePool := &GenericInstancePool{
				insConfig: map[resspeckey.ResSpecKey]*instanceconfig.Configuration{
					resspeckey.ResSpecKey{}: {
						InstanceMetaData: commonTypes.InstanceMetaData{},
					},
				},
				minScaleUpdatedTime:   time.Now(),
				reservedInstanceQueue: map[resspeckey.ResSpecKey]*instancequeue.ScaledInstanceQueue{resspeckey.ResSpecKey{}: {}},
				FuncSpec: &types.FunctionSpecification{
					FuncKey: "testFunc",
				},
				minScaleAlarmSign:  map[string]bool{},
				pendingInstanceNum: map[string]int{},
			}
			pm := &PoolManager{
				instancePool: map[string]InstancePool{
					"TenantID/FunctionName/Version": mockInstancePool,
				},
			}
			pm.judgeAndReport("clusterID")
			convey.So(mockInstancePool.minScaleAlarmSign[DefaultInstanceLabel], convey.ShouldBeFalse)
		})
		convey.Convey("reservedInstanceQueue is nill", func() {
			mockInstancePool := &GenericInstancePool{
				insConfig: map[resspeckey.ResSpecKey]*instanceconfig.Configuration{
					resspeckey.ResSpecKey{}: {
						InstanceMetaData: commonTypes.InstanceMetaData{},
					},
				},
				FuncSpec: &types.FunctionSpecification{
					FuncKey: "testFunc",
				},
			}
			pm := &PoolManager{
				instancePool: map[string]InstancePool{
					"TenantID/FunctionName/Version": mockInstancePool,
				},
				instanceConfigRecord: map[string]map[string]*instanceconfig.Configuration{},
			}
			pm.judgeAndReport("clusterID")
			convey.So(mockInstancePool.minScaleAlarmSign[DefaultInstanceLabel], convey.ShouldBeFalse)
		})
	})
}

func TestGetAndDeleteState(t *testing.T) {
	convey.Convey("Test GetAndDeleteState", t, func() {
		convey.Convey("create instance pool error", func() {
			patch := gomonkey.ApplyFunc(NewGenericInstancePool,
				func(funcSpec *types.FunctionSpecification,
					faasManagerInfo faasManagerInfo) (InstancePool, error) {
					return &GenericInstancePool{}, errors.New("create instance pool error")
				})
			defer patch.Reset()
			mockInstancePool := &GenericInstancePool{}
			pm := &PoolManager{
				instancePool: map[string]InstancePool{
					"TenantID/FunctionName/Version": mockInstancePool,
				},
			}
			res := pm.GetAndDeleteState("stateID", "testFunc",
				&types.FunctionSpecification{}, log.GetLogger())
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("get state id success", func() {
			patch := gomonkey.ApplyFunc(NewGenericInstancePool,
				func(funcSpec *types.FunctionSpecification,
					faasManagerInfo faasManagerInfo) (InstancePool, error) {
					return &GenericInstancePool{}, nil
				})
			defer patch.Reset()
			patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&StateRoute{}),
				"GetAndDeleteState", func(s *StateRoute, stateID string) bool {
					return true
				})
			defer patch1.Reset()
			mockInstancePool := &GenericInstancePool{}
			pm := &PoolManager{
				instancePool: map[string]InstancePool{
					"TenantID/FunctionName/Version": mockInstancePool,
				},
			}
			res := pm.GetAndDeleteState("stateID", "testFunc",
				&types.FunctionSpecification{}, log.GetLogger())
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

func TestReleaseStateThread(t *testing.T) {
	convey.Convey("Test ReleaseStateThread", t, func() {
		convey.Convey("leaser is nil", func() {
			pm := &PoolManager{
				stateLeaseManager: make(map[string]*stateinstance.Leaser),
			}
			thread := &types.InstanceAllocation{
				Instance: &types.Instance{
					InstanceID: "testID",
				},
			}
			err := pm.ReleaseStateThread(thread)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("thread id invalid", func() {
			pm := &PoolManager{
				stateLeaseManager: map[string]*stateinstance.Leaser{"testID": {}},
			}
			thread := &types.InstanceAllocation{
				Instance: &types.Instance{
					InstanceID: "testID",
				},
				AllocationID: "InsAlloc",
			}
			err := pm.ReleaseStateThread(thread)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("thread id error", func() {
			pm := &PoolManager{
				stateLeaseManager: map[string]*stateinstance.Leaser{"testID": {}},
			}
			thread := &types.InstanceAllocation{
				Instance: &types.Instance{
					InstanceID: "testID",
				},
				AllocationID: "InsAlloc-stateThreada",
			}
			err := pm.ReleaseStateThread(thread)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("release state thread success", func() {
			pm := &PoolManager{
				stateLeaseManager: map[string]*stateinstance.Leaser{
					"testID": stateinstance.NewLeaser(1, func(stateID string, instanceID string) {},
						"stateID", "instanceID", 1*time.Second),
				},
			}
			thread := &types.InstanceAllocation{
				Instance: &types.Instance{
					InstanceID: "testID",
				},
				AllocationID: "InsAlloc-stateThread1",
			}
			err := pm.ReleaseStateThread(thread)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestRetainStateThread(t *testing.T) {
	convey.Convey("Test RetainStateThread", t, func() {
		convey.Convey("leaser is nil", func() {
			pm := &PoolManager{
				stateLeaseManager: make(map[string]*stateinstance.Leaser),
			}
			thread := &types.InstanceAllocation{
				Instance: &types.Instance{
					InstanceID: "testID",
				},
			}
			err := pm.RetainStateThread(thread)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("thread id invalid", func() {
			pm := &PoolManager{
				stateLeaseManager: map[string]*stateinstance.Leaser{"testID": {}},
			}
			thread := &types.InstanceAllocation{
				Instance: &types.Instance{
					InstanceID: "testID",
				},
				AllocationID: "InsAlloc",
			}
			err := pm.RetainStateThread(thread)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("thread id error", func() {
			pm := &PoolManager{
				stateLeaseManager: map[string]*stateinstance.Leaser{"testID": {}},
			}
			thread := &types.InstanceAllocation{
				Instance: &types.Instance{
					InstanceID: "testID",
				},
				AllocationID: "InsAlloc-stateThreada",
			}
			err := pm.RetainStateThread(thread)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("lease not found", func() {
			pm := &PoolManager{
				stateLeaseManager: map[string]*stateinstance.Leaser{
					"testID": stateinstance.NewLeaser(1, func(stateID string, instanceID string) {},
						"stateID", "instanceID", 1*time.Second),
				},
			}
			thread := &types.InstanceAllocation{
				Instance: &types.Instance{
					InstanceID: "testID",
				},
				AllocationID: "InsAlloc-stateThread1",
			}
			err := pm.RetainStateThread(thread)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestRecoverStateLeaser(t *testing.T) {
	convey.Convey("Test RetainStateThread", t, func() {
		convey.Convey("pool is nil", func() {
			mockInstancePool := &GenericInstancePool{}
			stateInstance := map[string]*types.Instance{
				"state1": {
					InstanceStatus: commonTypes.InstanceStatus{Code: int32(-1)},
				},
				"state2": {
					InstanceStatus: commonTypes.InstanceStatus{Code: int32(0)},
				},
			}
			pm := &PoolManager{
				instancePool: map[string]InstancePool{
					"testFunc": mockInstancePool,
				},
			}
			funcSpec := &types.FunctionSpecification{
				FuncKey: "testFunc1",
			}
			pm.recoverStateLeaser(stateInstance, funcSpec)
			convey.So(len(pm.stateLeaseManager), convey.ShouldEqual, 0)
		})
		convey.Convey("recover success", func() {
			mockInstancePool := &GenericInstancePool{}
			stateInstance := map[string]*types.Instance{
				"state1": {
					InstanceStatus: commonTypes.InstanceStatus{Code: int32(-1)},
					InstanceID:     "id1",
				},
				"state2": {
					InstanceStatus: commonTypes.InstanceStatus{Code: int32(0)},
					InstanceID:     "id2",
				},
			}
			pm := &PoolManager{
				instancePool: map[string]InstancePool{
					"testFunc": mockInstancePool,
				},
				stateLeaseManager: make(map[string]*stateinstance.Leaser),
			}
			funcSpec := &types.FunctionSpecification{
				FuncKey: "testFunc",
				InstanceMetaData: commonTypes.InstanceMetaData{
					ConcurrentNum: 1,
				},
			}
			pm.recoverStateLeaser(stateInstance, funcSpec)
			convey.So(len(pm.stateLeaseManager), convey.ShouldNotEqual, 0)
		})
	})
}

func TestHandleSchedulerManaged(t *testing.T) {
	convey.Convey("HandleSchedulerManaged", t, func() {
		pm := &PoolManager{
			stateLeaseManager: map[string]*stateinstance.Leaser{
				"testID": stateinstance.NewLeaser(1, func(stateID string, instanceID string) {},
					"stateID", "instanceID", 1*time.Second),
			},
			instancePool: map[string]InstancePool{},
		}
		pm.HandleSchedulerManaged(registry.SubEventTypeUpdate, &commonTypes.InstanceSpecification{})
	})
}

func TestHandleRolloutRatioChange(t *testing.T) {
	var patches []*gomonkey.Patches
	expectRatio := 0
	patches = append(patches, gomonkey.ApplyFunc(
		(*PoolManager).HandleRolloutRatioChange,
		func(pm *PoolManager, ratio int) {
			expectRatio = ratio
		},
	))
	defer func() {
		for _, p := range patches {
			p.Reset()
		}
	}()

	var mockInstancePool = &GenericInstancePool{}
	pm := &PoolManager{
		stateLeaseManager: map[string]*stateinstance.Leaser{
			"testID": stateinstance.NewLeaser(1, func(stateID string, instanceID string) {},
				"stateID", "instanceID", 1*time.Second),
		},
		instancePool: map[string]InstancePool{
			"instance1": mockInstancePool,
		},
	}
	pm.HandleRolloutRatioChange(50)
	assert.Equal(t, expectRatio, 50)
}

func TestPoolManagerCreateInstance(t *testing.T) {
	tests := []struct {
		name         string
		setupPool    bool
		createReq    *types.InstanceCreateRequest
		mockInstance *types.Instance
		mockError    snerror.SNError
		expectError  bool
	}{
		{
			name:      "pool_exists_success",
			setupPool: true,
			createReq: &types.InstanceCreateRequest{
				FuncSpec: &types.FunctionSpecification{
					FuncKey: "test-func",
				},
			},
			mockInstance: &types.Instance{
				InstanceID: "test-instance",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:      "pool_not_exist",
			setupPool: false,
			createReq: &types.InstanceCreateRequest{
				FuncSpec: &types.FunctionSpecification{
					FuncKey: "test-func",
				},
			},
			mockInstance: nil,
			mockError:    nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := &PoolManager{
				instancePool: make(map[string]InstancePool),
			}

			if tt.setupPool {
				pool := &GenericInstancePool{}
				pm.instancePool[tt.createReq.FuncSpec.FuncKey] = pool

				patches := gomonkey.NewPatches()
				defer patches.Reset()

				patches.ApplyMethod(reflect.TypeOf(pool), "CreateInstance",
					func(_ *GenericInstancePool, req *types.InstanceCreateRequest) (*types.Instance, snerror.SNError) {
						assert.Equal(t, tt.createReq, req)
						return tt.mockInstance, tt.mockError
					})
			}

			_, err := pm.CreateInstance(tt.createReq)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestPoolManagerDeleteInstance(t *testing.T) {
	tests := []struct {
		name        string
		setupPool   bool
		instance    *types.Instance
		mockError   snerror.SNError
		expectError bool
	}{
		{
			name:        "Pool does not exist",
			setupPool:   false,
			instance:    &types.Instance{FuncKey: "test-func-key"},
			expectError: true,
		},
		{
			name:        "Pool exists, delete succeeds",
			setupPool:   true,
			instance:    &types.Instance{FuncKey: "test-func-key"},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Pool exists, delete fails",
			setupPool:   true,
			instance:    &types.Instance{FuncKey: "test-func-key"},
			mockError:   snerror.New(statuscode.InternalErrorCode, "delete error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := &PoolManager{
				instancePool: make(map[string]InstancePool),
			}

			if tt.setupPool {
				pool := &GenericInstancePool{}
				pm.instancePool[tt.instance.FuncKey] = pool

				patches := gomonkey.NewPatches()
				defer patches.Reset()

				patches.ApplyMethod(reflect.TypeOf(pool), "DeleteInstance",
					func(_ *GenericInstancePool, instance *types.Instance) snerror.SNError {
						assert.Equal(t, tt.instance, instance)
						return tt.mockError
					})
			}

			err := pm.DeleteInstance(tt.instance)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
