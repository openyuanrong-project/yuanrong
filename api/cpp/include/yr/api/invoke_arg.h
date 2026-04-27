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

namespace YR {
namespace internal {
struct InvokeArg {
    InvokeArg() = default;
    ~InvokeArg() = default;

    InvokeArg(InvokeArg &&rhs) noexcept
    {
        if (&rhs != this) {
            buf = std::move(rhs.buf);
            yrBuf = std::move(rhs.yrBuf);
            isRef = rhs.isRef;
            objId = std::move(rhs.objId);
            nestedObjects = std::move(rhs.nestedObjects);
        }
    }

    InvokeArg &operator=(InvokeArg &&rhs)
    {
        if (&rhs != this) {
            buf = std::move(rhs.buf);
            yrBuf = std::move(rhs.yrBuf);
            isRef = rhs.isRef;
            objId = std::move(rhs.objId);
            nestedObjects = std::move(rhs.nestedObjects);
        }
        return *this;
    }

    InvokeArg(const InvokeArg &) = delete;
    InvokeArg &operator=(InvokeArg const &) = delete;

    msgpack::sbuffer buf;
    Buffer yrBuf;
    bool isRef = false;
    std::string objId;                              // objId records objs for dependency resolver
    std::unordered_set<std::string> nestedObjects;  // nestedObjects records objs for ref count
};
}  // namespace internal
}  // namespace YR
