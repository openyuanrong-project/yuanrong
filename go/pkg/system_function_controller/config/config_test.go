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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/sts"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
	ftypes "yuanrong.org/kernel/pkg/frontend/types"
	stypes "yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/system_function_controller/state"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

var (
	configString = `{
		"routerEtcd": {
			"servers": ["1.2.3.4:1234"],
			"user": "tom",
			"password": "**"
		},
		"metaEtcd": {
			"servers": ["1.2.3.4:5678"],
			"user": "tom",
			"password": "**"
		},
		"alarmConfig":{"enableAlarm": true}
	}
	`
	schedulerConfigString = `{
		"cpu":999,
		"memory":999,
        "autoScaleConfig":{
              "slaQuota": 1000,
			  "scaleDownTime": 60000,
			  "burstScaleNum": 1000
         },
		"leaseSpan":600000,
		"functionLimitRate":0,
		"routerEtcd":{"servers":["1.2.3.4:1234"],"user":"tom","password":"**","sslEnable":false,"CaFile":"","CertFile":"","KeyFile":""},
		"metaEtcd":{"servers":["1.2.3.4:5678"],"user":"tom","password":"**","sslEnable":false,"CaFile":"","CertFile":"","KeyFile":""},
		"schedulerNum":100
		}`
	frontendConfigString = `{
		"instanceNum":100,
		"cpu":777,
		"memory":777,
		"slaQuota":1000,
		"trafficLimitDisable":true,
		"routerEtcd":{
			"servers": ["1.2.3.4:1234"],
			"user": "tom",
			"password": "**"
		},
		"metaEtcd":{
			"servers": ["1.2.3.4:5678"],
			"user": "tom",
			"password": "**"
		},
		"http":{"maxRequestBodySize": 6}
		}`
	managerConfigString = `{
		"managerInstanceNum":0,
		"routerEtcd":{
			"servers": ["1.2.3.4:1234"],
			"user": "tom",
			"password": "**"
		},
		"metaEtcd":{
			"servers": ["1.2.3.4:5678"],
			"user": "tom",
			"password": "**"
		},
		"alarmConfig":{"enableAlarm": true}
		}`
	lostSTSConfigString = `{
		"routerEtcd": {
			"servers": ["1.2.3.4:1234"],
			"user": "tom",
			"password": "**"
		},
		"metaEtcd": {
			"servers": ["1.2.3.4:5678"],
			"user": "tom",
			"password": "**"
		},
		"rawStsConfig": {
			"stsEnable": true
		}
	}
	`

	invalidConfigString = `{
		"x3": {
			"url": ["1.2.3.4:1234"],
			"username": "tom",
			"password": "**"
		}
	}
	`

	routerEtcdConfig = etcd3.EtcdConfig{
		Servers:  []string{"1.2.3.4:1234"},
		User:     "tom",
		Password: "**",
	}

	metaEtcdConfig = etcd3.EtcdConfig{
		Servers:  []string{"1.2.3.4:5678"},
		User:     "tom",
		Password: "**",
	}

	schedulerBasicConfig = types.SchedulerConfig{
		Configuration: stypes.Configuration{
			CPU:    999,
			Memory: 999,
			AutoScaleConfig: stypes.AutoScaleConfig{
				SLAQuota:      1000,
				ScaleDownTime: 60000,
				BurstScaleNum: 1000,
			},
			LeaseSpan:        600000,
			RouterETCDConfig: routerEtcdConfig,
			MetaETCDConfig:   metaEtcdConfig,
		},
		SchedulerNum: 100,
	}

	frontendConfig = types.FrontendConfig{
		Config: ftypes.Config{
			InstanceNum:          100,
			CPU:                  777,
			Memory:               777,
			SLAQuota:             1000,
			AuthenticationEnable: false,
			HTTPConfig: &ftypes.FrontendHTTP{
				MaxRequestBodySize: 6,
			},
			RouterEtcd: routerEtcdConfig,
			MetaEtcd:   metaEtcdConfig,
		},
	}

	managerConf = &types.ManagerConfig{
		ManagerInstanceNum: 0,
		RouterEtcd:         routerEtcdConfig,
		MetaEtcd:           metaEtcdConfig,
		AlarmConfig:        alarm.Config{EnableAlarm: true},
	}

	expectedFaaSControllerConfig = types.Config{
		RouterETCD:  routerEtcdConfig,
		MetaETCD:    metaEtcdConfig,
		AlarmConfig: alarm.Config{EnableAlarm: true},
	}

	expectedFaaSSchedulerConfig = &types.SchedulerConfig{
		Configuration: schedulerBasicConfig.Configuration,
		SchedulerNum:  100,
	}

	expectedFaaSFrontendConfig = &frontendConfig
)

