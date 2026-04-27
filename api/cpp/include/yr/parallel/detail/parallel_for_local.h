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
#include <boost/callable_traits/is_invocable.hpp>

#include "barrier.h"
#include "native_sem.h"
#include "yr/api/function_handler.h"

namespace YR {
extern thread_local int g_threadid;
static inline int GetThreadid()
{
    return g_threadid;
}

template<typename F, typename... Args>
constexpr bool is_invocable_v = boost::callable_traits::is_invocable<F, Args...>::value;
namespace Parallel {

/*!
   @struct Context
   @brief thread context info
 */
struct Context {
    /*!
       @brief thread identify id
     */
    size_t id;
};

int LocalSubmit(std::function<void()> &&func);

template <typename Index, typename Handler>
struct ParallelForLocal : public std::enable_shared_from_this<ParallelForLocal<Index, Handler>> {
public:
    ParallelForLocal(const ParallelForLocal<Index, Handler> &) = delete;
    ParallelForLocal(ParallelForLocal<Index, Handler> &&) = delete;
    ParallelForLocal<Index, Handler> &operator=(const ParallelForLocal<Index, Handler> &) = delete;
    ParallelForLocal<Index, Handler> &operator=(ParallelForLocal<Index, Handler> &&) = delete;

    ParallelForLocal(const Index &start, const Index &end, const Handler &handler, const size_t &chunkSize)
        : startIndex(start), endIndex(end), bodyHandler(handler), chunksize(chunkSize)
    {
        threadBarrier = new Barrier<NativeSem>();
    }

    ~ParallelForLocal()
    {
        if (threadBarrier != nullptr) {
            delete threadBarrier;
            threadBarrier = nullptr;
        }
    }

    void DoParallelFor(const int &parallelDegree)
    {
        size_t chunkCount = (endIndex - startIndex + chunksize - 1) / chunksize;
        // master thread fork, this will init barrier
        threadBarrier->ForkBarrier(chunkCount);

        // worker threads do work
        for (int i = 0; i < parallelDegree - 1; i++) {
            Context ctx;
            ctx.id = i;
            auto weak = this->weak_from_this();
            LocalSubmit([weak, ctx]() {
                if (auto ptr = weak.lock(); ptr) {
                    ptr->ParallelForLocal<Index, Handler>::ParallelForDynamicEntryTask(ctx);
                }
            });
        }
        Context ctx;
        ctx.id = parallelDegree - 1;
        ParallelForDynamicEntryTask(ctx);

        // master thread join
        threadBarrier->JoinBarrier(true);
    }

    static constexpr bool HandlerTypeCheck()
    {
        if constexpr (is_invocable_v<Handler, Index, Index>) {
            return true;
        } else if constexpr (is_invocable_v<Handler, Index, Index, const YR::Parallel::Context &>) {
            return true;
        }
        return false;
    }

    static void CallBodyHandler(Index start, Index end, const Handler &handler, const Context &ctx)
    {
        //  match the argument format of Handler
        if constexpr (is_invocable_v<Handler, Index, Index>) {
            handler(start, end);
        } else if constexpr (is_invocable_v<Handler, Index, Index, const YR::Parallel::Context &>) {
            handler(start, end, ctx);
        }
    }

private:
    YR_ALWAYS_FORCE_INLINE void ParallelForDynamicEntryTask(const Context &ctx)
    {
        Index start;
        Index end;
        for (;;) {
            bool hasNextSlice = GetNextSliceDynamic(&start, &end);
            if (hasNextSlice) {
                CallBodyHandler(start, end, bodyHandler, ctx);
                threadBarrier->JoinBarrier(false);
            } else {
                break;
            }
        }
    }

    // modify startIndex concurrency
    YR_ALWAYS_FORCE_INLINE bool GetNextSliceDynamic(Index *start, Index *end)
    {
        bool success;
        do {
            *start = startIndex;

            if (YR_UNLIKELY(*start >= endIndex)) {
                return false;
            }

            *end = *start + (Index)chunksize;
            success = __sync_bool_compare_and_swap(&startIndex, *start, *end);
        } while (!success);

        if (*end > endIndex) {
            *end = endIndex;
        }

        return true;
    }

    Index startIndex;
    const Index endIndex;
    const Handler bodyHandler;
    const size_t chunksize;
    Barrier<NativeSem> *threadBarrier;
};
}  // namespace Parallel
}  // namespace YR