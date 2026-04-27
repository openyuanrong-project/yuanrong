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

#include "Function.h"

#include "json.hpp"

#include "FunctionError.h"
#include "context_impl.h"
#include "api/cpp/src/utils/utils.h"
#include "src/dto/data_object.h"
#include "src/dto/invoke_options.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/utility/logger/logger.h"

namespace Function {

const int NAME_AND_VERSION_SIZE = 2;
const char *const FUNC_NAME_PATTERN_STRING = "^[a-zA-Z]([a-zA-Z0-9._-]*[a-zA-Z0-9])?$";
const char *const VERSION_PATTERN_STRING = "^[a-zA-Z0-9]([a-zA-Z0-9_-]*\\\\.)*[a-zA-Z0-9_-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$";
const char *const ALIAS_PATTERN_STRING = "^[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$";
const char *const INSTANCE_NAME_PATTERN_STRING = "^[a-zA-Z0-9]+$";
const char *const ENV_KEY_RUNTIME_SERVICE_FUNC_VERSION = "RUNTIME_SERVICE_FUNC_VERSION";
const char *const DEFAULT_VERSION = "latest";
const char *const DEFAULT_TENANT_ID = "default";
const char FUNCTION_URN_SEPERATOR = '@';
const char FUNCTION_NAME_SEPERATOR = ':';
const int FUNCTION_NAME_LENGTH = 3;
const int SERVICE_NAME_INDEX_FOR_FUNCTION_NAME = 1;
const char ALIAS_PREFIX = '!';
const int FUNC_NAME_LENGTH_LIMIT = 128;
const int VERSION_LENGTH_LIMIT = 32;
const int ALIAS_LENGTH_LIMIT = 32;
const int INSTANCE_NAME_LENGTH_LIMIT = 128;

void CheckFuncName(const std::string &funcName)
{
    const re2::RE2 pattern(FUNC_NAME_PATTERN_STRING);
    bool match = RE2::FullMatch(funcName, pattern);
    if (!match || funcName.length() > FUNC_NAME_LENGTH_LIMIT) {
        YRLOG_WARN("Invalid function name, {}", funcName);
        throw FunctionError(ErrorCode::INVALID_PARAMETER,
                            "Invalid funcName, not match regular expression or length exceeds "
                            "limit");
    }
}
void CheckFuncVersion(const std::string &version)
{
    const re2::RE2 pattern(VERSION_PATTERN_STRING);
    bool match = RE2::FullMatch(version, pattern);
    if (!match || version.length() > VERSION_LENGTH_LIMIT) {
        YRLOG_WARN("Invalid function version, {}", version);
        throw FunctionError(ErrorCode::INVALID_PARAMETER,
                            "Invalid func version, not match regular expression or length exceeds "
                            "limit");
    }
}
void CheckFuncAlias(const std::string &alias)
{
    const re2::RE2 pattern(ALIAS_PATTERN_STRING);
    bool match = RE2::FullMatch(alias, pattern);
    if (!match || alias.length() > ALIAS_LENGTH_LIMIT) {
        YRLOG_WARN("Invalid function alias, {}", alias);
        throw FunctionError(ErrorCode::INVALID_PARAMETER,
                            "Invalid func alias, not match regular expression or length exceeds "
                            "limit");
    }
}

std::string GetFunctionName(const std::string &funcString)
{
    if (funcString.empty()) {
        YRLOG_WARN("function name is invalid: {}", funcString);
        throw FunctionError(ErrorCode::INVALID_PARAMETER, "invalid funcName, expect not null");
    }
    std::string funcName = "";
    std::string version = DEFAULT_VERSION;

    if (funcString.find(FUNCTION_NAME_SEPERATOR) != std::string::npos) {
        std::vector<std::string> splitRet;
        YR::utility::Split(funcString, splitRet, FUNCTION_NAME_SEPERATOR);
        if (splitRet.size() != NAME_AND_VERSION_SIZE) {
            YRLOG_WARN("function name is invalid: {}", funcString);
            throw FunctionError(ErrorCode::INVALID_PARAMETER, "invalid funcName, not match regular expression");
        }
        funcName = splitRet[0];
        version = splitRet[1];

        CheckFuncName(funcName);
        if (version[0] == ALIAS_PREFIX) {
            CheckFuncAlias(version);
        } else {
            CheckFuncVersion(version);
        }
    } else {
        funcName = funcString;
        CheckFuncName(funcName);
    }
    return funcName + "/" + version;
}

Function::Function(Context &context, const std::string &funcName, const std::string &instanceName)
    : funcName_(funcName), instanceName_(instanceName)
{
    context_ = std::make_shared<ContextImpl>(*dynamic_cast<ContextImpl *>(&context));
}

Function::Function(Context &context, const std::string &funcName) : funcName_(funcName)
{
    context_ = std::make_shared<ContextImpl>(*dynamic_cast<ContextImpl *>(&context));
}

Function::Function(Context &context)
{
    context_ = std::make_shared<ContextImpl>(*dynamic_cast<ContextImpl *>(&context));
    funcName_ = context_->GetFunctionName();
}

ObjectRef Function::Invoke(const std::string &payload)
{
    std::string tenantId = context_->GetProjectID();
    std::string traceId = context_->GetTraceId();
    if (tenantId.empty()) {
        tenantId = DEFAULT_TENANT_ID;
    }
    std::stringstream ss;
    // default/0@faas@java:latest
    ss << tenantId << "/"
       << "0@" << context_->GetPackage() << "@" << GetFunctionName(funcName_);
    YR::Libruntime::FunctionMeta libFunctionMeta;
    libFunctionMeta.languageType = libruntime::LanguageType::Cpp;
    libFunctionMeta.functionId = ss.str();
    libFunctionMeta.apiType = libruntime::ApiType::Faas;
    std::vector<YR::Libruntime::InvokeArg> libArgs;
    std::vector<std::string> argsStr;
    argsStr.emplace_back("{}");
    nlohmann::json returnValJson;
    returnValJson["body"] = payload;
    argsStr.emplace_back(returnValJson.dump());
    for (auto &arg : argsStr) {
        YR::Libruntime::InvokeArg libArg;
        libArg.dataObj = std::make_shared<YR::Libruntime::DataObject>(0, arg.size());
        YR::WriteDataObject(static_cast<const void *>(arg.data()), libArg.dataObj, arg.size(), {});
        libArg.tenantId = tenantId;
        libArgs.emplace_back(std::move(libArg));
    }
    YR::Libruntime::InvokeOptions libOpts;
    libOpts.cpu = options_.cpu;
    libOpts.memory = options_.memory;
    libOpts.aliasParams = options_.aliasParams;
    libOpts.traceId = traceId;
    std::vector<YR::Libruntime::DataObject> returnObjs{{""}};
    auto err = YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->InvokeByFunctionName(
        libFunctionMeta, libArgs, libOpts, returnObjs);
    if (!err.OK()) {
        YRLOG_WARN("failed to invoke function {}, err: {}", funcName_, err.Msg());
        throw FunctionError(ErrorCode::INVALID_PARAMETER, err.Msg());
    }
    return ObjectRef(returnObjs[0].id, instanceID_);
}

Function &Function::Options(const InvokeOptions &opt)
{
    options_ = opt;
    return *this;
}

const std::string Function::GetObjectRef(ObjectRef &objectRef)
{
    return objectRef.Get();
}

void Function::GetInstance(const std::string &functionName, const std::string &instanceName)
{
    throw FunctionError(ErrorCode::INVALID_PARAMETER, "not support this function");
}

void Function::GetLocalInstance(const std::string &functionName, const std::string &instanceName)
{
    throw FunctionError(ErrorCode::INVALID_PARAMETER, "not support this function");
}

ObjectRef Function::Terminate()
{
    throw FunctionError(ErrorCode::INVALID_PARAMETER, "not support this function");
}

void Function::SaveState()
{
    throw FunctionError(ErrorCode::INVALID_PARAMETER, "not support this function");
}

const std::shared_ptr<Context> Function::GetContext() const
{
    return context_;
}

std::string Function::GetInstanceId() const
{
    throw FunctionError(ErrorCode::INVALID_PARAMETER, "not support this function");
}
}  // namespace Function