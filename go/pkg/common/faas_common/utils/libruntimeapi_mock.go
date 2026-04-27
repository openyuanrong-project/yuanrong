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

// Package utils is sdk
package utils

import (
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/uuid"
)

// FakeLibruntimeSdkClient -
type FakeLibruntimeSdkClient struct{}

// CreateInstance -
func (f *FakeLibruntimeSdkClient) CreateInstance(funcMeta api.FunctionMeta, args []api.Arg,
	invokeOpt api.InvokeOptions) (string, error) {

	InstanceID := uuid.New().String()
	return InstanceID, nil
}

// InvokeByInstanceId -
func (f *FakeLibruntimeSdkClient) InvokeByInstanceId(funcMeta api.FunctionMeta,
	instanceID string, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	return "", nil
}

// InvokeByFunctionName -
func (f *FakeLibruntimeSdkClient) InvokeByFunctionName(funcMeta api.FunctionMeta,
	args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	return "", nil
}

// AcquireInstance -
func (f *FakeLibruntimeSdkClient) AcquireInstance(state string, funcMeta api.FunctionMeta,
	acquireOpt api.InvokeOptions) (api.InstanceAllocation, error) {
	return api.InstanceAllocation{}, nil
}

// ReleaseInstance -
func (f *FakeLibruntimeSdkClient) ReleaseInstance(allocation api.InstanceAllocation,
	stateID string, abnormal bool, option api.InvokeOptions) {
	return
}

// Kill -
func (f *FakeLibruntimeSdkClient) Kill(instanceID string, signal int, payload []byte) error {
	return nil
}

// CreateInstanceRaw -
func (f *FakeLibruntimeSdkClient) CreateInstanceRaw(createReqRaw []byte) ([]byte, error) {
	return nil, nil
}

// InvokeByInstanceIdRaw -
func (f *FakeLibruntimeSdkClient) InvokeByInstanceIdRaw(invokeReqRaw []byte) ([]byte, error) {
	return nil, nil
}

// KillRaw -
func (f *FakeLibruntimeSdkClient) KillRaw(killReqRaw []byte) ([]byte, error) {
	return nil, nil
}

// SaveState -
func (f *FakeLibruntimeSdkClient) SaveState(state []byte) (string, error) {
	return "", nil
}

// LoadState -
func (f *FakeLibruntimeSdkClient) LoadState(checkpointID string) ([]byte, error) {
	return nil, nil
}

// Exit -
func (f *FakeLibruntimeSdkClient) Exit(code int, message string) {
	return
}

// Finalize -
func (f *FakeLibruntimeSdkClient) Finalize() {
	return
}

// KVSet -
func (f *FakeLibruntimeSdkClient) KVSet(key string, value []byte, param api.SetParam) error {
	return nil
}

// KVSetWithoutKey -
func (f *FakeLibruntimeSdkClient) KVSetWithoutKey(value []byte, param api.SetParam) (string, error) {
	return "", nil
}

// KVMSetTx -
func (f *FakeLibruntimeSdkClient) KVMSetTx(keys []string, values [][]byte, param api.MSetParam) error {
	return nil
}

// KVGet -
func (f *FakeLibruntimeSdkClient) KVGet(key string, timeoutms uint) ([]byte, error) {
	return nil, nil
}

// KVGetMulti -
func (f *FakeLibruntimeSdkClient) KVGetMulti(keys []string, timeoutms uint) ([][]byte, error) {
	return nil, nil
}

// KVDel -
func (f *FakeLibruntimeSdkClient) KVDel(key string) error {
	return nil
}

// KVDelMulti -
func (f *FakeLibruntimeSdkClient) KVDelMulti(keys []string) ([]string, error) {
	return []string{}, nil
}

// CreateProducer -
func (f *FakeLibruntimeSdkClient) CreateProducer(streamName string,
	producerConf api.ProducerConf) (api.StreamProducer, error) {
	return &FakeStreamProducer{}, nil
}

// Subscribe -
func (f *FakeLibruntimeSdkClient) Subscribe(streamName string,
	config api.SubscriptionConfig) (api.StreamConsumer, error) {
	return &FakeStreamConsumer{}, nil
}

// DeleteStream -
func (f *FakeLibruntimeSdkClient) DeleteStream(streamName string) error {
	return nil
}

// QueryGlobalProducersNum -
func (f *FakeLibruntimeSdkClient) QueryGlobalProducersNum(streamName string) (uint64, error) {
	return 0, nil
}

