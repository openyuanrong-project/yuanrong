/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

#include "grpc_utils.h"
#include "src/libruntime/utils/hash_utils.h"
#include "src/libruntime/utils/utils.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
const double TIMESTAMP_EXPIRE_DURATION_SECONDS = 60;
std::pair<std::string, bool> SerializeBodyToString(const ::runtime_rpc::StreamingMessage &message)
{
    const google::protobuf::Descriptor *descriptor = message.GetDescriptor();
    const google::protobuf::Reflection *reflection = message.GetReflection();
    const google::protobuf::OneofDescriptor *oneofDescriptor = descriptor->FindOneofByName("body");
    const google::protobuf::FieldDescriptor *fieldDescriptor =
        reflection->GetOneofFieldDescriptor(message, oneofDescriptor);
    if (fieldDescriptor) {
        const google::protobuf::Message &bodyField = reflection->GetMessage(message, fieldDescriptor);
        std::stringstream ss;
        SHA256AndHex(bodyField.DebugString(), ss);
        return std::make_pair<std::string, bool>(ss.str(), true);
    } else {
        YRLOG_ERROR("failed to get body string of message: {}", message.messageid());
        return std::make_pair<std::string, bool>("", false);
    }
}

bool SignStreamingMessage(const std::string &accessKey, const SensitiveValue &secretKey,
                          ::runtime_rpc::StreamingMessage &message)
{
    auto timestamp = GetCurrentUTCTime();
    auto [body, serializeSuccess] = SerializeBodyToString(message);
    if (body.empty() && !serializeSuccess) {
        YRLOG_ERROR("body is empty and serialize failed, message is {}", message.DebugString());
        return false;
    }
    std::string signKey = accessKey + ":" + timestamp + ":" + body;
    auto signature = GetHMACSha256(secretKey, signKey);

    (*message.mutable_metadata())["access_key"] = accessKey;
    (*message.mutable_metadata())["signature"] = signature;
    (*message.mutable_metadata())["timestamp"] = timestamp;
    return true;
}

std::string SignStreamingMessage(const std::string &accessKey, const SensitiveValue &secretKey)
{
    auto timestamp = GetCurrentUTCTime();
    std::string signKey = accessKey + ":" + timestamp;
    auto signature = GetHMACSha256(secretKey, signKey);
    return signature;
}

bool VerifyStreamingMessage(const std::string &accessKey, const SensitiveValue &secretKey,
                            const ::runtime_rpc::StreamingMessage &message)
{
    auto timestamp = message.metadata().find("timestamp");
    if (timestamp == message.metadata().end() || timestamp->second.empty()) {
        YRLOG_ERROR("failed to verify message: {}, failed to find timestamp in meta-data", message.DebugString());
        return false;
    }

    auto currentTimestamp = GetCurrentUTCTime();
    if (IsLaterThan(currentTimestamp, timestamp->second, TIMESTAMP_EXPIRE_DURATION_SECONDS)) {
        YRLOG_ERROR("failed to verify message: {}, failed to verify timestamp, difference is more than 1 min {} vs {}",
                    message.messageid(), currentTimestamp, timestamp->second);
        return false;
    }

    auto signature = message.metadata().find("signature");
    if (signature == message.metadata().end() || signature->second.empty()) {
        YRLOG_ERROR("failed to verify message: {}, failed to find signature in meta-data", message.DebugString());
        return false;
    }

    std::string signKey = accessKey + ":" + timestamp->second + ":" + SerializeBodyToString(message).first;
    if (GetHMACSha256(secretKey, signKey) != signature->second) {
        YRLOG_ERROR("failed to verify message: {},", message.DebugString());
        return false;
    }
    return true;
}

std::string SignTimestamp(const std::string &accessKey, const SensitiveValue &secretKey, const std::string &timestamp)
{
    return GetHMACSha256(secretKey, accessKey + ":" + timestamp);
}

bool VerifyTimestamp(const std::string &accessKey, const SensitiveValue &secretKey, const std::string &timestamp,
                     const std::string &signature)
{
    auto currentTimestamp = GetCurrentUTCTime();
    if (IsLaterThan(currentTimestamp, timestamp, TIMESTAMP_EXPIRE_DURATION_SECONDS)) {
        YRLOG_ERROR("failed to verify timestamp, difference is more than 1 min {} vs {}", currentTimestamp, timestamp);
        return false;
    }

    std::string signKey = accessKey + ":" + timestamp;
    if (GetHMACSha256(secretKey, signKey) != signature) {
        YRLOG_ERROR("failed to verify timestamp, signature isn't the same");
        return false;
    }
    return true;
}

}  // namespace Libruntime
}  // namespace YR