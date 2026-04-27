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
#include "src/libruntime/groupmanager/group.h"
#include "src/utility/id_generator.h"

namespace YR {
namespace Libruntime {
Group::Group(const std::string &name, const std::string &inputTenantId, GroupOpts &inputOpts,
             std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
             std::shared_ptr<MemoryStore> memStore)
    : groupName(name),
      tenantId(inputTenantId),
      opts(inputOpts),
      fsClient(client),
      waitManager_(waitManager),
      memStore_(memStore)
{
}

Group::Group(const std::string &name, const std::string &inputTenantId, InstanceRange &inputRange,
             std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
             std::shared_ptr<MemoryStore> memStore)
    : groupName(name),
      tenantId(inputTenantId),
      range(inputRange),
      fsClient(client),
      waitManager_(waitManager),
      memStore_(memStore)
{
}

Group::Group(const std::string &name, const std::string &inputTenantId, FunctionGroupOptions &inputOpts,
             std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
             std::shared_ptr<MemoryStore> memStore)
    : groupName(name),
      tenantId(inputTenantId),
      functionGroupOpts(inputOpts),
      fsClient(client),
      waitManager_(waitManager),
      memStore_(memStore)
{
}

ErrorInfo Group::GroupCreate()
{
    if (isSendReq) {
        YRLOG_DEBUG("this group: {} has send reqs before", groupName);
        return ErrorInfo();
    }
    if (createSpecs.empty()) {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                         "there is no create req in this group, please select correct group");
    }
    isSendReq = true;
    CreateRequests reqs = BuildCreateReqs();
    YRLOG_DEBUG("start send group create req, req id is {}", reqs.requestid());
    auto weak_this = weak_from_this();
    this->fsClient->GroupCreateAsync(
        reqs,
        [weak_this](const CreateResponses &responses) {
            if (auto this_ptr = weak_this.lock(); this_ptr) {
                this_ptr->CreateRespHandler(responses);
            }
        },
        [weak_this](const NotifyRequest &req) {
            if (auto this_ptr = weak_this.lock(); this_ptr) {
                this_ptr->CreateNotifyHandler(req);
            }
        });
    return ErrorInfo();
}

ErrorInfo Group::Wait()
{
    int64_t timeoutMs = opts.timeout != NO_TIMEOUT ? opts.timeout * S_TO_MS : NO_TIMEOUT;
    std::vector<std::string> idList;
    for (auto spec : createSpecs) {
        idList.push_back(spec->returnIds[0].id);
    }
    if (idList.empty()) {
        return ErrorInfo();
    }
    auto internalWaitResult = waitManager_->WaitUntilReady(idList, createSpecs.size(), timeoutMs);
    if (internalWaitResult->exceptionIds.size() > 0) {
        return internalWaitResult->exceptionIds.begin()->second;
    }
    if (internalWaitResult->unreadyIds.size() > 0) {
        return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, ModuleCode::CORE,
                         "group create timeout, unready obj ids is [" + internalWaitResult->unreadyIds[0] + ", ...]");
    }
    isReady = true;
    return ErrorInfo();
}

void Group::Terminate()
{
    runFlag = false;
    std::string tmpGroupId;
    {
        std::lock_guard<std::mutex> lock(groupIdMtx);
        tmpGroupId = groupId;
    }
    YRLOG_DEBUG("start terminate group ins, group name is {}, group id is {}", groupName, tmpGroupId);
    KillRequest killReq;
    killReq.set_requestid(YR::utility::IDGenerator::GenRequestId());
    killReq.set_instanceid(tmpGroupId);
    killReq.set_payload("");
    killReq.set_signal(libruntime::Signal::KillGroupInstance);
    this->fsClient->KillAsync(killReq, [](KillResponse resp, const ErrorInfo &err) -> void {
        YRLOG_ERROR("get termiate group ins response, resp code is {}, resp msg is {}", fmt::underlying(resp.code()),
                    resp.message());
    });
    SetTerminateError();
}

ErrorInfo Group::Suspend()
{
    YRLOG_DEBUG("start suspend group ins, group name is {}", groupName);
    return Signal(libruntime::Signal::GroupSuspend);
}

ErrorInfo Group::Resume()
{
    YRLOG_DEBUG("start resume group ins, group name is {}", groupName);
    return Signal(libruntime::Signal::GroupResume);
}

ErrorInfo Group::Signal(libruntime::Signal signal)
{
    if (!runFlag) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "group already terminate");
    }
    if (!isSendReq) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "group not create");
    }
    KillRequest killReq;
    {
        std::lock_guard<std::mutex> lock(groupIdMtx);
        killReq.set_instanceid(groupId);
        if (groupId.empty()) {
            return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "groupId is empty");
        }
    }
    killReq.set_signal(signal);
    killReq.set_payload("");
    auto promise = std::promise<ErrorInfo>();
    auto future = promise.get_future();
    this->fsClient->KillAsync(killReq, [&promise](KillResponse resp, ErrorInfo err) -> void {
        if (resp.code() != common::ERR_NONE) {
            ErrorInfo errInfo(static_cast<ErrorCode>(resp.code()), ModuleCode::RUNTIME, resp.message());
            promise.set_value(errInfo);
        } else {
            promise.set_value(ErrorInfo());
        }
    });
    auto err = future.get();
    return err;
}

void Group::SetRunFlag()
{
    this->runFlag = false;
}

bool Group::IsReady()
{
    return this->isReady;
}

std::string Group::GetGroupName()
{
    return this->groupName;
}

void Group::SetCreateSpecs(std::vector<std::shared_ptr<InvokeSpec>> &specs)
{
    this->createSpecs = specs;
}

InstanceRange Group::GetInstanceRange()
{
    return this->range;
}

FunctionGroupOptions Group::GetFunctionGroupOptions()
{
    return this->functionGroupOpts;
}

}  // namespace Libruntime
}  // namespace YR