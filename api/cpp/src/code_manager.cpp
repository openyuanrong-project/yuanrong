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

#include <dlfcn.h>
#if __has_include(<filesystem>)
#include <filesystem>
#else
#include <experimental/filesystem>
#endif
#include <set>

#include "boost/range/iterator_range.hpp"

#include "api/cpp/src/utils/utils.h"
#include "code_manager.h"
#include "src/dto/constant.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/utility/logger/logger.h"
#include "src/utility/timer_worker.h"
#include "yr/api/function_manager.h"

namespace YR {
namespace internal {
using YR::Libruntime::ErrorCode;
using YR::Libruntime::ModuleCode;
using YR::utility::ExecuteByGlobalTimer;
#ifdef __cpp_lib_filesystem
namespace filesystem = std::filesystem;
#elif __cpp_lib_experimental_filesystem
namespace filesystem = std::experimental::filesystem;
#endif

const int MAX_READFILE_TIME = 30;  // seconds

void AddLibrary(const filesystem::path &path, std::set<filesystem::path> &libPaths)
{
    YRLOG_DEBUG("path: {}", path.string());
    if (path.extension().string() == ".so") {
        libPaths.emplace(path);
    }
}

std::string GetEnv(const std::string &key)
{
    if (const char *env = std::getenv(key.c_str())) {
        return std::string(env);
    }
    return std::string("");
}

std::shared_ptr<msgpack::sbuffer> CodeManager::instancePtr = nullptr;
std::string CodeManager::className = "";

void CodeManager::DoLoadFunctions(const std::vector<std::string> &paths)
{
    std::set<filesystem::path> libPaths;
    for (auto const &path : paths) {
        filesystem::path p = path;
        if (filesystem::is_directory(p)) {
            for (auto const &entry : boost::make_iterator_range(filesystem::directory_iterator(p), {})) {
                AddLibrary(entry, libPaths);
            }
        } else if (filesystem::exists(p)) {
            AddLibrary(p, libPaths);
        } else {
            YRLOG_WARN("failed to open lib path: {}", path);
        }
    }

    YRLOG_INFO("{}={}", DynamicLibraryEnvKey, GetEnv(DynamicLibraryEnvKey));
    if (libPaths.size() == 0) {
        throw YR::Exception(ErrorCode::ERR_USER_CODE_LOAD, ModuleCode::RUNTIME_CREATE,
                            "cannot find shared library file");
    }
    for (auto &lib : libPaths) {
        OpenLibrary(lib.string());
    }
}

void CodeManager::OpenLibrary(const std::string &path)
{
    YRLOG_INFO("Begin to open library: {}", path);
    if (libs.find(path) != libs.end()) {
        return;
    }

    if (!filesystem::exists(path)) {
        YRLOG_ERROR("Library path {} does not exist!", path);
        return;
    }

    auto t = ExecuteByGlobalTimer([]() { YRLOG_WARN("ReadFile timeout, timeConsumption = {}", MAX_READFILE_TIME); },
                                  MAX_READFILE_TIME * YR::Libruntime::MILLISECOND_UNIT, -1);

    dlerror();
    void *handle = dlopen(path.c_str(), RTLD_LAZY | RTLD_GLOBAL);
    t->cancel();
    if (handle == nullptr) {
        std::string err(dlerror());
        YRLOG_ERROR("Failed to open library from path {}, reason: {}", path, err);
        throw YR::Exception(ErrorCode::ERR_USER_CODE_LOAD, "path: " + path + ", error: " + err);
    }

    libs.emplace(path, handle);

    YRLOG_INFO("Success to open library {}", path);
}

YR::Libruntime::ErrorInfo CodeManager::LoadFunctions(const std::vector<std::string> &paths)
{
    YR::Libruntime::ErrorInfo err;
    try {
        CodeManager::Singleton().DoLoadFunctions(paths);
    } catch (YR::Exception &e) {
        err.SetErrCodeAndMsg(static_cast<ErrorCode>(e.Code()), ModuleCode::RUNTIME, e.Msg());
    } catch (std::exception &e) {
        err.SetErrCodeAndMsg(ErrorCode::ERR_USER_CODE_LOAD, ModuleCode::RUNTIME, e.what());
    } catch (...) {
        err.SetErrCodeAndMsg(ErrorCode::ERR_USER_CODE_LOAD, ModuleCode::RUNTIME, "unkown reason");
    }
    return err;
}

ErrorInfo CodeManager::ExecuteFunction(const YR::Libruntime::FunctionMeta &function,
                                       const libruntime::InvokeType invokeType,
                                       const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
                                       std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects)
{
    ErrorInfo err;
    std::shared_ptr<msgpack::sbuffer> retVal;
    bool putDone;  // it is a return value of ExecInstanceFunction, ExecNormalFunction. True: needn't Put outside.
    std::vector<msgpack::sbuffer> rawBuffers(rawArgs.size());
    for (size_t i = 0; i < rawArgs.size(); i++) {
        rawBuffers[i].write(static_cast<const char *>(rawArgs[i]->data->MutableData()), rawArgs[i]->data->GetSize());
    }
    if (invokeType == libruntime::InvokeType::CreateInstance) {
        err = CodeManager::Singleton().ExecNormalFunction(function.funcName, "", rawBuffers, retVal, putDone);
        if (err.OK()) {
            instancePtr = retVal;
        }
        CodeManager::className = function.className;
        return err;
    } else if (invokeType == libruntime::InvokeType::CreateInstanceStateless) {
        return err;
    } else if (invokeType == libruntime::InvokeType::InvokeFunction) {
        err = CodeManager::Singleton().ExecInstanceFunction(function.funcName, returnObjects[0]->id, rawBuffers,
                                                            instancePtr, retVal, putDone);
    } else if (invokeType == libruntime::InvokeType::InvokeFunctionStateless) {
        err = CodeManager::Singleton().ExecNormalFunction(function.funcName, returnObjects[0]->id, rawBuffers, retVal,
                                                          putDone);
    } else {
        err.SetErrCodeAndMsg(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                             "Invalid invoke type" + libruntime::InvokeType_Name(invokeType));
    }

    if (!err.OK()) {
        return err;
    }

    uint64_t totalNativeBufferSize = 0;
    if (putDone) {
        // has already put
        returnObjects[0]->putDone = true;
        return ErrorInfo();
    }
    err = YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->AllocReturnObject(
        returnObjects[0], 0, retVal->size(), {}, totalNativeBufferSize);
    if (!err.OK()) {
        return err;
    }

    return WriteDataObject(retVal->data(), returnObjects[0], retVal->size(), {});
}

// @param rawBuffers: invoke arg raw buffers
// @param bufPtr: return value buffer
ErrorInfo CodeManager::ExecNormalFunction(const std::string &funcName, const std::string &returnObjId,
                                          const std::vector<msgpack::sbuffer> &rawBuffers,
                                          std::shared_ptr<msgpack::sbuffer> &bufPtr, bool &putDone)
{
    auto func = FunctionManager::Singleton().GetNormalFunction(funcName);
    if (!func.has_value()) {
        std::string msg = funcName + " is not found in FunctionHelper, check if func is decorated by YR_INVOKE";
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    }

    try {
        auto result = func.get()(returnObjId, rawBuffers);
        bufPtr = result.first;
        putDone = result.second;
    } catch (std::exception &e) {
        std::string msg = "exception happens when executing user's function: " + std::string(e.what());
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    } catch (...) {
        std::string msg = "unknown exception happens when executing user function";
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    }

    return ErrorInfo();
}

ErrorInfo CodeManager::ExecInstanceFunction(const std::string &funcName, const std::string &returnObjId,
                                            const std::vector<msgpack::sbuffer> &rawBuffers,
                                            std::shared_ptr<msgpack::sbuffer> namedObject,
                                            std::shared_ptr<msgpack::sbuffer> &bufPtr, bool &putDone)
{
    auto func = FunctionManager::Singleton().GetInstanceFunction(funcName);
    if (!func.has_value()) {
        std::string msg = funcName + " is not found in FunctionHelper, check if func is decorated by YR_INVOKE";
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    }

    try {
        auto result = func.get()(returnObjId, *namedObject, rawBuffers);
        bufPtr = result.first;
        putDone = result.second;
    } catch (std::exception &e) {
        std::string msg = "failed to invoke " + funcName + ", exception: " + std::string(e.what()) +
                          " return obj id is: " + returnObjId;
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    } catch (...) {
        std::string msg = "failed to invoke " + funcName + " with unknown exception, return obj id is: " + returnObjId;
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    }
    return ErrorInfo();
}

ErrorInfo CodeManager::ExecuteShutdownFunction(uint64_t gracePeriodSec)
{
    if (instancePtr == nullptr) {
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                         "Instance pointer is null, stateful function may not be initialized.");
    }
    std::string clsName = CodeManager::Singleton().GetClassName();
    auto shutdownFunc = FunctionManager::Singleton().GetShutdownFunction(clsName);
    if (!shutdownFunc.has_value()) {
        std::string msg = clsName + " shutdown func is not found in CodeManager";
        YRLOG_DEBUG(msg);
        return ErrorInfo();
    }

    try {
        shutdownFunc.get()(*instancePtr, gracePeriodSec);
    } catch (const std::exception &e) {
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                         "Failed to invoke shutdown function: " + std::string(e.what()));
    } catch (...) {
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                         "Failed to invoke shutdown function with an unknown exception.");
    }
    return ErrorInfo();
}

std::string CodeManager::GetClassName()
{
    return CodeManager::className;
}

void CodeManager::SetClassName(std::string &clsName)
{
    CodeManager::className = clsName;
}

std::shared_ptr<msgpack::sbuffer> CodeManager::GetInstanceBuffer()
{
    return instancePtr;
}

void CodeManager::SetInstanceBuffer(std::shared_ptr<msgpack::sbuffer> sbuf)
{
    instancePtr = sbuf;
}
}  // namespace internal
}  // namespace YR
