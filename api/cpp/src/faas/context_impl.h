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

#include <functional>
#include <memory>
#include <mutex>
#include <string>
#include <unordered_map>

#include "Context.h"
#include "api/cpp/src/faas/context_env.h"
#include "api/cpp/src/faas/context_invoke_params.h"

namespace Function {
class ContextImpl : public Context {
public:
    ContextImpl() = default;
    explicit ContextImpl(std::shared_ptr<ContextInvokeParams> &invokeParams, std::shared_ptr<ContextEnv> &contextEnv);
    ContextImpl(const ContextImpl &contextImpl);
    ContextImpl &operator=(const ContextImpl &contextImpl);
    ~ContextImpl() = default;
    const std::string GetAccessKey() const override;

    const std::string GetSecretKey() const override;

    const std::string GetSecurityAccessKey() const override;

    const std::string GetSecuritySecretKey() const override;

    const std::string GetToken() const override;

    const std::string GetAlias() const override;

    const std::string GetTraceId() const override;

    const std::string GetInvokeId() const override;

    const FunctionLogger &GetLogger() override;

    const std::string GetInstanceId() const override;

    const std::string GetInstanceLabel() const override;

    void SetStateId(const std::string &stateId);

    const std::string GetState() const override;

    void SetState(const std::string &state) override;

    const std::string GetInvokeProperty() const override;

    const std::string GetRequestID() const override;

    const std::string GetUserData(std::string key) const override;

    const std::string GetFunctionName() const override;

    int GetRemainingTimeInMilliSeconds() const override;

    int GetRunningTimeInSeconds() const override;

    const std::string GetVersion() const override;

    int GetMemorySize() const override;

    int GetCPUNumber() const override;

    const std::string GetProjectID() const override;

    const std::string GetPackage() const override;

    void SetInvokeProperty(const std::string &property);

    void SetFuncStartTime(const long startTime);

private:
    FunctionLogger logger_;
    std::string stateId_;
    std::string state_;
    std::string property_;
    long funcStartTime_;

    std::shared_ptr<ContextInvokeParams> contextInvokeParams_;
    std::shared_ptr<ContextEnv> contextEnv_;

    mutable std::mutex stateMtx_;
};
}  // namespace Function
