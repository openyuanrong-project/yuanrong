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

#include "faas_instance_manager.h"
#include <json.hpp>
#include "src/dto/acquire_options.h"

namespace YR {
namespace Libruntime {
const std::string INSTANCE_REQUIREMENT_RESOURKEY = "resourcesData";
const std::string INSTANCE_REQUIREMENT_INSKEY = "designateInstanceID";
const std::string INSTANCE_REQUIREMENT_POOLLABELKEY = "poolLabel";
const std::string INSTANCE_REQUIREMENT_INVOKE_LABEL = "instanceInvokeLabel";
const std::string META_PREFIX = "0000000000000000";
const std::string INSTANCETRAFFICLIMITED = "instanceTrafficLimited";
const std::string HOST_NAME_ENV = "HOSTNAME";
const std::string POD_NAME_ENV = "POD_NAME";
const std::string INSTANCE_CALLER_POD_NAME = "instanceCallerPodName";
const std::string INSTANCE_SESSION_CONFIG = "instanceSessionConfig";
const int64_t BEFOR_RETAIN_TIME = 30;  // millisecond
const int64_t RETAIN_TIME_RATE = 2;
const int RELEASE_DELAYTIME = 100;  // millisecond
const int FAAS_INS_REQ_SUCCESS_CODE = 6030;
const int FAAS_USER_CODE_THRESHOLD = 10000;
const int ERR_FUNC_META_NOT_FOUND = 150424;
const int ERR_TARGET_INSTANCE_NOT_FOUND = 150425;
const int ERR_RENEW_INSTANCE_LEASE_NOT_FOUND = 150463;
const int ERR_NO_INSTANCE_AVAILABLE = 150431;
const int INSTANCE_LABEL_NOT_FOUND = 150444;
const int ERR_USER_FUNC_ENTRY_NOT_FOUND = 4001;
const int WISECLOUD_NUWA_INVOKE_ERROR = 161915;
const int64_t BATCH_RENEW_LEASE_NUM = 1000;
const std::string SPEC_ERR_MSG = "invoke request timeout";

const static std::set<int> userErrSet =
    std::set<int>{ERR_FUNC_META_NOT_FOUND, INSTANCE_LABEL_NOT_FOUND, ERR_USER_FUNC_ENTRY_NOT_FOUND,
                  WISECLOUD_NUWA_INVOKE_ERROR, ERR_TARGET_INSTANCE_NOT_FOUND};

void FaasInsManager::UpdateConfig(int recycleTimeMs) {}

void FaasInsManager::RecordRequest(const RequestResource &resource, const std::shared_ptr<InvokeSpec> spec,
                                   bool isInstanceNormal)
{
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return;
    }
    std::shared_ptr<InstanceInfo> faasInfo;
    {
        absl::ReaderMutexLock lock(&info->mtx);
        if (info->instanceInfos.find(spec->invokeLeaseId) == info->instanceInfos.end()) {
            return;
        }
        faasInfo = info->instanceInfos[spec->invokeLeaseId];
    }
    absl::WriterMutexLock insLk(&faasInfo->mtx);
    // Except for user option errors, ins will only be immediately deleted in case of instance abnormal
    if (!isInstanceNormal) {
        faasInfo->reporter->RecordAbnormal();
        faasInfo->claimTime = 0;
        return;
    }
    long long claimTime = 0;
    if (faasInfo->claimTime != 0) {
        claimTime = faasInfo->claimTime;
        auto currentMillTime = std::chrono::time_point_cast<std::chrono::milliseconds>(std::chrono::steady_clock::now())
                                   .time_since_epoch()
                                   .count();
        faasInfo->claimTime = currentMillTime;
        faasInfo->reporter->RecordRequest(currentMillTime - claimTime);
    }
}

void FaasInsManager::DelRelatedLease(const std::string &insId, const RequestResource &resource)
{
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return;
    }
    absl::WriterMutexLock lock(&info->mtx);
    auto &insInfos = info->instanceInfos;
    for (auto it = insInfos.begin(); it != insInfos.end();) {
        std::string leaseId;
        std::string instanceId;
        {
            absl::ReaderMutexLock lock(&it->second->mtx);
            leaseId = it->second->leaseId;
            instanceId = it->second->instanceId;
        }
        YRLOG_DEBUG("lease id is {}, instance id is {}, input ins id is {}", leaseId, instanceId, insId);
        if (instanceId != insId) {
            ++it;
            continue;
        }
        ReleaseInstanceAsync(it->second);
        it = insInfos.erase(it);
        info->avaliableInstanceInfos.erase(leaseId);
        DecreaseCreatedInstanceNum();
    }
}

void FaasInsManager::ScaleCancel(const RequestResource &resource, size_t reqNum, bool cleanAll)
{
    return;
}

void FaasInsManager::ScaleDown(const std::shared_ptr<InvokeSpec> spec, bool isInstanceNormal)
{
    auto resource = GetRequestResource(spec);
    this->RecordRequest(resource, spec, isInstanceNormal);
    auto id = spec->invokeLeaseId;
    auto instanceId = spec->invokeInstanceId;
    if (isInstanceNormal) {
        return StartReleaseTimer(resource, id);
    }
    this->DelRelatedLease(instanceId, resource);
    EraseResourceInfoMap(resource);
}

bool FaasInsManager::ScaleUp(std::shared_ptr<InvokeSpec> invokeSpec, size_t reqNum)
{
    AddRequestResourceInfo(invokeSpec);
    return this->AcquireFaasIns(invokeSpec, reqNum);
}

bool FaasInsManager::AcquireFaasIns(const std::shared_ptr<InvokeSpec> invokeSpec, size_t reqNum)
{
    auto resource = GetRequestResource(invokeSpec);
    auto needCreatePair = NeedCreateNewIns(resource, reqNum, false);
    if (!needCreatePair.first) {
        YRLOG_INFO("No need create new ins for reqid: {}, trace id: {}", invokeSpec->requestId, invokeSpec->traceId);
        return false;
    }
    SendAcquireReq(invokeSpec, needCreatePair.second);
    return true;
}

void FaasInsManager::SendAcquireReq(const std::shared_ptr<InvokeSpec> invokeSpec, size_t delayTime)
{
    auto resource = GetRequestResource(invokeSpec);
    this->AddCreatingInsInfo(resource, std::make_shared<CreatingInsInfo>("", 0));
    auto weak_this = weak_from_this();
    tw_->CreateTimer(delayTime * MILLISECOND_UNIT, 1, [weak_this, invokeSpec]() {
        if (auto this_ptr = weak_this.lock(); this_ptr) {
            this_ptr->AcquireInstanceAsync(invokeSpec);
        }
    });
}

