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

#include <sys/select.h>
#include <functional>
#include <iostream>
#include <memory>
#include <mutex>
#include <string>
#include <thread>
#include <unordered_map>
#include <vector>

#include "boost/asio.hpp"
#include "boost/chrono/chrono.hpp"

namespace YR {
namespace internal {
class WaitRequest {
    friend class WaitRequestManager;

public:
    ~WaitRequest() = default;
    WaitRequest(size_t waitNum, int timeout);

private:
    void Wait() const;
    void Notify();
    void SetException(const std::exception_ptr &exception);

private:
    size_t waitNum = 0;
    int timeout = -1;
    size_t finishCount = 0;
    std::chrono::time_point<std::chrono::steady_clock> start;
    mutable std::mutex mutex_;
    mutable std::condition_variable cv;
    std::exception_ptr exceptionPtr = nullptr;
};
class LocalModeRuntime;
class WaitRequestManager {
    friend class LocalModeRuntime;

public:
    WaitRequestManager();
    ~WaitRequestManager();

private:
    void SetReady(const std::string &id);
    void SetException(const std::string &id, const std::exception_ptr &exceptionPtr);

    // for cluster mdoe, T = std::string
    // for local mode, T = ObjectRef<>
    // isReady and getId are customed functions defined by local mode/cluster mode
    template <typename T>
    std::vector<bool> Wait(const std::vector<T> &objs, std::size_t waitNum, int timeoutMS,
                           std::function<bool(const T &)> isReady, std::function<std::string(const T &)> getId);

    void Add(const std::vector<std::string> &ids, const std::shared_ptr<WaitRequest> &waitRequest);

    void Add(const std::string &id, const std::shared_ptr<WaitRequest> &waitRequest);

    void Remove(const std::string &id, const std::shared_ptr<WaitRequest> &waitRequest);

    void NotifyTimeout(const std::shared_ptr<WaitRequest> &waitRequest);

    void WaitTimer(boost::asio::steady_timer &timer, int timeout, const std::shared_ptr<WaitRequest> &waitRequest);

    std::mutex mu;
    std::unordered_map<std::string, std::vector<std::shared_ptr<WaitRequest>>> requestStore;

    std::shared_ptr<boost::asio::io_context> ioc;
    std::unique_ptr<boost::asio::executor_work_guard<boost::asio::io_context::executor_type>> work;
    std::unique_ptr<std::thread> asyncRunner;
};

template <typename T>
std::vector<bool> WaitRequestManager::Wait(const std::vector<T> &objs, std::size_t waitNum, int timeout,
                                           std::function<bool(const T &)> isReady,
                                           std::function<std::string(const T &)> getId)
{
    std::vector<bool> results(objs.size());
    std::size_t finishNum = 0;
    std::vector<std::size_t> remainingIndex;
    remainingIndex.reserve(objs.size());

    // stop set ready when init a wait request in case of missing notification
    // check finished requests before enter waiting
    std::shared_ptr<WaitRequest> waitRequest;
    {
        std::lock_guard<std::mutex> lk(mu);
        for (std::size_t i = 0; i < objs.size(); ++i) {
            if (isReady(objs[i])) {
                results[i] = true;
                finishNum++;
            } else {
                remainingIndex.push_back(i);
            }
        }
        if (finishNum >= waitNum) {
            return results;
        }
        waitRequest = std::make_shared<WaitRequest>(waitNum - finishNum, timeout);
        for (auto &i : remainingIndex) {
            Add(getId(objs[i]), waitRequest);
        }
    }

    boost::asio::steady_timer timer(*this->ioc);
    WaitTimer(timer, timeout, waitRequest);
    waitRequest->Wait();
    for (std::size_t i = 0; i < objs.size(); ++i) {
        if (results[i] == false && isReady(objs[i])) {
            results[i] = true;
            if (++finishNum >= waitNum) {
                break;
            }
        }
    }
    {
        std::lock_guard<std::mutex> lk(mu);
        for (auto &i : remainingIndex) {
            Remove(getId(objs[i]), waitRequest);
        }
    }
    return results;
}
}  // namespace internal
}  // namespace YR
