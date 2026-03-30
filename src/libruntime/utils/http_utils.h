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
#include <sstream>
#include <string>
#include <unordered_map>
#include "datasystem/utils/sensitive_value.h"
#include "src/libruntime/utils/hash_utils.h"
#include "src/libruntime/utils/utils.h"
namespace YR {
namespace Libruntime {
const double TIMESTAMP_EXPIRE_DURATION_SECONDS = 60;
const std::string TRACE_ID_KEY_NEW = "X-Trace-Id";
const std::string AUTHORIZATION_KEY = "Authorization";
const std::string INSTANCE_CPU_KEY = "X-Instance-Cpu";
const std::string INSTANCE_MEMORY_KEY = "X-Instance-Memory";
const std::string TIMESTAMP_KEY = "X-Timestamp";
const std::string ACCESS_KEY = "X-Access-Key";
const std::string METHOD_GET = "GET";
const std::string X_SIGNATURE = "X-Signature";
const std::string X_SIGNED_HEADER = "X-Signed-Header";
const std::string X_AUTH = "X-Auth";
const std::string HEADER_CONNECTION = "connection";
const std::string HEADER_AUTHORIZATION = "authorization";
const std::string EMPTY_CONTENT_SHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";
const std::string HEX_STRING_SET_CAP = "0123456789ABCDEF";  // NOLINI
const int32_t FIRST_FOUR_BIT_MOVE = 4;
const std::string STR_WHITESPACE = " \t\n\r";
// Flags indicating how 'remove' or 'trim' should operate.
enum class Mode { PREFIX, SUFFIX, ANY };
using SensitiveValue = datasystem::SensitiveValue;

std::string HashToHex(const std::string &message);

std::string GenerateRequestDigest(const std::unordered_map<std::string, std::string> &headers, const std::string &body,
                                  const std::string &url);

void SignHttpRequest(const std::string &accessKey, const SensitiveValue &secretKey,
                     std::unordered_map<std::string, std::string> &headers, const std::string &body,
                     const std::string &url);

bool VerifyHttpRequest(const std::string &accessKey, const SensitiveValue &secretKey,
                       std::unordered_map<std::string, std::string> &headers, const std::string &body,
                       const std::string &url);

bool ShouldQueryEscape(char c);

std::string EscapeQuery(const std::string &value);

std::string DoReplace(std::string &source, const std::string &replace, const std::string &replacement);

std::string EscapeURL(const std::string &url, bool replacePath);

void ToLower(std::string &source);

std::string &Trim(std::string &from, Mode mode = Mode::ANY, const std::string &chars = STR_WHITESPACE);

std::string GetCanonicalQueries(const std::shared_ptr<std::map<std::string, std::string>> &queries);

std::string GetCanonicalHeaders(const std::map<std::string, std::string> &headers);

std::string GetSignedHeaders(const std::map<std::string, std::string> &headers);

std::string GetCanonicalRequest(const std::string &method, const std::string &path,
                                const std::shared_ptr<std::map<std::string, std::string>> &queries,
                                const std::map<std::string, std::string> &headers, const std::string &sha256);

std::map<std::string, std::string> UnorderedMapToMapSorted(const std::unordered_map<std::string, std::string> &um);

std::string HashToHex2(const std::string &message);

void SignHttpRequest2(std::shared_ptr<Security> security, const std::string &method,
                      const std::string &path, std::unordered_map<std::string, std::string> &headers,
                      const std::string &body);
}  // namespace Libruntime
}  // namespace YR