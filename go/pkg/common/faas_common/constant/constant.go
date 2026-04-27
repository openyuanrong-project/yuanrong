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

import "time"

const (
	// LibruntimeHeaderSize is the header length of libruntime package
	LibruntimeHeaderSize = 16
)

const (
	// BusinessTypeFG is the business type of FunctionGraph
	BusinessTypeFG = iota
	// BusinessTypeWiseCloud is the business type of WiseCloud
	BusinessTypeWiseCloud
)

const (
	// BackendTypeKernel -
	BackendTypeKernel = iota
	// BackendTypeFG -
	BackendTypeFG
)

const (
	// DeployModeContainer -
	DeployModeContainer = "Container"
	// DeployModeProcesses -
	DeployModeProcesses = "Processes"
)

const (
	// HeaderRequestID -
	HeaderRequestID = "X-Request-Id"
	// HeaderTraceID -
	HeaderTraceID = "X-Trace-Id"
	// HeaderTraceParent -
	HeaderTraceParent = "Traceparent"
)

const (
	// KernelResourceNotEnoughErrCode is the error code of kernel resource not enough
	KernelResourceNotEnoughErrCode = 1002
	// KernelInnerSystemErrCode is the error code of kernel inner system error
	KernelInnerSystemErrCode = 3003
	// KernelRequestErrBetweenRuntimeAndBus is the error code of bus communicate with runtime
	KernelRequestErrBetweenRuntimeAndBus = 3001
	// KernelUserCodeLoadErrCode is the error code if kernel user code load error
	KernelUserCodeLoadErrCode = 2001
	// KernelUserFunctionExceptionErrCode is the error code of kernel when user function exception
	KernelUserFunctionExceptionErrCode = 2002
	// KernelCreateLimitErrCode is the error code of kernel when create limited
	KernelCreateLimitErrCode = 1012
	// KernelWriteEtcdCircuitErrCode is the error code of kernel when write etcd failed or circuit
	KernelWriteEtcdCircuitErrCode = 3005
	// KernelDataSystemUnavailable is the error code of kernel when data system is unavailable
	KernelDataSystemUnavailable = 3015
	// KernelNPUFAULTErrCode is the error code of kernel when user exit with npu card is fault
	KernelNPUFAULTErrCode = 3016
)

