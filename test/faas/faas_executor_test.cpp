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
#include <filesystem>
#include <stdexcept>
#include <string>

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include "Constant.h"
#include "FunctionError.h"
#include "common/mock_libruntime.h"
#include "api/cpp/src/utils/utils.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/utility/id_generator.h"
#define private public
#include "api/cpp/src/faas/context_impl.h"
#include "api/cpp/src/faas/faas_executor.h"
#include "api/cpp/src/faas/register_runtime_handler.h"

namespace YR {
namespace test {
using namespace testing;
using namespace YR::internal;
using namespace Function;
using namespace YR::utility;
namespace fs = std::filesystem;
class FaasExecutorTest : public testing::Test {
public:
    FaasExecutorTest(){};
    ~FaasExecutorTest(){};

    void SetUp() override
    {
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
        auto lc = std::make_shared<YR::Libruntime::LibruntimeConfig>();
        lc->jobId = YR::utility::IDGenerator::GenApplicationId();
        auto clientsMgr = std::make_shared<YR::Libruntime::ClientsManager>();
        auto metricsAdaptor = std::make_shared<YR::Libruntime::MetricsAdaptor>();
        auto sec = std::make_shared<YR::Libruntime::Security>();
        auto socketClient = std::make_shared<YR::Libruntime::DomainSocketClient>("/home/snuser/socket/runtime.sock");
        lr = std::make_shared<YR::Libruntime::MockLibruntime>(lc, clientsMgr, metricsAdaptor, sec, socketClient);
        YR::Libruntime::LibruntimeManager::Instance().SetLibRuntime(lr);
        exec_ = std::make_shared<FaasExecutor>();
    }

