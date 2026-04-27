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

#include <memory>
#include <thread>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "api/cpp/include/yr/api/exception.h"
#include "api/cpp/src/cluster_mode_runtime.h"
#include "api/cpp/src/config_manager.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/proto/libruntime.pb.h"
#include "src/utility/logger/logger.h"

using namespace testing;
namespace YR {
namespace test {
using namespace testing;
using namespace YR::utility;

class ConfigManagerTest : public testing::Test {
public:
    ConfigManagerTest() {};
    ~ConfigManagerTest() {};
    void SetUp() override
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
        InitLog(g_logParam);
    }
};

TEST_F(ConfigManagerTest, ConfigManagerInitFailTest)
{
    Config conf;
    conf.isDriver = true;
    conf.functionUrn = "abc123";

    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    try {
        ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    } catch (YR::Exception &e) {
        ASSERT_EQ(e.Code(), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);
    }
}

Config GetMockConf()
{
    Config conf;
    conf.isDriver = true;
    conf.mode = Config::Mode::CLUSTER_MODE;
    conf.functionUrn = "sn:cn:yrk:default:function:0-opc-opc:$latest";
    conf.javaFunctionUrn = "sn:cn:yrk:default:function:0-opc-opc:$latest";
    conf.pythonFunctionUrn = "sn:cn:yrk:default:function:0-opc-opc:$latest";
    conf.serverAddr = "127.0.0.1:1234";
    conf.threadPoolSize = 4;
    return conf;
}

TEST_F(ConfigManagerTest, ConfigManagerInitTest1)
{
    Config conf = GetMockConf();
    conf.dataSystemAddr = "127.0.0.1:1235";
    conf.maxTaskInstanceNum = 65537;

    int mockArgc = 5;
    char *mockArgv[] = {"--logDir=/tmp/log", "--logLevel=DEBUG", "--grpcAddress=127.0.0.1:1234", "--runtimeId=driver",
                        "jobId=job123"};

    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    EXPECT_EQ(ConfigManager::Singleton().IsLocalMode(), false) << "Test failed, it should be cluster mode";
}

TEST_F(ConfigManagerTest, ConfigManagerInitTest2)
{
    Config conf = GetMockConf();
    conf.maxTaskInstanceNum = 1;

    int mockArgc = 5;
    char *mockArgv[] = {"--logDir=/tmp/log", "--logLevel=DEBUG", "--grpcAddress=127.0.0.1:1234", "--runtimeId=driver",
                        "jobId=job123"};

    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
}

TEST_F(ConfigManagerTest, ConfigManagerInitTest3)
{
    Config conf = GetMockConf();
    conf.loadPaths = std::vector<std::string>(1025, std::string(1, 'a'));

    int mockArgc = 5;
    char *mockArgv[] = {"--logDir=/tmp/log", "--logLevel=DEBUG", "--grpcAddress=127.0.0.1:1234", "--runtimeId=driver",
                        "jobId=job123"};

    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
}

TEST_F(ConfigManagerTest, ConfigManagerInitTest4)
{
    Config conf = GetMockConf();
    conf.loadPaths = std::vector<std::string>(1, std::string(1, 'a'));
    conf.logDir = "/tmp/log";
    conf.logLevel = "DEBUG";
    int mockArgc = 5;
    char *mockArgv[] = {"--runtimeConfigPath=/home/snuser/config/runtime.json", "--logLevel=DEBUG",
                        "--grpcAddress=127.0.0.1:1234", "--runtimeId=driver", "jobId=job123"};

    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
}

TEST_F(ConfigManagerTest, ConfigManagerInitTest5)
{
    Config conf = GetMockConf();
    conf.loadPaths = std::vector<std::string>(1, std::string(1, 'a'));
    conf.maxConcurrencyCreateNum = -1;
    int mockArgc = 5;
    char *mockArgv[] = {"--logDir=/tmp/log", "--logLevel=DEBUG", "--grpcAddress=127.0.0.1:1234", "--runtimeId=driver",
                        "jobId=job123"};
    try {
        ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    } catch (YR::Exception &e) {
        ASSERT_EQ(e.Code(), YR::Libruntime::ErrorCode::ERR_INCORRECT_INIT_USAGE);
    }
}

TEST_F(ConfigManagerTest, ConfigManagerInitTest6)
{
    Config conf = GetMockConf();
    setenv("YR_LOG_COMPRESS", "false", 1);

    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    try {
        ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    } catch (const std::invalid_argument &e) {
        EXPECT_EQ(1, 0) << "Test failed";
    }
}

TEST_F(ConfigManagerTest, ConfigManagerInitTest7)
{
    Config conf;
    conf.mode = Config::Mode::LOCAL_MODE;
    conf.threadPoolSize = 65;
    setenv("YR_LOG_COMPRESS", "false", 1);

    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    int wantSize = static_cast<int>(std::thread::hardware_concurrency());
    try {
        ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    } catch (const std::invalid_argument &e) {
        EXPECT_EQ(1, 0) << "Test failed";
    }
    EXPECT_EQ(ConfigManager::Singleton().threadPoolSize, wantSize) << "Test failed";
}

TEST_F(ConfigManagerTest, GetValidLocalThreadPoolSizeTest)
{
    Config conf = GetMockConf();
    conf.localThreadPoolSize = 1000;
    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    int wantSize = static_cast<int>(std::thread::hardware_concurrency());
    try {
        ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    } catch (const std::invalid_argument &e) {
        EXPECT_EQ(1, 0) << "Test failed";
    }
    EXPECT_EQ(ConfigManager::Singleton().localThreadPoolSize, wantSize) << "Test failed";
}

TEST_F(ConfigManagerTest, ConfigManagerInitTest8)
{
    Config conf;
    conf.mode = Config::Mode::LOCAL_MODE;
    setenv("YR_LOG_COMPRESS", "false", 1);

    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    try {
        ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    } catch (const std::invalid_argument &e) {
        EXPECT_EQ(1, 0) << "Test failed";
    }
    EXPECT_EQ(ConfigManager::Singleton().GetClientInfo().serverVersion, "")
        << "Test failed, serverVersion should be empty";
    EXPECT_EQ(ConfigManager::Singleton().IsLocalMode(), true) << "Test failed, it should be local mode";
}

TEST_F(ConfigManagerTest, ConfigManagerInitTest9)
{
    Config conf = GetMockConf();
    conf.functionUrn = "";

    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
}

TEST_F(ConfigManagerTest, GetValidMaxLogFileNumTest)
{
    Config conf = GetMockConf();
    conf.maxLogFileNum = 0;

    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
}

TEST_F(ConfigManagerTest, GetValidMaxLogSizeMbTest)
{
    Config conf = GetMockConf();
    conf.maxLogSizeMb = 10;
    conf.maxLogFileNum = 10;
    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    ASSERT_EQ(conf.maxLogSizeMb, ConfigManager::Singleton().maxLogFileSize);
    ASSERT_EQ(conf.maxLogFileNum, ConfigManager::Singleton().maxLogFileNum);
}

TEST_F(ConfigManagerTest, ConfigManagerInitEnableMTLSTest)
{
    Config conf = GetMockConf();
    conf.enableMTLS = true;
    conf.privateKeyPath = "ddd/module.key";
    conf.certificateFilePath = "ddd/module.crt";
    conf.verifyFilePath = "ddd/ca.crt";
    std::strcpy(conf.privateKeyPaaswd, "paaswd");
    conf.encryptPrivateKeyPasswd = "abcd";
    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    ASSERT_EQ(conf.privateKeyPath, ConfigManager::Singleton().privateKeyPath);
    ASSERT_EQ(conf.certificateFilePath, ConfigManager::Singleton().certificateFilePath);
    ASSERT_EQ(conf.verifyFilePath, ConfigManager::Singleton().verifyFilePath);
    ASSERT_EQ(std::string(conf.privateKeyPaaswd), std::string(ConfigManager::Singleton().privateKeyPaaswd));
}

TEST_F(ConfigManagerTest, GetValidLogCompressTest)
{
    Config conf = GetMockConf();
    conf.logCompress = false;
    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
}

TEST_F(ConfigManagerTest, IsLowReliabilityTask)
{
    Config conf = GetMockConf();
    conf.isLowReliabilityTask = true;

    int mockArgc = 1;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    YR::Libruntime::LibruntimeConfig libConfig;
    libConfig.isLowReliabilityTask = ConfigManager::Singleton().isLowReliabilityTask;
    libruntime::MetaConfig metaConfig;
    libConfig.BuildMetaConfig(metaConfig);
    EXPECT_EQ(metaConfig.islowreliabilitytask(), true);
    libConfig.InitConfig(metaConfig);
    EXPECT_EQ(libConfig.isLowReliabilityTask, true);
}

TEST_F(ConfigManagerTest, ConfigManagerInitLogDirTest)
{
    Config conf = GetMockConf();

    int mockArgc = 0;
    char *mockArgv[] = {"--logDir=/tmp/log"};
    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    ASSERT_EQ(ConfigManager::Singleton().logDir, "./");
}
}  // namespace test
}  // namespace YR