void FaasInsManager::ProcessInstanceInfo(std::shared_ptr<InvokeSpec> spec, const InstanceAllocation &inst)
{
    auto resource = GetRequestResource(spec);
    auto faasInsInfo = std::make_shared<InstanceInfo>();
    faasInsInfo->instanceId = inst.instanceId;
    faasInsInfo->leaseId = inst.leaseId;
    faasInsInfo->idleTime = 0;
    faasInsInfo->unfinishReqNum = 0;
    faasInsInfo->available = true;
    faasInsInfo->traceId = spec->traceId;
    faasInsInfo->faasInfo = FaasAllocationInfo{inst.functionId, inst.funcSig, inst.tLeaseInterval,
                                               spec->GetSchedulerInstanceId(), spec->opts.schedulerFunctionId};
    faasInsInfo->reporter = std::make_shared<ReportRecord>();
    this->HandleFaasInsInfo(faasInsInfo, resource);
}

std::pair<InstanceResponse, ErrorInfo> GetFaasInstanceRsp(const NotifyRequest &notifyReq)
{
    InstanceResponse instanceResp;
    if (notifyReq.smallobjects_size() == 0) {
        return std::make_pair(instanceResp,
                              ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                        "invalid acquire notify result, req id is " + notifyReq.requestid()));
    }
    auto &smallObj = notifyReq.smallobjects(0);
    if (smallObj.value().size() <= MetaDataLen) {
        return std::make_pair(instanceResp,
                              ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                        "invalid acquire notify result, req id is " + notifyReq.requestid()));
    }
    const std::string &bufStr = smallObj.value().substr(MetaDataLen);
    YRLOG_DEBUG("get acquire notify result, req id is {}, value is {}", notifyReq.requestid(), bufStr);
    auto err = ConvertStringToInsResp(instanceResp, bufStr);
    if (!err.OK()) {
        YRLOG_ERROR("failed to convert acquire notify req to instance info, req id is {}, err msg is {}",
                    notifyReq.requestid(), err.Msg());
    }
    return std::make_pair(instanceResp, err);
}

std::pair<BatchInstanceResponse, ErrorInfo> GetFaasBatchInstanceRsp(const NotifyRequest &notifyReq)
{
    BatchInstanceResponse instanceResp;
    if (notifyReq.smallobjects_size() == 0) {
        return std::make_pair(instanceResp,
                              ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                        "invalid acquire notify result, req id is " + notifyReq.requestid()));
    }
    auto &smallObj = notifyReq.smallobjects(0);
    if (smallObj.value().size() <= MetaDataLen) {
        return std::make_pair(instanceResp,
                              ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                        "invalid acquire notify result, req id is " + notifyReq.requestid()));
    }
    const std::string &bufStr = smallObj.value().substr(MetaDataLen);
    YRLOG_DEBUG("get retain notify result, req id is {}, value is {}", notifyReq.requestid(), bufStr);
    ConvertStringToBatchInsResp(instanceResp, bufStr);
    return std::make_pair(instanceResp, ErrorInfo());
}

void FaasInsManager::ProcessAsynAcquireResult(const NotifyRequest &notifyReq, std::shared_ptr<InvokeSpec> acquireSpec,
                                              const ErrorInfo &errInput, std::shared_ptr<InvokeSpec> invokeSpec)
{
    auto err = ErrorInfo(ErrorCode(notifyReq.code()), notifyReq.message(), errInput.IsAckTimeout());
    if (errInput.IsTimeout()) {
        err.SetErrorCode(ErrorCode::ERR_ACQUIRE_TIMEOUT);
        err.SetErrorMsg("acquire instance timeout");
    }
    if (!err.OK()) {
        AcquireCallback(acquireSpec, err, InstanceResponse{}, invokeSpec);
        return;
    }
    auto [instanceResp, errInfo] = GetFaasInstanceRsp(notifyReq);
    AcquireCallback(acquireSpec, errInfo, instanceResp, invokeSpec);
}

void FaasInsManager::ProcecssAcquireResult(
    const NotifyRequest &notifyReq, std::shared_ptr<InvokeSpec> spec,
    std::shared_ptr<std::promise<std::pair<InstanceAllocation, ErrorInfo>>> acquirePromise)
{
    auto err = ErrorInfo(ErrorCode(notifyReq.code()), notifyReq.message());
    if (notifyReq.code() != common::ERR_NONE) {
        acquirePromise->set_value(std::make_pair(InstanceAllocation{}, err));
        return;
    }

    auto [instanceResp, errInfo] = GetFaasInstanceRsp(notifyReq);
    if (!errInfo.OK()) {
        acquirePromise->set_value(std::make_pair(InstanceAllocation{}, errInfo));
        return;
    }

    ProcessInstanceInfo(spec, instanceResp.info);
    acquirePromise->set_value(std::make_pair(instanceResp.info, err));
}

std::pair<InstanceAllocation, ErrorInfo> FaasInsManager::AcquireInstance(const std::string &stateId,
                                                                         std::shared_ptr<InvokeSpec> spec)
{
    YRLOG_DEBUG("start acquire instance, functon id is {}, request id is {}, state id is {}",
                spec->functionMeta.functionId, spec->requestId, stateId);
    auto acquireSpec = BuildAcquireRequest(spec, stateId);
    auto acquirePromise = std::make_shared<std::promise<std::pair<InstanceAllocation, ErrorInfo>>>();
    auto acquireFuture = acquirePromise->get_future().share();
    auto weak_this = weak_from_this();
    this->fsClient->InvokeAsync(
        acquireSpec->requestInvoke,
        [weak_this, spec, acquirePromise](const NotifyRequest &req, const ErrorInfo &err) -> void {
            if (auto this_ptr = weak_this.lock(); this_ptr) {
                this_ptr->ProcecssAcquireResult(req, spec, acquirePromise);
            }
        });
    auto status = acquireFuture.wait_for(std::chrono::seconds(
        spec->opts.acquireTimeout == 0 ? FAAS_DEFALUT_ACQUIRE_TIMEOUT : spec->opts.acquireTimeout));
    if (status != std::future_status::ready) {
        YRLOG_ERROR("acquire instance timeout, state id is {}, req id is {}", stateId, spec->requestId);
        return std::make_pair(InstanceAllocation{}, ErrorInfo(ErrorCode::ERR_ACQUIRE_TIMEOUT, ModuleCode::RUNTIME,
                                                              "acquire instance timeout, req id is " + spec->requestId +
                                                                  " , state id is " + stateId));
    }
    return acquireFuture.get();
}

