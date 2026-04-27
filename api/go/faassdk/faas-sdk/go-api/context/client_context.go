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

// Package context for userCode
package context

import (
	"yuanrong.org/kernel/runtime/faassdk/faas-sdk/pkg/runtime/userlog"
)

// RuntimeContext for userCode
type RuntimeContext interface {
	GetLogger() userlog.RuntimeLogger

	GetRequestID() string

	GetRemainingTimeInMilliSeconds() int

	GetAccessKey() string

	GetSecretKey() string

	GetUserData(string) string

	GetFunctionName() string

	GetRunningTimeInSeconds() int

	GetVersion() string

	GetMemorySize() int

	GetCPUNumber() int

	GetPackage() string

	GetToken() string

	GetProjectID() string

	GetState() string

	SetState(string) error

	GetInvokeProperty() string

	GetTraceID() string

	GetInvokeID() string

	GetAlias() string

	GetSecurityToken() string
}
