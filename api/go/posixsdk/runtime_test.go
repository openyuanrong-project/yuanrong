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

// Package posixsdk for init and start
package posixsdk

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/common"
)

func TestInitRuntimeAndRun(t *testing.T) {
	convey.Convey(
		"Test InitRuntimeAndRun", t, func() {
			cfg := common.GetConfig()
			cfg.DriverMode = true
			os.Create("./test.json")
			defer os.Remove("./test.json")
			os.Setenv(initArgsFilePathEnvKey, "./test.json")
			defer os.Setenv(initArgsFilePathEnvKey, "")
			os.Setenv("FUNCTION_LIB_PATH", "/tmp")
			intfs, err := NewPosixFuncExecutionIntfs()
			convey.So(err, convey.ShouldBeNil)
			convey.Convey("Test InitRuntime Failed", func() {
				err = InitRuntime(cfg, intfs)
				Run()
				convey.So(err, convey.ShouldBeNil)
			})
		})
}

func TestExecShutdownHandler(t *testing.T) {
	convey.Convey(
		"Test ExecShutdownHandler", t, func() {
			ExecShutdownHandler(15)
			convey.So(nil, convey.ShouldBeNil)
		})
}
