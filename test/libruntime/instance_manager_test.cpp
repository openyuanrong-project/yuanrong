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
#include <unistd.h>
#include "mock/mock_fs_intf.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"

#define protected public
#include "src/libruntime/invokeadaptor/normal_instance_manager.h"

using namespace YR::Libruntime;
using namespace YR::utility;
using YR::utility::CloseGlobalTimer;
using YR::utility::InitGlobalTimer;
namespace YR {
namespace test {

class NormalInstanceManagerTest : public testing::Test {
public:
    NormalInstanceManagerTest(){};
    ~NormalInstanceManagerTest(){};
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
        std::function<void(const RequestResource &resource, const ErrorInfo &err, bool isRemainIns)> cb;
        cb = {};
        auto reqMgr = std::make_shared<RequestManager>();
        auto librtCfg = std::make_shared<LibruntimeConfig>();
        auto mockFsIntf = std::make_unique<MockGwClient>();
        auto fsClient = std::make_shared<FSClient>(std::move(mockFsIntf));
        std::shared_ptr<MemoryStore> memoryStore = std::make_shared<MemoryStore>();
        auto dsObjectStore = std::make_shared<DSCacheObjectStore>();
        dsObjectStore->Init("127.0.0.1", 8080);
        auto wom = std::make_shared<WaitingObjectManager>();
        memoryStore->Init(dsObjectStore, wom);
        auto deleteInsCb = [](const std::string &) {};
        insManager = std::make_shared<NormalInsManager>(cb, fsClient, memoryStore, reqMgr, librtCfg);
        insManager->SetDeleleInsCallback(deleteInsCb);
    }
    void TearDown() override
    {
        CloseGlobalTimer();
        insManager.reset();
    }

protected:
    std::shared_ptr<InsManager> insManager;
};

TEST_F(NormalInstanceManagerTest, ScheduleInsTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->jobId = "job-7c8e6fab";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    auto resource = GetRequestResource(spec);
    auto summary = insManager->GetAvailableIns(resource);
    ASSERT_EQ(summary.instanceId.empty(), true);

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
    insManager->requestResourceInfoMap[resource] = requestResourceInfo;
    summary = insManager->GetAvailableIns(resource);
    ASSERT_EQ(summary.instanceId.empty(), false);
    insManager->Stop();
}

TEST_F(NormalInstanceManagerTest, NeedCreateNewIns)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->jobId = "job-7c8e6fab";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    auto resource = GetRequestResource(spec);
    auto [needCreate, size] = insManager->NeedCreateNewIns(resource, 1);
    ASSERT_EQ(needCreate, false);

    auto requestResourceInfo = std::make_shared<RequestResourceInfo>();
    insManager->requestResourceInfoMap[resource] = requestResourceInfo;
    auto [needCreate1, size1] = insManager->NeedCreateNewIns(resource, 1);
    ASSERT_EQ(needCreate1, true);
    requestResourceInfo->createFailInstanceNum = 1;
    auto [needCreate2, size2] = insManager->NeedCreateNewIns(resource, 1);
    ASSERT_EQ(needCreate2, true);
    insManager->AddCreatingInsInfo(resource, std::make_shared<CreatingInsInfo>("", 0));
    auto [needCreate3, size3] = insManager->NeedCreateNewIns(resource, 1);
    ASSERT_EQ(needCreate3, false);
}

TEST_F(NormalInstanceManagerTest, NeedCancelCreatingInsTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->jobId = "job-7c8e6fab";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    auto resource = GetRequestResource(spec);

    auto [needCancel, cancelIns] = insManager->NeedCancelCreatingIns(resource, 0, false);
    ASSERT_FALSE(needCancel);

    auto requestResourceInfo = std::make_shared<RequestResourceInfo>();
    insManager->requestResourceInfoMap[resource] = requestResourceInfo;
    requestResourceInfo->createTime = 1;

    auto [needCancel1, cancelIns1] = insManager->NeedCancelCreatingIns(resource, 0, false);
    ASSERT_FALSE(needCancel1);

    for (auto i = 0; i < 10; ++i) {
        auto insInfo = std::make_shared<CreatingInsInfo>("", 0);
        insManager->AddCreatingInsInfo(resource, insInfo);
        insInfo->instanceId = "instance" + std::to_string(i);
    }
    ASSERT_EQ(requestResourceInfo->creatingIns.size(), 10);
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 10);

    auto [needCancel2, cancelIns2] = insManager->NeedCancelCreatingIns(resource, 0, true);
    ASSERT_TRUE(needCancel2);
    ASSERT_EQ(cancelIns2.size(), 10);
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 0);

    for (auto i = 0; i < 10; ++i) {
        auto insInfo = std::make_shared<CreatingInsInfo>("", 0);
        insManager->AddCreatingInsInfo(resource, insInfo);
    }
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 10);
    auto [needCancel3, cancelIns3] = insManager->NeedCancelCreatingIns(resource, 0, true);

    ASSERT_EQ(requestResourceInfo->creatingIns.size(), 10);
    ASSERT_EQ(cancelIns3.size(), 0);
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 10);

    for (auto i = 0; i < 10; ++i) {
        requestResourceInfo->creatingIns[i]->instanceId = "instance" + std::to_string(i);
        i = i + 1;
    }
    auto [needCancel4, cancelIns4] = insManager->NeedCancelCreatingIns(resource, 0, true);
    ASSERT_EQ(requestResourceInfo->creatingIns.size(), 5);
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 5);
    ASSERT_TRUE(needCancel4);
    ASSERT_EQ(cancelIns4.size(), 5);
}

