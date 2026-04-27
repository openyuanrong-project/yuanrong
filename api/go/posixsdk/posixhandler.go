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
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"

	"yuanrong.org/kernel/runtime/libruntime"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/log"
	"yuanrong.org/kernel/runtime/libruntime/config"
	"yuanrong.org/kernel/runtime/libruntime/execution"
	"yuanrong.org/kernel/runtime/libruntime/libruntimesdkimpl"
)

type initHandlerType = func([]api.Arg, api.LibruntimeAPI) ([]byte, error)
type callHandlerType = func([]api.Arg) ([]byte, error)
type checkpointHandlerType = func(string) ([]byte, error)
type recoverHandlerType = func([]byte, api.LibruntimeAPI) error
type shutdownHandlerType = func(gracePeriodSecond uint64) error
type signalHandlerType = func(signal int, payload []byte) error
type healthCheckHandlerType = func() (api.HealthType, error)

// RegisterHandler -
type RegisterHandler struct {
	InitHandler        initHandlerType
	CallHandler        callHandlerType
	CheckPointHandler  checkpointHandlerType
	RecoverHandler     recoverHandlerType
	ShutDownHandler    shutdownHandlerType
	SignalHandler      signalHandlerType
	HealthCheckHandler healthCheckHandlerType
}

type posixHandlers struct {
	sync.Once
	InitHandler        initHandlerType
	CallHandler        callHandlerType
	CheckPointHandler  checkpointHandlerType
	RecoverHandler     recoverHandlerType
	ShutDownHandler    shutdownHandlerType
	SignalHandler      signalHandlerType
	HealthCheckHandler healthCheckHandlerType
	execType           execution.ExecutionType
	onceType           sync.Once
	libruntimeAPI      api.LibruntimeAPI
}

const (
	functionLibraryPath string = "YR_FUNCTION_LIB_PATH"
	initHandler         string = "INIT_HANDLER"
	callHandler         string = "CALL_HANDLER"
	checkpointHandler   string = "CHECKPOINT_HANDLER"
	recoverHandler      string = "RECOVER_HANDLER"
	shutdownHandler     string = "SHUTDOWN_HANDLER"
	signalHandler       string = "SIGNAL_HANDLER"
	healthCheckHandler  string = "HEALTH_CHECK_HANDLER"
	actorHandlersString string = "yrlib-executor.so"
	faaSHandlersString  string = "faas-executor.so"
	minPathPartLength          = 2
)

var (
	posixHandler = &posixHandlers{
		libruntimeAPI: libruntimesdkimpl.NewLibruntimeSDKImpl(),
	}
	posixHandlersEnvs []string = []string{
		initHandler, callHandler, checkpointHandler, recoverHandler, shutdownHandler, signalHandler, healthCheckHandler,
	}
	libMap map[string]*plugin.Plugin
)

// Load getEnvs to load code and setHandler
func Load() (execution.ExecutionType, error) {
	functionLibPath := os.Getenv(functionLibraryPath)
	if functionLibPath == "" {
		return execution.ExecutionTypeInvalid, fmt.Errorf("YR_FUNCTION_LIB_PATH not found")
	}

	var err error
	posixHandler.Do(
		func() {
			libMap = make(map[string]*plugin.Plugin)
			for _, envKey := range posixHandlersEnvs {
				if err = setHandler(functionLibPath, envKey); err != nil {
					fmt.Printf("failed to open lib %v\n", err)
				}
			}
		},
	)
	return posixHandler.execType, err
}

func setHandler(functionLibPath, envKey string) error {
	handler := os.Getenv(envKey)
	if handler == "" {
		return fmt.Errorf("%s not found", envKey)
	}

	symbol, err := getLib(functionLibPath, handler)
	if err != nil {
		return err
	}
	ok := true
	switch envKey {
	case initHandler:
		posixHandler.InitHandler, ok = symbol.(initHandlerType)
	case callHandler:
		posixHandler.CallHandler, ok = symbol.(callHandlerType)
	case checkpointHandler:
		posixHandler.CheckPointHandler, ok = symbol.(checkpointHandlerType)
	case recoverHandler:
		posixHandler.RecoverHandler, ok = symbol.(recoverHandlerType)
	case shutdownHandler:
		posixHandler.ShutDownHandler, ok = symbol.(shutdownHandlerType)
	case signalHandler:
		posixHandler.SignalHandler, ok = symbol.(signalHandlerType)
	case healthCheckHandler:
		posixHandler.HealthCheckHandler, ok = symbol.(healthCheckHandlerType)
	default:
		return fmt.Errorf("not support %s", handler)
	}
	if !ok {
		return fmt.Errorf("%s type error", handler)
	}
	return nil
}

func getLib(functionLibPath, handler string) (plugin.Symbol, error) {
	path, name := getLibInfo(functionLibPath, handler)
	if path == "" {
		return nil, fmt.Errorf("invalid handler name :%s", handler)
	}
	posixHandler.onceType.Do(
		func() {
			posixHandler.execType = parseHandlersType(path)
		},
	)

	lib, ok := libMap[path]
	if !ok {
		log.GetLogger().Infof("start to open lib %s", path)
		plug, err := plugin.Open(path)
		if err != nil {
			log.GetLogger().Errorf("failed to open lib %v", err)
			return nil, fmt.Errorf("failed to open %s", handler)
		}
		lib = plug
		libMap[path] = plug
	}
	symbol, err := lib.Lookup(name)
	if err != nil {
		log.GetLogger().Errorf("failed to look up %v", err)
		return nil, fmt.Errorf("failed to look up %s", name)
	}
	return symbol, nil
}

