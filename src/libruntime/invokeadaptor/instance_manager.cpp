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

#include "instance_manager.h"

namespace YR {
namespace Libruntime {
using json = nlohmann::json;
ErrorInfo ConvertStringToInsResp(InstanceResponse &resp, const std::string &bufStr)
{
    /*
    bufstr pattern
    {"funcKey":"","funcSig":"","instanceID":"","threadID":"","leaseInterval":0,
    "errorCode":150428,"errorMessage":"","schedulerTime":30}
    */
    ErrorInfo err;
    try {
        json jsonData = json::parse(bufStr);
        InstanceAllocation info;
        info.functionId = jsonData["funcKey"].get<std::string>();
        info.funcSig = jsonData["funcSig"].get<std::string>();
        info.instanceId = jsonData["instanceID"].get<std::string>();
        info.leaseId = jsonData["threadID"].get<std::string>();
        info.tLeaseInterval = jsonData["leaseInterval"].get<int>();
        info.forceInvoke = jsonData.value("forceInvoke", false);
        resp.info = info;
        resp.errorCode = jsonData["errorCode"].get<int>();
        resp.errorMessage = jsonData["errorMessage"].get<std::string>();
        resp.schedulerTime = jsonData["schedulerTime"].get<float>();
    } catch (const std::exception &e) {
        err.SetErrCodeAndMsg(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME,
                             "failed parse acquire notify request");
    }
    return err;
}

void ConvertStringToBatchInsResp(BatchInstanceResponse &resp, const std::string &bufStr)
{
    try {
        json jsonData = json::parse(bufStr);
        resp.tLeaseInterval = jsonData["leaseInterval"].get<int>();
        resp.schedulerTime = jsonData["schedulerTime"].get<float>();
        auto allocSucceed = jsonData["instanceAllocSucceed"].get<std::unordered_map<std::string, json>>();
        auto allocFailed = jsonData["instanceAllocFailed"].get<std::unordered_map<std::string, json>>();
        for (const auto &item : allocSucceed) {
            InstanceAllocation info;
            info.functionId = item.second["funcKey"].get<std::string>();
            info.funcSig = item.second["funcSig"].get<std::string>();
            info.instanceId = item.second["instanceID"].get<std::string>();
            info.leaseId = item.second["threadID"].get<std::string>();
            resp.instanceAllocSucceed[item.first] = info;
        }
        for (const auto &item : allocFailed) {
            InstanceAllocationFailedRsp info;
            info.errorCode = item.second["errorCode"].get<int>();
            info.errorMessage = item.second["errorMessage"].get<std::string>();
            resp.instanceAllocFailed[item.first] = info;
        }
    } catch (const std::exception &e) {
        YRLOG_WARN("failed to convert renew or release notify req to instance info, err msg is {}", e.what());
    }
}

size_t GetDelayTime(size_t failedCnt)
{
    static const std::vector<size_t> RETRY_TIME_SEQ = {0, 1, 2, 4, 8, 16, 32};
    if (failedCnt < RETRY_TIME_SEQ.size()) {
        return RETRY_TIME_SEQ[failedCnt];
    }
    return RETRY_TIME_SEQ[RETRY_TIME_SEQ.size() - 1];
}

RequestResource GetRequestResource(std::shared_ptr<InvokeSpec> spec)
{
    size_t concurrency = DEFAULT_CONCURRENCY;
    auto iter = spec->opts.customExtensions.find(CONCURRENCY);
    if (iter != spec->opts.customExtensions.end()) {
        concurrency = static_cast<unsigned short>(std::stoull(iter->second));
    }
    RequestResource resource;
    resource.functionMeta = spec->functionMeta;
    resource.concurrency = concurrency;
    resource.opts = spec->opts;
    return resource;
}

void CancelScaleDownTimer(std::shared_ptr<InstanceInfo> insInfo)
{
    absl::WriterMutexLock insLock(&insInfo->mtx);
    if (insInfo->scaleDownTimer != nullptr) {
        insInfo->scaleDownTimer->cancel();
        insInfo->scaleDownTimer.reset();
    }
}

std::shared_ptr<RequestResourceInfo> InsManager::GetOrAddRequestResourceInfo(const RequestResource &resource)
{
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        absl::WriterMutexLock lock(&insMtx);
        if (requestResourceInfoMap.find(resource) == requestResourceInfoMap.end()) {
            requestResourceInfoMap[resource] = std::make_shared<RequestResourceInfo>();
        }
        info = requestResourceInfoMap[resource];
    }
    return info;
}

std::shared_ptr<RequestResourceInfo> InsManager::GetRequestResourceInfo(const RequestResource &resource)
{
    std::shared_ptr<RequestResourceInfo> info;
    {
        absl::ReaderMutexLock lock(&insMtx);
        if (requestResourceInfoMap.find(resource) != requestResourceInfoMap.end()) {
            info = requestResourceInfoMap[resource];
        }
    }
    return info;
}

std::shared_ptr<InstanceInfo> InsManager::GetInstanceInfo(const RequestResource &resource, const std::string &insId)
{
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return nullptr;
    }
    std::shared_ptr<InstanceInfo> insInfo;
    {
        absl::ReaderMutexLock lock(&info->mtx);
        if (info->instanceInfos.find(insId) == info->instanceInfos.end()) {
            return nullptr;
        }
        insInfo = info->instanceInfos[insId];
    }
    return insInfo;
}

