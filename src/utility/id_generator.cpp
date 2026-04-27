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

#include "id_generator.h"

#include <sstream>

namespace YR {
namespace utility {

const std::string IDGenerator::appIdPrefix_ = "job-";
const std::string IDGenerator::traceIdPrefix_ = "job-";
const std::string IDGenerator::traceIdSuffix_ = "-trace-X";

std::string IDGenerator::GenApplicationId()
{
    return appIdPrefix_ + GenerateUUID(appIdLen_ / DOUBLE);
}

std::string IDGenerator::GenTraceId()
{
    auto id = GenerateUUID(traceIdLen_ / DOUBLE);
    std::string s;
    s.reserve(traceIdPrefix_.size() + traceIdLen_ + traceIdPrefix_.size());
    s.append(traceIdPrefix_);
    s.append(id);
    s.append(traceIdSuffix_);
    return s;
}

std::string IDGenerator::GenTraceId(const std::string &appId)
{
    std::string s;
    s.reserve(traceIdPrefix_.size() + traceIdLen_ + traceIdPrefix_.size());
    s.append(traceIdPrefix_);
    s.append(appId.substr(appIdPrefix_.size()));
    s.append(traceIdSuffix_);
    return s;
}

std::string IDGenerator::GenRequestId(uint8_t index)
{
    auto id = GenerateUUID((requestIdLen_ - DEFAULT_SUBSTR_LENTH) / DOUBLE);
    if (index == 0) {
        return id + RAW_SEQ_STR;
    }
    auto seqStr = Uint8ToHex(index);
    return id + seqStr;
}

std::string IDGenerator::GenRequestId(const std::string &requestId, uint8_t index)
{
    auto seqStr = Uint8ToHex(index);
    return requestId.substr(0, requestIdLen_ - DEFAULT_SUBSTR_LENTH) + seqStr;
}

// remove @initcall
std::string IDGenerator::GetRealRequestId(const std::string &requestId)
{
    return ParseRealRequestId(requestId);
}

// reset suffix with 00
std::pair<std::string, uint8_t> IDGenerator::DecodeRawRequestId(const std::string &requestId)
{
    if (requestId.empty()) {
        return std::make_pair(requestNilId, 0);
    }
    std::string realReqId;
    if (requestId.find('@') == std::string::npos && requestId.find('-') == std::string::npos) {
        auto id = requestId.substr(0, requestIdLen_ - DEFAULT_SUBSTR_LENTH);
        auto seqStr = requestId.substr(requestIdLen_ - DEFAULT_SUBSTR_LENTH);
        uint8_t seq = HexToUint8(seqStr);
        return std::make_pair(id + RAW_SEQ_STR, seq);
    }
    realReqId = ParseRealRequestId(requestId);
    auto id = realReqId.substr(0, requestIdLen_ - DEFAULT_SUBSTR_LENTH);
    auto seqStr = realReqId.substr(requestIdLen_ - DEFAULT_SUBSTR_LENTH);
    uint8_t seq = HexToUint8(seqStr);
    return std::make_pair(id + RAW_SEQ_STR, seq);
}

std::string IDGenerator::GetRequestIdFromMsg(const std::string &messageId)
{
    return messageId.substr(0, requestIdLen_);
}

std::string IDGenerator::GetRequestIdFromObj(const std::string &objectId)
{
    return objectId.substr(0, requestIdLen_);
}

std::string IDGenerator::GenPacketId()
{
    return GenerateUUID(packIdLen_ / DOUBLE);
}

std::string IDGenerator::GenMessageId(const std::string &requestId, uint8_t index)
{
    auto seqStr = Uint8ToHex(index);
    return requestId + seqStr;
}

std::string IDGenerator::GenObjectId(std::function<void(std::string &, const std::string &)> generateKey)
{
    std::string dsObjId;
    auto objId = GenerateUUID(objIdLen_ / DOUBLE);
    if (generateKey) {
        generateKey(dsObjId, objId);
        return dsObjId;
    }
    return objId;
}

std::string IDGenerator::GenObjectId(const std::string &requestId, uint8_t index)
{
    auto seqStr = Uint8ToHex(index);
    return requestId + seqStr;
}

std::string IDGenerator::GenGroupId(const std::string &appId)
{
    return appId.substr(appIdPrefix_.size()) + GenerateUUID(groupIdLen_ / DOUBLE);
}

std::string IDGenerator::GenerateUUID(size_t size)
{
    std::string hex;
    hex.reserve(size * DOUBLE);
    thread_local struct {
        absl::BitGen gen;
        pid_t pid;
    } tls{absl::BitGen(), getpid()};
    if (tls.pid != getpid()) {
        tls.gen = absl::BitGen();
        tls.pid = getpid();
    }
    auto &gen = tls.gen;
    for (size_t i = 0; i < size; i++) {
        uint8_t value = absl::uniform_int_distribution<uint8_t>(0, std::numeric_limits<uint8_t>::max())(gen);
        hex.push_back(hexDigits_[value >> 0x04]);
        hex.push_back(hexDigits_[value & 0x0F]);
    }
    return hex;
}

std::string IDGenerator::Uint8ToHex(uint8_t value)
{
    std::string hex;
    hex.reserve(DOUBLE);
    hex.push_back(hexDigits_[value >> 0x04]);
    hex.push_back(hexDigits_[value & 0x0F]);
    return hex;
}

uint8_t IDGenerator::HexToUint8(const std::string &hex)
{
    if (hex == RAW_SEQ_STR) {
        return 0;
    }
    return (HexCharToUint8(hex[0]) << 0x04) | HexCharToUint8(hex[1]);
}

uint8_t IDGenerator::HexCharToUint8(char c)
{
    if (c >= '0' && c <= '9') {
        return c - '0';
    }

    if (c >= 'A' && c <= 'F') {
        return c - 'A' + HEX_OFFSET;
    }

    if (c >= 'a' && c <= 'f') {
        return c - 'a' + HEX_OFFSET;
    }
    return 0;
}

}  // namespace utility
}  // namespace YR