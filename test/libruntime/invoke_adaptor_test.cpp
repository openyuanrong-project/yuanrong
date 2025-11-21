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
#include <chrono>
#include <future>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <boost/beast/http.hpp>
#include <json.hpp>
#include <string>
#include <fstream>
#include "mock/mock_datasystem_client.h"
#include "mock/mock_fs_intf.h"
#include "mock/mock_fs_intf_with_callback.h"
#include "mock/mock_task_submitter.h"
#include "src/libruntime/dependency_resolver.h"
#include "src/libruntime/invoke_spec.h"

#define private public

#include "src/libruntime/err_type.h"
#include "src/libruntime/generator/stream_generator_notifier.h"
#include "src/libruntime/generator/stream_generator_receiver.h"
#include "src/libruntime/groupmanager/group.h"
#include "src/libruntime/invokeadaptor/invoke_adaptor.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/libruntime/objectstore/memory_store.h"
#include "src/libruntime/runtime_context.h"
#include "src/libruntime/utils/serializer.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;
using namespace std::chrono_literals;
using json = nlohmann::json;

namespace YR {

namespace Libruntime {
bool ParseMetaData(const CallRequest &request, bool isPosix, libruntime::MetaData &metaData);
bool ParseFunctionGroupRunningInfo(const CallRequest &request, bool isPosix,
                                   common::FunctionGroupRunningInfo &runningInfo);
bool ParseFaasController(const CallRequest &request, std::vector<std::shared_ptr<DataObject>> &rawArgs, bool isPosix);
}  // namespace Libruntime
namespace test {

class InvokeAdaptorTest : public testing::Test {
public:
    InvokeAdaptorTest(){};
    ~InvokeAdaptorTest(){};
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
        auto dsObjectStore = std::make_shared<DSCacheObjectStore>();
        dsObjectStore->Init("127.0.0.1", 8080);
        auto wom = std::make_shared<WaitingObjectManager>();
        auto runtimeContext = std::make_shared<RuntimeContext>();
        libConfig = std::make_shared<LibruntimeConfig>();
        LibruntimeOptions option;
        option.loadFunctionCallback = [](const std::vector<std::string> &codePaths) -> ErrorInfo {
            return ErrorInfo();
        };
        libConfig->libruntimeOptions = option;
        libConfig->isDriver = false;
        libConfig->inCluster = false;
        this->memoryStore->Init(dsObjectStore, wom);
        auto dependencyResolver = std::make_shared<DependencyResolver>(this->memoryStore);
        this->gwClient = std::make_shared<MockFsIntf>();
        this->taskSubmitter = std::make_shared<MockTaskSubmitter>();
        auto cb = []() { return; };
        auto clientsMgr = std::make_shared<ClientsManager>();
        auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
        auto fsClient = std::make_shared<FSClient>(this->gwClient);
        auto mapper = std::make_shared<GeneratorIdMap>();
        auto dsStreamStore = std::shared_ptr<StreamStore>();
        auto generatorReceiver = std::make_shared<StreamGeneratorReceiver>(libConfig, dsStreamStore, this->memoryStore);
        auto generatorNotifier = std::make_shared<StreamGeneratorNotifier>(dsStreamStore, mapper);
        auto rGroupManager = std::make_shared<ResourceGroupManager>();
        auto security = std::make_shared<Security>();
        this->invokeAdaptor =
            std::make_shared<InvokeAdaptor>(libConfig, dependencyResolver, fsClient, this->memoryStore, runtimeContext,
                                            cb, nullptr, std::make_shared<InvokeOrderManager>(), clientsMgr,
                                            metricsAdaptor, mapper, generatorReceiver, generatorNotifier);
        invokeAdaptor->SetRGroupManager(rGroupManager);
        invokeAdaptor->SetCallbackOfSetTenantId([]() {});
        this->invokeAdaptor->Init(*runtimeContext, nullptr);
    }

