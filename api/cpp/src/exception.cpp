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

#include "yr/api/exception.h"
#include "src/libruntime/err_type.h"

namespace YR {
Exception::Exception() : code(YR::Libruntime::ErrorCode::ERR_OK), mCode(YR::Libruntime::ModuleCode::RUNTIME) {}

Exception::Exception(const std::string &msg)
    : code(YR::Libruntime::ErrorCode::ERR_OK), mCode(YR::Libruntime::ModuleCode::RUNTIME), msg(msg)
{
    codeMsg = "ErrCode: " + std::to_string(code) + ", ModuleCode: " + std::to_string(mCode) + ", ErrMsg: " + msg;
}

Exception::Exception(int code, const std::string &msg)
    : code(code), mCode(YR::Libruntime::ModuleCode::RUNTIME), msg(msg)
{
    codeMsg = "ErrCode: " + std::to_string(code) + ", ModuleCode: " + std::to_string(mCode) + ", ErrMsg: " + msg;
}

Exception::Exception(int code, int moduleCode, const std::string &msg) : code(code), mCode(moduleCode), msg(msg)
{
    codeMsg = "ErrCode: " + std::to_string(code) + ", ModuleCode: " + std::to_string(moduleCode) + ", ErrMsg: " + msg;
}

Exception Exception::RegisterRecoverFunctionException()
{
    return Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_INIT_USAGE, YR::Libruntime::ModuleCode::RUNTIME,
                     "YR_RECOVER only supported member function");
}

Exception Exception::RegisterShutdownFunctionException()
{
    return Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_INIT_USAGE, YR::Libruntime::ModuleCode::RUNTIME,
                     "YR_SHUTDOWN only supported member function");
}

Exception Exception::DeserializeException(const std::string &exceptionMsg)
{
    return Exception(YR::Libruntime::ErrorCode::ERR_DESERIALIZATION_FAILED, YR::Libruntime::ModuleCode::RUNTIME,
                     exceptionMsg);
}

Exception Exception::RegisterFunctionException(const std::string &funcName)
{
    return Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_INVOKE_USAGE, YR::Libruntime::ModuleCode::RUNTIME_INVOKE,
                     "YR_INVOKE function is duplicated, function name: " + funcName);
}

Exception Exception::InvalidParamException(const std::string &exceptionMsg)
{
    return Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME, exceptionMsg);
}

Exception Exception::GetException(const std::string &exceptionMsg)
{
    return {YR::Libruntime::ErrorCode::ERR_GET_OPERATION_FAILED, YR::Libruntime::ModuleCode::RUNTIME, exceptionMsg};
}

Exception Exception::InnerSystemException(const std::string &exceptionMsg)
{
    return {YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME, exceptionMsg};
}

Exception Exception::UserCodeException(const std::string &exceptionMsg)
{
    return {YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION, YR::Libruntime::ModuleCode::RUNTIME, exceptionMsg};
}

Exception Exception::InstanceIdEmptyException(const std::string &exceptionMsg)
{
    return {YR::Libruntime::ErrorCode::ERR_INSTANCE_ID_EMPTY, YR::Libruntime::ModuleCode::RUNTIME, exceptionMsg};
}

Exception Exception::IncorrectInvokeUsageException(const std::string &exceptionMsg)
{
    return {YR::Libruntime::ErrorCode::ERR_INCORRECT_INVOKE_USAGE, YR::Libruntime::ModuleCode::RUNTIME, exceptionMsg};
}

Exception Exception::IncorrectFunctionUsageException(const std::string &exceptionMsg)
{
    return {YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE, YR::Libruntime::ModuleCode::RUNTIME, exceptionMsg};
}
}  // namespace YR