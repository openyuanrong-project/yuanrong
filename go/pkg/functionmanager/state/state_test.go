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

package state

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/state"
	commonstate "yuanrong.org/kernel/pkg/common/faas_common/state"
	"yuanrong.org/kernel/pkg/functionmanager/types"
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
	managerState = &ManagerState{
		PatKeyList: make(map[string]map[string]struct{}),
	}
	stateByte, _ := json.Marshal(managerState)

	convey.Convey("set state", t, func() {
		err := SetState(stateByte)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("get state byte", t, func() {
		defer gomonkey.ApplyMethod(reflect.TypeOf(managerHandlerQueue), "GetState",
			func(q *state.Queue, key string) ([]byte, error) {
				return stateByte, nil
			}).Reset()
		msByte, err := GetStateByte()
		outPut := &ManagerState{}
		json.Unmarshal(msByte, outPut)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestUpdateState(t *testing.T) {
	defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
		return &etcd3.EtcdClient{}
	}).Reset()
	defer gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdClient{}), "Put", func(_ *etcd3.EtcdClient,
		ctxInfo etcd3.EtcdCtxInfo, key string, value string, opts ...clientv3.OpOption) error {
		return nil
	}).Reset()
	InitState()
	managerState = &ManagerState{
		PatKeyList: make(map[string]map[string]struct{}),
	}
	stateByte, _ := json.Marshal(managerState)
	SetState(stateByte)

	convey.Convey("update PatKeyList success", t, func() {
		Update(map[string]map[string]struct{}{
			"123": map[string]struct{}{
				"356": struct{}{},
			},
		})
		time.Sleep(100 * time.Millisecond)
		convey.So(GetState().PatKeyList["123"], convey.ShouldContainKey, "356")
	})
	convey.Convey("type is error", t, func() {
		type custom struct{}
		temp1 := *GetState()
		Update(&custom{})
		time.Sleep(100 * time.Millisecond)
		temp2 := *GetState()
		convey.So(temp1, convey.ShouldResemble, temp2)
	})
	convey.Convey("managerHandlerQueue is nil", t, func() {
		managerHandlerQueue = nil
		temp1 := *GetState()
		Update(&types.ManagerConfig{
			FunctionCapability: 3,
		})
		time.Sleep(100 * time.Millisecond)
		temp2 := *GetState()
		convey.So(temp1, convey.ShouldResemble, temp2)
	})

}

func Test_updateState(t *testing.T) {
	convey.Convey("updateState", t, func() {
		convey.Convey("manager state managerHandlerQueue is nil", func() {
			managerHandlerQueue = nil
			updateState("state1", "add")
		})

		defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}).Reset()
		convey.Convey("json.Marshal error", func() {
			defer gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, errors.New("json marshal error")
			}).Reset()
			managerHandlerQueue = state.NewStateQueue(defaultHandlerQueueSize)
			updateState(&types.ManagerConfig{}, "add")
		})

		convey.Convey("save manager state error", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(managerHandlerQueue), "SaveState",
				func(q *commonstate.Queue, state []byte) error {
					return errors.New("save error")
				}).Reset()
			updateState(&types.ManagerConfig{}, "add")
		})
	})
}
