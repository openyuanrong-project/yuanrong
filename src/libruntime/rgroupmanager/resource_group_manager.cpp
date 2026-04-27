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

#include "resource_group_manager.h"

namespace YR {
namespace Libruntime {
void ResourceGroupManager::StoreRGDetail(const std::string &rgName, const std::string &requestId, int bundleSize)
{
    auto rGroupDetail = std::make_shared<ResourceGroupDetail>(bundleSize);
    absl::WriterMutexLock lock(&mu_);
    auto res = storeMap_.emplace(rgName, rGroupDetail);
    auto rgCreateErrInfo = std::promise<ErrorInfo>();
    auto rgCreateErrInfoFuture = rgCreateErrInfo.get_future();
    absl::WriterMutexLock rGroupDetailLock(&res.first->second->mu_);
    res.first->second->rgCreateErrInfo_.emplace(requestId, std::move(rgCreateErrInfo));
    res.first->second->rgCreateErrInfoFuture_.emplace(requestId, std::move(rgCreateErrInfoFuture));
}

void ResourceGroupManager::RemoveRGDetail(const std::string &rgName)
{
    absl::WriterMutexLock lock(&mu_);
    storeMap_.erase(rgName);
}

bool ResourceGroupManager::IsRGDetailExist(const std::string &rgName)
{
    absl::ReaderMutexLock lock(&mu_);
    auto it = storeMap_.find(rgName);
    if (it != storeMap_.end()) {
        return true;
    }
    return false;
}

int ResourceGroupManager::GetRGroupBundleSize(const std::string &rgName)
{
    absl::ReaderMutexLock lock(&mu_);
    auto it = storeMap_.find(rgName);
    if (it != storeMap_.end()) {
        return it->second->bundleSize_;
    }
    return -1;
}

void ResourceGroupManager::SetRGCreateErrInfo(const std::string &rgName, const std::string &requestId,
                                               const ErrorInfo &err)
{
    absl::ReaderMutexLock lock(&mu_);
    auto it = storeMap_.find(rgName);
    if (it == storeMap_.end()) {
        YRLOG_WARN("resource group: {} does not exist in store.", rgName);
        return;
    }
    std::shared_ptr<ResourceGroupDetail> rGroupDetail = it->second;
    absl::WriterMutexLock rGroupDetailLock(&rGroupDetail->mu_);
    auto rgInfoIt = rGroupDetail->rgCreateErrInfo_.find(requestId);
    if (rgInfoIt == rGroupDetail->rgCreateErrInfo_.end()) {
        YRLOG_WARN("requestId: {} of resource group: {} does not exist in rgroup detail.", requestId, rgName);
        return;
    }
    try {
        rgInfoIt->second.set_value(err);
    } catch (const std::future_error &e) {
        YRLOG_WARN("the value has already been set, rgName: {}, requestId: {}.", rgName, requestId);
    }
}

ErrorInfo ResourceGroupManager::GetRGCreateErrInfo(const std::string &rgName, const std::string &requestId,
                                                    int timeoutSec)
{
    std::shared_future<ErrorInfo> f;
    std::shared_ptr<ResourceGroupDetail> rGroupDetail;
    {
        absl::ReaderMutexLock lock(&mu_);
        auto it = storeMap_.find(rgName);
        if (it == storeMap_.end()) {
            std::string msg = "rgName: " + rgName + " does not exist in storeMap.";
            YRLOG_ERROR(msg);
            return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, msg);
        }
        rGroupDetail = it->second;
        absl::ReaderMutexLock rGroupDetailLock(&rGroupDetail->mu_);
        auto rgInfoIt = rGroupDetail->rgCreateErrInfoFuture_.find(requestId);
        if (rgInfoIt == rGroupDetail->rgCreateErrInfoFuture_.end()) {
            std::string msg =
                "requestId: " + requestId + " of resource group: " + rgName + " does not exist in rgroup detail.";
            YRLOG_ERROR(msg);
            return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, msg);
        }
        f = rgInfoIt->second;
    }
    if (timeoutSec != NO_TIMEOUT && f.wait_for(std::chrono::seconds(timeoutSec)) != std::future_status::ready) {
        std::string msg = "get resource group create errorinfo timeout, failed rgName: " + rgName + ".";
        YRLOG_ERROR(msg);
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, msg);
    }
    return f.get();
}
}  // namespace Libruntime
}  // namespace YR