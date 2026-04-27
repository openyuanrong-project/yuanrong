/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#include <map>
#include <memory>
#include <string>
#include <unordered_map>
#include <vector>
#include "Constant.h"
#include "Context.h"
#include "ObjectRef.h"

namespace Function {

/*! @struct InvokeOptions Function.h "include/faas/Function.h"
 *  @brief The struct is used to provide parameters for configuring function calls.
 */
struct InvokeOptions {
    /*!
     * @brief CPU size (unit: mi/millicores).
     */
    int cpu = 0;

    /*!
     * @brief Memory size (unit: MB/megabytes).
     */
    int memory = 0;

    /*!
     * @brief Alias params.
     */
    std::unordered_map<std::string, std::string> aliasParams;
};

/*! @class Function Function.h "include/faas/Function.h"
 *  @brief The class provides the functionality for inter-function calls.
 */
class Function {
public:
    /*!
     * @brief The constructor of the Function class, where the callee function is the function itself.
     * @param context A parameter of the `Context` object.
     */
    explicit Function(Context &context);

    /*!
     * @brief The constructor of the Function class.
     * @param context A parameter of the `Context` object.
     * @param funcName Then callee function name.
     * @snippet{trimleft} faas_caller_example.cpp faasCallerHandler
     */
    explicit Function(Context &context, const std::string &funcName);

    // not exposed
    explicit Function(Context &context, const std::string &funcName, const std::string &instanceName);

    /*!
     * @brief Default destructor.
     */
    virtual ~Function() = default;

    /*!
     * @brief Disable the copy constructor.
     */
    Function(const Function &) = delete;

    /*!
     * @brief Disable the copy assignment operator.
     */
    Function &operator=(const Function &) = delete;

    /*!
     * @brief Invoke the callee function.
     * @param payload Request parameters, which are required to be in JSON string format.
     * @return The response result of type ObjectRef.
     * @throws FunctionError Exceptions thrown when executing the request.
     */
    ObjectRef Invoke(const std::string &payload);

    /*!
     * @brief Set the invoke Options.
     * @param opt The reference of the `InvokeOptions` type.
     */
    Function &Options(const InvokeOptions &opt);

    /*!
     * @brief Get the invoke result.
     * @param objectRef The reference of the `ObjectRef` type.
     * @return The invoke result.
     */
    const std::string GetObjectRef(ObjectRef &objectRef);

    // not exposed
    void GetInstance(const std::string &functionName, const std::string &instanceName);

    // not exposed
    void GetLocalInstance(const std::string &functionName, const std::string &instanceName);

    // not exposed
    ObjectRef Terminate();

    // not exposed
    void SaveState();

    /*!
     * @brief Get the configured Context parameters.
     * @return The Configured Context.
     * @throws FunctionError Exceptions thrown when get the invoke result.
     */
    const std::shared_ptr<Context> GetContext() const;

    // not exposed
    std::string GetInstanceId() const;

private:
    std::shared_ptr<Context> context_;
    std::string funcName_;
    std::string instanceName_;
    std::string instanceID_;
    InvokeOptions options_;
};
}  // namespace Function
