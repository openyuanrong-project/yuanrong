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
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/uuid"

	"yuanrong.org/kernel/runtime/faassdk/common/alarm"
	"yuanrong.org/kernel/runtime/faassdk/common/aliasroute"
	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/common/faasscheduler"
	"yuanrong.org/kernel/runtime/faassdk/common/functionlog"
	"yuanrong.org/kernel/runtime/faassdk/common/monitor"
	"yuanrong.org/kernel/runtime/faassdk/common/tokentosecret"
	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/faassdk/handler"
	"yuanrong.org/kernel/runtime/faassdk/handler/http/crossclusterinvoke"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/faassdk/utils"
	"yuanrong.org/kernel/runtime/faassdk/utils/urnutils"
	"yuanrong.org/kernel/runtime/libruntime/common"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const (
	invokeServerAddr        = "127.0.0.1:31538"
	funcNameSeparator       = "@"
	defaultInvokeTimeout    = 900 * time.Second
	customContainerCallPath = "invoke"
	// CustomGracefulShutdownClientTimeout -
	CustomGracefulShutdownClientTimeout = time.Duration(3) * time.Second
	// CustomGracefulShutdownPeriod -
	CustomGracefulShutdownPeriod  = 5
	maxInvokeRetries              = 5
	customContainerSystemTenantID = "default" // same with systemTenantID in rpc pkg
)

const (
	invokeRoute            = "/invoke"
	getFutureRoute         = "/future"
	stateNewRoute          = "/state/new"
	stateGetRoute          = "/state/get"
	stateDeleteRoute       = "/state/delete"
	circuitBreakRoute      = "/circuitbreak"
	exitRoute              = "/exit"
	credentialRequireRoute = "/serverless/v1/credential/require"
)

var (
	stsInitOnce sync.Once
	stsInitErr  error
)

type status int

const (
	uninitialized status = iota
	completed
	failed
)

// CustomContainerHandler -
type CustomContainerHandler struct {
	*basicHandler
	invokeServer        *http.Server
	invokeClient        *http.Client
	logger              *RuntimeContainerLogger
	futureMap           map[string]chan types.GetFutureResponse
	alarmConfig         types.AlarmConfig
	customUserArgs      types.CustomUserArgs
	remoteDsClient      api.KvClient
	stateMgr            *stateManager
	stateMgrState       status
	stateMgrLock        sync.RWMutex
	crossClusterInvoker *crossclusterinvoke.Invoker
}

// NewCustomContainerHandler creates CustomContainerHandler
func NewCustomContainerHandler(funcSpec *types.FuncSpec, client api.LibruntimeAPI) handler.ExecutorHandler {
	handler := &CustomContainerHandler{
		basicHandler: newBasicHandler(funcSpec, client),
		invokeServer: &http.Server{
			Addr: invokeServerAddr,
		},
		invokeClient: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   defaultInvokeTimeout,
		},
		futureMap: make(map[string]chan types.GetFutureResponse, constants.DefaultMapSize),
	}
	handler.crossClusterInvoker = crossclusterinvoke.NewInvoker(funcSpec.FuncMetaData.FunctionVersionURN)
	return handler
}

func (ch *CustomContainerHandler) initDsClient(apiClient api.LibruntimeAPI) error {
	cr := apiClient.GetCredential()
	funcInfo, err := urnutils.GetFunctionInfo(ch.funcSpec.FuncMetaData.FunctionVersionURN)
	if err != nil {
		return err
	}
	createConfig := api.ConnectArguments{
		Host:                      os.Getenv("HOST_IP"),
		Port:                      31501,
		TimeoutMs:                 60 * 1000, // 60 seconds
		TenantID:                  funcInfo.TenantID,
		AccessKey:                 cr.AccessKey,
		SecretKey:                 cr.SecretKey,
		EnableCrossNodeConnection: true,
	}

	client, err := apiClient.CreateClient(createConfig)
	if err != nil {
		return fmt.Errorf("creating ds client failed, err: %s", err)
	}
	ch.remoteDsClient = client
	return nil
}

func (ch *CustomContainerHandler) setStateMgr(mgr *stateManager, status status) {
	ch.stateMgrLock.Lock()
	ch.stateMgr = mgr
	ch.stateMgrState = status
	ch.stateMgrLock.Unlock()
}

func (ch *CustomContainerHandler) getStateMgr() *stateManager {
	if ch.stateMgrState == failed {
		logger.GetLogger().Errorf("stateMgrState is failed!")
		return nil
	} else if ch.stateMgrState == completed {
		return ch.stateMgr
	}
	for i := 0; i < 60; i++ { // retry times 60
		ch.stateMgrLock.RLock()
		if ch.stateMgr != nil {
			ch.stateMgrLock.RUnlock()
			break
		}
		ch.stateMgrLock.RUnlock()
		logger.GetLogger().Warnf("get state mgr = nil, times %d", i)
		time.Sleep(1 * time.Second) // sleep 1 second
	}
	return ch.stateMgr
}

