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
#include <future>
#include <iostream>
#include <thread>

#include "mock/mock_datasystem.h"
#include "mock/mock_datasystem_client.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#define private public
#include "src/libruntime/objectstore/memory_store.h"
#include "src/libruntime/waiting_object_manager.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {

class MemoryStoreTest : public testing::Test {
public:
    MemoryStoreTest(){};
    ~MemoryStoreTest(){};
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
        memoryStore = std::make_shared<MemoryStore>();
        dsObjectStore = std::make_shared<MockObjectStore>();
        wom = std::make_shared<WaitingObjectManager>();
        memoryStore->Init(dsObjectStore, wom);
        wom->SetMemoryStore(memoryStore);
    }

    void TearDown() override
    {
        memoryStore.reset();
        wom.reset();
    }

    std::shared_ptr<MemoryStore> memoryStore;
    std::shared_ptr<WaitingObjectManager> wom;
    std::shared_ptr<MockObjectStore> dsObjectStore;
};

TEST_F(MemoryStoreTest, InitPutGetTest)
{
    EXPECT_CALL(*dsObjectStore.get(), IncreGlobalReference(_)).WillRepeatedly(Return(ErrorInfo()));
    EXPECT_CALL(*dsObjectStore.get(), DecreGlobalReference(_)).WillRepeatedly(Return(ErrorInfo()));
    EXPECT_CALL(*dsObjectStore.get(), QueryGlobalReference(_)).WillOnce(Return(std::vector<int>{1}));
    EXPECT_CALL(*dsObjectStore.get(), Put(_, _, _, _)).WillRepeatedly(Return(ErrorInfo()));
    auto singleRet = std::make_pair<ErrorInfo, std::shared_ptr<Buffer>>(ErrorInfo(), std::make_shared<NativeBuffer>(0));
    EXPECT_CALL(*dsObjectStore.get(), Get(Matcher<const std::string &>(_), _)).WillRepeatedly(Return(singleRet));
    std::string str = "Hello, world!";
    auto data = std::make_shared<YR::Libruntime::NativeBuffer>(str.size());
    data->MemoryCopy(str.data(), str.size());
    std::string id = "mock-objid-1";
    ErrorInfo errInfo = memoryStore->Put(data, id, {});
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_PARAM_INVALID);  // Didn't Incre before Put. Should be ERR_PARAM_INVALID.

    errInfo = memoryStore->IncreGlobalReference({id});  // ds
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_OK);

    errInfo = memoryStore->Put(data, id, {});  // ds
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_OK);

    auto [errInfo2, sbufferPtr] = memoryStore->Get(id, 1);  // ds fake Get
    ASSERT_EQ(errInfo2.Code(), ErrorCode::ERR_OK);

    errInfo = memoryStore->DecreGlobalReference({id});  // ds
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_OK);

    // === In Mem ===
    errInfo = memoryStore->Put(data, id, {}, false);
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_PARAM_INVALID);  // Decreased to 0 before Put. Should be ERR_PARAM_INVALID.

    errInfo = memoryStore->IncreGlobalReference({id}, false);  // mem
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_OK);

    std::vector<int> vec = memoryStore->QueryGlobalReference({id});  // mem
    ASSERT_EQ(vec[0], 1);

    errInfo = memoryStore->Put(data, id, {}, false);  // mem
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_OK);

    auto [errInfo3, sbufferPtr2] = memoryStore->Get(id, 1);  // mem Get
    ASSERT_EQ(errInfo3.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(sbufferPtr2.get(), data.get());

    std::vector<std::string> idVec = {id};
    memoryStore->AlsoPutToDS(idVec);

    errInfo = memoryStore->DecreGlobalReference({id});  // mem
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_OK);

    // === Not in storeMap ===
    auto [errInfo4, sbufferPtr3] = memoryStore->Get(id, -1);  // mem Get
    //  Decreased to 0, not in storeMap, should force Get from ds.
    ASSERT_EQ(errInfo4.Code(), ErrorCode::ERR_OK);

    auto [errInfo5, sbufferPtr4] = memoryStore->Get("mock-abc123", 1000);
    ASSERT_EQ(errInfo5.Code(), ErrorCode::ERR_OK);

    // Circular references
    errInfo = memoryStore->Put(data, id, {id}, false);
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_PARAM_INVALID);
}

bool is_timeout(std::future<bool> &f, int seconds)
{
    return f.wait_for(std::chrono::seconds(seconds)) == std::future_status::timeout;
}

