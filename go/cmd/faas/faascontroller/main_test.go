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

//go:build function

package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/faascontroller"
	"yuanrong.org/kernel/pkg/system_function_controller/state"
)

func TestCallHandler(t *testing.T) {
	type args struct {
		args      []*api.Arg
		createOpt map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "nil response",
			args: args{
				args:      nil,
				createOpt: make(map[string]string),
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CallHandler(tt.args.args, tt.args.createOpt)
			if (err != nil) != tt.wantErr {
				t.Errorf("CallHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CallHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckpointHandler(t *testing.T) {
	type args struct {
		checkpointID string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "nil response",
			args:    args{checkpointID: "123456"},
			wantErr: false,
			want:    []byte(`{"FaaSControllerConfig":null,"FaasInstance":null}`),
		},
	}
	defer gomonkey.ApplyFunc(state.GetStateByte, func() ([]byte, error) {
		return json.Marshal(&state.ControllerState{})
	}).Reset()
	state.InitState()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckpointHandler(tt.args.checkpointID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckpointHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckpointHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var configString = `{
		"frontendInstanceNum": 100,
		"schedulerInstanceNum": 100,
		"faasschedulerConfig": {
			"cpu": 777,
			"memory": 777,
			"slaQuota": 1000,
			"scaleDownTime": 60000,
			"burstScaleNum": 1000,
			"leaseSpan": 600000
		},
		"faasfrontendConfig": {
			"cpu": 777,
			"memory": 777,
			"slaQuota": 1000,
			"functionCapability": 1,
			"authenticationEnable": false,
			"trafficLimitDisable": true,
			"http": {
                "resptimeout": 5,
                "workerInstanceReadTimeOut": 5,
                "maxRequestBodySize": 6
            }
		},
		"routerEtcd": {
			"servers": ["1.2.3.4:1234"],
			"user": "tom",
			"password": "**"
		},
		"metaEtcd": {
			"servers": ["1.2.3.4:5678"],
			"user": "tom",
			"password": "**"
		}
	}
	`

func TestInitHandler(t *testing.T) {
	type args struct {
		args         []*api.Arg
		fsClient     api.FunctionSystemClient
		dsClient     api.DataSystemClient
		formatLogger api.FormatLogger
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "init_args_error",
			args: args{
				args:         []*api.Arg{},
				fsClient:     nil,
				dsClient:     nil,
				formatLogger: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "init_args_type_error",
			args: args{
				args: []*api.Arg{&api.Arg{
					ArgType: 1,
					Data:    nil,
				}},
				fsClient: nil,
				dsClient: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "init_config_error",
			args: args{
				args: []*api.Arg{&api.Arg{
					ArgType: api.Value,
					Data:    nil,
				}},
				fsClient: nil,
				dsClient: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "init_etcd_error",
			args: args{
				args: []*api.Arg{&api.Arg{
					ArgType: api.Value,
					Data:    []byte(configString),
				}},
				fsClient: nil,
				dsClient: nil,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyFunc(utils.DelayExit, func(module string, err error) { return }),
			}
			got, err := InitHandler(tt.args.args, tt.args.fsClient, tt.args.dsClient, tt.args.formatLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitHandler() got = %v, want %v", got, tt.want)
			}
			for _, patch := range patches {
				patch.Reset()
			}
		})
	}

	convey.Convey("init success", t, func() {
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFunc(faascontroller.NewFaaSController,
				func(sdkClient api.FunctionSystemClient, stopCh chan struct{}) (*faascontroller.FaaSController, error) {
					return &faascontroller.FaaSController{}, nil
				}),
			gomonkey.ApplyMethod(reflect.TypeOf(&faascontroller.FaaSController{}), "RegistryList",
				func(_ *faascontroller.FaaSController, stopCh <-chan struct{}) error {
					return nil
				}),
			gomonkey.ApplyMethod(reflect.TypeOf(&faascontroller.FaaSController{}), "CreateExpectedAllInstanceCount",
				func(_ *faascontroller.FaaSController) error {
					return nil
				}),
			gomonkey.ApplyFunc(state.InitState, func() {
				return
			}),
			gomonkey.ApplyFunc(config.InitEtcd, func(stopCh <-chan struct{}) error {
				return nil
			}),
		}
		defer func() {
			close(stopCh)
			for _, patch := range patches {
				patch.Reset()
			}
		}()
		initArgs := []*api.Arg{
			{
				ArgType: 0,
				Data:    []byte(configString),
			},
		}
		resp, err := InitHandler(initArgs, nil, nil, nil)
		convey.So(resp, convey.ShouldEqual, "")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestRecoverHandler(t *testing.T) {
	convey.Convey("recover success", t, func() {
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFunc(faascontroller.NewFaaSController,
				func(sdkClient api.FunctionSystemClient, stopCh chan struct{}) (*faascontroller.FaaSController, error) {
					return &faascontroller.FaaSController{}, nil
				}),
			gomonkey.ApplyMethod(reflect.TypeOf(&faascontroller.FaaSController{}), "RegistryList",
				func(_ *faascontroller.FaaSController, stopCh <-chan struct{}) error {
					return nil
				}),
			gomonkey.ApplyFunc(config.RecoverConfig, func() error {
				return nil
			}),
			gomonkey.ApplyFunc(state.InitState, func() {
				return
			}),
			gomonkey.ApplyFunc(config.InitEtcd, func(stopCh <-chan struct{}) error {
				return nil
			}),
		}
		defer func() {
			for _, patch := range patches {
				patch.Reset()
			}
		}()
		convey.Convey("success", func() {
			patches = append(patches, gomonkey.ApplyMethod(reflect.TypeOf(&faascontroller.FaaSController{}), "CreateExpectedAllInstanceCount",
				func(_ *faascontroller.FaaSController) error {
					return nil
				}))
			err := RecoverHandler([]byte(configString), nil, nil, nil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("failed", func() {
			patches = append(patches, gomonkey.ApplyMethod(reflect.TypeOf(&faascontroller.FaaSController{}), "CreateExpectedAllInstanceCount",
				func(_ *faascontroller.FaaSController) error {
					return errors.New("fail to CreateExpectedAllInstanceCount")
				}))
			err := RecoverHandler([]byte(configString), nil, nil, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestShutdownHandler(t *testing.T) {
	stopCh = make(chan struct{})
	convey.Convey("ShutdownHandler", t, func() {
		faasController = &faascontroller.FaaSController{}
		convey.Convey("success", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyMethod(reflect.TypeOf(faasController), "KillAllInstances",
					func(_ *faascontroller.FaaSController) {

					}),
			}
			defer func() {
				for _, patch := range patches {
					patch.Reset()
				}
			}()
			err := ShutdownHandler(30)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSignalHandler(t *testing.T) {
	convey.Convey("SignalHandler", t, func() {
		err := SignalHandler(0, []byte{})
		convey.So(err, convey.ShouldBeNil)
	})
}
