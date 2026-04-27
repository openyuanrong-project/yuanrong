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

#include "api/cpp/src/executor/function_executor.h"

#include <dlfcn.h>

#if __has_include(<filesystem>)
#include <filesystem>
#else
#include <experimental/filesystem>
#endif

#include "boost/range/iterator_range.hpp"

#include "api/cpp/src/utils/utils.h"
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
const char *DynamicLibraryEnvKey = "LD_LIBRARY_PATH";

void AddLibraryInternal(const filesystem::path &path, std::set<filesystem::path> &libPaths)
{
    YRLOG_DEBUG("path: {}", path.string());
    if (path.extension().string() == ".so") {
        libPaths.emplace(path);
    }
}

void FunctionExecutor::OpenLibrary(const std::string &path)
{
    YRLOG_INFO("Begin to open library: {}", path);
    if (libs_.find(path) != libs_.end()) {
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

    libs_.emplace(path, handle);

    YRLOG_INFO("Success to open library {}", path);
}

void FunctionExecutor::DoLoadFunctions(const std::vector<std::string> &paths)
{
    std::set<filesystem::path> libPaths;
    for (auto const &path : paths) {
        filesystem::path p = path;
        if (filesystem::is_directory(p)) {
            for (auto const &entry : boost::make_iterator_range(filesystem::directory_iterator(p), {})) {
                AddLibraryInternal(entry, libPaths);
            }
        } else if (filesystem::exists(p)) {
            AddLibraryInternal(p, libPaths);
        } else {
            YRLOG_WARN("failed to open lib path: {}", path);
        }
    }
    YRLOG_INFO("{}={}", DynamicLibraryEnvKey, GetEnv(DynamicLibraryEnvKey));
    if (libPaths.empty()) {
        throw YR::Exception(ErrorCode::ERR_USER_CODE_LOAD, ModuleCode::RUNTIME_CREATE, "cannot find shared library file.");
    }
    for (auto &lib : libPaths) {
        OpenLibrary(lib.string());
    }
}

Libruntime::ErrorInfo FunctionExecutor::LoadFunctions(const std::vector<std::string> &paths)
{
    Libruntime::ErrorInfo err;
    try {
        this->DoLoadFunctions(paths);
    } catch (YR::Exception &e) {
        err.SetErrCodeAndMsg(static_cast<ErrorCode>(e.Code()), ModuleCode::RUNTIME, e.Msg());
    } catch (std::exception &e) {
        err.SetErrCodeAndMsg(ErrorCode::ERR_USER_CODE_LOAD, ModuleCode::RUNTIME, e.what());
    } catch (...) {
        err.SetErrCodeAndMsg(ErrorCode::ERR_USER_CODE_LOAD, ModuleCode::RUNTIME, "unkown reason");
    }
    return err;
}

// @param rawBuffers: invoke arg raw buffers
// @param bufPtr: return value buffer
Libruntime::ErrorInfo FunctionExecutor::ExecNormalFunction(const std::string &funcName, const std::string &returnObjId,
                                                        const std::vector<msgpack::sbuffer> &rawBuffers,
                                                        std::shared_ptr<msgpack::sbuffer> &bufPtr, bool &putDone)
{
    auto func = FunctionManager::Singleton().GetNormalFunction(funcName);
    if (!func.has_value()) {
        std::string msg = funcName + " is not found in FunctionHelper, check if func is decorated by YR_INVOKE";
        return Libruntime::ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    }

    try {
        auto result = func.get()(returnObjId, rawBuffers);
        bufPtr = result.first;
        putDone = result.second;
    } catch (std::exception &e) {
        std::string msg = "exception happens when executing user's function: " + std::string(e.what());
        return Libruntime::ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    } catch (...) {
        std::string msg = "unknown exception happens when executing user function";
        return Libruntime::ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    }

    return Libruntime::ErrorInfo();
}

Libruntime::ErrorInfo FunctionExecutor::ExecInstanceFunction(const std::string &funcName,
                                                             const std::string &returnObjId,
                                                             const std::vector<msgpack::sbuffer> &rawBuffers,
                                                             std::shared_ptr<msgpack::sbuffer> namedObject,
                                                             std::shared_ptr<msgpack::sbuffer> &bufPtr, bool &putDone)
{
    auto func = FunctionManager::Singleton().GetInstanceFunction(funcName);
    if (!func.has_value()) {
        std::string msg = funcName + " is not found in FunctionHelper, check if func is decorated by YR_INVOKE";
        return Libruntime::ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    }

    try {
        auto result = func.get()(returnObjId, *namedObject, rawBuffers);
        bufPtr = result.first;
        putDone = result.second;
    } catch (std::exception &e) {
        std::string msg = "failed to invoke " + funcName + ", exception: " + std::string(e.what()) +
                          " return obj id is: " + returnObjId;
        return Libruntime::ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    } catch (...) {
        std::string msg = "failed to invoke " + funcName + " with unknown exception, return obj id is: " + returnObjId;
        return Libruntime::ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, msg);
    }
    return Libruntime::ErrorInfo();
}

