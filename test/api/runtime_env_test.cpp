/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#include "yr/api/runtime_env.h"
#include <gtest/gtest.h>
#include <string>
#include "yr/api/constant.h"

namespace YR {
namespace internal {

class RuntimeEnvTest : public ::testing::Test {
protected:
    void SetUp() override {
        env.Set<int>("test_int", 100);
        env.Set<std::string>("test_str", "测试字符串");
        env.SetJsonStr("test_json", R"({"key":"value","num":123})");
    }

    RuntimeEnv env;
};

// 测试Set/Get基本数据类型
TEST_F(RuntimeEnvTest, ShouldHandlePrimitiveTypes) {
    env.Set<int>("int_val", 42);
    EXPECT_EQ(env.Get<int>("int_val"), 42);

    env.Set<double>("double_val", 3.14);
    EXPECT_DOUBLE_EQ(env.Get<double>("double_val"), 3.14);

    env.Set<bool>("bool_val", true);
    EXPECT_TRUE(env.Get<bool>("bool_val"));
}

// 测试Set/Get字符串
TEST_F(RuntimeEnvTest, ShouldHandleString) {
    env.Set<std::string>("str_val", "测试字符串");
    EXPECT_EQ(env.Get<std::string>("str_val"), "测试字符串");
}

// 测试Set/Get容器类型
TEST_F(RuntimeEnvTest, ShouldHandleContainers) {
    // 测试vector
    std::vector<int> vec{1, 2, 3};
    env.Set<std::vector<int>>("vec_val", vec);
    auto vec_ret = env.Get<std::vector<int>>("vec_val");
    EXPECT_EQ(vec_ret, vec);

    // 测试map
    std::map<std::string, int> map_val{{"a", 1}, {"b", 2}};
    env.Set<std::map<std::string, int>>("map_val", map_val);
    auto map_ret = env.Get<std::map<std::string, int>>("map_val");
    EXPECT_EQ(map_ret, map_val);
}

// 测试异常情况 - Get不存在的字段
TEST_F(RuntimeEnvTest, ShouldThrowWhenFieldNotExist) {
    EXPECT_THROW(env.Get<int>("non_exist"), YR::Exception);
}

// 测试异常情况 - 类型不匹配
TEST_F(RuntimeEnvTest, ShouldThrowWhenTypeMismatch) {
    env.Set<int>("int_field", 100);

    // 尝试用错误类型获取
    EXPECT_THROW(env.Get<std::string>("int_field"), YR::Exception);

    // 测试错误消息包含类型信息
    try {
        env.Get<std::string>("int_field");
    } catch (const YR::Exception& e) {
        EXPECT_NE(std::string(e.what()).find("Failed to get the field"), std::string::npos);
    }
}

// 测试边界值
TEST_F(RuntimeEnvTest, ShouldHandleEmptyValues) {
    env.Set<std::string>("empty_str", "");
    EXPECT_TRUE(env.Get<std::string>("empty_str").empty());

    std::vector<int> empty_vec;
    env.Set<std::vector<int>>("empty_vec", empty_vec);
    EXPECT_TRUE(env.Get<std::vector<int>>("empty_vec").empty());
}

// 测试特殊字符字段名
TEST_F(RuntimeEnvTest, ShouldHandleSpecialCharNames) {
    std::string special_name = "field@name#123";
    env.Set<int>(special_name, 100);
    EXPECT_EQ(env.Get<int>(special_name), 100);
}

// 测试基础Set/Get功能
TEST_F(RuntimeEnvTest, ShouldCorrectlySetAndGetValues) {
    EXPECT_EQ(env.Get<int>("test_int"), 100);
    EXPECT_EQ(env.Get<std::string>("test_str"), "测试字符串");

    auto jsonStr = env.GetJsonStr("test_json");
    nlohmann::json j = nlohmann::json::parse(jsonStr);
    EXPECT_EQ(j["key"], "value");
    EXPECT_EQ(j["num"], 123);
}

// 测试异常场景
TEST_F(RuntimeEnvTest, ShouldThrowWhenGettingNonexistentField) {
    EXPECT_THROW(env.Get<int>("non_exist"), YR::Exception);
    EXPECT_THROW(env.GetJsonStr("non_exist"), YR::Exception);
}

// 测试JSON字符串处理
TEST_F(RuntimeEnvTest, ShouldHandleJsonStringsProperly) {
    // 测试无效JSON
    EXPECT_THROW(env.SetJsonStr("bad_json", "{invalid}"), YR::Exception);

    // 测试嵌套JSON
    env.SetJsonStr("nested_json", R"({"nested":{"key":"value"}})");
    auto jsonStr = env.GetJsonStr("nested_json");
    nlohmann::json j = nlohmann::json::parse(jsonStr);
    EXPECT_EQ(j["nested"]["key"], "value");
}

// 测试Contains方法
TEST_F(RuntimeEnvTest, ShouldCorrectlyCheckFieldExistence) {
    EXPECT_TRUE(env.Contains("test_int"));
    EXPECT_FALSE(env.Contains("non_exist"));
}

// 测试Remove方法
TEST_F(RuntimeEnvTest, ShouldRemoveFieldsCorrectly) {
    EXPECT_TRUE(env.Remove("test_int"));
    EXPECT_FALSE(env.Contains("test_int"));
    EXPECT_FALSE(env.Remove("non_exist"));  // 删除不存在的字段应返回false
}

// 测试Empty方法
TEST_F(RuntimeEnvTest, ShouldCheckEmptyState) {
    RuntimeEnv emptyEnv;
    EXPECT_TRUE(emptyEnv.Empty());

    emptyEnv.Set<int>("temp", 1);
    EXPECT_FALSE(emptyEnv.Empty());

    emptyEnv.Remove("temp");
    EXPECT_TRUE(emptyEnv.Empty());
}

// 测试模板特化
TEST_F(RuntimeEnvTest, ShouldSupportVariousDataTypes) {
    env.Set<double>("double_val", 3.14);
    env.Set<bool>("bool_val", true);

    EXPECT_DOUBLE_EQ(env.Get<double>("double_val"), 3.14);
    EXPECT_TRUE(env.Get<bool>("bool_val"));
}
}
}