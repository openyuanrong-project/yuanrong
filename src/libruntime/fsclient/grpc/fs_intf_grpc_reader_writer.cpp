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

#include "fs_intf_grpc_reader_writer.h"

#include "src/dto/config.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
const uint32_t SIZE_MEGA_BYTES = 1024 * 1024;
const std::string FUNCTION_PROXY = "function-proxy";
bool FSIntfGrpcReaderWriter::Available()
{
    return this->isConnect_ && !this->abnormal_;
}

bool FSIntfGrpcReaderWriter::Abnormal()
{
    return this->abnormal_;
}

void FSIntfGrpcReaderWriter::Write(const std::shared_ptr<StreamingMessage> &msg,
                                   std::function<void(bool, ErrorInfo)> callback, std::function<void(bool)> preWrite)
{
    if (msg == nullptr) {
        YRLOG_ERROR("Invalid GRPC message pointer.");
        return;
    }
    if (preWrite != nullptr) {
        preWrite(isDirectConnection_);
    }
    MessageInfo msgInfo{msg, callback};
    {
        std::unique_lock<std::mutex> lk(mux_);
        if (YR_UNLIKELY(stop_)) {
            auto err = ErrorInfo(ERR_INNER_COMMUNICATION, "posix stream is stopped.");
            if (callback != nullptr) {
                callback(isDirectConnection_, err);
            }
            return;
        }
        msgQueue_.push(msgInfo);
    }
    cv_.notify_all();
}

void FSIntfGrpcReaderWriter::SingleWrite(MessageInfo &msgInfo)
{
    auto msg = std::move(msgInfo.msg);
    auto callback = std::move(msgInfo.callback);
    ErrorInfo err;
    uint32_t maxGrpcSize = Config::Instance().YR_MAX_GRPC_SIZE() * SIZE_MEGA_BYTES;
    if (msg->ByteSizeLong() > maxGrpcSize) {
        auto message = "Failed to send GRPC message (message ID: " + msg->messageid() +
                       "), the message size (" + std::to_string(msg->ByteSizeLong()) +
                       " bytes) exceeds the limit(" + std::to_string(maxGrpcSize) + " bytes)";
        YRLOG_ERROR(message);
        // This code should be neither
        // 'ERR_INNER_COMMUNICATION' nor 'ERR_REQUEST_BETWEEN_RUNTIME_BUS',
        // otherwise retry of sending would be performed.
        err.SetErrorCode(ERR_PARAM_INVALID);
        err.SetErrorMsg(message);
    } else if (!this->Available()) {
        err.SetErrorCode(ERR_INNER_COMMUNICATION);
        err.SetErrorMsg("Function system client is unavailable.");
    } else if (!GrpcWrite(msg)) {
        YRLOG_ERROR("Stream write message, message ID: {}", msg->messageid());
        err.SetErrorCode(ERR_INNER_COMMUNICATION);
        err.SetErrorMsg("Function system client rpc error.");
    }

    if (callback != nullptr) {
        callback(isDirectConnection_, err);
    }
}


void FSIntfGrpcReaderWriter::BatchWrite(std::queue<FSIntfGrpcReaderWriter::MessageInfo> &msgInfos)
{
    auto batchs = std::make_shared<BatchStreamingMessage>();
    std::vector<std::function<void(bool, ErrorInfo)>> callbacks;
    static uint32_t maxGrpcSize = Config::Instance().YR_MAX_GRPC_SIZE() * SIZE_MEGA_BYTES;
    size_t totalSize = 0;
    while (!msgInfos.empty()) {
        auto &msg = msgInfos.front();
        auto size = msg.msg->ByteSizeLong();
        if (size > maxGrpcSize) {
            auto message = "Failed to send GRPC message (message ID: " + msg.msg->messageid() +
                           "), the message size (" + std::to_string(size) + " bytes) exceeds the limit(" +
                           std::to_string(maxGrpcSize) + " bytes)";
            YRLOG_ERROR(message);
            ErrorInfo err;
            err.SetErrorCode(ERR_PARAM_INVALID);
            err.SetErrorMsg(message);
            msg.callback(isDirectConnection_, err);
            msgInfos.pop();
            continue;
        }
        totalSize += size;
        if (totalSize > maxGrpcSize) {
            break;
        }
        *batchs->add_messages() = *msg.msg;
        callbacks.emplace_back(std::move(msg.callback));
        msgInfos.pop();
    }
    ErrorInfo err;
    if (!this->Available()) {
        err.SetErrorCode(ERR_INNER_COMMUNICATION);
        err.SetErrorMsg("client is unavailable.");
    } else if (!GrpcBatchWrite(batchs)) {
        err.SetErrorCode(ERR_INNER_COMMUNICATION);
        err.SetErrorMsg("client rpc error.");
    }
    for (auto &callback : callbacks) {
        if (callback == nullptr) {
            continue;
        }
        callback(isDirectConnection_, err);
    }
}

void FSIntfGrpcReaderWriter::Run()
{
    while (!stop_) {
        auto msgInfos = std::queue<MessageInfo>();
        {
            std::unique_lock<std::mutex> lk(mux_);
            cv_.wait(lk, [this] { return msgQueue_.size() != 0 || stop_; });
            if (stop_) {
                break;
            }
            msgInfos.swap(msgQueue_);
        }
        while (!msgInfos.empty()) {
            if (IsBatched()) {
                BatchWrite(msgInfos);
            } else {
                auto &msgInfo = msgInfos.front();
                SingleWrite(msgInfo);
                msgInfos.pop();
            }
        }
    }
}

