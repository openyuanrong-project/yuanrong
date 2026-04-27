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

// Package constants
package constants

const (
	// NoneError -
	NoneError = 0
	// FaaSError -
	FaaSError = 500
)

// executor error code
const (
	// ExecutorErrCodeInitFail -
	ExecutorErrCodeInitFail = 6001
	// AcquireLeaseTrafficLimitErrorCode is reach max limit of acquiring lease concurrently
	AcquireLeaseTrafficLimitErrorCode = 6037
)

// user error code
const (
	// EntryNotFound user code entry not found
	EntryNotFound = 4001
	// FunctionRunError user function failed to run
	FunctionRunError = 4002
	// StateContentTooLarge state content is too large
	StateContentTooLarge = 4003
	// ResponseExceedLimit response of user function exceeds the platform limit
	ResponseExceedLimit = 4004
	// UndefinedState state is undefined
	UndefinedState = 4005
	// HeartBeatFunctionInvalid heart beat function of user invalid
	HeartBeatFunctionInvalid = 4006
	// FunctionResultInvalid user function result is invalid
	FunctionResultInvalid = 4007
	// InitializeFunctionError user initialize function error
	InitializeFunctionError = 4009
	// HeartBeatInvokeError failed to invoke heart beat function
	HeartBeatInvokeError = 4010
	// InvokeFunctionTimeout user function invoke timeout
	InvokeFunctionTimeout = 4010

	// InstanceCircuitBreakError function is circuit break
	InstanceCircuitBreakError = 4011

	// InitFunctionTimeout user function init timeout
	InitFunctionTimeout = 4211
	// RequestBodyExceedLimit request body exceeds limit
	RequestBodyExceedLimit = 4140
	// InitFunctionFail function initialization failed
	InitFunctionFail = 4201
	// MemoryLimitExceeded	runtime memory limit exceeded
	MemoryLimitExceeded = 4205
	// DiskUsageExceed disk usage exceed code
	DiskUsageExceed = 4207
)

// frontend error code
const (
	// FrontendStatusOk ok code
	FrontendStatusOk = 200200
	// ClusterIsUpgrading -
	ClusterIsUpgrading = 150439
)
