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
#include <unordered_map>
#include <unordered_set>

#include "src/dto/acquire_options.h"
#include "src/libruntime/invokeadaptor/alias_routing.h"
#include "src/libruntime/invokeadaptor/instance_manager.h"
#include "src/libruntime/invokeadaptor/request_manager.h"
#include "src/libruntime/invokeadaptor/request_queue.h"
#include "src/libruntime/invokeadaptor/task_scheduler.h"
#include "src/libruntime/metricsadaptor/metrics_adaptor.h"
#include "src/libruntime/objectstore/memory_store.h"
#include "src/libruntime/utils/utils.h"
#include "src/scene/downgrade.h"
#include "src/utility/time_measurement.h"
namespace YR {
namespace Libruntime {
using KillFunc = std::function<ErrorInfo(const std::string &, const std::string &, int32_t)>;
using CancelFunc = std::function<void(const std::string &, const std::string &, int32_t)>;
using YR::utility::TimeMeasurement;
class TaskSubmitter : public std::enable_shared_from_this<TaskSubmitter> {
public:
    TaskSubmitter() = default;
    TaskSubmitter(std::shared_ptr<LibruntimeConfig> config, std::shared_ptr<MemoryStore> store,
                  std::shared_ptr<FSClient> client, std::shared_ptr<RequestManager> reqMgr, CancelFunc cancelFunc,
                  std::shared_ptr<AliasRouting> ar = nullptr, std::shared_ptr<MetricsAdaptor> adaptor = nullptr,
                  std::shared_ptr<YR::scene::DowngradeController> downgrade = nullptr);
    virtual ~TaskSubmitter();
    void SubmitFunction(std::shared_ptr<InvokeSpec> spec);
    void ScheduleFunction(const RequestResource &resource);
    void ScheduleFunctions();
    void HandleInvokeNotify(const NotifyRequest &req, const ErrorInfo &err);
    void HandleFailInvokeNotify(const NotifyRequest &req, std::shared_ptr<InvokeSpec> spec,
                                const RequestResource &resource, const ErrorInfo &err);
    void HandleSuccessInvokeNotify(const NotifyRequest &req, std::shared_ptr<InvokeSpec> spec,
                                   const RequestResource &resource);
    bool NeedRetry(const ErrorInfo &errInfo, const std::shared_ptr<InvokeSpec> spec, bool &isConsumeRetryTime);
    bool NeedRetryCreate(const ErrorInfo &errInfo);
    virtual ErrorInfo CancelStatelessRequest(std::shared_ptr<InvokeSpec> spec, const KillFunc &killCallBack, bool isForce,
                                     bool isRecursive, const std::string &objId);
    void Finalize(void);
    std::vector<std::string> GetInstanceIds();
    std::vector<std::string> GetCreatingInsIds();
    void UpdateConfig();
    void Init();
    virtual std::pair<InstanceAllocation, ErrorInfo> AcquireInstance(const std::string &stateId,
                                                                     std::shared_ptr<InvokeSpec> spec);
    virtual ErrorInfo ReleaseInstance(const std::string &leaseId, const std::string &stateId, bool abnormal,
                                      std::shared_ptr<InvokeSpec> spec);
    virtual void UpdateFaaSSchedulerInfo(std::string schedulerFuncKey,
                                         const std::vector<SchedulerInstance> &schedulerInstanceList);
    void RecordFaasInvokeData(const std::shared_ptr<InvokeSpec> spec);
    void CancelFaasScheduleTimeoutReq(const std::string &reqId, int timeoutSeconds);
    void EraseFaasCancelTimer(const std::string &reqId);
    virtual void UpdateSchdulerInfo(const std::string &schedulerName, const std::string &schedulerId,
                                    const std::string &option);
    YR::Libruntime::GaugeData ConvertToGaugeData(const std::shared_ptr<FaasInvokeData> data, const std::string &reqId);
    void UpdateFaasInvokeSendTime(const std::string &reqId);
    void UpdateFaasInvokeLog(const std::string &reqId, const ErrorInfo &err);

private:
    void EraseInsResourceMap();
    std::vector<RequestResource> GetScheduleResources();
    void AddFaasCancelTimer(std::shared_ptr<InvokeSpec> spec);
    void EraseTaskSchedulerMap(const RequestResource &resource);
    bool CancelAndScheOtherRes(const RequestResource &resource);
    std::shared_ptr<TaskSchedulerWrapper> GetTaskScheduler(const RequestResource &resource);
    std::shared_ptr<TaskSchedulerWrapper> GetOrAddTaskScheduler(const RequestResource &resource);
    void ScheduleIns(const RequestResource &resource, const ErrorInfo &err, bool isRemainIns);
    void NotifyScheduler(const RequestResource &resource, const ErrorInfo &err = ErrorInfo());
    bool HandleFailInvokeIsDelayScaleDown(const NotifyRequest &req, const ErrorInfo &err);
    void DeleteInsCallback(const std::string &instanceId);
    bool ScheduleRequest(const RequestResource &resource, std::shared_ptr<BaseQueue> requestQueue);
    void SendInvokeReq(const RequestResource &resource, std::shared_ptr<InvokeSpec> invokeSpec);
    void DowngradeCallback(const std::string &requestId, Libruntime::ErrorCode code, const std::string &result);

    std::shared_ptr<LibruntimeConfig> libRuntimeConfig;
    std::atomic<bool> runFlag{true};
    mutable absl::Mutex reqMtx_;
    mutable absl::Mutex cancelTimerMtx_;
    std::shared_ptr<MemoryStore> memoryStore;
    std::shared_ptr<FSClient> fsClient;
    std::shared_ptr<RequestManager> requestManager;
    int recycleTimeMs;
    std::unordered_map<libruntime::ApiType, std::shared_ptr<InsManager>> insManagers;
    std::unordered_map<std::string, TimeMeasurement> invokeCostMap ABSL_GUARDED_BY(invokeCostMtx_);
    mutable absl::Mutex invokeCostMtx_;
    std::unordered_map<RequestResource, std::shared_ptr<TaskSchedulerWrapper>, HashFn> taskSchedulerMap_
        ABSL_GUARDED_BY(reqMtx_);
    std::unordered_map<std::string, std::shared_ptr<YR::utility::Timer>> faasCancelTimerWorkers
        ABSL_GUARDED_BY(cancelTimerMtx_);
    CancelFunc cancelCb;
    std::unordered_map<std::string, std::shared_ptr<FaasInvokeData>> faasInvokeDataMap_ ABSL_GUARDED_BY(invokeDataMtx_);
    mutable absl::Mutex invokeDataMtx_;
    std::shared_ptr<AliasRouting> ar_;
    std::shared_ptr<MetricsAdaptor> metricsAdaptor_;
    std::atomic<bool> metricsEnable_{false};
    std::shared_ptr<TaskScheduler> taskScheduler_;
    bool enablePriority_ = false;
    std::shared_ptr<YR::scene::DowngradeController> downgrade_;
    std::shared_ptr<YR::utility::Timer> eraseTimer_;
};

}  // namespace Libruntime
}  // namespace YR