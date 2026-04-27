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

#include "yr/yr.h"
#include "yr/api/function_manager.h"

namespace YR {
namespace test {
using namespace testing;
class FunctionManagerTest : public testing::Test {
public:
    FunctionManagerTest(){};
    ~FunctionManagerTest(){};
    void SetUp() override
    {}
};

class Counter {
public:
    Counter() {}
    Counter(int init)
    {
        count = init;
    }

    void Shutdown(uint64_t gracePeriodSecond) {
        return;
    }

    int count;
    std::string key = "";
    YR_STATE(key, count);
};


TEST_F(FunctionManagerTest, RegisterShutdownFunctionsTest) {
    YR_SHUTDOWN(&Counter::Shutdown);
    auto func = internal::FunctionManager::Singleton().GetShutdownFunction("Counter");
    EXPECT_TRUE(func.has_value());
}

TEST_F(FunctionManagerTest, CheckpointRecoverTest) {
    auto clsPtr = new Counter();
    clsPtr->key = "1234";
    auto instancePtr = YR::internal::Serialize((uint64_t)clsPtr);
    msgpack::sbuffer instanceBuf = YR::internal::Checkpoint<Counter>(instancePtr);
    instancePtr = YR::internal::Recover<Counter>(instanceBuf);
    Counter clsRef = YR::internal::ParseClassRef<Counter>(instancePtr);
    EXPECT_EQ(clsRef.key, clsPtr->key);
}

}  // namespace test
}  // namespace YR