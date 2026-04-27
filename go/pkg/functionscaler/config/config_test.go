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

// Package config -
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/sts"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/raw"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestInitConfig(t *testing.T) {
	cfg := &types.Configuration{
		CPU:    999,
		Memory: 999,
		AutoScaleConfig: types.AutoScaleConfig{
			SLAQuota:      1,
			ScaleDownTime: 1,
			BurstScaleNum: 1,
		},
		LeaseSpan: 1,
		RouterETCDConfig: etcd3.EtcdConfig{
			Password: "321",
		},
		MetaETCDConfig: etcd3.EtcdConfig{
			Password: "123",
		},
		SchedulerNum:   1,
		DockerRootPath: "dockerRootPath",
		RawStsConfig: raw.StsConfig{
			StsEnable:        true,
			SensitiveConfigs: raw.SensitiveConfigs{},
			ServerConfig:     raw.ServerConfig{},
			MgmtServerConfig: raw.MgmtServerConfig{},
		},
	}
	p1 := gomonkey.ApplyFunc(sts.InitStsSDK, func(serverCfg raw.ServerConfig) error {
		return nil
	})
	cfgByte, _ := json.Marshal(cfg)
	convey.Convey("success", t, func() {
		err := InitConfig(cfgByte)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("DockerRootPath is empty", t, func() {
		cfg := &types.Configuration{
			CPU:    999,
			Memory: 999,
			AutoScaleConfig: types.AutoScaleConfig{
				SLAQuota:      1,
				ScaleDownTime: 1,
				BurstScaleNum: 1,
			},
			LeaseSpan: 1,
			RouterETCDConfig: etcd3.EtcdConfig{
				Password: "321",
			},
			MetaETCDConfig: etcd3.EtcdConfig{
				Password: "123",
			},
			SchedulerNum:   1,
			DockerRootPath: "",
			RawStsConfig: raw.StsConfig{
				StsEnable:        true,
				SensitiveConfigs: raw.SensitiveConfigs{},
				ServerConfig:     raw.ServerConfig{},
				MgmtServerConfig: raw.MgmtServerConfig{},
			},
		}
		cfgByte, _ := json.Marshal(cfg)
		err := InitConfig(cfgByte)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("GetDecryptFromEnv success", t, func() {
		defer gomonkey.ApplyFunc(localauth.GetDecryptFromEnv, func() (map[string]string, error) {
			return map[string]string{"metaEtcdPwd": "qwerdf"}, nil
		}).Reset()
		err := InitConfig(cfgByte)
		convey.So(err, convey.ShouldBeNil)
		convey.So(GlobalConfig.MetaETCDConfig.Password, convey.ShouldEqual, "qwerdf")
	})
	convey.Convey("GetDecryptFromEnv error", t, func() {
		defer gomonkey.ApplyFunc(localauth.GetDecryptFromEnv, func() (map[string]string, error) {
			return nil, fmt.Errorf("get decrypt from env error")
		}).Reset()
		err := InitConfig(cfgByte)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("Unmarshal error", t, func() {
		err := InitConfig(cfgByte[2:20])
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("ValidateStruct error", t, func() {
		cfg.MetaETCDConfig = etcd3.EtcdConfig{}
		cfgByte, _ := json.Marshal(cfg)
		err := InitConfig(cfgByte)
		convey.So(err, convey.ShouldNotBeNil)
	})
	p1.Reset()

	convey.Convey("sts init error", t, func() {
		p2 := gomonkey.ApplyFunc(sts.InitStsSDK, func(serverCfg raw.ServerConfig) error {
			return errors.New("init sts error")
		})
		err := InitConfig(cfgByte)
		convey.So(err, convey.ShouldNotBeNil)
		p2.Reset()
	})

	convey.Convey("sts init error", t, func() {
		defer gomonkey.ApplyFunc(sts.InitStsSDK, func(serverCfg raw.ServerConfig) error {
			return nil
		}).Reset()
		defer gomonkey.ApplyFunc(os.Setenv, func(key, value string) error {
			return errors.New("set env error")
		}).Reset()
		err := InitConfig(cfgByte)
		convey.So(err, convey.ShouldNotBeNil)
	})

}

func TestInitModuleConfig(t *testing.T) {
	cfg := &types.Configuration{
		CPU:    999,
		Memory: 999,
		AutoScaleConfig: types.AutoScaleConfig{
			SLAQuota:      1,
			ScaleDownTime: 1,
			BurstScaleNum: 1,
		},
		LeaseSpan: 1,
		RouterETCDConfig: etcd3.EtcdConfig{
			Password: "321",
		},
		MetaETCDConfig: etcd3.EtcdConfig{
			Password: "123",
		},
		SchedulerNum: 1,
	}
	cfgByte, _ := json.Marshal(cfg)
	convey.Convey("success", t, func() {
		defer gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			return string(cfgByte)
		}).Reset()
		err := InitModuleConfig()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("Unmarshal error", t, func() {
		defer gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			return "{"
		}).Reset()
		err := InitModuleConfig()
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestInitEtcd(t *testing.T) {
	convey.Convey("Test InitEtcd", t, func() {
		convey.Convey("global config is nil", func() {
			rawGConfig := GlobalConfig
			GlobalConfig = types.Configuration{}
			stopCh := make(chan struct{})
			err := InitEtcd(stopCh)
			GlobalConfig = rawGConfig
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("init route etcd error", func() {
			rawGConfig := GlobalConfig
			GlobalConfig = types.Configuration{
				RouterETCDConfig: etcd3.EtcdConfig{},
			}
			err := InitEtcd(nil)
			GlobalConfig = rawGConfig
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("init metadata etcd error", func() {
			rawGConfig := GlobalConfig
			GlobalConfig = types.Configuration{
				MetaETCDConfig: etcd3.EtcdConfig{},
			}
			err := InitEtcd(nil)
			GlobalConfig = rawGConfig
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestClearSensitiveInfo(t *testing.T) {
	convey.Convey("Test ClearSensitiveInfo", t, func() {
		convey.Convey("global config is nil", func() {
			rawGConfig := GlobalConfig
			GlobalConfig = types.Configuration{}
			ClearSensitiveInfo()
			GlobalConfig = rawGConfig
		})
		convey.Convey("clear sensitive Info", func() {
			rawGConfig := GlobalConfig
			GlobalConfig = types.Configuration{
				RouterETCDConfig: etcd3.EtcdConfig{},
				MetaETCDConfig:   etcd3.EtcdConfig{},
			}
			ClearSensitiveInfo()
			GlobalConfig = rawGConfig
		})
	})
}

func TestSetAlarmEnv(t *testing.T) {
	convey.Convey("Test SetAlarmEnv", t, func() {
		convey.Convey("success", func() {
			rawGConfig := GlobalConfig
			GlobalConfig = types.Configuration{
				ClusterName: "cluster1",
				AlarmConfig: alarm.Config{
					EnableAlarm: true,
				},
			}
			err := setAlarmEnv()
			GlobalConfig = rawGConfig
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
