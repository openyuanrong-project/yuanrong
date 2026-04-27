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
#include <unistd.h>
#include <boost/beast/http.hpp>
#include <future>
#define private public
#include "httpserver/async_https_server.h"
#include "src/libruntime/gwclient/http/client_manager.h"
#include "src/libruntime/gwclient/http/http_client.h"
namespace YR {
namespace test {
using namespace YR::Libruntime;

class HttpsClientTest : public ::testing::Test {
public:
    HttpsClientTest() {}
    ~HttpsClientTest() {}
    void SetUp() override
    {
        httpsServer_ = std::make_shared<AsyncHttpsServer>();
        YR::utility::LogParam logParam = {
            .logLevel = "DEBUG",
            .logDir = "/tmp/log",
            .nodeName = "test-https",
            .modelName = "test",
            .maxSize = 100,
            .maxFiles = 1,
            .logFileWithTime = false,
            .logBufSecs = 30,
            .maxAsyncQueueSize = 1048510,
            .asyncThreadCount = 1,
            .alsoLog2Stderr = true,
        };
        YR::utility::InitLog(logParam);
    }
    void TearDown() override
    {
        if (httpsServer_) {
            httpsServer_.reset();
        }
    }

private:
    std::shared_ptr<AsyncHttpsServer> httpsServer_;
    std::string ip_ = "127.0.0.1";
    unsigned short port_ = 12346;
    int threadNum = 8;
};

std::shared_ptr<LibruntimeConfig> ConstructLibruntimeConfig()
{
    std::shared_ptr<LibruntimeConfig> librtCfg = std::make_shared<LibruntimeConfig>();
    librtCfg->enableMTLS = true;
    librtCfg->verifyFilePath = "./test/data/cert/ca.crt";
    librtCfg->certificateFilePath = "./test/data/cert/client.crt";
    std::strcpy(librtCfg->privateKeyPaaswd, "test");
    librtCfg->privateKeyPath = "./test/data/cert/client.key";
    // The serverName is not verified.
    librtCfg->serverName = "test";
    return librtCfg;
}

std::shared_ptr<ssl::context> ConstructSslContext()
{
    try {
        auto ctx = std::make_shared<ssl::context>(ssl::context::tlsv12);
        ctx->set_options(boost::asio::ssl::context::default_workarounds | boost::asio::ssl::context::no_sslv2);
        ctx->load_verify_file("./test/data/cert/ca.crt");
        ctx->use_certificate_chain_file("./test/data/cert/server.crt");
        ctx->set_password_callback(
            [](std::size_t max_length, ssl::context_base::password_purpose purpose) { return "test"; });
        ctx->use_private_key_file("./test/data/cert/server.key", ssl::context::pem);
        return ctx;
    } catch (const std::exception &e) {
        std::cerr << e.what() << std::endl;
        return nullptr;
    }
}

TEST_F(HttpsClientTest, InitFailed)
{
    auto librtCfg = ConstructLibruntimeConfig();
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1.0", "12346"});
    ASSERT_EQ(err.OK(), false);
}

TEST_F(HttpsClientTest, SubmitTask)
{
    auto ctx = ConstructSslContext();
    ASSERT_TRUE(ctx != nullptr);
    if (httpsServer_->StartServer(ip_, port_, threadNum, ctx)) {
        std::cout << "start https server success" << std::endl;
    } else {
        std::cout << "start https server failed" << std::endl;
    }
    auto librtCfg = ConstructLibruntimeConfig();
    librtCfg->httpIocThreadsNum = 5;
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1", "12346"});
    ASSERT_EQ(err.OK(), true);

    std::unordered_map<std::string, std::string> headers;
    headers.emplace("type", "test");
    std::string urn = "/test";
    auto retPromise = std::make_shared<std::promise<std::string>>();
    auto future = retPromise->get_future();
    auto requestId = std::make_shared<std::string>("requestID");
    httpClient->SubmitInvokeRequest(
        GET, urn, headers, "", requestId,
        [retPromise](const std::string &result, const boost::beast::error_code &errorCode, const uint statusCode) {
            if (errorCode) {
                std::cerr << "network error, error_code: " << errorCode.message() << std::endl;
            } else {
                retPromise->set_value(result);
            }
        });
    ASSERT_EQ("ok", future.get());
    httpsServer_->StopServer();
}

/*case
 * @title: Server故障恢复后发送请求成功
 * @precondition:
 * @step:  1. 启动HttpsServer
 * @step:  2. 创建客户端连接
 * @step:  3. 停止HttpsServer
 * @step:  4. 恢复HttpsServer
 * @step:  5. 发送https请求
 * @expect:  1.用例不卡住
 */
TEST_F(HttpsClientTest, after_https_server_recover_request_should_return)
{
    auto ctx = ConstructSslContext();
    ASSERT_TRUE(ctx != nullptr);
    if (httpsServer_->StartServer(ip_, port_, threadNum, ctx)) {
        std::cout << "start https server success" << std::endl;
    } else {
        std::cout << "start https server failed" << std::endl;
    }
    auto librtCfg = ConstructLibruntimeConfig();
    librtCfg->httpIocThreadsNum = 5;
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1", "12346"});
    ASSERT_EQ(err.OK(), true);
    httpsServer_->StopServer();
    if (httpsServer_->StartServer(ip_, port_, threadNum, ctx)) {
        std::cout << "start https server success" << std::endl;
    } else {
        std::cout << "start https server failed" << std::endl;
    }
    std::unordered_map<std::string, std::string> headers;
    headers.emplace("type", "test");
    std::string urn = "/test";
    int num = 0;
    auto promise = std::make_shared<std::promise<int>>();
    auto future = promise->get_future();

    auto requestId = std::make_shared<std::string>("requestID");
    std::mutex mtx;
    int sendTimes = 10;
    auto sendMsgHandler = [&]() {
        httpClient->SubmitInvokeRequest(
            GET, urn, headers, "", requestId,
            [&](const std::string &result, const boost::beast::error_code &errorCode, const uint statusCode) {
                if (errorCode) {
                    std::cerr << "network error, error_code: " << errorCode.message() << std::endl;
                } else {
                    std::cout << "request success" << std::endl;
                }
                std::lock_guard<std::mutex> lockGuard(mtx);
                std::cout << "num: " << num << std::endl;
                num++;
                if (num == sendTimes) {
                    promise->set_value(num);
                }
            });
    };
    for (int i = 0; i < sendTimes; i++) {
        sendMsgHandler();
    }
    ASSERT_EQ(sendTimes, future.get());
    httpsServer_->StopServer();
}

TEST_F(HttpsClientTest, after_reinit_should_not_coredump)
{
    auto ctx = ConstructSslContext();
    ASSERT_TRUE(ctx != nullptr);
    if (httpsServer_->StartServer(ip_, port_, threadNum, ctx)) {
        std::cout << "start https server success" << std::endl;
    } else {
        std::cout << "start https server failed" << std::endl;
    }
    auto librtCfg = ConstructLibruntimeConfig();
    librtCfg->httpIocThreadsNum = 5;
    // Change the value of idleTime to a smaller value to trigger reinit.
    librtCfg->httpIdleTime = 1;
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1", "12346"});
    std::unordered_map<std::string, std::string> headers;
    headers.emplace("type", "test");
    std::string urn = "/test";
    for (int i = 0; i < 3; i++) {
        int num = 0;
        auto promise = std::make_shared<std::promise<int>>();
        auto future = promise->get_future();
        auto requestId = std::make_shared<std::string>("requestID");
        std::mutex mtx;
        int sendTimes = 10;
        auto sendMsgHandler = [&]() {
            httpClient->SubmitInvokeRequest(
                GET, urn, headers, "", requestId,
                [&](const std::string &result, const boost::beast::error_code &errorCode, const uint statusCode) {
                    std::lock_guard<std::mutex> lockGuard(mtx);
                    std::cout << "num: " << num << std::endl;
                    num++;
                    if (num == sendTimes) {
                        promise->set_value(num);
                    }
                });
        };
        for (int j = 0; j < sendTimes; j++) {
            sendMsgHandler();
        }
        ASSERT_EQ(sendTimes, future.get());
        sleep(1);
    }
    httpsServer_->StopServer();
}

TEST_F(HttpsClientTest, trigger_init_more_client_should_not_coredump)
{
    auto ctx = ConstructSslContext();
    ASSERT_TRUE(ctx != nullptr);
    if (httpsServer_->StartServer(ip_, port_, threadNum, ctx)) {
        std::cout << "start https server success" << std::endl;
    } else {
        std::cout << "start https server failed" << std::endl;
    }
    auto librtCfg = ConstructLibruntimeConfig();
    librtCfg->httpIocThreadsNum = 20;
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1", "12346"});
    std::unordered_map<std::string, std::string> headers;
    headers.emplace("type", "test");
    std::string urn = "/test";
    auto promise = std::make_shared<std::promise<int>>();
    auto future = promise->get_future();
    auto sharedFuture = future.share();
    auto requestId = std::make_shared<std::string>("requestID");
    // default http init is 10, sendTimes = 15 trigger init more client
    int sendTimes = 15;
    std::atomic<int> get{0};
    auto sendMsgHandler = [&]() {
        httpClient->SubmitInvokeRequest(
            GET, urn, headers, "", requestId,
            [&](const std::string &result, const boost::beast::error_code &errorCode, const uint statusCode) {
                sharedFuture.get();
                get++;
            });
    };
    for (int j = 0; j < sendTimes; j++) {
        sendMsgHandler();
    }
    promise->set_value(1);
    for (;;) {
        if (get == sendTimes) {
            ASSERT_EQ(get, sendTimes);
            break;
        }
        std::this_thread::yield();
    }
    EXPECT_NO_THROW(httpsServer_->StopServer());
}

// Manual test case requires using tc to construct network latency.
TEST_F(HttpsClientTest, DISABLED_Reinit_Retry_When_Network_latency)
{
    auto ctx = ConstructSslContext();
    ASSERT_TRUE(ctx != nullptr);
    if (httpsServer_->StartServer(ip_, port_, threadNum, ctx)) {
        std::cout << "start https server success" << std::endl;
    } else {
        std::cout << "start https server failed" << std::endl;
    }
    auto librtCfg = ConstructLibruntimeConfig();
    librtCfg->httpIocThreadsNum = 5;
    // Change the value of idleTime to a smaller value to trigger reinit.
    librtCfg->httpIdleTime = 1;
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1", "12346", 1});
    std::unordered_map<std::string, std::string> headers;
    headers.emplace("type", "test");
    std::string urn = "/test";

    std::promise<bool> waitPromise;
    auto waitFuture = waitPromise.get_future();
    waitFuture.wait_for(std::chrono::seconds(10));
    // tc qdisc add dev eth0 root netem delay 3s

    auto requestId = std::make_shared<std::string>("requestID");
    auto promise = std::make_shared<std::promise<int>>();
    auto future = promise->get_future();
    httpClient->SubmitInvokeRequest(
        GET, urn, headers, "", requestId,
        [&](const std::string &result, const boost::beast::error_code &errorCode, const uint statusCode) {
            if (errorCode) {
                YRLOG_DEBUG("error {}", errorCode.to_string());
                promise->set_value(0);
            } else {
                YRLOG_DEBUG("response ok");
                promise->set_value(1);
            }
        });
    ASSERT_EQ(1, future.get());
    httpsServer_->StopServer();
}

TEST_F(HttpsClientTest, DISABLED_TLSVerify)
{
    auto ctx = ConstructSslContext();
    ASSERT_TRUE(ctx != nullptr);
    if (httpsServer_->StartServer(ip_, port_, threadNum, ctx)) {
        std::cout << "start https server success" << std::endl;
    } else {
        std::cout << "start https server failed" << std::endl;
    }
    auto librtCfg = std::make_shared<LibruntimeConfig>();
    librtCfg->httpIocThreadsNum = 5;
    librtCfg->enableTLS = true;
    librtCfg->serverName = "test";
    librtCfg->verifyFilePath = "./test/data/cert/ca.crt";
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1", "12346"});
    ASSERT_EQ(err.OK(), true);

    std::unordered_map<std::string, std::string> headers;
    headers.emplace("type", "test");
    std::string urn = "/test";
    auto retPromise = std::make_shared<std::promise<std::string>>();
    auto future = retPromise->get_future();
    auto requestId = std::make_shared<std::string>("requestID");
    httpClient->SubmitInvokeRequest(
        GET, urn, headers, "", requestId,
        [retPromise](const std::string &result, const boost::beast::error_code &errorCode, const uint statusCode) {
            if (errorCode) {
                std::cerr << "network error, error_code: " << errorCode.message() << std::endl;
            } else {
                retPromise->set_value(result);
            }
        });
    ASSERT_EQ("ok", future.get());
    httpsServer_->StopServer();
}
}  // namespace test
}  // namespace YR
