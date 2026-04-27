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

// Package context
package context

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestInitializeContext(t *testing.T) {
	convey.Convey(
		"Test InitializeContext", t, func() {
			convey.Convey(
				"InitializeContext success", func() {
					p := &Provider{
						CtxEnv:      &Env{},
						CtxHTTPHead: &InvokeContext{},
					}
					os.Setenv("RUNTIME_PROJECT_ID", "rtProjectID")
					os.Setenv("RUNTIME_PACKAGE", "rtPackage")
					os.Setenv("RUNTIME_FUNC_NAME", "rtFcName")
					os.Setenv("RUNTIME_FUNC_VERSION", "rtFcVersion")
					os.Setenv("RUNTIME_MEMORY", "1")
					os.Setenv("RUNTIME_CPU", "1")
					err := p.InitializeContext()
					convey.So(err, convey.ShouldBeNil)
					os.Setenv("RUNTIME_USERDATA", "1")
					err = p.InitializeContext()
					convey.So(err, convey.ShouldNotBeNil)
					os.Setenv("RUNTIME_USERDATA", "")
					os.Setenv("ENV_DELEGATE_DECRYPT", "1")
					err = p.InitializeContext()
					convey.So(err, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestAtoi(t *testing.T) {
	convey.Convey(
		"Test atoi", t, func() {
			convey.Convey(
				"atoi success", func() {
					i := atoi("a")
					convey.So(i, convey.ShouldBeZeroValue)
				},
			)
		},
	)
}
