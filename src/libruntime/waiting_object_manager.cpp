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
#include "waiting_object_manager.h"

#include <algorithm>
#include <chrono>

#include "src/libruntime/utils/exception.h"
#include "src/utility/logger/logger.h"
namespace YR {
namespace Libruntime {

const int64_t BATCH_WAIT_TIMEOUT_MS = 1000;
const int64_t WAIT_TIMEOUT_MS = 990;
const int64_t WAIT_INTERNAL_TIMEOUT_MS = 10;
// WaitingEntity

WaitingEntity::WaitingEntity(int minReadyNumber)
{
    minReadyNum = minReadyNumber;
}

bool WaitingEntity::Wait(int64_t timeoutMs)
{
    std::future<void> fut = p.get_future();
    if (timeoutMs < 0) {
        fut.wait();
        return true;
    }

    std::future_status status = fut.wait_for(std::chrono::milliseconds(timeoutMs));
    if (status == std::future_status::ready) {
        return true;
    } else if (status == std::future_status::timeout) {
        failedTimeout = true;
    }
    return false;
}

void WaitingEntity::Notify(const std::string &readyObjId)
{
    std::lock_guard<std::mutex> lock(mu);
    readyNum++;
    readyObjIds.push_back(readyObjId);
    if (readyNum + exceptionNum >= minReadyNum && !isFinished) {
        isFinished = true;
        p.set_value();  // notify thread in this->Wait().
    }
}

void WaitingEntity::SetError(const std::string &id, const ErrorInfo &err)
{
    YRLOG_DEBUG("set id {}, msg: {}", id, err.Msg());
    std::lock_guard<std::mutex> lock(mu);
    exceptionIds[id] = err;
    exceptionNum++;
    if (readyNum + exceptionNum >= minReadyNum && !isFinished) {
        isFinished = true;
        p.set_value();  // notify thread in this->Wait().
    }
}

int WaitingEntity::GetReadyNum()
{
    return minReadyNum;
}

int WaitingEntity::GetExceptionNum()
{
    return exceptionNum;
}

std::vector<std::string> WaitingEntity::GetReadyObjIds()
{
    std::lock_guard<std::mutex> lock(mu);
    return readyObjIds;
}

std::unordered_map<std::string, ErrorInfo> WaitingEntity::GetExceptionIds()
{
    std::lock_guard<std::mutex> lock(mu);
    return exceptionIds;
}

// WaitingObjectManager

bool WaitingObjectManager::CheckReady(const std::string &id)
{
    std::lock_guard<std::mutex> lock(unreadyObjectMapMutex);
    return !(bool)(unreadyObjectMap.count(id));
}

bool WaitingObjectManager::SetUnready(const std::string &id)
{
    YRLOG_DEBUG("set id {}", id);
    std::lock_guard<std::mutex> lock(unreadyObjectMapMutex);
    if (unreadyObjectMap.count(id) > 0) {
        return false;
    }
    unreadyObjectMap[id] = std::vector<std::shared_ptr<WaitingEntity>>();
    return true;
}

bool WaitingObjectManager::SetReady(const std::string &id)
{
    std::lock_guard<std::mutex> lock(unreadyObjectMapMutex);
    if (unreadyObjectMap.count(id) == 0) {
        return false;
    }
    for (std::shared_ptr<WaitingEntity> &waitingEntity : unreadyObjectMap[id]) {
        waitingEntity->Notify(id);
    }
    unreadyObjectMap.erase(id);
    return true;
}

void WaitingObjectManager::SetError(const std::string &id, const ErrorInfo &err)
{
    YRLOG_DEBUG("set id {}, msg: {}", id, err.Msg());
    std::lock_guard<std::mutex> lock(unreadyObjectMapMutex);
    if (unreadyObjectMap.count(id) > 0) {
        for (std::shared_ptr<WaitingEntity> &waitingEntity : unreadyObjectMap[id]) {
            waitingEntity->SetError(id, err);
        }
    }
    unreadyObjectMap.erase(id);
}

bool WaitingObjectManager::GetWaitResult(std::shared_ptr<InternalWaitResult> internalWaitResult,
                                         std::vector<std::string> &idList, int64_t currentWaitTimeout,
                                         std::size_t minReadyNum, bool lastWait)
{
    std::vector<std::string> unreadyIds;
    std::vector<std::string> readyIds;
    std::shared_ptr<WaitingEntity> waitingEntity;
    {
        std::lock_guard<std::mutex> lock(unreadyObjectMapMutex);
        std::shared_ptr<MemoryStore> memoryStore = memoryStoreWeak.lock();
        if (memoryStore == nullptr) {
            auto errMsg = "the memstore is null pointer";
            YRLOG_ERROR(errMsg);
            auto err = ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, errMsg);
            for (auto id : idList) {
                internalWaitResult->exceptionIds.emplace(id, err);
            }
            return true;
        }
        for (std::string &id : idList) {
            if (unreadyObjectMap.count(id) == 0) {
                auto errInfo = memoryStore->GetLastError(id);  // fetch history error
                if (errInfo.Code() != ErrorCode::ERR_OK) {
                    internalWaitResult->exceptionIds.emplace(id, errInfo);  // has history error
                } else {
                    readyIds.push_back(id);  // no history error. ready.
                }
                continue;
            }
            // This should not be determined by memoryStore to check if it's ready, needs to be refactored later;
            if (!memoryStore->IsReady(id)) {
                unreadyIds.push_back(id);
                continue;
            }
            auto errInfo = memoryStore->GetLastError(id);
            if (errInfo.Code() != ErrorCode::ERR_OK) {
                internalWaitResult->exceptionIds.emplace(id, errInfo);
            } else {
                readyIds.push_back(id);
            }
            unreadyObjectMap.erase(id);
        }
        auto exceptionIdsNum = internalWaitResult->exceptionIds.size();
        if (readyIds.size() + exceptionIdsNum >= minReadyNum) {
            internalWaitResult->readyIds = readyIds;
            internalWaitResult->unreadyIds = unreadyIds;
            return true;
        }
        waitingEntity = std::make_shared<WaitingEntity>(minReadyNum - readyIds.size() - exceptionIdsNum);
        for (std::string &id : unreadyIds) {
            unreadyObjectMap[id].push_back(waitingEntity);
        }
    }
    waitingEntity->Wait(std::min(currentWaitTimeout, WAIT_TIMEOUT_MS));
    auto exceptionIds = waitingEntity->GetExceptionIds();
    internalWaitResult->exceptionIds.insert(exceptionIds.begin(), exceptionIds.end());
    auto waitingEntityReadyObjIds = waitingEntity->GetReadyObjIds();
    auto needReturn =
        waitingEntityReadyObjIds.size() + readyIds.size() + internalWaitResult->exceptionIds.size() >= minReadyNum;
    if (needReturn || lastWait) {
        readyIds.insert(readyIds.end(), waitingEntityReadyObjIds.begin(), waitingEntityReadyObjIds.end());
        internalWaitResult->readyIds = readyIds;
        std::unordered_set<std::string> readyIdSet(internalWaitResult->readyIds.begin(),
                                                   internalWaitResult->readyIds.end());
        std::copy_if(idList.begin(), idList.end(), std::back_inserter(internalWaitResult->unreadyIds),
                     [&](const std::string &s) {
                         return readyIdSet.count(s) == 0 && internalWaitResult->exceptionIds.count(s) == 0;
                     });
        return needReturn;
    }
    return false;
}

std::shared_ptr<YR::InternalWaitResult> WaitingObjectManager::WaitUntilReady(std::vector<std::string> idList,
                                                                             std::size_t minReadyNum, int64_t timeoutMs)
{
    int64_t remainTimeout = timeoutMs;
    bool shouldBreak = false;
    std::shared_ptr<InternalWaitResult> internalWaitResult;
    while (!shouldBreak) {
        internalWaitResult = std::make_shared<InternalWaitResult>();
        auto currentWaitTimeout =
            timeoutMs == -1 ? BATCH_WAIT_TIMEOUT_MS : std::min(remainTimeout, BATCH_WAIT_TIMEOUT_MS);
        if (remainTimeout >= 0) {
            remainTimeout = remainTimeout - currentWaitTimeout;
        }
        auto needReturn = GetWaitResult(internalWaitResult, idList, currentWaitTimeout, minReadyNum, false);
        if (needReturn) {
            return internalWaitResult;
        }
        if (checkSignals_) {
            auto errInfo = checkSignals_();
            if (!errInfo.OK()) {
                for (auto id : idList) {
                    internalWaitResult->exceptionIds.emplace(id, errInfo);
                }
                return internalWaitResult;
            }
        }
        if (!shouldBreak) {
            std::this_thread::sleep_for(std::chrono::milliseconds(WAIT_INTERNAL_TIMEOUT_MS));
        }
        shouldBreak = (remainTimeout <= 0 && timeoutMs != -1);
        if (shouldBreak) {
            GetWaitResult(internalWaitResult, idList, 0, minReadyNum, shouldBreak);
        }
    }
    return internalWaitResult;
}

void WaitingObjectManager::SetMemoryStore(std::shared_ptr<MemoryStore> &_memoryStore)
{
    this->memoryStoreWeak = _memoryStore;
}

}  // namespace Libruntime
}  // namespace YR