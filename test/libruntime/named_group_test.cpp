/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.
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
#include "src/libruntime/groupmanager/named_group.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/libruntime/objectstore/memory_store.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {

class NamedGroupTest : public testing::Test {
public:
    NamedGroupTest(){};
    ~NamedGroupTest(){};
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
        YR::Libruntime::GroupOpts libOpts;
        libOpts.groupName = "gName";
        libOpts.timeout = -1;
        libOpts.bind.resource = "NUMA";
        libOpts.bind.strategy = "SPREAD";

        this->memoryStore = std::make_shared<MemoryStore>();
        this->invokeOrderMgr = std::make_shared<InvokeOrderManager>();
        auto dsObjectStore = std::make_shared<DSCacheObjectStore>();
        dsObjectStore->Init("127.0.0.1", 8080);
        waitManager = std::make_shared<WaitingObjectManager>();
        this->memoryStore->Init(dsObjectStore, waitManager);
        auto mockIntf = std::make_shared<MockFsIntf>();
        this->fsClient = std::make_shared<FSClient>(mockIntf);
        this->namedGroup = std::make_shared<NamedGroup>("groupName", "tenantId", libOpts, this->fsClient,
                                                        this->waitManager, this->memoryStore, this->invokeOrderMgr);
    }

    void TearDown() override
    {
        CloseGlobalTimer();
        this->waitManager.reset();
        this->memoryStore.reset();
        this->fsClient.reset();
        this->invokeOrderMgr.reset();
        this->namedGroup.reset();
    }
    std::shared_ptr<FSClient> fsClient;
    std::shared_ptr<MemoryStore> memoryStore;
    std::shared_ptr<NamedGroup> namedGroup;
    std::shared_ptr<WaitingObjectManager> waitManager;
    std::shared_ptr<InvokeOrderManager> invokeOrderMgr;
};

TEST_F(NamedGroupTest, BuildCreateReqsTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    CreateRequest createReq;
    spec->requestCreate = createReq;
    namedGroup->createSpecs.push_back(spec);
    auto createReqs = namedGroup->BuildCreateReqs();
    ASSERT_EQ(createReqs.requests_size(), 1);
    ASSERT_EQ(createReqs.groupopt().timeout(), -1);
    GroupOptions *options = createReqs.mutable_groupopt();
    ASSERT_EQ(options->bind().resource(), "NUMA");
    ASSERT_EQ(options->bind().policy(), ::common::BindStrategy::BIND_Spread);

    namedGroup->opts.bind.resource = "aaa";
    namedGroup->opts.bind.strategy = "bbb";
    auto invalidCreateReqs = namedGroup->BuildCreateReqs();
    GroupOptions *invalidOptions = invalidCreateReqs.mutable_groupopt();
    ASSERT_EQ(invalidOptions->bind().resource(), "NONE");
    ASSERT_EQ(invalidOptions->bind().policy(), ::common::BindStrategy::BIND_None);
}
}  // namespace test
}  // namespace YR
