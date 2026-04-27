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
#include <vector>
#include "src/dto/affinity.h"
#include "src/dto/debug_config.h"
#include "src/dto/device.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/stacktrace/stack_trace_info.h"
#include "src/libruntime/utils/constants.h"
#include "src/proto/libruntime.pb.h"
#include "src/libruntime/fsclient/protobuf/core_service.pb.h"
#include "src/libruntime/fsclient/protobuf/common.pb.h"

namespace YR {
namespace Libruntime {
const size_t FAAS_DEFALUT_INVOKE_TIMEOUT = 900;   // second
const size_t FAAS_DEFALUT_ACQUIRE_TIMEOUT = 120;  // second

enum class BundleAffinity : int {
    COMPACT,
    DISCRETE,
};

struct FunctionGroupOptions {
    int functionGroupSize{0};
    int bundleSize{0};
    BundleAffinity bundleAffinity{BundleAffinity::COMPACT};
    int timeout{NO_TIMEOUT};
    bool sameLifecycle = true;
};

struct ResourceGroupOptions {
    std::string resourceGroupName{""};
    int bundleIndex{-1};
};

struct RangeOptions {
    int timeout = NO_TIMEOUT;
};

struct InstanceRange {
    int min = DEFAULT_INSTANCE_RANGE_NUM;
    int max = DEFAULT_INSTANCE_RANGE_NUM;
    int step = DEFAULT_INSTANCE_RANGE_STEP;
    bool sameLifecycle = true;
    RangeOptions rangeOpts;
};

struct InstanceSession {
    std::string sessionID = "";
    int sessionTTL;
    int concurrency;
};

struct FaasInvokeData {
    FaasInvokeData() = default;
    FaasInvokeData(const std::string &teId, const std::string &funcName, const std::string &inputAliAs,
                   const std::string &inputTraceId, const long long inputSubmitTime)
        : tenantId(teId),
          functionName(funcName),
          aliAs(inputAliAs),
          traceId(inputTraceId),
          submitTime(inputSubmitTime) {};
    std::string businessId;
    std::string tenantId;
    std::string srcAppId;
    std::string functionName;
    std::string aliAs;
    std::string version;
    std::string traceId;
    std::string code;
    std::string innerCode;
    std::string describeMsg;
    long long submitTime = 0;
    long long sendTime = 0;
    long long endTime = 0;
};

struct InvokeOptions {
    int cpu = 500;

    int memory = 500;

    std::unordered_map<std::string, float> customResources;

    std::unordered_map<std::string, std::string> customExtensions;

    std::unordered_map<std::string, std::string> createOptions;

    std::unordered_map<std::string, std::string> podLabels;

    std::vector<std::string> labels;

    std::unordered_map<std::string, std::string> affinity;

    std::list<std::shared_ptr<Affinity>> scheduleAffinities;

    size_t retryTimes = 0;

    int maxRetryTime = -1;

    std::function<bool(const ErrorInfo &errInfo)> retryChecker = nullptr;

    size_t priority = 0;

    int instancePriority = 0;

    std::vector<std::string> codePaths;

    std::string schedulerFunctionId;

    std::vector<std::string> schedulerInstanceIds;

    std::string traceId;

    Device device;

    int maxInvokeLatency = 5000;

    int minInstances = 0;

    int maxInstances = 0;

    int groupTimeout = -1;

    std::string groupName;

    bool isDataAffinity = false;

    bool needOrder = false;

    int64_t scheduleTimeoutMs = 30000;

    bool preemptedAllowed = false;

    int recoverRetryTimes = 0;

    int timeout = 0;

    int acquireTimeout = 0;

    bool trafficLimited;

    bool forceInvoke = false;

    InstanceRange instanceRange;

    ResourceGroupOptions resourceGroupOpts;

    FunctionGroupOptions functionGroupOpts;

    std::unordered_map<std::string, std::string> envVars;

    bool isGetInstance = false;

    bool isDeleteRemoteTensor = false;

    std::unordered_map<std::string, std::string> invokeLabels;

    std::unordered_map<std::string, std::string> aliasParams;

    std::shared_ptr<InstanceSession> instanceSession;

    DebugConfig debug;

    std::string workingDir;

    bool isInterrupted = false;
};

struct FunctionMeta {
    std::string appName;
    std::string moduleName;
    std::string funcName;
    std::string className;
    libruntime::LanguageType languageType;
    std::string codeId;     // transferring python code with data object, use codeId as key to it
    std::string signature;  // java function signature
    std::string poolLabel;
    libruntime::ApiType apiType;
    std::string functionId;
    std::string name;
    std::string ns;
    std::string initializerCodeId;
    bool isAsync = false;
    bool isGenerator = false;
    bool needOrder = false;
    std::string tensorTransportTarget = "";
    bool enableTensorTransport = false;
    std::vector<char> code;

    /// Serialized instance state from recover (e.g. for GetInstance on caller).
    std::string recoveredData;
    bool IsServiceApiType()
    {
        return (apiType == libruntime::ApiType::Faas or apiType == libruntime::ApiType::Serve);
    }
};

struct BindOpts {
    std::string resource = "NONE";
    std::string strategy = "NONE";
};

struct GroupOpts {
    std::string groupName;
    int timeout;
    bool sameLifecycle = true;
    std::string strategy;
    BindOpts bind;
};

struct InstanceOptions {
    bool needOrder;
};

struct DoubleCounterData {
    std::string name;
    std::string description;
    std::string unit;
    std::unordered_map<std::string, std::string> labels;
    double value;
};

struct UInt64CounterData {
    std::string name;
    std::string description;
    std::string unit;
    std::unordered_map<std::string, std::string> labels;
    uint64_t value;
};

struct GaugeData {
    std::string name;
    std::string description;
    std::string unit;
    std::unordered_map<std::string, std::string> labels;
    double value;
};

enum class AlarmSeverity { OFF, INFO, MINOR, MAJOR, CRITICAL };

struct AlarmInfo {
    std::string id;
    std::string alarmName;
    AlarmSeverity alarmSeverity = AlarmSeverity::OFF;
    std::string locationInfo;
    std::string cause;
    long startsAt = DEFAULT_ALARM_TIMESTAMP;
    long endsAt = DEFAULT_ALARM_TIMESTAMP;
    long timeout = DEFAULT_ALARM_TIMEOUT;
    std::unordered_map<std::string, std::string> customOptions;
};

struct SnapOptions {
    common::SnapType type = common::SnapType::SNAPSHOT;
    int32_t ttl = -1;  // seconds
    bool leaveRunning = false;
};

struct SnapStartOptions {
    common::SnapType type = common::SnapType::SNAPSHOT;
    // Reserved for future use - can add SchedulingOptions later
};

struct Credential {
    std::string ak;
    std::string sk;
    std::string dk;
};

struct OwnerSchedulerInfo {
    std::string schedulerInstanceID;
    int retryTimes = 0;
    int maxRetryTimes = 3;  // 单个scheduler实例重试3次
    int currentRetryTimeSpent = 0;
    int maxRetryTimeSpent = 5000;  // Owner节点总重试时间最大5s
};
}  // namespace Libruntime
}  // namespace YR

namespace std {
template <>
class hash<YR::Libruntime::ResourceGroupOptions> {
public:
    size_t operator()(const YR::Libruntime::ResourceGroupOptions &d) const
    {
        return std::hash<std::string>()(d.resourceGroupName) ^ std::hash<int>()(d.bundleIndex);
    }
};
}  // namespace std