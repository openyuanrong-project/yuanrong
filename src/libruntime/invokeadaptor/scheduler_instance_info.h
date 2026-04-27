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
#include <string>
#include "absl/synchronization/mutex.h"

#include "src/libruntime/err_type.h"
#include "src/utility/json_utility.h"

namespace YR {
namespace Libruntime {
struct SchedulerInstance {
    std::string InstanceName;
    std::string InstanceID;
    long long updateTime;
    bool isAvailable = true;
};

struct SchedulerInfo {
    std::string schedulerFuncKey;
    std::vector<SchedulerInstance> schedulerInstanceList;
};

struct AvailableSchedulerInfos {
    std::vector<std::shared_ptr<SchedulerInstance>> schedulerInstanceList;
    mutable absl::Mutex schedulerMtx;
};

void from_json(const nlohmann::json &j, std::vector<SchedulerInstance> &schedulerInstanceList);

ErrorInfo ParseSchedulerInfo(const std::string &payload, SchedulerInfo &schedulerInfo);

}  // namespace Libruntime
}  // namespace YR