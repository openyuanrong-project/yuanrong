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

#include "src/libruntime/utils/grpc_utils.h"
#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include "datasystem/utils/sensitive_value.h"
#include "src/libruntime/utils/utils.h"

namespace YR {
namespace test {
using SensitiveValue = datasystem::SensitiveValue;
class GrpcUtilTest : public ::testing::Test {
protected:
    void SetUp() override {}

    void TearDown() override {}
};

TEST_F(GrpcUtilTest, SignAndVerifyStreamingMessageTest)
{
    std::string accessKey = "access_key";
    SensitiveValue secretKey = std::string("secret_key");
    runtime_rpc::StreamingMessage message;
    message.set_messageid("message_id");

    // fail to sign empty StreamingMessage
    ASSERT_FALSE(YR::Libruntime::SignStreamingMessage(accessKey, secretKey, message));
    ASSERT_FALSE(message.metadata().contains("access_key"));
    ASSERT_FALSE(message.metadata().contains("signature"));
    ASSERT_FALSE(message.metadata().contains("timestamp"));

    message.mutable_callreq()->set_requestid("123");
    ASSERT_TRUE(YR::Libruntime::SignStreamingMessage(accessKey, secretKey, message));
    ASSERT_TRUE(message.metadata().contains("access_key"));
    ASSERT_EQ(message.metadata().at("access_key"), accessKey);
    ASSERT_TRUE(message.metadata().contains("signature"));
    ASSERT_TRUE(message.metadata().contains("timestamp"));

    ASSERT_TRUE(YR::Libruntime::VerifyStreamingMessage(accessKey, secretKey, message));

    ASSERT_FALSE(YR::Libruntime::VerifyStreamingMessage("fake_access_key", secretKey, message));

    SensitiveValue fakeSecretKey = std::string("fake_secret_key");
    ASSERT_FALSE(YR::Libruntime::VerifyStreamingMessage(accessKey, fakeSecretKey, message));

    message.mutable_callreq()->set_requestid("1234");
    ASSERT_FALSE(YR::Libruntime::VerifyStreamingMessage(accessKey, secretKey, message));

    message.clear_callreq();
    message.mutable_callresultack()->set_message("123");
    ASSERT_FALSE(YR::Libruntime::VerifyStreamingMessage(accessKey, secretKey, message));

    message.Clear();
    ASSERT_FALSE(YR::Libruntime::VerifyStreamingMessage(accessKey, secretKey, message));
}

TEST_F(GrpcUtilTest, SignAndVerifyTimestampTest)
{
    std::string accessKey = "access_key";
    SensitiveValue secretKey = std::string("secret_key");
    std::cout << std::string(secretKey.GetData(), secretKey.GetSize()) << "<>=======" << std::endl;
    auto timestamp = YR::GetCurrentUTCTime();
    auto signature = YR::Libruntime::SignTimestamp(accessKey, secretKey, timestamp);
    ASSERT_FALSE(signature.empty());

    ASSERT_TRUE(YR::Libruntime::VerifyTimestamp(accessKey, secretKey, timestamp, signature));

    ASSERT_FALSE(YR::Libruntime::VerifyTimestamp("fake_access_key", secretKey, timestamp, signature));

    SensitiveValue fakeSecretKey = std::string("fake_secret_key");
    ASSERT_FALSE(YR::Libruntime::VerifyTimestamp(accessKey, fakeSecretKey, timestamp, signature));

    auto fakeTimestamp = YR::GetCurrentUTCTime();
    ASSERT_FALSE(YR::Libruntime::VerifyTimestamp(accessKey, fakeSecretKey, fakeTimestamp, signature));

    ASSERT_FALSE(YR::Libruntime::VerifyTimestamp(accessKey, fakeSecretKey, timestamp, "fake_signature"));
}
}  // namespace test
}  // namespace YR