void InsManager::AddRequestResourceInfo(std::shared_ptr<InvokeSpec> spec)
{
    auto resource = GetRequestResource(spec);
    GetOrAddRequestResourceInfo(resource);
}

InstanceSummary InsManager::ScheduleInsWithDevice(const RequestResource &resource,
                                                  std::shared_ptr<RequestResourceInfo> resourceInfo)
{
    absl::ReaderMutexLock infoLock(&resourceInfo->mtx);
    int64_t minCost = resource.opts.maxInvokeLatency;
    std::string minCostInstance = "";
    for (auto &instanceInfo : resourceInfo->instanceInfos) {
        if (instanceInfo.first.empty()) {
            continue;
        }
        int64_t nextCost = 0;
        {
            absl::WriterMutexLock lock(&invokeCostMtx);
            if (invokeCostMap.find(instanceInfo.first) == invokeCostMap.end()) {
                invokeCostMap[instanceInfo.first] = TimeMeasurement(DEFAULT_INVOKE_DURATION);
            }
            nextCost = invokeCostMap[instanceInfo.first].GetEstimatedCostofNextRequest();
        }
        if (nextCost <= resource.opts.maxInvokeLatency && nextCost <= minCost) {
            minCost = nextCost;
            minCostInstance = instanceInfo.first;
            YRLOG_DEBUG("instance: {} estimate cost of next request is: {} ms", instanceInfo.first, nextCost);
        }
    }
    if (!minCostInstance.empty()) {
        auto insInfo = resourceInfo->instanceInfos[minCostInstance];
        absl::WriterMutexLock insInfoLock(&insInfo->mtx);
        insInfo->unfinishReqNum++;
    }
    return InstanceSummary{minCostInstance, ""};
}

// return std::pair<insId, leaseId>
InstanceSummary InsManager::GetAvailableIns(const RequestResource &resource)
{
    if (!runFlag) {
        return InstanceSummary{"", ""};
    }
    auto resourceInfo = GetRequestResourceInfo(resource);
    if (resourceInfo == nullptr) {
        return InstanceSummary{"", ""};
    }
    if (!resource.opts.device.name.empty()) {
        return ScheduleInsWithDevice(resource, resourceInfo);
    }
    absl::WriterMutexLock infoLock(&resourceInfo->mtx);
    if (resourceInfo->avaliableInstanceInfos.begin() == resourceInfo->avaliableInstanceInfos.end()) {
        return InstanceSummary{"", ""};
    }
    auto insInfo = *resourceInfo->avaliableInstanceInfos.begin();
    absl::WriterMutexLock insInfoLock(&insInfo.second->mtx);
    insInfo.second->unfinishReqNum++;
    if (insInfo.second->unfinishReqNum >= static_cast<int>(resource.concurrency)) {
        insInfo.second->available = false;
        resourceInfo->avaliableInstanceInfos.erase(insInfo.first);
    }
    insInfo.second->idleTime = 0;
    if (insInfo.second->reporter) {
        insInfo.second->claimTime =
            std::chrono::time_point_cast<std::chrono::milliseconds>(std::chrono::steady_clock::now())
                .time_since_epoch()
                .count();
    }
    return InstanceSummary{insInfo.second->instanceId, insInfo.second->leaseId, insInfo.second->forceInvoke};
}

