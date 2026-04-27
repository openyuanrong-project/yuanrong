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

#include <thread>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <json.hpp>
#include <random>
#include "mock/mock_fs_intf.h"
#include "mock/mock_fs_intf_with_callback.h"
#include "src/dto/invoke_options.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/invoke_spec.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"
#include "src/utility/timer_worker.h"
#define protected public
#include "src/libruntime/invokeadaptor/instance_manager.h"
#define private public
#include "src/libruntime/invokeadaptor/faas_instance_manager.h"
#include "src/libruntime/invokeadaptor/normal_instance_manager.h"
#include "src/libruntime/invokeadaptor/task_submitter.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;
using YR::utility::CloseGlobalTimer;
using YR::utility::InitGlobalTimer;

namespace YR {

namespace test {
class TaskSubmitterTest : public testing::Test {
public:
    TaskSubmitterTest(){};
    ~TaskSubmitterTest(){};
    static void SetUpTestCase() {};
    static void TearDownTestCase() {};
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
        };
        InitLog(g_logParam);
        InitGlobalTimer();
        auto reqMgr = std::make_shared<RequestManager>();
        auto librtCfg = std::make_shared<LibruntimeConfig>();
        mockFsIntf = std::make_shared<MockFsIntf>();
        auto fsClient = std::make_shared<FSClient>(mockFsIntf);
        std::shared_ptr<MemoryStore> memoryStore = std::make_shared<MemoryStore>();
        auto dsObjectStore = std::make_shared<DSCacheObjectStore>();
        dsObjectStore->Init("127.0.0.1", 8080);
        auto wom = std::make_shared<WaitingObjectManager>();
        memoryStore->Init(dsObjectStore, wom);
        KillFunc f = [](const std::string &instanceId, const std::string &payload, int signal) -> ErrorInfo {
            return ErrorInfo();
        };
        std::string functionId = "";
        auto clientsMgr = std::make_shared<ClientsManager>();
        auto security = std::make_shared<Security>();
        auto downgrade = std::make_shared<YR::scene::DowngradeController>(functionId, clientsMgr, security);
        downgrade->Init();
        taskSubmitter =
            std::make_shared<TaskSubmitter>(librtCfg, memoryStore, fsClient, reqMgr, f, nullptr, nullptr, downgrade);
    }
    void TearDown() override
    {
        CloseGlobalTimer();
        mockFsIntf.reset();
        if (taskSubmitter) {
            taskSubmitter->Finalize();
        }
        taskSubmitter.reset();
    }

protected:
    bool NeedRetryWrapper(std::shared_ptr<InvokeSpec> &spec, ErrorCode errcode, bool &consume);
    std::shared_ptr<MockGwClient> SetMaxConcurrencyInstanceNum(int concurrencyCreateNum = 100000);
    void SubmitFunction(int total, bool differentResource = false);
    void CommonAssert(std::shared_ptr<TimerWorker> timerWorker, std::mutex &timersMtx,
                      std::shared_ptr<MockGwClient> mockFsIntf,
                      std::vector<std::shared_ptr<YR::utility::Timer>> &timers, bool differentResource = false,
                      int total = 10000);
    std::shared_ptr<TaskSubmitter> taskSubmitter;
    std::shared_ptr<FSIntf> mockFsIntf;
};
int GetRandomInt(int min, int max);
TEST_F(TaskSubmitterTest, ScheduleFunction)
{
    auto spec = std::make_shared<InvokeSpec>();

    spec->jobId = "job-7c8e6fab";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->opts = {};
    spec->returnIds = std::vector<DataObject>({DataObject{"obj-id"}});
    spec->invokeArgs = std::vector<InvokeArg>();
    auto resource = GetRequestResource(spec);
    taskSubmitter->SubmitFunction(spec);
    absl::ReaderMutexLock lock(&taskSubmitter->reqMtx_);
    ASSERT_EQ(taskSubmitter->taskSchedulerMap_[resource]->queue->Empty(), false);
    sleep(3);
}

TEST_F(TaskSubmitterTest, HandleInvokeNotify)
{
    NotifyRequest req;
    req.set_requestid("cae7c30c8d63f5ed00");
    req.set_code(common::ErrorCode::ERR_NONE);
    taskSubmitter->HandleInvokeNotify(req, ErrorInfo());
    req.set_requestid("cae7c30c8d63f5ee00");
    auto spec = std::make_shared<InvokeSpec>();
    spec->jobId = "job-7c8e6fab";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    Device device;
    device.name = "device";
    InvokeOptions opts;
    opts.device = device;
    spec->opts = opts;
    spec->requestId = "cae7c30c8d63f5ee00";
    spec->returnIds = std::vector<DataObject>({DataObject{"obj-id"}});
    taskSubmitter->SubmitFunction(spec);
    std::this_thread::sleep_for(std::chrono::milliseconds(3000));
    taskSubmitter->HandleInvokeNotify(req, ErrorInfo());
    auto resource = GetRequestResource(spec);
    absl::ReaderMutexLock lock(&taskSubmitter->reqMtx_);
    ASSERT_TRUE(taskSubmitter->taskSchedulerMap_[resource]->queue->Size() <= 1);
}

