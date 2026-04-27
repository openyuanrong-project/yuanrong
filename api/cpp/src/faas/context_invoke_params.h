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
#include <unordered_map>

namespace Function {

class ContextInvokeParams {
public:
    ContextInvokeParams() = default;
    explicit ContextInvokeParams(const std::unordered_map<std::string, std::string> &params);
    ~ContextInvokeParams() = default;
    const std::string GetAccessKey() const;

    const std::string GetSecretKey() const;

    const std::string GetSecurityAccessKey() const;

    const std::string GetSecuritySecretKey() const;

    const std::string GetRequestId() const;

    const std::string GetTraceId() const;

    const std::string GetInvokeId() const;

    const std::string GetToken() const;

    const std::string GetAlias() const;

    void SetInvokeID(const std::string &id);
    void SetRequestID(const std::string &id);

private:
    std::string accessKey = "";
    std::string secretKey = "";
    std::string securityAccessKey = "";
    std::string securitySecretKey = "";
    std::string requestID = "";
    std::string invokeID = "";
    std::string token = "";
    std::string securityToken = "";
    std::string alias = "";
};
}  // namespace Function
