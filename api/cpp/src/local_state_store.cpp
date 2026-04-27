/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 * Description: the KV interface provided by yuanrong
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
#include "yr/api/local_state_store.h"
#include <vector>
#include "re2/re2.h"
#include "parallel_for/complier.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/utility/logger/logger.h"
#include "yr/api/constant.h"
#include "yr/api/exception.h"
namespace YR {
namespace internal {

using YR::Libruntime::ErrorCode;
using YR::Libruntime::ModuleCode;

const int MIN_CHECK_INTERVAL_MS = 200;
const int MAX_CHECK_INTERVAL_MS = 1000;
const int GET_RETRY_MAX_TIME = 5;
const int MAX_MSET_SIZE = 8;
const re2::RE2 KEY_REGEX("^[a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;]*$");

LocalStateStore::LocalStateStore() {}

LocalStateStore::~LocalStateStore() {}

void LocalStateStore::Write(const std::string &key, std::shared_ptr<msgpack::sbuffer> value, ExistenceOpt existence)
{
    std::lock_guard<std::mutex> lock(mtx);
    if (key.empty() || value->size() == 0) {
        throw YR::Exception(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, "the length of key or value is 0");
    }
    if (existence == ExistenceOpt::NX && kv_map.count(key) > 0) {
        throw YR::Exception(ErrorCode::ERR_KEY_ALREADY_EXIST, ModuleCode::DATASYSTEM, "key already exist");
    }
    kv_map[key] = value;
}

bool IsRegexMatch(const re2::RE2 &re, const std::string &key)
{
    if (key.size() <= UINT8_MAX && (RE2::PartialMatch(key, re) || key.empty())) {
        return true;
    }
    return false;
}

void LocalStateStore::MSetTx(const std::vector<std::string> &keys,
                             const std::vector<std::shared_ptr<msgpack::sbuffer>> &vals, ExistenceOpt existence)
{
    if (keys.size() > MAX_MSET_SIZE) {
        std::string msg =
            "Invalid parameter. The maximum size of keys in single operation is " + std::to_string(MAX_MSET_SIZE) + ".";
        throw YR::Exception(ErrorCode::ERR_PARAM_INVALID, ModuleCode::DATASYSTEM, msg);
    }
    if (keys.empty()) {
        std::string msg = "The keys should not be empty.";
        throw YR::Exception(ErrorCode::ERR_PARAM_INVALID, ModuleCode::DATASYSTEM, msg);
    }
    for (auto &key : keys) {
        if (key.empty() || !IsRegexMatch(KEY_REGEX, key)) {
            std::string msg = "Invalid key: " + key;
            throw YR::Exception(ErrorCode::ERR_PARAM_INVALID, ModuleCode::DATASYSTEM, msg);
        }
    }
    std::lock_guard<std::mutex> lock(mtx);
    for (size_t i = 0; i < keys.size(); i++) {
        // ALL keys should not exist
        if (kv_map.find(keys[i]) != kv_map.end()) {
            // should throw error same as that in DSCacheStateStore::MSetTx
            throw YR::Exception(ErrorCode::ERR_KEY_ALREADY_EXIST, ModuleCode::DATASYSTEM,
                                "key " + keys[i] + " already exist");
        }
    }
    for (size_t i = 0; i < keys.size(); i++) {
        kv_map[keys[i]] = vals[i];
    }
}

SingleReadResult LocalStateStore::Read(const std::string &key, int timeoutMS)
{
    std::vector<std::string> keys = {key};
    MultipleReadResult result = Read(keys, timeoutMS);
    return std::make_pair(result.first[0], result.second);
}

MultipleReadResult LocalStateStore::Read(const std::vector<std::string> &keys, int timeoutMS)
{
    bool isExist = false;
    std::vector<std::shared_ptr<msgpack::sbuffer>> bufs(keys.size());
    Libruntime::ErrorInfo err = GetValueWithTimeout(keys, bufs, isExist, timeoutMS);
    if (!isExist && !err.OK()) {
        throw YR::Exception(err.Code(), err.MCode(), err.Msg());
    }
    return std::make_pair(bufs, err);
}

void LocalStateStore::Del(const std::string &key)
{
    std::lock_guard<std::mutex> lock(mtx);
    kv_map.erase(key);
}

std::vector<std::string> LocalStateStore::Del(const std::vector<std::string> &keys)
{
    std::lock_guard<std::mutex> lock(mtx);
    std::vector<std::string> failedKeys;
    for (auto &key : keys) {
        // erase will not throw exception
        kv_map.erase(key);
    }
    return failedKeys;
}

std::vector<bool> LocalStateStore::Exist(const std::vector<std::string> &keys)
{
    std::vector<bool> exists;
    exists.reserve(keys.size());
    std::lock_guard<std::mutex> lock(mtx);
    for (auto &key : keys) {
        if (kv_map.count(key) > 0) {
            exists.push_back(true);
        } else {
            exists.push_back(false);
        }
    }
    return exists;
}

bool LocalStateStore::IsEmpty()
{
    std::lock_guard<std::mutex> lock(mtx);
    return kv_map.empty();
}

int LocalStateStore::GetRetryInterval(int timeoutMS) const
{
    int interval = timeoutMS / GET_RETRY_MAX_TIME;
    if (interval > MAX_CHECK_INTERVAL_MS) {
        return MAX_CHECK_INTERVAL_MS;
    }
    if (interval < MIN_CHECK_INTERVAL_MS) {
        return MIN_CHECK_INTERVAL_MS;
    }
    return interval;
}

Libruntime::ErrorInfo LocalStateStore::GetValueWithTimeout(const std::vector<std::string> &keys,
                                                           std::vector<std::shared_ptr<msgpack::sbuffer>> &res,
                                                           bool &isExist, int timeoutMS)
{
    Libruntime::ErrorInfo err;
    size_t successCount = 0;
    int interval = GetRetryInterval(timeoutMS);
    std::vector<std::string> remainKeys(keys);
    std::unordered_map<std::string, std::list<size_t>> keyToIndices;
    for (size_t i = 0; i < keys.size(); i++) {
        keyToIndices[keys[i]].push_back(i);
    }
    auto start = std::chrono::high_resolution_clock::now();
    auto getElapsedTime = [start]() {
        return std::chrono::duration_cast<std::chrono::milliseconds>(std::chrono::high_resolution_clock::now() - start)
            .count();
    };

    while ((getElapsedTime() <= timeoutMS) || (timeoutMS == NO_TIMEOUT) || (timeoutMS == 0)) {
        std::vector<std::string> failedKeys;
        {
            std::lock_guard<std::mutex> lock(mtx);
            for (const std::string &key : remainKeys) {
                if (kv_map.find(key) == kv_map.end()) {
                    failedKeys.emplace_back(key);
                    continue;
                }
                auto p = kv_map.at(key);
                auto &indices = keyToIndices[key];
                if (UNLIKELY(indices.empty())) {
                    YRLOG_ERROR("Indices should not be empty. key: {}", key);
                    continue;
                }
                res[indices.front()] = p;
                indices.pop_front();
                successCount++;
                isExist = true;
            }
        }
        if (failedKeys.size() == 0) {
            return err;
        }
        if ((timeoutMS != NO_TIMEOUT && getElapsedTime() > timeoutMS) || (timeoutMS == 0)) {
            break;
        }
        remainKeys.swap(failedKeys);
        YRLOG_INFO("Datasystem retry to get objects failed. Elapsed: {} s", getElapsedTime() / S_TO_MS);
        std::this_thread::sleep_for(std::chrono::milliseconds(interval));
    }

    std::stringstream ss;
    ss << "get keys timeout, remain keys count: " << keys.size() - successCount;
    err.SetErrCodeAndMsg(ErrorCode::ERR_GET_OPERATION_FAILED, ModuleCode::DATASYSTEM, ss.str());
    return err;
}

void LocalStateStore::Clear() noexcept
{
    YRLOG_DEBUG("Clear all key-values in state store");
    kv_map.clear();
}
}  // namespace internal
}  // namespace YR