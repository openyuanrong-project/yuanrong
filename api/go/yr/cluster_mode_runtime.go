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

// Package yr for actor
package yr

import (
	"fmt"

	"yuanrong.org/kernel/runtime/libruntime"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/clibruntime"
)

// ClusterModeRuntime implements the libRuntimeAPI interface
type ClusterModeRuntime struct {
}

// ReleaseInstance -
func (r *ClusterModeRuntime) ReleaseInstance(allocation api.InstanceAllocation, stateID string,
	abnormal bool, option api.InvokeOptions) {
	return
}

// AcquireInstance -
func (r *ClusterModeRuntime) AcquireInstance(state string, funcMeta api.FunctionMeta,
	invokeOpt api.InvokeOptions) (api.InstanceAllocation, error) {
	return api.InstanceAllocation{}, fmt.Errorf("not implement")
}

// Init initializes the ClusterModeRuntime
func (r *ClusterModeRuntime) Init() error {
	return libruntime.Init(GetConfigManager().Config)
}

// InvokeByFunctionName stateless function invoke
func (r *ClusterModeRuntime) InvokeByFunctionName(
	funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions,
) (string, error) {
	return clibruntime.InvokeByFunctionName(funcMeta, args, invokeOpt)
}

// CreateInstance create instance
func (r *ClusterModeRuntime) CreateInstance(
	funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions,
) (string, error) {
	return clibruntime.CreateInstance(funcMeta, args, invokeOpt)
}

// InvokeByInstanceId invoke by instanceId
func (r *ClusterModeRuntime) InvokeByInstanceId(
	funcMeta api.FunctionMeta, instanceID string, args []api.Arg, invokeOpt api.InvokeOptions,
) (string, error) {
	return clibruntime.InvokeByInstanceId(funcMeta, instanceID, args, api.InvokeOptions{})
}

// Kill send kill instance request
func (r *ClusterModeRuntime) Kill(instanceID string, signal int, payload []byte) error {
	return clibruntime.Kill(instanceID, signal, payload)
}

// CreateInstanceRaw not support raw interface
func (r *ClusterModeRuntime) CreateInstanceRaw(createReqRaw []byte) ([]byte, error) {
	return nil, nil
}

// InvokeByInstanceIdRaw not support raw interface
func (r *ClusterModeRuntime) InvokeByInstanceIdRaw(invokeReqRaw []byte) ([]byte, error) {
	return nil, nil
}

// KillRaw not support raw interface
func (r *ClusterModeRuntime) KillRaw(killReqRaw []byte) ([]byte, error) { return nil, nil }

// SaveState no implement
func (r *ClusterModeRuntime) SaveState(state []byte) (string, error) { return "", nil }

// LoadState no implement
func (r *ClusterModeRuntime) LoadState(checkpointID string) ([]byte, error) { return nil, nil }

// Exit no implement
func (r *ClusterModeRuntime) Exit(code int, message string) {}

// Finalize release resources
func (r *ClusterModeRuntime) Finalize() {
	clibruntime.Finalize()
}

// KVSet save binary data to the data system.
func (r *ClusterModeRuntime) KVSet(key string, value []byte, param api.SetParam) error {
	return clibruntime.KVSet(key, value, param)
}

// KVSetWithoutKey no implement
func (r *ClusterModeRuntime) KVSetWithoutKey(value []byte, param api.SetParam) (string, error) {
	return "", nil
}

// KVMSetTx no implement
func (r *ClusterModeRuntime) KVMSetTx(keys []string, values [][]byte, param api.MSetParam) error {
	return clibruntime.KVMSetTx(keys, values, param)
}

// KVGet get binary data from data system.
func (r *ClusterModeRuntime) KVGet(key string, timeoutms uint) ([]byte, error) {
	return clibruntime.KVGet(key, timeoutms)
}

// KVGetMulti get multi binary data from data system.
func (r *ClusterModeRuntime) KVGetMulti(keys []string, timeoutms uint) ([][]byte, error) {
	return clibruntime.KVGetMulti(keys, timeoutms)
}

// KVDel del data from data system.
func (r *ClusterModeRuntime) KVDel(key string) error {
	return clibruntime.KVDel(key)
}

// KVDelMulti del multi data from data system.
func (r *ClusterModeRuntime) KVDelMulti(keys []string) ([]string, error) {
	return clibruntime.KVDelMulti(keys)
}

