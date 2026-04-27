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

#include <chrono>
#include <thread>

#include "mock/mock_datasystem.h"

#define private public

#include "src/libruntime/driverlog/driverlog_receiver.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace test {
using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;
using namespace std::chrono_literals;
class DriverLogTest : public testing::Test {
public:
    std::stringstream buffer;
    std::streambuf* coutbuf;
    void SetUp() override
    {
        Mkdir("/tmp/log");
        LogParam g_logParam = {
            .logLevel = "DEBUG",
            .logDir = "/tmp/log",
            .nodeName = "test-driver-runtime",
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
        coutbuf = std::cout.rdbuf();
        std::cout.rdbuf(buffer.rdbuf());
    }

    void TearDown() override
    {
        std::cout.rdbuf(coutbuf);
    }
};

TEST_F(DriverLogTest, DriverLogTestReceiver)
{
    std::shared_ptr<MockStreamStore> streamStore = std::make_shared<MockStreamStore>();
    auto c = std::make_shared<MockStreamConsumer>();
    std::string jobId = "job-8d638c95";
    EXPECT_CALL(*streamStore, CreateStreamConsumer("/log/runtime/std/job-8d638c95", _, _, true))
        .WillOnce([&c](const std::string &streamName, const SubscriptionConfig &config,
                       std::shared_ptr<StreamConsumer> &consumer, bool autoAck) {
            consumer = c;
            return ErrorInfo();
        });

    std::string logInfo = "this is driver log";
    EXPECT_CALL(*c, Receive(_, _))
        .WillOnce([&logInfo](uint32_t timeoutMs, std::vector<Element> &outElements) {
            outElements.resize(1);
            outElements[0].ptr = reinterpret_cast<uint8_t*>(const_cast<char*>(logInfo.data()));
            outElements[0].size = logInfo.size();
            return ErrorInfo();
        })
        .WillRepeatedly([](uint32_t timeoutMs, std::vector<Element> &outElements) {
            std::this_thread::sleep_for(500ms);
            return ErrorInfo();
        });

    {
        auto r = std::make_shared<DriverLogReceiver>();
        r->Init(streamStore, jobId, false);
        std::this_thread::sleep_for(50ms);
        size_t found = buffer.str().find(logInfo);
        ASSERT_FALSE(found == std::string::npos);
    }
}

TEST_F(DriverLogTest, DriverLogTestDedup)
{
    std::shared_ptr<MockStreamStore> streamStore = std::make_shared<MockStreamStore>();
    auto c = std::make_shared<MockStreamConsumer>();
    std::string jobId = "job-8d638c94";
    EXPECT_CALL(*streamStore, CreateStreamConsumer("/log/runtime/std/job-8d638c94", _, _, true))
        .WillOnce([&c](const std::string &streamName, const SubscriptionConfig &config,
                       std::shared_ptr<StreamConsumer> &consumer, bool autoAck) {
            consumer = c;
            return ErrorInfo();
        });


    std::vector<std::string> logInfos;
    for (auto i = 0; i < 5; i++) {
        auto msg = "(runtime-" + std::to_string(i) + ") " + "this is driver log";
        logInfos.push_back(msg);
    }

    EXPECT_CALL(*c, Receive(_, _))
        .WillOnce([&logInfos](uint32_t timeoutMs, std::vector<Element> &outElements) {
            outElements.resize(5);
            for (auto i = 0; i < 5; i++) {
                outElements[i].ptr = reinterpret_cast<uint8_t*>(const_cast<char*>(logInfos[i].data()));
                outElements[i].size = logInfos[i].size();
            }
            return ErrorInfo();
        })
        .WillRepeatedly([](uint32_t timeoutMs, std::vector<Element> &outElements) {
            std::this_thread::sleep_for(500ms);
            return ErrorInfo();
        });

    {
        auto r = std::make_shared<DriverLogReceiver>();
        r->Init(streamStore, jobId, true);
        std::this_thread::sleep_for(50ms);
        r.reset();
        size_t found = buffer.str().find("this is driver log");
        ASSERT_FALSE(found == std::string::npos);
        auto dedueInfo = "across cluster";
        YRLOG_ERROR("receive info is : {}", buffer.str());
        std::cout << buffer.str() << std::endl;
        found = buffer.str().find(dedueInfo);

        ASSERT_FALSE(found == std::string::npos);
    }
}

TEST_F(DriverLogTest, TestParse)
{
    std::string largeContent(40000, 'x');
    auto r = std::make_shared<DriverLogReceiver>();
    struct Case {
        std::string input;
        std::string rtId;
        std::string logContent;
    } tests[] = {{"(rtid1)content1", "rtid1", "content1"},
                 {"(rt(i)d)content1", "rt(i)d", "content1"},
                 {"(rtId)" + largeContent, "rtId", largeContent}};

    for (auto &t : tests) {
        auto [id, cont] = r->ParseLine(t.input);
        ASSERT_EQ(id, t.rtId);
        ASSERT_EQ(cont, t.logContent);
    }
}
}
}