TEST_F(MemoryStoreTest, ReadyTest)
{
    std::string str = "Hello, world!";
    auto data = std::make_shared<YR::Libruntime::NativeBuffer>(str.size());
    data->MemoryCopy(str.data(), str.size());
    std::string id = "mock-objid-1";

    bool ok = memoryStore->AddReturnObject(id);
    ASSERT_TRUE(ok);
    std::string mockInstanceId = "mock-instance-id-1";
    ok = memoryStore->SetInstanceId(id, mockInstanceId);
    ASSERT_TRUE(ok);
    auto instanceId = memoryStore->GetInstanceId(id);
    ASSERT_EQ(instanceId, mockInstanceId);

    bool triggeredExecption = false;
    ok = memoryStore->AddReadyCallback(id, [this, &triggeredExecption](const ErrorInfo &err) {
        if (!err.OK()) {
            triggeredExecption = true;
        }
        std::string id2 = "mock-objid-2";
        // Dead lock test.
        memoryStore->IncreGlobalReference({id2}, false);
        memoryStore->DecreGlobalReference({id2});
    });
    ASSERT_FALSE(triggeredExecption);
    ASSERT_TRUE(ok);
    ErrorInfo err = memoryStore->Put(data, id, {}, false);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    // if there is a Dead lock, it will time out.
    // same as memoryStore->SetReady(id); but with timeout check
    std::future<bool> f =
        std::async(std::launch::async, static_cast<bool (MemoryStore::*)(const std::string &)>(&MemoryStore::SetReady),
                   memoryStore, id);
    EXPECT_FALSE(is_timeout(f, 3)) << "function timed out after 3 seconds";
    ASSERT_FALSE(triggeredExecption);

    // Dead lock test: in AddReadyCallback, exec callback
    std::string retInstanceId;

    std::future<bool> f2 = std::async(
        std::launch::async,
        static_cast<bool (MemoryStore::*)(const std::string &, YR::Libruntime::ObjectReadyCallback)>(
            &MemoryStore::AddReadyCallback),
        memoryStore, id,
        [this, &id, &retInstanceId](const ErrorInfo &err) { retInstanceId = memoryStore->GetInstanceId(id); });
    EXPECT_FALSE(is_timeout(f2, 3)) << "function f2 timed out after 3 seconds";
    ASSERT_EQ(retInstanceId, mockInstanceId);
}

TEST_F(MemoryStoreTest, DuplicatedSetInsIdTest)
{
    ASSERT_NO_THROW({
        memoryStore->SetInstanceId("objid", "instanceId");
        memoryStore->SetInstanceId("objid", "instanceId");
    });
}

TEST_F(MemoryStoreTest, ExceptionTest)
{
    std::string id = "mock-objid-1";

    bool ok = memoryStore->AddReturnObject(id);
    ASSERT_TRUE(ok);
    std::string mockInstanceId = "mock-instance-id-1";
    ok = memoryStore->SetInstanceId(id, mockInstanceId);
    ASSERT_TRUE(ok);
    auto instanceId = memoryStore->GetInstanceId(id);
    ASSERT_EQ(instanceId, mockInstanceId);

    bool triggeredExecption = false;
    ok = memoryStore->AddReadyCallback(id, [&triggeredExecption](const ErrorInfo &err) {
        if (!err.OK()) {
            triggeredExecption = true;
        }
    });
    ASSERT_FALSE(triggeredExecption);
    ASSERT_TRUE(ok);
    ErrorInfo err(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, "mock error message");
    ok = memoryStore->SetError(id, err);
    ASSERT_TRUE(ok);
    ASSERT_TRUE(triggeredExecption);

    ErrorInfo err2 = memoryStore->GetLastError(id);
    ASSERT_EQ(err2, err);
}

