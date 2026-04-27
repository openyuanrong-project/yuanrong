/*
* Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
* Description: the ErrerInfo interface provided by yuanrong
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

namespace YR {
enum ErrorCode : int {
    ERR_NONE = 1,
    ERR_PARAM_INVALID = 1001,
    ERR_RESOURCE_NOT_ENOUGH = 1002,
    ERR_INSTANCE_NOT_FOUND = 1003,
    ERR_INSTANCE_EXITED = 1007,

    ERR_INSTANCE_SUSPEND = 1020,
    ERR_INSTANCE_BUSY = 1021,
    ERR_USER_FUNCTION_EXCEPTION = 2002,

    ERR_REQUEST_BETWEEN_RUNTIME_BUS = 3001,
    ERR_INNER_COMMUNICATION = 3002,
    ERR_BUS_DISCONNECTION = 3006,

    ERR_GET_OPERATION_FAILED = 4005,
    ERR_ROCKSDB_FAILED = 4201,
    ERR_SHARED_MEMORY_LIMITED = 4202,
    ERR_OPERATE_DISK_FAILED = 4203,
    ERR_INSUFFICIENT_DISK_SPACE = 4204,
    ERR_CONNECTION_FAILED = 4205,
    ERR_KEY_ALREADY_EXIST = 4206,
    ERR_DATASYSTEM_FAILED = 4299
};

enum ModuleCode : int {
    CORE_ = 10,
    RUNTIME_ = 20,
    RUNTIME_CREATE_ = 21,
    RUNTIME_INVOKE_ = 22,
    RUNTIME_KILL_ = 23,
    DATASYSTEM_ = 30
};

class ErrorInfo {
public:
    ErrorInfo() = default;
    ~ErrorInfo() = default;
    void SetErrorCode(ErrorCode errCode)
    {
        code = errCode;
    }

    ErrorCode Code() const
    {
        return code;
    }

    void SetModuleCode(ModuleCode errCode)
    {
        mCode = errCode;
    }

    ModuleCode MCode() const
    {
        return mCode;
    }

    void SetErrorMsg(const std::string &errMsg)
    {
        msg = errMsg;
    }

    std::string Msg() const
    {
        return msg;
    }

    void SetCodeAndMsg(ErrorCode errCode, const std::string &errMsg)
    {
        code = errCode;
        msg = errMsg;
    }

private:
    ErrorCode code = ErrorCode::ERR_NONE;
    ModuleCode mCode = ModuleCode::RUNTIME_;
    std::string msg;
};
} // namespace YR
