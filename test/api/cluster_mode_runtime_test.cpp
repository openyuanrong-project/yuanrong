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

#include <memory>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "common/mock_libruntime.h"
#include "api/cpp/include/yr/api/exception.h"
#include "api/cpp/include/yr/api/hetero_exception.h"
#include "api/cpp/src/cluster_mode_runtime.h"
#include "api/cpp/src/config_manager.h"
#include "api/cpp/src/internal.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/utility/logger/logger.h"
#include "yr/api/err_type.h"
#include "yr/yr.h"

using namespace testing;
namespace YR {
namespace test {
using namespace testing;
using testing::AtLeast;
using testing::HasSubstr;
using testing::SetArgReferee;
using namespace YR::utility;
class ClusterModeRuntimeTest : public testing::Test {
public:
    ClusterModeRuntimeTest(){};
    ~ClusterModeRuntimeTest(){};
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
        auto lc = std::make_shared<YR::Libruntime::LibruntimeConfig>();
        lc->jobId = YR::utility::IDGenerator::GenApplicationId();
        auto clientsMgr = std::make_shared<YR::Libruntime::ClientsManager>();
        auto metricsAdaptor = std::make_shared<YR::Libruntime::MetricsAdaptor>();
        auto sec = std::make_shared<YR::Libruntime::Security>();
        auto socketClient = std::make_shared<YR::Libruntime::DomainSocketClient>("/home/snuser/socket/runtime.sock");
        lr = std::make_shared<YR::Libruntime::MockLibruntime>(lc, clientsMgr, metricsAdaptor, sec, socketClient);
        YR::Libruntime::LibruntimeManager::Instance().SetLibRuntime(lr);
        rt = std::make_shared<ClusterModeRuntime>();
    }
    void TearDown()
    {
        YR::Libruntime::LibruntimeManager::Instance().Finalize();
        lr.reset();
        rt.reset();
    }
    std::shared_ptr<ClusterModeRuntime> rt;
    std::shared_ptr<YR::Libruntime::MockLibruntime> lr;
};

TEST_F(ClusterModeRuntimeTest, InitClusterModeRuntimeTest)
{
    Config conf;
    conf.isDriver = true;
    conf.mode = Config::Mode::CLUSTER_MODE;
    conf.functionUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    conf.javaFunctionUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    conf.pythonFunctionUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    conf.serverAddr = "127.0.0.1:1234";
    conf.threadPoolSize = 4;
    conf.loadPaths = std::vector<std::string>(1025, std::string(1, 'a'));
    int mockArgc = 5;
    char *mockArgv[] = {"--logDir=/tmp/log", "--logLevel=DEBUG", "--grpcAddress=127.0.0.1:1234", "--runtimeId=driver",
                        "jobId=job123"};

    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);

    EXPECT_THROW(rt->Init(), YR::Exception);

    conf.dataSystemAddr = "127.0.0.1:11111";
    ConfigManager::Singleton().Init(conf, mockArgc, mockArgv);
    EXPECT_NO_THROW(rt->Init());
}

