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
#include "Context.h"

namespace Function {
/*!
 *  @brief Customer implement this interface to handle runtime message.
 */
class RuntimeHandler {
public:
    RuntimeHandler() = default;
    virtual ~RuntimeHandler() = default;

    /*!
     * @brief Handle incoming runtime requests and return a response.
     * @param request Incoming request string (typically in JSON/string format).
     * @param context Runtime context object (contains trace ID, invoke ID, log utilities).
     * @return std::string Response string to the request.
     */
    virtual std::string HandleRequest(const std::string &request, Context &context) = 0;

    // not exposed
    virtual void InitState(const std::string &request, Context &context) = 0;

    /*!
     * @brief Perform pre-stop cleanup operations before runtime shutdown.
     * This method is triggered when the runtime is about to shut down (e.g., receiving SIGINT/SIGTERM).
     * It should release resources (e.g., close file handles, database connections) and log
     * cleanup status for troubleshooting.
     * @param context The runtime context object (for logging shutdown status).
     */
    virtual void PreStop(Context &context) = 0;

    /*!
     * @brief Perform global initialization for the handler.
     * This method is called once during runtime startup (before Start() completes) to
     * initialize global resources (e.g., load configuration files, initialize loggers).
     * It is executed before any requests are processed.
     * @param context Runtime context object (contains runtime configuration)
     */
    virtual void Initializer(Context &context) = 0;
};
}  // namespace Function

