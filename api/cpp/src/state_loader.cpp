/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include <algorithm>

#include <boost/beast/core/detail/base64.hpp>
#include <msgpack.hpp>

#include "api/cpp/include/yr/api/function_manager.h"
#include "api/cpp/include/yr/api/serdes.h"
#include "api/cpp/src/code_manager.h"
#include "src/dto/buffer.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_manager.h"

using namespace boost::beast::detail;

namespace YR {
namespace internal {

YR::Libruntime::ErrorInfo LoadInstance(std::shared_ptr<YR::Libruntime::Buffer> data)
{
    if (data->GetSize() == 0) {
        return YR::Libruntime::ErrorInfo();
    }

    // deserialize data buffer format: [uint_8(size of buf1)|buf1(instanceBuf)|buf2(clsName)]
    char *ptr = reinterpret_cast<char *>(data->MutableData());
    if (data->GetSize() < sizeof(size_t)) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, "data size invalid");
    }
    size_t instanceBufSize;
    std::copy_n(ptr, sizeof(size_t), reinterpret_cast<char *>(&instanceBufSize));
    ptr += sizeof(size_t);
    msgpack::sbuffer instanceBuf;
    instanceBuf.write(ptr, instanceBufSize);
    ptr += instanceBufSize;
    if (data->GetSize() - sizeof(size_t) <= instanceBufSize) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, "data size invalid");
    }
    size_t clsNameSize = data->GetSize() - sizeof(size_t) - instanceBufSize;
    std::string clsName(ptr, clsNameSize);

    auto func = FunctionManager::Singleton().GetRecoverFunction(clsName);
    if (!func.has_value()) {
        std::string msg = clsName + " recover func is not found in FunctionHelper";
        YRLOG_ERROR(msg);
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION, YR::Libruntime::ModuleCode::RUNTIME,
                         msg);
    }
    std::shared_ptr<msgpack::sbuffer> bufPtr;
    auto buffer = func.get()(instanceBuf);
    if (buffer.size() == 0) {
        std::string msg = clsName + " load failed: deserialize failed";
        YRLOG_ERROR(msg);
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION, YR::Libruntime::ModuleCode::RUNTIME,
                         msg);
    }
    auto recoverCallback = FunctionManager::Singleton().GetRecoverCallbackFunction(clsName);
    if (recoverCallback.has_value()) {
        YRLOG_INFO("execute the recover callback function of the user");
        recoverCallback.get()(buffer);
    }
    bufPtr = std::make_shared<msgpack::sbuffer>(std::move(buffer));

    try {
        CodeManager::Singleton().SetInstanceBuffer(bufPtr);
        CodeManager::Singleton().SetClassName(clsName);
    } catch (Exception &e) {
        std::string msg = "exception happens when save instance or classname: " + std::string(e.what());
        YRLOG_ERROR(msg);
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME, msg);
    }
    return YR::Libruntime::ErrorInfo();
}

YR::Libruntime::ErrorInfo DumpInstance(const std::string &instanceID, std::shared_ptr<YR::Libruntime::Buffer> &data)
{
    std::shared_ptr<msgpack::sbuffer> namedObject;
    namedObject = CodeManager::Singleton().GetInstanceBuffer();
    if (namedObject == nullptr) {
        YRLOG_INFO("object is null, instanceID: {}", instanceID);
        return YR::Libruntime::ErrorInfo();
    }
    std::string clsName = CodeManager::Singleton().GetClassName();
    auto func = FunctionManager::Singleton().GetCheckpointFunction(clsName);
    if (!func.has_value()) {
        std::string msg = clsName + " checkpoint func is not found in FunctionHelper";
        YRLOG_ERROR(msg);
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION, YR::Libruntime::ModuleCode::RUNTIME,
                         msg);
    }

    msgpack::sbuffer instanceBuf = func.get()(*namedObject);

    // data buffer format: [uint_8(size of buf1)|buf1(instanceBuf)|buf2(clsName)]
    // nativeBuffer is the combination of instanceBuf and clsName
    size_t nativeBufferSize = sizeof(size_t) + instanceBuf.size() + clsName.size();
    size_t instanceBufSize = instanceBuf.size();
    std::shared_ptr<YR::Libruntime::NativeBuffer> nativeBuffer =
        std::make_shared<YR::Libruntime::NativeBuffer>(nativeBufferSize);
    char *ptr = reinterpret_cast<char *>(nativeBuffer->MutableData());
    std::copy_n(reinterpret_cast<char *>(&instanceBufSize), sizeof(size_t), ptr);
    ptr += sizeof(size_t);
    std::copy_n(instanceBuf.data(), instanceBuf.size(), ptr);
    ptr += instanceBufSize;
    std::copy_n(clsName.data(), clsName.size(), ptr);

    data = nativeBuffer;
    return YR::Libruntime::ErrorInfo();
}

}  // namespace internal
}  // namespace YR
