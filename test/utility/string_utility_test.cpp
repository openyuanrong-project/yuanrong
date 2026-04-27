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

#include <tuple>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/utility/id_generator.h"
#include "src/utility/string_utility.h"

namespace YR {
namespace test {
using namespace YR::utility;
class StringUtilityTest : public ::testing::Test {
    void SetUp() override {}

    void TearDown() override {}
};

TEST_F(StringUtilityTest, TrimSpaceTest)
{
    std::vector<std::pair<std::string, std::string>> tests = {
        {" test", "test"}, {"test ", "test"}, {" test ", "test"},       {"test", "test"},
        {"", ""},          {"   ", ""},       {" t e s t ", "t e s t"},
    };

    for (auto &test : tests) {
        auto res = TrimSpace(test.first);
        ASSERT_EQ(res, test.second) << "raw string: " << test.first << " == "
                                    << "first not of: " << test.first.find_first_not_of(' ') << " == "
                                    << "last not of: " << test.first.find_last_not_of(' ');
    }
}

TEST_F(StringUtilityTest, CompareIntFromStringTest)
{
    std::vector<std::tuple<std::string, std::string, int, std::string>> tests = {
        {"2", "1", 1, ""},
        {"-123456789123456789", "123456789123456789", -1, ""},
        {"3", "3", 0, ""},
        {"x", "y", 0, "stoll"},
    };

    for (auto &test : tests) {
        auto [res, exception] = CompareIntFromString(std::get<0>(test), std::get<1>(test));
        ASSERT_EQ(std::get<2>(test), res);
        ASSERT_EQ(std::get<3>(test), exception);
    }
}

TEST_F(StringUtilityTest, JoinTest)
{
    std::vector<std::string> parts = {"a", "b", "c"};
    auto joinedStr = Join(parts, "-");
    ASSERT_EQ(joinedStr, "a-b-c");
}

}  // namespace test
}  // namespace YR