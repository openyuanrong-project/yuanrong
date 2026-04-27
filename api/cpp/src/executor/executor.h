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

#include "src/dto/buffer.h"
#include "src/dto/data_object.h"
#include "src/dto/invoke_options.h"
#include "src/libruntime/err_type.h"

namespace YR {
namespace internal {

class Executor {
public:
    virtual Libruntime::ErrorInfo LoadFunctions(const std::vector<std::string> &paths) = 0;

    virtual Libruntime::ErrorInfo ExecuteFunction(
        const YR::Libruntime::FunctionMeta &function, const libruntime::InvokeType invokeType,
        const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
        std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects) = 0;

    virtual Libruntime::ErrorInfo Checkpoint(const std::string &instanceID,
                                             std::shared_ptr<YR::Libruntime::Buffer> &data) = 0;

    virtual Libruntime::ErrorInfo Recover(std::shared_ptr<YR::Libruntime::Buffer> data) = 0;

    virtual Libruntime::ErrorInfo ExecuteShutdownFunction(uint64_t gracePeriodSecond) = 0;

    virtual Libruntime::ErrorInfo Signal(int sigNo, std::shared_ptr<YR::Libruntime::Buffer> payload) = 0;
};

}  // namespace internal
}  // namespace YR