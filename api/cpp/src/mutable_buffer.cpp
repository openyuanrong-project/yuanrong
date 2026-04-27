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

#include "yr/api/mutable_buffer.h"
#include "yr/api/object_ref.h"

namespace YR {
void *MutableBuffer::MutableData()
{
    return nullptr;
}

ObjectRef<MutableBuffer> MutableBuffer::Publish()
{
    return ObjectRef<MutableBuffer>();
}

int64_t MutableBuffer::GetSize()
{
    return 0;
}
}  // namespace YR