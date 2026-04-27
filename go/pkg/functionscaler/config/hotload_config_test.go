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
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/monitor"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	serverPort              = "8888"
	defaultMetaEtcdCafile   = "/home/sn/resource/ca/ca.pem"
	defaultMetaEtcdCertfile = "/home/sn/resource/ca/cert.pem"
	defaultMetaEtcdKeyfile  = "/home/sn/resource/ca/key.pem"

	defaultRouterEtcdCafile   = "/home/sn/resource/routerEtcd/ca.pem"
	defaultRouterEtcdCertfile = "/home/sn/resource/routerEtcd/cert.pem"
	defaultRouterEtcdKeyfile  = "/home/sn/resource/routerEtcd/key.pem"
)

var (
	watcher    *monitor.MockFileWatcher
	maxTimeout = 100*24*3600 + 1
	testConfig = types.Configuration{
		CPU:    5,
		Memory: 100,
		AutoScaleConfig: types.AutoScaleConfig{
			SLAQuota:      10,
			ScaleDownTime: 1,
			BurstScaleNum: 1,
		},
		LeaseSpan: 1,
		MetaETCDConfig: etcd3.EtcdConfig{
			Servers:   []string{"127.0.0.1:2379"},
			User:      "root",
			Password:  "0000",
			CaFile:    defaultMetaEtcdCafile,
			CertFile:  defaultMetaEtcdCertfile,
			KeyFile:   defaultMetaEtcdKeyfile,
			SslEnable: true,
		},
		RouterETCDConfig: etcd3.EtcdConfig{
			Servers:   []string{"127.0.0.2:2379"},
			User:      "root",
			Password:  "1111",
			CaFile:    defaultRouterEtcdCafile,
			CertFile:  defaultRouterEtcdCertfile,
			KeyFile:   defaultRouterEtcdKeyfile,
			SslEnable: true,
		},
	}

	testConfig2 = &types.Configuration{
		CPU:    5,
		Memory: 100,
		AutoScaleConfig: types.AutoScaleConfig{
			SLAQuota:      10,
			ScaleDownTime: 1,
			BurstScaleNum: 1,
		},
		LeaseSpan: 1,
		MetaETCDConfig: etcd3.EtcdConfig{
			SslEnable: true,
		},
		RouterETCDConfig: etcd3.EtcdConfig{
			SslEnable: true,
		},
	}
)

func createMockFileWatcher(stopCh <-chan struct{}) (monitor.FileWatcher, error) {
	watcher = &monitor.MockFileWatcher{
		Callbacks: map[string]monitor.FileChangedCallback{},
		StopCh:    stopCh,
		EventChan: make(chan string, 1),
	}
	return watcher, nil
}

func TestWatchConfig(t *testing.T) {
	convey.Convey("TestWatchConfig error", t, func() {
		defer gomonkey.ApplyFunc(monitor.CreateFileWatcher, func(stopCh <-chan struct{}) (monitor.FileWatcher, error) {
			return nil, fmt.Errorf("ioutil.ReadFile error")
		}).Reset()
		stopCh := make(chan struct{}, 1)
		err := WatchConfig(configFilePath, stopCh, nil)
		if err != nil {
			return
		}
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestHotLoadConfig(t *testing.T) {
	convey.Convey("TestHotLoadConfig OK", t, func() {
		patches := gomonkey.NewPatches()
		data, _ := json.Marshal(testConfig)
		patches.ApplyFunc(ioutil.ReadFile, func() ([]byte, error) {
			fmt.Println(string(data))
			return data, nil
		})
		defer func() {
			patches.Reset()
		}()

		GlobalConfig = testConfig

		monitor.SetCreator(createMockFileWatcher)

		stopCh := make(chan struct{}, 1)
		callbackChan := make(chan int, 1)
		WatchConfig(configFilePath, stopCh, func() {
			callbackChan <- 1
			fmt.Println("do call back")
		})

		watcher.EventChan <- configFilePath

		<-callbackChan
		convey.So(GlobalConfig.MetaETCDConfig.SslEnable, convey.ShouldEqual,
			testConfig.MetaETCDConfig.SslEnable)
		convey.So(GlobalConfig.RouterETCDConfig.SslEnable, convey.ShouldEqual,
			testConfig.RouterETCDConfig.SslEnable)
		close(stopCh)
	})

	convey.Convey("TestHotLoadConfig OK 2", t, func() {
		patches := gomonkey.NewPatches()
		data, _ := json.Marshal(testConfig2)
		patches.ApplyFunc(ioutil.ReadFile, func() ([]byte, error) {
			fmt.Println(string(data))
			return data, nil
		})
		defer func() {
			patches.Reset()
		}()

		GlobalConfig = testConfig
		monitor.SetCreator(createMockFileWatcher)

		stopCh := make(chan struct{}, 1)
		callbackChan := make(chan int, 1)
		WatchConfig(configFilePath, stopCh, func() {
			callbackChan <- 1
			fmt.Println("do call back")
		})

		watcher.EventChan <- configFilePath

		<-callbackChan
		convey.So(GlobalConfig.MetaETCDConfig.SslEnable, convey.ShouldEqual,
			testConfig.MetaETCDConfig.SslEnable)
		convey.So(GlobalConfig.RouterETCDConfig.SslEnable, convey.ShouldEqual,
			testConfig.RouterETCDConfig.SslEnable)
		close(stopCh)
	})
}

func TestLoadConfig(t *testing.T) {
	convey.Convey("TestLoadConfig error 0", t, func() {
		defer gomonkey.ApplyFunc(ioutil.ReadFile, func() ([]byte, error) {
			return nil, fmt.Errorf("ioutil.ReadFile error")
		}).Reset()
		_, err := loadConfig(configFilePath)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("TestLoadConfig error 1", t, func() {
		patches := gomonkey.NewPatches()
		data, _ := json.Marshal(testConfig)
		patches.ApplyFunc(ioutil.ReadFile, func() ([]byte, error) {
			fmt.Println(string(data))
			return data, nil
		})
		patches.ApplyFunc(json.Unmarshal, func(data []byte, v any) error {
			return fmt.Errorf("json.Unmarshal error")
		})
		defer func() {
			patches.Reset()
		}()
		_, err := loadConfig(configFilePath)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("TestLoadConfig error 2", t, func() {
		patches := gomonkey.NewPatches()
		data, _ := json.Marshal(testConfig)
		patches.ApplyFunc(ioutil.ReadFile, func() ([]byte, error) {
			fmt.Println(string(data))
			return data, nil
		})
		patches.ApplyFunc(loadFunctionConfig, func(config *types.Configuration) error {
			return fmt.Errorf("loadFunctionConfig error")
		})
		defer func() {
			patches.Reset()
		}()
		_, err := loadConfig(configFilePath)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
