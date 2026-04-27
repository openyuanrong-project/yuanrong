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

package instancepool

import (
	"reflect"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/stateinstance"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

func mockStateRoute() *StateRoute {
	stateRoute := &StateRoute{
		stateRoute: make(map[string]*StateInstance),
		stateConfig: commonTypes.StateConfig{
			LifeCycle: types.InstanceLifeCycleConsistentWithState,
		},
		funcSpec: &types.FunctionSpecification{
			InstanceMetaData: commonTypes.InstanceMetaData{
				ConcurrentNum: 100,
			},
		},
		resSpec: &resspeckey.ResourceSpecification{
			CPU:    100,
			Memory: 100,
		},
		deleteInstanceFunc: func(instance *types.Instance) error { return nil },
		createInstanceFunc: func(resSpec *resspeckey.ResourceSpecification, instanceType types.InstanceType,
			callerPodName string) (*types.Instance, error) {
			return nil, nil
		},
		stateLeaseManager: make(map[string]*stateinstance.Leaser, utils.DefaultMapSize),
		logger:            log.NewConsoleLogger(),
		RWMutex:           sync.RWMutex{},
		stateLocks:        sync.Map{},
	}
	return stateRoute
}

func TestNewStateRoute(t *testing.T) {
	convey.Convey("testNewStateRoute", t, func() {
		convey.So(mockStateRoute(), convey.ShouldNotBeNil)
	})
}

func TestStateRoute_Destroy(t *testing.T) {
	convey.Convey("testDestroy", t, func() {
		stateRoute := mockStateRoute()
		stateRoute.stateRoute["1"] = &StateInstance{
			StateID:  "1",
			status:   0,
			instance: &types.Instance{},
		}
		stateRoute.stateRoute["2"] = &StateInstance{
			StateID:  "2",
			status:   0,
			instance: &types.Instance{},
		}
		stateRoute.Destroy()
		convey.So(len(stateRoute.stateRoute), convey.ShouldEqual, 0)
	})
}

func TestStateRoute_HandleInstanceUpdate(t *testing.T) {
	convey.Convey("TestStateRoute_HandleInstanceUpdate", t, func() {
		stateRoute := mockStateRoute()
		var targetStateInstance *StateInstance
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(stateRoute), "processStateRoute",
			func(_ *StateRoute, instance *StateInstance, stateID string, opType string) {
				targetStateInstance = instance
			})
		defer patch.Reset()
		testInstance := &types.Instance{
			ResKey:         resspeckey.ResSpecKey{},
			InstanceStatus: commonTypes.InstanceStatus{},
			InstanceType:   types.InstanceTypeReserved,
			InstanceID:     "",
			FuncKey:        "",
			FuncSig:        "",
			ConcurrentNum:  0,
		}
		stateRoute.HandleInstanceUpdate(testInstance)
		convey.So(targetStateInstance, convey.ShouldBeNil)

		testInstance.InstanceType = types.InstanceTypeState
		testInstance.InstanceID = "1"
		testInstance.InstanceStatus = commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
			Msg:  "",
		}
		stateRoute.stateRoute["1"] = &StateInstance{
			StateID:  "1",
			status:   InstanceOk,
			instance: testInstance,
		}
		testInstance.InstanceStatus.Code = int32(constant.KernelInstanceStatusSubHealth)
		stateRoute.HandleInstanceUpdate(testInstance)
		convey.So(targetStateInstance, convey.ShouldNotBeNil)
		convey.So(targetStateInstance.status, convey.ShouldEqual, InstanceAbnormal)
	})
}

