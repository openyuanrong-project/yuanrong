/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

#include "src/libruntime/libruntime_manager.h"

using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class LibruntimeManagerTest : public testing::Test {
public:
    LibruntimeManagerTest(){};
    ~LibruntimeManagerTest(){};
    void SetUp() override {}
    void TearDown() override {}
};

TEST_F(LibruntimeManagerTest, InitFinalizeTest)
{
    YR::Libruntime::LibruntimeConfig libConfig;
    libConfig.inCluster = true;
    libConfig.isDriver = true;
    libConfig.jobId = YR::utility::IDGenerator::GenApplicationId();
    libConfig.functionSystemIpAddr = "127.0.0.1";
    libConfig.functionSystemPort = 1110;
    libConfig.dataSystemIpAddr = "127.0.0.1";
    libConfig.dataSystemPort = 1100;
    auto rt = LibruntimeManager::Instance().GetLibRuntime("");
    ASSERT_EQ(rt, nullptr);
    bool isInitialized = LibruntimeManager::Instance().IsInitialized("");
    ASSERT_FALSE(isInitialized);
    auto errInfo = LibruntimeManager::Instance().Init(libConfig, "");
    rt = LibruntimeManager::Instance().GetLibRuntime("");
    ASSERT_EQ(rt, nullptr) << errInfo.Code() << errInfo.Msg();
    isInitialized = LibruntimeManager::Instance().IsInitialized("");
    ASSERT_FALSE(isInitialized) << errInfo.Code() << errInfo.Msg();

    LibruntimeManager::Instance().Finalize("");
    rt = LibruntimeManager::Instance().GetLibRuntime("");
    ASSERT_EQ(rt, nullptr);
    isInitialized = LibruntimeManager::Instance().IsInitialized("");
    ASSERT_FALSE(isInitialized);
}

TEST_F(LibruntimeManagerTest, InitFailedWhenInputInvalidRecycleTime)
{
    YR::Libruntime::LibruntimeConfig libConfig;
    libConfig.recycleTime = 0;
    auto errInfo = LibruntimeManager::Instance().Init(libConfig, "");
    ASSERT_FALSE(errInfo.OK());
    libConfig.recycleTime = 3001;
    errInfo = LibruntimeManager::Instance().Init(libConfig, "");
    ASSERT_FALSE(errInfo.OK());
}

TEST_F(LibruntimeManagerTest, HandleInitializedTest)
{
    YR::Libruntime::LibruntimeConfig libConfig;
    libConfig.functionIds[libruntime::LanguageType::Cpp] = "cpp";
    auto errInfo = LibruntimeManager::Instance().HandleInitialized(libConfig, "test");
    ASSERT_EQ(errInfo.OK(), true);
}
}  // namespace test
}  // namespace YR