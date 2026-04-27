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

#include <sstream>

#include "src/libruntime/utils/utils.h"
#include "src/proto/socket.pb.h"
#include "src/utility/logger/logger.h"

#include <google/protobuf/descriptor.h>
#include <google/protobuf/io/coded_stream.h>
#include <google/protobuf/io/zero_copy_stream_impl.h>
#include <google/protobuf/io/zero_copy_stream_impl_lite.h>
#include <google/protobuf/message.h>

namespace YR {
namespace Libruntime {
namespace ProtoIO = google::protobuf::io;
using SocketMessage = ::libruntime::SocketMessage;
using BusinessMessage = ::libruntime::BusinessMessage;
using MessageType = ::libruntime::MessageType;
using FunctionLog = ::libruntime::FunctionLog;

const char MAGIC_NUMBER = 0x01;
const char MESSAGE_REQUEST_BYTE = 0x00;
const char X_VERSION = 0x01;
const int BYTES_SIZE = 4;

class MessageCoder {
public:
    MessageCoder();
    std::shared_ptr<SocketMessage> GenerateSocketMsg(char magicNumber, char version, char packetType,
                                                     const std::string &packetId, FunctionLog &functionLog);
    std::string Encode(std::shared_ptr<SocketMessage> socketMsg);
    std::shared_ptr<SocketMessage> Decode(int fd, google::protobuf::uint32 socketMsgSize);
    google::protobuf::uint32 DecodeMsgSize(char *buf);
};
}  // namespace Libruntime
}  // namespace YR
