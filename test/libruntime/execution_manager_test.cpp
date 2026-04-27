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

#include <future>

#include <gtest/gtest.h>

#include "src/libruntime/invokeadaptor/ordered_execution_manager.h"
#include "src/libruntime/invokeadaptor/general_execution_manager.h"

namespace YR {
namespace test {

using namespace YR::Libruntime;
class ExecutionManagerTest : public testing::Test {
public:
    ExecutionManagerTest() = default;
    ~ExecutionManagerTest() = default;

protected:
    std::shared_ptr<ExecutionManager> execMgr;
};

TEST_F(ExecutionManagerTest, HandleOrderedRequestTest)
{
    execMgr = std::make_shared<OrderedExecutionManager>(1, nullptr);
    bool handled[10];
    for (int i = 0; i < 10; i++) {
        handled[i] = false;
    }

    std::vector<std::promise<int>> ps(10);
    std::vector<std::future<int>> fs;
    for (int i = 0; i < 10; i++) {
        fs.emplace_back(ps[i].get_future());
    }
    for (int i = 0; i < 10; i++) {

        libruntime::InvocationMeta meta;
        meta.set_invokerruntimeid("x");
        meta.set_invocationsequenceno(i);
        meta.set_minunfinishedsequenceno(0);
        execMgr->Handle(meta, [&handled, &ps, i]() {
            handled[i] = true;
            ps[i].set_value(1);
        });
    }
    for (int i = 0; i < 10; i++) {
        fs[i].get();
        ASSERT_TRUE(handled[i]);
    }
}

TEST_F(ExecutionManagerTest, HandleMisorderedRequestTest)
{
    execMgr = std::make_shared<OrderedExecutionManager>(1, nullptr);
    bool handled[10];
    for (int i = 0; i < 10; i++) {
        handled[i] = false;
    }
    std::vector<std::promise<int>> ps(10);
    std::vector<std::future<int>> fs;
    for (int i = 0; i < 10; i++) {
        fs.emplace_back(ps[i].get_future());
    }
    int index[10] = {0, 1, 5, 4, 3, 2, 6, 9, 7, 8};
    for (int i = 0; i < 10; i++) {
        libruntime::InvocationMeta meta;
        meta.set_invokerruntimeid("x");
        meta.set_invocationsequenceno(index[i]);
        meta.set_minunfinishedsequenceno(0);
        execMgr->Handle(meta, [&handled, &ps, index, i]() {
            handled[index[i]] = true;
            ps[index[i]].set_value(1);
        });
    }
    for (int i = 0; i < 10; i++) {
        fs[i].get();
        ASSERT_TRUE(handled[i]);
    }
}

TEST_F(ExecutionManagerTest, HandleNormalRequestTest)
{
    execMgr = std::make_shared<GeneralExecutionManager>(2, nullptr);
    ASSERT_EQ(execMgr->DoInit(2).OK(), true);

    bool handled = false;
    std::promise<int> promise;
    auto future = promise.get_future();
    libruntime::InvocationMeta meta;
    meta.set_minunfinishedsequenceno(0);
    execMgr->Handle(meta, [&handled, &promise]() {
        handled = true;
        promise.set_value(1);
    });
    future.get();
    ASSERT_TRUE(handled);
}

TEST_F(ExecutionManagerTest, isMultipleConcurrencyTest)
{
    execMgr = std::make_shared<GeneralExecutionManager>(1, nullptr);
    ASSERT_TRUE(!(execMgr->isMultipleConcurrency()));
}

TEST_F(ExecutionManagerTest, DoInitWithUserHookTest)
{
    std::function func = [](std::function<void(void)> userFuc)->void{};
    execMgr = std::make_shared<GeneralExecutionManager>(1, func);
    ASSERT_TRUE(execMgr->DoInit(1).OK());
    ASSERT_NO_THROW(execMgr->ErasePendingThread("reqID"));
}

TEST_F(ExecutionManagerTest, ErasePendingThreadTest)
{
    execMgr = std::make_shared<GeneralExecutionManager>(2, nullptr);
    execMgr->DoInit(2);
    ASSERT_NO_THROW(execMgr->ErasePendingThread("reqID"));
}

}  // namespace test
}  // namespace YR