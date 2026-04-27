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

#include <exception>
#include <string>
#include "Constant.h"

namespace Function {
/*! @class FunctionError FunctionError.h "include/faas/FunctionError.h"
 *  @brief The exception class.
 */
class FunctionError : public std::exception {
public:
    /*!
    * @brief Constructor for FunctionError (initializes error code and message).
    * @param code Integer error code (maps to `ErrorCode` enum).
    * @param message Descriptive error message.
    * @snippet{trimleft} faas_example1.cpp functionerror_usage
    * @see Function::ErrorCode
     */
    FunctionError(int code, const std::string message) : errCode((ErrorCode)code), errMsg(message) {}
    
    /*!
    * @brief The default destructor.
     */
    virtual ~FunctionError() = default;

    /*!
     * @brief Get the char* type output of the exception's message.
     * @return The exception message.
     */
    const char *what() const noexcept override;

    /*!
     * @brief Get  the exception's Code.
     * @return The exception code.
     */
    ErrorCode GetErrorCode() const;

    /*!
     * @brief Get the exception's Message.
     * @return The exception message.
     */
    const std::string GetMessage() const;

    /*!
     * @brief Get the exception information and display it in JSON format.
     * @return The exception info.
     */
    const std::string GetJsonString() const;

private:
    ErrorCode errCode;
    std::string errMsg;
};
}  // namespace Function
