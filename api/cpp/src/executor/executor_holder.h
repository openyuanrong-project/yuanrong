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

#include "api/cpp/src/executor/executor.h"

namespace YR {
namespace internal {
class ExecutorHolder {
public:
    static ExecutorHolder &Singleton();

    void SetExecutor(std::shared_ptr<Executor> executor);

    std::shared_ptr<Executor> GetExecutor();

private:
    std::shared_ptr<Executor> executor_;
};

Libruntime::ErrorInfo LoadFunctions(const std::vector<std::string> &paths);

Libruntime::ErrorInfo LoadNoneFunctions(const std::vector<std::string> &paths);

Libruntime::ErrorInfo ExecuteFunction(const YR::Libruntime::FunctionMeta &function,
                                      const libruntime::InvokeType invokeType,
                                      const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
                                      std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects);

Libruntime::ErrorInfo Checkpoint(const std::string &instanceID, std::shared_ptr<YR::Libruntime::Buffer> &data);

Libruntime::ErrorInfo Recover(std::shared_ptr<YR::Libruntime::Buffer> data);

Libruntime::ErrorInfo ExecuteShutdownFunction(uint64_t gracePeriodSecond);

Libruntime::ErrorInfo Signal(int sigNo, std::shared_ptr<YR::Libruntime::Buffer> payload);

}  // namespace internal
}  // namespace YR