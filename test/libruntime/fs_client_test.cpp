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

#include "mock/mock_fs_client_and_server.h"
#include "mock/mock_security.h"
#include "src/libruntime/fsclient/fs_client.h"
#include "src/libruntime/fsclient/fs_intf.h"
#include "src/libruntime/fsclient/fs_intf_impl.h"
#include "src/utility/notification_utility.h"

using namespace testing;
using namespace YR::utility;
using namespace YR::Libruntime;
using namespace std::placeholders;

namespace YR {
namespace test {
class FSClientGrpcTest : public testing::Test {
public:
    FSClientGrpcTest()
    {
        handlers_.init = std::bind(&FSClientGrpcTest::EmptyInitHandler, this, _1);
        handlers_.call = std::bind(&FSClientGrpcTest::EmptyCallHandler, this, _1);
        handlers_.checkpoint = std::bind(&FSClientGrpcTest::EmptyCheckpointHandler, this, _1);
        handlers_.recover = std::bind(&FSClientGrpcTest::EmptyRecoverHandler, this, _1);
        handlers_.shutdown = std::bind(&FSClientGrpcTest::EmptyShutdownHandler, this, _1);
        handlers_.signal = std::bind(&FSClientGrpcTest::EmptySignalHandler, this, _1);
        clientsMgr = std::make_shared<ClientsManager>();
        grpcServer = std::make_shared<FakeGrpcServer>(Config::Instance().HOST_IP());
        security_ = std::make_shared<MockSecurity>();
    }

    ~FSClientGrpcTest() {}

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

        grpcServer->Start();
    }

    void TearDown() override
    {
        fsClient_->Stop();
        grpcServer->Stop();
        security_.reset();
        if (t.joinable()) {
            t.join();
        }
    }

    void DoStartGrpcClient()
    {
        fsClient_ = std::make_shared<FSClient>();
        auto err = fsClient_->Start(Config::Instance().HOST_IP(), grpcServer->GetPort(), handlers_,
                                    FSClient::ClientType::GRPC_CLIENT, false, security_, clientsMgr, "12345678",
                                    "instanceID", "runtimeID", "function");
        EXPECT_TRUE(err.OK());
        EXPECT_EQ(err.Msg(), "");
        t = std::thread(&FSClient::ReceiveRequestLoop, *fsClient_);
    }

    CallResponse EmptyInitHandler(const std::shared_ptr<CallMessageSpec> &req)
    {
        CallResponse resp;
        return resp;
    }
    CallResponse EmptyCallHandler(const std::shared_ptr<CallMessageSpec> &req)
    {
        CallResponse resp;
        return resp;
    }
    CheckpointResponse EmptyCheckpointHandler(const CheckpointRequest &req)
    {
        CheckpointResponse resp;
        return resp;
    }
    RecoverResponse EmptyRecoverHandler(const RecoverRequest &req)
    {
        RecoverResponse resp;
        return resp;
    }
    ShutdownResponse EmptyShutdownHandler(const ShutdownRequest &req)
    {
        ShutdownResponse resp;
        return resp;
    }
    SignalResponse EmptySignalHandler(const SignalRequest &req)
    {
        SignalResponse resp;
        return resp;
    }