const (
	// InsReqSuccessCode is the return code when instance request succeeds
	InsReqSuccessCode = 6030
	// InsReqSuccessMessage is the return message when instance request succeeds
	InsReqSuccessMessage = "instance request successfully"
	// UnsupportedOperationErrorCode is the return code when operation is not supported
	UnsupportedOperationErrorCode = 6031
	// UnsupportedOperationErrorMessage is the return message when operation is not supported
	UnsupportedOperationErrorMessage = "operation not supported"
	// FuncNotExistErrorCode is the return code when function does not exist
	FuncNotExistErrorCode = 6032
	// FuncNotExistErrorMessage is the return message when function does not exist
	FuncNotExistErrorMessage = "function not exist"
	// InsNotExistErrorCode is the return code when instance does not exist
	InsNotExistErrorCode = 6033
	// InsNotExistErrorMessage is the return message when instance does not exist
	InsNotExistErrorMessage = "instance not exist"
	// InsAcquireFailedErrorCode is the return code when acquire instance fails
	InsAcquireFailedErrorCode = 6034
	// InsAcquireLeaseExistErrorCode - is the return code when acquire repeated lease
	InsAcquireLeaseExistErrorCode = 6035
	// InsAcquireFailedErrorMessage is the return message when acquire instance fails
	InsAcquireFailedErrorMessage = "failed to acquire instance"
	// LeaseExpireOrDeletedErrorCode is the return code when lease expires or be deleted
	LeaseExpireOrDeletedErrorCode = 6036
	// LeaseExpireOrDeletedErrorMessage is the return message when lease expires or be deleted
	LeaseExpireOrDeletedErrorMessage = "lease expires or deleted"
	// AcquireLeaseTrafficLimitErrorCode -
	AcquireLeaseTrafficLimitErrorCode = 6037
	// AcquireLeaseTrafficLimitErrorMessage is reach max limit of acquiring lease concurrently
	AcquireLeaseTrafficLimitErrorMessage = "reach max limit of acquiring lease concurrently"
	// LeaseErrorInstanceIsAbnormalMessage - lease op failed, instance is abnormal
	LeaseErrorInstanceIsAbnormalMessage = "lease op failed, instance is abnormal"
	// AllSchedulerUnavailableErrorMessage - all scheduler unavailable
	AllSchedulerUnavailableErrorMessage = "all scheduler unavailable"
	// InsAcquireTimeOutErrorCode is the return code when acquire instance timout
	InsAcquireTimeOutErrorCode = 6038
	// AcquireLeaseVPCConflictErrorCode The called function instance has a VPC conflict
	AcquireLeaseVPCConflictErrorCode = 6039
	// AcquireVPCPatInfoErrorCode -
	AcquireVPCPatInfoErrorCode = 6040
	// InstancesConfigEtcdPrefix -
	InstancesConfigEtcdPrefix = "/instances"
	// InstancePathPrefix is the etcd path where the instance info will be placed
	InstancePathPrefix = "/sn/instance"
	// SchedulerHashPrefix is the etcd path where the scheduler hash info will be placed
	SchedulerHashPrefix = "/sn/faas-scheduler/instances"
	// SchedulerRolloutPrefix -
	SchedulerRolloutPrefix = "/sn/faas-scheduler/rollout"
	// RolloutConfigPrefix -
	RolloutConfigPrefix = "/sn/faas-scheduler/rolloutConfig"
	// HTTPTriggerPrefix -
	HTTPTriggerPrefix = "/sn/triggers/triggerType/HTTP/business/"
	// FunctionPrefix -
	FunctionPrefix = "/sn/functions"
	// AliasPrefix -
	AliasPrefix = "/sn/aliases"
	// LeasePrefix -
	LeasePrefix = "/sn/lease"
	// FunctionAvailClusterPrefix Used to identify whether the called function vpc conflicts with the cluster network
	FunctionAvailClusterPrefix = "/sn/function/available/clusters/"
	// FrontendInstancePrefix frontend instance information recorded in meta etcd
	FrontendInstancePrefix = "/sn/frontend/instances"
	// TenantQuotaPrefix define the key prefix of etcd for tenant metadata
	TenantQuotaPrefix = "/sn/quota/cluster"

	// ETCDEventKeySeparator is the separator of ETCD event key
	ETCDEventKeySeparator = "/"

	// DefaultMaxRequestBodySize frontend maximum request body size
	DefaultMaxRequestBodySize = 100 * 1024 * 1024

	// DefaultMapSize default map size
	DefaultMapSize = 3
	// DefaultHostAliasesSliceSize default host aliases slice size
	DefaultHostAliasesSliceSize = 4
	// MinCustomResourcesSize is min custom resource size of invoke
	MinCustomResourcesSize = 0

	// SchedulerExclusivityKey is the key for tenant exclusivity scheduler
	SchedulerExclusivityKey = "exclusivity"
	// FaaSSchedulerExclusivityEnvKey -
	FaaSSchedulerExclusivityEnvKey = "FaaSScheduler_Exclusivity"

	// SchedulerRecoverTime -
	SchedulerRecoverTime = 30 * time.Second
	// DefaultServerWriteTimeOut 1300s
	DefaultServerWriteTimeOut = 1300 * time.Second
	// SchedulerKeyTypeFunction -
	SchedulerKeyTypeFunction = "function"
	// SchedulerKeyTypeModule -
	SchedulerKeyTypeModule = "module"
	// SchedulerInstanceKeyType -
	SchedulerInstanceKeyType = "instance"
	// StaticInstanceApplier mark the instance is created by static function
	StaticInstanceApplier = "static_function"
)

