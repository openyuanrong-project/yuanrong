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
#include <boost/beast/http.hpp>
#include "mock/mock_datasystem.h"
#include "mock/mock_fs_intf.h"
#include "mock/mock_invoke_adaptor.h"
#include "mock/mock_security.h"
#include "src/proto/libruntime.pb.h"
#include "src/utility/timer_worker.h"
#define private public
#include "src/libruntime/libruntime.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;
namespace YR {
namespace Libruntime {
extern int g_killTimeout;
}
}  // namespace YR
namespace YR {
namespace test {

using YR::utility::CloseGlobalTimer;
using YR::utility::InitGlobalTimer;
class LibruntimeTest : public testing::Test {
public:
    LibruntimeTest(){};
    ~LibruntimeTest(){};
    void SetUp() override
    {
        g_killTimeout = 1000;
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
        DsConnectOptions opts;
        lc = std::make_shared<YR::Libruntime::LibruntimeConfig>();
        lc->jobId = YR::utility::IDGenerator::GenApplicationId();
        lc->tenantId = "tenantId";
        auto clientsMgr = std::make_shared<YR::Libruntime::ClientsManager>();
        auto metricsAdaptor = std::make_shared<YR::Libruntime::MetricsAdaptor>();
        sec_ = std::make_shared<MockSecurity>();
        auto socketClient = std::make_shared<DomainSocketClient>("/home/snuser/socket/runtime.sock");
        lr = std::make_shared<YR::Libruntime::Libruntime>(lc, clientsMgr, metricsAdaptor, sec_, socketClient);
        gwClient_ = std::make_shared<MockGwClient>();
        auto fsClient = std::make_shared<YR::Libruntime::FSClient>(gwClient_);
        objectStore_ = std::make_shared<MockObjectStore>();
        stateStore_ = std::make_shared<MockStateStore>();
        streamStore_ = std::make_shared<MockStreamStore>();
        heteroStore_ = std::make_shared<MockHeretoStore>();
        auto finalizeHandler = []() { return; };
        DatasystemClients dsclients{objectStore_, stateStore_, streamStore_, heteroStore_};
        lr->Init(fsClient, dsclients, finalizeHandler);
    }

    void TearDown() override
    {
        EXPECT_CALL(*this->gwClient_, KillAsync(_, _, _))
            .WillRepeatedly([=](const YR::Libruntime::KillRequest &, YR::Libruntime::KillCallBack cb, int) {
                if (cb != nullptr) {
                    YR::Libruntime::KillResponse resp;
                    cb(resp, ErrorInfo());
                }
            });
        CloseGlobalTimer();
        if (lr) {
            lr->Finalize(true);
            lr.reset();
        }
        g_killTimeout = 7;
    }

