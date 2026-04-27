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
	"reflect"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetFuncInfos(t *testing.T) {
	convey.Convey(
		"Test getFuncInfos", t, func() {
			var str, str1, str2, str3 string
			var err error

			convey.Convey(
				"when s == \"\" success", func() {
					str = ""
					str1, str2, str3, err = getFuncInfos(str)
					convey.So(str1, convey.ShouldBeEmpty)
					convey.So(str2, convey.ShouldBeEmpty)
					convey.So(str3, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldNotBeNil)
				},
			)

			convey.Convey(
				"when sByte[l-1] != ')' success", func() {
					str = "1.1"
					str1, str2, str3, err = getFuncInfos(str)
					convey.So(str1, convey.ShouldEqual, defaultClassName)
					convey.So(str2, convey.ShouldEqual, "1")
					convey.So(str3, convey.ShouldBeEmpty)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func TestGetStackTraceInfos(t *testing.T) {
	stacks := getStackTraceInfos()
	if len(stacks) != 0 {
		t.Errorf("failed to test getStackTraceInfos")
	}
}

func TestParseStack(t *testing.T) {
	GetConfigManager().EnableCallStack = true
	funcInfo := "plugin/unnamed-6748bd019e0aa0de0891ed0eb3826f47e13fb900.Functiond(0x45,0x57)"
	fileInfo := "./test_exception.go:34 +0x265"
	stack, err := parseStack(funcInfo, fileInfo)
	if err != nil {
		t.Errorf("parse stack shoud success %s", err.Error())
		return
	}
	if stack.MethodName != "Functiond" {
		t.Errorf("methodName should be Functiond, but get %s", stack.MethodName)
		return
	}
	if stack.LineNumber != 34 {
		t.Errorf("methodName should be 34, but get %d", stack.LineNumber)
		return
	}
	fileInfo = "./test_exception.go:aa +0x265"
	stack, err = parseStack(funcInfo, fileInfo)
	if err == nil {
		t.Errorf("parse stack %s shoud be failed", fileInfo)
		return
	}
}

func TestGetCallStackTracesMgr(t *testing.T) {
	callNum, enbaleCall := getCallStackTracesMgr()
	if callNum != defaultStackTracesNum || enbaleCall == false {
		t.Error("failed call stack configure")
		return
	}
}

func TestGetStackTrace(t *testing.T) {
	GetConfigManager().EnableCallStack = true

	type mockInterface struct{}
	m := &mockInterface{}
	r := reflect.ValueOf(m)
	stack := GetStackTrace(&r, "mockFunction")
	if stack.ClassName != defaultClassName || stack.MethodName != "(*mockInterface).mockFunction" {
		t.Errorf("failed to get stack, %v", stack)
	}
	stack = GetStackTrace(nil, "mockFunction")
	if stack.ClassName != defaultClassName || stack.MethodName != "mockFunction" {
		t.Errorf("failed to get stack, %v", stack)
	}
}

func TestGetMethodName(t *testing.T) {
	convey.Convey(
		"Test getMethodName", t, func() {
			convey.Convey(
				"when len(infos) < 2 success", func() {
					str := "main"
					instance := reflect.ValueOf(str)

					res := getMethodName(&instance, "funcName")
					convey.So(res, convey.ShouldEqual, "funcName")
				},
			)
		},
	)
}