// InitHandler will send init request to custom container http server
// there are some features (oom, log) will to be added to this method, so keep this InitHandler for now
func (ch *CustomContainerHandler) InitHandler(args []api.Arg, apiClient api.LibruntimeAPI) ([]byte, error) {
	http.HandleFunc(invokeRoute, ch.handleInvoke)
	http.HandleFunc(getFutureRoute, ch.handleGetFuture)
	http.HandleFunc(stateNewRoute, ch.handleStateNew)
	http.HandleFunc(stateGetRoute, ch.handleStateGet)
	http.HandleFunc(stateDeleteRoute, ch.handleStateDel)
	http.HandleFunc(circuitBreakRoute, ch.handleCircuitBreak)
	http.HandleFunc(exitRoute, ch.handleExit)
	http.HandleFunc(credentialRequireRoute, ch.handleCredentialRequire)
	go func() {
		logger.GetLogger().Infof("start invoke server")
		err := ch.invokeServer.ListenAndServe()
		if err != nil {
			logger.GetLogger().Errorf("invoke server exit error %s", err.Error())
		}
	}()

	if err := ch.setCustomUserArgs(args); err != nil {
		logger.GetLogger().Errorf("set custom user args error:%s", err.Error())
		return utils.HandleInitResponse(fmt.Sprintf("set custom user args error: %s", err.Error()), constants.FaaSError)
	}
	ch.crossClusterInvoker.StsServerConfig = ch.customUserArgs.StsServerConfig
	funcMonitor, err := monitor.CreateFunctionMonitor(ch.funcSpec, stopCh)
	if err != nil {
		logger.GetLogger().Errorf("create function monitor error:", err.Error())
		return utils.HandleInitResponse(fmt.Sprintf("create function monitor error: %s", err.Error()), constants.FaaSError)
	}
	ch.monitor = funcMonitor

	err = ch.setAlarmInfo()
	if err != nil {
		logger.GetLogger().Errorf("set alarm info error:", err.Error())
	}
	common.GetTokenMgr().SetCallback(tokentosecret.GetSecretMgr().SetAuthContext)
	go func() {
		if err = ch.initDsClient(apiClient); err != nil {
			logger.GetLogger().Errorf("create ds client fail: %s", err.Error())
			ch.setStateMgr(nil, failed)
		} else {
			logger.GetLogger().Infof("create ds client successfully")
			ch.setStateMgr(GetStateManager(ch.remoteDsClient), completed)
		}
	}()
	resp, err := ch.basicHandler.InitHandler(args, apiClient)
	if err != nil {
		return resp, err
	}

	go ch.checkHealth()
	return resp, err
}

func (ch *CustomContainerHandler) setCustomUserArgs(args []api.Arg) error {
	if len(args) != constants.ValidCustomImageCreateParamSize {
		return errors.New("invalid create params number")
	}
	customUserArgs := types.CustomUserArgs{}
	if err := json.Unmarshal(args[constants.CustomImageUserArgIndex].Data, &customUserArgs); err != nil {
		return err
	}
	ch.customUserArgs = customUserArgs
	return nil
}

func (ch *CustomContainerHandler) setAlarmInfo() error {
	logger.GetLogger().Infof("start to setAlarmInfo")
	if &ch.customUserArgs == nil {
		return errors.New("custom user args is empty")
	}
	if !ch.customUserArgs.AlarmConfig.EnableAlarm {
		logger.GetLogger().Infof("enableAlarm is false")
		return nil
	}
	ch.alarmConfig = ch.customUserArgs.AlarmConfig
	alarm.SetClusterNameEnv(ch.customUserArgs.ClusterName)
	alarm.SetAlarmEnv(ch.alarmConfig.AlarmLogConfig)
	alarm.SetXiangYunFourConfigEnv(ch.alarmConfig.XiangYunFourConfig)

	if err := alarm.SetPodIP(); err != nil {
		return err
	}
	logger.GetLogger().Infof("set alarm info ok")
	return nil
}

func (ch *CustomContainerHandler) checkHealth() {
	if ch.funcSpec.ExtendedMetaData.CustomHealthCheck.TimeoutSeconds == 0 ||
		ch.funcSpec.ExtendedMetaData.CustomHealthCheck.PeriodSeconds == 0 ||
		ch.funcSpec.ExtendedMetaData.CustomHealthCheck.FailureThreshold == 0 {
		logger.GetLogger().Info("user has not configured custom health check and will not enable it")
		return
	}
	healthURL := fmt.Sprintf("%s/%s", ch.baseURL, ch.healthRoute)
	request, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, healthURL, nil)
	if err != nil {
		logger.GetLogger().Errorf("failed to create request to health route error %s", err.Error())
		return
	}
	queries := request.URL.Query()
	request.URL.RawQuery = queries.Encode()
	request.Header.Add("Content-Type", "application/json")
	if err := ch.checkRuntime(request, ch.funcSpec.ExtendedMetaData.CustomHealthCheck); err != nil {
		logger.GetLogger().Errorf("check runtime health failed, err: %s", err)
		ch.sdkClient.Exit(0, "")
	}
}

