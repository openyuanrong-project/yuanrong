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

#pragma once

#include "Context.h"
#include "Runtime.h"

namespace Function {
class RegisterRuntimeHandler : public Function::RuntimeHandler {
public:
    RegisterRuntimeHandler()
    {
    }
    virtual ~RegisterRuntimeHandler()
    {
    }

    std::string HandleRequest(const std::string &request, Function::Context &context) override;

    void InitState(const std::string &request, Function::Context &context) override;

    void PreStop(Function::Context &context) override;

    void Initializer(Function::Context &context) override;

    void RegisterHandler(
        std::function<std::string(const std::string &request, Function::Context &context)> handleRequestFunc);

    void RegisterInitializerFunction(std::function<void(Function::Context &context)> initializerFunc);

    void RegisterPreStopFunction(std::function<void(Function::Context &context)> preStopFunc);

    void InitState(std::function<void(const std::string &request, Function::Context &context)> initStateFunc);

private:
    std::function<std::string(const std::string &, Function::Context &)> handleRequest;
    std::function<void(Function::Context &)> initializerFunction;
    std::function<void(Function::Context &)> preStopFunction;
    std::function<void(const std::string &, Function::Context &)> initStateFunction;
};
}  // namespace Function
