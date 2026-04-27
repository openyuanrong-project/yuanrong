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

// Package http handler
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/faassdk/common/aliasroute"
	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/common/faasscheduler"
	"yuanrong.org/kernel/runtime/faassdk/common/functionlog"
	"yuanrong.org/kernel/runtime/faassdk/common/monitor"
	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/faassdk/utils"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
	runtimeLog "yuanrong.org/kernel/runtime/libruntime/common/logger/log"
)

const (
	listenIP              = "127.0.0.1"
	mebibyte              = 1024 * 1024
	maxResponseSize       = 6 * mebibyte
	defaultTimeout        = 3
	healthRoute           = "healthz"
	killSignalAliasUpdate = 64
	schedulerArgIndex     = 2
	shutdownCheckInterval = 100 * time.Millisecond
	checkReadyInterval    = 10 * time.Millisecond
	argsMinLength         = 2
)

var (
	// ErrExecuteTimeout is the error of execute function timeout
	ErrExecuteTimeout = errors.New("execute function timeout")

	stopCh    = make(chan struct{})
	closeOnce sync.Once
)

type basicHandler struct {
	funcSpec          *types.FuncSpec
	sdkClient         api.LibruntimeAPI
	functionLogger    *functionlog.FunctionLogger
	logger            *RuntimeContainerLogger
	monitor           *monitor.FunctionMonitorManager
	config            *config.Configuration
	bootstrapFunc     func() error
	requestNum        int32
	baseURL           string
	initRoute         string
	callRoute         string
	healthRoute       string
	disableAPIGFormat bool
	port              int
	initTimeout       int
	callTimeout       int
	gracefulShutdown  bool
	circuitBreaker    bool
	circuitLock       sync.RWMutex
	once              sync.Once
	sync.RWMutex
}

func newBasicHandler(funcSpec *types.FuncSpec, client api.LibruntimeAPI) *basicHandler {
	rtmConf, err := config.GetConfig(funcSpec)
	if err != nil {
		logger.GetLogger().Errorf("failed to get runtime config for function %s error %s",
			funcSpec.FuncMetaData.FunctionName, err.Error())
		return nil
	}
	return &basicHandler{
		funcSpec:          funcSpec,
		sdkClient:         client,
		monitor:           &monitor.FunctionMonitorManager{},
		config:            rtmConf,
		gracefulShutdown:  false,
		disableAPIGFormat: os.Getenv("DISABLE_APIG_FORMAT") == "true",
	}
}

var transport *http.Transport

func (bh *basicHandler) getHTTPClient(timeout time.Duration) *http.Client {
	bh.once.Do(func() {
		defaultTransport := http.DefaultTransport
		var ok bool
		transport, ok = defaultTransport.(*http.Transport)
		if !ok {
			logger.GetLogger().Warnf("not a http transport type")
			transport = &http.Transport{}
		}
		// The server can close a previous "keep-alive" TCP connection (mainly because of a read timeout) at any time.
		// If we send a request while the connection is closing, the HTTP client will return an EOF error. As a result,
		// we disable keep-alive completely.
		transport.DisableKeepAlives = true
	})
	return &http.Client{
		Transport: transport,
		Timeout: timeout,
	}
}

func (bh *basicHandler) parseCreateParams(args []api.Arg) error {
	if len(args) != constants.ValidCustomImageCreateParamSize &&
		len(args) != constants.ValidBasicCreateParamSize {
		return errors.New("invalid create params number")
	}
	funcSpec := &types.FuncSpec{}
	err := json.Unmarshal(args[0].Data, funcSpec)
	if err != nil {
		logger.GetLogger().Errorf("failed to unmarshal funcSpec from %s", string(args[0].Data))
		return err
	}
	bh.initTimeout = funcSpec.ExtendedMetaData.Initializer.Timeout
	if bh.initTimeout <= 0 {
		bh.initTimeout = defaultTimeout
	}
	bh.callTimeout = funcSpec.FuncMetaData.Timeout
	if bh.callTimeout <= 0 {
		bh.callTimeout = defaultTimeout
	}
	createParams := &types.HttpCreateParams{}
	err = json.Unmarshal(args[1].Data, createParams)
	if err != nil {
		logger.GetLogger().Errorf("failed to unmarshal create params from %s", string(args[1].Data))
		return err
	}
	err = faasscheduler.ParseSchedulerData(args[schedulerArgIndex])
	if err != nil {
		return err
	}
	bh.port = createParams.Port
	bh.baseURL = fmt.Sprintf("http://%s:%d", listenIP, createParams.Port)
	bh.initRoute = createParams.InitRoute
	bh.callRoute = createParams.CallRoute
	bh.healthRoute = healthRoute
	return nil
}

