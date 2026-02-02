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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	FakeHTTP "github.com/stretchr/testify/http"

	"yuanrong.org/kernel/runtime/faassdk/common/alarm"
	"yuanrong.org/kernel/runtime/faassdk/common/faasscheduler"
	"yuanrong.org/kernel/runtime/faassdk/common/monitor"
	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/faassdk/handler/http/crossclusterinvoke"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/faassdk/utils/urnutils"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/uuid"
)

type fakeSnError struct {
	ErrorCode    int
	ErrorMessage string
}

// Code Returned error code
func (s *fakeSnError) Code() int {
	return s.ErrorCode
}

// Error Implement the native error interface.
func (s *fakeSnError) Error() string {
	return s.ErrorMessage
}

type fakeSDKClient struct {
	returnErr bool
}

func (f *fakeSDKClient) CreateInstance(funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	InstanceID := uuid.New().String()
	return InstanceID, nil
}

func (f *fakeSDKClient) InvokeByInstanceId(funcMeta api.FunctionMeta, instanceID string, args []api.Arg, invokeOpt api.InvokeOptions) (returnObjectID string, err error) {
	//TODO implement me
	return "", nil
}

func (f *fakeSDKClient) InvokeByFunctionName(funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	return "success", nil
}

func (f *fakeSDKClient) AcquireInstance(state string, funcMeta api.FunctionMeta, acquireOpt api.InvokeOptions) (api.InstanceAllocation, error) {
	return api.InstanceAllocation{}, nil
}

func (f *fakeSDKClient) ReleaseInstance(allocation api.InstanceAllocation, stateID string, abnormal bool, option api.InvokeOptions) {
	//TODO implement me
	return
}

func (f *fakeSDKClient) Kill(instanceID string, signal int, payload []byte) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) CreateInstanceRaw(createReqRaw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) InvokeByInstanceIdRaw(invokeReqRaw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) KillRaw(killReqRaw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) Finalize() {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) KVSet(key string, value []byte, param api.SetParam) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) KVSetWithoutKey(value []byte, param api.SetParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) KVMSetTx(keys []string, values [][]byte, param api.MSetParam) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) KVGet(key string, timeoutms uint) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) KVGetMulti(keys []string, timeoutms uint) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) KVQuerySize(keys []string) ([]uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) KVDel(key string) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) KVDelMulti(keys []string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) CreateProducer(streamName string, producerConf api.ProducerConf) (api.StreamProducer, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) Subscribe(streamName string, config api.SubscriptionConfig) (api.StreamConsumer, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) DeleteStream(streamName string) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) QueryGlobalProducersNum(streamName string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) QueryGlobalConsumersNum(streamName string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) SetTraceID(traceID string) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) SetTenantID(tenantID string) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) Put(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) PutRaw(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) Get(objectIDs []string, timeoutMs int) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) GetRaw(objectIDs []string, timeoutMs int) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) Wait(objectIDs []string, waitNum uint64, timeoutMs int) ([]string, []string, map[string]error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) GIncreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) GIncreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) GDecreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return nil, nil
}

func (f *fakeSDKClient) GDecreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	panic("implement me")
}

func (f *fakeSDKClient) GetAsync(objectID string, cb api.GetAsyncCallback) {
	cb([]byte("success"), nil)
}

func (f *fakeSDKClient) GetEvent(objectID string, cb api.GetEventCallback) {
	cb([]byte("success"), nil)
}

func (f *fakeSDKClient) DeleteGetEvent(objectID string) {

}

func (f *fakeSDKClient) GetFormatLogger() api.FormatLogger {
	//TODO implement me
	panic("implement me")
}

func (f *fakeSDKClient) CreateClient(config api.ConnectArguments) (api.KvClient, error) {
	return &FakeDataSystemClinet{}, nil
}

func (f *fakeSDKClient) SaveState(state []byte) (string, error) {
	return "", nil
}

func (f *fakeSDKClient) LoadState(checkpointID string) ([]byte, error) {
	return nil, nil
}

func (f *fakeSDKClient) Exit(code int, message string) {
	return
}

func (f *fakeSDKClient) GetCredential() api.Credential {
	return api.Credential{}
}

func (f *fakeSDKClient) UpdateSchdulerInfo(schedulerName string, schedulerId string, option string) {
	return
}

func (f *fakeSDKClient) IsHealth() bool {
	return true
}

func (f *fakeSDKClient) IsDsHealth() bool {
	return true
}

func newFuncSpec() *types.FuncSpec {
	return &types.FuncSpec{
		FuncMetaData: types.FuncMetaData{
			FunctionName:       "test-future-fuction",
			Runtime:            "go1.x",
			TenantId:           "123456789",
			Version:            "$latest",
			Timeout:            3,
			FunctionVersionURN: "sn:cn:yrk:12345678901234561234567890123456:function:0@yrservice@test-image-env",
		},
		ResourceMetaData: types.ResourceMetaData{},
		ExtendedMetaData: types.ExtendedMetaData{
			Initializer: types.Initializer{Timeout: 3},
			LogTankService: types.LogTankService{
				GroupID:  "gid",
				StreamID: "sid",
			},
			CustomGracefulShutdown: types.CustomGracefulShutdown{
				MaxShutdownTimeout: 5,
			},
		},
	}
}

func newFuncSpecWithHealthCheck() *types.FuncSpec {
	return &types.FuncSpec{
		FuncMetaData: types.FuncMetaData{
			FunctionName:       "test-future-fuction",
			Runtime:            "go1.x",
			TenantId:           "123456789",
			Version:            "$latest",
			Timeout:            3,
			FunctionVersionURN: "sn:cn:yrk:12345678901234561234567890123456:function:0@yrservice@test-image-env",
		},
		ResourceMetaData: types.ResourceMetaData{},
		ExtendedMetaData: types.ExtendedMetaData{
			Initializer: types.Initializer{Timeout: 3},
			CustomHealthCheck: types.CustomHealthCheck{
				TimeoutSeconds:   30,
				PeriodSeconds:    5,
				FailureThreshold: 1,
			},
		},
	}
}

