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

#include <filesystem>
#include <fstream>
#include <iostream>
#include <thread>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#define private public
#include "src/dto/config.h"
#include "src/libruntime/auto_init.h"
#include "src/utility/logger/logger.h"
using namespace YR::utility;
const std::string masterInfoString =
    "master_ip:127.0.0.1,etcd_ip:127.0.0.1,local_ip:127.0.0.1,host_ip:127.0.0.1,etcd_port:11393,global_scheduler_port:14210,ds_"
    "master_port:11090,etcd_"
    "peer_port:15580,bus-proxy:30495,bus:34834,ds-worker:31499,";

class AutoInitTest : public testing::Test {
public:
    AutoInitTest(){};
    ~AutoInitTest(){};
    void SetUp() override
    {
        unsetenv("YR_SERVER_ADDRESS");
        unsetenv("YR_DS_ADDRESS");
        unsetenv("YR_MASTER_INFO");
        unsetenv("YR_MASTER_INFO_PATH");
        YR::Libruntime::Config::c = YR::Libruntime::Config();
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
        InitLog(g_logParam);
    }
    void TearDown() override {}
};

static void MakeMasterInfoFile(const std::string &filepath, const std::string &content = masterInfoString)
{
    std::remove(filepath.c_str());

    std::filesystem::path dirPath = std::filesystem::path(filepath).parent_path();
    if (!dirPath.empty()) {
        if (!std::filesystem::exists(dirPath)) {
            std::filesystem::create_directories(dirPath);
        }
    }

    std::ofstream ofs(filepath);
    ofs << content << std::endl;
    ofs.close();
}

static void RemoveMasterInfoFile(const std::string &filepath)
{
    std::remove(filepath.c_str());
}

TEST_F(AutoInitTest, AutoCreateYuanRongClusterFailed)
{
    RemoveMasterInfoFile(YR::Libruntime::Config::Instance().YR_MASTER_INFO_PATH());
    YR::Libruntime::ClusterAccessInfo info;
    auto info2 = YR::Libruntime::AutoGetClusterAccessInfo(info);

    ASSERT_EQ(info2.serverAddr, "");
    ASSERT_EQ(info2.dsAddr, "");
    ASSERT_EQ(info2.inCluster, false);
}

TEST_F(AutoInitTest, AutoInitWithClusterAccessInfo)
{
    MakeMasterInfoFile(YR::Libruntime::Config::Instance().YR_MASTER_INFO_PATH(), masterInfoString);

    YR::Libruntime::ClusterAccessInfo info;
    auto info2 = YR::Libruntime::AutoGetClusterAccessInfo(info);

    ASSERT_EQ(info2.serverAddr, "127.0.0.1:34834");
    ASSERT_EQ(info2.dsAddr, "127.0.0.1:31499");
    ASSERT_EQ(info2.inCluster, true);
}

TEST_F(AutoInitTest, ParseFromMasterInfo)
{
    MakeMasterInfoFile(YR::Libruntime::Config::Instance().YR_MASTER_INFO_PATH(), masterInfoString);
    YR::Libruntime::ClusterAccessInfo info;
    info.ParseFromMasterInfo();
    ASSERT_EQ(info.serverAddr, "127.0.0.1:34834");
    ASSERT_EQ(info.dsAddr, "127.0.0.1:31499");
    ASSERT_EQ(info.inCluster, true);
}