/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

#include "limiter_consistant_hash.h"

namespace YR {
namespace Libruntime {
using namespace std::chrono;
void LimiterCsHash::Add(const std::string &schedulerName, const std::string &schedulerId, int weight)
{
    std::lock_guard<std::mutex> lk(limiterMtx_);
    this->AddWithoutLock(schedulerName, schedulerId, weight);
}

void LimiterCsHash::AddWithoutLock(const std::string &schedulerName, const std::string &schedulerId, int weight)
{
    if (schedulerName.empty() || schedulerId.empty()) {
        YRLOG_WARN("scheduler name: {} or scheduler id {} is empty, no need add and return directly", schedulerName,
                   schedulerId);
        return;
    }
    nodeCount_++;
    loadBalancer_->Add(schedulerName, weight);
    schedulerInfoMap[schedulerName] = schedulerId;
    YRLOG_DEBUG("add scheduler name {}, id: {}, current node count is {}", schedulerName, schedulerId,
                nodeCount_.load());
}

std::string LimiterCsHash::NextRetry(const std::string &functionId, bool move)
{
    std::vector<std::shared_ptr<SchedulerInstance>> vec;
    return NextRetry(functionId, nullptr, move);
}

std::string LimiterCsHash::NextRetry(const std::string &functionId,
                                     std::shared_ptr<AvailableSchedulerInfos> schedulerInfos, bool move)
{
    auto schedulerInstanceId = Next(functionId, schedulerInfos, move);
    if (schedulerInstanceId.empty()) {
        return Next(functionId, schedulerInfos, true);
    }
    return schedulerInstanceId;
}

std::string LimiterCsHash::Next(const std::string &functionId, bool move)
{
    return Next(functionId, nullptr, move);
}

std::string LimiterCsHash::Next(const std::string &functionId, std::shared_ptr<AvailableSchedulerInfos> schedulerInfos,
                                bool moveFlag)
{
    std::lock_guard<std::mutex> lk(limiterMtx_);
    auto res = loadBalancer_->Next(functionId, moveFlag);
    size_t pos = res.find(":");
    if (pos != std::string::npos) {
        std::string schedulerName = res.substr(0, pos);
        long long updateTime = std::stoll(res.substr(pos + 1));
        YRLOG_DEBUG("res is {}, schedulerName is {}, update time is {}", res, schedulerName, updateTime);
        if (schedulerName.empty()) {
            return "";
        }
        if (schedulerInfoMap.find(schedulerName) != schedulerInfoMap.end()) {
            if (moveFlag && isAllInsUnavailable(schedulerInfoMap[schedulerName], updateTime, schedulerInfos)) {
                return ALL_SCHEDULER_UNAVAILABLE;
            }
            return schedulerInfoMap[schedulerName];
        }
    }
    return "";
}

bool LimiterCsHash::isAllInsUnavailable(const std::string &instanceKey, long long updateTime,
                                                         std::shared_ptr<AvailableSchedulerInfos> schedulerInfos)
{
    if (schedulerInfos == nullptr) {
        return false;
    }
    absl::WriterMutexLock insLk(&schedulerInfos->schedulerMtx);
    YRLOG_DEBUG("instanceKey is {}, update time is {}", instanceKey, updateTime);
    bool hasAvailableInsInVec = false;
    bool hasInsIdMatch = false;
    std::shared_ptr<SchedulerInstance> matchedIns;
    for (auto &scheduler : schedulerInfos->schedulerInstanceList) {
        if (scheduler->isAvailable) {
            hasAvailableInsInVec = true;
        }
        if (scheduler->InstanceID == instanceKey) {
            hasInsIdMatch = true;
            matchedIns = scheduler;
        }
    }

    if (hasAvailableInsInVec) {
        if (!hasInsIdMatch) {
            schedulerInfos->schedulerInstanceList.push_back(
                std::make_shared<SchedulerInstance>(SchedulerInstance{"", instanceKey, updateTime, true}));
        } else {
            matchedIns->updateTime = updateTime;
        }
        return false;
    }

    if (!hasInsIdMatch) {
        schedulerInfos->schedulerInstanceList.push_back(
            std::make_shared<SchedulerInstance>(SchedulerInstance{"", instanceKey, updateTime, true}));
        return false;
    } else {
        if (updateTime <= matchedIns->updateTime) {
            YRLOG_WARN(
                "current scheduler vecs has no available ins, next scheduler res is {}, add into hash pool at {}, used "
                "at {}",
                instanceKey, updateTime, matchedIns->updateTime);
            return true;
        }
        matchedIns->isAvailable = true;
        matchedIns->updateTime = updateTime;
    }
    return false;
}

void LimiterCsHash::Remove(const std::string &name)
{
    std::lock_guard<std::mutex> lk(limiterMtx_);
    nodeCount_--;
    YRLOG_DEBUG("start remove schedule id: {}, current node count is {}", name, nodeCount_.load());
    loadBalancer_->Remove(name);
    schedulerInfoMap.erase(name);
}

void LimiterCsHash::RemoveAll()
{
    std::lock_guard<std::mutex> lk(limiterMtx_);
    nodeCount_ = 0;
    YRLOG_DEBUG("start remove all scheduler");
    loadBalancer_->RemoveAll();
    schedulerInfoMap.clear();
}

bool LimiterCsHash::IsSameWithHashPool(const std::vector<SchedulerInstance> &schedulerInfoList)
{
    if (schedulerInfoList.size() != schedulerInfoMap.size()) {
        return false;
    }
    for (auto &schedulerIns : schedulerInfoList) {
        if (schedulerInfoMap.find(schedulerIns.InstanceName) == schedulerInfoMap.end() ||
            schedulerInfoMap[schedulerIns.InstanceName] != schedulerIns.InstanceID) {
            return false;
        }
    }
    return true;
}

void LimiterCsHash::ResetAll(const std::vector<SchedulerInstance> &schedulerInfoList, int weight)
{
    std::lock_guard<std::mutex> lk(limiterMtx_);
    if (auto isSame = IsSameWithHashPool(schedulerInfoList); isSame) {
        return;
    }
    nodeCount_ = 0;
    YRLOG_DEBUG("start remove all scheduler");
    loadBalancer_->RemoveAll();
    schedulerInfoMap.clear();
    for (auto info : schedulerInfoList) {
        this->AddWithoutLock(info.InstanceName, info.InstanceID, weight);
    }
}

}  // namespace Libruntime
}  // namespace YR