func newFuncSpecWithoutTimeout() *types.FuncSpec {
	return &types.FuncSpec{
		FuncMetaData: types.FuncMetaData{
			FunctionName:       "test-future-fuction",
			Runtime:            "go1.x",
			TenantId:           "123456789",
			Version:            "$latest",
			Timeout:            0,
			FunctionVersionURN: "sn:cn:yrk:12345678901234561234567890123456:function:0@yrservice@test-image-env",
		},
		ResourceMetaData: types.ResourceMetaData{},
		ExtendedMetaData: types.ExtendedMetaData{
			Initializer: types.Initializer{Timeout: 0},
			LogTankService: types.LogTankService{
				GroupID:  "gid",
				StreamID: "sid",
			},
			CustomGracefulShutdown: types.CustomGracefulShutdown{
				MaxShutdownTimeout: 5,
			},
		},
	}
}

func newHttpCreateParams() *types.HttpCreateParams {
	return &types.HttpCreateParams{
		Port:      33333,
		InitRoute: "initRoute",
		CallRoute: "callRote",
	}
}

func newHttpSchedulerParams() *faasscheduler.SchedulerInfo {
	return &faasscheduler.SchedulerInfo{
		SchedulerFuncKey: "scheduler func key",
		SchedulerIDList:  []string{"1111"},
	}
}

func TestNewCustomContainerHandler(t *testing.T) {
	convey.Convey("NewCustomContainerHandler", t, func() {
		handler := NewCustomContainerHandler(newFuncSpec(), nil)
		convey.So(handler, convey.ShouldNotBeNil)
	})
}

func TestHandleGetFutureResponse(t *testing.T) {
	convey.Convey("Test Handle GetFutureResponse", t, func() {
		convey.Convey("call response unmarshal error", func() {
			response := &types.GetFutureResponse{}
			handleGetFutureResponse([]byte("result"), response, false)
			convey.So(response.StatusCode, convey.ShouldEqual, constants.FaaSError)
		})
		convey.Convey("failed to get the innerCode", func() {
			callResponse := &types.CallResponse{
				InnerCode: "aaa",
			}
			result, _ := json.Marshal(callResponse)
			response := &types.GetFutureResponse{}
			handleGetFutureResponse(result, response, false)
			convey.So(response.StatusCode, convey.ShouldEqual, constants.FaaSError)
		})
		convey.Convey("inner code not zero", func() {
			callResponse := &types.CallResponse{
				InnerCode: "1",
			}
			result, _ := json.Marshal(callResponse)
			response := &types.GetFutureResponse{}
			handleGetFutureResponse(result, response, false)
			convey.So(response.StatusCode, convey.ShouldEqual, 1)
		})
		convey.Convey("unmarshal http response error", func() {
			callResponse := &types.CallResponse{
				InnerCode: "0",
				Body:      []byte("aaa"),
			}
			result, _ := json.Marshal(callResponse)
			response := &types.GetFutureResponse{}
			handleGetFutureResponse(result, response, false)
			convey.So(response.StatusCode, convey.ShouldEqual, constants.FaaSError)
		})
		convey.Convey("response ok", func() {
			apigResponse := APIGTriggerResponse{
				IsBase64Encoded: true,
				Body:            "b2s=",
				StatusCode:      http.StatusOK,
			}
			body, _ := json.Marshal(apigResponse)
			callResponse := &types.CallResponse{
				InnerCode: "0",
				Body:      body,
			}
			result, _ := json.Marshal(callResponse)
			response := &types.GetFutureResponse{}
			handleGetFutureResponse(result, response, false)
			convey.So(response.Content, convey.ShouldEqual, "ok")
		})
		convey.Convey("response not ok", func() {
			apigResponse := APIGTriggerResponse{
				IsBase64Encoded: true,
				Body:            "b2s=",
				StatusCode:      http.StatusBadRequest,
			}
			body, _ := json.Marshal(apigResponse)
			callResponse := &types.CallResponse{
				InnerCode: "0",
				Body:      body,
			}
			result, _ := json.Marshal(callResponse)
			response := &types.GetFutureResponse{}
			handleGetFutureResponse(result, response, false)
			convey.So(response.StatusCode, convey.ShouldEqual, constants.FunctionRunError)
		})
		convey.Convey("response base64 decode failed", func() {
			apigResponse := APIGTriggerResponse{
				IsBase64Encoded: true,
				Body:            "????",
				StatusCode:      http.StatusOK,
			}
			body, _ := json.Marshal(apigResponse)
			callResponse := &types.CallResponse{
				InnerCode: "0",
				Body:      body,
			}
			result, _ := json.Marshal(callResponse)
			response := &types.GetFutureResponse{}
			handleGetFutureResponse(result, response, false)
			convey.So(response.StatusCode, convey.ShouldEqual, constants.FaaSError)
		})
	})
}

