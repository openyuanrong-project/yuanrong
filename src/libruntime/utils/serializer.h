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
#include <msgpack.hpp>
#include "src/libruntime/utils/exception.h"

namespace YR {
namespace Libruntime {

class Serializer {
public:
    template <typename T>
    static msgpack::sbuffer Serialize(const T &value)
    {
        msgpack::sbuffer buffer;
        try {
            msgpack::pack(buffer, value);
            return buffer;
        } catch (const msgpack::container_size_overflow &e) {
            auto msg = "Serializer::Serialize container_size_overflow: " + std::string(e.what());
            YRLOG_ERROR(msg);
            THROW_ERRCODE_EXCEPTION(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, msg);
        } catch (const std::exception &e) {
            auto msg = "Serializer::Serialize exception msg: " + std::string(e.what());
            YRLOG_ERROR(msg);
            THROW_ERRCODE_EXCEPTION(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, msg);
        }
        return buffer;
    }

    template <typename T>
    static T Deserialize(const msgpack::sbuffer &data)
    {
        try {
            msgpack::unpacked unpacked = msgpack::unpack(data.data(), data.size(), 0);
            return unpacked.get().as<T>();
        } catch (std::exception &e) {
            std::string msg = "failed to deserialize input argument whose type=" + std::string(typeid(T).name()) +
                              " and len=" + std::to_string(data.size()) +
                              ", original exception message: " + std::string(e.what());
            THROW_ERRCODE_EXCEPTION(ErrorCode::ERR_DESERIALIZATION_FAILED, ModuleCode::RUNTIME, msg);
        }
    }
};
}  // namespace Libruntime
}  // namespace YR