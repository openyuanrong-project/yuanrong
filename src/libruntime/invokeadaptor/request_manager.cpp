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

#include "request_manager.h"

namespace YR {
namespace Libruntime {

void RequestManager::PushFaasRequest(const std::string &leaseId, const std::shared_ptr<InvokeSpec> spec)
{
    absl::WriterMutexLock lock(&reqMtx_);
    requestMap[leaseId] = spec;
}

void RequestManager::PushRequest(const std::shared_ptr<InvokeSpec> spec)
{
    absl::WriterMutexLock lock(&reqMtx_);
    requestMap[spec->requestId] = spec;
}

bool RequestManager::PopRequest(const std::string &requestId, std::shared_ptr<InvokeSpec> &spec)
{
    absl::WriterMutexLock lock(&reqMtx_);
    auto it = requestMap.find(requestId);
    if (it == requestMap.end()) {
        return false;
    }
    spec = it->second;
    requestMap.erase(it);
    return true;
}

std::shared_ptr<InvokeSpec> RequestManager::GetRequest(const std::string &requestId)
{
    // 改成读写锁
    absl::ReaderMutexLock lock(&reqMtx_);
    auto it = requestMap.find(requestId);
    if (it == requestMap.end()) {
        return nullptr;
    }
    return it->second;
}

bool RequestManager::RemoveRequest(const std::string &requestId)
{
    absl::WriterMutexLock lock(&reqMtx_);
    auto it = requestMap.find(requestId);
    if (it == requestMap.end()) {
        return false;
    }
    (void)requestMap.erase(it);
    return true;
}

std::vector<std::string> RequestManager::GetObjIds()
{
    absl::ReaderMutexLock lock(&reqMtx_);
    std::vector<std::string> result;
    for (auto it = requestMap.begin(); it != requestMap.end(); it++) {
        auto spec = it->second;
        if (spec != nullptr) {
            for (auto &id : spec->returnIds) {
                result.emplace_back(id.id);
            }
        }
    }
    return result;
}

}  // namespace Libruntime
}  // namespace YR