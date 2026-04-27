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

#include "RuntimeHandler.h"
#include "api/cpp/src/executor/executor.h"
#include "api/cpp/src/faas/context_env.h"
#include "api/cpp/src/faas/context_invoke_params.h"

namespace Function {
void SetRuntimeHandler(std::unique_ptr<RuntimeHandler> r);
class FaasExecutor : public YR::internal::Executor {
public:
    FaasExecutor() = default;
    ~FaasExecutor() = default;

    YR::Libruntime::ErrorInfo LoadFunctions(const std::vector<std::string> &paths) override;

    YR::Libruntime::ErrorInfo ExecuteFunction(
        const YR::Libruntime::FunctionMeta &function, const libruntime::InvokeType invokeType,
        const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
        std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects) override;

    YR::Libruntime::ErrorInfo Checkpoint(const std::string &instanceID,
                                         std::shared_ptr<YR::Libruntime::Buffer> &data) override;

    YR::Libruntime::ErrorInfo Recover(std::shared_ptr<YR::Libruntime::Buffer> data) override;

    YR::Libruntime::ErrorInfo ExecuteShutdownFunction(uint64_t gracePeriodSecond) override;

    YR::Libruntime::ErrorInfo Signal(int sigNo, std::shared_ptr<YR::Libruntime::Buffer> payload) override;

private:
    // The reason for returning ErrorInfo here is to convey the failure of initialization to the functionsystem.
    // The failed instance needs to be reclaimed by the functionsystem.
    YR::Libruntime::ErrorInfo FaasInitHandler(const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs);
    std::string FaasCallHandler(const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs);
    void ParseContextMeta(const std::string &contextMeta);
    void ParseCreateParams(const std::string &createParam);
    void ParseDelegateDecrypt(const std::string &delegateDecrypt);

    std::shared_ptr<ContextEnv> contextEnv_;
    std::shared_ptr<ContextInvokeParams> contextInvokeParams_;
};

}  // namespace Function