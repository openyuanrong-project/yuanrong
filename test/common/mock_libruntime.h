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

#pragma once

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/libruntime/libruntime.h"

using namespace testing;
namespace YR {
namespace Libruntime {
class MockLibruntime : public Libruntime {
public:
    MockLibruntime(std::shared_ptr<LibruntimeConfig> config, std::shared_ptr<ClientsManager> clientsMgr,
                   std::shared_ptr<MetricsAdaptor> metricsAdaptor, std::shared_ptr<Security> security,
                   std::shared_ptr<DomainSocketClient> socketClient)
        : Libruntime(config, clientsMgr, metricsAdaptor, security, socketClient)
    {
    }
    MOCK_METHOD3(CreateInstance, std::pair<ErrorInfo, std::string>(const YR::Libruntime::FunctionMeta &functionMeta,
                                                                   std::vector<YR::Libruntime::InvokeArg> &invokeArgs,
                                                                   YR::Libruntime::InvokeOptions &opts));
    MOCK_METHOD5(InvokeByInstanceId,
                 ErrorInfo(const YR::Libruntime::FunctionMeta &funcMeta, const std::string &instanceId,
                           std::vector<YR::Libruntime::InvokeArg> &args, YR::Libruntime::InvokeOptions &opts,
                           std::vector<DataObject> &returnObjs));

    MOCK_METHOD4(InvokeByFunctionName,
                 ErrorInfo(const YR::Libruntime::FunctionMeta &funcMeta, std::vector<YR::Libruntime::InvokeArg> &args,
                           YR::Libruntime::InvokeOptions &opt, std::vector<DataObject> &returnObjs));

    MOCK_METHOD2(CreateInstanceRaw, void(std::shared_ptr<Buffer> reqRaw, RawCallback cb));

    MOCK_METHOD2(InvokeByInstanceIdRaw, void(std::shared_ptr<Buffer> reqRaw, RawCallback cb));

    MOCK_METHOD2(KillRaw, void(std::shared_ptr<Buffer> reqRaw, RawCallback cb));

    MOCK_METHOD3(Wait, std::shared_ptr<YR::InternalWaitResult>(const std::vector<std::string> &objs,
                                                               std::size_t waitNum, int timeoutSec));

    MOCK_METHOD3(Put, std::pair<ErrorInfo, std::string>(std::shared_ptr<DataObject> dataObj,
                                                        const std::unordered_set<std::string> &nestedIds,
                                                        const CreateParam &createParam));

    MOCK_METHOD4(Put, ErrorInfo(const std::string &objId, std::shared_ptr<DataObject> dataObj,
                                const std::unordered_set<std::string> &nestedId, const CreateParam &createParam));

    // allowPartial is true means YR::GET support to return when getting partial object ref success
    // allowPartial is false means YR::GET will throw exception when getting partial object ref success
    MOCK_METHOD3(Get,
                 std::pair<ErrorInfo, std::vector<std::shared_ptr<DataObject>>>(const std::vector<std::string> &ids,
                                                                                int timeoutMs, bool allowPartial));

    MOCK_METHOD2(IncreaseReference, ErrorInfo(const std::vector<std::string> &objIds, bool toDataSystem));

    MOCK_METHOD2(IncreaseReference,
                 std::pair<ErrorInfo, std::vector<std::string>>(const std::vector<std::string> &objIds,
                                                                const std::string &remoteId));

    MOCK_METHOD1(DecreaseReference, void(const std::vector<std::string> &objids));

    MOCK_METHOD2(DecreaseReference,
                 std::pair<ErrorInfo, std::vector<std::string>>(const std::vector<std::string> &objIds,
                                                                const std::string &remoteId));

    MOCK_METHOD5(AllocReturnObject,
                 ErrorInfo(std::shared_ptr<DataObject> &returnObj, size_t metaSize, size_t dataSize,
                           const std::vector<std::string> &nestedObjIds, uint64_t &totalNativeBufferSize));

    MOCK_METHOD5(AllocReturnObject,
                 ErrorInfo(DataObject *returnObj, size_t metaSize, size_t dataSize,
                           const std::vector<std::string> &nestedObjIds, uint64_t &totalNativeBufferSize));

