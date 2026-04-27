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

/*
Package yr
This package provides methods to obtain the execution interface.
*/
package yr

import (
	"errors"
	"os"
	"plugin"
	"reflect"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/config"
	"yuanrong.org/kernel/runtime/libruntime/execution"
)

func TestLoadFunction(t *testing.T) {
	convey.Convey(
		"Test ActorHandler LoadFunction", t, func() {
			intfs := newActorFuncExecutionIntfs()
			convey.So(intfs.GetExecType(), convey.ShouldEqual, execution.ExecutionTypeActor)
			convey.Convey("LoadFunction success when path error", func() {
				err := intfs.LoadFunction([]string{"path"})
				convey.So(err, convey.ShouldNotBeNil)
			})
			convey.Convey("LoadFunction success when plugin error", func() {
				os.Create("lib.so")
				err := intfs.LoadFunction([]string{"./"})
				convey.So(err, convey.ShouldNotBeNil)
				os.Remove("lib.so")
			})
			convey.Convey("LoadFunction success when codePaths is empty", func() {
				err := intfs.LoadFunction([]string{})
				convey.So(err, convey.ShouldBeNil)
			})
		})
}

func TestFunctionExecute(t *testing.T) {
	convey.Convey(
		"Test ActorHandler FunctionExecute", t, func() {
			intfs := newActorFuncExecutionIntfs()
			convey.Convey("LoadFunction success when case config.CreateInstance:", func() {
				convey.So(func() {
					intfs.FunctionExecute(api.FunctionMeta{}, config.CreateInstance, []api.Arg{}, []config.DataObject{})
				}, convey.ShouldNotPanic)
			})
			convey.Convey("LoadFunction success when case config.CreateInstanceStateless:", func() {
				convey.So(func() {
					intfs.FunctionExecute(api.FunctionMeta{}, config.CreateInstanceStateless, []api.Arg{}, []config.DataObject{})
				}, convey.ShouldNotPanic)
			})
			convey.Convey("LoadFunction success when case config.InvokeInstance:", func() {
				convey.So(func() {
					intfs.FunctionExecute(api.FunctionMeta{}, config.InvokeInstance, []api.Arg{}, []config.DataObject{})
				}, convey.ShouldNotPanic)
			})
			convey.Convey("LoadFunction success when case config.InvokeInstanceStateless:", func() {
				convey.So(func() {
					intfs.FunctionExecute(api.FunctionMeta{}, config.InvokeInstanceStateless, []api.Arg{}, []config.DataObject{})
				}, convey.ShouldNotPanic)
			})
			convey.Convey("LoadFunction success when case default:", func() {
				convey.So(func() {
					intfs.FunctionExecute(api.FunctionMeta{}, 5, []api.Arg{}, []config.DataObject{})
				}, convey.ShouldNotPanic)
			})
		},
	)
}

func TestCheckpoint(t *testing.T) {
	convey.Convey(
		"Test Checkpoint", t, func() {
			intfs := newActorFuncExecutionIntfs()
			convey.Convey("Checkpoint success", func() {
				bytes, err := intfs.Checkpoint("checkpointID1")
				convey.So(bytes, convey.ShouldBeEmpty)
				convey.So(err, convey.ShouldBeNil)
			})
		},
	)
}

func TestRecover(t *testing.T) {
	convey.Convey(
		"Test Recover", t, func() {
			intfs := newActorFuncExecutionIntfs()
			convey.Convey("Recover success", func() {
				err := intfs.Recover([]byte{})
				convey.So(err, convey.ShouldBeNil)
			})
		})
}

func TestShutdown(t *testing.T) {
	convey.Convey(
		"Test Shutdown", t, func() {
			intfs := newActorFuncExecutionIntfs()
			convey.Convey("Shutdown success", func() {
				err := intfs.Shutdown(1)
				convey.So(err, convey.ShouldBeNil)
			})
		})
}

func TestSignal(t *testing.T) {
	convey.Convey(
		"Test Signal", t, func() {
			intfs := newActorFuncExecutionIntfs()
			convey.Convey("Signal success", func() {
				err := intfs.Signal(1, []byte{})
				convey.So(err, convey.ShouldBeNil)
			})
		})
}

