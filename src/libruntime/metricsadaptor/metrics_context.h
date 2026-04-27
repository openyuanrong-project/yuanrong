/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

#include <map>
#include <mutex>
#include <string>
#include <vector>

namespace YR {
namespace Libruntime {
using NodeLabelsType = std::map<std::string, std::map<std::string, uint64_t>>;

class MetricsContext {
public:
    MetricsContext() = default;
    ~MetricsContext() noexcept = default;

    std::string GetAttr(const std::string &attr) const;
    void SetAttr(const std::string &attr, const std::string &value);

private:
    std::map<std::string, std::string> attribute_;

    std::mutex invokeMtx_{};
    std::mutex functionMtx_{};
};
} // namespace metrics
} // namespace functionsystem
