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

#include "api/cpp/src/faas/faas_executor.h"

#include "json.hpp"

#include "FunctionError.h"
#include "api/cpp/src/faas/context_impl.h"
#include "api/cpp/src/utils/utils.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/utility/json_utility.h"
#include "src/utility/logger/logger.h"
namespace Function {

using YR::utility::JsonGet;

const int META_DATA_INDEX = 0;
const int USER_EVENT_INDEX = 1;
const int CREATE_PARAM_INDEX = 1;
const int RUNTIME_MAX_RESP_BODY_SIZE = 6 * 1024 * 1024;

const char *const PRE_STOP_INVOKE_ID = "preStop";
const char *const INITIALIZER_INVOKE_ID = "initializer";
const char *const INSTANCE_LABEL_KEY = "instanceLabel";

const std::string LD_LIBRARY_PATH = "LD_LIBRARY_PATH";

std::string MakeErrorResult(int code, std::string message)
{
    nlohmann::json returnValJson;
    returnValJson["innerCode"] = std::to_string(code);
    returnValJson["body"] = message;
    return returnValJson.dump();
}

static std::unique_ptr<RuntimeHandler> runtimeHandler;

void SetRuntimeHandler(std::unique_ptr<RuntimeHandler> r)
{
    runtimeHandler = std::move(r);
}

YR::Libruntime::ErrorInfo FaasExecutor::LoadFunctions(const std::vector<std::string> &paths)
{
    return {};
}

YR::Libruntime::ErrorInfo FaasExecutor::ExecuteFunction(
    const YR::Libruntime::FunctionMeta &function, const libruntime::InvokeType invokeType,
    const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects)
{
    std::string result;
    if (invokeType == libruntime::InvokeType::CreateInstance ||
        invokeType == libruntime::InvokeType::CreateInstanceStateless) {
        auto initErr = FaasInitHandler(rawArgs);
        if (!initErr.OK()) {
            return initErr;
        }
    } else if (invokeType == libruntime::InvokeType::InvokeFunctionStateless ||
               invokeType == libruntime::InvokeType::InvokeFunction) {
        result = FaasCallHandler(rawArgs);
    }

    YRLOG_DEBUG("finish to execute faas function, result: {}", result);
    if (result.empty()) {
        return {};
    }

    uint64_t totalNativeBufferSize = 0;
    auto err = YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->AllocReturnObject(
        returnObjects[0], 0, result.size(), {}, totalNativeBufferSize);
    if (!err.OK()) {
        return err;
    }
    return YR::WriteDataObject(result.data(), returnObjects[0], result.size(), {});
}

YR::Libruntime::ErrorInfo FaasExecutor::Checkpoint(const std::string &instanceID,
                                                   std::shared_ptr<YR::Libruntime::Buffer> &data)
{
    return {};
}

YR::Libruntime::ErrorInfo FaasExecutor::Recover(std::shared_ptr<YR::Libruntime::Buffer> data)
{
    return {};
}

YR::Libruntime::ErrorInfo FaasExecutor::ExecuteShutdownFunction(uint64_t gracePeriodSecond)
{
    if (contextInvokeParams_ == nullptr || contextEnv_ == nullptr) {
        return YR::Libruntime::ErrorInfo(
            YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION, YR::Libruntime::ModuleCode::RUNTIME,
            FunctionError(ErrorCode::FUNCTION_EXCEPTION, "can not call prestop before initialize").GetJsonString());
    }
    contextInvokeParams_->SetInvokeID(PRE_STOP_INVOKE_ID);
    auto context = ContextImpl(contextInvokeParams_, contextEnv_);
    try {
        runtimeHandler->PreStop(context);
    } catch (FunctionError &functionError) {
        YRLOG_ERROR("execute prestop failed, {}", functionError.GetJsonString());
        return YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                                         YR::Libruntime::ModuleCode::RUNTIME, functionError.GetJsonString());
    } catch (std::exception &e) {
        YRLOG_ERROR("execute prestop failed, {}", e.what());
        return YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                                         YR::Libruntime::ModuleCode::RUNTIME,
                                         FunctionError(ErrorCode::FUNCTION_EXCEPTION, e.what()).GetJsonString());
    }
    return {};
}

YR::Libruntime::ErrorInfo FaasExecutor::Signal(int sigNo, std::shared_ptr<YR::Libruntime::Buffer> payload)
{
    return {};
}