    FSIntfHandlers handlers_;
    std::shared_ptr<FSClient> fsClient_;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::shared_ptr<FakeGrpcServer> grpcServer;
    std::thread t;
    std::shared_ptr<MockSecurity> security_;
};

TEST_F(FSClientGrpcTest, GrpcClientTest_GroupCreateAsync)
{
    DoStartGrpcClient();

    NotificationUtility responsed;
    auto rspHandler = [&responsed](const CreateResponses &rsp) -> void { responsed.Notify(); };

    NotificationUtility notified;
    auto notifyHandler = [&notified](const NotifyRequest &req) -> void { notified.Notify(); };

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    CreateRequests req;
    req.set_requestid(reqId);
    fsClient_->GroupCreateAsync(req, rspHandler, notifyHandler, -1);

    CreateResponses resp;
    grpcServer->SendAfterRead(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), resp));
    {
        auto err = responsed.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }

    NotifyRequest notifyReq;
    notifyReq.set_requestid(reqId);
    grpcServer->Send(*GenStreamMsg("", notifyReq));
    {
        auto err = notified.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_CreateAsync)
{
    DoStartGrpcClient();

    NotificationUtility responsed;
    auto rspHandler = [&responsed](const CreateResponse &rsp) -> void { responsed.Notify(); };

    NotificationUtility notified;
    auto notifyHandler = [&notified](const NotifyRequest &req) -> void { notified.Notify(); };

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    CreateRequest req;
    req.set_requestid(reqId);
    fsClient_->CreateAsync(req, rspHandler, notifyHandler, -1);

    CreateResponse resp;
    grpcServer->SendAfterRead(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), resp));
    {
        auto err = responsed.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }

    NotifyRequest notifyReq;
    notifyReq.set_requestid(reqId);
    grpcServer->Send(*GenStreamMsg("", notifyReq));
    {
        auto err = notified.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_CreateAsyncTimeout)
{
    DoStartGrpcClient();

    NotificationUtility responsed;
    auto rspHandler = [&responsed](const CreateResponse &rsp) -> void { responsed.Notify(); };

    NotificationUtility notified;
    auto notifyHandler = [&notified](const NotifyRequest &req) -> void {
        auto err = ErrorInfo(static_cast<ErrorCode>(req.code()), req.message());
        notified.Notify(err);
    };

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    CreateRequest req;
    req.set_requestid(reqId);
    fsClient_->CreateAsync(req, rspHandler, notifyHandler, 1);
    {
        auto err = notified.WaitForNotification();
        EXPECT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_InvokeAsync)
{
    DoStartGrpcClient();

    NotificationUtility notified;
    auto notifyHandler = [&notified](const NotifyRequest &req, const ErrorInfo &err) -> void { notified.Notify(); };

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    InvokeRequest req;
    req.set_requestid(reqId);
    auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
    fsClient_->InvokeAsync(messageSpec, notifyHandler, -1);

    InvokeResponse resp;
    grpcServer->SendAfterRead(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), resp));

    NotifyRequest notifyReq;
    notifyReq.set_requestid(reqId);
    grpcServer->Send(*GenStreamMsg("", notifyReq));
    {
        auto err = notified.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_InvokeAsyncTimeout)
{
    DoStartGrpcClient();

    NotificationUtility notified;
    auto notifyHandler = [&notified](const NotifyRequest &req, const ErrorInfo &errInput) -> void {
        auto err = ErrorInfo(static_cast<ErrorCode>(req.code()), req.message());
        notified.Notify(err);
    };

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    InvokeRequest req;
    req.set_requestid(reqId);
    auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
    fsClient_->InvokeAsync(messageSpec, notifyHandler, 1);
    {
        auto err = notified.WaitForNotification();
        EXPECT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);
        EXPECT_TRUE(err.Msg().find("invoke request timeout with") != std::string::npos);
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_CallResultAsync)
{
    DoStartGrpcClient();

    NotificationUtility acked;
    auto ackHandler = [&acked](const CallResultAck &req) -> void { acked.Notify(); };

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    CallResult req;
    req.set_requestid(reqId);
    auto messageSpec = std::make_shared<CallResultMessageSpec>();
    messageSpec->Mutable() = std::move(req);
    fsClient_->CallResultAsync(messageSpec, ackHandler);

    CallResultAck ack;
    grpcServer->SendAfterRead(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), ack));
    {
        auto err = acked.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_ReturnCallResult)
{
    DoStartGrpcClient();

    NotificationUtility acked;
    auto ackHandler = [&acked](const CallResultAck &req) -> void { acked.Notify(); };

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    CallResult req;
    req.set_requestid(reqId);
    auto messageSpec = std::make_shared<CallResultMessageSpec>();
    messageSpec->Mutable() = std::move(req);
    fsClient_->ReturnCallResult(messageSpec, false, ackHandler);

    CallResultAck ack;
    grpcServer->SendAfterRead(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), ack));
    {
        auto err = acked.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_KillAsync)
{
    DoStartGrpcClient();

    NotificationUtility notify;
    auto cb = [&notify](const KillResponse &req, const ErrorInfo &err) -> void { notify.Notify(); };

    KillRequest req;
    fsClient_->KillAsync(req, cb, -1);

    StreamingMessage msg;
    auto ret = grpcServer->Read(msg);
    EXPECT_TRUE(ret);

    KillResponse resp;
    grpcServer->Send(*GenStreamMsg(msg.messageid(), resp));
    {
        auto err = notify.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_CreateRGroupAsync)
{
    DoStartGrpcClient();

    NotificationUtility notify;
    auto cb = [&notify](const CreateResourceGroupResponse &resp) -> void { notify.Notify(); };
    auto reqId = YR::utility::IDGenerator::GenRequestId();
    CreateResourceGroupRequest req;
    req.set_requestid(reqId);
    fsClient_->CreateRGroupAsync(req, cb, 1000);
    StreamingMessage msg;
    auto ret = grpcServer->Read(msg);
    EXPECT_TRUE(ret);

    CreateResourceGroupResponse resp;
    grpcServer->Send(*GenStreamMsg(msg.messageid(), resp));
    {
        auto err = notify.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_KillAsyncTimeout)
{
    DoStartGrpcClient();

    NotificationUtility notify;
    auto cb = [&notify](const KillResponse &req, const ErrorInfo &fakeErr) -> void {
        auto err = ErrorInfo(static_cast<ErrorCode>(req.code()), req.message());
        notify.Notify(err);
    };

    KillRequest req;
    fsClient_->KillAsync(req, cb, 1);
    {
        auto err = notify.WaitForNotification();
        EXPECT_EQ(err.Code(), ERR_INNER_SYSTEM_ERROR);
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_ExitAsync)
{
    DoStartGrpcClient();

    NotificationUtility notify;
    auto cb = [&notify](const ExitResponse &req) -> void { notify.Notify(); };

    ExitRequest req;
    fsClient_->ExitAsync(req, cb);

    StreamingMessage msg;
    auto ret = grpcServer->Read(msg);
    EXPECT_TRUE(ret);

    ExitResponse resp;
    grpcServer->Send(*GenStreamMsg(msg.messageid(), resp));
    {
        auto err = notify.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_StateSaveAsync)
{
    DoStartGrpcClient();

    NotificationUtility notify;
    auto cb = [&notify](const StateSaveResponse &req) -> void { notify.Notify(); };

    StateSaveRequest req;
    fsClient_->StateSaveAsync(req, cb);

    StreamingMessage msg;
    auto ret = grpcServer->Read(msg);
    EXPECT_TRUE(ret);

    StateSaveResponse resp;
    grpcServer->Send(*GenStreamMsg(msg.messageid(), resp));
    {
        auto err = notify.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_StateLoadAsync)
{
    DoStartGrpcClient();

    NotificationUtility notify;
    auto cb = [&notify](const StateLoadResponse &req) -> void { notify.Notify(); };

    StateLoadRequest req;
    fsClient_->StateLoadAsync(req, cb);

    StreamingMessage msg;
    auto ret = grpcServer->Read(msg);
    EXPECT_TRUE(ret);

    StateLoadResponse resp;
    grpcServer->Send(*GenStreamMsg(msg.messageid(), resp));
    {
        auto err = notify.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_CallRequestInit)
{
    NotificationUtility called;
    handlers_.init = [&called](const std::shared_ptr<CallMessageSpec> &req) -> CallResponse {
        called.Notify();
        return CallResponse();
    };

    DoStartGrpcClient();

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    CallRequest req;
    req.set_requestid(reqId);
    req.set_iscreate(true);
    grpcServer->Send(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), req));
    {
        auto err = called.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_CallRequestCall)
{
    NotificationUtility called;
    handlers_.init = [&called](const std::shared_ptr<CallMessageSpec> &req) -> CallResponse {
        called.Notify();
        return CallResponse();
    };

    NotificationUtility called2;
    handlers_.call = [&called2](const std::shared_ptr<CallMessageSpec> &req) -> CallResponse {
        called2.Notify();
        return CallResponse();
    };

    DoStartGrpcClient();

    {
        auto reqId = YR::utility::IDGenerator::GenRequestId();
        CallRequest req;
        req.set_requestid(reqId);
        req.set_iscreate(true);
        grpcServer->Send(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), req));
        {
            auto err = called.WaitForNotification();
            EXPECT_TRUE(err.OK());
        }
        auto ackHandler = [](const CallResultAck &req) -> void {};
        CallResult res;
        res.set_requestid(reqId);
        auto messageSpec = std::make_shared<CallResultMessageSpec>();
        messageSpec->Mutable() = std::move(res);
        fsClient_->ReturnCallResult(messageSpec, true, ackHandler);
    }
    {
        auto reqId = YR::utility::IDGenerator::GenRequestId();
        CallRequest req;
        req.set_requestid(reqId);
        req.set_iscreate(false);
        grpcServer->Send(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), req));
        {
            auto err = called2.WaitForNotification();
            EXPECT_TRUE(err.OK());
        }
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_CheckpointRequest)
{
    NotificationUtility called;
    handlers_.checkpoint = [&called](const CheckpointRequest &req) -> CheckpointResponse {
        called.Notify();
        return CheckpointResponse();
    };

    DoStartGrpcClient();

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    CheckpointRequest req;
    grpcServer->Send(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), req));
    {
        auto err = called.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_RecoverRequest)
{
    NotificationUtility called;
    handlers_.recover = [&called](const RecoverRequest &req) -> RecoverResponse {
        called.Notify();
        return RecoverResponse();
    };

    DoStartGrpcClient();

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    RecoverRequest req;
    grpcServer->Send(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), req));
    {
        auto err = called.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_ShutdownRequest)
{
    NotificationUtility called;
    handlers_.shutdown = [&called](const ShutdownRequest &req) -> ShutdownResponse {
        called.Notify();
        return ShutdownResponse();
    };

    DoStartGrpcClient();

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    ShutdownRequest req;
    grpcServer->Send(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), req));
    {
        auto err = called.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, DISABLED_GrpcClientTest_SignalRequest)
{
    NotificationUtility called;
    handlers_.signal = [&called](const SignalRequest &req) -> SignalResponse {
        called.Notify();
        return SignalResponse();
    };

    DoStartGrpcClient();

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    SignalRequest req;
    grpcServer->Send(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), req));
    {
        auto err = called.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_HeartbeatRequest)
{
    NotificationUtility called;
    handlers_.heartbeat = [&called](const HeartbeatRequest &req) -> HeartbeatResponse {
        called.Notify();
        return HeartbeatResponse();
    };

    DoStartGrpcClient();

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    HeartbeatRequest req;
    grpcServer->Send(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), req));
    {
        auto err = called.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GrpcClientTest_HeartbeatRequestSync)
{
    DoStartGrpcClient();

    auto reqId = YR::utility::IDGenerator::GenRequestId();
    HeartbeatRequest req;
    grpcServer->Send(*GenStreamMsg(YR::utility::IDGenerator::GenMessageId(reqId, 0), req));

    StreamingMessage msg;
    auto ret = grpcServer->Read(msg);
    EXPECT_TRUE(ret);
    EXPECT_TRUE(msg.has_heartbeatrsp());
}

