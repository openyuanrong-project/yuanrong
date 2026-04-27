/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

// Implement GenericModeRuntime

#include <unordered_set>

#include "config_manager.h"
#include "src/libruntime/libruntime_config.h"
#include "src/libruntime/libruntime_options.h"
#include "yr/api/runtime.h"
#include "yr/api/wait_result.h"
namespace YR {
YR::Libruntime::InvokeOptions BuildOptions(const YR::InvokeOptions &opts);
class ClusterModeRuntime : public Runtime {
public:
    ClusterModeRuntime() = default;

    void Init() override;

    std::string GetServerVersion() override;

    // return objid
    std::string Put(std::shared_ptr<msgpack::sbuffer> data, const std::unordered_set<std::string> &nestedId);

    std::string Put(std::shared_ptr<Buffer> data, const std::unordered_set<std::string> &nestedId);

    std::string Put(std::shared_ptr<msgpack::sbuffer> data, const std::unordered_set<std::string> &nestedId,
                    const CreateParam &createParam);

    std::string Put(std::shared_ptr<Buffer> data, const std::unordered_set<std::string> &nestedId,
                    const CreateParam &createParam);

    void Put(const std::string &objId, std::shared_ptr<msgpack::sbuffer> data,
             const std::unordered_set<std::string> &nestedId);

    // KV
    void KVWrite(const std::string &key, std::shared_ptr<msgpack::sbuffer> value, SetParam setParam);

    void KVWrite(const std::string &key, const char *value, SetParam setParam);

    void KVWrite(const std::string &key, const char *value, size_t len, SetParam setParam);

    void KVWrite(const std::string &key, std::shared_ptr<msgpack::sbuffer> value, SetParamV2 setParam);

    void KVMSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<msgpack::sbuffer>> &vals,
                  ExistenceOpt existence);

    void KVMSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<msgpack::sbuffer>> &vals,
                  const MSetParam &mSetParam);

    std::shared_ptr<Buffer> KVRead(const std::string &key, int timeoutMs);

    std::vector<std::shared_ptr<Buffer>> KVRead(const std::vector<std::string> &keys, int timeoutMs,
                                                bool allowPartial = false);

    void KVDel(const std::string &key, const DelParam &delParam = {});

    std::vector<std::shared_ptr<Buffer>> KVGetWithParam(const std::vector<std::string> &keys,
                                                        const YR::GetParams &params, int timeoutMs);

    std::vector<std::string> KVDel(const std::vector<std::string> &keys, const DelParam &delParam = {});

    std::vector<bool> KVExist(const std::vector<std::string> &keys);

    std::shared_ptr<Producer> CreateStreamProducer(const std::string &streamName, ProducerConf producerConf);

    std::shared_ptr<Consumer> CreateStreamConsumer(const std::string &streamName, const SubscriptionConfig &config,
                                                   bool autoAck = false);

    void DeleteStream(const std::string &streamName);

    void QueryGlobalProducersNum(const std::string &streamName, uint64_t &gProducerNum);

    void QueryGlobalConsumersNum(const std::string &streamName, uint64_t &gConsumerNum);

    std::string InvokeByName(const internal::FuncMeta &funcMeta, std::vector<YR::internal::InvokeArg> &args,
                             const InvokeOptions &opt) override;

    std::string InvokeInstance(const internal::FuncMeta &funcMeta, const std::string &instanceId,
                               std::vector<YR::internal::InvokeArg> &args, const InvokeOptions &opt) override;

    std::string CreateInstance(const internal::FuncMeta &funcMeta, std::vector<YR::internal::InvokeArg> &args,
                               YR::InvokeOptions &opt) override;

    std::string GetRealInstanceId(const std::string &objectId);

    void SaveRealInstanceId(const std::string &objectId, const std::string &instanceId, const InvokeOptions &opts);

    void Cancel(const std::vector<std::string> &objs, bool isForce, bool isRecursive);

    std::pair<internal::RetryInfo, std::vector<std::shared_ptr<Buffer>>> Get(const std::vector<std::string> &ids,
                                                                             int timeoutMS,
                                                                             int &limitedRetryTime) override;

    YR::internal::WaitResult Wait(const std::vector<std::string> &objs, std::size_t waitNum, int timeoutSec) override;

    int64_t WaitBeforeGet(const std::vector<std::string> &ids, int timeoutMs, bool allowPartial) override;

    std::vector<std::string> GetInstances(const std::string &objId, int timeoutSec) override;

    std::string GenerateGroupName() override;

    // throw YR::Libruntime::Exception
    void IncreGlobalReference(const std::vector<std::string> &objids, bool toDatasystem = true);

    void DecreGlobalReference(const std::vector<std::string> &objids);

    void TerminateInstance(const std::string &instanceId);

    void Exit(void);

    static void StopRuntime(void);

    bool IsOnCloud();

    void GroupCreate(const std::string &name, GroupOptions &opts);

    void GroupTerminate(const std::string &name);

    void GroupWait(const std::string &name);

    void SaveState(const int &timeout);

    void LoadState(const int &timeout);

    void DevDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds) override;

    void DevLocalDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds) override;

    void DevSubscribe(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                      std::vector<std::shared_ptr<YR::Future>> &futureVec) override;

    void DevPublish(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                    std::vector<std::shared_ptr<YR::Future>> &futureVec) override;

    void DevMSet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                 std::vector<std::string> &failedKeys) override;

    void DevMGet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                 std::vector<std::string> &failedKeys, int32_t timeout) override;

    internal::FuncMeta GetInstance(const std::string &name, const std::string &nameSpace, int timeoutSec) override;

    std::string GetGroupInstanceIds(const std::string &objectId);

    void SaveGroupInstanceIds(const std::string &objectId, const std::string &groupInsIds, const InvokeOptions &opts);

    std::string GetInstanceRoute(const std::string &objectId);

    void SaveInstanceRoute(const std::string &objectId, const std::string &instanceRoute);

    void TerminateInstanceSync(const std::string &instanceId);

    std::vector<Node> Nodes();

    std::shared_ptr<MutableBuffer> CreateMutableBuffer(uint64_t size) override;

    std::vector<std::shared_ptr<MutableBuffer>> GetMutableBuffer(const std::vector<std::string> &ids,
                                                                 int timeout) override;

    std::shared_future<void> TerminateInstanceAsync(const std::string &instanceId, bool isSync) override;

private:
    std::string Put(const void *data, uint64_t dataSize, const std::unordered_set<std::string> &nestedId,
                    const CreateParam &createParam);
};
}  // namespace YR
