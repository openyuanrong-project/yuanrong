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

#include <condition_variable>
#include <memory>
#include <mutex>
#include <queue>
#include <thread>

#include "src/libruntime/fsclient/fs_intf_reader_writer.h"

#include "src/utility/thread_pool.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/fsclient/protobuf/runtime_rpc.grpc.pb.h"

namespace YR {
namespace Libruntime {
using YR::utility::ThreadPool;
using StreamingMessage = ::runtime_rpc::StreamingMessage;

class FSIntfGrpcReaderWriter : public FSIntfReaderWriter {
public:
    FSIntfGrpcReaderWriter(const std::string &srcInstance, const std::string &dstInstance, const std::string &runtimeID)
        : FSIntfReaderWriter(srcInstance, dstInstance, runtimeID)
    {
        Init();
    };
    ~FSIntfGrpcReaderWriter() override
    {
        Stop();
    }
    virtual void PreStart() = 0;
    virtual ErrorInfo Start() = 0;
    void Init();
    virtual bool IsHealth() = 0;
    virtual bool IsSameDstAddr(const std::string &dstIp, const int &dstPort) = 0;
    void Stop() override;
    bool Available() override;
    bool Abnormal() override;
    void Write(const std::shared_ptr<StreamingMessage> &msg, std::function<void(bool, ErrorInfo)> callback = nullptr,
               std::function<void(bool)> preWrite = nullptr) override;
    void RecvFunc();

protected:
    struct MessageInfo {
        std::shared_ptr<StreamingMessage> msg;
        std::function<void(bool, ErrorInfo)> callback;
    };
    void Run();
    std::shared_ptr<StreamingMessage> TransDirectInvokeRequest(const std::shared_ptr<StreamingMessage> &);
    std::shared_ptr<StreamingMessage> TransDirectCallResponse(const std::shared_ptr<StreamingMessage> &);
    std::shared_ptr<StreamingMessage> TransDirectCallResult(const std::shared_ptr<StreamingMessage> &);
    std::shared_ptr<StreamingMessage> TransDirectNotifyResponse(const std::shared_ptr<StreamingMessage> &);
    typedef std::shared_ptr<StreamingMessage> (FSIntfGrpcReaderWriter::*TransHdlr)(
        const std::shared_ptr<StreamingMessage> &);
    std::unordered_map<BodyCase, TransHdlr> transHdlrs_ = {
        {StreamingMessage::kInvokeReq, &FSIntfGrpcReaderWriter::TransDirectInvokeRequest},
        {StreamingMessage::kCallRsp, &FSIntfGrpcReaderWriter::TransDirectCallResponse},
        {StreamingMessage::kCallResultReq, &FSIntfGrpcReaderWriter::TransDirectCallResult},
        {StreamingMessage::kNotifyRsp, &FSIntfGrpcReaderWriter::TransDirectNotifyResponse}};
    void TransDirectSendMsg(std::shared_ptr<StreamingMessage> &);
    void SingleRecv();
    void BatchRecv();
    void SingleWrite(MessageInfo &msgInfo);
    void BatchWrite(std::queue<MessageInfo> &msgInfos);
    virtual bool GrpcRead(const std::shared_ptr<StreamingMessage> &message) = 0;
    virtual bool GrpcWrite(const std::shared_ptr<StreamingMessage> &request) = 0;
    virtual bool GrpcBatchRead(const std::shared_ptr<BatchStreamingMessage> &message) = 0;
    virtual bool GrpcBatchWrite(const std::shared_ptr<BatchStreamingMessage> &request) = 0;
    virtual bool IsBatched() = 0;
    std::atomic<bool> isConnect_ = {false};
    std::atomic<bool> abnormal_ = {false};
    std::atomic<bool> stop_ = {false};
    std::thread worker_;
    std::mutex mux_;
    std::condition_variable cv_;
    std::queue<MessageInfo> msgQueue_;
    std::atomic<bool> isDirectConnection_;
};
}  // namespace Libruntime
}  // namespace YR