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
Package posixsdk
This package provides methods to obtain the execution interface.
*/
package posixsdk

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/config"
	"yuanrong.org/kernel/runtime/libruntime/execution"
)

func TestLoad(t *testing.T) {
	convey.Convey(
		"Test Load", t, func() {
			convey.Convey("Test invalid handler name", func() {
				os.Setenv(functionLibraryPath, "/tmp")
				_, err := Load()
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "not found")
			})
		})
}

func TestPosixHandler_FunctionExecute(t *testing.T) {
	convey.Convey(
		"Test PosixHandler FunctionExecute", t, func() {
			convey.Convey("Test no such invokeType", func() {
				err := posixHandler.FunctionExecute(api.FunctionMeta{}, 4, []api.Arg{}, []config.DataObject{})
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "no such invokeType")
			})
			posixHandler.InitHandler = func([]api.Arg, api.LibruntimeAPI) ([]byte, error) { return nil, nil }
			invokeType := config.CreateInstance
			convey.Convey("FunctionExecute when case config.CreateInstance", func() {
				err := posixHandler.FunctionExecute(api.FunctionMeta{}, invokeType, []api.Arg{}, []config.DataObject{})
				convey.So(err, convey.ShouldBeNil)
			})
			posixHandler.CallHandler = func([]api.Arg) ([]byte, error) { return nil, nil }
			invokeType = config.InvokeInstance
			convey.Convey("FunctionExecute when case config.InvokeInstance", func() {
				err := posixHandler.FunctionExecute(api.FunctionMeta{}, invokeType, []api.Arg{}, []config.DataObject{})
				convey.So(err, convey.ShouldBeNil)
			})
			convey.Convey("FunctionExecute when len(returnobjs) > 0", func() {
				err := posixHandler.FunctionExecute(api.FunctionMeta{}, invokeType, []api.Arg{}, []config.DataObject{{}})
				convey.So(err, convey.ShouldBeNil)
			})
			convey.Convey("FunctionExecute when do.ID == \"returnByMsg\"", func() {
				returnObjs := []config.DataObject{{ID: "returnByMsg"}}
				err := posixHandler.FunctionExecute(api.FunctionMeta{}, invokeType, []api.Arg{}, returnObjs)
				convey.So(err, convey.ShouldBeNil)
			})
		})
}

func TestSetHandler(t *testing.T) {
	convey.Convey("Test setHandler", t, func() {
		functionLibPath := "/tmp"
		envKey := "initHandler"
		convey.Convey("setHandler success", func() {
			os.Setenv(envKey, "example.init")
			convey.So(func() {
				setHandler(functionLibPath, envKey)
			}, convey.ShouldNotPanic)
		})
	})
}

func TestGetLib(t *testing.T) {
	convey.Convey("Test getLib", t, func() {
		convey.Convey("getLib when path == \"\"", func() {
			symbol, err := getLib("", "handler")
			convey.So(symbol, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid handler name")
		})
	})
}

func TestGetLibInfo(t *testing.T) {
	convey.Convey("Test getLibInfo", t, func() {
		convey.Convey("getLibInfo when length > minPathPartLength", func() {
			fileName, handlerName := getLibInfo("/tmp", "test.faasfrontend.init")
			convey.So(fileName, convey.ShouldEqual, "/tmp/test/faasfrontend.so")
			convey.So(handlerName, convey.ShouldEqual, "initLibruntime")
		})
	})
}

func TestParseHandlersType(t *testing.T) {
	convey.Convey("Test parseHandlersType", t, func() {
		convey.Convey("parseHandlersType when hasSuffix actorHandlersString", func() {
			executionType := parseHandlersType("lib" + actorHandlersString)
			convey.So(executionType, convey.ShouldEqual, execution.ExecutionTypeActor)
		})
		convey.Convey("parseHandlersType when hasSuffix faaSHandlersString", func() {
			executionType := parseHandlersType("lib" + faaSHandlersString)
			convey.So(executionType, convey.ShouldEqual, execution.ExecutionTypeFaaS)
		})
	})
}

func TestPosixHandlers(t *testing.T) {
	convey.Convey("Test posixHandlers", t, func() {
		convey.Convey("LoadFunction success", func() {
			err := posixHandler.LoadFunction([]string{})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("Checkpoint success", func() {
			bytes, err := posixHandler.Checkpoint("checkpointID1")
			convey.So(bytes, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeNil)
			posixHandler.CheckPointHandler = func(string) ([]byte, error) { return nil, nil }
			bytes, err = posixHandler.Checkpoint("checkpointID1")
			convey.So(bytes, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("Recover success", func() {
			err := posixHandler.Recover([]byte{})
			convey.So(err, convey.ShouldBeNil)
			posixHandler.RecoverHandler = func([]byte, api.LibruntimeAPI) error { return nil }
			err = posixHandler.Recover([]byte{})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("Shutdown success", func() {
			err := posixHandler.Shutdown(0)
			convey.So(err, convey.ShouldBeNil)
			posixHandler.ShutDownHandler = func(gracePeriodSecond uint64) error { return nil }
			err = posixHandler.Shutdown(0)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("Signal success", func() {
			err := posixHandler.Signal(0, []byte{})
			convey.So(err, convey.ShouldBeNil)
			posixHandler.SignalHandler = func(signal int, payload []byte) error { return nil }
			err = posixHandler.Signal(0, []byte{})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("HealthCheck success", func() {
			healthType, err := posixHandler.HealthCheck()
			convey.So(healthType, convey.ShouldEqual, api.Healthy)
			convey.So(err, convey.ShouldBeNil)
			posixHandler.HealthCheckHandler = func() (api.HealthType, error) { return api.Healthy, nil }
			healthType, err = posixHandler.HealthCheck()
			convey.So(healthType, convey.ShouldEqual, api.Healthy)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
