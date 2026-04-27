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
class Buffer {
public:
    Buffer() = default;

    Buffer(void *data, uint64_t size)
    {
        data_ = data;
        size_ = size;
    }

    virtual uint64_t GetSize() const
    {
        return size_;
    }

    virtual const void *ImmutableData() const
    {
        return data_;
    }

private:
    void *data_ = nullptr;
    uint64_t size_ = 0;
};
}  // namespace YR