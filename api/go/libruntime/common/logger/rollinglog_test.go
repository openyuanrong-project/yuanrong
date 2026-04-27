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
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

func TestInitRollingLog(t *testing.T) {
	convey.Convey(
		"Test initRollingLog", t, func() {
			convey.Convey(
				"initRollingLog success", func() {
					coreInfo := config.CoreInfo{FilePath: "test", IsUserLog: true}
					r, err := initRollingLog(coreInfo, os.O_WRONLY|os.O_APPEND|os.O_CREATE, defaultPerm)
					convey.So(r, convey.ShouldNotBeNil)
					convey.So(err, convey.ShouldBeNil)
					cleanFile("test")
				},
			)
		},
	)
}

func TestRollingLogHandler(t *testing.T) {
	convey.Convey(
		"Test RollingLogHandler", t, func() {
			fp := "test"
			coreInfo := config.CoreInfo{FilePath: fp, IsUserLog: true}
			r, _ := initRollingLog(coreInfo, os.O_WRONLY|os.O_APPEND|os.O_CREATE, defaultPerm)
			defer cleanFile(fp)
			convey.Convey(
				"tidySinks success", func() {
					err := r.tidySinks()
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"cleanRedundantSinks success", func() {
					r.maxBackups = 0
					convey.So(func() {
						r.cleanRedundantSinks()
					}, convey.ShouldNotPanic)
				},
			)
			convey.Convey(
				"newName success", func() {
					str := r.newName()
					convey.So(str, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestNewName(t *testing.T) {
	convey.Convey(
		"Test newName", t, func() {
			fp := "test@#"
			coreInfo := config.CoreInfo{FilePath: fp, IsUserLog: true}
			r, _ := initRollingLog(coreInfo, os.O_WRONLY|os.O_APPEND|os.O_CREATE, defaultPerm)
			defer cleanFile(fp)
			convey.Convey(
				"newName success", func() {
					str := r.newName()
					convey.So(str, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestRollingLogWrite(t *testing.T) {
	convey.Convey(
		"Test RollingLogWrite", t, func() {

			convey.Convey(
				"Write success when file==nil", func() {
					r := &rollingLog{}
					i, err := r.Write([]byte{})
					convey.So(i, convey.ShouldEqual, 0)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Write success", func() {
					fp := "test"
					coreInfo := config.CoreInfo{FilePath: fp, IsUserLog: true}
					r, _ := initRollingLog(coreInfo, os.O_WRONLY|os.O_APPEND|os.O_CREATE, defaultPerm)
					r.maxSize = 0
					defer cleanFile(fp)

					i, err := r.Write([]byte{0})
					convey.So(i, convey.ShouldNotEqual, 0)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestTryRotate(t *testing.T) {
	convey.Convey(
		"Test tryRotate", t, func() {
			fp := "test"
			coreInfo := config.CoreInfo{FilePath: fp, IsUserLog: true}
			r, _ := initRollingLog(coreInfo, os.O_WRONLY|os.O_APPEND|os.O_CREATE, defaultPerm)
			r.maxSize = 0
			cleanFile(fp)
			convey.Convey(
				"newName success", func() {
					convey.So(func() {
						r.tryRotate()
					}, convey.ShouldNotPanic)
				},
			)
		},
	)
}

func TestGetLogName(t *testing.T) {
	convey.Convey(
		"Test GetLogName", t, func() {
			convey.Convey(
				"GetLogName success", func() {
					str := GetLogName("templateTest")
					convey.So(str, convey.ShouldEqual, "")
				},
			)
		},
	)
}
