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
#include <boost/asio/ssl.hpp>
#include <boost/beast/http.hpp>
#include "mock/mock_datasystem_client.h"
#include "mock/mock_fs_intf.h"
#include "mock/mock_fs_intf_with_callback.h"
#include "src/libruntime/invoke_spec.h"
#define private public

#include "src/libruntime/err_type.h"
#define protected public
#include "src/libruntime/groupmanager/function_group.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/libruntime/objectstore/memory_store.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {

namespace Libruntime {
bool ParseRequest(const CallRequest &request, std::vector<std::shared_ptr<DataObject>> &rawArgs,
                  libruntime::MetaData &metaData, std::shared_ptr<MemoryStore> memStore, bool isPosix);
}
namespace test {

class FunctionGroupTest : public testing::Test {
public:
    FunctionGroupTest(){};
    ~FunctionGroupTest(){};
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
        InitGlobalTimer();
        InvokeOptions opts;
        opts.groupName = "groupName";
        FunctionGroupOptions opt;
        opt.functionGroupSize = 8;
        opt.bundleSize = 2;
        this->memoryStore = std::make_shared<MemoryStore>();
        this->invokeOrderMgr = std::make_shared<InvokeOrderManager>();
        auto dsObjectStore = std::make_shared<DSCacheObjectStore>();
        dsObjectStore->Init("127.0.0.1", 8080);
        waitManager = std::make_shared<WaitingObjectManager>();
        this->memoryStore->Init(dsObjectStore, waitManager);
        gwClient_ = std::make_shared<MockGwClient>();
        this->fsClient = std::make_shared<FSClient>(gwClient_);
        this->fnGroup = std::make_shared<FunctionGroup>("groupName", "tenantId", opt, this->fsClient, this->waitManager,
                                                        this->memoryStore, this->invokeOrderMgr, nullptr, nullptr);
        spec = std::make_shared<InvokeSpec>();
    }

    void TearDown() override
    {
        CloseGlobalTimer();
        this->waitManager.reset();
        this->memoryStore.reset();
        this->fsClient.reset();
        this->invokeOrderMgr.reset();
        this->spec.reset();
        this->fnGroup.reset();
    }
    std::shared_ptr<FSClient> fsClient;
    std::shared_ptr<MemoryStore> memoryStore;
    std::shared_ptr<FunctionGroup> fnGroup;
    std::shared_ptr<WaitingObjectManager> waitManager;
    std::shared_ptr<InvokeOrderManager> invokeOrderMgr;
    std::shared_ptr<InvokeSpec> spec;
    std::shared_ptr<MockGwClient> gwClient_;
};

TEST_F(FunctionGroupTest, CreateRespHandlerTest)
{
    auto groupOpts = fnGroup->GetFunctionGroupOptions();
    ASSERT_EQ(groupOpts.functionGroupSize, 8);
    CreateResponses resps;
    resps.set_groupid("groupid");
    spec->returnIds = {DataObject{"objectId"}};
    std::vector<std::shared_ptr<InvokeSpec>> specs = {spec};
    fnGroup->SetCreateSpecs(specs);
    fnGroup->SetInvokeSpec(spec);
    memoryStore->AddReturnObject("objectId");
    memoryStore->SetInstanceIds("objectId", std::vector<std::string>{"instanceId"});
    resps.set_code(::common::ErrorCode::ERR_PARAM_INVALID);
    fnGroup->CreateRespHandler(resps);
    ASSERT_TRUE(fnGroup->invokeSpec_ != nullptr);
}

TEST_F(FunctionGroupTest, TerminateTest)
{
    std::vector<std::shared_ptr<InvokeSpec>> specs = {spec};
    fnGroup->SetCreateSpecs(specs);
    fnGroup->instanceIds_ = {"123", "456"};
    ASSERT_NO_THROW(fnGroup->SetTerminateError());
}

TEST_F(FunctionGroupTest, HandleReturnObjectLoopTest)
{
    auto handle = AccelerateMsgQueueHandle();
    handle.rank = 0;
    handle.maxChunks = 10;
    handle.maxChunkBytes = 10 * 1024 * 1024;
    handle.worldSize = 1;
    auto buffer = std::make_shared<NativeBuffer>(handle.maxChunks * handle.maxChunkBytes);
    auto queue = std::make_shared<AccelerateMsgQueue>(handle, buffer);
    fnGroup->queues_.emplace_back(queue);
    std::vector<std::shared_ptr<InvokeSpec>> specs = {spec};
    fnGroup->SetCreateSpecs(specs);
    fnGroup->instanceIds_ = {"123", "456"};
    auto t = std::thread([&]() {
        auto promise = std::make_shared<std::promise<void>>();
        auto future = promise->get_future();
        future.wait_for(std::chrono::milliseconds(1));
        fnGroup->SetTerminateError();
    });
    ASSERT_NO_THROW(fnGroup->HandleReturnObjectLoop());
    t.join();
}

TEST_F(FunctionGroupTest, AccelerateMsgQueueHandleTest)
{
    auto handle = AccelerateMsgQueueHandle();
    auto data = handle.ToJson();
    auto newHandle = AccelerateMsgQueueHandle::FromJson(data);
    ASSERT_TRUE(newHandle.name == handle.name);
}

TEST_F(FunctionGroupTest, AccelerateMsgQueueTest)
{
    auto handle = AccelerateMsgQueueHandle();
    handle.rank = 0;
    handle.maxChunks = 10;
    handle.maxChunkBytes = 10 * 1024 * 1024;
    handle.worldSize = 1;
    auto buffer = std::make_shared<NativeBuffer>(handle.maxChunks * handle.maxChunkBytes);
    auto queue = std::make_shared<AccelerateMsgQueue>(handle, buffer);
    auto shmBuffer = std::make_shared<ShmRingBuffer>(handle.worldSize, handle.maxChunks, handle.maxChunkBytes, buffer);
    ASSERT_TRUE(shmBuffer->GetMetadata(0) != nullptr);
    ASSERT_TRUE(shmBuffer->GetData(0) != nullptr);
    queue->SetReadFlag();
}

TEST_F(FunctionGroupTest, AccelerateMsgQueueDequeueTest)
{
    auto handle = AccelerateMsgQueueHandle();
    handle.rank = 0;
    handle.maxChunks = 10;
    handle.maxChunkBytes = 10 * 1024 * 1024;
    handle.worldSize = 1;
    auto buffer = std::make_shared<NativeBuffer>(handle.maxChunks * handle.maxChunkBytes);
    auto queue = std::make_shared<AccelerateMsgQueue>(handle, buffer);
    queue->Stop();
    ASSERT_TRUE(queue->Dequeue() == nullptr);
}

TEST_F(FunctionGroupTest, AccelerateTest)
{
    fnGroup->AddInstance({std::string("insId")});
    auto fsIntf = std::make_shared<MockFsIntf>();
    fnGroup->fsClient = std::make_shared<FSClient>(fsIntf);
    YR::Libruntime::AccelerateMsgQueueHandle h;
    YR::Libruntime::HandleReturnObjectCallback cb;
    fnGroup->Stop();
    auto err = fnGroup->Accelerate(h, cb);
    fnGroup->Stop();
    ASSERT_EQ(err.OK(), true);
}
}  // namespace test
}  // namespace YR
