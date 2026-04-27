/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
#include <atomic>
#include <exception>
#include <future>
#include <memory>
#include <mutex>
#include <string>
#include <unordered_map>
#include <unordered_set>
#include <vector>

#include "src/dto/internal_wait_result.h"
#include "src/libruntime/utils/utils.h"
#include "src/libruntime/objectstore/memory_store.h"

namespace YR {
namespace Libruntime {
class MemoryStore;
class WaitingEntity {
public:
    WaitingEntity(int minReadyNumber);

    // Be called in SetReady
    void Notify(const std::string &readyObjId);

    // Be called in WaitUntilReady. return true: notified; false: timeout.
    bool Wait(int64_t timeoutMs);

    // set exception. if minReadyNum unreachable, NOTIFY.
    void SetError(const std::string &id, const ErrorInfo &err);

    int GetReadyNum();

    int GetExceptionNum();

    std::vector<std::string> GetReadyObjIds();

    std::unordered_map<std::string, ErrorInfo> GetExceptionIds();

private:
    int minReadyNum;
    std::atomic_int readyNum{0};
    std::atomic_int exceptionNum{0};
    std::promise<void> p;
    std::vector<std::string> readyObjIds;
    std::unordered_map<std::string, ErrorInfo> exceptionIds;
    bool failedTimeout = false;
    bool isFinished = false;
    std::mutex mu;  // for readyObjIds, exceptionIds thread safety access.
};

class WaitingObjectManager {
public:
    WaitingObjectManager() = default;
    explicit WaitingObjectManager(std::function<ErrorInfo()> checkSignals)
    {
        checkSignals_ = checkSignals;
    };
    bool CheckReady(const std::string &id);  // check whether id is in unreadyObjectMap

    // set the 'return obj' unready (insert into unreadyObjectMap) when send Invoke
    bool SetUnready(const std::string &id);

    bool SetReady(const std::string &id);  // notify subscribers of the id. remove it from unreadyObjectMap

    void SetError(const std::string &id, const ErrorInfo &err);

    // Wait a group of ids, until minReadyNum of which ready, or timeout, or too many exception objects
    // In avoid of dead lock, Should not call in MemoryStore.
    std::shared_ptr<YR::InternalWaitResult> WaitUntilReady(std::vector<std::string> idList, std::size_t minReadyNum,
                                                           int64_t timeoutMs);

    void SetMemoryStore(std::shared_ptr<MemoryStore> &_memoryStore);

private:
    std::unordered_map<std::string, std::vector<std::shared_ptr<WaitingEntity>>> unreadyObjectMap;
    bool GetWaitResult(std::shared_ptr<InternalWaitResult> internalWaitResult,
                                             std::vector<std::string> &idList, int64_t currentWaitTimeout,
                                             std::size_t minReadyNum, bool lastWait);
    std::mutex unreadyObjectMapMutex;  // protect unreadyObjectMap
    std::weak_ptr<MemoryStore> memoryStoreWeak;
    std::function<ErrorInfo()> checkSignals_;
};
}  // namespace Libruntime
}  // namespace YR