func TestHandleInvoke(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
			config:    &config.Configuration{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}
	convey.Convey("Test Handle Invoke", t, func() {
		convey.Convey("read request body error", func() {
			patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
				return []byte{}, errors.New("read error")
			})
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(w.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
		convey.Convey("unmarshal request body error", func() {
			patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
				return []byte("aaa"), nil
			})
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(w.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
		convey.Convey("process invoke error", func() {
			patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
				invokeRequest := types.InvokeRequest{}
				data, _ := json.Marshal(invokeRequest)
				return data, nil
			})
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(w.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
		convey.Convey("process invoke success", func() {
			patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
				invokeRequest := types.InvokeRequest{}
				data, _ := json.Marshal(invokeRequest)
				return data, nil
			})
			faasscheduler.Proxy.Add("faasScheduler1")
			defer patch.Reset()
			defer faasscheduler.Proxy.Remove("faasScheduler1")
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(w.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
		convey.Convey("process invoke kernel error", func() {
			ch.sdkClient = &fakeSDKClient{returnErr: true}
			defer gomonkey.ApplyMethod(reflect.TypeOf(ch.sdkClient), "InvokeByFunctionName",
				func(_ *fakeSDKClient, funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
					return "123", errors.New("kernel error")
				}).Reset()
			patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
				invokeRequest := types.InvokeRequest{}
				data, _ := json.Marshal(invokeRequest)
				return data, nil
			})
			faasscheduler.Proxy.Add("faasScheduler1")
			defer patch.Reset()
			defer faasscheduler.Proxy.Remove("faasScheduler1")
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			for id, _ := range ch.futureMap {
				rep := <-ch.futureMap[id]
				convey.So(rep, convey.ShouldNotBeNil)
			}
		})
	})
}

func TestHandleGetFuture(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}
	convey.Convey("Test Handle Get Future", t, func() {
		convey.Convey("read request body error", func() {
			patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
				return []byte{}, errors.New("read error")
			})
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleGetFuture(w, r)
			convey.So(w.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
		convey.Convey("unmarshal request body error", func() {
			patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
				return []byte("aaa"), nil
			})
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleGetFuture(w, r)
			convey.So(w.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
		convey.Convey("process futureCh not exit", func() {
			patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
				invokeRequest := types.InvokeRequest{}
				data, _ := json.Marshal(invokeRequest)
				return data, nil
			})
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleGetFuture(w, r)
			convey.So(w.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
	})
}

func TestInitHandler(t *testing.T) {
	convey.Convey("InitHandler", t, func() {
		handler := NewCustomContainerHandler(newFuncSpec(), nil)
		convey.So(handler, convey.ShouldNotBeNil)
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFunc(monitor.CreateFunctionMonitor, func(funcSpec *types.FuncSpec, stopCh chan struct{}) (*monitor.FunctionMonitorManager, error) {
				return &monitor.FunctionMonitorManager{}, nil
			}),
			gomonkey.ApplyFunc(http.HandleFunc, func(pattern string, handler func(http.ResponseWriter, *http.Request)) {
				return
			}),
			gomonkey.ApplyMethod(reflect.TypeOf(&http.Server{}), "ListenAndServe", func(_ *http.Server) error {
				return nil
			}),
		}
		defer func() {
			for _, patch := range patches {
				patch.Reset()
			}
		}()
		convey.Convey("failed to parse create params: invalid create params number", func() {
			args := []api.Arg{}
			res, err := handler.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldNotBeEmpty)
			convey.So(res, convey.ShouldEqual, []byte{})
		})

		convey.Convey("failed to parse create params: failed to unmarshal funcSpec from", func() {
			args := []api.Arg{
				{
					Type: 0,
					Data: []byte("invalid json"),
				},
				{
					Type: 0,
					Data: []byte{},
				},
			}
			res, err := handler.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldNotBeEmpty)
			convey.So(res, convey.ShouldEqual, []byte{})
		})
		convey.Convey("test initExecuteWithTimeout", func() {
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
					Type: 0,
					Data: []byte(""),
				},
				{
					Type: 0,
					Data: []byte(""),
				},
			}
			convey.Convey("failed to test ExecuteTimeout cause timeout", func() {
				res, err := handler.InitHandler(args, nil)
				convey.So(err.Error(), convey.ShouldNotBeEmpty)
				convey.So(res, convey.ShouldEqual, []byte{})
			})
			convey.Convey("succeed to test ExecuteTimeout", func() {
				defer gomonkey.ApplyFunc(net.Dial, func(_, _ string) (net.Conn, error) {
					n := tls.Conn{}
					return &n, nil
				}).Reset()
				defer gomonkey.ApplyFunc((*tls.Conn).Close, func(_ *tls.Conn) error {
					return nil
				}).Reset()
				res, err := handler.InitHandler(args, nil)
				convey.So(err.Error(), convey.ShouldNotBeEmpty)
				convey.So(res, convey.ShouldBeEmpty)
			})
		})

		convey.Convey("failed to create request to init route error", func() {
			defer gomonkey.ApplyFunc(http.NewRequestWithContext,
				func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
					return nil, errors.New("NewRequestWithContext error")
				}).Reset()
			defer gomonkey.ApplyFunc((*CustomContainerHandler).awaitReady, func(_ *CustomContainerHandler, timeout time.Duration) error {
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
					Type: 0,
					Data: []byte(""),
				},
				{
					Type: 0,
					Data: []byte(""),
				},
			}
			res, err := handler.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldNotBeEmpty)
			convey.So(res, convey.ShouldBeEmpty)
		})

		convey.Convey("init failed request timed out after 3s", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do",
				func(_ *http.Client, req *http.Request) (*http.Response, error) {
					time.Sleep(3*time.Second + 10*time.Millisecond)
					return nil, errors.New("timeout 3s")
				}).Reset()
			defer gomonkey.ApplyFunc((*CustomContainerHandler).awaitReady, func(_ *CustomContainerHandler, timeout time.Duration) error {
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
					Type: 0,
					Data: []byte(""),
				},
				{
					Type: 0,
					Data: []byte(""),
				},
			}
			res, err := handler.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldNotBeEmpty)
			convey.So(res, convey.ShouldBeEmpty)
		})

		convey.Convey("init failed request error", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do",
				func(_ *http.Client, req *http.Request) (*http.Response, error) {
					return nil, errors.New("mock init request error")
				}).Reset()
			defer gomonkey.ApplyFunc((*CustomContainerHandler).awaitReady, func(_ *CustomContainerHandler, timeout time.Duration) error {
				return nil
			}).Reset()
			handlerWithHealthCheck := NewCustomContainerHandler(newFuncSpecWithHealthCheck(), &fakeSDKClient{})
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
					Type: 0,
					Data: []byte(""),
				},
				{
					Type: 0,
					Data: []byte(""),
				},
			}
			res, err := handlerWithHealthCheck.InitHandler(args, nil)
			convey.So(err.Error(), convey.ShouldNotBeEmpty)
			convey.So(res, convey.ShouldBeEmpty)
		})
	})
}

