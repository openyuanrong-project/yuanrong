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

package http

import (
	"github.com/agiledragon/gomonkey/v2"
	"reflect"
	"testing"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/uuid"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

type FakeDataSystemClinet struct{}

func (f *FakeDataSystemClinet) CreateClient(config api.ConnectArguments) (api.KvClient, error) {
	return &FakeDataSystemClinet{}, nil
}

func (f *FakeDataSystemClinet) KVSet(key string, value []byte, param api.SetParam) api.ErrorInfo {
	return api.ErrorInfo{
		Code: 0,
		Err:  nil,
	}
}

func (f *FakeDataSystemClinet) KVSetWithoutKey(value []byte, param api.SetParam) (string, api.ErrorInfo) {
	return string(value), api.ErrorInfo{
		Code: 0,
		Err:  nil,
	}
}

func (f *FakeDataSystemClinet) KVGet(key string, timeoutms ...uint32) ([]byte, api.ErrorInfo) {
	return []byte{}, api.ErrorInfo{
		Code: 0,
		Err:  nil,
	}
}

func (f *FakeDataSystemClinet) KVGetMulti(keys []string, timeoutms ...uint32) ([][]byte, api.ErrorInfo) {
	return [][]byte{}, api.ErrorInfo{
		Code: 0,
		Err:  nil,
	}
}

func (f *FakeDataSystemClinet) KVQuerySize(keys []string) ([]uint64, api.ErrorInfo) {
	return []uint64{}, api.ErrorInfo{
		Code: 0,
		Err:  nil,
	}
}

func (f *FakeDataSystemClinet) KVDel(key string) api.ErrorInfo {
	return api.ErrorInfo{
		Code: 0,
		Err:  nil,
	}
}

func (f *FakeDataSystemClinet) KVDelMulti(keys []string) ([]string, api.ErrorInfo) {
	return []string{}, api.ErrorInfo{
		Code: 0,
		Err:  nil,
	}
}

func (f *FakeDataSystemClinet) GenerateKey() string {
	return ""
}

func (f *FakeDataSystemClinet) SetTraceID(traceID string) {
}

func (f *FakeDataSystemClinet) DestroyClient() {
}

func (f *FakeDataSystemClinet) Put(objectId string, value []byte, param api.PutParam,
	nestedObjectIds ...[]string) error {
	return nil
}

func (f *FakeDataSystemClinet) Get(objectIds []string, timeoutMs ...int64) ([][]byte, error) {
	return [][]byte{}, nil
}

func (f *FakeDataSystemClinet) GIncreaseRef(objectIds []string, remoteClientId ...string) ([]string, error) {
	return []string{}, nil
}

func (f *FakeDataSystemClinet) GDecreaseRef(objectIds []string, remoteClientId ...string) ([]string, error) {
	return []string{}, nil
}

func TestGetStateManager(t *testing.T) {
	dsClient := &FakeDataSystemClinet{}

	sm1 := GetStateManager(dsClient)
	assert.NotNil(t, sm1, "Expected non-nil stateManager instance")

	sm2 := GetStateManager(dsClient)
	assert.Equal(t, sm1, sm2, "Expected the same instance of stateManager")
}

func TestNewState(t *testing.T) {
	convey.Convey("state is already existed", t, func() {
		dsClient := &FakeDataSystemClinet{}
		sm := GetStateManager(dsClient)
		patches := gomonkey.NewPatches()
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVGet", func(_ *FakeDataSystemClinet, key string, timeoutms ...uint32) (string, api.ErrorInfo) {
			return "", api.ErrorInfo{
				Code: cOK,
				Err:  nil,
			}
		})
		defer patches.Reset()
		err := sm.newState("testFunc", "testStateID", "testTraceID")
		if stateErr, ok := err.(api.ErrorInfo); !ok {
			t.Errorf("Expected error type stateError")
		} else {
			convey.So(stateErr.Code, convey.ShouldEqual, StateExistedErrCode)
		}
	})

	convey.Convey("standard flow", t, func() {
		dsClient := &FakeDataSystemClinet{}
		sm := GetStateManager(dsClient)
		patches := gomonkey.NewPatches()
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVGet", func(_ *FakeDataSystemClinet, key string, timeoutms ...uint32) (string, api.ErrorInfo) {
			return "", api.ErrorInfo{
				Code: cNotFound,
				Err:  nil,
			}
		})
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVSet", func(_ *FakeDataSystemClinet, key string, value []byte, param api.SetParam) api.ErrorInfo {
			return api.ErrorInfo{
				Code: cOK,
				Err:  nil,
			}
		})
		defer patches.Reset()
		err := sm.newState("testFunc", "testStateID", "testTraceID")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("datasystem put failed", t, func() {
		dsClient := &FakeDataSystemClinet{}
		sm := GetStateManager(dsClient)
		patches := gomonkey.NewPatches()
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVGet", func(_ *FakeDataSystemClinet, key string, timeoutms ...uint32) (string, api.ErrorInfo) {
			return "", api.ErrorInfo{
				Code: cNotFound,
				Err:  nil,
			}
		})
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVSet", func(_ *FakeDataSystemClinet, key string, value []byte, param api.SetParam) api.ErrorInfo {
			return api.ErrorInfo{
				Code: 4, // todo
				Err:  nil,
			}
		})
		defer patches.Reset()
		err := sm.newState("testFunc", "testStateID", "testTraceID")
		if stateErr, ok := err.(api.ErrorInfo); !ok {
			t.Errorf("Expected error type stateError")
		} else {
			convey.So(stateErr.Code, convey.ShouldEqual, DataSystemInternalErrCode)
		}
	})
}

