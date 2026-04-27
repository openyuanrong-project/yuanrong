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

#include <mutex>
#include <unordered_map>
#include <vector>

namespace YR {
namespace Libruntime {
class RefCountMap {
public:
    void IncreRefCount(const std::vector<std::string> &ids);
    std::vector<std::string> DecreRefCount(const std::vector<std::string> &ids);
    std::vector<std::string> ToArray() noexcept;
    void Clear();

private:
    std::unordered_map<std::string, size_t> refCountMap;  // key:objectId, value: reference conut
    mutable std::mutex mapMtx;
};
}  // namespace Libruntime
}  // namespace YR