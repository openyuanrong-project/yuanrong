/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

#include "src/libruntime/rgroupmanager/resource_group_create_spec.h"
#include "src/libruntime/rgroupmanager/resource_group_manager.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class ResourceGroupTest : public testing::Test {
public:
    ResourceGroupTest(){};
    ~ResourceGroupTest(){};
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
        rGroupManager_ = std::make_shared<ResourceGroupManager>();
    }
    void TearDown() override
    {
        rGroupManager_.reset();
    }

    std::shared_ptr<ResourceGroupManager> rGroupManager_;
};

TEST_F(ResourceGroupTest, ResourceGroupManagerTest)
{
    ErrorInfo err;
    auto requestId1 = YR::utility::IDGenerator::GenRequestId();
    auto requestId2 = YR::utility::IDGenerator::GenRequestId();
    auto isExist = rGroupManager_->IsRGDetailExist("rgName");
    ASSERT_TRUE(!isExist);
    auto bundleSize = rGroupManager_->GetRGroupBundleSize("rgName");
    ASSERT_TRUE(bundleSize == -1);
    rGroupManager_->SetRGCreateErrInfo("rgName", requestId1, err);
    auto err1 = rGroupManager_->GetRGCreateErrInfo("rgName", requestId1, 1);
    ASSERT_TRUE(err1.Msg() == "rgName: rgName does not exist in storeMap.");
    rGroupManager_->StoreRGDetail("rgName", requestId1, 10);
    isExist = rGroupManager_->IsRGDetailExist("rgName");
    ASSERT_TRUE(isExist);
    bundleSize = rGroupManager_->GetRGroupBundleSize("rgName");
    ASSERT_TRUE(bundleSize == 10);
    err1 = rGroupManager_->GetRGCreateErrInfo("rgName", requestId1, 0);
    ASSERT_TRUE(err1.Msg() == "get resource group create errorinfo timeout, failed rgName: rgName.");
    rGroupManager_->SetRGCreateErrInfo("rgName", requestId1, err);
    err1 = rGroupManager_->GetRGCreateErrInfo("rgName", requestId1, 0);
    ASSERT_TRUE(err1.Msg() == "");
    rGroupManager_->StoreRGDetail("rgName", requestId2, 10);
    rGroupManager_->SetRGCreateErrInfo("rgName", requestId2,
                                         ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::CORE, "msg"));
    err1 = rGroupManager_->GetRGCreateErrInfo("rgName", requestId2, 0);
    ASSERT_TRUE(err1.Msg() == "msg");
    rGroupManager_->RemoveRGDetail("rgName");
    isExist = rGroupManager_->IsRGDetailExist("rgName");
    ASSERT_TRUE(!isExist);
}

TEST_F(ResourceGroupTest, ResourceGroupTest)
{
    std::vector<std::unordered_map<std::string, double>> bundles = {
        {{"CPU", 500.0}, {"Memory", 200.0}}, {{"CPU", 300.0}}, {}};
    ResourceGroupSpec resourceGroupSpec;
    resourceGroupSpec.name = "rgName";
    resourceGroupSpec.bundles = bundles;
    auto spec =
        std::make_shared<ResourceGroupCreateSpec>(resourceGroupSpec, "requestId", "traceId", "jobId", "tenantId");
    spec->BuildCreateResourceGroupRequest();
    ASSERT_TRUE(spec->requestCreateRGroup.DebugString() != "");
}
}  // namespace test
}  // namespace YR
