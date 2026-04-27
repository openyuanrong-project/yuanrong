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

#include "src/utility/file_watcher.h"
#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <filesystem>
#include "json.hpp"
namespace YR {
namespace test {
using namespace YR::utility;
class FileWatcherTest : public ::testing::Test {
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

    void TearDown() override {}
};

bool CreateFile(const std::string &path)
{
    std::ofstream ofs(path, std::ios::out);
    if (!ofs.is_open()) {
        std::cerr << "Failed to create file: " << path << std::endl;
        return false;
    }
    ofs.close();
    return true;
}

bool DeleteFile(const std::string &path)
{
    std::error_code ec;  // 避免抛异常
    bool removed = std::filesystem::remove(path, ec);
    if (!removed || ec) {
        std::cerr << "Failed to delete file: " << path << " , error: " << ec.message() << std::endl;
        return false;
    }
    return true;
}

bool WriteStringToFile(const std::string &path, const std::string &content)
{
    std::ofstream ofs(path, std::ios::out);
    if (!ofs.is_open()) {
        std::cerr << "Failed to open file: " << path << std::endl;
        return false;
    }
    ofs << content;
    ofs.close();
    return true;
}

TEST_F(FileWatcherTest, FileWatcher)
{
    bool isDowngradeEnabled = false;
    std::string fileName = "/tmp/log/a.json";
    FileWatcher f(fileName, [&isDowngradeEnabled](const std::string &fileName) {
        std::ifstream file(fileName);
        nlohmann::json j;
        try {
            file >> j;
            isDowngradeEnabled = j.value("downgrade", false);
        } catch (const nlohmann::json::parse_error &e) {
            std::stringstream buffer;
            buffer << file.rdbuf();
            YRLOG_WARN("{} json parse error: {}", buffer.str(), e.what());
            isDowngradeEnabled = false;
        }
    });
    f.Start();
    CreateFile(fileName);
    WriteStringToFile(fileName, "{\"downgrade\": true}");
    while (true) {
        if (isDowngradeEnabled) {
            break;
        }
        std::this_thread::yield();
    }
    ASSERT_TRUE(isDowngradeEnabled);
    WriteStringToFile(fileName, "{\"downgrade\": false}");
    while (true) {
        if (!isDowngradeEnabled) {
            break;
        }
        std::this_thread::yield();
    }
    DeleteFile(fileName);
    ASSERT_FALSE(isDowngradeEnabled);
    f.Stop();
}
}  // namespace test
}  // namespace YR