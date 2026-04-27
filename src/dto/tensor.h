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
namespace YR {
namespace Libruntime {
enum class DataType : uint8_t {
    INT8 = 0,
    INT16,
    INT32,
    INT64,
    UINT8,
    UINT16,
    UINT32,
    UINT64,
    FLOAT16,
    FLOAT32,
    FLOAT64,
    BFLOAT16,
    COMPLEX64,
    COMPLEX128,

    INVALID
};

struct Tensor {
    uint8_t *ptr = nullptr;
    uint64_t count = 0;
    DataType dtype = DataType::INVALID;
    int32_t deviceIdx = -1;
    bool operator==(const Tensor &other) const
    {
        return this->ptr == other.ptr && this->count == other.count && this->dtype == other.dtype &&
               this->deviceIdx == other.deviceIdx;
    }
};
}  // namespace Libruntime
}  // namespace YR