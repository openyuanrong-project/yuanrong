/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

#include "src/utility/id_generator.h"
#include "src/utility/string_utility.h"

namespace YR {
namespace test {
using namespace YR::utility;

class IdTest : public ::testing::Test {
    void SetUp() override {}

    void TearDown() override {}
};

TEST_F(IdTest, ApplicationIdTest)
{
    auto appId = IDGenerator::GenApplicationId();
    ASSERT_EQ(appId.size(), 12);
    ASSERT_EQ(appId.substr(0, 4), "job-");
}

TEST_F(IdTest, RequestIdTest)
{
    auto requestId = IDGenerator::GenRequestId();
    ASSERT_EQ(requestId.size(), 18);
    ASSERT_EQ(requestId.substr(16), "00");

    requestId = IDGenerator::GenRequestId(1);
    ASSERT_EQ(requestId.size(), 18);
    ASSERT_EQ(requestId.substr(16), "01");

    requestId = IDGenerator::GenRequestId(requestId, 3);
    ASSERT_EQ(requestId.size(), 18);
    ASSERT_EQ(requestId.substr(16), "03");

    std::string initCallId = "c943c1890198b057Aa-18@initcall";
    auto rawId = IDGenerator::DecodeRawRequestId(initCallId).first;
    ASSERT_EQ(rawId, "c943c1890198b05700");
}

TEST_F(IdTest, MessageIdTest)
{
    auto requestId = IDGenerator::GenRequestId();
    auto messageId = IDGenerator::GenMessageId(requestId, 1);
    ASSERT_EQ(messageId.size(), 20);
    ASSERT_EQ(messageId.substr(18), "01");
}

TEST_F(IdTest, ObjectIdTest)
{
    auto requestId = IDGenerator::GenRequestId();
    auto objId = IDGenerator::GenObjectId(requestId);
    ASSERT_EQ(objId.size(), 20);
    ASSERT_EQ(objId.substr(18), "00");

    objId = IDGenerator::GenObjectId(requestId, 1);
    ASSERT_EQ(objId.size(), 20);
    ASSERT_EQ(objId.substr(18), "01");
    const std::string suffix = ";0b7c77fa-ef0e-4a34-b5c1-deab89db82e6";
    auto generateKey = [suffix](std::string &dsObjId, const std::string &objId) { dsObjId = objId + suffix; };

    auto dsObjId = IDGenerator::GenObjectId(generateKey);
    ASSERT_EQ(dsObjId.size(), 20 + suffix.size());
    ASSERT_EQ(dsObjId.substr(20), suffix);
}

TEST_F(IdTest, TraceIdTest)
{
    auto traceId = IDGenerator::GenTraceId();
    ASSERT_EQ(traceId.size(), 20);
    ASSERT_EQ(traceId.substr(0, 4), "job-");
    ASSERT_EQ(traceId.substr(12), "-trace-X");

    auto appId = IDGenerator::GenApplicationId();
    traceId = IDGenerator::GenTraceId(appId);
    ASSERT_EQ(traceId.substr(4, 8), appId.substr(4));
}

TEST_F(IdTest, GroupIdTest)
{
    auto appId = IDGenerator::GenApplicationId();
    auto groupId = IDGenerator::GenGroupId(appId);
    ASSERT_EQ(groupId.size(), 24);
    ASSERT_EQ(groupId.substr(0, 8), appId.substr(4));
}

TEST_F(IdTest, Test_When_Input_Longer_Id_Should_Truncate_Length)
{
    const std::string requestId = "9527789565bcd37900";
    const std::string suffix = "@initcall";
    std::string initCallRequestId = requestId + suffix;
    EXPECT_EQ(requestId, YR::utility::IDGenerator::GetRealRequestId(initCallRequestId));
}

TEST_F(IdTest, PacketIdTest)
{
    auto id = IDGenerator::GenPacketId();
    EXPECT_EQ(id.size(), 36);
}

TEST_F(IdTest, TestGangInitCallRequestId)
{
    auto id = "c943c1890198b05700";
    auto id1 = "c943c1890198b05700-18@initcall";
    ASSERT_EQ(id, YR::utility::IDGenerator::GetRealRequestId(id1));
    auto id2 = "c943c1890198b05700@18@initcall";
    ASSERT_EQ(id, YR::utility::IDGenerator::GetRealRequestId(id2));
}

}  // namespace test
}  // namespace YR