func (ch *CustomContainerHandler) checkRuntime(request *http.Request, check types.CustomHealthCheck) error {
	log := logger.GetLogger().With(zap.Any("timeoutSecond", check.TimeoutSeconds),
		zap.Any("periodSecond", check.PeriodSeconds), zap.Any("failureThreshold", check.FailureThreshold))
	log.Infof("start custom health check")
	defer log.Infof("end custom health check")
	timer := time.NewTimer(time.Duration(check.PeriodSeconds) * time.Second)
	defer timer.Stop()
	failCount := 0
	for {
		res, err := ch.sendRequest(request, time.Duration(check.TimeoutSeconds)*time.Second)
		if err != nil {
			failCount++
		} else if res != nil && res.StatusCode != http.StatusOK {
			err = fmt.Errorf("check health res status code is %d", res.StatusCode)
			failCount++
		}
		if res != nil && res.StatusCode == http.StatusOK {
			failCount = 0
		}

		if failCount >= check.FailureThreshold {
			log.Errorf("do custom health check failed, err: %s, failCount: %d", err.Error(), failCount)
			return err
		}
		select {
		case <-stopCh:
			logger.GetLogger().Warnf("stop channel closed")
			return nil
		case <-timer.C:
			timer.Reset(time.Duration(check.PeriodSeconds) * time.Second)
			continue
		}
	}
}

// ShutDownHandler handles shutdown
func (ch *CustomContainerHandler) ShutDownHandler(gracePeriodSecond uint64) error {
	conf, err := config.GetConfig(ch.funcSpec)
	if err != nil {
		return err
	}
	functionlog.Sync(conf)
	startGracefulShutdown := time.Now()
	gracefulTime := uint64(ch.funcSpec.ExtendedMetaData.CustomGracefulShutdown.MaxShutdownTimeout)
	if gracefulTime <= 0 {
		gracefulTime = gracePeriodSecond
	}
	err = ch.basicHandler.ShutDownHandler(gracefulTime)
	if err != nil {
		logger.GetLogger().Errorf("failed to shutdown handler error %s", err.Error())
	}
	leftTime := time.Duration(ch.funcSpec.ExtendedMetaData.CustomGracefulShutdown.MaxShutdownTimeout)*time.Second -
		time.Since(startGracefulShutdown)
	// custom container's runtime graceful shutdown
	if ch.funcSpec.ExtendedMetaData.CustomGracefulShutdown.MaxShutdownTimeout > 0 && leftTime > 0 {
		logger.GetLogger().Infof("wait runtime to shutdown gracefully")
		ch.customGracefulShutdownNewVer(leftTime)
	}
	// close invoke server after custom graceful shutdown
	// timeout set to 1 second because gracefulshutdown timeout is already handled above
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second) // wait for 1 second
	defer cancel()
	err = ch.invokeServer.Shutdown(ctx)
	if err != nil {
		logger.GetLogger().Errorf("failed to close invoke server error %s", err.Error())
	}
	logger.GetLogger().Infof("exit custom runtime")
	return err
}

func checkGracefulRsp(rsp *http.Response, err error) bool {
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Err.Error() == "EOF" {
				logger.GetLogger().Infof("server is shut down")
				return true
			}
		}
		return false
	}
	logger.GetLogger().Infof("graceful http rsp code is %d", rsp.StatusCode)
	if rsp.StatusCode == http.StatusOK || rsp.StatusCode == http.StatusNotFound {
		return true
	}
	return false
}

func (ch *CustomContainerHandler) customGracefulShutdownNewVer(duration time.Duration) {
	t := time.NewTicker(duration)
	defer t.Stop()
	rspChan := make(chan struct{})
	stopChan := make(chan struct{})
	go func() {
		client := http.Client{
			Timeout: time.Duration(duration) * time.Second,
		}
		for {
			select {
			case <-stopChan:
				return
			default:
				rsp, err := client.Get("http://127.0.0.1:8000/shutdown")
				if checkGracefulRsp(rsp, err) {
					rspChan <- struct{}{}
					return
				}
				time.Sleep(CustomGracefulShutdownPeriod * time.Second)
			}
		}
	}()

	select {
	case <-t.C:
		logger.GetLogger().Infof("Time to graceful shutdown")
		close(stopChan)
		return
	case <-rspChan:
		logger.GetLogger().Infof("stop by http")
		return
	}
}

// CallHandler handles call
func (ch *CustomContainerHandler) CallHandler(args []api.Arg, ctx map[string]string) ([]byte, error) {
	ch.basicHandler.logger = ch.logger
	return ch.basicHandler.CallHandler(args, ctx)
}

func (ch *CustomContainerHandler) handleInvoke(w http.ResponseWriter, r *http.Request) {
	logger.GetLogger().Infof("handle function invoke from function %s", ch.funcSpec.FuncMetaData.FunctionName)
	response := types.InvokeResponse{
		StatusCode: constants.NoneError,
	}
	defer handleResponse(&response, w, invokeRoute)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to read request body error %s", err.Error())
		return
	}
	request := types.InvokeRequest{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to unmarshal request body error %s", err.Error())
		return
	}
	objectID := uuid.New().String()
	ch.Lock()
	ch.futureMap[objectID] = make(chan types.GetFutureResponse, 1)
	ch.Unlock()
	if request.StateID == "" {
		err = ch.processInvoke(request, objectID)
	} else {
		err = ch.processInvokeByState(request, objectID)
	}

	if err != nil {
		if snErr, ok := err.(api.ErrorInfo); ok {
			logger.GetLogger().Errorf("invoke funcKey %s err type is SNError, code is %d, msg is %s, traceid: %s",
				request.FuncName, snErr.Code, snErr.Error(), request.TraceID)
			response.StatusCode = snErr.Code
			response.ErrorMessage = snErr.Error()
		} else {
			response.StatusCode = constants.FaaSError
		}
		ch.Lock()
		delete(ch.futureMap, objectID)
		ch.Unlock()
		return
	}
	response.ObjectID = objectID
}