func TestShutDownHander(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		invokeServer:        &http.Server{},
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}
	convey.Convey("Test ShutDown Hander", t, func() {
		convey.Convey("shutdown success", func() {
			h := &basicHandler{}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(h),
				"ShutDownHandler", func(b *basicHandler, gracePeriodSecond uint64) error {
					return nil
				})
			defer patch.Reset()
			err := ch.ShutDownHandler(30)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetCallRequestBody(t *testing.T) {
	request := types.InvokeRequest{
		Payload: "hello",
	}
	convey.Convey("Test GetCallRequestBody", t, func() {
		convey.Convey("disableAPIGFormat is ture", func() {
			requestData, err := getCallRequestBody(request, true)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(requestData), convey.ShouldEqual, "hello")
		})
		convey.Convey("disableAPIGFormat is false", func() {
			requestData, err := getCallRequestBody(request, false)
			APIGData := APIGTriggerEvent{}
			_ = json.Unmarshal(requestData, &APIGData)
			convey.So(err, convey.ShouldBeNil)
			convey.So(APIGData.Body, convey.ShouldEqual, "hello")
		})
	})
}

func TestHandleResponse(t *testing.T) {
	convey.Convey("Test HandleResponse", t, func() {
		convey.Convey("json marshal error", func() {
			patch := gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return []byte{}, errors.New("marshal error")
			})
			defer patch.Reset()
			writer := &fakeWriter{}
			handleResponse(&types.GetFutureResponse{}, writer, "test")
			convey.So(writer.StatusCode, convey.ShouldEqual, http.StatusInternalServerError)
		})
		convey.Convey("Write error", func() {
			writer := &fakeWriter{}
			response := &types.GetFutureResponse{
				StatusCode: constants.NoneError,
			}
			handleResponse(response, writer, "test")
			convey.So(response.GetStatusCode(), convey.ShouldEqual, constants.NoneError)
		})
	})
}

type fakeWriter struct {
	StatusCode int
}

func (w *fakeWriter) Header() http.Header {
	return nil
}

func (w *fakeWriter) Write([]byte) (int, error) {
	return 0, errors.New("write error")
}

func (w *fakeWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func TestCustomGracefulShutdownNewVer(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		invokeServer:        &http.Server{},
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}
	convey.Convey("test customGracefulShutdownNewVer", t, func() {
		convey.Convey("test stop by cust-image", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
				return &http.Response{
					StatusCode: 200,
				}, nil
			}).Reset()
			convey.So(func() {
				ch.customGracefulShutdownNewVer(1 * time.Second)
			}, convey.ShouldNotPanic)
		})

		convey.Convey("test stop by timeout", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&http.Client{}), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
				t.Logf("sleep 2s")
				time.Sleep(2 * time.Second)
				return &http.Response{
					StatusCode: 200,
				}, nil
			}).Reset()
			convey.So(func() {
				ch.customGracefulShutdownNewVer(1 * time.Second)
			}, convey.ShouldNotPanic)
		})
	})
}

