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

#include <string>
#include "FunctionLogger.h"

namespace Function {

/*! @class Context Context.h "include/faas/Context.h"
 *  @brief This is the class of context parameters passed when executing a function, and this object contains
 *  information such as the call, function and environment.
 */
class Context {
public:
    /*!
     * @brief The default Constructor.
     * @snippet{trimleft} faas_example1.cpp context_usage
     */
    Context() = default;

    /*!
     * @brief The default Destructor.
     */
    virtual ~Context() = default;

    // not exposed
    virtual const std::string GetAccessKey() const = 0;

    // not exposed
    virtual const std::string GetSecretKey() const = 0;

    // not exposed
    virtual const std::string GetSecurityAccessKey() const = 0;

    // not exposed
    virtual const std::string GetSecuritySecretKey() const = 0;

    // not exposed
    virtual const std::string GetToken() const = 0;

    // not exposed
    virtual const std::string GetAlias() const = 0;

    /*!
     * @brief Get the traceId of the current execution request.
     * @return The traceId.
     */
    virtual const std::string GetTraceId() const = 0;

    // not exposed
    virtual const std::string GetInvokeId() const = 0;

    /*!
     * @brief Get the reference of to the instance of the `FunctionLogger` class provided by the context, which can
     * be used to print logs. For details, refer to the usage instructions of `FunctionLogger`.
     * @return An instance of FunctionLogger class.
     */
    virtual const FunctionLogger &GetLogger() = 0;

    // not exposed
    virtual const std::string GetState() const = 0;

    // not exposed
    virtual const std::string GetInstanceId() const = 0;

    /*!
     * @brief Get the instance label.
     * @return An instance label.
     */
    virtual const std::string GetInstanceLabel() const = 0;

    // not exposed
    virtual void SetState(const std::string &state) = 0;

    // not exposed
    virtual const std::string GetInvokeProperty() const = 0;

    /*!
     * @brief Get the requestId of the current execution request.
     * @return The reuqestId.
     */
    virtual const std::string GetRequestID() const = 0;

    /*!
     * @brief Get the environment variables set by the user.
     * @param key The key of the environment variable.
     * @return The value of the environment variable.
     */
    virtual const std::string GetUserData(std::string key) const = 0;

    /*!
     * @brief Get the name of the function.
     * @return The function name.
     */
    virtual const std::string GetFunctionName() const = 0;

    // not exposed
    virtual int GetRemainingTimeInMilliSeconds() const = 0;

    // not exposed
    virtual int GetRunningTimeInSeconds() const = 0;

    /*!
     * @brief Get the version of the function.
     * @return The function version.
     */
    virtual const std::string GetVersion() const = 0;

    /*!
     * @brief Get memory size (unit: MB/megabytes).
     * @return The memory size.
     */
    virtual int GetMemorySize() const = 0;

    /*!
     * @brief Get cpu size (unit: mi/millicores).
     * @return The cpu size.
     */
    virtual int GetCPUNumber() const = 0;

    /*!
     * @brief Get tenantId of the function.
     * @return The tenantId.
     */
    virtual const std::string GetProjectID() const = 0;

    /*!
     * @brief Get service name of the function.
     * @return The service.
     */
    virtual const std::string GetPackage() const = 0;
};

}  // namespace Function
