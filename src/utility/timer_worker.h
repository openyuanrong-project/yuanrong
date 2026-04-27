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

#pragma once
#include <boost/asio.hpp>
#include "absl/synchronization/mutex.h"
#include <atomic>

using BoostTimer = boost::asio::deadline_timer;

namespace YR {
namespace utility {

class TimerWorker;

class Timer : public std::enable_shared_from_this<Timer> {
public:
    Timer(boost::asio::io_context &io, int timeoutMS, std::weak_ptr<TimerWorker> tw);
    ~Timer() = default;
    std::shared_ptr<BoostTimer> &GetTimer();
    void cancel();
    std::string ID();
    void Cancel();
private:
    std::string id_;
    std::shared_ptr<BoostTimer> timer_;
    std::weak_ptr<TimerWorker> weakTW_;
    inline static std::atomic<int> counter_ = 0;
};

class TimerWorker : public std::enable_shared_from_this<TimerWorker> {
public:
    TimerWorker();
    ~TimerWorker();
    TimerWorker(const TimerWorker &) = delete;
    TimerWorker &operator=(const TimerWorker &) = delete;
    std::shared_ptr<YR::utility::Timer> CreateTimer(int timeoutSeconds, int execTimes, std::function<void()> fn);
    void ExecuteByTimer(std::shared_ptr<YR::utility::Timer> t, int timeoutMilliSeconds, std::function<void()> fn);
    void ExecuteByTimer(std::shared_ptr<YR::utility::Timer> t, int timeoutMilliSeconds, int execTimes,
                        std::function<void()> fn);
    void EarseTimer(std::shared_ptr<YR::utility::Timer> t);
    void CancelTimer(std::shared_ptr<YR::utility::Timer> t);
    void Stop();

private:
    void TimeoutHandler(std::weak_ptr<Timer> t, int timeoutSeconds, int execTimes, std::function<void()> fn,
                        const boost::system::error_code &ec);
    absl::Mutex mu;
    absl::Mutex timerMu;
    bool isRunning ABSL_GUARDED_BY(mu);
    std::thread th;
    boost::asio::io_context io ABSL_GUARDED_BY(mu);
    std::unique_ptr<boost::asio::executor_work_guard<boost::asio::io_context::executor_type>> work;
    std::unordered_map<std::string, std::shared_ptr<YR::utility::Timer>> timerStore_ ABSL_GUARDED_BY(timerMu);
};

void InitGlobalTimer(void);

void CloseGlobalTimer(void);

std::shared_ptr<YR::utility::Timer> ExecuteByGlobalTimer(std::function<void()> fn, size_t timeoutSeconds,
                                                         int execTimes);

void ExecuteByGlobalTimer(std::function<void()> fn, size_t timeoutMilliSeconds, std::shared_ptr<YR::utility::Timer> t);

void CancelGlobalTimer(std::shared_ptr<YR::utility::Timer> t);
}  // namespace utility
}  // namespace YR