void FSIntfGrpcReaderWriter::RecvFunc()
{
    if (IsBatched()) {
        return BatchRecv();
    }
    return SingleRecv();
}

void FSIntfGrpcReaderWriter::SingleRecv()
{
    for (; Available();) {
        auto message = std::make_shared<StreamingMessage>();
        if (!GrpcRead(message)) {
            YRLOG_INFO("Read failed from {}", dstInstance);
            break;
        }
        if (!Available()) {
            YRLOG_INFO("{} Not available", dstInstance);
            break;
        }
        TransDirectSendMsg(message);
        HandleRequest(message);
    }
}

void FSIntfGrpcReaderWriter::BatchRecv()
{
    for (; Available();) {
        auto message = std::make_shared<BatchStreamingMessage>();
        if (!GrpcBatchRead(message)) {
            YRLOG_INFO("Read failed from {}", dstInstance);
            break;
        }
        if (!Available()) {
            YRLOG_INFO("{} Not available", dstInstance);
            break;
        }
        for (auto &msg : *message->mutable_messages()) {
            auto ptr = std::make_shared<StreamingMessage>(std::move(msg));
            TransDirectSendMsg(ptr);
            HandleRequest(ptr);
        }
    }
}

void FSIntfGrpcReaderWriter::TransDirectSendMsg(std::shared_ptr<StreamingMessage> &message)
{
    auto it = transHdlrs_.find(message->body_case());
    if (it != transHdlrs_.end()) {
        message = (this->*(it->second))(message);
    }
}

void FSIntfGrpcReaderWriter::Init()
{
    isDirectConnection_ = (dstInstance != FUNCTION_PROXY);
    stop_ = false;
    worker_ = std::thread([this] { Run(); });
    std::string name = "writer." + dstInstance.substr(dstInstance.size() - 6, 6);
    pthread_setname_np(worker_.native_handle(), name.c_str());
}

void FSIntfGrpcReaderWriter::Stop()
{
    std::queue<MessageInfo> cache;
    {
        std::unique_lock<std::mutex> lock(mux_);
        if (stop_) {
            return;
        }
        if (!msgQueue_.empty()) {
            msgQueue_.swap(cache);
        }
        stop_ = true;
    }
    cv_.notify_all();
    if (worker_.joinable()) {
        worker_.join();
    }
    YRLOG_DEBUG("{} writer thread is already stopped, unprocessed {}", dstInstance, cache.size());
    ErrorInfo err(ERR_INNER_COMMUNICATION, "posix stream is closed");
    // to avoid request not processed
    while (cache.size() != 0) {
        if (cache.front().callback != nullptr) {
            cache.front().callback(isDirectConnection_, err);
        }
        cache.pop();
    }
}

std::shared_ptr<StreamingMessage> FSIntfGrpcReaderWriter::TransDirectInvokeRequest(
    const std::shared_ptr<StreamingMessage> &message)
{
    auto msg = std::make_shared<StreamingMessage>();
    auto callreq = msg->mutable_callreq();
    auto invokereq = message->mutable_invokereq();
    msg->set_messageid(std::move(*message->mutable_messageid()));
    callreq->set_function(std::move(*invokereq->mutable_function()));
    callreq->set_traceid(std::move(*invokereq->mutable_traceid()));
    callreq->set_requestid(std::move(*invokereq->mutable_requestid()));
    callreq->set_iscreate(false);
    *callreq->mutable_args() = std::move(*invokereq->mutable_args());
    *callreq->mutable_createoptions() = std::move(*invokereq->mutable_invokeoptions()->mutable_customtag());
    *callreq->mutable_returnobjectids() = std::move(*invokereq->mutable_returnobjectids());
    callreq->set_senderid(dstInstance);
    return msg;
}

std::shared_ptr<StreamingMessage> FSIntfGrpcReaderWriter::TransDirectCallResponse(
    const std::shared_ptr<StreamingMessage> &message)
{
    auto msg = std::make_shared<StreamingMessage>();
    auto invokersp = msg->mutable_invokersp();
    auto callrsp = message->mutable_callrsp();
    msg->set_messageid(std::move(*message->mutable_messageid()));
    invokersp->set_code(callrsp->code());
    invokersp->set_message(std::move(*callrsp->mutable_message()));
    return msg;
}

std::shared_ptr<StreamingMessage> FSIntfGrpcReaderWriter::TransDirectCallResult(
    const std::shared_ptr<StreamingMessage> &message)
{
    auto msg = std::make_shared<StreamingMessage>();
    auto notifyreq = msg->mutable_notifyreq();
    auto callresult = message->mutable_callresultreq();
    msg->set_messageid(std::move(*message->mutable_messageid()));
    notifyreq->set_code(callresult->code());
    notifyreq->set_message(std::move(*callresult->mutable_message()));
    *notifyreq->mutable_smallobjects() = std::move(*callresult->mutable_smallobjects());
    *notifyreq->mutable_stacktraceinfos() = std::move(*callresult->mutable_stacktraceinfos());
    *notifyreq->mutable_requestid() = std::move(*callresult->mutable_requestid());
    // runtime info is removed to avoid duplicated build stream.
    return msg;
}

std::shared_ptr<StreamingMessage> FSIntfGrpcReaderWriter::TransDirectNotifyResponse(
    const std::shared_ptr<StreamingMessage> &message)
{
    auto msg = std::make_shared<StreamingMessage>();
    auto callresultack = msg->mutable_callresultack();
    msg->set_messageid(std::move(*message->mutable_messageid()));
    // directly call would never to got err ack from remote
    callresultack->set_code(common::ErrorCode::ERR_NONE);
    return msg;
}
}  // namespace Libruntime
}  // namespace YR