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
#include <exception>
#include <string>

namespace YR {

/**
  @brief The openYuanRong Exception class.
 */
class Exception : public std::exception {
public:
    Exception();
    explicit Exception(const std::string &msg);
    Exception(int code, const std::string &msg);
    Exception(int code, int moduleCode, const std::string &msg);

    static Exception RegisterRecoverFunctionException();
    static Exception RegisterShutdownFunctionException();
    static Exception RegisterFunctionException(const std::string &funcName);
    static Exception DeserializeException(const std::string &exceptionMsg);
    static Exception InvalidParamException(const std::string &exceptionMsg);
    static Exception GetException(const std::string &exceptionMsg);
    static Exception InnerSystemException(const std::string &exceptionMsg);
    static Exception UserCodeException(const std::string &exceptionMsg);
    static Exception InstanceIdEmptyException(const std::string &exceptionMsg);
    static Exception IncorrectInvokeUsageException(const std::string &exceptionMsg);
    static Exception IncorrectFunctionUsageException(const std::string &exceptionMsg);
    ~Exception() = default;
    /**
     * @brief Obtain detailed information about the exception, including errCode, moduleCode and errMsg.
     *
     * @return The detailed information about the exception.
     */
    const char *what() const noexcept override
    {
        return codeMsg.c_str();
    }

    /**
     * @brief Get the error code.
     *
     * @return The error code.
     */
    int Code() const
    {
        return code;
    }

    /**
     * @brief Get the module code.
     *
     * @return The module code.
     */
    int MCode() const
    {
        return mCode;
    }

    /**
     * @brief Get the error message.
     *
     * @return The error message.
     */
    std::string Msg() const
    {
        return msg;
    }

private:
    /*!
       @brief The error code.
    */
    int code;
    /*!
       @brief The module code.
    */
    int mCode;
    /*!
       @brief The error message.
    */
    std::string msg;
    /*!
       @brief The detailed information about the exception.
    */
    std::string codeMsg = "";
};
}  // namespace YR