    void TearDown() override
    {
        CloseGlobalTimer();
        this->gwClient.reset();
        this->libConfig.reset();
        this->memoryStore.reset();
        this->invokeAdaptor.reset();
    }
    std::shared_ptr<MockFsIntf> gwClient;
    std::shared_ptr<MemoryStore> memoryStore;
    std::shared_ptr<InvokeAdaptor> invokeAdaptor;
    std::shared_ptr<LibruntimeConfig> libConfig;
    std::shared_ptr<MockTaskSubmitter> taskSubmitter;
};

TEST_F(InvokeAdaptorTest, ParseInvokeRequestTest)
{
    CallRequest request;
    auto pbArg = request.add_args();
    pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    InvokeSpec invokeSpec;
    invokeSpec.invokeType = libruntime::InvokeType::InvokeFunction;
    pbArg->set_value(invokeSpec.BuildInvokeMetaData(*invokeAdaptor->librtConfig));

    auto pbArg2 = request.add_args();
    pbArg2->set_type(common::Arg::OBJECT_REF);
    std::string objId("mock-123");
    pbArg2->set_value(objId.c_str(), objId.size());

    std::vector<std::shared_ptr<DataObject>> rawArgs;
    bool ok = invokeAdaptor->ParseRequest(request, rawArgs, false);
    ASSERT_EQ(ok, true);
    libruntime::MetaData metaData;
    ok = YR::Libruntime::ParseMetaData(request, false, metaData);
    ASSERT_EQ(ok, true);
    ASSERT_EQ(metaData.invoketype(), invokeSpec.invokeType);

    auto pbArg3 = request.add_args();
    pbArg3->set_type(common::Arg::VALUE);
    pbArg3->set_value(objId.c_str(), objId.size());
    ok = invokeAdaptor->ParseRequest(request, rawArgs, false);
    ASSERT_EQ(ok, true);
}

TEST_F(InvokeAdaptorTest, ParseCreateRequestTest)
{
    CallRequest request;
    auto pbArg = request.add_args();
    pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    InvokeSpec invokeSpec;
    invokeSpec.invokeType = libruntime::InvokeType::CreateInstance;
    std::string meta;
    pbArg->set_value(invokeSpec.BuildCreateMetaData(*invokeAdaptor->librtConfig, meta));

    auto pbArg2 = request.add_args();
    pbArg2->set_type(common::Arg::OBJECT_REF);
    std::string objId("mock-123");
    pbArg2->set_value(objId.c_str(), objId.size());

    std::vector<std::shared_ptr<DataObject>> rawArgs;
    bool ok = invokeAdaptor->ParseRequest(request, rawArgs, false);
    ASSERT_EQ(ok, true);
    libruntime::MetaData metaData;
    ok = YR::Libruntime::ParseMetaData(request, false, metaData);
    ASSERT_EQ(ok, true);
    ASSERT_EQ(metaData.invoketype(), invokeSpec.invokeType);
}

TEST_F(InvokeAdaptorTest, ParseYrCreateRequestTest)
{
    CallRequest request;
    request.set_iscreate(true);
    std::vector<std::shared_ptr<DataObject>> rawArgs;
    bool ok = invokeAdaptor->ParseRequest(request, rawArgs, true);
    ASSERT_EQ(ok, true);
    ASSERT_EQ(rawArgs.empty(), true);
    libruntime::MetaData metaData;
    ok = YR::Libruntime::ParseMetaData(request, true, metaData);
    ASSERT_EQ(ok, true);
    ASSERT_EQ(metaData.invoketype(), libruntime::InvokeType::CreateInstance);
}

TEST_F(InvokeAdaptorTest, ParseYrInvokeRequestTest)
{
    CallRequest request;
    request.set_iscreate(false);
    std::vector<std::shared_ptr<DataObject>> rawArgs;
    bool ok = invokeAdaptor->ParseRequest(request, rawArgs, true);
    ASSERT_EQ(ok, true);
    ASSERT_EQ(rawArgs.empty(), true);
    libruntime::MetaData metaData;
    ok = YR::Libruntime::ParseMetaData(request, true, metaData);
    ASSERT_EQ(ok, true);
    ASSERT_EQ(metaData.invoketype(), libruntime::InvokeType::InvokeFunction);
}

TEST_F(InvokeAdaptorTest, PrepareCallExecutorTest)
{
    auto memStore = std::make_shared<MemoryStore>();
    auto dependencyResolver = std::make_shared<DependencyResolver>(memStore);
    auto librtCfg = std::make_shared<LibruntimeConfig>();
    auto runtimeContext = std::make_shared<RuntimeContext>(YR::utility::IDGenerator::GenApplicationId());
    auto execMgr = std::make_shared<InvokeOrderManager>();
    FSIntfHandlers handlers;

    auto gwClient = std::make_shared<GwClient>(librtCfg->functionIds[libruntime::LanguageType::Cpp], handlers);
    auto cb = []() { return; };
    auto clientsMgr = std::make_shared<ClientsManager>();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    auto fsClient = std::make_shared<FSClient>(gwClient);
    auto mapper = std::make_shared<GeneratorIdMap>();
    auto dsStreamStore = std::shared_ptr<StreamStore>();
    auto generatorReceiver = std::make_shared<StreamGeneratorReceiver>(librtCfg, dsStreamStore, memStore);
    auto generatorNotifier = std::make_shared<StreamGeneratorNotifier>(dsStreamStore, mapper);
    auto rGroupManager = std::make_shared<ResourceGroupManager>();
    auto invokeAdaptor = std::make_shared<InvokeAdaptor>(librtCfg, dependencyResolver, fsClient, memStore,
                                                         runtimeContext, cb, nullptr, execMgr, clientsMgr,
                                                         metricsAdaptor, mapper, generatorReceiver, generatorNotifier);
    invokeAdaptor->SetRGroupManager(rGroupManager);
    invokeAdaptor->SetCallbackOfSetTenantId([]() {});
    invokeAdaptor->Init(*runtimeContext, nullptr);

    struct MyTests {
        int concurrency;
        common::ErrorCode errCode;
    };
    std::vector<MyTests> myTests = {
        {100, common::ERR_NONE},
        {-1, common::ERR_PARAM_INVALID},
    };
    for (const auto &myTest : myTests) {
        CallRequest req;
        auto *createOptions = req.mutable_createoptions();
        createOptions->insert({CONCURRENT_NUM, std::to_string(myTest.concurrency)});
        auto [code, msg] = invokeAdaptor->PrepareCallExecutor(req);
        ASSERT_EQ(code, myTest.errCode);
    }

    CallRequest req;
    auto *createOptions = req.mutable_createoptions();
    createOptions->insert({CONCURRENT_NUM, "100"});
    createOptions->insert({NEED_ORDER, "true"});
    auto [code, msg] = invokeAdaptor->PrepareCallExecutor(req);
    ASSERT_EQ(code, common::ERR_PARAM_INVALID);
}

TEST_F(InvokeAdaptorTest, CallTest)
{
    LibruntimeOptions options{};
    options.functionExecuteCallback = [](const FunctionMeta &function, const libruntime::InvokeType invokeType,
                                         const std::vector<std::shared_ptr<DataObject>> &rawArgs,
                                         std::vector<std::shared_ptr<DataObject>> &returnValues) -> ErrorInfo {
        return ErrorInfo();
    };
    std::vector<std::string> objectsInDs;
    CallRequest req;
    req.set_requestid("fff87cc506e547d9");
    req.set_senderid("instance_id");
    req.set_iscreate(true);


    libruntime::MetaData metaData;
    auto result = invokeAdaptor->Call(req, metaData, options, objectsInDs);
    ASSERT_EQ(result.code(), ::common::ERR_NONE);

    auto pbArg1 = req.add_args();
    pbArg1->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    InvokeSpec invokeSpec;
    pbArg1->set_value(invokeSpec.BuildInvokeMetaData(*invokeAdaptor->librtConfig));
    auto pbArg = req.add_args();
    pbArg->set_type(common::Arg::OBJECT_REF);
    std::string objId("mock-123");
    pbArg->set_value(objId.c_str(), objId.size());
    auto result1 = invokeAdaptor->Call(req, metaData, options, objectsInDs);
    ASSERT_EQ(result1.code(), ::common::ERR_NONE);

    options.functionExecuteCallback = [](const FunctionMeta &function, const libruntime::InvokeType invokeType,
                                         const std::vector<std::shared_ptr<DataObject>> &rawArgs,
                                         std::vector<std::shared_ptr<DataObject>> &returnValues) -> ErrorInfo {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "test");
    };
    auto result2 = invokeAdaptor->Call(req, metaData, options, objectsInDs);
    ASSERT_EQ(result2.code(), ::common::ErrorCode::ERR_INNER_SYSTEM_ERROR);

    req.add_args();
    auto arg2 = req.add_args();
    arg2->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    arg2->set_value("{\"schedulerFuncK{\"schedulerFuncKey\":\"0/0-system-faasscheduler/$latest\",\"schedulerInstanceList\":[{\"instanceName\":\"abfe9e68-9221-4b97-8e85-87b5b5faf69c\",\"instanceId\":\"abfe9e68-9221-4b97-8e85-87b5b5faf69c\"},{\"instanceName\":\"2db4a71b-157c-4ec2-95d7-c70fccc85dfa\",\"instanceId\":\"abfe9e68-9221-4b97-8e85-87b5b5faf69c\"}]}");

    options.functionExecuteCallback = [](const FunctionMeta &function, const libruntime::InvokeType invokeType,
                                         const std::vector<std::shared_ptr<DataObject>> &rawArgs,
                                         std::vector<std::shared_ptr<DataObject>> &returnValues) -> ErrorInfo {
        return ErrorInfo(ErrorCode::ERR_OK, ModuleCode::RUNTIME, "test");
    };

    std::vector<SchedulerInstance> schedulerInstanceList;
    this->taskSubmitter = std::make_shared<MockTaskSubmitter>();
    EXPECT_CALL(*(this->taskSubmitter), UpdateFaaSSchedulerInfo(_, _))
        .WillOnce([=](std::string schedulerFuncKey, const std::vector<SchedulerInstance> &sInstanceList) {
       ASSERT_EQ(sInstanceList.size(), 2);
        return;
        });
    invokeAdaptor->taskSubmitter = this->taskSubmitter;
    libruntime::FunctionMeta functionMeta;
    functionMeta.set_apitype(libruntime::ApiType::Faas);
    metaData.set_allocated_functionmeta(&functionMeta);
    invokeAdaptor->Call(req, metaData, options, objectsInDs);
    metaData.release_functionmeta();
}

TEST_F(InvokeAdaptorTest, InitCallTest)
{
    CallRequest req;
    req.set_requestid("fff87cc506e547d9");
    req.set_senderid("instanceid");
    req.set_traceid("fff87cc506e547d9");
    req.set_function("function");
    libruntime::MetaData metaData;
    auto res = invokeAdaptor->InitCall(req, metaData);
    ASSERT_EQ(res.code(), ::common::ERR_NONE);

    auto pbArg1 = req.add_args();
    pbArg1->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    InvokeSpec invokeSpec;
    std::string meta;
    pbArg1->set_value(invokeSpec.BuildCreateMetaData(*invokeAdaptor->librtConfig, meta));
    auto res1 = invokeAdaptor->InitCall(req, metaData);
    ASSERT_EQ(res1.code(), ::common::ERR_NONE);
    libruntime::MetaConfig *config = metaData.mutable_config();
    config->add_codepaths("path");
    libConfig->libruntimeOptions.loadFunctionCallback = [](const std::vector<std::string> &codePaths) -> ErrorInfo {
        ErrorInfo err;
        err.SetErrorCode(ErrorCode::ERR_PARAM_INVALID);
        return err;
    };
    auto res2 = invokeAdaptor->InitCall(req, metaData);
    ASSERT_EQ(res2.code(), ::common::ERR_PARAM_INVALID);
}