TEST_F(TaskSubmitterTest, HandleFailInvokeNotify)
{
    NotifyRequest req;
    req.set_requestid("cae7c30c8d63f5ed00");
    req.set_code(common::ErrorCode::ERR_PARAM_INVALID);
    auto spec = std::make_shared<InvokeSpec>();
    spec->jobId = "job-7c8e6fab";
    InvokeOptions opts;
    opts.retryTimes = 1;
    spec->seq = 0;
    spec->opts = opts;
    spec->requestId = "cae7c30c8d63f5ed00";
    spec->invokeInstanceId = "insId";
    spec->invokeLeaseId = "leaseId";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Faas};
    spec->invokeType = libruntime::InvokeType::InvokeFunctionStateless;
    auto resource = GetRequestResource(spec);
    taskSubmitter->HandleFailInvokeNotify(req, spec, resource, ErrorInfo());
    std::this_thread::sleep_for(std::chrono::milliseconds(200));
    ASSERT_EQ(spec->opts.retryTimes, 1);

    req.set_code(common::ErrorCode::ERR_INSTANCE_EVICTED);
    {
        absl::WriterMutexLock lock(&taskSubmitter->reqMtx_);
        taskSubmitter->taskSchedulerMap_[resource] = std::make_shared<TaskSchedulerWrapper>(false);
    }
    taskSubmitter->HandleFailInvokeNotify(req, spec, resource, ErrorInfo());
    auto [rawRequestId, seq] = YR::utility::IDGenerator::DecodeRawRequestId(spec->requestInvoke->Mutable().requestid());
    ASSERT_EQ(rawRequestId, "cae7c30c8d63f5ed00");
    ASSERT_EQ(seq, spec->seq);
}

TEST_F(TaskSubmitterTest, HandleFailInvokeNotifyUserCodeError)
{
    NotifyRequest req;
    req.set_requestid("cae7c30c8d63f5ee00");
    req.set_code(common::ErrorCode::ERR_USER_FUNCTION_EXCEPTION);
    auto spec = std::make_shared<InvokeSpec>();
    spec->jobId = "job-7c8e6fab";
    InvokeOptions opts;
    opts.retryTimes = 0;
    spec->opts = opts;
    spec->requestId = "cae7c30c8d63f5ed00";
    spec->invokeInstanceId = "insId";
    spec->invokeLeaseId = "leaseId";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    spec->invokeType = libruntime::InvokeType::InvokeFunctionStateless;
    auto resource = GetRequestResource(spec);
    auto normInsMgr =
        std::dynamic_pointer_cast<NormalInsManager>(taskSubmitter->insManagers[spec->functionMeta.apiType]);
    auto createSpec = normInsMgr->BuildCreateSpec(spec);
    ASSERT_TRUE(std::find(createSpec->requestCreate.labels().begin(), createSpec->requestCreate.labels().end(),
                          "task") != createSpec->requestCreate.labels().end());
    normInsMgr->AddInsInfo(createSpec, resource);
    taskSubmitter->HandleFailInvokeNotify(req, spec, resource, ErrorInfo());
    std::shared_ptr<RequestResourceInfo> insInfo;
    {
        absl::ReaderMutexLock lock(&taskSubmitter->insManagers[spec->functionMeta.apiType]->insMtx);
        auto requestResourceInfoMap = taskSubmitter->insManagers[spec->functionMeta.apiType]->requestResourceInfoMap;
        ASSERT_TRUE(requestResourceInfoMap.find(resource) != requestResourceInfoMap.end());
        insInfo = requestResourceInfoMap[resource];
    }
    {
        absl::ReaderMutexLock lock(&insInfo->mtx);
        ASSERT_TRUE(insInfo->instanceInfos.find(spec->invokeInstanceId) == insInfo->instanceInfos.end());
    }
}

bool TaskSubmitterTest::NeedRetryWrapper(std::shared_ptr<InvokeSpec> &spec, ErrorCode errcode, bool &consume)
{
    ErrorInfo err;
    err.SetErrorCode(errcode);
    return taskSubmitter->NeedRetry(err, spec, consume);
}

