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

#pragma once
#include "datasystem/utils/sensitive_value.h"
#include "src/libruntime/fsclient/protobuf/runtime_rpc.grpc.pb.h"

namespace YR {
namespace Libruntime {
using SensitiveValue = datasystem::SensitiveValue;
std::pair<std::string, bool> SerializeBodyToString(const ::runtime_rpc::StreamingMessage &message);

bool SignStreamingMessage(const std::string &accessKey, const SensitiveValue &secretKey,
                          ::runtime_rpc::StreamingMessage &message);

std::string SignStreamingMessage(const std::string &accessKey, const SensitiveValue &secretKey);

bool VerifyStreamingMessage(const std::string &accessKey, const SensitiveValue &secretKey,
                            const ::runtime_rpc::StreamingMessage &message);

std::string SignTimestamp(const std::string &accessKey, const SensitiveValue &secretKey, const std::string &timestamp);

bool VerifyTimestamp(const std::string &accessKey, const SensitiveValue &secretKey, const std::string &timestamp,
                     const std::string &signature);

}  // namespace Libruntime
}  // namespace YR