func TestStateRoute_GetAndDeleteState(t *testing.T) {
	convey.Convey("TestStateRoute_GetAndDeleteState", t, func() {
		stateRoute := mockStateRoute()
		var targetStateInstance *StateInstance
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(stateRoute), "processStateRoute",
			func(_ *StateRoute, instance *StateInstance, stateID string, opType string) {
				if opType == stateDelete {
					delete(stateRoute.stateRoute, stateID)
				}
			})
		defer patch.Reset()
		testInstance := &types.Instance{
			ResKey:         resspeckey.ResSpecKey{},
			InstanceStatus: commonTypes.InstanceStatus{},
			InstanceType:   types.InstanceTypeReserved,
			InstanceID:     "",
			FuncKey:        "",
			FuncSig:        "",
			ConcurrentNum:  0,
		}
		stateRoute.HandleInstanceUpdate(testInstance)
		convey.So(targetStateInstance, convey.ShouldBeNil)

		testInstance.InstanceType = types.InstanceTypeState
		testInstance.InstanceID = "1"
		testInstance.InstanceStatus = commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
			Msg:  "",
		}
		stateRoute.stateRoute["1"] = &StateInstance{
			StateID:  "1",
			status:   InstanceOk,
			instance: testInstance,
		}
		exist := stateRoute.GetAndDeleteState("1")
		convey.So(exist, convey.ShouldEqual, true)
		_, exist = stateRoute.stateRoute["1"]
		convey.So(exist, convey.ShouldEqual, false)
	})
}

func TestStateRoute_DeleteStateInstance(t *testing.T) {
	convey.Convey("TestStateRoute_DeleteStateInstance", t, func() {
		stateRoute := mockStateRoute()
		patch := gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {})
		defer patch.Reset()
		testInstance := &types.Instance{
			ResKey:         resspeckey.ResSpecKey{},
			InstanceStatus: commonTypes.InstanceStatus{},
			InstanceType:   types.InstanceTypeReserved,
			InstanceID:     "1111",
			FuncKey:        "",
			FuncSig:        "",
			ConcurrentNum:  0,
		}

		testInstance.InstanceType = types.InstanceTypeState
		testInstance.InstanceID = "1111"
		testInstance.InstanceStatus = commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
			Msg:  "",
		}
		stateRoute.stateRoute["1111"] = &StateInstance{
			StateID:  "1111",
			status:   InstanceOk,
			instance: testInstance,
		}

		stateRoute.DeleteStateInstance("1111", "1111")
		stateInstance, ok := stateRoute.stateRoute["1111"]
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(stateInstance.status, convey.ShouldEqual, InstanceExit)
	})
}

func TestStateRoute_DeleteStateInstanceByInstanceID(t *testing.T) {
	convey.Convey("TestStateRoute_DeleteStateInstanceByInstanceID", t, func() {
		stateRoute := mockStateRoute()
		patch := gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {})
		defer patch.Reset()
		testInstance := &types.Instance{
			ResKey:         resspeckey.ResSpecKey{},
			InstanceStatus: commonTypes.InstanceStatus{},
			InstanceType:   types.InstanceTypeState,
			InstanceID:     "3",
			FuncKey:        "",
			FuncSig:        "",
			ConcurrentNum:  0,
		}

		testInstance.InstanceType = types.InstanceTypeState
		testInstance.InstanceID = "3"
		testInstance.InstanceStatus = commonTypes.InstanceStatus{
			Code: int32(constant.KernelInstanceStatusRunning),
			Msg:  "",
		}
		stateRoute.stateRoute["3"] = &StateInstance{
			StateID:  "3",
			status:   InstanceOk,
			instance: testInstance,
		}

		stateRoute.DeleteStateInstanceByInstanceID("3")
		stateInstance, ok := stateRoute.stateRoute["3"]
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(stateInstance.status, convey.ShouldEqual, InstanceExit)
	})
}

