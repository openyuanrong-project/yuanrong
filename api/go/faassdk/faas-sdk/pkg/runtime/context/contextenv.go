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

// Package context
package context

import (
	"os"
	"time"

	"yuanrong.org/kernel/runtime/faassdk/faas-sdk/pkg/runtime/userlog"
)

// GetRemainingTimeInMilliSeconds get remaining time
func (p Provider) GetRemainingTimeInMilliSeconds() int {
	timeoutInMilliSeconds := p.CtxEnv.rtTimeout * (int)(time.Millisecond)
	remainingTime := p.CtxEnv.rtStartTime + (timeoutInMilliSeconds) - getCurrentTime()
	if remainingTime < 0 {
		return 0
	}
	return remainingTime
}

// GetFunctionName get this functionName
func (p Provider) GetFunctionName() string {
	return p.CtxEnv.rtFcName
}

// GetRunningTimeInSeconds get timeout interval
func (p Provider) GetRunningTimeInSeconds() int {
	return p.CtxEnv.rtTimeout
}

// GetVersion get runtime version
func (p Provider) GetVersion() string {
	return p.CtxEnv.rtFcVersion
}

// GetMemorySize get memory size of runtime instances
func (p Provider) GetMemorySize() int {
	return p.CtxEnv.rtMemory
}

// GetCPUNumber get CPU usage of runtime instance
func (p Provider) GetCPUNumber() int {
	return p.CtxEnv.rtCPU
}

// GetUserData get userData from env
func (p Provider) GetUserData(key string) string {
	if p.CtxEnv.rtUserData != nil {
		return p.CtxEnv.rtUserData[key]
	}
	return ""
}

// GetLogger get logger
func (p Provider) GetLogger() userlog.RuntimeLogger {
	return p.Logger
}

// GetProjectID get projectId
func (p Provider) GetProjectID() string {
	return p.CtxEnv.rtProjectID
}

// GetPackage get package
func (p Provider) GetPackage() string {
	return p.CtxEnv.rtPackage
}

// GetAccessKey get AccessKey
func (p Provider) GetAccessKey() string {
	return ""
}

// GetSecretKey get SecretKey
func (p Provider) GetSecretKey() string {
	return ""
}

// GetToken get token
func (p Provider) GetToken() string {
	return ""
}

// GetRequestID get requestId
func (p Provider) GetRequestID() string {
	return p.CtxHTTPHead.RequestID
}

// GetState get instance status
func (p Provider) GetState() string {
	return os.Getenv("USER_DEFINED_STATE")
}

// SetState set instance status
func (p Provider) SetState(state string) error {
	err := os.Setenv("USER_DEFINED_STATE", state)
	if err != nil {
		return err
	}
	return nil
}

// GetInvokeProperty get invoke property
func (p Provider) GetInvokeProperty() string {
	return p.CtxHTTPHead.InvokeProperty
}

// GetTraceID get traceId
func (p Provider) GetTraceID() string {
	return p.CtxHTTPHead.TraceID
}

// GetInvokeID  get invokeId
func (p Provider) GetInvokeID() string {
	return p.CtxHTTPHead.InvokeID
}

// GetAlias get function alias
func (p Provider) GetAlias() string {
	return p.CtxHTTPHead.Alias
}

// GetSecurityToken get token
func (p Provider) GetSecurityToken() string {
	return ""
}
