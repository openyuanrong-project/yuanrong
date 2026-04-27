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

#include <algorithm>
#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <boost/beast/http.hpp>
#define private public
#include "src/libruntime/err_type.h"
#include "src/libruntime/fmclient/fm_client.h"
#include "src/libruntime/gwclient/gw_client.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

class MockHttpClient : public HttpClient {
public:
    MockHttpClient() = default;
    ~MockHttpClient() = default;
    ErrorInfo Init(const ConnectionParam &param) override
    {
        ResetConnActive();
        return ErrorInfo();
    }

    void SubmitInvokeRequest(const http::verb &method, const std::string &target,
                             const std::unordered_map<std::string, std::string> &headers, const std::string &body,
                             const std::shared_ptr<std::string> requestId,
                             const HttpCallbackFunction &receiver) override
    {
        if (GLOBAL_SCHEDULER_QUERY_RESOURCES == target) {
            auto rsp = BuildQueryResponse();
            std::string rspBody;
            rsp.SerializeToString(&rspBody);
            if (isSuccess_) {
                receiver(rspBody, boost::beast::error_code(), 200);
            } else {
                receiver(rspBody, boost::beast::error_code(), 400);
            }
        } else {
            std::cout << "unknown target:" << target;
        }
    }

    ErrorInfo ReInit(const std::shared_ptr<std::string> requestId) override
    {
        return ErrorInfo();
    }

    QueryResourcesInfoResponse BuildQueryResponse()
    {
        QueryResourcesInfoResponse rsp;
        rsp.set_requestid("requestid");
        auto *resource = rsp.mutable_resource();
        resource->set_id("resourceid");
        resource->set_status(0);
        resources::ResourceUnit ru;
        ru.set_id("id");
        ru.set_status(0);

        resources::Resource r1;
        r1.set_type(ResourceType::Value_Type_SCALAR);
        r1.mutable_scalar()->set_value(400.00);
        ru.mutable_capacity()->mutable_resources()->insert({"CPU", r1});

        resources::Resource r2;
        r2.set_type(ResourceType::Value_Type_VECTORS);
        resources::Value_Vectors_Vector ve2;
        ve2.add_values(333.14);
        ve2.add_values(222.88);
        resources::Value_Vectors_Category c2;
        c2.mutable_vectors()->insert({"uuid", ve2});
        r2.mutable_vectors()->mutable_values()->insert({"ids", c2});
        ru.mutable_capacity()->mutable_resources()->insert({"NPU", r2});

        resources::Value::Counter cnter1;
        cnter1.mutable_items()->insert( { "value1_1", 1 } );
        cnter1.mutable_items()->insert( { "value1_2", 1 } );
        ru.mutable_nodelabels()->insert({ "key1", cnter1 });
        resources::Value::Counter cnter2;
        cnter2.mutable_items()->insert( { "value2", 3 } );
        ru.mutable_nodelabels()->insert({ "key2", cnter2 });

        resource->mutable_fragment()->insert({"resource", ru});
        return rsp;
    }
    bool isSuccess_ = true;
};

class FmClientTest : public testing::Test {
public:
    FmClientTest(){};
    ~FmClientTest(){};
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
    }
    void TearDown() override {}
};

TEST_F(FmClientTest, TestGetResourcesSuccessfully)
{
    std::shared_ptr<FMClient> fmClient_ = std::make_shared<FMClient>();
    auto httpClient = std::make_shared<MockHttpClient>();
    httpClient->Init(ConnectionParam{"127.0.0.1", "8888"});
    httpClient->SetAvailable();
    fmClient_->UpdateActiveMaster("127.0.0.1:8888");
    fmClient_->activeMasterHttpClient_ = httpClient;
    auto res = fmClient_->GetResources();
    EXPECT_TRUE(res.first.OK());
    auto unit = res.second.at(0);
    EXPECT_TRUE(unit.capacity["CPU"] == 400.00);
    EXPECT_TRUE(unit.capacity["NPU"] == 2.00);
    EXPECT_TRUE(unit.nodeLabels.size() == 2);
    EXPECT_TRUE(unit.nodeLabels.count("key1"));
    EXPECT_TRUE(unit.nodeLabels["key1"].size() == 2);
    EXPECT_TRUE(std::find(unit.nodeLabels["key1"].begin(), unit.nodeLabels["key1"].end(), "value1_1") !=
        unit.nodeLabels["key1"].end());
    EXPECT_TRUE(std::find(unit.nodeLabels["key1"].begin(), unit.nodeLabels["key1"].end(), "value1_2") !=
        unit.nodeLabels["key1"].end());
    EXPECT_TRUE(unit.nodeLabels.count("key2"));
    EXPECT_TRUE(std::find(unit.nodeLabels["key2"].begin(), unit.nodeLabels["key2"].end(), "value2") !=
        unit.nodeLabels["key2"].end());

    httpClient->isSuccess_ = false;
    fmClient_->cb_ = [](){};
    fmClient_->maxWaitTimeSec = 1;
    res = fmClient_->GetResources();
    EXPECT_TRUE(!res.first.OK());
}

TEST_F(FmClientTest, TestGetResourcesWithRetryFailed)
{
    std::shared_ptr<FMClient> fmClient_ = std::make_shared<FMClient>();
    fmClient_->UpdateActiveMaster("127.0.0.1");
    auto res = fmClient_->GetResourcesWithRetry();
    EXPECT_TRUE(!res.first.OK());
}

TEST_F(FmClientTest, TestUpdateActiveMasterWithStopSuccessfully)
{
    std::shared_ptr<FMClient> fmClient_ = std::make_shared<FMClient>();
    fmClient_->Stop();
    EXPECT_NO_THROW(fmClient_->UpdateActiveMaster("127.0.0.1"));
}

TEST_F(FmClientTest, TestCheckResponseCode)
{
    // Test case 1: errorCode is set
    boost::beast::error_code errorCode(1, boost::system::system_category());
    uint statusCode = 200;
    std::string result = "Success";
    std::string requestId = "12345";
    ErrorInfo errorInfo = YR::Libruntime::CheckResponseCode(errorCode, statusCode, result, requestId);
    EXPECT_TRUE(errorInfo.Code() == YR::Libruntime::ErrorCode::ERR_INNER_COMMUNICATION);
    EXPECT_TRUE(errorInfo.Msg() == "network error between runtime and function master, error_code: Operation not permitted, requestId: 12345");

    // Test case 2: statusCode is not successful
    errorCode = boost::beast::error_code();
    statusCode = 400;
    result = "Bad Request";
    requestId = "67890";
    errorInfo = YR::Libruntime::CheckResponseCode(errorCode, statusCode, result, requestId);
    EXPECT_TRUE(errorInfo.Code() == YR::Libruntime::ErrorCode::ERR_PARAM_INVALID);
    EXPECT_TRUE(errorInfo.Msg() == "response is error, status_code: 400, result: Bad Request, requestId: 67890");

    // Test case 3: no error
    errorCode = boost::beast::error_code();
    statusCode = 200;
    result = "Success";
    requestId = "12345";
    errorInfo = YR::Libruntime::CheckResponseCode(errorCode, statusCode, result, requestId);
    EXPECT_TRUE(errorInfo.OK());
}

TEST_F(FmClientTest, TestRemoveActiveMaster)
{
    std::shared_ptr<FMClient> fmClient_ = std::make_shared<FMClient>();
    fmClient_->activeMasterAddr_ = "activeMasterAddr_";
    fmClient_->RemoveActiveMaster();
    EXPECT_TRUE(fmClient_->activeMasterAddr_.empty());
}