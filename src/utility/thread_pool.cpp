/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include "src/utility/thread_pool.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace utility {
constexpr int QUEUE_WARN_SIZE = 10000;
ThreadPool::~ThreadPool() noexcept
{
    if (!stop_) {
        Shutdown();
    }
}

bool ThreadPool::IsInit()
{
    return init_;
}

std::string ThreadPool::TruncateThreadNamePrefix(const std::string &threadNamePrefix)
{
    static const size_t MAX_THREAD_NAME_LEN = 16;
    static const size_t THREAD_INDEX_LEN = 4;
    static const size_t MAX_THREAD_NAME_PREFIX_LEN = MAX_THREAD_NAME_LEN - THREAD_INDEX_LEN;
    return (threadNamePrefix.size() > MAX_THREAD_NAME_PREFIX_LEN)
               ? threadNamePrefix.substr(0, MAX_THREAD_NAME_PREFIX_LEN)
               : threadNamePrefix;
}

std::string ThreadPool::Init(size_t n, const std::string &threadNamePrefix)
{
    std::unique_lock<std::mutex> lock(mux_);
    if (init_) {
        return "";
    }
    stop_ = false;
    std::string prefix = TruncateThreadNamePrefix(threadNamePrefix);
    threadNamePrefix_ = prefix;
    for (size_t i = 0; i < n; i++) {
        try {
            std::thread t([this] { Work(); });
            std::string name = prefix + "." + std::to_string(i);
            pthread_setname_np(t.native_handle(), name.c_str());
            workers_[t.native_handle()] = std::move(t);
        } catch (const std::system_error &e) {
            return e.what();
        } catch (std::bad_alloc &e) {
            return e.what();
        }
    }
    init_ = true;
    return "";
}

void ThreadPool::InitAndRun(void)
{
    {
        std::unique_lock<std::mutex> lock(mux_);
        if (stop_) {
            return;
        }
        stop_ = false;
    }
    Work();
}

void ThreadPool::Handle(HandleFunc &&fn, const std::string &reqId)
{
    {
        std::lock_guard<std::mutex> lk(mux_);
        Job job;
        job.fn = std::move(fn);
        job.reqId = reqId;
        jobQueue_.emplace(std::move(job));
        if (jobQueue_.size() % QUEUE_WARN_SIZE == 0) {
            YRLOG_WARN("thread pool {} queue size reached {}", threadNamePrefix_, jobQueue_.size());
        }
    }
    cv_.notify_one();
}

void ThreadPool::Work()
{
    for (;;) {
        std::string reqId = "";
        HandleFunc fn;
        {
            std::unique_lock<std::mutex> lk(mux_);
            cv_.wait(lk, [this] { return jobQueue_.size() != 0 || stop_; });
            if (stop_) {
                break;
            }
            fn = std::move(jobQueue_.front().fn);
            reqId = jobQueue_.front().reqId;
            if (reqId != "") {
                std::unique_lock<std::mutex> lock(workThreadMutex_);
                workThread_[reqId] = pthread_self();
            }
            jobQueue_.pop();
        }
        fn();
        if (reqId != "") {
            std::unique_lock<std::mutex> lock(workThreadMutex_);
            workThread_.erase(reqId);
        }
    }
}

void ThreadPool::Shutdown()
{
    {
        std::unique_lock<std::mutex> lock(mux_);
        if (stop_) {
            return;
        }
        stop_ = true;
        init_ = false;
    }

    cv_.notify_all();
    for (auto &w : workers_) {
        if (w.second.joinable()) {
            w.second.join();
        }
    }
    for (auto &w : abandonedWorkers_) {
        if (w.joinable()) {
            w.join();
        }
    }
    workers_.clear();
}

void ThreadPool::ErasePendingThread(const std::string &reqId)
{
    char threadName[THREAD_NAME_LEN];
    pthread_t threadId = 0;
    {
        std::unique_lock<std::mutex> lock(workThreadMutex_);
        auto it = workThread_.find(reqId);
        if (it == workThread_.end()) {
            return;
        }
        pthread_getname_np(workThread_[reqId], threadName, THREAD_NAME_LEN);
        threadId = it->second;
    }

    std::thread workerThread;
    {
        std::unique_lock<std::mutex> lock(mux_);
        auto it = workers_.find(threadId);
        if (it != workers_.end()) {
            workerThread = std::move(it->second);
            YRLOG_DEBUG("erase pending thread from workers, req id is {}", reqId);
            workers_.erase(it);
        } else {
            return;
        }
    }

    {
        std::unique_lock<std::mutex> lock(workThreadMutex_);
        workThread_.erase(reqId);
    }

    {
        std::unique_lock<std::mutex> lock(mux_);
        std::thread t([this] { Work(); });
        pthread_setname_np(t.native_handle(), threadName);
        workers_[t.native_handle()] = std::move(t);
    }

    abandonedWorkers_.emplace_back(std::move(workerThread));
}

void ThreadPool::Stop(const std::vector<std::string> &requestIds)
{
    char threadName[THREAD_NAME_LEN];
    int ret;

    for (auto &reqId : requestIds) {
        if (reqId.empty()) {
            continue;
        }
        std::unique_lock<std::mutex> lock(workThreadMutex_);
        if (workThread_.find(reqId) != workThread_.end()) {
            pthread_getname_np(workThread_[reqId], threadName, THREAD_NAME_LEN);
            ret = pthread_cancel(workThread_[reqId]);
            if (ret != 0) {
                continue;
            }
            workers_.erase(workThread_[reqId]);
            workThread_.erase(reqId);

            std::thread t([this] { Work(); });
            pthread_setname_np(t.native_handle(), threadName);
            workers_[t.native_handle()] = std::move(t);
        }
    }
}
}  // namespace utility
}  // namespace YR