std::pair<bool, std::vector<std::string>> InsManager::NeedCancelCreatingIns(const RequestResource &resource,
                                                                            size_t reqNum, bool cleanAll)
{
    auto cancelIns = std::vector<std::string>();
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return std::make_pair(false, cancelIns);
    }
    {
        absl::ReaderMutexLock lock(&info->mtx);
        if (info->creatingIns.empty()) {
            return std::make_pair(false, cancelIns);
        }
    }
    if (cleanAll) {
        absl::WriterMutexLock lock(&info->mtx);
        if (reqNum > 0) {
            return std::make_pair(false, cancelIns);
        }
        auto deleteIns = std::remove_if(info->creatingIns.begin(), info->creatingIns.end(),
                                        [&cancelIns](const std::shared_ptr<CreatingInsInfo> &ins) {
                                            if (!ins->instanceId.empty()) {
                                                cancelIns.push_back(ins->instanceId);
                                                return true;
                                            }
                                            return false;
                                        });
        info->creatingIns.erase(deleteIns, info->creatingIns.end());
        {
            absl::WriterMutexLock lock(&createInstanceNumMutex);
            auto updateCreatingIns = totalCreatingInstanceNum_ - cancelIns.size();
            totalCreatingInstanceNum_ = (updateCreatingIns > 0) ? updateCreatingIns : 0;
        }
        YRLOG_DEBUG("add cancel all ins num {}, {}", cancelIns.size(), info->creatingIns.size());
        return std::make_pair(true, cancelIns);
    }
    absl::WriterMutexLock lock(&info->mtx);
    auto availableInsNum = info->avaliableInstanceInfos.size();
    auto creatingInsNum = info->creatingIns.size();
    auto requiredInsNum = GetRequiredInstanceNum(reqNum, resource.concurrency);
    YRLOG_DEBUG("requiredInsNum {}, creatingInsNum {}, availableInsNum {}", requiredInsNum, creatingInsNum,
                availableInsNum);
    if (requiredInsNum >= static_cast<int>(creatingInsNum) + static_cast<int>(availableInsNum)) {
        return std::make_pair(false, cancelIns);
    }
    auto &cancelCreatingIns = info->creatingIns.back();
    auto nowTime = std::chrono::time_point_cast<std::chrono::milliseconds>(std::chrono::steady_clock::now())
                       .time_since_epoch()
                       .count();
    auto waitingTime = nowTime - cancelCreatingIns->startTime;
    if (waitingTime < info->createTime || cancelCreatingIns->instanceId.empty()) {
        return std::make_pair(false, cancelIns);
    }
    cancelIns.push_back(cancelCreatingIns->instanceId);
    YRLOG_DEBUG("add cancel ins {}, creating ins {}, waitingTime {}, createTime {}", cancelCreatingIns->instanceId,
                info->creatingIns.size(), waitingTime, info->createTime);
    info->creatingIns.pop_back();
    DecreaseCreatingInstanceNum();
    return std::make_pair(true, cancelIns);
}

