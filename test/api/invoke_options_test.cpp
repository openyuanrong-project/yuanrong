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

#include "api/cpp/include/yr/api/invoke_options.h"
#include "common/common.h"

using namespace testing;
namespace YR {
namespace test {
class InvokeOptionsTest : public testing::Test {
public:
    InvokeOptionsTest() = default;
    ~InvokeOptionsTest() = default;
};

bool RetryForEverything(const Exception &e) noexcept
{
    return true;
}

TEST_F(InvokeOptionsTest, RetryChecker)
{
    InvokeOptions opts;
    opts.retryTimes = 5;
    opts.retryChecker = RetryForEverything;
    EXPECT_NO_THROW(opts.CheckOptionsValid());

    opts.retryTimes = 20;
    EXPECT_THROW_WITH_CODE_AND_MSG(opts.CheckOptionsValid(), 1001, "invalid opts retryTimes");

    opts.retryTimes = 0;
    EXPECT_NO_THROW(opts.CheckOptionsValid());

    YR::InstanceRange instanceRange;
    instanceRange.min = 1;
    instanceRange.max = 10;
    opts.instanceRange = instanceRange;
    EXPECT_NO_THROW(opts.CheckOptionsValid());

    opts.groupName = "groupName";
    EXPECT_THROW_WITH_CODE_AND_MSG(
        opts.CheckOptionsValid(), 1001,
        "gang scheduling and range scheduling cannot be used at the same time, please select one scheduling to set.");

    opts.groupName = "";
    instanceRange.min = -1;
    instanceRange.max = -1;
    opts.instanceRange = instanceRange;
    EXPECT_NO_THROW(opts.CheckOptionsValid());

    instanceRange.min = -1;
    instanceRange.max = 10;
    opts.instanceRange = instanceRange;
    EXPECT_THROW_WITH_CODE_AND_MSG(opts.CheckOptionsValid(), 1001,
                                   "please set the min and the max as follows: max = min = -1 or max >= min > 0");

    instanceRange.min = -2;
    instanceRange.max = -2;
    opts.instanceRange = instanceRange;
    EXPECT_THROW_WITH_CODE_AND_MSG(opts.CheckOptionsValid(), 1001,
                                   "please set the min and the max as follows: max = min = -1 or max >= min > 0");
}
}  // namespace test
}  // namespace YR