func TestProcessState(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		invokeServer:        &http.Server{},
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}

	req := types.StateRequest{
		FuncName:    "test-func-name",
		FuncVersion: "test-func-version",
		Params:      nil,
		StateID:     "test-state-id" + uuid.New().String(),
		TraceID:     "test-trace-id-1",
	}

	convey.Convey("func urn is invalid", t, func() {
		defer gomonkey.ApplyFunc(urnutils.GetFunctionInfo, func(urn string) (urnutils.BaseURN, error) {
			return urnutils.BaseURN{}, errors.New("mock urn error")
		}).Reset()
		err := ch.processState(req, "new")
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("state id is empty", t, func() {
		req.StateID = ""
		err := ch.processState(req, "new")
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("stateMgr cannot be obtained", t, func() {
		ch.stateMgrState = failed
		err := ch.processState(req, "new")
		convey.So(err, convey.ShouldNotBeNil)
	})

	ch.stateMgrState = completed
	ch.stateMgr = &stateManager{}
	req.StateID = "test-state-id"

	convey.Convey("new state success", t, func() {
		defer gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch.stateMgr), "newState", func(_ *stateManager, funcKey string, stateID string, traceID string) error {
			return nil
		}).Reset()
		err := ch.processState(req, "new")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("new state fail due to stateMgr failure", t, func() {
		defer gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch.stateMgr), "newState", func(_ *stateManager, funcKey, stateID, traceID string) error {
			return errors.New("mock state mgr err")
		}).Reset()
		err := ch.processState(req, "new")
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("get state success", t, func() {
		defer gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch.stateMgr), "getState", func(_ *stateManager, funcKey, stateID, traceID string) error {
			return nil
		}).Reset()
		err := ch.processState(req, "get")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("get state fail due to stateMgr failure", t, func() {
		defer gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch.stateMgr), "getState", func(_ *stateManager, funcKey, stateID, traceID string) error {
			return errors.New("mock state mgr err")
		}).Reset()
		err := ch.processState(req, "get")
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("del state fail due to stateMgr failure", t, func() {
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch.stateMgr), "delState", func(_ *stateManager, funcKey, stateID, traceID string) error {
			return errors.New("mock state mgr err")
		})
		defer patch.Reset()
		err := ch.processState(req, "del")
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("del state fail due to faasscheduler failure", t, func() {
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch.stateMgr), "delState", func(_ *stateManager, funcKey, stateID, traceID string) error {
			return nil
		})
		defer patch.Reset()
		err := ch.processState(req, "del")
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("del state success", t, func() {
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch.stateMgr), "delState", func(_ *stateManager, funcKey, stateID, traceID string) error {
			return nil
		})
		faasscheduler.Proxy.Add("faasScheduler1")
		defer faasscheduler.Proxy.Remove("faasScheduler1")
		defer patch.Reset()
		err := ch.processState(req, "del")
		convey.So(err, convey.ShouldBeNil)
	})
}

func mockHttpWriter(rspCode *int, patch *gomonkey.Patches) {
	patch.ApplyMethod(reflect.TypeOf(&FakeHTTP.TestResponseWriter{}), "Write", func(_ *FakeHTTP.TestResponseWriter, data []byte) (int, error) {
		response := types.StateResponse{}
		err := json.Unmarshal(data, &response)
		if err != nil {
			*rspCode = -1
		}
		*rspCode = response.StatusCode
		return 0, nil
	})
}

func TestHandleStateNew(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		invokeServer:        &http.Server{},
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}

	convey.Convey("read req body error", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			return []byte{}, errors.New("read error")
		})
		rspCode := 0
		mockHttpWriter(&rspCode, patch)
		defer patch.Reset()

		w := &FakeHTTP.TestResponseWriter{}
		r := &http.Request{}
		ch.handleStateNew(w, r)
		convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
	})

	convey.Convey("unmarshal body failed", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			return []byte("aaa"), nil
		})
		rspCode := 0
		mockHttpWriter(&rspCode, patch)
		defer patch.Reset()
		w := &FakeHTTP.TestResponseWriter{}
		r := &http.Request{}
		ch.handleStateNew(w, r)
		convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
	})

	convey.Convey("read req body successfully", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			invokeRequest := types.InvokeRequest{}
			data, _ := json.Marshal(invokeRequest)
			return data, nil
		})
		defer patch.Reset()
		convey.Convey("processState return snErr", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return api.ErrorInfo{
					Code: StateInstanceNotExistedErrCode,
					Err:  fmt.Errorf("mock snerror"),
				}
			})

			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateNew(w, r)
			convey.So(rspCode, convey.ShouldEqual, StateInstanceNotExistedErrCode)
		})

		convey.Convey("processState return stateError", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return api.ErrorInfo{
					Code: InvalidState,
					Err:  fmt.Errorf("error"),
				}
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateNew(w, r)
			convey.So(rspCode, convey.ShouldEqual, InvalidState)
		})

		convey.Convey("processState return normal error", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return errors.New("mock normal error")
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateNew(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
		})

		convey.Convey("processState success", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return nil
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateNew(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.NoneError)
		})
	})
}

func TestHandleStateGet(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		invokeServer:        &http.Server{},
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}

	convey.Convey("read req body error", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			return []byte{}, errors.New("read error")
		})
		rspCode := 0
		mockHttpWriter(&rspCode, patch)
		defer patch.Reset()

		w := &FakeHTTP.TestResponseWriter{}
		r := &http.Request{}
		ch.handleStateGet(w, r)
		convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
	})

	convey.Convey("unmarshal body error", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			return []byte("aaa"), nil
		})
		rspCode := 0
		mockHttpWriter(&rspCode, patch)
		defer patch.Reset()
		w := &FakeHTTP.TestResponseWriter{}
		r := &http.Request{}
		ch.handleStateGet(w, r)
		convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
	})

	convey.Convey("read req body successfully", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			invokeRequest := types.InvokeRequest{}
			data, _ := json.Marshal(invokeRequest)
			return data, nil
		})
		defer patch.Reset()
		convey.Convey("processState return snErr", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return api.ErrorInfo{
					Code: StateInstanceNotExistedErrCode,
					Err:  fmt.Errorf("mock snerror"),
				}
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateGet(w, r)
			convey.So(rspCode, convey.ShouldEqual, StateInstanceNotExistedErrCode)
		})

		convey.Convey("processState return stateError", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return api.ErrorInfo{
					Code: InvalidState,
					Err:  fmt.Errorf("mock snerror"),
				}
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateGet(w, r)
			convey.So(rspCode, convey.ShouldEqual, InvalidState)
		})

		convey.Convey("processState return normal error", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return errors.New("mock normal error")
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateGet(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
		})

		convey.Convey("processState success", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return nil
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateGet(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.NoneError)
		})
	})
}

func TestHandleStateDel(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		invokeServer:        &http.Server{},
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}

	convey.Convey("read req body error", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			return []byte{}, errors.New("read error")
		})
		rspCode := 0
		mockHttpWriter(&rspCode, patch)
		defer patch.Reset()

		w := &FakeHTTP.TestResponseWriter{}
		r := &http.Request{}
		ch.handleStateDel(w, r)
		convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
	})

	convey.Convey("unmarshal body error", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			return []byte("aaa"), nil
		})
		rspCode := 0
		mockHttpWriter(&rspCode, patch)
		defer patch.Reset()
		w := &FakeHTTP.TestResponseWriter{}
		r := &http.Request{}
		ch.handleStateDel(w, r)
		convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
	})

	convey.Convey("read req body successfully", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			invokeRequest := types.InvokeRequest{}
			data, _ := json.Marshal(invokeRequest)
			return data, nil
		})
		defer patch.Reset()
		convey.Convey("processState return snErr", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return api.ErrorInfo{
					Code: StateInstanceNotExistedErrCode,
					Err:  fmt.Errorf("mock snerror"),
				}
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateDel(w, r)
			convey.So(rspCode, convey.ShouldEqual, StateInstanceNotExistedErrCode)
		})

		convey.Convey("processState return stateError", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return api.ErrorInfo{
					Code: InvalidState,
					Err:  fmt.Errorf("mock snerror"),
				}
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateDel(w, r)
			convey.So(rspCode, convey.ShouldEqual, InvalidState)
		})

		convey.Convey("processState return normal error", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return errors.New("mock normal error")
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateDel(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
		})

		convey.Convey("processState success", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(ch), "processState", func(_ *CustomContainerHandler, request types.StateRequest, opType string) error {
				return nil
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleStateDel(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.NoneError)
		})
	})
}