    void TearDown() override
    {
        YR::Libruntime::LibruntimeManager::Instance().Finalize();
        lr.reset();
        exec_.reset();
    }

private:
    std::shared_ptr<YR::Libruntime::MockLibruntime> lr;
    std::shared_ptr<FaasExecutor> exec_;
};

TEST_F(FaasExecutorTest, LoadFunctionsTest)
{
    auto err = exec_->LoadFunctions({});
    ASSERT_TRUE(err.OK()) << err.Msg();
}

TEST_F(FaasExecutorTest, SignalTest)
{
    auto payload = std::make_shared<YR::Libruntime::NativeBuffer>(1);
    auto err = exec_->Signal(65, payload);
    ASSERT_TRUE(err.OK()) << err.Msg();
}

std::string contextMetaStr = R"(
{"funcMetaData":
    {
        "layers":[],"name":"0@test@cpp","description":"this is my app",
        "functionUrn":"sn:cn:yrk:default:function:0@test@cpp",
        "tenantId":"default","tags":null,"functionUpdateTime":"",
        "functionVersionUrn":"sn:cn:yrk:default:function:0@test@cpp:latest",
        "revisionId":"20241017135254074","codeSize":0,"codeSha512":"","handler":"bin/start.sh",
        "runtime":"posix-custom-runtime","timeout":600,"version":"latest","deadLetterConfig":"",
        "businessId":"yrk","functionType":"","func_id":"","func_name":"cpp","domain_id":"",
        "project_name":"","service":"test","poolLabel":"","dependencies":"",
        "enable_cloud_debug":"","isStatefulFunction":false,"isBridgeFunction":false,
        "isStreamEnable":false,"type":"","enable_auth_in_header":false,"dns_domain_cfg":null,"vpcTriggerImage":"",
        "stateConfig":{"lifeCycle":""}},
        "s3MetaData":{"appId":"","bucketId":"","objectId":"","bucketUrl":"","code_type":"","code_url":"","code_filename":"",
        "func_code":{"file":"","link":""}},
        "codeMetaData":{"sha512":"","storage_type":"s3","code_path":"",
        "appId":"","bucketId":"test","objectId":"cpp.zip","bucketUrl":"https://127.0.0,1:30110","code_type":"",
        "code_url":"","code_filename":"","func_code":{"file":"","link":""}},
        "envMetaData":{"environment":"fa3f8de0d7ac57ed34babf52:db67e7193a6d9ee35a15f92dc8de08f16a58d7810806921ad404bdccde",
        "encrypted_user_data":"",
        "envKey":"d79a80e56bd11a37c35ea5e7:d31ee89d52ab7f36094faa015a599149f1d27e295010bd328b319d3316c80bcfdeeefeffe1a3b622fb3c1ea8ffa86511dc7682086c711dc33bdb67d3c21dc93dbfdf487e90010b1905c1d168c1fe57c3",
        "cryptoAlgorithm":"GCM"},
        "stsMetaData":{"enableSts":false},
        "resourceMetaData":{"cpu":1000,"memory":1024,"gpu_memory":0,"enable_dynamic_memory":false,"customResources":"",
        "enable_tmp_expansion":false,"ephemeral_storage":0,"CustomResourcesSpec":""},
        "instanceMetaData":{"maxInstance":100,"minInstance":0,"concurrentNum":100,"diskLimit":0,"instanceType":"",
        "schedulePolicy":"concurrency","scalePolicy":"","idleMode":false},
        "extendedMetaData":{"image_name":"","role":{"xrole":"","app_xrole":""},
        "func_vpc":null,"endpoint_tenant_vpc":null,"mount_config":null,
        "strategy_config":{"concurrency":0},"extend_config":"",
        "initializer":{"initializer_handler":"","initializer_timeout":0},
        "heartbeat":{"heartbeat_handler":""},"enterprise_project_id":"",
        "log_tank_service":{"logGroupId":"","logStreamId":""},
        "tracing_config":{"tracing_ak":"","tracing_sk":"","project_name":""},
        "custom_container_config":{"control_path":"","image":"","command":null,"args":null,
        "working_dir":"","uid":0,"gid":0},"async_config_loaded":false,"restore_hook":{},
        "network_controller":{"disable_public_network":false,"trigger_access_vpcs":null},
        "user_agency":{"accessKey":"","secretKey":"","token":"","securityAk":"","securitySk":"",
        "securityToken":""},"custom_filebeat_config":{"sidecarConfigInfo":null,"cpu":0,"memory":0,
        "version":"","imageAddress":""},"custom_health_check":{"timeoutSeconds":0,"periodSeconds":0,
        "failureThreshold":0},"dynamic_config":{"enabled":false,"update_time":"","config_content":null},
        "runtime_graceful_shutdown":{"maxShutdownTimeout":0},"pre_stop":{"pre_stop_handler":"",
        "pre_stop_timeout":0},"rasp_config":{"init-image":"","rasp-image":"","rasp-server-ip":"","rasp-server-port":""}}}
)";

std::string createParamStr = R"(
{"instanceLabel": "aaaaa"}
)";

std::string deleteDecryptStr = R"(
    {
        "environment": "{\"key\":\"value\"}",
        "encrypted_user_data": "{\"aaa\":\"bbb\"}"
    }
)";

