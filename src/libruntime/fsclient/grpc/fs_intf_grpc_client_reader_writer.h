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

#pragma once
#include "fs_intf_grpc_reader_writer.h"

#include "src/libruntime/clientsmanager/clients_manager.h"
#include "src/utility/thread_pool.h"

namespace YR {
namespace Libruntime {
using Stream = std::unique_ptr<grpc::ClientReaderWriter<runtime_rpc::StreamingMessage, runtime_rpc::StreamingMessage>>;
using BatchStream =
    std::unique_ptr<grpc::ClientReaderWriter<runtime_rpc::BatchStreamingMessage, runtime_rpc::BatchStreamingMessage>>;
class FSIntfGrpcClientReaderWriter : public FSIntfGrpcReaderWriter {
public:
    const std::string INSTANCE_ID_META = "instance_id";
    const std::string RUNTIME_ID_META = "runtime_id";
    const std::string TOKEN_META = "authorization";
    const std::string TENANT_ACCESS_KEY = "access_key";
    const std::string SIGNATURE = "signature";
    const std::string TIMESTAMP = "timestamp";
    const std::string TENANT_ID = "tenant_id";
    const std::string SOURCE_ID_META = "source_id";
    const std::string DST_ID_META = "dst_id";
    const std::string JOB_ID_META = "job_id";
    const std::string RPC_IP_NAME = "HOST_IP";
    const std::string RPC_PORT_NAME = "PROXY_GRPC_SERVER_PORT";

    FSIntfGrpcClientReaderWriter(const std::string &srcInstance, const std::string &dstInstance,
                                 const std::string &runtimeID, const std::shared_ptr<ClientsManager> &clientsMgr,
                                 const ReaderWriterClientOption &option)
        : FSIntfGrpcReaderWriter(srcInstance, dstInstance, runtimeID),
          ip(option.ip),
          port(option.port),
          security(option.security),
          clientsMgr(clientsMgr),
          resendCb(option.resendCb),
          disconnectedCb(option.disconnectedCb),
          disconnectedTimeout(option.disconnectedTimeout),
          isKeepAlive(option.isKeepAlive)
    {
        disconnTime = std::chrono::steady_clock::now();
    }

    ~FSIntfGrpcClientReaderWriter()
    {
        this->Stop();
    }

    ErrorInfo NewGrpcClientWithRetry(const int retryTimes = RETRY_TIME);
    ErrorInfo BuildStreamWithRetry(std::shared_ptr<grpc::Channel> channel, const int retryTimes = RETRY_TIME);
    void PreStart() override {}
    ErrorInfo Start() override;
    bool IsHealth() override;
    bool IsSameDstAddr(const std::string &dstIp, const int &dstPort) override;
    ErrorInfo Reconnect();
    void Stop() override;
    bool GrpcRead(const std::shared_ptr<StreamingMessage> &message) override;
    bool GrpcWrite(const std::shared_ptr<StreamingMessage> &request) override;
    bool GrpcBatchRead(const std::shared_ptr<BatchStreamingMessage> &message) override;
    bool GrpcBatchWrite(const std::shared_ptr<BatchStreamingMessage> &request) override;
    bool IsBatched() override;

private:
    void ReceiveHandler();
    void ReconnectHandler();
    bool StreamEmpty();
    void WritesDone();
    grpc::Status Finish();
    ErrorInfo BuildStream(std::shared_ptr<grpc::Channel> channel);
    void Reset();
    std::string ip;
    int port;
    std::atomic<bool> stopped = {false};
    std::unique_ptr<runtime_rpc::RuntimeRPC::Stub> stub_ = nullptr;
    std::shared_ptr<::grpc::ClientContext> context = nullptr;
    Stream stream_ = nullptr;
    BatchStream batchStream_ = nullptr;
    ThreadPool receiver_;
    std::chrono::time_point<std::chrono::steady_clock> disconnTime;
    std::shared_ptr<Security> security;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::function<void(const std::string &)> resendCb;
    std::function<void(const std::string &)> disconnectedCb;
    int disconnectedTimeout;
    bool isKeepAlive = false;
};
}  // namespace Libruntime
}  // namespace YR
