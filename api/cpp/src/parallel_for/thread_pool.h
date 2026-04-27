/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: memory pool implementation
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
#include <cstdint>
#include <functional>
#include <mutex>
#include <thread>
#include <vector>

#include "event.h"
#include "mem_pool.h"
#include "task_queue.h"

namespace YR {
namespace Parallel {
extern const std::string threadNamePrefix;
constexpr uint32_t DEFAULT_WORKER_NUM = 8;
constexpr uint32_t MAX_WORKER_NUM = 64;
constexpr uint32_t MAX_MEM_POOL_NUM = 1024 * 128 + 64;
constexpr uint32_t MAX_TASK_QUEUE_NUM = 1024 * 128;
constexpr uint64_t MAX_64_BIT_MAP = 0x8000000000000000ul;
struct Task {
    std::function<void()> func;
};

class ThreadPool {
public:
    ThreadPool(const ThreadPool &) = delete;
    ThreadPool(ThreadPool &&) = delete;
    ThreadPool &operator=(const ThreadPool &) = delete;
    ThreadPool &operator=(ThreadPool &&) = delete;
    static ThreadPool &GetInstance();
    void Init(uint32_t threads = DEFAULT_WORKER_NUM);
    int SubmitTaskToPool(std::function<void()> &&task);
    void Stop();
    bool IsStop();

private:
    ThreadPool() = default;
    ~ThreadPool();
    void ThreadInit();
    void ThreadTask(int workerId);
    inline void WorkerIdle(int workerId);
    inline void WorkerWakeup(uint32_t taskNum);
    void WorkerJoin(int workerId);

    std::vector<std::thread> workers{};
    volatile uint64_t sleepWorkerBitmap = 0;
    TaskQueue tasks;
    MemPool memPool;
    uint32_t workerNum = DEFAULT_WORKER_NUM;
    std::atomic<bool> stopped = {true};
    struct WorkerSleepCtrl {
        Event event;
    };

    WorkerSleepCtrl sleepWorkers[MAX_WORKER_NUM];
};

inline void ThreadPool::WorkerWakeup(uint32_t taskNum)
{
    uint64_t wakeupId;
    bool success;
    uint32_t wakeupNum;
    uint64_t bitmap;
    wakeupNum = taskNum;
    bitmap = sleepWorkerBitmap;
    while (bitmap > 0) {
        wakeupId = GetLmb(bitmap);
        success = AtomicCmpset64(&sleepWorkerBitmap, bitmap, bitmap & (~(MAX_64_BIT_MAP >> wakeupId)));
        if (LIKELY(success)) {
            sleepWorkers[wakeupId].event.EventWakeUp();
            wakeupNum--;
            if (wakeupNum == 0) {
                return;
            }
        }
        bitmap = sleepWorkerBitmap;
    }
}

inline void ThreadPool::WorkerIdle(int workerId)
{
    uint64_t myBitmap = MAX_64_BIT_MAP >> static_cast<uint32_t>(workerId);
    sleepWorkers[workerId].event.EventReady();
    AtomicOr64(&sleepWorkerBitmap, myBitmap);
    sleepWorkers[workerId].event.EventWait();
}
}  // namespace Parallel
}  // namespace YR