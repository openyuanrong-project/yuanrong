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

#include <iomanip>
#include <openssl/err.h>
#include <openssl/evp.h>
#include <string>
#include <vector>

#include "datasystem/utils/sensitive_value.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/utils/raii.h"

namespace YR {
namespace Libruntime {
using SensitiveValue = datasystem::SensitiveValue;
const size_t AES_IV_MIN_LEN = 12;
const size_t AES_IV_MAX_LEN = 16;
const size_t AES_TAG_LEN = 16;
const size_t BYTE_PRE_HEX = 2;
const int AES_128_GCM_KEY_LEN = 16;
const int AES_192_GCM_KEY_LEN = 24;
const int AES_256_GCM_KEY_LEN = 32;
const uint8_t HEX_TO_BYTE_MIN_LEN = 2;

const EVP_CIPHER *GetGcmAlg(uint32_t keyLen);
std::vector<unsigned char> HexDecodeToBytes(const std::string &hex);
std::string HexEncodeToString(const std::vector<unsigned char>& data, bool upperCase = false);
YR::Libruntime::ErrorInfo CipherCut(std::string &cipher, std::string &iv, std::string &tag, std::string &content);
YR::Libruntime::ErrorInfo gcmDecrypt(std::vector<unsigned char> &iv, std::vector<unsigned char> &tag,
                                     std::vector<unsigned char> &content, const std::vector<unsigned char> &key,
                                     std::vector<unsigned char> &plaintext);
YR::Libruntime::ErrorInfo DecryptWithGCMSpecifyKey(SensitiveValue &encryptData, SensitiveValue &dataKey);
}  // namespace Libruntime
}  // namespace YR
