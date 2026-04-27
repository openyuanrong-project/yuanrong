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
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
)

func TestParseHeader(t *testing.T) {
	convey.Convey("test ParseHeader", t, func() {
		convey.Convey("when header is empty", func() {
			var ctx *gin.Context
			result := ParseHeader(ctx)
			convey.So(result, convey.ShouldBeEmpty)
		})

		convey.Convey("when header is not empty", func() {
			ctx := &gin.Context{
				Request: &http.Request{
					Header: map[string][]string{
						"aa": {"bb", "cc"},
					},
				},
			}
			result := ParseHeader(ctx)
			convey.So(result, convey.ShouldNotBeEmpty)
			convey.So(len(result), convey.ShouldEqual, 1)
		})
	})
}
