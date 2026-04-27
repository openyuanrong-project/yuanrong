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

#include <gtest/gtest.h>
#include <future>
#include <unordered_map>
#include "mock/mock_security.h"
#include "src/libruntime/utils/grpc_utils.h"
#include "src/libruntime/clientsmanager/clients_manager.h"
#include "src/libruntime/fsclient/grpc/fs_intf_grpc_client_reader_writer.h"
#include "src/utility/logger/logger.h"
#define private public
#include "src/libruntime/fsclient/grpc/grpc_posix_service.h"

using namespace YR::Libruntime;
using namespace YR::utility;
using namespace std::placeholders;

class FSIntfGrpcRWTest : public ::testing::Test {
protected:
    std::shared_ptr<GrpcPosixService> service;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::shared_ptr<FSIntfManager> fsIntfManager;
    std::shared_ptr<YR::Libruntime::Security> security_;
    std::string ip = "127.0.0.1";
    int listenPort = 12345;
    ThreadPool writer;
    std::unordered_map<std::string, std::shared_ptr<std::promise<StreamingMessage>>> msgs;
    mutable absl::Mutex mu;

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
        clientsMgr = std::make_shared<ClientsManager>();
        fsIntfManager = std::make_shared<FSIntfManager>(clientsMgr);
        security_ = std::make_shared<YR::test::MockSecurity>();
        writer.Init(1, "FSIntfGrpcRWTest");
    }

    void TearDown() override
    {
        clientsMgr = nullptr;
        fsIntfManager->Clear();
        fsIntfManager = nullptr;
        security_.reset();
        writer.Shutdown();
    }

    void RegisterMessagePromise(const std::string msgID, const std::shared_ptr<std::promise<StreamingMessage>> &promise)
    {
        absl::MutexLock lock(&this->mu);
        msgs[msgID] = promise;
    }
    void RecvMsg(const std::string &, const std::shared_ptr<StreamingMessage> &messages)
    {
        absl::MutexLock lock(&this->mu);
        if (msgs.find(messages->messageid()) != msgs.end()) {
            msgs[messages->messageid()]->set_value(*messages);

        }
    }
    std::unordered_map<BodyCase, MsgHdlr> rtMsgHdlrs = {
        {StreamingMessage::kCallReq, std::bind(&FSIntfGrpcRWTest::RecvMsg, this, _1, _2)},
        {StreamingMessage::kInvokeRsp, std::bind(&FSIntfGrpcRWTest::RecvMsg, this, _1, _2)},
        {StreamingMessage::kNotifyReq, std::bind(&FSIntfGrpcRWTest::RecvMsg, this, _1, _2)},
        {StreamingMessage::kCallResultAck, std::bind(&FSIntfGrpcRWTest::RecvMsg, this, _1, _2)},
    };
    std::unordered_map<BodyCase, MsgHdlr> registerEventHdlrs = {
        {StreamingMessage::kEventReq, std::bind(&FSIntfGrpcRWTest::RecvMsg, this, _1, _2)},
    };
    void StartService(std::function<void(const std::string &)> resendCb = nullptr,
                      std::function<void(const std::string &)> disconnectedCb = nullptr, bool enableEvent = false)
    {
        service = std::make_shared<GrpcPosixService>("server", "runtime-server", ip, listenPort,
                                                     std::make_shared<TimerWorker>(),
                                                     std::make_shared<absl::Notification>(), fsIntfManager, security_,
                                                     true, enableEvent);
        service->rtDisconnectedTimeout = 100;
        service->fsDisconnectedTimeout = 100;
        service->RegisterRTHandler(rtMsgHdlrs);
        service->RegisterEventHdlrs(registerEventHdlrs);
        service->RegisterResendCallback(resendCb);
        service->RegisterDisconnectedCallback(disconnectedCb);
        service->Start();
    }
    void StopService()
    {
        fsIntfManager->Clear();
        if (service != nullptr) {
            service->Stop();
        }
        service = nullptr;
    }

    std::shared_ptr<FSIntfGrpcClientReaderWriter> StartClient(
        std::string clientName, std::function<void(const std::string &)> resendCb = nullptr,
        std::function<void(const std::string &)> disconnectedCb = nullptr)
    {
        auto clientRw =
            std::make_shared<FSIntfGrpcClientReaderWriter>(clientName, "server", "runtime-client", clientsMgr,
                                                           ReaderWriterClientOption{.ip = ip,
                                                                                    .port = listenPort,
                                                                                    .disconnectedTimeout = 100,
                                                                                    .security = this->security_,
                                                                                    .resendCb = resendCb,
                                                                                    .disconnectedCb = disconnectedCb});
        clientRw->RegisterMessageHandler(rtMsgHdlrs);
        clientRw->Start();
        auto ready = std::promise<void>();
        writer.Handle(
            [&]() {
                auto begin = std::chrono::steady_clock::now();
                while (std::chrono::steady_clock::now() - begin < std::chrono::milliseconds(5000)) {
                    if (fsIntfManager->TryGet(clientName) != nullptr) {
                        ready.set_value();
                        return;
                    }
                }
            },
            "");
        ready.get_future().wait();
        return clientRw;
    }

    std::shared_ptr<FSIntfGrpcClientReaderWriter> StartEventClient(
        std::string clientName, std::function<void(const std::string &)> resendCb = nullptr,
        std::function<void(const std::string &)> disconnectedCb = nullptr, bool enableAuth = false)
    {
        if (enableAuth) {
            std::string ak = "ak";
            SensitiveValue sk = std::string("sk");
            this->security_->SetAKSKAndCredential(ak, sk);
            Config::Instance().YR_TENANT_ID() = ak;
        }
        auto clientRw =
            std::make_shared<FSIntfGrpcClientReaderWriter>(clientName, "server", "runtime-client", clientsMgr,
                                                           ReaderWriterClientOption{.ip = ip,
                                                                                    .port = listenPort,
                                                                                    .disconnectedTimeout = 100,
                                                                                    .security = this->security_,
                                                                                    .resendCb = resendCb,
                                                                                    .disconnectedCb = disconnectedCb});
        clientRw->RegisterMessageHandler(registerEventHdlrs);
        clientRw->Start();
        auto ready = std::promise<void>();
        writer.Handle(
            [&]() {
                auto begin = std::chrono::steady_clock::now();
                while (std::chrono::steady_clock::now() - begin < std::chrono::milliseconds(5000)) {
                    if (fsIntfManager->TryGetEventIntfs(clientName) != nullptr) {
                        ready.set_value();
                        return;
                    }
                }
            },
            "");
        ready.get_future().wait();
        return clientRw;
    }
};