TEST_F(TaskSubmitterTest, NeedRetry)
{
    bool consume = false;
    ErrorInfo err;
    auto spec = std::make_shared<InvokeSpec>();
    spec->requestId = "cae7c30c8d63f5ed00";
    spec->invokeType = libruntime::InvokeType::InvokeFunctionStateless;
    spec->opts.retryTimes = 0;
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_OK, consume), false);
    ASSERT_EQ(consume, false);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_USER_FUNCTION_EXCEPTION, consume), false);
    ASSERT_EQ(consume, false);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_INSTANCE_EVICTED, consume), true);
    ASSERT_EQ(consume, false);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_INSTANCE_NOT_FOUND, consume), true);
    ASSERT_EQ(consume, false);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_INSTANCE_EXITED, consume), true);
    ASSERT_EQ(consume, false);

    spec->opts.retryTimes = 7;
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_USER_FUNCTION_EXCEPTION, consume), true);
    ASSERT_EQ(consume, true);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_BUS, consume), true);
    ASSERT_EQ(consume, true);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_INNER_COMMUNICATION, consume), true);
    ASSERT_EQ(consume, true);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_SHARED_MEMORY_LIMITED, consume), true);
    ASSERT_EQ(consume, true);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_OPERATE_DISK_FAILED, consume), true);
    ASSERT_EQ(consume, true);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_INSUFFICIENT_DISK_SPACE, consume), true);
    ASSERT_EQ(consume, true);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_INSTANCE_EVICTED, consume), true);
    ASSERT_EQ(consume, false);
    ASSERT_EQ(NeedRetryWrapper(spec, ErrorCode::ERR_USER_CODE_LOAD, consume), false);
    ASSERT_EQ(consume, false);

    ASSERT_EQ(taskSubmitter->NeedRetryCreate(ErrorInfo(ErrorCode::ERR_RESOURCE_NOT_ENOUGH, "")), true);
    ASSERT_EQ(taskSubmitter->NeedRetryCreate(ErrorInfo(ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_BUS, "")), true);
    ASSERT_EQ(taskSubmitter->NeedRetryCreate(ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, "")), false);
    ASSERT_EQ(taskSubmitter->NeedRetryCreate(ErrorInfo(ErrorCode::ERR_BUS_DISCONNECTION, "")), false);
}

TEST_F(TaskSubmitterTest, NeedRetryWithRetryChecker)
{
    bool consume = false;
    auto spec = std::make_shared<InvokeSpec>();

    spec->invokeType = libruntime::InvokeType::InvokeFunctionStateless;
    spec->opts.retryTimes = 5;
    spec->opts.retryChecker = [](const ErrorInfo &errInfo) -> bool {
        if (errInfo.Code() == ErrorCode::ERR_USER_FUNCTION_EXCEPTION) {
            std::string msg = errInfo.Msg();
            if (msg.find("123") != std::string::npos) {
                return true;
            }
        }
        return false;
    };
    ErrorInfo err(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, "123");
    ASSERT_EQ(taskSubmitter->NeedRetry(err, spec, consume), true);
    ASSERT_EQ(consume, true);
    err.SetErrorMsg("456");
    ASSERT_EQ(taskSubmitter->NeedRetry(err, spec, consume), false);
    ASSERT_EQ(consume, false);
}

TEST_F(TaskSubmitterTest, CancelStatelessRequest)
{
    std::vector<std::string> objids;
    std::string requestId = YR::utility::IDGenerator::GenRequestId();
    objids.emplace_back(YR::utility::IDGenerator::GenObjectId(requestId, 0));
    KillFunc f = [](const std::string &instanceId, const std::string &payload, int signal) -> ErrorInfo {
        return ErrorInfo();
    };
    auto spec = std::make_shared<InvokeSpec>();
    spec->jobId = "jobId";
    spec->requestId = requestId;
    spec->invokeInstanceId = "instanceId";
    taskSubmitter->requestManager->PushRequest(spec);
    taskSubmitter->CancelStatelessRequest(spec, f, true, true, objids[0]);
    auto res = taskSubmitter->requestManager->GetRequest(requestId);
    ASSERT_EQ(res, nullptr);

    spec->opts.customExtensions["Concurrency"] = "3";
    taskSubmitter->requestManager->PushRequest(spec);
    taskSubmitter->CancelStatelessRequest(spec, f, true, true, objids[0]);
    auto res1 = taskSubmitter->requestManager->GetRequest(requestId);
    ASSERT_EQ(res1, spec);

    auto res2 = taskSubmitter->CancelStatelessRequest(spec, f, true, false, objids[0]);
    EXPECT_TRUE(res2.OK());
}

std::list<std::shared_ptr<LabelOperator>> GetMockLabelOperators()
{
    std::list<std::shared_ptr<LabelOperator>> operators;
    std::shared_ptr<LabelOperator> inOperator = std::make_shared<LabelInOperator>();
    inOperator->SetKey("k1");
    inOperator->SetValues({"v1", "v2"});
    std::shared_ptr<LabelOperator> notInOperator = std::make_shared<LabelNotInOperator>();
    notInOperator->SetKey("k1");
    notInOperator->SetValues({"v1", "v2"});
    std::shared_ptr<LabelOperator> existsOperator = std::make_shared<LabelExistsOperator>();
    existsOperator->SetKey("k1");
    std::shared_ptr<LabelOperator> notExistOperator = std::make_shared<LabelDoesNotExistOperator>();
    notExistOperator->SetKey("k1");
    operators.push_back(inOperator);
    operators.push_back(notInOperator);
    operators.push_back(existsOperator);
    operators.push_back(notExistOperator);
    return operators;
}