TEST_F(InvokeAdaptorTest, CreateInstanceTest)
{
    auto cfg = LibruntimeConfig();
    auto invokeSpec = std::make_shared<InvokeSpec>();
    invokeSpec->requestId = "cae7c30c8d63f5ed00";
    invokeSpec->IncrementSeq();
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    invokeSpec->returnIds = returnObjs;
    invokeSpec->BuildInstanceCreateRequest(cfg);
    invokeAdaptor->CreateInstance(invokeSpec);
    auto [rawRequestId, seq] =YR::utility::IDGenerator::DecodeRawRequestId(invokeSpec->requestCreate.requestid());
    EXPECT_EQ(rawRequestId, invokeSpec->requestId);
    EXPECT_EQ(seq, 1);
}

TEST_F(InvokeAdaptorTest, InvokeGroupInstanceFunctionTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    InvokeOptions opts;
    InstanceRange range;
    opts.groupName = "invokeGroup";
    spec->opts = opts;
    spec->requestId = "reqId";
    ASSERT_NO_THROW(invokeAdaptor->InvokeInstanceFunction(spec));
}

TEST_F(InvokeAdaptorTest, CreateInstanceWithFunctionGroupTest)
{
    auto cfg = LibruntimeConfig();
    auto invokeSpec = std::make_shared<InvokeSpec>();
    invokeSpec->requestId = "cae7c30c8d63f5ed00";
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    invokeSpec->returnIds = returnObjs;
    auto opts = InvokeOptions();
    opts.groupName = "groupName";
    FunctionGroupOptions opt;
    opt.functionGroupSize = 8;
    opt.bundleSize = 2;
    opts.functionGroupOpts = opt;
    invokeSpec->opts = opts;
    invokeSpec->BuildInstanceCreateRequest(cfg);
    invokeAdaptor->CreateInstance(invokeSpec);
    ASSERT_TRUE(invokeAdaptor->groupManager->IsGroupExist("groupName"));
}

TEST_F(InvokeAdaptorTest, SubmitFunctionWithFunctionGroupTest)
{
    auto cfg = LibruntimeConfig();
    auto invokeSpec = std::make_shared<InvokeSpec>();
    invokeSpec->requestId = "cae7c30c8d63f5ed00";
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    invokeSpec->returnIds = returnObjs;
    invokeSpec->jobId = YR::utility::IDGenerator::GenApplicationId();
    auto opts = InvokeOptions();
    opts.groupName = "groupName";
    FunctionGroupOptions opt;
    opt.functionGroupSize = 8;
    opt.bundleSize = 2;
    opts.functionGroupOpts = opt;
    invokeSpec->opts = opts;
    invokeSpec->BuildInstanceCreateRequest(cfg);
    invokeAdaptor->SubmitFunction(invokeSpec);
    ASSERT_TRUE(invokeAdaptor->groupManager->IsGroupExist("groupName"));
}

TEST_F(InvokeAdaptorTest, SubmitFunctionWithAliasTest)
{
    auto cfg = LibruntimeConfig();
    auto invokeSpec = std::make_shared<InvokeSpec>();
    invokeSpec->requestId = "cae7c30c8d63f5ed00";
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    invokeSpec->returnIds = returnObjs;
    invokeSpec->jobId = YR::utility::IDGenerator::GenApplicationId();
    auto opts = InvokeOptions();
    opts.groupName = "groupName";
    FunctionGroupOptions opt;
    opt.functionGroupSize = 8;
    opt.bundleSize = 2;
    opts.functionGroupOpts = opt;

    std::unordered_map<std::string, std::string> params;
    params["userType"] = "VIP";
    params["age"] = "10";
    params["devType"] = "MATE40";

    opts.aliasParams = params;
    invokeSpec->functionMeta.functionId = "12345678901234561234567890123456/helloworld/myaliasv1";
    invokeSpec->opts = opts;
    invokeSpec->BuildInstanceCreateRequest(cfg);

    std::vector<AliasElement> g_aes_rule = {
        {
            .aliasUrn = "sn:cn:yrk:12345678901234561234567890123456:function:helloworld:myaliasv1",
            .functionUrn = "sn:cn:yrk:12345678901234561234567890123456:function:helloworld",
            .functionVersionUrn = "sn:cn:yrk:12345678901234561234567890123456:function:helloworld:$latest",
            .name = "myaliasv1",
            .functionVersion = "$latest",
            .revisionId = "20210617023315921",
            .description = "fake_description",
            .routingType = "rule",
            .routingRules =
                {
                    .ruleLogic = "and",
                    .rules =
                        {
                            "userType:=:VIP",
                            "age:<=:20",
                            "devType:in:P40,P50,MATE40",
                        },
                    .grayVersion = "sn:cn:yrk:12345678901234561234567890123456:function:helloworld:1",
                },
        },
    };
    invokeAdaptor->ar->UpdateAliasInfo(g_aes_rule);
    invokeAdaptor->SubmitFunction(invokeSpec);
    ASSERT_EQ(invokeSpec->functionMeta.functionId, "12345678901234561234567890123456/helloworld/1");
}

TEST_F(InvokeAdaptorTest, CreateResponseHandlerTest)
{
    CreateResponse resp;
    resp.set_instanceid("instanceId");
    resp.set_code(common::ERR_INSTANCE_DUPLICATED);
    auto spec = std::make_shared<InvokeSpec>();
    spec->invokeType = libruntime::InvokeType::CreateInstance;
    spec->requestId = "cae7c30c8d63f5ed00";
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->returnIds = returnObjs;
    invokeAdaptor->CreateResponseHandler(spec, resp);
    resp.set_code(common::ERR_NONE);
    invokeAdaptor->CreateResponseHandler(spec, resp);
    resp.set_code(common::ERR_USER_FUNCTION_EXCEPTION);
    invokeAdaptor->CreateResponseHandler(spec, resp);
    ASSERT_EQ(spec->opts.retryTimes, 0);
    InvokeOptions opts;
    opts.retryTimes = 1;
    spec->opts = opts;
    resp.set_code(common::ERR_RESOURCE_NOT_ENOUGH);
    invokeAdaptor->CreateResponseHandler(spec, resp);
    ASSERT_EQ(spec->opts.retryTimes, 0);
}

TEST_F(InvokeAdaptorTest, CreateGroupInstanceTest)
{
    auto cfg = LibruntimeConfig();
    auto spec = std::make_shared<InvokeSpec>();
    InvokeOptions opts;
    opts.groupName = "test";
    spec->opts = opts;
    spec->requestId = "reqId";
    invokeAdaptor->CreateInstance(spec);
    ASSERT_EQ(invokeAdaptor->groupManager->groupSpecs_.find(spec->opts.groupName) !=
                  invokeAdaptor->groupManager->groupSpecs_.end(),
              true);
}

TEST_F(InvokeAdaptorTest, CreateRangeInstanceTest)
{
    auto cfg = LibruntimeConfig();
    auto spec = std::make_shared<InvokeSpec>();
    InvokeOptions opts;
    InstanceRange range;
    opts.groupName = "group";
    range.max = 10;
    range.min = 2;
    range.step = 2;
    opts.instanceRange = range;
    spec->opts = opts;
    spec->requestId = "reqId";
    auto group = std::make_shared<RangeGroup>("group", "tenantId", range, invokeAdaptor->fsClient,
                                              invokeAdaptor->waitingObjectManager, invokeAdaptor->memStore,
                                              invokeAdaptor->invokeOrderMgr);
    invokeAdaptor->groupManager->AddGroup(group);
    invokeAdaptor->CreateInstance(spec);
    ASSERT_EQ(invokeAdaptor->groupManager->groupSpecs_.find(spec->opts.groupName) !=
                  invokeAdaptor->groupManager->groupSpecs_.end(),
              true);
}

