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

#include <chrono>
#include <memory>
#include <thread>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/dto/constant.h"
#include "src/utility/time_measurement.h"
#include "src/utility/timer_worker.h"

namespace YR {
namespace test {
using namespace YR::utility;
using namespace std::chrono_literals;
const static int delay = 10;
class TimerTest : public ::testing::Test {
    void SetUp() override
    {
        InitGlobalTimer();
    }

    void TearDown() override
    {
        CloseGlobalTimer();
    }
};

TEST_F(TimerTest, GlobalTimerTest)
{
    int counter = 0;
    size_t timeout_1 = 1;
    size_t timeout_2 = 2;
    int exec_once = 1;
    int exec_twice = 2;
    ExecuteByGlobalTimer([&counter]() { counter++; }, timeout_1, exec_once);
    ExecuteByGlobalTimer([&counter]() { counter++; }, timeout_2, exec_twice);
    std::this_thread::sleep_for(5.5ms * delay);
    ASSERT_EQ(counter, 3);
}

TEST_F(TimerTest, GlobalTimerCancelTest)
{
    int counter = 0;
    size_t timeout_1 = 10;
    size_t timeout_2 = 20;
    int exec_once = 1;
    auto t1 = ExecuteByGlobalTimer([&counter]() { counter++; }, timeout_1, exec_once);
    auto t2 = ExecuteByGlobalTimer([&counter]() { counter++; }, timeout_2, exec_once);
    std::this_thread::sleep_for(10ms);
    t2->cancel();
    std::this_thread::sleep_for(40ms * delay);
    ASSERT_EQ(counter, 1);
    int counter2 = 0;
    auto t3 = ExecuteByGlobalTimer([&counter2]() { counter2++; }, timeout_1, exec_once);
    auto t4 = ExecuteByGlobalTimer([&counter2]() { counter2++; }, timeout_2, exec_once);
    std::this_thread::sleep_for(10ms);
    CancelGlobalTimer(t4);
    std::this_thread::sleep_for(40ms * delay);
    ASSERT_EQ(counter2, 1);
}
TEST_F(TimerTest, GlobalTimerNullptrTest)
{
    CloseGlobalTimer();
    int counter = 0;
    size_t timeout_1 = 1;
    int exec_once = 1;
    auto ptr = ExecuteByGlobalTimer([&counter]() { counter++; }, timeout_1, exec_once);
    ASSERT_EQ(ptr, nullptr);
}

TEST_F(TimerTest, Test_timeworker_execute_times_should_be_right)
{
    int timeoutMs = 1;
    int executeTimes = 3;
    // TimerWorker uses weak_from_this() in CreateTimer; the object must be owned by shared_ptr.
    auto t = std::make_shared<TimerWorker>();
    std::atomic<int> count{0};
    auto f = [&count]() { count++; };
    ASSERT_NE(t->CreateTimer(timeoutMs, executeTimes, f), nullptr);
    std::this_thread::sleep_for(std::chrono::milliseconds(timeoutMs * executeTimes * delay));
    EXPECT_EQ(count.load(), executeTimes);
}

TEST_F(TimerTest, Test_TimeMeasurement)
{
    TimeMeasurement timeMeasurement(-1);
    auto res = timeMeasurement.GetEstimatedCostofNextRequest();
    ASSERT_EQ(res, -1);
    for (int i = 0; i < 8; ++i) {
        timeMeasurement.StartTimer("req" + std::to_string(i));
    }
    timeMeasurement.StopTimer("req-1", true);
    timeMeasurement.StopTimer("req0", true);
    res = timeMeasurement.GetEstimatedCostofNextRequest();
    ASSERT_NE(res, -1);
    for (int i = 1; i < 8; ++i) {
        timeMeasurement.StopTimer("req" + std::to_string(i), true);
    }
}

}  // namespace test
}  // namespace YR