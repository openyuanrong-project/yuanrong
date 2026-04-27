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

// this is Package libruntimesdkimpl implements
package libruntimesdkimpl

import (
	"fmt"
	"os"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/clibruntime"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/log"
	"yuanrong.org/kernel/runtime/libruntime/common/utils"
)

type libruntimeSDKImpl struct{}

// NewLibruntimeSDKImpl creates and returns a new libruntimeSDK instance
func NewLibruntimeSDKImpl() api.LibruntimeAPI {
	return &libruntimeSDKImpl{}
}

func (l *libruntimeSDKImpl) CreateInstance(
	funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	return clibruntime.CreateInstance(funcMeta, args, invokeOpt)
}

func (l *libruntimeSDKImpl) InvokeByInstanceId(
	funcMeta api.FunctionMeta, instanceID string, args []api.Arg, invokeOpt api.InvokeOptions,
) (string, error) {
	return clibruntime.InvokeByInstanceId(funcMeta, instanceID, args, invokeOpt)
}

func (l *libruntimeSDKImpl) InvokeByFunctionName(
	funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions,
) (string, error) {
	return clibruntime.InvokeByFunctionName(funcMeta, args, invokeOpt)
}

func (l libruntimeSDKImpl) AcquireInstance(state string, funcMeta api.FunctionMeta,
	acquireOpt api.InvokeOptions) (api.InstanceAllocation, error) {
	return clibruntime.AcquireInstance(state, funcMeta, acquireOpt)
}

func (l libruntimeSDKImpl) ReleaseInstance(allocation api.InstanceAllocation, stateID string,
	abnormal bool, option api.InvokeOptions) {
	clibruntime.ReleaseInstance(allocation, stateID, abnormal, option)
}

func (l *libruntimeSDKImpl) Kill(instanceID string, signal int, payload []byte) error {
	return clibruntime.Kill(instanceID, signal, payload)
}

func (l *libruntimeSDKImpl) CreateInstanceRaw(createReqRaw []byte) ([]byte, error) {
	return clibruntime.CreateInstanceRaw(createReqRaw)
}

func (l *libruntimeSDKImpl) InvokeByInstanceIdRaw(invokeReqRaw []byte) ([]byte, error) {
	return clibruntime.InvokeByInstanceIdRaw(invokeReqRaw)
}

func (l *libruntimeSDKImpl) KillRaw(killReqRaw []byte) ([]byte, error) {
	return clibruntime.KillRaw(killReqRaw)
}

func (l *libruntimeSDKImpl) SaveState(state []byte) (string, error) {
	return "", fmt.Errorf("not support")
}

func (l *libruntimeSDKImpl) LoadState(checkpointID string) ([]byte, error) {
	return []byte{}, fmt.Errorf("not support")
}

func (l *libruntimeSDKImpl) Exit(code int, message string) {
	clibruntime.Exit(code, message)
}

func (l *libruntimeSDKImpl) Finalize() {
	clibruntime.Finalize()
}

func (l *libruntimeSDKImpl) KVSet(key string, value []byte, param api.SetParam) error {
	return clibruntime.KVSet(key, value, param)
}

func (l *libruntimeSDKImpl) KVSetWithoutKey(value []byte, param api.SetParam) (string, error) {
	return "", fmt.Errorf("not support")
}

func (l *libruntimeSDKImpl) KVMSetTx(keys []string, values [][]byte, param api.MSetParam) error {
	return clibruntime.KVMSetTx(keys, values, param)
}

func (l *libruntimeSDKImpl) KVGet(key string, timeoutms uint) ([]byte, error) {
	return clibruntime.KVGet(key, timeoutms)
}

func (l *libruntimeSDKImpl) KVGetMulti(keys []string, timeoutms uint) ([][]byte, error) {
	return clibruntime.KVGetMulti(keys, timeoutms)
}

func (l *libruntimeSDKImpl) KVDel(key string) error {
	return clibruntime.KVDel(key)
}

func (l *libruntimeSDKImpl) KVDelMulti(keys []string) ([]string, error) {
	return clibruntime.KVDelMulti(keys)
}

func (l *libruntimeSDKImpl) CreateProducer(
	streamName string, producerConf api.ProducerConf) (api.StreamProducer, error) {
	return clibruntime.CreateStreamProducer(streamName, producerConf)
}

func (l *libruntimeSDKImpl) Subscribe(
	streamName string, config api.SubscriptionConfig) (api.StreamConsumer, error) {
	return clibruntime.CreateStreamConsumer(streamName, config)
}

func (l *libruntimeSDKImpl) DeleteStream(streamName string) error {
	return clibruntime.DeleteStream(streamName)
}

