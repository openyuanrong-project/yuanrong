/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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
namespace Libruntime {
enum class WriteMode : int {
    NONE_L2_CACHE = 0,
    WRITE_THROUGH_L2_CACHE = 1,  // sync write
    WRITE_BACK_L2_CACHE = 2,     // async write
    NONE_L2_CACHE_EVICT = 3,     // evictable objects
};

enum class CacheType : int {
    MEMORY = 0,
    DISK = 1,
};
}  // namespace Libruntime
}  // namespace YR
