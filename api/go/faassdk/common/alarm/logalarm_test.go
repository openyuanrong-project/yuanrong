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

// Package alarm alarm log by filebeat
package alarm

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/libruntime/common/logger"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

func TestGetAlarmLogger(t *testing.T) {
	convey.Convey("TestGetAlarmLogger", t, func() {
		convey.Convey("success", func() {
			defer gomonkey.ApplyFunc(config.ExtractCoreInfoFromEnv, func(env string) (config.CoreInfo, error) {
				return config.CoreInfo{FilePath: "./"}, nil
			}).Reset()
			defer gomonkey.ApplyFunc(logger.CreateSink, func(coreInfo config.CoreInfo) (io.Writer, error) {
				return nil, nil
			}).Reset()
			defer gomonkey.ApplyFunc(zapcore.AddSync, func(w io.Writer) zapcore.WriteSyncer {
				return nil
			}).Reset()
			// 设置环境变量
			os.Setenv("WiseCloudSite", "testSite")
			os.Setenv("TenantID", "testTenantID")
			os.Setenv("ApplicationID", "testApplicationID")
			os.Setenv("ServiceID", "testServiceID")
			// 测试正常情况
			logger, err := GetAlarmLogger()
			convey.So(logger, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestReportOrClearAlarm(t *testing.T) {
	convey.Convey("TestReportOrClearAlarm", t, func() {
		convey.Convey("success", func() {
			defer func() {
				alarmLogger = nil
				createLoggerErr = nil
				createLoggerOnce = sync.Once{}
				os.Remove("WiseCloudSite")
				os.Remove("TenantID")
				os.Remove("ApplicationID")
				os.Remove("ServiceID")
			}()
			defer gomonkey.ApplyMethod(reflect.TypeOf(&zap.Logger{}), "Info", func(_ *zap.Logger, msg string, fields ...zap.Field) {
			}).Reset()
			ReportOrClearAlarm(&LogAlarmInfo{}, &Detail{})
		})
	})
}

func TestSetAlarmEnv(t *testing.T) {
	convey.Convey("TestSetAlarmEnv", t, func() {
		SetAlarmEnv(config.CoreInfo{FilePath: "./test"})
		bytes := []byte(os.Getenv(ConfigKey))
		var coreInfo config.CoreInfo
		_ = json.Unmarshal(bytes, &coreInfo)
		convey.So(coreInfo.FilePath, convey.ShouldEqual, "./test")
		os.Remove(ConfigKey)
	})
}

func TestSetClusterNameEnv(t *testing.T) {
	convey.Convey("TestSetClusterNameEnv", t, func() {
		SetXiangYunFourConfigEnv(types.XiangYunFourConfig{Site: "www", TenantID: "testTenantID", ApplicationID: "app", ServiceID: "service"})
		convey.So(os.Getenv(constants.WiseCloudSite), convey.ShouldEqual, "www")
		convey.So(os.Getenv(constants.TenantID), convey.ShouldEqual, "testTenantID")
		convey.So(os.Getenv(constants.ApplicationID), convey.ShouldEqual, "app")
		convey.So(os.Getenv(constants.ServiceID), convey.ShouldEqual, "service")
		os.Remove(constants.WiseCloudSite)
		os.Remove(constants.TenantID)
		os.Remove(constants.ApplicationID)
		os.Remove(constants.ServiceID)
	})
}

func TestSetPodIP(t *testing.T) {
	convey.Convey("TestSetPodIP", t, func() {
		SetPodIP()
		getenv := os.Getenv(constants.PodIPEnvKey)
		convey.So(getenv, convey.ShouldNotEqual, "")
	})
}

func TestSetXiangYunFourConfigEnv(t *testing.T) {
	convey.Convey("TestSetPodIP", t, func() {
		SetClusterNameEnv("cluster")
		getenv := os.Getenv(constants.ClusterName)
		convey.So(getenv, convey.ShouldEqual, "cluster")
		os.Remove(constants.ClusterName)
	})
}