func (bh *basicHandler) setBootstrapFunc(bootstrapFunc func() error) {
	bh.bootstrapFunc = bootstrapFunc
}

func (bh *basicHandler) setHTTPHeader(header http.Header) {
	header.Set("X-CFF-Memory", strconv.Itoa(bh.funcSpec.ResourceMetaData.Memory))
	header.Set("X-CFF-Timeout", strconv.Itoa(bh.funcSpec.FuncMetaData.Timeout))
	header.Set("X-CFF-Func-Version", bh.funcSpec.FuncMetaData.Version)
	header.Set("X-CFF-Func-Name", bh.funcSpec.FuncMetaData.FunctionName)
	header.Set("X-CFF-Project-Id", bh.funcSpec.FuncMetaData.TenantId)
	header.Set("X-CFF-Package", "")
	header.Set("X-CFF-Region", "")
	header.Set("X-CFF-Access-Key", "")
	header.Set("X-CFF-Secret-Key", "")
	header.Set("X-CFF-Security-Access-Key", "")
	header.Set("X-CFF-Security-Secret-Key", "")
	header.Set("X-CFF-Auth-Token", "")
}

func (bh *basicHandler) setEnvContext() {
	var err error
	// deal with env
	err, envMap := utils.DealEnv()
	if err != nil {
		logger.GetLogger().Errorf("deal env from createOpt failed, err: %s", err)
	}
	// http handler set environment and encrypted_user_data all to env
	for key, value := range envMap {
		if key != constants.LDLibraryPath {
			err = os.Setenv(key, value)
		} else {
			err = os.Setenv(key, os.Getenv(constants.LDLibraryPath)+fmt.Sprintf(":%s", value))
		}
	}
	if err != nil {
		logger.GetLogger().Errorf("set env from envMap failed, err: %s", err)
	}
	userDataStr, err := json.Marshal(envMap)
	if err != nil {
		logger.GetLogger().Errorf("setEnvContext failed to marshal Userdata, error: %s", err)
	}
	err = os.Setenv("RUNTIME_USERDATA", string(userDataStr))
	err = os.Setenv("RUNTIME_PROJECT_ID", bh.funcSpec.FuncMetaData.TenantId)
	err = os.Setenv("RUNTIME_FUNC_NAME", bh.funcSpec.FuncMetaData.FunctionName)
	err = os.Setenv("RUNTIME_FUNC_VERSION", bh.funcSpec.FuncMetaData.Version)
	err = os.Setenv("RUNTIME_HANDLER", bh.funcSpec.FuncMetaData.Handler)

	err = os.Setenv("RUNTIME_TIMEOUT", strconv.Itoa(bh.funcSpec.FuncMetaData.Timeout))
	nameSplit := strings.Split(bh.funcSpec.FuncMetaData.FunctionName, "@")
	if len(nameSplit) >= constants.RuntimePkgNameSplit {
		err = os.Setenv("RUNTIME_PACKAGE", nameSplit[1])
	}
	err = os.Setenv("RUNTIME_CPU", strconv.Itoa(bh.funcSpec.ResourceMetaData.Cpu))
	err = os.Setenv("RUNTIME_MEMORY", strconv.Itoa(bh.funcSpec.ResourceMetaData.Memory))
	err = os.Setenv("RUNTIME_MAX_RESP_BODY_SIZE", strconv.Itoa(constants.RuntimeMaxRespBodySize))
	err = os.Setenv("RUNTIME_INITIALIZER_HANDLER", bh.funcSpec.ExtendedMetaData.Initializer.Handler)
	err = os.Setenv("RUNTIME_INITIALIZER_TIMEOUT",
		strconv.FormatInt(int64(bh.funcSpec.ExtendedMetaData.Initializer.Timeout), constants.INT64ToINT))
	err = os.Setenv("RUNTIME_ROOT", constants.RuntimeRoot)
	err = os.Setenv("RUNTIME_CODE_ROOT", constants.RuntimeCodeRoot)
	err = os.Setenv("RUNTIME_LOG_DIR", constants.RuntimeLogDir)
	if err != nil {
		logger.GetLogger().Errorf("set env from createOpt failed, err: %s", err)
	}
}

