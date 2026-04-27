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

#pragma once
#include <atomic>
#include <condition_variable>
#include <functional>
#include <mutex>
#include <thread>
#include <unordered_map>
#include <unordered_set>

#include "src/libruntime/err_type.h"
#include "src/libruntime/invokeadaptor/request_queue.h"

namespace YR {
namespace Libruntime {
using ScheduleFunc = std::function<void()>;
class TaskScheduler {
public:
    TaskScheduler() = default;
    ~TaskScheduler() = default;
    explicit TaskScheduler(const ScheduleFunc &func) : func_(func) {}

    void Run();
    void Stop();
    void Notify();

private:
    void Schedule();
    std::atomic<bool> runFlag_{false};
    bool scheduleFlag_ = false;
    std::mutex mtx_;
    std::condition_variable condVar_;
    ScheduleFunc func_;
    std::thread t;
};

class TaskSchedulerWrapper {
public:
    TaskSchedulerWrapper(bool enablePriority)
    {
        if (enablePriority) {
            queue = std::make_shared<PriorityQueue>();
        } else {
            queue = std::make_shared<Queue>();
        }
    }

    explicit TaskSchedulerWrapper(std::shared_ptr<TaskScheduler> taskScheduler, bool enablePriority)
        : taskScheduler_(taskScheduler)
    {
        if (enablePriority) {
            queue = std::make_shared<PriorityQueue>();
        } else {
            queue = std::make_shared<Queue>();
        }
    }

    ~TaskSchedulerWrapper() = default;
    void Notify();
    void SetLastError(const YR::Libruntime::ErrorInfo &err);
    YR::Libruntime::ErrorInfo GetLastError();
    bool IsLastErrorOk();
    std::shared_ptr<BaseQueue> queue;
private:
    std::shared_ptr<TaskScheduler> taskScheduler_;
    YR::Libruntime::ErrorInfo lastError_ = ErrorInfo();
    std::mutex errLock_;
};
}  // namespace Libruntime
}  // namespace YR