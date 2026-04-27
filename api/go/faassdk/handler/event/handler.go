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
	"errors"
	"fmt"
	"os"
	"plugin"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/config"
	contextApi "yuanrong.org/kernel/runtime/faassdk/faas-sdk/go-api/context"
	contextImpl "yuanrong.org/kernel/runtime/faassdk/faas-sdk/pkg/runtime/context"
	"yuanrong.org/kernel/runtime/faassdk/handler"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/faassdk/utils"
	"yuanrong.org/kernel/runtime/libruntime/api"
	log "yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
	runtimeLog "yuanrong.org/kernel/runtime/libruntime/common/logger/log"
)

const (
	userEntryIndex = 1
	userLogLevel   = "info"
	defaultTimeout = 3
)

var (
	defaultContext      *contextImpl.Provider
	stopCh              = make(chan struct{})
	preSetUserInitEntry InitEntry
	preSetUserCallEntry CallEntry
)

// InitEntry of user code
type InitEntry = func(ctx contextApi.RuntimeContext)

// PreStopEntry of user code
type PreStopEntry = func(ctx contextApi.RuntimeContext)

// CallEntry of user code
type CallEntry = func([]byte, contextApi.RuntimeContext) (interface{}, error)

// EventHandler Data required by the executor
type EventHandler struct {
	funcSpec      *types.FuncSpec
	libMap        map[string]*plugin.Plugin
	functionName  string
	args          []*api.Arg
	client        api.LibruntimeAPI
	userInitEntry InitEntry
	userCallEntry CallEntry
	initTimeout   int
	invokeTimeout int
}

// SetUserInitEntry set EventHandler's userInitEntry
func SetUserInitEntry(initEntry InitEntry) {
	preSetUserInitEntry = initEntry
}

// SetUserCallEntry set EventHandler's userCallEntry
func SetUserCallEntry(callEntry CallEntry) {
	preSetUserCallEntry = callEntry
}

// NewEventHandler creates EventHandler
func NewEventHandler(funcSpec *types.FuncSpec, client api.LibruntimeAPI) handler.ExecutorHandler {
	var initTimeout, invokeTimeout int
	if initTimeout = funcSpec.ExtendedMetaData.Initializer.Timeout; initTimeout == 0 {
		initTimeout = defaultTimeout
	}
	if invokeTimeout = funcSpec.FuncMetaData.Timeout; invokeTimeout == 0 {
		invokeTimeout = defaultTimeout
	}
	return &EventHandler{
		funcSpec:      funcSpec,
		libMap:        make(map[string]*plugin.Plugin),
		args:          make([]*api.Arg, 1),
		client:        client,
		initTimeout:   initTimeout,
		invokeTimeout: invokeTimeout,
		userInitEntry: preSetUserInitEntry,
		userCallEntry: preSetUserCallEntry,
	}
}

// getHandlerName from api.Arg
func getHandlerName(args []api.Arg) (*types.CreateParams, error) {
	if len(args) != constants.ValidBasicCreateParamSize {
		return nil, errors.New("invalid args number")
	}
	createParams := &types.CreateParams{}
	err := json.Unmarshal(args[userEntryIndex].Data, createParams)
	if err != nil {
		return nil, err
	}
	return createParams, nil
}

// InitHandler -
func (eh *EventHandler) InitHandler(args []api.Arg, rt api.LibruntimeAPI) ([]byte, error) {
	log.GetLogger().Infof("start to init user code")
	log.SetupUserLogger(userLogLevel)
	path, err := handler.GetUserCodePath()
	if err != nil {
		return utils.HandleInitResponse(err.Error(), constants.ExecutorErrCodeInitFail)
	}
	if eh.userCallEntry == nil {
		userHook, err := getHandlerName(args)
		if err != nil {
			return utils.HandleInitResponse(err.Error(), constants.ExecutorErrCodeInitFail)
		}
		if err = eh.SetHandler(path, userHook.InitEntry); err != nil {
			return utils.HandleInitResponse(err.Error(), constants.InitFunctionFail)
		}
		if err = eh.SetHandler(path, userHook.CallEntry); err != nil {
			return utils.HandleInitResponse(err.Error(), constants.InitFunctionFail)
		}
	}
	eh.setEnvContext()
	if err = initContext(); err != nil {
		log.GetLogger().Errorf("init context failed, error: %s", err)
		return utils.HandleInitResponse(fmt.Sprintf("init context failed, error: %s", err.Error()),
			constants.ExecutorErrCodeInitFail)
	}
	if eh.userInitEntry != nil {
		if result, err := eh.initWithTimeout(); err != nil {
			log.GetLogger().Errorf("faas-executor failed to init user code, err: %s", err.Error())
			return result, err
		}
	}
	log.GetLogger().Infof("faas-executor succeed to init user code")
	return []byte{}, nil
}