TEST_F(FSIntfGrpcRWTest, TestSignRequest)
{
    std::string ak = "ak";
    SensitiveValue sk = std::string("sk");
    this->security_->SetAKSKAndCredential(ak, sk);
    StartService();
    auto clientRw = StartClient("client");
    auto recvPromise = std::make_shared<std::promise<StreamingMessage>>();
    std::string msgID = "invokereq";
    RegisterMessagePromise(msgID, recvPromise);
    auto msg = StreamingMessage();
    msg.set_messageid(msgID);
    auto rsp = msg.mutable_invokereq();
    rsp->set_requestid("request");
    rsp->set_function("function");
    rsp->set_traceid("traceid");
    rsp->set_instanceid("server");
    (*rsp->mutable_invokeoptions()->mutable_customtag())["custom"] = "value";
    auto arg = rsp->add_args();
    arg->set_value("args_value");
    YR::Libruntime::SignStreamingMessage(ak, sk, msg);
    std::promise<ErrorInfo> writecb;
    auto ptr = std::make_shared<StreamingMessage>(std::move(msg));
    clientRw->Write(ptr, [&](bool, ErrorInfo err) { writecb.set_value(err); });
    auto err = writecb.get_future().get();
    ASSERT_EQ(err.OK(), true);
    StopService();
}