func (ch *CustomContainerHandler) processInvoke(request types.InvokeRequest, objectID string) error {
	logWith := logger.GetLogger().With(zap.Any("function", ch.funcSpec.FuncMetaData.FunctionName),
		zap.Any("objectID", objectID), zap.Any("traceID", request.TraceID))
	logWith.Infof("processing function invoke, request %+v", request)
	funcKey, funcUrn, err := ch.getInvokeFuncKey(request.FuncName, request.FuncVersion, request.Params)
	if err != nil {
		return err
	}
	request.FuncUrn = funcUrn
	logWith = logWith.With(zap.Any("funcKey", funcKey))
	logWith.Infof("main function start to call function:%s", funcKey)
	start := time.Now()
	acquireTimeout := int(request.AcquireTimeout)
	defaultTimeout := 120 // 默认120s
	oldAcquireTimeout := acquireTimeout
	crossClusterIsUpgrading := false
	if acquireTimeout == 0 && ch.crossClusterInvoker.InvokeConfig.Enable {
		acquireTimeout = ch.crossClusterInvoker.AcquireTimeout
	} else if acquireTimeout == 0 {
		acquireTimeout = defaultTimeout
	}
	functionMeta := api.FunctionMeta{FuncID: funcKey, Api: api.FaaSApi}
	arg, err := prepareInvokeArg(request, ch.disableAPIGFormat)
	if err != nil {
		return err
	}
	go func() {
		wait := make(chan struct{}, 1)
		response := types.GetFutureResponse{
			StatusCode: constants.NoneError,
		}
		defer func() {
			ch.RLock()
			futureCh, exist := ch.futureMap[objectID]
			ch.RUnlock()
			if !exist {
				logWith.Errorf("future channel doesn't exist")
				return
			}
			if futureCh != nil {
				futureCh <- response
			}
		}()
		leftRetryTimes := maxInvokeRetries
		for {
			trafficLimit := false
			shouldRetry := false
			shouldCrossClusterInvoke := false
			invokeOptions := api.InvokeOptions{
				SchedulerFunctionID: faasscheduler.Proxy.GetSchedulerFuncKey(),
				RetryTimes:          leftRetryTimes,
				TraceID:             request.TraceID,
				Timeout:             int(request.Timeout),
				AcquireTimeout:      acquireTimeout,
			}
			returnObjectID, InvokeErr := ch.sdkClient.InvokeByFunctionName(functionMeta, arg, invokeOptions)
			if InvokeErr != nil {
				response.StatusCode = constants.FaaSError
				response.ErrorMessage = fmt.Sprintf("invoke funcKey %s,return error %s", funcKey, InvokeErr.Error())
				return
			}
			ch.sdkClient.GetAsync(returnObjectID, func(result []byte, err error) {
				defer func() {
					wait <- struct{}{}
				}()
				if _, decreaseErr := ch.sdkClient.GDecreaseRef([]string{returnObjectID}); decreaseErr != nil {
					fmt.Printf("failed to decrease object ref,err: %s", decreaseErr.Error())
				}
				if checkTrafficLimitResp(result) {
					trafficLimit = true
					return
				}
				if err != nil {
					logWith.Errorf("invoke function err, %s", err.Error())
					response.StatusCode = constants.FaaSError
					response.ErrorMessage = fmt.Sprintf("invoke funcKey %s,return error %s traceID: %s",
						funcKey, err.Error(), request.TraceID)
					if err != nil && ch.crossClusterInvoker.NeedCrossClusterInvoke(err) && !crossClusterIsUpgrading {
						shouldCrossClusterInvoke = true
						return
					}
					response.StatusCode = constants.FaaSError
					response.ErrorMessage = fmt.Sprintf("invoke funcKey %s,return error %s", funcKey, err.Error())
					if snErr, ok := err.(api.ErrorInfo); ok {
						logWith.Errorf("invoke err type is SNError, code is %d, msg is %s",
							snErr.Code, snErr.Error())
						response.StatusCode = snErr.Code
						response.ErrorMessage = snErr.Error()
					} else {
						logWith.Debugf("invoke err type is not SNError")
					}
					return
				}
				handleGetFutureResponse(result, &response, ch.disableAPIGFormat)
			})
			<-wait
			if trafficLimit && request.StateID == "" { // invoking by stateID don't do trafficlimit
				continue
			}
			if shouldCrossClusterInvoke {
				timeout := time.Duration(request.Timeout)*time.Second - time.Now().Sub(start)
				crossClusterIsUpgrading = ch.crossClusterInvoker.DoInvoke(request, &response, timeout, logWith)
				if !crossClusterIsUpgrading {
					break
				}
				invokeOptions.Timeout = oldAcquireTimeout
				shouldRetry = true
				leftRetryTimes--
			}
			if shouldRetry && leftRetryTimes > 0 {
				continue
			}
			break
		}
	}()
	return nil
}