    MOCK_METHOD3(CreateBuffer, ErrorInfo(size_t dataSize, std::shared_ptr<Buffer> &dataBuf,
                                         const std::vector<std::string> &nestedObjIds));

    MOCK_METHOD2(CreateBuffer, std::pair<ErrorInfo, std::string>(size_t dataSize, std::shared_ptr<Buffer> &dataBuf));

    MOCK_METHOD3(GetBuffers,
                 std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>>(const std::vector<std::string> &ids,
                                                                            int timeoutMs, bool allowPartial));

    MOCK_METHOD2(GetDataObjectsWithoutWait,
                 std::pair<RetryInfo, std::vector<std::shared_ptr<DataObject>>>(const std::vector<std::string> &ids,
                                                                                int timeoutMS));

    MOCK_METHOD6(CreateDataObject,
                 ErrorInfo(const std::string &objId, size_t metaSize, size_t dataSize,
                           std::shared_ptr<DataObject> &dataObj, const std::vector<std::string> &nestedObjIds,
                           const CreateParam &createParam));

    MOCK_METHOD5(CreateDataObject, std::pair<ErrorInfo, std::string>(size_t metaSize, size_t dataSize,
                                                                     std::shared_ptr<DataObject> &dataObj,
                                                                     const std::vector<std::string> &nestedObjIds,
                                                                     const CreateParam &createParam));

    MOCK_METHOD3(GetDataObjects,
                 std::pair<ErrorInfo, std::vector<std::shared_ptr<DataObject>>>(const std::vector<std::string> &ids,
                                                                                int timeoutMs, bool allowPartial));

    MOCK_METHOD1(IsObjectExistingInLocal, bool(const std::string &objId));

    MOCK_METHOD3(Cancel, ErrorInfo(const std::vector<std::string> &objids, bool isForce, bool isRecursive));

    MOCK_METHOD0(Exit, void(void));

    MOCK_METHOD2(Exit, void(const int code, const std::string &message));

    MOCK_METHOD2(Kill, ErrorInfo(const std::string &instanceId, int sigNo));

    MOCK_METHOD3(Kill, ErrorInfo(const std::string &instanceId, int sigNo, std::shared_ptr<Buffer> data));

    MOCK_METHOD3(KillAsync,
                 void(const std::string &instanceId, int sigNo, std::function<void(const ErrorInfo &err)> cb));

    MOCK_METHOD1(Finalize, void(bool isDriver));

    MOCK_METHOD3(WaitAsync, void(const std::string &objectId, WaitAsyncCallback callback, void *userData));

    MOCK_METHOD3(GetEvent, void(const std::string &objectId, GetEventCallback callback, void *userData));

    MOCK_METHOD3(Init, ErrorInfo(std::shared_ptr<FSClient> fsClient, DatasystemClients &datasystemClients,
                                 FinalizeCallback cb));

    MOCK_METHOD0(ReceiveRequestLoop, void(void));

    MOCK_METHOD2(GetRealInstanceId, std::string(const std::string &objectId, int timeout));

    MOCK_METHOD2(SaveRealInstanceId,
                 void(const std::string &objectId, const std::string &instanceId));  // to be deprecated

    MOCK_METHOD3(SaveRealInstanceId,
                 void(const std::string &objectId, const std::string &instanceId, const InstanceOptions &opts));

    MOCK_METHOD2(GetGroupInstanceIds, std::string(const std::string &objectId, int timeout));
    MOCK_METHOD3(SaveGroupInstanceIds,
                 void(const std::string &objectId, const std::string &groupInsIds, const InstanceOptions &opts));

    MOCK_METHOD1(ProcessLog, ErrorInfo(std::string functionLog));

    // KV
    MOCK_METHOD3(KVWrite, ErrorInfo(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam));
    MOCK_METHOD3(KVMSetTx, ErrorInfo(const std::vector<std::string> &keys,
                                     const std::vector<std::shared_ptr<Buffer>> &vals, const MSetParam &mSetParam));
    MOCK_METHOD2(KVRead, SingleReadResult(const std::string &key, int timeoutMS));
    MOCK_METHOD3(KVRead, MultipleReadResult(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial));
    MOCK_METHOD3(KVGetWithParam,
                 MultipleReadResult(const std::vector<std::string> &keys, const GetParams &params, int timeoutMs));
    MOCK_METHOD1(KVDel, ErrorInfo(const std::string &key));
    MOCK_METHOD1(KVDel, MultipleDelResult(const std::vector<std::string> &keys));

