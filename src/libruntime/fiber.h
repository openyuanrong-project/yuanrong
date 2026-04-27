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

#include <boost/fiber/all.hpp>
#include <chrono>

#include "event_notify.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
class FiberSemaphore {
public:
    explicit FiberSemaphore(int maxConcurrency)
        : count_(maxConcurrency) {
    }

    void Acquire()
    {
        std::unique_lock<boost::fibers::mutex> lock(mutex_);
        cv_.wait(lock, [this] { return count_ > 0; });
        --count_;
    }

    void Release()
    {
        {
            std::unique_lock<boost::fibers::mutex> lock(mutex_);
            ++count_;
        }
        cv_.notify_one();
    }

private:
    boost::fibers::mutex mutex_;
    boost::fibers::condition_variable cv_;
    int count_ = 1;
};

class FiberConcurrencyGuard {
public:
    explicit FiberConcurrencyGuard(FiberSemaphore& sem)
        : sem_(sem)
    {
        sem_.Acquire();
    }

    ~FiberConcurrencyGuard()
    {
        sem_.Release();
    }

    FiberConcurrencyGuard(const FiberConcurrencyGuard&) = delete;
    FiberConcurrencyGuard& operator=(const FiberConcurrencyGuard&) = delete;

private:
    FiberSemaphore& sem_;
};

class FiberPool {
public:
    FiberPool(size_t stackSize, int maxConcurrency)
        : stackAllocator_(stackSize), sem_(maxConcurrency), stoppedEvent_(std::make_shared<EventNotify<>>())
    {
        auto t = std::thread([this] { Work(); });
        t.detach();
    }

    void Work()
    {
        while (isRunning_) {
            std::function<void()> task;
            auto status = queue_.pop(task);
            if (status == boost::fibers::channel_op_status::success) {
                boost::fibers::fiber(boost::fibers::launch::dispatch, std::allocator_arg, stackAllocator_, task)
                    .detach();
            } else if (status == boost::fibers::channel_op_status::closed) {
                break;
            } else {
                YRLOG_ERROR("async actor fiber channel returned unexpected error code.");
                return;
            }
        }
        stoppedEvent_->Notify();
        while (true) {
            std::this_thread::sleep_for(std::chrono::hours(1));
        }
    }

    void Handle(std::function<void()> &&handler)
    {
        if (!isRunning_.load()) {
            return;
        }

        auto status = queue_.push([this, handler] {
            FiberConcurrencyGuard guard(sem_);
            handler();
        });
        if (status != boost::fibers::channel_op_status::success) {
            YRLOG_ERROR("something happend wrong when exec aysn func");
        }
    }

    void Shutdown()
    {
        isRunning_.store(false);
        queue_.close();
        stoppedEvent_->Wait();
    }

private:
    boost::fibers::fixedsize_stack stackAllocator_;
    boost::fibers::unbuffered_channel<std::function<void()>> queue_;
    FiberSemaphore sem_;
    std::shared_ptr<EventNotify<>> stoppedEvent_;

    std::atomic<bool> isRunning_ { true };
};

}  // namespace Libruntime
}  // namespace YR