// CreateProducer creates and returns api.StreamProducer
func (r *ClusterModeRuntime) CreateProducer(
	streamName string, producerConf api.ProducerConf,
) (api.StreamProducer, error) {
	return clibruntime.CreateStreamProducer(streamName, producerConf)
}

// Subscribe creates and returns api.StreamConsumer
func (r *ClusterModeRuntime) Subscribe(streamName string, config api.SubscriptionConfig) (api.StreamConsumer, error) {
	return clibruntime.CreateStreamConsumer(streamName, config)
}

// DeleteStream Delete a data flow.
// When the number of global producers and consumers is 0,
// the data flow is no longer used and the metadata related to the data flow is deleted from each worker and the
// master. This function can be invoked on any host node.
func (r *ClusterModeRuntime) DeleteStream(streamName string) error {
	return clibruntime.DeleteStream(streamName)
}

// QueryGlobalProducersNum Specifies the flow name to query the number of all producers of the flow.
func (r *ClusterModeRuntime) QueryGlobalProducersNum(streamName string) (uint64, error) {
	return clibruntime.QueryGlobalProducersNum(streamName)
}

// QueryGlobalConsumersNum Specifies the flow name to query the number of all consumers of the flow.
func (r *ClusterModeRuntime) QueryGlobalConsumersNum(streamName string) (uint64, error) {
	return clibruntime.QueryGlobalConsumersNum(streamName)
}

// SetTraceID no implement
func (r *ClusterModeRuntime) SetTraceID(traceID string) {}

// SetTenantID -
func (r *ClusterModeRuntime) SetTenantID(tenantID string) error {
	return clibruntime.SetTenantID(tenantID)
}

// Put put obj data to data system.
func (r *ClusterModeRuntime) Put(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	return clibruntime.Put(objectID, value, param, nestedObjectIDs...)
}

// PutRaw -
func (r *ClusterModeRuntime) PutRaw(
	objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string,
) error {
	return clibruntime.PutRaw(objectID, value, param, nestedObjectIDs...)
}

// Get to get objs from data system.
func (r *ClusterModeRuntime) Get(objectIDs []string, timeoutMs int) ([][]byte, error) {
	return clibruntime.Get(objectIDs, timeoutMs)
}

// GetRaw -
func (r *ClusterModeRuntime) GetRaw(objectIDs []string, timeoutMs int) ([][]byte, error) {
	return clibruntime.GetRaw(objectIDs, timeoutMs)
}

// Wait until result return or timeout
func (r *ClusterModeRuntime) Wait(
	objectIDs []string, waitNum uint64, timeoutMs int,
) ([]string, []string, map[string]error) {
	return clibruntime.Wait(objectIDs, waitNum, timeoutMs)
}

// GIncreaseRef increase object reference count
func (r *ClusterModeRuntime) GIncreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return clibruntime.GIncreaseRef(objectIDs, remoteClientID...)
}

// GIncreaseRefRaw -
func (r *ClusterModeRuntime) GIncreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return clibruntime.GIncreaseRefRaw(objectIDs, remoteClientID...)
}

// GDecreaseRef no implement
func (r *ClusterModeRuntime) GDecreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return nil, nil
}

// GDecreaseRefRaw -
func (r *ClusterModeRuntime) GDecreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return clibruntime.GDecreaseRefRaw(objectIDs, remoteClientID...)
}

// GetAsync no implement
func (r *ClusterModeRuntime) GetAsync(objectID string, cb api.GetAsyncCallback) {}

// GetEvent no implement
func (r *ClusterModeRuntime) GetEvent(objectID string, cb api.GetEventCallback) {}

// DeleteGetEventCallback no implement
func (r *ClusterModeRuntime) DeleteGetEventCallback(objectID string) {}

// GetFormatLogger no implement
func (r *ClusterModeRuntime) GetFormatLogger() api.FormatLogger { return nil }

// CreateClient -
func (r *ClusterModeRuntime) CreateClient(config api.ConnectArguments) (api.KvClient, error) {
	return nil, nil
}

// GetCredential -
func (r *ClusterModeRuntime) GetCredential() api.Credential {
	return api.Credential{}
}

// UpdateSchdulerInfo -
func (r *ClusterModeRuntime) UpdateSchdulerInfo(schedulerName string, schedulerId string, option string) {
}

// IsHealth -
func (r *ClusterModeRuntime) IsHealth() bool {
	return true
}

// IsDsHealth -
func (r *ClusterModeRuntime) IsDsHealth() bool {
	return true
}
