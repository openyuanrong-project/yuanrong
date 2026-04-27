/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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

#include "yr/api/wait_request_manager.h"
#include "src/utility/logger/logger.h"
#include "yr/api/constant.h"

namespace {
const int S_TO_MS = 1000;
}

namespace YR {
namespace internal {
WaitRequest::WaitRequest(size_t waitNum_, int timeout_) : waitNum(waitNum_), timeout(timeout_)
{
    this->start = std::chrono::steady_clock::now();
}

void WaitRequest::Wait() const
{
    for (;;) {
        auto condition_func = [this]() { return finishCount >= waitNum || this->exceptionPtr != nullptr; };
        std::unique_lock<std::mutex> lk(mutex_);
        if (timeout == NO_TIMEOUT) {
            cv.wait(lk, condition_func);
        } else {
            cv.wait_for(lk, std::chrono::seconds(timeout), condition_func);
        }
        lk.unlock();
        if (this->exceptionPtr != nullptr) {
            YRLOG_DEBUG("WaitRequest exception throw");
            std::rethrow_exception(this->exceptionPtr);
        }

        if (timeout != NO_TIMEOUT &&
            std::chrono::duration_cast<std::chrono::milliseconds>(std::chrono::steady_clock::now() - start).count() >=
                this->timeout * S_TO_MS) {
            return;
        }

        if (finishCount >= waitNum) {
            YRLOG_DEBUG("Wait finishCount greater than or equal to waitNum, not need wait");
            return;
        }
    }
}

void WaitRequest::Notify()
{
    std::unique_lock<std::mutex> lk(mutex_);
    finishCount++;
    cv.notify_all();
}

void WaitRequest::SetException(const std::exception_ptr &exception)
{
    if (exceptionPtr == nullptr) {
        this->exceptionPtr = exception;
    }
}

WaitRequestManager::WaitRequestManager()
{
    this->ioc = std::make_shared<boost::asio::io_context>();
    this->work = std::make_unique<boost::asio::executor_work_guard<boost::asio::io_context::executor_type>>(
        boost::asio::make_work_guard(*ioc));
    this->asyncRunner = std::make_unique<std::thread>([&] { this->ioc->run(); });
    pthread_setname_np(this->asyncRunner->native_handle(), "wait_request_handler");
}

WaitRequestManager::~WaitRequestManager()
{
    this->work.reset();
    this->ioc->stop();
    if (this->asyncRunner->joinable()) {
        this->asyncRunner->join();
    }
}

void WaitRequestManager::Add(const std::string &id, const std::shared_ptr<WaitRequest> &waitRequest)
{
    requestStore[id].push_back(waitRequest);
}

void WaitRequestManager::Add(const std::vector<std::string> &ids, const std::shared_ptr<WaitRequest> &waitRequest)
{
    std::lock_guard<std::mutex> lk(mu);
    for (auto &id : ids) {
        Add(id, waitRequest);
    }
}

void WaitRequestManager::Remove(const std::string &id, const std::shared_ptr<WaitRequest> &waitRequest)
{
    auto iter = requestStore.find(id);
    if (iter == requestStore.end())
        return;
    auto &requests = iter->second;
    auto it = std::find(requests.begin(), requests.begin(), waitRequest);
    if (it == requests.end())
        return;
    std::iter_swap(it, std::prev(requests.end()));  // swap with the last element
    requests.pop_back();                            // pop the last element => erase it
    if (requests.empty()) {
        requestStore.erase(iter);
    }
}

void WaitRequestManager::NotifyTimeout(const std::shared_ptr<WaitRequest> &waitRequest)
{
    waitRequest->Notify();
}

void WaitRequestManager::SetReady(const std::string &id)
{
    std::lock_guard<std::mutex> lk(mu);
    auto iter = requestStore.find(id);
    if (iter != requestStore.end()) {
        auto &requests = iter->second;
        for (auto &request : requests) {
            request->Notify();
        }
    }
}

void WaitRequestManager::SetException(const std::string &id, const std::exception_ptr &exceptionPtr)
{
    std::lock_guard<std::mutex> lk(mu);
    YRLOG_INFO("Wait result Exception, id = {}", id);
    auto iter = requestStore.find(id);
    if (iter != requestStore.end()) {
        auto &requests = iter->second;
        for (auto &request : requests) {
            YRLOG_DEBUG("Set WaitRequest exception, id = {}", id);
            request->SetException(exceptionPtr);
            request->Notify();
        }
    }
}

void WaitRequestManager::WaitTimer(boost::asio::steady_timer &timer, int timeout,
                                   const std::shared_ptr<WaitRequest> &waitRequest)
{
    if (timeout != NO_TIMEOUT) {
        timer.expires_after(std::chrono::seconds(timeout));
        timer.async_wait([this, waitRequest](const boost::system::error_code &ec) {
            if (ec) {
                return;
            }
            this->NotifyTimeout(waitRequest);
        });
    }
}
}  // namespace internal
}  // namespace YR