func isFaaSSchedulerStateErrorCode(errCode int) bool {
	faasschedulerErrorCodes := []int{StateInstanceNotExistedErrCode, StateInstanceNoLease, FaaSSchedulerInternalErrCode}
	for _, code := range faasschedulerErrorCodes {
		if errCode == code {
			return true
		}
	}
	return false
}

func (ch *CustomContainerHandler) processInvokeByState(request types.InvokeRequest, objectID string) error {
	loggerWith := logger.GetLogger().With(zap.Any("traceID", request.TraceID), zap.Any("stateID", request.StateID),
		zap.Any("funcName", request.FuncName), zap.Any("objectID", objectID))
	loggerWith.Infof("processing function state invoke from function %s, request %v",
		ch.funcSpec.FuncMetaData.FunctionName, request)
	funcKey, _, err := ch.getInvokeFuncKey(request.FuncName, request.FuncVersion, request.Params)
	if err != nil {
		return err
	}
	arg, err := prepareInvokeArg(request, ch.disableAPIGFormat)
	if err != nil {
		return err
	}
	schedulerID, err := faasscheduler.Proxy.Get(funcKey)
	if err != nil {
		return err
	}
	functionMeta := api.FunctionMeta{FuncID: funcKey, Api: api.FaaSApi}
	option := api.InvokeOptions{
		SchedulerFunctionID:  faasscheduler.Proxy.GetSchedulerFuncKey(),
		SchedulerInstanceIDs: []string{schedulerID},
		TraceID:              request.TraceID,
		Timeout:              120, // 120 seconds
		AcquireTimeout:       120,
	}
	stateMgr := ch.getStateMgr()
	if stateMgr == nil {
		loggerWith.Errorf("stateMgr is nil!!")
		return errors.New("stateMgr is nil")
	}
	lease, err := ch.sdkClient.AcquireInstance(request.StateID, functionMeta, option)
	loggerWith.Debugf("get lease err = %v %T", err, err)
	if snErr, ok := err.(api.ErrorInfo); ok {
		loggerWith.Errorf("invoke funcKey %s err type is SNError, code is %d, msg is %s",
			funcKey, snErr.Code, snErr.Error())
		if isFaaSSchedulerStateErrorCode(snErr.Code) {
			stateMgr.delInstance(funcKey, request.StateID)
			return err
		}
	}
	if err != nil {
		leasePtr := stateMgr.getInstance(funcKey, request.StateID)
		if leasePtr == nil {
			loggerWith.Errorf("failed to get lease, err: %s", err.Error())
			return err
		}
		lease = *leasePtr
		loggerWith.Warnf("failed to get lease, err: %s, to use cached lease %s", err.Error(), lease.InstanceID)
	} else {
		stateMgr.addInstance(funcKey, request.StateID, &lease)
		loggerWith.Infof("aquire lease ok: %s", lease.InstanceID)
	}

	go func() {
		wait := make(chan struct{}, 1)
		response := types.GetFutureResponse{
			StatusCode: constants.NoneError,
		}
		defer func() {
			ch.RLock()
			futureCh, exist := ch.futureMap[objectID]
			ch.RUnlock()
			if !exist {
				loggerWith.Errorf("future channel for objectID %s doesn't exist for function %s traceID %s",
					ch.funcSpec.FuncMetaData.FunctionName, request.TraceID)
				return
			}
			if futureCh != nil {
				futureCh <- response
			} else {
				loggerWith.Errorf("futureCh is nil")
			}
		}()
		option.Timeout = int(request.Timeout)
		returnObjectID, invokeErr := ch.sdkClient.InvokeByInstanceId(functionMeta, lease.InstanceID, arg, option)
		if invokeErr != nil {
			response.StatusCode = constants.FaaSError
			response.ErrorMessage = fmt.Sprintf("invoke funcKey %s,return error %s", funcKey, invokeErr.Error())
			return
		}
		ch.sdkClient.GetAsync(returnObjectID, func(result []byte, cbErr error) {
			defer func() {
				wait <- struct{}{}
			}()
			loggerWith.Infof("finish invoke, err: %v, result len = %d", cbErr, len(result))
			if cbErr != nil {
				if snErr, ok := cbErr.(api.ErrorInfo); ok {
					loggerWith.Errorf("invoke funcKey %s err type is SNError, code is %d, msg is %s",
						funcKey, snErr.Code, snErr.Error())
					response.StatusCode = snErr.Code
					response.ErrorMessage = snErr.Error()
				} else {
					loggerWith.Errorf("invoke funcKey %s err type is not SNError: %s", funcKey, err.Error())
					response.StatusCode = constants.FaaSError
					response.ErrorMessage = fmt.Sprintf("invoke funcKey %s,return error %s", funcKey, err.Error())
				}
				return
			}
			loggerWith.Debugf("inovke result: %v", result)
			handleGetFutureResponse(result, &response, ch.disableAPIGFormat)
		})
		<-wait
		ch.sdkClient.ReleaseInstance(lease, "", false, option)
	}()
	return nil
}