TEST_F(FSIntfGrpcRWTest, SendInvokeMsg)
{
    StartService();
    auto clientRw = StartClient("client");
    // send invoke, received callreq
    auto recvPromise = std::make_shared<std::promise<StreamingMessage>>();
    std::string msgID = "invokereq";
    RegisterMessagePromise(msgID, recvPromise);
    auto msg = StreamingMessage();
    msg.set_messageid(msgID);
    auto invokereq = msg.mutable_invokereq();
    invokereq->set_requestid("request");
    invokereq->set_function("function");
    invokereq->set_traceid("traceid");
    invokereq->set_instanceid("server");
    (*invokereq->mutable_invokeoptions()->mutable_customtag())["custom"] = "value";
    auto arg = invokereq->add_args();
    arg->set_value("args_value");
    std::promise<ErrorInfo> writecb;
    auto ptr = std::make_shared<StreamingMessage>(std::move(msg));
    clientRw->Write(ptr, [&](bool, ErrorInfo err) { writecb.set_value(err); });
    auto err = writecb.get_future().get();
    ASSERT_EQ(err.OK(), true);

    auto callmsg = recvPromise->get_future().get();
    ASSERT_EQ(callmsg.messageid(), msgID);
    auto callreq = callmsg.callreq();
    ASSERT_EQ(callreq.requestid(), invokereq->requestid());
    ASSERT_EQ(callreq.function(), invokereq->function());
    ASSERT_EQ(callreq.traceid(), invokereq->traceid());
    ASSERT_EQ(callreq.args_size(), invokereq->args_size());
    ASSERT_EQ(callreq.args(0).value(), invokereq->args(0).value());
    ASSERT_EQ(callreq.createoptions().at("custom"), invokereq->invokeoptions().customtag().at("custom"));
    clientRw->Stop();
    StopService();
}

TEST_F(FSIntfGrpcRWTest, SendCallRspMsg)
{
    StartService();
    auto clientRw = StartClient("client");
    // send callrsp, received invokersp
    auto recvPromise = std::make_shared<std::promise<StreamingMessage>>();
    std::string msgID = "callrsp";
    RegisterMessagePromise(msgID, recvPromise);
    auto msg = StreamingMessage();
    msg.set_messageid(msgID);
    auto rsp = msg.mutable_callrsp();
    rsp->set_code(common::ErrorCode::ERR_INSTANCE_EXITED);
    rsp->set_message("err");
    std::promise<ErrorInfo> writecb;
    auto ptr = std::make_shared<StreamingMessage>(std::move(msg));
    clientRw->Write(ptr, [&](bool, ErrorInfo err) { writecb.set_value(err); });
    auto err = writecb.get_future().get();
    ASSERT_EQ(err.OK(), true);

    auto recvrsp = recvPromise->get_future().get();
    ASSERT_EQ(recvrsp.messageid(), msgID);
    auto invokersp = recvrsp.invokersp();
    ASSERT_EQ(invokersp.code(), rsp->code());
    ASSERT_EQ(invokersp.message(), rsp->message());
    clientRw->Stop();
    StopService();
}

