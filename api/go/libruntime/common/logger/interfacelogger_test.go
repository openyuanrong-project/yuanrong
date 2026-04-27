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

// Package logger for log
package logger

import (
	"os"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

func TestNewInterfaceLogger(t *testing.T) {
	convey.Convey(
		"Test NewInterfaceLogger", t, func() {
			convey.Convey(
				"NewInterfaceLogger success", func() {
					logger, err := NewInterfaceLogger("logPath", "fileName", InterfaceEncoderConfig{})
					convey.So(logger, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestWrite(t *testing.T) {
	convey.Convey(
		"Test Write", t, func() {
			convey.Convey(
				"Write success", func() {
					logger, _ := NewInterfaceLogger("logPath", "fileName", InterfaceEncoderConfig{})
					convey.So(func() {
						logger.Write("msg")
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestNewCore(t *testing.T) {
	convey.Convey(
		"Test newCore", t, func() {
			convey.Convey(
				"newCore success", func() {
					core, err := newCore(config.CoreInfo{}, InterfaceEncoderConfig{})
					convey.So(core, convey.ShouldBeNil)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestCreateSink(t *testing.T) {
	convey.Convey(
		"Test CreateSink", t, func() {
			convey.Convey(
				"CreateSink success", func() {
					fp := "test"
					w, err := CreateSink(config.CoreInfo{FilePath: fp})
					convey.So(w, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
					cleanFile(fp)
				},
			)
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
