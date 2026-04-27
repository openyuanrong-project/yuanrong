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

// Package types -
package types

import "time"

const (
	// SystemTenantID is the tenantID for system functions
	SystemTenantID = "0"
	// FaasManagerFuncName is the function name of faas manager
	FaasManagerFuncName = "faasmanager"
	// FaasSchedulerFuncName is the function name of faas scheduler
	FaasSchedulerFuncName = "faasscheduler"
	// FaasFrontendFuncName is the function name of faas frontend
	FaasFrontendFuncName = "faasfrontend"
	// FaasControllerFuncName is the function name of faas controller
	FaasControllerFuncName = "faascontroller"
	// UserInitEntryKey is the key for user init entry
	UserInitEntryKey = "initializer"
	// UserCallEntryKey is the key for user call entry
	UserCallEntryKey = "handler"
	// ConcurrentNumKey is the key for concurrency in CreateOption
	ConcurrentNumKey = "ConcurrentNum"
	// InitCallTimeoutKey is the key for init call timeout in CreateOption
	InitCallTimeoutKey = "init_call_timeout"
	// CallTimeoutKey is the key for init call timeout in CreateOption
	CallTimeoutKey = "call_timeout"
	// NetworkConfigKey is the key for NetworkConfig in CreateOption
	NetworkConfigKey = "networkConfig"
	// ProberConfigKey is the key for ProberConfig in CreateOption
	ProberConfigKey = "proberConfig"
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
	// PermanentInstance 不论scheduler怎么扩缩，这个实例都由创建该实例的scheduler纳管
	PermanentInstance = "-permanent"
	// TemporaryInstance 普通实例，可能会在scheduler扩缩容时改变纳管关系
	TemporaryInstance = "-temporary"
	// FunctionSign 函数签名
	FunctionSign = "FUNCTION_SIGNATURE"
	// TenantID -
	TenantID = "tenantId"
	// HTTPRuntimeType is the runtime type for http function
	HTTPRuntimeType = "http"
	// CustomContainerRuntimeType is the runtime type for http function
	CustomContainerRuntimeType = "custom image"
	// HTTPFuncPort is the listening port for http function
	HTTPFuncPort = 8000
	// HTTPCallRoute is the call route for http function
	HTTPCallRoute = "invoke"
	// GracefulShutdownTime is the key for GRACEFUL_SHUTDOWN_TIME in CreateOption
	GracefulShutdownTime = "GRACEFUL_SHUTDOWN_TIME"
	// MaxShutdownTimeout used to be the default request timeout,now is used to graceful shutdown
	MaxShutdownTimeout = 900
	// MinLeaseInterval is the minimum interval for lease
	MinLeaseInterval = 500 * time.Millisecond
	// HeaderInstanceLabel -
	HeaderInstanceLabel = "X-Instance-Label"
	// RegisterTypeSelf -
	RegisterTypeSelf = "registerBySelf"
	// RegisterTypeContend -
	RegisterTypeContend = "registerByContend"
)

const (
	// AscendResourcePrefix is the prefix of ascend resource
	AscendResourcePrefix = "huawei.com/ascend"
	// AscendRankTableFileEnvKey is the env key of ascend ranktable
	AscendRankTableFileEnvKey = "RANK_TABLE_FILE"
	// AscendRankTableFileEnvValue is the env value of ascend ranktable
	AscendRankTableFileEnvValue = "/opt/config/ascend_config/ranktable_file.json"
	// AscendResourceD910B is one type of ascend resource
	AscendResourceD910B = "huawei.com/ascend-1980"
	// AscendResourceD910BInstanceType is type of D910B
	AscendResourceD910BInstanceType = "instanceType"
	// SystemNodeInstanceType is type of node instance type ,such as 280T
	SystemNodeInstanceType = "X_SYSTEM_NODE_INSTANCE_TYPE"
)

const (
	// IncrementTimeout - add 5 second
	IncrementTimeout = 5
	// DefaultCommonQueueTimeout -
	DefaultCommonQueueTimeout = 20
	// DefaultMaxInsQueueTimeout -
	DefaultMaxInsQueueTimeout = 10
)

const (
	// InstanceSchedulePolicyConcurrency is the schedule policy based on concurrency
	InstanceSchedulePolicyConcurrency = "concurrency"
	// InstanceSchedulePolicyRoundRobin is the schedule policy based on round-robin
	InstanceSchedulePolicyRoundRobin = "round-robin"
	// InstanceSchedulePolicyMicroService is the schedule policy based on microservice
	InstanceSchedulePolicyMicroService = "microservice"
)

const (
	// InstanceScalePolicyConcurrency is the auto scaler policy based on concurrency
	InstanceScalePolicyConcurrency = "concurrency"
	// InstanceScalePolicyPredict is the auto scaler policy based on qps predict
	InstanceScalePolicyPredict = "qpsPredict"
	// InstanceScalePolicyStaticFunction is the schedule policy for static function
	InstanceScalePolicyStaticFunction = "staticFunction"
	// InstanceScalePolicyOneshot is the schedule policy for one-shot function
	InstanceScalePolicyOneshot = "oneshot"
)

const (
	// ScenarioWiseCloud is the scenario of CaaS
	ScenarioWiseCloud = "WiseCloud"
	// ScenarioFunctionGraph is the scenario of FunctionGraph
	ScenarioFunctionGraph = "FunctionGraph"
)

const (
	// DefaultMaxOnDemandInstanceNumPerTenant define the default value for maximum on-demand instance per tenant
	DefaultMaxOnDemandInstanceNumPerTenant = 1000
	// DefaultMaxReversedInstanceNumPerTenant define the default value for maximum reversed instance per tenant
	DefaultMaxReversedInstanceNumPerTenant = 1000
)

const (
	// FgsTunnelName -
	FgsTunnelName = "tunl_fgs_vpc"
	// IPIPMode -
	IPIPMode = "ipip"

	// ICMPProtocol -
	ICMPProtocol = "ICMP"
	// HTTPProtocol -
	HTTPProtocol = "HTTP"
)