TEST_F(MemoryStoreTest, RangeSetGetTest)
{
    msgpack::sbuffer buffer;
    msgpack::pack(buffer, "Hello, world!");
    std::shared_ptr<msgpack::sbuffer> data = std::make_shared<msgpack::sbuffer>(std::move(buffer));
    std::string id = "mock-objid-1";

    bool ok = memoryStore->AddReturnObject(id);
    ASSERT_TRUE(ok);
    std::vector<std::string> mockInstanceIds;
    mockInstanceIds.push_back("mock-instance-id-1");
    mockInstanceIds.push_back("mock-instance-id-2");
    mockInstanceIds.push_back("mock-instance-id-3");
    ok = memoryStore->SetInstanceIds(id, mockInstanceIds);
    ASSERT_TRUE(ok);
    auto [instanceIds, err] = memoryStore->GetInstanceIds(id, 2);
    ASSERT_EQ(err.OK(), true);
    ASSERT_EQ(instanceIds.size(), 3);

    std::string id2 = "mock-objid-2";
    auto res = memoryStore->GetInstanceIds(id2, 0);
    ASSERT_EQ(res.second.Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);
    ASSERT_EQ(res.second.Msg(), "objId " + id2 + " does not exist in storeMap.");
    ok = memoryStore->AddReturnObject(id2);
    ASSERT_TRUE(ok);
    res = memoryStore->GetInstanceIds(id2, 0);
    ASSERT_EQ(res.second.Code(), ErrorCode::ERR_GET_OPERATION_FAILED);
    ASSERT_EQ(res.second.Msg(), "get instances timeout, failed objectID: mock-objid-2.");
}

/**
 * Description:
 *     Test whether two locks in the code cause a deadlock.
 *     The mu lock should be released when 'SetUnready' is
 *     called in the 'AddReturnObject' function.
 * Steps:
 * 1. Use std::async(std::launch::async... to start two threads to execute 'AddReturnObject' and 'WaitUntilReady'
 *    respectively.
 * Expectation:
 *  Neither thread times out, and the function executed by the thread returns true.
 */
TEST_F(MemoryStoreTest, MuLockReleaseTest)
{
    // start two threads
    std::future<bool> async_task_set = std::async(std::launch::async, [this]() -> bool {
        for (int i = 0; i < 10; i++) {
            std::string objID = "mock-objid-" + std::to_string(i);
            memoryStore->AddReturnObject(objID);

            std::this_thread::sleep_for(std::chrono::milliseconds(50));
            memoryStore->SetReady(objID);
        }
        return true;
    });

    std::future<bool> async_task_wait = std::async(std::launch::async, [this]() -> bool {
        std::vector<std::string> objIDs;
        for (int i = 0; i < 10; i++) {
            std::string objID = "mock-objid-" + std::to_string(i);
            objIDs.emplace_back(objID);
            wom->WaitUntilReady(objIDs, objIDs.size(), -1);
        }
        return true;
    });
    EXPECT_FALSE(is_timeout(async_task_set, 3)) << "async_task_set timed out after 3 seconds";
    EXPECT_FALSE(is_timeout(async_task_wait, 3)) << "async_task_wait timed out after 3 seconds";

    ASSERT_TRUE(async_task_set.get());
    ASSERT_TRUE(async_task_wait.get());
}

TEST_F(MemoryStoreTest, TestAddReturnObjectWhenInputDuplicateReturnFalse)
{
    std::vector<DataObject> objIDs;
    for (int i = 0; i < 2; i++) {
        std::string objID = "mock-objid-" + std::to_string(i);
        objIDs.push_back(DataObject(objID));
    }
    ASSERT_TRUE(memoryStore->AddReturnObject(objIDs));
    ASSERT_FALSE(memoryStore->AddReturnObject(objIDs));
}

TEST_F(MemoryStoreTest, TestIncreaseObjRefAndRefCntRight)
{
    EXPECT_CALL(*dsObjectStore.get(), IncreGlobalReference(_)).WillRepeatedly(Return(ErrorInfo()));
    EXPECT_CALL(*dsObjectStore.get(), DecreGlobalReference(_)).WillRepeatedly(Return(ErrorInfo()));
    EXPECT_CALL(*dsObjectStore.get(), QueryGlobalReference(_)).WillOnce(Return(std::vector<int>{1}));
    std::string objID = "mock-objid-1";
    auto err = memoryStore->IncreaseObjRef({objID});
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID);
    ASSERT_TRUE(memoryStore->AddReturnObject(objID));
    auto refCnts = memoryStore->QueryGlobalReference({objID});
    ASSERT_EQ(refCnts[0], 1);
    auto err2 = memoryStore->IncreaseObjRef({objID});
    ASSERT_TRUE(err2.OK());
    ASSERT_TRUE(memoryStore->IsExistedInLocal(objID));
    memoryStore->DecreGlobalReference({objID});
    ASSERT_TRUE(memoryStore->IsExistedInLocal(objID));
    memoryStore->DecreGlobalReference({objID});
    ASSERT_FALSE(memoryStore->IsExistedInLocal(objID));
}