TEST_F(FSIntfGrpcRWTest, SendCallResultMsg)
{
    StartService();
    auto clientRw = StartClient("client");
    // send callresult, received notifyreq
    auto recvPromise = std::make_shared<std::promise<StreamingMessage>>();
    std::string msgID = "callresult";
    RegisterMessagePromise(msgID, recvPromise);
    auto msg = StreamingMessage();
    msg.set_messageid(msgID);
    auto rsp = msg.mutable_callresultreq();
    rsp->set_code(common::ErrorCode::ERR_USER_FUNCTION_EXCEPTION);
    rsp->set_message("err");
    rsp->set_requestid("requestid");
    rsp->set_instanceid("instanceid");
    auto obj = rsp->add_smallobjects();
    obj->set_id("small");
    auto stack = rsp->add_stacktraceinfos();
    stack->set_message("stack");
    rsp->mutable_runtimeinfo()->set_serveripaddr("127.0.0.1");
    std::promise<ErrorInfo> writecb;
    auto ptr = std::make_shared<StreamingMessage>(std::move(msg));
    clientRw->Write(ptr, [&](bool, ErrorInfo err) { writecb.set_value(err); });
    auto err = writecb.get_future().get();
    ASSERT_EQ(err.OK(), true);

    auto recvrsp = recvPromise->get_future().get();
    ASSERT_EQ(recvrsp.messageid(), msgID);
    auto notifyreq = recvrsp.notifyreq();
    ASSERT_EQ(notifyreq.code(), rsp->code());
    ASSERT_EQ(notifyreq.message(), rsp->message());
    ASSERT_EQ(notifyreq.requestid(), rsp->requestid());
    ASSERT_EQ(notifyreq.smallobjects_size(), rsp->smallobjects_size());
    ASSERT_EQ(notifyreq.smallobjects(0).id(), rsp->smallobjects(0).id());
    ASSERT_EQ(notifyreq.stacktraceinfos_size(), rsp->stacktraceinfos_size());
    ASSERT_EQ(notifyreq.stacktraceinfos(0).message(), rsp->stacktraceinfos(0).message());
    ASSERT_EQ(notifyreq.has_runtimeinfo(), false);
    clientRw->Stop();
    StopService();
}

TEST_F(FSIntfGrpcRWTest, SendNotifyResponseMsg)
{
    StartService();
    auto clientRw = StartClient("client");
    // send notifyrsp, received callresultack
    auto recvPromise = std::make_shared<std::promise<StreamingMessage>>();
    std::string msgID = "notifyrsp";
    RegisterMessagePromise(msgID, recvPromise);
    auto msg = StreamingMessage();
    msg.set_messageid(msgID);
    auto rsp = msg.mutable_notifyrsp();
    auto serverRw = fsIntfManager->Get("client");
    std::promise<ErrorInfo> writecb;
    auto ptr = std::make_shared<StreamingMessage>(std::move(msg));
    serverRw->Write(ptr, [&](bool, ErrorInfo err) { writecb.set_value(err); });
    auto err = writecb.get_future().get();
    ASSERT_EQ(err.OK(), true);
    auto recvrsp = recvPromise->get_future().get();
    ASSERT_EQ(recvrsp.messageid(), msgID);
    auto callresultack = recvrsp.callresultack();
    ASSERT_EQ(callresultack.code(), common::ErrorCode::ERR_NONE);
    clientRw->Stop();
    StopService();
}

TEST_F(FSIntfGrpcRWTest, ClientDisconnected)
{
    std::promise<std::string> resend;
    auto resendCB = [&](const std::string &remote){
        static std::once_flag flag;
        std::call_once(flag, [&]{
            resend.set_value(remote);
        });
    };
    std::promise<std::string> disconnect;
    auto disconnectedCB = [&](const std::string &remote){
        static std::once_flag flag;
        std::call_once(flag, [&]{
            disconnect.set_value(remote);
        });
    };
    StartService(resendCB, disconnectedCB);
    auto clientRw = StartClient("client");
    clientRw->Stop();
    ASSERT_EQ(disconnect.get_future().get(), "client");
    clientRw = nullptr;
    clientRw = StartClient("client");
    ASSERT_EQ(resend.get_future().get(), "client");
    clientRw->Stop();
    StopService();
}

