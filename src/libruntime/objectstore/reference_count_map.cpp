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

#include "reference_count_map.h"
namespace YR {
namespace Libruntime {
void RefCountMap::IncreRefCount(const std::vector<std::string> &ids)
{
    std::lock_guard<std::mutex> lock(mapMtx);
    for (auto &objectId : ids) {
        refCountMap[objectId]++;
    }
}

std::vector<std::string> RefCountMap::DecreRefCount(const std::vector<std::string> &ids)
{
    std::vector<std::string> objectIds;
    std::lock_guard<std::mutex> lock(mapMtx);
    for (auto &objectId : ids) {
        if (refCountMap.find(objectId) != refCountMap.end()) {
            objectIds.emplace_back(objectId);
            refCountMap[objectId]--;
            if (refCountMap[objectId] == 0) {
                refCountMap.erase(objectId);
            }
        }
    }
    return objectIds;
}

std::vector<std::string> RefCountMap::ToArray() noexcept
{
    std::vector<std::string> objectIds;
    std::lock_guard<std::mutex> lock(mapMtx);
    for (auto &[id, count] : refCountMap) {
        objectIds.insert(objectIds.end(), count, id);
    }
    return objectIds;
}

void RefCountMap::Clear()
{
    refCountMap.clear();
}
}  // namespace Libruntime
}  // namespace YR