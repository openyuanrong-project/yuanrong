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

// Package utils
package utils

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestHandleInitResponse(t *testing.T) {
	convey.Convey("HandleInitResponse", t, func() {
		convey.Convey("ExecutorErrCodeInitFail", func() {
			defer gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, errors.New("json marshal error")
			}).Reset()
			resp, err := HandleInitResponse("hello world", 0)
			convey.So(resp, convey.ShouldEqual, "")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestHandleCallResponse(t *testing.T) {
	convey.Convey("HandleCallResponse", t, func() {
		totalTime := ExecutionDuration{
			ExecutorBeginTime: time.Now(),
			UserFuncBeginTime: time.Now(),
			UserFuncTotalTime: 0,
		}
		response, err := HandleCallResponse(createLargeStr(), 200, "", totalTime, nil)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(response) > 6291456, convey.ShouldEqual, true)
	})
}

func createLargeStr() string {
	builder := strings.Builder{}
	for i := 0; i < 1024*1024; i++ {
		builder.WriteString("aaaaaa")
	}
	return builder.String()
}

func TestHandleResponseBody(t *testing.T) {
	convey.Convey("HandleResponseBody", t, func() {
		body, err := HandleResponseBody([]byte{})
		convey.So(body, convey.ShouldEqual, []byte{})
		convey.So(err, convey.ShouldBeNil)
	})
}
