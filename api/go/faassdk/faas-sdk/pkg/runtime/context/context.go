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
	"yuanrong.org/kernel/runtime/faassdk/faas-sdk/pkg/runtime/userlog"
)

// Env ContextEnv for InvokeHandler
type Env struct {
	rtProjectID          string
	rtFcName             string
	rtFcVersion          string
	rtPackage            string
	rtMemory             int
	rtCPU                int
	rtTimeout            int
	rtStartTime          int
	rtUserData           map[string]string
	rtInitializerTimeout int
}

// InvokeContext InvokeContext
type InvokeContext struct {
	RequestID      string `json:"request_id"`
	Alias          string `json:"alias"`
	InvokeProperty string `json:"invoke_property"`
	InvokeID       string `json:"invoke_id"`
	TraceID        string `json:"trace_id"`
}

// Provider ContextProvider for RuntimeContext
type Provider struct {
	CtxEnv      *Env
	CtxHTTPHead *InvokeContext
	Logger      userlog.RuntimeLogger
}