// QueryGlobalConsumersNum -
func (f *FakeLibruntimeSdkClient) QueryGlobalConsumersNum(streamName string) (uint64, error) {
	return 0, nil
}

// SetTraceID -
func (f *FakeLibruntimeSdkClient) SetTraceID(traceID string) {
	return
}

// SetTenantID -
func (f *FakeLibruntimeSdkClient) SetTenantID(tenantID string) error {
	return nil
}

// Put -
func (f *FakeLibruntimeSdkClient) Put(objectID string, value []byte,
	param api.PutParam, nestedObjectIDs ...string) error {
	return nil
}

// PutRaw -
func (f *FakeLibruntimeSdkClient) PutRaw(objectID string, value []byte,
	param api.PutParam, nestedObjectIDs ...string) error {
	return nil
}

// Get -
func (f *FakeLibruntimeSdkClient) Get(objectIDs []string, timeoutMs int) ([][]byte, error) {
	return nil, nil
}

// GetRaw -
func (f *FakeLibruntimeSdkClient) GetRaw(objectIDs []string, timeoutMs int) ([][]byte, error) {
	return nil, nil
}

// Wait -
func (f *FakeLibruntimeSdkClient) Wait(objectIDs []string,
	waitNum uint64, timeoutMs int) ([]string, []string, map[string]error) {
	return nil, nil, nil
}

// GIncreaseRef -
func (f *FakeLibruntimeSdkClient) GIncreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return nil, nil
}

// GIncreaseRefRaw -
func (f *FakeLibruntimeSdkClient) GIncreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return nil, nil
}

// GDecreaseRef -
func (f *FakeLibruntimeSdkClient) GDecreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return nil, nil
}

// GDecreaseRefRaw -
func (f *FakeLibruntimeSdkClient) GDecreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return nil, nil
}

// GetAsync -
func (f *FakeLibruntimeSdkClient) GetAsync(objectID string, cb api.GetAsyncCallback) {
	return
}

// GetEvent -
func (f *FakeLibruntimeSdkClient) GetEvent(objectID string, cb api.GetEventCallback) {
	return
}

// DeleteGetEventCallback -
func (f *FakeLibruntimeSdkClient) DeleteGetEventCallback(objectID string) {
	return
}

// GetFormatLogger -
func (f *FakeLibruntimeSdkClient) GetFormatLogger() api.FormatLogger {
	return nil
}

// CreateClient -
func (f *FakeLibruntimeSdkClient) CreateClient(config api.ConnectArguments) (api.KvClient, error) {
	return nil, nil
}

// ReleaseGRefs -
func (f *FakeLibruntimeSdkClient) ReleaseGRefs(remoteClientID string) error {
	return nil
}

// GetCredential -
func (f *FakeLibruntimeSdkClient) GetCredential() api.Credential {
	return api.Credential{}
}

// UpdateSchdulerInfo -
func (f *FakeLibruntimeSdkClient) UpdateSchdulerInfo(schedulerName string, schedulerId string, option string) {
	return
}

// IsHealth  -
func (f *FakeLibruntimeSdkClient) IsHealth() bool {
	return true
}

// IsDsHealth  -
func (f *FakeLibruntimeSdkClient) IsDsHealth() bool {
	return true
}

// GetActiveMasterAddr for getting active master address
func (f *FakeLibruntimeSdkClient) GetActiveMasterAddr() string {
	return "mockMasterAddr"
}

// FakeStreamProducer -
type FakeStreamProducer struct{}

// Send -
func (fsp *FakeStreamProducer) Send(element api.Element) error {
	return nil
}

// SendWithTimeout -
func (fsp *FakeStreamProducer) SendWithTimeout(element api.Element, timeoutMs int64) error {
	return nil
}

// Flush -
func (fsp *FakeStreamProducer) Flush() error {
	return nil
}

// Close -
func (fsp *FakeStreamProducer) Close() error {
	return nil
}

// FakeStreamConsumer -
type FakeStreamConsumer struct{}

// ReceiveExpectNum -
func (fsc *FakeStreamConsumer) ReceiveExpectNum(expectNum uint32, timeoutMs uint32) ([]api.Element, error) {
	return nil, nil
}

// Receive -
func (fsc *FakeStreamConsumer) Receive(timeoutMs uint32) ([]api.Element, error) {
	return nil, nil
}

// Ack -
func (fsc *FakeStreamConsumer) Ack(elementId uint64) error {
	return nil
}

// Close -
func (fsc *FakeStreamConsumer) Close() error {
	return nil
}
