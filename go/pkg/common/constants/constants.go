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

// Package constants implements vars of all
package constants

import (
	"os"
	"strconv"
	"time"
)

const (
	// ZoneKey zone key
	ZoneKey = "KUBERNETES_IO_AVAILABLEZONE"
	// ZoneNameLen define zone length
	ZoneNameLen = 255
	// DefaultAZ default az
	DefaultAZ = "defaultaz"

	// PodIPEnvKey define pod ip env key
	PodIPEnvKey = "POD_IP"

	// HostNameEnvKey defines the hostname env key
	HostNameEnvKey = "HOSTNAME"

	// NodeID defines the node name env key
	NodeID = "NODE_ID"

	// HostIPEnvKey defines the host ip env key
	HostIPEnvKey = "HOST_IP"

	// PodNamespaceEnvKey define pod namespace env key
	PodNamespaceEnvKey = "POD_NAMESPACE"

	// ResourceLimitsMemory Memory limit, in bytes
	ResourceLimitsMemory = "MEMORY_LIMIT_BYTES"

	// ResourceLimitsCPU CPU limit, in m(1/1000)
	ResourceLimitsCPU = "CPU_LIMIT"

	// FuncBranchEnvKey is branch env key
	FuncBranchEnvKey = "FUNC_BRANCH"

	// DataSystemBranchEnvKey is branch env key
	DataSystemBranchEnvKey = "DATASYSTEM_CAPABILITY"

	// HTTPort busproxy httpserver listen port
	HTTPort = "22423"
	// GRPCPort busproxy gRPCserver listen port
	GRPCPort = "22769"
	// WorkerAgentPort is the listen port of worker agent grpc server
	WorkerAgentPort = "22888"
	// DataSystemPort is the port of data system
	DataSystemPort = "31501"
	// LocalSchedulerPort is the listen port string of local scheduler grpc server
	LocalSchedulerPort = GRPCPort
	// DomainSchedulerPort is the listen port of domain scheduler grpc server
	DomainSchedulerPort = 22771
	// MaxPort maximum number of ports
	MaxPort = 65535
	// SchedulerAddressSeparator is the separator of domain scheduler address
	SchedulerAddressSeparator = ":"
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

	// TenantID config from function task
	TenantID = "0"

	// BackpressureCode indicate that frontend should choose another proxy/worker and retry
	BackpressureCode = 211429
	// HeaderBackpressure indicate that proxy can backpressure this request
	HeaderBackpressure = "X-Backpressure"

	// SrcInstanceID gRPC context of metadata
	SrcInstanceID = "src_instance_id"
	// ReturnObjID gRPC context of metadata
	ReturnObjID = "return_obj_id"

	// DelWorkerAgentEvent delete workerAgent
	DelWorkerAgentEvent = "WorkerAgent-Del"
	// UpdWorkerAgentEvent update workerAgent
	UpdWorkerAgentEvent = "WorkerAgent-Upd"

	// DefaultLatestVersion is default function name
	DefaultLatestVersion = "$latest"
	// DefaultLatestFaaSVersion is default faas function name
	DefaultLatestFaaSVersion = "latest"
	// DefaultJavaRuntimeName is default java runtime name
	DefaultJavaRuntimeName = "java1.8"
	// DefaultJavaRuntimeNameForFaas is defualt
	DefaultJavaRuntimeNameForFaas = "java8"
)

// grpc parameters
const (
	// MaxMsgSize grpc client max message size(bit)
	MaxMsgSize = 1024 * 1024 * 2
	// MaxWindowSize grpc flow control window size(bit)
	MaxWindowSize = 1024 * 1024 * 2
	// MaxBufferSize grpc read/write buffer size(bit)
	MaxBufferSize = 1024 * 1024 * 2
)

// functionBus userData key flag
const (
	// FrontendCallFlag invoke from task
	FrontendCallFlag = "FrontendCallFlag"
)

const (
	// DynamicRouterParamPrefix 动态路由参数前缀
	DynamicRouterParamPrefix = "/:"
)