const (
	// KeySeparator is the separator in an ETCD key
	KeySeparator = "/"
	// ValidEtcdKeyLenForInstance is the valid len of an instance ETCD key
	ValidEtcdKeyLenForInstance = 14
	// SysFunctionTenantID is the tenantID of a system function
	SysFunctionTenantID = "0"
	// FaasFrontendMark is a part of the function name of a faasfrontend system function
	FaasFrontendMark = "system-faasfrontend"
	// FaasSchedulerMark is a part of the function name of a faasscheduler system function
	FaasSchedulerMark = "system-faasscheduler"
	// FunctionsIndexForInstance is the functions index of an valid instance ETCD key
	FunctionsIndexForInstance = 2
	// TenantIndexForInstance is the tenant index of an valid instance ETCD key
	TenantIndexForInstance = 5
	// TenantIDIndexForInstance is the tenantID index of an valid instance ETCD key
	TenantIDIndexForInstance = 6
	// FunctionIndexForInstance is the functon index of an valid instance ETCD key
	FunctionIndexForInstance = 7
	// FunctionNameIndexForInstance is the functon name index of an valid instance ETCD key
	FunctionNameIndexForInstance = 8
	// InstanceIDIndexForInstance is the instanceID index of an valid instance ETCD key
	InstanceIDIndexForInstance = 13
	// FaasSchedulerName is  function name of a faasscheduler system function
	FaasSchedulerName = "0-system-faasscheduler"
)

// InstanceStatus is stauts of instance_status object
type InstanceStatus int

const (
	// KernelInstanceStatusExited instance is exited
	KernelInstanceStatusExited InstanceStatus = -1
	// KernelInstanceStatusNew instance is not created
	KernelInstanceStatusNew InstanceStatus = 0
	// KernelInstanceStatusScheduling instance is scheduling
	KernelInstanceStatusScheduling InstanceStatus = 1
	// KernelInstanceStatusCreating instance is creating
	KernelInstanceStatusCreating InstanceStatus = 2
	// KernelInstanceStatusRunning instance is running
	KernelInstanceStatusRunning InstanceStatus = 3
	// KernelInstanceStatusFailed instance is failed
	KernelInstanceStatusFailed InstanceStatus = 4
	// KernelInstanceStatusExiting instance is exiting
	KernelInstanceStatusExiting InstanceStatus = 5
	// KernelInstanceStatusFatal instance abnormal exits
	KernelInstanceStatusFatal InstanceStatus = 6
	// KernelInstanceStatusScheduleFailed instance is schedule failed
	KernelInstanceStatusScheduleFailed InstanceStatus = 7
	// KernelInstanceStatusEvicting instance is evicting
	KernelInstanceStatusEvicting InstanceStatus = 9
	// KernelInstanceStatusEvicted instance is evicted
	KernelInstanceStatusEvicted InstanceStatus = 10
	// KernelInstanceStatusSubHealth instance is sub health
	KernelInstanceStatusSubHealth InstanceStatus = 11
)

// InstanceStatusType is  EXIT_TYPE of instance_status object
type InstanceStatusType int

const (
	// KernelInstanceStatusTypeNoneExit -
	KernelInstanceStatusTypeNoneExit InstanceStatusType = 0
	// KernelInstanceStatusTypeReturn -
	KernelInstanceStatusTypeReturn InstanceStatusType = 1
	// KernelInstanceStatusTypeExceptionInfo -
	KernelInstanceStatusTypeExceptionInfo InstanceStatusType = 2
	// KernelInstanceStatusTypeOomInfo -
	KernelInstanceStatusTypeOomInfo InstanceStatusType = 3
	// KernelInstanceStatusTypeStandardInfo -
	KernelInstanceStatusTypeStandardInfo InstanceStatusType = 4
	// KernelInstanceStatusTypeUnknownError -
	KernelInstanceStatusTypeUnknownError InstanceStatusType = 5
	// KernelInstanceStatusTypeUserKillInfo -
	KernelInstanceStatusTypeUserKillInfo InstanceStatusType = 6
)

