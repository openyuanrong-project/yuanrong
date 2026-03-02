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

#include <memory>

#include "src/libruntime/err_type.h"
#include "src/libruntime/fsclient/protobuf/runtime_rpc.grpc.pb.h"
#include "src/libruntime/utils/security.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
using StreamingMessage = ::runtime_rpc::StreamingMessage;
using BatchStreamingMessage = ::runtime_rpc::BatchStreamingMessage;
using BodyCase = ::runtime_rpc::StreamingMessage::BodyCase;
using MsgHdlr = std::function<void(const std::string &from, const std::shared_ptr<StreamingMessage> &)>;
using DiscoverDriverCb = std::function<ErrorInfo()>;
struct ReaderWriterClientOption {
    std::string ip;
    int port;
    int disconnectedTimeout;
    std::shared_ptr<Security> security;
    std::function<void(const std::string &)> resendCb;
    std::function<void(const std::string &)> disconnectedCb;
    bool isKeepAlive = true;
};
class FSIntfReaderWriter {
public:
    FSIntfReaderWriter(const std::string &srcInstance, const std::string &dstInstance, const std::string &runtimeID)
        : srcInstance(srcInstance), dstInstance(dstInstance), runtimeID(runtimeID){};
    virtual ~FSIntfReaderWriter() = default;
    virtual ErrorInfo Start() = 0;
    virtual void Stop() = 0;
    virtual bool Available() = 0;
    virtual bool Abnormal() = 0;
    virtual void Write(const std::shared_ptr<StreamingMessage> &msg,
                       std::function<void(bool, ErrorInfo)> callback = nullptr,
                       std::function<void(bool)> preWrite = nullptr) = 0;
    virtual bool IsHealth() = 0;
    virtual bool IsSameDstAddr(const std::string &dstIp, const int &dstPort) = 0;
    void SetDiscoverDriverCb(const DiscoverDriverCb &cb)
    {
        discoverDriverCb = cb;
    }
    inline void RegisterMessageHandler(const std::unordered_map<BodyCase, MsgHdlr> &hdlrs)
    {
        msgHdlrs = hdlrs;
    }

    void HandleRequest(const std::shared_ptr<StreamingMessage> &message)
    {
        auto it = msgHdlrs.find(message->body_case());
        if (it != msgHdlrs.end()) {
            it->second(dstInstance, message);
        } else {
            YRLOG_ERROR("Invalid received message body type {} from {}", fmt::underlying(message->body_case()),
                        dstInstance);
        }
    }

protected:
    std::string srcInstance;
    std::string dstInstance;
    std::string runtimeID;
    std::unordered_map<BodyCase, MsgHdlr> msgHdlrs;
    DiscoverDriverCb discoverDriverCb = nullptr;
};
}  // namespace Libruntime
}  // namespace YR