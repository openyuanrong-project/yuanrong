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

#include <chrono>
#include <thread>

#include <msgpack.hpp>

#include "datasystem/kv_client.h"
#include "src/dto/buffer.h"
#include "src/libruntime/metricsadaptor/metrics_adaptor.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/libruntime/utils/constants.h"
#include "src/libruntime/utils/datasystem_utils.h"
#include "src/utility/logger/logger.h"
#include "src/utility/timer_worker.h"

namespace YR {
namespace Libruntime {

const int DS_HEALTHCHECK_INTERVAL = 3000;  // 1s
const int DS_HEALTHCHECK_TIMES = -1;            // unlimited retry
const int DS_HEALTHCHECK_FAILED_LIMIT = 10;

class DataSystemReadOnlyBuffer : public ReadOnlySharedBuffer {
public:
    DataSystemReadOnlyBuffer(std::shared_ptr<datasystem::ReadOnlyBuffer> buf)
        : ReadOnlySharedBuffer(buf->ImmutableData(), buf->GetSize()), buffer(buf)
    {
    }
    DataSystemReadOnlyBuffer(datasystem::ReadOnlyBuffer &&buf)
        : ReadOnlySharedBuffer(buf.ImmutableData(), buf.GetSize())
    {
        buffer = std::make_shared<datasystem::ReadOnlyBuffer>(std::move(buf));
    }

    virtual ErrorInfo MemoryCopy(const void *data, uint64_t length) override
    {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "Memory copy not supported");
    }

    virtual ErrorInfo Seal(const std::unordered_set<std::string> &nestedIds) override
    {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "Seal not supported");
    }

    virtual ErrorInfo WriterLatch() override
    {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "WLatch not supported");
    }

    virtual ErrorInfo WriterUnlatch() override
    {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "UnWLatch not supported");
    }

    virtual ErrorInfo ReaderLatch() override
    {
        datasystem::Status status = buffer->RLatch();
        RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                          status.ToString());
        return ErrorInfo();
    }

    virtual ErrorInfo ReaderUnlatch() override
    {
        datasystem::Status status = buffer->UnRLatch();
        RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                          status.ToString());
        return ErrorInfo();
    }

private:
    std::shared_ptr<datasystem::ReadOnlyBuffer> buffer;
};

class DSCacheStateStore : public StateStore {
public:
    DSCacheStateStore() = default;
    ~DSCacheStateStore() override = default;

    ErrorInfo Init(const std::string &ip, int port, std::int32_t connectTimeout = DS_CONNECT_TIMEOUT) override;

    ErrorInfo Init(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                   const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                   const std::string &dsPublicKey, const datasystem::SensitiveValue &token, const std::string &ak,
                   const datasystem::SensitiveValue &sk, std::int32_t connectTimeout) override;
    ErrorInfo Init(datasystem::ConnectOptions &inputConnOpt) override;
    void InitOnce(void);

    ErrorInfo Init(const DsConnectOptions &options) override;

    ErrorInfo Write(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam) override;

    //  Set Multiple kv pair
    ErrorInfo MSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<Buffer>> &vals,
                     const MSetParam &mSetParam);

    ErrorInfo Del(const std::string &key) override;

    MultipleDelResult Del(const std::vector<std::string> &keys) override;

    MultipleExistResult Exist(const std::vector<std::string> &keys) override;

    SingleReadResult Read(const std::string &key, int timeoutMS) override;

    MultipleReadResult Read(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial) override;

    MultipleReadResult GetWithParam(const std::vector<std::string> &keys, const GetParams &params,
                                    int timeoutMs) override;

    ErrorInfo QuerySize(const std::vector<std::string> &keys, std::vector<uint64_t> &outSizes) override;

    void Shutdown() override;

    ErrorInfo UpdateToken(datasystem::SensitiveValue token) override;

    ErrorInfo UpdateAkSk(std::string ak, datasystem::SensitiveValue sk) override;

    ErrorInfo GenerateKey(std::string &returnKey) override;

    ErrorInfo Write(std::shared_ptr<Buffer> value, SetParam setParam, std::string &returnKey) override;

    ErrorInfo StartHealthCheck() override;

    ErrorInfo DoHealthCheck();

    ErrorInfo HealthCheck() override;

