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

#include "src/dto/buffer.h"
#include "yr/api/buffer.h"

namespace YR {
class ReadOnlyBuffer : public Buffer {
public:
    ReadOnlyBuffer(std::shared_ptr<YR::Libruntime::Buffer> buf) : Buffer(nullptr, 0), buf_(buf) {
    }

    virtual uint64_t GetSize() const override
    {
        return buf_->GetSize();
    }

    virtual const void *ImmutableData() const override
    {
        return buf_->ImmutableData();
    }

private:
    std::shared_ptr<YR::Libruntime::Buffer> buf_;
};

}  // namespace YR