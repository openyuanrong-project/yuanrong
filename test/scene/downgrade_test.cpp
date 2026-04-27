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
#include <boost/beast/http.hpp>
#include <filesystem>
#define private public
#include "src/scene/downgrade.h"
namespace YR {
namespace test {
using namespace YR::utility;
using namespace YR::scene;
using namespace YR::Libruntime;
using namespace testing;
class MockClientManager : public ClientManager {
public:
    explicit MockClientManager(const std::shared_ptr<LibruntimeConfig> &librtCfg) : ClientManager(librtCfg) {}
    MOCK_METHOD1(Init, ErrorInfo(const ConnectionParam &param));
    MOCK_METHOD6(SubmitInvokeRequest,
                 void(const http::verb &method, const std::string &target,
                      const std::unordered_map<std::string, std::string> &headers, const std::string &body,
                      const std::shared_ptr<std::string> requestId, const HttpCallbackFunction &receiver));
};

class DowngradeControllerTest : public ::testing::Test {
public:
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
        std::string address = "127.0.0.1:11111";
        std::string urlPrefix = "/serverless/v2/functions/wisefunction:cn";
        clientsMgr = std::make_shared<ClientsManager>();
        auto config = std::make_shared<LibruntimeConfig>();
        clientMgr = std::make_shared<MockClientManager>(config);
        clientsMgr->httpClients[address] = clientMgr;

        security = std::make_shared<Security>();
        std::string accessKey = "access_key";
        SensitiveValue secretKey = std::string("secret_key");
        security->SetAKSKAndCredential(accessKey, secretKey);
        setenv(FRONTEND_ADDRESS_ENV.c_str(), address.c_str(), 1);
        setenv(INVOCATION_URL_PREFIX_ENV.c_str(), urlPrefix.c_str(), 1);
        spec = std::make_shared<InvokeSpec>();
        spec->traceId = "traceId";
        spec->requestId = "requestId";
        spec->functionMeta.functionId = "tenantId/hello/version";
        std::string arg = "{\"key\":\"invoke\"}";
        InvokeArg libArg;
        libArg.dataObj = std::make_shared<YR::Libruntime::DataObject>(0, arg.size());
        libArg.dataObj->data->MemoryCopy(arg.data(), arg.size());
        spec->invokeArgs.emplace_back(std::move(libArg));
    }

    void TearDown() override {}
    std::shared_ptr<MockClientManager> clientMgr;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::shared_ptr<Security> security;
    std::shared_ptr<InvokeSpec> spec;
};

TEST_F(DowngradeControllerTest, ApiClient)
{
    EXPECT_CALL(*this->clientMgr, SubmitInvokeRequest(_, _, _, _, _, _)).WillOnce(testing::Return());
    std::string functionId = "tenantId/hello/version";
    auto apiClient = std::make_shared<ApiClient>(functionId, this->clientsMgr, this->security);
    ASSERT_TRUE(apiClient->Init().OK());
    ASSERT_NO_THROW(apiClient->InvocationAsync(this->spec, nullptr));
}

TEST_F(DowngradeControllerTest, DowngradeControllerTest)
{
    std::string functionId = "tenantId/hello/version";
    EXPECT_CALL(*this->clientMgr, SubmitInvokeRequest(_, _, _, _, _, _)).WillOnce(testing::Return());
    auto controller = std::make_shared<DowngradeController>(functionId, this->clientsMgr, this->security);
    ASSERT_TRUE(controller->Init().OK());
    ASSERT_FALSE(controller->ShouldDowngrade(this->spec));
    controller->ParseDowngrade("/tmp/log/a.json");  // test not exit
    controller->Downgrade(this->spec, nullptr);
    controller->Stop();
}

TEST_F(DowngradeControllerTest, DowngradeControllerInitTest)
{
    setenv(DOWNGRADE_FILE_ENV.c_str(), "/tmp/log/downgrade.json", 1);
    std::string functionId = "tenantId/hello/version";
    auto controller = std::make_shared<DowngradeController>(functionId, this->clientsMgr, this->security);
    ASSERT_TRUE(controller->Init().OK());
    unsetenv(DOWNGRADE_FILE_ENV.c_str());
    controller->Stop();
}
}  // namespace test
}  // namespace YR