TEST_F(ClusterModeRuntimeTest, When_In_Cluster_With_Empty_DatasystemAddr_Should_Throw_Exception)
{
    Config conf;
    conf.inCluster = true;
    conf.mode = Config::Mode::CLUSTER_MODE;
    conf.serverAddr = "127.0.0.1:1234";
    conf.functionUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    ConfigManager::Singleton().Init(conf, 0, nullptr);
    EXPECT_THROW(rt->Init(), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, CreateInstanceFailedTest)
{
    internal::FuncMeta funcMeta;
    std::vector<YR::internal::InvokeArg> args;
    YR::InvokeOptions opts;
    EXPECT_THROW(rt->CreateInstance(funcMeta, args, opts), YR::Exception);

    funcMeta.appName = "appName";
    funcMeta.moduleName = "moduleName";
    funcMeta.funcName = "funcName";
    funcMeta.funcUrn = "abc123";
    funcMeta.className = "className";
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_JAVA;
    try {
        rt->CreateInstance(funcMeta, args, opts);
    } catch (YR::Exception &e) {
        ASSERT_EQ(e.Code(), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);
    }
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_PYTHON;
    try {
        rt->CreateInstance(funcMeta, args, opts);
    } catch (YR::Exception &e) {
        ASSERT_EQ(e.Code(), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);
    }
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    try {
        rt->CreateInstance(funcMeta, args, opts);
    } catch (YR::Exception &e) {
        ASSERT_EQ(e.Code(), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);
    }

    funcMeta.funcUrn = "sn:cn:yrk:default:function:0-f-a:latest";
    YR::Libruntime::ErrorInfo err(YR::Libruntime::ErrorCode::ERR_DEPENDENCY_FAILED, YR::Libruntime::ModuleCode::RUNTIME,
                                  "dependency resolve failed");
    EXPECT_CALL(*lr.get(), CreateInstance(_, _, _)).WillOnce(Return(std::make_pair(err, "instanceID")));
    EXPECT_THROW(rt->CreateInstance(funcMeta, args, opts), YR::Exception);

    funcMeta.language = static_cast<YR::internal::FunctionLanguage>(10);
    try {
        rt->CreateInstance(funcMeta, args, opts);
    } catch (YR::Exception &e) {
        ASSERT_EQ(e.Code(), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);
    }
}

bool checkFunction(const Exception &e) noexcept
{
    return true;
}

TEST_F(ClusterModeRuntimeTest, BuildOptionsTest)
{
    const std::string key = "key1";
    YR::LabelExistsOperator labelExistsOp(key);
    YR::ResourcePreferredAffinity affinity(labelExistsOp);

    YR::InvokeOptions invokeOptions;
    invokeOptions.retryChecker = checkFunction;
    invokeOptions.preferredPriority = false;
    invokeOptions.preferredAntiOtherLabels = false;
    invokeOptions.AddAffinity(affinity);

    auto af2 = YR::InstancePreferredAffinity(YR::LabelInOperator("key", {"value"}));
    auto af3 = YR::ResourcePreferredAntiAffinity(YR::LabelNotInOperator("key", {"value"}));
    auto af4 = YR::InstancePreferredAntiAffinity(YR::LabelExistsOperator("key"));
    auto af5 = YR::ResourceRequiredAffinity(YR::LabelDoesNotExistOperator("key"));
    auto af6 = YR::InstanceRequiredAffinity(YR::LabelNotInOperator("key", {"value"}));
    auto af7 = YR::ResourceRequiredAntiAffinity(YR::LabelInOperator("key", {"value"}));
    auto af8 = YR::InstanceRequiredAntiAffinity(YR::LabelNotInOperator("key", {"value"}));
    auto af9 = YR::Affinity(INSTANCE, REQUIRED_ANTI, {YR::LabelNotInOperator("key", {"value"})}, "POD");
    af9.SetAffinityScope(YR::AFFINITYSCOPE_NODE);
    invokeOptions.AddAffinity({af2, af3, af4, af5, af6, af7, af8, af9});

    YR::InstanceRange instanceRange;
    YR::RangeOptions rangeOpts;
    instanceRange.min = 1;
    instanceRange.max = 10;
    instanceRange.sameLifecycle = true;
    rangeOpts.timeout = 60;
    instanceRange.rangeOpts = rangeOpts;
    invokeOptions.instanceRange = instanceRange;
    invokeOptions.preemptedAllowed = true;
    invokeOptions.instancePriority = 200;
    invokeOptions.scheduleTimeoutMs = 50000;

    YR::Libruntime::InvokeOptions libInvokeOptions = BuildOptions(std::move(invokeOptions));
    auto firstAffinity = libInvokeOptions.scheduleAffinities.front();
    EXPECT_FALSE(firstAffinity->GetPreferredAntiOtherLabels());
    auto lastAffinity = libInvokeOptions.scheduleAffinities.back();
    ASSERT_EQ(lastAffinity->GetAffinityScope(), YR::AFFINITYSCOPE_NODE);
    ASSERT_EQ(libInvokeOptions.instanceRange.min, instanceRange.min);
    ASSERT_EQ(libInvokeOptions.instanceRange.max, instanceRange.max);
    ASSERT_EQ(libInvokeOptions.instanceRange.step, 2);
    ASSERT_EQ(libInvokeOptions.instanceRange.sameLifecycle, instanceRange.sameLifecycle);
    ASSERT_EQ(libInvokeOptions.instanceRange.rangeOpts.timeout, instanceRange.rangeOpts.timeout);
    ASSERT_EQ(libInvokeOptions.preemptedAllowed, true);
    ASSERT_EQ(libInvokeOptions.instancePriority, 200);
    ASSERT_EQ(libInvokeOptions.scheduleTimeoutMs, 50000);

    YR::InvokeOptions invokeOptions1;
    invokeOptions1.requiredPriority = false;
    invokeOptions1.preferredPriority = true;
    invokeOptions1.preferredAntiOtherLabels = false;
    auto aff1 = YR::ResourcePreferredAntiAffinity(YR::LabelExistsOperator("test"));
    invokeOptions1.AddAffinity(aff1);
    YR::Libruntime::InvokeOptions libInvokeOptions1 = BuildOptions(std::move(invokeOptions1));
    for (auto aff : libInvokeOptions1.scheduleAffinities) {
        EXPECT_EQ(aff->GetPreferredAntiOtherLabels(), false);
    }
}

TEST_F(ClusterModeRuntimeTest, TestCreateInstanceSuccessfully)
{
    EXPECT_CALL(*lr.get(), CreateInstance(_, _, _))
        .WillOnce(Return(std::make_pair<YR::Libruntime::ErrorInfo, std::string>(YR::Libruntime::ErrorInfo(), "111")));
    internal::FuncMeta funcMeta;
    funcMeta.appName = "appName";
    funcMeta.moduleName = "moduleName";
    funcMeta.funcName = "funcName";
    funcMeta.funcUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    funcMeta.className = "className";
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    funcMeta.name = "name";
    funcMeta.ns = "ns";
    std::vector<YR::internal::InvokeArg> args;
    auto arg = YR::internal::InvokeArg();
    std::string s = "aa";
    arg.buf.write(s.c_str(), s.size());
    args.push_back(std::move(arg));
    YR::InvokeOptions opts;
    auto instanceId = rt->CreateInstance(funcMeta, args, opts);
    ASSERT_EQ(instanceId, "111");
}

MATCHER_P(ScheduleAffinitiesMatcher, inputOp, "Schedule affinities matcher")
{
    const auto &op = (const YR::Libruntime::InvokeOptions &)arg;
    if (op.scheduleAffinities.size() != inputOp.scheduleAffinities.size()) {
        std::cerr << "scheduleAffinities size: " << op.scheduleAffinities.size()
                  << " != " << inputOp.scheduleAffinities.size() << std::endl;
        return false;
    }
    auto it2 = inputOp.scheduleAffinities.begin();
    for (const auto &it1 : op.scheduleAffinities) {
        if (it1->GetAffinityHash() != (*it2)->GetAffinityHash()) {
            std::cerr << it1->GetString() << " != " << (*it2)->GetString() << std::endl;
            return false;
        }
        it2++;
    }
    return true;
}

TEST_F(ClusterModeRuntimeTest, TestCreateInstanceSuccessfullyWithAffinity)
{
    internal::FuncMeta funcMeta;
    funcMeta.appName = "appName";
    funcMeta.moduleName = "moduleName";
    funcMeta.funcName = "funcName";
    funcMeta.funcUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    funcMeta.className = "className";
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    std::vector<YR::internal::InvokeArg> args;
    YR::InvokeOptions opts;
    auto af = YR::ResourcePreferredAffinity(YR::LabelInOperator("key", {"value"}));
    opts.AddAffinity(af);
    YR::Libruntime::InvokeOptions op;
    auto af2 = std::make_shared<YR::Libruntime::ResourcePreferredAffinity>();
    auto labelInOperator = std::make_shared<YR::Libruntime::LabelInOperator>();
    labelInOperator->SetKey("key");
    labelInOperator->SetValues({"value"});
    af2->SetLabelOperators(std::list<std::shared_ptr<YR::Libruntime::LabelOperator>>{labelInOperator});
    op.scheduleAffinities.push_back(af2);
    EXPECT_CALL(*lr.get(), CreateInstance(_, _, ScheduleAffinitiesMatcher(op)))
        .WillOnce(Return(std::make_pair<YR::Libruntime::ErrorInfo, std::string>(YR::Libruntime::ErrorInfo(), "111")));
    auto instanceId = rt->CreateInstance(funcMeta, args, opts);
    ASSERT_EQ(instanceId, "111");
}

TEST_F(ClusterModeRuntimeTest, TestInvokeInstanceSuccessfully)
{
    internal::FuncMeta funcMeta;
    funcMeta.appName = "appName";
    funcMeta.moduleName = "moduleName";
    funcMeta.funcName = "funcName";
    funcMeta.funcUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    funcMeta.className = "className";
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    std::vector<YR::internal::InvokeArg> args;
    YR::InvokeOptions opts;
    std::vector<YR::Libruntime::DataObject> returnObjs{{"111"}};
    EXPECT_CALL(*lr.get(), InvokeByInstanceId(_, _, _, _, _))
        .WillOnce(DoAll(SetArgReferee<4>(returnObjs), Return(YR::Libruntime::ErrorInfo())));
    auto objectId = rt->InvokeInstance(funcMeta, "instanceid", args, opts);
    ASSERT_EQ(objectId, "111");
}

TEST_F(ClusterModeRuntimeTest, TestInvokeInstanceFailed)
{
    internal::FuncMeta funcMeta;
    std::vector<YR::internal::InvokeArg> args;
    YR::InvokeOptions opts;
    EXPECT_THROW(rt->InvokeInstance(funcMeta, "instanceID", args, opts), YR::Exception);

    funcMeta.appName = "appName";
    funcMeta.moduleName = "moduleName";
    funcMeta.funcName = "funcName";
    funcMeta.funcUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    funcMeta.className = "className";
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    YR::Libruntime::ErrorInfo err(YR::Libruntime::ErrorCode::ERR_DEPENDENCY_FAILED, YR::Libruntime::ModuleCode::RUNTIME,
                                  "dependency resolve failed");
    EXPECT_CALL(*lr.get(), InvokeByInstanceId(_, _, _, _, _)).WillOnce(Return(err));
    EXPECT_THROW(rt->InvokeInstance(funcMeta, "instanceID", args, opts), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestInvokeByNameSuccessfully)
{
    internal::FuncMeta funcMeta;
    funcMeta.appName = "appName";
    funcMeta.moduleName = "moduleName";
    funcMeta.funcName = "funcName";
    funcMeta.funcUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    funcMeta.className = "className";
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    std::vector<YR::internal::InvokeArg> args;
    YR::InvokeOptions opts;
    std::vector<YR::Libruntime::DataObject> returnObjs{{"111"}};
    EXPECT_CALL(*lr.get(), InvokeByFunctionName(_, _, _, _))
        .WillOnce(DoAll(SetArgReferee<3>(returnObjs), Return(YR::Libruntime::ErrorInfo())));
    auto objectId = rt->InvokeByName(funcMeta, args, opts);
    ASSERT_EQ(objectId, "111");
}

TEST_F(ClusterModeRuntimeTest, TestInvokeByNameFailed)
{
    internal::FuncMeta funcMeta;
    std::vector<YR::internal::InvokeArg> args;
    YR::InvokeOptions opts;
    EXPECT_THROW(rt->InvokeByName(funcMeta, args, opts), YR::Exception);

    funcMeta.appName = "appName";
    funcMeta.moduleName = "moduleName";
    funcMeta.funcName = "funcName";
    funcMeta.funcUrn = "sn:cn:yrk:default:function:0-test-test:$latest";
    funcMeta.className = "className";
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_CPP;
    YR::Libruntime::ErrorInfo err(YR::Libruntime::ErrorCode::ERR_DEPENDENCY_FAILED, YR::Libruntime::ModuleCode::RUNTIME,
                                  "dependency resolve failed");
    EXPECT_CALL(*lr.get(), InvokeByFunctionName(_, _, _, _)).WillOnce(Return(err));
    EXPECT_THROW(rt->InvokeByName(funcMeta, args, opts), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestTerminateInstanceSuccessfully)
{
    EXPECT_CALL(*lr.get(), Kill(_, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->TerminateInstance("111"));
}

TEST_F(ClusterModeRuntimeTest, TestTerminateInstanceFailed)
{
    EXPECT_CALL(*lr.get(), Kill(_, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "111")));
    EXPECT_THROW(rt->TerminateInstance("111"), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestTerminateInstanceSyncSuccessfully)
{
    EXPECT_CALL(*lr.get(), Kill(_, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->TerminateInstanceSync("111"));
}

TEST_F(ClusterModeRuntimeTest, TestTerminateInstanceSyncFailed)
{
    EXPECT_CALL(*lr.get(), Kill(_, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "111")));
    EXPECT_THROW(rt->TerminateInstanceSync("111"), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestPutWithoutObjIDFailed)
{
    EXPECT_THROW(rt->Put(std::make_shared<msgpack::sbuffer>(), {}), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestPutFailed)
{
    EXPECT_CALL(*lr.get(), CreateDataObject(_, _, _, _, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "111")));
    EXPECT_THROW(rt->Put("111", std::make_shared<msgpack::sbuffer>(), {}), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestGetUnlimitedRetry)
{
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> ret;
    ret.push_back(std::make_shared<YR::Libruntime::DataObject>(0, 1));
    Libruntime::RetryInfo retryInfo;
    retryInfo.retryType = Libruntime::RetryType::UNLIMITED_RETRY;
    EXPECT_CALL(*lr.get(), GetDataObjectsWithoutWait(_, _))
        .WillRepeatedly(
            Return(std::make_pair<YR::Libruntime::RetryInfo, std::vector<std::shared_ptr<YR::Libruntime::DataObject>>>(
                std::move(retryInfo), std::move(ret))));
    int limitedRetryTime = 0;
    for (int i = 0; i < LIMITED_RETRY_TIME; i++) {
        ASSERT_TRUE(rt->Get({"111"}, 10, limitedRetryTime).first.needRetry);
    }
}

TEST_F(ClusterModeRuntimeTest, TestGetLimitedRetry)
{
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> ret;
    ret.push_back(std::make_shared<YR::Libruntime::DataObject>(0, 1));
    Libruntime::RetryInfo retryInfo;
    retryInfo.retryType = Libruntime::RetryType::LIMITED_RETRY;
    EXPECT_CALL(*lr.get(), GetDataObjectsWithoutWait(_, _))
        .WillRepeatedly(
            Return(std::make_pair<YR::Libruntime::RetryInfo, std::vector<std::shared_ptr<YR::Libruntime::DataObject>>>(
                std::move(retryInfo), std::move(ret))));
    int limitedRetryTime = 0;
    for (int i = 0; i < LIMITED_RETRY_TIME - 1; i++) {
        ASSERT_TRUE(rt->Get({"111"}, 10, limitedRetryTime).first.needRetry);
    }
    ASSERT_FALSE(rt->Get({"111"}, 10, limitedRetryTime).first.needRetry);
}

TEST_F(ClusterModeRuntimeTest, TestWaitSuccessfully)
{
    auto ret = std::make_shared<InternalWaitResult>();
    ret->readyIds.emplace_back("111");
    EXPECT_CALL(*lr.get(), Wait(_, _, _)).WillOnce(Return(ret));
    auto ret2 = rt->Wait({"111"}, 1, 10);
    EXPECT_EQ(ret2.readyIds.size(), 1);
    EXPECT_EQ(ret2.readyIds[0], "111");
}

TEST_F(ClusterModeRuntimeTest, TestWaitFailed)
{
    auto ret = std::make_shared<InternalWaitResult>();
    ret->exceptionIds.emplace("111",
                              YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_CONNECTION_FAILED, "aaa"));
    EXPECT_CALL(*lr.get(), Wait(_, _, _)).WillOnce(Return(ret));
    EXPECT_THROW(rt->Wait({"111"}, 1, 10), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestGetInstancesSuccessfully)
{
    EXPECT_CALL(*lr.get(), GetInstances(_, Matcher<int>(1)))
        .WillOnce(Return(
            std::make_pair<std::vector<std::string>, YR::Libruntime::ErrorInfo>({"111"}, YR::Libruntime::ErrorInfo())));
    auto ret2 = rt->GetInstances("111", 1);
    EXPECT_EQ(ret2.size(), 1);
    EXPECT_EQ(ret2[0], "111");
}

TEST_F(ClusterModeRuntimeTest, TestGetInstancesFailed)
{
    EXPECT_CALL(*lr.get(), GetInstances(_, Matcher<int>(1)))
        .WillOnce(Return(std::make_pair<std::vector<std::string>, YR::Libruntime::ErrorInfo>(
            {}, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_CONNECTION_FAILED, "aaa"))));
    EXPECT_THROW(rt->GetInstances("111", 1), YR::Exception);
    EXPECT_THROW(rt->GetInstances("111", -2), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestGenerateGroupNameSuccessfully)
{
    EXPECT_CALL(*lr.get(), GenerateGroupName()).WillOnce(Return("111"));
    auto ret2 = rt->GenerateGroupName();
    EXPECT_EQ(ret2, "111");
}

TEST_F(ClusterModeRuntimeTest, TestIncreGlobalReferenceSuccessfully)
{
    EXPECT_CALL(*lr.get(), IncreaseReference(_, Matcher<bool>(_))).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->IncreGlobalReference({"111"}));
}

TEST_F(ClusterModeRuntimeTest, TestIncreGlobalReferenceFailed)
{
    EXPECT_CALL(*lr.get(), IncreaseReference(_, Matcher<bool>(_)))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_CONNECTION_FAILED, "aaa")));
    EXPECT_THROW(rt->IncreGlobalReference({"111"}), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestDecreGlobalReference)
{
    EXPECT_CALL(*lr.get(), DecreaseReference(_)).WillOnce(Return());
    EXPECT_NO_THROW(rt->DecreGlobalReference({"111"}));
}

TEST_F(ClusterModeRuntimeTest, TestKVWriteSuccessfully)
{
    SetParam param;
    SetParamV2 paramV2;
    std::string val = "val";
    EXPECT_CALL(*lr.get(), KVWrite(_, _, _)).WillRepeatedly(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->KVWrite("111", std::make_shared<msgpack::sbuffer>(), param));
    EXPECT_NO_THROW(rt->KVWrite("111", val.c_str(), param));
    EXPECT_NO_THROW(rt->KVWrite("111", std::make_shared<msgpack::sbuffer>(), paramV2));
}

TEST_F(ClusterModeRuntimeTest, TestKVWriteFailed)
{
    SetParam param;
    SetParamV2 paramV2;
    std::string val = "val";
    EXPECT_CALL(*lr.get(), KVWrite(_, _, _))
        .WillRepeatedly(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->KVWrite("111", std::make_shared<msgpack::sbuffer>(), param), YR::Exception);
    EXPECT_THROW(rt->KVWrite("111", val.c_str(), param), YR::Exception);
    EXPECT_THROW(rt->KVWrite("111", std::make_shared<msgpack::sbuffer>(), paramV2), YR::Exception);
    EXPECT_CALL(*lr.get(), SetTraceId(_))
        .WillRepeatedly(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->KVWrite("111", std::make_shared<msgpack::sbuffer>(), param), YR::Exception);
    EXPECT_THROW(rt->KVWrite("111", val.c_str(), param), YR::Exception);
    EXPECT_THROW(rt->KVWrite("111", std::make_shared<msgpack::sbuffer>(), paramV2), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestKVMSetTxSuccessfully)
{
    EXPECT_CALL(*lr.get(), KVMSetTx(_, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->KVMSetTx({"111"}, {std::make_shared<msgpack::sbuffer>()}, YR::ExistenceOpt()));
}

TEST_F(ClusterModeRuntimeTest, TestKVMSetTxFailed)
{
    EXPECT_CALL(*lr.get(), KVMSetTx(_, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->KVMSetTx({"111"}, {std::make_shared<msgpack::sbuffer>()}, YR::ExistenceOpt()), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestKVReadSuccessfully)
{
    std::shared_ptr<YR::Libruntime::Buffer> ret = std::make_shared<YR::Libruntime::NativeBuffer>(1);
    EXPECT_CALL(*lr.get(), KVRead(_, _)).WillOnce(Return(std::make_pair(ret, YR::Libruntime::ErrorInfo())));
    EXPECT_NO_THROW(rt->KVRead("111", 1000));
}

TEST_F(ClusterModeRuntimeTest, TestKVReadFailed)
{
    EXPECT_CALL(*lr.get(), KVRead(_, _))
        .WillOnce(Return(
            std::make_pair(nullptr, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa"))));
    EXPECT_THROW(rt->KVRead("111", 1000), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestKVReadAllowPartialSuccessfully)
{
    std::vector<std::shared_ptr<YR::Libruntime::Buffer>> ret;
    ret.push_back(std::make_shared<YR::Libruntime::NativeBuffer>(1));
    EXPECT_CALL(*lr.get(), KVRead(_, _, _)).WillOnce(Return(std::make_pair(ret, YR::Libruntime::ErrorInfo())));
    EXPECT_NO_THROW(rt->KVRead({"111"}, 1000, true));
}

TEST_F(ClusterModeRuntimeTest, TestKVReadAllowPartialFailed)
{
    std::vector<std::shared_ptr<YR::Libruntime::Buffer>> ret;
    EXPECT_CALL(*lr.get(), KVRead(_, _, _))
        .WillOnce(Return(
            std::make_pair(ret, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa"))));
    EXPECT_THROW(rt->KVRead({"111"}, 1000, true), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestKVGetWithParamSuccessfully)
{
    YR::GetParam param1;
    YR::GetParam param2;
    YR::GetParams params;
    params.getParams = {param1, param2};
    std::vector<std::shared_ptr<YR::Libruntime::Buffer>> ret;
    ret.push_back(std::make_shared<YR::Libruntime::NativeBuffer>(1));
    ret.push_back(nullptr);
    EXPECT_CALL(*lr.get(), KVGetWithParam(_, _, _)).WillOnce(Return(std::make_pair(ret, YR::Libruntime::ErrorInfo())));
    EXPECT_NO_THROW(rt->KVGetWithParam({"111", "222"}, params, 1000));
}

TEST_F(ClusterModeRuntimeTest, TestKVGetWithParamFailed)
{
    YR::GetParam param1;
    YR::GetParam param2;
    YR::GetParams params;
    params.getParams = {param1, param2};
    std::vector<std::shared_ptr<YR::Libruntime::Buffer>> ret;
    EXPECT_CALL(*lr.get(), KVGetWithParam(_, _, _))
        .WillOnce(Return(
            std::make_pair(ret, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa"))));
    EXPECT_THROW(rt->KVGetWithParam({"111", "222"}, params, 1000), YR::Exception);
    EXPECT_CALL(*lr.get(), SetTraceId(_))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->KVGetWithParam({"111", "222"}, params, 1000), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestKVDelSuccessfully)
{
    EXPECT_CALL(*lr.get(), KVDel(Matcher<const std::string &>("111"))).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->KVDel("111"));
}

TEST_F(ClusterModeRuntimeTest, TestKVExistSuccessfully)
{
    std::vector<bool> exists = {false, false, true};
    EXPECT_CALL(*lr.get(), KVExist(_)).WillOnce(Return(std::make_pair(exists, YR::Libruntime::ErrorInfo())));
    ASSERT_EQ(rt->KVExist({"key1", "key2", "key3"}), exists);
}

TEST_F(ClusterModeRuntimeTest, TestKVExistFailed)
{
    std::vector<bool> exists = {false, false, false};
    EXPECT_CALL(*lr.get(), KVExist(_))
        .WillOnce(Return(
            std::make_pair(exists, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa"))));
    EXPECT_THROW(rt->KVExist({"key1", "key2", "key3"}), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestKVDelFailed)
{
    EXPECT_CALL(*lr.get(), KVDel(Matcher<const std::string &>("111")))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->KVDel("111"), YR::Exception);
    EXPECT_CALL(*lr.get(), SetTraceId(_))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->KVDel("111"), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestKVDelMultiKeysSuccessfully)
{
    std::vector<std::string> input = {"111"};
    std::vector<std::string> ret = {"res"};
    EXPECT_CALL(*lr.get(), KVDel(Matcher<const std::vector<std::string> &>({"111"})))
        .WillOnce(Return(std::make_pair(ret, YR::Libruntime::ErrorInfo())));
    EXPECT_NO_THROW(rt->KVDel(input));
}

TEST_F(ClusterModeRuntimeTest, TestKVDelMultiKeysFailed)
{
    std::vector<std::string> input = {"111"};
    std::vector<std::string> ret;
    EXPECT_CALL(*lr.get(), KVDel(Matcher<const std::vector<std::string> &>({"111"})))
        .WillOnce(Return(
            std::make_pair(ret, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa"))));
    EXPECT_THROW(rt->KVDel(input), YR::Exception);
    EXPECT_CALL(*lr.get(), SetTraceId(_))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->KVDel(input), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestCreateStreamProducerSuccessfully)
{
    EXPECT_CALL(*lr.get(), CreateStreamProducer(_, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->CreateStreamProducer("streamName", YR::ProducerConf()));
}

TEST_F(ClusterModeRuntimeTest, TestCreateStreamProducerFailed)
{
    EXPECT_CALL(*lr.get(), CreateStreamProducer(_, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->CreateStreamProducer("streamName", YR::ProducerConf()), YR::Exception);
    EXPECT_CALL(*lr.get(), SetTraceId(_))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->CreateStreamProducer("streamName", YR::ProducerConf()), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestCreateStreamConsumerSuccessfully)
{
    EXPECT_CALL(*lr.get(), CreateStreamConsumer(_, _, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->CreateStreamConsumer("streamName", YR::SubscriptionConfig(), true));
}

TEST_F(ClusterModeRuntimeTest, TestCreateStreamConsumerFailed)
{
    EXPECT_CALL(*lr.get(), CreateStreamConsumer(_, _, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->CreateStreamConsumer("streamName", YR::SubscriptionConfig(), true), YR::Exception);
    EXPECT_CALL(*lr.get(), SetTraceId(_))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->CreateStreamConsumer("streamName", YR::SubscriptionConfig(), true), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestDeleteStreamSuccessfully)
{
    EXPECT_CALL(*lr.get(), DeleteStream(_)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->DeleteStream("streamName"));
}

TEST_F(ClusterModeRuntimeTest, TestDeleteStreamFailed)
{
    EXPECT_CALL(*lr.get(), DeleteStream(_))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->DeleteStream("streamName"), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestQueryGlobalProducersNumSuccessfully)
{
    uint64_t gProducerNum = 1;
    EXPECT_CALL(*lr.get(), QueryGlobalProducersNum(_, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->QueryGlobalProducersNum("streamName", gProducerNum));
}

TEST_F(ClusterModeRuntimeTest, TestQueryGlobalProducersNumFailed)
{
    uint64_t gProducerNum = 1;
    EXPECT_CALL(*lr.get(), QueryGlobalProducersNum(_, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_CONNECTION_FAILED, "aaa")));
    EXPECT_THROW(rt->QueryGlobalProducersNum("streamName", gProducerNum), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestQueryGlobalConsumersNumSuccessfully)
{
    uint64_t gConsumerNum = 1;
    EXPECT_CALL(*lr.get(), QueryGlobalConsumersNum(_, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->QueryGlobalConsumersNum("streamName", gConsumerNum));
}

TEST_F(ClusterModeRuntimeTest, TestQueryGlobalConsumersNumFailed)
{
    uint64_t gConsumerNum = 1;
    EXPECT_CALL(*lr.get(), QueryGlobalConsumersNum(_, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_CONNECTION_FAILED, "aaa")));
    EXPECT_THROW(rt->QueryGlobalConsumersNum("streamName", gConsumerNum), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestGetRealInstanceIdSuccessfully)
{
    EXPECT_CALL(*lr.get(), GetRealInstanceId(_, _)).WillOnce(Return("realInstanceID"));
    EXPECT_NO_THROW(rt->GetRealInstanceId("objID"));
}

TEST_F(ClusterModeRuntimeTest, TestSaveRealInstanceIdSuccessfully)
{
    EXPECT_CALL(*lr.get(), SaveRealInstanceId(_, _, _)).WillOnce(Return());
    EXPECT_NO_THROW(rt->SaveRealInstanceId("objID", "insID", YR::InvokeOptions()));
}

TEST_F(ClusterModeRuntimeTest, TestGetGroupInstanceIdsSuccessfully)
{
    EXPECT_CALL(*lr.get(), GetGroupInstanceIds(_, _)).WillOnce(Return("groupInsIds"));
    EXPECT_NO_THROW(rt->GetGroupInstanceIds("objID"));
}

TEST_F(ClusterModeRuntimeTest, TestSaveGroupInstanceIdsSuccessfully)
{
    EXPECT_CALL(*lr.get(), SaveGroupInstanceIds(_, _, _)).WillOnce(Return());
    EXPECT_NO_THROW(rt->SaveGroupInstanceIds("objID", "groupInsIds", YR::InvokeOptions()));
}

TEST_F(ClusterModeRuntimeTest, TestCancelSuccessfully)
{
    EXPECT_CALL(*lr.get(), Cancel(_, _, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->Cancel({"111"}, true, true));
}

TEST_F(ClusterModeRuntimeTest, TestCancelFailed)
{
    EXPECT_CALL(*lr.get(), Cancel(_, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, "aaa")));
    EXPECT_THROW(rt->Cancel({"111"}, true, true), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestExit)
{
    EXPECT_CALL(*lr.get(), Exit()).WillOnce(Return());
    EXPECT_NO_THROW(rt->Exit());
}

TEST_F(ClusterModeRuntimeTest, TestIsOnCloud)
{
    EXPECT_FALSE(rt->IsOnCloud());
}

TEST_F(ClusterModeRuntimeTest, TestGroupCreateSuccessfully)
{
    YR::GroupOptions gOpts;
    EXPECT_CALL(*lr.get(), GroupCreate(_, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->GroupCreate("111", gOpts));
}

TEST_F(ClusterModeRuntimeTest, TestGroupCreateSuccessfullyWithBindOptions)
{
    YR::GroupOptions gOpts;
    std::string gName = "gName";
    gOpts.bind.resource = "NUMA";
    gOpts.bind.strategy = "SPREAD";
    YR::Libruntime::GroupOpts libOpts;
    EXPECT_CALL(*lr.get(), GroupCreate("gName", _))
    .WillOnce(::testing::DoAll(
        ::testing::SaveArg<1>(&libOpts),  // 保存第二个参数
        Return(YR::Libruntime::ErrorInfo())));
    EXPECT_NO_THROW(rt->GroupCreate("gName", gOpts));
    ASSERT_EQ(libOpts.bind.resource, "NUMA");
    ASSERT_EQ(libOpts.bind.strategy, "SPREAD");
}

TEST_F(ClusterModeRuntimeTest, TestGroupCreateFailed)
{
    YR::GroupOptions gOpts;
    gOpts.timeout = -2;
    EXPECT_THROW(rt->GroupCreate("111", gOpts), YR::Exception);
    EXPECT_CALL(*lr.get(), GroupCreate(_, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    gOpts.timeout = 2;
    EXPECT_THROW(rt->GroupCreate("111", gOpts), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestGroupTerminate)
{
    EXPECT_CALL(*lr.get(), GroupTerminate(_)).WillOnce(Return());
    EXPECT_NO_THROW(rt->GroupTerminate("111"));
}

TEST_F(ClusterModeRuntimeTest, TestGroupWaitSuccessfully)
{
    EXPECT_CALL(*lr.get(), GroupWait(_)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->GroupWait("111"));
}

TEST_F(ClusterModeRuntimeTest, TestGroupWaitFailed)
{
    EXPECT_CALL(*lr.get(), GroupWait(_))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->GroupWait("111"), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestSaveStateSuccessfully)
{
    EXPECT_CALL(*lr.get(), SaveState(_, _)).WillOnce(Return(YR::Libruntime::ErrorInfo()));
    EXPECT_NO_THROW(rt->SaveState(100));
}

TEST_F(ClusterModeRuntimeTest, TestSaveStateFailed)
{
    EXPECT_CALL(*lr.get(), SaveState(_, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->SaveState(100), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestLoadStateFailed)
{
    EXPECT_CALL(*lr.get(), LoadState(_, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "aaa")));
    EXPECT_THROW(rt->LoadState(100), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestDelete)
{
    std::vector<std::string> objectIds;
    std::vector<std::string> failedObjectIds;
    ASSERT_NO_THROW(rt->DevDelete(objectIds, failedObjectIds));
    EXPECT_CALL(*lr.get(), DevDelete(_, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "111")));
    ASSERT_THROW(rt->DevDelete(objectIds, failedObjectIds), YR::HeteroException);
}

TEST_F(ClusterModeRuntimeTest, TestDevLocalDelete)
{
    std::vector<std::string> objectIds;
    std::vector<std::string> failedObjectIds;
    ASSERT_NO_THROW(rt->DevLocalDelete(objectIds, failedObjectIds));
    EXPECT_CALL(*lr.get(), DevLocalDelete(_, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "111")));
    ASSERT_THROW(rt->DevLocalDelete(objectIds, failedObjectIds), YR::HeteroException);
}

TEST_F(ClusterModeRuntimeTest, TestDevSubscribe)
{
    std::vector<std::string> keys;
    std::vector<YR::DeviceBlobList> blob2dList;
    std::vector<std::shared_ptr<YR::Future>> futureVec;
    ASSERT_NO_THROW(rt->DevSubscribe(keys, blob2dList, futureVec));
    EXPECT_CALL(*lr.get(), DevSubscribe(_, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "111")));
    ASSERT_THROW(rt->DevSubscribe(keys, blob2dList, futureVec), YR::HeteroException);
}

TEST_F(ClusterModeRuntimeTest, TestDevPublish)
{
    std::vector<std::string> keys;
    std::vector<YR::DeviceBlobList> blob2dList;
    std::vector<std::shared_ptr<YR::Future>> futureVec;
    ASSERT_NO_THROW(rt->DevPublish(keys, blob2dList, futureVec));
    EXPECT_CALL(*lr.get(), DevPublish(_, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "111")));
    ASSERT_THROW(rt->DevPublish(keys, blob2dList, futureVec), YR::HeteroException);
}

TEST_F(ClusterModeRuntimeTest, TestDevMSet)
{
    std::vector<std::string> keys;
    std::vector<YR::DeviceBlobList> blob2dList;
    std::vector<std::string> failedKeys;
    ASSERT_NO_THROW(rt->DevMSet(keys, blob2dList, failedKeys));
    EXPECT_CALL(*lr.get(), DevMSet(_, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "111")));
    ASSERT_THROW(rt->DevMSet(keys, blob2dList, failedKeys), YR::HeteroException);
}

TEST_F(ClusterModeRuntimeTest, TestDevMGet)
{
    std::vector<std::string> keys;
    std::vector<YR::DeviceBlobList> blob2dList;
    std::vector<std::string> failedKeys;
    ASSERT_NO_THROW(rt->DevMGet(keys, blob2dList, failedKeys, 1));
    EXPECT_CALL(*lr.get(), DevMGet(_, _, _, _))
        .WillOnce(Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "111")));
    ASSERT_THROW(rt->DevMGet(keys, blob2dList, failedKeys, 1), YR::HeteroException);
}

TEST_F(ClusterModeRuntimeTest, GetInstanceTest)
{
    std::string name = "name";
    std::string ns = "ns";
    YR::Libruntime::FunctionMeta funcMeta;
    funcMeta.name = "ins-name";
    EXPECT_CALL(*lr.get(), GetInstance(_, _, _))
        .WillOnce(testing::Return(std::make_pair(funcMeta, YR::Libruntime::ErrorInfo())));
    auto res = rt->GetInstance(name, ns, 60);
    ASSERT_EQ(res.name, "ins-name");

    EXPECT_CALL(*lr.get(), GetInstance(_, _, _))
        .WillOnce(testing::Return(
            std::make_pair(funcMeta, YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                                               YR::Libruntime::ModuleCode::RUNTIME, "111"))));
    ASSERT_THROW(rt->GetInstance(name, ns, 60), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestGetInstanceRouteSuccessfully)
{
    EXPECT_CALL(*lr.get(), GetInstanceRoute(_)).WillOnce(Return("instanceRoute"));
    EXPECT_NO_THROW(rt->GetInstanceRoute("objID"));
}

TEST_F(ClusterModeRuntimeTest, TestSaveInstanceRouteSuccessfully)
{
    EXPECT_CALL(*lr.get(), SaveInstanceRoute(_, _)).WillOnce(Return());
    EXPECT_NO_THROW(rt->SaveInstanceRoute("objID", "insRoute"));
}

TEST_F(ClusterModeRuntimeTest, TestInvokeByName)
{
    YR::SetInitialized(true);
    internal::FuncMeta funcMeta;
    funcMeta.funcName = "common";
    funcMeta.moduleName = "echo";
    funcMeta.language = YR::internal::FunctionLanguage::FUNC_LANG_PYTHON;
    funcMeta.funcUrn = "sn:cn:yrk:default:function:0-yr-stpython:$latest";
    InvokeOptions opts;
    std::string str = "success";
    YR::Buffer buf(str.data(), str.size());
    std::vector<YR::internal::InvokeArg> invokeArgs;
    YR::internal::InvokeArg invokeArg1{};
    invokeArg1.buf = std::move(YR::internal::Serialize(PY_PLACEHOLDER));
    invokeArg1.isRef = false;
    invokeArgs.emplace_back(std::move(invokeArg1));
    YR::internal::InvokeArg invokeArg2{};
    invokeArg2.yrBuf = buf;
    invokeArg2.isRef = false;
    invokeArgs.emplace_back(std::move(invokeArg2));

    std::vector<YR::Libruntime::InvokeArg> libArgs;
    std::vector<YR::Libruntime::DataObject> returnObjs{{"abc"}};
    EXPECT_CALL(*lr.get(), InvokeByFunctionName(_, _, _, _))
        .WillOnce(DoAll(SetArgReferee<3>(returnObjs),
                        ::testing::Invoke([this, &libArgs](auto &&arg1, auto &&arg2, auto &&arg3, auto &&arg4) {
                            libArgs = std::move(arg2);
                            return YR::Libruntime::ErrorInfo();
                        })));
    auto ret = rt->InvokeByName(funcMeta, invokeArgs, opts);
    ASSERT_EQ(ret, "abc");
    // 校验mock函数入参
    auto meta = libArgs[1].dataObj->meta;
    auto data = libArgs[1].dataObj->data;
    auto metaBuf = std::make_shared<YR::Buffer>(const_cast<void *>(meta->ImmutableData()), meta->GetSize());
    auto metaResult = YR::internal::Deserialize<uint8_t>(metaBuf);
    ASSERT_EQ(metaResult, 3);
    std::string dataResult = std::string(static_cast<const char *>(data->ImmutableData()), data->GetSize());
    ASSERT_EQ(dataResult, str);
}

TEST_F(ClusterModeRuntimeTest, APIPutNullptrFailed)
{
    ASSERT_THROW(rt->Put(std::make_shared<YR::Buffer>(nullptr, 0), {}, {}), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestCreateBuffer)
{
    ASSERT_NO_THROW(rt->CreateMutableBuffer(60));
    EXPECT_CALL(*lr.get(), CreateBuffer(_, _))
        .WillOnce(Return(std::make_pair(YR::Libruntime::ErrorInfo(YR::Libruntime::ERR_CLIENT_ALREADY_CLOSED,
                                                                  Libruntime::ModuleCode::RUNTIME, "111"),
                                        "nullptr")));
    EXPECT_THROW(rt->CreateMutableBuffer(60), YR::Exception);
}

TEST_F(ClusterModeRuntimeTest, TestGetMutableBuffer)
{
    std::vector<std::shared_ptr<YR::Libruntime::Buffer>> buffers{};
    std::vector<std::string> ids;
    ASSERT_NO_THROW(rt->GetMutableBuffer(ids, 60));
    EXPECT_CALL(*lr.get(), GetBuffers(_, _, _))
        .WillOnce(Return(std::make_pair(YR::Libruntime::ErrorInfo(YR::Libruntime::ERR_CLIENT_ALREADY_CLOSED,
                                                                  Libruntime::ModuleCode::RUNTIME, "111"),
                                        buffers)));
    ASSERT_THROW(rt->GetMutableBuffer(ids, 60), YR::Exception);
}

class A {
public:
    int a;
    A() {}
    A(int init)
    {
        a = init;
    }

    static A *Create(int init)
    {
        return new A(init);
    }
    int Add(int x)
    {
        a += x;
        return a;
    }
    MSGPACK_DEFINE(a)
};

struct Y {
    YR::ObjectRef<int> obj;
    YR_STATE(obj)
};

int OneArg(int x)
{
    return x;
}

int TwoArgs(int x, int y)
{
    return x + y;
}

int NestedArg(Y y)
{
    return *YR::Get(y.obj);
}
class ClusterModeTest : public testing::Test {
public:
    ClusterModeTest(){};
    ~ClusterModeTest(){};

    static void SetUpTestCase()
    {
        YR_INVOKE(A::Create, &A::Add)
        YR_INVOKE(OneArg, TwoArgs, NestedArg)
    }
    void SetUp() override
    {
        MockInit(10);
    }
    void TearDown()
    {
        MockFinalize();
    }

protected:
    Config CreateConfig(uint32_t size)
    {
        Config config;
        config.mode = Config::Mode::CLUSTER_MODE;
        config.dataSystemAddr = "127.0.0.1:31501";
        config.serverAddr = "127.0.0.1:31220";
        config.localThreadPoolSize = size;
        config.logLevel = "DEBUG";
        config.logDir = "/tmp/log";
        return config;
    }

    void MockInit(uint32_t size)
    {
        auto lc = std::make_shared<YR::Libruntime::LibruntimeConfig>();
        lc->jobId = YR::utility::IDGenerator::GenApplicationId();
        auto clientsMgr = std::make_shared<YR::Libruntime::ClientsManager>();
        auto metricsAdaptor = std::make_shared<YR::Libruntime::MetricsAdaptor>();
        auto sec = std::make_shared<YR::Libruntime::Security>();
        auto socketClient = std::make_shared<YR::Libruntime::DomainSocketClient>("/home/snuser/socket/runtime.sock");
        lr = std::make_shared<YR::Libruntime::MockLibruntime>(lc, clientsMgr, metricsAdaptor, sec, socketClient);
        YR::Libruntime::LibruntimeManager::Instance().SetLibRuntime(lr);
        EXPECT_CALL(*lr.get(), GetLocalThreadPoolSize()).Times(AtLeast(0)).WillRepeatedly(Return(size));

        Mkdir("/tmp/log");
        ConfigManager::Singleton().Init(CreateConfig(size), 0, nullptr);
        YR::internal::RuntimeManager::GetInstance().Initialize(std::make_shared<ClusterModeRuntime>());
        SetInitialized(true);
    }

    void MockFinalize()
    {
        YR::internal::RuntimeManager::GetInstance().Stop();
        SetInitialized(false);
        lr = nullptr;
    }

    std::shared_ptr<YR::Libruntime::MockLibruntime> lr;
};

TEST_F(ClusterModeTest, TestHybridClusterInvokeLocalEmptyThreadPool)
{
    MockFinalize();
    MockInit(0);
    YR::InvokeOptions opt;
    opt.alwaysLocalMode = true;
    int x = 1;
    EXPECT_THROW(
        {
            try {
                YR::Function(OneArg).Options(opt).Invoke(x);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot submit task to empty thread pool"));
                throw;
            }
        },
        Exception);

    EXPECT_THROW(
        {
            try {
                YR::Instance(A::Create).Options(opt).Invoke(x);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot submit task to empty thread pool"));
                throw;
            }
        },
        Exception);
}

TEST_F(ClusterModeTest, TestHybridClusterInvokeLocal)
{
    YR::InvokeOptions opt;
    opt.alwaysLocalMode = true;
    int x = 1;
    auto obj = YR::Function(OneArg).Options(opt).Invoke(x);
    EXPECT_EQ(*YR::Get(obj), 1);
    auto ins = YR::Instance(A::Create).Options(opt).Invoke(x);
    auto obj2 = ins.Function(&A::Add).Invoke(x);
    EXPECT_EQ(*YR::Get(obj2), 2);
}

TEST_F(ClusterModeTest, TestHybridClusterPassLocal)
{
    auto x = YR::ObjectRef<int>("123", false, true);  // local object
    // throw on serialize
    EXPECT_THROW(
        {
            try {
                YR::Function(OneArg).Invoke(x);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot serialize local object ref"));
                throw;
            }
        },
        Exception);

    EXPECT_THROW(
        {
            try {
                YR::Instance(A::Create).Invoke(x);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot serialize local object ref"));
                throw;
            }
        },
        Exception);

    EXPECT_THROW(
        {
            try {
                auto ins = YR::Instance(A::Create).Invoke(0);
                auto obj = ins.Function(&A::Add).Invoke(x);
                YR::Get(obj);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot serialize local object ref"));
                throw;
            }
        },
        Exception);
}

TEST_F(ClusterModeTest, TestHybridClusterPassMix)
{
    auto x = YR::ObjectRef<int>("123", false, true);   // local object
    auto y = YR::ObjectRef<int>("124", false, false);  // cluster object
    EXPECT_THROW(
        {
            try {
                YR::Function(TwoArgs).Invoke(x, y);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot serialize local object ref"));
                throw;
            }
        },
        Exception);
}

TEST_F(ClusterModeTest, TestHybridClusterPassNested)
{
    YR::InvokeOptions opt;
    opt.alwaysLocalMode = true;
    int x = 1;
    auto obj = YR::Function(OneArg).Options(opt).Invoke(x);
    Y y;
    y.obj = obj;
    EXPECT_THROW(
        {
            try {
                YR::Function(NestedArg).Invoke(y);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot serialize local object ref"));
                throw;
            }
        },
        Exception);
}

TEST_F(ClusterModeTest, TestHybridLocalPassCluster)
{
    auto buf = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(1));
    auto data = std::make_shared<YR::Libruntime::DataObject>(0, buf->size());
    data->data->MemoryCopy(static_cast<void *>(buf->data()), buf->size());
    EXPECT_CALL(*lr.get(), CreateDataObject(_, _, _, _, _))
        .WillOnce(DoAll(SetArgReferee<2>(data), Return(std::make_pair(YR::Libruntime::ErrorInfo(), "123"))));
    YR::InvokeOptions opt;
    opt.alwaysLocalMode = true;
    auto x = YR::Put(1);
    auto obj = YR::Function(OneArg).Options(opt).Invoke(x);
    EXPECT_THROW(
        {
            try {
                YR::Get(obj);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot pass cluster object ref as local invoke args"));
                throw;
            }
        },
        Exception);

    auto ins = YR::Instance(A::Create).Options(opt).Invoke(x);
    auto obj2 = ins.Function(&A::Add).Invoke(x);
    EXPECT_THROW(
        {
            try {
                YR::Get(obj2);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot pass cluster object ref as local invoke args"));
                throw;
            }
        },
        Exception);
}

TEST_F(ClusterModeTest, TestHybridLocalPassMix)
{
    YR::InvokeOptions opt;
    opt.alwaysLocalMode = true;
    auto x = YR::ObjectRef<int>("123", false, false);
    auto y = YR::ObjectRef<int>("124", false, false);
    auto obj = YR::Function(TwoArgs).Options(opt).Invoke(x, y);
    EXPECT_THROW(
        {
            try {
                YR::Get(obj);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot pass cluster object ref as local invoke args"));
                throw;
            }
        },
        Exception);
}

TEST_F(ClusterModeTest, TestHybridLocalPassNested)
{
    auto buf = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(1));
    auto data = std::make_shared<YR::Libruntime::DataObject>(0, buf->size());
    data->data->MemoryCopy(static_cast<void *>(buf->data()), buf->size());
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> ret;
    ret.push_back(data);
    EXPECT_CALL(*lr.get(), CreateDataObject(_, _, _, _, _))
        .WillOnce(DoAll(SetArgReferee<2>(data), Return(std::make_pair(YR::Libruntime::ErrorInfo(), "123"))));
    Libruntime::RetryInfo retryInfo;
    retryInfo.retryType = Libruntime::RetryType::UNLIMITED_RETRY;
    EXPECT_CALL(*lr.get(), WaitBeforeGet(_, _, _))
        .WillOnce(Return(std::make_pair<YR::Libruntime::ErrorInfo, int64_t>(YR::Libruntime::ErrorInfo(), 1)));
    EXPECT_CALL(*lr.get(), GetDataObjectsWithoutWait(_, _))
        .WillOnce(
            Return(std::make_pair<YR::Libruntime::RetryInfo, std::vector<std::shared_ptr<YR::Libruntime::DataObject>>>(
                std::move(retryInfo), std::move(ret))));
    YR::InvokeOptions opt;
    opt.alwaysLocalMode = true;
    auto obj = YR::Put(1);  // cluster obj
    Y y;
    y.obj = obj;
    auto r = YR::Function(NestedArg).Options(opt).Invoke(y);
    EXPECT_EQ(*YR::Get(r), 1);
}

TEST_F(ClusterModeTest, TestHybridClusterWaitGetLocal)
{
    YR::InvokeOptions opt;
    opt.alwaysLocalMode = true;
    int x = 1;
    std::vector<YR::ObjectRef<int>> results;
    for (int i = 0; i < 5; i++) {
        results.emplace_back(YR::Function(OneArg).Options(opt).Invoke(x));
    }
    auto wait_result = YR::Wait(results, 2, -1);
    EXPECT_GE(wait_result.first.size(), 2);
    EXPECT_EQ(wait_result.first.size() + wait_result.second.size(), 5);
    auto wait_value = YR::Get(results);
    EXPECT_EQ(wait_value.size(), 5);
    for (int i = 0; i < 5; i++) {
        EXPECT_EQ(*wait_value[i], 1);
    }
}

TEST_F(ClusterModeTest, TestHybridClusterWaitGetMix)
{
    auto x = YR::ObjectRef<int>("123", false, true);   // local object
    auto y = YR::ObjectRef<int>("124", false, false);  // cluster object
    std::vector<YR::ObjectRef<int>> v = {x, y};
    EXPECT_THROW(
        {
            try {
                YR::Wait(v, v.size());
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot mix local and cluster object refs"));
                throw;
            }
        },
        Exception);
    EXPECT_THROW(
        {
            try {
                YR::Get(v);
            } catch (const Exception &e) {
                EXPECT_THAT(e.what(), HasSubstr("cannot mix local and cluster object refs"));
                throw;
            }
        },
        Exception);
}

TEST_F(ClusterModeTest, TestGetNodes)
{
    // 定义资源单元数据
    std::unordered_map<std::string, float> capacityMap = {{"cpu", 500.0f}, {"memory", 500.0f}};
    std::unordered_map<std::string, std::vector<std::string>> nodeLabelsMap = {{"NODE_ID", {"node1", "node2"}}};
    std::vector<YR::Libruntime::ResourceUnit> resourceUnitVector = {{
        .id = "node1",
        .capacity = capacityMap,
        .allocatable = capacityMap,
        .nodeLabels = nodeLabelsMap,
        .status = 0,
    }};
    EXPECT_CALL(*lr.get(), GetResources())
        .WillOnce(Return(std::make_pair(YR::Libruntime::ErrorInfo(), resourceUnitVector)));

    std::vector<YR::Node> nodes = YR::Nodes();
    EXPECT_EQ(nodes.size(), 1);
    EXPECT_EQ(nodes[0].id, "node1");
    EXPECT_TRUE(nodes[0].alive);
    EXPECT_EQ(nodes[0].resources["cpu"], 500.0f);
    EXPECT_EQ(nodes[0].resources["memory"], 500.0f);
    EXPECT_EQ(nodes[0].labels["NODE_ID"][0], "node1");
}

TEST_F(ClusterModeRuntimeTest, TestTerminateInstanceAsyncSuccessfully)
{
    EXPECT_CALL(*lr.get(), KillAsync(_, _, _))
        .WillOnce(Invoke(
            [](const std::string &instanceId, int sigNo, std::function<void(const YR::Libruntime::ErrorInfo &err)> cb) {
                cb(YR::Libruntime::ErrorInfo());
            }));
    auto f = rt->TerminateInstanceAsync("111", false);
    auto status = f.wait_for(std::chrono::milliseconds(100));
    EXPECT_EQ(status, std::future_status::ready);
    EXPECT_NO_THROW(f.get());
}

TEST_F(ClusterModeRuntimeTest, TestTerminateInstanceAsyncFailed)
{
    EXPECT_CALL(*lr.get(), KillAsync(_, _, _))
        .WillOnce(Invoke(
            [](const std::string &instanceId, int sigNo, std::function<void(const YR::Libruntime::ErrorInfo &err)> cb) {
                cb(YR::Libruntime::ErrorInfo(YR::Libruntime::ERR_INSTANCE_NOT_FOUND, Libruntime::ModuleCode::RUNTIME,
                                             "111"));
            }));
    auto f = rt->TerminateInstanceAsync("111", false);
    auto status = f.wait_for(std::chrono::milliseconds(100));
    EXPECT_EQ(status, std::future_status::ready);
    EXPECT_THROW(f.get(), YR::Exception);
}

}  // namespace test
}  // namespace YR
