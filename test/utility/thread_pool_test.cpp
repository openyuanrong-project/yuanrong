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
#include <chrono>
#include <future>

#define private public
#include "src/utility/thread_pool.h"

namespace YR {
namespace test {
using namespace YR::utility;

class ThreadPoolTest : public ::testing::Test {
public:
    ThreadPoolTest() = default;
    ~ThreadPoolTest() = default;
};

TEST_F(ThreadPoolTest, Handle)
{
    std::atomic<int> sum(0);
    ThreadPool pool;
    pool.Init(4, "thread_pool_test");
    for (size_t i = 0; i < 50; i++) {
        pool.Handle([&]() { sum++; }, "");
    }
    while (sum < 50)
        std::this_thread::yield();
    pool.Shutdown();
    EXPECT_EQ(sum, 50);
}

TEST_F(ThreadPoolTest, VerifyThreadName)
{
    std::atomic<int> total(0);
    std::promise<void> prom;
    std::future<void> fut = prom.get_future();
    ThreadPool pool;
    pool.Init(2, "thread_pool_test");
    for (size_t i = 0; i < 2; i++) {
        pool.Handle(
            [&]() {
                char name[20];
                ASSERT_EQ(pthread_getname_np(pthread_self(), name, 20), 0);
                std::string name_str = name;
                ASSERT_TRUE(name_str == "thread_pool_.0" || name_str == "thread_pool_.1");
                total++;
                if (total >= 2) {
                    prom.set_value();
                }
            },
            "");
    }
    fut.get();
    pool.Shutdown();
}

TEST_F(ThreadPoolTest, TruncateThreadNamePrefix)
{
    ThreadPool pool;
    ASSERT_TRUE(pool.TruncateThreadNamePrefix("123") == "123");
    ASSERT_TRUE(pool.TruncateThreadNamePrefix("01234567890123456") == "012345678901");
    ASSERT_TRUE(pool.TruncateThreadNamePrefix("") == "");
    pool.Shutdown();
}

TEST_F(ThreadPoolTest, ErasePendingThreadTest)
{
    ThreadPool pool;
    pool.Init(1, "thread_pool_test");
    std::promise<int> prom1;
    std::future<int> fut1 = prom1.get_future();
    std::promise<void> prom2;
    std::future<void> fut2 = prom2.get_future();
    pool.Handle(
        [&]() {
            prom2.set_value();
            auto value = fut1.get();
            ASSERT_EQ(value, 1);
        },
        "reqId");
    fut2.get();
    ASSERT_EQ(pool.workThread_.size(), 1);
    ASSERT_EQ(pool.workers_.size(), 1);
    pool.ErasePendingThread("reqId");
    ASSERT_EQ(pool.workThread_.size(), 0);
    ASSERT_EQ(pool.abandonedWorkers_.size(), 1);
    ASSERT_EQ(pool.workers_.size(), 1);
    prom1.set_value(1);
    pool.Shutdown();
}

TEST_F(ThreadPoolTest, StopTest)
{
    ThreadPool pool;
    std::vector<std::string> requestIds = {"", "t1"};
    ASSERT_NO_THROW(pool.Stop(requestIds));
}
}  // namespace test
}  // namespace YR