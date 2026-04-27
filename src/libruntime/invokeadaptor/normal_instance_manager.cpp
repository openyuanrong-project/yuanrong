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

#include "normal_instance_manager.h"
namespace YR {
namespace Libruntime {
const std::string TASK_INSTANCE_TYPE = "task";

void NormalInsManager::UpdateConfig(int inputRecycleTimeMs)
{
    YRLOG_DEBUG("update recycle time value: {}", inputRecycleTimeMs);
    this->recycleTimeMs = inputRecycleTimeMs;
}

void NormalInsManager::SendKillReq(const std::string &insId)
{
    KillRequest killReq;
    killReq.set_requestid(YR::utility::IDGenerator::GenRequestId());
    killReq.set_instanceid(insId);
    killReq.set_payload("");
    killReq.set_signal(libruntime::Signal::KillInstance);
    YRLOG_DEBUG("start send kill req, ins id is {}", insId);
    this->fsClient->KillAsync(killReq, [insId](KillResponse rsp, const ErrorInfo &err) -> void {
        if (rsp.code() != common::ERR_NONE) {
            YRLOG_WARN("kill req send failed, instance id is {}", insId);
        }
    });
    if (deleteInsCallback_) {
        deleteInsCallback_(insId);
    }
}

void NormalInsManager::ScaleCancel(const RequestResource &resource, size_t reqNum, bool cleanAll)
{
    auto [cancel, cancelIns] = NeedCancelCreatingIns(resource, reqNum, cleanAll);
    if (cancel && !cancelIns.empty()) {
        for (const auto &ins : cancelIns) {
            YRLOG_DEBUG("start to cancel creating instance {}", ins);
            SendKillReq(ins);
        }
    }
    EraseResourceInfoMap(resource);
    return;
}

void NormalInsManager::ScaleDown(const std::shared_ptr<InvokeSpec> spec, bool isInstanceNormal)
{
    auto resource = GetRequestResource(spec);
    auto id = spec->invokeInstanceId;
    YRLOG_DEBUG("start scale down ins, ins id is : {}, with delay time : {}, ins is normal : {}", id,
                libRuntimeConfig->recycleTime, isInstanceNormal);
    if (isInstanceNormal) {
        return StartNormalInsScaleDownTimer(resource, id);
    }
    this->DelInsInfo(id, resource);
    SendKillReq(id);
    return;
}

bool NormalInsManager::ScaleUp(std::shared_ptr<InvokeSpec> spec, size_t reqNum)
{
    AddRequestResourceInfo(spec);
    return this->CreateInstance(spec, reqNum);
}

std::shared_ptr<InvokeSpec> NormalInsManager::BuildCreateSpec(std::shared_ptr<InvokeSpec> spec)
{
    auto createSpec = std::make_shared<InvokeSpec>();
    createSpec->jobId = spec->jobId;
    createSpec->functionMeta = spec->functionMeta;
    createSpec->opts = spec->opts;
    createSpec->invokeType = libruntime::InvokeType::CreateInstanceStateless;
    createSpec->traceId = YR::utility::IDGenerator::GenTraceId(spec->jobId);
    createSpec->requestId = YR::utility::IDGenerator::GenRequestId();
    createSpec->returnIds = spec->returnIds;
    if (createSpec->opts.customExtensions.find(LIFECYCLE) != createSpec->opts.customExtensions.end()) {
        YRLOG_WARN("task does not support detached mode");
        createSpec->opts.customExtensions.erase(LIFECYCLE);
    }
    createSpec->opts.labels.push_back(TASK_INSTANCE_TYPE);
    createSpec->BuildInstanceCreateRequest(*libRuntimeConfig);
    return createSpec;
}

bool NormalInsManager::CreateInstance(std::shared_ptr<InvokeSpec> spec, size_t reqNum)
{
    auto resource = GetRequestResource(spec);
    auto needCreatePair = NeedCreateNewIns(resource, reqNum);
    if (!needCreatePair.first) {
        return false;
    }
    SendCreateReq(spec, needCreatePair.second);
    return true;
}

void NormalInsManager::SendCreateReq(std::shared_ptr<InvokeSpec> spec, size_t deleyTime)
{
    auto createSpec = BuildCreateSpec(spec);
    auto resource = GetRequestResource(spec);
    auto insInfo = std::make_shared<CreatingInsInfo>("", 0);
    this->AddCreatingInsInfo(resource, insInfo);
    auto weakThis = weak_from_this();
    tw_->CreateTimer(deleyTime * MILLISECOND_UNIT, 1, [this, createSpec, insInfo, weakThis]() {
        auto thisPtr = weakThis.lock();
        if (thisPtr == nullptr) {
            return;
        }
        YRLOG_DEBUG("send create instance request, req id is {}", createSpec->requestId);
        requestManager->PushRequest(createSpec);
        auto respCallback = [this, createSpec, insInfo, weakThis](const CreateResponse &resp) -> void {
            auto thisPtr = weakThis.lock();
            if (thisPtr == nullptr) {
                return;
            }
            HandleCreateResponse(createSpec, resp, insInfo);
        };
        if (!createSpec->opts.device.name.empty()) {
            absl::ReaderMutexLock lock(&createCostMtx);
            createCostMap[createSpec->opts.device.name] = TimeMeasurement(DEFAULT_CREATE_DURATION);
            createCostMap[createSpec->opts.device.name].StartTimer(createSpec->requestId);
            YRLOG_DEBUG("start timer for {}, reqID: {}", createSpec->opts.device.name, createSpec->requestId);
        }
        this->fsClient->CreateAsync(createSpec->requestCreate, respCallback,
                                    std::bind(&NormalInsManager::HandleCreateNotify, this, std::placeholders::_1));
    });
}

void NormalInsManager::HandleCreateResponse(std::shared_ptr<InvokeSpec> spec, const CreateResponse &resp,
                                            std::shared_ptr<CreatingInsInfo> insInfo)
{
    auto instanceId = resp.instanceid();
    spec->instanceId = instanceId;
    {
        absl::WriterMutexLock lock(&insInfo->mtx);
        insInfo->instanceId = instanceId;
    }
    if (resp.code() != common::ERR_NONE) {
        YRLOG_ERROR(
            "start handle fail create response, req id is {}, trace id is {}, instance id is {}, code is {}, message "
            "is {}",
            spec->requestId, spec->traceId, instanceId, fmt::underlying(resp.code()), resp.message());
        auto resource = GetRequestResource(spec);
        HandleFailCreateNotify(spec, resource);
        scheduleInsCb(resource, ErrorInfo(static_cast<ErrorCode>(resp.code()), ModuleCode::CORE, resp.message(), true),
                      IsRemainIns(resource));
    }
}

void NormalInsManager::HandleCreateNotify(const NotifyRequest &req)
{
    if (!runFlag) {
        return;
    }
    ErrorInfo errInfo(static_cast<ErrorCode>(req.code()), ModuleCode::CORE, req.message(), true);
    if (errInfo.Finalized()) {
        YRLOG_WARN("req id: {} is finalized, ignore it.", req.requestid());
        return;
    }

    auto reqId = req.requestid();
    auto rawRequestId = YR::utility::IDGenerator::DecodeRawRequestId(reqId).first;
    std::shared_ptr<InvokeSpec> createSpec;
    bool specExist = requestManager->PopRequest(rawRequestId, createSpec);
    if (!specExist) {
        YRLOG_DEBUG(
            "create request id : {} did not exit in request manager, may be the normal function instance create "
            "request has been canceled or finished.",
            reqId);
        return;
    }
    YRLOG_DEBUG("start process create notify id is: {}, {}", req.requestid(), createSpec->instanceId);
    RequestResource resource = GetRequestResource(createSpec);
    if (req.code() != common::ERR_NONE) {
        YRLOG_ERROR(
            "handle normal function instance create failed notify or response, request id is: {}, instance id is: {}, "
            "trace id is: {},err code is {}, err msg is {}",
            reqId, createSpec->instanceId, createSpec->traceId, fmt::underlying(errInfo.Code()), errInfo.Msg());
        HandleFailCreateNotify(createSpec, resource);
    } else {
        HandleSuccessCreateNotify(createSpec, resource, req);
    }
    bool existCreateCost = false;
    {
        absl::ReaderMutexLock lock(&createCostMtx);
        if (createCostMap.find(createSpec->opts.device.name) != createCostMap.end()) {
            existCreateCost = true;
        }
    }
    if (existCreateCost) {
        absl::WriterMutexLock lock(&createCostMtx);
        if (createCostMap.find(createSpec->opts.device.name) != createCostMap.end()) {
            createCostMap[createSpec->opts.device.name].StopTimer(reqId, req.code() == common::ERR_NONE);
        }
    }
    scheduleInsCb(resource, errInfo, IsRemainIns(resource));
}

void NormalInsManager::HandleFailCreateNotify(const std::shared_ptr<InvokeSpec> createSpec,
                                              const RequestResource &resource)
{
    if (!createSpec->instanceId.empty()) {
        this->DelInsInfo(createSpec->instanceId, resource);
        this->SendKillReq(createSpec->instanceId);
    }
    this->ChangeCreateFailNum(resource, true);
    this->EraseCreatingInsInfo(resource, createSpec->instanceId, false);
}

void NormalInsManager::HandleSuccessCreateNotify(const std::shared_ptr<InvokeSpec> createSpec,
                                                 const RequestResource &resource, const NotifyRequest &req)
{
    this->ChangeCreateFailNum(resource, false);
    auto info = GetOrAddRequestResourceInfo(resource);
    // ensure atomicity of erase creating and add instances.
    // avoid creating unnecessary instances when judge NeedCreateNewIns
    bool isErased;
    {
        absl::WriterMutexLock lock(&info->mtx);
        // If isErased is false, it means the instance has already been canceled and should not be added to insInfo.
        isErased = this->EraseCreatingInsInfoBare(info, createSpec->instanceId, true);
        if (isErased) {
            this->AddInsInfoBare(createSpec, info);
            if (req.has_runtimeinfo() && !req.runtimeinfo().route().empty()) {
                this->memoryStore->SetInstanceRoute(createSpec->returnIds[0].id, req.runtimeinfo().route());
            }
        }
    }
    if (isErased) {
        this->StartNormalInsScaleDownTimer(resource, createSpec->instanceId);
    } else {
        SendKillReq(createSpec->instanceId);
        EraseResourceInfoMap(resource, REQUEST_RESOURCE_USE_COUNT);
    }
}

void NormalInsManager::ScaleDownHandler(const RequestResource &resource, const std::string &id)
{
    auto info = GetRequestResourceInfo(resource);
    if (info == nullptr) {
        return;
    }
    std::shared_ptr<InstanceInfo> insInfo;
    {
        absl::WriterMutexLock infoLock(&info->mtx);
        if (info->instanceInfos.find(id) == info->instanceInfos.end()) {
            return;
        }
        insInfo = info->instanceInfos[id];
        absl::ReaderMutexLock insInfoLock(&insInfo->mtx);
        if (insInfo->unfinishReqNum > 0) {
            YRLOG_DEBUG("instance still has unfinish req, do not kill, id is {}, unfinish req num is {}", id,
                        insInfo->unfinishReqNum);
            return;
        }
        DelInsInfoBare(id, info);
    }
    SendKillReq(id);
    EraseResourceInfoMap(resource, REQUEST_RESOURCE_USE_COUNT);
    scheduleInsCb(resource, ErrorInfo(), IsRemainIns(resource));
}

void NormalInsManager::StartNormalInsScaleDownTimer(const RequestResource &resource, const std::string &id)
{
    auto weakPtr = weak_from_this();
    std::shared_ptr<InstanceInfo> insInfo = GetInstanceInfo(resource, id);
    if (insInfo == nullptr) {
        return;
    }
    auto timer = tw_->CreateTimer(libRuntimeConfig->recycleTime * MILLISECOND_UNIT, 1, [this, weakPtr, resource, id]() {
        if (auto thisPtr = weakPtr.lock(); thisPtr) {
            ScaleDownHandler(resource, id);
        }
    });
    CancelScaleDownTimer(insInfo);
    absl::WriterMutexLock insLock(&insInfo->mtx);
    insInfo->scaleDownTimer = timer;
}

void NormalInsManager::AddInsInfo(const std::shared_ptr<InvokeSpec> createSpec, const RequestResource &resource)
{
    auto info = GetOrAddRequestResourceInfo(resource);
    absl::WriterMutexLock lock(&info->mtx);
    AddInsInfoBare(createSpec, info);
}

void NormalInsManager::AddInsInfoBare(const std::shared_ptr<InvokeSpec> createSpec,
                                      std::shared_ptr<RequestResourceInfo> info)
{
    bool happened;
    auto insInfo = std::make_shared<InstanceInfo>();
    insInfo->instanceId = createSpec->instanceId;
    happened = info->instanceInfos.emplace(createSpec->instanceId, insInfo).second;
    info->avaliableInstanceInfos.emplace(createSpec->instanceId, insInfo);
    if (happened) {
        IncreaseCreatedInstanceNum();
    }
}

void NormalInsManager::StartBatchRenewTimer() {}
}  // namespace Libruntime
}  // namespace YR