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
Package faassdk
This package provides methods to obtain the execution interface.
*/
package faassdk

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
	"yuanrong.org/kernel/runtime/libruntime"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/handler"
	"yuanrong.org/kernel/runtime/faassdk/handler/http"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/config"
	"yuanrong.org/kernel/runtime/libruntime/execution"
)

func TestFaasHandler_LoadFunction(t *testing.T) {
	convey.Convey(
		"Test FaasHandler LoadFunction", t, func() {
			convey.Convey(
				"Load function success", func() {
					intfs := newFaaSFuncExecutionIntfs()
					convey.So(intfs.GetExecType(), convey.ShouldEqual, execution.ExecutionTypeFaaS)
					codePaths := []string{"/tmp"}
					err := intfs.LoadFunction(codePaths)
					convey.So(err, convey.ShouldBeNil)
				},
			)
		},
	)
}

func Test_faasHandlers_FunctionExecute(t *testing.T) {
	convey.Convey("Test faasHandlers_FunctionExecute", t, func() {
		convey.Convey("CreateInstance and InvokeInstance", func() {
			// 测试创建实例的情况
			h := &faasHandlers{}
			funcMeta := api.FunctionMeta{}
			invokeType := config.CreateInstance
			var args []api.Arg
			returnobjs := []config.DataObject{}

			// 设置预期的错误
			err := h.FunctionExecute(funcMeta, invokeType, args, returnobjs)
			convey.So(err.Error(), convey.ShouldEqual, "invalid funcSpec")
		})
		convey.Convey("InvokeInstance", func() {
			// 测试调用实例的情况
			h := &faasHandlers{}
			funcMeta := api.FunctionMeta{FuncName: "test function"}
			funcSpec := &types.FuncSpec{FuncMetaData: types.FuncMetaData{FunctionName: "test function", Runtime: constants.RuntimeTypeCustomContainer}}
			bytes, _ := json.Marshal(funcSpec)
			invokeType := config.CreateInstance
			args := []api.Arg{
				{
					Type: api.Value,
					Data: bytes,
				},
				{},
			}
			returnobjs := []config.DataObject{}
			cch := &http.CustomContainerHandler{}
			var patches []*gomonkey.Patches
			patches = append(patches,
				gomonkey.ApplyFunc(http.NewCustomContainerHandler, func(funcSpec *types.FuncSpec, client api.LibruntimeAPI) handler.ExecutorHandler {
					return cch
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(cch), "InitHandler", func(ch *http.CustomContainerHandler, args []api.Arg, rt api.LibruntimeAPI) ([]byte, error) {
					return []byte("success"), nil
				}),
				gomonkey.ApplyFunc(libruntime.AllocReturnObject, func(do *config.DataObject, size uint, nestedIds []string, totalNativeBufferSize *uint) error {
					return nil
				}),
				gomonkey.ApplyFunc(libruntime.WriterLatch, func(do *config.DataObject) error {
					return nil
				}),
				gomonkey.ApplyFunc(libruntime.WriterUnlatch, func(do *config.DataObject) error {
					return nil
				}),
				gomonkey.ApplyFunc(libruntime.MemoryCopy, func(do *config.DataObject, src []byte) error {
					return nil
				}),
				gomonkey.ApplyFunc(libruntime.Seal, func(do *config.DataObject) error {
					return nil
				}),
			)
			defer func() {
				for _, patch := range patches {
					time.Sleep(50 * time.Millisecond)
					patch.Reset()
				}
			}()
			// 设置预期的错误
			err := h.FunctionExecute(funcMeta, invokeType, args, returnobjs)
			convey.So(err, convey.ShouldBeNil)

			invokeType = config.InvokeInstance
			args = []api.Arg{
				{
					Type: api.Value,
					Data: []byte("trace ID"),
				},
				{
					Type: api.Value,
				},
			}
			defer gomonkey.ApplyMethod(reflect.TypeOf(cch), "CallHandler", func(ch *http.CustomContainerHandler, args []api.Arg, context map[string]string) ([]byte, error) {
				return []byte("success"), nil
			}).Reset()
			err = h.FunctionExecute(funcMeta, invokeType, args, returnobjs)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("InvalidInvokeType", func() {
			// 测试无效的调用类型的情况
			h := &faasHandlers{}
			funcMeta := api.FunctionMeta{}
			invokeType := 99999
			args := []api.Arg{}
			returnobjs := []config.DataObject{}

			// 设置预期的错误
			err := h.FunctionExecute(funcMeta, config.InvokeType(invokeType), args, returnobjs)
			convey.So(err.Error(), convey.ShouldEqual, "no such invokeType 99999")
		})
	})
}

func Test_faasHandlers_Checkpoint(t *testing.T) {
	convey.Convey("Test_faasHandlers_Checkpoint", t, func() {
		convey.Convey("faasexecutor is nil", func() {
			// 测试faasexecutor为nil的情况
			h := &faasHandlers{
				faasexecutor: nil,
			}
			_, err := h.Checkpoint("123")
			convey.So(err.Error(), convey.ShouldEqual, "faasexcutor not initialized")
		})
		convey.Convey("faasexecutor is not nil", func() {
			// 测试faasexecutor不为nil的情况
			executor := &http.CustomContainerHandler{}
			h := &faasHandlers{
				faasexecutor: executor,
			}
			defer gomonkey.ApplyMethod(reflect.TypeOf(executor), "CheckPointHandler", func(_ *http.CustomContainerHandler, checkPointId string) ([]byte, error) {
				return nil, nil
			}).Reset()
			_, err := h.Checkpoint("123")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_faasHandlers_Recover(t *testing.T) {
	convey.Convey("Test_faasHandlers_Recover", t, func() {
		convey.Convey("faasexecutor is nil", func() {
			// 测试faasexecutor为nil的情况
			h := &faasHandlers{
				faasexecutor: nil,
			}
			err := h.Recover([]byte("123"))
			convey.So(err.Error(), convey.ShouldEqual, "faasexcutor not initialized")
		})
		convey.Convey("faasexecutor is not nil", func() {
			// 测试faasexecutor不为nil的情况
			executor := &http.CustomContainerHandler{}
			h := &faasHandlers{
				faasexecutor: executor,
			}
			defer gomonkey.ApplyMethod(reflect.TypeOf(executor), "RecoverHandler", func(_ *http.CustomContainerHandler, state []byte) error {
				return nil
			}).Reset()
			err := h.Recover([]byte("123"))
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_faasHandlers_Shutdown(t *testing.T) {
	convey.Convey("Test_faasHandlers_Shutdown", t, func() {
		convey.Convey("faasexecutor is nil", func() {
			// 测试faasexecutor为nil的情况
			h := &faasHandlers{
				faasexecutor: nil,
			}
			err := h.Shutdown(60)
			convey.So(err.Error(), convey.ShouldEqual, "faasexcutor not initialized")
		})
		convey.Convey("faasexecutor is not nil", func() {
			// 测试faasexecutor不为nil的情况
			executor := &http.CustomContainerHandler{}
			h := &faasHandlers{
				faasexecutor: executor,
			}
			defer gomonkey.ApplyMethod(reflect.TypeOf(executor), "ShutDownHandler", func(_ *http.CustomContainerHandler, gracePeriodSecond uint64) error {
				return nil
			}).Reset()
			err := h.Shutdown(60)
			convey.So(err, convey.ShouldBeNil)
		})
	})

}

func Test_faasHandlers_Signal(t *testing.T) {
	convey.Convey("Test_faasHandlers_Signal", t, func() {
		convey.Convey("faasexecutor is nil", func() {
			// 测试faasexecutor为nil的情况
			h := &faasHandlers{
				faasexecutor: nil,
			}
			err := h.Signal(65, []byte("123"))
			convey.So(err.Error(), convey.ShouldEqual, "faasexcutor not initialized")
		})
		convey.Convey("faasexecutor is not nil", func() {
			// 测试faasexecutor不为nil的情况
			executor := &http.CustomContainerHandler{}
			h := &faasHandlers{
				faasexecutor: executor,
			}
			defer gomonkey.ApplyMethod(reflect.TypeOf(executor), "SignalHandler", func(_ *http.CustomContainerHandler, signal int32, payload []byte) error {
				return nil
			}).Reset()
			err := h.Signal(65, []byte("123"))
			convey.So(err, convey.ShouldBeNil)
		})
	})

}

func Test_faasHandlers_HealthCheck(t *testing.T) {
	convey.Convey("Test_faasHandlers_HealthCheck", t, func() {
		convey.Convey("faasexecutor is nil", func() {
			// 测试faasexecutor为nil的情况
			h := &faasHandlers{
				faasexecutor: nil,
			}
			health, _ := h.HealthCheck()
			convey.So(health, convey.ShouldEqual, api.Healthy)
		})
		convey.Convey("faasexecutor is not nil", func() {
			// 测试faasexecutor不为nil的情况
			faaExeecutorDoneChan = make(chan struct{})
			defer func() {
				faaExeecutorDoneChan = make(chan struct{})
			}()
			executor := &http.CustomContainerHandler{}
			h := &faasHandlers{
				faasexecutor: executor,
			}
			defer gomonkey.ApplyMethod(reflect.TypeOf(executor), "HealthCheckHandler", func(_ *http.CustomContainerHandler) (api.HealthType, error) {
				return api.SubHealth, nil
			}).Reset()
			go func() {
				close(faaExeecutorDoneChan)
			}()
			health, _ := h.HealthCheck()
			convey.So(health, convey.ShouldEqual, api.SubHealth)
		})
	})
}

func TestInitFaaSExecutor(t *testing.T) {
	convey.Convey("Test InitFaaSExecutor", t, func() {
		convey.Convey("InitFaaSExecutor when case constants.RuntimeTypeHttp", func() {
			funcSpec := &types.FuncSpec{FuncMetaData: types.FuncMetaData{Runtime: "http"}}
			convey.So(func() {
				faasHdlrs.InitFaaSExecutor(funcSpec, faasHdlrs.libruntimeAPI)
			}, convey.ShouldNotPanic)
		})
		convey.Convey("InitFaaSExecutor when default", func() {
			funcSpec := &types.FuncSpec{FuncMetaData: types.FuncMetaData{Runtime: "default"}}
			convey.So(func() {
				faasHdlrs.InitFaaSExecutor(funcSpec, faasHdlrs.libruntimeAPI)
			}, convey.ShouldNotPanic)
		})
	})
}