TEST_F(FaasExecutorTest, ExecuteFunctionSuccessFullyTest)
{
    auto handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterHandler([](const std::string &event, Context &context) -> std::string {
        Function::FunctionLogger logger = context.GetLogger();
        logger.setLevel("INFO");
        logger.Info("hello cpp %s ", "user info log");
        logger.Error("hello cpp %s ", "user error log");
        logger.Warn("hello cpp %s ", "user warn log");
        logger.Debug("hello cpp %s ", "user debug log");

        EXPECT_EQ(context.GetAccessKey(), "");
        EXPECT_EQ(context.GetSecretKey(), "");
        EXPECT_EQ(context.GetSecurityAccessKey(), "");
        EXPECT_EQ(context.GetSecuritySecretKey(), "");
        EXPECT_EQ(context.GetToken(), "");
        EXPECT_EQ(context.GetAlias(), "");
        EXPECT_EQ(context.GetTraceId(), "traceid");
        EXPECT_EQ(context.GetInvokeId(), "initializer");
        EXPECT_EQ(context.GetState(), "");
        EXPECT_EQ(context.GetInstanceId(), "");
        EXPECT_EQ(context.GetInvokeProperty(), "");
        EXPECT_EQ(context.GetRequestID(), "traceid");
        EXPECT_EQ(context.GetUserData("aaa"), "bbb");
        EXPECT_EQ(context.GetFunctionName(), "cpp");
        EXPECT_EQ(context.GetInstanceLabel(), "aaaaa");
        EXPECT_EQ(context.GetRemainingTimeInMilliSeconds(), 0);
        EXPECT_EQ(context.GetRunningTimeInSeconds(), 0);
        EXPECT_EQ(context.GetVersion(), "latest");
        EXPECT_EQ(context.GetMemorySize(), 1024);
        EXPECT_EQ(context.GetCPUNumber(), 1000);
        EXPECT_EQ(context.GetProjectID(), "default");
        EXPECT_EQ(context.GetPackage(), "test");

        return event;
    });
    std::string traceId = "traceid";
    handlerPtr->RegisterInitializerFunction([](Context &context) {});
    SetRuntimeHandler(std::move(handlerPtr));
    YR::SetEnv("key", "1");
    YR::SetEnv("ENV_DELEGATE_DECRYPT", deleteDecryptStr);

    YR::Libruntime::ErrorInfo err;
    YR::Libruntime::FunctionMeta function;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> rawArgs;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> returnObjects;
    returnObjects.push_back(std::make_shared<YR::Libruntime::DataObject>());

    auto contextMetaObj = std::make_shared<YR::Libruntime::DataObject>(0, contextMetaStr.size());
    contextMetaObj->data->MemoryCopy(contextMetaStr.data(), contextMetaStr.size());
    rawArgs.push_back(contextMetaObj);
    auto createParamObj = std::make_shared<YR::Libruntime::DataObject>(0, createParamStr.size());
    createParamObj->data->MemoryCopy(createParamStr.data(), createParamStr.size());
    rawArgs.push_back(createParamObj);

    err = exec_->ExecuteFunction(function, libruntime::InvokeType::CreateInstance, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    rawArgs.pop_back();

    std::string eventStr = "{\"body\":\"event\", \"header\": {\"X-Trace-Id\":\"traceid\"}}";
    auto eventObj = std::make_shared<YR::Libruntime::DataObject>(0, eventStr.size());
    eventObj->data->MemoryCopy(eventStr.data(), eventStr.size());
    rawArgs.push_back(eventObj);

    auto traceIdObj = std::make_shared<YR::Libruntime::DataObject>(0, traceId.size());
    traceIdObj->data->MemoryCopy(traceId.data(), traceId.size());
    rawArgs.push_back(traceIdObj);

    auto returnObj2 = std::make_shared<Libruntime::DataObject>(0, 200);
    EXPECT_CALL(*lr.get(), AllocReturnObject(Matcher<std::shared_ptr<YR::Libruntime::DataObject> &>(_), _, _, _, _))
        .WillOnce(DoAll(SetArgReferee<0>(returnObj2), Return(YR::Libruntime::ErrorInfo())));
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::InvokeFunction, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    auto result = std::string(static_cast<const char *>(returnObjects[0]->data->ImmutableData()),
                              returnObjects[0]->data->GetSize());
    ASSERT_TRUE(result.find("event")) << result;

    ASSERT_EQ(YR::GetEnv("key"), "value");
}

TEST_F(FaasExecutorTest, ExecuteFunctionFailedTest)
{
    YR::Libruntime::ErrorInfo err;
    YR::Libruntime::FunctionMeta function;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> rawArgs;
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> returnObjects;
    returnObjects.push_back(std::make_shared<YR::Libruntime::DataObject>());
    auto contextMetaObj = std::make_shared<YR::Libruntime::DataObject>(0, contextMetaStr.size());
    contextMetaObj->data->MemoryCopy(contextMetaStr.data(), contextMetaStr.size());
    rawArgs.push_back(contextMetaObj);
    auto returnObj = std::make_shared<Libruntime::DataObject>(0, 200);
    EXPECT_CALL(*lr.get(), AllocReturnObject(Matcher<std::shared_ptr<YR::Libruntime::DataObject> &>(_), _, _, _, _))
        .WillRepeatedly(DoAll(SetArgReferee<0>(returnObj), Return(YR::Libruntime::ErrorInfo())));

    auto handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    SetRuntimeHandler(std::move(handlerPtr));

    auto eventObj = std::make_shared<YR::Libruntime::DataObject>(0, 0);
    rawArgs.push_back(eventObj);
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::InvokeFunction, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    auto result = std::string(static_cast<const char *>(returnObjects[0]->data->ImmutableData()),
                              returnObjects[0]->data->GetSize());
    ASSERT_TRUE(result.find("call req is empty"));

    std::string eventStr = "{}";
    eventObj = std::make_shared<YR::Libruntime::DataObject>(0, eventStr.size());
    eventObj->data->MemoryCopy(eventStr.data(), eventStr.size());
    rawArgs[1] = eventObj;
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::InvokeFunction, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    result = std::string(static_cast<const char *>(returnObjects[0]->data->ImmutableData()),
                         returnObjects[0]->data->GetSize());
    ASSERT_TRUE(result.find("can not find body"));

    eventStr = "{\"body\": 1}";
    eventObj = std::make_shared<YR::Libruntime::DataObject>(0, eventStr.size());
    eventObj->data->MemoryCopy(eventStr.data(), eventStr.size());
    rawArgs[1] = eventObj;
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::InvokeFunction, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    result = std::string(static_cast<const char *>(returnObjects[0]->data->ImmutableData()),
                         returnObjects[0]->data->GetSize());
    ASSERT_TRUE(result.find("event type is not string"));

    eventStr = "{\"body\": \"event\"}";
    eventObj = std::make_shared<YR::Libruntime::DataObject>(0, eventStr.size());
    eventObj->data->MemoryCopy(eventStr.data(), eventStr.size());
    rawArgs[1] = eventObj;
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::InvokeFunction, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    result = std::string(static_cast<const char *>(returnObjects[0]->data->ImmutableData()),
                         returnObjects[0]->data->GetSize());
    ASSERT_TRUE(result.find("can not call handlerequest before initialize"));

    contextMetaObj = std::make_shared<YR::Libruntime::DataObject>(0, 0);
    rawArgs[0] = contextMetaObj;
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::CreateInstance, rawArgs, returnObjects);
    ASSERT_FALSE(err.OK()) << err.Msg();

    contextMetaObj = std::make_shared<YR::Libruntime::DataObject>(0, contextMetaStr.size());
    contextMetaObj->data->MemoryCopy(contextMetaStr.data(), contextMetaStr.size());
    rawArgs[0] = contextMetaObj;
    handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterInitializerFunction(
        [](Context &context) { throw FunctionError(ErrorCode::FUNCTION_EXCEPTION, "function error"); });
    SetRuntimeHandler(std::move(handlerPtr));
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::CreateInstance, rawArgs, returnObjects);
    ASSERT_FALSE(err.OK()) << err.Msg();
    ASSERT_TRUE(result.find("function error"));

    handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterInitializerFunction([](Context &context) { throw std::runtime_error("runtime error"); });
    SetRuntimeHandler(std::move(handlerPtr));
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::CreateInstance, rawArgs, returnObjects);
    ASSERT_FALSE(err.OK()) << err.Msg();

    handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    SetRuntimeHandler(std::move(handlerPtr));
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::InvokeFunction, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    result = std::string(static_cast<const char *>(returnObjects[0]->data->ImmutableData()),
                         returnObjects[0]->data->GetSize());
    ASSERT_TRUE(result.find("undefined HandleRequest"));

    handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterHandler([](const std::string &event, Context &context) -> std::string {
        throw FunctionError(ErrorCode::FUNCTION_EXCEPTION, "function error");
    });
    SetRuntimeHandler(std::move(handlerPtr));
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::InvokeFunction, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    result = std::string(static_cast<const char *>(returnObjects[0]->data->ImmutableData()),
                         returnObjects[0]->data->GetSize());
    ASSERT_TRUE(result.find("function error"));

    handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterHandler(
        [](const std::string &event, Context &context) -> std::string { throw std::runtime_error("runtime error"); });
    SetRuntimeHandler(std::move(handlerPtr));
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::InvokeFunction, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    result = std::string(static_cast<const char *>(returnObjects[0]->data->ImmutableData()),
                         returnObjects[0]->data->GetSize());
    ASSERT_TRUE(result.find("runtime error"));

    handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterHandler([](const std::string &event, Context &context) -> std::string {
        std::string a;
        a.resize(6 * 1024 * 1024 + 1, '0');
        return a;
    });
    SetRuntimeHandler(std::move(handlerPtr));
    std::string expectResult =
        "{\"body\":\"function result size: 6291457, exceed limit(6291456)\",\"innerCode\":\"4004\"}";
    err = exec_->ExecuteFunction(function, libruntime::InvokeType::InvokeFunction, rawArgs, returnObjects);
    ASSERT_TRUE(err.OK()) << err.Msg();
    result = std::string(static_cast<const char *>(returnObjects[0]->data->ImmutableData()), expectResult.size());
    auto resultJson = nlohmann::json::parse(result);
    std::string value = resultJson["body"];
    ASSERT_TRUE(value.find("exceed limit("));
}

