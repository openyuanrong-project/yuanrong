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

// Package state -
package state

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/state"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestInitState(t *testing.T) {
	convey.Convey("InitState success", t, func() {
		defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}).Reset()
		InitState()
	})
}

func TestOptState(t *testing.T) {
	defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
		return &etcd3.EtcdClient{}
	}).Reset()
	defer gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "Put", func(_ *etcd3.EtcdClient,
		ctxInfo etcd3.EtcdCtxInfo, key string, value string, opts ...clientv3.OpOption) error {
		return nil
	}).Reset()
	InitState()
	schedulerState = &SchedulerState{
		InstancePool: make(map[string]*types.InstancePoolState),
	}
	stateByte, _ := json.Marshal(schedulerState)

	convey.Convey("set state", t, func() {
		err := SetState(stateByte)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("get state", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(schedulerHandlerQueue), "GetState",
			func(q *state.Queue, key string) ([]byte, error) {
				return stateByte, nil
			}).Reset()
		ssByte, err := GetStateByte()
		outPut := &SchedulerState{}
		json.Unmarshal(ssByte, outPut)
		convey.So(err, convey.ShouldBeNil)
	})
	time.Sleep(50 * time.Millisecond)
}

func TestUpdateState(t *testing.T) {
	defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
		return &etcd3.EtcdClient{}
	}).Reset()
	defer gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "Put", func(_ *etcd3.EtcdClient,
		ctxInfo etcd3.EtcdCtxInfo, key string, value string, opts ...clientv3.OpOption) error {
		return nil
	}).Reset()
	// client := mockUtils.FakeLibruntimeSdkClient{}
	InitState()
	schedulerState = &SchedulerState{
		InstancePool: make(map[string]*types.InstancePoolState),
	}
	schedulerState.InstancePool["testStateFuncKey"] = &types.InstancePoolState{
		StateInstance: map[string]*types.Instance{
			"testState": {},
		},
	}
	stateByte, _ := json.Marshal(schedulerState)
	SetState(stateByte)

	convey.Convey("delete funcKey success", t, func() {
		Update(&types.InstancePoolStateInput{
			FuncKey: "testFuncKey2",
		}, types.StateDelete)
		time.Sleep(100 * time.Millisecond)
		convey.So(GetState().InstancePool, convey.ShouldNotContainKey, "testFuncKey2")
	})

	convey.Convey("type is error", t, func() {
		type custom struct{}
		temp1 := *GetState()
		Update(&custom{})
		time.Sleep(100 * time.Millisecond)
		temp2 := *GetState()
		convey.So(temp1, convey.ShouldResemble, temp2)
	})
	convey.Convey("instancepoll tags is error", t, func() {
		Update(&types.InstancePoolStateInput{
			FuncKey:      "testFuncKey3",
			InstanceType: types.InstanceTypeScaled,
			ResKey:       resspeckey.ResSpecKey{CPU: 300, Memory: 128},
			InstanceID:   "InstanceID-891011",
		})
		time.Sleep(100 * time.Millisecond)
		convey.So(GetState().InstancePool, convey.ShouldNotContainKey, "testFuncKey3")

		Update(&types.InstancePoolStateInput{
			FuncKey:      "testFuncKey3",
			InstanceType: types.InstanceTypeScaled,
			ResKey:       resspeckey.ResSpecKey{CPU: 300, Memory: 128},
			InstanceID:   "InstanceID-891011",
		}, "error Opt")
		time.Sleep(100 * time.Millisecond)
		convey.So(GetState().InstancePool, convey.ShouldNotContainKey, "testFuncKey3")
	})

	convey.Convey("update state instance", t, func() {
		Update(&types.InstancePoolStateInput{
			FuncKey:      "testStateFuncKey",
			StateID:      "testState",
			InstanceType: types.InstanceTypeState,
			ResKey:       resspeckey.ResSpecKey{CPU: 300, Memory: 128},
			InstanceID:   "InstanceID-891011",
		}, types.StateUpdate)
		time.Sleep(100 * time.Millisecond)
		convey.So(GetState().InstancePool, convey.ShouldContainKey, "testStateFuncKey")
	})

	convey.Convey("delete state instance", t, func() {
		Update(&types.InstancePoolStateInput{
			FuncKey:      "testStateFuncKey",
			StateID:      "testState",
			InstanceType: types.InstanceTypeState,
			ResKey:       resspeckey.ResSpecKey{CPU: 300, Memory: 128},
			InstanceID:   "InstanceID-891011",
		}, types.StateDelete)
		time.Sleep(100 * time.Millisecond)
		convey.So(len(GetState().InstancePool), convey.ShouldEqual, 1)
	})

	convey.Convey("schedulerHandlerQueue is nil", t, func() {
		schedulerHandlerQueue = nil
		temp1 := *GetState()
		Update(&types.Configuration{
			AutoScaleConfig: types.AutoScaleConfig{
				SLAQuota: 1,
			},
		})
		time.Sleep(100 * time.Millisecond)
		temp2 := *GetState()
		convey.So(temp1, convey.ShouldResemble, temp2)
		updateState(&types.Configuration{
			AutoScaleConfig: types.AutoScaleConfig{
				SLAQuota: 1,
			},
		})
		time.Sleep(100 * time.Millisecond)
		temp2 = *GetState()
		convey.So(temp1, convey.ShouldResemble, temp2)
	})

}

func TestRecoverStateRev(t *testing.T) {
	convey.Convey("RecoverStateRev", t, func() {
		schedulerHandlerQueue = &state.Queue{}
		defer gomonkey.ApplyMethod(reflect.TypeOf(schedulerHandlerQueue), "GetState",
			func(q *state.Queue, key string) ([]byte, error) {
				return []byte(`{"lastInstanceRevision": 10086}`), nil
			}).Reset()
		RecoverStateRev()
		stateRev := GetStateRev()
		convey.So(stateRev, convey.ShouldNotBeNil)
		convey.So(stateRev.LastInstanceRevision, convey.ShouldEqual, 10086)
	})
}
