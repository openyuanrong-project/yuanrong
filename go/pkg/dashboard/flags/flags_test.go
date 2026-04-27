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

// Package flags for obtain command line params
package flags

import (
	"os"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/reader"
)

func TestInitConfigNoFile(t *testing.T) {
	convey.Convey("Test initConfig:", t, func() {
		dashboardConfigPath = "config.json"
		convey.Convey("init config when no file", func() {
			err := initConfig(dashboardConfigPath)
			convey.So(os.IsNotExist(err), convey.ShouldBeTrue)
		})
		convey.Convey("init config when format error", func() {
			cofPatches := gomonkey.ApplyFunc(reader.ReadFileWithTimeout, func(configFile string) ([]byte, error) {
				configStr := `{
		"ip": "0.0.0.0",
		"port": "9080"
		}`
				return []byte(configStr), nil
			})
			defer cofPatches.Reset()
			err := initConfig(dashboardConfigPath)
			convey.So(err.Error(), convey.ShouldContainSubstring, "cannot unmarshal")
		})
		convey.Convey("init config when validate error", func() {
			cofPatches := gomonkey.ApplyFunc(reader.ReadFileWithTimeout, func(configFile string) ([]byte, error) {
				configStr := `{
		"ip": "0.0.0.0",
		"port": 9080,
		"functionMasterAddr": "0.0.0.0:1234"
		}`
				return []byte(configStr), nil
			})
			defer cofPatches.Reset()
			err := initConfig(dashboardConfigPath)
			convey.So(err.Error(), convey.ShouldContainSubstring, "0.0.0.0:1234 does not validate as url")
		})
	})
}

func TestLoadConfig(t *testing.T) {
	convey.Convey("Test loadConfig:", t, func() {
		convey.Convey("load config success", func() {
			cofPatches := gomonkey.ApplyFunc(reader.ReadFileWithTimeout, func(configFile string) ([]byte, error) {
				configStr := `{
		"ip": "0.0.0.0",
		"port": 9080,
		"functionMasterAddr": "127.0.0.1:1234",
		"frontendAddr": "127.0.0.1:8888",
		"prometheusAddr": "127.0.0.1:9090",
		"routerEtcdConfig": {
		  "servers": ["127.0.0.1:5678"],
		  "sslEnable": false,
		  "authType": "NOAUTH"
		},
		"metaEtcdConfig": {
		  "servers": ["127.0.0.1:5678"],
		  "sslEnable": false,
		  "authType": "NOAUTH"
		}
		}`
				return []byte(configStr), nil
			})
			defer cofPatches.Reset()
			convey.So(func() {
				loadConfig(dashboardConfigPath, initConfig)
			}, convey.ShouldNotPanic)
		})
		convey.Convey("load config when loadFunc is nil", func() {
			convey.So(func() {
				loadConfig(dashboardConfigPath, nil)
			}, convey.ShouldNotPanic)
		})
	})
}

func TestLoadLogConfig(t *testing.T) {
	convey.Convey("Test loadLogConfig:", t, func() {
		dashboardLogConfigPath = "log.json"
		logPatches := gomonkey.ApplyFunc(log.InitRunLog, func(fileName string, async bool) error {
			return nil
		})
		defer logPatches.Reset()

		convey.Convey("load log config success", func() {
			convey.So(initLog(), convey.ShouldBeNil)
		})
	})
}