func (l *libruntimeSDKImpl) QueryGlobalProducersNum(streamName string) (uint64, error) {
	return clibruntime.QueryGlobalProducersNum(streamName)
}

func (l *libruntimeSDKImpl) QueryGlobalConsumersNum(streamName string) (uint64, error) {
	return clibruntime.QueryGlobalConsumersNum(streamName)
}

func (l *libruntimeSDKImpl) SetTraceID(traceID string) {
	return
}

func (l *libruntimeSDKImpl) SetTenantID(tenantID string) error {
	return clibruntime.SetTenantID(tenantID)
}

func (l *libruntimeSDKImpl) Put(
	objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	return clibruntime.Put(objectID, value, param, nestedObjectIDs...)
}

func (l *libruntimeSDKImpl) PutRaw(
	objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string,
) error {
	return clibruntime.PutRaw(objectID, value, param, nestedObjectIDs...)
}

func (l *libruntimeSDKImpl) Get(objectIDs []string, timeoutMs int) ([][]byte, error) {
	return clibruntime.Get(objectIDs, timeoutMs)
}

func (l *libruntimeSDKImpl) GetRaw(objectIDs []string, timeoutMs int) ([][]byte, error) {
	return clibruntime.GetRaw(objectIDs, timeoutMs)
}

func (l *libruntimeSDKImpl) Wait(
	objectIDs []string, waitNum uint64, timeoutMs int,
) ([]string, []string, map[string]error) {
	return clibruntime.Wait(objectIDs, waitNum, timeoutMs)
}

func (l *libruntimeSDKImpl) GIncreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return clibruntime.GIncreaseRef(objectIDs, remoteClientID...)
}

func (l *libruntimeSDKImpl) GIncreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return clibruntime.GIncreaseRefRaw(objectIDs, remoteClientID...)
}

func (l *libruntimeSDKImpl) GDecreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return clibruntime.GDecreaseRef(objectIDs, remoteClientID...)
}

func (l *libruntimeSDKImpl) GDecreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return clibruntime.GDecreaseRefRaw(objectIDs, remoteClientID...)
}

// ReleaseGRefs release object refs by remote client id
func (l *libruntimeSDKImpl) ReleaseGRefs(remoteClientID string) error {
	err := clibruntime.ReleaseGRefs(remoteClientID)
	return err
}

func (l *libruntimeSDKImpl) GetAsync(objectID string, cb api.GetAsyncCallback) {
	clibruntime.GetAsync(objectID, cb)
}

// GetEvent -
func (l *libruntimeSDKImpl) GetEvent(objectID string, cb api.GetEventCallback) {
	clibruntime.GetEvent(objectID, cb)
}

// DeleteGetEventCallback -
func (l *libruntimeSDKImpl) DeleteGetEventCallback(objectID string) {
	clibruntime.DeleteGetEventCallback(objectID)
}

// UpdateSchdulerInfo -
func (l *libruntimeSDKImpl) UpdateSchdulerInfo(schedulerName string, schedulerId string, option string) {
	clibruntime.UpdateSchdulerInfo(schedulerName, schedulerId, option)
}

func getLogName() string {
	logName := "runtime-go"
	funcLibPath := os.Getenv("FUNCTION_LIB_PATH")
	funcName := utils.GetFuncNameFromFuncLibPath(funcLibPath)
	if len(funcName) == 0 {
		return logName
	}
	logName = funcName
	if logName == "go1.x" {
		logName = "faas-executor"
	}
	return logName
}

func (l *libruntimeSDKImpl) GetFormatLogger() api.FormatLogger {
	logName := getLogName()
	log.GetLogger().Infof("logName is: %s", logName)
	logImpl, err := log.InitRunLog(logName, true)
	if err != nil {
		panic("InitRunLog failed")
	}
	return logImpl
}

// CreateClient -
func (l *libruntimeSDKImpl) CreateClient(config api.ConnectArguments) (api.KvClient, error) {
	return clibruntime.CreateClient(config)
}

// GetCredential
func (l *libruntimeSDKImpl) GetCredential() api.Credential {
	return clibruntime.GetCredential()
}

// IsHealth -
func (l *libruntimeSDKImpl) IsHealth() bool {
	return clibruntime.IsHealth()
}

// IsDsHealth -
func (l *libruntimeSDKImpl) IsDsHealth() bool {
	return clibruntime.IsDsHealth()
}

// GetActiveMasterAddr for getting active master address
func (l *libruntimeSDKImpl) GetActiveMasterAddr() string {
	return clibruntime.GetActiveMasterAddr()
}