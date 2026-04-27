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

#include "report_record.h"

void ReportRecord::RecordAbnormal()
{
    std::lock_guard<std::mutex> insLk(reportMtx);
    isAbnormal = true;
}

void ReportRecord::RecordRequest(int64_t inputDuration)
{
    std::lock_guard<std::mutex> insLk(reportMtx);
    requestsCount++;
    totalDuration += inputDuration;
    if (inputDuration > maxDuration) {
        maxDuration = inputDuration;
    }
}

InstanceReport ReportRecord::Report(bool reset)
{
    std::lock_guard<std::mutex> insLk(reportMtx);
    InstanceReport report{requestsCount, maxDuration, isAbnormal};
    if (requestsCount == 0) {
        report.avgProcTime = -1;
    } else {
        report.avgProcTime = totalDuration / requestsCount;
    }

    if (reset) {
        requestsCount = 0;
        totalDuration = 0;
    }
    return report;
}

int64_t ReportRecord::GetTotalDuration()
{
    return totalDuration;
}

bool ReportRecord::IsAbnormal()
{
    return isAbnormal;
}
