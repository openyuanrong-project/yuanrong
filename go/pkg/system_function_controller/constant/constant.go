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

// Package constant -
package constant

import (
	"math"
	"time"
)

const (
	// ConcurrencyKey is the key for concurrency in CreateOption
	ConcurrencyKey = "ConcurrentNum"
)

const (
	// NamespaceDefault -
	NamespaceDefault = "default"
	// SystemFuncName for pod labels
	SystemFuncName = "systemFuncName"
	// FuncNameFaasfrontend -
	FuncNameFaasfrontend = "faasfrontend"
	// FuncNameFaasscheduler -
	FuncNameFaasscheduler = "faasscheduler"
	// FuncNameFaasmanager -
	FuncNameFaasmanager = "faasmanager"
	// ConcurrentNumKey -
	ConcurrentNumKey = "ConcurrentNum"
	// DefaultConcurrentNum -
	DefaultConcurrentNum = 32
	// SchedulerExclusivity -
	SchedulerExclusivity = "exclusivity"
	// RetryCounts -
	RetryCounts = 3
	// RetryInterval -
	RetryInterval = 3 * time.Second
	// DefaultInstanceNum -
	DefaultInstanceNum = 1
	// RetryIntervalIncrement -
	RetryIntervalIncrement = 2
	// MinSleepTime -
	MinSleepTime = 1 * time.Second
	// MaxSleepTime -
	MaxSleepTime = 60 * time.Second
	// SystemFunctionKinds -
	SystemFunctionKinds = 3
	// DefaultCreateRetryTime -
	DefaultCreateRetryTime = math.MaxInt
	// DefaultCreateRetryDuration -
	DefaultCreateRetryDuration = 2 * time.Second
	// DefaultCreateRetryFactor -
	DefaultCreateRetryFactor = 1
	// DefaultCreateRetryJitter -
	DefaultCreateRetryJitter = 5
	// DefaultChannelSize -
	DefaultChannelSize = 100
	// DefaultMultiples -
	DefaultMultiples = 2
	// RecreateSleepTime -
	RecreateSleepTime = 10 * MinSleepTime
	// MaxConcurrency -
	MaxConcurrency = 1000
	// InitCallTimeoutKey is the key for init call timeout in CreateOption
	InitCallTimeoutKey = "init_call_timeout"
)

const (
	// ServiceFrontendPort -
	ServiceFrontendPort = 8888
	// ServiceFrontendTargetPort -
	ServiceFrontendTargetPort = 8888
	// ServiceFrontendNodePort is k8s service NodePort
	ServiceFrontendNodePort = 31222
)