func TestGetState(t *testing.T) {
	convey.Convey("datesystem InternalErrCode", t, func() {
		dsClient := &FakeDataSystemClinet{}
		sm := GetStateManager(dsClient)
		patches := gomonkey.NewPatches()
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVGet", func(_ *FakeDataSystemClinet, key string, timeoutms ...uint32) (string, api.ErrorInfo) {
			return "", api.ErrorInfo{
				Code: 100,
				Err:  nil,
			}
		})
		defer patches.Reset()
		err := sm.getState("testFunc", "testStateID", "testTraceID")
		if stateErr, ok := err.(api.ErrorInfo); !ok {
			t.Errorf("Expected error type stateError")
		} else {
			convey.So(stateErr.Code, convey.ShouldEqual, DataSystemInternalErrCode)
		}
	})
}

func TestDelState(t *testing.T) {
	convey.Convey("state is not existed", t, func() {
		dsClient := &FakeDataSystemClinet{}
		sm := GetStateManager(dsClient)
		patches := gomonkey.NewPatches()
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVGet", func(_ *FakeDataSystemClinet, key string, timeoutms ...uint32) (string, api.ErrorInfo) {
			return "", api.ErrorInfo{
				Code: cNotFound,
				Err:  nil,
			}
		})
		defer patches.Reset()
		err := sm.delState("testFunc", "testStateID", "testTraceID")
		if stateErr, ok := err.(api.ErrorInfo); !ok {
			t.Errorf("Expected error type stateError")
		} else {
			convey.So(stateErr.Code, convey.ShouldEqual, StateNotExistedErrCode)
		}
	})

	convey.Convey("standard flow", t, func() {
		dsClient := &FakeDataSystemClinet{}
		sm := GetStateManager(dsClient)
		patches := gomonkey.NewPatches()
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVGet", func(_ *FakeDataSystemClinet, key string, timeoutms ...uint32) (string, api.ErrorInfo) {
			return "", api.ErrorInfo{
				Code: cOK,
				Err:  nil,
			}
		})
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVDel", func(_ *FakeDataSystemClinet, key string) api.ErrorInfo {
			return api.ErrorInfo{
				Code: cOK,
				Err:  nil,
			}
		})
		defer patches.Reset()
		err := sm.delState("testFunc", "testStateID", "testTraceID")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("datesystem InternalErrCode", t, func() {
		dsClient := &FakeDataSystemClinet{}
		sm := GetStateManager(dsClient)
		patches := gomonkey.NewPatches()
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVGet", func(_ *FakeDataSystemClinet, key string, timeoutms ...uint32) (string, api.ErrorInfo) {
			return "", api.ErrorInfo{
				Code: cOK,
				Err:  nil,
			}
		})
		patches.ApplyMethod(reflect.TypeOf(dsClient), "KVDel", func(_ *FakeDataSystemClinet, key string) api.ErrorInfo {
			return api.ErrorInfo{
				Code: 100,
				Err:  nil,
			}
		})
		defer patches.Reset()
		err := sm.delState("testFunc", "testStateID", "testTraceID")
		if stateErr, ok := err.(api.ErrorInfo); !ok {
			t.Errorf("Expected error type stateError")
		} else {
			convey.So(stateErr.Code, convey.ShouldEqual, DataSystemInternalErrCode)
		}
	})
}

func TestInstanceFuncs(t *testing.T) {
	dsClient := &FakeDataSystemClinet{}
	sm := GetStateManager(dsClient)
	testFuncKey := "test-func-key"
	testStateID := "test-state-id" + uuid.New().String()
	testLease := api.InstanceAllocation{
		LeaseID: "test-lease-id",
	}
	getLease := sm.getInstance(testFuncKey, testStateID)
	if getLease != nil {
		t.Errorf("Expected getLease to be nil, but it was not")
	}

	sm.addInstance(testFuncKey, testStateID, &testLease)
	getLease = sm.getInstance(testFuncKey, testStateID)

	if testLease != *getLease {
		t.Errorf("Expected getLease and testLease are same")
	}

	sm.delInstance(testFuncKey, testStateID)
	getLease = sm.getInstance(testFuncKey, testStateID)
	if getLease != nil {
		t.Errorf("Expected getLease to be nil, but it was not")
	}
}
