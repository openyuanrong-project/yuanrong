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

#include "src/libruntime/err_type.h"
#include "src/libruntime/fsclient/fs_client.h"
#include "src/libruntime/invoke_spec.h"
#include "src/libruntime/invokeadaptor/limiter_consistant_hash.h"
#include "src/libruntime/invokeadaptor/request_manager.h"
#include "src/libruntime/invokeadaptor/request_queue.h"
#include "src/libruntime/objectstore/memory_store.h"
#include "src/utility/time_measurement.h"
#include "src/utility/timer_worker.h"

namespace YR {
namespace Libruntime {
const int DEFAULT_INVOKE_DURATION = 1000;
const int DEFAULT_CREATE_DURATION = 1000;  // ms
const int64_t DEFAULT_LEASE_INTERVAL = 500;
const int REQUEST_RESOURCE_USE_COUNT = 3;

using YR::utility::TimeMeasurement;
using ScheduleInsCallback = std::function<void(const RequestResource &resource, const ErrorInfo &err, bool isRemainIs)>;

void CancelScaleDownTimer(std::shared_ptr<InstanceInfo> insInfo);
using DeleteInsCallback = std::function<void(const std::string &instanceId)>;
class InsManager {
public:
    enum class UpdateSchedulerOption { ADD, REMOVE, UNKNOWN };
    UpdateSchedulerOption stringToOption(const std::string &s)
    {
        static const std::unordered_map<std::string, UpdateSchedulerOption> mapping = {
            {"ADD", UpdateSchedulerOption::ADD}, {"REMOVE", UpdateSchedulerOption::REMOVE}};
        auto it = mapping.find(s);
        return (it != mapping.end()) ? it->second : UpdateSchedulerOption::UNKNOWN;
    }
    InsManager()
    {
        tw_ = std::make_shared<YR::utility::TimerWorker>();
    };
    InsManager(ScheduleInsCallback cb, std::shared_ptr<FSClient> client, std::shared_ptr<MemoryStore> store,
               std::shared_ptr<RequestManager> reqMgr, std::shared_ptr<LibruntimeConfig> config)
        : scheduleInsCb(cb), fsClient(client), memoryStore(store), requestManager(reqMgr), libRuntimeConfig(config)
    {
        tw_ = std::make_shared<YR::utility::TimerWorker>();
    }
    ~InsManager() = default;
    void DelInsInfo(const std::string &insId, const RequestResource &resource);
    InstanceSummary GetAvailableIns(const RequestResource &resource);
    virtual bool ScaleUp(std::shared_ptr<InvokeSpec> spec, size_t reqNum) = 0;
    virtual void ScaleDown(const std::shared_ptr<InvokeSpec> spec, bool isInstanceNormal = false) = 0;
    virtual void ScaleCancel(const RequestResource &resource, size_t reqNum, bool cleanAll = false) = 0;
    virtual void StartBatchRenewTimer() = 0;
    virtual void UpdateConfig(int recycleTimeMs) = 0;
    virtual void UpdateSchedulerInfo(const std::string &schedulerFuncKey,
                                     const std::vector<SchedulerInstance> &schedulerInfoList)
    {
    }
    void DecreaseUnfinishReqNum(const std::shared_ptr<InvokeSpec> spec, bool isInstanceNormal = true);
    void Stop();
    std::vector<std::string> GetInstanceIds();
    std::vector<std::string> GetCreatingInsIds();
    void SetDeleleInsCallback(const DeleteInsCallback &cb);
    void UpdateSchdulerInfo(const std::string &schedulerName, const std::string &schedulerId,
                            const std::string &option);
    void EraseResourceInfoMap();
protected:
    int recycleTimeMs;
    ScheduleInsCallback scheduleInsCb;
    std::shared_ptr<FSClient> fsClient;
    std::shared_ptr<MemoryStore> memoryStore;
    std::shared_ptr<RequestManager> requestManager;
    std::shared_ptr<LibruntimeConfig> libRuntimeConfig;
    DeleteInsCallback deleteInsCallback_;
    void AddRequestResourceInfo(std::shared_ptr<InvokeSpec> spec);
    void DelInsInfoBare(const std::string &insId, std::shared_ptr<RequestResourceInfo> info);
    std::pair<bool, std::vector<std::string>> NeedCancelCreatingIns(const RequestResource &resource, size_t reqNum,
                                                                    bool cleanAll);
    std::pair<bool, size_t> NeedCreateNewIns(const RequestResource &resource, size_t reqNum,
                                             bool considerWithFailNum = true);
    void AddCreatingInsInfo(const RequestResource &resource, std::shared_ptr<CreatingInsInfo> insInfo);
    bool EraseCreatingInsInfo(const RequestResource &resource, const std::string &instanceId,
                              bool createSuccess = true);
    // erase creating instance info without locking; the caller must ensure locking.
    bool EraseCreatingInsInfoBare(std::shared_ptr<RequestResourceInfo> info, const std::string &instanceId,
                                  bool createSuccess = true);
    void ChangeCreateFailNum(const RequestResource &resource, bool isIncreaseOps);
    int GetRequiredInstanceNum(int reqNum, int concurrency) const;
    bool IsRemainIns(const RequestResource &resource);
    std::shared_ptr<RequestResourceInfo> GetRequestResourceInfo(const RequestResource &resource);
    std::shared_ptr<RequestResourceInfo> GetOrAddRequestResourceInfo(const RequestResource &resource);
    std::shared_ptr<InstanceInfo> GetInstanceInfo(const RequestResource &resource, const std::string &insId);
    int GetCreatedInstanceNum()
    {
        absl::ReaderMutexLock lock(&createInstanceNumMutex);
        return totalCreatedInstanceNum_;
    }
    int GetCreatingInstanceNum()
    {
        absl::ReaderMutexLock lock(&createInstanceNumMutex);
        return totalCreatingInstanceNum_;
    }
    int GetCreateInstanceNum()
    {
        absl::ReaderMutexLock lock(&createInstanceNumMutex);
        return totalCreatedInstanceNum_ + totalCreatingInstanceNum_;
    }
    void DecreaseCreatedInstanceNum()
    {
        absl::WriterMutexLock lock(&createInstanceNumMutex);
        totalCreatedInstanceNum_--;
    }
    void DecreaseCreatedInstanceNums(int instanceNum)
    {
        absl::WriterMutexLock lock(&createInstanceNumMutex);
        totalCreatedInstanceNum_ -= instanceNum;
    }
    void DecreaseCreatingInstanceNum()
    {
        absl::WriterMutexLock lock(&createInstanceNumMutex);
        totalCreatingInstanceNum_--;
    }
    void IncreaseCreatedInstanceNum()
    {
        absl::WriterMutexLock lock(&createInstanceNumMutex);
        totalCreatedInstanceNum_++;
    }
    void IncreaseCreatingInstanceNum()
    {
        absl::WriterMutexLock lock(&createInstanceNumMutex);
        totalCreatingInstanceNum_++;
    }
    std::vector<RequestResource> GetScheduleResources();
    void EraseResourceInfoMap(const RequestResource &resource, int currentCount = 2);
    mutable absl::Mutex insMtx;
    std::atomic<bool> runFlag{true};
    std::shared_ptr<LimiterCsHash> csHash;
    int totalCreatedInstanceNum_{0} ABSL_GUARDED_BY(createInstanceNumMutex);
    int totalCreatingInstanceNum_{0} ABSL_GUARDED_BY(createInstanceNumMutex);
    mutable absl::Mutex createInstanceNumMutex;
    std::unordered_map<std::string, YR::utility::TimeMeasurement> createCostMap ABSL_GUARDED_BY(createCostMtx);
    mutable absl::Mutex createCostMtx;
    std::unordered_map<std::string, YR::utility::TimeMeasurement> invokeCostMap ABSL_GUARDED_BY(invokeCostMtx);
    mutable absl::Mutex invokeCostMtx;
    std::unordered_map<RequestResource, std::shared_ptr<RequestResourceInfo>, HashFn> requestResourceInfoMap
        ABSL_GUARDED_BY(insMtx);
    std::unordered_map<std::string, RequestResource> globalLeases ABSL_GUARDED_BY(leaseMtx);
    int64_t tLeaseInterval{DEFAULT_LEASE_INTERVAL} ABSL_GUARDED_BY(leaseMtx);
    std::shared_ptr<YR::utility::Timer> leaseTimer ABSL_GUARDED_BY(leaseMtx);
    mutable absl::Mutex leaseMtx;
    mutable absl::Mutex schedulerMtx;
    std::shared_ptr<YR::utility::TimerWorker> tw_;

private:
    InstanceSummary ScheduleInsWithDevice(const RequestResource &resource,
                                          std::shared_ptr<RequestResourceInfo> resourceInfo);
};

RequestResource GetRequestResource(std::shared_ptr<InvokeSpec> spec);
size_t GetDelayTime(size_t failedCnt);
ErrorInfo ConvertStringToInsResp(InstanceResponse &resp, const std::string &bufStr);
void ConvertStringToBatchInsResp(BatchInstanceResponse &resp, const std::string &bufStr);

}  // namespace Libruntime
}  // namespace YR
