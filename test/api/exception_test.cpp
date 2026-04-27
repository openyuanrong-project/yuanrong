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

#include <memory>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "api/cpp/include/yr/api/exception.h"
#include "src/libruntime/err_type.h"

using namespace testing;
namespace YR {
namespace test {
class ExceptionTest : public testing::Test {
public:
    ExceptionTest(){};
    ~ExceptionTest(){};
};

TEST_F(ExceptionTest, ExceptionTest1)
{
    Exception e = Exception();
    ASSERT_EQ(e.Code(), 0);
    e = Exception("mock-exception");
    ASSERT_EQ(e.Code(), 0);
    ASSERT_EQ(e.MCode(), Libruntime::ModuleCode::RUNTIME);
    ASSERT_EQ(e.Msg(), "mock-exception");
    e = Exception(1001, "mock-exception");
    ASSERT_EQ(std::string(e.what()),"ErrCode: 1001, ModuleCode: 20, ErrMsg: mock-exception");
    e = Exception().RegisterRecoverFunctionException();
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_INCORRECT_INIT_USAGE);
    std::string mockMsg = "mockMsg";
    e = Exception().DeserializeException(mockMsg);
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_DESERIALIZATION_FAILED);
    mockMsg = "mockMsg";
    e = Exception().RegisterFunctionException(mockMsg);
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_INCORRECT_INVOKE_USAGE);
    e = Exception().InvalidParamException(mockMsg);
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_PARAM_INVALID);
    e = Exception().GetException(mockMsg);
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_GET_OPERATION_FAILED);
    e = Exception().InnerSystemException(mockMsg);
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    e = Exception().UserCodeException(mockMsg);
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION);
    e = Exception().InstanceIdEmptyException(mockMsg);
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_INSTANCE_ID_EMPTY);
    e = Exception().IncorrectInvokeUsageException(mockMsg);
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_INCORRECT_INVOKE_USAGE);
    e = Exception().IncorrectFunctionUsageException(mockMsg);
    ASSERT_EQ(e.Code(), Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE);
}
}  // namespace test
}  // namespace YR