void FaasInsManager::AcquireInstanceAsync(std::shared_ptr<InvokeSpec> invokeSpec)
{
    if (invokeSpec->opts.acquireTimeout == 0) {
        invokeSpec->opts.acquireTimeout = FAAS_DEFALUT_ACQUIRE_TIMEOUT;
    }
    YRLOG_DEBUG("start acquire instance async, function: {}, timeout: {}, request: {}, trace: {}",
                invokeSpec->functionMeta.functionId, invokeSpec->opts.acquireTimeout, invokeSpec->requestId,
                invokeSpec->traceId);
    auto acquireSpec = BuildAcquireRequest(invokeSpec);
    YRLOG_INFO("acquire instance to {} for {}, trace: {}, acquire req id :{}, invoke req id : {}",
               acquireSpec->invokeInstanceId, invokeSpec->functionMeta.functionId, invokeSpec->traceId,
               acquireSpec->requestId, invokeSpec->requestId);
    auto weak_this = weak_from_this();
    this->fsClient->InvokeAsync(
        acquireSpec->requestInvoke,
        [weak_this, acquireSpec, invokeSpec](const NotifyRequest &notifyReq, const ErrorInfo &err) -> void {
            if (auto this_ptr = weak_this.lock(); this_ptr) {
                this_ptr->ProcessAsynAcquireResult(notifyReq, acquireSpec, err, invokeSpec);
            }
        },
        acquireSpec->opts.timeout);
}

/***********************************************************************************************************************
The acquire request can fail for the following reasons:
    1. The function system reports an error with error codes 1003 or 1007. This is considered a FaaSScheduler failure,
    and the scheduler ID needs to be the next from the Hash ring.
    2. The function system reports an error with an error code other than 1003 or 1007. No action is taken.
    3. The FaaS reports an error with an error code less than 10000. This is considered a user error.
    Scheduling is interrupted, and the error is throw.
    4. The FaaS reports an error with an error code greater than 10000. No action is taken.
***********************************************************************************************************************/
void FaasInsManager::AcquireCallback(const std::shared_ptr<InvokeSpec> acquireSpec, const ErrorInfo &errInfo,
                                     const InstanceResponse &resp, const std::shared_ptr<InvokeSpec> invokeSpec)
{
    YRLOG_DEBUG("finished to acquire, acquire req id: {}, invoke req id: {}, invoke trace id: {}",
                acquireSpec->requestId, invokeSpec->requestId, invokeSpec->traceId);
    auto resource = GetRequestResource(invokeSpec);
    auto returnErr = errInfo;

    if (!errInfo.OK() || resp.errorCode != FAAS_INS_REQ_SUCCESS_CODE) {
        auto code = errInfo.OK() ? resp.errorCode : errInfo.Code();
        auto msg = errInfo.OK() ? resp.errorMessage : errInfo.Msg();
        YRLOG_WARN(
            "failed to acquire, acquire req id: {}, invoke trace id: {}, invoke req id: {}, invoke scheduler id: {},  "
            "code: {}, msg: {}",
            acquireSpec->requestId, acquireSpec->traceId, invokeSpec->requestId, acquireSpec->invokeInstanceId, code,
            msg);
        this->EraseCreatingInsInfo(resource, "");
        this->ChangeCreateFailNum(resource, true);
    }

    if (!errInfo.OK()) {
        // When the caller specifies a scheduler ID, but the ID is invalid, obtain the new scheduler ID in the hash
        // ring.
        if (errInfo.Code() == YR::Libruntime::ErrorCode::ERR_INSTANCE_EXITED ||
            errInfo.Code() == YR::Libruntime::ErrorCode::ERR_INSTANCE_NOT_FOUND ||
            (errInfo.Code() == YR::Libruntime::ERR_REQUEST_BETWEEN_RUNTIME_BUS && errInfo.IsAckTimeout()) ||
            errInfo.Code() == YR::Libruntime::ERR_FINALIZED) {
            UpdateSpecSchedulerIds(invokeSpec, acquireSpec->invokeInstanceId);
            auto schedulerId =
                csHash->NextRetry(invokeSpec->functionMeta.functionId, invokeSpec->schedulerInfos, true);
            if (schedulerId == ALL_SCHEDULER_UNAVAILABLE || schedulerId.empty()) {
                returnErr =
                    ErrorInfo(ErrorCode::ERR_ALL_SCHEDULER_UNAVALIABLE, "all scheduler instance is unavailable");
            } else {
                invokeSpec->SetSchedulerInstanceId(schedulerId);
                returnErr = ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION,
                                      "send acquire to faas scheduler failed, id: " + acquireSpec->invokeInstanceId);
            }
        }
        return scheduleInsCb(resource, returnErr, IsRemainIns(resource));
    }
    if (resp.errorCode != FAAS_INS_REQ_SUCCESS_CODE) {
        nlohmann::json faasMessageJson;
        faasMessageJson["code"] = resp.errorCode;
        faasMessageJson["message"] = resp.errorMessage;
        std::string faasMessage = faasMessageJson.dump();
        if (resp.errorCode < FAAS_USER_CODE_THRESHOLD || userErrSet.find(resp.errorCode) != userErrSet.end()) {
            returnErr = ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, faasMessage);
        } else if (resp.errorCode == ERR_NO_INSTANCE_AVAILABLE) {
            UpdateSpecSchedulerIds(invokeSpec, acquireSpec->invokeInstanceId);
            auto schedulerId = csHash->Next(invokeSpec->functionMeta.functionId, invokeSpec->schedulerInfos, true);
            if (schedulerId == ALL_SCHEDULER_UNAVAILABLE || schedulerId.empty()) {
                returnErr =
                    ErrorInfo(ErrorCode::ERR_ALL_SCHEDULER_UNAVALIABLE, "all scheduler instance is unavailable");
            } else {
                invokeSpec->SetSchedulerInstanceId(schedulerId);
                returnErr = ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, faasMessage);
            }
        } else {
            returnErr = ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, faasMessage);
        }
        return scheduleInsCb(resource, returnErr, IsRemainIns(resource));
    }

    auto faasInsInfo = std::make_shared<InstanceInfo>();
    faasInsInfo->instanceId = resp.info.instanceId;
    faasInsInfo->leaseId = resp.info.leaseId;
    faasInsInfo->idleTime = 0;
    faasInsInfo->forceInvoke = resp.info.forceInvoke;
    faasInsInfo->unfinishReqNum = 0;
    faasInsInfo->available = true;
    faasInsInfo->traceId = invokeSpec->traceId;
    faasInsInfo->faasInfo = FaasAllocationInfo{resp.info.functionId, resp.info.funcSig, resp.info.tLeaseInterval,
                                               acquireSpec->invokeInstanceId, acquireSpec->functionMeta.functionId};
    faasInsInfo->reporter = std::make_shared<ReportRecord>();
    this->HandleFaasInsInfo(faasInsInfo, resource);
    if (faasInsInfo->faasInfo.tLeaseInterval > 0) {
        this->StartReleaseTimer(resource, faasInsInfo->leaseId);
    }
    this->ChangeCreateFailNum(resource, false);
    YRLOG_INFO("succeed to acquire, lease: {} req: {} trace: {}", faasInsInfo->leaseId, acquireSpec->requestId,
               acquireSpec->traceId);
    return scheduleInsCb(resource, returnErr, IsRemainIns(resource));
}

