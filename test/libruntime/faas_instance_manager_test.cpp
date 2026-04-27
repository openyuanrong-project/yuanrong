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
#include "mock/mock_fs_intf.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"

#define protected public
#include "mock/mock_fs_intf_with_callback.h"
#include "src/libruntime/invokeadaptor/faas_instance_manager.h"

namespace YR {
namespace Libruntime {
std::pair<InstanceResponse, ErrorInfo> GetFaasInstanceRsp(const NotifyRequest &notifyReq);
std::pair<BatchInstanceResponse, ErrorInfo> GetFaasBatchInstanceRsp(const NotifyRequest &notifyReq);
std::vector<unsigned char> BuildReacquireInstanceData(const RequestResource &resource);
}
}  // namespace YR

using namespace YR::Libruntime;
using namespace YR::utility;
using YR::utility::CloseGlobalTimer;
using YR::utility::InitGlobalTimer;
namespace YR {
namespace test {

class FaasInstanceManagerTest : public testing::Test {
public:
    FaasInstanceManagerTest(){};
    ~FaasInstanceManagerTest(){};
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
        insManager = std::make_shared<FaasInsManager>(cb, fsClient, memoryStore, reqMgr, librtCfg);
        insManager->UpdateSchdulerInfo("scheduler1", "scheduler1", "ADD");
    }
    void TearDown() override
    {
        if (insManager->leaseTimer) {
            insManager->leaseTimer->cancel();
            insManager->leaseTimer.reset();
        }
        insManager.reset();
        mockFsIntf.reset();
        CloseGlobalTimer();
    }

protected:
    std::shared_ptr<FaasInsManager> insManager;
    std::shared_ptr<FSIntf> mockFsIntf;
};

TEST_F(FaasInstanceManagerTest, BuildAcquireRequestTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->returnIds = returnObjs;
    spec->jobId = "jobId";
    spec->requestId = "requestId";
    spec->traceId = "traceId";
    spec->instanceId = "instanceId";
    spec->invokeLeaseId = "leaseId";
    spec->invokeInstanceId = "insId";
    spec->functionMeta = {"",
                          "",
                          "funcname",
                          "classname",
                          libruntime::LanguageType::Cpp,
                          "",
                          "",
                          "poollabel",
                          libruntime::ApiType::Function};
    InvokeOptions opts;
    opts.schedulerInstanceIds.push_back("shcedulerInstanceId");
    std::unordered_map<std::string, std::string> invokelabels;
    invokelabels["xxx"] = "xxx";
    opts.invokeLabels = invokelabels;
    spec->opts = opts;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isAcquireResponse = false;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = false;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->needCheckArgs = true;
    auto [instanceAllocation, err] = insManager->AcquireInstance("", spec);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(FaasInstanceManagerTest, RecordRequestTest)
{
    auto reqInsInfo = std::make_shared<RequestResourceInfo>();
    reqInsInfo->instanceInfos["leaseId"] = std::make_shared<InstanceInfo>();
    reqInsInfo->instanceInfos["leaseId"]->instanceId = "insId";
    reqInsInfo->instanceInfos["leaseId"]->leaseId = "leaseId";
    reqInsInfo->instanceInfos["leaseId"]->idleTime = 0;
    reqInsInfo->instanceInfos["leaseId"]->unfinishReqNum = 0;
    reqInsInfo->instanceInfos["leaseId"]->available = true;
    reqInsInfo->instanceInfos["leaseId"]->traceId = "traceId";
    reqInsInfo->instanceInfos["leaseId"]->faasInfo = FaasAllocationInfo{"", "", 100};
    reqInsInfo->instanceInfos["leaseId"]->reporter = std::make_shared<ReportRecord>();
    reqInsInfo->instanceInfos["leaseId"]->stateId = "";
    reqInsInfo->instanceInfos["leaseId"]->claimTime = 10000LL;
    auto spec = std::make_shared<InvokeSpec>();
    spec->invokeLeaseId = "leaseId";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    auto resource = GetRequestResource(spec);
    insManager->requestResourceInfoMap[resource] = reqInsInfo;

    insManager->RecordRequest(resource, spec, true);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos["leaseId"]->reporter->GetTotalDuration() > 0,
              true);

    insManager->RecordRequest(resource, spec, false);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos["leaseId"]->reporter->IsAbnormal(), true);
}

