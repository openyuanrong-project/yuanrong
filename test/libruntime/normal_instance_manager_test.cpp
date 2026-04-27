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
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"

#define protected public
#include "mock/mock_fs_intf_with_callback.h"
#include "src/libruntime/invokeadaptor/normal_instance_manager.h"

using namespace YR::Libruntime;
using namespace YR::utility;
using YR::utility::CloseGlobalTimer;
using YR::utility::InitGlobalTimer;
namespace YR {
namespace test {
std::shared_ptr<RequestResourceInfo> BuildRequestResourceInfo(const RequestResource &resource)
{
    std::unordered_map<std::string, std::shared_ptr<InstanceInfo>> instanceInfos;
    instanceInfos["insId"] = std::make_shared<InstanceInfo>();
    instanceInfos["insId"]->instanceId = "insId";
    instanceInfos["insId"]->leaseId = "leaseId";
    instanceInfos["insId"]->idleTime = 0;
    instanceInfos["insId"]->unfinishReqNum = 0;
    instanceInfos["insId"]->available = true;
    auto requestResourceInfo = std::make_shared<RequestResourceInfo>();
    requestResourceInfo->instanceInfos = instanceInfos;
    requestResourceInfo->avaliableInstanceInfos = instanceInfos;
    return requestResourceInfo;
}
class NormalInsManagerTest : public testing::Test {
public:
    NormalInsManagerTest(){};
    ~NormalInsManagerTest(){};
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
        std::function<void(const RequestResource &resource, const ErrorInfo &err, bool isRemainIns)> cb =
            [](const RequestResource &resource, const ErrorInfo &err, bool isRemainIns) {};
        auto reqMgr = std::make_shared<RequestManager>();
        auto librtCfg = std::make_shared<LibruntimeConfig>();
        mockFsIntf = std::make_shared<MockFsIntf>();
        auto fsClient = std::make_shared<FSClient>(mockFsIntf);
        std::shared_ptr<MemoryStore> memoryStore = std::make_shared<MemoryStore>();
        auto dsObjectStore = std::make_shared<DSCacheObjectStore>();
        dsObjectStore->Init("127.0.0.1", 8080);
        auto wom = std::make_shared<WaitingObjectManager>();
        memoryStore->Init(dsObjectStore, wom);
        auto deleteInsCb = [](const std::string &) {};
        insManager = std::make_shared<NormalInsManager>(cb, fsClient, memoryStore, reqMgr, librtCfg);
        insManager->SetDeleleInsCallback(deleteInsCb);
        spec = std::make_shared<InvokeSpec>();
        spec->jobId = "jobId";
        spec->requestId = "requestId";
        spec->traceId = "traceId";
        spec->instanceId = "instanceId";
        spec->invokeLeaseId = "leaseId";
        spec->invokeInstanceId = "insId";
        spec->functionMeta = {
            "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    }
    void TearDown() override
    {
        insManager->Stop();
        spec.reset();
        mockFsIntf.reset();
        insManager.reset();
        CloseGlobalTimer();
    }

    std::shared_ptr<NormalInsManager> insManager;
    std::shared_ptr<InvokeSpec> spec;
    std::shared_ptr<FSIntf> mockFsIntf;
};

TEST_F(NormalInsManagerTest, ScaleUpSuccessTest)
{
    auto queue = std::make_shared<PriorityQueue>();
    auto resource = GetRequestResource(spec);
    insManager->ScaleUp(spec, queue->Size());
    // requestResourceInfoMap 测试用例里加锁
    ASSERT_EQ(insManager->requestResourceInfoMap.find(resource) != insManager->requestResourceInfoMap.end(), true);

    queue->Push(spec);
    insManager->ScaleUp(spec, queue->Size());
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 1);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 1);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->callbackFuture.get();
    std::this_thread::sleep_for(std::chrono::milliseconds(2100));
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos.size(), 0);
}