func TestHealthCheck(t *testing.T) {
	convey.Convey(
		"Test HealthCheck", t, func() {
			intfs := newActorFuncExecutionIntfs()
			convey.Convey("HealthCheck success", func() {
				healthType, err := intfs.HealthCheck()
				convey.So(healthType, convey.ShouldBeZeroValue)
				convey.So(err, convey.ShouldBeNil)
			})
		})
}

func TestInvokeFunction(t *testing.T) {
	convey.Convey(
		"Test invokeFunction", t, func() {
			intfs := newActorFuncExecutionIntfs()
			h := intfs.(*actorHandlers)
			h.plugins = map[string]*plugin.Plugin{"p1": {}}
			convey.Convey("invokeFunction success", func() {
				values, err := h.invokeFunction(api.FunctionMeta{}, []api.Arg{})
				convey.So(values, convey.ShouldBeEmpty)
				convey.So(err, convey.ShouldNotBeNil)
			})
			convey.Convey("invokeMemberFunction when h.instance == nil", func() {
				values, err := h.invokeMemberFunction(api.FunctionMeta{}, []api.Arg{})
				convey.So(values, convey.ShouldBeEmpty)
				convey.So(err, convey.ShouldNotBeNil)

			})
			v := reflect.ValueOf("v")
			h.instance = &v
			convey.Convey("invokeMemberFunction success", func() {
				values, err := h.invokeMemberFunction(api.FunctionMeta{}, []api.Arg{})
				convey.So(values, convey.ShouldBeEmpty)
				convey.So(err, convey.ShouldNotBeNil)
			})
		})
}

func TestCatchUserErr(t *testing.T) {
	convey.Convey("Test catchUserErr", t, func() {
		values := []reflect.Value{reflect.ValueOf(&api.FunctionMeta{}), reflect.ValueOf("add")}
		convey.Convey("catchUserErr success", func() {
			err := catchUserErr(values)
			convey.So(err, convey.ShouldBeNil)
		})
		values = []reflect.Value{reflect.ValueOf(errors.New("err1"))}
		convey.Convey("catchUserErr when error", func() {
			err := catchUserErr(values)
			convey.So(err.Error(), convey.ShouldEqual, "err1")
		})
	})
}

func TestProcessInvokeRes(t *testing.T) {
	convey.Convey("Test processInvokeRes", t, func() {
		values := []reflect.Value{reflect.ValueOf(&api.FunctionMeta{}), reflect.ValueOf("add")}
		convey.Convey("processInvokeRes when len(returnValues) == 0", func() {
			err := processInvokeRes([]reflect.Value{}, []config.DataObject{})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("processInvokeRes success", func() {
			err := processInvokeRes(values, []config.DataObject{config.DataObject{}})
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestParseArgsValue(t *testing.T) {
	convey.Convey("Test parseArgsValue", t, func() {
		convey.Convey("parseArgsValue when len(args) != method.Type().NumIn()", func() {
			values, err := parseArgsValue(reflect.ValueOf(func(foo int) {}), []api.Arg{})
			convey.So(len(values), convey.ShouldBeZeroValue)
			convey.So(err.Error(), convey.ShouldEqual, "args number is not valid")
		})
		convey.Convey("parseArgsValue when parseArgValue error", func() {
			args := []api.Arg{{Data: []byte("a")}}
			values, err := parseArgsValue(reflect.ValueOf(func(foo int) {}), args)
			convey.So(len(values), convey.ShouldEqual, 1)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("parseArgsValue success", func() {
			values, err := parseArgsValue(reflect.ValueOf(func(foo int) {}), []api.Arg{{Data: make([]byte, 1)}})
			convey.So(len(values), convey.ShouldEqual, 1)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestParseArgValue(t *testing.T) {
	convey.Convey("Test parseArgValue", t, func() {
		convey.Convey("parseArgsValue when arg.Type == api.ObjectRef ", func() {
			value, err := parseArgValue(api.Arg{Type: api.ObjectRef}, reflect.TypeOf("v"))
			convey.So(value, convey.ShouldEqual, reflect.Value{})
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