TEST_F(NormalInstanceManagerTest, ChangeInsNum)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    auto resource = GetRequestResource(spec);
    auto requestResourceInfo = std::make_shared<RequestResourceInfo>();
    insManager->requestResourceInfoMap[resource] = requestResourceInfo;
    insManager->AddCreatingInsInfo(resource, std::make_shared<CreatingInsInfo>("instance1", 0));
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 1);
    insManager->EraseCreatingInsInfo(resource, "instance1");
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 0);

    insManager->EraseCreatingInsInfo(resource, "");
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 0);

    insManager->ChangeCreateFailNum(resource, true);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->createFailInstanceNum, 1);
    insManager->ChangeCreateFailNum(resource, false);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->createFailInstanceNum, 0);
}

TEST_F(NormalInstanceManagerTest, AddCreatingInsInfo)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    auto resource = GetRequestResource(spec);
    std::unordered_map<std::string, std::shared_ptr<InstanceInfo>> instanceInfos;
    instanceInfos["insId"] = std::make_shared<InstanceInfo>();
    instanceInfos["insId"]->instanceId = "insId";
    instanceInfos["insId"]->leaseId = "leaseId";
    instanceInfos["insId"]->idleTime = 0;
    instanceInfos["insId"]->unfinishReqNum = 0;
    instanceInfos["insId"]->available = true;
    auto requestResourceInfo = std::make_shared<RequestResourceInfo>();
    requestResourceInfo->instanceInfos = instanceInfos;
    insManager->requestResourceInfoMap[resource] = requestResourceInfo;

    for (auto i = 0; i < 10; ++i) {
        auto insInfo = std::make_shared<CreatingInsInfo>("", 0);
        insManager->AddCreatingInsInfo(resource, insInfo);
        insInfo->instanceId = "instance" + std::to_string(i);
    }
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 10);
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 10);

    for (auto i = 0; i < 10; ++i) {
        auto insInfo = std::make_shared<CreatingInsInfo>("", 0);
        insManager->AddCreatingInsInfo(resource, insInfo);
        insInfo->instanceId = "instance-" + std::to_string(i);
    }
    ASSERT_EQ(requestResourceInfo->creatingIns.size(), 20);
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 20);

    insManager->EraseCreatingInsInfo(resource, "instance0");
    ASSERT_EQ(requestResourceInfo->creatingIns.size(), 19);
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 19);

    insManager->EraseCreatingInsInfo(resource, "instance-0");
    ASSERT_EQ(requestResourceInfo->creatingIns.size(), 18);
    ASSERT_EQ(insManager->totalCreatingInstanceNum_, 18);

    ASSERT_TRUE(insManager->requestResourceInfoMap[resource]->instanceInfos.find("instance0") ==
                insManager->requestResourceInfoMap[resource]->instanceInfos.end());
}

TEST_F(NormalInstanceManagerTest, EraseCreatingFailNumWhenCreatingInsNumIsZero)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    auto resource = GetRequestResource(spec);
    auto requestResourceInfo = std::make_shared<RequestResourceInfo>();
    insManager->requestResourceInfoMap[resource] = requestResourceInfo;
    insManager->AddCreatingInsInfo(resource, std::make_shared<CreatingInsInfo>("instance0", 0));
    insManager->ChangeCreateFailNum(resource, true);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 1);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->createFailInstanceNum, 1);
    insManager->EraseCreatingInsInfo(resource, "instance0", false);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 0);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->createFailInstanceNum, 0);
}

TEST_F(NormalInstanceManagerTest, DelInsInfo)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    auto resource = GetRequestResource(spec);
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
    insManager->requestResourceInfoMap[resource] = requestResourceInfo;
    insManager->totalCreatedInstanceNum_ = 1;
    insManager->DelInsInfo("insId", resource);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos.find("insId") ==
                  insManager->requestResourceInfoMap[resource]->instanceInfos.end(),
              true);
    ASSERT_EQ(insManager->totalCreatedInstanceNum_, 0);
    ASSERT_EQ(0, insManager->requestResourceInfoMap[resource]->instanceInfos.size());
    ASSERT_EQ(0, insManager->requestResourceInfoMap[resource]->avaliableInstanceInfos.size());
}

TEST_F(NormalInstanceManagerTest, DecreaseUnfinishReqNum)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    spec->invokeInstanceId = "insId";
    auto resource = GetRequestResource(spec);
    std::unordered_map<std::string, std::shared_ptr<InstanceInfo>> instanceInfos;
    instanceInfos["insId"] = std::make_shared<InstanceInfo>();
    instanceInfos["insId"]->instanceId = "insId";
    instanceInfos["insId"]->leaseId = "leaseId";
    instanceInfos["insId"]->idleTime = 0;
    instanceInfos["insId"]->unfinishReqNum = 1;
    instanceInfos["insId"]->available = true;
    auto requestResourceInfo = std::make_shared<RequestResourceInfo>();
    requestResourceInfo->instanceInfos = instanceInfos;
    insManager->requestResourceInfoMap[resource] = requestResourceInfo;

    insManager->DecreaseUnfinishReqNum(spec, true);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos["insId"]->unfinishReqNum, 0);
}
}  // namespace test
}  // namespace YR