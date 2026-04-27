/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include <boost/format.hpp>
#include <json.hpp>
#include <msgpack.hpp>
#include <optional>

#include <google/protobuf/message.h>
#include <google/protobuf/util/json_util.h>
#include "src/dto/config.h"
#include "src/dto/constant.h"
#include "src/libruntime/fsclient/protobuf/core_service.grpc.pb.h"
#include "src/libruntime/invoke_spec.h"
#include "src/libruntime/invokeadaptor/faas_instance_manager.h"
#include "src/libruntime/invokeadaptor/normal_instance_manager.h"
#include "src/libruntime/utils/constants.h"
#include "src/libruntime/utils/exception.h"
#include "src/libruntime/utils/serializer.h"
#include "src/proto/libruntime.pb.h"
#include "src/scene/downgrade.h"
#include "src/utility/id_generator.h"
#include "src/utility/timer_worker.h"
#include "task_submitter.h"
namespace YR {
namespace Libruntime {
const std::string INSTANCE_REQUIREMENT_RESOURKEY = "resourcesData";
const std::string INSTANCE_REQUIREMENT_INSKEY = "designateInstanceID";
const std::string INSTANCE_REQUIREMENT_POOLLABELKEY = "poolLabel";
const int64_t BEFOR_RETAIN_TIME = 30;           // millisecond
const int SECONDS_TO_MILLISECONDS_UNIT = 1000;  // millisecond
const int64_t IDLE_TIMER_INTERNAL = 10;
const int DEFALUT_CANCEL_DELAY_TIME = 5;  // second
const int ERASE_DELAY_TIME = 30;
const static int EVENT_INFO_TIMEOUT = 30000;  // millisecond
using namespace YR::utility;
using json = nlohmann::json;

ErrorInfo PackageNotifyErr(const NotifyRequest &notifyReq, bool isCreate)
{
    std::vector<YR::Libruntime::StackTraceInfo> stackTraceInfos = GetStackTraceInfos(notifyReq);
    YRLOG_DEBUG("get stackTraceInfos from notify request in task_submitter package size: {}", stackTraceInfos.size());
    ErrorInfo errInfo(static_cast<ErrorCode>(notifyReq.code()), ModuleCode::CORE, notifyReq.message(), isCreate,
                      stackTraceInfos);
    return errInfo;
}

TaskSubmitter::TaskSubmitter(std::shared_ptr<LibruntimeConfig> config, std::shared_ptr<MemoryStore> store,
                             std::shared_ptr<FSClient> client, std::shared_ptr<RequestManager> reqMgr,
                             CancelFunc cancelFunc, std::shared_ptr<AliasRouting> ar,
                             std::shared_ptr<MetricsAdaptor> adaptor,
                             std::shared_ptr<YR::scene::DowngradeController> downgrade)
    : libRuntimeConfig(config),
      memoryStore(store),
      fsClient(client),
      requestManager(reqMgr),
      cancelCb(cancelFunc),
      ar_(ar),
      metricsAdaptor_(adaptor),
      downgrade_(downgrade)
{
    enablePriority_ = Config::Instance().ENABLE_PRIORITY();
    this->Init();
}

void TaskSubmitter::Init()
{
    auto normalInsManager =
        std::make_shared<NormalInsManager>(std::bind(&TaskSubmitter::ScheduleIns, this, std::placeholders::_1,
                                                     std::placeholders::_2, std::placeholders::_3),
                                           fsClient, memoryStore, requestManager, libRuntimeConfig);
    normalInsManager->SetDeleleInsCallback(std::bind(&TaskSubmitter::DeleteInsCallback, this, std::placeholders::_1));
    auto faasInsManager =
        std::make_shared<FaasInsManager>(std::bind(&TaskSubmitter::ScheduleIns, this, std::placeholders::_1,
                                                   std::placeholders::_2, std::placeholders::_3),
                                         fsClient, memoryStore, requestManager, libRuntimeConfig);
    insManagers[libruntime::ApiType::Faas] = faasInsManager;
    insManagers[libruntime::ApiType::Serve] = faasInsManager;
    insManagers[libruntime::ApiType::Function] = normalInsManager;
    this->UpdateConfig();
    if (Config::Instance().ENABLE_METRICS()) {
        this->metricsEnable_ = true;
    }
    faasInsManager->StartBatchRenewTimer();
    taskScheduler_ = std::make_shared<TaskScheduler>(std::bind(&TaskSubmitter::ScheduleFunctions, this));
    taskScheduler_->Run();
    if (downgrade_) {
        downgrade_->Init();
    }
    eraseTimer_ = YR::utility::ExecuteByGlobalTimer([this]() { this->EraseInsResourceMap(); },
                                                    ERASE_DELAY_TIME * MILLISECOND_UNIT, -1);
}

void TaskSubmitter::EraseInsResourceMap()
{
    for (auto it = insManagers.begin(); it != insManagers.end(); ++it) {
        it->second->EraseResourceInfoMap();
    }
}

void TaskSubmitter::UpdateConfig()
{
    int recycleTime = libRuntimeConfig->recycleTime;
    if (libRuntimeConfig->recycleTime <= 0) {
        YRLOG_WARN("recycle time is invalid, expect > 0,  actual {}", libRuntimeConfig->recycleTime);
        recycleTime = DEFAULT_RECYCLETIME;
    }
    this->recycleTimeMs = recycleTime * S_TO_MS;
    insManagers[libruntime::ApiType::Function]->UpdateConfig(this->recycleTimeMs);
}

TaskSubmitter::~TaskSubmitter()
{
    Finalize();
}

void TaskSubmitter::HandleInvokeNotify(const NotifyRequest &req, const ErrorInfo &err)
{
    if (!runFlag) {
        return;
    }

    ErrorInfo errInfo(static_cast<ErrorCode>(req.code()), req.message());
    if (errInfo.Finalized()) {
        return;
    }
    std::string requestId = req.requestid();
    auto [rawRequestId, seq] = YR::utility::IDGenerator::DecodeRawRequestId(requestId);
    std::shared_ptr<InvokeSpec> spec = requestManager->GetRequest(rawRequestId);
    if (spec == nullptr) {
        YRLOG_WARN(
            "request id : {} did not exit in request manager, may be the invoke request has been canceled or finished.",
            requestId);
        return;
    }
    if (spec->IsStaleDuplicateNotify(seq)) {
        return;
    }
    spec->endTimestamp =
        std::chrono::duration_cast<std::chrono::milliseconds>(std::chrono::system_clock::now().time_since_epoch())
            .count();
    insManagers[spec->functionMeta.apiType]->DecreaseUnfinishReqNum(spec, HandleFailInvokeIsDelayScaleDown(req, err));
    RequestResource resource = GetRequestResource(spec);
    if (req.code() != common::ERR_NONE) {
        HandleFailInvokeNotify(req, spec, resource, err);
    } else {
        HandleSuccessInvokeNotify(req, spec, resource);
    }
    {
        absl::ReaderMutexLock lock(&invokeCostMtx_);
        if (invokeCostMap.find(spec->invokeInstanceId) == invokeCostMap.end()) {
            return;
        }
    }
    absl::WriterMutexLock lock(&invokeCostMtx_);
    if (invokeCostMap.find(spec->invokeInstanceId) != invokeCostMap.end()) {
        invokeCostMap[spec->invokeInstanceId].StopTimer(requestId, req.code() == common::ERR_NONE);
    }
}

void TaskSubmitter::DowngradeCallback(const std::string &requestId, ErrorCode code, const std::string &result)
{
    auto spec = requestManager->GetRequest(requestId);
    if (spec == nullptr) {
        YRLOG_WARN(
            "request id : {} did not exit in request manager, may be the invoke request has been canceled or finished.",
            requestId);
        return;
    }
    ErrorInfo errInfo;
    if (code != ErrorCode::ERR_OK) {
        errInfo = ErrorInfo(code, result);
        memoryStore->SetError(spec->returnIds, errInfo);
    } else {
        auto &dataObj = spec->returnIds.front();
        std::shared_ptr<Buffer> buf = std::make_shared<NativeBuffer>(result.size() + MetaDataLen);
        dataObj.SetBuffer(buf);
        (void)memset_s(dataObj.meta->MutableData(), dataObj.meta->GetSize(), 0, dataObj.meta->GetSize());
        dataObj.data->MemoryCopy(result.data(), result.size());
        memoryStore->Put(dataObj.buffer, dataObj.id, {}, false);
        memoryStore->SetReady(dataObj.id);
    }
    auto ids = this->memoryStore->UnbindObjRefInReq(spec->requestId);
    auto errorInfo = this->memoryStore->DecreGlobalReference(ids);
    if (!errorInfo.OK()) {
        YRLOG_WARN("failed to decrease obj ref [{},...] by requestid {}. Code: {}, MCode: {}, Msg: {}", ids[0],
                   spec->requestId, fmt::underlying(errorInfo.Code()), fmt::underlying(errorInfo.MCode()),
                   errorInfo.Msg());
    }
    requestManager->RemoveRequest(spec->requestId);
    this->UpdateFaasInvokeLog(spec->requestId, errInfo);
}

bool TaskSubmitter::HandleFailInvokeIsDelayScaleDown(const NotifyRequest &req, const ErrorInfo &err)
{
    // If error code is less than 2000, it means user operation error, and the instance is normal,
    // so we can hold the instance, delay scale down.
    // Otherwise, invoke failure is either because of fault or only user code error,
    // it cannot be distinguished here, so just scale down the instance with no delay,
    // to avoid instance/node fault.
    YRLOG_INFO("check if invoke is abnormal notify request code {} requestid {}", fmt::underlying(req.code()),
               req.requestid());
    if (req.code() == common::ErrorCode::ERR_INSTANCE_NOT_FOUND ||
        req.code() == common::ErrorCode::ERR_INSTANCE_EXITED || req.code() == common::ErrorCode::ERR_INSTANCE_EVICTED ||
        req.code() == common::ErrorCode::ERR_USER_FUNCTION_EXCEPTION) {
        return false;
    }
    return true;
}

void TaskSubmitter::HandleFailInvokeNotify(const NotifyRequest &req, const std::shared_ptr<InvokeSpec> spec,
                                           const RequestResource &resource, const ErrorInfo &err)
{
    insManagers[spec->functionMeta.apiType]->ScaleDown(spec, HandleFailInvokeIsDelayScaleDown(req, err));
    auto isCreate = spec->invokeType == libruntime::InvokeType::CreateInstanceStateless ||
                    spec->invokeType == libruntime::InvokeType::CreateInstance;
    auto errInfo = PackageNotifyErr(req, isCreate);
    bool isConsumeRetryTime = false;
    if (NeedRetry(errInfo, spec, isConsumeRetryTime)) {
        spec->IncrementRequestID(spec->requestInvoke->Mutable());
        YRLOG_ERROR(
            "normal invoke requet fail, need retry, raw request id is {}, code is: {}, trace id is {}, seq is {}, "
            "complete request id is {}",
            req.requestid(), fmt::underlying(req.code()), spec->traceId, spec->seq,
            spec->requestInvoke->Mutable().requestid());
        if (isConsumeRetryTime) {
            spec->ConsumeRetryTime();
            YRLOG_DEBUG("consumed invoke retry time to {}, req id is {}", spec->opts.retryTimes, req.requestid());
        }
        auto taskScheduler = GetOrAddTaskScheduler(resource);
        {
            std::lock_guard<std::mutex> lockGuard(taskScheduler->queue->atomicMtx);
            taskScheduler->queue->Push(spec);
        }
    } else {
        YRLOG_ERROR(
            "normal invoke requet fail, don't need retry, raw request id is {}, code is: {}, trace id is {}, seq is "
            "{}, complete request id is {}",
            req.requestid(), fmt::underlying(req.code()), spec->traceId, spec->seq,
            spec->requestInvoke->Mutable().requestid());
        if (this->libRuntimeConfig->inCluster) {
            std::vector<std::string> dsObjs;
            for (auto it = spec->returnIds.begin(); it != spec->returnIds.end(); ++it) {
                if (!it->id.empty()) {
                    dsObjs.emplace_back(it->id);
                }
            }
        }
        auto ids = this->memoryStore->UnbindObjRefInReq(spec->requestId);
        auto errorInfo = this->memoryStore->DecreGlobalReference(ids);
        if (!errorInfo.OK()) {
            YRLOG_WARN("failed to decrease obj ref [{},...] by requestid {}. Code: {}, MCode: {}, Msg: {}", ids[0],
                       spec->requestId, fmt::underlying(errorInfo.Code()), fmt::underlying(errorInfo.MCode()),
                       errorInfo.Msg());
        }
        this->memoryStore->SetError(spec->returnIds, errInfo);
        (void)requestManager->RemoveRequest(spec->requestId);
        this->UpdateFaasInvokeLog(spec->requestId, errInfo);
    }
    NotifyScheduler(resource);
}

void TaskSubmitter::HandleSuccessInvokeNotify(const NotifyRequest &req, const std::shared_ptr<InvokeSpec> spec,
                                              const RequestResource &resource)
{
    YRLOG_DEBUG("handle nomarl invoke finish, request id: {}, trace id: {}", req.requestid(), spec->traceId);
    insManagers[spec->functionMeta.apiType]->ScaleDown(spec, true);
    (void)requestManager->RemoveRequest(spec->requestId);
    int smallObjSize = req.smallobjects_size();
    size_t curPos = 0;
    std::vector<std::string> dsObjs;

    if (smallObjSize <= 0) {
        for (size_t j = 0; j < spec->returnIds.size(); j++) {
            dsObjs.emplace_back(spec->returnIds[j].id);
        }
    } else {
        for (int i = 0; i < smallObjSize; i++) {
            auto &smallObj = req.smallobjects(i);
            const std::string &bufStr = smallObj.value();
            std::shared_ptr<Buffer> buf = std::make_shared<NativeBuffer>(bufStr.size());
            buf->MemoryCopy(bufStr.data(), bufStr.size());
            YRLOG_DEBUG("set small obj into memory store, obj id: {}, req id: {}, instace id: {}", smallObj.id(),
                        req.requestid(), spec->invokeInstanceId);
            memoryStore->Put(buf, smallObj.id(), {}, false);

            for (size_t j = curPos; j < spec->returnIds.size(); j++) {
                if (spec->returnIds[j].id == smallObj.id()) {
                    curPos++;
                    continue;
                }
                dsObjs.emplace_back(spec->returnIds[j].id);
                curPos++;
            }
        }
    }
    if (this->libRuntimeConfig->inCluster && !dsObjs.empty()) {
        this->memoryStore->IncreDSGlobalReference(dsObjs);
    }
    this->memoryStore->SetReady(spec->returnIds);
    auto ids = this->memoryStore->UnbindObjRefInReq(spec->requestId);
    auto errorInfo = this->memoryStore->DecreGlobalReference(ids);
    if (!errorInfo.OK()) {
        YRLOG_WARN("failed to decrease obj ref [{},...] by requestid {}. Code: {}, MCode: {}, Msg: {}", ids[0],
                   spec->requestId, fmt::underlying(errorInfo.Code()), fmt::underlying(errorInfo.MCode()),
                   errorInfo.Msg());
    }
    if (spec->functionMeta.apiType == libruntime::ApiType::Faas) {
        this->UpdateFaasInvokeLog(spec->requestId, ErrorInfo());
    }
    NotifyScheduler(resource);
    if (spec->invokeTimestamp > 0 && spec->createTimestamp > 0) {
        YRLOG_INFO("scheduler time: {} ms, invoke time: {} ms, req id: {}, trace id: {}",
                   spec->invokeTimestamp - spec->createTimestamp, spec->endTimestamp - spec->invokeTimestamp,
                   spec->requestId, spec->traceId);
    }
}

void TaskSubmitter::AddFaasCancelTimer(std::shared_ptr<InvokeSpec> spec)
{
    if (spec->functionMeta.apiType == libruntime::ApiType::Faas) {
        auto weakThis = weak_from_this();
        auto reqId = spec->requestId;
        auto timeoutSec =
            spec->opts.acquireTimeout > 0 ? spec->opts.acquireTimeout : Config::Instance().FASS_SCHEDULE_TIMEOUT();
        // The Purpose of this timer is to remove the corresponding spec from request manager and set the invoke req
        // as failed if it do not send invoke req after set timeout seconds(default 120s).
        {
            absl::WriterMutexLock lock(&cancelTimerMtx_);
            faasCancelTimerWorkers[reqId] = YR::utility::ExecuteByGlobalTimer(
                [weakThis, reqId, timeoutSec]() {
                    auto thisPtr = weakThis.lock();
                    if (!thisPtr) {
                        return;
                    }
                    thisPtr->CancelFaasScheduleTimeoutReq(reqId, timeoutSec);
                },
                (timeoutSec + DEFALUT_CANCEL_DELAY_TIME) * MILLISECOND_UNIT, 1);
        }
        /**
         * Record faas invoke function data in 3 steps:
           1. initializes the relevant data when submit the faas function
           2. Record the time of sending faas invoke request after the success of the acquire request, and record
         the end time of the faas request when the acquire fails
           3. After receiving notify (whether successful or failed), record the end time of the faas request
        */
        this->RecordFaasInvokeData(spec);
    }
}

std::shared_ptr<TaskSchedulerWrapper> TaskSubmitter::GetTaskScheduler(const RequestResource &resource)
{
    std::shared_ptr<TaskSchedulerWrapper> taskScheduler;
    {
        absl::ReaderMutexLock lock(&reqMtx_);
        if (taskSchedulerMap_.find(resource) != taskSchedulerMap_.end()) {
            taskScheduler = taskSchedulerMap_[resource];
        }
    }
    return taskScheduler;
}

std::shared_ptr<TaskSchedulerWrapper> TaskSubmitter::GetOrAddTaskScheduler(const RequestResource &resource)
{
    auto taskScheduler = GetTaskScheduler(resource);
    if (taskScheduler == nullptr) {
        absl::WriterMutexLock lock(&reqMtx_);
        if (taskSchedulerMap_.find(resource) == taskSchedulerMap_.end()) {
            taskScheduler = std::make_shared<TaskSchedulerWrapper>(taskScheduler_, enablePriority_);
            taskSchedulerMap_[resource] = taskScheduler;
        } else {
            taskScheduler = taskSchedulerMap_[resource];
        }
    }
    return taskScheduler;
}

void TaskSubmitter::SubmitFunction(std::shared_ptr<InvokeSpec> spec)
{
    YRLOG_DEBUG("start submit stateless function, req id is {}, return obj id is {}, trace id is {}", spec->requestId,
                spec->returnIds[0].id, spec->traceId);
    RequestResource resource = GetRequestResource(spec);
    auto taskScheduler = GetOrAddTaskScheduler(resource);
    {
        std::lock_guard<std::mutex> lockGuard(taskScheduler->queue->atomicMtx);
        taskScheduler->queue->Push(spec);
    }
    AddFaasCancelTimer(spec);
    taskScheduler->Notify();
}

void TaskSubmitter::RecordFaasInvokeData(const std::shared_ptr<InvokeSpec> spec)
{
    YRLOG_DEBUG("start recorde value, function id is {}, req id is {}", spec->functionMeta.functionId, spec->requestId);
    absl::WriterMutexLock lock(&invokeDataMtx_);
    if (!metricsEnable_) {
        return;
    }
    if (faasInvokeDataMap_.find(spec->requestId) == faasInvokeDataMap_.end()) {
        FaasInvokeData data(this->libRuntimeConfig->tenantId, spec->functionMeta.funcName,
                            ar_ ? this->ar_->ParseAlias(spec->functionMeta.functionId, spec->opts.aliasParams) : "",
                            spec->traceId, GetCurrentTimestampNs());
        auto pos = data.aliAs.find_last_of('/');
        if (pos != std::string::npos && pos != data.aliAs.length() - 1) {
            data.version = data.aliAs.substr(pos + 1);
        }
        faasInvokeDataMap_[spec->requestId] = std::make_shared<FaasInvokeData>(data);
    }
}

void TaskSubmitter::CancelFaasScheduleTimeoutReq(const std::string &reqId, int timeoutSeconds)
{
    auto spec = requestManager->GetRequest(reqId);
    if (!spec) {
        YRLOG_DEBUG("spec of request {} is nullptr, return directly", reqId);
        return;
    }
    if (spec->invokeInstanceId.empty()) {
        (void)requestManager->RemoveRequest(reqId);
        auto ids = memoryStore->UnbindObjRefInReq(spec->requestId);
        auto errorInfo = memoryStore->DecreGlobalReference(ids);
        if (!errorInfo.OK()) {
            YRLOG_WARN("failed to decrease obj ref [{},...] by requestid {}. Code: {}, MCode: {}, Msg: {}", ids[0],
                       spec->requestId, fmt::underlying(errorInfo.Code()), fmt::underlying(errorInfo.MCode()),
                       errorInfo.Msg());
        }
        auto resource = GetRequestResource(spec);
        ErrorInfo err = ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME,
                                  "request is still scheduled within " + std::to_string(timeoutSeconds) +
                                      " seconds, cancel the faas invoke request");
        {
            absl::ReaderMutexLock lock(&reqMtx_);
            auto it = taskSchedulerMap_.find(resource);
            if (it != taskSchedulerMap_.end()) {
                auto scheduler = it->second;
                if (!scheduler->IsLastErrorOk()) {
                    err = scheduler->GetLastError();
                }
            }
        }
        memoryStore->SetError(spec->returnIds, err);
        YRLOG_ERROR("cancel req {} with {}s no scheduled, code: {}, message: {}, trace: {}", spec->requestId,
                    timeoutSeconds, fmt::underlying(err.Code()), err.Msg(), spec->traceId);
        this->UpdateFaasInvokeLog(reqId, err);
    }
}

void TaskSubmitter::EraseFaasCancelTimer(const std::string &reqId)
{
    // three situations need to erase faas cancel timer
    // 1. create ins failed and no need retry, there are no running or creating instance in current resource, then the
    // faas req is considered a failure and then need to erase the cancel timer. Refer to TaskSubmitter::ScheduleIns
    // method else branch for details.
    // 2. RequestManager does not have a spec for reqId, indicating that timer has already been executed or the request
    // has been cancelled for other reasons, so need to erase the cancel timer. Refer to TaskSubmitter::ScheduleRequest
    // method for details.
    // 3. Before send faas invoke req, need to stop and erase the cancel timer. Refer to TaskSubmitter::ScheduleRequest
    // method for details.
    absl::WriterMutexLock lock(&cancelTimerMtx_);
    if (auto it = faasCancelTimerWorkers.find(reqId); it != faasCancelTimerWorkers.end()) {
        if (it->second) {
            YRLOG_DEBUG("start stop and erase the cancel timer of faas req: {}", reqId);
            it->second->cancel();
            it->second.reset();
            faasCancelTimerWorkers.erase(it);
        }
    }
}

void TaskSubmitter::NotifyScheduler(const RequestResource &resource, const ErrorInfo &err)
{
    auto taskScheduler = GetTaskScheduler(resource);
    if (taskScheduler) {
        if (!err.OK()) {
            taskScheduler->SetLastError(err);
        }
        taskScheduler->Notify();
        return;
    }
    auto resources = GetScheduleResources();
    for (const auto &resource : resources) {
        taskScheduler = GetTaskScheduler(resource);
        if (taskScheduler) {
            taskScheduler->Notify();
            return;
        }
    }
}

void TaskSubmitter::ScheduleIns(const RequestResource &resource, const ErrorInfo &errInfo, bool isRemainIns)
{
    if (errInfo.OK()) {
        NotifyScheduler(resource);
        return;
    }
    if (NeedRetryCreate(errInfo)) {
        YRLOG_INFO("start retry create task instance, code: {}, msg: {}", fmt::underlying(errInfo.Code()),
                   errInfo.Msg());
        NotifyScheduler(resource, errInfo);
        return;
    }
    auto taskScheduler = GetTaskScheduler(resource);
    if (taskScheduler == nullptr) {
        return;
    }
    // If there are still other instances existing or being created under resource, equals the
    // isRemainIns value is true, then all requests under that resource should not be set as failed
    if (!isRemainIns) {
        std::lock_guard<std::mutex> lockGuard(taskScheduler->queue->atomicMtx);
        for (; !taskScheduler->queue->Empty(); taskScheduler->queue->Pop()) {
            auto reqId = taskScheduler->queue->Top()->requestId;
            this->EraseFaasCancelTimer(reqId);
            if (downgrade_->ShouldFaultDowngrade()) {
                auto spec = requestManager->GetRequest(reqId);
                downgrade_->Downgrade(spec, std::bind(&TaskSubmitter::DowngradeCallback, this, std::placeholders::_1,
                                                      std::placeholders::_2, std::placeholders::_3));
                continue;
            }
            std::shared_ptr<InvokeSpec> invokeSpecNeedFailed;
            bool specExits = requestManager->PopRequest(reqId, invokeSpecNeedFailed);
            if (!specExits) {
                continue;
            }
            if (invokeSpecNeedFailed) {
                YRLOG_ERROR(
                    "there is no available ins of resource, start set error, req id is {}, trace id is {}, invoke ins "
                    "id is {}",
                    invokeSpecNeedFailed->requestId, invokeSpecNeedFailed->traceId,
                    invokeSpecNeedFailed->invokeInstanceId);
            }
            this->memoryStore->SetError(invokeSpecNeedFailed->returnIds[0].id, errInfo);
            this->UpdateFaasInvokeLog(reqId, errInfo);
        }
    }
    if (taskScheduler) {
        taskScheduler->Notify();
    }
}

void TaskSubmitter::SendInvokeReq(const RequestResource &resource, std::shared_ptr<InvokeSpec> invokeSpec)
{
    YRLOG_INFO("invoke function {}, instance: {}, lease: {}, req: {}, trace: {}", invokeSpec->functionMeta.funcName,
               invokeSpec->invokeInstanceId, invokeSpec->invokeLeaseId, invokeSpec->requestId, invokeSpec->traceId);
    invokeSpec->invokeTimestamp =
        std::chrono::duration_cast<std::chrono::milliseconds>(std::chrono::system_clock::now().time_since_epoch())
            .count();

    if (!invokeSpec->opts.device.name.empty()) {
        absl::WriterMutexLock lock(&invokeCostMtx_);
        invokeCostMap[invokeSpec->invokeInstanceId].StartTimer(invokeSpec->requestId);
        YRLOG_DEBUG("start timer for instance: {}, reqID: {}", invokeSpec->invokeInstanceId, invokeSpec->requestId);
    }

    if (invokeSpec->functionMeta.IsServiceApiType()) {
        this->UpdateFaasInvokeSendTime(invokeSpec->requestId);
    }
    auto weakThis = weak_from_this();
    auto insId = invokeSpec->invokeInstanceId;
    this->fsClient->InvokeAsync(
        invokeSpec->requestInvoke,
        [weakThis, insId](const NotifyRequest &notifyRequest, const ErrorInfo &err) {
            if (auto thisPtr = weakThis.lock(); thisPtr) {
                thisPtr->HandleInvokeNotify(notifyRequest, err);
                if (err.IsTimeout() && thisPtr->cancelCb) {
                    // if timeout,then send cancel req to runtime for erase pending thread
                    YRLOG_DEBUG("start send cancel req to runtime: {} for req: {}", insId, notifyRequest.requestid());
                    thisPtr->cancelCb(insId, notifyRequest.requestid(), libruntime::Signal::ErasePendingThread);
                }
            }
        },
        invokeSpec->opts.timeout);
}

bool TaskSubmitter::ScheduleRequest(const RequestResource &resource, std::shared_ptr<BaseQueue> requestQueue)
{
    std::unique_lock<std::mutex> atomicLock(requestQueue->atomicMtx);
    if (requestQueue->Empty()) {
        return true;
    }
    auto requestId = requestQueue->Top()->requestId;
    auto requestQueueSize = requestQueue->Size();
    YRLOG_DEBUG("queue size: {}, top request: {}", requestQueueSize, requestId);
    auto invokeSpec = requestManager->GetRequest(requestId);
    if (invokeSpec == nullptr) {
        YRLOG_WARN("The request {} has been cancelled", requestId);
        this->EraseFaasCancelTimer(requestId);
        requestQueue->Pop();
        return false;
    }
    if (downgrade_->ShouldDowngrade(invokeSpec)) {
        this->EraseFaasCancelTimer(requestId);
        requestQueue->Pop();
        atomicLock.unlock();
        auto weakPtr = weak_from_this();
        downgrade_->Downgrade(
            invokeSpec, [weakPtr](const std::string &requestId, ErrorCode code, const std::string &result) {
                if (auto thisPtr = weakPtr.lock(); thisPtr) {
                    thisPtr->DowngradeCallback(requestId, code, result);
                }
            });
        return false;
    }
    auto summary = insManagers[invokeSpec->functionMeta.apiType]->GetAvailableIns(resource);
    if (summary.instanceId.empty()) {
        atomicLock.unlock();
        YRLOG_DEBUG("invoke request {} can not be scheduled", requestId);
        bool needCreate = insManagers[invokeSpec->functionMeta.apiType]->ScaleUp(invokeSpec, requestQueueSize);
        return !needCreate;
    }
    this->EraseFaasCancelTimer(requestId);
    requestQueue->Pop();
    atomicLock.unlock();
    invokeSpec->invokeInstanceId = summary.instanceId;
    invokeSpec->invokeLeaseId = summary.leaseId;
    invokeSpec->requestInvoke->Mutable().set_instanceid(summary.instanceId);
    if (summary.forceInvoke) {
        auto invokeOptions = invokeSpec->requestInvoke->Mutable().mutable_invokeoptions();
        auto customTag = invokeOptions->mutable_customtag();
        customTag->insert({"ENABLE_FORCE_INVOKE", ""});
    }
    invokeSpec->instanceRoute = memoryStore->GetInstanceRoute(invokeSpec->returnIds[0].id);
    // If event stream is enabled, pass event server info in invoke options.
    if (libRuntimeConfig->enableEvent) {
        auto it = resource.opts.invokeLabels.find(INSTANCE_REQUIREMENT_ACCEPT);
        if (it != resource.opts.invokeLabels.end() && it->second == INSTANCE_REQUIREMENT_ACCEPT_EVENT_STREAM) {
            auto invokeOptions = invokeSpec->requestInvoke->Mutable().mutable_invokeoptions();
            auto customTag = invokeOptions->mutable_customtag();
            (*customTag)[YR_EVENT_SERVER_IP] = fsClient->GetEventServerIP();
            (*customTag)[YR_EVENT_SERVER_PORT] = std::to_string(fsClient->GetEventServerPort());
        }
    }
    SendInvokeReq(resource, invokeSpec);
    return false;
}

bool TaskSubmitter::CancelAndScheOtherRes(const RequestResource &resource)
{
    absl::ReaderMutexLock lock(&reqMtx_);
    auto it = taskSchedulerMap_.find(resource);
    if (it == taskSchedulerMap_.end()) {
        return true;
    }
    if (!it->second->queue->Empty()) {
        return false;
    }
    YRLOG_DEBUG("queue is empty, try schedule other queue. func name {}, class name {}", resource.functionMeta.funcName,
                resource.functionMeta.className);
    insManagers[resource.functionMeta.apiType]->ScaleCancel(resource, 0, true);
    for (auto &entry : taskSchedulerMap_) {
        if (taskSchedulerMap_.find(entry.first) == it) {
            continue;
        }
        if (!entry.second->queue->Empty()) {
            taskSchedulerMap_[entry.first]->Notify();
            return true;
        }
    }
    return true;
}

void TaskSubmitter::ScheduleFunctions()
{
    auto resources = GetScheduleResources();
    for (const auto &resource : resources) {
        ScheduleFunction(resource);
    }
}

std::vector<RequestResource> TaskSubmitter::GetScheduleResources()
{
    absl::ReaderMutexLock lock(&reqMtx_);
    std::vector<RequestResource> resources;
    for (const auto &pair : taskSchedulerMap_) {
        resources.push_back(pair.first);
    }
    return resources;
}

void TaskSubmitter::ScheduleFunction(const RequestResource &resource)
{
    YRLOG_DEBUG("schedule resource req. func name is {}, class name is {}", resource.functionMeta.funcName,
                resource.functionMeta.className);
    if (!runFlag) {
        return;
    }
    if (CancelAndScheOtherRes(resource)) {
        EraseTaskSchedulerMap(resource);
        return;
    }
    auto taskScheduler = GetTaskScheduler(resource);
    if (taskScheduler == nullptr) {
        return;
    }
    while (!taskScheduler->queue->Empty()) {
        if (auto finish = ScheduleRequest(resource, taskScheduler->queue); finish) {
            break;
        }
    }
    EraseTaskSchedulerMap(resource);
}

void TaskSubmitter::EraseTaskSchedulerMap(const RequestResource &resource)
{
    // yield make SubmitFunction can be executed first avoid frequent cleaning taskSchedulerMap
    std::this_thread::yield();
    auto taskScheduler = GetTaskScheduler(resource);
    if (taskScheduler == nullptr) {
        return;
    }
    const static int currentCount = 2;
    std::unique_lock<std::mutex> atomicLock(taskScheduler->queue->atomicMtx);
    if (taskScheduler->queue->Empty()) {
        absl::WriterMutexLock lock(&reqMtx_);
        // if there are still references elsewhere, do not erase.
        if (taskScheduler.use_count() <= currentCount) {
            YRLOG_DEBUG("remove taskScheduler");
            taskSchedulerMap_.erase(resource);
        }
    }
}

bool TaskSubmitter::NeedRetryCreate(const ErrorInfo &errInfo)
{
    auto errCode = errInfo.Code();
    if (errCode == ErrorCode::ERR_RESOURCE_NOT_ENOUGH || errCode == ErrorCode::ERR_INNER_COMMUNICATION ||
        errCode == ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_BUS) {
        return true;
    }
    return false;
}

bool TaskSubmitter::NeedRetry(const ErrorInfo &errInfo, const std::shared_ptr<InvokeSpec> spec,
                              bool &isConsumeRetryTime)
{
    auto errCode = errInfo.Code();
    if (spec->invokeType == libruntime::InvokeType::InvokeFunctionStateless &&
        (errCode == ErrorCode::ERR_INSTANCE_EVICTED || errCode == ErrorCode::ERR_INSTANCE_NOT_FOUND ||
         errCode == ErrorCode::ERR_INSTANCE_EXITED)) {
        if (spec->ExceedMaxRetryTime()) {
            return false;
        }
        isConsumeRetryTime = false;  // the only case to retry without consuming
        return true;
    }

    if (spec->opts.retryTimes <= 0) {
        isConsumeRetryTime = false;
        return false;
    }
    switch (spec->invokeType) {
        case libruntime::InvokeType::InvokeFunctionStateless: {
            static const std::unordered_set<ErrorCode> codesWorthRetry = {
                ErrorCode::ERR_USER_FUNCTION_EXCEPTION, ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_BUS,
                ErrorCode::ERR_INNER_COMMUNICATION,     ErrorCode::ERR_SHARED_MEMORY_LIMITED,
                ErrorCode::ERR_OPERATE_DISK_FAILED,     ErrorCode::ERR_INSUFFICIENT_DISK_SPACE,
                ErrorCode::ERR_INSTANCE_NOT_FOUND,      ErrorCode::ERR_INSTANCE_EXITED,
                ErrorCode::ERR_INSTANCE_SUB_HEALTH,     ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_FRONTEND};
            bool isRetryable = (codesWorthRetry.find(errCode) != codesWorthRetry.end());
            if (isRetryable && spec->opts.retryChecker) {
                isConsumeRetryTime = spec->opts.retryChecker(errInfo);
                return isConsumeRetryTime;
            }
            // no retry checker, then retry for all codes worth retrying
            isConsumeRetryTime = isRetryable;
            return isConsumeRetryTime;
        }
        default:
            break;
    }
    return (isConsumeRetryTime = false);
}

ErrorInfo TaskSubmitter::CancelStatelessRequest(std::shared_ptr<InvokeSpec> spec, const KillFunc &killCallBack,
                                                bool isForce, bool isRecursive, const std::string &objId)
{
    YRLOG_DEBUG("start cancel request: {}, object id is {}", spec->requestId, objId);
    if (spec->invokeInstanceId != "" && isForce) {
        RequestResource resource = GetRequestResource(spec);
        if (resource.concurrency > MIN_CONCURRENCY) {
            YRLOG_WARN(
                "request: {}  has been sent to the runtime, and concurrency is greater than 1.Cancellation is not "
                "supported, return directly",
                spec->requestId);
            return ErrorInfo();
        }
        // 在调用同步的killCallBack之前调用RemoveRequest，避免滞后
        requestManager->RemoveRequest(spec->requestId);
        YRLOG_INFO("request: {} has already be sent, need to send signal to remote instance: {}", spec->requestId,
                   spec->invokeInstanceId);
        if (isRecursive) {
            (void)killCallBack(spec->invokeInstanceId, "", libruntime::Signal::Cancel);
        } else {
            (void)killCallBack(spec->invokeInstanceId, "", libruntime::Signal::KillInstance);
        }
    } else {
        requestManager->RemoveRequest(spec->requestId);
    }
    ErrorInfo cancelErr(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                        "invalid get obj, the obj has been cancelled.");
    memoryStore->SetError(objId, cancelErr);
    this->UpdateFaasInvokeLog(spec->requestId, cancelErr);
    return ErrorInfo();
}

std::vector<std::string> TaskSubmitter::GetInstanceIds()
{
    if (insManagers[libruntime::ApiType::Function]) {
        return insManagers[libruntime::ApiType::Function]->GetInstanceIds();
    }
    return std::vector<std::string>();
}

std::vector<std::string> TaskSubmitter::GetCreatingInsIds()
{
    if (insManagers[libruntime::ApiType::Function]) {
        return insManagers[libruntime::ApiType::Function]->GetCreatingInsIds();
    }
    return std::vector<std::string>();
}

void TaskSubmitter::Finalize()
{
    if (!runFlag) {
        return;
    }
    runFlag = false;
    {
        absl::WriterMutexLock lock(&reqMtx_);
        taskSchedulerMap_.clear();
    }
    if (taskScheduler_) {
        taskScheduler_->Stop();
    }

    for (auto &pair : insManagers) {
        pair.second->Stop();
    }
    {
        absl::WriterMutexLock lock(&cancelTimerMtx_);
        for (auto it = faasCancelTimerWorkers.begin(); it != faasCancelTimerWorkers.end(); ++it) {
            if (it->second) {
                it->second->cancel();
                it->second.reset();
            }
        }
        faasCancelTimerWorkers.clear();
    }
    if (eraseTimer_) {
        eraseTimer_->cancel();
        eraseTimer_.reset();
    }
}

std::pair<InstanceAllocation, ErrorInfo> TaskSubmitter::AcquireInstance(const std::string &stateId,
                                                                        std::shared_ptr<InvokeSpec> spec)
{
    return std::static_pointer_cast<FaasInsManager>(insManagers[spec->functionMeta.apiType])
        ->AcquireInstance(stateId, spec);
}

ErrorInfo TaskSubmitter::ReleaseInstance(const std::string &leaseId, const std::string &stateId, bool abnormal,
                                         std::shared_ptr<InvokeSpec> spec)
{
    return std::static_pointer_cast<FaasInsManager>(insManagers[spec->functionMeta.apiType])
        ->ReleaseInstance(leaseId, stateId, abnormal, spec);
}

void TaskSubmitter::UpdateFaaSSchedulerInfo(std::string schedulerFuncKey,
                                            const std::vector<SchedulerInstance> &schedulerInfoList)
{
    if (insManagers[libruntime::ApiType::Faas]) {
        insManagers[libruntime::ApiType::Faas]->UpdateSchedulerInfo(schedulerFuncKey, schedulerInfoList);
    }
    if (insManagers[libruntime::ApiType::Serve]) {
        insManagers[libruntime::ApiType::Serve]->UpdateSchedulerInfo(schedulerFuncKey, schedulerInfoList);
    }
}

void TaskSubmitter::DeleteInsCallback(const std::string &instanceId)
{
    if (fsClient) {
        fsClient->RemoveInsRtIntf(instanceId);
    }
}

void TaskSubmitter::UpdateSchdulerInfo(const std::string &schedulerName, const std::string &schedulerId,
                                       const std::string &option)
{
    if (insManagers[libruntime::ApiType::Faas]) {
        YRLOG_INFO("start update scheduler info, scheduler name is {}, schduler id is {},  option is {}", schedulerName,
                   schedulerId, option);
        insManagers[libruntime::ApiType::Faas]->UpdateSchdulerInfo(schedulerName, schedulerId, option);
    }
}

YR::Libruntime::GaugeData TaskSubmitter::ConvertToGaugeData(const std::shared_ptr<FaasInvokeData> data,
                                                            const std::string &reqId)
{
    YR::Libruntime::GaugeData gaugeData;
    gaugeData.name = "report_faas_invoke_data";
    gaugeData.labels["requestId"] = reqId;
    gaugeData.labels["businessId"] = data->businessId;
    gaugeData.labels["tenantId"] = data->tenantId;
    gaugeData.labels["srcAppId"] = data->srcAppId;
    gaugeData.labels["functionName"] = data->functionName;
    gaugeData.labels["traceId"] = data->traceId;
    gaugeData.labels["version"] = data->version;
    gaugeData.labels["aliAs"] = data->aliAs;
    gaugeData.labels["code"] = data->code;
    gaugeData.labels["innerCode"] = data->innerCode;
    gaugeData.labels["describeMsg"] = data->describeMsg;
    gaugeData.labels["exec"] =
        data->endTime - data->sendTime > 0 ? std::to_string(data->endTime - data->sendTime) : "0";
    gaugeData.labels["cost"] =
        data->endTime - data->submitTime > 0 ? std::to_string(data->endTime - data->submitTime) : "0";
    return gaugeData;
}

void TaskSubmitter::UpdateFaasInvokeSendTime(const std::string &reqId)
{
    absl::WriterMutexLock lock(&invokeDataMtx_);
    if (!metricsEnable_) {
        return;
    }
    if (faasInvokeDataMap_.find(reqId) != faasInvokeDataMap_.end()) {
        faasInvokeDataMap_[reqId]->sendTime = GetCurrentTimestampNs();
    }
}

void TaskSubmitter::UpdateFaasInvokeLog(const std::string &reqId, const ErrorInfo &err)
{
    {
        absl::WriterMutexLock lock(&invokeDataMtx_);
        if (!metricsEnable_) {
            return;
        }
    }
    std::shared_ptr<FaasInvokeData> data;
    {
        absl::WriterMutexLock lock(&invokeDataMtx_);
        auto it = faasInvokeDataMap_.find(reqId);
        if (it == faasInvokeDataMap_.end()) {
            YRLOG_DEBUG("there is no invoke data of req: {}, no need update", reqId);
            return;
        }
        if (!this->metricsAdaptor_ || !Config::Instance().ENABLE_METRICS() || this->metricsAdaptor_->IsInited()) {
            return;
        }
        it->second->endTime = GetCurrentTimestampNs();
        if (err.OK()) {
            it->second->code = "200";
        } else {
            if (it->second->sendTime > 0) {
                it->second->code = "500";
            } else {
                it->second->code = "400";
            }
        }
        it->second->innerCode = std::to_string(err.Code());
        auto errIt = errCodeToString.find(err.Code());
        it->second->describeMsg = errIt == errCodeToString.end() ? "UNKOWN" : errIt->second;
        data = it->second;
        faasInvokeDataMap_.erase(it);
    }
    YRLOG_DEBUG("start report faas invoke data, req id is {}, function id is {}", reqId, data->aliAs);
    auto reportErr = this->metricsAdaptor_->ReportMetrics(ConvertToGaugeData(data, reqId));
    if (!reportErr.OK()) {
        YRLOG_WARN("failed to report metrics, req id is {}, trace id is {}, err code is {}, msg is {}", reqId,
                   data->traceId, fmt::underlying(reportErr.Code()), reportErr.Msg());
    }
}
}  // namespace Libruntime
}  // namespace YR