std::list<std::shared_ptr<YR::Libruntime::Affinity>> GetMockAffinity()
{
    std::list<std::shared_ptr<Affinity>> affinities;
    std::shared_ptr<Affinity> resourcePreferred = std::make_shared<ResourcePreferredAffinity>();
    resourcePreferred->SetLabelOperators(GetMockLabelOperators());
    std::shared_ptr<Affinity> instancePreferred = std::make_shared<InstancePreferredAffinity>();
    instancePreferred->SetLabelOperators(GetMockLabelOperators());
    std::shared_ptr<Affinity> resourcePreferredAnti = std::make_shared<ResourcePreferredAntiAffinity>();
    resourcePreferredAnti->SetLabelOperators(GetMockLabelOperators());
    std::shared_ptr<Affinity> instancePreferredAnti = std::make_shared<InstancePreferredAntiAffinity>();
    instancePreferredAnti->SetLabelOperators(GetMockLabelOperators());
    std::shared_ptr<Affinity> resourceRequired = std::make_shared<ResourceRequiredAffinity>();
    resourceRequired->SetLabelOperators(GetMockLabelOperators());
    std::shared_ptr<Affinity> instanceRequired = std::make_shared<InstanceRequiredAffinity>();
    instanceRequired->SetLabelOperators(GetMockLabelOperators());
    std::shared_ptr<Affinity> resourceRequiredAnti = std::make_shared<ResourceRequiredAntiAffinity>();
    resourceRequiredAnti->SetLabelOperators(GetMockLabelOperators());
    std::shared_ptr<Affinity> instanceRequiredAnti = std::make_shared<InstanceRequiredAntiAffinity>();
    instanceRequiredAnti->SetLabelOperators(GetMockLabelOperators());
    instanceRequiredAnti->SetAffinityScope(AFFINITYSCOPE_NODE);
    affinities.push_back(resourcePreferred);
    affinities.push_back(instancePreferred);
    affinities.push_back(resourcePreferredAnti);
    affinities.push_back(instancePreferredAnti);
    affinities.push_back(resourceRequired);
    affinities.push_back(instanceRequired);
    affinities.push_back(resourceRequiredAnti);
    affinities.push_back(instanceRequiredAnti);
    return affinities;
}

TEST_F(TaskSubmitterTest, TestAffinity)
{
    std::cout << "test start" << std::endl;
    auto spec = std::make_shared<InvokeSpec>();
    spec->jobId = "job-7c8e6fab";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    InvokeOptions opts = {};
    std::list<std::shared_ptr<Affinity>> affinities = GetMockAffinity();
    std::shared_ptr<Affinity> fst = affinities.front();
    ASSERT_EQ(fst->GetAffinityKind(), "Resource");
    ASSERT_EQ(fst->GetAffinityType(), "PreferredAffinity");
    std::list<std::shared_ptr<LabelOperator>> labelOperators = fst->GetLabelOperators();
    ASSERT_EQ(labelOperators.size(), 4);
    std::shared_ptr<LabelOperator> inOperator = labelOperators.front();
    ASSERT_EQ(inOperator->GetOperatorType(), "LabelIn");
    ASSERT_EQ(inOperator->GetKey(), "k1");
    ASSERT_EQ(inOperator->GetValues().size(), 2);
    std::shared_ptr<Affinity> last = affinities.back();
    ASSERT_EQ(last->GetAffinityScope(), "NODE");
    opts.scheduleAffinities = affinities;
    spec->opts = opts;
    spec->returnIds = std::vector<DataObject>({DataObject{"obj-id"}});
    spec->invokeArgs = std::vector<InvokeArg>();
    auto resource = GetRequestResource(spec);
    taskSubmitter->SubmitFunction(spec);
    absl::ReaderMutexLock lock(&taskSubmitter->reqMtx_);
    ASSERT_EQ(taskSubmitter->taskSchedulerMap_[resource]->queue->Empty(), false);
    sleep(3);
}

TEST_F(TaskSubmitterTest, Test_When_UpdateConfig_Should_Update_OK)
{
    auto cfg = std::make_shared<LibruntimeConfig>();
    auto insManager = std::make_shared<NormalInsManager>();
    taskSubmitter->insManagers[libruntime::ApiType::Function] = insManager;
    taskSubmitter->libRuntimeConfig->recycleTime = 1;
    taskSubmitter->UpdateConfig();
    EXPECT_EQ(taskSubmitter->recycleTimeMs, 1000);
    EXPECT_EQ(taskSubmitter->insManagers[libruntime::ApiType::Function]->recycleTimeMs, 1000);

    taskSubmitter->libRuntimeConfig->recycleTime = 5;
    taskSubmitter->UpdateConfig();
    EXPECT_EQ(taskSubmitter->recycleTimeMs, 5000);
    EXPECT_EQ(taskSubmitter->insManagers[libruntime::ApiType::Function]->recycleTimeMs, 5000);
}

TEST_F(TaskSubmitterTest, ScheduleInsTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->requestId = "reqId";
    spec->invokeInstanceId = "instanceId";
    ErrorInfo err(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, "errMsg");
    auto resource = GetRequestResource(spec);
    {
        absl::WriterMutexLock lock(&taskSubmitter->reqMtx_);
        taskSubmitter->taskSchedulerMap_[resource] = std::make_shared<TaskSchedulerWrapper>(false);
        taskSubmitter->taskSchedulerMap_[resource]->queue->Push(spec);
    }
    taskSubmitter->ScheduleIns(resource, err, false);
    absl::ReaderMutexLock lock(&taskSubmitter->reqMtx_);
    EXPECT_EQ(taskSubmitter->taskSchedulerMap_[resource]->queue->Empty(), true);
}

