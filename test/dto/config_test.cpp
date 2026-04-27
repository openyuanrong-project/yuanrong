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
#include <cstdlib>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#define private public
#include "src/dto/config.h"
using namespace YR::Libruntime;
class DTOConfigTest : public testing::Test {
public:
    DTOConfigTest(){};
    ~DTOConfigTest(){};
    void SetUp() override {
        Config::c = Config();
    }
    void TearDown() override {}
};

TEST_F(DTOConfigTest, TestConfig)
{
    ASSERT_EQ(Config::Instance().REQUEST_ACK_ACC_MAX_SEC(), 1800);
    ASSERT_EQ(Config::Instance().DS_CONNECT_TIMEOUT_SEC(), 60);

    setenv("MOCK_ENV1", "5", 0);
    size_t mockVal1 = Config::Instance().ParseFromEnv<size_t>(
        "MOCK_ENV1", 1800, [](const size_t &val) -> bool { return val >= REQUEST_ACK_TIMEOUT_SEC; });
    ASSERT_EQ(mockVal1, 5);

    setenv("MOCK_ENV2", "1", 0);
    size_t mockVal2 = Config::Instance().ParseFromEnv<size_t>(
        "MOCK_ENV2", 1800, [](const size_t &val) -> bool { return val >= REQUEST_ACK_TIMEOUT_SEC; });
    ASSERT_EQ(mockVal2, 1800);
}

TEST_F(DTOConfigTest, TestGetMaxArgsInMsgBytes)
{
    ASSERT_EQ(Config::Instance().MAX_ARGS_IN_MSG_BYTES(), 100 * 1024);
    setenv("RUNTIME_DIRECT_CONNECTION_ENABLE", "true", 1);
    Config::c = Config();
    ASSERT_EQ(Config::Instance().MAX_ARGS_IN_MSG_BYTES(), 10 * 1024 * 1024);
    unsetenv("RUNTIME_DIRECT_CONNECTION_ENABLE");

    setenv("MAX_ARGS_IN_MSG_BYTES", "10", 1);
    Config::c = Config();
    ASSERT_EQ(Config::Instance().MAX_ARGS_IN_MSG_BYTES(), 10);
    unsetenv("MAX_ARGS_IN_MSG_BYTES");

    setenv("RUNTIME_DIRECT_CONNECTION_ENABLE", "true", 1);
    setenv("MAX_ARGS_IN_MSG_BYTES", "100", 1);
    Config::c = Config();
    ASSERT_EQ(Config::Instance().MAX_ARGS_IN_MSG_BYTES(), 100);
    unsetenv("MAX_ARGS_IN_MSG_BYTES");
    unsetenv("RUNTIME_DIRECT_CONNECTION_ENABLE");

    ASSERT_EQ(Config::Instance().MEM_STORE_SIZE_THRESHOLD(), 100 * 1024);
    setenv("MEM_STORE_SIZE_THRESHOLD", "100", 1);
    Config::c = Config();
    ASSERT_EQ(Config::Instance().MEM_STORE_SIZE_THRESHOLD(), 100);
    unsetenv("MEM_STORE_SIZE_THRESHOLD");

    ASSERT_EQ(Config::Instance().FASS_SCHEDULE_TIMEOUT(), 120);
    setenv("FASS_SCHEDULE_TIMEOUT", "100", 1);
    Config::c = Config();
    ASSERT_EQ(Config::Instance().FASS_SCHEDULE_TIMEOUT(), 100);
    unsetenv("FASS_SCHEDULE_TIMEOUT");

    ASSERT_EQ(Config::Instance().YR_MAX_LOG_SIZE_MB(), 500);
    ASSERT_EQ(Config::Instance().YR_MAX_LOG_FILE_NUM(), 10);
    setenv("YR_MAX_LOG_SIZE_MB", "100", 1);
    setenv("YR_MAX_LOG_FILE_NUM", "5", 1);
    Config::c = Config();
    ASSERT_EQ(Config::Instance().YR_MAX_LOG_SIZE_MB(), 100);
    ASSERT_EQ(Config::Instance().YR_MAX_LOG_FILE_NUM(), 5);
    unsetenv("YR_MAX_LOG_SIZE_MB");
    unsetenv("YR_MAX_LOG_FILE_NUM");
}

TEST_F(DTOConfigTest, TestSetenv)
{
    ASSERT_FALSE(Config::Instance().YR_ENABLE_HTTP_PROXY());
    setenv("YR_ENABLE_HTTP_PROXY", "true", 1);
    Config::c = Config();
    ASSERT_TRUE(Config::Instance().YR_ENABLE_HTTP_PROXY());
    unsetenv("YR_ENABLE_HTTP_PROXY");
}