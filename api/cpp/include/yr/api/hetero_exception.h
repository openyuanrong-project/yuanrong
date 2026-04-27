/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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
#include <vector>

namespace YR {

/**
  @brief the YR Exception class
 */
class HeteroException : public std::exception {
public:
    HeteroException() = default;
    HeteroException(int code, int moduleCode, const std::string &msg, std::vector<std::string> failedList = {});
    static HeteroException IncorrectFunctionUsageException(const std::string &exceptionMsg);
    static HeteroException InvalidParamException(const std::string &exceptionMsg);
    ~HeteroException() = default;
    const char *what() const noexcept override
    {
        return codeMsg_.c_str();
    }

    int Code() const
    {
        return code_;
    }

    int MCode() const
    {
        return mCode_;
    }

    std::string Msg() const
    {
        return msg_;
    }

    std::vector<std::string> FailedList()
    {
        return failedList_;
    }

private:
    int code_;
    int mCode_;
    std::string msg_;
    std::string codeMsg_ = "";
    std::vector<std::string> failedList_;
};
}  // namespace YR