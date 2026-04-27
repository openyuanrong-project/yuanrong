/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.
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

#include <future>
#include <memory>
#include <string>
#include <unordered_set>
#include <vector>

#include <msgpack.hpp>

#include "yr/api/buffer.h"
#include "yr/api/err_type.h"
#include "yr/api/future.h"
#include "yr/api/invoke_arg.h"
#include "yr/api/invoke_options.h"
#include "yr/api/mutable_buffer.h"
#include "yr/api/node.h"
#include "yr/api/stream.h"
#include "yr/api/wait_result.h"

namespace YR {
namespace internal {
struct RetryInfo;
}
enum class ExistenceOpt : int {
    NONE = 0,
    NX = 1,
};

/**
 * @enum WriteMode
 * @brief Write Mode
 * @details Sets the reliability of data. When the server configuration supports a secondary cache (e.g., Redis), this
 * setting ensures data reliability.
 */
enum class WriteMode : int {
    /**
     * @brief Do not write to the secondary cache
     */
    NONE_L2_CACHE = 0,
    /**
     * @brief Synchronously write data to the secondary cache to ensure reliability
     */
    WRITE_THROUGH_L2_CACHE = 1,
    /**
     * @brief Asynchronously write data to the secondary cache to ensure reliability
     */
    WRITE_BACK_L2_CACHE = 2,
    /**
     * @brief Do not write to the secondary cache, and the data may be evicted when system resources are insufficient
     */
    NONE_L2_CACHE_EVICT = 3,
};

/**
 * @enum CacheType
 * @brief Type of allocated medium.
 * @details Indicates whether the allocated medium is memory or disk.
 */
enum class CacheType : int {
    /**
     * @brief Memory Medium
     */
    MEMORY = 0,
    /**
     * @brief Disk Medium
     */
    DISK = 1,
};

/**
 * @enum ConsistencyType
 * @brief Data Consistency Configuration
 * @details In a distributed scenario, different levels of consistency semantics can be configured.
 */
enum class ConsistencyType : int {
    /**
     * @brief Asynchronous
     */
    PRAM = 0,
    /**
     * @brief Causal Consistency
     */
    CAUSAL = 1,
};

/**
 * @struct SetParam
 * @brief Configures attributes for an object, such as reliability
 */
struct SetParam {
    /**
     * @brief Write mode
     * @details Sets the reliability of data. When the server configuration supports a secondary cache (e.g., Redis),
     * this setting ensures data reliability. Default value: YR::WriteMode::NONE_L2_CACHE
     */
    WriteMode writeMode = WriteMode::NONE_L2_CACHE;
    /**
     * @brief Time-to-Live (TTL) in seconds
     * @details Specifies the duration for which the data will be retained before being deleted. Default value: 0,
     * meaning the key will persist until explicitly deleted using the Del interface.
     */
    uint32_t ttlSecond = 0;
    /**
     * @brief Existence option
     * @details Determines if the key allows repeated writes. Optional parameters: YR::ExistenceOpt::NONE (allow,
     * default) and YR::ExistenceOpt::NX (do not allow). Default value: YR::ExistenceOpt::NONE.
     */
    ExistenceOpt existence = ExistenceOpt::NONE;
    /**
     * @brief Trace ID
     * @details A custom trace ID used for troubleshooting and performance optimization. Only supported within the cloud
     * environment; settings outside the cloud will not take effect. Maximum length: 36. Valid characters must match the
     * regular expression: ^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$
     */
    std::string traceId;
};

/**
 * @struct SetParamV2
 * @brief Configures attributes for an object, such as reliability
 */
struct SetParamV2 {
    /**
     * @brief Write mode
     * @details Sets the reliability of data. When the server configuration supports a secondary cache (e.g., Redis),
     * this setting ensures data reliability. Default value: YR::WriteMode::NONE_L2_CACHE
     */
    WriteMode writeMode = WriteMode::NONE_L2_CACHE;
    /**
     * @brief Time-to-Live (TTL) in seconds
     * @details Specifies the duration for which the data will be retained before being deleted. Default value: 0,
     * meaning the key will persist until explicitly deleted using the Del interface.
     */
    uint32_t ttlSecond = 0;
    /**
     * @brief Existence option
     * @details Determines if the key allows repeated writes. Optional parameters: YR::ExistenceOpt::NONE (allow,
     * default) and YR::ExistenceOpt::NX (do not allow). Default value: YR::ExistenceOpt::NONE.
     */
    ExistenceOpt existence = ExistenceOpt::NONE;
    /**
     * @brief Trace ID
     * @details A custom trace ID used for troubleshooting and performance optimization. Only supported within the cloud
     * environment; settings outside the cloud will not take effect. Maximum length: 36. Valid characters must match the
     * regular expression: ^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$
     */
    std::string traceId;
    /**
     * @brief Cache type
     * @details Specifies whether the data is allocated to memory or disk. Optional parameters: YR::CacheType::Memory
     * (memory) and YR::CacheType::Disk (disk). Default value: YR::CacheType::Memory.
     */
    CacheType cacheType = CacheType::MEMORY;
    /**
     * @brief Extended parameters
     * @details Configures additional parameters beyond those specified above
     */
    std::unordered_map<std::string, std::string> extendParams;
};

/**
 * @struct MSetParam
 * @brief Configures attributes for multiple objects, such as reliability
 */
struct MSetParam {
    /**
     * @brief Write mode
     * @details Sets the reliability of data. When the server configuration supports a secondary cache (e.g., Redis),
     * this setting ensures data reliability. Default value: YR::WriteMode::NONE_L2_CACHE
     */
    WriteMode writeMode = WriteMode::NONE_L2_CACHE;
    /**
     * @brief Time-to-Live (TTL) in seconds
     * @details Specifies the duration for which the data will be retained before being deleted. Default value: 0,
     * meaning the key will persist until explicitly deleted using the Del interface.
     */
    uint32_t ttlSecond = 0;
    /**
     * @brief Existence option
     * @details Determines if the key allows repeated writes. Optional parameters: YR::ExistenceOpt::NX (allow, default)
     * and YR::ExistenceOpt::NONE (do not allow). Default value: YR::ExistenceOpt::NX.
     */
    ExistenceOpt existence = ExistenceOpt::NX;
    /**
     * @brief Cache type
     * @details Specifies whether the data is allocated to memory or disk. Optional parameters: YR::CacheType::Memory
     * (memory) and YR::CacheType::Disk (disk). Default value: YR::CacheType::Memory.
     */
    CacheType cacheType = CacheType::MEMORY;
    /**
     * @brief Extended parameters
     * @details Configures additional parameters beyond those specified above
     */
    std::unordered_map<std::string, std::string> extendParams;
};

/**
 * @struct DelParam
 * @brief Specifies parameters for a key to be deleted, such as setting a traceId
 */
struct DelParam {
    /**
     * @brief traceId
     * @details Custom traceId used for troubleshooting and performance optimization;
     * only supported within the cloud environment; settings outside the cloud will not take effect.
     * Maximum length is 36. Valid characters must match the regular expression:
     * ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     */
    std::string traceId;
};

/**
 * @struct GetParam
 * @brief Specifies parameters for a query key
 */
struct GetParam {
    /**
     * @brief  The offset specifies the starting position of the data to be retrieved.
     */
    uint64_t offset;
    /**
     * @brief  The size specifies the number of elements or the amount of data to retrieve.
     */
    uint64_t size;
};

/**
 * @struct GetParams
 * @brief Specifies parameters for a set of query keys
 */
struct GetParams {
    /**
     * @brief  A vector containing multiple GetParam objects.
     */
    std::vector<GetParam> getParams;
    /**
     * @brief  A trace ID used for tracking and identifying specific requests.
     */
    std::string traceId;
};

/**
 * @struct CreateParam
 * @brief Configure attributes for the object, such as whether reliability is needed.
 */
struct CreateParam {
    /**
     * @brief write mode
     * @details Set the reliability of the data.
     * When the server configuration supports a secondary cache to ensure reliability,
     * such as the Redis service, this configuration can be used to guarantee data reliability.
     * Defaults to YR::WriteMode::NONE_L2_CACHE .
     */
    WriteMode writeMode = WriteMode::NONE_L2_CACHE;