TEST_F(FaasInstanceManagerTest, ScaleDownTest)
{
    auto reqInsInfo = std::make_shared<RequestResourceInfo>();
    reqInsInfo->instanceInfos["leaseId1"] = std::make_shared<InstanceInfo>();
    reqInsInfo->instanceInfos["leaseId1"]->instanceId = "insId";
    reqInsInfo->instanceInfos["leaseId1"]->leaseId = "leaseId1";
    reqInsInfo->instanceInfos["leaseId1"]->idleTime = 0;
    reqInsInfo->instanceInfos["leaseId1"]->unfinishReqNum = 0;
    reqInsInfo->instanceInfos["leaseId1"]->available = true;
    reqInsInfo->instanceInfos["leaseId1"]->traceId = "traceId";
    reqInsInfo->instanceInfos["leaseId1"]->faasInfo = FaasAllocationInfo{"", "", 100};
    reqInsInfo->instanceInfos["leaseId1"]->reporter = std::make_shared<ReportRecord>();
    reqInsInfo->instanceInfos["leaseId2"] = std::make_shared<InstanceInfo>();
    reqInsInfo->instanceInfos["leaseId2"]->instanceId = "insId";
    reqInsInfo->instanceInfos["leaseId2"]->leaseId = "leaseId2";
    reqInsInfo->instanceInfos["leaseId2"]->idleTime = 0;
    reqInsInfo->instanceInfos["leaseId2"]->unfinishReqNum = 0;
    reqInsInfo->instanceInfos["leaseId2"]->available = true;
    reqInsInfo->instanceInfos["leaseId2"]->traceId = "traceId";
    reqInsInfo->instanceInfos["leaseId2"]->faasInfo = FaasAllocationInfo{"", "", 100};
    reqInsInfo->instanceInfos["leaseId2"]->reporter = std::make_shared<ReportRecord>();
    auto spec = std::make_shared<InvokeSpec>();
    spec->invokeLeaseId = "leaseId1";
    spec->invokeInstanceId = "insId";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    auto resource = GetRequestResource(spec);
    insManager->requestResourceInfoMap[resource] = reqInsInfo;

    insManager->ScaleDown(spec, false);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos.find("leaseId1") ==
                  insManager->requestResourceInfoMap[resource]->instanceInfos.end(),
              true);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos.find("leaseId2") ==
                  insManager->requestResourceInfoMap[resource]->instanceInfos.end(),
              true);
}

TEST_F(FaasInstanceManagerTest, ScaleUpSuccessTest)
{
    auto queue = std::make_shared<PriorityQueue>();
    auto spec = std::make_shared<InvokeSpec>();
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->returnIds = returnObjs;
    spec->jobId = "jobId";
    spec->requestId = "requestId";
    spec->traceId = "traceId";
    spec->instanceId = "instanceId";
    spec->invokeLeaseId = "leaseId";
    spec->invokeInstanceId = "insId";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    InvokeOptions opts;
    opts.schedulerInstanceIds.push_back("shcedulerInstanceId");
    spec->opts = opts;
    insManager->ScaleUp(spec, queue->Size());
    auto resource = GetRequestResource(spec);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->creatingIns.size(), 0);

    queue->Push(spec);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isAcquireResponse = true;
    insManager->ScaleUp(spec, queue->Size());
    sleep(2);
    ASSERT_EQ(insManager->requestResourceInfoMap.find(resource), insManager->requestResourceInfoMap.end());
    ASSERT_EQ(insManager->totalCreatedInstanceNum_, 0);
}