std::pair<bool, size_t> InsManager::NeedCreateNewIns(const RequestResource &resource, size_t reqNum,
                                                     bool considerWithFailNum)
{
    auto resourceInsInfo = GetRequestResourceInfo(resource);
    if (resourceInsInfo == nullptr) {
        return std::make_pair(false, 0);
    }
    size_t creatingInsNum = 0;
    int createFailNum = 0;
    int requiredInsNum = GetRequiredInstanceNum(reqNum, resource.concurrency);
    int availableInsNum = 0;

    int totalInsNum = GetCreateInstanceNum();
    int currentResourceInsNum = 0;
    {
        absl::ReaderMutexLock lock(&resourceInsInfo->mtx);
        creatingInsNum = resourceInsInfo->creatingIns.size();
        createFailNum = resourceInsInfo->createFailInstanceNum;
        currentResourceInsNum = static_cast<int>(resourceInsInfo->instanceInfos.size() + creatingInsNum);
        availableInsNum = resourceInsInfo->avaliableInstanceInfos.size();
    }
    YRLOG_DEBUG("ins info: required({}) creating({}) available({}) total({}) current resource({}) ", requiredInsNum,
                creatingInsNum, availableInsNum, totalInsNum, currentResourceInsNum);

    if (considerWithFailNum && (createFailNum > 0 && creatingInsNum > 0)) {
        YRLOG_INFO("current createfailnum is {}, creating num is {}, no need create more ins for function: {}",
                   createFailNum, creatingInsNum, resource.functionMeta.functionId);
        return std::make_pair(false, 0);
    }

    if (requiredInsNum <= static_cast<int>(creatingInsNum) + availableInsNum) {
        YRLOG_INFO("required ({}) < creating ({}) + available ({}); no need to create more for function: {}",
                   requiredInsNum, creatingInsNum, availableInsNum, resource.functionMeta.functionId);
        return std::make_pair(false, 0);
    }

    auto totalCreatingNum = GetCreatingInstanceNum();
    if (totalCreatingNum >= libRuntimeConfig->maxConcurrencyCreateNum) {
        YRLOG_INFO(
            "total creating ins num : {} is more than max concurrency create num : {}, should not create more ins for "
            "function: {}",
            totalCreatingNum, libRuntimeConfig->maxConcurrencyCreateNum, resource.functionMeta.functionId);
        return std::make_pair(false, 0);
    }

    bool exceedMaxTaskInsNum =
        libRuntimeConfig->maxTaskInstanceNum > 0
            ? (totalInsNum >= libRuntimeConfig->maxTaskInstanceNum) ||
                  (resource.opts.maxInstances != 0 && currentResourceInsNum >= resource.opts.maxInstances)
            : false;
    if (exceedMaxTaskInsNum) {
        YRLOG_INFO(
            "total ins num : {} is more than max concurrency create num: {} or {} is more than resource ins num limit: "
            "{}, should "
            "not create more ins for function: {}",
            totalInsNum, libRuntimeConfig->maxTaskInstanceNum, currentResourceInsNum, resource.opts.maxInstances,
            resource.functionMeta.functionId);
        return std::make_pair(false, 0);
    }
    return std::make_pair(true, GetDelayTime(createFailNum));
}

int InsManager::GetRequiredInstanceNum(int reqNum, int concurrency) const
{
    if (concurrency <= 0) {
        concurrency = DEFAULT_CONCURRENCY;
    }
    return ((reqNum - 1) / concurrency) + 1;  // ceil divide
}

void InsManager::AddCreatingInsInfo(const RequestResource &resource, std::shared_ptr<CreatingInsInfo> insInfo)
{
    auto resourceInfo = GetOrAddRequestResourceInfo(resource);
    IncreaseCreatingInstanceNum();
    absl::WriterMutexLock lock(&resourceInfo->mtx);
    auto &creatingInfo = resourceInfo->creatingIns;
    auto startTime = std::chrono::time_point_cast<std::chrono::milliseconds>(std::chrono::steady_clock::now())
                         .time_since_epoch()
                         .count();
    insInfo->startTime = startTime;
    creatingInfo.push_back(insInfo);
    YRLOG_DEBUG("add creating instance {}, {}", insInfo->instanceId, creatingInfo.size());
    return;
}

// If returns false, it means the instance has already been canceled
bool InsManager::EraseCreatingInsInfo(const RequestResource &resource, const std::string &instanceId,
                                      bool createSuccess)
{
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return false;
    }
    bool isErase;
    {
        absl::WriterMutexLock lock(&info->mtx);
        isErase = EraseCreatingInsInfoBare(info, instanceId, createSuccess);
    }
    info.reset();
    EraseResourceInfoMap(resource);
    return isErase;
}

