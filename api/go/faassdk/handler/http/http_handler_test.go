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
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/runtime/faassdk/common/functionlog"
	"yuanrong.org/kernel/runtime/faassdk/common/monitor"
	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/libruntime/api"
)

var userCodePath = "../test/main/user_code_test.so"

func TestHttpHandler_InitHandler(t *testing.T) {
	convey.Convey("HttpHandler_InitHandler", t, func() {
		handler := NewHttpHandler(newFuncSpec(), nil)
		os.Setenv(config.DelegateDownloadPath, userCodePath)
		defer gomonkey.ApplyFunc(syscall.Kill, func(gid int, signal syscall.Signal) error {
			return nil
		}).Reset()
		convey.Convey("invalid create params number", func() {
			res, err := handler.InitHandler([]api.Arg{}, nil)
			convey.So(err.Error(), convey.ShouldEqual, "{\"errorCode\":\"6001\",\"message\":\"invalid create params number\"}")
			convey.So(res, convey.ShouldEqual, []byte{})
		})

		convey.Convey("init failed bootstrap timed out after 3s", func() {
			defer gomonkey.ApplyFunc(functionlog.GetFunctionLogger, func(cfg *config.Configuration) (*functionlog.FunctionLogger, error) {
				return &functionlog.FunctionLogger{}, nil
			}).Reset()
			defer gomonkey.ApplyFunc((*HttpHandler).bootstrap, func(_ *HttpHandler) error {
				return nil
			}).Reset()
			funcSpecBytes, _ := json.Marshal(newFuncSpec())
			createParamsBytes, _ := json.Marshal(newHttpCreateParams())
			schedulerParamsBytes, _ := json.Marshal(newHttpSchedulerParams())
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
					Data: schedulerParamsBytes,
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			res, err := handler.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldContainSubstring, `4211`)
			convey.So(res, convey.ShouldEqual, []byte{})
			healthType, err := handler.HealthCheckHandler()
			convey.So(healthType, convey.ShouldEqual, api.Healthy)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("bootstrap file not exist", func() {
			defer gomonkey.ApplyFunc(functionlog.GetFunctionLogger, func(cfg *config.Configuration) (*functionlog.FunctionLogger, error) {
				return &functionlog.FunctionLogger{}, nil
			}).Reset()
			funcSpecBytes, _ := json.Marshal(newFuncSpec())
			createParamsBytes, _ := json.Marshal(newHttpCreateParams())
			args := []api.Arg{
				{
					Type: 0,
					Data: funcSpecBytes,
				},
				{
					Type: 0,
					Data: createParamsBytes,
				},
			}
			res, err := handler.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldNotBeEmpty)
			convey.So(res, convey.ShouldEqual, []byte{})
		})

		convey.Convey("failed to execute bootstrap for http function", func() {
			defer gomonkey.ApplyFunc(functionlog.GetFunctionLogger, func(cfg *config.Configuration) (*functionlog.FunctionLogger, error) {
				return &functionlog.FunctionLogger{}, nil
			}).Reset()
			defer gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, nil
			}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(&exec.Cmd{}), "Start", func(_ *exec.Cmd) error {
				return errors.New("run bootstrap error")
			}).Reset()
			funcSpecBytes, _ := json.Marshal(newFuncSpec())
			createParamsBytes, _ := json.Marshal(newHttpCreateParams())
			schedulerParamsBytes, _ := json.Marshal(newHttpSchedulerParams())
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
					Data: schedulerParamsBytes,
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			res, err := handler.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldEqual, "{\"errorCode\":\"4201\",\"message\":\"init failed bootstrap failed error run bootstrap error\"}")
			convey.So(res, convey.ShouldEqual, []byte{})
		})

		convey.Convey("timeout waiting for http server running", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyFunc(functionlog.GetFunctionLogger, func(cfg *config.Configuration) (*functionlog.FunctionLogger, error) {
					return &functionlog.FunctionLogger{}, nil
				}),
				gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return nil, nil
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&exec.Cmd{}), "Start", func(c *exec.Cmd) error {
					c.Process = &os.Process{}
					return nil
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&os.Process{}), "Wait", func(_ *os.Process) (*os.ProcessState, error) {
					// mock process running
					time.Sleep(5 * time.Second)
					return nil, nil
				}),
				gomonkey.ApplyFunc(net.Dial, func(network, address string) (net.Conn, error) {
					time.Sleep(3*time.Second + 10*time.Millisecond)
					return nil, errors.New("timeout")
				}),
			}
			defer func() {
				for _, patch := range patches {
					patch.Reset()
				}
			}()
			funcSpecBytes, _ := json.Marshal(newFuncSpec())
			createParamsBytes, _ := json.Marshal(newHttpCreateParams())
			args := []api.Arg{
				{
					Type: 0,
					Data: funcSpecBytes,
				},
				{
					Type: 0,
					Data: createParamsBytes,
				},
			}
			res, err := handler.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldNotBeEmpty)
			convey.So(res, convey.ShouldEqual, []byte{})
		})

		convey.Convey("http server process exited", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyFunc(functionlog.GetFunctionLogger, func(cfg *config.Configuration) (*functionlog.FunctionLogger, error) {
					return &functionlog.FunctionLogger{}, nil
				}),
				gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return nil, nil
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&exec.Cmd{}), "Start", func(c *exec.Cmd) error {
					c.Process = &os.Process{}
					return nil
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&os.Process{}), "Wait", func(_ *os.Process) (*os.ProcessState, error) {
					return &os.ProcessState{}, nil
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&os.ProcessState{}), "ExitCode", func(_ *os.ProcessState) int {
					return 0
				}),
				gomonkey.ApplyFunc(net.Dial, func(network, address string) (net.Conn, error) {
					return nil, errors.New("process has done")
				}),
			}
			defer func() {
				for _, patch := range patches {
					patch.Reset()
				}
			}()
			funcSpecBytes, _ := json.Marshal(newFuncSpec())
			createParamsBytes, _ := json.Marshal(newHttpCreateParams())
			args := []api.Arg{
				{
					Type: 0,
					Data: funcSpecBytes,
				},
				{
					Type: 0,
					Data: createParamsBytes,
				},
			}
			res, err := handler.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldNotBeEmpty)
			convey.So(res, convey.ShouldEqual, []byte{})
		})

		convey.Convey("success", func() {
			patches := []*gomonkey.Patches{
				gomonkey.ApplyFunc(functionlog.GetFunctionLogger, func(cfg *config.Configuration) (*functionlog.FunctionLogger, error) {
					return &functionlog.FunctionLogger{}, nil
				}),
				gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return nil, nil
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&exec.Cmd{}), "Start", func(c *exec.Cmd) error {
					c.Process = &os.Process{}
					return nil
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&os.Process{}), "Wait", func(_ *os.Process) (*os.ProcessState, error) {
					// mock process running
					time.Sleep(5 * time.Second)
					return nil, nil
				}),
				gomonkey.ApplyFunc(net.Dial, func(network, address string) (net.Conn, error) {
					return &net.TCPConn{}, nil
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do",
					func(_ *http.Client, req *http.Request) (*http.Response, error) {
						return &http.Response{}, nil
					}),
			}
			defer func() {
				for _, patch := range patches {
					patch.Reset()
				}
			}()
			funcSpecBytes, _ := json.Marshal(newFuncSpec())
			createParamsBytes, _ := json.Marshal(newHttpCreateParams())
			schedulerParamsBytes, _ := json.Marshal(newHttpSchedulerParams())
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
					Data: schedulerParamsBytes,
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}
			res, err := handler.InitHandler(args, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldEqual, []byte{})
		})
	})
}

