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

#include "aes_crypto.h"

#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
const EVP_CIPHER *GetGcmAlg(uint32_t keyLen)
{
    auto alg = EVP_aes_256_gcm();
    switch (keyLen) {
        case YR::Libruntime::AES_128_GCM_KEY_LEN:
            alg = EVP_aes_128_gcm();
            break;
        case YR::Libruntime::AES_192_GCM_KEY_LEN:
            alg = EVP_aes_192_gcm();
            break;
        case YR::Libruntime::AES_256_GCM_KEY_LEN:
            alg = EVP_aes_256_gcm();
            break;
        default:
            YRLOG_WARN("invalid key length({}), using aes_256_gcm", keyLen);
    }
    return alg;
}

// 将十六进制字符串转换为字节向量
std::vector<unsigned char> HexDecodeToBytes(const std::string &hex)
{
    std::vector<unsigned char> bytes;
    for (size_t i = 0; i < hex.length(); i += YR::Libruntime::HEX_TO_BYTE_MIN_LEN) {
        std::string byte_str = hex.substr(i, YR::Libruntime::HEX_TO_BYTE_MIN_LEN);
        auto byte = static_cast<unsigned char>(std::stoul(byte_str, nullptr, 16));
        bytes.push_back(byte);
    }
    return bytes;
}

// 将字节数据转换为十六进制字符串
std::string HexEncodeToString(const std::vector<unsigned char>& data, bool upperCase)
{
    std::stringstream ss;
    for (const auto& byte : data) {
        // 将字节转换为十六进制表示
        ss << std::hex << std::setw(YR::Libruntime::BYTE_PRE_HEX) << std::setfill('0') << static_cast<int>(byte);
    }
    std::string hexString = ss.str();
    // 根据参数决定是否转换为大写
    if (upperCase) {
        for (auto& c : hexString) {
            c = std::toupper(c);
        }
    }
    return hexString;
}

// iv(12):content+tag(16) -> iv(12):tag(16):content
YR::Libruntime::ErrorInfo CipherCut(std::string &cipher, std::string &iv, std::string &tag, std::string &content)
{
    size_t ivHexMinLen = YR::Libruntime::AES_IV_MIN_LEN * YR::Libruntime::BYTE_PRE_HEX;
    size_t ivHexMaxLen = YR::Libruntime::AES_IV_MAX_LEN * YR::Libruntime::BYTE_PRE_HEX;
    size_t tagHexLen = YR::Libruntime::AES_TAG_LEN * YR::Libruntime::BYTE_PRE_HEX;

    // cipher format is: "iv:cipher+tag"
    if (cipher.size() < ivHexMinLen + tagHexLen + 1) {
        return ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID,
                         "Incorrect cipher format, cipher length is less than minimum value!");
    }

    // Cut tag hex string.
    tag = std::move(cipher.substr(cipher.size() - tagHexLen));
    cipher = cipher.substr(0, cipher.size() - tagHexLen);

    // The range of ':' is [24, 32].
    auto cutPos = cipher.find(':', 0);
    if (cipher.find(':') == std::string::npos || cutPos < ivHexMinLen || cutPos > ivHexMaxLen ||
        cutPos == cipher.size() - 1) {
        return ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, "Incorrect cipher text format!");
    }
    
    // the range of iv is [0, cutPos), and the range of cipher is [cutPos+1, str.size-cutPos-1]
    iv = std::move(cipher.substr(0, cutPos)); // salt
    content = std::move(cipher.substr(cutPos + 1));
    return ErrorInfo();
}

// AES-GCM 解密函数
YR::Libruntime::ErrorInfo gcmDecrypt(std::vector<unsigned char> &iv, std::vector<unsigned char> &tag,
                                     std::vector<unsigned char> &content, const std::vector<unsigned char> &key,
                                     std::vector<unsigned char> &plaintext)
{
    EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();
    [[maybe_unused]] Raii evpRaii([ctx]() { EVP_CIPHER_CTX_free(ctx); });
    auto ret = EVP_DecryptInit_ex(ctx, GetGcmAlg(key.size()), nullptr, key.data(), iv.data());
    if (ret != 1) {
        std::string msg = "Decrypt init failed! ret = " + std::to_string(ret);
        return ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, msg);
    }

    if (ret = EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_AEAD_SET_TAG, YR::Libruntime::AES_TAG_LEN, tag.data()); ret != 1) {
        std::string msg = "Tag authentication failed! ret = " + std::to_string(ret);
        return ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, msg);
    }
    plaintext.resize(content.size());
    int len = 0;
    ret = EVP_DecryptUpdate(ctx, plaintext.data(), &len, content.data(), content.size());
    if (ret != 1) {
        std::string msg = "Decrypt update failed! ret = " + std::to_string(ret);
        return ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, msg);
    }
    plaintext.resize(len);
    if (ret = EVP_DecryptFinal_ex(ctx, plaintext.data() + len, &len); ret != 1) {
        std::string msg = "Decrypt final failed! ret = " + std::to_string(ret);
        return ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, msg);
    }
    plaintext.resize(plaintext.size() + len);
    YRLOG_INFO("encrypt with GCM finish");
    return ErrorInfo();
}

YR::Libruntime::ErrorInfo DecryptWithGCMSpecifyKey(SensitiveValue &encryptData, SensitiveValue &dataKey)
{
    if (dataKey.Empty()) {
        return ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, "dataKey is empty");
    }
    std::string dataKeyText(dataKey.GetData(), dataKey.GetSize());
    Raii dataKeyRaii([&dataKeyText] { std::fill(dataKeyText.begin(), dataKeyText.end(), 0); });
    auto dk = HexDecodeToBytes(dataKeyText);

    std::string cipherText(encryptData.GetData(), encryptData.GetSize());
    Raii cipherRaii([&cipherText] { std::fill(cipherText.begin(), cipherText.end(), 0); });
    std::string ivStr;
    std::string tagStr;
    std::string contentStr;
    auto err = CipherCut(cipherText, ivStr, tagStr, contentStr);
    if (!err.OK()) {
        YRLOG_ERROR("split cipher failed, err: {}", err.Msg());
        return err;
    }
    auto iv = HexDecodeToBytes(ivStr);
    auto tag = HexDecodeToBytes(tagStr);
    auto content = HexDecodeToBytes(contentStr);

    std::vector<unsigned char> result;
    Raii resRaii([&result] { std::fill(result.begin(), result.end(), 0); });
    err = gcmDecrypt(iv, tag, content, dk, result);
    if (!err.OK()) {
        YRLOG_ERROR("decrypt by gcm failed, err: {}", err.Msg());
        return err;
    }
    encryptData = std::move(std::string(reinterpret_cast<const char*>(result.data()), result.size()));
    return ErrorInfo();
}
}  // namespace Libruntime
}  // namespace YR