TEST_F(FSClientGrpcTest, StartByServerTest)
{
    // driver true
    fsClient_ = std::make_shared<FSClient>();
    auto err = fsClient_->Start(Config::Instance().HOST_IP(), grpcServer->GetPort(), handlers_,
                                FSClient::ClientType::GRPC_SERVER, true, security_, clientsMgr, "12345678",
                                "instanceID", "runtimeID", "function");
    EXPECT_TRUE(err.OK());
    EXPECT_EQ(err.Msg(), "");
    fsClient_->Stop();
}

TEST_F(FSClientGrpcTest, StartByClientTest)
{
    {  // driver true
        fsClient_ = std::make_shared<FSClient>();
        auto err = fsClient_->Start(Config::Instance().HOST_IP(), grpcServer->GetPort(), handlers_,
                                    FSClient::ClientType::GRPC_CLIENT, true, security_, clientsMgr, "12345678",
                                    "instanceID", "runtimeID", "function");
        EXPECT_TRUE(err.OK());
        EXPECT_EQ(err.Msg(), "");
        EXPECT_EQ(fsClient_->GetNodeId().second, "nodeId");
        EXPECT_EQ(fsClient_->GetNodeIp().second, "hostIp");
        fsClient_->Stop();
    }

    {  // driver false
        fsClient_ = std::make_shared<FSClient>();
        auto err = fsClient_->Start(Config::Instance().HOST_IP(), grpcServer->GetPort(), handlers_,
                                    FSClient::ClientType::GRPC_CLIENT, false, security_, clientsMgr, "12345678",
                                    "instanceID", "runtimeID", "function");
        EXPECT_TRUE(err.OK());
        EXPECT_EQ(err.Msg(), "");
        fsClient_->Stop();
    }
}

