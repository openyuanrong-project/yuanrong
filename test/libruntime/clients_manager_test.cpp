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
#include <pthread.h>
#include <boost/asio/ssl.hpp>
#include <boost/beast/http.hpp>
#include "mock/mock_datasystem.h"
#define private public
#include "httpserver/async_http_server.h"
#include "src/libruntime/clientsmanager/clients_manager.h"
#include "src/libruntime/utils/utils.h"
#include "src/libruntime/err_type.h"
#include "src/utility/logger/logger.h"
using namespace testing;
using namespace YR::utility;
using namespace YR::Libruntime;
namespace YR {
namespace test {
class ClientsManagerTest : public testing::Test {
public:
    ClientsManagerTest(){};
    ~ClientsManagerTest(){};
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
        httpServer_ = std::make_shared<AsyncHttpServer>();
        httpServer_->StartServer("127.0.0.1", port, 5);
    }
    void TearDown() override
    {
        httpServer_->StopServer();
    }

private:
    int port = 22222;
    std::shared_ptr<AsyncHttpServer> httpServer_;
};

TEST_F(ClientsManagerTest, DsClientsTest)
{
    auto clientsMgr = std::make_shared<ClientsManager>();
    datasystem::SensitiveValue runtimePrivateKey;
    datasystem::SensitiveValue token;
    auto librtCfg = std::make_shared<LibruntimeConfig>();
    librtCfg->dataSystemIpAddr = "127.0.0.1";
    librtCfg->dataSystemPort = 22222;
    librtCfg->runtimePrivateKey = runtimePrivateKey;

    auto res = clientsMgr->GetOrNewDsClient(librtCfg, "", datasystem::SensitiveValue{}, 30);
    EXPECT_EQ(res.second.Code(), ErrorCode::ERR_OK);
    EXPECT_EQ(clientsMgr->dsClientsReferCounter["127.0.0.1:22222"], 1);
    res = clientsMgr->GetOrNewDsClient(librtCfg, "", datasystem::SensitiveValue{}, 30);
    EXPECT_EQ(res.second.Code(), ErrorCode::ERR_OK);
    EXPECT_EQ(clientsMgr->dsClientsReferCounter["127.0.0.1:22222"], 2);
    auto err = clientsMgr->ReleaseDsClient("127.0.0.1", 22222);
    EXPECT_EQ(err.Code(), ErrorCode::ERR_OK);
    EXPECT_EQ(clientsMgr->dsClientsReferCounter["127.0.0.1:22222"], 1);
    err = clientsMgr->ReleaseDsClient("127.0.0.1", 22222);
    EXPECT_EQ(err.Code(), ErrorCode::ERR_OK);
    EXPECT_EQ(clientsMgr->dsClientsReferCounter["127.0.0.1:22222"], 0);
    EXPECT_EQ(clientsMgr->dsClients.find("127.0.0.1:22222"), clientsMgr->dsClients.end());
}

TEST_F(ClientsManagerTest, HttpClientsTest)
{
    auto clientsMgr = std::make_shared<ClientsManager>();
    std::shared_ptr<LibruntimeConfig> librtCfg = std::make_shared<LibruntimeConfig>();
    auto res = clientsMgr->GetOrNewHttpClient("127.0.0.1", 22222, librtCfg);
    EXPECT_EQ(res.second.Code(), ErrorCode::ERR_OK);
    EXPECT_EQ(clientsMgr->httpClientsReferCounter["127.0.0.1:22222"], 1);
    res = clientsMgr->GetOrNewHttpClient("127.0.0.1", 22222, librtCfg);
    EXPECT_EQ(res.second.Code(), ErrorCode::ERR_OK);
    EXPECT_EQ(clientsMgr->httpClientsReferCounter["127.0.0.1:22222"], 2);
    auto err = clientsMgr->ReleaseHttpClient("127.0.0.1", 22222);
    EXPECT_EQ(err.Code(), ErrorCode::ERR_OK);
    EXPECT_EQ(clientsMgr->httpClientsReferCounter["127.0.0.1:22222"], 1);
    err = clientsMgr->ReleaseHttpClient("127.0.0.1", 22222);
    EXPECT_EQ(err.Code(), ErrorCode::ERR_OK);
    EXPECT_EQ(clientsMgr->httpClientsReferCounter["127.0.0.1:22222"], 0);
    EXPECT_EQ(clientsMgr->httpClients.find("127.0.0.1:22222"), clientsMgr->httpClients.end());
}

TEST_F(ClientsManagerTest, GetFsConnTest)
{
    auto clientsMgr = std::make_shared<ClientsManager>();
    auto res = clientsMgr->GetFsConn("127.0.0.1", 8080, "");
    ASSERT_TRUE(res.first == nullptr);
    ASSERT_TRUE(res.second.OK());
}

TEST_F(ClientsManagerTest, ReleaseDsClientTest)
{
    auto clientsMgr = std::make_shared<ClientsManager>();
    clientsMgr->dsClientsReferCounter["127.0.0.1:80"] = 1;
    DatasystemClients dsClients{.dsObjectStore = std::make_shared<MockObjectStore>(),
                                .dsStateStore = std::make_shared<MockStateStore>(),
                                .dsStreamStore = std::make_shared<MockStreamStore>(),
                                .dsHeteroStore = std::make_shared<MockHeretoStore>()};
    clientsMgr->dsClients["127.0.0.1:80"] = dsClients;
    ASSERT_EQ(clientsMgr->ReleaseDsClient("127.0.0.1", 80).OK(), true);
    ASSERT_EQ(clientsMgr->dsClientsReferCounter["127.0.0.1:80"] == 0, true);
}

TEST_F(ClientsManagerTest, FsConnErrorHandlingTest)
{
    auto clientsMgr = std::make_shared<ClientsManager>();
    
    auto result = clientsMgr->GetFsConn("invalid_ip", -1, "test_instance");
    EXPECT_EQ(result.first, nullptr);
    EXPECT_NE(result.second.Code(), ErrorCode::ERR_OK);
}

TEST_F(ClientsManagerTest, GetIpAddrTest)
{
    std::string ip = "127.0.0.1";
    int port = 8080;
    std::string expected = "127.0.0.1:8080";
    auto result = GetIpAddr(ip, port);
    EXPECT_EQ(result, expected);
}
}  // namespace test
}  // namespace YR