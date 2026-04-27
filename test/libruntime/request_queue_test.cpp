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

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/libruntime/invokeadaptor/request_queue.h"

using namespace testing;
using namespace YR::Libruntime;

namespace YR {
namespace test {
class RequestQueueTest : public ::testing::Test {
public:
    RequestQueueTest() {}
    ~RequestQueueTest() {}
};

TEST_F(RequestQueueTest, TestPriorityQueue)
{
    auto queue = PriorityQueue();
    ASSERT_TRUE(queue.Empty());
    ASSERT_EQ(queue.Size(), 0);
    auto spec = std::make_shared<InvokeSpec>();
    spec->requestId = "1";
    spec->opts.priority = 1;
    queue.Push(spec);
    auto spec2 = std::make_shared<InvokeSpec>();
    spec2->requestId = "2";
    spec2->opts.priority = 2;
    queue.Push(spec2);
    ASSERT_FALSE(queue.Empty());
    ASSERT_EQ(queue.Size(), 2);
    ASSERT_EQ(queue.Top()->requestId, "2");
    queue.Pop();
    queue.Pop();
    ASSERT_TRUE(queue.Empty());
    ASSERT_EQ(queue.Size(), 0);
}

TEST_F(RequestQueueTest, TestQueue)
{
    auto queue = Queue();
    ASSERT_TRUE(queue.Empty());
    ASSERT_EQ(queue.Size(), 0);
    auto spec = std::make_shared<InvokeSpec>();
    spec->requestId = "1";
    spec->opts.priority = 1;
    queue.Push(spec);
    auto spec2 = std::make_shared<InvokeSpec>();
    spec2->requestId = "2";
    spec2->opts.priority = 2;
    queue.Push(spec2);
    ASSERT_FALSE(queue.Empty());
    ASSERT_EQ(queue.Size(), 2);
    ASSERT_EQ(queue.Top()->requestId, "1");
    queue.Pop();
    queue.Pop();
    ASSERT_TRUE(queue.Empty());
    ASSERT_EQ(queue.Size(), 0);
}

}  // namespace test
}  // namespace YR