func (ch *CustomContainerHandler) handleStateNew(w http.ResponseWriter, r *http.Request) {
	logger.GetLogger().Infof("handle state new from function %s", ch.funcSpec.FuncMetaData.FunctionName)
	response := types.StateResponse{
		StatusCode: constants.NoneError,
	}
	defer handleResponse(&response, w, stateNewRoute)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to read request body error %s", err.Error())
		return
	}
	logger.GetLogger().Debugf("state new req data is %s", data)
	request := types.StateRequest{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to unmarshal request body error %s", err.Error())
		return
	}
	logger := logger.GetLogger().With(zap.Any("traceID", request.TraceID))
	err = ch.processState(request, "new")
	if err != nil {
		if snErr, ok := err.(api.ErrorInfo); ok {
			logger.Errorf("state new rsp err type is SNError, code is %d, msg is %s",
				snErr.Code, snErr.Error())
			response.StatusCode = snErr.Code
			response.ErrorMessage = snErr.Error()
		} else {
			logger.Errorf("state new req %v rsp err type is not SNError", request)
			response.StatusCode = constants.FaaSError
			response.ErrorMessage = fmt.Sprintf("failed process state new error %s", err.Error())
		}
		return
	}
	response.StateID = request.StateID
}

func (ch *CustomContainerHandler) handleStateGet(w http.ResponseWriter, r *http.Request) {
	logger.GetLogger().Infof("handle state get from function %s", ch.funcSpec.FuncMetaData.FunctionName)
	response := types.StateResponse{
		StatusCode: constants.NoneError,
	}
	defer handleResponse(&response, w, stateGetRoute)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to read request body error %s", err.Error())
		return
	}
	logger.GetLogger().Debugf("state get req data is %s", data)
	request := types.StateRequest{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to unmarshal request body error %s", err.Error())
		return
	}
	logger := logger.GetLogger().With(zap.Any("traceID", request.TraceID))
	err = ch.processState(request, "get")
	if err != nil {
		if snErr, ok := err.(api.ErrorInfo); ok {
			logger.Errorf("state get rsp err type is SNError, code is %d, msg is %s",
				snErr.Code, snErr.Error())
			response.StatusCode = snErr.Code
			response.ErrorMessage = snErr.Error()
		} else {
			logger.Errorf("state get req %v rsp err type is not SNError", request)
			response.StatusCode = constants.FaaSError
			response.ErrorMessage = fmt.Sprintf("failed process state new error %s", err.Error())
		}
		return
	}
	response.StateID = request.StateID
}

func (ch *CustomContainerHandler) handleStateDel(w http.ResponseWriter, r *http.Request) {
	logger.GetLogger().Infof("handle function invoke from function %s", ch.funcSpec.FuncMetaData.FunctionName)
	response := types.TerminateResponse{
		StatusCode: constants.NoneError,
	}
	defer handleResponse(&response, w, stateDeleteRoute)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to read request body error %s", err.Error())
		return
	}
	logger.GetLogger().Infof("state del req data is %s", data)
	request := types.StateRequest{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to unmarshal request body error %s", err.Error())
		return
	}
	logger := logger.GetLogger().With(zap.Any("traceID", request.TraceID))
	err = ch.processState(request, "del")
	logger.Infof("processState del err: %v, %T", err, err)
	if err != nil {
		if snErr, ok := err.(api.ErrorInfo); ok {
			logger.Errorf("state del rsp err type is SNError, code is %d, msg is %s",
				snErr.Code, snErr.Error())
			response.StatusCode = snErr.Code
			response.ErrorMessage = snErr.Error()
		} else {
			logger.Errorf("state del rsp err type is not SNError")
			response.StatusCode = constants.FaaSError
			response.ErrorMessage = fmt.Sprintf("failed process state new error %s", err.Error())
		}
		return
	}
	objectID := uuid.New().String()
	furtureRspChan := make(chan types.GetFutureResponse, 1)
	ch.Lock()
	ch.futureMap[objectID] = furtureRspChan
	ch.Unlock()
	logger.Infof("set object %s to futureMap", objectID)
	response.ObjectID = objectID
	furtureRspChan <- types.GetFutureResponse{
		StatusCode: constants.NoneError,
		Content:    `{"result": "DELETE SUCCESSFULLY"}`,
	}
}

