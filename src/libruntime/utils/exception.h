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
#include <iostream>
#include <string>
#include "src/libruntime/err_type.h"
#include "src/utility/logger/logger.h"
namespace YR {
namespace Libruntime {

#define CHECK_ERRORCODE_THROW_EXCEPTION(errInfo)                             \
    do {                                                                     \
        if (errInfo.Code() != YR::Libruntime::ErrorCode::ERR_OK) {           \
            std::cerr << errInfo.Msg() << std::endl;                         \
            throw Exception(errInfo.Code(), errInfo.MCode(), errInfo.Msg()); \
        }                                                                    \
    } while (0)

#define THROW_ERRCODE_EXCEPTION(code, moduleCode, msg) (throw Exception(code, moduleCode, msg))

#define STDERR_AND_THROW_EXCEPTION(code, moduleCode, msg)                                \
    do {                                                                                 \
        std::stringstream ss;                                                            \
        ss << "ErrCode: " << code << ", ModuleCode: " << moduleCode << ", Msg: " << msg; \
        YRLOG_ERROR(ss.str());                                                           \
        std::cerr << ss.str() << std::endl;                                              \
        throw Exception(code, moduleCode, msg);                                          \
    } while (0)

class Exception : public std::exception {
public:
    explicit Exception()
    {
        code = YR::Libruntime::ErrorCode::ERR_OK;
        codeMsg = "";
    }
    explicit Exception(const std::string &msg) : msg(msg)
    {
        code = YR::Libruntime::ErrorCode::ERR_OK;
        codeMsg = "ErrCode: " + std::to_string(code) + ", ModuleCode: " + std::to_string(mCode) + ", ErrMsg: " + msg;
    }
    explicit Exception(const int &code, const std::string &msg) : code(code), msg(msg)
    {
        codeMsg = "ErrCode: " + std::to_string(code) + ", ModuleCode: " + std::to_string(mCode) + ", ErrMsg: " + msg;
    }
    explicit Exception(const int &code, const int &moduleCode, const std::string &msg)
        : code(code), mCode(moduleCode), msg(msg)
    {
        codeMsg =
            "ErrCode: " + std::to_string(code) + ", ModuleCode: " + std::to_string(moduleCode) + ", ErrMsg: " + msg;
    }
    explicit Exception(const YR::Libruntime::ErrorCode &eCode, const YR::Libruntime::ModuleCode &moduleCode,
                       const std::string &msg)
        : msg(msg)
    {
        code = static_cast<int>(eCode);
        mCode = static_cast<int>(moduleCode);
        codeMsg =
            "ErrCode: " + std::to_string(eCode) + ", ModuleCode: " + std::to_string(moduleCode) + ", ErrMsg: " + msg;
    }

    ~Exception() = default;
    const char *what() const noexcept override
    {
        return codeMsg.c_str();
    }

    int Code() const
    {
        return code;
    }
    int MCode() const
    {
        return mCode;
    }
    std::string Msg() const
    {
        return msg;
    }

private:
    int code;
    int mCode{static_cast<int>(YR::Libruntime::ModuleCode::RUNTIME)};
    std::string msg;
    std::string codeMsg = "";
};
}  // namespace Libruntime
}  // namespace YR