const (
	// RuntimeTypeCpp -
	RuntimeTypeCpp = "cpp"
	// RuntimeTypeCppBin -
	RuntimeTypeCppBin = "cppbin"
	// RuntimeTypeJava -
	RuntimeTypeJava = "java"
	// RuntimeTypeNodejs -
	RuntimeTypeNodejs = "nodejs"
	// RuntimeTypePython -
	RuntimeTypePython = "python"
	// RuntimeTypeCustom -
	RuntimeTypeCustom = "custom"
	// RuntimeTypeFusion -
	RuntimeTypeFusion = "fusion"
	// RuntimeTypeHTTP -
	RuntimeTypeHTTP = "http"
)

const (
	// ExtendedCallHandler used as kernel metadata extendedMetaData.extended_handler.handler field
	ExtendedCallHandler = "handler"
	// ExtendedInitHandler used as kernel metadata extendedMetaData.extended_handler.initializer field
	ExtendedInitHandler = "initializer"
	// CallHandler -
	CallHandler = "call"
	// InitHandler -
	InitHandler = "init"
	// CheckPointHandler -
	CheckPointHandler = "checkpoint"
	// RecoverHandler -
	RecoverHandler = "recover"
	// ShutdownHandler -
	ShutdownHandler = "shutdown"
	// SignalHandler -
	SignalHandler = "signal"
)

const (
	// PythonCallExecutor -
	PythonCallExecutor = "faas_executor.faasCallHandler"
	// PythonInitExecutor -
	PythonInitExecutor = "faas_executor.faasInitHandler"
	// PythonCheckPointExecutor -
	PythonCheckPointExecutor = "faas_executor.faasCheckPointHandler"
	// PythonRecoverExecutor -
	PythonRecoverExecutor = "faas_executor.faasRecoverHandler"
	// PythonShutDownExecutor -
	PythonShutDownExecutor = "faas_executor.faasShutDownHandler"
	// PythonSignalExecutor -
	PythonSignalExecutor = "faas_executor.faasSignalHandler"
)

const (
	// MaxTraceIDLength is the max length of traceID
	MaxTraceIDLength = 128
)

const (
	// DefaultListenIP -
	DefaultListenIP = "127.0.0.1"
	// BusProxyHTTPPort -
	BusProxyHTTPPort = "22423"
)

const (
	// TraceIDRuntimeCallCtx Key value of the traceID parameter in the context input parameter of CallHandler
	TraceIDRuntimeCallCtx = "traceID"
)

const (
	// DefaultURNVersion is the default version of a URN
	DefaultURNVersion = "latest"
	// DefaultNameSpace is the default namespace
	DefaultNameSpace = "default"
)

const (
	// ClusterNameEnvKey defines env key for cluster name
	ClusterNameEnvKey = "CLUSTER_NAME"
	// PodIPEnvKey define pod ip env key
	PodIPEnvKey = "POD_IP"
	// HostNameEnvKey defines the hostname env key
	HostNameEnvKey = "HOSTNAME"
	// HostIPEnvKey defines the host ip env key
	HostIPEnvKey = "HOST_IP"
	// ResourceCPUName -
	ResourceCPUName = "CPU"
	// ResourceMemoryName -
	ResourceMemoryName = "Memory"
	// ResourceEphemeralStorage -
	ResourceEphemeralStorage = "ephemeral-storage"
	// CustomContainerRuntimeType is the runtime type for http function
	CustomContainerRuntimeType = "custom image"
	// CustomImageExtraTimeout is the timeout to offset non-pool start of custom image
	CustomImageExtraTimeout = 300
	// PosixCustomRuntimeType is the runtime type for posix custom
	PosixCustomRuntimeType = "posix-custom-runtime"
	// CommonExtraTimeout -
	CommonExtraTimeout = 2
	// TrafficRedundantRate  limit redundancy rate for traffic limitation
	TrafficRedundantRate = 1.1
	// SystemExtraTimeout -
	SystemExtraTimeout = 5
	// KernelScheduleTimeout is the timeout set in kernel to avoid instance schedule timeout
	KernelScheduleTimeout = 5
	// ModuleScheduler -
	ModuleScheduler = "ModuleScheduler"
	// AffinityPoolIDKey -
	AffinityPoolIDKey = "AFFINITY_POOL_ID"
	// UnUseAntiOtherLabelsKey -
	UnUseAntiOtherLabelsKey = "unUseAntiOtherLabels"

	// BusinessTypeServe -
	BusinessTypeServe = "serve"
	// URLSeparator is the separator of http url
	URLSeparator = "/"
	// ApplicationIndex -
	ApplicationIndex = 0

	// EnableStream 用于识别黄区测试桩环境frontend是否开启流式场景监听本地数据系统状态
	EnableStream = "ENABLE_STREAM"
)