    /**
     * @brief consistency type
     * @details Data consistency configuration. In a distributed scenario,
     * different levels of consistency semantics can be configured.
     * The optional parameters are:
     * - YR::ConsistencyType::PRAM (Asynchronous)
     * - YR::ConsistencyType::CAUSAL (Causal consistency)<br>
     * <br>
     * Defaults to YR::ConsistencyType::PRAM .
     */
    ConsistencyType consistencyType = ConsistencyType::PRAM;

    /**
     * @brief  cacheType
     * @details Identifies whether the allocated medium is memory or disk.
     * The optional parameters are:
     * - YR::CacheType::MEMORY (Memory medium)
     * - YR::CacheType::DISK (Disk medium)<br>
     * <br>
     * Defaults to YR::CacheType::MEMORY .
     */
    CacheType cacheType = CacheType::MEMORY;
};

struct Blob {
    void *pointer;
    uint64_t size;
};

struct DeviceBlobList {
    std::vector<Blob> blobs;
    int32_t deviceIdx = -1;
};

struct AsyncResult {
    ErrorInfo error;
    std::vector<std::string> failedList;
};

class Runtime {
public:
    virtual void Init() = 0;

    virtual std::string GetServerVersion() = 0;

    // return objid
    virtual std::string Put(std::shared_ptr<msgpack::sbuffer> data,
                            const std::unordered_set<std::string> &nestedObjectIds) = 0;

