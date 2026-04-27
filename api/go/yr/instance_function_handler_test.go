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

func TestInstanceFunctionHandler(t *testing.T) {
	convey.Convey(
		"Test InstanceFunctionHandler", t, func() {
			convey.Convey(
				"InstanceFunctionHandler success when !isFunction", func() {
					convey.So(func() {
						NewInstanceFunctionHandler("", "")
					}, convey.ShouldPanic)
				},
			)
			funcHandler := NewInstanceFunctionHandler((*CounterTest).Add, "")
			convey.Convey(
				"InstanceFunctionHandler success ", func() {
					convey.So(funcHandler, convey.ShouldNotBeNil)
				},
			)
			convey.Convey(
				"Invoke success ", func() {
					refs := funcHandler.Invoke(1, 2)
					convey.So(refs, convey.ShouldNotBeNil)
				},
			)
		},
	)
}
