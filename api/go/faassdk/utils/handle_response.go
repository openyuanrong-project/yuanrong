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

// Package utils
package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/types"
)

// ExecutionDuration records the execution time of Executor and user functions.
type ExecutionDuration struct {
	ExecutorBeginTime time.Time
	UserFuncBeginTime time.Time
	UserFuncTotalTime time.Duration
}

const responseHeaderSize = 2

// HandleInitResponse format init response processing error return.
func HandleInitResponse(body interface{}, statusCode int) ([]byte, error) {
	var err error
	initResponse := &types.InitResponse{}

	initResponse.ErrorCode = strconv.Itoa(statusCode)
	if initResponse.Message, err = HandleResponseBody(body); err != nil {
		errStr := fmt.Sprintf(`{"errorCode":"%d","message":"%s"}`,
			constants.ExecutorErrCodeInitFail, err.Error())
		return []byte{}, errors.New(errStr)
	}

	result, err := json.Marshal(initResponse)
	if err != nil {
		errStr := fmt.Sprintf(`{"errorCode":"%d","message":"json marsh response error:%s"}`,
			constants.ExecutorErrCodeInitFail, err.Error())
		return []byte{}, errors.New(errStr)
	}
	return []byte{}, errors.New(string(result))
}

// HandleCallResponse format call response
func HandleCallResponse(body interface{}, statusCode int, logResult string,
	totalTime ExecutionDuration, headers map[string][]string) ([]byte, error) {
	callResponse := &types.CallResponse{}
	var err error
	callResponse.Headers = HandleCallResponseHeaders(headers)
	if callResponse.Body, err = HandleResponseBody(body); err != nil {
		return []byte{}, fmt.Errorf("handle response header error: %s", err)
	}
	callResponse.BillingDuration = "this is billing duration TODO"
	callResponse.InnerCode = strconv.Itoa(statusCode)
	callResponse.InvokerSummary = "this is summary TODO"
	callResponse.LogResult = encodeBase64([]byte("this is user log TODO"))
	if !totalTime.ExecutorBeginTime.IsZero() {
		callResponse.ExecutorTime = float64(time.Since(totalTime.ExecutorBeginTime).Milliseconds())
	}
	callResponse.UserFuncTime = float64(totalTime.UserFuncTotalTime.Milliseconds())
	if logResult != "" {
		callResponse.LogResult = logResult
	}
	result, err := json.Marshal(callResponse)
	if err != nil {
		return []byte{}, fmt.Errorf("json marsh response error:%s", err)
	}

	return result, nil
}

// HandleResponseBody convert interface body to []byte or handle error message
func HandleResponseBody(body interface{}) ([]byte, error) {
	if _, ok := body.([]byte); ok {
		return body.([]byte), nil
	}
	if result, err := json.Marshal(body); err != nil {
		return []byte{}, fmt.Errorf("json marshal response error:%s", err)
	} else {
		return result, nil
	}
}

// HandleCallResponseHeaders deal with response header
func HandleCallResponseHeaders(rawHeaders map[string][]string) map[string]string {
	headers := make(map[string]string, responseHeaderSize)
	headers["Content-Type"] = "application/json"
	headers["X-Log-Type"] = "base64"
	if rawHeaders != nil {
		for k, v := range rawHeaders {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
	}
	return headers
}

func encodeBase64(data []byte) string {
	base64Result := base64.StdEncoding.EncodeToString(data)
	return base64Result
}