func (ch *CustomContainerHandler) handleExit(w http.ResponseWriter, r *http.Request) {
	logger.GetLogger().Infof("handle exit")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.GetLogger().Errorf("read body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.GetLogger().Infof("get exit body %s", string(body))
	req := &types.ExitRequest{}
	if len(body) > 0 {
		err = json.Unmarshal(body, req)
		if err != nil {
			logger.GetLogger().Errorf("unmarshal exit body failed: %v", err)
		}
	}
	response := types.ExitResponse{
		StatusCode: constants.NoneError,
	}
	handleResponse(&response, w, exitRoute)
	go ch.sdkClient.Exit(req.Code, req.Message)
}

func (ch *CustomContainerHandler) handleCredentialRequire(w http.ResponseWriter, r *http.Request) {
	logger.GetLogger().Infof("handle credential require")
	response := types.CredentialResponse{
		StatusCode: constants.NoneError,
	}
	credential := ch.sdkClient.GetCredential()
	response.Credential = credential
	handleResponse(&response, w, credentialRequireRoute)

}

func (ch *CustomContainerHandler) handleCircuitBreak(w http.ResponseWriter, r *http.Request) {
	logger.GetLogger().Infof("handle circuit break")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.GetLogger().Errorf("read body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req := &types.CircuitBreakRequest{}
	err = json.Unmarshal(body, req)
	if err != nil {
		logger.GetLogger().Errorf("unmarshal body failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := types.ExitResponse{
		StatusCode: constants.NoneError,
	}
	ch.circuitLock.Lock()
	ch.circuitBreaker = req.Switch
	ch.circuitLock.Unlock()
	logger.GetLogger().Infof("circuit break flag set to %v", req.Switch)
	handleResponse(&response, w, circuitBreakRoute)
}

func (ch *CustomContainerHandler) releaseState(funcKey, stateID, traceID string) error {
	schedulerID, err := faasscheduler.Proxy.Get(funcKey)
	if err != nil {
		return err
	}
	option := api.InvokeOptions{
		SchedulerFunctionID:  faasscheduler.Proxy.GetSchedulerFuncKey(),
		SchedulerInstanceIDs: []string{schedulerID},
		TraceID:              traceID,
	}
	instanceAllocation := api.InstanceAllocation{
		FuncKey: funcKey,
	}
	ch.sdkClient.ReleaseInstance(instanceAllocation, fmt.Sprintf("%s;%s", funcKey, stateID),
		false, option)
	return nil
}

func (ch *CustomContainerHandler) processState(request types.StateRequest, opType string) error {
	funcKey, _, err := ch.getInvokeFuncKey(request.FuncName, request.FuncVersion, request.Params)
	if err != nil {
		return err
	}
	if request.StateID == "" {
		return api.ErrorInfo{
			Code: InvalidState,
			Err:  fmt.Errorf(InvalidStateErrMsg),
		}
	}
	stateMgr := ch.getStateMgr()
	if stateMgr == nil {
		logger.GetLogger().Errorf("stateMgr is nil!!")
		return errors.New("stateMgr is nil")
	}
	switch opType {
	case "new":
		err = stateMgr.newState(funcKey, request.StateID, request.TraceID)
	case "get":
		err = stateMgr.getState(funcKey, request.StateID, request.TraceID)
	case "del":
		err = stateMgr.delState(funcKey, request.StateID, request.TraceID)
		if err == nil {
			logger.GetLogger().Infof("releaseState to faasscheduler")
			err = ch.releaseState(funcKey, request.StateID, request.TraceID)
			if err != nil {
				alarmDetail := &alarm.Detail{
					SourceTag: os.Getenv(constants.PodNameEnvKey) + "|" + os.Getenv(constants.PodIPEnvKey) +
						"|" + os.Getenv(constants.ClusterName) + "|" + os.Getenv(constants.HostIPEnvKey),
					OpType: alarm.GenerateAlarmLog,
					Details: fmt.Sprintf("terminate State failed, faasscheduler error: %v, "+
						"stateKey: %s, statefuncKey: %s", err, request.StateID, funcKey),
					StartTimestamp: int(time.Now().Unix()),
					EndTimestamp:   0,
				}
				alarmInfo := &alarm.LogAlarmInfo{
					AlarmID:    "TerminateStateForFaasScheduler00001",
					AlarmName:  "TerminateStateForFaasScheduler",
					AlarmLevel: alarm.Level2,
				}
				alarm.ReportOrClearAlarm(alarmInfo, alarmDetail)
			}
			logger.GetLogger().Debugf("releaseState to faasscheduler, err: %v, %T", err, err)
		}
	default:
		err = errors.New(fmt.Sprintf("unknow opType %s", opType))
	}
	return err
}

func checkTrafficLimitResp(notifyMsg []byte) bool {
	if notifyMsg != nil && len(notifyMsg) != 0 {
		var insResponse struct {
			ErrorCode    int    `json:"errorCode"`
			ErrorMessage string `json:"errorMessage"`
		}
		if unMarshalErr := json.Unmarshal(notifyMsg, &insResponse); unMarshalErr != nil {
			logger.GetLogger().Errorf("unmarshal notifyMsg error : %s", unMarshalErr.Error())
		}
		// current faasscheduler has reached instance limit, should retry and chose another faasscheduler
		return insResponse.ErrorCode == constants.AcquireLeaseTrafficLimitErrorCode
	}
	return false
}

func handleGetFutureResponse(result []byte, response *types.GetFutureResponse, disableAPIGFormat bool) {
	callResponse := &types.CallResponse{}
	err := json.Unmarshal(result, callResponse)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("call response unmarshal error %s", err.Error())
		return
	}
	if len(callResponse.InnerCode) != 0 {
		innerCode, err := strconv.Atoi(callResponse.InnerCode)
		if err != nil {
			response.StatusCode = constants.FaaSError
			response.ErrorMessage = fmt.Sprintf("failed to get the innerCode, err: %s", err)
			return
		}
		if innerCode != constants.NoneError {
			response.StatusCode = innerCode
			response.ErrorMessage = string(callResponse.Body)
			return
		}
	}
	if disableAPIGFormat {
		response.Content = string(callResponse.Body)
		logger.GetLogger().Infof("set rsp content: %s", response.Content)
		return
	}
	apigResponse := APIGTriggerResponse{}
	err = json.Unmarshal(callResponse.Body, &apigResponse)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to unmarshal http response error %s", err.Error())
		return
	}
	content := apigResponse.Body
	if apigResponse.IsBase64Encoded {
		decodeContent, err := base64.StdEncoding.DecodeString(apigResponse.Body)
		if err != nil {
			response.StatusCode = constants.FaaSError
			response.ErrorMessage = fmt.Sprintf("failed to decode http response error %s", err.Error())
			return
		}
		content = string(decodeContent)
	}
	if apigResponse.StatusCode != http.StatusOK {
		response.StatusCode = constants.FunctionRunError
		response.ErrorMessage = content
	} else {
		response.Content = content
		logger.GetLogger().Infof("set rsp content: %s", response.Content)
	}
}

func prepareInvokeArg(request types.InvokeRequest, disableAPIGFormat bool) ([]api.Arg, error) {
	requestData, err := getCallRequestBody(request, disableAPIGFormat)
	if err != nil {
		return nil, err
	}
	callRequest := &types.CallRequest{
		Body: requestData,
		Header: map[string]string{
			constants.CaaSTraceIDHeaderKey:  request.TraceID,
			constants.CffRequestIDHeaderKey: request.TraceID,
			"Content-Type":                  "application/json",
		},
		Timeout: request.Timeout,
	}
	data, err := json.Marshal(callRequest)
	if err != nil {
		return nil, err
	}
	arg := []api.Arg{
		{
			Type: api.Value,
			Data: []byte(request.TraceID),
		},
		{
			Type: api.Value,
			Data: data,
		},
	}
	return arg, nil
}

func getCallRequestBody(request types.InvokeRequest, disableAPIGFormat bool) ([]byte, error) {
	if disableAPIGFormat {
		return []byte(request.Payload), nil
	}
	apigRequestData, err := buildAPIGRequestData(request.Payload)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to build APIG request error %s", err.Error()))
	}
	return apigRequestData, nil
}