void FaasInsManager::UpdateSpecSchedulerIds(std::shared_ptr<InvokeSpec> spec, const std::string &schedulerId)
{
    absl::WriterMutexLock insLk(&spec->schedulerInfos->schedulerMtx);
    if (schedulerId.empty()) {
        YRLOG_WARN("scheduler id for req: {} is empty, no need update, trace id is {}", spec->requestId, spec->traceId);
        return;
    }
    bool updated = false;
    for (auto &scheduler : spec->schedulerInfos->schedulerInstanceList) {
        if (scheduler->InstanceID == schedulerId) {
            scheduler->isAvailable = false;
            scheduler->updateTime = YR::GetCurrentTimestampNs();
            updated = true;
            YRLOG_WARN("success update scheduler info, update status of scheduler:{} to unavailable, trace id is {}",
                       schedulerId, spec->traceId);
            break;
        }
    }
    if (!updated) {
        spec->schedulerInfos->schedulerInstanceList.push_back(std::make_shared<SchedulerInstance>(
            SchedulerInstance{"", schedulerId, YR::GetCurrentTimestampNs(), false}));
        YRLOG_WARN("success update scheduler info, update status of scheduler:{} to unavailable, trace id is {}",
                   schedulerId, spec->traceId);
    }
}

void FaasInsManager::HandleFaasInsInfo(std::shared_ptr<InstanceInfo> &faasInsInfo, const RequestResource &resource)
{
    YRLOG_DEBUG("start handler fass acquire info, instance id: {}, lease id: {}", faasInsInfo->instanceId,
                faasInsInfo->leaseId);
    auto info = GetOrAddRequestResourceInfo(resource);
    // ensure atomicity of erase creating and add instances.
    // avoid creating unnecessary instances when judge NeedCreateNewIns
    {
        absl::WriterMutexLock lock(&info->mtx);
        this->EraseCreatingInsInfoBare(info, "");
        this->AddInsInfoBare(info, faasInsInfo);
    }
    {
        if (faasInsInfo->faasInfo.tLeaseInterval <= 0) {
            return;
        }
        absl::WriterMutexLock lock(&this->leaseMtx);
        this->globalLeases[faasInsInfo->leaseId] = resource;
    }
}

void FaasInsManager::ReleaseHandler(const RequestResource &resource, const std::string &leaseId)
{
    std::shared_ptr<InstanceInfo> faasInfoDel;
    bool needDel = false;

    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return;
    }
    std::shared_ptr<InstanceInfo> faasInfo;
    {
        absl::ReaderMutexLock lock(&info->mtx);
        if (info->instanceInfos.find(leaseId) == info->instanceInfos.end()) {
            YRLOG_DEBUG("lease has already release, id: {}", leaseId);
            return;
        }
        faasInfo = info->instanceInfos[leaseId];
    }

    std::string functionId = "";
    {
        absl::WriterMutexLock resInfoLock(&info->mtx);
        absl::ReaderMutexLock faasInfoLock(&faasInfo->mtx);
        if (faasInfo->unfinishReqNum >= 1) {
            YRLOG_DEBUG("instance is not available, do not release, lease id{}", leaseId);
            return;
        }
        info->avaliableInstanceInfos.erase(leaseId);
        functionId = faasInfo->faasInfo.functionId;
    }
    faasInfoDel = faasInfo;
    needDel = true;
    {
        absl::WriterMutexLock lock(&info->mtx);
        info->instanceInfos.erase(leaseId);
    }
    DecreaseCreatedInstanceNum();
    EraseResourceInfoMap(resource, REQUEST_RESOURCE_USE_COUNT);

    if (needDel) {
        YRLOG_DEBUG("start send release instance req, function id {}, lease id {}", functionId, leaseId);
        this->ReleaseInstanceAsync(faasInfoDel);
    }
}

void FaasInsManager::StartReleaseTimer(const RequestResource &resource, const std::string &leaseId)
{
    YRLOG_DEBUG("start release timer, lease id: {}", leaseId);
    auto weakThis = weak_from_this();
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return;
    }
    int tLeaseInterval;
    {
        absl::ReaderMutexLock lock(&info->mtx);
        tLeaseInterval = info->tLeaseInterval;
    }
    if (tLeaseInterval <= 0) {
        return ReleaseHandler(resource, leaseId);
    }
    tw_->CreateTimer(RELEASE_DELAYTIME, 1, [weakThis, this, leaseId, resource]() {
        auto thisPtr = weakThis.lock();
        if (!thisPtr) {
            return;
        }
        ReleaseHandler(resource, leaseId);
    });
}

void FaasInsManager::ReleaseInstanceAsync(const std::shared_ptr<InstanceInfo> &ins)
{
    std::string leaseId;
    {
        absl::ReaderMutexLock lock(&ins->mtx);
        if (ins->faasInfo.tLeaseInterval <= 0) {
            return;
        }
        leaseId = ins->leaseId;
        YRLOG_DEBUG("start aysnc release instance, leaseId id is {}", leaseId);
    }
    auto req = BuildReleaseReq(ins);
    auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
    this->fsClient->InvokeAsync(messageSpec, [leaseId](const NotifyRequest &notifyReq, const ErrorInfo &err) -> void {
        if (notifyReq.code() != common::ERR_NONE) {
            YRLOG_ERROR("release instance async failed, leaseId id is {}, req id is {}, msg is {}", leaseId,
                        notifyReq.requestid(), notifyReq.message());
        } else {
            YRLOG_DEBUG("release instance async success, lease id is {}", leaseId);
        }
    });
}

