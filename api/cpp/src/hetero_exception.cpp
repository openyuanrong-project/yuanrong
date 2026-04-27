/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

#include "yr/api/hetero_exception.h"
#include "src/libruntime/err_type.h"

namespace YR {

HeteroException::HeteroException(int code, int moduleCode, const std::string &msg, std::vector<std::string> failedList)
    : code_(code), mCode_(moduleCode), msg_(msg), failedList_(failedList)
{
    codeMsg_ = "ErrCode: " + std::to_string(code) + ", ModuleCode: " + std::to_string(moduleCode) + ", ErrMsg: " + msg;
}

HeteroException HeteroException::IncorrectFunctionUsageException(const std::string &exceptionMsg)
{
    return HeteroException{YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE, YR::Libruntime::ModuleCode::RUNTIME,
                           exceptionMsg};
}

HeteroException HeteroException::InvalidParamException(const std::string &exceptionMsg)
{
    return HeteroException(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                           exceptionMsg);
}
}  // namespace YR