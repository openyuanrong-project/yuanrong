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

#include "gtest/gtest.h"
#include "gmock/gmock.h"

#include <filesystem>

#include "yr/yr.h"
#include "user_common_func.h"

class InitTest : public testing::Test {
public:
    InitTest(){};
    ~InitTest(){};
    static void SetUpTestCase(){};
    static void TearDownTestCase(){};
    void SetUp()
    {
    };
    void TearDown()
    {
        YR::Finalize();
    };
};

size_t GetRuntimeNum(std::string jobId)
{
    int cnt = 0;
    std::string tmpJobId = jobId.substr(4);

    const char *deployPath = ::getenv("DEPLOY_PATH");
    if (deployPath == nullptr) {
        std::cout << "Failed to get DEPLOY_PATH" << std::endl;
        return cnt;
    }
    auto tmp = std::string(deployPath) + "/yr_agent/";

    std::cout << "Deploy path " << deployPath << std::endl;

    using std::filesystem::recursive_directory_iterator;

    for (auto& v : recursive_directory_iterator(tmp))
    {
        std::string fileName = v.path().filename().string();
        if (fileName.find(tmpJobId) != std::string::npos && fileName.find("INFO.log") != std::string::npos) {
            cnt++;
        }
    }
    std::cout << "runtime num: " << cnt << std::endl;
    return cnt;
}

TEST_F(InitTest, InitFailedWhenMaxConcurrencyCreateNumIs0)
{
    YR::Config config;
    config.mode = YR::Config::Mode::CLUSTER_MODE;
    config.maxConcurrencyCreateNum = 0;
    ASSERT_THROW(YR::Init(config), YR::Exception);
    config.mode = YR::Config::Mode::CLUSTER_MODE;
    config.maxConcurrencyCreateNum = 1;
    YR::Init(config);
}

TEST_F(InitTest, InitSuccessAndReturnServerVersionRight)
{
    YR::Config config;
    config.mode = YR::Config::Mode::CLUSTER_MODE;
    config.logLevel = "DEBUG";
    auto clientInfo = YR::Init(config);
    std::cout << "client version: " << clientInfo.version << std::endl;
    std::cout << "server version: " << clientInfo.serverVersion << std::endl;
    EXPECT_NE(clientInfo.serverVersion, "");
    EXPECT_NE(clientInfo.version, "");
}