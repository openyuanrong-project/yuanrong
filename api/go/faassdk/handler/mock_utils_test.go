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

// Package event -
package handler

import (
	"github.com/agiledragon/gomonkey/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"yuanrong.org/kernel/runtime/libruntime/api"
)

// PatchSlice -
type PatchSlice []*gomonkey.Patches

// PatchesFunc -
type PatchesFunc func() PatchSlice

// InitPatchSlice -
func InitPatchSlice() PatchSlice {
	return make([]*gomonkey.Patches, 0)
}

type mockLibruntimeClient struct {
}

func (m mockLibruntimeClient) CreateInstance(funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) InvokeByInstanceId(funcMeta api.FunctionMeta, instanceID string, args []api.Arg, invokeOpt api.InvokeOptions) (returnObjectID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) InvokeByFunctionName(funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) AcquireInstance(state string, funcMeta api.FunctionMeta, acquireOpt api.InvokeOptions) (api.InstanceAllocation, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) ReleaseInstance(allocation api.InstanceAllocation, stateID string, abnormal bool, option api.InvokeOptions) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Kill(instanceID string, signal int, payload []byte) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) CreateInstanceRaw(createReqRaw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) InvokeByInstanceIdRaw(invokeReqRaw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KillRaw(killReqRaw []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) SaveState(state []byte) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) LoadState(checkpointID string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Exit(code int, message string) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Finalize() {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVSet(key string, value []byte, param api.SetParam) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVSetWithoutKey(value []byte, param api.SetParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVGet(key string, timeoutms uint) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVGetMulti(keys []string, timeoutms uint) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVDel(key string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) KVDelMulti(keys []string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) CreateProducer(streamName string, producerConf api.ProducerConf) (api.StreamProducer, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Subscribe(streamName string, config api.SubscriptionConfig) (api.StreamConsumer, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) DeleteStream(streamName string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) QueryGlobalProducersNum(streamName string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) QueryGlobalConsumersNum(streamName string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) SetTraceID(traceID string) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) SetTenantID(tenantID string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Put(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) PutRaw(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Get(objectIDs []string, timeoutMs int) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GetRaw(objectIDs []string, timeoutMs int) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) Wait(objectIDs []string, waitNum uint64, timeoutMs int) ([]string, []string, map[string]error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GIncreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GIncreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GDecreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GDecreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GetAsync(objectID string, cb api.GetAsyncCallback) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GetEvent(objectID string, cb api.GetEventCallback) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) DeleteEventCallback(objectID string, cb api.GetEventCallback) {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) GetFormatLogger() api.FormatLogger {
	//TODO implement me
	panic("implement me")
}

func (m mockLibruntimeClient) CreateClient(config api.ConnectArguments) (api.KvClient, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLibruntimeClient) UpdateSchdulerInfo(schedulerName string, option string) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLibruntimeClient) IsHealth() bool {
	return true
}

func (m *mockLibruntimeClient) IsDsHealth() bool {
	return true
}

func (m *mockLibruntimeClient) GetActiveMasterAddr() string {
	return "mockMasterAddr"
}

// Append -
func (p *PatchSlice) Append(patches PatchSlice) {
	if len(patches) > 0 {
		*p = append(*p, patches...)
	}
}

// ResetAll -
func (p PatchSlice) ResetAll() {
	for _, item := range p {
		item.Reset()
	}
}

// FakeLogger -
type FakeLogger struct{}

// With -
func (f *FakeLogger) With(fields ...zapcore.Field) api.FormatLogger {
	return f
}

// Infof -
func (f *FakeLogger) Infof(format string, paras ...interface{}) {}

// Errorf -
func (f *FakeLogger) Errorf(format string, paras ...interface{}) {}

// Warnf -
func (f *FakeLogger) Warnf(format string, paras ...interface{}) {}

// Debugf -
func (f *FakeLogger) Debugf(format string, paras ...interface{}) {}

// Fatalf -
func (f *FakeLogger) Fatalf(format string, paras ...interface{}) {}

// Info -
func (f *FakeLogger) Info(msg string, fields ...zap.Field) {}

// Error -
func (f *FakeLogger) Error(msg string, fields ...zap.Field) {}

// Warn -
func (f *FakeLogger) Warn(msg string, fields ...zap.Field) {}

// Debug -
func (f *FakeLogger) Debug(msg string, fields ...zap.Field) {}

// Fatal -
func (f *FakeLogger) Fatal(msg string, fields ...zap.Field) {}

// Sync -
func (f *FakeLogger) Sync() {}