bool InsManager::EraseCreatingInsInfoBare(std::shared_ptr<RequestResourceInfo> info, const std::string &instanceId,
                                          bool createSuccess)
{
    auto &creatingInfo = info->creatingIns;
    if (instanceId.empty() && !creatingInfo.empty()) {
        creatingInfo.pop_back();
        DecreaseCreatingInstanceNum();
        if (creatingInfo.size() == 0) {
            info->createFailInstanceNum = 0;
        }
        return true;
    }
    bool isErased = false;
    for (auto it = creatingInfo.begin(); it != creatingInfo.end();) {
        if (((*it))->instanceId == instanceId) {
            if (!createSuccess) {
                it = creatingInfo.erase(it);
                DecreaseCreatingInstanceNum();
                isErased = true;
                break;
            }
            auto nowTime = std::chrono::time_point_cast<std::chrono::milliseconds>(std::chrono::steady_clock::now())
                               .time_since_epoch()
                               .count();
            auto updateTime = nowTime - (*it)->startTime;
            if (info->createTime <= 0 && updateTime > 0) {
                info->createTime = updateTime;
            } else if (info->createTime > updateTime && updateTime > 0) {
                info->createTime = updateTime;
            }
            it = creatingInfo.erase(it);
            YRLOG_DEBUG("delete creating instance {}, {}", instanceId, creatingInfo.size());
            DecreaseCreatingInstanceNum();
            isErased = true;
            break;
        } else {
            ++it;
        }
    }
    if (creatingInfo.size() == 0) {
        info->createFailInstanceNum = 0;
    }
    return isErased;
}

void InsManager::ChangeCreateFailNum(const RequestResource &resource, bool isIncreaseOps)
{
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return;
    }
    absl::WriterMutexLock lock(&info->mtx);
    if (isIncreaseOps) {
        info->createFailInstanceNum++;
    } else {
        info->createFailInstanceNum = 0;
    }
}

void InsManager::DelInsInfo(const std::string &insId, const RequestResource &resource)
{
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return;
    }
    {
        absl::WriterMutexLock infoLock(&info->mtx);
        YRLOG_DEBUG("start delete ins info, ins id is {}, current totol ins num is {}", insId, GetCreatedInstanceNum());
        if (info->instanceInfos.find(insId) != info->instanceInfos.end()) {
            auto &insInfo = info->instanceInfos[insId];
            CancelScaleDownTimer(insInfo);
            DelInsInfoBare(insId, info);
        }
    }
    info.reset();
    EraseResourceInfoMap(resource);
}

void InsManager::EraseResourceInfoMap()
{
    auto resources = GetScheduleResources();
    for (const auto &resource : resources) {
        EraseResourceInfoMap(resource);
    }
}

std::vector<RequestResource> InsManager::GetScheduleResources()
{
    absl::ReaderMutexLock lock(&insMtx);
    std::vector<RequestResource> resources;
    for (const auto &pair : requestResourceInfoMap) {
        resources.push_back(pair.first);
    }
    return resources;
}

void InsManager::EraseResourceInfoMap(const RequestResource &resource, int currentCount)
{
    // yield avoid frequent cleaning requestResourceInfoMap
    std::this_thread::yield();
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return;
    }
    absl::WriterMutexLock lock(&insMtx);
    absl::WriterMutexLock infoLock(&info->mtx);
    if (info->instanceInfos.empty() && info->creatingIns.empty()) {
        // If the reference count is greater than the specified value,
        // it indicates that other threads are operating on RequestResourceInfo,
        // and the deletion operation should not be performed.
        if (info.use_count() <= currentCount) {
            YRLOG_DEBUG("remove resource info");
            requestResourceInfoMap.erase(resource);
        }
    }
}

// Deletes instance info without any locking; the caller must ensure locking.
void InsManager::DelInsInfoBare(const std::string &insId, std::shared_ptr<RequestResourceInfo> info)
{
    if (info->instanceInfos.erase(insId) > 0) {
        info->avaliableInstanceInfos.erase(insId);
        DecreaseCreatedInstanceNum();
    }
}

