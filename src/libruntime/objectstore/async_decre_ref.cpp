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

#include "async_decre_ref.h"

namespace YR {
namespace Libruntime {
AsyncDecreRef::~AsyncDecreRef()
{
    Stop();
}
void AsyncDecreRef::Init(std::shared_ptr<DatasystemClientWrapper> clientWrapper)
{
    client = clientWrapper;
    running = true;
    nonEmpty = false;
    bgThread = std::thread(&AsyncDecreRef::Process, this);
    pthread_setname_np(bgThread.native_handle(), "async_decrease_reference");
}

void AsyncDecreRef::Stop() noexcept
{
    {
        std::lock_guard<std::mutex> lk(mu);
        running = false;
        cv.notify_one();
    }
    if (bgThread.joinable()) {
        bgThread.join();
    }
    client.reset();
}

bool AsyncDecreRef::Push(const std::vector<std::string> &objs, const std::string tenantId)
{
    std::lock_guard<std::mutex> lk(mu);
    if (!running) {
        return false;
    }
    if (!objs.empty()) {
        objQueue[tenantId].insert(objQueue[tenantId].end(), objs.begin(), objs.end());
        nonEmpty = true;
        cv.notify_one();
    }
    return true;
}

bool AsyncDecreRef::IsEmpty()
{
    std::lock_guard<std::mutex> lk(mu);
    return objQueue.empty();
}

bool AsyncDecreRef::PopBatch(std::vector<std::string> &objs, std::string &tenantId, std::size_t size)
{
    objs.clear();
    std::lock_guard<std::mutex> lk(mu);
    std::vector<std::string> popTenantId;
    bool isPopSuccessfully = false;
    for (auto it = objQueue.begin(); it != objQueue.end(); ++it) {
        tenantId = it->first;
        if (it->second.empty()) {
            popTenantId.push_back(it->first);
            continue;
        } else if (it->second.size() <= size) {
            objs.swap(it->second);
            popTenantId.push_back(it->first);
        } else {
            // Take out "size" elements from the end of the objQueue and return as objs.
            std::size_t splitBeginPos = it->second.size() - size;
            std::vector<std::string> selectedObjs(
                std::make_move_iterator(std::next(it->second.begin(), (int)splitBeginPos)),
                std::make_move_iterator(it->second.end()));
            it->second.erase(std::next(it->second.begin(), (int)splitBeginPos), it->second.end());
            objs.swap(selectedObjs);
        }
        isPopSuccessfully = true;
        break;
    }
    for (auto &id : popTenantId) {
        objQueue.erase(id);
    }
    if (!objQueue.empty()) {
        nonEmpty = true;
    } else {
        nonEmpty = false;
    }
    return isPopSuccessfully;
}

void AsyncDecreRef::Process()
{
    int limitedRetryTime = 0;
    while (running || !IsEmpty()) {
        {
            std::unique_lock<std::mutex> lk(mu);
            cv.wait(lk, [this] { return nonEmpty || !running; });
            nonEmpty = false;
        }
        std::vector<std::string> objs;
        std::string tenantId;
        if (!PopBatch(objs, tenantId, DECRE_REF_BATCH_SIZE)) {
            continue;
        }
        if (client == nullptr) {
            return;
        }
        client->SetTenantId(tenantId);
        std::vector<std::string> failedIds;
        datasystem::Status status = client->GDecreaseRef(objs, failedIds);
        auto err = DecreaseRefReturnCheck(status, failedIds);
        if (failedIds.empty()) {
            continue;
        }
        YRLOG_ERROR(err.Msg());
        // Error cannot retry. Just skip these ids.
        if (!IsRetryableStatus(status)) {
            YRLOG_WARN(
                "the StatusCode of GDecreaseRef returned is not Retryable, so skip all ids in objQueue. reason: {}",
                status.ToString());
            std::lock_guard<std::mutex> lk(mu);
            objQueue.clear();
        } else if (IsLimitedRetryEnd(status, limitedRetryTime)) {
            // Error retried for limited times. If limitedRetryTime>LIMITED_RETRY_TIME AND these retry are failed,
            // just skip these ids.
            YRLOG_WARN("GDecreaseRef has failed for {} times so skip all ids in objQueue, reason: {}",
                       LIMITED_RETRY_TIME, status.ToString());
            std::lock_guard<std::mutex> lk(mu);
            objQueue.clear();
        } else {
            std::lock_guard<std::mutex> lk(mu);
            objQueue[tenantId].reserve(objQueue[tenantId].size() + failedIds.size());
            std::move(std::begin(failedIds), std::end(failedIds), std::back_inserter(objQueue[tenantId]));  // append
            failedIds.clear();
            nonEmpty = true;
        }
        sleep(DECRE_RETRY_INTERVAL);
    }
}
}  // namespace Libruntime
}  // namespace YR