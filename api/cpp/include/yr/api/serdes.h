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

#pragma once
#include <msgpack.hpp>

#include "yr/api/buffer.h"
#include "yr/api/exception.h"


namespace YR {
namespace internal {
template <typename T>
static msgpack::sbuffer Serialize(const T &value)
{
    msgpack::sbuffer buffer;
    msgpack::pack(buffer, value);
    return buffer;
}

static bool referenceFunc(msgpack::type::object_type type, std::size_t length, void* user_data)
{
    return true;
}

template <typename T>
static T Deserialize(const msgpack::sbuffer &data)
{
    try {
        msgpack::object_handle oh = msgpack::unpack(data.data(), data.size(), referenceFunc);
        return oh.get().as<T>();
    } catch (std::exception &e) {
        std::string msg = "failed to deserialize input argument whose type=" + std::string(typeid(T).name()) +
            " and len=" + std::to_string(data.size()) + ", original exception message: " + std::string(e.what());
        throw YR::Exception::DeserializeException(msg);
    }
}

template <typename T>
static T Deserialize(const std::shared_ptr<Buffer> data)
{
    try {
        msgpack::object_handle oh =
            msgpack::unpack(static_cast<const char *>(data->ImmutableData()), data->GetSize(), referenceFunc);
        return oh.get().as<T>();
    } catch (std::exception &e) {
        std::string msg = "failed to deserialize input argument whose type=" + std::string(typeid(T).name()) +
                          " and len=" + std::to_string(data->GetSize()) +
                          ", original exception message: " + std::string(e.what());
        throw YR::Exception::DeserializeException(msg);
    }
}
}  // namespace internal
}  // namespace YR