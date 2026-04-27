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
package config

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/asaskevich/govalidator/v11"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionmanager/state"
	"yuanrong.org/kernel/pkg/functionmanager/types"
)

var (
	configString = `{
		"httpsEnable": true,
		"functionCapability": 100,
		"authenticationEnable": true
	}
	`
)

func TestGetConfig(t *testing.T) {
	defer gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient {
		return &etcd3.EtcdClient{}
	}).Reset()
	state.InitState()
	tests := []struct {
		name string
		want types.ManagerConfig
	}{
		{"case1", types.ManagerConfig{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitConfig(t *testing.T) {
	state.InitState()
	type args struct {
		data []byte
	}
	a := args{}
	b := args{
		data: []byte(configString),
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", a, true},
		{"case2", b, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitConfig(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("InitConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	convey.Convey("ValidateStruct error", t, func() {
		defer gomonkey.ApplyFunc(govalidator.ValidateStruct, func(s interface{}) (bool, error) {
			return false, fmt.Errorf("check error")
		}).Reset()
		err := InitConfig([]byte(configString))
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestInitEtcd(t *testing.T) {
	type args struct {
		stopCh <-chan struct{}
	}
	tests := []struct {
		name        string
		args        args
		wantErr     assert.ErrorAssertionFunc
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1 succeed to init etcd", args{stopCh: make(<-chan struct{})}, assert.NoError,
			func() mockUtils.PatchSlice {
				patches := mockUtils.InitPatchSlice()
				patches.Append(mockUtils.PatchSlice{
					gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdInitParam{}), "InitClient",
						func(_ *etcd3.EtcdInitParam) error { return nil }),
				})
				return patches
			}},
		{"case2 failed to init etcd", args{stopCh: make(<-chan struct{})}, assert.Error, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdInitParam{}), "InitClient",
					func(_ *etcd3.EtcdInitParam) error { return errors.New("e") }),
			})
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			tt.wantErr(t, InitEtcd(tt.args.stopCh), fmt.Sprintf("InitEtcd(%v)", tt.args.stopCh))
			patches.ResetAll()
		})
	}
}
