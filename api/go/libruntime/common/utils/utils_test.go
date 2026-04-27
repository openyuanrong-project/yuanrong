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

// Package utils for common functions
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestValidateFilePath1(t *testing.T) {
	convey.Convey("TestValidateFilePath", t, func() {
		convey.Convey("ValidateFilePath success", func() {
			err := ValidateFilePath("./")
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid file path")
		})
		convey.Convey("ValidateFilePath success when path==/home/test", func() {
			err := ValidateFilePath("/home/test")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetFuncNameFromFuncLibPath(t *testing.T) {
	convey.Convey("TestGetFuncNameFromFuncLibPath", t, func() {
		convey.Convey("GetFuncNameFromFuncLibPath success", func() {
			str := GetFuncNameFromFuncLibPath("/home/test/function_lib")
			convey.So(str, convey.ShouldEqual, "function_lib")
		})
		convey.Convey("ValidateFilePath success when funcLibPath==\"\"", func() {
			str := GetFuncNameFromFuncLibPath("")
			convey.So(str, convey.ShouldBeEmpty)
		})
	})
}