ErrorInfo FaasInsManager::ReleaseInstance(const std::string &leaseId, const std::string &stateId, bool abnormal,
                                          std::shared_ptr<InvokeSpec> spec)
{
    ErrorInfo err;
    if (!stateId.empty()) {
        auto insInfo = std::make_shared<InstanceInfo>();
        insInfo->faasInfo.schedulerInstanceID = spec->GetSchedulerInstanceId();
        insInfo->faasInfo.schedulerFunctionID = spec->opts.schedulerFunctionId;
        insInfo->reporter = std::make_shared<ReportRecord>();
        insInfo->stateId = stateId;

        err = SendReleaseInstanceReq(insInfo, spec);
        return err;
    }

    auto resource = GetRequestResource(spec);
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return err;
    }
    absl::WriterMutexLock lock(&info->mtx);
    auto &insInfos = info->instanceInfos;
    for (auto it = insInfos.begin(); it != insInfos.end();) {
        YRLOG_DEBUG("state id is {}, lease id is {}, instance id is {}", it->second->stateId, it->second->leaseId,
                    it->second->instanceId);
        if (it->second->leaseId == leaseId) {
            err = SendReleaseInstanceReq(it->second, spec);
            it = insInfos.erase(it);
            info->avaliableInstanceInfos.erase(leaseId);
            DecreaseCreatedInstanceNum();
            break;
        } else {
            ++it;
        }
    }
    return err;
}

ErrorInfo FaasInsManager::SendReleaseInstanceReq(const std::shared_ptr<InstanceInfo> &ins,
                                                 std::shared_ptr<InvokeSpec> spec)
{
    std::string leaseId;
    std::string stateId;
    {
        absl::ReaderMutexLock lock(&ins->mtx);
        leaseId = ins->leaseId;
        stateId = ins->stateId;
    }
    auto req = BuildReleaseReq(ins);
    auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
    auto releasePromise = std::make_shared<std::promise<ErrorInfo>>();
    auto releaseFuture = releasePromise->get_future().share();
    YRLOG_DEBUG("start release instance, leaseId id is {}, state id is {}, req id is {}", leaseId, stateId,
                req.requestid());
    this->fsClient->InvokeAsync(
        messageSpec, [releasePromise, stateId](const NotifyRequest &notifyReq, const ErrorInfo &err) -> void {
            if (notifyReq.code() != common::ERR_NONE) {
                YRLOG_ERROR("receive release notify, req id is {}, state id is {} code is {}, msg is {}",
                            notifyReq.requestid(), stateId, fmt::underlying(notifyReq.code()), notifyReq.message());
            }
            releasePromise->set_value(ErrorInfo(ErrorCode(notifyReq.code()), notifyReq.message()));
        });
    auto status = releaseFuture.wait_for(
        std::chrono::milliseconds(spec->opts.timeout == 0 ? FAAS_DEFALUT_INVOKE_TIMEOUT : spec->opts.timeout));
    if (status != std::future_status::ready) {
        YRLOG_ERROR("release instance timeout, state id is {}, lease id is {}, req id is {}", stateId, leaseId,
                    spec->requestId);
        return ErrorInfo(ErrorCode::ERR_INIT_CONNECTION_FAILED,
                         "release instance timeout, req id is " + spec->requestId + " , state id is " + stateId);
    }
    return releaseFuture.get();
}

InvokeRequest FaasInsManager::BuildReleaseReq(const std::shared_ptr<InstanceInfo> &ins)
{
    absl::ReaderMutexLock lock(&ins->mtx);
    YRLOG_DEBUG("start build release instance req, instance id {}, lease id {}", ins->instanceId, ins->leaseId);
    InvokeRequest req;
    auto requestId = YR::utility::IDGenerator::GenRequestId();
    req.set_requestid(requestId);
    req.set_traceid(ins->traceId);
    req.set_instanceid(ins->faasInfo.schedulerInstanceID);
    req.set_function(ins->faasInfo.schedulerFunctionID);
    req.add_returnobjectids(YR::utility::IDGenerator::GenObjectId(requestId, 0));
    std::string acquireOps;
    if (ins->stateId.empty()) {
        acquireOps = "release#" + ins->leaseId;
    } else {
        acquireOps = "release#" + ins->stateId;
    }
    YRLOG_DEBUG("requst id is {}, instance id is {}, function is {}, acquire ops is {}", req.requestid(),
                req.instanceid(), req.function(), acquireOps);
    Arg *pbArg = req.add_args();
    pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    pbArg->set_value(META_PREFIX + acquireOps);
    auto report = ins->reporter->Report(true);
    nlohmann::json instanceReport;
    instanceReport["ProcReqNum"] = report.procReqNum;
    instanceReport["AvgProcTime"] = report.avgProcTime;
    instanceReport["MaxProcTime"] = report.maxProcTime;
    instanceReport["IsAbnormal"] = report.isAbnormal;
    YRLOG_DEBUG(
        "lease id is {}, instance id is {}, procReqNum is {}, avgProcTime is {}, maxProcTime is {}, isAbnormal: {}",
        ins->leaseId, ins->instanceId, report.procReqNum, report.avgProcTime, report.maxProcTime, report.isAbnormal);
    pbArg = req.add_args();
    pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    pbArg->set_value(META_PREFIX + instanceReport.dump());
    pbArg = req.add_args();
    pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
    pbArg->set_value(META_PREFIX + ins->traceId);
    return req;
}

