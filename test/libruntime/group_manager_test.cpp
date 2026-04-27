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
#include "src/libruntime/groupmanager/function_group.h"
#include "src/libruntime/groupmanager/group_manager.h"
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

class GroupManagerTest : public testing::Test {
public:
    GroupManagerTest(){};
    ~GroupManagerTest(){};
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
        this->memoryStore = std::make_shared<MemoryStore>();
        this->invokeOrderMgr = std::make_shared<InvokeOrderManager>();
        auto dsObjectStore = std::make_shared<DSCacheObjectStore>();
        dsObjectStore->Init("127.0.0.1", 8080);
        waitManager = std::make_shared<WaitingObjectManager>();
        waitManager->SetMemoryStore(memoryStore);
        this->memoryStore->Init(dsObjectStore, waitManager);
        mockFsIntf = std::make_shared<MockFsIntf>();
        this->fsClient = std::make_shared<FSClient>(mockFsIntf);
        groupManager = std::make_shared<GroupManager>();
    }

    void TearDown() override
    {
        CloseGlobalTimer();
        this->waitManager.reset();
        this->memoryStore.reset();
        this->mockFsIntf.reset();
        this->fsClient.reset();
        this->groupManager.reset();
        this->invokeOrderMgr.reset();
    }
    std::shared_ptr<FSClient> fsClient;
    std::shared_ptr<FSIntf> mockFsIntf;
    std::shared_ptr<MemoryStore> memoryStore;
    std::shared_ptr<GroupManager> groupManager;
    std::shared_ptr<WaitingObjectManager> waitManager;
    std::shared_ptr<InvokeOrderManager> invokeOrderMgr;
};

TEST_F(GroupManagerTest, IsGroupExistTest)
{
    InvokeOptions opts;
    opts.groupName = "groupName";
    GroupOpts groupOpts{"groupName", -1};
    auto group = std::make_shared<NamedGroup>("groupName", "tenantId", groupOpts, this->fsClient, this->waitManager,
                                              this->memoryStore, this->invokeOrderMgr);
    auto isGroupExist1 = groupManager->IsGroupExist(opts.groupName);
    ASSERT_EQ(isGroupExist1, false);

    auto spec = std::make_shared<InvokeSpec>();
    CreateRequest req;
    spec->requestCreate = req;
    spec->opts = opts;
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->returnIds = returnObjs;
    groupManager->AddSpec(spec);
    groupManager->AddGroup(group);
    auto isGroupExist2 = groupManager->IsGroupExist(opts.groupName);
    ASSERT_EQ(isGroupExist2, true);

    auto insIsReady = groupManager->IsInsReady(opts.groupName);
    ASSERT_EQ(insIsReady, false);

    groupManager->GroupCreate(opts.groupName);
    groupManager->Terminate(opts.groupName);
    auto isGroupExist3 = groupManager->IsGroupExist(opts.groupName);
    ASSERT_EQ(isGroupExist3, false);
}

TEST_F(GroupManagerTest, CreateFunctionGroupSuccessTest)
{
    InvokeOptions opts;
    opts.groupName = "groupName";
    FunctionGroupOptions opt;
    opt.functionGroupSize = 8;
    opt.bundleSize = 2;
    auto group = std::make_shared<FunctionGroup>("groupName", "tenantId", opt, this->fsClient, this->waitManager,
                                                 this->memoryStore, this->invokeOrderMgr, nullptr, nullptr);
    auto spec = std::make_shared<InvokeSpec>();
    CreateRequest req;
    spec->requestCreate = req;
    spec->opts = opts;
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->returnIds = returnObjs;
    groupManager->AddSpec(spec);
    groupManager->AddGroup(group);
    groupManager->GroupCreate(opts.groupName);
    groupManager->Terminate(opts.groupName);
    auto isGroupExist = groupManager->IsGroupExist(opts.groupName);
    ASSERT_EQ(isGroupExist, false);
}

TEST_F(GroupManagerTest, CreateFunctionGroupFailedTest)
{
    InvokeOptions opts;
    opts.groupName = "groupName";
    FunctionGroupOptions opt;
    opt.functionGroupSize = 8;
    opt.bundleSize = 2;
    auto group = std::make_shared<FunctionGroup>("groupName", "tenantId", opt, this->fsClient, this->waitManager,
                                                 this->memoryStore, this->invokeOrderMgr, nullptr, nullptr);
    auto spec = std::make_shared<InvokeSpec>();
    CreateRequest req;
    spec->requestCreate = req;
    spec->opts = opts;
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->returnIds = returnObjs;
    groupManager->AddSpec(spec);
    groupManager->AddGroup(group);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = false;
    groupManager->GroupCreate(opts.groupName);
    groupManager->Terminate(opts.groupName);
    auto isGroupExist = groupManager->IsGroupExist(opts.groupName);
    ASSERT_EQ(isGroupExist, false);
}

TEST_F(GroupManagerTest, GroupTerminateTest)
{
    InvokeOptions opts;
    opts.groupName = "groupName";
    GroupOpts groupOpts{"groupName", -1};
    auto group = std::make_shared<NamedGroup>("groupName", "tenantId", groupOpts, this->fsClient, this->waitManager,
                                              this->memoryStore, this->invokeOrderMgr);
    auto spec = std::make_shared<InvokeSpec>();
    CreateRequest req;
    spec->requestCreate = req;
    spec->opts = opts;
    spec->returnIds = std::vector<DataObject>{DataObject("returnID")};
    groupManager->AddSpec(spec);
    groupManager->AddGroup(group);
    ASSERT_EQ(groupManager->groupSpecs_[opts.groupName].size(), 1);
    ASSERT_EQ(groupManager->groups_.size(), 1);
    groupManager->Terminate(opts.groupName);
    groupManager->AddSpec(spec);
    groupManager->AddGroup(group);
    ASSERT_EQ(groupManager->groupSpecs_[opts.groupName].size(), 1);
    ASSERT_EQ(groupManager->groups_.size(), 1);
}

TEST_F(GroupManagerTest, GroupWaitTest)
{
    InvokeOptions opts;
    opts.groupName = "groupName";
    GroupOpts groupOpts{"groupName", -1};
    auto group = std::make_shared<NamedGroup>("groupName", "tenantId", groupOpts, this->fsClient, this->waitManager,
                                              this->memoryStore, this->invokeOrderMgr);
    auto groupGet1 = groupManager->GetGroup(opts.groupName);
    ASSERT_EQ(groupGet1, nullptr);
    auto errorInfo1 = groupManager->Wait(opts.groupName);
    ASSERT_EQ(errorInfo1.Code(), ErrorCode::ERR_PARAM_INVALID);
    groupManager->AddGroup(group);
    auto groupGet2 = groupManager->GetGroup(opts.groupName);
    ASSERT_NE(groupGet2, nullptr);
    auto spec = std::make_shared<InvokeSpec>();
    spec->opts = opts;
    spec->returnIds = std::vector<DataObject>{DataObject("returnID")};
    std::vector<std::shared_ptr<InvokeSpec>> specs{spec};
    auto errorInfo2 = groupManager->Wait(opts.groupName);
    ASSERT_EQ(errorInfo2.Code(), common::ErrorCode::ERR_NONE);

    groupGet2->SetCreateSpecs(specs);
    auto errorInfo3 = groupManager->Wait(opts.groupName);
    ASSERT_EQ(errorInfo3.Code(), common::ErrorCode::ERR_NONE);
}
}  // namespace test
}  // namespace YR
