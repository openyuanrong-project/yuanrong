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

#include "src/libruntime/err_type.h"

using namespace YR::Libruntime;
class ErrorInfoTest : public testing::Test {
public:
    ErrorInfoTest() = default;
    ~ErrorInfoTest() = default;
};

TEST_F(ErrorInfoTest, Test_When_Construct_And_Copy_ErrorInfo_Should_Be_Same)
{
    ErrorInfo errOne(ErrorCode::ERR_DEPENDENCY_FAILED, ModuleCode::RUNTIME, "dependency resolve failed", true);

    ErrorInfo errConstruct(errOne);
    ASSERT_TRUE(errOne.Code() == errConstruct.Code());
    ASSERT_TRUE(errOne.MCode() == errConstruct.MCode());
    ASSERT_TRUE(errOne.Msg() == errConstruct.Msg());
    ASSERT_TRUE(errOne.IsCreate() == errConstruct.IsCreate());

    ErrorInfo errCopy = errOne;
    ASSERT_TRUE(errOne.Code() == errCopy.Code());
    ASSERT_TRUE(errOne.MCode() == errCopy.MCode());
    ASSERT_TRUE(errOne.Msg() == errCopy.Msg());
    ASSERT_TRUE(errOne.IsCreate() == errCopy.IsCreate());
}