TEST_F(FaasInstanceManagerTest, ScaleUpFailTest)
{
    auto queue = std::make_shared<PriorityQueue>();
    auto spec = std::make_shared<InvokeSpec>();
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->returnIds = returnObjs;
    spec->jobId = "jobId";
    spec->requestId = "requestId";
    spec->traceId = "traceId";
    spec->instanceId = "instanceId";
    spec->invokeLeaseId = "leaseId";
    spec->invokeInstanceId = "insId";
    spec->functionMeta = {"",
                          "",
                          "funcname",
                          "classname",
                          libruntime::LanguageType::Cpp,
                          "",
                          "",
                          "poollabel",
                          libruntime::ApiType::Function};
    InvokeOptions opts;
    opts.schedulerInstanceIds.push_back("shcedulerInstanceId");
    spec->opts = opts;
    queue->Push(spec);
    auto resource = GetRequestResource(spec);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = false;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isAcquireResponse = true;
    insManager->ScaleUp(spec, queue->Size());
    sleep(2);
    std::shared_ptr<RequestResourceInfo> info;
    {
        absl::ReaderMutexLock lock(&insManager->insMtx);
        ASSERT_EQ(insManager->requestResourceInfoMap.find(resource) == insManager->requestResourceInfoMap.end(), true);
    }
}

TEST_F(FaasInstanceManagerTest, StartBatchRenewTimer)
{
    auto spec = std::make_shared<InvokeSpec>();
    auto resource = GetRequestResource(spec);
    insManager->StartBatchRenewTimer();
    ASSERT_EQ(insManager->requestResourceInfoMap.find(resource) != insManager->requestResourceInfoMap.end(), false);

    auto reqInsInfo = std::make_shared<RequestResourceInfo>();
    insManager->tLeaseInterval = 1000;
    reqInsInfo->instanceInfos["leaseId"] = std::make_shared<InstanceInfo>();
    reqInsInfo->instanceInfos["leaseId"]->instanceId = "insId";
    reqInsInfo->instanceInfos["leaseId"]->leaseId = "leaseId";
    reqInsInfo->instanceInfos["leaseId"]->idleTime = 0;
    reqInsInfo->instanceInfos["leaseId"]->unfinishReqNum = 0;
    reqInsInfo->instanceInfos["leaseId"]->available = true;
    reqInsInfo->instanceInfos["leaseId"]->traceId = "traceId";
    reqInsInfo->instanceInfos["leaseId"]->faasInfo = FaasAllocationInfo{"", "", 100};
    reqInsInfo->instanceInfos["leaseId"]->reporter = std::make_shared<ReportRecord>();
    insManager->requestResourceInfoMap[resource] = reqInsInfo;
    insManager->globalLeases["leaseId"] = resource;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isAcquireResponse = true;
    insManager->StartBatchRenewTimer();
    ASSERT_EQ(insManager->leaseTimer != nullptr, true);
    if (insManager->leaseTimer) {
        insManager->leaseTimer->cancel();
    }
}

TEST_F(FaasInstanceManagerTest, BatchRenewHandlerReleased)
{
    auto reqInsInfo = std::make_shared<RequestResourceInfo>();
    insManager->tLeaseInterval = 10;
    reqInsInfo->instanceInfos["leaseId"] = std::make_shared<InstanceInfo>();
    reqInsInfo->instanceInfos["leaseId"]->instanceId = "insId";
    reqInsInfo->instanceInfos["leaseId"]->leaseId = "leaseId";
    reqInsInfo->instanceInfos["leaseId"]->idleTime = 0;
    reqInsInfo->instanceInfos["leaseId"]->unfinishReqNum = 0;
    reqInsInfo->instanceInfos["leaseId"]->available = true;
    reqInsInfo->instanceInfos["leaseId"]->traceId = "traceId";
    reqInsInfo->instanceInfos["leaseId"]->faasInfo = FaasAllocationInfo{"functionid", "sig111", 10000};
    reqInsInfo->instanceInfos["leaseId"]->reporter = std::make_shared<ReportRecord>();
    ASSERT_EQ(reqInsInfo->instanceInfos.size(), 1);

    auto spec = std::make_shared<InvokeSpec>();
    auto resource = GetRequestResource(spec);
    insManager->requestResourceInfoMap[resource] = reqInsInfo;
    insManager->globalLeases["leaseId"] = resource;

    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isBatchRenew = true;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = true;
    insManager->StartBatchRenewTimer();
    sleep(1);
    ASSERT_EQ(reqInsInfo->instanceInfos.size(), 0);
    if (insManager->leaseTimer) {
        insManager->leaseTimer->cancel();
    }
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = false;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isBatchRenew = false;
}

