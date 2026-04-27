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

const (
	// BusinessTypeWebSocket websocket business type
	BusinessTypeWebSocket = "WEBSOCKET"
	// BusinessTypeCAE cae business type
	BusinessTypeCAE = "CAE"
	// BusinessTypeFaaS FaaS business type
	BusinessTypeFaaS = "FaaS"
	// BusinessType business type key
	BusinessType = "BUSINESS_TYPE"
	// LanguageJava8 language java8
	LanguageJava8 = "java8"
)

const (
	// ZoneKey zone key
	ZoneKey = "KUBERNETES_IO_AVAILABLEZONE"
	// ZoneNameLen define zone length
	ZoneNameLen = 255
	// DefaultAZ default az
	DefaultAZ = "defaultaz"

	// PodNamespaceEnvKey define pod namespace env key
	PodNamespaceEnvKey = "POD_NAMESPACE"

	// FunctionLoadTimeoutEnvKey load Function timeout time
	FunctionLoadTimeoutEnvKey = "LOAD_FUNCTION_TIMEOUT"

	// ResourceLimitsMemory Memory limit, in bytes
	ResourceLimitsMemory = "MEMORY_LIMIT_BYTES"

	// FuncBranchEnvKey is branch env key
	FuncBranchEnvKey = "FUNC_BRANCH"

	// DataSystemBranchEnvKey is branch env key
	DataSystemBranchEnvKey = "DATASYSTEM_CAPABILITY"

	// HTTPort busproxy httpserver listen port
	HTTPort = "22423"
	// TCPort busproxy tcpserver listen port
	TCPort = "32568"
	// BusWorkerServerTCPort bus worker server listen port
	BusWorkerServerTCPort = "32569"
	// BusRuntimeServerPort bus listen port for
	BusRuntimeServerPort = "32570"
	// DefaultCachePort indicates the default port of a cache-manager server
	DefaultCachePort = "9993"

	// PlatformTenantID is tenant ID of platform function
	PlatformTenantID = "0"

	// RuntimeLogOptTail -
	RuntimeLogOptTail = "Tail"
	// RuntimeLayerDirName -
	RuntimeLayerDirName = "layer"
	// RuntimeFuncDirName -
	RuntimeFuncDirName = "func"

	// FunctionTaskAppID -
	FunctionTaskAppID = "function-task"

	// BackpressureCode indicate that frontend should choose another proxy/worker and retry
	BackpressureCode = 211429
	// HeaderBackpressure indicate that proxy can backpressure this request
	HeaderBackpressure = "X-Backpressure"
	// HeaderBackpressureNums Backpressure numbers counter
	HeaderBackpressureNums = "X-Backpressure-Nums"
	// MonitorFileName monitor file name
	MonitorFileName = "monitor-disk"

	// DefaultFuncLogIndex default function log's index
	DefaultFuncLogIndex = -2

	// IsClusterUpgrading indicate that the cluster is in upgrading phase
	IsClusterUpgrading = "FAAS_CLUSTER_IS_UPGRADING"
)

const (
	// WorkerManagerApplier mark the instance is created by minInstance
	WorkerManagerApplier = "worker-manager"
	// ASBResApplier mark the instance is created by ASBRes
	ASBResApplier = "ASBRes"
	// FunctionTaskApplier mark the instance is created by minInstance
	FunctionTaskApplier = "functiontask"
	// PredictionApplier mark the instance is created by smart warmer predict
	PredictionApplier = "prediction"
	// FaasSchedulerApplier the instance is created by faas scheduler
	FaasSchedulerApplier = "faas-scheduler"
	// PoolInfoPrefix pool info prefix in redis
	PoolInfoPrefix = "ClusterState_Pool"
	// PoolInfoSep pool info separator in redis
	PoolInfoSep = "_"
	// ClusterIDKey cluster id key in system env
	ClusterIDKey = "CLUSTER_ID"
	// DefaultRecordingInterval default pool info recording interval, unit is second
	DefaultRecordingInterval = 5
	// DefaultRecordExpiredTime default pool info record expired time, unit is second
	DefaultRecordExpiredTime = 900
)

const (
	// FunctionAccessor - defines the microservice component name.
	FunctionAccessor = "FunctionAccessor"
	// FunctionTask -
	FunctionTask = "FunctionTask"
	// InstanceManager -
	InstanceManager = "FunctionInstanceManager"
	// StateManager -
	StateManager = "StateManager"
	// CacheManager -
	CacheManager = "CacheManager"
	// CacheServiceName indicates the header of the cache service
	CacheServiceName = "cache-manager"
	// FaaSScheduler -
	FaaSScheduler = "faas-scheduler"
	// SnapshotManager -
	SnapshotManager = "SnapshotManager"
	// Autoscaler define the alarm type
	Autoscaler = "Autoscaler"
)

// header constant key for FG
const (
	// FGHeaderRequestID -
	FGHeaderRequestID = "X-Request-Id"
	// FGHeaderAccessKey -
	FGHeaderAccessKey = "X-Access-Key"
	// FGHeaderSecretKey -
	FGHeaderSecretKey = "X-Secret-Key"
	// FGHeaderSecurityAccessKey -
	FGHeaderSecurityAccessKey = "X-Security-Access-Key"
	// FGHeaderSecuritySecretKey -
	FGHeaderSecuritySecretKey = "X-Security-Secret-Key"
	// FGHeaderAuthToken -
	FGHeaderAuthToken = "X-Auth-Token"
	// FGHeaderSecurityToken -
	FGHeaderSecurityToken = "X-Security-Token"
)