    std::shared_ptr<YR::Libruntime::LibruntimeConfig> lc;
    std::shared_ptr<MockGwClient> gwClient_;
    std::shared_ptr<MockObjectStore> objectStore_;
    std::shared_ptr<MockStateStore> stateStore_;
    std::shared_ptr<MockStreamStore> streamStore_;
    std::shared_ptr<MockHeretoStore> heteroStore_;
    std::shared_ptr<MockSecurity> sec_;
    std::shared_ptr<YR::Libruntime::Libruntime> lr;
};

TEST_F(LibruntimeTest, GetTest)
{
    lr->IncreaseReference({"1", "2"});
    std::vector<std::shared_ptr<Buffer>> retBuffer;
    retBuffer.push_back(std::make_shared<NativeBuffer>(1));
    retBuffer.push_back(std::shared_ptr<Buffer>{});
    auto ret = std::make_pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>>(ErrorInfo(), std::move(retBuffer));

    EXPECT_CALL(*this->objectStore_, Get(Matcher<const std::vector<std::string> &>({"1", "2"}), _))
        .WillOnce(testing::Return(ret));
    auto [errInfo, vec] = lr->Get({"1", "2"}, 5, false);  // 1ok 2err
    EXPECT_EQ(errInfo.Code(), ErrorCode::ERR_GET_OPERATION_FAILED);

    // EXPECT_CALL
    EXPECT_CALL(*this->objectStore_, Get(Matcher<const std::vector<std::string> &>({"1", "2"}), _))
        .WillOnce(testing::Return(ret));
    auto [errInfo2, vec2] = lr->Get({"1", "2"}, 5, true);  // 1ok 2err
    EXPECT_EQ(errInfo2.Code(), ErrorCode::ERR_OK);
    lr->DecreaseReference({"1", "2"});
}

TEST_F(LibruntimeTest, PutTest)
{
    auto lc = std::make_shared<LibruntimeConfig>();
    libruntime::MetaConfig mockMetaConfig;
    lc->loadPaths = {"/tmp/test/123"};
    lc->functionIds.emplace(libruntime::LanguageType::Cpp, "mock-func-id-123");
    lc->BuildMetaConfig(mockMetaConfig);  // Test LibruntimeConfig
    lc->InitConfig(mockMetaConfig);

    lc->jobId = YR::utility::IDGenerator::GenApplicationId();
    auto clientsMgr = std::make_shared<ClientsManager>();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    auto socketClient = std::make_shared<DomainSocketClient>("/home/snuser/socket/runtime.sock");
    lr = std::make_shared<YR::Libruntime::Libruntime>(lc, clientsMgr, metricsAdaptor, sec_, socketClient);
    auto fsClient = std::make_shared<YR::Libruntime::FSClient>(gwClient_);
    DatasystemClients dsclients{objectStore_, stateStore_, streamStore_, heteroStore_};
    lr->Init(fsClient, dsclients);
    std::string str = "Hello, world!";
    auto dataObj = std::make_shared<YR::Libruntime::DataObject>(0, str.size());
    dataObj->data->MemoryCopy(str.data(), str.size());
    auto [errInfo, objId] = lr->Put(dataObj, {});
    ASSERT_EQ(errInfo.Code(), ErrorCode::ERR_OK);
}

TEST_F(LibruntimeTest, InvokeArgTest)
{
    std::string testStr = "1234567890/func";
    std::string testObjid = "test_obj_id";
    std::string testNestedObjId = "test_nested_obj_id";
    InvokeArg arg1, arg2, arg10;
    arg1.dataObj = std::make_shared<YR::Libruntime::DataObject>(0, testStr.size());
    arg1.dataObj->data->MemoryCopy(testStr.data(), testStr.size());
    arg1.objId = testObjid;
    arg1.nestedObjects.emplace(testNestedObjId);

    arg2 = std::move(arg1);
    ASSERT_EQ(testStr, std::string(static_cast<const char *>(arg2.dataObj->data->ImmutableData()),
                                   arg2.dataObj->data->GetSize()));
    ASSERT_EQ(testObjid, arg2.objId);
    ASSERT_EQ(std::unordered_set<std::string>({testNestedObjId}), arg2.nestedObjects);

    InvokeArg arg3 = std::move(arg2);
    ASSERT_EQ(testStr, std::string(static_cast<const char *>(arg3.dataObj->data->ImmutableData()),
                                   arg3.dataObj->data->GetSize()));
    ASSERT_EQ(testObjid, arg3.objId);
    ASSERT_EQ(std::unordered_set<std::string>({testNestedObjId}), arg3.nestedObjects);

    InvokeArg arg4(arg3);
    ASSERT_EQ(testStr, std::string(static_cast<const char *>(arg4.dataObj->data->ImmutableData()),
                                   arg4.dataObj->data->GetSize()));
    ASSERT_EQ(testObjid, arg4.objId);
    ASSERT_EQ(std::unordered_set<std::string>({testNestedObjId}), arg4.nestedObjects);

    arg10 = arg3;
    ASSERT_EQ(testStr, std::string(static_cast<const char *>(arg10.dataObj->data->ImmutableData()),
                                   arg10.dataObj->data->GetSize()));
    ASSERT_EQ(testObjid, arg10.objId);
    ASSERT_EQ(std::unordered_set<std::string>({testNestedObjId}), arg10.nestedObjects);
}

TEST_F(LibruntimeTest, When_Not_Driver_Finalize_Should_Kill_Instances)
{
    auto lc = std::make_shared<LibruntimeConfig>();
    lc->jobId = YR::utility::IDGenerator::GenApplicationId();
    lc->isDriver = false;
    lc->inCluster = true;
    lc->functionSystemRtServerIpAddr = "127.0.0.1";
    lc->functionSystemRtServerPort = 1110;
    lc->dataSystemIpAddr = "127.0.0.1";
    lc->dataSystemPort = 1100;
    auto clientsMgr = std::make_shared<ClientsManager>();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    auto socketClient = std::make_shared<DomainSocketClient>("/home/snuser/socket/runtime.sock");
    lr = std::make_shared<YR::Libruntime::Libruntime>(lc, clientsMgr, metricsAdaptor, sec_, socketClient);
    auto fsClient = std::make_shared<YR::Libruntime::FSClient>(gwClient_);
    DatasystemClients dsclients{objectStore_, stateStore_, streamStore_, heteroStore_};
    lr->Init(fsClient, dsclients);
    EXPECT_NO_THROW(lr->Finalize(false));
}

TEST_F(LibruntimeTest, PodLabelsTest)
{
    auto cfg = LibruntimeConfig();
    InvokeOptions opts;
    opts.podLabels["key1"] = "value1";
    opts.podLabels["key2"] = "value2";
    auto invokeSpec = std::make_shared<InvokeSpec>();
    invokeSpec->opts = opts;
    invokeSpec->requestId = "cae7c30c8d63f5ed00";
    invokeSpec->BuildInstanceCreateRequest(cfg);
    auto podLabelsIter = invokeSpec->requestCreate.createoptions().find("DELEGATE_POD_LABELS");
    ASSERT_NE(podLabelsIter, invokeSpec->requestCreate.createoptions().end());
    nlohmann::json j = nlohmann::json::parse(podLabelsIter->second);
    ASSERT_EQ(j["key1"], "value1");
    ASSERT_EQ(j["key2"], "value2");
}

TEST_F(LibruntimeTest, TenantIdTest)
{
    auto cfg = LibruntimeConfig();
    cfg.tenantId = "test-tenantId";
    auto invokeSpec = std::make_shared<InvokeSpec>();
    invokeSpec->requestId = "cae7c30c8d63f5ed00";
    invokeSpec->BuildInstanceCreateRequest(cfg);
    auto res = invokeSpec->requestCreate.createoptions().at("tenantId");
    ASSERT_EQ(res, "test-tenantId");
}

TEST_F(LibruntimeTest, CreateFailedTest)
{
    FunctionMeta meta;
    std::vector<InvokeArg> invokeArgs;
    InvokeOptions opts;
    opts.podLabels["key1"] = "value1";
    opts.podLabels["key2"] = "value2";
    opts.podLabels["key3"] = "value3";
    opts.podLabels["key4"] = "value4";
    opts.podLabels["key5"] = "value5";
    opts.podLabels["key6"] = "value6";
    auto res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_NE(res.first.Code(), ErrorCode::ERR_OK);
    opts.podLabels.clear();
    opts.podLabels[""] = "value1";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_NE(res.first.Code(), ErrorCode::ERR_OK);
    opts.podLabels.clear();
    opts.podLabels["-aa"] = "value1";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_NE(res.first.Code(), ErrorCode::ERR_OK);
    opts.podLabels.clear();
    opts.podLabels["key1"] = "-aa";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_NE(res.first.Code(), ErrorCode::ERR_OK);
    opts.podLabels.clear();
    opts.podLabels["aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"] = "aa";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_NE(res.first.Code(), ErrorCode::ERR_OK);
    opts.podLabels.clear();
    opts.podLabels["aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"] = "aa";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_EQ(res.first.Msg(),
              "The pod label key is invalid, please set the pod label key with letters, digits and '-' which cannot "
              "start or end with '-' and cannot exceed 63 characters.");
    opts.podLabels.clear();
    opts.podLabels["aa"] = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_EQ(res.first.Msg(),
              "The pod label value is invalid, please set the pod label value with letters, digits and '-' which "
              "cannot start or end with '-' and cannot exceed 63 characters. And empty string can also be set as pod "
              "label value too");

    opts.podLabels.clear();
    opts.customExtensions[CONCURRENCY] = "999";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_EQ(res.first.Code(), ErrorCode::ERR_OK);
    opts.customExtensions[CONCURRENCY] = "1001";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_NE(res.first.Code(), ErrorCode::ERR_OK);
    opts.customExtensions[CONCURRENCY] = "-1";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_NE(res.first.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(res.first.Msg(),
              "invalid opts concurrency, concurrency: -1, please set the concurrency range between 1 and 1000");
    opts.customExtensions[CONCURRENCY] = "1";

    meta.name =
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_EQ(res.first.Code(), ErrorCode::ERR_PARAM_INVALID);
    ASSERT_NE(res.first.Msg().find("exceeds the maximum length of 128 bytes"), std::string::npos);
    meta.name =
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
    meta.ns = "ns";
    res = lr->CreateInstance(meta, invokeArgs, opts);
    ASSERT_EQ(res.first.Code(), ErrorCode::ERR_PARAM_INVALID);
    ASSERT_NE(res.first.Msg().find("exceeds the maximum length of 128 bytes"), std::string::npos);
}

TEST_F(LibruntimeTest, AllocReturnObjectSmallTest)
{
    std::string testObjId("fake_id");
    size_t testDataSize = 100;
    uint64_t totalNativeBufferSize = 0;
    for (int i = 0; i < 2; i++) {
        auto dataObj = std::make_shared<DataObject>(testObjId);
        auto err = lr->AllocReturnObject(dataObj, 0, testDataSize, {}, totalNativeBufferSize);
        ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
        ASSERT_EQ(totalNativeBufferSize, (testDataSize + 16) * (i + 1));
        ASSERT_TRUE(dataObj->data->IsNative());
    }
}

TEST_F(LibruntimeTest, AllocReturnObjectBigTest)
{
    auto lc = std::make_shared<LibruntimeConfig>();
    lc->jobId = YR::utility::IDGenerator::GenApplicationId();
    lc->isDriver = false;
    lc->inCluster = true;
    lc->functionSystemRtServerIpAddr = "127.0.0.1";
    lc->functionSystemRtServerPort = 1110;
    lc->dataSystemIpAddr = "127.0.0.1";
    lc->dataSystemPort = 1100;
    auto clientsMgr = std::make_shared<ClientsManager>();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    auto socketClient = std::make_shared<DomainSocketClient>("/home/snuser/socket/runtime.sock");
    lr = std::make_shared<YR::Libruntime::Libruntime>(lc, clientsMgr, metricsAdaptor, sec_, socketClient);
    auto fsClient = std::make_shared<YR::Libruntime::FSClient>(gwClient_);
    DatasystemClients dsclients{objectStore_, stateStore_, streamStore_, heteroStore_};
    lr->Init(fsClient, dsclients);

    std::string testObjId("fake_id");
    size_t testDataSize = 200000;
    char testData[18] = {"a"};
    auto returnObjs = std::make_shared<SharedBuffer>(static_cast<void *>(testData), 18);
    EXPECT_CALL(*this->objectStore_, CreateBuffer(_, _, _, _))
        .WillRepeatedly(DoAll(SetArgReferee<2>(returnObjs), testing::Return(ErrorInfo())));
    uint64_t totalNativeBufferSize = 0;
    for (int i = 0; i < 2; i++) {
        auto dataObj = std::make_shared<DataObject>(testObjId);
        auto err = lr->AllocReturnObject(dataObj, 0, testDataSize, {}, totalNativeBufferSize);
        ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
        ASSERT_EQ(totalNativeBufferSize, 0);
        ASSERT_FALSE(dataObj->data->IsNative());
    }
}

TEST_F(LibruntimeTest, ConstructTraceIdTest)
{
    struct TestParam {
        std::string userTraceId;
        std::string expectTraceId;
    } tps[] = {
        {
            .userTraceId = "",
            .expectTraceId = YR::utility::IDGenerator::GenTraceId(lc->jobId),
        },
        {
            .userTraceId = "traceid_test",
            .expectTraceId = "traceid_test",
        },
    };

    for (auto &tp : tps) {
        YR::Libruntime::InvokeOptions opts;
        opts.traceId = tp.userTraceId;
        auto traceId = lr->ConstructTraceId(opts);
        ASSERT_EQ(traceId, tp.expectTraceId);
    }
}

TEST_F(LibruntimeTest, NonDriverSecurityInitTest)
{
    auto lc = std::make_shared<LibruntimeConfig>();
    lc->jobId = YR::utility::IDGenerator::GenApplicationId();
    lc->inCluster = true;
    lc->isDriver = false;
    auto clientsMgr = std::make_shared<ClientsManager>();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    auto sec = std::make_shared<MockSecurity>();
    auto socketClient = std::make_shared<DomainSocketClient>("/home/snuser/socket/runtime.sock");
    lr = std::make_shared<YR::Libruntime::Libruntime>(lc, clientsMgr, metricsAdaptor, sec, socketClient);
    auto fsClient = std::make_shared<YR::Libruntime::FSClient>(gwClient_);
    DatasystemClients dsclients{objectStore_, stateStore_, streamStore_, heteroStore_};
    lr->Init(fsClient, dsclients);
}

TEST_F(LibruntimeTest, DriverSecurityInitTest)
{
    auto lc = std::make_shared<LibruntimeConfig>();
    lc->jobId = YR::utility::IDGenerator::GenApplicationId();
    lc->inCluster = true;
    lc->isDriver = true;
    auto clientsMgr = std::make_shared<ClientsManager>();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    auto sec = std::make_shared<MockSecurity>();
    auto socketClient = std::make_shared<DomainSocketClient>("/home/snuser/socket/runtime.sock");
    lr = std::make_shared<YR::Libruntime::Libruntime>(lc, clientsMgr, metricsAdaptor, sec, socketClient);
    auto fsClient = std::make_shared<YR::Libruntime::FSClient>(gwClient_);
    DatasystemClients dsClients{objectStore_, stateStore_, streamStore_, heteroStore_};
    ASSERT_NO_THROW(lr->Init(fsClient, dsClients));
}

TEST_F(LibruntimeTest, CheckSpec)
{
    auto spec = std::make_shared<InvokeSpec>();
    InvokeOptions opts;
    spec->opts = opts;
    auto errorInfo = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(spec->opts.customExtensions["DELEGATE_DIRECTORY_QUOTA"], "512");

    std::unordered_map<std::string, std::string> customExtensions;
    customExtensions["DELEGATE_DIRECTORY_QUOTA"] = "abc";
    opts.customExtensions = customExtensions;
    spec->opts = opts;
    auto errorInfo1 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo1.Code(), ErrorCode::ERR_PARAM_INVALID);

    spec->opts.customExtensions["DELEGATE_DIRECTORY_QUOTA"] = "-1";
    errorInfo = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo.Code(), ErrorCode::ERR_OK);

