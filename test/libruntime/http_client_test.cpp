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
#include <pthread.h>
#include <unistd.h>
#include <boost/beast/http.hpp>
#include <future>
#define private public
#include "httpserver/async_http_server.h"
#include "src/libruntime/gwclient/http/client_manager.h"
#include "src/libruntime/gwclient/http/http_client.h"
namespace YR {
namespace test {
using namespace YR::Libruntime;

class HttpClientTest : public ::testing::Test {
public:
    HttpClientTest() {}
    ~HttpClientTest() {}

    void SetUp() override
    {
        httpServer_ = std::make_shared<AsyncHttpServer>();
        YR::utility::LogParam logParam = {
            .logLevel = "DEBUG",
            .logDir = "/tmp/log",
            .nodeName = "test-http",
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
        if (httpServer_) {
            httpServer_.reset();
        }
    }

private:
    std::shared_ptr<AsyncHttpServer> httpServer_;
    std::string ip_ = "127.0.0.1";
    unsigned short port_ = 12345;
    int threadNum = 8;
};

TEST_F(HttpClientTest, InitFailed)
{
    std::shared_ptr<LibruntimeConfig> librtCfg = std::make_shared<LibruntimeConfig>();
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1.0", "12345"});
    ASSERT_EQ(err.OK(), false);
}

TEST_F(HttpClientTest, SubmitTask)
{
    if (httpServer_->StartServer(ip_, port_, threadNum)) {
        std::cout << "start http server success" << std::endl;
    } else {
        std::cout << "start http server failed" << std::endl;
    }
    std::shared_ptr<LibruntimeConfig> librtCfg = std::make_shared<LibruntimeConfig>();
    librtCfg->httpIocThreadsNum = 5;
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1", "12345"});
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
    httpServer_->StopServer();
}

/*case
 * @title: Server故障恢复后发送请求成功
 * @precondition:
 * @step:  1. 启动HttpServer
 * @step:  2. 创建客户端连接
 * @step:  3. 停止HttpServer
 * @step:  4. 恢复HttpServer
 * @step:  5. 发送http请求
 * @expect:  1.用例不卡住
 */
TEST_F(HttpClientTest, DISABLED_after_httpserver_recover_request_should_return)
{
    if (httpServer_->StartServer(ip_, port_, threadNum)) {
        std::cout << "start http server success" << std::endl;
    } else {
        std::cout << "start http server failed" << std::endl;
    }
    std::shared_ptr<LibruntimeConfig> librtCfg = std::make_shared<LibruntimeConfig>();
    librtCfg->httpIocThreadsNum = 5;
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1", "12345"});
    ASSERT_EQ(err.OK(), true);
    httpServer_->StopServer();
    if (httpServer_->StartServer(ip_, port_, threadNum)) {
        std::cout << "start http server success" << std::endl;
    } else {
        std::cout << "start http server failed" << std::endl;
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
    httpServer_->StopServer();
}

/*case
 * @title: Server故障后发送请求失败
 * @precondition:
 * @step:  1. 启动HttpServer
 * @step:  2. 创建客户端连接
 * @step:  3. 停止HttpServer
 * @step:  5. 发送http请求
 * @expect:  1.callback执行一次
 */
TEST_F(HttpClientTest, after_httpserver_stop_request_should_return_once)
{
    if (httpServer_->StartServer(ip_, port_, threadNum)) {
        std::cout << "start http server success" << std::endl;
    } else {
        std::cout << "start http server failed" << std::endl;
    }
    std::shared_ptr<LibruntimeConfig> librtCfg = std::make_shared<LibruntimeConfig>();
    auto httpClient = std::make_unique<ClientManager>(librtCfg);
    auto err = httpClient->Init({"127.0.0.1", "12345"});
    ASSERT_EQ(err.OK(), true);
    httpServer_->StopServer();
    std::unordered_map<std::string, std::string> headers;
    headers.emplace("type", "test");
    std::string urn = "/test";
    int num = 0;
    auto promise = std::make_shared<std::promise<int>>();
    auto future = promise->get_future();
    auto requestId = std::make_shared<std::string>("requestID");
    auto sendMsgHandler = [&]() {
        httpClient->SubmitInvokeRequest(
            GET, urn, headers, "", requestId,
            [&](const std::string &result, const boost::beast::error_code &errorCode, const uint statusCode) {
                if (errorCode) {
                    std::cerr << "network error, error_code: " << errorCode.message() << std::endl;
                } else {
                    std::cout << "request success" << std::endl;
                }
                num++;
                promise->set_value(num);
            });
    };
    sendMsgHandler();
    ASSERT_EQ(1, future.get());
}
}  // namespace test
}  // namespace YR