const (
	// TrueStr -
	TrueStr = "true"
)

const (
	// DefaultFaaSSchedulerHttpPort -
	DefaultFaaSSchedulerHttpPort = "8889"
)

const (
	// HeaderInvokeURN  -
	HeaderInvokeURN = "X-Tag-VersionUrn"
	// HeaderStateKey -
	HeaderStateKey = "X-State-Key"
	// HeaderNodeLabel is node label
	HeaderNodeLabel = "X-Node-Label"
	// HeaderCPUSize is cpu size specified by invoke
	HeaderCPUSize = "X-Instance-Cpu"
	// HeaderMemorySize is cpu memory specified by invoke
	HeaderMemorySize = "X-Instance-Memory"
	// HeaderCustomResource is customResource specified by invoke
	HeaderCustomResource = "X-Instance-CustomResource"
	// HeaderCustomResourceNew is customResource specified by invoke
	HeaderCustomResourceNew = "X-Instance-Custom-Resource"
	// HeaderContentType -
	HeaderContentType = "Content-Type"
	// HeaderContentLength -
	HeaderContentLength = "Content-Length"
	// HeaderBillingDuration -
	HeaderBillingDuration = "X-Billing-Duration"
	// HeaderInnerCode -
	HeaderInnerCode = "X-Inner-Code"
	// HeaderInvokeSummary -
	HeaderInvokeSummary = "X-Invoke-Summary"
	// HeaderLogResult -
	HeaderLogResult = "X-Log-Result"
	// HeaderLogType -
	HeaderLogType = "X-Log-Type"
	// DefaultLogFlag is the default flag for log
	DefaultLogFlag = "None"
	// HeaderAuthTimestamp is the timestamp for authorization
	HeaderAuthTimestamp = "X-Timestamp-Auth"
	// HeaderAuthorization is authorization
	HeaderAuthorization = "Authorization"
	// HeaderConnection -
	HeaderConnection = "Connection"
	// HeaderSignedHeader -
	HeaderSignedHeader = "X-Signed-Header"
	// HeaderInvokeAlias indicates alias of current invocation
	HeaderInvokeAlias = "x-invoke-alias"
	// HeaderRetryFlag -
	HeaderRetryFlag = "X-Retry-Flag"
	// HeaderInstanceID -
	HeaderInstanceID = "X-Instance-Id"
	// HeaderInstanceIP -
	HeaderInstanceIP = "X-Instance-Ip"
	// HeaderWorkerCost -
	HeaderWorkerCost = "X-Worker-Cost"
	// HeaderCallInstance -
	HeaderCallInstance = "X-Call-Instance"
	// HeaderCallNode -
	HeaderCallNode = "X-Call-Node"

	// HeaderEventSourceID -
	HeaderEventSourceID = "X-Event-Source-Id"
	// HeaderCallType is the request type
	HeaderCallType = "X-Call-Type"
	// HeaderForceDeploy is Force Deploy
	HeaderForceDeploy = "X-Force-Deploy"
	// HeaderStreamAPIGEvent -
	HeaderStreamAPIGEvent = "X-Stream-Apig-Event"
	// HeaderRequestStreamName -
	HeaderRequestStreamName = "X-Request-Stream-Name"
	// HeaderResponseStreamName -
	HeaderResponseStreamName = "X-Response-Stream-Name"
	// HeaderFrontendResponseStreamName -
	HeaderFrontendResponseStreamName = "X-Frontend-Response-Stream-Name"
	// HeaderRemoteClientId -
	HeaderRemoteClientId = "X-Remote-Client-Id"
	// HeaderAcceptEventStream Accept:text/event-stream
	HeaderAcceptEventStream = "text/event-stream"
)