func TestInitConfig(t *testing.T) {
	defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	Convey("Test InitConfig", t, func() {
		Convey("Test InitConfig with invalid config", func() {
			faaSControllerConfig = nil
			err := InitConfig([]byte("123"))
			So(err, ShouldNotBeNil)

			err = InitConfig([]byte(invalidConfigString))
			So(err, ShouldNotBeNil)
		})

		Convey("Test InitConfig with valid config", func() {
			faaSControllerConfig = nil
			defer gomonkey.ApplyFunc(localauth.GetDecryptFromEnv, func() (map[string]string, error) {
				return map[string]string{"metaEtcdPwd": "123"}, nil
			}).Reset()
			err := InitConfig([]byte(configString))
			So(err, ShouldBeNil)
		})

		Convey("Test InitConfig decrypt error", func() {
			faaSControllerConfig = nil
			defer gomonkey.ApplyFunc(localauth.GetDecryptFromEnv, func() (map[string]string, error) {
				return nil, errors.New("decrypt error")
			}).Reset()
			err := InitConfig([]byte(configString))
			So(err, ShouldNotBeNil)
		})

		Convey("Test InitConfig STS error", func() {
			faaSControllerConfig = nil
			err := InitConfig([]byte(lostSTSConfigString))
			So(err, ShouldNotBeNil)
		})

		Convey("Test InitConfig STS success", func() {
			faaSControllerConfig = nil
			defer gomonkey.ApplyFunc(sts.InitStsSDK, func(serverCfg raw.ServerConfig) error {
				return nil
			}).Reset()
			err := InitConfig([]byte(lostSTSConfigString))
			So(err, ShouldBeNil)
		})
	})
}

func TestGetFaaSControllerConfig(t *testing.T) {
	defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	faaSControllerConfig = nil
	Convey("Test GetFaaSControllerConfig", t, func() {
		InitConfig([]byte(configString))
		got := GetFaaSControllerConfig()
		fmt.Printf("%+v \n", got)
		fmt.Printf("%+v", expectedFaaSControllerConfig)
		So(reflect.DeepEqual(got, expectedFaaSControllerConfig), ShouldBeTrue)
	})
}

func TestGetFaaSSchedulerConfig(t *testing.T) {
	defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	Convey("Test GetFaaSSchedulerConfig", t, func() {
		var cfg *types.SchedulerConfig
		_ = json.Unmarshal([]byte(schedulerConfigString), &cfg)
		UpdateSchedulerConfig(cfg)
		got := GetFaaSSchedulerConfig()
		So(reflect.DeepEqual(got, expectedFaaSSchedulerConfig), ShouldBeTrue)
	})
}

func TestGetFaaSFrontendConfig(t *testing.T) {
	defer gomonkey.ApplyFunc(state.Update, func(value interface{}, tags ...string) {
	}).Reset()
	Convey("Test GetFaaSFrontendConfig", t, func() {
		var cfg *types.FrontendConfig
		_ = json.Unmarshal([]byte(frontendConfigString), &cfg)
		UpdateFrontendConfig(cfg)
		got := GetFaaSFrontendConfig()
		So(got, ShouldResemble, expectedFaaSFrontendConfig)
	})
}

func TestGetFaaSManagerConfig(t *testing.T) {
	Convey("TestGetFaaSManagerConfig", t, func() {
		var cfg *types.ManagerConfig
		_ = json.Unmarshal([]byte(managerConfigString), &cfg)
		UpdateManagerConfig(cfg)
		got := GetFaaSManagerConfig()
		So(reflect.DeepEqual(got, managerConf), ShouldBeTrue)
	})
}

func TestRecoverConfig(t *testing.T) {
	Convey("RecoverConfig", t, func() {
		conf := []byte(`{"FaaSControllerConfig":{
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
	}}`)
		state.SetState(conf)
		err := RecoverConfig()
		So(err, ShouldBeNil)
	})
}

func TestInitEtcd(t *testing.T) {
	Convey("TestInitEtcd", t, func() {
		Convey("failed config is not initialized", func() {
			faaSControllerConfig = nil
			err := InitEtcd(make(chan struct{}))
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&etcd3.EtcdInitParam{}), "InitClient",
				func(_ *etcd3.EtcdInitParam) error {
					return nil
				}).Reset()
			InitConfig([]byte(configString))
			err := InitEtcd(make(chan struct{}))
			So(err, ShouldBeNil)
		})
	})
}
