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

#include "src/utility/time_measurement.h"

namespace YR {
namespace utility {
const int LIST_SIZE = 5;

TimeMeasurement::TimeMeasurement(int64_t defaultDuration) : defaultDuration_(defaultDuration)
{
}

void TimeMeasurement::StartTimer(const std::string &reqID)
{
    startTimeMap[reqID] = std::chrono::steady_clock::now();
}

void TimeMeasurement::StopTimer(const std::string &reqID, bool isSuccessfulInvoke)
{
    if (startTimeMap.find(reqID) == startTimeMap.end()) {
        YRLOG_ERROR("no start time record for req: {}", reqID);
        return;
    }

    if (isSuccessfulInvoke) {
        auto start = startTimeMap[reqID];
        auto end = std::chrono::steady_clock::now();
        auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start).count();
        durationMap[reqID] = duration;
        lastFiveReqIDs.push_back(reqID);
        YRLOG_DEBUG("{} cost {} ms", reqID, duration);
    }
    
    startTimeMap.erase(reqID);

    if (lastFiveReqIDs.size() > LIST_SIZE) {
        lastFiveReqIDs.pop_front();
        if (durationMap.find(reqID) != durationMap.end()) {
            durationMap.erase(reqID);
        }
    }
}

int64_t TimeMeasurement::GetEstimatedCostofNextRequest()
{
    int64_t sum = 0;
    for (auto &pair: durationMap) {
        sum += pair.second;
    }
    int64_t size = durationMap.size();
    if (size != 0) {
        return sum/size;
    }
    return defaultDuration_;
}

}  // namespace utility
}  // namespace YR
