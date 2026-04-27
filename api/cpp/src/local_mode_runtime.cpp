/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

#include "yr/api/local_mode_runtime.h"

#include "config_manager.h"
#include "api/cpp/src/read_only_buffer.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/libruntime/objectstore/object_store.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/libruntime/utils/datasystem_utils.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"
#include "src/utility/thread_pool.h"

namespace YR {
namespace internal {
const static std::string LOCAL_THREAD_NAME = "yr.local";

Libruntime::ErrorInfo ProcessKeyPartialResult(const std::vector<std::string> &keys,
                                              const std::vector<std::shared_ptr<msgpack::sbuffer>> &result,
                                              const Libruntime::ErrorInfo &errInfo, const int timeoutMs)
{
    Libruntime::ErrorInfo err;
    std::vector<std::string> failKeys;
    bool isPartialResult = false;
    for (unsigned int i = 0; i < result.size(); i++) {
        if (!(result[i])) {  // result[i] is nullptr
            isPartialResult = true;
            failKeys.push_back(keys[i]);
        }
    }
    if (isPartialResult) {
        err.SetErrCodeAndMsg(YR::Libruntime::ErrorCode::ERR_GET_OPERATION_FAILED,
                             YR::Libruntime::ModuleCode::DATASYSTEM, errInfo.GetExceptionMsg(failKeys, timeoutMs),
                             errInfo.GetDsStatusCode());
    }
    return err;
}

void LocalModeRuntime::Init()
{
    waitRequestManager_ = std::make_shared<WaitRequestManager>();
    stateStore_ = std::make_shared<LocalStateStore>();
    YRLOG_INFO("Job ID: {}, runtime ID: {}, log dir: {}", ConfigManager::Singleton().jobId,
               ConfigManager::Singleton().runtimeId, ConfigManager::Singleton().logDir);
}

void LocalModeRuntime::Stop()
{
    if (pool_) {
        pool_->Shutdown();
    }
    initPool_ = false;
    if (stateStore_) {
        stateStore_->Clear();
    }
    if (waitRequestManager_) {
        waitRequestManager_.reset();
    }
}

int GetLocalThreadPoolSize()
{
    if (internal::IsLocalMode()) {
        return ConfigManager::Singleton().localThreadPoolSize;
    }
    if (Libruntime::LibruntimeManager::Instance().GetLibRuntime() == nullptr) {
        YRLOG_WARN("libruntime is not initialized; use default local thread pool size: {}",
                   ConfigManager::Singleton().localThreadPoolSize);
        return ConfigManager::Singleton().localThreadPoolSize;
    }
    return YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->GetLocalThreadPoolSize();
}

std::string LocalModeRuntime::GenerateObjId()
{
    return YR::utility::IDGenerator::GenObjectId();
}

void LocalModeRuntime::LocalSubmit(std::function<void()> &&func)
{
    if (!initPool_) {
        initPool_ = true;
        threads_ = GetLocalThreadPoolSize();
        pool_ = std::make_shared<YR::utility::ThreadPool>();
        pool_->Init(threads_, LOCAL_THREAD_NAME);
    }
    if (threads_ == 0) {
        throw YR::Exception::InvalidParamException("cannot submit task to empty thread pool");
    }
    pool_->Handle(std::move(func), "");
}

bool LocalModeRuntime::SetReady(const std::string &objectId)
{
    waitRequestManager_->SetReady(objectId);
    return true;
}

void LocalModeRuntime::SetException(const std::string &id, const std::exception_ptr &exceptionPtr)
{
    waitRequestManager_->SetException(id, exceptionPtr);
}

// KV
void LocalModeRuntime::KVWrite(const std::string &key, std::shared_ptr<msgpack::sbuffer> value, ExistenceOpt existence)
{
    stateStore_->Write(key, value, existence);
}

void LocalModeRuntime::KVMSetTx(const std::vector<std::string> &keys,
                                const std::vector<std::shared_ptr<msgpack::sbuffer>> &vals, ExistenceOpt existence)
{
    stateStore_->MSetTx(keys, vals, existence);
}

std::shared_ptr<Buffer> LocalModeRuntime::KVRead(const std::string &key, int timeoutMs)
{
    auto result = stateStore_->Read(key, timeoutMs);
    return std::make_shared<YR::ReadOnlyBuffer>(std::make_shared<YR::Libruntime::MsgpackBuffer>(result.first));
}

std::vector<std::shared_ptr<Buffer>> LocalModeRuntime::KVRead(const std::vector<std::string> &keys, int timeoutMs,
                                                              bool allowPartial)
{
    auto result = stateStore_->Read(keys, timeoutMs);
    if (!allowPartial) {
        auto partialErrInfo = ProcessKeyPartialResult(keys, result.first, result.second, timeoutMs);
        if (partialErrInfo.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
            throw YR::Exception(static_cast<int>(partialErrInfo.Code()), static_cast<int>(partialErrInfo.MCode()),
                                partialErrInfo.Msg());
        }
    }
    std::vector<std::shared_ptr<Buffer>> buffers;
    buffers.resize(result.first.size());
    for (size_t i = 0; i < result.first.size(); i++) {
        if (result.first[i] == nullptr) {
            continue;
        }
        buffers[i] =
            std::make_shared<YR::ReadOnlyBuffer>(std::make_shared<YR::Libruntime::MsgpackBuffer>(result.first[i]));
    }
    return buffers;
}

void LocalModeRuntime::KVDel(const std::string &key)
{
    stateStore_->Del(key);
}

std::vector<std::string> LocalModeRuntime::KVDel(const std::vector<std::string> &keys)
{
    return stateStore_->Del(keys);
}

std::vector<bool> LocalModeRuntime::KVExist(const std::vector<std::string> &keys)
{
    return stateStore_->Exist(keys);
}
}  // namespace internal
}  // namespace YR