TEST_F(FaasInstanceManagerTest, BatchRenewHandlerRetained)
{
    auto reqInsInfo = std::make_shared<RequestResourceInfo>();
    insManager->tLeaseInterval = 10;
    int i;
    auto spec = std::make_shared<InvokeSpec>();
    auto resource = GetRequestResource(spec);
    for (i = 0; i < 1001; i++) {
        std::string leaseId = "leaseId" +  std::to_string(i);
        reqInsInfo->instanceInfos[leaseId] = std::make_shared<InstanceInfo>();
        reqInsInfo->instanceInfos[leaseId]->instanceId = "insId";
        reqInsInfo->instanceInfos[leaseId]->leaseId = leaseId;
        reqInsInfo->instanceInfos[leaseId]->idleTime = 0;
        reqInsInfo->instanceInfos[leaseId]->unfinishReqNum = 0;
        reqInsInfo->instanceInfos[leaseId]->available = true;
        reqInsInfo->instanceInfos[leaseId]->traceId = leaseId;
        reqInsInfo->instanceInfos[leaseId]->faasInfo = FaasAllocationInfo{"functionid", "sig111", 10000};
        reqInsInfo->instanceInfos[leaseId]->reporter = std::make_shared<ReportRecord>();
        insManager->globalLeases[leaseId] = resource;
    }
    ASSERT_EQ(reqInsInfo->instanceInfos.size(), 1001);

    insManager->requestResourceInfoMap[resource] = reqInsInfo;

    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isBatchRenew = true;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = true;
    insManager->StartBatchRenewTimer();
    sleep(1);
    ASSERT_EQ(reqInsInfo->instanceInfos.size(), 1001);
    if (insManager->leaseTimer) {
        insManager->leaseTimer->cancel();
    }
    for (i = 0; i < 1001; i++) {
        std::string leaseId = "leaseId" + std::to_string(i);
        reqInsInfo->instanceInfos.erase(leaseId);
        insManager->globalLeases.erase(leaseId);
    }
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = false;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isBatchRenew = false;
}

TEST_F(FaasInstanceManagerTest, BatchRenewHandlerFailed)
{
    auto reqInsInfo = std::make_shared<RequestResourceInfo>();
    insManager->tLeaseInterval = 10;
    reqInsInfo->instanceInfos["leaseId"] = std::make_shared<InstanceInfo>();
    reqInsInfo->instanceInfos["leaseId"]->instanceId = "insId";
    reqInsInfo->instanceInfos["leaseId"]->leaseId = "leaseId";
    reqInsInfo->instanceInfos["leaseId"]->idleTime = 0;
    reqInsInfo->instanceInfos["leaseId"]->unfinishReqNum = 0;
    reqInsInfo->instanceInfos["leaseId"]->available = true;
    reqInsInfo->instanceInfos["leaseId"]->traceId = "traceId";
    reqInsInfo->instanceInfos["leaseId"]->faasInfo = FaasAllocationInfo{"functionid", "sig111", 10000};
    reqInsInfo->instanceInfos["leaseId"]->reporter = std::make_shared<ReportRecord>();
    reqInsInfo->instanceInfos["leaseId2"] = std::make_shared<InstanceInfo>();
    reqInsInfo->instanceInfos["leaseId2"]->instanceId = "insId";
    reqInsInfo->instanceInfos["leaseId2"]->leaseId = "leaseId2";
    reqInsInfo->instanceInfos["leaseId2"]->idleTime = 0;
    reqInsInfo->instanceInfos["leaseId2"]->unfinishReqNum = 0;
    reqInsInfo->instanceInfos["leaseId2"]->available = true;
    reqInsInfo->instanceInfos["leaseId2"]->traceId = "traceId";
    reqInsInfo->instanceInfos["leaseId2"]->faasInfo = FaasAllocationInfo{"functionid", "sig111", 10000};
    reqInsInfo->instanceInfos["leaseId2"]->reporter = std::make_shared<ReportRecord>();
    ASSERT_EQ(reqInsInfo->instanceInfos.size(), 2);

    auto spec = std::make_shared<InvokeSpec>();
    auto resource = GetRequestResource(spec);
    insManager->requestResourceInfoMap[resource] = reqInsInfo;
    insManager->globalLeases["leaseId"] = resource;
    insManager->globalLeases["leaseId2"] = resource;

    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = false;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isBatchRenew = true;
    insManager->StartBatchRenewTimer();
    sleep(1);
    ASSERT_EQ(reqInsInfo->instanceInfos.size(), 0);
    if (insManager->leaseTimer) {
        insManager->leaseTimer->cancel();
    }
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isBatchRenew = false;
}