Libruntime::ErrorInfo FunctionExecutor::ExecuteFunction(
    const YR::Libruntime::FunctionMeta &function, const libruntime::InvokeType invokeType,
    const std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &rawArgs,
    std::vector<std::shared_ptr<YR::Libruntime::DataObject>> &returnObjects)
{
    Libruntime::ErrorInfo err;
    std::shared_ptr<msgpack::sbuffer> retVal;
    bool putDone;  // it is a return value of ExecInstanceFunction, ExecNormalFunction. True: needn't Put outside.
    std::vector<msgpack::sbuffer> rawBuffers(rawArgs.size());
    for (size_t i = 0; i < rawArgs.size(); i++) {
        rawBuffers[i].write(static_cast<const char *>(rawArgs[i]->data->MutableData()), rawArgs[i]->data->GetSize());
    }

    if (invokeType == libruntime::InvokeType::CreateInstance) {
        err = this->ExecNormalFunction(function.funcName, "", rawBuffers, retVal, putDone);
        if (err.OK()) {
            instancePtr_ = retVal;
        }
        className_ = function.className;
        return err;
    }
    if (invokeType == libruntime::InvokeType::CreateInstanceStateless) {
        return err;
    }
    if (invokeType == libruntime::InvokeType::InvokeFunctionStateless) {
        err = this->ExecNormalFunction(function.funcName, returnObjects[0]->id, rawBuffers, retVal, putDone);
    } else if (invokeType == libruntime::InvokeType::InvokeFunction) {
        err = this->ExecInstanceFunction(function.funcName, returnObjects[0]->id, rawBuffers, instancePtr_, retVal,
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
        return Libruntime::ErrorInfo();
    }
    err = YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->AllocReturnObject(
        returnObjects[0], 0, retVal->size(), {}, totalNativeBufferSize);
    if (!err.OK()) {
        return err;
    }

    return WriteDataObject(retVal->data(), returnObjects[0], retVal->size(), {});
}

Libruntime::ErrorInfo FunctionExecutor::Checkpoint(const std::string &instanceID,
                                                std::shared_ptr<YR::Libruntime::Buffer> &data)
{
    if (instancePtr_ == nullptr) {
        YRLOG_INFO("object is null, instanceID: {}", instanceID);
        return Libruntime::ErrorInfo();
    }
    auto func = FunctionManager::Singleton().GetCheckpointFunction(className_);
    if (!func.has_value()) {
        std::string msg = className_ + " checkpoint func is not found in FunctionHelper";
        YRLOG_ERROR(msg);
        return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                                     YR::Libruntime::ModuleCode::RUNTIME, msg);
    }

    msgpack::sbuffer instanceBuf = func.get()(*instancePtr_);

    // data buffer format: [uint_8(size of buf1)|buf1(instanceBuf)|buf2(className_)]
    // nativeBuffer is the combination of instanceBuf and className_
    size_t instanceBufSize = instanceBuf.size();
    if (WillSizeOverFlow(sizeof(size_t), instanceBufSize)) {
        return Libruntime::ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                     "data size overflow instanceBuf size: " + std::to_string(instanceBufSize));
    }
    size_t nativeBufferSize = sizeof(size_t) + instanceBufSize;
    if (WillSizeOverFlow(nativeBufferSize, className_.size())) {
        return Libruntime::ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                     "data size overflow className_ size: " + std::to_string(className_.size()));
    }
    nativeBufferSize = nativeBufferSize + className_.size();
    std::shared_ptr<YR::Libruntime::NativeBuffer> nativeBuffer =
        std::make_shared<YR::Libruntime::NativeBuffer>(nativeBufferSize);
    char *ptr = reinterpret_cast<char *>(nativeBuffer->MutableData());
    std::copy_n(reinterpret_cast<char *>(&instanceBufSize), sizeof(size_t), ptr);
    ptr += sizeof(size_t);
    std::copy_n(instanceBuf.data(), instanceBufSize, ptr);
    ptr += instanceBufSize;
    std::copy_n(className_.data(), className_.size(), ptr);

    data = nativeBuffer;
    return Libruntime::ErrorInfo();
}

