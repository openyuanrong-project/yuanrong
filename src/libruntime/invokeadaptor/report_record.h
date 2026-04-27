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

struct InstanceReport {
    int64_t procReqNum;
    int64_t avgProcTime;
    int64_t maxProcTime;
    bool isAbnormal;
    InstanceReport() {}
    InstanceReport(int64_t inputProcReqNum, int64_t inputMaxProcTime, bool inputIsAbnormal)
        : procReqNum(inputProcReqNum), maxProcTime(inputMaxProcTime), isAbnormal(inputIsAbnormal)
    {
    }
};

class ReportRecord {
public:
    void RecordAbnormal();
    void RecordRequest(int64_t inputDuration);
    InstanceReport Report(bool reset);
    int64_t GetTotalDuration();
    bool IsAbnormal();

private:
    // the requests completed at the current report period
    int64_t requestsCount;
    // the total time spent by the requests completed at the current report period, ms
    int64_t totalDuration;
    // the max of the time spent by all the requests yet, ms
    int64_t maxDuration;
    bool isAbnormal = false;
    std::mutex reportMtx;
};