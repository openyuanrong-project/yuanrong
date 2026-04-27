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

// Package log is common logger client
package log

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

func TestLog(t *testing.T) {
	convey.Convey(
		"Test Log", t, func() {
			convey.Convey(
				"Test console log", func() {
					consoleLog := "this is a console log"
					logWrapper := GetLogger()
					logWrapper.Info(consoleLog)
				},
			)
			convey.Convey(
				"Test GetLogger", func() {
					lg, err := InitRunLog("runtime-go", false)
					convey.So(lg, convey.ShouldNotBeNil)
					wrapperLogger := loggerWrapper{
						real: lg,
					}
					convey.So(err, convey.ShouldBeNil)
					convey.Convey(
						"Test log warn", func() {
							wrapperLogger.Info("this is a warn log 1")
							convey.SkipSo()
						},
					)
					convey.Convey(
						"Test log error", func() {
							wrapperLogger.Info("this is a error log 1")
							convey.SkipSo()
						},
					)
					convey.Convey(
						"Test log sync", func() {
							wrapperLogger.Sync()
							convey.SkipSo()
						},
					)
				},
			)
		},
	)
}

func TestInitRunLogWithConfig(t *testing.T) {
	convey.Convey(
		"Test InitRunLogWithConfig", t, func() {
			convey.Convey(
				"InitRunLogWithConfig success", func() {
					coreInfo := config.GetDefaultCoreInfo()
					fl, err := InitRunLogWithConfig("monitor-disk", false, coreInfo)
					convey.So(fl, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"InitRunLogWithConfig when ValidateStruct error", func() {
					fl, err := InitRunLogWithConfig("monitor-disk", false, config.CoreInfo{})
					convey.So(fl, convey.ShouldBeNil)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestConsoleLogger(t *testing.T) {
	convey.Convey(
		"Test ConsoleLogger", t, func() {
			fl := NewConsoleLogger()
			field := zapcore.Field{
				Key:    "key",
				Type:   zapcore.StringType,
				String: "value",
			}
			convey.Convey(
				"NewConsoleLogger success", func() {
					convey.So(fl, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"With success", func() {
					newFl := fl.With(field)
					convey.So(newFl, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Infof success", func() {
					convey.So(func() {
						fl.Infof("msg:", "ok", "err")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Errorf success", func() {
					convey.So(func() {
						fl.Errorf("msg:", "ok", "err")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Warnf success", func() {
					convey.So(func() {
						fl.Warnf("msg:", "ok", "err")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Debugf success", func() {
					convey.So(func() {
						fl.Debugf("msg:", "ok", "err")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Error success", func() {
					convey.So(func() {
						fl.Error("msg:", field)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Warn success", func() {
					convey.So(func() {
						fl.Warn("msg:", field)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Debugf when config.LogLevel == zapcore.DebugLevel", func() {
					convey.So(func() {
						fl.Debugf("msg:", field)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Debug when config.LogLevel == zapcore.DebugLevel", func() {
					convey.So(func() {
						fl.Debug("msg:", field)
					}, convey.ShouldNotPanic)
				},
			)
			config.LogLevel = zapcore.DebugLevel

		},
	)
}

func TestLoggerWrapper(t *testing.T) {
	convey.Convey(
		"Test loggerWrapper", t, func() {
			lw := GetLogger()
			field := zapcore.Field{
				Key:    "key",
				Type:   zapcore.StringType,
				String: "value",
			}
			convey.Convey(
				"With success", func() {
					fl := lw.With(field)
					convey.So(fl, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Infof success", func() {
					convey.So(func() {
						lw.Infof("msg:", "ok", "err")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Errorf success", func() {
					convey.So(func() {
						lw.Errorf("msg:", "ok", "err")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Warnf success", func() {
					convey.So(func() {
						lw.Warnf("msg:", "ok", "err")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Debugf success", func() {
					convey.So(func() {
						lw.Debugf("msg:", "ok", "err")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Error success", func() {
					convey.So(func() {
						lw.Error("msg:", field)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Warn success", func() {
					convey.So(func() {
						lw.Warn("msg:", field)
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Debug success", func() {
					convey.So(func() {
						lw.Debug("msg:", field)
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}
