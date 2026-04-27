/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: compiling and optimization for ARM and x86
 *
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
namespace Parallel {

#define LIKELY(x) __builtin_expect(!!(x), 1)
#define UNLIKELY(x) __builtin_expect(!!(x), 0)

inline bool AtomicCmpset32(volatile uint32_t *dst, uint32_t exp, uint32_t src)
{
    return __sync_bool_compare_and_swap(dst, exp, src);
}

inline bool AtomicCmpset64(volatile uint64_t *dst, uint64_t exp, uint64_t src)
{
    return __sync_bool_compare_and_swap(dst, exp, src);
}

inline void AtomicOr64(uint64_t *value, uint64_t n)
{
    (void)__sync_fetch_and_or(value, n);
}

inline void AtomicOr64(volatile uint64_t *value, uint64_t n)
{
    (void)__sync_fetch_and_or(value, n);
}

inline uint32_t GetLmb(uint64_t value)
{
    uint32_t index = static_cast<uint32_t>(__builtin_clzll(value));
    return index;
}

}  // namespace Parallel
}  // namespace YR