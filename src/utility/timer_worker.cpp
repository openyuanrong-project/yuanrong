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

#include "src/utility/timer_worker.h"

#include <boost/bind/bind.hpp>
#include <iostream>

#include "logger/logger.h"
const size_t timerIdLength = 16;

namespace YR {
namespace utility {
using namespace boost::placeholders;
static std::shared_ptr<TimerWorker> timerWorker = nullptr;

Timer::Timer(boost::asio::io_context &io, int timeoutMS, std::weak_ptr<TimerWorker> tw) : weakTW_(tw)
{
    id_ = std::to_string(counter_.fetch_add(1));
    timer_ = std::make_shared<BoostTimer>(io, boost::posix_time::milliseconds(timeoutMS));
}

std::shared_ptr<BoostTimer> &Timer::GetTimer()
{
    return timer_;
}

void Timer::cancel()
{
    if (auto tw = weakTW_.lock(); tw) {
        timer_->cancel();
        std::shared_ptr<Timer> self;
        try {
            self = shared_from_this();
        } catch (const std::exception &e) {
            YRLOG_ERROR("Timer has been destructed, {}", e.what());
            return;
        }
        tw->EarseTimer(self);
    }
}

void Timer::Cancel()
{
    timer_->cancel();
}

std::string Timer::ID()
{
    return id_;
}

TimerWorker::TimerWorker() : isRunning(true)
{
    th = std::thread([&] {
        work = std::make_unique<boost::asio::executor_work_guard<boost::asio::io_context::executor_type>>(
            boost::asio::make_work_guard(io));
        ;
        io.run();
    });
    pthread_setname_np(th.native_handle(), "TimerWorker");
}

TimerWorker::~TimerWorker()
{
    Stop();
}

void TimerWorker::TimeoutHandler(std::weak_ptr<Timer> weakT, int timeoutMilliSeconds, int execTimes,
                                 std::function<void()> fn, const boost::system::error_code &ec)
{
    if (ec) {
        return;
    }
    auto needErase = false;
    auto t = weakT.lock();
    if (t == nullptr) {
        return;
    }
    {
        absl::ReaderMutexLock lock(&this->mu);
        if (!isRunning) {
            return;
        }

        if (execTimes == -1) {
            t->GetTimer()->expires_at(t->GetTimer()->expires_at() +
                                      boost::posix_time::milliseconds(timeoutMilliSeconds));
            t->GetTimer()->async_wait(boost::bind(&TimerWorker::TimeoutHandler, this, weakT, timeoutMilliSeconds,
                                                  execTimes, fn, boost::asio::placeholders::error));
        } else if (auto remainTime = (execTimes - 1); remainTime > 0) {
            t->GetTimer()->expires_at(t->GetTimer()->expires_at() +
                                      boost::posix_time::milliseconds(timeoutMilliSeconds));
            t->GetTimer()->async_wait(boost::bind(&TimerWorker::TimeoutHandler, this, weakT, timeoutMilliSeconds,
                                                  remainTime, fn, boost::asio::placeholders::error));
        } else {
            needErase = true;
        }
    }
    if (needErase) {
        absl::WriterMutexLock lock(&this->timerMu);
        if (timerStore_.find(t->ID()) != timerStore_.end()) {
            timerStore_.erase(t->ID());
        }
    }
    fn();
}

std::shared_ptr<YR::utility::Timer> TimerWorker::CreateTimer(int timeoutMilliSeconds, int execTimes,
                                                             std::function<void()> fn)
{
    auto t = std::make_shared<Timer>(io, timeoutMilliSeconds, weak_from_this());
    {
        absl::ReaderMutexLock lock(&this->mu);
        if (!isRunning) {
            return nullptr;
        }

        std::weak_ptr<Timer> weakT = t;
        t->GetTimer()->async_wait(boost::bind(&TimerWorker::TimeoutHandler, this, weakT, timeoutMilliSeconds, execTimes,
                                              fn, boost::asio::placeholders::error));
    }

    {
        absl::WriterMutexLock lock(&this->timerMu);
        auto ret = timerStore_.emplace(t->ID(), t);
        if (!ret.second) {
            return nullptr;
        }
    }
    return t;
}

void TimerWorker::ExecuteByTimer(std::shared_ptr<YR::utility::Timer> t, int timeoutMilliSeconds,
                                 std::function<void()> fn)
{
    this->ExecuteByTimer(t, timeoutMilliSeconds, 1, fn);
}

void TimerWorker::ExecuteByTimer(std::shared_ptr<YR::utility::Timer> t, int timeoutMilliSeconds, int execTimes,
                                 std::function<void()> fn)
{
    {
        absl::ReaderMutexLock lock(&this->mu);
        if (!isRunning) {
            return;
        }
        if (t == nullptr) {
            YRLOG_WARN("timer is null, will not execute.");
            return;
        }

        std::weak_ptr<Timer> weakT = t;
        t->GetTimer()->expires_at(t->GetTimer()->expires_at() + boost::posix_time::milliseconds(timeoutMilliSeconds));
        t->GetTimer()->async_wait(boost::bind(&TimerWorker::TimeoutHandler, this, weakT, timeoutMilliSeconds, execTimes,
                                              fn, boost::asio::placeholders::error));
    }

    {
        absl::WriterMutexLock lock(&this->timerMu);
        auto ret = timerStore_.emplace(t->ID(), t);
        if (!ret.second) {
            YRLOG_DEBUG("timer {} already existed.", t->ID());
        }
    }
}

void TimerWorker::EarseTimer(std::shared_ptr<YR::utility::Timer> t)
{
    if (!t) {
        return;
    }
    absl::WriterMutexLock lock(&this->timerMu);
    if (timerStore_.find(t->ID()) != timerStore_.end()) {
        timerStore_.erase(t->ID());
    }
}

void TimerWorker::CancelTimer(std::shared_ptr<YR::utility::Timer> t)
{
    if (!t) {
        return;
    }
    (void)t->GetTimer()->cancel();
    this->EarseTimer(t);
}

void TimerWorker::Stop()
{
    {
        absl::WriterMutexLock lock(&this->timerMu);
        for (const auto &pair : timerStore_) {
            pair.second->Cancel();
        }
        timerStore_.clear();
    }

    {
        absl::WriterMutexLock lock(&this->mu);
        isRunning = false;
        if (work) {
            work->reset();
        }
        if (!io.stopped()) {
            io.stop();
        }
    }

    if (th.joinable()) {
        th.join();
    }
}

void InitGlobalTimer(void)
{
    if (timerWorker == nullptr) {
        timerWorker = std::make_shared<TimerWorker>();
    }
}

void CloseGlobalTimer(void)
{
    if (timerWorker != nullptr) {
        timerWorker->Stop();
    }
    timerWorker.reset();
}

std::shared_ptr<YR::utility::Timer> ExecuteByGlobalTimer(std::function<void()> fn, size_t timeoutMilliSeconds,
                                                         int execTimes)
{
    if (execTimes == 0) {
        return nullptr;
    }
    if (timerWorker == nullptr) {
        YRLOG_ERROR("timerWorker is nullptr.");
        return nullptr;
    }
    return timerWorker->CreateTimer(timeoutMilliSeconds, execTimes, fn);
}

void ExecuteByGlobalTimer(std::function<void()> fn, size_t timeoutMilliSeconds, std::shared_ptr<YR::utility::Timer> t)
{
    if (timerWorker == nullptr) {
        YRLOG_ERROR("timerWorker is nullptr.");
        return;
    }
    timerWorker->ExecuteByTimer(t, timeoutMilliSeconds, fn);
}

void CancelGlobalTimer(std::shared_ptr<YR::utility::Timer> t)
{
    timerWorker->CancelTimer(t);
}
}  // namespace utility
}  // namespace YR