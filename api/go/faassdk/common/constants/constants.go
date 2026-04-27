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

// Package constants gets function services from URNs
package constants

const (
	// FuncLogLevelDebug -
	FuncLogLevelDebug = "DEBUG"
	// FuncLogLevelInfo -
	FuncLogLevelInfo = "INFO"
	// FuncLogLevelWarn -
	FuncLogLevelWarn = "WARN"
	// FuncLogLevelWarning -
	FuncLogLevelWarning = "WARNING"
	// FuncLogLevelError -
	FuncLogLevelError = "ERROR"
)

// stage
const (
	InitializeStage = "initialize"
	RestoreStage    = "restore"
	InvokeStage     = "invoke"
	LoadStage       = "load"
	ExtensionStage  = "extension"
	ColdStartStage  = "coldstart"
)

const (
	// DefaultFuncLogIndex default function log's index
	DefaultFuncLogIndex = -2
	// RuntimeLogOptTail -
	RuntimeLogOptTail = "Tail"
	// RuntimeContainerIDEnvKey -
	RuntimeContainerIDEnvKey = "DELEGATE_CONTAINER_ID"
)

const (
	// ValidBasicCreateParamSize -
	ValidBasicCreateParamSize = 4
	// ValidCustomImageCreateParamSize -
	ValidCustomImageCreateParamSize = 5
	// CustomImageUserArgIndex -
	CustomImageUserArgIndex = 4
	// ValidInvokeArgumentSize -
	ValidInvokeArgumentSize = 2
)

// CaaS alarm
const (
	// WiseCloudSite site
	WiseCloudSite = "WISECLOUD_SITE"
	// TenantID WiseCloud tenantID
	TenantID = "WISECLOUD_TENANTID"
	// ApplicationID WiseCloud applicationId
	ApplicationID = "WISECLOUD_APPLICATIONID"
	// ServiceID WiseCloud serviceId
	ServiceID = "WISECLOUD_SERVICEID"
	// ClusterName define cluster env key
	ClusterName = "CLUSTER_NAME"
	// PodNameEnvKey define pod name env key
	PodNameEnvKey = "POD_NAME"
	// PodIPEnvKey define pod ip env key
	PodIPEnvKey = "POD_IP"
	// HostIPEnvKey define pod ip env key
	HostIPEnvKey = "HOST_IP"
)

const (
	// DefaultMapSize is the default map size
	DefaultMapSize = 16
	// KernelRequestIDKey is the requestID in kernel
	KernelRequestIDKey = "requestId"
	// RuntimeTypeHttp represents runtime type of HTTP
	RuntimeTypeHttp = "http"
	// RuntimeTypeCustomContainer represents runtime type of custom container
	RuntimeTypeCustomContainer = "custom image"
)

const (
	// CaaSTraceIDHeaderKey is the key of trace ID
	CaaSTraceIDHeaderKey = "X-CaaS-Trace-Id"
	// CffRequestIDHeaderKey is the key of trace ID
	CffRequestIDHeaderKey = "X-Cff-Request-Id"
)

const (
	// RuntimeMaxRespBodySize -
	RuntimeMaxRespBodySize = 6 * 1024 * 1024
	// RuntimeRoot -
	RuntimeRoot = "/home/snuser/runtime"
	// RuntimeCodeRoot -
	RuntimeCodeRoot = "/opt/function/code"
	// RuntimeLogDir -
	RuntimeLogDir = "/home/snuser/log"
	// RuntimePkgNameSplit -
	RuntimePkgNameSplit = 2
	// INT64ToINT -
	INT64ToINT = 10
	// LDLibraryPath -
	LDLibraryPath = "LD_LIBRARY_PATH"
)
