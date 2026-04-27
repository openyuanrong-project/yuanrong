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
#include "src/libruntime/invokeadaptor/instance_manager.h"

namespace YR {
namespace Libruntime {

std::pair<InstanceResponse, ErrorInfo> GetFaasInstanceRsp(const NotifyRequest &notifyReq);

class FaasInsManager : public InsManager, public std::enable_shared_from_this<FaasInsManager> {
public:
    FaasInsManager() = default;
    FaasInsManager(ScheduleInsCallback cb, std::shared_ptr<FSClient> client, std::shared_ptr<MemoryStore> store,
                   std::shared_ptr<RequestManager> reqMgr, std::shared_ptr<LibruntimeConfig> config)
        : InsManager(cb, client, store, reqMgr, config)
    {
        std::shared_ptr<LoadBalancer> lb(LoadBalancer::Factory(LoadBalancerType::ConsistantRoundRobin));
        this->csHash = std::make_shared<LimiterCsHash>(lb);
    }
    bool ScaleUp(std::shared_ptr<InvokeSpec> spec, size_t reqNum) override;
    void ScaleDown(const std::shared_ptr<InvokeSpec> spec, bool isInstanceNormal = false) override;
    void ScaleCancel(const RequestResource &resource, size_t reqNum, bool cleanAll = false) override;
    void StartBatchRenewTimer() override;
    virtual void UpdateConfig(int recycleTimeMs) override;
    void UpdateSchedulerInfo(const std::string &schedulerFuncKey,
                             const std::vector<SchedulerInstance> &schedulerInstanceList) override;
    void RecordRequest(const RequestResource &resource, const std::shared_ptr<InvokeSpec> spec, bool isInstanceNormal);
    void DelRelatedLease(const std::string &insId, const RequestResource &resource);
    std::pair<InstanceAllocation, ErrorInfo> AcquireInstance(const std::string &stateId,
                                                             std::shared_ptr<InvokeSpec> spec);
    void ProcessInstanceInfo(std::shared_ptr<InvokeSpec> spec, const InstanceAllocation &inst);
    ErrorInfo ReleaseInstance(const std::string &leaseId, const std::string &stateId, bool abnormal,
                              std::shared_ptr<InvokeSpec> spec);
    void ProcessBatchRenewResult(const NotifyRequest &notifyReq, const std::string &functionId, const ErrorInfo &err,
                                 std::vector<std::string> leaseIds);
    // add instance info without locking; the caller must ensure locking.
    void AddInsInfoBare(std::shared_ptr<RequestResourceInfo> info, std::shared_ptr<InstanceInfo> &faasInsInfo);
    void UpdateSpecSchedulerIds(std::shared_ptr<InvokeSpec> spec, const std::string &schedulerId);
    void AcquireCallback(const std::shared_ptr<InvokeSpec> acquireSpec, const ErrorInfo &errInfo,
                         const InstanceResponse &resp, const std::shared_ptr<InvokeSpec> invokeSpec);

private:
    std::shared_ptr<YR::utility::Timer> CreateBatchRenewTimer();
    void RenewHandler(std::shared_ptr<InstanceInfo> insInfo);
    void BatchRenewHandler();
    void SendAcquireReq(const std::shared_ptr<InvokeSpec> spec, size_t delayTime);
    std::shared_ptr<InvokeSpec> BuildAcquireRequest(std::shared_ptr<InvokeSpec> invokeSpec,
                                                    const std::string &stateId = "");
    InvokeRequest BuildReleaseReq(const std::shared_ptr<InstanceInfo> &ins);
    bool AcquireFaasIns(const std::shared_ptr<InvokeSpec> spec, size_t reqNum);
    void HandleFaasInsInfo(std::shared_ptr<InstanceInfo> &faasInsInfo, const RequestResource &resource);
    void ProcessAsynAcquireResult(const NotifyRequest &notifyReq, std::shared_ptr<InvokeSpec> acquireSpec,
                                  const ErrorInfo &errInput, std::shared_ptr<InvokeSpec> invokeSpec);
    void ProcecssAcquireResult(const NotifyRequest &req, std::shared_ptr<InvokeSpec> spec,
                               std::shared_ptr<std::promise<std::pair<InstanceAllocation, ErrorInfo>>> acquirePromise);
    void AcquireInstanceAsync(std::shared_ptr<InvokeSpec> spec);
    void StartReleaseTimer(const RequestResource &resource, const std::string &leaseId);
    void ReleaseHandler(const RequestResource &resource, const std::string &leaseId);
    void ReleaseInstanceAsync(const std::shared_ptr<InstanceInfo> &ins);
    ErrorInfo SendReleaseInstanceReq(const std::shared_ptr<InstanceInfo> &ins, std::shared_ptr<InvokeSpec> spec);
    std::string GetSchedulerKey();
    std::string GetFunctionIdWithLabel(const RequestResource &resource);
    void ChangeInstanceSchedulerId(const std::string &functionId, std::vector<std::string> &leaseIds);
    mutable absl::Mutex schedulerFuncKeyMtx;
    std::string schedulerFuncKey;
};
}  // namespace Libruntime
}  // namespace YR