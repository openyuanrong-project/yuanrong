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

#include "mock_fs_client_and_server.h"

namespace YR {
namespace test {
FakeGrpcClient::FakeGrpcClient(const std::string &ipAddr, const std::string &port) : ipAddr_(ipAddr), port_(port)
{
    this->security_ = std::make_shared<MockSecurity>();
}

FakeGrpcClient::~FakeGrpcClient()
{
    Stop();
}

void FakeGrpcClient::Start(void)
{
    grpc::ChannelArguments args;
    auto channel = grpc::CreateChannel(ipAddr_ + ":" + port_, grpc::InsecureChannelCredentials());
    stub_ = runtime_rpc::RuntimeRPC::NewStub(channel);
    EXPECT_NE(stub_, nullptr);
    t = std::thread(&FakeGrpcClient::RunMessageStream, this);
}

void FakeGrpcClient::Stop(void)
{
    stop.Notify();
    t.join();
}

void FakeGrpcClient::RunMessageStream(void)
{
    auto ctx = grpc::ClientContext();
    stream_ = stub_->MessageStream(&ctx);
    start.Notify();
    (void)stop.WaitForNotification();
}

FakeGrpcServer::FakeGrpcServer(const std::string &ipAddr) : ipAddr_(ipAddr)
{
    discoverFlagFuture = discoverFlagPromise.get_future();
}

void FakeGrpcServer::Start()
{
    grpc::ServerBuilder builder;
    builder.AddChannelArgument(GRPC_ARG_ALLOW_REUSEPORT, 0);
    builder.SetDefaultCompressionLevel(GRPC_COMPRESS_LEVEL_NONE);
    builder.AddListeningPort(ipAddr_ + ":" + std::to_string(0), grpc::InsecureServerCredentials(), &port_);
    builder.RegisterService(static_cast<runtime_rpc::RuntimeRPC::Service *>(this));
    builder.RegisterService(static_cast<bus_service::BusService::Service *>(this));
    server_ = builder.BuildAndStart();
    EXPECT_NE(server_, nullptr) << "start FakeGrpcServer failed";
}

void FakeGrpcServer::StartWithPort(int port)
{
    port_ = port;
    grpc::ServerBuilder builder;
    builder.AddChannelArgument(GRPC_ARG_ALLOW_REUSEPORT, 0);
    builder.SetDefaultCompressionLevel(GRPC_COMPRESS_LEVEL_NONE);
    builder.AddListeningPort(ipAddr_ + ":" + std::to_string(port_), grpc::InsecureServerCredentials());
    builder.RegisterService(static_cast<runtime_rpc::RuntimeRPC::Service *>(this));
    builder.RegisterService(static_cast<bus_service::BusService::Service *>(this));
    server_ = builder.BuildAndStart();
    EXPECT_NE(server_, nullptr) << "start FakeGrpcServer failed";
}

grpc::Status FakeGrpcServer::MessageStream(
    grpc::ServerContext *context,
    grpc::ServerReaderWriter<runtime_rpc::StreamingMessage, runtime_rpc::StreamingMessage> *stream)
{
    stream_ = stream;
    context_ = context;
    start.Notify();
    (void)stop.WaitForNotificationWithTimeout(absl::Seconds(30), YR::Libruntime::ErrorInfo());
    return grpc::Status::OK;
}

::grpc::Status FakeGrpcServer::DiscoverDriver(::grpc::ServerContext *context,
                                              const ::bus_service::DiscoverDriverRequest *request,
                                              ::bus_service::DiscoverDriverResponse *response)
{
    runtimeServerPort = request->driverport();
    fakeGrpcClient = std::make_shared<FakeGrpcClient>(ipAddr_, runtimeServerPort);
    fakeGrpcClient->Start();
    discoverFlagPromise.set_value(true);
    *response = ::bus_service::DiscoverDriverResponse();
    response->set_nodeid("nodeId");
    response->set_hostip("hostIp");
    return ::grpc::Status::OK;
}

void FakeGrpcServer::Stop()
{
    YRLOG_INFO("begin to stop FakeGrpcServer");
    stop.Notify();
    if (server_) {
        auto tmout = gpr_time_add(gpr_now(GPR_CLOCK_MONOTONIC), {1, 0, GPR_TIMESPAN});
        this->server_->Shutdown(tmout);
    }
}

int FakeGrpcServer::GetPort()
{
    return port_;
}

std::string FakeGrpcServer::GetRuntimeServerPort()
{
    return runtimeServerPort;
}

void FakeGrpcServer::Send(const runtime_rpc::StreamingMessage &msg)
{
    start.WaitForNotification();
    stream_->Write(msg);
}

void FakeGrpcServer::SendAfterRead(const runtime_rpc::StreamingMessage &msg)
{
    start.WaitForNotification();
    runtime_rpc::StreamingMessage req;
    Read(req);
    stream_->Write(msg);
}

bool FakeGrpcServer::Read(runtime_rpc::StreamingMessage &req)
{
    start.WaitForNotification();
    return stream_->Read(&req);
}
grpc::Status FakeGrpcServerOne::MessageStream(
    grpc::ServerContext *context,
    grpc::ServerReaderWriter<runtime_rpc::StreamingMessage, runtime_rpc::StreamingMessage> *stream)
{
    stream_ = stream;
    context_ = context;
    return grpc::Status(grpc::StatusCode::UNAUTHENTICATED, "err");
}
}  // namespace test
}  // namespace YR