    MOCK_METHOD1(KVExist, MultipleExistResult(const std::vector<std::string> &keys));

    MOCK_METHOD3(CreateStreamProducer, ErrorInfo(const std::string &streamName, ProducerConf producerConf,
                                                 std::shared_ptr<StreamProducer> &producer));
    MOCK_METHOD4(CreateStreamConsumer, ErrorInfo(const std::string &streamName, const SubscriptionConfig &config,
                                                 std::shared_ptr<StreamConsumer> &consumer, bool autoAck));
    MOCK_METHOD1(DeleteStream, ErrorInfo(const std::string &streamName));
    MOCK_METHOD2(QueryGlobalProducersNum, ErrorInfo(const std::string &streamName, uint64_t &gProducerNum));
    MOCK_METHOD2(QueryGlobalConsumersNum, ErrorInfo(const std::string &streamName, uint64_t &gConsumerNum));
    MOCK_METHOD2(SaveState, ErrorInfo(const std::shared_ptr<Buffer> data, const int &timeout));
    MOCK_METHOD2(LoadState, ErrorInfo(std::shared_ptr<Buffer> &data, const int &timeout));
    MOCK_METHOD0(GetInvokingRequestId, std::string(void));

    MOCK_CONST_METHOD0(GetThreadPoolSize, uint32_t());
    MOCK_CONST_METHOD0(GetLocalThreadPoolSize, uint32_t());
    MOCK_METHOD2(GroupCreate, ErrorInfo(const std::string &groupName, GroupOpts &opts));
    MOCK_METHOD1(GroupWait, ErrorInfo(const std::string &groupName));
    MOCK_METHOD1(GroupTerminate, void(const std::string &groupName));
    MOCK_METHOD2(GetInstances,
                 std::pair<std::vector<std::string>, ErrorInfo>(const std::string &objId, int timeoutSec));
    MOCK_METHOD2(GetInstances, std::pair<std::vector<std::string>, ErrorInfo>(const std::string &objId,
                                                                              const std::string &groupName));
    MOCK_METHOD0(GenerateGroupName, std::string());
    MOCK_METHOD2(CreateStateStore, ErrorInfo(const DsConnectOptions &opts, std::shared_ptr<StateStore> &stateStore));
    MOCK_METHOD1(SetTraceId, ErrorInfo(const std::string &traceId));
    MOCK_METHOD2(GenerateKeyByStateStore, ErrorInfo(std::shared_ptr<StateStore> stateStore, std::string &returnKey));
    MOCK_METHOD4(SetByStateStore, ErrorInfo(std::shared_ptr<StateStore> stateStore, const std::string &key,
                                            std::shared_ptr<ReadOnlyNativeBuffer> value, SetParam setParam));
    MOCK_METHOD4(SetValueByStateStore,
                 ErrorInfo(std::shared_ptr<StateStore> stateStore, std::shared_ptr<ReadOnlyNativeBuffer> value,
                           SetParam setParam, std::string &returnKey));
    MOCK_METHOD3(GetByStateStore,
                 SingleReadResult(std::shared_ptr<StateStore> stateStore, const std::string &key, int timeoutMs));
    MOCK_METHOD4(GetArrayByStateStore,
                 MultipleReadResult(std::shared_ptr<StateStore> stateStore, const std::vector<std::string> &keys,
                                    int timeoutMs, bool allowPartial));
    MOCK_METHOD3(QuerySizeByStateStore,
                 ErrorInfo(std::shared_ptr<StateStore> stateStore, const std::vector<std::string> &keys,
                           std::vector<uint64_t> &outSizes));
    MOCK_METHOD2(DelByStateStore, ErrorInfo(std::shared_ptr<StateStore> stateStore, const std::string &key));
    MOCK_METHOD2(DelArrayByStateStore,
                 MultipleDelResult(std::shared_ptr<StateStore> stateStore, const std::vector<std::string> &keys));
    MOCK_METHOD1(ExecShutdownCallback, ErrorInfo(int gracePeriodSec));
    MOCK_METHOD1(SetUInt64Counter, ErrorInfo(const YR::Libruntime::UInt64CounterData &data));
    MOCK_METHOD1(ResetUInt64Counter, ErrorInfo(const YR::Libruntime::UInt64CounterData &data));
    MOCK_METHOD1(IncreaseUInt64Counter, ErrorInfo(const YR::Libruntime::UInt64CounterData &data));
    MOCK_METHOD1(GetValueUInt64Counter, std::pair<ErrorInfo, uint64_t>(const YR::Libruntime::UInt64CounterData &data));
    MOCK_METHOD1(SetDoubleCounter, ErrorInfo(const YR::Libruntime::DoubleCounterData &data));
    MOCK_METHOD1(ResetDoubleCounter, ErrorInfo(const YR::Libruntime::DoubleCounterData &data));
    MOCK_METHOD1(IncreaseDoubleCounter, ErrorInfo(const YR::Libruntime::DoubleCounterData &data));
    MOCK_METHOD1(GetValueDoubleCounter, std::pair<ErrorInfo, double>(const YR::Libruntime::DoubleCounterData &data));
    MOCK_METHOD1(ReportGauge, ErrorInfo(const YR::Libruntime::GaugeData &gauge));
    MOCK_METHOD3(SetAlarm, ErrorInfo(const std::string &name, const std::string &description,
                                     const YR::Libruntime::AlarmInfo &alarmInfo));
    MOCK_METHOD1(SetTenantId, void(const std::string &tenantId));
    MOCK_METHOD1(SetTenantId, void(const FunctionMeta &functionmeta));
    MOCK_METHOD3(WaitBeforeGet,
                 std::pair<ErrorInfo, int64_t>(const std::vector<std::string> &ids, int timeoutMs, bool allowPartial));
    MOCK_METHOD0(GetServerVersion, std::string());
    MOCK_METHOD2(PeekObjectRefStream, std::pair<ErrorInfo, std::string>(const std::string &generatorId, bool blocking));
    MOCK_METHOD0(GetFunctionGroupRunningInfo, FunctionGroupRunningInfo());

