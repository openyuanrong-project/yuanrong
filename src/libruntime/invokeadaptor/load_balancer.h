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

#include <memory>
#include <mutex>
#include <vector>
#include <unordered_map>

#include "src/libruntime/invokeadaptor/scheduler_instance_info.h"

namespace YR {
namespace Libruntime {
const std::string ALL_SCHEDULER_UNAVAILABLE = "AllSchedulerUnavailable";
enum class LoadBalancerType : int {
    WeightedRoundRobin,
    ConsistantRoundRobin,
};

class LoadBalancer {
public:
    static LoadBalancer *Factory(LoadBalancerType type);

    LoadBalancer() = default;
    virtual ~LoadBalancer() = default;

    virtual std::string Next(const std::string &name, bool move = false) = 0;
    virtual void Add(const std::string &node, int weight) = 0;
    virtual void RemoveAll(void) = 0;
    virtual void Remove(const std::string &name) = 0;
};

}  // namespace Libruntime
}  // namespace YR