    spec->opts.customExtensions["DELEGATE_DIRECTORY_QUOTA"] = "-2";
    errorInfo = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo.Code(), ErrorCode::ERR_PARAM_INVALID);

    spec->opts.customExtensions["DELEGATE_DIRECTORY_QUOTA"] = std::to_string(1024 * 1024 + 1);
    errorInfo = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo.Code(), ErrorCode::ERR_PARAM_INVALID);

    spec->opts.customExtensions["DELEGATE_DIRECTORY_QUOTA"] = "0123";
    auto errorInfo2 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo2.Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(spec->opts.customExtensions["DELEGATE_DIRECTORY_QUOTA"], "123");

    spec->opts.recoverRetryTimes = -1;
    auto errorInfo3 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo3.Code(), ErrorCode::ERR_PARAM_INVALID);

    spec->opts.recoverRetryTimes = 1;
    errorInfo3 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo3.Code(), ErrorCode::ERR_OK);

    InstanceRange instanceRange;
    instanceRange.max = 20;
    instanceRange.min = 1;
    instanceRange.step = 0;
    spec->opts.instanceRange = instanceRange;
    auto errorInfo4 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo4.Code(), ErrorCode::ERR_PARAM_INVALID);
    EXPECT_THAT(errorInfo4.Msg(), testing::HasSubstr("please set the step > 0"));

    instanceRange.step = 1;
    spec->opts.instanceRange = instanceRange;
    errorInfo4 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo4.Code(), ErrorCode::ERR_OK);

    RangeOptions rangeOpts;
    rangeOpts.timeout = -2;
    instanceRange.rangeOpts = rangeOpts;
    spec->opts.instanceRange = instanceRange;
    errorInfo4 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo4.Code(), ErrorCode::ERR_PARAM_INVALID);
    EXPECT_THAT(errorInfo4.Msg(), testing::HasSubstr("please set the timeout >= -1"));
}

TEST_F(LibruntimeTest, CheckSpecRGOption)
{
    auto spec = std::make_shared<InvokeSpec>();
    InvokeOptions opts;
    spec->opts = opts;
    ResourceGroupOptions rgOpts;
    rgOpts.resourceGroupName = "primary";
    spec->opts.resourceGroupOpts = rgOpts;
    auto errorInfo5 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo5.Code(), ErrorCode::ERR_PARAM_INVALID);
    EXPECT_THAT(errorInfo5.Msg(), testing::HasSubstr("please set the name other than primary."));

    std::vector<std::unordered_map<std::string, double>> bundles = {
        {{"CPU", 500.0}, {"Memory", 200.0}}, {{"CPU", 300.0}}, {}};
    std::string reqId;
    ResourceGroupSpec resourcegroupSpec;
    resourcegroupSpec.name = "rgname";
    resourcegroupSpec.bundles = bundles;
    errorInfo5 = lr->CreateResourceGroup(resourcegroupSpec, reqId);
    ASSERT_EQ(errorInfo5.Code(), ErrorCode::ERR_OK);

    rgOpts.resourceGroupName = "rgname";
    spec->opts.resourceGroupOpts = rgOpts;
    errorInfo5 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo5.Code(), ErrorCode::ERR_OK);

    rgOpts.bundleIndex = -2;
    spec->opts.resourceGroupOpts = rgOpts;
    errorInfo5 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo5.Code(), ErrorCode::ERR_PARAM_INVALID);

    rgOpts.bundleIndex = 0;
    spec->opts.resourceGroupOpts = rgOpts;
    errorInfo5 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo5.Code(), ErrorCode::ERR_OK);

    rgOpts.bundleIndex = 10;
    spec->opts.resourceGroupOpts = rgOpts;
    errorInfo5 = lr->CheckSpec(spec);
    ASSERT_EQ(errorInfo5.Code(), ErrorCode::ERR_OK);
}

TEST_F(LibruntimeTest, SetTraceIdTest)
{
    std::string traceId = "traceId";
    auto err = lr->SetTraceId(traceId);
    ASSERT_TRUE(err.OK());
    threadLocalTraceId = "threadLocalTraceId";
    err = lr->SetTraceId();
    ASSERT_TRUE(err.OK());
}

TEST_F(LibruntimeTest, GenerateKeyByStateStoreTest)
{
    std::string key;
    EXPECT_CALL(*this->stateStore_, GenerateKey(_))
        .WillOnce(DoAll(SetArgReferee<0>("genKey"), testing::Return(ErrorInfo())));
    lr->GenerateKeyByStateStore(stateStore_, key);
    ASSERT_TRUE("genKey" == key);
}