TEST_F(FaasExecutorTest, ExecuteShutdownFunctionTest)
{
    YR::Libruntime::ErrorInfo err;
    auto handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    SetRuntimeHandler(std::move(handlerPtr));
    err = exec_->ExecuteShutdownFunction(1);
    ASSERT_FALSE(err.OK());
    ASSERT_TRUE(err.Msg().find("can not call prestop before initialize"));

    exec_->contextEnv_ = std::make_shared<ContextEnv>();
    std::unordered_map<std::string, std::string> params;
    exec_->contextInvokeParams_ = std::make_shared<ContextInvokeParams>(params);
    err = exec_->ExecuteShutdownFunction(1);
    ASSERT_TRUE(err.OK());

    handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterPreStopFunction(
        [](Context &context) { throw FunctionError(ErrorCode::FUNCTION_EXCEPTION, "function error"); });
    SetRuntimeHandler(std::move(handlerPtr));
    err = exec_->ExecuteShutdownFunction(1);
    ASSERT_FALSE(err.OK());
    ASSERT_TRUE(err.Msg().find("function error"));

    handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterPreStopFunction([](Context &context) { throw std::runtime_error("runtime error"); });
    SetRuntimeHandler(std::move(handlerPtr));
    err = exec_->ExecuteShutdownFunction(1);
    ASSERT_FALSE(err.OK());
    ASSERT_TRUE(err.Msg().find("runtime error"));

    handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterPreStopFunction([](Context &context) { return; });
    SetRuntimeHandler(std::move(handlerPtr));
    err = exec_->ExecuteShutdownFunction(1);
    ASSERT_TRUE(err.OK()) << err.Msg();
}