std::shared_ptr<MockGwClient> TaskSubmitterTest::SetMaxConcurrencyInstanceNum(int concurrencyCreateNum)
{
    // construct taskSubmitter
    KillFunc f = [](const std::string &instanceId, const std::string &payload, int signal) -> ErrorInfo {
        return ErrorInfo();
    };
    auto reqMgr = std::make_shared<RequestManager>();
    auto librtCfg = std::make_shared<LibruntimeConfig>();
    librtCfg->maxConcurrencyCreateNum = concurrencyCreateNum;
    auto mockFsIntf = std::make_shared<MockGwClient>();
    auto fsClient = std::make_shared<FSClient>(mockFsIntf);
    std::shared_ptr<MemoryStore> memoryStore = std::make_shared<MemoryStore>();
    auto dsObjectStore = std::make_shared<DSCacheObjectStore>();
    dsObjectStore->Init("127.0.0.1", 8080);
    auto wom = std::make_shared<WaitingObjectManager>();
    memoryStore->Init(dsObjectStore, wom);
    std::string functionId = "";
    auto clientsMgr = std::make_shared<ClientsManager>();
    auto security = std::make_shared<Security>();
    auto downgrade = std::make_shared<YR::scene::DowngradeController>(functionId, clientsMgr, security);
    downgrade->Init();
    taskSubmitter =
        std::make_shared<TaskSubmitter>(librtCfg, memoryStore, fsClient, reqMgr, f, nullptr, nullptr, downgrade);
    return mockFsIntf;
}