func (bh *basicHandler) sendRequest(request *http.Request, timeout time.Duration) (*http.Response, error) {
	bh.setHTTPHeader(request.Header)
	response, err := bh.getHTTPClient(timeout).Do(request)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request with timeout %d s, error %s", 
			int(timeout.Seconds()), err.Error())
		return nil, err
	}
	return response, nil
}

func (bh *basicHandler) awaitReady(timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	for {
		select {
		case <-timer.C:
			logger.GetLogger().Errorf("timeout waiting for http server running")
			return errors.New("timeout waiting for http server running")
		default:
		}
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", listenIP, bh.port))
		if err != nil {
			time.Sleep(checkReadyInterval)
			continue
		}
		err = conn.Close()
		if err != nil {
			logger.GetLogger().Warnf("failed to close connection for checking, err: %s", err.Error())
		}
		if bh.isCustomHealthCheckReady() {
			return nil
		}
		time.Sleep(checkReadyInterval)
	}
}

func (bh *basicHandler) processInitRequest(timeout time.Duration) ([]byte, error) {
	traceID := utils.UniqueID()
	logRecorder := bh.functionLogger.NewLogRecorder(traceID, traceID, constants.InitializeStage)
	logRecorder.StartSync()
	defer func() {
		logRecorder.FinishSync()
		logRecorder.Finish()
	}()
	var initURL string
	if bh.funcSpec != nil && bh.funcSpec.ExtendedMetaData.Initializer.Handler != "" {
		initURL = fmt.Sprintf("%s/%s", bh.baseURL, "init")
	} else {
		initURL = bh.baseURL
	}
	requestBody := []byte("{}")
	request, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, initURL, bytes.NewBuffer(requestBody))
	if err != nil {
		logger.GetLogger().Errorf("failed to create request to init route error %s", err.Error())
		return utils.HandleInitResponse(
			fmt.Sprintf("failed to create request to init route error %s", err.Error()),
			constants.ExecutorErrCodeInitFail)
	}
	queries := request.URL.Query()
	request.URL.RawQuery = queries.Encode()
	request.Header.Add("Content-Type", "application/json")
	if (bh.funcSpec != nil && bh.funcSpec.ExtendedMetaData.Initializer.Handler != "") || bh.initRoute != "" {
		_, err, _ = executeWithTimeout(func() (interface{}, error) {
			return bh.sendRequest(request, timeout)
		}, timeout, bh.monitor.ErrChan)
	}
	if err != nil {
		logger.GetLogger().Errorf("init failed request error %s", err.Error())
		switch err {
		case monitor.ErrExecuteDiskLimit:
			return utils.HandleInitResponse(fmt.Sprintf("init failed request disk limit"), constants.DiskUsageExceed)
		case monitor.ErrExecuteOOM:
			return utils.HandleInitResponse(fmt.Sprintf("init failed reqeust oom"), constants.MemoryLimitExceeded)
		case ErrExecuteTimeout:
			return utils.HandleInitResponse(fmt.Sprintf("init failed request timed out after %ds",
				int(timeout.Seconds())), constants.InitFunctionTimeout)
		default:
			return utils.HandleInitResponse(fmt.Sprintf("init failed request error %s", err.Error()),
				constants.ExecutorErrCodeInitFail)
		}
	}
	logger.GetLogger().Infof("succeed to process init request for function %s", bh.funcSpec.FuncMetaData.FunctionName)
	return []byte{}, nil
}