TEST_F(InvokeAdaptorTest, InvokeNotifyHandlerTest)
{
    NotifyRequest req;
    ErrorInfo err;
    req.set_code(common::ERR_NONE);
    req.set_requestid("cae7c30c8d63f5ed00");
    auto spec1 = invokeAdaptor->requestManager->GetRequest(req.requestid());
    EXPECT_TRUE(spec1 == nullptr);

    auto spec = std::make_shared<InvokeSpec>();
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->invokeType = libruntime::InvokeType::InvokeFunction;
    spec->returnIds = returnObjs;
    spec->requestId = "cae7c30c8d63f5ed01";
    spec->functionMeta = YR::Libruntime::FunctionMeta{
        .apiType = libruntime::ApiType::Function, .functionId = "functionId", .name = "name", .ns = "ns"};
    invokeAdaptor->requestManager->PushRequest(spec);
    invokeAdaptor->InvokeNotifyHandler(req, err);
    auto spec2 = invokeAdaptor->requestManager->GetRequest(req.requestid());
    EXPECT_TRUE(spec2 == nullptr);
    invokeAdaptor->requestManager->RemoveRequest("cae7c30c8d63f5ed01");

    spec->requestId = "cae7c30c8d63f5ed00";
    invokeAdaptor->requestManager->PushRequest(spec);
    invokeAdaptor->InvokeNotifyHandler(req, err);
    auto spec3 = invokeAdaptor->requestManager->GetRequest(req.requestid());
    EXPECT_TRUE(spec3 == nullptr);

    req.set_code(common::ERR_INNER_COMMUNICATION);
    spec->opts.retryTimes = 1;
    invokeAdaptor->requestManager->PushRequest(spec);
    invokeAdaptor->InvokeNotifyHandler(req, err);
    auto spec4 = invokeAdaptor->requestManager->GetRequest(req.requestid());
    EXPECT_TRUE(spec4 != nullptr);

    req.set_code(common::ERR_USER_CODE_LOAD);
    err.SetIsTimeout(true);
    spec->invokeInstanceId = "invokeinstanceid";
    spec->opts.retryTimes = 0;
    invokeAdaptor->requestManager->PushRequest(spec);
    invokeAdaptor->InvokeNotifyHandler(req, err);
    auto spec5 = invokeAdaptor->requestManager->GetRequest(req.requestid());
    EXPECT_TRUE(spec5 != nullptr);
}

TEST_F(InvokeAdaptorTest, HandleReturnedObjectTest)
{
    NotifyRequest req;
    common::SmallObject *object = req.add_smallobjects();
    object->set_id("objId");
    object->set_value("bytes_data");
    auto spec = std::make_shared<InvokeSpec>();
    spec->returnIds = {DataObject("objId")};
    invokeAdaptor->librtConfig->inCluster = true;
    ASSERT_NO_THROW(invokeAdaptor->HandleReturnedObject(req, spec));
}

TEST_F(InvokeAdaptorTest, GroupCreateTest)
{
    GroupOpts opts;
    ASSERT_EQ(invokeAdaptor->GroupCreate("groupName", opts).OK(), false);
    ASSERT_EQ(invokeAdaptor->GroupCreate("groupName", opts).Code(), ErrorCode::ERR_PARAM_INVALID);
}

TEST_F(InvokeAdaptorTest, RangeCreateTest)
{
    InstanceRange range;
    ASSERT_EQ(invokeAdaptor->RangeCreate("groupName", range).OK(), false);
    ASSERT_EQ(invokeAdaptor->RangeCreate("groupName", range).Code(), ErrorCode::ERR_PARAM_INVALID);
}

TEST_F(InvokeAdaptorTest, ReleaseInstanceTest)
{
    InvokeOptions opts;
    ASSERT_EQ(invokeAdaptor->ReleaseInstance("leaseId", "", true, opts).OK(), true);
}

TEST_F(InvokeAdaptorTest, SubscribeAllTest)
{
    libruntime::FunctionMeta meta;
    invokeAdaptor->metaMap["insId"] = meta;
    ASSERT_NO_THROW(invokeAdaptor->SubscribeAll());
}

TEST_F(InvokeAdaptorTest, CreateNotifyHandlerTest)
{
    NotifyRequest req;
    req.set_code(common::ERR_NONE);
    req.set_requestid("cae7c30c8d63f5ed00");
    invokeAdaptor->CreateNotifyHandler(req);
    auto spec1 = invokeAdaptor->requestManager->GetRequest(req.requestid());
    EXPECT_TRUE(spec1 == nullptr);
    auto spec = std::make_shared<InvokeSpec>();
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->invokeType = libruntime::InvokeType::CreateInstance;
    spec->returnIds = returnObjs;
    spec->requestId = "cae7c30c8d63f5ed00";
    spec->functionMeta = YR::Libruntime::FunctionMeta{
        .apiType = libruntime::ApiType::Function, .functionId = "functionId", .name = "name", .ns = "ns"};
    invokeAdaptor->requestManager->PushRequest(spec);
    invokeAdaptor->CreateNotifyHandler(req);
    auto spec2 = invokeAdaptor->requestManager->GetRequest(req.requestid());
    EXPECT_TRUE(spec2 == nullptr);
    req.set_code(common::ERR_INSTANCE_NOT_FOUND);
    invokeAdaptor->requestManager->PushRequest(spec);
    invokeAdaptor->CreateNotifyHandler(req);
    auto spec3 = invokeAdaptor->requestManager->GetRequest(req.requestid());
    EXPECT_TRUE(spec3 == nullptr);
    InvokeOptions opts;
    opts.retryTimes = 1;
    spec->opts = opts;
    spec->seq = 2;
    req.set_requestid("cae7c30c8d63f5ed02");  // seq = 2
    req.set_code(common::ERR_RESOURCE_NOT_ENOUGH);
    invokeAdaptor->requestManager->PushRequest(spec);
    invokeAdaptor->CreateNotifyHandler(req);
    ASSERT_EQ(spec->opts.retryTimes, 0);  // retryable, consume 1 time
}

TEST_F(InvokeAdaptorTest, TestFinalize)
{
    EXPECT_NO_THROW(invokeAdaptor->Finalize(false));
    EXPECT_NE(invokeAdaptor->functionMasterClient_->work_, nullptr);
}

std::string g_alias = R"(
[{
    "aliasUrn": "fake_alias_urn",
    "functionUrn": "fake_function_urn",
    "functionVersionUrn": "fake_function_version_urn",
    "name": "fake_name",
    "functionVersion": "fake_function_version",
    "revisionId": "fake_revision_id",
    "description": "fake_description",
    "routingType": "rule",
    "routingRules": {
        "ruleLogic": "and",
        "rules": ["userType:=:VIP", "age:<=:20", "devType:in:P40,P50,MATE40"],
        "grayVersion": "fake_gray_version"
    },
    "routingconfig": [
        {
            "functionVersionUrn": "fake_function_version_urn_1",
            "weight": 50.0
        },
        {
            "functionVersionUrn": "fake_function_version_urn_2",
            "weight": 50.0
        }
    ]
}]
)";