std::shared_ptr<InvokeSpec> FaasInsManager::BuildAcquireRequest(std::shared_ptr<InvokeSpec> invokeSpec,
                                                                const std::string &stateId)
{
    auto acquireSpec = std::make_shared<InvokeSpec>();
    acquireSpec->jobId = invokeSpec->jobId;
    acquireSpec->requestId = YR::utility::IDGenerator::GenRequestId();
    auto dsObjectId = YR::utility::IDGenerator::GenObjectId(acquireSpec->requestId, 0);
    acquireSpec->returnIds.push_back(DataObject(dsObjectId));
    acquireSpec->traceId = invokeSpec->traceId;

    acquireSpec->functionMeta.functionId =
        invokeSpec->opts.schedulerFunctionId.empty() ? this->schedulerFuncKey : invokeSpec->opts.schedulerFunctionId;
    acquireSpec->functionMeta.apiType = libruntime::ApiType::Posix;
    std::string schedulerId = invokeSpec->GetSchedulerInstanceId();
    if (!schedulerId.empty()) {
        acquireSpec->invokeInstanceId = schedulerId;
    } else {
        schedulerId = csHash->NextRetry(invokeSpec->functionMeta.functionId);
        acquireSpec->invokeInstanceId = schedulerId;
    }

    std::string acquireFunction;
    if (stateId == "") {
        acquireFunction = "acquire#" + invokeSpec->functionMeta.functionId;
    } else {
        acquireFunction = "acquire#" + invokeSpec->functionMeta.functionId + ";" + stateId;
    }
    InvokeArg acquireFunctionArg;
    acquireFunctionArg.dataObj = std::make_shared<DataObject>(0, acquireFunction.size());
    acquireFunctionArg.dataObj->data->MemoryCopy(acquireFunction.data(), acquireFunction.size());
    acquireSpec->invokeArgs.push_back(acquireFunctionArg);

    std::unordered_map<std::string, std::vector<unsigned char>> instanceRequirement;
    if (invokeSpec->designatedInstanceID.empty()) {
        std::unordered_map<std::string, u_int64_t> resourceSpecs;
        resourceSpecs.insert({"CPU", invokeSpec->opts.cpu});
        resourceSpecs.insert({"Memory", invokeSpec->opts.memory});
        for (auto &resource : invokeSpec->opts.customResources) {
            resourceSpecs.insert({resource.first, static_cast<u_int64_t>(resource.second)});
        }
        nlohmann::json resourceMapJson = resourceSpecs;
        std::string resourceMapString = resourceMapJson.dump();
        YRLOG_DEBUG("instance resource : {}", resourceMapString);
        std::vector<unsigned char> vec(resourceMapString.begin(), resourceMapString.end());
        instanceRequirement[INSTANCE_REQUIREMENT_RESOURKEY] = vec;
    } else {
        std::vector<unsigned char> vec(invokeSpec->designatedInstanceID.begin(),
                                       invokeSpec->designatedInstanceID.end());
        instanceRequirement[INSTANCE_REQUIREMENT_INSKEY] = vec;
    }
    if (invokeSpec->opts.trafficLimited) {
        std::string strTrue = "true";
        std::vector<unsigned char> vec(strTrue.begin(), strTrue.end());
        instanceRequirement[INSTANCETRAFFICLIMITED] = vec;
    }
    std::string podName =
        Config::Instance().POD_NAME().empty() ? Config::Instance().HOSTNAME() : Config::Instance().POD_NAME();
    if (!podName.empty()) {
        std::vector<unsigned char> vec(podName.begin(), podName.end());
        instanceRequirement[INSTANCE_CALLER_POD_NAME] = vec;
    }

    if (!invokeSpec->functionMeta.poolLabel.empty()) {
        std::vector<unsigned char> vec(invokeSpec->functionMeta.poolLabel.begin(),
                                       invokeSpec->functionMeta.poolLabel.end());
        instanceRequirement[INSTANCE_REQUIREMENT_POOLLABELKEY] = vec;
    }
    if (!invokeSpec->opts.invokeLabels.empty()) {
        nlohmann::json invokeLabelsJson = invokeSpec->opts.invokeLabels;
        std::string invokeLabelsString = invokeLabelsJson.dump();
        std::vector<unsigned char> vec(invokeLabelsString.begin(), invokeLabelsString.end());
        instanceRequirement[INSTANCE_REQUIREMENT_INVOKE_LABEL] = vec;
    }
    if (invokeSpec->opts.instanceSession) {
        nlohmann::json instanceSessionJson;
        instanceSessionJson["sessionID"] = invokeSpec->opts.instanceSession->sessionID;
        instanceSessionJson["sessionTTL"] = invokeSpec->opts.instanceSession->sessionTTL;
        instanceSessionJson["concurrency"] = invokeSpec->opts.instanceSession->concurrency;
        std::string instanceSessionJsonString = instanceSessionJson.dump();
        std::vector<unsigned char> vec(instanceSessionJsonString.begin(), instanceSessionJsonString.end());
        instanceRequirement[INSTANCE_SESSION_CONFIG] = vec;
    }
    nlohmann::json instanceRequirementJson = instanceRequirement;
    std::string instanceRequirementStr = instanceRequirementJson.dump();
    InvokeArg instanceRequirementArg;
    instanceRequirementArg.dataObj = std::make_shared<DataObject>(0, instanceRequirementStr.size());
    instanceRequirementArg.dataObj->data->MemoryCopy(instanceRequirementStr.data(), instanceRequirementStr.size());
    acquireSpec->invokeArgs.push_back(instanceRequirementArg);

    InvokeArg traceIdArg;
    traceIdArg.dataObj = std::make_shared<DataObject>(0, invokeSpec->traceId.size());
    traceIdArg.dataObj->data->MemoryCopy(invokeSpec->traceId.data(), invokeSpec->traceId.size());
    acquireSpec->invokeArgs.push_back(traceIdArg);

    acquireSpec->opts.timeout = invokeSpec->opts.acquireTimeout;

    acquireSpec->BuildInstanceInvokeRequest(*libRuntimeConfig);

    return acquireSpec;
}

void FaasInsManager::AddInsInfoBare(std::shared_ptr<RequestResourceInfo> info,
                                    std::shared_ptr<InstanceInfo> &faasInsInfo)
{
    bool needIncreaseInsNum = true;
    if (auto it = info->instanceInfos.find(faasInsInfo->leaseId); it != info->instanceInfos.end()) {
        YRLOG_DEBUG("already exist, instace id {}, current unfinish req num {}, lease id {}", faasInsInfo->instanceId,
                    faasInsInfo->unfinishReqNum, faasInsInfo->leaseId);
        needIncreaseInsNum = false;
    }
    info->instanceInfos[faasInsInfo->leaseId] = faasInsInfo;
    info->tLeaseInterval = faasInsInfo->faasInfo.tLeaseInterval;
    info->avaliableInstanceInfos[faasInsInfo->leaseId] = faasInsInfo;
    if (needIncreaseInsNum) {
        IncreaseCreatedInstanceNum();
    }
}

void FaasInsManager::StartBatchRenewTimer()
{
    absl::ReaderMutexLock lock(&this->leaseMtx);
    if (!this->leaseTimer) {
        YRLOG_DEBUG("start batch renew timer");
        auto timer = CreateBatchRenewTimer();
        this->leaseTimer = timer;
    }
}

