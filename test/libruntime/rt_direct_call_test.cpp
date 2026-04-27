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

#include "mock/fake_function_proxy_server.h"
#include "mock/mock_security.h"
#include "src/libruntime/fsclient/fs_client.h"
#include "src/libruntime/fsclient/fs_intf_impl.h"

using namespace testing;
using namespace YR::utility;
using namespace YR::Libruntime;
using namespace std::placeholders;

namespace YR {
namespace test {
class FakeCallee {
public:
    FakeCallee()
    {
        handlers_.call = std::bind(&FakeCallee::EmptyCallHandler, this, _1);
        handlers_.init = [](const std::shared_ptr<CallMessageSpec> &){ return CallResponse(); };
        handlers_.checkpoint = [](const CheckpointRequest &){ return CheckpointResponse (); };
        handlers_.recover = [](const RecoverRequest &){ return RecoverResponse (); };
        handlers_.shutdown = [](const ShutdownRequest &){ return ShutdownResponse (); };
        handlers_.signal = [](const SignalRequest &){ return SignalResponse (); };
        clientsMgr = std::make_shared<ClientsManager>();
        worker.Init(1, "FakeCallee");
    }

    void Start(const std::shared_ptr<FakeFunctionProxyServer> &functionProxy)
    {
        Config::Instance().POD_IP() = Config::Instance().HOST_IP();
        fsClient_ = std::make_shared<FSIntfImpl>(Config::Instance().HOST_IP(), functionProxy->GetPort(), handlers_,
                                                 false, security_, clientsMgr, true);
        auto err = fsClient_->Start("12345678", "callee", "callee");
        EXPECT_TRUE(err.OK());
        EXPECT_EQ(err.Msg(), "");
        t = std::thread([&](){
            fsClient_->ReceiveRequestLoop();
        });
        fsClient_->SetInitialized();
    }

    int GetPort()
    {
        return fsClient_->selfPort;
    }

    void Stop()
    {
        if (fsClient_ != nullptr) {
            fsClient_->Stop();
        }
        if (t.joinable()) {
            t.join();
        }
        fsClient_ = nullptr;
    }

    CallResponse EmptyCallHandler(const std::shared_ptr<CallMessageSpec> &req)
    {
        CallResponse resp;
        YRLOG_INFO("EmptyCallHandler {} {}", req->Immutable().requestid(), req->Immutable().senderid());
        this->worker.Handle([this, req] () {
            YRLOG_INFO("ReturnCallResult {} {}", req->Immutable().requestid(), req->Immutable().senderid());
            CallResult res;
            res.set_requestid(req->Immutable().requestid());
            res.set_instanceid(req->Immutable().senderid());
                auto result = std::make_shared<CallResultMessageSpec>();
                result->Mutable() = std::move(res);
                fsClient_->ReturnCallResult(result, false, [this](const CallResultAck &resp) {
                    if (resp.code() != common::ERR_NONE) {
                        YRLOG_WARN("failed to send CallResult, code: {}, message: {}", fmt::underlying(resp.code()),
                                   resp.message());
                    }
                    return;
                });
            }, "");
        return resp;
    }

    ThreadPool worker;
    FSIntfHandlers handlers_;
    std::shared_ptr<FSIntfImpl> fsClient_;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::thread t;
    std::shared_ptr<YR::Libruntime::Security> security_ = std::make_shared<YR::test::MockSecurity>();;
    int port;
};

class RTDirectCallTest : public testing::Test {
public:
    RTDirectCallTest()
    {
        handlers_.call = [](const std::shared_ptr<CallMessageSpec> &){ return CallResponse(); };
        handlers_.init = [](const std::shared_ptr<CallMessageSpec> &){ return CallResponse(); };
        handlers_.checkpoint = [](const CheckpointRequest &){ return CheckpointResponse (); };
        handlers_.recover = [](const RecoverRequest &){ return RecoverResponse (); };
        handlers_.shutdown = [](const ShutdownRequest &){ return ShutdownResponse (); };
        handlers_.signal = [](const SignalRequest &){ return SignalResponse (); };
    }

    ~RTDirectCallTest() {}

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
        Config::Instance().RUNTIME_DIRECT_CONNECTION_ENABLE() = true;
        clientsMgr = std::make_shared<ClientsManager>();
        functionProxy = std::make_shared<FakeFunctionProxyServer>(Config::Instance().HOST_IP());
        fakeCallee = std::make_shared<FakeCallee>();
        functionProxy->Start();
        fakeCallee->Start(functionProxy);
        calleePort = fakeCallee->GetPort();
    }

    void TearDown() override
    {
        if (fakeCallee) {
            fakeCallee->Stop();
        }
        if (caller) {
            caller->Stop();
            caller = nullptr;
        }
        functionProxy->Stop();
        Config::Instance().RUNTIME_DIRECT_CONNECTION_ENABLE() = false;
    }

    void DoStartCaller()
    {
        Config::Instance().POD_IP() = Config::Instance().HOST_IP();
        Config::Instance().DERICT_RUNTIME_SERVER_PORT() = callerPort;
        caller = std::make_shared<FSClient>();
        auto err = caller->Start(Config::Instance().HOST_IP(), functionProxy->GetPort(), handlers_,
                                 FSClient::ClientType::GRPC_CLIENT, false, security_, clientsMgr, "12345678", "caller",
                                 "caller", "function");
        EXPECT_TRUE(err.OK());
        EXPECT_EQ(err.Msg(), "");
    }