func TestInitDsClinet(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		invokeServer:        &http.Server{},
		remoteDsClient:      &FakeDataSystemClinet{},
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}

	convey.Convey("init sts success", t, func() {
		convey.Convey("create client fail", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&fakeSDKClient{}), "CreateClient",
				func(_ *fakeSDKClient, config api.ConnectArguments) (api.KvClient, error) {
					return &FakeDataSystemClinet{}, errors.New("mock create client error")
				}).Reset()
			err := ch.initDsClient(&fakeSDKClient{})
			t.Logf("error is %v", err)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("create client successufully", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&FakeDataSystemClinet{}), "CreateClient",
				func(_ *FakeDataSystemClinet, config api.ConnectArguments) (api.KvClient, error) {
					return &FakeDataSystemClinet{}, nil
				}).Reset()
			err := ch.initDsClient(&fakeSDKClient{})
			t.Logf("error is %v", err)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetStateMgr(t *testing.T) {
	ch := &CustomContainerHandler{}
	convey.Convey("stateMgrState is failed", t, func() {
		ch.stateMgrState = failed
		sm := ch.getStateMgr()
		convey.So(sm, convey.ShouldBeNil)
	})
	convey.Convey("stateMgrState is uninitialized", t, func() {
		ch.stateMgrState = uninitialized
		defer gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {
		}).Reset()
		convey.Convey("stateMgr is nil", func() {
			ch.stateMgr = nil
			sm := ch.getStateMgr()
			convey.So(sm, convey.ShouldBeNil)
		})
		convey.Convey("stateMgr is not nil", func() {
			ch.stateMgr = &stateManager{}
			sm := ch.getStateMgr()
			convey.So(sm, convey.ShouldNotBeNil)
		})
	})

}

func TestHandleInvokeStatePart(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{
			funcSpec:  newFuncSpec(),
			sdkClient: &fakeSDKClient{},
		},
		futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
		remoteDsClient:      &FakeDataSystemClinet{},
		crossClusterInvoker: &crossclusterinvoke.Invoker{},
	}
	ch.setStateMgr(GetStateManager(ch.remoteDsClient), completed)
	faasscheduler.Proxy.Add("faasScheduler1")
	defer faasscheduler.Proxy.Remove("faasScheduler1")

	convey.Convey("Test Handle Invoke by StateID", t, func() {
		patch := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
			invokeRequest := types.InvokeRequest{
				StateID: "test-state-id",
			}
			data, _ := json.Marshal(invokeRequest)
			return data, nil
		})
		rspCode := 0
		mockHttpWriter(&rspCode, patch)
		defer patch.Reset()
		convey.Convey("process invoke success", func() {
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.NoneError)
		})

		convey.Convey("func urn is invalid", func() {
			patch := gomonkey.ApplyFunc(urnutils.GetFunctionInfo, func(urn string) (urnutils.BaseURN, error) {
				return urnutils.BaseURN{}, errors.New("mock urn error")
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
		})

		convey.Convey("stateMgr is nil", func() {
			ch.stateMgrState = failed
			defer func() { ch.stateMgrState = completed }()
			patch := gomonkey.NewPatches()
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.FaaSError)
		})

		convey.Convey("invoke fail return normal error", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&fakeSDKClient{}), "InvokeByInstanceId", func(_ *fakeSDKClient, funcMeta api.FunctionMeta, instanceID string, args []api.Arg, invokeOpt api.InvokeOptions) (returnObjectID string, err error) {
				return "", errors.New("mock invoke err")
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.NoneError)
		})

		convey.Convey("invoke fail return SNError", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&fakeSDKClient{}), "InvokeByInstanceId", func(_ *fakeSDKClient, funcMeta api.FunctionMeta, instanceID string, args []api.Arg,
				invokeOpt api.InvokeOptions) (returnObjectID string, err error) {
				return "", api.ErrorInfo{Code: 4000, Err: fmt.Errorf("mock snerror")}
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.NoneError)
		})

		convey.Convey("acquire instance fail return normal error", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&fakeSDKClient{}), "AcquireInstance", func(_ *fakeSDKClient, state string, funcMeta api.FunctionMeta,
				acquireOpt api.InvokeOptions) (api.InstanceAllocation, error) {
				return api.InstanceAllocation{}, errors.New("mock AcquireInstance normal error")
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.NoneError)
		})

		convey.Convey("acquire instance fail return SNError but no faasscheduler error", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&fakeSDKClient{}), "AcquireInstance", func(_ *fakeSDKClient,
				state string, funcMeta api.FunctionMeta, acquireOpt api.InvokeOptions) (api.InstanceAllocation, error) {
				return api.InstanceAllocation{}, &fakeSnError{
					ErrorCode:    4999, // 4999 is not faasscheduler errorCodes
					ErrorMessage: "mock snerror",
				}
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(rspCode, convey.ShouldEqual, constants.NoneError)
		})

		convey.Convey("acquire instance fail return faas error", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&fakeSDKClient{}), "AcquireInstance", func(_ *fakeSDKClient,
				state string, funcMeta api.FunctionMeta, acquireOpt api.InvokeOptions) (api.InstanceAllocation, error) {
				return api.InstanceAllocation{}, api.ErrorInfo{
					Code: StateInstanceNotExistedErrCode,
					Err:  fmt.Errorf("mock snerror"),
				}
			})
			rspCode := 0
			mockHttpWriter(&rspCode, patch)
			defer patch.Reset()
			w := &FakeHTTP.TestResponseWriter{}
			r := &http.Request{}
			ch.handleInvoke(w, r)
			convey.So(rspCode, convey.ShouldEqual, StateInstanceNotExistedErrCode)
		})
	})
}

