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

// Package zap zapper log
package zap

import (
	"os"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

func TestNewDevelopmentLog(t *testing.T) {
	convey.Convey(
		"Test NewDevelopmentLog", t, func() {
			convey.Convey(
				"NewDevelopmentLog success", func() {
					logger, err := NewDevelopmentLog()
					convey.So(logger, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestNewConsoleLog(t *testing.T) {
	convey.Convey(
		"Test ConsoleLog", t, func() {
			logger, err := NewConsoleLog()
			z := &LoggerWithFormat{logger}
			convey.Convey(
				"NewConsoleLog success", func() {
					convey.So(logger.Name(), convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"Infof success", func() {
					convey.So(func() {
						z.Infof("msg:%s", "ok")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Errorf success", func() {
					convey.So(func() {
						z.Errorf("msg:%s", "err")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Warnf success", func() {
					convey.So(func() {
						z.Warnf("msg:%s", "warning")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Debugf success", func() {
					convey.So(func() {
						z.Debugf("msg:%s", "ok")
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"Debugf success when config.LogLevel==zapcore.DebugLevel", func() {
					convey.So(func() {
						config.LogLevel = zapcore.DebugLevel
						z.Debugf("msg:%s", "ok")
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestNewWithLevel(t *testing.T) {
	convey.Convey(
		"Test NewWithLevel", t, func() {
			convey.Convey(
				"NewWithLevel success", func() {
					logger, err := NewWithLevel(config.CoreInfo{FilePath: "./"}, false)
					convey.So(logger, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"NewWithLevel success when isAsync==true", func() {
					logger, err := NewWithLevel(config.CoreInfo{FilePath: "./"}, true)
					convey.So(logger, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			convey.Convey(
				"NewWithLevel success when CoreInfo is empty", func() {
					logger, err := NewWithLevel(config.CoreInfo{}, false)
					convey.So(logger, convey.ShouldBeNil)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"NewWithLevel success when CoreInfo is empty1", func() {
					coreInfo := config.CoreInfo{
						FilePath:   "./",
						Tick:       1,
						First:      1,
						Thereafter: 1,
					}
					logger, err := NewWithLevel(coreInfo, false)
					convey.So(logger, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
			cleanFile(".")
		},
	)
}

func cleanFile(fileName string) {
	files, _ := os.ReadDir("./")
	for _, file := range files {
		flag := strings.HasPrefix(file.Name(), fileName)
		if flag {
			os.Remove(file.Name())
		}
	}
}

func TestPriorityHandler(t *testing.T) {
	convey.Convey(
		"Test priorityHandler", t, func() {
			convey.Convey(
				"priorityHandler success", func() {
					flag := priorityHandler(zapcore.InfoLevel)
					convey.So(flag, convey.ShouldBeTrue)
				},
			)
		},
	)
}
