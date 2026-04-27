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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/state"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

func TestGetState(t *testing.T) {
	conf := []byte(`{"FaaSControllerConfig":{"frontendInstanceNum":100,"schedulerInstanceNum":100,"faasschedulerConfig":{  "autoScaleConfig":{"slaQuota": 1000,"scaleDownTime": 60000,"burstScaleNum": 1000},"leaseSpan":600000},"faasfrontendConfig":{"slaQuota":1000,"functionCapability":1,"authenticationEnable":false,"trafficLimitDisable":true,"instanceNum":0,"http":{"resptimeout":5,"workerInstanceReadTimeOut":5,"maxRequestBodySize":6},"metaEtcd":{"servers":null,"user":"","password":"","sslEnable":false,"CaFile":"","CertFile":"","KeyFile":""},"routerEtcd":{"servers":null,"user":"","password":"","sslEnable":false,"CaFile":"","CertFile":"","KeyFile":""}},"routerEtcd":{"servers":["1.2.3.4:1234"],"user":"tom","password":"**","sslEnable":false,"CaFile":"","CertFile":"","KeyFile":""},"metaEtcd":{"servers":["1.2.3.4:5678"],"user":"tom","password":"**","sslEnable":false,"CaFile":"","CertFile":"","KeyFile":""},"tlsConfig":{"caContent":"","keyContent":"","certContent":""}},"FaasInstance":{}}`)
	SetState(conf)
	_ = json.Unmarshal(conf, controllerState)
	tests := []struct {
		name    string
		want    *ControllerState
		wantErr bool
	}{
		{
			name:    "get state",
			want:    controllerState,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetState()
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetState() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_updateState(t *testing.T) {
	convey.Convey("updateState", t, func() {
		defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}).Reset()
		defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
			return &etcd3.EtcdClient{}
		}).Reset()
		defer gomonkey.ApplyMethod(reflect.TypeOf(&state.Queue{}), "SaveState",
			func(q *state.Queue, state []byte, key string) error {
				return nil
			}).Reset()
		defer gomonkey.ApplyMethod(reflect.TypeOf(controllerHandlerQueue), "GetState",
			func(q *state.Queue, key string) ([]byte, error) {
				return nil, fmt.Errorf("etcd not init")
			}).Reset()
		config := types.Config{}
		controllerState.FaaSControllerConfig = config
		convey.Convey("nil controllerHandlerQueue", func() {
			updateState(types.Config{ManagerInstanceNum: 1})
			convey.So(GetState().FaaSControllerConfig.ManagerInstanceNum, convey.ShouldEqual, 0)
		})
		InitState(nil)
		convey.Convey("config", func() {
			method := gomonkey.ApplyMethod(reflect.TypeOf(&state.Queue{}), "SaveState", func(_ *state.Queue, state []byte) error {
				return nil
			})
			defer method.Reset()
			updateState(types.Config{ManagerInstanceNum: 1})

			state := GetState()
			convey.So(state.FaaSControllerConfig.ManagerInstanceNum, convey.ShouldEqual, 1)
		})

		convey.Convey("string", func() {
			updateState("instanceID", "wrong tag")
			state := GetState()
			convey.So(state.FaaSControllerConfig.ManagerInstanceNum, convey.ShouldEqual, 0)

			Update("instanceID", types.StateUpdate, types.FaasFrontendInstanceState)
			time.Sleep(50 * time.Millisecond)
			convey.So(len(GetState().FaasInstance[types.FaasFrontendInstanceState]), convey.ShouldEqual, 1)

			Update("instanceID", types.StateDelete, types.FaasFrontendInstanceState)
			time.Sleep(50 * time.Millisecond)
			convey.So(len(GetState().FaasInstance[types.FaasFrontendInstanceState]), convey.ShouldEqual, 0)

			Update("instanceID", "wrong opt", types.FaasFrontendInstanceState)
			time.Sleep(50 * time.Millisecond)
			convey.So(len(GetState().FaasInstance[types.FaasFrontendInstanceState]), convey.ShouldEqual, 0)
			Update(123)
		})
	})

}

func TestGetStateByte(t *testing.T) {
	convey.Convey("GetStateByte", t, func() {
		stateBytes := []byte("123")
		defer gomonkey.ApplyFunc(state.NewStateQueue, func(size int) *state.Queue {
			return &state.Queue{}
		}).Reset()
		defer gomonkey.ApplyMethod(reflect.TypeOf(&state.Queue{}), "GetState",
			func(q *state.Queue, key string) ([]byte, error) {
				return stateBytes, nil
			}).Reset()
		InitState(nil)
		stateByte, err := GetStateByte()
		convey.So(err, convey.ShouldBeNil)
		convey.So(string(stateByte), convey.ShouldEqual, "123")
	})
}

func TestDeleteStateByte(t *testing.T) {
	convey.Convey("delete state byte test", t, func() {
		p := gomonkey.ApplyFunc(state.NewStateQueue, func(size int) *state.Queue {
			return &state.Queue{}
		})
		defer p.Reset()
		p2 := gomonkey.ApplyFunc((*state.Queue).DeleteState, func(_ *state.Queue, key string) error {
			return nil
		})
		defer p2.Reset()
		InitState(nil)
		err := DeleteStateByte()
		convey.So(err, convey.ShouldBeNil)
	})
}
