/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: memory pool implementation
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

#include "mem_pool.h"
#include "securec.h"

namespace YR {
namespace Parallel {
MemPool::~MemPool()
{
    if (fixMem != nullptr) {
        free(fixMem);
        fixMem = nullptr;
    }
}

int MemPool::InitMemPool(uint32_t objSize, uint32_t objNum)
{
    uint32_t i;
    uint32_t queSize;
    /* queSize = 2^n. max contain obj= 2^n -1 */
    queSize = objNum + 1;
    if ((queSize & objNum) > 0) {
        i = 0;
        while (queSize > 0) {
            queSize >>= 1;
            i++;
        }
        queSize = (1u << i);
    }
    que = TaskQueue(queSize);

    perObjSize = objSize;
    uint32_t mallocSize = perObjSize * objNum;
    fixMem = malloc(mallocSize);
    if (fixMem == nullptr) {
        return H_FAIL;
    }
    (void)memset_s(fixMem, mallocSize, 0, mallocSize);

    uint8_t *obj = (uint8_t *)fixMem;
    for (i = 0; i < objNum; i++) {
        (void)que.EnqueueParallel(reinterpret_cast<intptr_t>(obj));
        obj = obj + perObjSize;
    }

    return H_SUCCESS;
}

void *MemPool::AllocObjFromMemPool()
{
    intptr_t obj;
    if (que.DequeueParallel(&obj) == H_SUCCESS) {
        return reinterpret_cast<void *>(obj);
    }
    return nullptr;
}

void MemPool::FreeObjToMemPool(const void *objTable)
{
    (void)que.EnqueueParallel(reinterpret_cast<intptr_t>(objTable));
}

void MemPool::DestroyMemPool()
{
    if (fixMem != nullptr) {
        free(fixMem);
        fixMem = nullptr;
    }
}
}  // namespace Parallel
}  // namespace YR