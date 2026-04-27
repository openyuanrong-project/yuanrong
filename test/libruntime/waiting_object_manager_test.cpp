/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
#include <iostream>
#include <thread>
#include "mock/mock_datasystem.h"
#define private public
#include "src/libruntime/waiting_object_manager.h"
#include "src/utility/logger/logger.h"
#include "src/dto/constant.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
using std::chrono::high_resolution_clock;
using std::chrono::milliseconds;
class WaitingObjectManagerTest : public testing::Test {
public:
    WaitingObjectManagerTest(){};
    ~WaitingObjectManagerTest(){};
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

TEST_F(WaitingObjectManagerTest, SetUnreadyTest)
{
    auto wom = std::make_shared<WaitingObjectManager>();
    auto memStore = std::make_shared<MemoryStore>();
    auto objStore = std::make_shared<MockObjectStore>();
    memStore->Init(objStore, wom);
    wom->SetMemoryStore(memStore);
    bool ok = memStore->AddReturnObject("mock-objid-1");
    ASSERT_EQ(ok, true);

    bool idReady = wom->CheckReady("mock-objid-1");
    ASSERT_EQ(idReady, false);
    ASSERT_EQ(wom->unreadyObjectMap.size(), 1);
    idReady = wom->CheckReady("mock-objid-2");
    ASSERT_EQ(idReady, true);
    std::thread t([memStore]() {
        std::this_thread::sleep_for(std::chrono::seconds(1));
        memStore->SetReady("mock-objid-1");
    });
    if (t.joinable()) {
        auto waitResult = wom->WaitUntilReady({"mock-objid-1"}, 1, -1);
        t.join();
        ASSERT_EQ(waitResult->readyIds.size(), 1);
        ASSERT_EQ(waitResult->readyIds[0], "mock-objid-1");
    }

    // Wait multi object
    memStore->AddReturnObject("mock-objid-2");
    memStore->AddReturnObject("mock-objid-3");
    memStore->AddReturnObject("mock-objid-4");
    std::thread t2([memStore]() {
        memStore->SetReady("mock-objid-2");
        std::this_thread::sleep_for(std::chrono::seconds(1));
        memStore->SetReady("mock-objid-3");
    });
    if (t2.joinable()) {
        auto waitResult = wom->WaitUntilReady({"mock-objid-2", "mock-objid-3", "mock-objid-4"}, 2, -1);
        t2.join();
        ASSERT_EQ(waitResult->readyIds.size(), 2);
        ASSERT_EQ(waitResult->unreadyIds.size(), 1);
    }

    // Mock an exception for an object
    memStore->AddReturnObject("mock-objid-5");
    ErrorInfo err(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, "fake error message");
    std::thread t3([memStore, &err]() {
        memStore->SetReady("mock-objid-4");
        std::this_thread::sleep_for(std::chrono::seconds(1));
        memStore->SetError("mock-objid-5", err);
    });
    if (t3.joinable()) {
        auto waitResult = wom->WaitUntilReady({"mock-objid-4", "mock-objid-5"}, 2, 10 * MILLISECOND_UNIT);
        t3.join();
        ASSERT_EQ(waitResult->readyIds.size(), 1);
        ASSERT_EQ(waitResult->unreadyIds.size(), 0);
        ASSERT_EQ(waitResult->exceptionIds["mock-objid-5"], err);
    }
}

TEST_F(WaitingObjectManagerTest, WaitUntilReadyWaitTimeTest) 
{
    auto wom = std::make_shared<WaitingObjectManager>();
    auto memStore = std::make_shared<MemoryStore>();
    auto objStore = std::make_shared<MockObjectStore>();
    memStore->Init(objStore, wom);
    wom->SetMemoryStore(memStore);
    bool ok = memStore->AddReturnObject("mock-objid-1");
    ASSERT_EQ(ok, true);
    high_resolution_clock::time_point beginTime = high_resolution_clock::now();
    auto waitResult = wom->WaitUntilReady({"mock-objid-1"}, 1, 1 * MILLISECOND_UNIT);
    high_resolution_clock::time_point endTime = high_resolution_clock::now();
    ASSERT_EQ(waitResult->unreadyIds.size(), 1);
    ASSERT_GE(std::chrono::duration_cast<milliseconds>(endTime - beginTime).count(), 1000);
}

TEST_F(WaitingObjectManagerTest, WaitReturnWhenSetErrorBeforeWait)
{
    auto wom = std::make_shared<WaitingObjectManager>();
    auto memStore = std::make_shared<MemoryStore>();
    auto objStore = std::make_shared<MockObjectStore>();
    memStore->Init(objStore, wom);
    wom->SetMemoryStore(memStore);
    bool ok = memStore->AddReturnObject("mock-objid-1");
    ASSERT_EQ(ok, true);
    bool ok2 = memStore->AddReturnObject("mock-objid-2");
    ASSERT_EQ(ok2, true);

    ErrorInfo err(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, "fake error message");
    bool ok3 = memStore->SetError("mock-objid-1", err);
    ASSERT_EQ(ok3, true);
    bool ok4 = memStore->SetReady("mock-objid-2");
    ASSERT_EQ(ok4, true);
    auto waitResult = wom->WaitUntilReady({"mock-objid-1", "mock-objid-2"}, 2, -1);
    for(auto i:waitResult->exceptionIds){
        std::cout << i.first << std::endl;
    }
    ASSERT_TRUE(waitResult->exceptionIds.find("mock-objid-1") != waitResult->exceptionIds.end());
    ASSERT_EQ(waitResult->readyIds[0], "mock-objid-2");
}
}  // namespace test
}  // namespace YR