void FaasInsManager::ProcessBatchRenewResult(const NotifyRequest &notifyReq, const std::string &functionId,
                                             const ErrorInfo &err, std::vector<std::string> leaseIds)
{
    bool failedFlag = false;
    if (!err.OK()) {
        YRLOG_WARN("failed to batch renew instance, req id {}, {}", notifyReq.requestid(), err.CodeAndMsg());
        failedFlag = true;
    }
    if (notifyReq.code() != common::ERR_NONE) {
        YRLOG_WARN("failed to batch renew instance, req id {}, code {}, msg {}", notifyReq.requestid(),
                   fmt::underlying(notifyReq.code()), notifyReq.message());
        if (notifyReq.code() == common::ERR_INSTANCE_NOT_FOUND || notifyReq.code() == common::ERR_INSTANCE_EXITED) {
            ChangeInstanceSchedulerId(functionId, leaseIds);
            return;
        }
        failedFlag = true;
    }
    auto [instanceResp, errInfo] = GetFaasBatchInstanceRsp(notifyReq);
    if (!errInfo.OK()) {
        YRLOG_WARN("failed to batch renew instance, req id {}, {}", notifyReq.requestid(), errInfo.CodeAndMsg());
        failedFlag = true;
    }
    std::vector<std::string> decreaseLeaseIds;
    std::vector<std::string> reacquireLeaseIds;
    if (failedFlag) {
        decreaseLeaseIds.resize(leaseIds.size());
        std::copy(leaseIds.begin(), leaseIds.end(), decreaseLeaseIds.begin());
    } else {
        for (const auto &[leaseId, allocErrInfo] : instanceResp.instanceAllocFailed) {
            YRLOG_WARN("failed to renew instance, lease id {}, req id {}, errCode: {}, errMsg: {}", leaseId,
                       notifyReq.requestid(), allocErrInfo.errorCode, allocErrInfo.errorMessage);
            if (allocErrInfo.errorCode == ERR_RENEW_INSTANCE_LEASE_NOT_FOUND) {
                reacquireLeaseIds.push_back(leaseId);
            } else {
                decreaseLeaseIds.push_back(leaseId);
            }
        }
    }
    absl::WriterMutexLock lock(&this->leaseMtx);
    for (std::string &leaseId : reacquireLeaseIds) {
        if (this->globalLeases.find(leaseId) != this->globalLeases.end()) {
            auto info = GetRequestResourceInfo(this->globalLeases.find(leaseId)->second);
            if (info == nullptr) {
                continue;
            }
            absl::ReaderMutexLock infoLock(&info->mtx);
            if (info->instanceInfos.find(leaseId) != info->instanceInfos.end()) {
                auto insInfo = info->instanceInfos[leaseId];
                absl::WriterMutexLock instanceLock(&insInfo->mtx);
                insInfo->needReacquire = true;
            }
        }
    }
    for (std::string &leaseId : decreaseLeaseIds) {
        if (this->globalLeases.find(leaseId) != this->globalLeases.end()) {
            DelInsInfo(leaseId, this->globalLeases.find(leaseId)->second);
            this->globalLeases.erase(leaseId);
        }
    }
    if (!failedFlag) {
        this->tLeaseInterval = instanceResp.tLeaseInterval;
    }
}

void FaasInsManager::ChangeInstanceSchedulerId(const std::string &functionId, std::vector<std::string> &leaseIds)
{
    auto otherSchedulerInstanceId = this->csHash->Next(functionId, true);
    for (std::string &leaseId : leaseIds) {
        auto it = this->globalLeases.find(leaseId);
        if (it != this->globalLeases.end()) {
            auto instanceInfo = GetInstanceInfo(it->second, it->first);
            if (instanceInfo == nullptr) {
                it = this->globalLeases.erase(it);
                continue;
            }
            {
                YRLOG_WARN("failed to renew instance {}, scheduler {} change to {}", leaseId,
                           instanceInfo->faasInfo.schedulerInstanceID, otherSchedulerInstanceId);
                absl::WriterMutexLock insInfoLock(&instanceInfo->mtx);
                instanceInfo->faasInfo.schedulerInstanceID = otherSchedulerInstanceId;
            }
        }
    }
}

std::vector<unsigned char> BuildReacquireInstanceData(const RequestResource &resource)
{
    nlohmann::json reacquireData;
    std::unordered_map<std::string, u_int64_t> resourceSpecs;
    resourceSpecs.insert({"CPU", resource.opts.cpu});
    resourceSpecs.insert({"Memory", resource.opts.memory});
    for (auto &cr : resource.opts.customResources) {
        resourceSpecs.insert({cr.first, static_cast<u_int64_t>(cr.second)});
    }
    nlohmann::json resourceMapJson = resourceSpecs;
    std::string resourceMapString = resourceMapJson.dump();
    std::vector<unsigned char> resourceKeyVec(resourceMapString.begin(), resourceMapString.end());
    reacquireData[INSTANCE_REQUIREMENT_RESOURKEY] = resourceKeyVec;
    if (resource.opts.instanceSession) {
        nlohmann::json instanceSessionJson;
        instanceSessionJson["sessionID"] = resource.opts.instanceSession->sessionID;
        instanceSessionJson["sessionTTL"] = resource.opts.instanceSession->sessionTTL;
        instanceSessionJson["concurrency"] = resource.opts.instanceSession->concurrency;
        std::string instanceSessionJsonString = instanceSessionJson.dump();
        std::vector<unsigned char> vec(instanceSessionJsonString.begin(), instanceSessionJsonString.end());
        reacquireData["instanceSessionConfig"] = vec;
    }
    if (!resource.opts.invokeLabels.empty()) {
        nlohmann::json invokeLabelsJson = resource.opts.invokeLabels;
        std::string invokeLabelsString = invokeLabelsJson.dump();
        std::vector<unsigned char> vec(invokeLabelsString.begin(), invokeLabelsString.end());
        reacquireData[INSTANCE_REQUIREMENT_INVOKE_LABEL] = vec;
    }
    if (!resource.functionMeta.poolLabel.empty()) {
        std::vector<unsigned char> vec(resource.functionMeta.poolLabel.begin(), resource.functionMeta.poolLabel.end());
        reacquireData[INSTANCE_REQUIREMENT_POOLLABELKEY] = vec;
    }
    std::string reacquireDataStr = reacquireData.dump();
    std::vector<unsigned char> reacquireDataVec(reacquireDataStr.begin(), reacquireDataStr.end());
    return reacquireDataVec;
}

