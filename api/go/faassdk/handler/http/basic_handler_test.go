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

package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"reflect"
	"strings"
	"testing"
	"time"
	"yuanrong.org/kernel/runtime/faassdk/common/functionlog"
	"yuanrong.org/kernel/runtime/faassdk/common/monitor"
	"yuanrong.org/kernel/runtime/faassdk/config"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/faassdk/common/aliasroute"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/libruntime/api"
)

type mockReadCloser struct {
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	bytes := []byte("hello world")
	if p[0] == bytes[0] {
		return 0, io.EOF
	}
	for i, b := range bytes {
		p[i] = b
	}
	return len(bytes), nil
}

func (m *mockReadCloser) Close() error {
	return nil
}

func Test_basicHandler_CallHandler(t *testing.T) {
	config1, _ := config.GetConfig(newFuncSpec())
	fl, _ := functionlog.GetFunctionLogger(config1)
	convey.Convey("CallHandler", t, func() {
		convey.Convey("graceful shutdown, stop processing", func() {
			handler := basicHandler{funcSpec: newFuncSpec(), gracefulShutdown: true}
			res, err := handler.CallHandler(nil, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(strings.Contains(string(res), "graceful shutdown, stop processing"), convey.ShouldEqual, true)
		})

		convey.Convey("invalid invoke argument", func() {
			handler := basicHandler{funcSpec: newFuncSpec()}
			args := []api.Arg{
				{
					Type: 0,
					Data: nil,
				},
			}
			ctx := make(map[string]string)
			res, err := handler.CallHandler(args, ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(strings.Contains(string(res), "invalid invoke argument"), convey.ShouldEqual, true)
		})

		convey.Convey("unmarshal invoke call request data", func() {
			handler := basicHandler{
				funcSpec:       newFuncSpec(),
				baseURL:        fmt.Sprintf("http://%s:%d", listenIP, 8080),
				port:           8080,
				callRoute:      "/callRoute",
				functionLogger: fl,
			}
			args := []api.Arg{
				{
					Type: 0,
					Data: nil,
				},
				{
					Type: 0,
					Data: []byte(`{`),
				},
			}
			ctx := make(map[string]string)
			res, err := handler.CallHandler(args, ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(strings.Contains(string(res), "unmarshal invoke call request data"), convey.ShouldEqual, true)
		})

		convey.Convey("call request timed out", func() {
			defer gomonkey.ApplyFunc(http.NewRequestWithContext,
				func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
					u, _ := urlpkg.Parse(url)
					return &http.Request{URL: u, Header: make(map[string][]string)}, nil
				}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do",
				func(_ *http.Client, req *http.Request) (*http.Response, error) {
					time.Sleep(3*time.Second + 10*time.Millisecond)
					return nil, errors.New("timeout 3s")
				}).Reset()
			handler := basicHandler{
				funcSpec:       newFuncSpec(),
				baseURL:        fmt.Sprintf("http://%s:%d", listenIP, 8080),
				port:           8080,
				callRoute:      "callRoute",
				callTimeout:    3,
				functionLogger: fl,
				monitor:        &monitor.FunctionMonitorManager{},
			}
			apigEvent := &APIGTriggerEvent{
				Body: "hello",
				Path: "test",
			}
			apigEventBytes, _ := json.Marshal(apigEvent)
			userCallRequest := &types.CallRequest{
				Body: apigEventBytes,
			}
			userCallRequestBytes, _ := json.Marshal(userCallRequest)
			args := []api.Arg{
				{
					Type: 0,
					Data: nil,
				},
				{
					Type: 0,
					Data: userCallRequestBytes,
				},
			}
			ctx := make(map[string]string)
			res, err := handler.CallHandler(args, ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(strings.Contains(string(res), "call request timed out"), convey.ShouldEqual, true)
		})

		convey.Convey("call request failed error", func() {
			defer gomonkey.ApplyFunc(http.NewRequestWithContext,
				func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
					u, _ := urlpkg.Parse(url)
					return &http.Request{URL: u, Header: make(map[string][]string)}, nil
				}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do",
				func(_ *http.Client, req *http.Request) (*http.Response, error) {
					return nil, errors.New("user function error")
				}).Reset()
			handler := basicHandler{
				funcSpec:       newFuncSpec(),
				baseURL:        fmt.Sprintf("http://%s:%d", listenIP, 8080),
				port:           8080,
				callRoute:      "callRoute",
				callTimeout:    3,
				functionLogger: fl,
				monitor:        &monitor.FunctionMonitorManager{},
			}
			apigEvent := &APIGTriggerEvent{
				Body: "hello",
				Path: "test",
			}
			apigEventBytes, _ := json.Marshal(apigEvent)
			userCallRequest := &types.CallRequest{
				Body: apigEventBytes,
			}
			userCallRequestBytes, _ := json.Marshal(userCallRequest)
			args := []api.Arg{
				{
					Type: 0,
					Data: nil,
				},
				{
					Type: 0,
					Data: userCallRequestBytes,
				},
			}
			ctx := make(map[string]string)
			res, err := handler.CallHandler(args, ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(strings.Contains(string(res), "user function error"), convey.ShouldEqual, true)
		})

		convey.Convey("success", func() {
			defer gomonkey.ApplyFunc(http.NewRequestWithContext,
				func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
					u, _ := urlpkg.Parse(url)
					return &http.Request{URL: u, Header: make(map[string][]string)}, nil
				}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do",
				func(_ *http.Client, req *http.Request) (*http.Response, error) {
					return &http.Response{Body: &mockReadCloser{}}, nil
				}).Reset()
			handler := basicHandler{
				funcSpec:       newFuncSpec(),
				baseURL:        fmt.Sprintf("http://%s:%d", listenIP, 8080),
				port:           8080,
				callRoute:      "callRoute",
				callTimeout:    3,
				functionLogger: fl,
				monitor:        &monitor.FunctionMonitorManager{},
			}
			apigEvent := &APIGTriggerEvent{
				Body: "hello",
				Path: "test",
			}
			apigEventBytes, _ := json.Marshal(apigEvent)
			userCallRequest := &types.CallRequest{
				Body: apigEventBytes,
			}
			userCallRequestBytes, _ := json.Marshal(userCallRequest)
			args := []api.Arg{
				{
					Type: 0,
					Data: nil,
				},
				{
					Type: 0,
					Data: userCallRequestBytes,
				},
			}
			ctx := make(map[string]string)
			res, err := handler.CallHandler(args, ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(strings.Contains(string(res), base64.StdEncoding.EncodeToString([]byte("hello world"))),
				convey.ShouldEqual, true)
		})

		convey.Convey("circuitBreaker is true", func() {
			handler := basicHandler{funcSpec: newFuncSpec()}
			handler.circuitBreaker = true
			args := []api.Arg{
				{
					Type: 0,
					Data: nil,
				},
			}
			ctx := make(map[string]string)
			res, err := handler.CallHandler(args, ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(strings.Contains(string(res), "function circuit break"), convey.ShouldEqual, true)
		})
	})
}

func TestCheckPointHandler(t *testing.T) {
	convey.Convey("CheckPointHandler", t, func() {
		handler := basicHandler{
			funcSpec:    newFuncSpec(),
			baseURL:     fmt.Sprintf("http://%s:%d", listenIP, 8080),
			port:        8080,
			callRoute:   "callRoute",
			callTimeout: 3,
		}
		bytes, err := handler.CheckPointHandler("checkpointID")
		convey.So(err, convey.ShouldBeNil)
		convey.So(bytes, convey.ShouldBeNil)
	})
}

func TestRecoverHandler(t *testing.T) {
	convey.Convey("RecoverHandler", t, func() {
		handler := basicHandler{
			funcSpec:    newFuncSpec(),
			baseURL:     fmt.Sprintf("http://%s:%d", listenIP, 8080),
			port:        8080,
			callRoute:   "callRoute",
			callTimeout: 3,
		}
		err := handler.RecoverHandler([]byte{})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSignalHandler(t *testing.T) {
	convey.Convey("SignalHandler", t, func() {
		handler := basicHandler{
			funcSpec:    newFuncSpec(),
			baseURL:     fmt.Sprintf("http://%s:%d", listenIP, 8080),
			port:        8080,
			callRoute:   "callRoute",
			callTimeout: 3,
		}
		convey.Convey("handle signal success", func() {
			aliasList := []*aliasroute.AliasElement{{AliasURN: "aaa"}}
			payload, _ := json.Marshal(aliasList)
			err := handler.SignalHandler(64, payload)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("handle signal error", func() {
			err := handler.SignalHandler(64, []byte("aaa"))
			convey.So(err, convey.ShouldNotBeNil)
		})

	})
}

func TestParseCreateParams(t *testing.T) {
	convey.Convey("TestParseCreateParams", t, func() {
		handler := basicHandler{
			funcSpec:    newFuncSpec(),
			baseURL:     fmt.Sprintf("http://%s:%d", listenIP, 8080),
			port:        8080,
			callRoute:   "callRoute",
			callTimeout: 3,
		}
		funcSpecBytes, _ := json.Marshal(newFuncSpecWithoutTimeout())
		createParamsBytes, _ := json.Marshal(newHttpCreateParams())
		schedulerParamsBytes, _ := json.Marshal(newHttpSchedulerParams())
		convey.Convey("unmarshal args[0].Data failed", func() {
			args := []api.Arg{
				{
					Type: 0,
					Data: []byte(""),
				},
				{
					Type: 0,
					Data: createParamsBytes,
				},
				{
					Type: 0,
					Data: schedulerParamsBytes,
				},
				{
					Type: 0,
					Data: []byte(""),
				},
			}
			err := handler.parseCreateParams(args)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("unmarshal args[1].Data failed", func() {
			args := []api.Arg{
				{
					Type: 0,
					Data: funcSpecBytes,
				},
				{
					Type: 0,
					Data: []byte(""),
				},
				{
					Type: 0,
					Data: schedulerParamsBytes,
				},
				{
					Type: 0,
					Data: []byte(""),
				},
			}
			err := handler.parseCreateParams(args)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("unmarshal args[2].Data failed", func() {
			args := []api.Arg{
				{
					Type: 0,
					Data: funcSpecBytes,
				},
				{
					Type: 0,
					Data: createParamsBytes,
				},
				{
					Type: 0,
					Data: []byte(""),
				},
				{
					Type: 0,
					Data: []byte(""),
				},
			}
			err := handler.parseCreateParams(args)
			convey.So(err, convey.ShouldNotBeNil)
		})

	})
}

func Test_basicHandler_isCustomHealthCheckReady(t *testing.T) {
	convey.Convey("test isCustomHealthCheckReady", t, func() {
		convey.Convey("baseline", func() {
			handler := &basicHandler{
				funcSpec: &types.FuncSpec{ExtendedMetaData: types.ExtendedMetaData{
					CustomHealthCheck: types.CustomHealthCheck{
						TimeoutSeconds:   1,
						PeriodSeconds:    1,
						FailureThreshold: 10,
					},
				}},
				baseURL:     "127.0.0.1",
				healthRoute: healthRoute,
			}
			p := gomonkey.ApplyFunc((*basicHandler).sendRequest, func(_ *basicHandler, request *http.Request,
				timeout time.Duration) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK}, nil
			})
			defer p.Reset()
			convey.So(handler.isCustomHealthCheckReady(), convey.ShouldBeTrue)
		})
		convey.Convey("check failed", func() {
			handler := &basicHandler{
				funcSpec: &types.FuncSpec{ExtendedMetaData: types.ExtendedMetaData{
					CustomHealthCheck: types.CustomHealthCheck{
						TimeoutSeconds:   1,
						PeriodSeconds:    1,
						FailureThreshold: 10,
					},
				}},
				baseURL:     "127.0.0.1",
				healthRoute: healthRoute,
			}
			p := gomonkey.ApplyFunc((*basicHandler).sendRequest, func(_ *basicHandler, request *http.Request,
				timeout time.Duration) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError}, nil
			})
			defer p.Reset()
			convey.So(handler.isCustomHealthCheckReady(), convey.ShouldBeFalse)
		})
		convey.Convey("custom health check disabled", func() {
			handler := &basicHandler{
				funcSpec: &types.FuncSpec{ExtendedMetaData: types.ExtendedMetaData{
					CustomHealthCheck: types.CustomHealthCheck{
						TimeoutSeconds:   0,
						PeriodSeconds:    0,
						FailureThreshold: 0,
					},
				}},
				baseURL:     "127.0.0.1",
				healthRoute: healthRoute,
			}
			convey.So(handler.isCustomHealthCheckReady(), convey.ShouldBeTrue)
		})
	})
}
