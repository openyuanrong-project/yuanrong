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

#include "src/libruntime/groupmanager/range_group.h"

namespace YR {
namespace Libruntime {
RangeGroup::RangeGroup(const std::string &name, const std::string &inputTenantId, InstanceRange &inputRange,
                       std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
                       std::shared_ptr<MemoryStore> memStore, std::shared_ptr<InvokeOrderManager> invokeOrderMgr)
    : Group(name, inputTenantId, inputRange, client, waitManager, memStore), invokeOrderMgr_(invokeOrderMgr)
{
}

RangeGroup::RangeGroup(const std::string &name, const std::string &inputTenantId, FunctionGroupOptions &inputOpts,
                       std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
                       std::shared_ptr<MemoryStore> memStore, std::shared_ptr<InvokeOrderManager> invokeOrderMgr)
    : Group(name, inputTenantId, inputOpts, client, waitManager, memStore), invokeOrderMgr_(invokeOrderMgr)
{
}

void RangeGroup::CreateRespHandler(const CreateResponses &resps)
{
    HandleCreateResp(resps);
}

void RangeGroup::CreateNotifyHandler(const NotifyRequest &req)
{
    HandleCreateNotify(req);
}

void RangeGroup::HandleCreateResp(const CreateResponses &resps)
{
    YRLOG_DEBUG("recieve group create response, resp code is {}, message is {}, runflag is {}",
                fmt::underlying(resps.code()), resps.message(), runFlag.load());
    if (!runFlag) {
        return;
    }
    {
        std::lock_guard<std::mutex> lock(groupIdMtx);
        groupId = resps.groupid();
        YRLOG_DEBUG("group id is {}", groupId);
    }
    if (resps.code() != common::ERR_NONE) {
        this->memStore_->SetError(createSpecs[0]->returnIds[0].id, ErrorInfo(static_cast<ErrorCode>(resps.code()),
                                                                             ModuleCode::CORE, resps.message(), true));
        this->memStore_->SetInstanceIds(createSpecs[0]->returnIds[0].id, {});
        invokeOrderMgr_->RemoveInstance(createSpecs[0]);
    } else {
        for (int i = 0; i < resps.instanceids_size(); i++) {
            YRLOG_DEBUG("instance_{} id is {}", i, resps.instanceids(i));
            if (createSpecs[0]->opts.needOrder) {
                invokeOrderMgr_->CreateGroupInstance(resps.instanceids(i));
            }
            auto ret = instanceIds_.insert(resps.instanceids(i));
            if (!ret.second) {
                YRLOG_DEBUG("instance id: {} already exist in group instance set", resps.instanceids(i));
                continue;
            }
            this->memStore_->AddReturnObject(resps.instanceids(i));
            this->memStore_->SetInstanceId(resps.instanceids(i), resps.instanceids(i));
        }
        std::vector<std::string> instanceIds(instanceIds_.begin(), instanceIds_.end());
        this->memStore_->SetInstanceIds(createSpecs[0]->returnIds[0].id, instanceIds);
    }
}

void RangeGroup::HandleCreateNotify(const NotifyRequest &req)
{
    YRLOG_DEBUG("recieve group create notify, req code is {}, message is {}, runflag is {}",
                fmt::underlying(req.code()), req.message(), runFlag.load());
    if (!runFlag) {
        return;
    }
    if (req.code() != common::ERR_NONE) {
        this->memStore_->SetError(createSpecs[0]->returnIds[0].id,
                                  ErrorInfo(static_cast<ErrorCode>(req.code()), ModuleCode::CORE, req.message(), true));
        SetInstancesError(ErrorInfo(static_cast<ErrorCode>(req.code()), ModuleCode::CORE, req.message(), true));
        invokeOrderMgr_->RemoveInstance(createSpecs[0]);
    } else {
        NotifyInstances();
        this->memStore_->SetReady(createSpecs[0]->returnIds[0].id);
        SetInstancesReady();
        isReady = true;
    }
}

CreateRequests RangeGroup::BuildCreateReqs()
{
    CreateRequests reqs;
    reqs.set_tenantid(tenantId);
    reqs.set_requestid(YR::utility::IDGenerator::GenRequestId());
    reqs.set_traceid(createSpecs[0]->traceId);
    CreateRequest *req = reqs.add_requests();
    *req = createSpecs[0]->requestCreate;
    GroupOptions *options = reqs.mutable_groupopt();
    options->set_groupname(groupName);
    options->set_timeout(range.rangeOpts.timeout);
    options->set_samerunninglifecycle(range.sameLifecycle);
    return reqs;
}

void RangeGroup::SetTerminateError()
{
    for (auto &instanceId : instanceIds_) {
        this->memStore_->SetError(instanceId, ErrorInfo(ErrorCode::ERR_FINALIZED, ModuleCode::RUNTIME,
                                                        "group ins had been terminated, return obj id is: " +
                                                            instanceId + " , instance id is: " + instanceId,
                                                        true));
    }
    RemoveInstances();
}

void RangeGroup::SetInstancesError(ErrorInfo err)
{
    for (auto &instanceId : instanceIds_) {
        this->memStore_->SetError(instanceId, err);
    }
}

void RangeGroup::SetInstancesReady()
{
    for (auto &instanceId : instanceIds_) {
        this->memStore_->SetReady(instanceId);
    }
}

void RangeGroup::NotifyInstances()
{
    if (!createSpecs[0]->opts.needOrder) {
        return;
    }
    for (auto &instanceId : instanceIds_) {
        this->invokeOrderMgr_->NotifyGroupInstance(instanceId);
    }
}

void RangeGroup::RemoveInstances()
{
    if (!createSpecs[0]->opts.needOrder) {
        return;
    }
    for (auto &instanceId : instanceIds_) {
        this->invokeOrderMgr_->RemoveGroupInstance(instanceId);
    }
}
}  // namespace Libruntime
}  // namespace YR
