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

#include "api/cpp/src/faas/register_runtime_handler.h"

#include "FunctionError.h"
#include "src/utility/logger/logger.h"

namespace Function {
void RegisterRuntimeHandler::RegisterHandler(
    std::function<std::string(const std::string &request, Function::Context &context)> handleRequestFunc)
{
    this->handleRequest = std::move(handleRequestFunc);
}

void RegisterRuntimeHandler::RegisterInitializerFunction(std::function<void(Function::Context &)> initializerFunc)
{
    this->initializerFunction = std::move(initializerFunc);
}

void RegisterRuntimeHandler::RegisterPreStopFunction(std::function<void(Function::Context &)> preStopFunc)
{
    this->preStopFunction = std::move(preStopFunc);
}

void RegisterRuntimeHandler::InitState(std::function<void(const std::string &, Function::Context &)> initStateFunc)
{
    this->initStateFunction = std::move(initStateFunc);
}

std::string RegisterRuntimeHandler::HandleRequest(const std::string &request, Context &context)
{
    if (this->handleRequest == nullptr) {
        YRLOG_WARN("The HandleRequest Function is null");
        throw FunctionError(ErrorCode::FUNCTION_EXCEPTION, "undefined HandleRequest()");
    }
    return this->handleRequest(request, context);
}

void RegisterRuntimeHandler::InitState(const std::string &request, Context &context)
{
    if (this->initStateFunction == nullptr) {
        YRLOG_WARN("The InitState Function is null");
        throw FunctionError(ErrorCode::FUNCTION_EXCEPTION, "undefined InitState()");
    }
    this->initStateFunction(request, context);
}

void RegisterRuntimeHandler::PreStop(Context &context)
{
    if (this->preStopFunction == nullptr) {
        return;
    }
    this->preStopFunction(context);
}

void RegisterRuntimeHandler::Initializer(Context &context)
{
    if (this->initializerFunction == nullptr) {
        return;
    }
    this->initializerFunction(context);
}
}  // namespace Function