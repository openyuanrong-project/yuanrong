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

// Package event for faas executor api
package event

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"plugin"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/common/faasscheduler"
	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/faassdk/faas-sdk/go-api/context"
	"yuanrong.org/kernel/runtime/faassdk/handler"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/libruntime/api"
	log "yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

var userCodePath = "../test/main/user_code_test.so"
var initHandlerName = "user_code_test.InitFunction"
var callHandlerName = "user_code_test.HandlerFunction"
var createParams = `{
			"userInitEntry" : "user_code_test.InitFunction",
			"userCallEntry" : "user_code_test.HandlerFunction"
		}`
var invokeEnvFunctionParam = `{
		"CreateParams": {
			"initEntry" : "user_code_test.InitFunction",
			"callEntry" : "user_code_test.GetEnvFunction"
			}
		}`

type mockLibruntimeClient struct {
}

func (m mockLibruntimeClient) CreateInstance(funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) InvokeByInstanceId(funcMeta api.FunctionMeta, instanceID string, args []api.Arg, invokeOpt api.InvokeOptions) (returnObjectID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) InvokeByFunctionName(funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) AcquireInstance(state string, funcMeta api.FunctionMeta, acquireOpt api.InvokeOptions) (api.InstanceAllocation, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) ReleaseInstance(allocation api.InstanceAllocation, stateID string, abnormal bool, option api.InvokeOptions) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Kill(instanceID string, signal int, payload []byte) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) CreateInstanceRaw(createReqRaw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) InvokeByInstanceIdRaw(invokeReqRaw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KillRaw(killReqRaw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) SaveState(state []byte) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) LoadState(checkpointID string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Exit(code int, message string) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Finalize() {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVSet(key string, value []byte, param api.SetParam) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVSetWithoutKey(value []byte, param api.SetParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f *mockLibruntimeClient) KVMSetTx(keys []string, values [][]byte, param api.MSetParam) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVGet(key string, timeoutms uint) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVGetMulti(keys []string, timeoutms uint) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVDel(key string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVDelMulti(keys []string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) CreateProducer(streamName string, producerConf api.ProducerConf) (api.StreamProducer, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Subscribe(streamName string, config api.SubscriptionConfig) (api.StreamConsumer, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) DeleteStream(streamName string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) QueryGlobalProducersNum(streamName string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) QueryGlobalConsumersNum(streamName string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) SetTraceID(traceID string) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) SetTenantID(tenantID string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Put(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) PutRaw(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Get(objectIDs []string, timeoutMs int) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GetRaw(objectIDs []string, timeoutMs int) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Wait(objectIDs []string, waitNum uint64, timeoutMs int) ([]string, []string, map[string]error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GIncreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GIncreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GDecreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GDecreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GetAsync(objectID string, cb api.GetAsyncCallback) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GetEvent(objectID string, cb api.GetEventCallback) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) DeleteGetEventCallback(objectID string) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GetFormatLogger() api.FormatLogger {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) ReleaseGRefs(remoteClientID string) error {
	// TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GetCredential() api.Credential {
	return api.Credential{}
}

func (m mockLibruntimeClient) UpdateSchdulerInfo(schedulerName string, schedulerId string, option string) {
	return
}

func (m *mockLibruntimeClient) IsHealth() bool {
	return true
}

func (m *mockLibruntimeClient) IsDsHealth() bool {
	return true
}

func (m *mockLibruntimeClient) GetActiveMasterAddr() string {
	return "mockMasterAddr"
}

// FakeLogger -
type FakeLogger struct{}

// With -
func (f *FakeLogger) With(fields ...zapcore.Field) api.FormatLogger {
	return f
}

// Infof -
func (f *FakeLogger) Infof(format string, paras ...interface{}) {}

// Errorf -
func (f *FakeLogger) Errorf(format string, paras ...interface{}) {}

// Warnf -
func (f *FakeLogger) Warnf(format string, paras ...interface{}) {}

// Debugf -
func (f *FakeLogger) Debugf(format string, paras ...interface{}) {}

// Fatalf -
func (f *FakeLogger) Fatalf(format string, paras ...interface{}) {}

// Info -
func (f *FakeLogger) Info(msg string, fields ...zap.Field) {}

// Error -
func (f *FakeLogger) Error(msg string, fields ...zap.Field) {}

// Warn -
func (f *FakeLogger) Warn(msg string, fields ...zap.Field) {}

// Debug -
func (f *FakeLogger) Debug(msg string, fields ...zap.Field) {}

// Fatal -
func (f *FakeLogger) Fatal(msg string, fields ...zap.Field) {}

// Sync -
func (f *FakeLogger) Sync() {}

func TestMain(m *testing.M) {
	log.SetupLogger(&FakeLogger{})
	os.Exit(m.Run())
}

func TestNewEventHandler(t *testing.T) {
	eventHandler := NewEventHandler(newFuncSpec(), nil)
	assert.NotNil(t, eventHandler)
}

func TestGetUserCodePath(t *testing.T) {
	err := os.Setenv(config.DelegateDownloadPath, userCodePath)
	assert.Nil(t, err)
	functionLibPath, err := handler.GetUserCodePath()
	assert.Nil(t, err)
	log.GetLogger().Infof("functionLibPath:%s", functionLibPath)
	assert.NotNil(t, functionLibPath)
}

func TestGetHandlerName(t *testing.T) {
	args := []api.Arg{
		{
			Type: api.Value,
			Data: []byte(""),
		},
		{
			Type: api.Value,
			Data: []byte(createParams),
		},
		{
			Type: api.Value,
			Data: []byte(""),
		},
		{
			Type: api.Value,
			Data: []byte(""),
		},
	}
	userHook, err := getHandlerName(args)
	log.GetLogger().Infof("getInitEntryName:%s", userHook.InitEntry)
	log.GetLogger().Infof("getCallEntryName:%s", userHook.CallEntry)
	assert.NotNil(t, userHook.CallEntry)
	assert.NotNil(t, userHook.InitEntry)
	assert.Nil(t, err)
}

func TestInitHandler(t *testing.T) {
	handler := NewEventHandler(&types.FuncSpec{}, nil)
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
	_, err := handler.InitHandler([]api.Arg{}, nil)
	var res *types.InitResponse
	_ = json.Unmarshal([]byte(err.Error()), &res)
	assert.Equal(t, res.ErrorCode, strconv.Itoa(constants.ExecutorErrCodeInitFail))
	assert.Contains(t, string(res.Message), "invalid args number")
	args, _ := getInitArgs(createParams)
	schedulerParamsBytes, _ := json.Marshal(newHttpSchedulerParams())
	result1, err := handler.InitHandler([]api.Arg{
		{
			Type: api.Value,
			Data: []byte(""),
		},
		args[0],
		{
			Type: api.Value,
			Data: schedulerParamsBytes,
		},
		{
			Type: api.Value,
			Data: []byte(""),
		},
	}, nil)
	assert.Nil(t, err)
	res = &types.InitResponse{}
	_ = json.Unmarshal(result1, &res)
	assert.Equal(t, "", string(res.Message))
}

func newHttpSchedulerParams() *faasscheduler.SchedulerInfo {
	return &faasscheduler.SchedulerInfo{
		SchedulerFuncKey: "scheduler func key",
		SchedulerIDList:  []string{"1111"},
	}
}

// CallHandler, impl of CallHandler
func TestCallHandlerError(t *testing.T) {
	handler := NewEventHandler(&types.FuncSpec{}, nil)

	convey.Convey("CallHandlerError", t, func() {
		convey.Convey("args length invalid", func() {
			result, err := handler.CallHandler([]api.Arg{}, nil)
			convey.So(err, convey.ShouldBeNil)
			var res *types.CallResponse
			_ = json.Unmarshal(result, &res)
			convey.So(string(res.Body), convey.ShouldEqual, `"invalid invoke argument"`)
			convey.So(res.InnerCode, convey.ShouldEqual, "500")
		})

		convey.Convey("the size of the invoke payload exceeds the limit", func() {
			args := []api.Arg{
				{},
				{Type: 0, Data: createLargeBytes(config.MaxPayloadSize)},
			}
			result, err := handler.CallHandler(args, nil)
			convey.So(err, convey.ShouldBeNil)
			var res *types.CallResponse
			_ = json.Unmarshal(result, &res)
			convey.So(string(res.Body), convey.ShouldEqual, fmt.Sprintf(`"the size of the invoke payload exceeds the limit %d"`, config.MaxPayloadSize))
			convey.So(res.InnerCode, convey.ShouldEqual, "500")
		})

		convey.Convey("invoke handler is nil", func() {
			args := []api.Arg{
				{},
				{Type: 0, Data: []byte("hello")},
			}
			result, err := handler.CallHandler(args, nil)
			convey.So(err, convey.ShouldBeNil)
			var res *types.CallResponse
			_ = json.Unmarshal(result, &res)
			convey.So(string(res.Body), convey.ShouldEqual, `"invoke handler is nil"`)
			convey.So(res.InnerCode, convey.ShouldEqual, "4001")
		})

		convey.Convey("unmarshal invoke call request data", func() {
			defer ApplyMethod(reflect.TypeOf(&EventHandler{}), "SetHandler",
				func(eh *EventHandler, functionLibPath, handler string) error {
					if handler == initHandlerName {
						eh.userInitEntry = func(ctx context.RuntimeContext) {
							fmt.Println("userCode init start")
							fmt.Println("userCode init success")
						}
					} else if handler == callHandlerName {
						eh.userCallEntry = func(payload []byte, ctx context.RuntimeContext) (interface{}, error) {
							logger := ctx.GetLogger()
							if logger == nil {
								return nil, fmt.Errorf("user logger not initialized")
							}
							logger.Infof("payload:%s", string(payload))
							fmt.Println("Handler function")
							fmt.Println("payload:", string(payload))
							return "success", nil
						}
					}
					return nil
				}).Reset()
			os.Setenv("RUNTIME_LOG_DIR", "./")
			args, _ := getInitArgs(createParams)
			_, err := handler.InitHandler([]api.Arg{
				{
					Type: api.Value,
					Data: []byte(""),
				},
				args[0],
				{
					Type: api.Value,
					Data: []byte(""),
				},
				{
					Type: api.Value,
					Data: []byte(""),
				},
			}, nil)
			convey.So(err, convey.ShouldBeNil)

			args = []api.Arg{
				{},
				{Type: 0, Data: []byte("hello")},
			}
			result, err := handler.CallHandler(args, nil)
			convey.So(err, convey.ShouldBeNil)
			var res *types.CallResponse
			_ = json.Unmarshal(result, &res)
			convey.So(strings.Contains(string(res.Body), "unmarshal invoke call request data"), convey.ShouldEqual, true)
			convey.So(res.InnerCode, convey.ShouldEqual, "500")
		})
	})
}

func createLargeBytes(size int32) []byte {
	arg := bytes.Buffer{}
	arg.Grow(int(size))
	for arg.Len() < int(size) {
		arg.WriteString("max limit")
	}
	return arg.Bytes()
}

func TestCallHandler_OK(t *testing.T) {
	handler := NewEventHandler(newFuncSpec(), nil)
	patches := []*Patches{
		ApplyMethod(reflect.TypeOf(&EventHandler{}), "SetHandler",
			func(eh *EventHandler, functionLibPath, handler string) error {
				if handler == initHandlerName {
					eh.userInitEntry = func(ctx context.RuntimeContext) {
						fmt.Println("userCode init start")
						fmt.Println("userCode init success")
					}
				} else if handler == callHandlerName {
					eh.userCallEntry = func(payload []byte, ctx context.RuntimeContext) (interface{}, error) {
						logger := ctx.GetLogger()
						if logger == nil {
							return nil, fmt.Errorf("user logger not initialized")
						}
						log.GetLogger().Infof("payload:%s", string(payload))
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
	os.Setenv("RUNTIME_LOG_DIR", "./")
	args, _ := getInitArgs(createParams)
	_, err := handler.InitHandler([]api.Arg{
		{
			Type: api.Value,
			Data: []byte(""),
		},
		args[0],
		{
			Type: api.Value,
			Data: []byte(""),
		},
		{
			Type: api.Value,
			Data: []byte(""),
		},
	}, nil)
	assert.Nil(t, err)
	result, err := handler.CallHandler(getCallArgs(), nil)
	assert.Nil(t, err)
	var res *types.CallResponse
	_ = json.Unmarshal(result, &res)
	assert.Contains(t, string(res.Body), "success")
}

func TestSetHandler(t *testing.T) {
	convey.Convey("SetHandler", t, func() {
		convey.Convey("getLib error: invalid handler name", func() {
			handler := &EventHandler{}
			err := handler.SetHandler("/xxx", "callHandlerName")
			convey.So(err.Error(), convey.ShouldEqual, "getLib error: invalid handler name :callHandlerName")
		})

		convey.Convey("getLib error: failed to open lib", func() {
			handler := &EventHandler{}
			err := handler.SetHandler("/xxx", "test.callHandlerName")
			convey.So(err.Error(), convey.ShouldEqual, "getLib error: failed to open test.callHandlerName")
		})

		convey.Convey("getLib error: failed to look up", func() {
			defer ApplyFunc(plugin.Open, func(path string) (*plugin.Plugin, error) {
				return &plugin.Plugin{}, nil
			}).Reset()
			handler := &EventHandler{libMap: make(map[string]*plugin.Plugin)}
			err := handler.SetHandler("/xxx", "test.callHandlerName")
			convey.So(err.Error(), convey.ShouldNotBeNil)
		})

		convey.Convey("type error", func() {
			defer ApplyFunc(plugin.Open, func(path string) (*plugin.Plugin, error) {
				return &plugin.Plugin{}, nil
			}).Reset()
			defer ApplyMethod(reflect.TypeOf(&plugin.Plugin{}), "Lookup",
				func(_ *plugin.Plugin, symName string) (plugin.Symbol, error) {
					return "type error", nil
				}).Reset()
			handler := &EventHandler{libMap: make(map[string]*plugin.Plugin)}
			err := handler.SetHandler("/xxx", "test.callHandlerName")
			convey.So(err.Error(), convey.ShouldNotBeNil)
		})

		convey.Convey("success", func() {
			defer ApplyFunc(plugin.Open, func(path string) (*plugin.Plugin, error) {
				return &plugin.Plugin{}, nil
			}).Reset()
			defer ApplyMethod(reflect.TypeOf(&plugin.Plugin{}), "Lookup",
				func(_ *plugin.Plugin, symName string) (plugin.Symbol, error) {
					var symbol func([]byte, context.RuntimeContext) (interface{}, error)
					symbol = func(i []byte, runtimeContext context.RuntimeContext) (interface{}, error) {
						return "success", nil
					}
					return symbol, nil
				}).Reset()
			handler := &EventHandler{libMap: make(map[string]*plugin.Plugin)}
			err := handler.SetHandler("/xxx", "test.callHandlerName")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCheckPointHandler(t *testing.T) {
	handler := NewEventHandler(newFuncSpec(), nil)
	pointHandler, err := handler.CheckPointHandler("checkPointID")
	assert.Nil(t, pointHandler)
	assert.Nil(t, err)
}

func TestRecoverHandler(t *testing.T) {
	handler := NewEventHandler(newFuncSpec(), nil)
	err := handler.RecoverHandler([]byte{})
	assert.Nil(t, err)
}

// create args and client, to invoke InvokeHandlerService
func getInitArgs(param string) ([]api.Arg, api.LibruntimeAPI) {
	err := os.Setenv(config.DelegateDownloadPath, userCodePath)
	if err != nil {
		return nil, nil
	}
	args := make([]api.Arg, 2, 2)
	args[0] = api.Arg{
		Type: 0,
		Data: []byte(param),
	}
	client := &mockLibruntimeClient{}
	return args, client
}

// create args to invoke CallHandler
func getCallArgs() []api.Arg {
	args := make([]api.Arg, 2)
	argsData0 := "This is payload argsData0 for userHandler."
	argsData1 := make(map[string]interface{}, 2)
	argsData1["headers"] = make(map[string]string)
	argsData1["body"] = []byte("This is payload argsData1 for userHandler.")
	args1, _ := json.Marshal(argsData1)
	args[0] = api.Arg{
		Type: 0,
		Data: []byte(argsData0),
	}
	args[1] = api.Arg{
		Type: 0,
		Data: args1,
	}
	return args
}

func Test_initContext(t *testing.T) {
	convey.Convey("initContext", t, func() {
		convey.Convey("init user function log error", func() {
			defer ApplyFunc(log.GetUserLogger, func() *log.UserFunctionLogger {
				return nil
			}).Reset()
			err := initContext()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("", func() {
			defer ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
				return errors.New("json unmarshal error")
			}).Reset()
			err := initContext()
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("success", func() {
			os.Setenv("RUNTIME_TIMEOUT", "10")
			os.Setenv("RUNTIME_PROJECT_ID", "123456789")
			os.Setenv("RUNTIME_PACKAGE", "service")
			os.Setenv("RUNTIME_FUNC_NAME", "test-go-runtime")
			os.Setenv("RUNTIME_FUNC_VERSION", "latest")
			os.Setenv("RUNTIME_MEMORY", "500")
			os.Setenv("RUNTIME_CPU", "500")
			os.Setenv("RUNTIME_USERDATA", `{"name":"runtime-go"}`)
			log.SetupUserLogger(userLogLevel)
			err := initContext()
			convey.So(err, convey.ShouldBeNil)
			convey.So(defaultContext.GetRunningTimeInSeconds(), convey.ShouldEqual, 10)
			convey.So(defaultContext.GetRemainingTimeInMilliSeconds() <= int(10*time.Millisecond), convey.ShouldEqual, true)
			convey.So(defaultContext.GetProjectID(), convey.ShouldEqual, "123456789")
			convey.So(defaultContext.GetPackage(), convey.ShouldEqual, "service")
			convey.So(defaultContext.GetFunctionName(), convey.ShouldEqual, "test-go-runtime")
			convey.So(defaultContext.GetVersion(), convey.ShouldEqual, "latest")
			convey.So(defaultContext.GetCPUNumber(), convey.ShouldEqual, 500)
			convey.So(defaultContext.GetMemorySize(), convey.ShouldEqual, 500)
			convey.So(defaultContext.GetUserData("name"), convey.ShouldEqual, "runtime-go")
			convey.So(defaultContext.GetAccessKey(), convey.ShouldEqual, "")
			convey.So(defaultContext.GetSecretKey(), convey.ShouldEqual, "")
			convey.So(defaultContext.GetToken(), convey.ShouldEqual, "")
			convey.So(defaultContext.GetRequestID(), convey.ShouldEqual, "")
			defaultContext.SetState("new state")
			convey.So(defaultContext.GetState(), convey.ShouldEqual, "new state")
			convey.So(defaultContext.GetInvokeProperty(), convey.ShouldEqual, "")
			convey.So(defaultContext.GetTraceID(), convey.ShouldEqual, "")
			convey.So(defaultContext.GetInvokeID(), convey.ShouldEqual, "")
			convey.So(defaultContext.GetAlias(), convey.ShouldEqual, "")
			convey.So(defaultContext.GetSecurityToken(), convey.ShouldEqual, "")
			defaultContext.GetLogger().Infof("info log")
			defaultContext.GetLogger().Warnf("warn log")
			defaultContext.GetLogger().Debugf("debug log")
			defaultContext.GetLogger().Errorf("error log")
		})
	})
}

func TestSetUserEntry(t *testing.T) {
	cleanFile("user-function")
	convey.Convey(
		"Test SetUserEntry", t, func() {
			convey.Convey("SetUserInitEntry success", func() {
				initE := func(ctx context.RuntimeContext) {}
				convey.So(func() {
					SetUserInitEntry(initE)
				}, convey.ShouldNotPanic)
			})
			convey.Convey("SetUserCallEntry  success", func() {
				callE := func([]byte, context.RuntimeContext) (interface{}, error) {
					return nil, nil
				}
				convey.So(func() {
					SetUserCallEntry(callE)
				}, convey.ShouldNotPanic)
			})
		},
	)
}

func TestEventHandler(t *testing.T) {
	convey.Convey(
		"Test EventHandler", t, func() {
			eh := NewEventHandler(newFuncSpec(), nil)
			convey.Convey("ShutDownHandler success", func() {
				err := eh.ShutDownHandler(1)
				convey.So(err, convey.ShouldBeNil)
			})
			convey.Convey("SignalHandler success", func() {
				err := eh.SignalHandler(1, []byte{0})
				convey.So(err, convey.ShouldBeNil)
			})
			convey.Convey("HealthCheckHandler success", func() {
				ht, err := eh.HealthCheckHandler()
				convey.So(ht, convey.ShouldNotBeNil)
				convey.So(err, convey.ShouldBeNil)
			})
			convey.Convey("setHandlerByEntry success", func() {
				err := eh.(*EventHandler).setHandlerByEntry("symbol", "handler")
				convey.So(err, convey.ShouldNotBeNil)
				symbol1 := func(ctx context.RuntimeContext) {}
				eh.(*EventHandler).userInitEntry = nil
				err = eh.(*EventHandler).setHandlerByEntry(symbol1, "handler")
				convey.So(err, convey.ShouldBeNil)
				symbol2 := func([]byte, context.RuntimeContext) (interface{}, error) {
					return nil, nil
				}
				eh.(*EventHandler).userCallEntry = nil
				err = eh.(*EventHandler).setHandlerByEntry(symbol2, "handler")
				convey.So(err, convey.ShouldBeNil)
			})
		},
	)
}

func cleanFile(fileName string) {
	files, _ := os.ReadDir("./")
	for _, file := range files {
		flag := strings.HasPrefix(file.Name(), fileName)
		if flag {
			os.Remove(file.Name())
		}
	}
}