TEST_F(FaasInstanceManagerTest, BuildReacquireInstanceData)
{
    RequestResource r1;
    r1.opts.instanceSession = std::make_shared<YR::Libruntime::InstanceSession>();
    r1.opts.invokeLabels["label1"] = "1";
    r1.opts.podLabels["label1"] = "2";
    r1.opts.customResources["ccc"] = 1;
    std::vector<unsigned char> vec = BuildReacquireInstanceData(r1);
    ASSERT_EQ(vec.size(), 384);
}

TEST_F(FaasInstanceManagerTest, AcquireAndReleaseInstanceTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->returnIds = returnObjs;
    spec->jobId = "jobId";
    spec->requestId = "requestId";
    spec->traceId = "traceId";
    spec->instanceId = "instanceId";
    spec->invokeLeaseId = "leaseId";
    spec->invokeInstanceId = "insId";
    spec->functionMeta = {"",
                          "",
                          "funcname",
                          "classname",
                          libruntime::LanguageType::Cpp,
                          "",
                          "",
                          "poollabel",
                          libruntime::ApiType::Function};
    InvokeOptions opts;
    opts.schedulerInstanceIds.push_back("shcedulerInstanceId");
    spec->opts = opts;
    auto resource = GetRequestResource(spec);
    // std::string stateId = ;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isAcquireResponse = true;
    auto [instanceAllocation1, err1] = insManager->AcquireInstance("", spec);
    ASSERT_EQ(instanceAllocation1.leaseId, "leaseId");
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos.find("leaseId") !=
                  insManager->requestResourceInfoMap[resource]->instanceInfos.end(),
              true);
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isAcquireResponse = false;
    auto err = insManager->ReleaseInstance("leaseId", "", false, spec);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(insManager->requestResourceInfoMap[resource]->instanceInfos.find("leaseId") ==
                  insManager->requestResourceInfoMap[resource]->instanceInfos.end(),
              true);

    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isAcquireResponse = true;
    std::dynamic_pointer_cast<MockFsIntf>(mockFsIntf)->isReqNormal = false;
    auto [instanceAllocation3, err3] = insManager->AcquireInstance("", spec);
    ASSERT_EQ(err3.Code(), ErrorCode::ERR_PARAM_INVALID);
}

TEST_F(FaasInstanceManagerTest, GetFaasInstanceRspTest)
{
    NotifyRequest notifyReq;
    auto object = notifyReq.add_smallobjects();
    object->set_id("123");

    std::string json_str =
        R"({"funcKey":"","funcSig":"","instanceID":"","threadID":"","leaseInterval":0,"errorCode":150428,"errorMessage":"","schedulerTime":30})";
    std::string value = "0000000000000000" + json_str;
    object->set_value(value);
    auto [resp, errorInfo] = GetFaasInstanceRsp(notifyReq);
    ASSERT_TRUE(errorInfo.OK());
    ASSERT_EQ(resp.errorCode, 150428);
}

TEST_F(FaasInstanceManagerTest, AddInsInfoBareTest)
{
    auto info = std::make_shared<RequestResourceInfo>();
    auto faasInfo = std::make_shared<InstanceInfo>();
    faasInfo->leaseId = "leaseId";

    insManager->AddInsInfoBare(info, faasInfo);
    ASSERT_EQ(info->instanceInfos.size(), 1);
    ASSERT_EQ(info->avaliableInstanceInfos.size(), 1);
    ASSERT_EQ(insManager->totalCreatedInstanceNum_, 1);

    auto faasInfo1 = std::make_shared<InstanceInfo>();
    faasInfo1->leaseId = "leaseId";
    faasInfo1->instanceId = "instanceId";
    insManager->AddInsInfoBare(info, faasInfo1);
    ASSERT_EQ(info->instanceInfos.size(), 1);
    ASSERT_EQ(info->instanceInfos[faasInfo->leaseId]->instanceId, "instanceId");
    ASSERT_EQ(insManager->totalCreatedInstanceNum_, 1);
}