private:
    ErrorInfo DoInitOnce(void);

    size_t ExtractSuccessObjects(std::vector<std::string> &remainkeys, std::vector<GetParam> &remainParams,
                                 std::vector<datasystem::Optional<datasystem::ReadOnlyBuffer>> &remainList,
                                 std::vector<std::shared_ptr<Buffer>> &resultList,
                                 std::unordered_map<std::string, std::list<size_t>> &keyToIndices)
    {
        size_t newSuccessCount = 0;
        std::vector<std::string> newRemainkeys;
        newRemainkeys.reserve(remainkeys.size());
        std::vector<GetParam> newRemainParams;
        for (size_t i = 0; i < remainkeys.size(); i++) {
            if (i < remainList.size() && !!(remainList[i])) {
                auto &indices = keyToIndices[remainkeys[i]];
                if (UNLIKELY(indices.empty())) {
                    YRLOG_ERROR("Indices should not be empty. key: {}", remainkeys[i]);
                    continue;
                }
                std::shared_ptr<Buffer> buf = std::make_shared<DataSystemReadOnlyBuffer>(std::move(*remainList[i]));
                resultList[indices.front()] = buf;
                indices.pop_front();
                newSuccessCount++;
            } else {
                newRemainkeys.emplace_back(std::move(remainkeys[i]));
                if (!remainParams.empty()) {
                    newRemainParams.emplace_back(std::move(remainParams[i]));
                }
            }
        }
        if (newRemainkeys.size() > 0) {
            YRLOG_INFO("Datasystem get partial values; success keys: ({}/{}); retrying [{}, ...]", newSuccessCount,
                       remainkeys.size(), newRemainkeys[0]);
        }
        remainkeys.swap(newRemainkeys);      // remainIds = newRemainIds
        remainParams.swap(newRemainParams);  // remainParams = newRemainParams.
        return newSuccessCount;
    }

    std::vector<datasystem::ReadParam> BuildDsReadParam(const std::vector<std::string> &keys,
                                                        const std::vector<GetParam> &params)
    {
        std::vector<datasystem::ReadParam> dsParams;
        for (size_t i = 0; i < keys.size(); i++) {
            datasystem::ReadParam dsParam = {.key = keys[i], .offset = params[i].offset, .size = params[i].size};
            dsParams.emplace_back(std::move(dsParam));
        }
        return dsParams;
    }

    ErrorInfo GetValueWithTimeout(const std::vector<std::string> &keys, std::vector<std::shared_ptr<Buffer>> &result,
                                  int timeoutMS, const GetParams &params)
    {
        datasystem::Status status;
        ErrorInfo err;
        std::unordered_map<std::string, std::list<size_t>> keyToIndices;
        std::vector<std::string> remainkeys(keys);
        std::vector<GetParam> remainParams(params.getParams);
        size_t successCount = 0;
        for (size_t i = 0; i < keys.size(); i++) {
            keyToIndices[keys[i]].push_back(i);
        }
        int limitedRetryTime = 0;
        auto start = std::chrono::high_resolution_clock::now();
        auto getElapsedTime = [start]() {
            return std::chrono::duration_cast<std::chrono::milliseconds>(std::chrono::high_resolution_clock::now() -
                                                                         start)
                .count();
        };

        while ((getElapsedTime() <= timeoutMS) || (timeoutMS == NO_TIMEOUT) || (timeoutMS == 0)) {
            std::vector<datasystem::Optional<datasystem::ReadOnlyBuffer>> remainList;
            remainList.reserve(remainkeys.size());
            int to =
                (timeoutMS == NO_TIMEOUT) ? (DEFAULT_TIMEOUT_MS) : (timeoutMS - static_cast<int>(getElapsedTime()));
            to = to < 0 ? 0 : to;
            if (remainParams.empty()) {
                status = dsStateClient->Get(remainkeys, remainList, to);
            } else {
                auto dsParams = BuildDsReadParam(remainkeys, remainParams);
                status = dsStateClient->Read(dsParams, remainList);
            }
            // put remainkeys to completedkeys
            if (!IsRetryableStatus(status)) {
                YRLOG_WARN("the StatusCode of KVGet/KVRead returned is not Retryable: {}", status.ToString());
                break;
            }
            if (IsLimitedRetryEnd(status, limitedRetryTime)) {
                YRLOG_WARN("the StatusCode of KVGet/KVRead returned is limited Retry end: {}", status.ToString());
                break;
            }
            size_t newSuccessCount = ExtractSuccessObjects(remainkeys, remainParams, remainList, result, keyToIndices);
            successCount += newSuccessCount;
            if (successCount == keys.size()) {
                return err;  // return StatusCode::K_OK
            }
            if ((timeoutMS != NO_TIMEOUT && getElapsedTime() > timeoutMS) || (timeoutMS == 0)) {
                break;
            }
            YRLOG_INFO("Datasystem retry to get objects: {}. Elapsed: {}s", status.ToString(),
                       getElapsedTime() / S_TO_MS);
            std::this_thread::sleep_for(std::chrono::seconds(GET_RETRY_INTERVAL));
        }
        return GenerateErrorInfo(successCount, status, timeoutMS, remainkeys, keys);
    }

    bool isInit{false};
    std::shared_ptr<datasystem::KVClient> dsStateClient{nullptr};
    std::once_flag initFlag;
    ErrorInfo initErr;
    datasystem::SensitiveValue tokenUpdated;
    datasystem::ConnectOptions connectOpts;
    std::shared_ptr<YR::utility::TimerWorker> timerWorker_;
    std::shared_ptr<YR::utility::Timer> timer_;
    int lostHealthTimes_ = 0;
};

#define STATE_STORE_INIT_ONCE() \
    do {                        \
        InitOnce();             \
        if (!initErr.OK()) {    \
            return initErr;     \
        }                       \
    } while (0)

#define STATE_STORE_INIT_ONCE_RETURN_PAIR(_first_)     \
    do {                                               \
        InitOnce();                                    \
        if (!initErr.OK()) {                           \
            return std::make_pair((_first_), initErr); \
        }                                              \
    } while (0)

#define STATE_STORE_INIT_ONCE_RETURN(_ret_) \
    do {                                    \
        InitOnce();                         \
        if (!initErr.OK()) {                \
            return (_ret_);                 \
        }                                   \
    } while (0)

}  // namespace Libruntime
}  // namespace YR