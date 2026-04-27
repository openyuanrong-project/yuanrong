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

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include "src/libruntime/utils/hash_utils.h"

namespace YR {
namespace test {
using namespace YR::Libruntime;
class HashUtilTest : public ::testing::Test {
protected:
    void SetUp() override {}

    void TearDown() override {}
};

TEST_F(HashUtilTest, CorrectHMACSHA256Test2)
{
    SensitiveValue key = std::string("secret");
    std::string worldSha256 = GetHMACSha256(key, "Hello, World!");
    EXPECT_EQ(worldSha256, "fcfaffa7fef86515c7beb6b62d779fa4ccf092f2e61c164376054271252821ff");
}

}  // namespace test
}  // namespace YR