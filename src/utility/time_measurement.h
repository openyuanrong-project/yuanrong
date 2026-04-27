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

#include <chrono>
#include <list>
#include "src/utility/logger/logger.h"

namespace YR {
namespace utility {

class TimeMeasurement {
public:
    explicit TimeMeasurement(int64_t defaultDuration);
    TimeMeasurement() = default;
    int64_t GetEstimatedCostofNextRequest();
    void StartTimer(const std::string &reqID);
    void StopTimer(const std::string &reqID, bool isSuccessfulInvoke);

private:
    std::list<std::string> lastFiveReqIDs;
    int64_t defaultDuration_;
    std::unordered_map<std::string, int64_t> durationMap;
    std::unordered_map<std::string, std::chrono::time_point<std::chrono::steady_clock>> startTimeMap;
};

}  // namespace utility
}  // namespace YR
