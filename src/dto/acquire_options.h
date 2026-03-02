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
#include <string>
#include <unordered_map>

namespace YR {
namespace Libruntime {

struct AcquireOption {
    std::string designatedInstanceID;
    std::string schedulerFunctionID;
    std::string schedulerInstanceID;
    std::string poolLabel;
    std::string requestId;
    std::string traceId;
    std::unordered_map<std::string, u_int64_t> resourceSpecs;
    size_t timeout;
};

struct InstanceAllocation {
    std::string functionId;
    std::string funcSig;
    std::string instanceId;
    std::string leaseId;
    int tLeaseInterval;
    bool forceInvoke;
};

struct InstanceResponse {
    InstanceAllocation info;
    int errorCode;  //`json:"errorCode"`
    std::string errorMessage; //`json:"errorMessage"`
    float schedulerTime; // `json:"schedulerTime"`
};

struct InstanceAllocationFailedRsp {
    int errorCode;  //`json:"errorCode"`
    std::string errorMessage; //`json:"errorMessage"`
};

struct BatchInstanceResponse {
    std::unordered_map<std::string, InstanceAllocation> instanceAllocSucceed;
    std::unordered_map<std::string, InstanceAllocationFailedRsp> instanceAllocFailed;
    int tLeaseInterval;
    float schedulerTime;  // `json:"schedulerTime"`
};
}  // namespace Libruntime
}  // namespace YR