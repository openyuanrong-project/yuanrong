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

// Package snerror -
package snerror

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSnError(t *testing.T) {
	convey.Convey("New", t, func() {
		snErr := New(1000, "test error")
		convey.So(snErr.Code(), convey.ShouldEqual, 1000)
		convey.So(snErr.Error(), convey.ShouldEqual, "test error")
		res := IsUserError(snErr)
		convey.So(res, convey.ShouldEqual, false)
	})
	convey.Convey("NewWithError", t, func() {
		snErr := NewWithError(1000, fmt.Errorf("test error"))
		convey.So(snErr.Code(), convey.ShouldEqual, 1000)
		convey.So(snErr.Error(), convey.ShouldEqual, "test error")
	})
}
