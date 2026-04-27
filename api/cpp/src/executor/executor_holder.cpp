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

#include "api/cpp/src/executor/executor_holder.h"
#include "api/cpp/src/executor/function_executor.h"
#include "api/cpp/src/executor/executor.h"

namespace YR {
namespace internal {
ExecutorHolder &ExecutorHolder::Singleton()
{
    static ExecutorHolder singleton{};
    return singleton;
}

void ExecutorHolder::SetExecutor(std::shared_ptr<Executor> executor)
{
    executor_ = executor;
}

std::shared_ptr<Executor> ExecutorHolder::GetExecutor()
{
    if (executor_) {
        return executor_;
    }
    executor_ = std::make_shared<FunctionExecutor>();
    return executor_;
}

Libruntime::ErrorInfo LoadFunctions(const std::vector<std::string> &paths)
{
    return ExecutorHolder::Singleton().GetExecutor()->LoadFunctions(paths);
}

Libruntime::ErrorInfo LoadNoneFunctions(const std::vector<std::string> &paths)
{
    return Libruntime::ErrorInfo();
}

Libruntime::ErrorInfo ExecuteFunction(const YR::Libruntime::FunctionMeta &function,
                                      const libruntime::InvokeType invokeType,
                                      const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
                                      std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects)
{
    return ExecutorHolder::Singleton().GetExecutor()->ExecuteFunction(function, invokeType, rawArgs, returnObjects);
}

Libruntime::ErrorInfo Checkpoint(const std::string &instanceID, std::shared_ptr<YR::Libruntime::Buffer> &data)
{
    return ExecutorHolder::Singleton().GetExecutor()->Checkpoint(instanceID, data);
}

Libruntime::ErrorInfo Recover(std::shared_ptr<YR::Libruntime::Buffer> data)
{
    return ExecutorHolder::Singleton().GetExecutor()->Recover(data);
}

Libruntime::ErrorInfo ExecuteShutdownFunction(uint64_t gracePeriodSecond)
{
    return ExecutorHolder::Singleton().GetExecutor()->ExecuteShutdownFunction(gracePeriodSecond);
}

Libruntime::ErrorInfo Signal(int sigNo, std::shared_ptr<YR::Libruntime::Buffer> payload)
{
    return ExecutorHolder::Singleton().GetExecutor()->Signal(sigNo, payload);
}

}  // namespace internal
}  // namespace YR