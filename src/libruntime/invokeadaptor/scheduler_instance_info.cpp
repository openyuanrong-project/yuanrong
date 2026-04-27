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

#include "scheduler_instance_info.h"

namespace YR {
namespace Libruntime {
using nlohmann::json;
using namespace YR::utility;


const static std::string SCHEDULER_FUNC_KEY = "schedulerFuncKey";
const static std::string SCHEDULER_INSTANCE_LIST = "schedulerInstanceList";

void from_json(const json &j, std::vector<SchedulerInstance> &schedulerInstanceList)
{
    for (auto &info : j) {
        SchedulerInstance instanceInfo;
        JsonGetTo(info, "instanceName", instanceInfo.InstanceName);
        JsonGetTo(info, "instanceId", instanceInfo.InstanceID);
        schedulerInstanceList.push_back(instanceInfo);
    }
}

ErrorInfo ParseSchedulerInfo(const std::string &payload, SchedulerInfo &schedulerInfo)
{
    try {
        json j = json::parse(payload);
        schedulerInfo.schedulerFuncKey = j[SCHEDULER_FUNC_KEY].get<std::string>();
        schedulerInfo.schedulerInstanceList = j[SCHEDULER_INSTANCE_LIST].get<std::vector<SchedulerInstance>>();
    } catch (std::exception &e) {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID,
                         std::string("parse schedulerInfo info: ") + e.what());
    }
    return ErrorInfo();
}

}  // namespace Libruntime
}  // namespace YR