func (bh *basicHandler) processCallRequest(args []api.Arg, traceID string,
	totalTime utils.ExecutionDuration) ([]byte, error) {
	if len(args) < argsMinLength {
		return nil, errors.New("args.length should not less than 2")
	}
	logRecorder := bh.functionLogger.NewLogRecorder(traceID, traceID, constants.InvokeStage)
	logRecorder.StartSync()
	defer func() {
		logRecorder.FinishSync()
		logRecorder.Finish()
	}()
	userCallRequest := &types.CallRequest{}
	if err := json.Unmarshal(args[1].Data, userCallRequest); err != nil {
		return utils.HandleCallResponse(fmt.Sprintf("unmarshal invoke call request data: %s, err: %s",
			string(args[1].Data), err), constants.FaaSError, "", totalTime, nil)
	}
	timeoutSeconds := bh.callTimeout
	if userCallRequest.Timeout > defaultTimeout {
		timeoutSeconds = int(userCallRequest.Timeout)
	}
	logger := logger.GetLogger().With(zap.Any("traceID", traceID), zap.Any("timeout", timeoutSeconds))
	request, err := bh.parseRequest(context.TODO(), userCallRequest.Body, userCallRequest.Header)
	if err != nil {
		logger.Errorf("failed to create request to call route,error: %s", err.Error())
		return utils.HandleCallResponse(fmt.Sprintf("failed to parse request, err:%s", err.Error()),
			constants.FaaSError, "", totalTime, nil)
	}
	totalTime.UserFuncBeginTime = time.Now()
	result, err, _ := executeWithTimeout(func() (interface{}, error) {
		request.Header.Set("Content-Type", "application/json")
		return bh.sendRequest(request, time.Duration(timeoutSeconds) * time.Second)
	}, time.Duration(timeoutSeconds) * time.Second, bh.monitor.ErrChan)
	totalTime.UserFuncTotalTime = time.Since(totalTime.UserFuncBeginTime)
	if bh.logger != nil {
		bh.logger.syncTo(time.Now())
	}
	if err != nil {
		logger.Errorf("call request failed error %s", err.Error())
		switch err {
		case monitor.ErrExecuteDiskLimit:
			return utils.HandleCallResponse(fmt.Sprintf("call request disk limit"),
				constants.DiskUsageExceed, "", totalTime, nil)
		case monitor.ErrExecuteOOM:
			return utils.HandleCallResponse(fmt.Sprintf("call request oom"),
				constants.MemoryLimitExceeded, "", totalTime, nil)
		case ErrExecuteTimeout:
			return utils.HandleCallResponse(fmt.Sprintf("call request timed out after %ds", bh.callTimeout),
				constants.InvokeFunctionTimeout, "", totalTime, nil)
		default:
			return utils.HandleCallResponse(err.Error(), constants.FunctionRunError,
				"", totalTime, nil)
		}
	}
	response, ok := result.(*http.Response)
	if !ok {
		logger.Errorf("call response type error")
		return utils.HandleCallResponse("call response type error", constants.InvokeFunctionTimeout,
			"", totalTime, nil)
	}
	responseData, err := buildResponseData(response, traceID)
	if err != nil {
		return utils.HandleCallResponse(fmt.Sprintf("failed to build response data, err: %s", err.Error()),
			constants.FaaSError, "", totalTime, nil)
	}
	if int32(responseData.Len()) > config.MaxReturnSize {
		return utils.HandleCallResponse(fmt.Sprintf("response body size %d exceeds the limit of 6291456",
			responseData.Len()), constants.ResponseExceedLimit, "", totalTime, nil)
	}

	resp, err := bh.constructResponse(response, responseData)
	if err != nil {
		return utils.HandleCallResponse(fmt.Sprintf("failed to build apig response, err: %s", err.Error()),
			constants.FaaSError, "", totalTime, nil)
	}

	innerCode := constants.NoneError
	if response.StatusCode/100 != 2 { // 非200,202,则认为失败
		innerCode = constants.FunctionRunError
	}

	logger.Infof("check call response data %s, StatusCode %d, innerCode: %d",
		string(resp), response.StatusCode, innerCode)
	logRecorder.FinishSync()
	logResult := logRecorder.MarshalAll()
	return utils.HandleCallResponse(resp, innerCode, logResult, totalTime, response.Header)
}

