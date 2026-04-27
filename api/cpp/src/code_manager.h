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

#include <memory>
#include <string>
#include <unordered_map>
#include <vector>

#include "msgpack.hpp"

#include "src/dto/data_object.h"
#include "src/dto/status.h"
#include "src/libruntime/err_type.h"

namespace YR {
namespace internal {
using YR::Libruntime::ErrorInfo;
class CodeManager {
public:
    static CodeManager &Singleton()
    {
        static CodeManager singleton{};
        return singleton;
    }

    void DoLoadFunctions(const std::vector<std::string> &paths);

    static ErrorInfo LoadFunctions(const std::vector<std::string> &paths);

    static ErrorInfo ExecuteFunction(const YR::Libruntime::FunctionMeta &function,
                                     const libruntime::InvokeType invokeType,
                                     const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
                                     std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects);

    std::string GetClassName();

    void SetClassName(std::string &clsName);

    std::shared_ptr<msgpack::sbuffer> GetInstanceBuffer();

    void SetInstanceBuffer(std::shared_ptr<msgpack::sbuffer> sbuf);

    static ErrorInfo ExecuteShutdownFunction(uint64_t gracePeriodSecond);
private:
    void OpenLibrary(const std::string &path);

    ErrorInfo ExecNormalFunction(const std::string &funcName, const std::string &returnObjId,
                                 const std::vector<msgpack::sbuffer> &rawBuffers,
                                 std::shared_ptr<msgpack::sbuffer> &bufPtr, bool &putDone);

    ErrorInfo ExecInstanceFunction(const std::string &funcName, const std::string &returnObjId,
                                   const std::vector<msgpack::sbuffer> &rawBuffers,
                                   std::shared_ptr<msgpack::sbuffer> namedObject,
                                   std::shared_ptr<msgpack::sbuffer> &bufPtr, bool &putDone);

    std::unordered_map<std::string, void *> libs;
    static std::shared_ptr<msgpack::sbuffer> instancePtr;
    static std::string className;
};

const std::string DynamicLibraryEnvKey = "LD_LIBRARY_PATH";
}  // namespace internal
}  // namespace YR