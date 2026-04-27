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
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/utils"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const defaultResNums = 2

type timeoutCallBack func() ([]byte, error)
type respCallBack func(res []interface{}) ([]byte, error)
type handleErrorRespCallBack func(body interface{}, statusCode int) ([]byte, error)

func (eh *EventHandler) initWithTimeout() ([]byte, error) {
	timeout := eh.initTimeout
	resCh := make(chan []interface{}, 1)
	go func() {
		eh.userInitEntry(updateInvokeContext(map[string]string{}))
		resCh <- nil
	}()

	handleErrorResp := func(body interface{}, statusCode int) ([]byte, error) {
		return utils.HandleInitResponse(body, statusCode)
	}
	timeoutCB := func() ([]byte, error) {
		logger.GetLogger().Errorf("runtime initialization timed out after %ds", timeout)
		return handleErrorResp(fmt.Sprintf("runtime initialization timed out after %ds",
			timeout), constants.InitFunctionTimeout)
	}
	respCB := func(res []interface{}) ([]byte, error) {
		return []byte("success"), nil
	}

	return waitRespProcess(time.Duration(timeout), resCh, timeoutCB, respCB, handleErrorResp)
}

func (eh *EventHandler) invokeWithTimeout(event []byte, context map[string]string,
	traceID string, totalTime utils.ExecutionDuration) ([]byte, error) {
	resCh := make(chan []interface{}, 1)
	logWithID := logger.GetLogger().With(zap.Any("traceID", traceID))
	go func() {
		totalTime.UserFuncBeginTime = time.Now()
		userRes, userErr := eh.userCallEntry(event, updateInvokeContext(context))
		resAndErr := []interface{}{userRes, userErr}
		resCh <- resAndErr
		totalTime.UserFuncTotalTime = time.Since(totalTime.UserFuncBeginTime)
	}()

	handleErrorResp := func(body interface{}, statusCode int) ([]byte, error) {
		return utils.HandleCallResponse(body, statusCode, "", totalTime, nil)
	}
	timeout := eh.invokeTimeout
	timeoutCB := func() ([]byte, error) {
		totalTime.UserFuncTotalTime = time.Since(totalTime.UserFuncBeginTime)
		logWithID.Errorf("call invoke timeout %d", timeout)
		return handleErrorResp(fmt.Sprintf("call invoke timeout %ds",
			timeout), constants.InvokeFunctionTimeout)
	}
	respCB := func(res []interface{}) ([]byte, error) {
		if len(res) < defaultResNums {
			logWithID.Errorf("invalid res nums")
			return handleErrorResp(fmt.Sprintf("invalid res num, traceID: %s", traceID),
				constants.FunctionRunError)
		}
		userRes := res[0]
		userErr := res[1]
		if userErr != nil {
			logWithID.Errorf("faas invoke user function,error: %s", userErr.(error))
			return handleErrorResp(fmt.Sprintf("faas invoke user function error: %s", userErr.(error)),
				constants.FunctionRunError)
		}
		return utils.HandleCallResponse(userRes, constants.NoneError, "", totalTime, nil)
	}

	return waitRespProcess(time.Duration(timeout), resCh, timeoutCB, respCB, handleErrorResp)
}

func waitRespProcess(timeout time.Duration, resCh <-chan []interface{}, timeoutCB timeoutCallBack,
	respCallBack respCallBack, handleErrorResp handleErrorRespCallBack) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		logger.GetLogger().Errorf("invoke timed out after %d ms", timeout)
		return timeoutCB()
	case _, ok := <-stopCh:
		logger.GetLogger().Warnf("failed to call invoke method, the runtime process exit")
		if !ok {
			return handleErrorResp("failed to call invoke method, the runtime process exit",
				constants.FunctionRunError)
		}
	case res, ok := <-resCh:
		if !ok {
			return handleErrorResp("invoke response channel closed error", constants.ExecutorErrCodeInitFail)
		}
		return respCallBack(res)
	}
	return []byte{}, nil
}
