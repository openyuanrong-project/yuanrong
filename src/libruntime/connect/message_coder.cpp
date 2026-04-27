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

#include <sstream>

#include "message_coder.h"

namespace YR {
namespace Libruntime {
MessageCoder::MessageCoder() {}

std::shared_ptr<SocketMessage> MessageCoder::GenerateSocketMsg(char magicNumber, char version, char packetType,
                                                               const std::string &packetId, FunctionLog &functionLog)
{
    std::shared_ptr<SocketMessage> socketMsg = std::make_shared<SocketMessage>();
    socketMsg->set_magicnumber(std::string(1, magicNumber));
    socketMsg->set_version(std::string(1, version));
    socketMsg->set_packettype(std::string(1, packetType));
    socketMsg->set_packetid(packetId);
    BusinessMessage *msg = socketMsg->mutable_businessmsg();
    msg->set_type(MessageType::LogProcess);
    auto *funcLog = msg->mutable_functionlog();
    *funcLog = functionLog;
    return socketMsg;
}

std::string MessageCoder::Encode(std::shared_ptr<SocketMessage> socketMsg)
{
    uint32_t size = socketMsg->ByteSize() + BYTES_SIZE;
    auto pkt = std::make_unique<char[]>(size);
    ProtoIO::ArrayOutputStream aos(pkt.get(), size);
    std::unique_ptr<ProtoIO::CodedOutputStream> codedOutput = std::make_unique<ProtoIO::CodedOutputStream>(&aos);
    codedOutput->WriteVarint32(socketMsg->ByteSize());
    socketMsg->SerializeToCodedStream(codedOutput.get());
    return std::string(pkt.get(), size);
}

std::shared_ptr<SocketMessage> MessageCoder::Decode(int fd, google::protobuf::uint32 socketMsgSize)
{
    int byteCount;
    auto buffer = std::make_unique<char[]>(socketMsgSize + BYTES_SIZE);
    if ((byteCount = recv(fd, (void *)buffer.get(), BYTES_SIZE + socketMsgSize, MSG_WAITALL)) == -1) {
        YRLOG_ERROR("Failed to receive data, code:{}", errno);
    }
    YRLOG_DEBUG("Size and msg byte count is {}", byteCount);

    ProtoIO::ArrayInputStream arrayInputStream(buffer.get(), socketMsgSize + BYTES_SIZE);
    ProtoIO::CodedInputStream codedInputStream(&arrayInputStream);

    codedInputStream.ReadVarint32(&socketMsgSize);
    YRLOG_DEBUG("SocketMsgLen: {}", socketMsgSize);

    ProtoIO::CodedInputStream::Limit msgLimit = codedInputStream.PushLimit(socketMsgSize);
    auto socketMsg = std::make_shared<SocketMessage>();
    socketMsg->ParseFromCodedStream(&codedInputStream);
    codedInputStream.PopLimit(msgLimit);
    YRLOG_DEBUG("SocketMsg: {}", socketMsg->DebugString());
    return socketMsg;
}

google::protobuf::uint32 MessageCoder::DecodeMsgSize(char *buf)
{
    google::protobuf::uint32 socketMsgSize;
    ProtoIO::ArrayInputStream arrayInputStream(buf, BYTES_SIZE);
    ProtoIO::CodedInputStream codedInputStream(&arrayInputStream);
    codedInputStream.ReadVarint32(&socketMsgSize);
    return socketMsgSize;
}
}  // namespace Libruntime
}  // namespace YR