TEST_F(LibruntimeTest, SetByStateStoreTest)
{
    std::string key = "key";
    auto nativeBuffer = std::make_shared<ReadOnlyNativeBuffer>(nullptr, 0);
    SetParam param;
    auto err = lr->SetByStateStore(stateStore_, key, nativeBuffer, param);
    ASSERT_TRUE(err.OK());
}

TEST_F(LibruntimeTest, SetValueByStateStoreTest)
{
    std::string key = "";
    std::string value = "value";
    EXPECT_CALL(*this->stateStore_, Write(_, _, Matcher<std::string &>(key)))
        .WillOnce(DoAll(SetArgReferee<2>("returnKey"), testing::Return(ErrorInfo())));
    auto nativeBuffer = std::make_shared<ReadOnlyNativeBuffer>(value.c_str(), value.size());
    SetParam param;
    auto err = lr->SetValueByStateStore(stateStore_, nativeBuffer, param, key);
    ASSERT_TRUE(err.OK());
    ASSERT_TRUE("returnKey" == key);
}

TEST_F(LibruntimeTest, GetByStateStoreTest)
{
    std::string key = "rightKey";
    std::shared_ptr<YR::Libruntime::Buffer> ret = std::make_shared<YR::Libruntime::NativeBuffer>(1);
    EXPECT_CALL(*this->stateStore_, Read(Matcher<const std::string &>(key), _))
        .WillOnce(testing::Return(
            std::make_pair<std::shared_ptr<YR::Libruntime::Buffer>, ErrorInfo>(std::move(ret), ErrorInfo())));
    auto result = lr->GetByStateStore(stateStore_, key, 0);
    ASSERT_TRUE(result.second.OK());
    ASSERT_TRUE(result.first != nullptr);

    std::string wrongKey = "wrongKey";
    EXPECT_CALL(*this->stateStore_, Read(Matcher<const std::string &>(wrongKey), _))
        .WillOnce(testing::Return(std::make_pair<std::shared_ptr<YR::Libruntime::Buffer>, ErrorInfo>(
            std::shared_ptr<YR::Libruntime::Buffer>(),
            ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME, ""))));
    result = lr->GetByStateStore(stateStore_, wrongKey, 0);
    ASSERT_TRUE(!result.second.OK());
    ASSERT_TRUE(result.first == nullptr);
}

TEST_F(LibruntimeTest, GetArrayByStateStoreTest)
{
    std::vector<std::shared_ptr<YR::Libruntime::Buffer>> ret;
    ret.push_back(std::make_shared<YR::Libruntime::NativeBuffer>(1));
    ret.push_back(std::shared_ptr<YR::Libruntime::Buffer>());
    std::vector<std::string> keys{"123", "456"};
    EXPECT_CALL(*this->stateStore_, Read(_, _, _))
        .WillOnce(testing::Return(std::make_pair<std::vector<std::shared_ptr<YR::Libruntime::Buffer>>, ErrorInfo>(
            std::move(ret),
            ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME, ""))));

    auto result = lr->GetArrayByStateStore(stateStore_, keys, 0);
    ASSERT_TRUE(!result.second.OK());
    ASSERT_TRUE(result.first[0] != nullptr);
    ASSERT_TRUE(result.first[1] == nullptr);
}

TEST_F(LibruntimeTest, QuerySizeByStateStoreTest)
{
    std::vector<std::string> keys{"123", "456"};
    std::vector<uint64_t> outSizes;
    EXPECT_CALL(*this->stateStore_, QuerySize(_, _))
        .WillOnce([=](const std::vector<std::string> &, std::vector<uint64_t> &param) {
            param = {10, 10};
            return YR::Libruntime::ErrorInfo();
        });
    auto result = lr->QuerySizeByStateStore(stateStore_, keys, outSizes);
    ASSERT_TRUE(result.OK());
    ASSERT_TRUE(outSizes.size() == 2);
}

TEST_F(LibruntimeTest, DelByStateStoreTest)
{
    std::string key;
    auto err = lr->DelByStateStore(stateStore_, key);
    ASSERT_TRUE(err.OK());
}

TEST_F(LibruntimeTest, DelArrayByStateStoreTest)
{
    std::vector<std::string> keys{"123", "456"};
    auto result = lr->DelArrayByStateStore(stateStore_, keys);
    ASSERT_TRUE(result.second.OK());
}

TEST_F(LibruntimeTest, SetTenantIdTest)
{
    EXPECT_NO_THROW(lr->SetTenantIdWithPriority());
    auto tenantID = "tenantId";
    EXPECT_NO_THROW(lr->SetTenantId(tenantID));
}

TEST_F(LibruntimeTest, GetTenantIdTest)
{
    ASSERT_EQ(lr->GetTenantId(), "tenantId");
}

TEST_F(LibruntimeTest, TestInvokeByInstanceIdSuccessfully)
{
    YR::Libruntime::FunctionMeta funcMeta{};
    std::string instanceId = "instanceid";
    std::vector<YR::Libruntime::InvokeArg> args{};
    YR::Libruntime::InvokeArg libArg1;
    libArg1.dataObj = std::make_shared<YR::Libruntime::DataObject>(0, instanceId.size());
    libArg1.dataObj->data->MemoryCopy(instanceId.c_str(), instanceId.size());
    libArg1.isRef = false;
    args.push_back(libArg1);
    YR::Libruntime::InvokeArg libArg2;
    libArg2.dataObj = std::make_shared<YR::Libruntime::DataObject>(0, 0);
    libArg2.isRef = true;
    libArg2.objId = "objId";
    args.push_back(libArg2);
    YR::Libruntime::InvokeOptions opts{};
    std::vector<YR::Libruntime::DataObject> returnObjs{{""}};
    EXPECT_CALL(*this->objectStore_, IncreGlobalReference(_)).WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    auto result1 = lr->IncreaseReference({"objId"});
    ASSERT_TRUE(result1.OK()) << result1.Code() << " " << result1.Msg();
    auto result2 = lr->InvokeByInstanceId(funcMeta, instanceId, args, opts, returnObjs);
    ASSERT_TRUE(result2.OK()) << result2.Code() << " " << result2.Msg();
}

TEST_F(LibruntimeTest, TestCreateDataObject)
{
    auto dataObj = std::make_shared<YR::Libruntime::DataObject>();
    std::vector<std::string> nestedIds;
    EXPECT_CALL(*this->objectStore_, GenerateKey(_, _, _)).WillRepeatedly(testing::Return(YR::Libruntime::ErrorInfo()));
    EXPECT_CALL(*this->objectStore_, IncreGlobalReference(_)).WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    char testData[2] = {"a"};
    auto returnObjs = std::make_shared<SharedBuffer>(static_cast<void *>(testData), 2);
    EXPECT_CALL(*this->objectStore_, CreateBuffer(_, _, _, _))
        .WillOnce(DoAll(SetArgReferee<2>(returnObjs), testing::Return(ErrorInfo())));
    CreateParam param;
    auto result1 = lr->CreateDataObject(0, 0, dataObj, nestedIds, param);
    ASSERT_TRUE(result1.first.OK());
    ASSERT_FALSE(lr->CreateDataObject("objId", 0, 0, dataObj, {"objId"}, param).OK());
}

TEST_F(LibruntimeTest, TestGetRealInstanceId)
{
    std::string objectId = "aaa";
    auto instanceId = lr->GetRealInstanceId(objectId, 1);
    ASSERT_EQ(instanceId, objectId);

    std::string realInstanceId = "bbb";
    lr->SaveRealInstanceId(objectId, realInstanceId);
    instanceId = lr->GetRealInstanceId(objectId, 1);
    ASSERT_EQ(instanceId, realInstanceId);

    InstanceOptions opts;
    opts.needOrder = true;
    std::string objectId2 = "ccc";
    std::string realInstanceId2 = "ddd";
    lr->SaveRealInstanceId(objectId2, realInstanceId2, opts);
    instanceId = lr->GetRealInstanceId(objectId2, 1);
    ASSERT_EQ(instanceId, realInstanceId2);
}

