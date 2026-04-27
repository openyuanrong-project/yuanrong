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

#include <msgpack.hpp>

namespace YR {
namespace internal {
class FunctionExecutor : public Executor {
public:
    FunctionExecutor() = default;
    ~FunctionExecutor() = default;
    Libruntime::ErrorInfo LoadFunctions(const std::vector<std::string> &paths) override;

    Libruntime::ErrorInfo ExecuteFunction(
        const YR::Libruntime::FunctionMeta &function, const libruntime::InvokeType invokeType,
        const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
        std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects) override;

    Libruntime::ErrorInfo Checkpoint(const std::string &instanceID,
                                         std::shared_ptr<YR::Libruntime::Buffer> &data) override;

    Libruntime::ErrorInfo Recover(std::shared_ptr<YR::Libruntime::Buffer> data) override;

    Libruntime::ErrorInfo ExecuteShutdownFunction(uint64_t gracePeriodSecond) override;

    Libruntime::ErrorInfo Signal(int sigNo, std::shared_ptr<YR::Libruntime::Buffer> payload) override;

private:
    void OpenLibrary(const std::string &path);
    void DoLoadFunctions(const std::vector<std::string> &paths);
    Libruntime::ErrorInfo ExecNormalFunction(const std::string &funcName, const std::string &returnObjId,
                                                 const std::vector<msgpack::sbuffer> &rawBuffers,
                                                 std::shared_ptr<msgpack::sbuffer> &bufPtr, bool &putDone);
    Libruntime::ErrorInfo ExecInstanceFunction(const std::string &funcName, const std::string &returnObjId,
                                                   const std::vector<msgpack::sbuffer> &rawBuffers,
                                                   std::shared_ptr<msgpack::sbuffer> namedObject,
                                                   std::shared_ptr<msgpack::sbuffer> &bufPtr, bool &putDone);

    std::unordered_map<std::string, void *> libs_;
    std::shared_ptr<msgpack::sbuffer> instancePtr_;
    std::string className_;
};

}  // namespace internal
}  // namespace YR