TEST_F(InvokeAdaptorTest, ParseAliasInfoTest)
{
    SignalRequest req;
    req.set_payload(g_alias);
    std::vector<AliasElement> aliasInfo;
    invokeAdaptor->ParseAliasInfo(req, aliasInfo);
    ASSERT_EQ(aliasInfo.size(), 1);

    auto &alias = aliasInfo[0];
    ASSERT_EQ(alias.aliasUrn, "fake_alias_urn");
    ASSERT_EQ(alias.functionUrn, "fake_function_urn");
    ASSERT_EQ(alias.functionVersionUrn, "fake_function_version_urn");
    ASSERT_EQ(alias.name, "fake_name");
    ASSERT_EQ(alias.functionVersion, "fake_function_version");
    ASSERT_EQ(alias.revisionId, "fake_revision_id");
    ASSERT_EQ(alias.description, "fake_description");
    ASSERT_EQ(alias.routingType, "rule");
    ASSERT_EQ(alias.routingRules.ruleLogic, "and");
    ASSERT_EQ(alias.routingRules.rules.size(), 3);
    ASSERT_EQ(alias.routingRules.rules[0], "userType:=:VIP");
    ASSERT_EQ(alias.routingRules.rules[1], "age:<=:20");
    ASSERT_EQ(alias.routingRules.rules[2], "devType:in:P40,P50,MATE40");
    ASSERT_EQ(alias.routingRules.grayVersion, "fake_gray_version");
}

TEST_F(InvokeAdaptorTest, ExecSignalCallbackNullptrTest)
{
    libConfig->libruntimeOptions.signalCallback = nullptr;
    SignalRequest req;
    auto resp = invokeAdaptor->ExecSignalCallback(req);
    ASSERT_EQ(resp.code(), common::ErrorCode::ERR_NONE);
}

TEST_F(InvokeAdaptorTest, ExecSignalCallbackNonNullTest)
{
    std::promise<int> prom;
    std::future<int> fut = prom.get_future();
    libConfig->libruntimeOptions.signalCallback = [&prom](int sigNo, std::shared_ptr<Buffer> payload) -> ErrorInfo {
        prom.set_value(3);
        return ErrorInfo();
    };
    SignalRequest req;
    auto resp = invokeAdaptor->ExecSignalCallback(req);
    auto status = fut.wait_for(1s);
    ASSERT_EQ(status, std::future_status::ready);
    auto got = fut.get();
    ASSERT_EQ(got, 3);
}

TEST_F(InvokeAdaptorTest, ShutDownHandlerWithoutCbTest)
{
    int gracePeriodSec = 10;
    ShutdownRequest shutdownReq;
    shutdownReq.set_graceperiodsecond(gracePeriodSec);
    auto resp = invokeAdaptor->ShutdownHandler(shutdownReq);
    ASSERT_EQ(resp.code(), common::ErrorCode::ERR_NONE);
}

TEST_F(InvokeAdaptorTest, ShutDownHandlerTest)
{
    int gracePeriodSec = 10;
    ShutdownRequest shutdownReq;
    shutdownReq.set_graceperiodsecond(gracePeriodSec);
    libConfig->libruntimeOptions.shutdownCallback = [](uint64_t gracePeriodSeconds) -> ErrorInfo {
        return ErrorInfo();
    };
    auto resp = invokeAdaptor->ShutdownHandler(shutdownReq);
    ASSERT_EQ(resp.code(), common::ErrorCode::ERR_NONE);
}

TEST_F(InvokeAdaptorTest, SignalHandlerTest)
{
    SignalRequest req;
    auto spec = std::make_shared<InvokeSpec>();
    spec->requestId = "reqId";
    spec->invokeInstanceId = "invokeInstanceId";
    std::vector<DataObject> returnObjs{DataObject("returnID")};
    spec->returnIds = returnObjs;
    invokeAdaptor->requestManager->PushRequest(spec);

    req.set_signal(libruntime::Signal::Cancel);
    invokeAdaptor->SignalHandler(req);
    ASSERT_NE(invokeAdaptor->requestManager->GetRequest("reqId"), nullptr);

    req.set_signal(libruntime::Signal::ErasePendingThread);
    ASSERT_NO_THROW(invokeAdaptor->SignalHandler(req));

    req.set_signal(libruntime::Signal::UpdateAlias);
    std::vector<YR::Libruntime::AliasElement> aliasInfo = {YR::Libruntime::AliasElement()};
    json j = aliasInfo;
    req.set_payload(j.dump());
    auto response = invokeAdaptor->SignalHandler(req);
    ASSERT_EQ(response.code(), ::common::ErrorCode::ERR_NONE);

    req.set_signal(libruntime::Signal::Update);
    NotificationPayload notifyscription;
    InstanceTermination *termination = notifyscription.mutable_instancetermination();
    termination->set_instanceid("insId");
    std::string serializedPayload;
    notifyscription.SerializeToString(&serializedPayload);
    req.set_payload(serializedPayload);
    libruntime::FunctionMeta funcMeta;
    invokeAdaptor->metaMap["insId"] = funcMeta;
    response = invokeAdaptor->SignalHandler(req);
    ASSERT_EQ(invokeAdaptor->metaMap.size() == 0, true);

    req.set_signal(libruntime::Signal::UpdateSchedulerHash);
    req.set_payload("{\"schedulerFuncKey\":\"0/0-system-faasscheduler/$latest\",\"schedulerIDList\":null,\"schedulerInstanceList\":[{\"instanceName\":\"abfe9e68-9221-4b97-8e85-87b5b5faf69c\",\"instanceId\":\"abfe9e68-9221-4b97-8e85-87b5b5faf69c\"},{\"instanceName\":\"2db4a71b-157c-4ec2-95d7-c70fccc85dfa\",\"instanceId\":\"abfe9e68-9221-4b97-8e85-87b5b5faf69c\"}]}");
    response = invokeAdaptor->SignalHandler(req);
    ASSERT_EQ(response.code(), ::common::ErrorCode::ERR_NONE);

    req.set_signal(libruntime::Signal::Signal_INT_MIN_SENTINEL_DO_NOT_USE_);
    response = invokeAdaptor->SignalHandler(req);
    ASSERT_EQ(response.code(), ::common::ErrorCode::ERR_NONE);

    req.set_signal(libruntime::Signal::UpdateManager);
    response = invokeAdaptor->SignalHandler(req);
    ASSERT_EQ(response.code(), ::common::ErrorCode::ERR_NONE);

    req.set_signal(libruntime::Signal::QueryDsAddress);
    response = invokeAdaptor->SignalHandler(req);
    ASSERT_EQ(response.code(), ::common::ErrorCode::ERR_NONE);

    req.set_signal(libruntime::Signal::UpdateSchedulerHash);
    response = invokeAdaptor->SignalHandler(req);
    ASSERT_EQ(response.code(), ::common::ErrorCode::ERR_NONE);

    core_service::NotificationPayload notifyPayload;
    auto *functionmasterevent = notifyPayload.mutable_functionmasterevent();
    functionmasterevent->set_address("127.0.0.1:8080");
    std::string payload;
    notifyPayload.SerializeToString(&payload);
    req.set_signal(libruntime::Signal::Update);
    req.set_payload(payload);
    response = invokeAdaptor->SignalHandler(req);
    ASSERT_EQ(response.code(), ::common::ErrorCode::ERR_NONE);

    invokeAdaptor->isRunning = false;
    response = invokeAdaptor->SignalHandler(req);
    ASSERT_EQ(response.code(), ::common::ErrorCode::ERR_NONE);
    invokeAdaptor->isRunning = true;
}