TEST_F(LibruntimeTest, TestInvokeByFunctionName)
{
    YR::Libruntime::FunctionMeta funcMeta{};
    std::string instanceId = "instanceid";
    std::vector<YR::Libruntime::InvokeArg> args{};
    YR::Libruntime::InvokeArg libArg1;
    libArg1.dataObj = std::make_shared<YR::Libruntime::DataObject>(0, instanceId.size());
    libArg1.dataObj->data->MemoryCopy(instanceId.c_str(), instanceId.size());
    libArg1.isRef = false;
    args.push_back(libArg1);
    YR::Libruntime::InvokeArg libArg2;
    libArg2.dataObj = std::make_shared<YR::Libruntime::DataObject>(0, 0);
    libArg2.isRef = true;
    libArg2.objId = "objId";
    args.push_back(libArg2);
    YR::Libruntime::InvokeOptions opts{};
    std::vector<YR::Libruntime::DataObject> returnObjs{{""}};
    EXPECT_CALL(*this->objectStore_, IncreGlobalReference(_)).WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    auto result1 = lr->IncreaseReference({"objId"});
    ASSERT_TRUE(result1.OK()) << result1.Code() << " " << result1.Msg();
    auto result2 = lr->InvokeByFunctionName(funcMeta, args, opts, returnObjs);
    ASSERT_TRUE(result2.OK()) << result2.Code() << " " << result2.Msg();
}

TEST_F(LibruntimeTest, TestCreateInstanceRaw)
{
    CreateRequest req;
    req.set_requestid(YR::utility::IDGenerator::GenRequestId());
    std::string body;
    req.SerializeToString(&body);
    auto buffer = std::make_shared<NativeBuffer>(body.size());
    buffer->MemoryCopy(body.c_str(), body.size());
    auto callback = [](const ErrorInfo &err, std::shared_ptr<Buffer> resultRaw) {};
    EXPECT_CALL(*this->gwClient_, CreateAsync(_, _, _, _)).WillOnce(testing::Return());
    lr->CreateInstanceRaw(buffer, callback);
}

TEST_F(LibruntimeTest, TestInvokeByInstanceIdRaw)
{
    InvokeRequest req;
    req.set_requestid(YR::utility::IDGenerator::GenRequestId());
    std::string body;
    req.SerializeToString(&body);
    auto buffer = std::make_shared<NativeBuffer>(body.size());
    buffer->MemoryCopy(body.c_str(), body.size());
    auto callback = [](const ErrorInfo &err, std::shared_ptr<Buffer> resultRaw) {};
    EXPECT_CALL(*this->gwClient_, InvokeAsync(_, _, _)).WillOnce(testing::Return());
    lr->InvokeByInstanceIdRaw(buffer, callback);
}

TEST_F(LibruntimeTest, TestKillRaw)
{
    KillRequest req;
    std::string body;
    req.SerializeToString(&body);
    auto buffer = std::make_shared<NativeBuffer>(body.size());
    buffer->MemoryCopy(body.c_str(), body.size());
    auto callback = [](const ErrorInfo &err, std::shared_ptr<Buffer> resultRaw) {};
    EXPECT_CALL(*this->gwClient_, KillAsync(_, _, _)).WillOnce(testing::Return());
    lr->KillRaw(buffer, callback);
}

TEST_F(LibruntimeTest, TestPutRaw)
{
    std::string body = "aaa";
    auto buffer = std::make_shared<NativeBuffer>(body.size());
    CreateParam param;
    buffer->MemoryCopy(body.c_str(), body.size());
    EXPECT_CALL(*this->objectStore_, Put(_, Matcher<const std::string &>(body), _, _))
        .WillOnce(testing::Return(ErrorInfo()));
    lr->PutRaw(body, buffer, {}, param);
    lr->dsClients.dsObjectStore = nullptr;
    ASSERT_EQ(lr->PutRaw(body, buffer, {}, param).Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);
}

TEST_F(LibruntimeTest, TestIncreaseReferenceRaw)
{
    std::vector<std::string> objIds;
    ASSERT_EQ(lr->IncreaseReferenceRaw(objIds).OK(), true);
    ASSERT_EQ(lr->IncreaseReferenceRaw(objIds, "remoteId").first.OK(), true);
    EXPECT_CALL(*this->objectStore_, IncreGlobalReference(_)).WillOnce(testing::Return(YR::Libruntime::ErrorInfo()));
    lr->IncreaseReferenceRaw({"aaa"});
    EXPECT_CALL(*this->objectStore_, IncreGlobalReference(_, _))
        .WillOnce(testing::Return(std::make_pair<YR::Libruntime::ErrorInfo, std::vector<std::string>>({}, {})));
    lr->IncreaseReferenceRaw({"aaa"}, "bbb");

    lr->dsClients.dsObjectStore = nullptr;
    ASSERT_EQ(lr->IncreaseReferenceRaw({"aaa"}).Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    ASSERT_EQ(lr->IncreaseReferenceRaw({"aaa"}, "remoteId").first.Code(),
              YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
}

TEST_F(LibruntimeTest, TestDecreaseReferenceRaw)
{
    std::vector<std::string> objIds;
    lr->DecreaseReferenceRaw(objIds);
    lr->DecreaseReferenceRaw(objIds, "remoteId");
    EXPECT_CALL(*this->objectStore_, DecreGlobalReference(_))
        .WillRepeatedly(testing::Return(YR::Libruntime::ErrorInfo()));
    lr->DecreaseReferenceRaw({"aaa"});
    EXPECT_CALL(*this->objectStore_, DecreGlobalReference(_))
        .WillRepeatedly(
            testing::Return(YR::Libruntime::ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, "err")));
    lr->DecreaseReferenceRaw({"aaa"});
    EXPECT_CALL(*this->objectStore_, DecreGlobalReference(_, _))
        .WillRepeatedly(testing::Return(std::make_pair<YR::Libruntime::ErrorInfo, std::vector<std::string>>({}, {})));
    lr->DecreaseReferenceRaw({"aaa"}, "bbb");
    EXPECT_CALL(*this->objectStore_, DecreGlobalReference(_, _))
        .WillRepeatedly(testing::Return(std::make_pair<YR::Libruntime::ErrorInfo, std::vector<std::string>>(
            ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, "err"), {})));
    lr->DecreaseReferenceRaw({"aaa"}, "bbb");
    lr->dsClients.dsObjectStore = nullptr;
    lr->DecreaseReferenceRaw({"aaa"});
    lr->DecreaseReferenceRaw({"aaa"}, "bbb");
}

TEST_F(LibruntimeTest, TestGetRaw)
{
    std::string body = "aaa";
    std::shared_ptr<Buffer> buffer = std::make_shared<NativeBuffer>(body.size());
    buffer->MemoryCopy(body.c_str(), body.size());
    EXPECT_CALL(*this->objectStore_, Get(Matcher<const std::vector<std::string> &>(_), _))
        .WillOnce(testing::Return(std::make_pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>>({}, {buffer})));
    lr->GetRaw({"aaa"}, 30, true);
    lr->dsClients.dsObjectStore = nullptr;
    ASSERT_EQ(lr->GetRaw({"aaa"}, 30, true).first.OK(), false);
}

TEST_F(LibruntimeTest, GetCredentialTest)
{
    datasystem::SensitiveValue sk = std::string("sk");
    lr->security_->SetAKSKAndCredential("ak", sk);
    auto result = lr->GetCredential();
    ASSERT_EQ(result.ak, "ak");
    ASSERT_EQ(result.sk, "sk");
}