void TaskSubmitterTest::SubmitFunction(int total, bool differentResource)
{
    for (int i = 0; i < total; i++) {
        auto spec = std::make_shared<InvokeSpec>();
        spec->jobId = "job-7c8e6fab";
        spec->functionMeta = {
            "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
        InvokeOptions options = {};
        if (differentResource) {
            options.instanceSession = std::make_shared<InstanceSession>();
            options.instanceSession->sessionID = YR::utility::IDGenerator::GenRequestId();
        }
        spec->opts = options;
        spec->returnIds = std::vector<DataObject>({DataObject{"obj-id"}});
        spec->invokeArgs = std::vector<InvokeArg>();
        spec->requestId = YR::utility::IDGenerator::GenRequestId();
        LibruntimeConfig config;
        spec->BuildInstanceInvokeRequest(config);
        taskSubmitter->requestManager->PushRequest(spec);
        taskSubmitter->SubmitFunction(spec);
    }
}

void TaskSubmitterTest::CommonAssert(std::shared_ptr<TimerWorker> timerWorker, std::mutex &timersMtx,
                                     std::shared_ptr<MockGwClient> mockFsIntf,
                                     std::vector<std::shared_ptr<YR::utility::Timer>> &timers, bool differentResource,
                                     int total)
{
    std::mutex mtx;
    int get = 0;
    auto promise = std::promise<bool>();
    auto future = promise.get_future();
    EXPECT_CALL(*mockFsIntf, InvokeAsync(_, _, _))
        .WillRepeatedly(
            [&, total](const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback, int timeoutSec) {
                auto invokeRspReturnTime = GetRandomInt(20, 100);
                auto timer = timerWorker->CreateTimer(invokeRspReturnTime, 1, [&, req, callback, total] {
                    NotifyRequest notifyReq;
                    notifyReq.set_requestid(req->Immutable().requestid());
                    notifyReq.set_code(::common::ErrorCode::ERR_NONE);
                    callback(notifyReq, ErrorInfo());
                    std::lock_guard<std::mutex> lockGuard(mtx);
                    get++;
                    if (get == total) {
                        promise.set_value(true);
                    }
                });
                std::lock_guard<std::mutex> lockGuard(timersMtx);
                timers.push_back(timer);
            });
    EXPECT_CALL(*mockFsIntf, KillAsync(_, _, _)).WillRepeatedly(Return());
    SubmitFunction(total, differentResource);
    ASSERT_TRUE(future.get());
    auto instanceIds = taskSubmitter->GetInstanceIds();
    std::cout << " create " << instanceIds.size() << std::endl;
    timers.clear();
}

int GetRandomInt(int min, int max)
{
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<int> dist(min, max);
    return dist(gen);
}

// invoke notify don't return
TEST_F(TaskSubmitterTest, ScheduleFunction_Benchmark)
{
    auto mockFsIntf = SetMaxConcurrencyInstanceNum(100000);
    auto timerWorker = std::make_shared<TimerWorker>();
    auto timers = std::vector<std::shared_ptr<YR::utility::Timer>>();
    // mock function
    // 再构造一下create 20ms~30ms之间才返回的
    EXPECT_CALL(*mockFsIntf, CreateAsync(_, _, _, _))
        .WillRepeatedly([timerWorker, &timers](const CreateRequest &req, CreateRespCallback respCallback,
                                               CreateCallBack callback, int timeoutSec) {
            auto createRspReturnTime = GetRandomInt(10, 20);
            auto timer = timerWorker->CreateTimer(createRspReturnTime, 1, [req, respCallback, callback] {
                CreateResponse response;
                auto instanceId = YR::utility::IDGenerator::GenRequestId();
                response.set_instanceid(instanceId);
                response.set_code(::common::ErrorCode::ERR_NONE);
                respCallback(response);
                NotifyRequest notifyReq;
                notifyReq.set_requestid(req.requestid());
                notifyReq.set_code(::common::ErrorCode::ERR_NONE);
                callback(notifyReq);
            });
            timers.push_back(timer);
        });

    EXPECT_CALL(*mockFsIntf, InvokeAsync(_, _, _))
        .WillRepeatedly(
            [](const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback, int timeoutSec) { return; });
    EXPECT_CALL(*mockFsIntf, KillAsync(_, _, _)).WillRepeatedly(Return());
    // benchmark
    auto start = std::chrono::high_resolution_clock::now();
    size_t total = 50000;
    if (const char *env = std::getenv("YR_BENCHMARK_SCALE")) {
        total = std::stoi(env);
    }

    SubmitFunction(total);
    auto submitEnd = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double, std::milli> submitCostMs = submitEnd - start;
    std::cout << "【benchmark】 submit " << total << " function cost time: " << submitCostMs.count() << " milliseconds"
              << std::endl;
    for (;;) {
        auto instanceIds = taskSubmitter->GetInstanceIds();
        auto second = std::chrono::high_resolution_clock::now();
        std::chrono::duration<double, std::milli> durationMs = second - start;
        std::cout << "【benchmark】 create " << instanceIds.size() << " instances cost time: " << durationMs.count()
                  << " milliseconds" << std::endl;
        if (instanceIds.size() >= total) {
            break;
        }
        std::promise<bool> promise;
        auto future = promise.get_future();
        future.wait_for(std::chrono::milliseconds(1000));
    }
    auto end = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double, std::milli> durationMs = end - start;
    std::cout << "【benchmark】 create " << total << " instances cost time: " << durationMs.count() << " milliseconds"
              << std::endl;
    ASSERT_TRUE(static_cast<int>(durationMs.count()) <= 80000);
}

TEST_F(TaskSubmitterTest, ScheduleFunction_InvokeAllSuccess)
{
    auto mockFsIntf = SetMaxConcurrencyInstanceNum(10000);
    auto timerWorker = std::make_shared<TimerWorker>();
    std::mutex timersMtx;
    auto timers = std::vector<std::shared_ptr<YR::utility::Timer>>();
    // mock function
    EXPECT_CALL(*mockFsIntf, CreateAsync(_, _, _, _))
        .WillRepeatedly([timerWorker, &timers, &timersMtx](const CreateRequest &req, CreateRespCallback respCallback,
                                                           CreateCallBack callback, int timeoutSec) {
            auto createRspReturnTime = GetRandomInt(10, 60);
            auto timer = timerWorker->CreateTimer(createRspReturnTime, 1, [&timersMtx, req, respCallback, callback] {
                CreateResponse response;
                auto instanceId = YR::utility::IDGenerator::GenRequestId();
                response.set_instanceid(instanceId);
                response.set_code(::common::ErrorCode::ERR_NONE);
                respCallback(response);
                NotifyRequest notifyReq;
                notifyReq.set_requestid(req.requestid());
                notifyReq.set_code(::common::ErrorCode::ERR_NONE);
                callback(notifyReq);
            });
            std::lock_guard<std::mutex> lockGuard(timersMtx);
            timers.push_back(timer);
        });
    CommonAssert(timerWorker, timersMtx, mockFsIntf, timers);
}

TEST_F(TaskSubmitterTest, ScheduleFunction_CreateRandomAbnormal)
{
    auto mockFsIntf = SetMaxConcurrencyInstanceNum(10000);
    auto timerWorker = std::make_shared<TimerWorker>();
    std::mutex timersMtx;
    auto timers = std::vector<std::shared_ptr<YR::utility::Timer>>();
    // mock function
    EXPECT_CALL(*mockFsIntf, CreateAsync(_, _, _, _))
        .WillRepeatedly([timerWorker, &timers, &timersMtx](const CreateRequest &req, CreateRespCallback respCallback,
                                                           CreateCallBack callback, int timeoutSec) {
            auto createRspReturnTime = GetRandomInt(10, 60);
            auto timer = timerWorker->CreateTimer(
                createRspReturnTime, 1, [&timersMtx, req, respCallback, callback, createRspReturnTime] {
                    CreateResponse response;
                    auto instanceId = YR::utility::IDGenerator::GenRequestId();
                    response.set_instanceid(instanceId);
                    response.set_code(::common::ErrorCode::ERR_NONE);
                    if (createRspReturnTime > 50) {
                        response.set_code(::common::ErrorCode::ERR_RESOURCE_NOT_ENOUGH);
                        respCallback(response);
                        return;
                    }
                    respCallback(response);
                    NotifyRequest notifyReq;
                    notifyReq.set_requestid(req.requestid());
                    notifyReq.set_code(::common::ErrorCode::ERR_NONE);
                    callback(notifyReq);
                });
            std::lock_guard<std::mutex> lockGuard(timersMtx);
            timers.push_back(timer);
        });
    CommonAssert(timerWorker, timersMtx, mockFsIntf, timers);
}

TEST_F(TaskSubmitterTest, ScheduleFunction_DifferentSessionResource)
{
    auto mockFsIntf = SetMaxConcurrencyInstanceNum(10000);
    auto timerWorker = std::make_shared<TimerWorker>();
    std::mutex timersMtx;
    auto timers = std::vector<std::shared_ptr<YR::utility::Timer>>();
    // mock function
    EXPECT_CALL(*mockFsIntf, CreateAsync(_, _, _, _))
        .WillRepeatedly([timerWorker, &timers, &timersMtx](const CreateRequest &req, CreateRespCallback respCallback,
                                                           CreateCallBack callback, int timeoutSec) {
            auto createRspReturnTime = GetRandomInt(10, 60);
            auto timer = timerWorker->CreateTimer(createRspReturnTime, 1, [&timersMtx, req, respCallback, callback] {
                CreateResponse response;
                auto instanceId = YR::utility::IDGenerator::GenRequestId();
                response.set_instanceid(instanceId);
                response.set_code(::common::ErrorCode::ERR_NONE);
                respCallback(response);
                NotifyRequest notifyReq;
                notifyReq.set_requestid(req.requestid());
                notifyReq.set_code(::common::ErrorCode::ERR_NONE);
                callback(notifyReq);
            });
            std::lock_guard<std::mutex> lockGuard(timersMtx);
            timers.push_back(timer);
        });
    CommonAssert(timerWorker, timersMtx, mockFsIntf, timers, true, 3000);
}

TEST_F(TaskSubmitterTest, CancelFaasScheduleTimeoutReqTest)
{
    ASSERT_NO_THROW(taskSubmitter->CancelFaasScheduleTimeoutReq("reqId", 100 * 1000));

    auto spec = std::make_shared<InvokeSpec>();
    spec->requestId = "reqId";
    spec->returnIds = {DataObject("objId")};
    auto resource = GetRequestResource(spec);
    taskSubmitter->taskSchedulerMap_[resource] = std::make_shared<TaskSchedulerWrapper>(false);
    taskSubmitter->taskSchedulerMap_[resource]->queue->Push(spec);
    taskSubmitter->requestManager->PushRequest(spec);
    taskSubmitter->CancelFaasScheduleTimeoutReq("reqId", 100 * 1000);
    ASSERT_EQ(taskSubmitter->requestManager->GetRequest("reqId"), nullptr);

    taskSubmitter->requestManager->PushRequest(spec);
    resource = GetRequestResource(spec);
    taskSubmitter->taskSchedulerMap_[resource] = std::make_shared<TaskSchedulerWrapper>(false);
    taskSubmitter->taskSchedulerMap_[resource]->SetLastError(
        ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, "error"));
    taskSubmitter->CancelFaasScheduleTimeoutReq("reqId", 100 * 1000);
    ASSERT_EQ(taskSubmitter->requestManager->GetRequest("reqId"), nullptr);
}

