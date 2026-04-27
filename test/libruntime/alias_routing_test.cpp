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

#include "src/libruntime/invokeadaptor/alias_routing.h"

using namespace YR::Libruntime;

namespace YR {
namespace test {
class AliasRoutingTest : public ::testing::Test {
    void SetUp() override {}

    void TearDown() override {}
};

static std::vector<AliasElement> g_aes = {
    {
        .aliasUrn = "fake_alias_urn",
        .functionUrn = "fake_function_urn",
        .functionVersionUrn = "fake_function_version_urn",
        .name = "fake_name",
        .functionVersion = "fake_function_version",
        .revisionId = "fake_revision_id",
        .description = "fake_description",
        .routingType = "",
        .routingConfig =
            {
                {
                    .functionVersionUrn = "fake_function_version_urn_1",
                    .weight = 50.0,
                },
                {
                    .functionVersionUrn = "fake_function_version_urn_2",
                    .weight = 50.0,
                },
            },
    },
};

TEST_F(AliasRoutingTest, CheckAliasTest)
{
    AliasRouting ar;
    auto ok = ar.CheckAlias("tenantId");
    ASSERT_EQ(ok, false);
    ok = ar.CheckAlias("tenantId/fullName/0");
    ASSERT_EQ(ok, false);
    ok = ar.CheckAlias("tenantId/fullName/latest");
    ASSERT_EQ(ok, false);
    ok = ar.CheckAlias("tenantId/fullName/alias");
    ASSERT_EQ(ok, true);
    ok = ar.CheckAlias("tenantId/fullName/alia-_s0");
    ASSERT_EQ(ok, true);
}

TEST_F(AliasRoutingTest, ParseAliasTest)
{
    AliasRouting ar;

    std::unordered_map<std::string, std::string> m;
    auto functionId = ar.ParseAlias("tenantId/function/version", m);
    ASSERT_EQ(functionId, "tenantId/function/version");
    functionId = ar.ParseAlias("tenantId/function", m);
    ASSERT_EQ(functionId, "tenantId/function");

    std::vector<AliasElement> g_aes_rule1 = {
        {
            .aliasUrn = "sn:cn:yrk:default:function:helloworld:myaliasv1",
            .functionUrn = "sn:cn:yrk:default:function:helloworld",
            .functionVersionUrn = "sn:cn:yrk:default:function:helloworld:$latest",
            .name = "myaliasv1",
            .functionVersion = "$latest",
            .revisionId = "20210617023315921",
            .description = "fake_description",
            .routingType = "rule",
            .routingRules =
                {
                    .ruleLogic = "and",
                    .rules =
                        {
                            "userType:=:VIP",
                            "age:<=:20",
                            "devType:in:P40,P50,MATE40",
                        },
                    .grayVersion = "sn:cn:yrk:default:function:helloworld:1",
                },
        },
    };
    ar.UpdateAliasInfo(g_aes_rule1);


    std::unordered_map<std::string, std::string> params;
    params["userType"] = "VIP";
    params["age"] = "10";
    params["devType"] = "MATE40";
    functionId = ar.ParseAlias("default/helloworld/myaliasv1", params);
    ASSERT_EQ(functionId, "default/helloworld/1");

    g_aes_rule1[0].routingRules.grayVersion = "sn:cn:yrk:default:function:helloworld:2";
    ar.UpdateAliasInfo(g_aes_rule1);
    functionId = ar.ParseAlias("default/helloworld/myaliasv1", params);
    ASSERT_EQ(functionId, "default/helloworld/2");
}

TEST_F(AliasRoutingTest, UpdateAliasInfoTest)
{
    AliasRouting ar;
    ar.UpdateAliasInfo(g_aes);
    std::unordered_map<std::string, std::string> params;
    auto urn1 = ar.GetFuncVersionUrnWithParams("fake_alias_urn", params);
    auto urn2 = ar.GetFuncVersionUrnWithParams("fake_alias_urn", params);
    ASSERT_EQ(urn1, "fake_function_version_urn_1");
    ASSERT_EQ(urn2, "fake_function_version_urn_2");
}

TEST_F(AliasRoutingTest, GetFuncVersionUrnWithParamsNotFoundTest)
{
    AliasRouting ar;
    ar.UpdateAliasInfo(g_aes);
    std::unordered_map<std::string, std::string> params;
    auto urn1 = ar.GetFuncVersionUrnWithParams("xxx", params);
    ASSERT_EQ(urn1, "xxx");
}

std::vector<AliasElement> g_aes_rule = {
    {
        .aliasUrn = "fake_alias_urn",
        .functionUrn = "fake_function_urn",
        .functionVersionUrn = "fake_function_version_urn",
        .name = "fake_name",
        .functionVersion = "fake_function_version",
        .revisionId = "fake_revision_id",
        .description = "fake_description",
        .routingType = "rule",
        .routingRules =
            {
                .ruleLogic = "and",
                .rules =
                    {
                        "userType:=:VIP",
                        "age:<=:20",
                        "devType:in:P40,P50,MATE40",
                    },
                .grayVersion = "fake_gray_version",
            },
    },
};

TEST_F(AliasRoutingTest, GetFuncVersionUrnWithParamsMatchTest)
{
    AliasRouting ar;
    ar.UpdateAliasInfo(g_aes_rule);
    std::unordered_map<std::string, std::string> params;
    params["userType"] = "VIP";
    params["age"] = "10";
    params["devType"] = "MATE40";
    auto urn1 = ar.GetFuncVersionUrnWithParams("fake_alias_urn", params);
    ASSERT_EQ(urn1, "fake_gray_version");
}

TEST_F(AliasRoutingTest, GetFuncVersionUrnWithParamsIntNoMatchTest)
{
    AliasRouting ar;
    ar.UpdateAliasInfo(g_aes_rule);
    std::unordered_map<std::string, std::string> params;
    params["userType"] = "VIP";
    params["age"] = "50";
    params["devType"] = "P40";
    auto urn1 = ar.GetFuncVersionUrnWithParams("fake_alias_urn", params);
    ASSERT_EQ(urn1, "fake_function_version_urn");
}

TEST_F(AliasRoutingTest, GetFuncVersionUrnWithParamsInNoMatchTest)
{
    AliasRouting ar;
    ar.UpdateAliasInfo(g_aes_rule);
    std::unordered_map<std::string, std::string> params;
    params["userType"] = "VIP";
    params["age"] = "10";
    params["devType"] = "P40X";
    auto urn1 = ar.GetFuncVersionUrnWithParams("fake_alias_urn", params);
    ASSERT_EQ(urn1, "fake_function_version_urn");
}

TEST_F(AliasRoutingTest, GetFuncVersionUrnWithParamsStringEqNoMatchTest)
{
    AliasRouting ar;
    ar.UpdateAliasInfo(g_aes_rule);
    std::unordered_map<std::string, std::string> params;
    params["userType"] = "VVIP";
    params["age"] = "10";
    params["devType"] = "P40";
    auto urn1 = ar.GetFuncVersionUrnWithParams("fake_alias_urn", params);
    ASSERT_EQ(urn1, "fake_function_version_urn");
}
}  // namespace test
}  // namespace YR