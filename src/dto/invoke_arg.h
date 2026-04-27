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
#include <vector>
#include "src/dto/buffer.h"
#include "src/dto/data_object.h"

namespace YR {
namespace Libruntime {
struct InvokeArg {
    InvokeArg() = default;
    ~InvokeArg() = default;

    InvokeArg(InvokeArg &&rhs) noexcept
    {
        if (&rhs != this) {
            dataObj = std::move(rhs.dataObj);
            isRef = rhs.isRef;
            objId = std::move(rhs.objId);
            nestedObjects = std::move(rhs.nestedObjects);
            tenantId = std::move(rhs.tenantId);
        }
    }

    InvokeArg &operator=(InvokeArg &&rhs)
    {
        if (&rhs != this) {
            dataObj = std::move(rhs.dataObj);
            isRef = rhs.isRef;
            objId = std::move(rhs.objId);
            nestedObjects = std::move(rhs.nestedObjects);
            tenantId = std::move(rhs.tenantId);
        }
        return *this;
    }

    InvokeArg(const InvokeArg &lhs)
    {
        if (&lhs != this) {
            dataObj = lhs.dataObj;
            isRef = lhs.isRef;
            objId = lhs.objId;
            nestedObjects = lhs.nestedObjects;
            tenantId = lhs.tenantId;
        }
    }

    InvokeArg &operator=(const InvokeArg &lhs)
    {
        if (&lhs != this) {
            dataObj = lhs.dataObj;
            isRef = lhs.isRef;
            objId = lhs.objId;
            nestedObjects = lhs.nestedObjects;
            tenantId = lhs.tenantId;
        }
        return *this;
    }

    std::shared_ptr<DataObject> dataObj;
    bool isRef = false;
    std::string objId;                              // objId records objs for dependency resolver
    std::unordered_set<std::string> nestedObjects;  // nestedObjects records objs for ref count;
    std::string tenantId;
};
}  // namespace Libruntime
}  // namespace YR