func TestShutDownHandler(t *testing.T) {
	convey.Convey("ShutDownHandler", t, func() {
		defer gomonkey.ApplyFunc(syscall.Getpgid, func(pid int) (int, error) {
			return 0, errors.New("get gid error")
		}).Reset()
		defer gomonkey.ApplyFunc(syscall.Kill, func(gid int, signal syscall.Signal) error {
			return errors.New("kill error")
		}).Reset()
		handler := NewHttpHandler(newFuncSpec(), nil)
		var err1 error
		go func() {
			time.Sleep(1 * time.Second)
			err1 = handler.ShutDownHandler(30)
		}()
		err := handler.ShutDownHandler(30)
		convey.So(err, convey.ShouldBeNil)
		convey.So(err1, convey.ShouldBeNil)
	})
}

func TestKillAllProcesses(t *testing.T) {
	convey.Convey("TestKillAllProcesses", t, func() {
		patch1 := gomonkey.ApplyFunc(syscall.Getpgid, func(pid int) (int, error) {
			return 0, errors.New("get gid error")
		})
		defer patch1.Reset()
		patch2 := gomonkey.ApplyFunc(syscall.Kill, func(gid int, signal syscall.Signal) error {
			return errors.New("kill error")
		})
		defer patch2.Reset()
		handler := HttpHandler{
			process: &os.Process{
				Pid: 100,
			},
			basicHandler: &basicHandler{
				funcSpec: newFuncSpec(),
			},
		}
		handler.killAllProcesses()
		convey.So(handler.process, convey.ShouldNotBeNil)
	})
}

func TestWaitProcess(t *testing.T) {
	convey.Convey("TestWaitProcess", t, func() {
		convey.Convey("precess is nil", func() {
			handler := HttpHandler{
				waitChan: make(chan processState, 1),
			}
			handler.waitProcess()
			convey.So(handler.process, convey.ShouldBeNil)
		})
		convey.Convey("precess wait error", func() {
			d := &os.Process{}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(d),
				"Wait", func(p *os.Process) (*os.ProcessState, error) {
					return nil, errors.New("error")
				})
			defer patch.Reset()
			handler := HttpHandler{
				process: &os.Process{
					Pid: 100,
				},
				waitChan: make(chan processState, 1),
			}
			handler.waitProcess()
			convey.So(handler.process, convey.ShouldNotBeNil)
		})
		convey.Convey("wait process success ", func() {
			d := &os.Process{}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(d),
				"Wait", func(p *os.Process) (*os.ProcessState, error) {
					return &os.ProcessState{}, nil
				})
			defer patch.Reset()
			handler := HttpHandler{
				process: &os.Process{
					Pid: 100,
				},
				basicHandler: &basicHandler{
					monitor: &monitor.FunctionMonitorManager{
						ErrChan: make(chan error, 1),
					},
				},
				waitChan: make(chan processState, 1),
			}
			handler.waitProcess()
			_, ok := <-handler.basicHandler.monitor.ErrChan
			convey.So(ok, convey.ShouldBeTrue)
		})
	})
}