func (ch *CustomContainerHandler) getInvokeFuncKey(funcName, funcVer string,
	params map[string]string) (string, string, error) {
	funcUrnPrefixIndex := strings.LastIndex(ch.funcSpec.FuncMetaData.FunctionVersionURN, funcNameSeparator)
	funcUrnPathPrefix := ch.funcSpec.FuncMetaData.FunctionVersionURN[:funcUrnPrefixIndex]
	aliasUrn := fmt.Sprintf("%s@%s:%s", funcUrnPathPrefix, funcName, funcVer)
	funcUrn := aliasroute.GetAliases().GetFuncVersionURNWithParams(aliasUrn, params)
	functionInfo, err := urnutils.GetFunctionInfo(funcUrn)
	if err != nil {
		return "", "", err
	}
	funcKey := urnutils.CombineFunctionKey(functionInfo.TenantID, functionInfo.Name, functionInfo.Version)
	return funcKey, funcUrn, err
}

func (ch *CustomContainerHandler) handleGetFuture(w http.ResponseWriter, r *http.Request) {
	response := types.GetFutureResponse{}
	defer handleResponse(&response, w, getFutureRoute)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to read future request body error %s", err.Error())
		return
	}
	logger.GetLogger().Infof("state getfuture req data is %s", data)
	request := types.GetFutureRequest{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("failed to read future request body error %s", err.Error())
		return
	}
	logger.GetLogger().Infof("getfuture objID is %s", request.ObjectID)
	ch.RLock()
	futureCh, exist := ch.futureMap[request.ObjectID]
	ch.RUnlock()
	if !exist {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("objectID %s doesn't exit", request.ObjectID)
		return
	}
	if futureCh != nil {
		response = <-futureCh
	}
	ch.Lock()
	delete(ch.futureMap, request.ObjectID)
	ch.Unlock()
}

func handleResponse(response types.Response, w http.ResponseWriter, handleType string) {
	logger.GetLogger().Infof("handle %s response code %d message %s", handleType, response.GetStatusCode(),
		response.GetErrorMessage())
	if response.GetStatusCode() != constants.NoneError {
		logger.GetLogger().Errorf("handle %s error %s", handleType, response.GetErrorMessage())
	}
	data, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.GetLogger().Errorf("failed to marshal %s response error %s", handleType, err.Error())
		if _, err = w.Write([]byte("failed to marshal response")); err != nil {
			logger.GetLogger().Errorf("failed to write marshal error %s", err.Error())
		}
		return
	}
	if _, err = w.Write(data); err != nil {
		logger.GetLogger().Errorf("failed to write %s response data error %s", handleType, err.Error())
	}
}

func buildAPIGRequestData(payload string) ([]byte, error) {
	apigRequest := APIGTriggerEvent{
		Path:            customContainerCallPath,
		Body:            payload,
		IsBase64Encoded: false,
	}
	data, err := json.Marshal(apigRequest)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// HealthCheckHandler health check
func (ch *CustomContainerHandler) HealthCheckHandler() (api.HealthType,
	error) {
	code := api.Healthy
	ch.circuitLock.RLock()
	defer ch.circuitLock.RUnlock()
	if ch.circuitBreaker {
		code = api.SubHealth
		return code, nil
	}
	return code, nil
}
