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

namespace Function {
/*!
 * @class FunctionLogger FunctionLogger.h "include/faas/FunctionLogger.h"
 *  @brief The log class.
 */
class FunctionLogger {
public:
    // not exposed
    FunctionLogger(){};
    // not exposed
    virtual ~FunctionLogger(){};
    // not exposed
    FunctionLogger(const std::string &traceId, const std::string &invokeId) : traceId(traceId), invokeId(invokeId)
    {
    }

    /*!
     * @brief Set the log persistence level. The available levels are `DEBUG`, `INFO`, `WARN`, and `ERROR`,
     * with `INFO` level as the default.
     * @param level The log level parameter.
     * @snippet{trimleft} faas_example1.cpp logger_usage
     */
    void setLevel(const std::string &level);

    /*!
     * @brief Print logs at the `INFO` level.
     * @param message Formatted string for the log.
     * @param ... Variable argument list, which corresponds to the placeholders in the formatted string one by one.
     */
    void Info(std::string message, ...);

    /*!
     * @brief Print logs at the `WARN` level.
     * @param message Formatted string for the log.
     * @param ... Variable argument list, which corresponds to the placeholders in the formatted string one by one.
     */
    void Warn(std::string message, ...);

    /*!
     * @brief Print logs at the `DEBUG` level.
     * @param message Formatted string for the log.
     * @param ... Variable argument list, which corresponds to the placeholders in the formatted string one by one.
     */
    void Debug(std::string message, ...);

    /*!
     * @brief Print logs at the `ERROR` level.
     * @param message Formatted string for the log.
     * @param ... Variable argument list, which corresponds to the placeholders in the formatted string one by one.
     */
    void Error(std::string message, ...);

    // not exposed
    void SetInvokeID(std::string invokeID);

    // not exposed
    void SetTraceID(std::string traceID);

private:
    std::string traceId;
    std::string invokeId;
    std::string logLevel = "INFO";
    void Log(const std::string &level, const std::string &logMessage);
    bool sendEmptyLog(const std::string &message, const std::string &level);
};
}  // namespace Function