// HTTP invoke request header key
const (
	// HeaderExecutedDuration -
	HeaderExecutedDuration = "X-Executed-Duration"
	// HeaderTraceID -
	HeaderTraceID = "X-Trace-Id"
	// HeaderEventSourceID -
	HeaderEventSourceID = "X-Event-Source-Id"
	// HeaderBusinessID -
	HeaderBusinessID = "X-Business-ID"
	// HeaderTenantID -
	HeaderTenantID = "X-Tenant-ID"
	// HeaderTenantId -
	HeaderTenantId = "X-Tenant-Id"
	// HeaderPoolLabel -
	HeaderPoolLabel = "X-Pool-Label"
	// HeaderLogType -
	HeaderLogType = "X-Log-Type"
	// HeaderLogResult -
	HeaderLogResult = "X-Log-Result"
	// HeaderTriggerFlag -
	HeaderTriggerFlag = "X-Trigger-Flag"
	// HeaderInnerCode -
	HeaderInnerCode = "X-Inner-Code"
	// HeaderInvokeURN  -
	HeaderInvokeURN = "X-Tag-VersionUrn"
	// HeaderStateKey -
	HeaderStateKey = "X-State-Key"
	// HeaderCallType is the request type
	HeaderCallType = "X-Call-Type"
	// HeaderLoadDuration duration of loading function
	HeaderLoadDuration = "X-Load-Duration"
	// HeaderNodeLabel is node label
	HeaderNodeLabel = "X-Node-Label"
	// HeaderForceDeploy is Force Deploy
	HeaderForceDeploy = "X-Force-Deploy"
	// HeaderAuthorization is authorization
	HeaderAuthorization = "authorization"
	// HeaderFutureID is futureID of invocation
	HeaderFutureID = "X-Future-ID"
	// HeaderAsync indicate whether it is an async request
	HeaderAsync = "X-ASYNC"
	// HeaderRuntimeID represents runtime instance identification
	HeaderRuntimeID = "X-Runtime-ID"
	// HeaderRuntimePort represents runtime rpc port
	HeaderRuntimePort = "X-Runtime-Port"
	// HeaderCPUSize is cpu size specified by invoke
	HeaderCPUSize = "X-Instance-CPU"
	// HeaderMemorySize is cpu memory specified by invoke
	HeaderMemorySize = "X-Instance-Memory"
	HeaderFileDigest = "X-File-Digest"
	HeaderProductID  = "X-Product-Id"
	HeaderPrivilege  = "X-Privilege"
	HeaderUserID     = "X-User-Id"
	HeaderVersion    = "X-Version"
	HeaderKind       = "X-Kind"
	// HeaderCompatibleRuntimes -
	HeaderCompatibleRuntimes = "X-Header-Compatible-Runtimes"
	// HeaderDescription -
	HeaderDescription = "X-Description"
	// HeaderLicenseInfo -
	HeaderLicenseInfo = "X-License-Info"
	// HeaderGroupID is group id
	HeaderGroupID = "X-Group-ID"
	// ApplicationJSON -
	ApplicationJSON = "application/json"
	// ContentType -
	ContentType = "Content-Type"
	// PriorityHeader -
	PriorityHeader = "priority"
	// HeaderDataContentType -
	HeaderDataContentType = "X-Content-Type"
	// ErrorDuration duration when error happened,
	// used with key $LoadDuration
	ErrorDuration = -1
)

// Extra Request Header
const (
	// HeaderRequestID -
	HeaderRequestID = "x-request-id"
	// HeaderAccessKey -
	HeaderAccessKey = "x-access-key"
	// HeaderSecretKey -
	HeaderSecretKey = "x-secret-key"
	// HeaderAuthToken -
	HeaderAuthToken = "x-auth-token"
	// HeaderSecurityToken -
	HeaderSecurityToken = "x-security-token"
	// HeaderStorageType code storage type
	HeaderStorageType = "x-storage-type"
)

const (
	// FunctionStatusUnavailable function status is unavailable
	FunctionStatusUnavailable = "unavailable"

	// FunctionStatusAvailable function status is available
	FunctionStatusAvailable = "available"
)

const (
	// OndemandKey is used in ondemand scenario
	OndemandKey = "ondemand"
)

// stage
const (
	InitializeStage = "initialize"
)

// default UIDs and GIDs
const (
	DefaultWorkerGID    = 1002
	DefaultRuntimeUID   = 1003
	DefaultRuntimeUName = "snuser"
	DefaultRuntimeGID   = 1003
)

const (
	// WorkerManagerApplier mark the instance is created by minInstance
	WorkerManagerApplier = "worker-manager"
)

const (
	DialBaseDelay       = 300 * time.Millisecond
	DialMultiplier      = 1.2
	DialJitter          = 0.1
	DialMaxDelay        = 15 * time.Second
	RuntimeDialMaxDelay = 100 * time.Second
)

