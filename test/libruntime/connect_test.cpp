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

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include <string>

#include "mock/mock_datasystem.h"
#include "mock/mock_fs_intf.h"
#include "src/libruntime/libruntime.h"
#include "src/utility/id_generator.h"

using namespace testing;
using namespace YR::utility;
using namespace YR::Libruntime;
using SocketMessage = ::libruntime::SocketMessage;
using MessageType = ::libruntime::MessageType;
using FunctionLog = ::libruntime::FunctionLog;

namespace YR {
namespace test {
const std::string SOCK_PATH = "/tmp/runtime.sock";

class FakeDomainSocketServer {
public:
    FakeDomainSocketServer(std::string socketPath) : socketPath(socketPath) {}

    virtual ~FakeDomainSocketServer()
    {
        Stop();
    }

    void InitOnce()
    {
        messageCoder = std::make_shared<MessageCoder>();
        std::call_once(this->initFlag, [this]() { this->DoInitOnce(); });
    }

    void DoInitOnce()
    {
        sockFd = socket(AF_UNIX, SOCK_STREAM, 0);
        EXPECT_NE(sockFd, -1);
        struct sockaddr_un serverAddr;
        memset_s(&serverAddr, sizeof(struct sockaddr_un), 0, sizeof(struct sockaddr_un));
        serverAddr.sun_family = AF_UNIX;
        unlink(socketPath.c_str());
        strncpy(serverAddr.sun_path, socketPath.c_str(), sizeof(serverAddr.sun_path) - 1);
        int bindRes = bind(sockFd, (struct sockaddr *)&serverAddr, sizeof(struct sockaddr_un));
        EXPECT_NE(bindRes, -1);
        int listenRes = listen(sockFd, 5) == -1;
        EXPECT_NE(listenRes, -1);
        epollFd = epoll_create1(0);
        EXPECT_GE(bindRes, 0);
        running = true;
        socketThread = std::thread(&FakeDomainSocketServer::Run, this);
        pthread_setname_np(socketThread.native_handle(), "handle_run");
        receivedFunctionLog = std::promise<FunctionLog>();
        receivedFunctionLogFuture = receivedFunctionLog.get_future();
    }

    void Stop()
    {
        running = false;
        if (sockFd >= 0) {
            close(sockFd);
            CleanupSocket();
        }
        if (epollFd >= 0) {
            close(epollFd);
        }
        if (socketThread.joinable()) {
            socketThread.join();
        }
    }

    void CleanupSocket()
    {
        if (access(socketPath.c_str(), F_OK) != -1) {
            YRLOG_INFO("Clean up socket in {}", socketPath);
            unlink(socketPath.c_str());
        }
    }

    void Run()
    {
        struct epoll_event event;
        event.events = EPOLLIN;
        event.data.fd = sockFd;
        if (epoll_ctl(epollFd, EPOLL_CTL_ADD, sockFd, &event) == -1) {
            close(sockFd);
            return;
        }

        while (running) {
            struct epoll_event events[20];
            int numFds = epoll_wait(epollFd, events, 20, 1);
            if (numFds < 0) {
                continue;
            }
            YRLOG_DEBUG("numFds: {}", numFds);
            for (int i = 0; i < numFds; i++) {
                if (events[i].data.fd == sockFd) {
                    int clientFd = accept(sockFd, NULL, NULL);
                    if (clientFd == -1) {
                        continue;
                    }
                    int flags = fcntl(clientFd, F_GETFL, 0);
                    fcntl(clientFd, F_SETFL, flags | O_NONBLOCK);
                    event.events = EPOLLIN | EPOLLET;
                    event.data.fd = clientFd;
                    if (epoll_ctl(epollFd, EPOLL_CTL_ADD, clientFd, &event) == -1) {
                        YRLOG_ERROR("Failed to add clientFd to epoll, code:{}", errno);
                        close(clientFd);
                    }
                } else if ((events[i].events & EPOLLERR) || (events[i].events & EPOLLHUP) ||
                           (!(events[i].events & EPOLLIN))) {
                    RemoveSocketFromEpoll(events[i].data.fd);
                } else {
                    HandleRead(events[i].data.fd);
                    YRLOG_DEBUG("After handle read");
                }
            }
        }
    }