TEST_F(FaasExecutorTest, CheckpointRecoverSuccessfullyTest)
{
    YR::Libruntime::ErrorInfo err;

    std::shared_ptr<YR::Libruntime::Buffer> data;
    err = exec_->Checkpoint("instanceid", data);
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK) << err.Msg();

    err = exec_->Recover(data);
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK) << err.Msg();
}

TEST_F(FaasExecutorTest, ContextImplTest)
{
    std::unordered_map<std::string, std::string> params;
    auto contextInvokeParams = std::make_shared<ContextInvokeParams>(params);
    auto contextEnv = std::make_shared<ContextEnv>();
    ContextImpl ctx(contextInvokeParams, contextEnv);
    ContextImpl ctx2 = ctx;
    ctx.SetFuncStartTime(1000000);
    ctx.SetStateId("stateId");
    ASSERT_EQ(ctx.GetInstanceId(), "stateId");
    ctx.SetState("state");
    ASSERT_EQ(ctx.GetState(), "state");
    ctx.SetInvokeProperty("prop");
    ASSERT_EQ(ctx.GetInvokeProperty(), "prop");
}

TEST_F(FaasExecutorTest, RuntimeTest)
{
    std::shared_ptr<Runtime> rt = std::make_shared<Runtime>();
    rt->InitRuntimeLogger();
    rt->RegisterHandler([](const std::string &request, Function::Context &context) -> std::string { return ""; });
    rt->RegisterInitializerFunction([](Function::Context &context) { return; });
    rt->RegisterPreStopFunction([](Function::Context &context) { return; });
    rt->InitState([](const std::string &, Function::Context &) { return; });
    ASSERT_NO_THROW(rt->BuildRegisterRuntimeHandler());
    rt->ReleaseRuntimeLogger();
}

}  // namespace test
}  // namespace YR