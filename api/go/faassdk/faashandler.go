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
	"errors"
	"fmt"
	"sync"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/handler"
	"yuanrong.org/kernel/runtime/faassdk/handler/event"
	"yuanrong.org/kernel/runtime/faassdk/handler/http"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/libruntime"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
	"yuanrong.org/kernel/runtime/libruntime/config"
	"yuanrong.org/kernel/runtime/libruntime/execution"
	"yuanrong.org/kernel/runtime/libruntime/libruntimesdkimpl"
)

type faasHandlers struct {
	execType      execution.ExecutionType
	faasexecutor  handler.ExecutorHandler
	libruntimeAPI api.LibruntimeAPI
}

var (
	faasHdlrs faasHandlers = faasHandlers{
		execType:      execution.ExecutionTypeFaaS,
		libruntimeAPI: libruntimesdkimpl.NewLibruntimeSDKImpl(),
	}
)

// GetExecType get execution type
func (h *faasHandlers) GetExecType() execution.ExecutionType {
	return h.execType
}

// LoadFunction load user function
func (h *faasHandlers) LoadFunction(codePaths []string) error {
	// 初始化日志
	logger.SetupLogger(h.libruntimeAPI.GetFormatLogger())
	return nil
}

// FunctionExecute function execute hook
func (h *faasHandlers) FunctionExecute(funcMeta api.FunctionMeta, invokeType config.InvokeType,
	args []api.Arg, returnobjs []config.DataObject) error {
	var ret []byte
	var err error
	switch invokeType {
	case config.CreateInstance, config.CreateInstanceStateless:
		funcSpec := getFuncSpecFromArgs(args)
		if funcSpec == nil {
			return errors.New("invalid funcSpec")
		}
		h.InitFaaSExecutor(funcSpec, h.libruntimeAPI)
		if h.faasexecutor != nil {
			ret, err = h.faasexecutor.InitHandler(args, h.libruntimeAPI)
		}
	case config.InvokeInstance, config.InvokeInstanceStateless:
		traceID := string(args[0].Data)
		context := map[string]string{"traceID": traceID}
		ret, err = h.faasexecutor.CallHandler(args, context)
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
	}
	if do != nil && do.ID == "returnByMsg" {
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

// Checkpoint check point
func (h *faasHandlers) Checkpoint(checkpointID string) ([]byte, error) {
	if h.faasexecutor == nil {
		return nil, errors.New("faasexcutor not initialized")
	}
	return h.faasexecutor.CheckPointHandler(checkpointID)
}

// Recover recover hook
func (h *faasHandlers) Recover(state []byte) error {
	if h.faasexecutor == nil {
		return errors.New("faasexcutor not initialized")
	}
	return h.faasexecutor.RecoverHandler(state)
}

// Shutdown hook
func (h *faasHandlers) Shutdown(gracePeriod uint64) error {
	if h.faasexecutor == nil {
		return errors.New("faasexcutor not initialized")
	}
	return h.faasexecutor.ShutDownHandler(gracePeriod)
}

// Signal hook
func (h *faasHandlers) Signal(sig int, data []byte) error {
	if h.faasexecutor == nil {
		return errors.New("faasexcutor not initialized")
	}
	waitFaaSExecutorDone()
	return h.faasexecutor.SignalHandler(int32(sig), data)
}

// HealthCheck 函数是一个健康检查函数，用于检查 faaexecutor 的状态
func (h *faasHandlers) HealthCheck() (api.HealthType, error) {
	if h.faasexecutor == nil {
		return api.Healthy, nil
	}
	waitFaaSExecutorDone()
	return h.faasexecutor.HealthCheckHandler()
}

func newFaaSFuncExecutionIntfs() execution.FunctionExecutionIntfs {
	return &faasHdlrs
}

func getFuncSpecFromArgs(args []api.Arg) *types.FuncSpec {
	if len(args) < 1 {
		logger.GetLogger().Errorf("invalid args number %d", len(args))
		return nil
	}
	funcSpec := &types.FuncSpec{}
	err := json.Unmarshal(args[0].Data, funcSpec)
	if err != nil {
		logger.GetLogger().Errorf("failed to unmarshal funcSpec %s", err.Error())
		return nil
	}
	return funcSpec
}

var (
	faasExecutorOnce     sync.Once
	faaExeecutorDoneChan = make(chan struct{})
)

func (h *faasHandlers) InitFaaSExecutor(funcSpec *types.FuncSpec, client api.LibruntimeAPI) {
	faasExecutorOnce.Do(func() {
		switch funcSpec.FuncMetaData.Runtime {
		case constants.RuntimeTypeHttp:
			h.faasexecutor = http.NewHttpHandler(funcSpec, client)
		case constants.RuntimeTypeCustomContainer:
			h.faasexecutor = http.NewCustomContainerHandler(funcSpec, client)
		default:
			h.faasexecutor = event.NewEventHandler(funcSpec, client)
		}
		close(faaExeecutorDoneChan)
	})
}

func waitFaaSExecutorDone() {
	select {
	case <-faaExeecutorDoneChan:
	}
}