void FaasExecutor::ParseContextMeta(const std::string &contextMeta)
{
    auto contextMetaJson = nlohmann::json::parse(contextMeta);
    std::string emptyStr = "";
    if (contextMetaJson.contains("funcMetaData")) {
        auto funcMetaData = contextMetaJson.at("funcMetaData");
        contextEnv_->SetProjectID(JsonGet(funcMetaData, "tenantId", emptyStr));
        contextEnv_->SetFuncPackage(JsonGet(funcMetaData, "service", emptyStr));
        contextEnv_->SetFunctionName(JsonGet(funcMetaData, "func_name", emptyStr));
        contextEnv_->SetVersion(JsonGet(funcMetaData, "version", emptyStr));
    }
    if (contextMetaJson.contains("resourceMetaData")) {
        auto resourceMetaData = contextMetaJson.at("resourceMetaData");
        contextEnv_->SetCPUNumber(JsonGet(resourceMetaData, "cpu", 0));
        contextEnv_->SetMemorySize(JsonGet(resourceMetaData, "memory", 0));
    }
}

void FaasExecutor::ParseCreateParams(const std::string &createParam)
{
    auto createParamJson = nlohmann::json::parse(createParam);
    std::string emptyStr = "";
    if (createParamJson.contains(std::string(INSTANCE_LABEL_KEY))) {
        contextEnv_->SetInstanceLabel(JsonGet(createParamJson, std::string(INSTANCE_LABEL_KEY), emptyStr));
    }
}

void FaasExecutor::ParseDelegateDecrypt(const std::string &delegateDecrypt)
{
    if (delegateDecrypt.empty()) {
        return;
    }
    auto delegateDecryptJson = nlohmann::json::parse(delegateDecrypt);
    std::unordered_map<std::string, std::string> userData;
    if (delegateDecryptJson.contains("environment")) {
        std::string envStr = delegateDecryptJson.at("environment");
        if (!envStr.empty()) {
            auto env = nlohmann::json::parse(envStr);
            for (auto &[key, value] : env.items()) {
                YRLOG_DEBUG("setenv {}={}", key, value.dump());
                if (key == LD_LIBRARY_PATH) {
                    auto newPath = YR::GetEnv(key) + ":" + value.dump();
                    userData[key] = newPath;
                    YR::SetEnv(key, newPath);
                } else {
                    userData[key] = value;
                    YR::SetEnv(key, value);
                }
            }
        }
    }
    if (delegateDecryptJson.contains("encrypted_user_data")) {
        std::string userDataStr = delegateDecryptJson.at("encrypted_user_data");
        if (!userDataStr.empty()) {
            auto userDataJson = nlohmann::json::parse(userDataStr);
            for (auto &[key, value] : userDataJson.items()) {
                YRLOG_DEBUG("set user data {}={}", key, value.dump());
                if (key == LD_LIBRARY_PATH) {
                    auto newPath = YR::GetEnv(key) + ":" + value.dump();
                    userData[key] = newPath;
                } else {
                    userData[key] = value;
                }
            }
        }
    }

    contextEnv_->SetUserData(userData);
}

void SetTraceID(nlohmann::json callReqJson, std::shared_ptr<ContextInvokeParams> &contextInvokeParams)
{
    std::string traceId = "";
    std::string headerKey = "header";
    if (callReqJson.contains(headerKey)) {
        auto header = callReqJson.at(headerKey);
        std::string traceIdKey = "X-Trace-Id";
        if (header.contains(traceIdKey)) {
            traceId = header.at(traceIdKey);
        }
    }
    contextInvokeParams->SetRequestID(traceId);
}