TEST_F(LibruntimeTest, FiberEventTest)
{
    FiberEventNotify event;
    int value = 0;
    std::shared_ptr<YR::Libruntime::Libruntime> lb = lr;
    std::thread t1([&event, &value, lb]() { lb->WaitEvent(event); });

    std::thread t2([&event, &value, lb]() {
        value++;
        lb->NotifyEvent(event);
    });
    t2.join();
    t1.join();
    ASSERT_EQ(value, 1);
}

TEST_F(LibruntimeTest, HeteroDeleteTest)
{
    std::vector<std::string> objectIds;
    std::vector<std::string> failedObjectIds;
    ASSERT_EQ(lr->DevDelete(objectIds, failedObjectIds).OK(), true);
}

TEST_F(LibruntimeTest, HeteroDevLocalDeleteTest)
{
    std::vector<std::string> objectIds;
    std::vector<std::string> failedObjectIds;
    ASSERT_EQ(lr->DevLocalDelete(objectIds, failedObjectIds).OK(), true);
}

TEST_F(LibruntimeTest, HeteroDevSubscribeTest)
{
    std::vector<std::string> keys;
    std::vector<DeviceBlobList> blob2dList;
    std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> futureVec;
    ASSERT_EQ(lr->DevSubscribe(keys, blob2dList, futureVec).OK(), true);
}

TEST_F(LibruntimeTest, HeteroDevPublishTest)
{
    std::vector<std::string> keys;
    std::vector<DeviceBlobList> blob2dList;
    std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> futureVec;
    ASSERT_EQ(lr->DevPublish(keys, blob2dList, futureVec).OK(), true);
}

TEST_F(LibruntimeTest, HeteroDevMSetTest)
{
    std::vector<std::string> keys;
    std::vector<DeviceBlobList> blob2dList;
    std::vector<std::string> failedKeys;
    ASSERT_EQ(lr->DevMSet(keys, blob2dList, failedKeys).OK(), true);
}

TEST_F(LibruntimeTest, HeteroDevMGetTest)
{
    std::vector<std::string> keys;
    std::vector<DeviceBlobList> blob2dList;
    std::vector<std::string> failedKeys;
    ASSERT_EQ(lr->DevMGet(keys, blob2dList, failedKeys, 1000).OK(), true);
}