func TestStateRoute_Recover(t *testing.T) {
	convey.Convey("TestStateRoute_recover", t, func() {
		instanceMap := map[string]*types.Instance{
			"1": &types.Instance{
				ResKey: resspeckey.ResSpecKey{},
				InstanceStatus: commonTypes.InstanceStatus{
					Code: int32(constant.KernelInstanceStatusRunning),
				},
				InstanceType:  "",
				InstanceID:    "1",
				FuncKey:       "",
				FuncSig:       "sig",
				ConcurrentNum: 0,
			},
			"2": &types.Instance{
				ResKey: resspeckey.ResSpecKey{},
				InstanceStatus: commonTypes.InstanceStatus{
					Code: int32(constant.KernelInstanceStatusSubHealth),
				},
				InstanceType:  "",
				InstanceID:    "2",
				FuncKey:       "",
				FuncSig:       "sig",
				ConcurrentNum: 0,
			},
			"3": &types.Instance{
				ResKey: resspeckey.ResSpecKey{},
				InstanceStatus: commonTypes.InstanceStatus{
					Code: int32(constant.KernelInstanceStatusExited),
				},
				InstanceType:  "",
				InstanceID:    "3",
				FuncKey:       "",
				FuncSig:       "sig",
				ConcurrentNum: 0,
			},
		}
		stateRoute := mockStateRoute()
		stateRoute.recover(instanceMap)
		stateInstance, ok := stateRoute.stateRoute["1"]
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(stateInstance.status, convey.ShouldEqual, InstanceOk)
		stateInstance, ok = stateRoute.stateRoute["2"]
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(stateInstance.status, convey.ShouldEqual, InstanceAbnormal)
		stateInstance, ok = stateRoute.stateRoute["3"]
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(stateInstance.status, convey.ShouldEqual, InstanceExit)
	})
}

func TestStateRoute_acquireStateInstanceThread(t *testing.T) {
	convey.Convey("TestStateRoute_acquireStateInstanceThread_1", t, func() {
		stateRoute := mockStateRoute()
		stateRoute.stateRoute["1"] = &StateInstance{
			StateID: "1",
			status:  InstanceAbnormal,
			instance: &types.Instance{
				ResKey:     resspeckey.ResSpecKey{},
				InstanceID: "1",
			},
		}
		thread, snerr := stateRoute.acquireStateInstanceThread(&types.InstanceAcquireRequest{
			StateID: "1",
		})
		convey.So(thread, convey.ShouldBeNil)
		convey.So(snerr.Code(), convey.ShouldEqual, statuscode.InstanceStatusAbnormalCode)
	})
	convey.Convey("TestStateRoute_acquireStateInstanceThread_2", t, func() {
		stateRoute := mockStateRoute()
		patch := gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {})
		defer patch.Reset()
		stateRoute.createInstanceFunc = func(resSpec *resspeckey.ResourceSpecification, instanceType types.InstanceType,
			callerPod string) (*types.Instance, error) {
			return &types.Instance{
				ResKey: resspeckey.ResSpecKey{},
				InstanceStatus: commonTypes.InstanceStatus{
					Code: int32(constant.KernelInstanceStatusRunning),
				},
				InstanceID:    "2",
				FuncKey:       "",
				FuncSig:       "sig",
				ConcurrentNum: 0,
			}, nil
		}
		stateRoute.stateRoute["1"] = &StateInstance{
			StateID: "1",
			status:  InstanceAbnormal,
			instance: &types.Instance{
				ResKey:     resspeckey.ResSpecKey{},
				InstanceID: "1",
			},
		}
		thread, _ := stateRoute.acquireStateInstanceThread(&types.InstanceAcquireRequest{
			StateID: "2",
		})
		convey.So(thread, convey.ShouldNotBeNil)
		convey.So(thread.Instance.InstanceID, convey.ShouldEqual, "2")
	})
	convey.Convey("TestStateRoute_acquireStateInstanceThread_3", t, func() {
		stateRoute := mockStateRoute()
		stateRoute.stateRoute["1"] = &StateInstance{
			StateID:  "1",
			status:   InstanceExit,
			instance: nil,
		}
		thread, snerr := stateRoute.acquireStateInstanceThread(&types.InstanceAcquireRequest{
			StateID: "1",
		})
		convey.So(thread, convey.ShouldBeNil)
		convey.So(snerr.Code(), convey.ShouldEqual, statuscode.StateInstanceNotExistedErrCode)
	})
}
