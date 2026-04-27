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

/*! @enum ErrorCode Constant.h "include/faas/Constant.h"
 *  @brief Error codes for runtime and user handler operations.
 */
enum ErrorCode {
    /*!
     * @brief Success.
     */
    OK = 0,
    /*!
     * @brief Generic error.
     */
    ERROR = 1,

    /*!
     * @brief Illegal access to runtime.
     */
    ILLEGAL_ACCESS = 4001,

    /*!
     * @brief User function exception.
     */
    FUNCTION_EXCEPTION = 4002,

    // not exposed
    USER_STATE_LARGE_ERROR = 4003,

    /*!
     * @brief Illegal return value from user handler.
     */
    ILLEGAL_RETURN = 4004,

    USER_STATE_UNDEFINED_ERROR = 4005,

    /*!
     * @brief User Initializer() function exception.
     */
    USER_INITIALIZATION_FUNCTION_EXCEPTION = 4009,

    /*!
     * @brief User load function exception.
     */
    USER_LOAD_FUNCTION_EXCEPTION = 4014,

    /*!
     * @brief Instance name not found.
     */
    NO_SUCH_INSTANCE_NAME_ERROR_CODE = 4026,

    /*!
     * @brief Invalid parameter.
     */
    INVALID_PARAMETER = 4040,

    /*!
     * @brief Illegal access to runtime.
     */
    NO_SUCH_STATE_ERROR_CODE = 4041,

    /*!
     * @brief Internal system error.
     */
    INTERNAL_ERROR = 110500,
};

// not exposed
// UserErrorMax is the maximum value of user errors
const int USER_ERROR_MAX = 10000;

// not exposed
const std::string SUCCESS_RESPONSE = "OK";
const std::string ILLEGAL_ACCESS_MESSAGE = "function entry cannot be found";
const std::string FUNCTION_EXCEPTION_MESSAGE = "function invocation exception: ";
const std::string USER_STATE_LARGE_ERROR_MESSAGE = "state content is too large";
const std::string ILLEGAL_RETURN_MESSAGE = "function return value is too large";
const std::string USER_STATE_UNDEFINED_ERROR_MESSAGE = "state is undefined";

// not exposed
const std::string INTERNAL_ERROR_MESSAGE = "internal system error";
const uint32_t MAX_USER_EXCEPTION_LENGTH = 1024;
const uint32_t MAX_USER_STATE_LENGTH = 3584 * 1024;
}  // namespace Function