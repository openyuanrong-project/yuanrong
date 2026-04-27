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

#include "api/cpp/src/executor/posix_executor.h"

namespace YR {
namespace internal {
Libruntime::ErrorInfo PosixExecutor::LoadFunctions(const std::vector<std::string> &paths)
{
    return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "POSIX Executor not support");
}

Libruntime::ErrorInfo PosixExecutor::ExecuteFunction(
    const YR::Libruntime::FunctionMeta &function, const libruntime::InvokeType invokeType,
    const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects)
{
    return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "POSIX Executor not support");
}

Libruntime::ErrorInfo PosixExecutor::Checkpoint(const std::string &instanceID,
                                                std::shared_ptr<YR::Libruntime::Buffer> &data)
{
    return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "POSIX Executor not support");
}

Libruntime::ErrorInfo PosixExecutor::Recover(std::shared_ptr<YR::Libruntime::Buffer> data)
{
    return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "POSIX Executor not support");
}

Libruntime::ErrorInfo PosixExecutor::ExecuteShutdownFunction(uint64_t gracePeriodSecond)
{
    return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "POSIX Executor not support");
}

Libruntime::ErrorInfo PosixExecutor::Signal(int sigNo, std::shared_ptr<YR::Libruntime::Buffer> payload)
{
    return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "POSIX Executor not support");
}
}  // namespace internal
}  // namespace YR