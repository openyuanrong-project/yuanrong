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
#include <stdlib.h>
#include <fstream>
#include <iostream>

#include "src/libruntime/libruntime_config.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class LibruntimeConfigTest : public ::testing::Test {
public:
    LibruntimeConfigTest()
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
    }
    ~LibruntimeConfigTest() {}
    static void SetUpTestSuite()
    {
    }
};

extern bool fileExists(const std::string &path);

TEST_F(LibruntimeConfigTest, MergeConfigTest)
{
    LibruntimeConfig config;
    LibruntimeConfig configInput;
    configInput.jobId = "jobId";
    config.MergeConfig(configInput);
    ASSERT_EQ(config.jobId, configInput.jobId);
}

TEST_F(LibruntimeConfigTest, InitFunctionGroupRunningInfoTest)
{
    LibruntimeConfig config;
    common::FunctionGroupRunningInfo runningInfo;
    runningInfo.set_devicename("devicename");
    auto serverInfo = runningInfo.add_serverlist();
    serverInfo->set_serverid("serverid");
    auto deviceInfo = serverInfo->add_devices();
    deviceInfo->set_deviceid(123456);
    config.InitFunctionGroupRunningInfo(runningInfo);
    ASSERT_EQ(config.groupRunningInfo.deviceName, "devicename");
    ASSERT_EQ(config.groupRunningInfo.serverList.size(), 1);
}

TEST_F(LibruntimeConfigTest, GetInstanceIdTest)
{
    LibruntimeConfig config;
    libruntime::FunctionMeta meta;
    meta.set_name("name");
    config.funcMeta = meta;
    auto insId = config.GetInstanceId();
    ASSERT_EQ(insId, "yr_defalut_namespace-name");
    meta.set_ns("ns");
    config.funcMeta = meta;
    insId = config.GetInstanceId();
    ASSERT_EQ(insId, "ns-name");
}

}  // namespace test
}  // namespace YR