// getLibInfo will parse handler info, example:
// libPath="/tmp" libName="example.init"      --> fileName="/tmp/example.so" handlerName="init"
// libPath="/tmp" libName="test.example.init" --> fileName="/tmp/test/example.so" handlerName="init"
func getLibInfo(libPath, libName string) (string, string) {
	path := libPath
	parts := strings.Split(libName, ".")
	length := len(parts)
	handlerName := parts[length-1]
	if length < minPathPartLength {
		return "", ""
	} else if length > minPathPartLength {
		tmpPath := filepath.Join(parts[:length-minPathPartLength]...)
		path = filepath.Join(path, tmpPath)
	}
	fileName := filepath.Join(path, parts[length-minPathPartLength]+".so")

	// hack for handlers for runtime-go and libruntime both present
	if strings.HasSuffix(fileName, "faasfrontend.so") ||
		strings.HasSuffix(fileName, "faascontroller.so") ||
		strings.HasSuffix(fileName, "faasscheduler.so") ||
		strings.HasSuffix(fileName, "faasmanager.so") {
		handlerName = handlerName + "Libruntime" // _Libruntime is not good format in golang
	}
	return fileName, handlerName
}

func parseHandlersType(libPath string) execution.ExecutionType {
	if strings.HasSuffix(libPath, actorHandlersString) {
		return execution.ExecutionTypeActor
	} else if strings.HasSuffix(libPath, faaSHandlersString) {
		return execution.ExecutionTypeFaaS
	}
	return execution.ExecutionTypePosix
}

// GetExecType get execution type
func (h *posixHandlers) GetExecType() execution.ExecutionType {
	return h.execType
}

// LoadFunction load function hook
func (h *posixHandlers) LoadFunction(codePaths []string) error {
	return nil
}

// FunctionExecute function execute hook
func (h *posixHandlers) FunctionExecute(
	funcMeta api.FunctionMeta, invokeType config.InvokeType, args []api.Arg, returnobjs []config.DataObject,
) error {
	var ret []byte
	var err error
	switch invokeType {
	case config.CreateInstance, config.CreateInstanceStateless:
		ret, err = h.InitHandler(args, h.libruntimeAPI)
	case config.InvokeInstance, config.InvokeInstanceStateless:
		ret, err = h.CallHandler(args)
	default:
		err = fmt.Errorf("no such invokeType %d", invokeType)
	}

	if err != nil {
		return err
	}

	var totalNativeBufferSize uint = 0
	var do *config.DataObject
	if len(returnobjs) > 0 {
		do = &returnobjs[0]
	} else {
		return nil
	}
	if do.ID == "returnByMsg" {
		libruntime.SetReturnObject(do, uint(len(ret)))
	} else {
		if err = libruntime.AllocReturnObject(do, uint(len(ret)), []string{}, &totalNativeBufferSize); err != nil {
			return err
		}
	}

	if err = libruntime.WriterLatch(do); err != nil {
		return err
	}
	defer func() {
		if err = libruntime.WriterUnlatch(do); err != nil {
			log.GetLogger().Errorf("%v", err)
		}
	}()

	if err = libruntime.MemoryCopy(do, ret); err != nil {
		return err
	}

	if err = libruntime.Seal(do); err != nil {
		return err
	}
	return nil
}

// Checkpoint check point hook
func (h *posixHandlers) Checkpoint(checkpointID string) ([]byte, error) {
	if h.CheckPointHandler == nil {
		return nil, nil
	}
	return h.CheckPointHandler(checkpointID)
}

// Recover recover hook
func (h *posixHandlers) Recover(state []byte) error {
	if h.RecoverHandler == nil {
		return nil
	}
	return h.RecoverHandler(state, h.libruntimeAPI)
}

// Shutdown graceful shutdown hook
func (h *posixHandlers) Shutdown(gracePeriod uint64) error {
	if h.ShutDownHandler == nil {
		return nil
	}
	return h.ShutDownHandler(gracePeriod)
}

// Signal receive signal hook
func (h *posixHandlers) Signal(sig int, data []byte) error {
	if h.SignalHandler == nil {
		return nil
	}
	return h.SignalHandler(sig, data)
}

// HealthCheck -
func (h *posixHandlers) HealthCheck() (api.HealthType, error) {
	if h.HealthCheckHandler == nil {
		return api.Healthy, nil
	}
	return h.HealthCheckHandler()
}

// NewPosixFuncExecutionIntfs -
func NewPosixFuncExecutionIntfs() (execution.FunctionExecutionIntfs, error) {
	if _, err := Load(); err != nil {
		fmt.Printf("Load function execution Intfs error:%s\n", err.Error())
	}
	if posixHandler.execType != execution.ExecutionTypePosix {
		return nil, fmt.Errorf("executionType is not posix")
	}
	return posixHandler, nil
}

// NewSDKPosixFuncExecutionWithHandler -
func NewSDKPosixFuncExecutionWithHandler(registerHandler RegisterHandler) execution.FunctionExecutionIntfs {
	SDKPosixHandler := &posixHandlers{
		libruntimeAPI:      libruntimesdkimpl.NewLibruntimeSDKImpl(),
		execType:           execution.ExecutionTypePosix,
		InitHandler:        registerHandler.InitHandler,
		CallHandler:        registerHandler.CallHandler,
		CheckPointHandler:  registerHandler.CheckPointHandler,
		RecoverHandler:     registerHandler.RecoverHandler,
		ShutDownHandler:    registerHandler.ShutDownHandler,
		SignalHandler:      registerHandler.SignalHandler,
		HealthCheckHandler: registerHandler.HealthCheckHandler,
	}
	return SDKPosixHandler
}
