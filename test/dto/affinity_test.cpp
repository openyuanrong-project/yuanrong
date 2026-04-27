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
#include <cstdlib>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#define protected public
#include "src/dto/affinity.h"
using namespace YR::Libruntime;
class DTOAffinityTest : public testing::Test {
public:
    DTOAffinityTest(){};
    ~DTOAffinityTest(){};
    void SetUp() override {}
    void TearDown() override {}
};

TEST_F(DTOAffinityTest, TestInstanceAntiOthersIsFalse)
{
    LabelOperator label{"LabelIn"};
    label.SetValues({"value1", "value2"});
    auto affinity = InstancePreferredAffinity();
    affinity.SetLabelOperators({std::make_shared<LabelOperator>(label)});
    affinity.preferredAntiOtherLabels = true;

    PBAffinity pbAffinity;
    affinity.UpdatePbAffinity(&pbAffinity);

    EXPECT_TRUE(pbAffinity.instance().preferredaffinity().condition().orderpriority());
    EXPECT_TRUE(pbAffinity.instance().preferredaffinity().condition().subconditions_size() != 0);
}

TEST_F(DTOAffinityTest, TestResourcePreferredAntiAffinity)
{
    LabelOperator label{"LabelIn"};
    label.SetValues({"value1", "value2"});
    auto affinity = ResourcePreferredAntiAffinity();
    affinity.SetLabelOperators({std::make_shared<LabelOperator>(label)});
    affinity.preferredAntiOtherLabels = true;

    PBAffinity pbAffinity;
    affinity.UpdatePbAffinity(&pbAffinity);

    EXPECT_EQ(pbAffinity.resource().requiredantiaffinity().condition().subconditions_size() != 0, true);
}