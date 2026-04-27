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
#include "src/dto/invoke_options.h"

namespace YR {
enum StatusCode : int {
    ERR_OK = 0,
    ERR_RPC_SEND_REQUEST = 110501,
    ERR_START_RPC_SERVER,
    ERR_START_GW_CLIENT,
    ERR_DRIVER_DISCOVERY,
};

class Status {
public:
    Status() : code(StatusCode::ERR_OK), msg("") {}
    Status(StatusCode statusCode, std::string errMsg) : code(statusCode), msg(errMsg) {}
    int Code() const
    {
        return this->code;
    }

    bool OK() const
    {
        return this->code == 0;
    }

    std::string ErrorMessage() const
    {
        return this->msg;
    }
    void SetErrorCode(int code)
    {
        this->code = static_cast<YR::StatusCode>(code);
    }
    void SetErrorMessage(const std::string &message)
    {
        this->msg = message;
    }
    void SetErrorMessage(std::string &&message)
    {
        this->msg = std::move(message);
    }

private:
    StatusCode code;
    std::string msg;
};
}  // namespace YR