void InsManager::DecreaseUnfinishReqNum(const std::shared_ptr<InvokeSpec> spec, bool isInstanceNormal)
{
    auto resource = GetRequestResource(spec);
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return;
    }
    absl::WriterMutexLock lock(&info->mtx);
    auto id = spec->functionMeta.IsServiceApiType() ? spec->invokeLeaseId : spec->invokeInstanceId;
    if (info->instanceInfos.find(id) == info->instanceInfos.end()) {
        return;
    }
    absl::WriterMutexLock insLock(&info->instanceInfos[id]->mtx);
    auto insInfo = info->instanceInfos[id];
    insInfo->unfinishReqNum--;
    if (!insInfo->leaseId.empty() && insInfo->faasInfo.tLeaseInterval <= 0) {
        insInfo->available = false;
    } else if (insInfo->unfinishReqNum < static_cast<int>(resource.concurrency) && isInstanceNormal) {
        insInfo->available = true;
        if (info->avaliableInstanceInfos.find(id) == info->avaliableInstanceInfos.end()) {
            info->avaliableInstanceInfos.emplace(id, insInfo);
        }
    }
    YRLOG_DEBUG("unfinish req num: {}, req id: {}, ins id: {}", insInfo->unfinishReqNum, spec->requestId, id);
}

void InsManager::Stop()
{
    runFlag = false;
    {
        absl::WriterMutexLock lock(&leaseMtx);
        if (leaseTimer) {
            leaseTimer->cancel();
            leaseTimer.reset();
        }
    }
    if (tw_ != nullptr) {
        tw_->Stop();
    }
    tw_.reset();
    absl::WriterMutexLock lock(&insMtx);
    for (auto &pair : requestResourceInfoMap) {
        auto &requestResourceInfo = pair.second;
        absl::WriterMutexLock infoLock(&requestResourceInfo->mtx);
        requestResourceInfo->instanceInfos.clear();
        requestResourceInfo->avaliableInstanceInfos.clear();
    }
    requestResourceInfoMap.clear();
}

std::vector<std::string> InsManager::GetInstanceIds()
{
    std::vector<std::string> instanceIds;
    absl::ReaderMutexLock lock(&insMtx);
    for (auto &reqResInfo : requestResourceInfoMap) {
        auto &insMap = reqResInfo.second->instanceInfos;
        for (const auto &insInfo : insMap) {
            instanceIds.push_back(insInfo.second->instanceId);
        }
    }
    return instanceIds;
}

std::vector<std::string> InsManager::GetCreatingInsIds()
{
    std::vector<std::string> instanceIds;
    absl::ReaderMutexLock lock(&insMtx);
    for (auto &reqResInfo : requestResourceInfoMap) {
        for (const auto &ins : reqResInfo.second->creatingIns) {
            if (!ins->instanceId.empty()) {
                instanceIds.push_back(ins->instanceId);
            }
        }
    }
    return instanceIds;
}

bool InsManager::IsRemainIns(const RequestResource &resource)
{
    auto resourceInsInfo = GetRequestResourceInfo(resource);
    if (resourceInsInfo == nullptr) {
        return false;
    }
    absl::ReaderMutexLock lock(&resourceInsInfo->mtx);
    if (resourceInsInfo->creatingIns.size() > 0 || resourceInsInfo->instanceInfos.size() > 0) {
        return true;
    }
    return false;
}

void InsManager::SetDeleleInsCallback(const DeleteInsCallback &cb)
{
    deleteInsCallback_ = cb;
}

void InsManager::UpdateSchdulerInfo(const std::string &schedulerName, const std::string &schedulerId,
                                    const std::string &option)
{
    UpdateSchedulerOption opt = stringToOption(option);
    switch (opt) {
        case UpdateSchedulerOption::ADD:
            this->csHash->Add(schedulerName, schedulerId);
            break;
        case UpdateSchedulerOption::REMOVE:
            this->csHash->Remove(schedulerName);
            break;
        case UpdateSchedulerOption::UNKNOWN:
            YRLOG_ERROR("option: {} is not correct, do nothing about scheudler: {}", option, schedulerName);
            break;
    }
}
}  // namespace Libruntime
}  // namespace YR