Libruntime::ErrorInfo FunctionExecutor::Recover(std::shared_ptr<YR::Libruntime::Buffer> data)
{
    if (data->GetSize() == 0) {
        return Libruntime::ErrorInfo();
    }

    // deserialize data buffer format: [uint_8(size of buf1)|buf1(instanceBuf)|buf2(className_)]
    char *ptr = reinterpret_cast<char *>(data->MutableData());
    size_t instanceBufSize;
    if (data->GetSize() < sizeof(size_t)) {
        return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, "data invalid");
    }
    std::copy_n(ptr, sizeof(size_t), reinterpret_cast<char *>(&instanceBufSize));
    ptr += sizeof(size_t);
    msgpack::sbuffer instanceBuf;
    instanceBuf.write(ptr, instanceBufSize);
    ptr += instanceBufSize;
    if (data->GetSize() - sizeof(size_t) < instanceBufSize) {
        return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, "data invalid");
    }
    size_t clsNameSize = data->GetSize() - sizeof(size_t) - instanceBufSize;
    className_ = std::string(ptr, clsNameSize);

    auto func = FunctionManager::Singleton().GetRecoverFunction(className_);
    if (!func.has_value()) {
        std::string msg = className_ + " recover func is not found in FunctionHelper";
        YRLOG_ERROR(msg);
        return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                                     YR::Libruntime::ModuleCode::RUNTIME, msg);
    }
    auto buffer = func.get()(instanceBuf);
    if (buffer.size() == 0) {
        std::string msg = className_ + " load failed: deserialize failed";
        YRLOG_ERROR(msg);
        return Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                                     YR::Libruntime::ModuleCode::RUNTIME, msg);
    }
    auto recoverCallback = FunctionManager::Singleton().GetRecoverCallbackFunction(className_);
    if (recoverCallback.has_value()) {
        YRLOG_INFO("execute the recover callback function of the user");
        recoverCallback.get()(buffer);
    }
    instancePtr_ = std::make_shared<msgpack::sbuffer>(std::move(buffer));

    return Libruntime::ErrorInfo();
}

Libruntime::ErrorInfo FunctionExecutor::ExecuteShutdownFunction(uint64_t gracePeriodSecond)
{
    if (instancePtr_ == nullptr) {
        return Libruntime::ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                                     "Instance pointer is null, stateful function may not be initialized.");
    }
    auto shutdownFunc = FunctionManager::Singleton().GetShutdownFunction(className_);
    if (!shutdownFunc.has_value()) {
        std::string msg = className_ + " shutdown func is not found in CodeManager";
        YRLOG_DEBUG(msg);
        return Libruntime::ErrorInfo();
    }

    try {
        shutdownFunc.get()(*instancePtr_, gracePeriodSecond);
    } catch (const std::exception &e) {
        return Libruntime::ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                                     "Failed to invoke shutdown function: " + std::string(e.what()));
    } catch (...) {
        return Libruntime::ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION,
                                     "Failed to invoke shutdown function with an unknown exception.");
    }
    return Libruntime::ErrorInfo();
}

Libruntime::ErrorInfo FunctionExecutor::Signal(int sigNo, std::shared_ptr<YR::Libruntime::Buffer> payload)
{
    return {};
}
}  // namespace internal
}  // namespace YR