/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#define YR_UNUSED __attribute__((unused))
#define YR_LIKELY(x) __builtin_expect(!!(x), 1)
#define YR_UNLIKELY(x) __builtin_expect(!!(x), 0)

#ifdef DEBUG
#define YR_ALWAYS_FORCE_INLINE inline
#else
#define YR_ALWAYS_FORCE_INLINE __attribute__((always_inline)) inline
#endif

#ifdef DEBUG
#define YR_ASSERT(expr) assert(expr)
#else
#define YR_ASSERT(expr)                                            \
    do {                                                           \
        if (!(expr)) {                                             \
            throw YR::Exception("Assertion " #expr " failed !!!"); \
            (void)raise(SIGINT);                                   \
        }                                                          \
    } while (0)
#endif

#define CHECK_FAIL_THROW_EXCEPTION(condition, eMsg)                \
    do {                                                           \
        if (!(condition)) {                                        \
            throw YR::Exception(eMsg);                             \
        }                                                          \
    } while (0)
