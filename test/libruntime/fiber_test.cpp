/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <atomic>

#include "src/libruntime/fiber.h"


using namespace testing;
using namespace YR::utility;
using namespace YR::Libruntime;

namespace YR {
namespace test {
class FiberTest : public testing::Test {
public:
    FiberTest(){}
    ~FiberTest(){}
    void SetUp() override
    {
        Mkdir("/tmp/log");
        LogParam g_logParam = {
            .logLevel = "DEBUG",
            .logDir = "/tmp/log",
            .nodeName = "test-runtime",
            .modelName = "test",
            .maxSize = 100,
            .maxFiles = 1,
            .logFileWithTime = false,
            .logBufSecs = 30,
            .maxAsyncQueueSize = 1048510,
            .asyncThreadCount = 1,
            .alsoLog2Stderr = true,
        };
        InitLog(g_logParam);
    }
};

constexpr size_t FIBER_STACK_SIZE = 1024 * 256;

TEST_F(FiberTest, EmptyTest) 
{
    FiberPool fiberPool(FIBER_STACK_SIZE, 2);
    fiberPool.Shutdown();
}


class Counter {
  boost::fibers::condition_variable cv_;
  boost::fibers::mutex mtx_;
  std::atomic<int> total_{0};

 public:
  void Increment() {
    std::unique_lock<boost::fibers::mutex> lock(mtx_);
    total_++;
    cv_.notify_one();
  }

  void Waitfor(int total) {
    std::unique_lock<boost::fibers::mutex> lock(mtx_);
    cv_.wait(lock, [this, total]() { return this->total_ >= total; });
  }
};

class ConcuCounter {
 public:
  std::atomic<int> current_concurrency_num_{0};
  std::atomic<int> max_concurrency_num_{0};

  void increase_yeid_decrease() {
    current_concurrency_num_++;
    max_concurrency_num_.store(std::max(current_concurrency_num_, max_concurrency_num_));
    boost::this_fiber::sleep_for(std::chrono::milliseconds(10));
    current_concurrency_num_--;
  }
};

TEST_F(FiberTest, TaskFinishTest)
{
    FiberPool fiberPool(FIBER_STACK_SIZE, 2);
    fiberPool.Handle([&](){
        while (true) {
            boost::this_fiber::sleep_for(std::chrono::milliseconds(10));
            boost::this_fiber::yield();
        }
    });
    boost::this_fiber::sleep_for(std::chrono::seconds(1));
    fiberPool.Shutdown();
}

TEST(FiberStateTest, RespectsConcurrencyLimit)
{
    FiberPool fiberPool(FIBER_STACK_SIZE, 2);
    Counter couter;
    ConcuCounter conCouter;

    for (int i = 0; i < 100; i++) {
        fiberPool.Handle([&](){
            conCouter.increase_yeid_decrease();
            couter.Increment();
        });
    }

    couter.Waitfor(100);
    EXPECT_EQ(conCouter.max_concurrency_num_, 2);

    fiberPool.Shutdown();
}


}  // namespace test
}  // namespace YR