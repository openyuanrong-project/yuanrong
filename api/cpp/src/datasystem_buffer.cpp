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

#include "datasystem_buffer.h"

#include "yr/api/err_type.h"
#include "yr/api/exception.h"
#include "yr/api/object_ref.h"

namespace YR {
DataSystemBuffer::DataSystemBuffer(std::string objId, std::shared_ptr<YR::Libruntime::Buffer> buffer)
    : objId_(std::move(objId)), buffer_(std::move(buffer))
{
}

void *DataSystemBuffer::MutableData()
{
    return buffer_->MutableData();
}

ObjectRef<MutableBuffer> DataSystemBuffer::Publish()
{
    if (const YR::Libruntime::ErrorInfo err = buffer_->Publish(); !err.OK()) {
        throw YR::Exception(err.Code(), YR::ModuleCode::DATASYSTEM_, "MutableBuffer->Publish() fail");
    }
    return ObjectRef<MutableBuffer>(objId_);
}

int64_t DataSystemBuffer::GetSize()
{
    return buffer_->GetSize();
}
}  // namespace YR
