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

// Package api for libruntime
package api

// LibruntimeAPI defines libruntime's interfaces, supporting function-system data-system and bridge-system operations
type LibruntimeAPI interface {
	CreateInstance(funcMeta FunctionMeta, args []Arg, invokeOpt InvokeOptions) (string, error)
	InvokeByInstanceId(
		funcMeta FunctionMeta, instanceID string, args []Arg, invokeOpt InvokeOptions,
	) (returnObjectID string, err error)
	InvokeByFunctionName(funcMeta FunctionMeta, args []Arg, invokeOpt InvokeOptions) (string, error)
	AcquireInstance(state string, funcMeta FunctionMeta, acquireOpt InvokeOptions) (InstanceAllocation, error)
	ReleaseInstance(allocation InstanceAllocation, stateID string, abnormal bool, option InvokeOptions)
	Kill(instanceID string, signal int, payload []byte) error

	CreateInstanceRaw(createReqRaw []byte) ([]byte, error)
	InvokeByInstanceIdRaw(invokeReqRaw []byte) ([]byte, error)
	KillRaw(killReqRaw []byte) ([]byte, error)

	SaveState(state []byte) (string, error)
	LoadState(checkpointID string) ([]byte, error)

	Exit(code int, message string)
	Finalize()

	KVSet(key string, value []byte, param SetParam) error
	KVSetWithoutKey(value []byte, param SetParam) (string, error)
	KVMSetTx(keys []string, values [][]byte, param MSetParam) error
	KVGet(key string, timeoutms uint) ([]byte, error)
	KVGetMulti(keys []string, timeoutms uint) ([][]byte, error)
	KVDel(key string) error
	KVDelMulti(keys []string) ([]string, error)

	CreateProducer(streamName string, producerConf ProducerConf) (StreamProducer, error)
	Subscribe(streamName string, config SubscriptionConfig) (StreamConsumer, error)
	DeleteStream(streamName string) error
	QueryGlobalProducersNum(streamName string) (uint64, error)
	QueryGlobalConsumersNum(streamName string) (uint64, error)

	SetTraceID(traceID string)
	SetTenantID(tenantID string) error

	Put(objectID string, value []byte, param PutParam, nestedObjectIDs ...string) error
	PutRaw(objectID string, value []byte, param PutParam, nestedObjectIDs ...string) error
	Get(objectIDs []string, timeoutMs int) ([][]byte, error)
	GetRaw(objectIDs []string, timeoutMs int) ([][]byte, error)
	Wait(objectIDs []string, waitNum uint64, timeoutMs int) ([]string, []string, map[string]error)
	GIncreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error)
	GIncreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error)
	GDecreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error)
	GDecreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error)
	GetAsync(objectID string, cb GetAsyncCallback)
	GetEvent(objectID string, cb GetEventCallback)
	DeleteGetEventCallback(objectID string)

	GetFormatLogger() FormatLogger

	CreateClient(config ConnectArguments) (KvClient, error)
	ReleaseGRefs(remoteClientID string) error
	GetCredential() Credential
	UpdateSchdulerInfo(schedulerName string, schedulerId string, option string)
	IsHealth() bool
	IsDsHealth() bool
	GetActiveMasterAddr() string
}

// KvClient -
type KvClient interface {
	KVSet(key string, value []byte, param SetParam) ErrorInfo

	KVSetWithoutKey(value []byte, param SetParam) (string, ErrorInfo)

	KVGet(key string, timeoutMs ...uint32) ([]byte, ErrorInfo)

	KVGetMulti(keys []string, timeoutMs ...uint32) ([][]byte, ErrorInfo)

	KVQuerySize(keys []string) ([]uint64, ErrorInfo)

	KVDel(key string) ErrorInfo

	KVDelMulti(keys []string) ([]string, ErrorInfo)

	GenerateKey() string

	SetTraceID(traceID string)

	DestroyClient()
}