TEST_F(LibruntimeTest, StreamProducerAndConsumerTest)
{
    std::string streamName = "streamname";
    uint64_t value = 1000;
    ProducerConf producerConf;
    std::shared_ptr<StreamProducer> producer = std::make_shared<StreamProducer>();
    ASSERT_EQ(lr->CreateStreamProducer(streamName, producerConf, producer).OK(), true);

    SubscriptionConfig config;
    std::shared_ptr<StreamConsumer> consumer = std::make_shared<StreamConsumer>();
    ASSERT_EQ(lr->CreateStreamConsumer(streamName, config, consumer, true).OK(), true);
    ASSERT_EQ(lr->DeleteStream(streamName).OK(), true);
    ASSERT_EQ(lr->QueryGlobalProducersNum(streamName, value).OK(), true);
    ASSERT_EQ(lr->QueryGlobalConsumersNum(streamName, value).OK(), true);

    lr->dsClients.dsStreamStore = nullptr;
    ASSERT_EQ(lr->CreateStreamProducer(streamName, producerConf, producer).Code(),
              YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    ASSERT_EQ(lr->CreateStreamConsumer(streamName, config, consumer, true).Code(),
              YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    ASSERT_EQ(lr->DeleteStream(streamName).Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    ASSERT_EQ(lr->QueryGlobalProducersNum(streamName, value).Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    ASSERT_EQ(lr->QueryGlobalConsumersNum(streamName, value).Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
}

TEST_F(LibruntimeTest, SetTenatIdTest)
{
    lr->config->enableAuth = true;
    std::string tenantId = "";
    bool isReturnErrWhenTenantIDEmpty = true;
    ASSERT_EQ(lr->SetTenantId(tenantId, isReturnErrWhenTenantIDEmpty).Code(),
              YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);

    std::string id = "tenantId";
    isReturnErrWhenTenantIDEmpty = false;
    ASSERT_EQ(lr->SetTenantId(id, isReturnErrWhenTenantIDEmpty).OK(), true);

    lr->dsClients.dsObjectStore = nullptr;
    ASSERT_EQ(lr->SetTenantId(id, isReturnErrWhenTenantIDEmpty).Code(),
              YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
}

TEST_F(LibruntimeTest, GetInstancesTest)
{
    std::string objId = "objId";
    auto [insIds, err] = lr->GetInstances(objId, 60);
    ASSERT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);

    lr->invokeAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    ASSERT_EQ(lr->GetInstances("objId", "groupName").second.OK(), true);
}

TEST_F(LibruntimeTest, GetInstanceTest)
{
    auto adaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    lr->invokeAdaptor = adaptor;
    EXPECT_CALL(*adaptor, GetInstance(_, _, _))
        .WillOnce(Return(
            std::make_pair(YR::Libruntime::FunctionMeta{.name = "name", .ns = "ns", .needOrder = true}, ErrorInfo())));
    auto res = lr->GetInstance(std::string("name"), std::string("namespace"), 300);
    ASSERT_EQ(res.first.name, "name");
}

TEST_F(LibruntimeTest, AcquireAndReleaseInstanceTest)
{
    lr->invokeAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    FunctionMeta meta;
    InvokeOptions opts;
    ASSERT_EQ(lr->AcquireInstance("stateId", meta, opts).second.OK(), true);
    ASSERT_EQ(lr->ReleaseInstance("leaseId", "stateId", true, opts).OK(), true);
}

TEST_F(LibruntimeTest, ExecShutdownCallbackTest)
{
    lr->invokeAdaptor = nullptr;
    ASSERT_EQ(lr->ExecShutdownCallback(600).Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    lr->invokeAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    ASSERT_EQ(lr->ExecShutdownCallback(600).OK(), true);
}

TEST_F(LibruntimeTest, SaveAndLoadStateTest)
{
    lr->invokeAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    std::shared_ptr<Buffer> buffer = std::make_shared<NativeBuffer>(16);
    ASSERT_EQ(lr->SaveState(buffer, 300).OK(), true);
    ASSERT_EQ(lr->LoadState(buffer, 300).OK(), true);
}

TEST_F(LibruntimeTest, GroupTest)
{
    lr->invokeAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    GroupOpts opts;
    ASSERT_EQ(lr->GroupCreate("groupName", opts).OK(), true);
    ASSERT_EQ(lr->GroupWait("groupName").OK(), true);
    ASSERT_NO_THROW(lr->GroupTerminate("groupName"));
}

TEST_F(LibruntimeTest, CancelTest)
{
    lr->invokeAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    ASSERT_EQ(lr->Cancel({"objId"}, true, true).OK(), true);
}

TEST_F(LibruntimeTest, ReceiveRequestLoopTest)
{
    lr->invokeAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    ASSERT_NO_THROW(lr->ReceiveRequestLoop());
}

TEST_F(LibruntimeTest, UpdateSchdulerInfoTest)
{
    lr->invokeAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    ASSERT_NO_THROW(lr->UpdateSchdulerInfo("scheduleName", "schedulerId", "ADD"));
}

TEST_F(LibruntimeTest, SaveGroupInstanceIdsTest)
{
    const std::string groupInsIds = "aa;bb;cc";
    const std::string objId = "objId";
    YR::Libruntime::InstanceOptions opts;
    opts.needOrder = true;
    lr->SaveGroupInstanceIds(objId, groupInsIds, opts);
    ASSERT_EQ(lr->memStore->GetInstanceIds("objId", 60).first[0], "aa");
}

TEST_F(LibruntimeTest, PutOKTest)
{
    std::string objID = "objID";
    std::shared_ptr<DataObject> dataObj = std::make_shared<DataObject>();
    std::unordered_set<std::string> nestedID;
    CreateParam createParam;
    auto err = lr->Put(objID, dataObj, nestedID, createParam);
    ASSERT_EQ(err.OK(), false);

    std::shared_ptr<Buffer> buffer = std::make_shared<NativeBuffer>(16);
    err = lr->Put(buffer, "objId_1", nestedID, false, createParam);
    ASSERT_EQ(err.OK(), false);
}

TEST_F(LibruntimeTest, IncreaseReferenceTest)
{
    std::vector<std::string> objIds{"objId"};
    std::string remoteId = "remoteId";
    ASSERT_EQ(lr->IncreaseReference(objIds).OK(), true);
    ASSERT_EQ(lr->IncreaseReference(objIds, remoteId).first.OK(), true);
}

TEST_F(LibruntimeTest, DecreaseReferenceTest)
{
    std::vector<std::string> objIds{"objId"};
    EXPECT_CALL(*this->objectStore_, DecreGlobalReference(_))
        .WillRepeatedly(testing::Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                                                  ModuleCode::RUNTIME, "err increase")));
    lr->DecreaseReference(objIds);
    EXPECT_CALL(*this->objectStore_, DecreGlobalReference(_))
        .WillRepeatedly(
            testing::Return(YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                                                      "err increase", true, {StackTraceInfo()})));
    lr->DecreaseReference(objIds, "remoteId");

    lr->memStore = nullptr;
    EXPECT_NO_THROW(lr->DecreaseReference(objIds));
}

TEST_F(LibruntimeTest, ReleaseGRefsTest)
{
    ASSERT_EQ(lr->ReleaseGRefs("remoteId").OK(), true);
}

TEST_F(LibruntimeTest, WaitTest)
{
    ASSERT_EQ(lr->Wait({"objId"}, 1, 0)->readyIds.size(), 1);
}

TEST_F(LibruntimeTest, GetBuffersTest)
{
    std::vector<std::string> ids{std::string("objId")};
    lr->AddReturnObject(ids);
    lr->SetError(ids[0], ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, "err"));
    ASSERT_EQ(lr->GetBuffers(ids, 300, false).first.OK(), false);
}

TEST_F(LibruntimeTest, PeekObjectRefStreamTest)
{
    ASSERT_EQ(lr->PeekObjectRefStream("generatorId", false).first.OK(), false);
}

TEST_F(LibruntimeTest, GetFunctionGroupRunningInfoTest)
{
    ASSERT_EQ(lr->GetFunctionGroupRunningInfo().instanceRankId, 0);
}

TEST_F(LibruntimeTest, GenerateGroupNameTest)
{
    ASSERT_EQ(lr->GenerateGroupName().empty(), false);
}

TEST_F(LibruntimeTest, IsObjectExistingInLocalTest)
{
    ASSERT_EQ(lr->IsObjectExistingInLocal("objId"), false);
}

TEST_F(LibruntimeTest, WaitAndGetAsyncTest)
{
    auto waitPromise = std::promise<ErrorInfo>();
    auto waitFut = waitPromise.get_future();
    YR::Libruntime::WaitAsyncCallback cbWait = [&waitPromise](const std::string &id, const ErrorInfo &err, void *data) {
        waitPromise.set_value(err);
    };
    char *str = "hello";
    lr->WaitAsync("objId", cbWait, str);
    ASSERT_EQ(waitFut.get().OK(), true);

    auto getPromise = std::promise<ErrorInfo>();
    auto getFut = getPromise.get_future();
    YR::Libruntime::GetAsyncCallback cbGet = [&getPromise](const std::shared_ptr<DataObject> &dataObj,
                                                           const ErrorInfo &err,
                                                           void *data) { getPromise.set_value(err); };
    lr->GetAsync("objId", cbGet, str);
    ASSERT_EQ(getFut.get().OK(), true);
}

TEST_F(LibruntimeTest, GetEventTest)
{
    YR::Libruntime::GetEventCallback callback = [](const std::shared_ptr<DataObject>& dataObj,
                                                           const ErrorInfo& err,
                                                           void* userData) {
        YRLOG_DEBUG("GetEvent callback is called");
    };
    std::string objectId = "testObjectId";
    void* userData = nullptr;
    EXPECT_NO_THROW(lr->GetEvent(objectId, callback, userData));
}

TEST_F(LibruntimeTest, GetGroupInstanceIdsTest)
{
    ASSERT_EQ(lr->GetGroupInstanceIds("objId", 100).empty(), true);
}

TEST_F(LibruntimeTest, ExitTest)
{
    EXPECT_CALL(*this->gwClient_, ExitAsync(_, _))
        .WillOnce([=](const YR::Libruntime::ExitRequest &, YR::Libruntime::ExitCallBack cb) {
            if (cb != nullptr) {
                YR::Libruntime::ExitResponse resp;
                cb(resp);
            }
        });
    EXPECT_NO_THROW(lr->Exit());
}

TEST_F(LibruntimeTest, GetDataObjectsTest)
{
    ASSERT_EQ(lr->GetDataObjects({"objId"}, 300, false).first.OK(), false);
}

TEST_F(LibruntimeTest, GetBuffersWithoutWaitTest)
{
    ASSERT_EQ(lr->GetBuffersWithoutWait({"objId"}, 300).first.errorInfo.OK(), true);
}

TEST_F(LibruntimeTest, GetDataObjectsWithoutWaitTest)
{
    ASSERT_EQ(lr->GetDataObjectsWithoutWait({"objId"}, 300).first.errorInfo.OK(), true);
}

TEST_F(LibruntimeTest, CreateBufferTest)
{
    std::shared_ptr<Buffer> buffer = std::make_shared<NativeBuffer>(16);
    ASSERT_EQ(lr->CreateBuffer(1, buffer).first.OK(), true);
}

TEST_F(LibruntimeTest, ProcessErrTest)
{
    std::shared_ptr<InvokeSpec> spec = std::make_shared<InvokeSpec>();
    ASSERT_NO_THROW(lr->ProcessErr(spec, ErrorInfo()));
}

TEST_F(LibruntimeTest, FinalizeHandlerTest)
{
    ASSERT_NO_THROW(lr->FinalizeHandler());
}

TEST_F(LibruntimeTest, GetServerVersionTest)
{
    ASSERT_EQ(lr->GetServerVersion().empty(), true);
}

TEST_F(LibruntimeTest, KillTest)
{
    lr->invokeAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    ASSERT_EQ(lr->Kill("instanceId", 1).OK(), true);
}

TEST_F(LibruntimeTest, GetThreadPoolSizeTest)
{
    ASSERT_EQ(lr->GetThreadPoolSize(), 0);
    ASSERT_EQ(lr->GetLocalThreadPoolSize(), 0);
}

TEST_F(LibruntimeTest, ResourceGroupTest)
{
    EXPECT_CALL(*this->gwClient_, CreateRGroupAsync(_, _, _))
        .WillOnce([=](const YR::Libruntime::CreateResourceGroupRequest &,
                      YR::Libruntime::CreateResourceGroupCallBack cb, int) {
            if (cb != nullptr) {
                YR::Libruntime::CreateResourceGroupResponse resp;
                resp.set_code(static_cast<common::ErrorCode>(1001));
                resp.set_message("error");
                cb(resp);
            }
        });
    EXPECT_CALL(*this->gwClient_, KillAsync(_, _, _))
        .WillOnce([=](const YR::Libruntime::KillRequest &, YR::Libruntime::KillCallBack cb, int) {
            if (cb != nullptr) {
                YR::Libruntime::KillResponse resp;
                resp.set_code(static_cast<common::ErrorCode>(1001));
                cb(resp, ErrorInfo());
            }
        });
    std::vector<std::unordered_map<std::string, double>> bundles1 = {
        {{"CPU", 500.0}, {"Memory", 200.0}}, {{"CPU", 300.0}}, {}};
    std::vector<std::unordered_map<std::string, double>> bundles2 = {{{"CPU", -500.0}, {"Memory", 200.0}}};
    std::vector<std::unordered_map<std::string, double>> bundles3 = {{{"", 500.0}, {"Memory", 200.0}}};
    std::string reqId;
    ResourceGroupSpec resourcegroupSpec;
    resourcegroupSpec.name = "";
    resourcegroupSpec.bundles = bundles1;
    auto err = lr->CreateResourceGroup(resourcegroupSpec, reqId);
    ASSERT_EQ(err.Msg(), "invalid resource group name, name: , please set the name other than primary or empty.");
    resourcegroupSpec.name = "rgname";
    resourcegroupSpec.bundles = bundles2;
    err = lr->CreateResourceGroup(resourcegroupSpec, reqId);
    ASSERT_EQ(err.Msg(), "invalid bundle, bundle index: 0, please set the value of CPU >= 0.");
    resourcegroupSpec.bundles = bundles3;
    err = lr->CreateResourceGroup(resourcegroupSpec, reqId);
    ASSERT_EQ(err.Msg(), "invalid bundle, bundle index: 0, please set a non-empty and correct key.");
    resourcegroupSpec.bundles = bundles1;
    err = lr->CreateResourceGroup(resourcegroupSpec, reqId);
    ASSERT_EQ(err.Code(), 0);

    err = lr->RemoveResourceGroup("rgname");
    ASSERT_EQ(err.Code(), 0);
    err = lr->RemoveResourceGroup("primary");
    ASSERT_EQ(err.Msg(),
              "invalid resource group name, name: primary, please set the name other than primary or empty.");
}

TEST_F(LibruntimeTest, KVTest)
{
    SetParam setParam;
    MSetParam mSetParam;
    GetParams getParam;
    std::string key = "kv-key";
    std::string vstr = "kv-value";
    std::shared_ptr<msgpack::sbuffer> sbuf = std::make_shared<msgpack::sbuffer>();
    sbuf->write(vstr.c_str(), vstr.length());
    auto value = std::make_shared<YR::Libruntime::MsgpackBuffer>(sbuf);
    std::vector<uint64_t> outSizes;
    auto err = lr->KVWrite(key, value, setParam);
    ASSERT_EQ(err.Code(), 0);
    std::vector<std::string> keys = {key};
    std::vector<std::shared_ptr<Buffer>> vals = {value};
    err = lr->KVMSetTx(keys, vals, mSetParam);
    ASSERT_EQ(err.Code(), 0);
    auto [res1, err1] = lr->KVRead(key, 10);
    ASSERT_EQ(err1.Code(), 0);
    auto [res2, err2] = lr->KVRead(keys, 10, true);
    ASSERT_EQ(err2.Code(), 0);
    auto [res3, err3] = lr->KVGetWithParam(keys, getParam, 10);
    ASSERT_EQ(err3.Code(), 0);
    err = lr->KVDel(key);
    ASSERT_EQ(err.Code(), 0);
    auto [res4, err4] = lr->KVDel(keys);
    ASSERT_EQ(err4.Code(), 0);
    auto [res5, err5] = lr->KVExist(keys);
    ASSERT_EQ(err5.Code(), 0);
}

TEST_F(LibruntimeTest, TestAccelerate)
{
    g_killTimeout = 1;
    std::string groupName = "group";
    AccelerateMsgQueueHandle handle;
    HandleReturnObjectCallback callback;
    auto ret = lr->Accelerate(groupName, handle, callback);
    ASSERT_EQ(ret.Code(), ErrorCode::ERR_PARAM_INVALID);
}

TEST_F(LibruntimeTest, TestIsLocalInstances)
{
    g_killTimeout = 1;
    std::vector<std::string> instanceIds = {"instance_1", "instance_2"};
    auto ret = lr->IsLocalInstances(instanceIds);
    ASSERT_FALSE(ret);
}

TEST_F(LibruntimeTest, TestIsDsHealth)
{
    lr->dsClients.dsStreamStore = nullptr;
    lr->dsClients.dsStateStore = nullptr;
    ASSERT_TRUE(lr->IsDsHealth());
    auto mockstateStore = std::make_shared<MockStateStore>();
    lr->dsClients.dsStateStore = mockstateStore;
    auto mockStreamStore = std::make_shared<MockStreamStore>();
    lr->dsClients.dsStreamStore = mockStreamStore;
    EXPECT_CALL(*mockstateStore, HealthCheck()).WillOnce(Return(ErrorInfo()));
    ASSERT_TRUE(lr->IsDsHealth());
}

TEST_F(LibruntimeTest, TestIsHealth)
{
    lr->invokeAdaptor = nullptr;
    ASSERT_FALSE(lr->IsHealth());
    auto mockInvokeAdaptor = std::make_shared<MockInvokeAdaptor>();
    lr->invokeAdaptor = mockInvokeAdaptor;
    EXPECT_CALL(*mockInvokeAdaptor, IsHealth()).WillOnce(Return(true));
    ASSERT_TRUE(lr->IsHealth());
}

TEST_F(LibruntimeTest, KillAsyncTest)
{
    auto mock_adaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    lr->invokeAdaptor = mock_adaptor;
    EXPECT_CALL(*mock_adaptor, KillAsyncCB(_, _, _, _, _))
        .WillOnce(Invoke([](const std::string &instanceId, const std::string &payload, int signal,
                            std::function<void(const ErrorInfo &err)> cb, int timeoutSec) { cb(YR::Libruntime::ErrorInfo()); }));
    auto promise = std::make_shared<std::promise<ErrorInfo>>();
    auto f = promise->get_future();
    lr->KillAsync("instanceId", 1, [promise](const ErrorInfo &err) { promise->set_value(err); });
    auto status = f.wait_for(std::chrono::milliseconds(100));
    EXPECT_EQ(status, std::future_status::ready);
    EXPECT_TRUE(f.get().OK());
}

TEST_F(LibruntimeTest, GetRequestAndInstanceIDTest)
{
    threadLocalRequestId = "req123";
    threadLocalInstanceId = "inst456";

    auto result = lr->GetRequestAndInstanceID();

    EXPECT_EQ(result.first, "req123");
    EXPECT_EQ(result.second, "inst456");

    threadLocalRequestId.clear();
    threadLocalInstanceId.clear();

    result = lr->GetRequestAndInstanceID();

    EXPECT_TRUE(result.first.empty());
    EXPECT_TRUE(result.second.empty());
}

TEST_F(LibruntimeTest, SessionWaitNotifyTest)
{
    lr->invokeAdaptor = nullptr;
    auto [err1, buf1] = lr->SessionWait("sessionId", 100);
    ASSERT_EQ(err1.Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);
    ASSERT_EQ(buf1, nullptr);
    ASSERT_EQ(lr->SessionNotify("sessionId", nullptr).Code(), ErrorCode::ERR_INNER_SYSTEM_ERROR);

    auto mockAdaptor = std::make_shared<YR::Libruntime::MockInvokeAdaptor>();
    lr->invokeAdaptor = mockAdaptor;
    std::string payload = "notify";
    auto data = std::make_shared<StringNativeBuffer>(payload.size());
    ASSERT_EQ(data->MemoryCopy(payload.data(), payload.size()).Code(), ErrorCode::ERR_OK);
    auto bufferData = std::static_pointer_cast<Buffer>(data);
    EXPECT_CALL(*mockAdaptor, SessionWait("sessionId", 200))
        .WillOnce(Return(std::make_pair(ErrorInfo(), bufferData)));
    EXPECT_CALL(*mockAdaptor, SessionNotify("sessionId", bufferData)).WillOnce(Return(ErrorInfo()));

    auto [err2, buf2] = lr->SessionWait("sessionId", 200);
    ASSERT_EQ(err2.Code(), ErrorCode::ERR_OK);
    ASSERT_NE(buf2, nullptr);
    ASSERT_EQ(lr->SessionNotify("sessionId", bufferData).Code(), ErrorCode::ERR_OK);
}

}  // namespace test
}  // namespace YR
