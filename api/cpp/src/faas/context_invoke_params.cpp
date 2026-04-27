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

#include "api/cpp/src/faas/context_invoke_params.h"
#include "FunctionLogger.h"

namespace Function {

ContextInvokeParams::ContextInvokeParams(const std::unordered_map<std::string, std::string> &params)
{
    auto aliasIt = params.find("X-Invoke-Alias");
    if (aliasIt != params.end()) {
        this->alias = aliasIt->second;
    }
    auto tokenIt = params.find("X-Auth-Token");
    if (tokenIt != params.end()) {
        this->token = tokenIt->second;
    }
    auto securitySecretKeyIt = params.find("X-Security-Secret-Key");
    if (securitySecretKeyIt != params.end()) {
        this->securitySecretKey = securitySecretKeyIt->second;
    }
    auto securityAccessKeyIt = params.find("X-Security-Access-Key");
    if (securityAccessKeyIt != params.end()) {
        this->securityAccessKey = securityAccessKeyIt->second;
    }
    auto accessKeyIt = params.find("X-Access-Key");
    if (accessKeyIt != params.end()) {
        this->accessKey = accessKeyIt->second;
    }
    auto secretKeyIt = params.find("X-Secret-Key");
    if (secretKeyIt != params.end()) {
        this->secretKey = secretKeyIt->second;
    }
    auto securityTokenIt = params.find("X-Security-Token");
    if (securityTokenIt != params.end()) {
        this->securityToken = securityTokenIt->second;
    }
}

const std::string ContextInvokeParams::GetAccessKey() const
{
    return accessKey;
}

const std::string ContextInvokeParams::GetSecretKey() const
{
    return secretKey;
}

const std::string ContextInvokeParams::GetSecurityAccessKey() const
{
    return securityAccessKey;
}

const std::string ContextInvokeParams::GetSecuritySecretKey() const
{
    return securitySecretKey;
}

const std::string ContextInvokeParams::GetRequestId() const
{
    return requestID;
}

const std::string ContextInvokeParams::GetTraceId() const
{
    return GetRequestId();
}

const std::string ContextInvokeParams::GetInvokeId() const
{
    return invokeID;
}

const std::string ContextInvokeParams::GetToken() const
{
    return token;
}

const std::string ContextInvokeParams::GetAlias() const
{
    return alias;
}

void ContextInvokeParams::SetInvokeID(const std::string &id)
{
    this->invokeID = id;
}
void ContextInvokeParams::SetRequestID(const std::string &id)
{
    this->requestID = id;
}

}  // namespace Function