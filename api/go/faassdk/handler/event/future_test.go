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

// Package event -
package event

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/faas-sdk/go-api/context"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/faassdk/utils"
)

func newFuncSpec() *types.FuncSpec {
	return &types.FuncSpec{
		FuncMetaData: types.FuncMetaData{
			FunctionName: "test-future-fuction",
			Runtime:      "go1.x",
			TenantId:     "123456789",
			Version:      "$latest",
			Timeout:      10,
		},
		ResourceMetaData: types.ResourceMetaData{},
		ExtendedMetaData: types.ExtendedMetaData{},
	}
}

func TestInitWithTimeout_OK(t *testing.T) {
	eventHandler := &EventHandler{
		funcSpec:      newFuncSpec(),
		libMap:        nil,
		functionName:  "test-future-fuction",
		args:          nil,
		client:        nil,
		userInitEntry: nil,
		userCallEntry: nil,
	}
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&EventHandler{}), "SetHandler",
			func(eh *EventHandler, functionLibPath, handler string) error {
				if handler == initHandlerName {
					eh.userInitEntry = func(ctx context.RuntimeContext) {
						fmt.Println("userCode init start")
						fmt.Println("userCode init success")
					}
				}
				return nil
			}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	eventHandler.SetHandler("", initHandlerName)
	_, err := eventHandler.initWithTimeout()
	convey.ShouldBeNil(err)
}

func TestInitWithTimeout_Fail(t *testing.T) {
	eventHandler := &EventHandler{
		funcSpec:      newFuncSpec(),
		libMap:        nil,
		functionName:  "test-future-fuction",
		args:          nil,
		client:        nil,
		userInitEntry: nil,
		userCallEntry: nil,
	}

	eventHandler.funcSpec.FuncMetaData.Timeout = 3
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&EventHandler{}), "SetHandler",
			func(eh *EventHandler, functionLibPath, handler string) error {
				if handler == initHandlerName {
					eh.userInitEntry = func(ctx context.RuntimeContext) {
						fmt.Println("userCode init start")
						fmt.Println("userCode init success")
						time.Sleep(5 * time.Second)
					}
				}
				return nil
			}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	eventHandler.SetHandler("", initHandlerName)
	_, err := eventHandler.initWithTimeout()
	convey.ShouldBeNil(err)
	var res *types.CallResponse
	_ = json.Unmarshal([]byte(err.Error()), &res)
	convey.ShouldEqual(res.InnerCode, constants.InitFunctionTimeout)
}

func TestInvokeWithTimeout_OK(t *testing.T) {
	eventHandler := &EventHandler{
		funcSpec:      newFuncSpec(),
		libMap:        nil,
		functionName:  "test-future-fuction",
		args:          nil,
		client:        nil,
		userInitEntry: nil,
		userCallEntry: nil,
	}
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&EventHandler{}), "SetHandler",
			func(eh *EventHandler, functionLibPath, handler string) error {
				if handler == callHandlerName {
					eh.userCallEntry = func(payload []byte, ctx context.RuntimeContext) (interface{}, error) {
						fmt.Println("Handler function")
						fmt.Println("payload:", string(payload))
						return "success", nil
					}
				}
				return nil
			}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	eventHandler.SetHandler("", callHandlerName)
	res, err := eventHandler.invokeWithTimeout(getCallArgs()[1].Data, map[string]string{}, "", utils.ExecutionDuration{})
	convey.ShouldBeNil(err)
	convey.ShouldContain(res, "success")
}

func TestInvokeWithTimeout_Fail(t *testing.T) {
	eventHandler := &EventHandler{
		funcSpec:      newFuncSpec(),
		libMap:        nil,
		functionName:  "test-future-fuction",
		args:          nil,
		client:        nil,
		userInitEntry: nil,
		userCallEntry: nil,
	}

	eventHandler.funcSpec.FuncMetaData.Timeout = 3
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&EventHandler{}), "SetHandler",
			func(eh *EventHandler, functionLibPath, handler string) error {
				if handler == callHandlerName {
					eh.userCallEntry = func(payload []byte, ctx context.RuntimeContext) (interface{}, error) {
						fmt.Println("Handler function")
						time.Sleep(5 * time.Second)
						fmt.Println("payload:", string(payload))
						return "success", nil
					}
				}
				return nil
			}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	eventHandler.SetHandler("", callHandlerName)
	result, err := eventHandler.invokeWithTimeout(getCallArgs()[1].Data, map[string]string{}, "", utils.ExecutionDuration{})
	convey.ShouldBeError(err)
	var res *types.CallResponse
	_ = json.Unmarshal(result, &res)
	convey.ShouldEqual(res.InnerCode, constants.InvokeFunctionTimeout)
}

func TestWaitRespProcess(t *testing.T) {
	convey.Convey(
		"Test waitRespProcess", t, func() {
			convey.Convey("waitRespProcess success", func() {
				go func() {
					stopCh <- struct{}{}
				}()
				resCh := make(chan []interface{})
				timeoutCB := func() ([]byte, error) { return nil, nil }
				respCB := func(res []interface{}) ([]byte, error) { return nil, nil }
				handleErrorRespCB := func(body interface{}, statusCode int) ([]byte, error) { return nil, nil }
				bytes, err := waitRespProcess(time.Second, resCh, timeoutCB, respCB, handleErrorRespCB)
				convey.So(bytes, convey.ShouldNotBeNil)
				convey.So(err, convey.ShouldBeNil)
				close(resCh)
				bytes, err = waitRespProcess(time.Second, resCh, timeoutCB, respCB, handleErrorRespCB)
				convey.So(bytes, convey.ShouldBeNil)
				convey.So(err, convey.ShouldBeNil)
			})
		},
	)
}