TEST_F(FSIntfGrpcRWTest, ServerDisconnected)
{
    std::promise<std::string> resend;
    auto resendCB = [&](const std::string &remote) {
        static std::once_flag flag;
        std::call_once(flag, [&]{
            resend.set_value(remote);
        }); };
    std::promise<std::string> disconnect;
    auto disconnectedCB = [&](const std::string &remote) {
        static std::once_flag flag;
        std::call_once(flag, [&]{
            disconnect.set_value(remote);
        }); };
    StartService();
    auto clientRw = StartClient("client", resendCB, disconnectedCB);
    auto serverRw = fsIntfManager->Get("client");
    serverRw->Stop();
    ASSERT_EQ(resend.get_future().get(), "server");
    StopService();
    ASSERT_EQ(disconnect.get_future().get(), "server");
    clientRw->Stop();
}

TEST_F(FSIntfGrpcRWTest, MultipleClient)
{
    StartService();
    auto clientRw1 = StartClient("clientRw1");
    auto clientRw2 = StartClient("clientRw2");

    {
        std::string msgID = "notifyrsp1";
        auto recvPromise = std::make_shared<std::promise<StreamingMessage>>();
        RegisterMessagePromise(msgID, recvPromise);
        auto msg = StreamingMessage();
        msg.set_messageid(msgID);
        auto rsp = msg.mutable_notifyrsp();
        std::promise<ErrorInfo> writecb;
        auto ptr = std::make_shared<StreamingMessage>(std::move(msg));
        clientRw1->Write(ptr, [&](bool, ErrorInfo err) { writecb.set_value(err); });
        auto err = writecb.get_future().get();
        ASSERT_EQ(err.OK(), true);
        auto recvrsp = recvPromise->get_future().get();
        ASSERT_EQ(recvrsp.messageid(), msgID);
    }
    {
        std::string msgID = "notifyrsp2";
        auto recvPromise = std::make_shared<std::promise<StreamingMessage>>();
        RegisterMessagePromise(msgID, recvPromise);
        auto msg = StreamingMessage();
        msg.set_messageid(msgID);
        auto rsp = msg.mutable_notifyrsp();
        std::promise<ErrorInfo> writecb;
        auto ptr = std::make_shared<StreamingMessage>(std::move(msg));
        clientRw2->Write(ptr, [&](bool, ErrorInfo err) { writecb.set_value(err); });
        auto err = writecb.get_future().get();
        ASSERT_EQ(err.OK(), true);
        auto recvrsp = recvPromise->get_future().get();
        ASSERT_EQ(recvrsp.messageid(), msgID);
    }
    clientRw1->Stop();
    clientRw2->Stop();
    StopService();
}

TEST_F(FSIntfGrpcRWTest, BatchWrite)
{
    StartService();
    auto clientRw1 = StartClient("clientRw1");
    ASSERT_EQ(clientRw1->IsHealth(), true);
    struct Promises {
        std::shared_ptr<std::promise<StreamingMessage>> recvPromise;
        std::shared_ptr<std::promise<ErrorInfo>> writecb;
    };
    std::unordered_map<std::string, Promises> infos;
    for (int i = 0; i < 100; i++)
    {
        std::string msgID = YR::utility::IDGenerator::GenRequestId();
        auto recvPromise = std::make_shared<std::promise<StreamingMessage>>();
        RegisterMessagePromise(msgID, recvPromise);
        auto msg = StreamingMessage();
        msg.set_messageid(msgID);
        auto rsp = msg.mutable_notifyrsp();
        auto writecb = std::make_shared<std::promise<ErrorInfo>>() ;
        infos[msgID] = Promises{recvPromise, writecb};
        auto ptr = std::make_shared<StreamingMessage>(std::move(msg));
        clientRw1->Write(ptr, [writecb](bool, ErrorInfo err) {
            writecb->set_value(err);
        });
    }
    for (auto &[msgID, promises] : infos) {
        auto writecb = promises.writecb;
        auto recvPromise = promises.recvPromise;
        auto err = writecb->get_future().get();
        ASSERT_EQ(err.OK(), true);
        auto recvrsp = recvPromise->get_future().get();
        ASSERT_EQ(recvrsp.messageid(), msgID);
    }
    clientRw1->Stop();
    StopService();
}