void FaasInsManager::BatchRenewHandler()
{
    std::unordered_map<FaasInfoForBatchRenew, nlohmann::json, FaasInfoForBatchRenewFn> instanceReportMap;
    std::unordered_map<FaasInfoForBatchRenew, std::string, FaasInfoForBatchRenewFn> targetNamesMap;
    std::unordered_map<FaasInfoForBatchRenew, std::vector<std::string>, FaasInfoForBatchRenewFn> leaseIdsMap;
    {
        absl::ReaderMutexLock lock(&this->leaseMtx);
        if (this->leaseTimer) {
            auto weakPtr = weak_from_this();
            tw_->ExecuteByTimer(this->leaseTimer, this->tLeaseInterval / RETAIN_TIME_RATE, [weakPtr]() {
                if (auto thisPtr = weakPtr.lock(); thisPtr) {
                    thisPtr->BatchRenewHandler();
                }
            });
        } else {
            YRLOG_DEBUG("batch renew is cancelled");
            return;
        }
        auto it = this->globalLeases.begin();
        while (it != this->globalLeases.end()) {
            auto instanceInfo = GetInstanceInfo(it->second, it->first);
            if (instanceInfo == nullptr) {
                it = this->globalLeases.erase(it);
                continue;
            }
            int i = 0;
            FaasInfoForBatchRenew faasInfo = FaasInfoForBatchRenew(instanceInfo->faasInfo, i);
            while (leaseIdsMap[faasInfo].size() >= BATCH_RENEW_LEASE_NUM) {
                i++;
                faasInfo = FaasInfoForBatchRenew(instanceInfo->faasInfo, i);
            }
            auto report = instanceInfo->reporter->Report(false);
            nlohmann::json instanceReport;
            instanceReport["procReqNum"] = report.procReqNum;
            instanceReport["avgProcTime"] = report.avgProcTime;
            instanceReport["maxProcTime"] = report.maxProcTime;
            instanceReport["isAbnormal"] = report.isAbnormal;
            if (instanceInfo->needReacquire) {
                instanceReport["reacquireData"] = BuildReacquireInstanceData(it->second);
                instanceReport["functionKey"] = instanceInfo->faasInfo.functionId;
                absl::WriterMutexLock instanceLock(&instanceInfo->mtx);
                instanceInfo->needReacquire = false;
            }
            nlohmann::json &reportMap = instanceReportMap[faasInfo];
            reportMap[it->first] = instanceReport;
            std::string &targetNames = targetNamesMap[faasInfo];
            if (targetNames.empty()) {
                targetNames = it->first;
            } else {
                targetNames += "," + it->first;
            }
            std::vector<std::string> &leaseIds = leaseIdsMap[faasInfo];
            leaseIds.push_back(it->first);
            it++;
        }
    }
    for (const auto &[faasInfo, targetNames] : targetNamesMap) {
        InvokeRequest req;
        req.set_requestid(YR::utility::IDGenerator::GenRequestId());
        std::string traceId = YR::utility::IDGenerator::GenTraceId();
        req.set_traceid(traceId);
        req.set_instanceid(faasInfo.schedulerInstanceID);
        req.set_function(faasInfo.schedulerFunctionID);
        req.add_returnobjectids("obj-" + YR::utility::IDGenerator::GenObjectId());

        auto acquireOps = "batchRetain#" + targetNames;
        Arg *pbArg = req.add_args();
        pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
        pbArg->set_value(META_PREFIX + acquireOps);
        pbArg = req.add_args();
        pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
        pbArg->set_value(META_PREFIX + instanceReportMap[faasInfo].dump());
        pbArg = req.add_args();
        pbArg->set_type(Arg_ArgType::Arg_ArgType_VALUE);
        pbArg->set_value(META_PREFIX + traceId);
        YRLOG_DEBUG("batch renew to {} {}, batchIndex: {}, req: {}, lease:{}", faasInfo.schedulerFunctionID,
                    faasInfo.schedulerInstanceID, faasInfo.batchIndex, req.requestid(), targetNames);
        auto messageSpec = std::make_shared<InvokeMessageSpec>(std::move(req));
        auto weak_this = weak_from_this();
        std::vector<std::string> &leaseIds = leaseIdsMap[faasInfo];
        auto functionId = faasInfo.functionId;
        this->fsClient->InvokeAsync(
            messageSpec,
            [weak_this, leaseIds, functionId](const NotifyRequest &notifyReq, const ErrorInfo &err) -> void {
                if (auto this_ptr = weak_this.lock(); this_ptr) {
                    this_ptr->ProcessBatchRenewResult(notifyReq, functionId, err, leaseIds);
                }
            });
    }
}

std::shared_ptr<YR::utility::Timer> FaasInsManager::CreateBatchRenewTimer()
{
    auto weakPtr = weak_from_this();
    return tw_->CreateTimer(this->tLeaseInterval / RETAIN_TIME_RATE, 1, [weakPtr]() {
        if (auto thisPtr = weakPtr.lock(); thisPtr) {
            thisPtr->BatchRenewHandler();
        }
    });
}

void FaasInsManager::UpdateSchedulerInfo(const std::string &schedulerFuncKey,
                                         const std::vector<SchedulerInstance> &schedulerInstanceList)
{
    YRLOG_INFO("recv update scheduler info");
    this->csHash->ResetAll(schedulerInstanceList);
    absl::WriterMutexLock lock(&this->schedulerFuncKeyMtx);
    if (this->schedulerFuncKey.empty()) {
        this->schedulerFuncKey = schedulerFuncKey;
    }
}

std::string FaasInsManager::GetSchedulerKey()
{
    absl::ReaderMutexLock lock(&this->schedulerFuncKeyMtx);
    return this->schedulerFuncKey;
}

std::string FaasInsManager::GetFunctionIdWithLabel(const RequestResource &resource)
{
    std::ostringstream functionKey;
    functionKey << resource.functionMeta.functionId << "-";
    functionKey << resource.opts.cpu << "-" << resource.opts.memory;
    if (!resource.opts.invokeLabels.empty()) {
        functionKey << "-{ ";
        for (auto it = resource.opts.invokeLabels.begin(); it != resource.opts.invokeLabels.end();) {
            functionKey << it->first << ":" << it->second;
            if (++it != resource.opts.invokeLabels.end()) {
                functionKey << ", ";
            }
        }
        functionKey << " }";
    }
    return functionKey.str();
}

}  // namespace Libruntime
}  // namespace YR
