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
#include <cstdlib>
#include <msgpack.hpp>
#include <string>

#include "src/libruntime/heterostore/hetero_store.h"
#include "src/libruntime/objectstore/object_store.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/libruntime/streamstore/stream_store.h"

namespace YR {
namespace Libruntime {
class MockObjectStore : public ObjectStore {
public:
    MOCK_METHOD(ErrorInfo, Init, (const std::string &addr, int port, std::int32_t connectTimeout), (override));
    MOCK_METHOD(ErrorInfo, Init,
                (const std::string &addr, int port, bool enableDsAuth, bool encryptEnable,
                 const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                 const std::string &dsPublicKey, const datasystem::SensitiveValue &token, const std::string &ak,
                 const datasystem::SensitiveValue &sk, std::int32_t connectTimeout),
                (override));
    MOCK_METHOD1(Init, ErrorInfo(datasystem::ConnectOptions &options));
    MOCK_METHOD(ErrorInfo, Put,
                (std::shared_ptr<Buffer> data, const std::string &objID,
                 const std::unordered_set<std::string> &nestedID, const CreateParam &createParam),
                (override));
    MOCK_METHOD(SingleResult, Get, (const std::string &objID, int timeoutMS), (override));
    MOCK_METHOD(MultipleResult, Get, (const std::vector<std::string> &ids, int timeoutMS), (override));
    MOCK_METHOD(ErrorInfo, IncreGlobalReference, (const std::vector<std::string> &objectIds), (override));
    MOCK_METHOD((std::pair<ErrorInfo, std::vector<std::string>>), IncreGlobalReference,
                (const std::vector<std::string> &objectIds, const std::string &remoteId), (override));
    MOCK_METHOD(ErrorInfo, DecreGlobalReference, (const std::vector<std::string> &objectIds), (override));
    MOCK_METHOD((std::pair<ErrorInfo, std::vector<std::string>>), DecreGlobalReference,
                (const std::vector<std::string> &objectIds, const std::string &remoteId), (override));
    MOCK_METHOD(std::vector<int>, QueryGlobalReference, (const std::vector<std::string> &objectIds), (override));
    MOCK_METHOD(ErrorInfo, ReleaseGRefs, (const std::string &remoteId), (override));
    MOCK_METHOD(ErrorInfo, GenerateKey, (std::string & key, const std::string &prefix, bool isPut), (override));
    MOCK_METHOD(ErrorInfo, GetPrefix, (const std::string &key, std::string &prefix), (override));
    MOCK_METHOD(ErrorInfo, CreateBuffer,
                (const std::string &objectId, size_t dataSize, std::shared_ptr<Buffer> &dataBuf,
                 const CreateParam &createParam),
                (override));
    MOCK_METHOD((std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>>), GetBuffers,
                (const std::vector<std::string> &ids, int timeoutMS), (override));
    MOCK_METHOD((std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>>), GetBuffersWithoutRetry,
                (const std::vector<std::string> &ids, int timeoutMS), (override));
    MOCK_METHOD(void, SetTenantId, (const std::string &tenantId), (override));
    MOCK_METHOD(void, Clear, (), (override));
    MOCK_METHOD(void, Shutdown, (), (override));
    MOCK_METHOD(ErrorInfo, UpdateToken, (datasystem::SensitiveValue token), (override));
    MOCK_METHOD(ErrorInfo, UpdateAkSk, (std::string ak, datasystem::SensitiveValue sk), (override));
};

class MockStateStore : public StateStore {
public:
    MockStateStore() : StateStore() {}
    MOCK_METHOD3(Init, ErrorInfo(const std::string &ip, int port, std::int32_t connectTimeout));
    MOCK_METHOD(ErrorInfo, Init,
                (const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                 const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                 const std::string &dsPublicKey, const datasystem::SensitiveValue &token, const std::string &ak,
                 const datasystem::SensitiveValue &sk, std::int32_t connectTimeout),
                (override));
    MOCK_METHOD1(Init, ErrorInfo(datasystem::ConnectOptions &options));
    MOCK_METHOD1(Init, ErrorInfo(const DsConnectOptions &options));
    MOCK_METHOD3(Write, ErrorInfo(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam));
    MOCK_METHOD3(Write, ErrorInfo(std::shared_ptr<Buffer> value, SetParam setParam, std::string &returnKey));
    MOCK_METHOD3(MSetTx, ErrorInfo(const std::vector<std::string> &keys,
                                   const std::vector<std::shared_ptr<Buffer>> &vals, const MSetParam &mSetParam));
    MOCK_METHOD2(Read, SingleReadResult(const std::string &key, int timeoutMS));
    MOCK_METHOD3(Read, MultipleReadResult(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial));
    MOCK_METHOD3(GetWithParam,
                 MultipleReadResult(const std::vector<std::string> &keys, const GetParams &params, int timeout));
    MOCK_METHOD2(QuerySize, ErrorInfo(const std::vector<std::string> &keys, std::vector<uint64_t> &outSizes));
    MOCK_METHOD1(Del, ErrorInfo(const std::string &key));
    MOCK_METHOD1(Del, MultipleDelResult(const std::vector<std::string> &keys));
    MOCK_METHOD1(Exist, MultipleExistResult(const std::vector<std::string> &keys));
    MOCK_METHOD0(Shutdown, void());
    MOCK_METHOD1(UpdateToken, ErrorInfo(datasystem::SensitiveValue token));
    MOCK_METHOD2(UpdateAkSk, ErrorInfo(std::string ak, datasystem::SensitiveValue sk));
    MOCK_METHOD1(GenerateKey, ErrorInfo(std::string &returnKey));
    MOCK_METHOD0(StartHealthCheck, ErrorInfo());
    MOCK_METHOD0(HealthCheck, ErrorInfo());
};

class MockStreamStore : public StreamStore {
public:
    MOCK_METHOD2(Init, ErrorInfo(const std::string &ip, int port));
    MOCK_METHOD10(Init,
                  ErrorInfo(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                            const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                            const std::string &dsPublicKey, const datasystem::SensitiveValue &token,
                            const std::string &ak, const datasystem::SensitiveValue &sk));
    MOCK_METHOD1(Init, ErrorInfo(datasystem::ConnectOptions &options));
    MOCK_METHOD2(Init, ErrorInfo(datasystem::ConnectOptions &options, std::shared_ptr<StateStore> dsStateStore));
    MOCK_METHOD3(CreateStreamProducer, ErrorInfo(const std::string &streamName,
                                                 std::shared_ptr<StreamProducer> &producer, ProducerConf producerConf));
    MOCK_METHOD4(CreateStreamConsumer, ErrorInfo(const std::string &streamName, const SubscriptionConfig &config,
                                                 std::shared_ptr<StreamConsumer> &consumer, bool autoAck));
    MOCK_METHOD1(DeleteStream, ErrorInfo(const std::string &streamName));
    MOCK_METHOD2(QueryGlobalProducersNum, ErrorInfo(const std::string &streamName, uint64_t &gProducerNum));
    MOCK_METHOD2(QueryGlobalConsumersNum, ErrorInfo(const std::string &streamName, uint64_t &gProducerNum));
    MOCK_METHOD0(Shutdown, void());
    MOCK_METHOD1(UpdateToken, ErrorInfo(datasystem::SensitiveValue token));
    MOCK_METHOD2(UpdateAkSk, ErrorInfo(std::string ak, datasystem::SensitiveValue sk));
};

class MockHeretoStore : public HeteroStore {
public:
    MOCK_METHOD1(Init, ErrorInfo(datasystem::ConnectOptions &options));
    MOCK_METHOD0(Shutdown, void());
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
};

class MockStreamProducer : public StreamProducer {
public:
    MOCK_METHOD1(Send, ErrorInfo(const Element &element));
    MOCK_METHOD2(Send, ErrorInfo(const Element &element, int64_t timeoutMs));
    MOCK_METHOD0(Close, ErrorInfo());
};

class MockStreamConsumer : public StreamConsumer {
public:
    MOCK_METHOD3(Receive, ErrorInfo(uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements));
    MOCK_METHOD2(Receive, ErrorInfo(uint32_t timeoutMs, std::vector<Element> &outElements));
    MOCK_METHOD1(Ack, ErrorInfo(uint64_t elementId));
    MOCK_METHOD0(Close, ErrorInfo());
};
}  // namespace Libruntime
}  // namespace YR
