/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
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

#include <pthread.h>
#include <atomic>
#include <condition_variable>
#include <functional>
#include <mutex>
#include <queue>
#include <thread>
#include <unordered_map>
#include <vector>

namespace YR {
namespace utility {
const int THREAD_NAME_LEN = 16;

class ThreadPool {
public:
    using HandleFunc = std::function<void()>;

    ThreadPool() {}

    virtual ~ThreadPool();

    ThreadPool(const ThreadPool &) = delete;
    ThreadPool(ThreadPool &&) = delete;
    ThreadPool &operator=(const ThreadPool &) = delete;
    ThreadPool &operator=(ThreadPool &&) = delete;

    std::string Init(size_t n, const std::string &threadNamePrefix = "rt_pool");

    void InitAndRun();

    void Handle(HandleFunc &&fn, const std::string &reqId);

    void Stop(const std::vector<std::string> &requestIds);

    void Shutdown();

    bool IsInit();

    void ErasePendingThread(const std::string &reqId);

private:
    std::string TruncateThreadNamePrefix(const std::string &threadNamePrefix);

    void Work();

    struct Job {
        HandleFunc fn;
        std::string reqId;
    };
    bool init_ = false;
    bool stop_ = false;
    std::unordered_map<pthread_t, std::thread> workers_;
    std::vector<std::thread> abandonedWorkers_;
    std::mutex mux_;
    std::condition_variable cv_;
    std::queue<Job> jobQueue_;
    std::unordered_map<std::string, pthread_t> workThread_;
    std::mutex workThreadMutex_;
    std::string threadNamePrefix_;
};
}  // namespace utility
}  // namespace YR