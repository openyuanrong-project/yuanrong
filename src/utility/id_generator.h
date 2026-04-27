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
#include <cstddef>
#include <cstdint>
#include <functional>
#include <string>
#include <vector>
#include "absl/random/random.h"
#include "src/utility/string_utility.h"
namespace YR {
namespace utility {
const int DOUBLE = 2;
const int DEFAULT_SUBSTR_LENTH = 2;
const uint8_t HEX_OFFSET = 10;
const std::string RAW_SEQ_STR = "00";
const std::string requestNilId = "ffffffffffffffff00";

inline std::string ParseRealRequestId(const std::string &reqIdStr)
{
    std::vector<std::string> result;
    Split(reqIdStr, result, '@');
    std::vector<std::string> tmpResult;
    if (result.empty()) {
        return requestNilId;
    }
    Split(result[0], tmpResult, '-');
    return tmpResult[0];
}

inline std::string ParseRealJobId(const std::string &hexString)
{
    std::vector<std::string> result;
    Split(hexString, result, '-');
    if (result.size() > 1) {
        return result[1];
    }
    return result[0];
}

class IDGenerator {
public:
    IDGenerator() = delete;

    static std::string GenApplicationId();
    static std::string GenTraceId();
    static std::string GenTraceId(const std::string &appId);
    static std::string GenRequestId(uint8_t index = 0);
    static std::string GenRequestId(const std::string &requestId, uint8_t index = 0);
    // remove @initcall
    static std::string GetRealRequestId(const std::string &requestId);
    // reset suffix with 00
    static std::pair<std::string, uint8_t> DecodeRawRequestId(const std::string &requestId);
    static std::string GetRequestIdFromMsg(const std::string &messageId);
    static std::string GetRequestIdFromObj(const std::string &objectId);
    static std::string GenPacketId();
    static std::string GenMessageId(const std::string &requestId, uint8_t index = 0);
    static std::string GenObjectId(std::function<void(std::string &, const std::string &)> generateKey = nullptr);
    static std::string GenObjectId(const std::string &requestId, uint8_t index = 0);
    static std::string GenGroupId(const std::string &appId);

private:
    static std::string GenerateUUID(size_t size);
    static std::string Uint8ToHex(uint8_t value);
    static uint8_t HexToUint8(const std::string &hex);
    static uint8_t HexCharToUint8(char c);
    static constexpr size_t appIdLen_ = 8;
    static constexpr size_t traceIdLen_ = 8;
    static constexpr size_t requestIdLen_ = 18;
    static constexpr size_t packIdLen_ = 36;
    static constexpr size_t objIdLen_ = requestIdLen_ + 2;
    static constexpr size_t groupIdLen_ = 16;
    static const std::string appIdPrefix_;
    static const std::string traceIdPrefix_;
    static const std::string traceIdSuffix_;
    static constexpr char hexDigits_[] = "0123456789abcdef";
};
}  // namespace utility
}  // namespace YR