TEST_F(InvokeAdaptorTest, CreateInstanceRawTest)
{
    CreateRequest req;
    ErrorInfo info;
    ErrorInfo info2;
    std::string instanceId;
    RawCallback cb = [&info, &info2, &instanceId](const ErrorInfo &err, std::shared_ptr<Buffer> resultRaw) {
        info.SetErrorCode(err.Code());
        if (!err.OK()) {
            return;
        }
        NotifyRequest notify;
        notify.ParseFromArray(resultRaw->ImmutableData(), resultRaw->GetSize());
        info2.SetErrorCode(static_cast<YR::Libruntime::ErrorCode>(notify.code()));
        instanceId = notify.instanceid();
    };
    size_t size = req.ByteSizeLong();
    auto reqRaw1 = std::make_shared<NativeBuffer>(size);
    req.SerializeToArray(reqRaw1->MutableData(), size);
    invokeAdaptor->CreateInstanceRaw(reqRaw1, cb);
    ASSERT_EQ(info.OK(), false);

    req.set_requestid("c51bbc05cf53e84304");
    size = req.ByteSizeLong();
    auto reqRaw2 = std::make_shared<NativeBuffer>(size);
    req.SerializeToArray(reqRaw2->MutableData(), size);
    invokeAdaptor->CreateInstanceRaw(reqRaw2, cb);
    ASSERT_EQ(info.OK(), true);
    ASSERT_EQ(info2.OK(), true);
    ASSERT_EQ(instanceId, "58f32000-0000-4000-8000-0ecfe00dd5e5");

    this->gwClient->isReqNormal = false;
    this->gwClient->callbackPromise = std::promise<int>();
    this->gwClient->callbackFuture = this->gwClient->callbackPromise.get_future();
    invokeAdaptor->CreateInstanceRaw(reqRaw2, cb);
    this->gwClient->callbackFuture.get();
    ASSERT_EQ(info.OK(), true);
    ASSERT_EQ(info2.OK(), false);
    ASSERT_EQ(instanceId, "58f32000-0000-4000-8000-0ecfe00dd5e5");
    this->gwClient->isReqNormal = true;
}

TEST_F(InvokeAdaptorTest, InvokeByInstanceIdRawTest)
{
    InvokeRequest req;
    ErrorInfo info;
    RawCallback cb = [&info](const ErrorInfo &err, std::shared_ptr<Buffer> resultRaw) {
        info.SetErrorCode(err.Code());
    };

    size_t size = req.ByteSizeLong();
    auto reqRaw1 = std::make_shared<NativeBuffer>(size);
    req.SerializeToArray(reqRaw1->MutableData(), size);
    invokeAdaptor->InvokeByInstanceIdRaw(reqRaw1, cb);
    ASSERT_EQ(info.OK(), false);

    req.set_requestid("c51bbc05cf53e84304");
    size = req.ByteSizeLong();
    auto reqRaw2 = std::make_shared<NativeBuffer>(size);
    req.SerializeToArray(reqRaw2->MutableData(), size);
    invokeAdaptor->InvokeByInstanceIdRaw(reqRaw2, cb);
    ASSERT_EQ(info.OK(), true);
}

TEST_F(InvokeAdaptorTest, KillRawTest)
{
    KillRequest req;
    req.set_instanceid("c51bbc05cf53e84304");
    size_t size = req.ByteSizeLong();
    auto reqRaw = std::make_shared<NativeBuffer>(size);
    req.SerializeToArray(reqRaw->MutableData(), size);
    ErrorInfo info;
    RawCallback cb = [&info](const ErrorInfo &err, std::shared_ptr<Buffer> resultRaw) {
        info.SetErrorCode(err.Code());
    };
    invokeAdaptor->KillRaw(reqRaw, cb);
    ASSERT_EQ(info.OK(), true);
}

TEST_F(InvokeAdaptorTest, ExecShutdownCallbackWithMinusDurationTest)
{
    libConfig->libruntimeOptions.shutdownCallback = [](uint64_t gracePeriodSeconds) -> ErrorInfo {
        return ErrorInfo(ErrorCode::ERR_OK, ModuleCode::RUNTIME, std::to_string(gracePeriodSeconds));
    };

    int gracePeriodSec = -10;
    auto err = invokeAdaptor->ExecShutdownCallback(gracePeriodSec);
    ASSERT_EQ(err.Msg(), "Execute user shutdown callback timeout");

    gracePeriodSec = 10;
    err = invokeAdaptor->ExecShutdownCallback(gracePeriodSec);
    ASSERT_EQ(err.Msg(), "10");
}

TEST_F(InvokeAdaptorTest, ParseFunctionGroupRunningInfoTest)
{
    CallRequest req;
    common::FunctionGroupRunningInfo runningInfo;
    auto res1 = ParseFunctionGroupRunningInfo(req, true, runningInfo);
    ASSERT_EQ(res1, true);
    auto res2 = ParseFunctionGroupRunningInfo(req, false, runningInfo);
    ASSERT_EQ(res2, true);
    common::FunctionGroupRunningInfo runningInfoInput;
    runningInfoInput.set_devicename("deviceName");
    std::string runningInfoStr;
    (void)google::protobuf::util::MessageToJsonString(runningInfoInput, &runningInfoStr);
    (*req.mutable_createoptions())["FUNCTION_GROUP_RUNNING_INFO"] = runningInfoStr;
    auto res4 = ParseFunctionGroupRunningInfo(req, true, runningInfo);
    ASSERT_EQ(res4, true);
}

TEST_F(InvokeAdaptorTest, ParseFaasControllerTest)
{
    CallRequest request;
    auto pbArg = request.add_args();
    pbArg->set_type(common::Arg::OBJECT_REF);
    std::string objId("mock-123");
    pbArg->set_value(objId.c_str(), objId.size());
    std::vector<std::shared_ptr<DataObject>> rawArgs;
    auto res = ParseFaasController(request, rawArgs, true);
    ASSERT_EQ(res, true);
}

TEST_F(InvokeAdaptorTest, InitHandlerTest)
{
    std::shared_ptr<CallMessageSpec> req = std::make_shared<CallMessageSpec>();
    req->Mutable().set_requestid("fff87cc506e547d9");
    req->Mutable().set_senderid("fff87cc506e547d9");
    req->Mutable().set_iscreate(true);
    int res = 0;
    libConfig->libruntimeOptions.loadFunctionCallback = [&res](const std::vector<std::string> &codePaths) -> ErrorInfo {
        res++;
        return ErrorInfo();
    };
    libConfig->libruntimeOptions.functionExecuteCallback =
        [&res](const FunctionMeta &function, const libruntime::InvokeType invokeType,
               const std::vector<std::shared_ptr<DataObject>> &rawArgs,
               std::vector<std::shared_ptr<DataObject>> &returnValues) -> ErrorInfo {
        res++;
        return ErrorInfo();
    };
    auto pbArg = req->Mutable().add_args();
    pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    InvokeSpec invokeSpec;
    invokeSpec.invokeType = libruntime::InvokeType::InvokeFunction;
    pbArg->set_value(invokeSpec.BuildInvokeMetaData(*invokeAdaptor->librtConfig));
    invokeAdaptor->InitHandler(req);
    ASSERT_EQ(res, 2);
}

TEST_F(InvokeAdaptorTest, PausedInitHandlerTest)
{
    std::shared_ptr<CallMessageSpec> req = std::make_shared<CallMessageSpec>();
    req->Mutable().set_requestid("fff87cc506e547d9");
    req->Mutable().set_senderid("fff87cc506e547d9");
    req->Mutable().set_iscreate(true);
    int res = 0;
    libConfig->libruntimeOptions.loadFunctionCallback = [&res](const std::vector<std::string> &codePaths) -> ErrorInfo {
        res++;
        return ErrorInfo();
    };
    libConfig->libruntimeOptions.functionExecuteCallback =
        [&res](const FunctionMeta &function, const libruntime::InvokeType invokeType,
               const std::vector<std::shared_ptr<DataObject>> &rawArgs,
               std::vector<std::shared_ptr<DataObject>> &returnValues) -> ErrorInfo {
        res++;
        return ErrorInfo();
    };
    auto pbArg = req->Mutable().add_args();
    pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    InvokeSpec invokeSpec;
    invokeSpec.invokeType = libruntime::InvokeType::InvokeFunction;
    pbArg->set_value(invokeSpec.BuildInvokeMetaData(*invokeAdaptor->librtConfig));

    auto createOpt = req->Mutable().mutable_createoptions();
    nlohmann::json debugJson;
    debugJson["enable"] = "true";
    createOpt->insert({"debug_config", debugJson.dump()});

    bool triggered = false;
    invokeAdaptor->setDebugBreakpoint_ = [&triggered]() { triggered = true; };
    invokeAdaptor->InitHandler(req);
    ASSERT_TRUE(triggered);
}

