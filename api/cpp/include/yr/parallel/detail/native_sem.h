/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2022. All rights reserved.
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
#include "yr/api/macro.h"
#include <semaphore.h>

namespace YR {
namespace Parallel {
typedef sem_t YrSem;

class NativeSem {
public:
    NativeSem(const NativeSem &) = delete;
    NativeSem(NativeSem &&) = delete;
    NativeSem &operator = (const NativeSem &) = delete;
    NativeSem &operator = (NativeSem &&) = delete;

    NativeSem()
    {
        sem = (YrSem *)malloc(sizeof(YrSem));
    }

    ~NativeSem()
    {
        if (sem != nullptr) {
            free(sem);
            sem = nullptr;
        }
    }

    YR_ALWAYS_FORCE_INLINE void SemInit(int32_t initCnt)
    {
        (void)sem_init(sem, 0, initCnt);
    }

    YR_ALWAYS_FORCE_INLINE void SemDestroy()
    {
        (void)sem_destroy(sem);
    }

    YR_ALWAYS_FORCE_INLINE void SemPend()
    {
        (void)sem_wait(sem);
    }

    YR_ALWAYS_FORCE_INLINE void SemPost()
    {
        (void)sem_post(sem);
    }

private:
    YrSem *sem;
};
} // Parallel
} // YR