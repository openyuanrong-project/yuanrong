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
#include <algorithm>
#include <climits>

#include "yr/api/macro.h"
#include "yr/parallel/detail/parallel_for_local.h"

namespace YR {
namespace Parallel {

int GetThreadPoolSize();

/**
 * @brief ParallelFor is a function framework for parallel computing, enabling tasks to be executed in parallel across
 * multiple threads to improve computational efficiency. ParallelFor internally implements parallel computing through
 * task allocation and scheduling, automatically distributing tasks to available threads.
 * @tparam Index The type of the iteration variable.
 * @tparam Handler The type of the function to be called.
 * @param start The starting value of the loop iteration range.
 * @param end The ending value of the loop iteration range (exclusive).
 * @param handler The function to be executed in the loop. The parameter list of the user-defined handler can be one of
 * the following two types:
 * 1. (Index, Index)
 * 2. (Index, Index, const YR::Parallel::Context&)
 * When the user's handler uses the YR::Parallel::Context parameter, the value of context.id in the handler will be in
 * the range [0, parallelism).
 * @param chunkSize The granularity of the task.
 * @param workThreadSize The number of worker threads. If set to -1 (default), it will be set to the number of threads
 * in the thread pool plus 1.
 * @throws Exception
 * 1. If ParallelFor is called before initialization, an exception "Assertion IsInitialized() failed !!!" will be
 * thrown.
 * 2. If the parameter list of the user-defined handler does not match the specified format, a compilation error will
 * occur: "error: static assertion failed: handler must have 2 or 3 arguments. And arguments should be (Index, Index) or
 * (Index, Index, const YR::Parallel::Context&)".
 *
 * @snippet{trimleft} parallel_for_example.cpp parallel for
 */
template <typename Index, typename Handler>
void ParallelFor(Index start, Index end, const Handler &handler, size_t chunkSize = -1, int workThreadSize = -1)
{
    if (end == start) {
        return;
    }
    YR_ASSERT(end > start);
    YR_ASSERT(workThreadSize == -1 || workThreadSize >= 1);
    YR_ASSERT(chunkSize == -1 || chunkSize >= 1);
    YR_ASSERT(chunkSize >= 1 && (Index)(INT_MAX - chunkSize) > end);  // performance considerations
    YR_ASSERT(IsInitialized());
    int parallelDegree;
    if (workThreadSize == -1) {
        parallelDegree = GetThreadPoolSize() + 1;
    } else {
        parallelDegree = std::min(workThreadSize, GetThreadPoolSize() + 1);
    }
    if (chunkSize == -1) {
        static const int DEFAULT_CHUNK_COUNT_PER_THREAD_ON_AVERAGE = 4;  // each thread processes 4 chunks on averge
        chunkSize = std::max(
            static_cast<int>(end - start) / parallelDegree / DEFAULT_CHUNK_COUNT_PER_THREAD_ON_AVERAGE, 1);
    }
    auto taskNum = static_cast<int>(ceil(static_cast<double>(end - start) / static_cast<double>(chunkSize)));
    parallelDegree = std::min(taskNum, parallelDegree);

    constexpr bool typeCheck = ParallelForLocal<Index, Handler>::HandlerTypeCheck();
    static_assert(typeCheck,
                  "handler must have 2 or 3 arguments. And arguments should be (Index, Index) or (Index, "
                  "Index, const YR::Parallel::Context&)");

    if (taskNum == 1) {
        ParallelForLocal<Index, Handler>::CallBodyHandler(start, end, handler, Context{0});
        return;
    }
    // allocate workshare
    auto local = std::make_shared<ParallelForLocal<Index, Handler>>(start, end, handler, chunkSize);
    local->DoParallelFor(parallelDegree);
}
}  // namespace Parallel
}  // namespace YR
