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

#include "src/libruntime/groupmanager/function_group.h"
namespace YR {
namespace Libruntime {

FunctionGroup::FunctionGroup(const std::string &name, const std::string &inputTenantId, FunctionGroupOptions &inputOpts,
                             std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
                             std::shared_ptr<MemoryStore> memStore, std::shared_ptr<InvokeOrderManager> invokeOrderMgr,
                             std::shared_ptr<RequestManager> requestManager, ReturnedObjectHandler handler)
    : RangeGroup(name, inputTenantId, inputOpts, client, waitManager, memStore, invokeOrderMgr),
      requestManager_(requestManager),
      returnedObjectHandler_(handler)
{
}

CreateRequests FunctionGroup::BuildCreateReqs()
{
    CreateRequests reqs;
    reqs.set_tenantid(tenantId);
    reqs.set_requestid(YR::utility::IDGenerator::GenRequestId());
    reqs.set_traceid(createSpecs[0]->traceId);
    AssembleAffinityRequest(reqs);
    GroupOptions *options = reqs.mutable_groupopt();
    options->set_groupname(groupName);
    options->set_timeout(functionGroupOpts.timeout);
    options->set_samerunninglifecycle(functionGroupOpts.sameLifecycle);
    return reqs;
}

void FunctionGroup::AssembleAffinityRequest(CreateRequests &reqs)
{
    auto bundleLabelPrefix = groupName + "_bundle_";
    for (int i = 0; i < functionGroupOpts.functionGroupSize; i++) {
        auto bundleIndex = i / functionGroupOpts.bundleSize;
        auto bundleLabel = bundleLabelPrefix + std::to_string(bundleIndex);
        CreateRequest *req = reqs.add_requests();
        *req = createSpecs[0]->requestCreate;
        SchedulingOptions *schedulingOps = req->mutable_schedulingops();
        auto *scheduleAffinity = schedulingOps->mutable_scheduleaffinity();
        req->set_requestid(YR::utility::IDGenerator::GenRequestId());
        if (i % functionGroupOpts.bundleSize == 0) {
            req->add_labels(bundleLabel);
        } else {
            auto op = std::make_shared<LabelExistsOperator>();
            op->SetKey(bundleLabel);
            auto aff = std::make_shared<InstanceRequiredAffinity>();
            aff->SetLabelOperators({op});
            aff->UpdatePbAffinity(scheduleAffinity);
        }
    }
}

void FunctionGroup::CreateRespHandler(const CreateResponses &resps)
{
    HandleCreateResp(resps);
    if (resps.code() != common::ERR_NONE) {
        if (invokeSpec_) {
            InvokeHandler();
        } else {
            auto ids = memStore_->UnbindObjRefInReq(createSpecs[0]->requestId);
            auto errorInfo = memStore_->DecreGlobalReference(ids);
            if (!errorInfo.OK()) {
                YRLOG_WARN("failed to decrease by requestid {}. Code: {}, MCode: {}, Msg: {}",
                           createSpecs[0]->requestId, fmt::underlying(errorInfo.Code()),
                           fmt::underlying(errorInfo.MCode()), errorInfo.Msg());
            }
        }
    }
}

void FunctionGroup::CreateNotifyHandler(const NotifyRequest &req)
{
    HandleCreateNotify(req);
    if (invokeSpec_) {
        InvokeHandler();
    } else {
        auto ids = memStore_->UnbindObjRefInReq(createSpecs[0]->requestId);
        auto errorInfo = memStore_->DecreGlobalReference(ids);
        if (!errorInfo.OK()) {
            YRLOG_WARN("failed to decrease by requestid {}. Code: {}, MCode: {}, Msg: {}", createSpecs[0]->requestId,
                       fmt::underlying(errorInfo.Code()), fmt::underlying(errorInfo.MCode()), errorInfo.Msg());
        }
    }
}

void FunctionGroup::InvokeHandler()
{
    auto [instanceIds, tmpErr] = memStore_->GetInstanceIds(createSpecs[0]->returnIds[0].id, functionGroupOpts.timeout);
    if (!tmpErr.OK()) {
        for (const auto &returnId : invokeSpec_->returnIds) {
            memStore_->SetError(returnId.id, tmpErr);
        }
        return;
    }
    InvokeByInstanceIds(invokeSpec_, instanceIds);
}

std::vector<std::shared_ptr<InvokeSpec>> FunctionGroup::BuildInvokeSpec(const std::shared_ptr<InvokeSpec> &spec,
                                                                        const std::vector<std::string> &instanceIds)
{
    auto totalReturnIdSize = spec->returnIds.size();
    if (instanceIds.empty()) {
        YRLOG_ERROR("instanceIds is empty");
        return {};
    }
    auto returnIdSize = totalReturnIdSize / instanceIds.size();
    size_t returnIdIndex = 0;
    std::vector<std::shared_ptr<InvokeSpec>> invokeSpecs;
    for (const auto &instanceId : instanceIds) {
        auto invokeSpec = std::make_shared<InvokeSpec>(*spec);
        auto req = spec->requestInvoke->Immutable();
        invokeSpec->requestInvoke = std::make_shared<InvokeMessageSpec>(std::move(req));
        invokeSpec->requestId = YR::utility::IDGenerator::GenRequestId();
        invokeSpec->requestInvoke->Mutable().set_requestid(invokeSpec->requestId);
        invokeSpec->requestInvoke->Mutable().set_traceid(YR::utility::IDGenerator::GenTraceId(spec->jobId));
        invokeSpec->requestInvoke->Mutable().set_instanceid(instanceId);
        invokeSpec->requestInvoke->Mutable().clear_returnobjectids();
        invokeSpec->returnIds.clear();
        for (size_t i = 0; i < returnIdSize; i++) {
            invokeSpec->returnIds.push_back(spec->returnIds[returnIdIndex]);
            invokeSpec->requestInvoke->Mutable().add_returnobjectids(spec->returnIds[returnIdIndex++].id);
        }
        invokeSpecs.push_back(invokeSpec);
    }
    return invokeSpecs;
}

void FunctionGroup::InvokeByInstanceIds(const std::shared_ptr<InvokeSpec> &spec,
                                        const std::vector<std::string> &instanceIds)
{
    auto totalReturnIdSize = spec->returnIds.size();
    YRLOG_DEBUG("start to invoke function by instance ids, request id: {}, instance num: {}, total return id size: {}",
                spec->requestId, instanceIds.size(), totalReturnIdSize);
    auto finishTaskNum = std::make_shared<int>(0);
    auto invokeSpecs = BuildInvokeSpec(spec, instanceIds);
    for (const auto &invokeSpec : invokeSpecs) {
        requestManager_->PushRequest(invokeSpec);
        fsClient->InvokeAsync(
            invokeSpec->requestInvoke, [this, spec, finishTaskNum](const NotifyRequest &req, const ErrorInfo &err) {
                auto totalTaskNum = spec->opts.functionGroupOpts.functionGroupSize;
                InvokeNotifyHandler(req, err);
                {
                    std::lock_guard<std::mutex> lockGuard(finishTaskMtx_);
                    (*finishTaskNum)++;
                    if (*finishTaskNum < totalTaskNum) {
                        YRLOG_DEBUG("{}/{} task finished, request id: {}, group name: {}", *finishTaskNum, totalTaskNum,
                                    spec->requestId, spec->opts.groupName);
                        return;
                    }
                }
                YRLOG_DEBUG("all task finished, start to terminate group, request id: {}, group name: {}",
                            spec->requestId, spec->opts.groupName);
                // After all tasks are executed, decrease reference and terminate the group.
                auto ids = memStore_->UnbindObjRefInReq(spec->requestId);
                auto errorInfo = memStore_->DecreGlobalReference(ids);
                if (!errorInfo.OK()) {
                    YRLOG_WARN("failed to decrease by requestid {}. Code: {}, MCode: {}, Msg: {}", req.requestid(),
                               fmt::underlying(errorInfo.Code()), fmt::underlying(errorInfo.MCode()), errorInfo.Msg());
                }
                this->Terminate();
            });
    }
}

void FunctionGroup::InvokeNotifyHandler(const NotifyRequest &req, const ErrorInfo &err)
{
    YRLOG_DEBUG("start handle instance function invoke notify, req id is {}", req.requestid());
    std::shared_ptr<InvokeSpec> spec = requestManager_->GetRequest(req.requestid());
    if (spec == nullptr) {
        return;
    }
    if (req.code() != common::ERR_NONE) {
        YRLOG_WARN("instance invoke failed, do not retry, request id: {}, instance id: {}, return id: {}",
                   req.requestid(), spec->invokeInstanceId, spec->returnIds[0].id);
        auto isCreate = false;
        std::vector<YR::Libruntime::StackTraceInfo> stackTraceInfos = GetStackTraceInfos(req);
        YRLOG_DEBUG("get stackTraceInfos from notify request size: {}", stackTraceInfos.size());
        memStore_->SetError(spec->returnIds, ErrorInfo(static_cast<ErrorCode>(req.code()), ModuleCode::CORE,
                                                       req.message(), isCreate, stackTraceInfos));
    } else {
        returnedObjectHandler_(req, spec);
    }
    (void)requestManager_->RemoveRequest(req.requestid());
}

void FunctionGroup::SetInvokeSpec(std::shared_ptr<InvokeSpec> invokeSpec)
{
    invokeSpec_ = invokeSpec;
}

ErrorInfo FunctionGroup::Accelerate(const AccelerateMsgQueueHandle &handle, HandleReturnObjectCallback callback)
{
    std::vector<std::string> instanceIdList(instanceIds_.begin(), instanceIds_.end());
    handleReturnObjectCallback_ = callback;
    std::vector<std::future<KillResponse>> killFutures;
    for (size_t i = 0; i < instanceIdList.size(); i++) {
        AccelerateMsgQueueHandle h = handle;
        h.rank = i;
        auto payload = h.ToJson();
        auto killPromise = std::make_shared<std::promise<KillResponse>>();
        killFutures.emplace_back(killPromise->get_future());
        KillRequest killReq;
        killReq.set_requestid(YR::utility::IDGenerator::GenRequestId());
        killReq.set_instanceid(instanceIdList[i]);
        killReq.set_payload(payload);
        killReq.set_signal(libruntime::Signal::Accelerate);
        fsClient->KillAsync(killReq,
                            [killPromise](KillResponse rsp, const ErrorInfo &err) { killPromise->set_value(rsp); });
    }
    std::vector<AccelerateMsgQueueHandle> handles;
    std::vector<std::string> objIds;
    for (size_t i = 0; i < killFutures.size(); i++) {
        auto killResponse = killFutures[i].get();
        if (killResponse.code() != common::ERR_NONE) {
            return ErrorInfo(static_cast<ErrorCode>(killResponse.code()), ModuleCode::CORE,
                             "Failed to kill instance " + instanceIdList[i] + ", err is: " + killResponse.message());
        }
        auto result = killResponse.message();
        auto h = AccelerateMsgQueueHandle::FromJson(result);
        h.rank = 0;
        objIds.emplace_back(h.name);
        handles.emplace_back(std::move(h));
    }
    auto result = memStore_->GetBuffers(objIds, -1);
    if (!result.first.OK()) {
        return result.first;
    }
    auto buffers = result.second;
    for (size_t i = 0; i < handles.size(); i++) {
        auto queue = std::make_shared<AccelerateMsgQueue>(handles[i], buffers[i]);
        queues_.emplace_back(queue);
    }
    HandleReturnObjectLoop();
    return ErrorInfo();
}

void FunctionGroup::HandleReturnObjectLoop()
{
    t_ = std::thread([this]() {
        while (!stop_) {
            for (size_t i = 0; i < queues_.size(); i++) {
                std::string objId;
                YRLOG_DEBUG("start dequeue invoke request result");
                auto buffer = queues_[i]->Dequeue();
                if (buffer == nullptr) {
                    break;
                }
                if (!handleReturnObjectCallback_) {
                    break;
                }
                auto [err, outBuffer] = handleReturnObjectCallback_(buffer, i, objId);
                YRLOG_DEBUG("{} invoke request result dequeued", objId);
                queues_[i]->SetReadFlag();
                if (!err.OK()) {
                    memStore_->SetError(objId, err);
                    continue;
                }
                memStore_->Put(outBuffer, objId, {}, false);
                memStore_->SetReady(objId);
            }
        }
    });
}

void FunctionGroup::SetTerminateError()
{
    for (auto &instanceId : instanceIds_) {
        this->memStore_->SetError(instanceId, ErrorInfo(ErrorCode::ERR_FINALIZED, ModuleCode::RUNTIME,
                                                        "group ins had been terminated, return obj id is: " +
                                                            instanceId + " , instance id is: " + instanceId,
                                                        true));
    }
    RemoveInstances();
    stop_ = true;
    for (auto &queue : queues_) {
        queue->Stop();
    }
    if (t_.joinable()) {
        t_.join();
    }
}

void FunctionGroup::AddInstance(const std::vector<std::string> &insIds)
{
    for (auto insId : insIds) {
        instanceIds_.emplace(insId);
    }
}

void FunctionGroup::Stop()
{
    this->stop_ = true;
    if (t_.joinable()) {
        t_.join();
    }
}
}  // namespace Libruntime
}  // namespace YR