TEST_F(FSClientGrpcTest, StartByClientWithDirectCallTest)
{
    Config::Instance().RUNTIME_DIRECT_CONNECTION_ENABLE() = true;
    {  // driver true
        fsClient_ = std::make_shared<FSClient>();
        auto err = fsClient_->Start(Config::Instance().HOST_IP(), grpcServer->GetPort(), handlers_,
                                    FSClient::ClientType::GRPC_CLIENT, true, security_, clientsMgr, "12345678",
                                    "instanceID", "runtimeID", "function");
        EXPECT_TRUE(err.OK());
        EXPECT_EQ(err.Msg(), "");
        fsClient_->Stop();
    }

    {  // driver false without POD_IP
        fsClient_ = std::make_shared<FSClient>();
        auto err = fsClient_->Start(Config::Instance().HOST_IP(), grpcServer->GetPort(), handlers_,
                                    FSClient::ClientType::GRPC_CLIENT, false, security_, clientsMgr, "12345678",
                                    "instanceID", "runtimeID", "function");
        EXPECT_FALSE(err.OK());
        fsClient_->Stop();
    }
    {  // driver false with POD_IP
        Config::Instance().POD_IP() = Config::Instance().HOST_IP();
        Config::Instance().DERICT_RUNTIME_SERVER_PORT() = 0;
        fsClient_ = std::make_shared<FSClient>();
        auto err = fsClient_->Start(Config::Instance().HOST_IP(), grpcServer->GetPort(), handlers_,
                                    FSClient::ClientType::GRPC_CLIENT, false, security_, clientsMgr, "12345678",
                                    "instanceID", "runtimeID", "function");
        EXPECT_TRUE(err.OK());
        EXPECT_EQ(err.Msg(), "");
        fsClient_->Stop();
    }
}

