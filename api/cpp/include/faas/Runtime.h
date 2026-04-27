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

#include <functional>
#include <memory>
#include <string>
#include "Constant.h"
#include "Context.h"
#include "RuntimeHandler.h"

namespace Function {

/*!
 * @class Runtime Runtime.h "include/faas/Runtime.h"
 *  @brief The Runtime class.
 */
class Runtime {
public:
    /*!
     * @brief The default constructor.
     */
    Runtime();

    /*!
     * @brief The default destructor.
     */
    virtual ~Runtime();

    /*!
     * @brief Register a handler function to the yuanrong cpp runtime.
     * @param handleRequestFunc The user-defined handler to be registered.
     * @snippet{trimleft} faas_example.cpp faasHandler
     */
    void RegisterHandler(
        std::function<std::string(const std::string &request, Function::Context &context)> handleRequestFunc);

    /*!
     * @brief Register an initial handler function to the yuanrong cpp runtime.
     * @param initializerFunc The user-defined initial handler to be registered.
     */
    void RegisterInitializerFunction(std::function<void(Function::Context &context)> initializerFunc);

    /*!
     * @brief Register a prestop handler function to the cpp yuanrong runtime.
     * @param preStopFunc The user-defined prestop handler to be registered.
     */
    void RegisterPreStopFunction(std::function<void(Function::Context &context)> preStopFunc);

    // not exposed
    void InitState(std::function<void(const std::string &request, Function::Context &context)> initStateFunc);

    /*!
     * @brief Start a yuanrong cpp runtime.
     * @param argc Count of command-line arguments.
     * @param argv Array of command-line argument strings (argv[0] = executable name).
     */
    void Start(int argc, char *argv[]);

private:
    std::function<std::string(const std::string &, Function::Context &)> handleRequest;
    std::function<void(Function::Context &)> initializerFunction;
    std::function<void(Function::Context &)> preStopFunction;
    std::function<void(const std::string &, Function::Context &)> initStateFunction;
    void InitRuntimeLogger() const;
    void ReleaseRuntimeLogger() const;
    void BuildRegisterRuntimeHandler() const;
};
}  // namespace Function