    virtual std::string Put(std::shared_ptr<Buffer> data, const std::unordered_set<std::string> &nestedObjectIds) = 0;

    virtual void Put(const std::string &objId, std::shared_ptr<msgpack::sbuffer> data,
                     const std::unordered_set<std::string> &nestedId) = 0;

    virtual std::pair<internal::RetryInfo, std::vector<std::shared_ptr<Buffer>>> Get(
        const std::vector<std::string> &ids, int timeoutMS, int &limitedRetryTime) = 0;

    virtual YR::internal::WaitResult Wait(const std::vector<std::string> &objs, std::size_t waitNum, int timeout) = 0;

    virtual int64_t WaitBeforeGet(const std::vector<std::string> &ids, int timeoutMs, bool allowPartial) = 0;

    virtual void KVWrite(const std::string &key, const char *value, SetParam setParam) = 0;
    virtual void KVWrite(const std::string &key, const char *value, size_t len, SetParam setParam) = 0;
    virtual void KVWrite(const std::string &key, std::shared_ptr<msgpack::sbuffer> value, SetParam setParam) = 0;

    virtual void KVMSetTx(const std::vector<std::string> &keys,
                          const std::vector<std::shared_ptr<msgpack::sbuffer>> &vals, ExistenceOpt existence) = 0;

    virtual std::shared_ptr<Buffer> KVRead(const std::string &key, int timeout) = 0;

    virtual std::vector<std::shared_ptr<Buffer>> KVRead(const std::vector<std::string> &keys, int timeout,
                                                        bool allowPartial = false) = 0;

    virtual std::vector<std::shared_ptr<Buffer>> KVGetWithParam(const std::vector<std::string> &keys,
                                                                const YR::GetParams &params, int timeout) = 0;

    virtual void KVDel(const std::string &key, const DelParam &delParam = {}) = 0;

    virtual std::vector<std::string> KVDel(const std::vector<std::string> &keys, const DelParam &delParam = {}) = 0;

    virtual std::vector<bool> KVExist(const std::vector<std::string> &keys) = 0;

    virtual std::shared_ptr<Producer> CreateStreamProducer(const std::string &streamName,
                                                           ProducerConf producerConf) = 0;

    virtual std::shared_ptr<Consumer> CreateStreamConsumer(const std::string &streamName,
                                                           const SubscriptionConfig &config, bool autoAck = false) = 0;

    virtual void DeleteStream(const std::string &streamName) = 0;

    virtual void QueryGlobalProducersNum(const std::string &streamName, uint64_t &gProducerNum) = 0;

    virtual void QueryGlobalConsumersNum(const std::string &streamName, uint64_t &gConsumerNum) = 0;