func TestCustomContainerHandler_checkRuntime(t *testing.T) {
	convey.Convey("check runtime test", t, func() {

		ch := &CustomContainerHandler{
			basicHandler: &basicHandler{
				funcSpec:  newFuncSpec(),
				sdkClient: &fakeSDKClient{},
			},
			futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
			remoteDsClient:      &FakeDataSystemClinet{},
			crossClusterInvoker: &crossclusterinvoke.Invoker{},
		}
		check := types.CustomHealthCheck{
			TimeoutSeconds:   1,
			PeriodSeconds:    5,
			FailureThreshold: 1,
		}
		convey.Convey("res is 500", func() {
			p := gomonkey.ApplyFunc((*basicHandler).sendRequest, func(_ *basicHandler,
				request *http.Request, timeout time.Duration) (*http.Response, error) {
				res := &http.Response{}
				res.StatusCode = http.StatusOK
				return res, nil
			})
			go func() {
				err := ch.checkRuntime(nil, check)
				assert.Equal(t, err.Error(), "check health res status code is 500")
			}()
			time.Sleep(2 * time.Second)
			p.Reset()
			p = gomonkey.ApplyFunc((*basicHandler).sendRequest, func(_ *basicHandler,
				request *http.Request, timeout time.Duration) (*http.Response, error) {
				res := &http.Response{}
				res.StatusCode = http.StatusInternalServerError
				return res, nil
			})
			time.Sleep(time.Duration(check.PeriodSeconds) * time.Second)
			p.Reset()
		})

		convey.Convey("err not nil", func() {
			p := gomonkey.ApplyFunc((*basicHandler).sendRequest, func(_ *basicHandler,
				request *http.Request, timeout time.Duration) (*http.Response, error) {
				return nil, fmt.Errorf("error")
			})
			defer p.Reset()
			err := ch.checkRuntime(nil, check)
			convey.So(err.Error(), convey.ShouldEqual, "error")
		})
		convey.Convey("fail fail success", func() {
			failCount := 2
			check.FailureThreshold = failCount + 1
			count := 0
			p := gomonkey.ApplyFunc((*basicHandler).sendRequest, func(_ *basicHandler,
				request *http.Request, timeout time.Duration) (*http.Response, error) {
				res := &http.Response{}
				if count < failCount {
					count++
					return nil, fmt.Errorf("error")
				}
				res.StatusCode = http.StatusOK
				return res, nil
			})
			defer func() {
				close(stopCh)
				time.Sleep(100 * time.Millisecond)
				stopCh = make(chan struct{})
				p.Reset()
			}()
			go func() {
				_ = ch.checkRuntime(nil, check)
			}()
			time.Sleep(20 * time.Second)
		})
	})
}

func TestCustomContainerHandler_checkHealth(t *testing.T) {
	convey.Convey("check health", t, func() {
		ch := &CustomContainerHandler{
			basicHandler: &basicHandler{
				funcSpec:  newFuncSpec(),
				sdkClient: &fakeSDKClient{},
			},
			futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
			remoteDsClient:      &FakeDataSystemClinet{},
			crossClusterInvoker: &crossclusterinvoke.Invoker{},
		}
		convey.Convey("check runtime health failed", func() {
			ch.funcSpec.ExtendedMetaData.CustomHealthCheck.TimeoutSeconds = 1
			ch.funcSpec.ExtendedMetaData.CustomHealthCheck.PeriodSeconds = 1
			ch.funcSpec.ExtendedMetaData.CustomHealthCheck.FailureThreshold = 1
			p := gomonkey.ApplyFunc((*CustomContainerHandler).checkRuntime, func(_ *CustomContainerHandler,
				request *http.Request, check types.CustomHealthCheck) error {
				return fmt.Errorf("error")
			})
			defer p.Reset()
			ch.checkHealth()
		})
	})
}

func TestCustomContainerHandler_handleExit(t *testing.T) {
	convey.Convey("handle exit test", t, func() {
		ch := &CustomContainerHandler{
			basicHandler: &basicHandler{
				funcSpec:  newFuncSpec(),
				sdkClient: &fakeSDKClient{},
			},
			futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
			remoteDsClient:      &FakeDataSystemClinet{},
			crossClusterInvoker: &crossclusterinvoke.Invoker{},
		}
		cnt := 0
		gomonkey.ApplyMethod(ch.sdkClient, "Exit", func(_ *fakeSDKClient, code int, message string) {
			if code != 0 {
				cnt++
			}
			return
		})
		convey.Convey("baseline", func() {
			p := gomonkey.ApplyFunc(handleResponse, func(response types.Response,
				w http.ResponseWriter, handleType string) {
				return
			})
			defer p.Reset()
			ch.handleExit(&FakeHTTP.TestResponseWriter{}, &http.Request{Body: io.NopCloser(strings.NewReader(""))})
		})
		convey.Convey("withexit", func() {
			p := gomonkey.ApplyFunc(handleResponse, func(response types.Response,
				w http.ResponseWriter, handleType string) {
				return
			})
			defer p.Reset()
			ch.handleExit(&FakeHTTP.TestResponseWriter{}, &http.Request{Body: io.NopCloser(strings.NewReader("{\"code\":3015, \"message\":\"\"}"))})
			time.Sleep(5 * time.Millisecond)
			convey.So(cnt, convey.ShouldEqual, 1)
		})
	})
}