TEST_F(TaskSubmitterTest, EraseFaasCancelTimerTest)
{
    taskSubmitter->faasCancelTimerWorkers["reqId"] = YR::utility::ExecuteByGlobalTimer([]() -> void {}, 1, 1);
    ASSERT_EQ(taskSubmitter->faasCancelTimerWorkers.size(), 1);
    taskSubmitter->EraseFaasCancelTimer("reqId");
    ASSERT_EQ(taskSubmitter->faasCancelTimerWorkers.size(), 0);
}

TEST_F(TaskSubmitterTest, UpdateSchdulerInfoTest)
{
    ASSERT_NO_THROW(taskSubmitter->UpdateSchdulerInfo("schedulerName", "schedulerId", "ADD"));
    ASSERT_NO_THROW(taskSubmitter->UpdateSchdulerInfo("schedulerName", "schedulerId", "REMOVE"));
    ASSERT_NO_THROW(taskSubmitter->UpdateSchdulerInfo("schedulerName", "schedulerId", "UNKOWN"));
}

TEST_F(TaskSubmitterTest, RecordFaasInvokeDataAndUpdateTest)
{
    taskSubmitter->metricsEnable_ = true;
    ASSERT_NO_THROW(taskSubmitter->UpdateFaasInvokeLog("reqId", ErrorInfo()));
    auto spec = std::make_shared<InvokeSpec>();
    spec->requestId = "reqId";
    spec->functionMeta = YR::Libruntime::FunctionMeta{.funcName = "funcName"};
    taskSubmitter->RecordFaasInvokeData(spec);
    ASSERT_TRUE(!taskSubmitter->faasInvokeDataMap_.empty());
    ASSERT_NO_THROW(taskSubmitter->UpdateFaasInvokeLog("reqId", ErrorInfo()));
    taskSubmitter->faasInvokeDataMap_.erase("reqId");
    ASSERT_NO_THROW(taskSubmitter->UpdateFaasInvokeLog(
        "reqId", ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, "err msg")));
}

