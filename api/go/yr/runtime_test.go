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

// Package yr for init and start
package yr

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
			os.Setenv("FUNCTION_LIB_PATH", "/tmp")
			convey.Convey("Test InitRuntime Failed", func() {
				err := InitRuntime()
				Run()
				convey.So(err, convey.ShouldBeNil)
			})
		})
}