const (
	// NewLease for add a lease of client
	NewLease = "NewLease"
	// KeepAlive for keep client alive
	KeepAlive = "KeepAlive"
	// DelLease for del a lease of client
	DelLease = "DelLease"
)
const (
	// MetaFuncKey key used to match functions within ETCD
	MetaFuncKey = "/sn/functions/business/yrk/tenant/%s/function/%s/version/%s"
	// SilentFuncKey key used to match silent functions within ETCD
	SilentFuncKey = "/silent/sn/functions/business/yrk/tenant/%s/function/%s/version/%s"
)

const (
	// RuntimeInstanceName is instance name specified by user
	RuntimeInstanceName = "instanceName"
	// InstanceCreateEvent key of instance create event
	InstanceCreateEvent = "instanceCreateEvent"
	// InstanceRequirementResourcesKey key of FunctionSystemClient.Invoke args[1]
	InstanceRequirementResourcesKey = "resourcesData"
	// InstanceRequirementInsIDKey key of FunctionSystemClient.Invoke args[1]
	InstanceRequirementInsIDKey = "designateInstanceID"
	// InstanceCallerPodName name of Instance Caller.Invoke args[1]
	InstanceCallerPodName = "instanceCallerPodName"
	// InstanceTrafficLimited - name of instance traffic limit key args[1]
	InstanceTrafficLimited = "instanceTrafficLimited"
	// InstanceRequirementPoolLabel - key of poolLabel
	InstanceRequirementPoolLabel = "poolLabel"
	// InstanceSessionConfig is the key of instance session config in instance acquiring
	InstanceSessionConfig = "instanceSessionConfig"
	// InstanceRequirementInvokeLabel - name of instance label args[1]
	InstanceRequirementInvokeLabel = "instanceInvokeLabel"
)

const (
	// HeaderTenantID -
	HeaderTenantID = "X-Tenant-Id"
	// HeaderFunctionName -
	HeaderFunctionName = "X-Function-Name"
	// HeaderDataSystemPayloadInfo -
	HeaderDataSystemPayloadInfo = "X-Data-System-Payload-Info"
	// HeaderClientID -
	HeaderClientID = "X-Client-Id"
	// HeaderTargetServiceID -
	HeaderTargetServiceID = "X-Target-Service-Id"
)

const (
	// PipInstallPrefix -
	PipInstallPrefix = "pip3.9 install"
	// WorkingDirType -
	WorkingDirType = "working_dir"
	// PipCheckSuffix -
	PipCheckSuffix = "pip3.9 check"
)

const (
	// KillSignalVal -
	KillSignalVal = 1
	// StopAppSignalVal used for stop-app
	StopAppSignalVal = 7

	// KillSignalAliasUpdate is signal for alias update
	KillSignalAliasUpdate = 64
	// KillSignalFaaSSchedulerUpdate is signal for faasscheduler update
	KillSignalFaaSSchedulerUpdate = 72
)

const (
	// InstanceNameNote notes instance name
	InstanceNameNote = "INSTANCE_NAME_NOTE"
	// FunctionKeyNote - is used to describe the function
	FunctionKeyNote = "FUNCTION_KEY_NOTE"
	// ResourceSpecNote - is used to describe the resource
	ResourceSpecNote = "RESOURCE_SPEC_NOTE"
	// SchedulerIDNote - is used to decribe the schedulerID
	SchedulerIDNote = "SCHEDULER_ID_NOTE"
	// InstanceTypeNote - is used to decribe the instance type: "scaled", "reserved", "state"
	InstanceTypeNote = "INSTANCE_TYPE_NOTE"
	// InstanceLabelNode -
	InstanceLabelNode = "INSTANCE_LABEL_NOTE"
)

const (
	// BusinessTypeAgent -
	BusinessTypeAgent = "AGENT"
	// EnableAgentCRDRegistry -
	EnableAgentCRDRegistry = "ENABLE_AGENT_CRD_REGISTRY"
	// NodeNameSpaceSelectorKey the function is not in the default namespace, mark the node selector key.
	NodeNameSpaceSelectorKey = "node-namespace"
)