// CallHandler -
func (eh *EventHandler) CallHandler(args []api.Arg, context map[string]string) ([]byte, error) {
	traceID := context["traceID"]
	logger := log.GetLogger().With(zap.Any("traceID", traceID))
	totalTime := utils.ExecutionDuration{
		ExecutorBeginTime: time.Now(),
	}
	logger.Infof("faas-executor call handler, function: %s", eh.funcSpec.FuncMetaData.FunctionName)
	if len(args) != constants.ValidInvokeArgumentSize {
		return utils.HandleCallResponse("invalid invoke argument",
			constants.FaaSError, "", totalTime, nil)
	}
	if int32(len(args[1].Data)) >= config.MaxPayloadSize {
		logger.Errorf("the size of the invoke payload exceeds the limit %d", config.MaxPayloadSize)
		return utils.HandleCallResponse(
			fmt.Sprintf("the size of the invoke payload exceeds the limit %d", config.MaxPayloadSize),
			constants.FaaSError, "", totalTime, nil)
	}
	if eh.userCallEntry == nil {
		logger.Errorf("invoke handler is nil")
		return utils.HandleCallResponse("invoke handler is nil",
			constants.EntryNotFound, "", totalTime, nil)
	}

	userCallRequest := &types.CallResponse{}
	if err := json.Unmarshal(args[1].Data, userCallRequest); err != nil {
		logger.Errorf("unmarshal invoke call request data: %s, err: %s", string(args[1].Data), err)
		return utils.HandleCallResponse(fmt.Sprintf("unmarshal invoke call request data: %s, err: %s",
			string(args[1].Data), err),
			constants.FaaSError, "", totalTime, nil)
	}
	return eh.invokeWithTimeout(userCallRequest.Body, context, traceID, totalTime)
}

// CheckPointHandler -
func (eh *EventHandler) CheckPointHandler(checkPointId string) ([]byte, error) {
	return nil, nil
}

// RecoverHandler -
func (eh *EventHandler) RecoverHandler(state []byte) error {
	return nil
}

// ShutDownHandler -
func (eh *EventHandler) ShutDownHandler(gracePeriodSecond uint64) error {
	log.GetLogger().Sync()
	runtimeLog.GetLogger().Sync()
	return nil
}

// SignalHandler -
func (eh *EventHandler) SignalHandler(signal int32, payload []byte) error {
	return nil
}

// HealthCheckHandler -
func (eh *EventHandler) HealthCheckHandler() (api.HealthType, error) {
	return api.Healthy, nil
}

// initContext init envContext from env
func initContext() error {
	userFunctionLog := log.GetUserLogger()
	if userFunctionLog == nil {
		log.GetLogger().Errorf("init user function log error: cannot create file user-function.log")
		return fmt.Errorf("init user function log error: cannot create file user-function.log")
	}

	defaultContext = &contextImpl.Provider{
		CtxEnv:      &contextImpl.Env{},
		CtxHTTPHead: &contextImpl.InvokeContext{},
		Logger:      userFunctionLog,
	}
	return defaultContext.InitializeContext()
}

// SetHandler will set user entries
func (eh *EventHandler) SetHandler(functionLibPath, handler string) error {
	symbol, err := eh.getLib(functionLibPath, handler)
	if err != nil {
		return fmt.Errorf("getLib error: %s", err)
	}
	return eh.setHandlerByEntry(symbol, handler)
}

func (eh *EventHandler) setHandlerByEntry(symbol plugin.Symbol, handler string) error {
	switch symbol.(type) {
	case InitEntry:
		if eh.userInitEntry == nil {
			eh.userInitEntry = symbol.(InitEntry)
		}
	case CallEntry:
		if eh.userCallEntry == nil {
			eh.userCallEntry = symbol.(CallEntry)
		}
	default:
		return fmt.Errorf("%s type error", handler)
	}
	return nil
}