TEST_F(MemoryStoreTest, TestIncreGlobalReferenceAndRefCntRight)
{
    EXPECT_CALL(*dsObjectStore.get(), IncreGlobalReference(_)).WillRepeatedly(Return(ErrorInfo()));
    EXPECT_CALL(*dsObjectStore.get(), DecreGlobalReference(_)).WillRepeatedly(Return(ErrorInfo()));
    EXPECT_CALL(*dsObjectStore.get(), QueryGlobalReference(_)).WillRepeatedly(Return(std::vector<int>{1}));
    std::string objID = "mock-objid-1";
    auto err = memoryStore->IncreGlobalReference({objID}, false);
    ASSERT_TRUE(err.OK());
    auto refCnts = memoryStore->QueryGlobalReference({objID});
    ASSERT_EQ(refCnts[0], 1);
    auto ret = memoryStore->IncreGlobalReference({objID}, true);
    ASSERT_TRUE(ret.OK());
    // query from dsObjectStore
    refCnts = memoryStore->QueryGlobalReference({objID});
    ASSERT_EQ(refCnts[0], 1);
    memoryStore->DecreGlobalReference({objID});
    memoryStore->DecreGlobalReference({objID});
    ASSERT_FALSE(memoryStore->IsExistedInLocal(objID));
}

TEST_F(MemoryStoreTest, TestIncreGlobalReferenceWithRemoteIdAndRefCntRight)
{
    EXPECT_CALL(*dsObjectStore.get(), QueryGlobalReference(_)).WillRepeatedly(Return(std::vector<int>{1}));
    EXPECT_CALL(*dsObjectStore.get(), IncreGlobalReference(_)).WillRepeatedly(Return(ErrorInfo()));
    std::string objID = "mock-objid-1";
    auto err = memoryStore->IncreGlobalReference({objID}, std::string("aaa"));
    ASSERT_TRUE(err.first.OK());
    ASSERT_EQ(err.second.size(), 0);
    auto refCnts = memoryStore->QueryGlobalReference({objID});
    ASSERT_EQ(refCnts[0], 1);
    auto ret = memoryStore->IncreGlobalReference({objID}, std::string("aaa"));
    ASSERT_TRUE(err.first.OK());
    ASSERT_EQ(err.second.size(), 0);
    // query from dsObjectStore
    refCnts = memoryStore->QueryGlobalReference({objID});
    ASSERT_EQ(refCnts[0], 1);
    memoryStore->Clear();
    ASSERT_FALSE(memoryStore->IsExistedInLocal(objID));
}

TEST_F(MemoryStoreTest, TestAddGenerator)
{
    auto res = memoryStore->AddGenerator("generatorId");
    ASSERT_TRUE(res);
    res = memoryStore->AddGenerator("generatorId");
    ASSERT_TRUE(!res);
}

TEST_F(MemoryStoreTest, TestGeneratorFinished)
{
    memoryStore->AddGenerator("generatorId");
    std::thread t([this]() {
        std::this_thread::sleep_for(std::chrono::seconds(2));
        memoryStore->AddOutput("generatorId", "objectid", 0);
    });
    auto [err, res] = memoryStore->GetOutput("generatorId", true);
    t.join();
    ASSERT_TRUE(err.OK());
    ASSERT_EQ(res, "objectid");

    memoryStore->GeneratorFinished("generatorId");
    ASSERT_EQ(memoryStore->storeMap["generatorId"]->finished, true);
}

TEST_F(MemoryStoreTest, TestGetFailedWhenDSInitFailed)
{
    auto ret = std::make_pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>>(
        ErrorInfo(ErrorCode::ERR_DATASYSTEM_FAILED, ""), {});
    EXPECT_CALL(*dsObjectStore.get(), Get(Matcher<const std::vector<std::string> &>(_), _)).WillOnce(Return(ret));
    std::vector<std::string> ids = {"aaa"};
    auto [err, result] = memoryStore->Get(ids, 100);
    ASSERT_FALSE(err.OK());
}

