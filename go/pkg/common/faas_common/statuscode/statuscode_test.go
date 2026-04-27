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

// Package statuscode define status code of Frontend
package statuscode

import (
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestStatusCode(t *testing.T) {
	convey.Convey("get code", t, func() {
		code := Code(InnerResponseSuccessCode)
		convey.So(code, convey.ShouldEqual, http.StatusOK)
	})
	convey.Convey("get message", t, func() {
		msg := Message(InnerResponseSuccessCode)
		convey.So(msg, convey.ShouldEqual, "OK")
	})
	convey.Convey("error code get message", t, func() {
		msg := Message(999999)
		convey.So(msg, convey.ShouldEqual, "")
	})
	convey.Convey("error code get message", t, func() {
		code := Code(999999)
		convey.So(code, convey.ShouldEqual, http.StatusInternalServerError)
	})
}

func TestGetKernelErrorCode(t *testing.T) {
	type args struct {
		errMsg string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"case1 unknow error", args{errMsg: "unknown error"}, InternalErrorCode},
		{"case2 get code", args{errMsg: "code: 1007,"}, 1007},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetKernelErrorCode(tt.args.errMsg); got != tt.want {
				t.Errorf("GetKernelErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKernelErrorMessage(t *testing.T) {
	type args struct {
		errMsg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case1 unknow message", args{errMsg: "unknown message"}, ""},
		{"case2 get message", args{errMsg: "message: yes"}, "yes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetKernelErrorMessage(tt.args.errMsg); got != tt.want {
				t.Errorf("GetKernelErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
