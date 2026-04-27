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

#include "task_scheduler.h"

namespace YR {
namespace Libruntime {
void TaskScheduler::Run()
{
    runFlag_ = true;
    // It is possible to optimize to have multiple schedulers share a single thread.
    t = std::thread([this]() { Schedule(); });
    pthread_setname_np(t.native_handle(), "task_scheduler");
}

void TaskScheduler::Schedule()
{
    for (;;) {
        std::unique_lock<std::mutex> lockGuard(mtx_);
        condVar_.wait(lockGuard, [this] { return scheduleFlag_ || !runFlag_; });
        if (!runFlag_) {
            break;
        }
        if (scheduleFlag_) {
            scheduleFlag_ = false;
            lockGuard.unlock();
            if (func_) {
                func_();
            }
        }
    }
}

void TaskScheduler::Stop()
{
    if (!runFlag_) {
        return;
    }
    {
        std::unique_lock<std::mutex> lockGuard(mtx_);
        runFlag_ = false;
        condVar_.notify_one();
    }
    if (t.joinable()) {
        t.join();
    }
}

void TaskScheduler::Notify()
{
    if (scheduleFlag_) {
        return;
    }
    std::unique_lock<std::mutex> lockGuard(mtx_);
    scheduleFlag_ = true;
    condVar_.notify_one();
}

void TaskSchedulerWrapper::SetLastError(const YR::Libruntime::ErrorInfo &err)
{
    std::lock_guard<std::mutex> lockGuard(errLock_);
    this->lastError_ = YR::Libruntime::ErrorInfo(err.Code(), err.MCode(), err.Msg());
}

YR::Libruntime::ErrorInfo TaskSchedulerWrapper::GetLastError()
{
    std::lock_guard<std::mutex> lockGuard(errLock_);
    return this->lastError_;
}

bool TaskSchedulerWrapper::IsLastErrorOk()
{
    std::lock_guard<std::mutex> lockGuard(errLock_);
    return this->lastError_.OK();
}

void TaskSchedulerWrapper::Notify()
{
    if (taskScheduler_) {
        taskScheduler_->Notify();
    }
}
}  // namespace Libruntime
}  // namespace YR