TEST_F(TaskSubmitterTest, InvokeExceedMaxRetryTimeShouldNotStuck)
{
    auto mockFsIntf = SetMaxConcurrencyInstanceNum(100000);
    // mock function
    EXPECT_CALL(*mockFsIntf, CreateAsync(_, _, _, _))
        .WillRepeatedly(
            [](const CreateRequest &req, CreateRespCallback respCallback, CreateCallBack callback, int timeoutSec) {
                CreateResponse response;
                auto instanceId = YR::utility::IDGenerator::GenRequestId();
                response.set_instanceid(instanceId);
                response.set_code(::common::ErrorCode::ERR_NONE);
                respCallback(response);
                NotifyRequest notifyReq;
                notifyReq.set_requestid(req.requestid());
                notifyReq.set_code(::common::ErrorCode::ERR_NONE);
                callback(notifyReq);
            });
    std::mutex mtx;
    int get = 0;
    int retry = 5;
    int want = retry + 1;
    auto promise = std::promise<bool>();
    auto future = promise.get_future();
    EXPECT_CALL(*mockFsIntf, InvokeAsync(_, _, _))
        .WillRepeatedly([&](const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback, int timeoutSec) {
            NotifyRequest notifyReq;
            notifyReq.set_requestid(req->Immutable().requestid());
            notifyReq.set_code(::common::ErrorCode::ERR_INSTANCE_EVICTED);
            callback(notifyReq, ErrorInfo());
            std::lock_guard<std::mutex> lockGuard(mtx);
            get++;
            if (get == want) {
                promise.set_value(true);
            }
        });
    EXPECT_CALL(*mockFsIntf, KillAsync(_, _, _)).WillRepeatedly(Return());
    auto spec = std::make_shared<InvokeSpec>();
    spec->jobId = "job-7c8e6fab";
    spec->functionMeta = {
        "", "", "funcname", "classname", libruntime::LanguageType::Cpp, "", "", "", libruntime::ApiType::Function};
    InvokeOptions options = {};
    options.maxRetryTime = retry;
    spec->opts = options;
    spec->invokeType = libruntime::InvokeType::InvokeFunctionStateless;
    spec->returnIds = std::vector<DataObject>({DataObject{"obj-id"}});
    spec->invokeArgs = std::vector<InvokeArg>();
    spec->requestId = YR::utility::IDGenerator::GenRequestId();
    LibruntimeConfig config;
    spec->BuildInstanceInvokeRequest(config);
    taskSubmitter->requestManager->PushRequest(spec);
    taskSubmitter->SubmitFunction(spec);
    ASSERT_TRUE(future.get());
    ASSERT_EQ(get, want);
}

TEST_F(TaskSubmitterTest, ConvertToGaugeDataTest)
{
    auto data = std::make_shared<YR::Libruntime::FaasInvokeData>(
        YR::Libruntime::FaasInvokeData("tenantId", "functionName", "inputAlias", "traceId", 0));
    auto res = taskSubmitter->ConvertToGaugeData(data, "reqId");
    ASSERT_EQ(res.name, "report_faas_invoke_data");
}

TEST_F(TaskSubmitterTest, AddFaasCancelTimerTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->opts.acquireTimeout = 300;
    spec->functionMeta.apiType = libruntime::ApiType::Faas;
    ASSERT_NO_THROW(taskSubmitter->AddFaasCancelTimer(spec));
}

TEST_F(TaskSubmitterTest, SendInvokeReqTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->opts.device = YR::Libruntime::Device{.name = "deviceName", .batch_size = 1};
    spec->functionMeta.apiType = libruntime::ApiType::Faas;
    auto resource = GetRequestResource(spec);
    ASSERT_NO_THROW(taskSubmitter->SendInvokeReq(resource, spec));
}

TEST_F(TaskSubmitterTest, UpdateFaasInvokeSendTimeTest)
{
    taskSubmitter->metricsEnable_ = true;
    auto currentTime = GetCurrentTimestampNs();
    auto data = std::make_shared<FaasInvokeData>();
    data->sendTime = currentTime;
    taskSubmitter->faasInvokeDataMap_["reqId"] = data;
    taskSubmitter->UpdateFaasInvokeSendTime("reqId");
    ASSERT_TRUE(taskSubmitter->faasInvokeDataMap_["reqId"]->sendTime > currentTime);
}

TEST_F(TaskSubmitterTest, HandleFailInvokeIsDelayScaleDownTest)
{
    NotifyRequest req;
    req.set_code(common::ErrorCode::ERR_PARAM_INVALID);
    ASSERT_TRUE(taskSubmitter->HandleFailInvokeIsDelayScaleDown(req, ErrorInfo()));
    req.set_code(common::ErrorCode::ERR_INSTANCE_NOT_FOUND);
    ASSERT_FALSE(taskSubmitter->HandleFailInvokeIsDelayScaleDown(req, ErrorInfo()));
    req.set_code(common::ErrorCode::ERR_INSTANCE_EXITED);
    ASSERT_FALSE(taskSubmitter->HandleFailInvokeIsDelayScaleDown(req, ErrorInfo()));
    req.set_code(common::ErrorCode::ERR_INSTANCE_EVICTED);
    ASSERT_FALSE(taskSubmitter->HandleFailInvokeIsDelayScaleDown(req, ErrorInfo()));
    req.set_code(common::ErrorCode::ERR_USER_FUNCTION_EXCEPTION);
    ASSERT_FALSE(taskSubmitter->HandleFailInvokeIsDelayScaleDown(req, ErrorInfo()));
}

}  // namespace test
}  // namespace YR
