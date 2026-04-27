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

// Package api for libruntime
package api

import (
	"errors"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNewErrInfo(t *testing.T) {
	info := StackTracesInfo{}
	err := NewErrInfo(0, "", info)
	if err.Err == nil {
		t.Errorf("test NewErrInfo failed")
	}
}

func TestNewErrorInfoWithStackInfo(t *testing.T) {
	stack := StackTrace{
		ClassName:     "main",
		MethodName:    "(*Manager.AddValue)",
		FileName:      "a.go",
		LineNumber:    12,
		ExtensionInfo: nil,
	}
	err := errors.New("error info")
	err = NewErrorInfoWithStackInfo(err, []StackTrace{stack})
	msg := fmt.Sprintf("%s\n%s", "error info", TurnErrInfo(err).getStackTracesInfo())
	if err.Error() != msg {
		t.Errorf("failed to new err to error with stack, %s %s", err.Error(), msg)
	}

	convey.Convey(
		"Test NewErrorInfoWithStackInfo", t, func() {
			convey.Convey(
				"NewErrorInfoWithStackInfo success when len(stacks)==0", func() {
					e := NewErrorInfoWithStackInfo(err, []StackTrace{})
					convey.So(e, convey.ShouldEqual, err)
				},
			)
		},
	)
}

func TestGetStackTracesInfo(t *testing.T) {
	className := "main"
	methodName := "(*Manager.AddValue)"
	fileName := "a.go"
	linNumber := 12
	stack := StackTrace{
		ClassName:     className,
		MethodName:    methodName,
		FileName:      "a.go",
		LineNumber:    12,
		ExtensionInfo: nil,
	}
	err := errors.New("error info")
	errInfo := ErrorInfo{
		Code: 2002,
		Err:  err,
		StackTracesInfo: StackTracesInfo{
			Code:        2002,
			MCode:       0,
			Message:     err.Error(),
			StackTraces: []StackTrace{stack},
		},
	}
	msg := errInfo.getStackTracesInfo()
	funcInfo := fmt.Sprintf("%s.%s%s\n", className, methodName, "")
	fileInfo := fmt.Sprintf("   %s:%d %s\n", fileName, linNumber, "")
	if msg != funcInfo+fileInfo {
		t.Errorf("error message %s shoud be %s", msg, funcInfo+fileInfo)
	}

	convey.Convey(
		"Test getStackTracesInfo", t, func() {
			convey.Convey(
				"getStackTracesInfo success", func() {
					st := StackTrace{LineNumber: 0}
					st.ExtensionInfo = make(map[string]string)
					var e ErrorInfo
					e.StackTracesInfo.StackTraces = []StackTrace{st}
					str := e.getStackTracesInfo()
					convey.So(str, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestIsError(t *testing.T) {
	convey.Convey(
		"Test IsError", t, func() {
			convey.Convey(
				"IsError success", func() {
					var e ErrorInfo
					flag := e.IsError()
					convey.So(flag, convey.ShouldBeFalse)
				},
			)
		},
	)
}

func TestError(t *testing.T) {
	convey.Convey(
		"Test Error", t, func() {
			convey.Convey(
				"Error success", func() {
					var e ErrorInfo
					e.Err = errors.New("error info")
					str := e.Error()
					convey.So(str, convey.ShouldNotBeNil)
				},
			)
		},
	)
}

func TestAddStack(t *testing.T) {
	convey.Convey(
		"Test AddStack", t, func() {
			var stack StackTrace
			err := errors.New("error info")
			convey.Convey(
				"AddStack success when stack is empty", func() {
					e := AddStack(err, stack)
					convey.So(e, convey.ShouldEqual, err)
				},
			)
			convey.Convey(
				"AddStack success", func() {
					stack.ClassName = "stackClassName"
					e := AddStack(err, stack)
					convey.So(e, convey.ShouldNotBeNil)
				},
			)
		},
	)
}