func (bh *basicHandler) parseRequest(ctx context.Context, serializedEvent []byte,
	header map[string]string) (*http.Request, error) {
	if bh.disableAPIGFormat {
		return parseNormalEvent(ctx, serializedEvent, header, bh.baseURL)
	}
	return parseAPIGEvent(ctx, serializedEvent, header, bh.baseURL, bh.callRoute)
}

func (bh *basicHandler) constructResponse(resp *http.Response, body *bytes.Buffer) ([]byte, error) {
	if bh.disableAPIGFormat {
		return body.Bytes(), nil
	}
	return constructAPIGResponse(resp, body)
}

func parseNormalEvent(ctx context.Context, serializedEvent []byte, header map[string]string,
	baseURLPath string) (*http.Request, error) {
	requestURI := fmt.Sprintf("%s/%s", baseURLPath, customContainerCallPath)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURI, bytes.NewBuffer(serializedEvent))
	if err != nil {
		logger.GetLogger().Errorf("failed to construct HTTP request, error: %s", err.Error())
		return nil, err
	}
	for key, val := range header {
		request.Header.Set(key, val)
	}
	return request, nil
}

func handleBootstrapError(initTimeout int, initErr error) ([]byte, error) {
	if initErr == ErrExecuteTimeout {
		logger.GetLogger().Errorf("init failed bootstrap timed out after %ds", initTimeout)
		return utils.HandleInitResponse(fmt.Sprintf("init failed bootstrap timed out after %ds", initTimeout),
			constants.InitFunctionTimeout)
	}
	logger.GetLogger().Errorf("init failed bootstrap failed error %s", initErr.Error())
	return utils.HandleInitResponse(fmt.Sprintf("init failed bootstrap failed error %s",
		initErr.Error()), constants.InitFunctionFail)
}

// InitHandler will send init request to user's http server
// args[0]: function specification
// args[1]: create params
// args[2]: scheduler info
func (bh *basicHandler) InitHandler(args []api.Arg, apiClient api.LibruntimeAPI) ([]byte, error) {
	logger.GetLogger().Infof("start to init function %s", bh.funcSpec.FuncMetaData.FunctionName)
	err := bh.parseCreateParams(args)
	if err != nil {
		return utils.HandleInitResponse(err.Error(), constants.ExecutorErrCodeInitFail)
	}
	bh.setEnvContext()
	bh.functionLogger, err = functionlog.GetFunctionLogger(bh.config)
	if err != nil {
		return utils.HandleInitResponse(err.Error(), constants.FaaSError)
	}
	initTimeout := time.Duration(bh.initTimeout) * time.Second
	var (
		bootErr  error
		costTime time.Duration
	)
	if bh.bootstrapFunc != nil {
		_, bootErr, costTime = executeWithTimeout(func() (interface{}, error) {
			return nil, bh.bootstrapFunc()
		}, initTimeout, bh.monitor.ErrChan)
		initTimeout -= costTime
	}
	if bootErr != nil {
		return handleBootstrapError(bh.initTimeout, bootErr)
	}
	_, bootErr, costTime = executeWithTimeout(func() (interface{}, error) {
		return nil, bh.awaitReady(initTimeout)
	}, initTimeout, bh.monitor.ErrChan)
	initTimeout -= costTime
	if bootErr != nil {
		logger.GetLogger().Errorf("failed to bootstrap function %s", bh.funcSpec.FuncMetaData.FunctionName)
		return handleBootstrapError(bh.initTimeout, bootErr)
	}
	logger.GetLogger().Infof("succeed to bootstrap function %s", bh.funcSpec.FuncMetaData.FunctionName)
	return bh.processInitRequest(initTimeout)
}