TEST_F(NormalInsManagerTest, ScaleUpFailTest)
{
    auto queue = std::make_shared<PriorityQueue>();
    queue->Push(spec);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = false;
    auto resource = GetRequestResource(spec);
    insManager->ScaleUp(spec, queue->Size());
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->createFailInstanceNum, 0);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->callbackFuture.get();
}

TEST_F(NormalInsManagerTest, ScaleDownInsIsAbnormalTest)
{
    auto resource = GetRequestResource(spec);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = false;
    auto reqInsInfo = std::make_shared<RequestResourceInfo>();
    reqInsInfo->instanceInfos["insId"] = std::make_shared<InstanceInfo>();
    reqInsInfo->instanceInfos["insId"]->instanceId = "insId";
    reqInsInfo->instanceInfos["insId"]->leaseId = "leaseId";
    reqInsInfo->instanceInfos["insId"]->idleTime = 0;
    reqInsInfo->instanceInfos["insId"]->unfinishReqNum = 0;
    reqInsInfo->instanceInfos["insId"]->available = true;
    insManager->requestResourceInfoMap[resource] = reqInsInfo;
    insManager->ScaleDown(spec, false);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos.size(), 0);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->killCallbackFuture.get();
}

TEST_F(NormalInsManagerTest, ScaleDownInsIsnormalTest)
{
    auto resource = GetRequestResource(spec);
    auto reqInsInfo = std::make_shared<RequestResourceInfo>();
    reqInsInfo->instanceInfos["insId"] = std::make_shared<InstanceInfo>();
    reqInsInfo->instanceInfos["insId"]->instanceId = "insId";
    reqInsInfo->instanceInfos["insId"]->leaseId = "leaseId";
    reqInsInfo->instanceInfos["insId"]->idleTime = 0;
    reqInsInfo->instanceInfos["insId"]->unfinishReqNum = 0;
    reqInsInfo->instanceInfos["insId"]->available = true;
    insManager->requestResourceInfoMap[resource] = reqInsInfo;
    insManager->ScaleDown(spec, true);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->killCallbackFuture.get();
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos.size(), 0);
}

TEST_F(NormalInsManagerTest, HandleCreateResponseTest)
{
    CreateResponse resp;
    resp.set_code(common::ERR_RESOURCE_NOT_ENOUGH);
    resp.set_instanceid("instanceId");
    spec->instanceId = "";
    auto insInfo = std::make_shared<CreatingInsInfo>("", 0);
    insManager->HandleCreateResponse(spec, resp, insInfo);
    ASSERT_EQ(spec->instanceId, "instanceId");
}

TEST_F(NormalInsManagerTest, ScaleCancelAll)
{
    auto queue = std::make_shared<PriorityQueue>();
    auto resource = GetRequestResource(spec);
    insManager->ScaleUp(spec, queue->Size());
    ASSERT_EQ(insManager->requestResourceInfoMap.find(resource) != insManager->requestResourceInfoMap.end(), true);

    queue->Push(spec);
    insManager->ScaleUp(spec, queue->Size());
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 1);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 1);

    insManager->ScaleCancel(resource, queue->Size());
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 1);

    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->callbackFuture.get();

    auto newQueue = std::make_shared<PriorityQueue>();
    insManager->ScaleCancel(resource, newQueue->Size(), true);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->killCallbackFuture.get();

    ASSERT_TRUE(insManager->requestResourceInfoMap.find(resource) == insManager->requestResourceInfoMap.end());
}

TEST_F(NormalInsManagerTest, When_StartNormalInsScaleDownTimer_Twice_Should_Be_Ok)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    auto resource = GetRequestResource(spec);
    auto info = BuildRequestResourceInfo(resource);
    insManager->requestResourceInfoMap[resource] = info;
    insManager->totalCreatedInstanceNum_ = 1;
    insManager->StartNormalInsScaleDownTimer(resource, "insId");
    insManager->StartNormalInsScaleDownTimer(resource, "insId");
    insManager->DelInsInfo("insId", resource);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos.find("insId") ==
                  insManager->requestResourceInfoMap[resource]->instanceInfos.end(),
              true);
}
}  // namespace test
}  // namespace YR