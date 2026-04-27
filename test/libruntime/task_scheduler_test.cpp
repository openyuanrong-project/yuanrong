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

#include <thread>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <future>
#include "src/libruntime/invokeadaptor/task_scheduler.h"
using namespace testing;
using namespace YR::Libruntime;

namespace YR {

namespace test {
class TaskSchedulerTest : public testing::Test {
public:
    TaskSchedulerTest(){};
    ~TaskSchedulerTest(){};
    static void SetUpTestCase() {};
    static void TearDownTestCase() {};
    void SetUp() override {}
    void TearDown() override {}
};
TEST_F(TaskSchedulerTest, DeduplicationTaskRequest)
{
    std::promise<int> promise;
    auto future = promise.get_future();
    auto scheduleFunc = [&promise]() { promise.set_value(1); };
    auto taskScheduler = std::make_shared<TaskScheduler>(scheduleFunc);
    taskScheduler->Run();
    taskScheduler->Notify();
    ASSERT_EQ(1, future.get());
    taskScheduler->Stop();
}
}  // namespace test
}  // namespace YR