    FSIntfHandlers handlers_;
    std::shared_ptr<FSClient> caller;
    std::shared_ptr<FakeCallee> fakeCallee;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::shared_ptr<FakeFunctionProxyServer> functionProxy;
    std::shared_ptr<Libruntime::Security> security_ = std::make_shared<YR::test::MockSecurity>();
    int callerPort = 5551;
    int calleePort = 6661;

    void DirectInvoke()
    {
        DoStartCaller();
        NotificationUtility notified;
        auto notifyHandler = [&notified](const NotifyRequest &req, const ErrorInfo &err) -> void { notified.Notify(); };

        auto reqId = YR::utility::IDGenerator::GenRequestId();
        InvokeRequest req;
        req.set_requestid(reqId);
        req.set_instanceid("callee");
        auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
        caller->InvokeAsync(messageSpec, notifyHandler, -1);
        InvokeResponse resp;
        functionProxy->SendAfterRead(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), resp));

        NotifyRequest notifyReq;
        notifyReq.set_requestid(reqId);
        notifyReq.mutable_runtimeinfo()->set_serveripaddr(Config::Instance().HOST_IP());
        notifyReq.mutable_runtimeinfo()->set_serverport(calleePort);
        functionProxy->Send(*GenStreamMsg("", notifyReq));
        {
            auto err = notified.WaitForNotification();
            EXPECT_TRUE(err.OK());
        }

        // directly call
        {
            NotificationUtility notified;
            auto notifyHandler = [&notified](const NotifyRequest &req, const ErrorInfo &err) -> void { notified.Notify(); };
            auto reqId = YR::utility::IDGenerator::GenRequestId();
            InvokeRequest req;
            req.set_requestid(reqId);
            req.set_instanceid("callee");
            auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
            caller->InvokeAsync(messageSpec, notifyHandler, -1);
            {
                auto err = notified.WaitForNotification();
                EXPECT_TRUE(err.OK());
            }
            auto [channel, err] = clientsMgr->GetFsConn(Config::Instance().HOST_IP(), calleePort, "callee");
            EXPECT_TRUE(channel != nullptr);
            EXPECT_TRUE(err.OK());
        }
    }
};

// call to function proxy, return to build rtIntf
TEST_F(RTDirectCallTest, CallResultToBuildRtInrf)
{
    DirectInvoke();
}

// After the direct connection is disconnected, the request is sent through the function proxy.
TEST_F(RTDirectCallTest, CallRtDowngradeTest)
{
    DirectInvoke();
    // directly call
    // close fake callee
    fakeCallee->Stop();
    NotificationUtility notified;
    auto notifyHandler = [&notified](const NotifyRequest &req, const ErrorInfo &err) -> void { notified.Notify(); };
    auto reqId = YR::utility::IDGenerator::GenRequestId();
    InvokeRequest req;
    req.set_requestid(reqId);
    req.set_instanceid("callee");
    auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
    caller->InvokeAsync(messageSpec, notifyHandler, -1);
    InvokeResponse resp;
    functionProxy->SendAfterRead(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), resp));

    NotifyRequest notifyReq;
    notifyReq.set_requestid(reqId);
    functionProxy->Send(*GenStreamMsg("", notifyReq));
    {
        auto err = notified.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

// function-proxy disconnected, rt connection still alive
TEST_F(RTDirectCallTest, FunctionProxyDisconnectedTest)
{
    DirectInvoke();
    // directly call
    // close function proxy
    functionProxy->Stop();
    for (int i = 0; i <= 3; i++) {
        NotificationUtility notified;
        auto notifyHandler = [&notified](const NotifyRequest &req, const ErrorInfo &err) -> void { notified.Notify(); };
        auto reqId = YR::utility::IDGenerator::GenRequestId();
        InvokeRequest req;
        req.set_requestid(reqId);
        req.set_instanceid("callee");
        auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
        caller->InvokeAsync(messageSpec, notifyHandler, -1);
        {
            auto err = notified.WaitForNotification();
            EXPECT_TRUE(err.OK());
        }
        auto [channel, err] = clientsMgr->GetFsConn(Config::Instance().HOST_IP(), calleePort, "callee");
        EXPECT_TRUE(channel != nullptr);
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(RTDirectCallTest, TestWhenMessageTooLarge)
{
    DirectInvoke();
    uint32_t SIZE_IN_BYTES = 11 * YR::Libruntime::SIZE_MEGA_BYTES;
    char *buffer = static_cast<char *>(malloc(SIZE_IN_BYTES));
    for (uint i = 0; i < SIZE_IN_BYTES; ++i) {
        buffer[i] = '\0';
    }
    NotificationUtility notified;
    std::promise<NotifyRequest> pro;
    auto notifyHandler = [&pro, &notified](const NotifyRequest &req, const ErrorInfo &err) -> void {
        notified.Notify();
        pro.set_value(req);
    };
    auto reqId = YR::utility::IDGenerator::GenRequestId();
    InvokeRequest req;
    req.set_requestid(reqId);
    req.set_instanceid("callee");
    std::string *value = req.add_args()->mutable_value();
    value->reserve(SIZE_IN_BYTES);
    value->append(buffer, SIZE_IN_BYTES);
    auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
    caller->InvokeAsync(messageSpec, notifyHandler, -1);
    {
        auto err = notified.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
    auto ret = pro.get_future().get();
    EXPECT_EQ(ret.code(), ERR_PARAM_INVALID);
    free(buffer);
}

}  // namespace test
}  // namespace YR