TEST_F(InvokeAdaptorTest, CallHandlerTest)
{
    std::shared_ptr<CallMessageSpec> req = std::make_shared<CallMessageSpec>();
    req->Mutable().set_requestid("fff87cc506e547d9");
    req->Mutable().set_senderid("fff87cc506e547d9");
    req->Mutable().set_iscreate(true);
    int res = 0;
    libConfig->libruntimeOptions.functionExecuteCallback =
        [&res](const FunctionMeta &function, const libruntime::InvokeType invokeType,
               const std::vector<std::shared_ptr<DataObject>> &rawArgs,
               std::vector<std::shared_ptr<DataObject>> &returnValues) -> ErrorInfo {
        res++;
        return ErrorInfo();
    };
    auto pbArg = req->Mutable().add_args();
    pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    InvokeSpec spec;
    spec.invokeType = libruntime::InvokeType::InvokeFunctionStateless;
    pbArg->set_value(spec.BuildInvokeMetaData(*invokeAdaptor->librtConfig));
    invokeAdaptor->CallHandler(req);
    ASSERT_EQ(res, 1);
}

TEST_F(InvokeAdaptorTest, CheckpointHandlerTest)
{
    CheckpointRequest req;
    req.set_checkpointid("checkpointId");
    auto resp1 = invokeAdaptor->CheckpointHandler(req);
    ASSERT_EQ(resp1.code(), ::common::ERR_NONE);

    libConfig->libruntimeOptions.checkpointCallback = [](const std::string &checkpointId,
                                                         std::shared_ptr<Buffer> &data) -> ErrorInfo {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, "test");
    };
    auto resp2 = invokeAdaptor->CheckpointHandler(req);
    ASSERT_EQ(resp2.code(), ::common::ERR_INNER_SYSTEM_ERROR);

    libConfig->libruntimeOptions.checkpointCallback = [](const std::string &checkpointId,
                                                         std::shared_ptr<Buffer> &data) -> ErrorInfo {
        std::string str = "test";
        data = std::make_shared<NativeBuffer>(&str, str.length());
        return ErrorInfo();
    };
    auto resp3 = invokeAdaptor->CheckpointHandler(req);
    ASSERT_EQ(resp3.code(), ::common::ERR_NONE);
}

TEST_F(InvokeAdaptorTest, RecoverHandlerTest)
{
    RecoverRequest req;
    auto resp1 = invokeAdaptor->RecoverHandler(req);
    ASSERT_EQ(resp1.code(), ::common::ERR_NONE);

    libConfig->libruntimeOptions.recoverCallback = [](std::shared_ptr<Buffer> data) -> ErrorInfo {
        return ErrorInfo();
    };
    auto resp2 = invokeAdaptor->RecoverHandler(req);
    ASSERT_EQ(resp2.code(), ::common::ERR_USER_FUNCTION_EXCEPTION);

    std::string str = "test";
    std::shared_ptr<Buffer> data = std::make_shared<NativeBuffer>(&str, str.length());
    auto bufInstanceSize = data->GetSize();
    auto headerSize = sizeof(size_t);
    libruntime::MetaConfig metaConfig;
    libConfig->BuildMetaConfig(metaConfig);
    std::string serializedMetaConfig = metaConfig.SerializeAsString();
    auto bufMetaSize = serializedMetaConfig.size();
    size_t stateSize = headerSize + bufInstanceSize + bufMetaSize;
    std::string *state = req.mutable_state();
    state->reserve(stateSize);
    state->append(reinterpret_cast<const char *>(&bufInstanceSize), headerSize);
    state->append(reinterpret_cast<const char *>(data->ImmutableData()), bufInstanceSize);
    state->append(serializedMetaConfig.c_str(), bufMetaSize);
    req.set_state(state->c_str(), state->size());
    auto resp3 = invokeAdaptor->RecoverHandler(req);
    ASSERT_EQ(resp3.code(), ::common::ERR_NONE);

    libConfig->libruntimeOptions.recoverCallback = [](std::shared_ptr<Buffer> data) -> ErrorInfo {
        return ErrorInfo(ErrorCode::ERR_INSTANCE_SUB_HEALTH, ModuleCode::RUNTIME, "test");
    };
    auto resp4 = invokeAdaptor->RecoverHandler(req);
    ASSERT_EQ(resp4.code(), ::common::ERR_INSTANCE_SUB_HEALTH);

    libConfig->libruntimeOptions.loadFunctionCallback = [](const std::vector<std::string> &codePaths) -> ErrorInfo {
        return ErrorInfo(ErrorCode::ERR_INSTANCE_DUPLICATED, ModuleCode::RUNTIME, "test");
    };
    auto resp5 = invokeAdaptor->RecoverHandler(req);
    ASSERT_EQ(resp5.code(), ::common::ERR_INSTANCE_DUPLICATED);
}

TEST_F(InvokeAdaptorTest, HeartbeatHandlerTest)
{
    HeartbeatRequest req;
    auto resp1 = invokeAdaptor->HeartbeatHandler(req);
    ASSERT_EQ(resp1.code(), ::common::ERR_NONE);

    libConfig->libruntimeOptions.healthCheckCallback = []() -> ErrorInfo {
        return ErrorInfo(ErrorCode::ERR_HEALTH_CHECK_HEALTHY, ModuleCode::RUNTIME, "test");
    };
    auto resp2 = invokeAdaptor->HeartbeatHandler(req);
    ASSERT_EQ(resp2.code(), common::HealthCheckCode::HEALTHY);

    libConfig->libruntimeOptions.healthCheckCallback = []() -> ErrorInfo {
        return ErrorInfo(ErrorCode::ERR_HEALTH_CHECK_FAILED, ModuleCode::RUNTIME, "test");
    };
    auto resp3 = invokeAdaptor->HeartbeatHandler(req);
    ASSERT_EQ(resp3.code(), common::HealthCheckCode::HEALTH_CHECK_FAILED);

    libConfig->libruntimeOptions.healthCheckCallback = []() -> ErrorInfo {
        return ErrorInfo(ErrorCode::ERR_HEALTH_CHECK_SUBHEALTH, ModuleCode::RUNTIME, "test");
    };
    auto resp4 = invokeAdaptor->HeartbeatHandler(req);
    ASSERT_EQ(resp4.code(), common::HealthCheckCode::SUB_HEALTH);
}

TEST_F(InvokeAdaptorTest, RetryInvokeInstanceFunctionTest)
{
    auto spec = std::make_shared<InvokeSpec>();
    spec->requestId = "reqId";
    InvokeOptions opts;
    opts.retryTimes = 1;
    spec->BuildInstanceInvokeRequest(*libConfig);
    invokeAdaptor->RetryInvokeInstanceFunction(spec, true);
    ASSERT_EQ(spec->opts.retryTimes, 0);
}

