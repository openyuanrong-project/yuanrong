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

package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClearStringMemory(t *testing.T) {
	Convey("Given a string", t, func() {
		testStr := "helloworld"

		b := []byte(testStr)
		s := string(b)
		Convey("When we clear the string", func() {
			ClearStringMemory(s)
			Convey("The string should be empty", func() {
				So(s, ShouldEqual, string([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
			})
		})
	})
}
