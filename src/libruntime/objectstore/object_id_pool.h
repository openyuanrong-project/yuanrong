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

#include <memory>
#include <mutex>
#include <string>
#include <vector>

#include "src/libruntime/objectstore/memory_store.h"

namespace YR {
namespace Libruntime {
/**
 * @brief The count of all objids in ObjectIdPool has been incremented by 1 in datasystem.
 */
class ObjectIdPool {
public:
    ObjectIdPool(std::shared_ptr<MemoryStore> memStore, size_t size = DEFAULT_OBJECT_POOL_SIZE);

    ObjectIdPool(const ObjectIdPool &) = delete;

    ObjectIdPool &operator=(const ObjectIdPool &) = delete;

    std::pair<ErrorInfo, std::string> Pop();

    void Clear();

private:
    ErrorInfo Scale();

    std::shared_ptr<MemoryStore> memoryStore;
    size_t poolSize;
    std::mutex mu;
    std::vector<std::string> idPool;
    static const int DEFAULT_OBJECT_POOL_SIZE = 100;
};
}  // namespace Libruntime
}  // namespace YR