// Open plugin and set Symbol from plugin
func (eh *EventHandler) getLib(functionLibPath, handler string) (plugin.Symbol, error) {
	path, name := utils.GetLibInfo(functionLibPath, handler)
	if path == "" {
		return nil, fmt.Errorf("invalid handler name :%s", handler)
	}
	lib, ok := eh.libMap[path]
	if !ok {
		log.GetLogger().Infof("start to open lib %s", path)
		userCodePlugin, err := plugin.Open(path)
		if err != nil {
			log.GetLogger().Errorf("failed to open lib %v", err)
			return nil, fmt.Errorf("failed to open %s", handler)
		}
		lib = userCodePlugin
		eh.libMap[path] = userCodePlugin
	}
	symbol, err := lib.Lookup(name)
	if err != nil {
		log.GetLogger().Errorf("failed to look up %v", err)
		return nil, fmt.Errorf("failed to look up %s", handler)
	}
	return symbol, nil
}

// updateInvokeContext update invoke context from CallRequest.createOpt
func updateInvokeContext(invokeContext map[string]string) contextApi.RuntimeContext {
	if defaultContext == nil || defaultContext.CtxHTTPHead == nil {
		log.GetLogger().Warnf("default context is not initialized")
		return &contextImpl.Provider{
			CtxEnv:      &contextImpl.Env{},
			CtxHTTPHead: &contextImpl.InvokeContext{},
		}
	}
	defaultContext.CtxHTTPHead.RequestID = invokeContext["requestId"]
	defaultContext.CtxHTTPHead.Alias = invokeContext["Alias"]
	defaultContext.CtxHTTPHead.InvokeProperty = invokeContext["InvokeProperty"]
	defaultContext.CtxHTTPHead.InvokeID = invokeContext["InvokeID"]
	defaultContext.CtxHTTPHead.TraceID = invokeContext["TraceID"]
	log.GetLogger().Infof("succeed to update context")
	return defaultContext
}

func (eh *EventHandler) setEnvContext() {
	var err error
	// deal with env
	err, envMap := utils.DealEnv()
	if err != nil {
		log.GetLogger().Errorf("deal env from createOpt failed, err: %s", err)
	}
	userDataStr, err := json.Marshal(envMap)
	if err != nil {
		log.GetLogger().Errorf("setEnvContext failed to marshal Userdata, error: %s", err)
	}
	err = os.Setenv(constants.LDLibraryPath,
		os.Getenv(constants.LDLibraryPath)+fmt.Sprintf(":%s", envMap[constants.LDLibraryPath]))
	err = os.Setenv("RUNTIME_USERDATA", string(userDataStr))
	err = os.Setenv("RUNTIME_PROJECT_ID", eh.funcSpec.FuncMetaData.TenantId)
	err = os.Setenv("RUNTIME_FUNC_NAME", eh.funcSpec.FuncMetaData.FunctionName)
	err = os.Setenv("RUNTIME_FUNC_VERSION", eh.funcSpec.FuncMetaData.Version)
	err = os.Setenv("RUNTIME_HANDLER", eh.funcSpec.FuncMetaData.Handler)

	err = os.Setenv("RUNTIME_TIMEOUT", strconv.Itoa(eh.funcSpec.FuncMetaData.Timeout))
	nameSplit := strings.Split(eh.funcSpec.FuncMetaData.FunctionName, "@")
	if len(nameSplit) >= constants.RuntimePkgNameSplit {
		err = os.Setenv("RUNTIME_PACKAGE", nameSplit[1])
	}
	err = os.Setenv("RUNTIME_CPU", strconv.Itoa(eh.funcSpec.ResourceMetaData.Cpu))
	err = os.Setenv("RUNTIME_MEMORY", strconv.Itoa(eh.funcSpec.ResourceMetaData.Memory))
	err = os.Setenv("RUNTIME_MAX_RESP_BODY_SIZE", strconv.Itoa(constants.RuntimeMaxRespBodySize))
	err = os.Setenv("RUNTIME_INITIALIZER_HANDLER", eh.funcSpec.ExtendedMetaData.Initializer.Handler)
	err = os.Setenv("RUNTIME_INITIALIZER_TIMEOUT",
		strconv.FormatInt(int64(eh.funcSpec.ExtendedMetaData.Initializer.Timeout), constants.INT64ToINT))
	err = os.Setenv("RUNTIME_ROOT", constants.RuntimeRoot)
	err = os.Setenv("RUNTIME_CODE_ROOT", constants.RuntimeCodeRoot)
	err = os.Setenv("RUNTIME_LOG_DIR", constants.RuntimeLogDir)
	if err != nil {
		log.GetLogger().Errorf("set env from createOpt failed, err: %s", err)
	}
}
