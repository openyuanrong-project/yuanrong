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

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/libruntime/utils/crypto/aes_crypto.h"

using namespace YR::Libruntime;

namespace YR {
namespace test {
class AesCryptoTest : public ::testing::Test {
    void SetUp() override {}
    void TearDown() override {}
};

TEST_F(AesCryptoTest, GetGcmAlgTest)
{
    auto alg128 = GetGcmAlg(YR::Libruntime::AES_128_GCM_KEY_LEN);
    EXPECT_EQ(alg128, EVP_aes_128_gcm());

    auto alg192 = GetGcmAlg(YR::Libruntime::AES_192_GCM_KEY_LEN);
    EXPECT_EQ(alg192, EVP_aes_192_gcm());

    auto alg256 = GetGcmAlg(YR::Libruntime::AES_256_GCM_KEY_LEN);
    EXPECT_EQ(alg256, EVP_aes_256_gcm());

    // 假设 0 是一个无效的密钥长度
    auto alg = GetGcmAlg(0);
    EXPECT_EQ(alg, EVP_aes_256_gcm());
}

TEST_F(AesCryptoTest, HexDecodeToBytes)
{
    std::string hex = "48656c6c6f20576f726c64";  // "Hello World" in hex
    std::vector<unsigned char> expected = {0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64};
    std::vector<unsigned char> result = HexDecodeToBytes(hex);
    EXPECT_EQ(result, expected);

    hex = "";
    expected = {};
    result = HexDecodeToBytes(hex);
    EXPECT_EQ(result, expected);
}

TEST_F(AesCryptoTest, HexEncodeToString)
{
    // 测试空字节数据
    std::vector<unsigned char> data = {};
    std::string result = HexEncodeToString(data, false);
    EXPECT_EQ(result, "");

    // 测试多个字节数据，不转换为大写
    data = {0x48, 0x65, 0x6c, 0x6c, 0x6f};  // "Hello" in hex
    result = HexEncodeToString(data, false);
    EXPECT_EQ(result, "48656c6c6f");

    // 测试多个字节数据，转换为大写
    data = {0x48, 0x65, 0x6c, 0x6c, 0x6f};  // "Hello" in hex
    result = HexEncodeToString(data, true);
    EXPECT_EQ(result, "48656C6C6F");
}

TEST_F(AesCryptoTest, CipherCut)
{
    // 测试有效的输入
    std::string cipher = "0123456789abcdef0123456789abcdef:0123456789abcdef0123456789abcdef0123456789abcdef";
    std::string iv, tag, content;
    YR::Libruntime::ErrorInfo error = CipherCut(cipher, iv, tag, content);
    EXPECT_TRUE(error.OK());
    EXPECT_EQ(iv, "0123456789abcdef0123456789abcdef");
    EXPECT_EQ(tag, "0123456789abcdef0123456789abcdef");
    EXPECT_EQ(content, "0123456789abcdef");

    // 测试无效的输入：长度不足
    cipher = "0123456789abcdef:0123456789abc";
    error = CipherCut(cipher, iv, tag, content);
    EXPECT_FALSE(error.OK());
    EXPECT_EQ(error.Msg(), "Incorrect cipher format, cipher length is less than minimum value!");

    // 测试无效的输入：缺少分隔符
    cipher = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef";
    error = CipherCut(cipher, iv, tag, content);
    EXPECT_FALSE(error.OK());
    EXPECT_EQ(error.Msg(), "Incorrect cipher text format!");

    // 测试无效的输入：IV长度不足
    cipher = "0123456789abcdef0:0123456789abcdef0123456789abcdef0123456789abcdef";
    error = CipherCut(cipher, iv, tag, content);
    EXPECT_FALSE(error.OK());
    EXPECT_EQ(error.Msg(), "Incorrect cipher text format!");

    // 测试无效的输入：IV长度过长
    cipher = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef";
    error = CipherCut(cipher, iv, tag, content);
    EXPECT_FALSE(error.OK());
    EXPECT_EQ(error.Msg(), "Incorrect cipher text format!");

    // 测试无效的输入：分隔符位置错误
    cipher = "0123456789abcdef:0123456789abcdef0123456789abcdef0123456789abcdef";
    error = CipherCut(cipher, iv, tag, content);
    EXPECT_FALSE(error.OK());
    EXPECT_EQ(error.Msg(), "Incorrect cipher text format!");
}

TEST_F(AesCryptoTest, gcmDecrypt)
{
    // 测试无效的输入：无效的密钥长度
    std::vector<unsigned char> key = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10};
    std::vector<unsigned char> iv = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B};
    std::vector<unsigned char> tag = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F};
    std::vector<unsigned char> content = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F};
    std::vector<unsigned char> plaintext;
    YR::Libruntime::ErrorInfo error = gcmDecrypt(iv, tag, content, key, plaintext);
    EXPECT_FALSE(error.OK());

    // 测试无效的输入：无效的IV长度
    key = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F};
    iv = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C};
    tag = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F};
    content = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F};
    error = gcmDecrypt(iv, tag, content, key, plaintext);
    EXPECT_FALSE(error.OK());

    // 测试无效的输入：无效的Tag长度
    key = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F};
    iv = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B};
    tag = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E};
    content = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F};
    error = gcmDecrypt(iv, tag, content, key, plaintext);
    EXPECT_FALSE(error.OK());

    // 测试无效的输入：无效的Content长度
    key = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F};
    iv = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B};
    tag = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F};
    content = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E};
    error = gcmDecrypt(iv, tag, content, key, plaintext);
    EXPECT_FALSE(error.OK());
}

TEST_F(AesCryptoTest, DecryptWithGCMSpecifyKey)
{
    // 测试有效的输入
    {
        SensitiveValue dataKey = std::string("ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890");
        SensitiveValue encryptData = std::string("677f190e08a97640fb086c9a:"
                                                 "cb5b9250f1d2966e273521cd613e883866999f3c1283f1ec70258ba0c2");
        YR::Libruntime::ErrorInfo error = DecryptWithGCMSpecifyKey(encryptData, dataKey);
        EXPECT_TRUE(error.OK());
    }

    // 测试无效的输入：空的dataKey
    {
        SensitiveValue dataKey = std::string("");
        SensitiveValue encryptData = std::string("0123456789abcdef:0123456789abcdef0123456789abcdef0123456789abcdef");
        YR::Libruntime::ErrorInfo error = DecryptWithGCMSpecifyKey(encryptData, dataKey);
        EXPECT_FALSE(error.OK());
    }

    // 测试无效的输入：空的encryptData
    {
        SensitiveValue dataKey = std::string("1234567890123456");
        SensitiveValue encryptData = std::string("");
        YR::Libruntime::ErrorInfo error = DecryptWithGCMSpecifyKey(encryptData, dataKey);
        EXPECT_FALSE(error.OK());
    }

    // 测试无效的输入：无效的gcmDecrypt结果
    {
        SensitiveValue dataKey = std::string("1234567890123456");
        SensitiveValue encryptData = std::string("0123456789abcdef:0123456789abcdef0123456789abcdef0123456789abc");
        YR::Libruntime::ErrorInfo error = DecryptWithGCMSpecifyKey(encryptData, dataKey);
        EXPECT_FALSE(error.OK());
    }
}
}  // namespace test
}  // namespace YR