    virtual void IncreGlobalReference(const std::vector<std::string> &objectIds, bool toDatasystem = true) = 0;

    virtual void DecreGlobalReference(const std::vector<std::string> &objectIds) = 0;

    virtual std::string InvokeByName(const internal::FuncMeta &funcMeta, std::vector<YR::internal::InvokeArg> &args,
                                     const InvokeOptions &opt) = 0;

    virtual std::string CreateInstance(const internal::FuncMeta &funcMeta, std::vector<YR::internal::InvokeArg> &args,
                                       YR::InvokeOptions &opt) = 0;

    virtual std::string InvokeInstance(const internal::FuncMeta &funcMeta, const std::string &instanceId,
                                       std::vector<YR::internal::InvokeArg> &args, const InvokeOptions &opt) = 0;

    virtual std::string GetRealInstanceId(const std::string &objectId) = 0;

    virtual void SaveRealInstanceId(const std::string &objectId, const std::string &instanceId,
                                    const InvokeOptions &opts) = 0;

    virtual void Cancel(const std::vector<std::string> &objs, bool isForce, bool isRecursive) = 0;

    virtual void TerminateInstance(const std::string &instanceId) = 0;

    virtual void Exit(void) = 0;

    virtual bool IsOnCloud() = 0;

    virtual void GroupCreate(const std::string &name, GroupOptions &opts) = 0;

    virtual void GroupTerminate(const std::string &name) = 0;

    virtual void GroupWait(const std::string &name) = 0;

    virtual std::vector<std::string> GetInstances(const std::string &objId, int timeoutSec) = 0;

    virtual std::string GenerateGroupName() = 0;

    virtual void SaveState(const int &timeout) = 0;

    virtual void LoadState(const int &timeout) = 0;

    virtual void DevDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds) = 0;

    virtual void DevLocalDelete(const std::vector<std::string> &objectIds,
                                std::vector<std::string> &failedObjectIds) = 0;

    virtual void DevSubscribe(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                              std::vector<std::shared_ptr<YR::Future>> &futureVec) = 0;

    virtual void DevPublish(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                            std::vector<std::shared_ptr<YR::Future>> &futureVec) = 0;

    virtual void DevMSet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                         std::vector<std::string> &failedKeys) = 0;

    virtual void DevMGet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                         std::vector<std::string> &failedKeys, int32_t timeout) = 0;

    virtual std::string Put(std::shared_ptr<msgpack::sbuffer> data,
                            const std::unordered_set<std::string> &nestedObjectIds, const CreateParam &createParam) = 0;

    virtual std::string Put(std::shared_ptr<Buffer> data, const std::unordered_set<std::string> &nestedObjectIds,
                            const CreateParam &createParam) = 0;

    virtual void KVWrite(const std::string &key, std::shared_ptr<msgpack::sbuffer> value, SetParamV2 setParam) = 0;

    virtual void KVMSetTx(const std::vector<std::string> &keys,
                          const std::vector<std::shared_ptr<msgpack::sbuffer>> &vals, const MSetParam &mSetParam) = 0;

    virtual internal::FuncMeta GetInstance(const std::string &name, const std::string &nameSpace, int timeoutSec) = 0;

    virtual std::string GetGroupInstanceIds(const std::string &objectId) = 0;

    virtual void SaveGroupInstanceIds(const std::string &objectId, const std::string &groupInsIds,
                                      const InvokeOptions &opts) = 0;

    virtual std::string GetInstanceRoute(const std::string &objectId) = 0;

    virtual void SaveInstanceRoute(const std::string &objectId, const std::string &instanceRoute) = 0;

    virtual void TerminateInstanceSync(const std::string &instanceId) = 0;

    virtual std::vector<Node> Nodes() = 0;

    virtual std::shared_ptr<MutableBuffer> CreateMutableBuffer(uint64_t size) = 0;

    virtual std::vector<std::shared_ptr<MutableBuffer>> GetMutableBuffer(const std::vector<std::string> &ids,
                                                                         int timeout) = 0;
    virtual std::shared_future<void> TerminateInstanceAsync(const std::string &instanceId, bool isSync) = 0;
};
}  // namespace YR