TEST_F(MemoryStoreTest, TestGetBuffersFailedWhenDSInitFailed)
{
    auto ret = std::make_pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>>(
        ErrorInfo(ErrorCode::ERR_DATASYSTEM_FAILED, ""), {});
    EXPECT_CALL(*dsObjectStore.get(), GetBuffers(_, _)).WillOnce(Return(ret));
    auto [err, result] = memoryStore->GetBuffers({"aaa"}, 100);
    ASSERT_FALSE(err.OK());
    RetryInfo retryInfo;
    retryInfo.errorInfo.SetErrCodeAndMsg(ErrorCode::ERR_DATASYSTEM_FAILED, YR::Libruntime::ModuleCode::DATASYSTEM, "");
    auto ret2 = std::make_pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>>(std::move(retryInfo), {});
    EXPECT_CALL(*dsObjectStore.get(), GetBuffersWithoutRetry(_, _)).WillOnce(Return(ret2));
    auto [err2, result2] = memoryStore->GetBuffersWithoutRetry({"aaa"}, 100);
    ASSERT_FALSE(err.OK());
}

TEST_F(MemoryStoreTest, TestAddReadyCallbackWithData)
{
    auto singleRet = std::make_pair<ErrorInfo, std::shared_ptr<Buffer>>(ErrorInfo(), std::make_shared<NativeBuffer>(0));
    EXPECT_CALL(*dsObjectStore.get(), Get(Matcher<const std::string &>(_), _)).WillOnce(Return(singleRet));
    auto cb = [](const ErrorInfo &err, std::shared_ptr<Buffer> buf) {};
    bool res = memoryStore->AddReadyCallbackWithData("objID111", cb);
    ASSERT_FALSE(res);
    auto detail = std::make_shared<ObjectDetail>();
    detail->err = ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "");
    memoryStore->storeMap["objID111"] = detail;
    res = memoryStore->AddReadyCallbackWithData("objID111", cb);
    ASSERT_FALSE(res);
    detail->err = ErrorInfo();
    detail->ready = false;
    res = memoryStore->AddReadyCallbackWithData("objID111", cb);
    ASSERT_TRUE(res);
    detail->ready = true;
    res = memoryStore->AddReadyCallbackWithData("objID111", cb);
    ASSERT_FALSE(res);
    detail->storeInMemory = true;
    res = memoryStore->AddReadyCallbackWithData("objID111", cb);
    ASSERT_FALSE(res);
}

TEST_F(MemoryStoreTest, HandleInstanceRouteTest)
{
    std::string id = "mock-objid-route";
    std::string route = "route";
    bool ok = memoryStore->SetInstanceRoute(id, route);
    ASSERT_TRUE(!ok);
    std::string res = memoryStore->GetInstanceRoute(id);
    ASSERT_TRUE(res == "");
    ok = memoryStore->AddReturnObject(id);
    ASSERT_TRUE(ok);
    ok = memoryStore->SetInstanceRoute(id, route);
    ASSERT_TRUE(ok);
    ok = memoryStore->SetInstanceRoute(id, route);
    ASSERT_TRUE(ok);
    res = memoryStore->GetInstanceRoute(id);
    ASSERT_TRUE(res == route);
    res = memoryStore->GetInstanceRoute(id, -1);
    ASSERT_TRUE(res == route);
}

TEST_F(MemoryStoreTest, TestAddEventCallbackWithData)
{
    auto callback = [](const ErrorInfo &err, std::shared_ptr<Buffer> buf) {};
    memoryStore->AddEventCallbackWithData("testID", callback);
    ASSERT_TRUE(memoryStore->reqEventCallbackMap_.find("testID") != memoryStore->reqEventCallbackMap_.end());
}

TEST_F(MemoryStoreTest, TestAddEventData)
{
    auto callback = [](const ErrorInfo &err, std::shared_ptr<Buffer> buf) {};
    auto res = memoryStore->AddEventData("testID", "testData");
    ASSERT_FALSE(res);

    memoryStore->AddEventCallbackWithData("testID", callback);
    res = memoryStore->AddEventData("testID", "testData");
    ASSERT_TRUE(res);

    res = memoryStore->AddEventData("testID", YR::Libruntime::EVENT_EOF);
    ASSERT_TRUE(res);
    auto it = memoryStore->reqEventCallbackMap_.find("testID");
    ASSERT_TRUE(it == memoryStore->reqEventCallbackMap_.end());
}

TEST_F(MemoryStoreTest, AddEventTimer)
{
    memoryStore->AddEventTimer("testID1", 1);
    auto it = memoryStore->reqEventCallbackMap_.find("testID1");
    ASSERT_TRUE(it != memoryStore->reqEventCallbackMap_.end());

    memoryStore->AddEventTimer("testID", 1);
    it = memoryStore->reqEventCallbackMap_.find("testID");
    ASSERT_TRUE(it != memoryStore->reqEventCallbackMap_.end());
}

}  // namespace test
}  // namespace YR