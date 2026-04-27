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

#pragma once
#include <grpcpp/grpcpp.h>

#include "mock_security.h"
#include "src/libruntime/fsclient/fs_intf.h"

namespace YR {
namespace test {
class FakeGrpcClient {
public:
    FakeGrpcClient(const std::string &ipAddr, const std::string &port);
    virtual ~FakeGrpcClient();
    void Start(void);
    void Stop(void);
    void RunMessageStream(void);

private:
    std::string ipAddr_;
    std::string port_;
    std::unique_ptr<runtime_rpc::RuntimeRPC::Stub> stub_;
    std::unique_ptr<grpc::ClientReaderWriter<runtime_rpc::StreamingMessage, runtime_rpc::StreamingMessage>> stream_;
    YR::utility::NotificationUtility start;
    YR::utility::NotificationUtility stop;
    std::thread t;
    std::shared_ptr<MockSecurity> security_;
};

class FakeGrpcServer : public runtime_rpc::RuntimeRPC::Service, public bus_service::BusService::Service {
public:
    FakeGrpcServer(const std::string &ipAddr);
    virtual ~FakeGrpcServer() = default;
    void Start();
    void StartWithPort(int port);
    virtual grpc::Status MessageStream(
        grpc::ServerContext *context,
        grpc::ServerReaderWriter<runtime_rpc::StreamingMessage, runtime_rpc::StreamingMessage> *stream) override;

    ::grpc::Status DiscoverDriver(::grpc::ServerContext *context, const ::bus_service::DiscoverDriverRequest *request,
                                  ::bus_service::DiscoverDriverResponse *response) override;
    void Stop();

    int GetPort();

    std::string GetRuntimeServerPort();

    void Send(const runtime_rpc::StreamingMessage &msg);

    void SendAfterRead(const runtime_rpc::StreamingMessage &msg);

    bool Read(runtime_rpc::StreamingMessage &req);
    std::promise<bool> discoverFlagPromise;
    std::future<bool> discoverFlagFuture;

protected:
    std::string ipAddr_;
    int port_;
    std::unique_ptr<grpc::Server> server_;
    grpc::ServerReaderWriter<runtime_rpc::StreamingMessage, runtime_rpc::StreamingMessage> *stream_ = nullptr;
    grpc::ServerContext *context_ = nullptr;
    YR::utility::NotificationUtility start;
    YR::utility::NotificationUtility stop;
    std::string runtimeServerPort;
    std::shared_ptr<FakeGrpcClient> fakeGrpcClient = nullptr;
};

class FakeGrpcServerOne : public FakeGrpcServer {
public:
    FakeGrpcServerOne(const std::string &ipAddr) : FakeGrpcServer(ipAddr) {}
    virtual ~FakeGrpcServerOne() = default;
    grpc::Status MessageStream(
        grpc::ServerContext *context,
        grpc::ServerReaderWriter<runtime_rpc::StreamingMessage, runtime_rpc::StreamingMessage> *stream) override;
};
}  // namespace test
}  // namespace YR