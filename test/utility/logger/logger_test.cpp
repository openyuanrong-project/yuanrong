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
#include <sys/stat.h>
#include <regex>

#include "src/utility/logger/fileutils.h"
#define private public
#include "src/utility/logger/log_handler.h"
#include "src/utility/logger/log_manager.h"
#include "src/utility/logger/logger.h"
#include "src/utility/logger/spd_logger.h"

namespace YR {
namespace test {
#define PATH_SIZE 255

using namespace YR::utility;

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

class Logger : public ::testing::Test {
    void SetUp() override
    {
        Mkdir("/tmp/log");
    }

    void TearDown() override
    {
        try {
            Rm("/tmp/log");
        } catch (const std::exception &e) {
            std::cout << "teardown exception: " << e.what() << std::endl;
        }
    }
};

TEST_F(Logger, LogFileCheck)
{
    InitLog(g_logParam);
    auto traceID = "traceID_123";
    YRLOG_DEBUG("{}|logger file debug, id:{}, name:{}", traceID, 123, "logger");
    YRLOG_INFO("{}|logger file info, id:{}, name:{}", traceID, 123, "logger");
    YRLOG_WARN("logger file warn");
    YRLOG_ERROR("{}|logger file error, id:{}, name:{}", traceID, 123, "logger");

    YRLOG_DEBUG("{}|logger debug, id:{}, name:{}", traceID, 131415, "logger");
    YRLOG_INFO("{}|logger info, id:{}, name:{}", traceID, 456, "logger");
    YRLOG_WARN("{}|logger warn, id:{}, name:{}", traceID, 789, "logger");
    YRLOG_ERROR("{}|logger error, id:{}, name:{}", traceID, 101112, "logger");

    YRLOG_DEBUG("test logger no args");
    YRLOG_INFO("test logger no args");
    YRLOG_WARN("test logger no args");
    YRLOG_ERROR("test logger no args");

    // log file is existed
    auto logFile = "/tmp/log/test-runtime-test.log";

    struct stat buffer;
    EXPECT_TRUE(stat(logFile, &buffer) == 0);
}

TEST_F(Logger, LogCountCheck)
{
    InitLog(g_logParam);
    if (SpdLogger::GetInstance().level() <= GetLogLevel("DEBUG")) {
        auto funcCnt1 = [](int i) { YRLOG_DEBUG_COUNT(1, "test log, id: {}, num: {}", 1, i); };
        auto funcCnt2 = [](int i) { YRLOG_DEBUG_COUNT(2, "test log, id: {}, num: {}", 2, i); };
        auto funcCnt5 = [](int i) { YRLOG_DEBUG_COUNT(5, "test log, id: {}, num: {}", 5, i); };

        for (uint32_t i = 0; i <= 5; ++i) {
            funcCnt1(i);
            funcCnt2(i);
            funcCnt5(i);
        }
    }
}

TEST_F(Logger, GetLogLevel)
{
    struct testArgs {
        std::string lvStr;
        yr_spdlog::level::level_enum spdLv;
    };
    std::vector<struct testArgs> ta = {
        {"DEBUG", yr_spdlog::level::debug}, {"INFO", yr_spdlog::level::info},      {"WARN", yr_spdlog::level::warn},
        {"ERR", yr_spdlog::level::err},     {"FATAL", yr_spdlog::level::critical},
    };

    for (auto &a : ta) {
        yr_spdlog::level::level_enum level = GetLogLevel(a.lvStr);
        EXPECT_EQ(level, a.spdLv);
    }
}

TEST_F(Logger, GetLogFile)
{
    struct testArgs {
        LogParam param;
        std::regex pattern;
    };
    LogParam lp1;
    lp1.logDir = "/var/paas/log";
    lp1.isLogMerge = true;
    LogParam lp2;
    lp2.logDir = "/var/paas/log";
    lp2.nodeName = YR::utility::IDGenerator::GenApplicationId();
    lp2.modelName = "driver";
    LogParam lp3;
    lp3.logDir = "/var/paas/log";
    lp3.nodeName = YR::utility::IDGenerator::GenApplicationId();
    lp3.modelName = "driver";
    lp3.logFileWithTime = true;
    std::regex pattern1("/var/paas/log/driver-[0-9]+\\.log");
    std::regex pattern2("/var/paas/log/job-[0-9a-f]{8}-[0-9a-z]+\\.log");
    std::regex pattern3("/var/paas/log/job-[0-9a-f]{8}-[0-9a-z]+-[0-9]{14}\\.log");
    std::vector<struct testArgs> ta = {
        {lp1, pattern1},
        {lp2, pattern2},
        {lp3, pattern3},
    };

    for (auto &a : ta) {
        auto logFile = SpdLogger::GetInstance().GetLogFile(a.param);
        EXPECT_TRUE(std::regex_match(logFile, a.pattern));
    }
}

TEST_F(Logger, CompressFileTest)
{
    std::string logDir = "/tmp";
    if (!ExistPath(logDir)) {
        Mkdir(logDir);
    }
    LogParam logParam;
    logParam.nodeName = "nodeName";
    logParam.modelName = "modelName";
    logParam.logFileWithTime = true;
    logParam.logDir = logDir;
    logParam.logLevel = "INFO";

    auto logpath = logDir + "/" + logParam.nodeName + "-" + logParam.modelName + ".0.log";
    std::ofstream outfile;
    outfile.open(logpath.c_str());
    outfile << "1";
    outfile.close();
    LogRollingCompress(logParam);

    auto filepath = logDir + "/" + logParam.nodeName + "-" + logParam.modelName + ".1.log.gz";
    outfile.open(filepath.c_str());
    outfile << "1";
    outfile.close();
    DoLogFileRolling(logParam);

    logParam.retentionDays = 0;
    DoLogFileRolling(logParam);

    Rm("nodeName-modelName.*");
}

TEST_F(Logger, FileutilsTest)
{
    std::string filepath_not_exist = "/tmp/filepath_not_exist";
    bool exist = FileExist(filepath_not_exist, 0);
    EXPECT_FALSE(exist);

    auto size = FileSize(filepath_not_exist);
    EXPECT_EQ(size, size_t(0));
    int64_t timestamp = 0;
    GetFileModifiedTime(filepath_not_exist, timestamp);
    EXPECT_EQ(timestamp, 0);

    std::vector<std::string> files;
    Glob(filepath_not_exist, files);

    int compress = CompressFile(filepath_not_exist, "dest");
    EXPECT_EQ(compress, -1);

    std::ofstream outfile;
    std::string filepath = "/tmp/temp.log";
    outfile.open(filepath.c_str());
    outfile << "1";
    outfile.close();

    compress = CompressFile(filepath, "/tmp/");
    EXPECT_EQ(compress, -1);

    DeleteFile("/tmp/");
    Rm(filepath);
}

TEST_F(Logger, FailureSignalHandlerTest)
{
    InitLog(g_logParam);
    FailureSignalWriter("xxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\nxxx\n");

    auto exceptionDir = "/tmp/log/exception";
    EXPECT_TRUE(ExistPath(exceptionDir));

    auto backtraceFilename = "/tmp/log/exception/BackTrace_test.log";
    struct stat buffer;
    EXPECT_TRUE(stat(backtraceFilename, &buffer) == 0);
}

TEST_F(Logger, LogManagerTest)
{
    LogManager logManager;
    LogParam logParam;
    EXPECT_TRUE(logManager.AddLogParam("", logParam));
    logParam.isLogMerge = true;
    EXPECT_TRUE(logManager.AddLogParam("", logParam));
}

TEST_F(Logger, LogPrefixTest)
{
    SpdLogger::GetInstance().AddLogPrefix("key", "value");
    std::string value;
    SpdLogger::GetInstance().GetLogPrefix("key", value);
    EXPECT_EQ(value, "value");
    SpdLogger::GetInstance().RemoveLogPrefix("key");
    std::string value2;
    SpdLogger::GetInstance().GetLogPrefix("key", value2);
    EXPECT_EQ(value2, "");
}

}  // namespace test
}  // namespace YR