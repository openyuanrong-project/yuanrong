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

// Package alarm
package alarm

import (
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
)

func TestGetAlarmLogger(t *testing.T) {
	convey.Convey("TestGetAlarmLogger", t, func() {
		convey.Convey("failed to new alarmLogger", func() {
			logger, err := GetAlarmLogger()
			convey.So(err, convey.ShouldBeError)
			convey.So(logger, convey.ShouldBeNil)
		})

		convey.Convey("success", func() {
			dir, _ := os.Getwd()
			defer gomonkey.ApplyFunc(config.ExtractCoreInfoFromEnv, func(env string) (config.CoreInfo, error) {
				return config.CoreInfo{FilePath: dir}, nil
			}).Reset()
			createLoggerOnce = sync.Once{}
			logger, err := GetAlarmLogger()
			convey.So(err, convey.ShouldBeNil)
			convey.So(logger, convey.ShouldNotBeNil)
		})
	})
}

func TestReportOrClearAlarm(t *testing.T) {
	convey.Convey("ReportOrClearAlarm", t, func() {
		convey.Convey("no test assert", func() {
			ReportOrClearAlarm(&LogAlarmInfo{}, &Detail{})
		})
	})
}

func TestSetAlarmEnv(t *testing.T) {
	convey.Convey("SetAlarmEnv", t, func() {
		convey.Convey("set env", func() {
			dir, _ := os.Getwd()
			SetAlarmEnv(config.CoreInfo{FilePath: dir})
			getenv := os.Getenv(ConfigKey)
			var cfg *config.CoreInfo
			err := json.Unmarshal([]byte(getenv), &cfg)
			convey.So(err, convey.ShouldBeNil)
			convey.So(cfg.FilePath, convey.ShouldEqual, dir)
			os.Unsetenv(ConfigKey)
		})
	})
}

func TestSetPodIP(t *testing.T) {
	convey.Convey("SetPodIP", t, func() {
		convey.Convey("", func() {
			ip, _ := urnutils.GetServerIP()
			SetPodIP()
			convey.So(os.Getenv(constant.PodIPEnvKey), convey.ShouldEqual, ip)
			os.Unsetenv(constant.PodIPEnvKey)
		})
	})

}