TEST_F(FSClientGrpcTest, StartByServerWithDirectCallTest)
{
    Config::Instance().RUNTIME_DIRECT_CONNECTION_ENABLE() = true;
    // driver true
    fsClient_ = std::make_shared<FSClient>();
    auto err = fsClient_->Start(Config::Instance().HOST_IP(), grpcServer->GetPort(), handlers_,
                                FSClient::ClientType::GRPC_SERVER, true, security_, clientsMgr, "12345678",
                                "instanceID", "runtimeID", "function");
    EXPECT_FALSE(err.OK());
    ASSERT_NO_THROW(fsClient_->RemoveInsRtIntf("fakeInsId"));
    fsClient_->Stop();
}

TEST_F(FSClientGrpcTest, ReconnectTest)
{
    DoStartGrpcClient();
    auto port = grpcServer->GetPort();
    grpcServer->Stop();
    grpcServer = std::make_shared<FakeGrpcServer>(Config::Instance().HOST_IP());
    grpcServer->StartWithPort(port);
    NotificationUtility notify;
    auto cb = [&notify](const KillResponse &req, const ErrorInfo &err) -> void { notify.Notify(); };

    KillRequest req;
    fsClient_->KillAsync(req, cb, -1);

    StreamingMessage msg;
    auto ret = grpcServer->Read(msg);
    EXPECT_TRUE(ret);

    KillResponse resp;
    grpcServer->Send(*GenStreamMsg(msg.messageid(), resp));
    {
        auto err = notify.WaitForNotification();
        EXPECT_TRUE(err.OK());
    }
}

TEST_F(FSClientGrpcTest, GRPC_STATUS_UNAUTHENTICATED_Should_Discover_Driver_Test)
{
    grpcServer->Stop();
    std::shared_ptr<FakeGrpcServer> grpcServerOne = std::make_shared<FakeGrpcServerOne>(Config::Instance().HOST_IP());
    grpcServerOne->Start();
    auto port = grpcServerOne->GetPort();
    fsClient_ = std::make_shared<FSClient>();
    auto err = fsClient_->Start(Config::Instance().HOST_IP(), grpcServerOne->GetPort(), handlers_,
                                FSClient::ClientType::GRPC_CLIENT, true, security_, clientsMgr, "12345678",
                                "instanceID", "runtimeID", "function");
    EXPECT_TRUE(grpcServerOne->discoverFlagFuture.get() == true);
    EXPECT_TRUE(err.OK());
    EXPECT_EQ(err.Msg(), "");
    grpcServerOne->Stop();
    grpcServerOne = std::make_shared<FakeGrpcServerOne>(Config::Instance().HOST_IP());
    grpcServerOne->StartWithPort(port);
    EXPECT_TRUE(grpcServerOne->discoverFlagFuture.get() == true);
    fsClient_->Stop();
    grpcServerOne->Stop();
}

TEST_F(FSClientGrpcTest, GrpcClientTest_EventAsync)
{
    DoStartGrpcClient();
    fsClient_->UpdateEventServerInfo(Config::Instance().HOST_IP(), grpcServer->GetPort(), "test_instace_id");
    EXPECT_EQ(fsClient_->GetEventServerIP(), Config::Instance().HOST_IP());
    EXPECT_NE(fsClient_->GetEventServerPort(), grpcServer->GetPort());
}
}  // namespace test
}  // namespace YR