type GetCredentialSuccessClient struct {
	fakeSDKClient
}

func (f *GetCredentialSuccessClient) GetCredential() api.Credential {
	return api.Credential{
		AccessKey: "ak",
		SecretKey: []byte("sk"),
		DataKey:   []byte("dk"),
	}
}

func TestCustomContainerHandler_handleCredentialRequire(t *testing.T) {
	ch := &CustomContainerHandler{
		basicHandler: &basicHandler{},
	}
	convey.Convey("handle credential require test", t, func() {
		convey.Convey("success", func() {
			ch.basicHandler.sdkClient = &GetCredentialSuccessClient{}
			w := &FakeHTTP.TestResponseWriter{}
			ch.handleCredentialRequire(w, nil)
			resp := &types.CredentialResponse{}
			_ = json.Unmarshal([]byte(w.Output), resp)
			convey.So(w.StatusCode, convey.ShouldEqual, http.StatusOK)
			convey.So(resp.StatusCode, convey.ShouldEqual, constants.NoneError)
			convey.So(resp.AccessKey, convey.ShouldEqual, "ak")
			convey.So(string(resp.SecretKey), convey.ShouldEqual, "sk")
			convey.So(string(resp.DataKey), convey.ShouldEqual, "dk")
		})
	})
}

func TestCustomContainerHandler_handleCircuitBreak(t *testing.T) {
	convey.Convey("handle circuit break test", t, func() {
		ch := &CustomContainerHandler{
			basicHandler: &basicHandler{
				funcSpec:  newFuncSpec(),
				sdkClient: &fakeSDKClient{},
			},
			futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
			remoteDsClient:      &FakeDataSystemClinet{},
			crossClusterInvoker: &crossclusterinvoke.Invoker{},
		}
		convey.Convey("baseline", func() {
			p := gomonkey.ApplyFunc(io.ReadAll, func(r io.Reader) ([]byte, error) {
				return []byte("{\"switch\":true}"), nil
			})
			p2 := gomonkey.ApplyFunc(handleResponse, func(response types.Response,
				w http.ResponseWriter, handleType string) {
				return
			})
			defer func() {
				p.Reset()
				p2.Reset()
			}()
			ch.handleCircuitBreak(nil, &http.Request{})
			convey.So(ch.circuitBreaker, convey.ShouldBeTrue)
		})
		convey.Convey("io read failed", func() {
			p := gomonkey.ApplyFunc(io.ReadAll, func(r io.Reader) ([]byte, error) {
				return nil, fmt.Errorf("error")
			})
			defer p.Reset()
			ch.circuitBreaker = false
			ch.handleCircuitBreak(&FakeHTTP.TestResponseWriter{}, &http.Request{})
			convey.So(ch.circuitBreaker, convey.ShouldBeFalse)
		})
		convey.Convey("json unmarshal failed", func() {
			p := gomonkey.ApplyFunc(io.ReadAll, func(r io.Reader) ([]byte, error) {
				return []byte("aaa"), nil
			})
			defer p.Reset()
			ch.circuitBreaker = true
			ch.handleCircuitBreak(&FakeHTTP.TestResponseWriter{}, &http.Request{})
			convey.So(ch.circuitBreaker, convey.ShouldBeTrue)
		})
	})
}

func TestCustomContainerHandler_HealthCheckHandler(t *testing.T) {
	convey.Convey("health check handler test", t, func() {
		ch := &CustomContainerHandler{
			basicHandler: &basicHandler{
				funcSpec:  newFuncSpec(),
				sdkClient: &fakeSDKClient{},
			},
			futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
			remoteDsClient:      &FakeDataSystemClinet{},
			crossClusterInvoker: &crossclusterinvoke.Invoker{},
		}
		convey.Convey("circuitBreaker is true", func() {
			ch.circuitBreaker = true
			res, err := ch.HealthCheckHandler()
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldEqual, api.SubHealth)
		})
		convey.Convey("circuitBreaker is false", func() {
			ch.circuitBreaker = false
			res, err := ch.HealthCheckHandler()
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldEqual, api.Healthy)
		})
	})
}

func TestCustomContainerHandler_setAlarmInfo(t *testing.T) {
	convey.Convey("set alarmInfo test", t, func() {
		ch := &CustomContainerHandler{
			basicHandler: &basicHandler{
				funcSpec:  newFuncSpec(),
				sdkClient: &fakeSDKClient{},
			},
			futureMap:           make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
			remoteDsClient:      &FakeDataSystemClinet{},
			crossClusterInvoker: &crossclusterinvoke.Invoker{},
		}
		convey.Convey("baseline", func() {
			ch.customUserArgs.AlarmConfig.EnableAlarm = true
			p := gomonkey.ApplyFunc(alarm.SetPodIP, func() error {
				return nil
			})
			defer p.Reset()
			err := ch.setAlarmInfo()
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("set podIp failed", func() {
			ch.customUserArgs.AlarmConfig.EnableAlarm = true
			p := gomonkey.ApplyFunc(alarm.SetPodIP, func() error {
				return fmt.Errorf("error")
			})
			defer p.Reset()
			err := ch.setAlarmInfo()
			convey.So(err.Error(), convey.ShouldEqual, "error")
		})
	})
}

func Test_checkTrafficLimitResp(t *testing.T) {
	convey.Convey("check traffic limit resp test", t, func() {
		convey.So(checkTrafficLimitResp([]byte("{\"errorCode\":0, \"errorMessage\": \"\"}")), convey.ShouldBeFalse)
		convey.So(checkTrafficLimitResp([]byte("{\"errorCode\":6037, \"errorMessage\": \"\"}")), convey.ShouldBeTrue)
		convey.So(checkTrafficLimitResp([]byte("aaa")), convey.ShouldBeFalse)
	})
}
