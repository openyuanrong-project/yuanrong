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

#include "hash_utils.h"

#include <iomanip>
#include <sstream>
#include <openssl/hmac.h>
#include <openssl/sha.h>

namespace YR {
namespace Libruntime {
const static unsigned int CHAR_TO_HEX = 2;
const static std::string HEX_STRING_SET = "0123456789abcdef";
const int32_t FIRST_FOUR_BIT_MOVE = 4;
std::string GetHMACSha256(const SensitiveValue &key, const std::string &data)
{
    HMAC_CTX *ctx = HMAC_CTX_new();
    HMAC_Init_ex(ctx, key.GetData(), static_cast<int>(key.GetSize()), EVP_sha256(), nullptr);
    HMAC_Update(ctx, reinterpret_cast<const unsigned char *>(&data[0]), data.length());
    unsigned int mdLength = EVP_MAX_MD_SIZE;
    unsigned char md[EVP_MAX_MD_SIZE];
    HMAC_Final(ctx, md, &mdLength);
    HMAC_CTX_free(ctx);
    std::stringstream ss;
    ss << std::hex << std::setfill('0');
    for (unsigned int i = 0; i < mdLength; i++) {
        ss << std::setw(CHAR_TO_HEX) << static_cast<unsigned int>(md[i]);
    }
    return ss.str();
}

void SHA256AndHex(const std::string &input, std::stringstream &output)
{
    unsigned char sha256Chars[SHA256_DIGEST_LENGTH];
    SHA256(reinterpret_cast<const unsigned char *>(input.c_str()), input.size(), sha256Chars);
    for (const auto &c : sha256Chars) {
        output << HEX_STRING_SET[c >> FIRST_FOUR_BIT_MOVE] << HEX_STRING_SET[c & 0xf];
    }
    output << "\n";
}
}  // namespace Libruntime
}  // namespace YR