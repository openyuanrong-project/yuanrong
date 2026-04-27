/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include <mutex>
#include <atomic>

#include "yr/api/macro.h"

namespace YR {
namespace Parallel {

template <typename T> class Barrier {
public:
    Barrier(const Barrier<T> &) = delete;
    Barrier(Barrier<T> &&) = delete;
    Barrier<T> &operator = (const Barrier<T> &) = delete;
    Barrier<T> &operator = (Barrier<T> &&) = delete;

    Barrier()
    {
        semData = new T();
    }

    ~Barrier()
    {
        if (semData != nullptr) {
            semData->SemDestroy();
            delete semData;
            semData = nullptr;
        }
    }

    YR_ALWAYS_FORCE_INLINE void ForkBarrier(uint32_t initCnt)
    {
        // just master thread call it
        semData->SemInit(0);
        awaited = initCnt;
    }

    YR_ALWAYS_FORCE_INLINE void JoinBarrier(bool isMaster)
    {
        if (YR_UNLIKELY(isMaster)) {
            semData->SemPend();
            // after wait, master should destroy it
            awaited = 0;
            return;
        }
        // workers meet join barrier point
        if (YR_UNLIKELY(awaited.fetch_add(-1, std::memory_order_relaxed) == 1)) {
            semData->SemPost();
        }
    }

private:
    T *semData;
    std::atomic<int32_t> awaited { 0 };
};
} // Parallel
} // YR
