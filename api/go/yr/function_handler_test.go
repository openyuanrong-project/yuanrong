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

// Package yr for actor
package yr

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestFunctionHandler(t *testing.T) {
	convey.Convey(
		"Test FunctionHandler", t, func() {
			convey.Convey(
				"Function success when !IsFunction", func() {
					convey.So(func() {
						Function("")
					}, convey.ShouldPanic)
				},
			)
			handler := Function(PlusOne)
			convey.Convey(
				"Function success", func() {
					convey.So(handler, convey.ShouldNotBeNil)
				},
			)
			opts := new(InvokeOptions)
			newHandler := handler.Options(opts)
			convey.Convey(
				"Options success", func() {
					convey.So(newHandler, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Invoke success", func() {
					refs := newHandler.Invoke(2, 3)
					convey.So(refs, convey.ShouldNotBeNil)
				},
			)
		},
	)
}