    void RemoveSocketFromEpoll(int fd)
    {
        struct epoll_event event;
        event.data.fd = fd;
        event.events = 0;
        if (epoll_ctl(epollFd, EPOLL_CTL_DEL, fd, &event) == -1) {
            YRLOG_ERROR("Failed to remove socket from epoll, code:{}", errno);
        }
    }

    void HandleRead(int clientFd)
    {
        char buffer[4];
        int byteCount = 0;
        while (1) {
            byteCount = recv(clientFd, buffer, 4, MSG_PEEK);
            YRLOG_DEBUG("Recv data byte count: {}", byteCount);
            if (byteCount == -1) {
                break;
            } else if (byteCount == 0) {
                break;
            }
            auto socketMsg = messageCoder->Decode(clientFd, messageCoder->DecodeMsgSize(buffer));
            HandleReceivedSocketMsg(socketMsg);
        }
    }

    void HandleReceivedSocketMsg(std::shared_ptr<SocketMessage> socketMsg)
    {
        auto businessMsg = socketMsg->businessmsg();
        if (businessMsg.type() == MessageType::LogProcess) {
            try {
                receivedFunctionLog.set_value(businessMsg.functionlog());
            } catch (const std::future_error &e) {
                YRLOG_DEBUG("set function log failed.");
            }
        } else {
            YRLOG_WARN("Unknown socket message type");
        }
    }

    FunctionLog GetFunctionLog(int timeoutSec)
    {
        FunctionLog functionLog;
        if (receivedFunctionLogFuture.wait_for(std::chrono::seconds(timeoutSec)) != std::future_status::ready) {
            YRLOG_DEBUG("get function log failed.");
            return functionLog;
        }
        return receivedFunctionLogFuture.get();
    }

private:
    std::string socketPath;
    int sockFd{-1};
    int epollFd{-1};
    std::thread socketThread;
    std::atomic<bool> running;
    std::once_flag initFlag;
    std::shared_ptr<MessageCoder> messageCoder;
    std::promise<FunctionLog> receivedFunctionLog;
    std::shared_future<FunctionLog> receivedFunctionLogFuture;
};

class ConnectTest : public testing::Test {
public:
    ConnectTest() {}

    ~ConnectTest() {}

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
        system("echo '' > /tmp/runtime.sock && chmod +770 /tmp/runtime.sock");
        server = std::make_shared<FakeDomainSocketServer>(SOCK_PATH);
        server->InitOnce();

        lc = std::make_shared<LibruntimeConfig>();
        lc->jobId = YR::utility::IDGenerator::GenApplicationId();
        lc->tenantId = "tenantId";
        auto clientsMgr = std::make_shared<ClientsManager>();
        auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
        auto sec = std::make_shared<Security>();
        auto socketClient = std::make_shared<DomainSocketClient>(SOCK_PATH);
        lr = std::make_shared<YR::Libruntime::Libruntime>(lc, clientsMgr, metricsAdaptor, sec, socketClient);
    }

    void TearDown() override
    {
        if (lr) {
            lr.reset();
        }
        if (server) {
            server.reset();
        }
        system("rm -rf /tmp/runtime.sock");
    }

    std::shared_ptr<FakeDomainSocketServer> server;
    std::shared_ptr<YR::Libruntime::Libruntime> lr;
    std::shared_ptr<LibruntimeConfig> lc;
};

TEST_F(ConnectTest, TestProcessLogSuccessfully)
{
    std::string level = "info";
    std::string content = "This is a test log content!";
    std::string logType = "tail";
    std::string functionInfo = "functionInfo";
    FunctionLog funcLog;
    funcLog.set_level(level);
    funcLog.set_content(content);
    funcLog.set_logtype(logType);
    funcLog.set_functioninfo(functionInfo);
    lr->ProcessLog(funcLog);
    FunctionLog receivedFunctionLog = server->GetFunctionLog(10);
    EXPECT_EQ(receivedFunctionLog.level(), level);
    EXPECT_EQ(receivedFunctionLog.content(), content);
    EXPECT_EQ(receivedFunctionLog.logtype(), logType);
    EXPECT_EQ(receivedFunctionLog.functioninfo(), functionInfo);
}
}  // namespace test
}  // namespace YR