// CallHandler parses invoke request and convert to http request which will be sent to user's http server
// args[0]: reserved
// args[1]: invoke payload to the target function
func (bh *basicHandler) CallHandler(args []api.Arg, ctx map[string]string) ([]byte, error) {
	traceID := ctx["traceID"]
	loggerWith := logger.GetLogger().With(zap.Any("traceID", traceID))
	loggerWith.Infof("start to call function %s", bh.funcSpec.FuncMetaData.FunctionVersionURN)
	totalTime := utils.ExecutionDuration{
		ExecutorBeginTime: time.Now(),
	}
	bh.RLock()
	isGracefulShutdown := bh.gracefulShutdown
	bh.RUnlock()
	if isGracefulShutdown {
		loggerWith.Infof("instances are gracefully exiting, function: %s", bh.funcSpec.FuncMetaData.FunctionName)
		return utils.HandleCallResponse("graceful shutdown, stop processing", constants.FaaSError,
			"", totalTime, nil)
	}
	bh.Lock()
	bh.circuitLock.RLock()
	if bh.circuitBreaker {
		bh.circuitLock.RUnlock()
		bh.Unlock()
		loggerWith.Infof("instance is circuit breaker, no need processing call request, function: %s",
			bh.funcSpec.FuncMetaData.FunctionName)
		return utils.HandleCallResponse("function circuit break, stop processing",
			constants.InstanceCircuitBreakError, "", totalTime, nil)
	}
	bh.circuitLock.RUnlock()
	atomic.AddInt32(&bh.requestNum, 1)
	bh.Unlock()
	defer func() {
		atomic.AddInt32(&bh.requestNum, -1)
	}()
	if len(args) != constants.ValidInvokeArgumentSize {
		return utils.HandleCallResponse("invalid invoke argument", constants.FaaSError,
			"", totalTime, nil)
	}
	return bh.processCallRequest(args, traceID, totalTime)
}

// CheckPointHandler handles checkpoint
func (bh *basicHandler) CheckPointHandler(checkPointId string) ([]byte, error) {
	logger.GetLogger().Infof("start to checkpoint function %s", bh.funcSpec.FuncMetaData.FunctionName)
	return nil, nil
}

// RecoverHandler handles recover
func (bh *basicHandler) RecoverHandler(state []byte) error {
	logger.GetLogger().Infof("start to recover function %s", bh.funcSpec.FuncMetaData.FunctionName)
	return nil
}

// ShutDownHandler handles shutdown
func (bh *basicHandler) ShutDownHandler(gracePeriodSecond uint64) error {
	logger.GetLogger().Infof("start to shutdown function %s", bh.funcSpec.FuncMetaData.FunctionName)
	bh.RLock()
	isGracefulShutdown := bh.gracefulShutdown
	bh.RUnlock()
	if isGracefulShutdown {
		time.Sleep(time.Duration(gracePeriodSecond) * time.Second)
		logger.GetLogger().Warnf("start to shutdown function %s second times", bh.funcSpec.FuncMetaData.FunctionName)
		return nil
	}
	bh.Lock()
	bh.gracefulShutdown = true
	bh.Unlock()
	timer := time.NewTimer(time.Duration(gracePeriodSecond) * time.Second)
	exitLoop := false
	logger.GetLogger().Infof("wait all request done")
	for !exitLoop {
		select {
		case <-timer.C:
			logger.GetLogger().Warnf("reach grace period second %d, kill http server now", gracePeriodSecond)
			exitLoop = true
			break
		default:
			bh.RLock()
			runningRequests := bh.requestNum
			bh.RUnlock()
			if runningRequests == 0 {
				exitLoop = true
				logger.GetLogger().Infof("all requests finished")
				closeOnce.Do(func() {
					close(stopCh)
				})
				break
			}
			time.Sleep(shutdownCheckInterval)
		}
	}
	logger.GetLogger().Infof("http server for function %s is killed", bh.funcSpec.FuncMetaData.FunctionName)
	logger.GetLogger().Sync()
	runtimeLog.GetLogger().Sync()
	return nil
}