TEST_F(InvokeAdaptorTest, SaveStateTest)
{
    auto err1 = invokeAdaptor->SaveState(nullptr, -2);
    ASSERT_EQ(err1.Code(), ErrorCode::ERR_PARAM_INVALID);

    auto err2 = invokeAdaptor->SaveState(nullptr, 1);
    ASSERT_EQ(err2.Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);

    std::string str = "test";
    std::shared_ptr<Buffer> data = std::make_shared<NativeBuffer>(&str, str.length());
    auto err3 = invokeAdaptor->SaveState(data, 1);
    ASSERT_EQ(err3.Code(), ErrorCode::ERR_INIT_CONNECTION_FAILED);
}

TEST_F(InvokeAdaptorTest, LoadStateTest)
{
    std::string str = "test";
    std::shared_ptr<Buffer> data = std::make_shared<NativeBuffer>(&str, str.length());
    auto err1 = invokeAdaptor->LoadState(data, -2);
    ASSERT_EQ(err1.Code(), ErrorCode::ERR_PARAM_INVALID);

    auto err2 = invokeAdaptor->LoadState(data, 1);
    ASSERT_EQ(err2.Code(), ErrorCode::ERR_INIT_CONNECTION_FAILED);
}

TEST_F(InvokeAdaptorTest, GetInstanceIdsTest)
{
    auto [vec1, err1] = invokeAdaptor->GetInstanceIds("objid", "groupname");
    ASSERT_EQ(err1.Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);
    GroupOpts opts;
    auto fsClient = std::make_shared<MockFsIntf>();
    auto group = std::make_shared<NamedGroup>("groupname", opts);
    invokeAdaptor->groupManager->AddGroup(group);
    auto [vec2, err2] = invokeAdaptor->GetInstanceIds("objid", "groupname");
    ASSERT_EQ(vec2.size() == 1, true);
}

TEST_F(InvokeAdaptorTest, AcquireInstanceTest)
{
    auto invokeAdaptor = std::make_shared<InvokeAdaptor>();
    invokeAdaptor->taskSubmitter = std::make_shared<MockTaskSubmitter>();
    invokeAdaptor->requestManager = std::make_shared<RequestManager>();
    FunctionMeta meta;
    meta.name = "name";
    InvokeOptions opts;
    opts.schedulerInstanceIds.push_back("id");
    opts.traceId = "traceId";
    auto [res, err] = invokeAdaptor->AcquireInstance("stateId", meta, opts);
    ASSERT_EQ(err.OK(), true);
}

TEST_F(InvokeAdaptorTest, UpdateSchdulerInfoTest)
{
    auto invokeAdaptor = std::make_shared<InvokeAdaptor>();
    invokeAdaptor->taskSubmitter = std::make_shared<MockTaskSubmitter>();
    ASSERT_NO_THROW(invokeAdaptor->UpdateSchdulerInfo("schedulerName", "schedulerId", "ADD"));
}

TEST_F(InvokeAdaptorTest, AdaptorGetInsTest)
{
    this->gwClient->isGetInstance = true;
    std::string name = "name";
    std::string ns = "ns";
    auto [res, err] = invokeAdaptor->GetInstance(name, ns, 60);
    ASSERT_EQ(res.className, "classname");
    auto [res1, err1] = invokeAdaptor->GetInstance(name, ns, 60);
    ASSERT_EQ(err1.OK(), true);
    ASSERT_EQ(invokeAdaptor->metaMap.size() == 1, true);

    libruntime::FunctionMeta meta;
    meta.set_name("name");
    invokeAdaptor->librtConfig->funcMeta = meta;
    auto [res2, err2] = invokeAdaptor->GetInstance(name, ns, 60);
    ASSERT_EQ(err2.OK(), true);
    meta.set_ns("ns");
    invokeAdaptor->librtConfig->funcMeta = meta;
    auto [res3, err3] = invokeAdaptor->GetInstance(name, ns, 60);
    ASSERT_EQ(err3.OK(), false);
    ASSERT_EQ(err3.Code(), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);
    this->gwClient->isGetInstance = false;
}

TEST_F(InvokeAdaptorTest, UpdateAndSubcribeInsStatusTest)
{
    libruntime::FunctionMeta funcMeta;
    funcMeta.set_classname("class_name");
    this->gwClient->isReqNormal = false;
    invokeAdaptor->UpdateAndSubcribeInsStatus("insId", funcMeta);
    this->gwClient->killCallbackFuture.get();
    ASSERT_EQ(invokeAdaptor->metaMap.size() == 0, true);
    this->gwClient->isReqNormal = true;
}

TEST_F(InvokeAdaptorTest, RemoveInsMetaInfoTest)
{
    libruntime::FunctionMeta funcMeta;
    invokeAdaptor->metaMap["insId"] = funcMeta;
    invokeAdaptor->RemoveInsMetaInfo("insId");
    ASSERT_EQ(invokeAdaptor->metaMap.size() == 0, true);
}

TEST_F(InvokeAdaptorTest, CallTimerTest)
{
    libruntime::FunctionMeta funcMeta;
    std::string reqId = "reqId";
    std::string insId = "insId";
    invokeAdaptor->CreateCallTimer(reqId, insId, 0);
    ASSERT_EQ(invokeAdaptor->callTimeoutTimerMap_.size() == 0, true);

    invokeAdaptor->CreateCallTimer(reqId, insId, 100);
    ASSERT_EQ(invokeAdaptor->callTimeoutTimerMap_.size() == 1, true);
    invokeAdaptor->EraseCallTimer(reqId);
    ASSERT_EQ(invokeAdaptor->callTimeoutTimerMap_.size() == 0, true);

    invokeAdaptor->CreateCallTimer(reqId, insId, 1);
    auto called = std::make_shared<std::promise<bool>>();
    EXPECT_CALL(*this->gwClient, ReturnCallResult)
        .WillOnce(::testing::Invoke(
            [called](const std::shared_ptr<CallResultMessageSpec> result, bool isCreate, CallResultCallBack callback) {
                called->set_value(true);
            }));
    ASSERT_EQ(called->get_future().get(), true);
}

TEST_F(InvokeAdaptorTest, MetricsTest)
{
    const std::string FileExporterJsonStr = R"(
{
    "backends": [
        {
            "immediatelyExport": {
                "name": "EDA",
                "enable": true,
                "exporters": [
                    {
                        "fileExporter": {
                            "enable": true,
                            "fileDir": "/tmp/",
                            "rolling": {
                                "enable": true,
                                "maxFiles": 3,
                                "maxSize": 10000
                            },
                            "contentType": "STANDARD"
                        }
                    }
                ]
            }
        }
    ]
}
    )";
    // empty metric -> expect init not to be called
    invokeAdaptor->InitMetricsAdaptor(true);
    ASSERT_EQ(MetricsAdaptor::GetInstance()->userEnable_, false);
    Config::Instance().METRICS_CONFIG_ = "invalid json";
    invokeAdaptor->InitMetricsAdaptor(true);
    ASSERT_EQ(MetricsAdaptor::GetInstance()->userEnable_, false);
    // 创建输出文件流
    std::string file = "./metric.json";
    std::ofstream outFile(file);
    if (!outFile) {
        return;
    }
    outFile << FileExporterJsonStr;
    outFile.close();
    Config::Instance().METRICS_CONFIG_FILE_ = file;
    Config::Instance().METRICS_CONFIG_ = "";
    // valid metric file -> expected successful called
    invokeAdaptor->InitMetricsAdaptor(true);
    ASSERT_EQ(MetricsAdaptor::GetInstance()->userEnable_, true);
    Config::Instance().ENABLE_METRICS_ = true;
    invokeAdaptor->ReportMetrics("request", "trace", 1);
    std::remove(file.c_str());
    Config::Instance().METRICS_CONFIG_FILE_ = "";
    Config::Instance().ENABLE_METRICS_ = false;
}
}  // namespace test
}  // namespace YR