TEST_F(FaasInstanceManagerTest, GetFaasBatchInstanceRsp)
{
    NotifyRequest notifyReq;
    auto object = notifyReq.add_smallobjects();
    object->set_id("123");

    std::string json_str =
        R"({"instanceAllocSucceed":{"f1a00e58-f2a1-4000-8000-000000f8e9e3-thread26":{"funcKey":"default/0@functest@functest/latest","funcSig":"4243308021","instanceID":"f1a00e58-f2a1-4000-8000-000000f8e9e3","threadID":"f1a00e58-f2a1-4000-8000-000000f8e9e3-thread26","instanceIP":"127.0.0.1","instancePort":"22771","nodeIP":"","nodePort":"","leaseInterval":0,"cpu":600,"memory":512}},"instanceAllocFailed":{},"leaseInterval":1000,"schedulerTime":0.000118108})";
    std::string value = "0000000000000000" + json_str;
    object->set_value(value);
    auto [resp, errorInfo] = GetFaasBatchInstanceRsp(notifyReq);
    ASSERT_TRUE(errorInfo.OK());
    ASSERT_EQ(resp.instanceAllocSucceed.size(), 1);
}


TEST_F(FaasInstanceManagerTest, UpdateSpecSchedulerIdsTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    insManager->UpdateSpecSchedulerIds(spec, "schedulerId");
    ASSERT_TRUE(spec->schedulerInfos->schedulerInstanceList[0]->InstanceID == "schedulerId");
    ASSERT_TRUE(!spec->schedulerInfos->schedulerInstanceList[0]->isAvailable);
    auto updateTime = spec->schedulerInfos->schedulerInstanceList[0]->updateTime;
    insManager->UpdateSpecSchedulerIds(spec, "schedulerId");
    ASSERT_TRUE(spec->schedulerInfos->schedulerInstanceList[0]->updateTime > updateTime);
}

TEST_F(FaasInstanceManagerTest, AcquireCallbackTest)
{
    auto acquireSpec = std::make_shared<InvokeSpec>();
    auto invokeSpec = std::make_shared<InvokeSpec>();

    insManager->UpdateSchedulerInfo("schedulerkey",
                                    {SchedulerInstance{.InstanceName = "instanceName1", .InstanceID = "instanceId1"},
                                     SchedulerInstance{.InstanceName = "instanceName2", .InstanceID = "instanceId2"}});

    ErrorInfo err(YR::Libruntime::ErrorCode::ERR_INSTANCE_EXITED, "err msg");
    ErrorInfo outputErr;
    insManager->scheduleInsCb = [&outputErr](const RequestResource &resource, const ErrorInfo &err, bool isRemainIs) -> void {
        outputErr = err;
    };
    insManager->AcquireCallback(acquireSpec, err, InstanceResponse{}, invokeSpec);
    std::cout << "size is :" << invokeSpec->schedulerInfos->schedulerInstanceList.size() << std::endl;
    ASSERT_NE(outputErr.Code(), YR::Libruntime::ErrorCode::ERR_ALL_SCHEDULER_UNAVALIABLE);

    acquireSpec->invokeInstanceId = invokeSpec->schedulerInfos->schedulerInstanceList[0]->InstanceID;
    insManager->AcquireCallback(acquireSpec, err, InstanceResponse{}, invokeSpec);
    ASSERT_NE(outputErr.Code(), YR::Libruntime::ErrorCode::ERR_ALL_SCHEDULER_UNAVALIABLE);

    acquireSpec->invokeInstanceId = invokeSpec->schedulerInfos->schedulerInstanceList[1]->InstanceID;
    insManager->AcquireCallback(acquireSpec, err, InstanceResponse{}, invokeSpec);
    ASSERT_EQ(outputErr.Code(), YR::Libruntime::ErrorCode::ERR_ALL_SCHEDULER_UNAVALIABLE);
}
}  // namespace test
}  // namespace YR