    MOCK_METHOD3(AcquireInstance,
                 std::pair<InstanceAllocation, ErrorInfo>(const std::string &stateId, const FunctionMeta &functionMeta,
                                                          InvokeOptions &opts));
    MOCK_METHOD4(ReleaseInstance,
                 ErrorInfo(const std::string &leaseId, const std::string &stateId, bool abnormal, InvokeOptions &opts));
    MOCK_METHOD0(GetCredential, Credential());

    // heteroclient
    MOCK_METHOD2(DevDelete,
                 ErrorInfo(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds));
    MOCK_METHOD2(DevLocalDelete,
                 ErrorInfo(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds));
    MOCK_METHOD3(DevSubscribe,
                 ErrorInfo(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                           std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec));
    MOCK_METHOD3(DevPublish,
                 ErrorInfo(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                           std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec));
    MOCK_METHOD3(DevMSet, ErrorInfo(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                    std::vector<std::string> &failedKeys));
    MOCK_METHOD4(DevMGet, ErrorInfo(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                    std::vector<std::string> &failedKeys, int32_t timeoutMs));
    MOCK_METHOD3(GetInstance, std::pair<YR::Libruntime::FunctionMeta, YR::Libruntime::ErrorInfo>(
                                  const std::string &name, const std::string &nameSpace, int timeoutSec));
    MOCK_METHOD3(UpdateSchdulerInfo,
                 void(const std::string &schedulerName, const std::string &schedulerId, const std::string &option));
    MOCK_METHOD1(GetInstanceRoute, std::string(const std::string &objectId));
    MOCK_METHOD2(SaveInstanceRoute, void(const std::string &objectId, const std::string &instanceRoute));
    MOCK_METHOD0(GetResources, std::pair<ErrorInfo, std::vector<YR::Libruntime::ResourceUnit>>());
    MOCK_METHOD0(IsHealth, bool());
    MOCK_METHOD0(IsDsHealth, bool());
    MOCK_METHOD1(StreamWrite, ErrorInfo(const std::string &streamMessage));
};
}  // namespace Libruntime
}  // namespace YR
