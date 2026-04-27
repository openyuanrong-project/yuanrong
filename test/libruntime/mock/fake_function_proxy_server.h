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

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/libruntime/fsclient/fs_client.h"
#include "src/libruntime/fsclient/fs_intf_impl.h"
using namespace testing;
using namespace YR::utility;
using namespace YR::Libruntime;
using namespace std::placeholders;

namespace YR {
namespace test {
class FakeFunctionProxyServer : public runtime_rpc::RuntimeRPC::Service {
public:
    FakeFunctionProxyServer(const std::string &ipAddr) : ipAddr_(ipAddr) {}
    virtual ~FakeFunctionProxyServer() = default;
    void Start()
    {
        grpc::ServerBuilder builder;
        builder.AddChannelArgument(GRPC_ARG_ALLOW_REUSEPORT, 0);
        builder.SetDefaultCompressionLevel(GRPC_COMPRESS_LEVEL_NONE);
        builder.AddListeningPort(ipAddr_ + ":" + std::to_string(0), grpc::InsecureServerCredentials(), &port_);
        builder.RegisterService(static_cast<runtime_rpc::RuntimeRPC::Service *>(this));
        server_ = builder.BuildAndStart();
        EXPECT_NE(server_, nullptr) << "start FakeFunctionProxyServer failed";
    }
    grpc::Status MessageStream(grpc::ServerContext *context,
                               grpc::ServerReaderWriter<StreamingMessage, StreamingMessage> *stream) override
    {
        auto metadata = context->client_metadata();
        if (auto iter = metadata.find("instance_id"); iter != metadata.end()) {
            auto srcInstance = std::string(iter->second.data(), iter->second.size());
            if (srcInstance == "callee") {
                (void)stop.WaitForNotificationWithTimeout(absl::Seconds(30), YR::Libruntime::ErrorInfo());
                YRLOG_INFO("FakeFunctionProxyServer callee stream stoped");
                return grpc::Status::OK;
            }
        }
        stream_ = stream;
        context_ = context;
        start.Notify();
        (void)stop.WaitForNotificationWithTimeout(absl::Seconds(30), YR::Libruntime::ErrorInfo());
        YRLOG_INFO("FakeFunctionProxyServer caller stoped ");
        return grpc::Status::OK;
    }

    void Stop()
    {
        YRLOG_INFO("begin to stop FakeFunctionProxyServer");
        stop.Notify();
    }

    int GetPort()
    {
        return port_;
    }

    std::string GetRuntimeServerPort()
    {
        return runtimeServerPort;
    }

    void Send(const StreamingMessage &msg)
    {
        start.WaitForNotification();
        stream_->Write(msg);
    }

    void SendAfterRead(const StreamingMessage &msg)
    {
        start.WaitForNotification();
        StreamingMessage req;
        Read(req);
        stream_->Write(msg);
    }

    bool Read(StreamingMessage &req)
    {
        start.WaitForNotification();
        return stream_->Read(&req);
    }

private:
    std::string ipAddr_;
    int port_;
    std::unique_ptr<grpc::Server> server_;
    grpc::ServerReaderWriter<StreamingMessage, StreamingMessage> *stream_ = nullptr;
    grpc::ServerContext *context_ = nullptr;
    NotificationUtility start;
    NotificationUtility stop;
    std::string runtimeServerPort;
};
} // namespace test
} // namespace YR