// constants of network connection
const (
	// DefaultConnectInterval is the default connect interval
	DefaultConnectInterval = 3 * time.Second
	// DefaultDialInterval is the default grpc dial request interval
	DefaultDialInterval = 3 * time.Second
	// DefaultRetryTimes is the default request retry times
	DefaultRetryTimes   = 3
	ConnectIntervalTime = 1 * time.Second
)

// request message
const (
	// RequestCPU -
	RequestCPU = "CPU"
	// RequestMemory -
	RequestMemory = "Memory"
	// MinCustomResourcesSize is min gpu size of invoke
	MinCustomResourcesSize = 0

	// CpuUnitConvert -
	CpuUnitConvert = 1000
	// MemoryUnitConvert -
	MemoryUnitConvert = 1024

	// minInvokeCPUSize is default min cpu size of invoke (One CPU core corresponds to 1000)
	minInvokeCPUSize = 300
	// MaxInvokeCPUSize is max cpu size of invoke (One CPU core corresponds to 1000)
	MaxInvokeCPUSize = 16000
	// minInvokeMemorySize is default min memory size of invoke (MB)
	minInvokeMemorySize = 128
	// MaxInvokeMemorySize is max memory size of invoke (MB)
	MaxInvokeMemorySize = 1024 * 1024 * 1024
	// InstanceConcurrency -
	InstanceConcurrency = "Concurrency"
	// DefaultMapSize default map size
	DefaultMapSize = 2
	// DefaultSliceSize default slice size
	DefaultSliceSize = 16
	// MaxUploadMemorySize is max memory size of upload (MB)
	MaxUploadMemorySize = 10 * 1024 * 1024
	// S3StorageType the code is stored in the minio
	S3StorageType = "s3"
	// LocalStorageType the code is stored in the disk
	LocalStorageType = "local"
	// CopyStorageType the code is stored in the disk and need to copy to container path
	CopyStorageType = "copy"
	// Faas kind of function creation
	Faas = "faas"
)

// prefixes of ETCD keys
const (
	WorkerETCDKeyPrefix = "/sn/workeragent"
	NodeETCDKeyPrefix   = "/sn/node"
	// InstanceETCDKeyPrefix is the prefix of etcd key for instance
	InstanceETCDKeyPrefix = "/sn/instance"
	// ResourceGroupETCDKeyPrefix is the prefix of etcd key for resource group
	ResourceGroupETCDKeyPrefix = "/sn/resourcegroup"
	// WorkersEtcdKeyPrefix is the prefix of etcd key for workers
	WorkersEtcdKeyPrefix = "/sn/workers"
	// AliasEtcdKeyPrefix is the key prefix of aliases in etcd
	AliasEtcdKeyPrefix = "/sn/aliases"
)

// constants of posix custom runtime
const (
	PosixCustomRuntime = "posix-custom-runtime"
	GORuntime          = "go"
	JavaRuntime        = "java"
	_
)

const (
	// OriginSchedulePolicy use origin scheduler policy
	OriginSchedulePolicy = 0
	// NewSchedulePolicy use new scheduler policy
	NewSchedulePolicy = 1
)

const (
	// LocalSchedulerLevel local scheduler level is 0
	LocalSchedulerLevel = iota
	// LowDomainSchedulerLevel low domain scheduler level is 0
	LowDomainSchedulerLevel
)

const (
	// Base10 is the decimal base number when use FormatInt
	Base10 = 10
)

// MinInvokeCPUSize is min cpu size of invoke (One CPU core corresponds to 1000)
// Return default minInvokeCPUSize or system env[MinInvokeCPUSize]
var MinInvokeCPUSize = func() float64 {
	minInvokeCPUSizeStr := os.Getenv("MinInvokeCPUSize")
	if minInvokeCPUSizeStr != "" {
		value, err := strconv.Atoi(minInvokeCPUSizeStr)
		if err != nil {
			return minInvokeCPUSize
		}
		return float64(value)
	}
	return minInvokeCPUSize
}()

// MinInvokeMemorySize is min memory size of invoke (MB)
// Return default minInvokeMemorySize or system env[MinInvokeMemorySize]
var MinInvokeMemorySize = func() float64 {
	minInvokeMemorySizeStr := os.Getenv("MinInvokeMemorySize")
	if minInvokeMemorySizeStr != "" {
		value, err := strconv.Atoi(minInvokeMemorySizeStr)
		if err != nil {
			return minInvokeMemorySize
		}
		return float64(value)
	}
	return minInvokeMemorySize
}()

// SelfNodeIP - node IP
var SelfNodeIP = os.Getenv(HostIPEnvKey)

// SelfNodeID - node ID
var SelfNodeID = os.Getenv(NodeID)
