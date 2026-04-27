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

// Package http -
package http

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func Test_parseAPIGEvent(t *testing.T) {
	convey.Convey("parseAPIGEvent", t, func() {
		event, err := parseAPIGEvent(context.TODO(), []byte(`{`), nil, "/baseURL", "invoke")
		convey.So(err, convey.ShouldBeError)
		convey.So(event, convey.ShouldBeNil)
	})
}

func Test_initRequestHeader(t *testing.T) {
	convey.Convey("initRequestHeader", t, func() {
		req := &http.Request{
			Header: make(map[string][]string),
		}

		event := &APIGTriggerEvent{
			Headers: make(map[string]interface{}),
		}
		event.Headers["test-header"] = []string{"testHeader"}
		initRequestHeader(req, event)
		convey.So(req.Header["Test-Header"][0], convey.ShouldEqual, "testHeader")
	})
}

func Test_fromHTTPResponse(t *testing.T) {
	convey.Convey("fromHTTPResponse", t, func() {
		resp := &http.Response{StatusCode: 200, Header: make(map[string][]string)}
		resp.Header["mockHeader"] = []string{"test-header"}
		body := &bytes.Buffer{}

		apigResponse, err := fromHTTPResponse(resp, body)
		convey.So(err, convey.ShouldBeNil)
		convey.So(apigResponse.StatusCode, convey.ShouldEqual, 200)
		convey.So(apigResponse.Headers["mockHeader"][0], convey.ShouldEqual, "test-header")
	})
}