// SignalHandler api for go-runtime, management function instance by semaphore.
// semaphore: 1-63 Yuanrong kernel reservation
// semaphore: 64-1024 User defined
func (bh *basicHandler) SignalHandler(signal int32, payload []byte) error {
	logger.GetLogger().Infof("start to signal function %s", bh.funcSpec.FuncMetaData.FunctionName)
	var aliasList []*aliasroute.AliasElement
	if signal == killSignalAliasUpdate {
		err := json.Unmarshal(payload, &aliasList)
		if err != nil {
			return err
		}
		aliasroute.UpdateAliasesMap(aliasList)
	}
	return nil
}

func executeWithTimeout(function func() (interface{}, error), timeout time.Duration, monitorErrChan <-chan error) (
	interface{}, error, time.Duration) {
	startTime := time.Now()
	type asyncResult struct {
		res interface{}
		err error
	}
	resCh := make(chan asyncResult, 1)
	timer := time.NewTimer(timeout)
	go func() {
		result, err := function()
		resCh <- asyncResult{
			res: result,
			err: err,
		}
	}()
	select {
	case result := <-resCh:
		return result.res, result.err, time.Now().Sub(startTime)
	case <-timer.C:
		return nil, ErrExecuteTimeout, time.Now().Sub(startTime)
	case err := <-monitorErrChan:
		return nil, err, time.Now().Sub(startTime)
	}
}

func buildResponseData(response *http.Response, traceID string) (*bytes.Buffer, error) {
	const base = 1024
	buffer := &bytes.Buffer{}
	buf := make([]byte, base)
	logger := logger.GetLogger().With(zap.Any("traceID", traceID))
	defer response.Body.Close()
	for {
		readLength, err := response.Body.Read(buf)
		if err != nil && err != io.EOF {
			logger.Errorf("found error in response body,error: %s", err.Error())
			if strings.Contains(err.Error(), "context deadline exceeded") {
				return nil, errors.New("http function request timeout")
			}
			return nil, errors.New("http function request error")
		}
		if readLength > 0 {
			buffer.Write(buf[:readLength])
		} else {
			break
		}
	}
	return buffer, nil
}

// HealthCheckHandler custom health check handler
func (bh *basicHandler) HealthCheckHandler() (api.HealthType, error) {
	return api.Healthy, nil
}

func (bh *basicHandler) isCustomHealthCheckReady() bool {
	check := bh.funcSpec.ExtendedMetaData.CustomHealthCheck
	if check.TimeoutSeconds == 0 || check.PeriodSeconds == 0 || check.FailureThreshold == 0 {
		logger.GetLogger().Infof("custom health check is disabled, no need check")
		return true
	}
	logger.GetLogger().Debugf("custom health check is available, waiting for health check success")
	healthURL := fmt.Sprintf("%s/%s", bh.baseURL, bh.healthRoute)
	request, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, healthURL, nil)
	if err != nil {
		logger.GetLogger().Errorf("failed to create request to health route error %s", err.Error())
		return false
	}
	queries := request.URL.Query()
	request.URL.RawQuery = queries.Encode()
	request.Header.Add("Content-Type", "application/json")
	res, err := bh.sendRequest(request,
		time.Duration(check.TimeoutSeconds)*time.Second)
	if err == nil && res != nil && res.StatusCode == http.StatusOK {
		logger.GetLogger().Infof("health check successfully, custom runtime is ready")
		return true
	}
	return false
}