YR::Libruntime::ErrorInfo FaasExecutor::FaasInitHandler(
    const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs)
{
    YRLOG_DEBUG("start to call faas init handler, arg size: {} ", rawArgs.size());
    contextEnv_ = std::make_shared<ContextEnv>();
    std::unordered_map<std::string, std::string> params;
    contextInvokeParams_ = std::make_shared<ContextInvokeParams>(params);
    auto contextMeta = std::string(static_cast<const char *>(rawArgs[META_DATA_INDEX]->data->ImmutableData()),
                                   rawArgs[META_DATA_INDEX]->data->GetSize());
    if (contextMeta.empty()) {
        YRLOG_WARN("failed to parse context with empty string.");
        return YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                         "failed to parse context with empty string");
    }
    try {
        ParseContextMeta(contextMeta);
        auto delegateDecrypt = YR::GetEnv("ENV_DELEGATE_DECRYPT");
        ParseDelegateDecrypt(delegateDecrypt);
    } catch (nlohmann::detail::exception &e) {
        YRLOG_WARN("failed to parse context with exception: {} ", e.what());
        return YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                         std::string("faild to parse context with exception: ") + e.what());
    }
    if (rawArgs.size() > CREATE_PARAM_INDEX) {
        auto createParam = std::string(static_cast<const char *>(rawArgs[CREATE_PARAM_INDEX]->data->ImmutableData()),
                                       rawArgs[CREATE_PARAM_INDEX]->data->GetSize());
        if (!createParam.empty()) {
            try {
                ParseCreateParams(createParam);
            } catch (nlohmann::detail::exception &e) {
                YRLOG_WARN("failed to parse create param with exception: {} ", e.what());
            }
        }
    }
    contextInvokeParams_->SetInvokeID(INITIALIZER_INVOKE_ID);
    auto context = ContextImpl(contextInvokeParams_, contextEnv_);
    try {
        runtimeHandler->Initializer(context);
    } catch (FunctionError &functionError) {
        YRLOG_WARN("failed to call initializer: {} ", functionError.GetJsonString());
        return YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                                         functionError.GetJsonString());
    } catch (std::exception &e) {
        YRLOG_WARN("failed to call initializer:: {} ", e.what());
        return YR::Libruntime::ErrorInfo(
            YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
            std::string("failed to call initializer function with exception: ") + e.what());
    }
    YRLOG_DEBUG("success to call faas init handler");
    return {};
}

std::string FaasExecutor::FaasCallHandler(const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs)
{
    YRLOG_DEBUG("start to call faas call handler, arg size: {} ", rawArgs.size());
    auto callReq = std::string(static_cast<const char *>(rawArgs[USER_EVENT_INDEX]->data->ImmutableData()),
                               rawArgs[USER_EVENT_INDEX]->data->GetSize());
    if (callReq.empty()) {
        return MakeErrorResult(ErrorCode::FUNCTION_EXCEPTION, "call req is empty");
    }
    auto callReqJson = nlohmann::json::parse(callReq);
    std::string bodyKey = "body";
    if (!callReqJson.contains(bodyKey)) {
        return MakeErrorResult(ErrorCode::FUNCTION_EXCEPTION, "can not find body");
    }
    std::string eventStr;
    auto event = callReqJson.at(bodyKey);
    if (event.type() != nlohmann::json::value_t::string) {
        eventStr = event.dump();
    } else {
        eventStr = event;
    }
    if (contextInvokeParams_ == nullptr || contextEnv_ == nullptr) {
        return MakeErrorResult(ErrorCode::FUNCTION_EXCEPTION, "can not call handlerequest before initialize");
    }
    SetTraceID(callReqJson, contextInvokeParams_);
    auto context = ContextImpl(contextInvokeParams_, contextEnv_);
    auto errCode = ErrorCode::OK;
    std::string errMsg = "";
    nlohmann::json returnValJson;
    std::string result;
    try {
        result = runtimeHandler->HandleRequest(eventStr, context);
    } catch (FunctionError &functionError) {
        errCode = functionError.GetErrorCode();
        errMsg = functionError.GetMessage();
    } catch (std::exception &e) {
        errCode = ErrorCode::FUNCTION_EXCEPTION;
        errMsg = std::string("failed to call function with exception: ") + e.what();
    }
    returnValJson["innerCode"] = std::to_string(errCode);
    if (errCode != ErrorCode::OK) {
        YRLOG_WARN("failed to call user function, err {}", errMsg);
        returnValJson["body"] = errMsg;
        return returnValJson.dump();
    }

    if (result.size() > RUNTIME_MAX_RESP_BODY_SIZE) {
        std::stringstream ss;
        ss << "function result size: " << result.size() << ", exceed limit(" << RUNTIME_MAX_RESP_BODY_SIZE << ")";
        YRLOG_WARN(ss.str());
        return MakeErrorResult(ErrorCode::ILLEGAL_RETURN, ss.str());
    }
    returnValJson["body"] = result;
    returnValJson["billingDuration"] = "this is billing duration TODO";
    returnValJson["logResult"] = "this is user log TODO";
    returnValJson["invokerSummary"] = "this is summary TODO";
    return returnValJson.dump();
}

}  // namespace Function