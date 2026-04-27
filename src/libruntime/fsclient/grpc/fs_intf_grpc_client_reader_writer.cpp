/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.
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
#include <future>

#include "fs_intf_grpc_client_reader_writer.h"
#include "src/libruntime/utils/grpc_utils.h"
#include "src/utility/string_utility.h"

namespace YR::Libruntime {
constexpr unsigned int GRPC_RECONNECT_INTERVAL = 1;
const std::string FUNCTION_PROXY = "function-proxy";
using namespace std::chrono_literals;
using namespace runtime_rpc;
using grpc::internal::ClientReactor;

bool FSIntfGrpcClientReaderWriter::IsBatched()
{
    return batchStream_ != nullptr;
}

bool FSIntfGrpcClientReaderWriter::GrpcBatchRead(const std::shared_ptr<BatchStreamingMessage> &message)
{
    if (!FSIntfGrpcReaderWriter::Available()) {
        YRLOG_WARN("client is not connected");
        return false;
    }
    if (batchStream_) {
        return batchStream_->Read(message.get());
    }
    return false;
}

bool FSIntfGrpcClientReaderWriter::GrpcBatchWrite(const std::shared_ptr<BatchStreamingMessage> &request)
{
    if (!FSIntfGrpcReaderWriter::Available()) {
        YRLOG_WARN("client is not connected");
        return false;
    }
    if (batchStream_ == nullptr) {
        YRLOG_WARN("stream has reset nullptr");
        return false;
    }
    return batchStream_->Write(*request.get());
}

bool FSIntfGrpcClientReaderWriter::GrpcRead(const std::shared_ptr<StreamingMessage> &message)
{
    if (!FSIntfGrpcReaderWriter::Available()) {
        YRLOG_WARN("client is not connected");
        return false;
    }
    if (stream_) {
        return stream_->Read(message.get());
    }
    return false;
}

bool FSIntfGrpcClientReaderWriter::GrpcWrite(const std::shared_ptr<StreamingMessage> &request)
{
    if (!FSIntfGrpcReaderWriter::Available()) {
        YRLOG_WARN("client is not connected");
        return false;
    }
    if (stream_ == nullptr) {
        YRLOG_WARN("stream has reset nullptr");
        return false;
    }
    return stream_->Write(*request.get());
}

bool FSIntfGrpcClientReaderWriter::StreamEmpty()
{
    return stream_ == nullptr && batchStream_ == nullptr;
}

void FSIntfGrpcClientReaderWriter::WritesDone()
{
    if (stream_ != nullptr) {
        stream_->WritesDone();
        return;
    }
    if (batchStream_ != nullptr) {
        batchStream_->WritesDone();
        return;
    }
}

grpc::Status FSIntfGrpcClientReaderWriter::Finish()
{
    if (stream_ != nullptr) {
        return stream_->Finish();
    }
    if (batchStream_ != nullptr) {
        return batchStream_->Finish();
    }
    return grpc::Status(grpc::StatusCode::OK, "stream is null");
}

void FSIntfGrpcClientReaderWriter::Reset()
{
    stream_ = nullptr;
    batchStream_ = nullptr;
}

void FSIntfGrpcClientReaderWriter::ReconnectHandler()
{
    FSIntfGrpcReaderWriter::Stop();
    if (!StreamEmpty()) {
        WritesDone();
        auto status = Finish();
        YRLOG_INFO("grpc status code: {}, msg: {}", fmt::underlying(status.error_code()), status.error_message());
        // instance id not match
        if (status.error_code() == grpc::StatusCode::INVALID_ARGUMENT) {
            abnormal_.store(true);
            Reset();
            return;
        }
        if (status.error_code() == grpc::StatusCode::UNAUTHENTICATED) {
            if (discoverDriverCb) {
                discoverDriverCb();
            }
        }
        Reset();
    }
    if (Reconnect().OK() && resendCb != nullptr) {
        FSIntfGrpcReaderWriter::Init();
        resendCb(dstInstance);
    }
}

void FSIntfGrpcClientReaderWriter::ReceiveHandler()
{
    YRLOG_INFO("begin to receive msg from {}", dstInstance);
    while (!abnormal_) {
        if (isConnect_) {
            RecvFunc();
            isConnect_.store(false);
            disconnTime = std::chrono::steady_clock::now();
        }
        if (!abnormal_ &&
            std::chrono::steady_clock::now() - disconnTime < std::chrono::milliseconds(disconnectedTimeout)) {
            sleep(GRPC_RECONNECT_INTERVAL);
            ReconnectHandler();
        } else {
            if (!stopped && disconnectedCb != nullptr) {
                abnormal_.store(true);
                disconnectedCb(dstInstance);
            }
            break;
        }
    }
    YRLOG_INFO("end to receiver from {}", dstInstance);
}

ErrorInfo FSIntfGrpcClientReaderWriter::Start()
{
    if (isConnect_) {
        return {ErrorCode::ERR_INIT_CONNECTION_FAILED, "The client has been started."};
    }
    auto error = NewGrpcClientWithRetry();
    if (!error.OK() || StreamEmpty()) {
        std::stringstream ss;
        ss << "failed to establish grpc connection after " << RETRY_TIME << " trys. instanceID(" << dstInstance
           << "), code:(" << error.Code() << "), message(" << error.Msg() << "), stream null: " << (stream_ == nullptr);
        YRLOG_ERROR(ss.str());
        return StreamEmpty() ? ErrorInfo(ErrorCode::ERR_CONNECTION_FAILED, ss.str()) : error;
    }
    receiver_.Init(1, "grpc.receiver." + dstInstance);
    receiver_.Handle([this] { ReceiveHandler(); }, "grpc.client.receiver");
    return {};
}

ErrorInfo FSIntfGrpcClientReaderWriter::Reconnect()
{
    YRLOG_INFO("begin to reconnect {}, abnormal_ {}", dstInstance, abnormal_.load());
    clientsMgr->ReleaseFsConn(ip, port, dstInstance);
    return NewGrpcClientWithRetry(1);
}

void FSIntfGrpcClientReaderWriter::Stop()
{
    bool expected = false;
    bool exchanged = stopped.compare_exchange_strong(expected, true);
    if (!exchanged) {
        return;
    }
    YRLOG_DEBUG("begin to close connection of {}, ip={}, port={}", dstInstance, ip, port);
    abnormal_.store(true);
    // before finish, receiver_ & writer should be stopped to avoid race condition
    FSIntfGrpcReaderWriter::Stop();
    try {
        if (context != nullptr) {
            context->TryCancel();
        }
        receiver_.Shutdown();
        if (!StreamEmpty()) {
            WritesDone();
            (void)Finish();
            Reset();
        }
    } catch (std::exception &e) {
        YRLOG_ERROR("failed to finalize grpc stream, exception: {}", e.what());
    }
    auto err = clientsMgr->ReleaseFsConn(ip, port, dstInstance);
    if (!err.OK()) {
        YRLOG_ERROR("failed to release function system conn, code:({}), message({})", fmt::underlying(err.Code()),
                    err.Msg());
    }
    YRLOG_DEBUG("connection of {} closed, ip={}, port={}", dstInstance, ip, port);
}

ErrorInfo FSIntfGrpcClientReaderWriter::NewGrpcClientWithRetry(const int retryTimes)
{
    grpc::EnableDefaultHealthCheckService(true);
    auto [channel, error] = clientsMgr->GetFsConn(ip, port, dstInstance);
    if (!error.OK()) {
        YRLOG_ERROR(
            "failed to get grpc connection from fsconns to instance({}), "
            "exception({})",
            dstInstance, error.Msg());
        isConnect_.store(false);
        Reset();
        return error;
    }
    return BuildStreamWithRetry(channel, retryTimes);
}

ErrorInfo FSIntfGrpcClientReaderWriter::BuildStream(std::shared_ptr<grpc::Channel> channel)
{
    stub_ = runtime_rpc::RuntimeRPC::NewStub(channel);
    if (dstInstance == FUNCTION_PROXY || security->IsFsAuthEnable()) {
        stream_ = stub_->MessageStream(context.get());
        if (stream_ == nullptr) {
            return ErrorInfo(ErrorCode::ERR_CONNECTION_FAILED, "failed to build posix stream");
        }
    } else {
        batchStream_ = stub_->BatchMessageStream(context.get());
        if (batchStream_ == nullptr) {
            return ErrorInfo(ErrorCode::ERR_CONNECTION_FAILED, "failed to build posix stream");
        }
    }
    isConnect_.store(true);
    return {};
}


ErrorInfo FSIntfGrpcClientReaderWriter::BuildStreamWithRetry(std::shared_ptr<grpc::Channel> channel,
                                                             const int retryTimes)
{
    context = std::make_shared<::grpc::ClientContext>();
    context->AddMetadata(INSTANCE_ID_META, srcInstance);
    context->AddMetadata(RUNTIME_ID_META, runtimeID);
    SensitiveValue token;
    std::string ak;
    SensitiveValue sk;
    if (security != nullptr) {
        security->GetToken(token);
        security->GetAKSK(ak, sk);
    }
    YRLOG_INFO(
        "start build grpc stream, src instance is {}, dst instance is {}, token is empty: {}, ak is: {}, sk is empty: "
        "{}",
        srcInstance, dstInstance, token.Empty(), ak, sk.Empty());
    if (!token.Empty()) {
        context->AddMetadata(TOKEN_META, std::string(token.GetData(), token.GetSize()));
    }
    if (!ak.empty() && !sk.Empty()) {
        context->AddMetadata(TENANT_ACCESS_KEY, ak);
        auto signature = SignStreamingMessage(ak, sk);
        context->AddMetadata(SIGNATURE, signature);
        auto timestamp = GetCurrentUTCTime();
        context->AddMetadata(TIMESTAMP, timestamp);
        // If it is the runtime direct connection mode, set the tenant ID.
        if (dstInstance != FUNCTION_PROXY) {
            context->AddMetadata(TENANT_ID, Config::Instance().YR_TENANT_ID());
        }
    }
    context->AddMetadata(SOURCE_ID_META, srcInstance);
    context->AddMetadata(DST_ID_META, dstInstance);
    ErrorInfo err;
    for (int32_t retry = 0; retry < retryTimes; ++retry) {
        if (channel == nullptr) {
            auto ret = clientsMgr->NewFsConn(ip, port, security, dstInstance, isKeepAlive);
            if (!ret.second.OK()) {
                err = ret.second;
                YRLOG_ERROR("get new fs connection err, ip is {}, port is {}, err code is {}, err msg is {}", ip, port,
                            fmt::underlying(err.Code()), err.Msg());
                continue;
            }
            channel = ret.first;
        }
        try {
            err = BuildStream(channel);
            if (err.OK()) {
                return err;
            }
        } catch (std::exception &e) {
            YRLOG_ERROR(
                "failed to establish grpc connection to instance({}) ({}:{}), "
                "exception({})",
                dstInstance, ip, port, e.what());
            isConnect_.store(false);
            Reset();
            err = ErrorInfo(ErrorCode::ERR_CONNECTION_FAILED,
                            "failed to establish grpc connection between instance and LocalScheduler");
            break;
        }
    }
    if (channel != nullptr) {
        clientsMgr->ReleaseFsConn(ip, port, dstInstance);
    }
    if (!err.OK()) {
        isConnect_.store(false);
        Reset();
    }
    return err;
}

bool FSIntfGrpcClientReaderWriter::IsHealth()
{
    if (stopped) {
        YRLOG_WARN("function system client has been stopped, return false derectly");
        return false;
    }
    auto res = clientsMgr->GetFsConn(ip, port, dstInstance);
    if (!res.first) {
        YRLOG_WARN("there is no channel of dstInstance:{}, address is {}:{}, return false directly", dstInstance, ip,
                   port);
        return false;
    }
    clientsMgr->ReleaseFsConn(ip, port, dstInstance);
    if (res.first->GetState(false) != GRPC_CHANNEL_READY) {
        YRLOG_WARN("channel of dstInstance:{}, address is {}:{}, cheeck health failed, return false directly",
                   dstInstance, ip, port);
        return false;
    }
    return true;
}

bool FSIntfGrpcClientReaderWriter::IsSameDstAddr(const std::string &dstIp, const int &dstPort)
{
    if (this->ip == dstIp && this->port == dstPort) {
        return true;
    }
    return false;
}
}  // namespace YR::Libruntime
