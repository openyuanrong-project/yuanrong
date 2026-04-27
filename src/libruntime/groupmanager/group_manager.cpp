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

#include "src/libruntime/groupmanager/group_manager.h"

namespace YR {
namespace Libruntime {

void GroupManager::AddGroup(std::shared_ptr<Group> group)
{
    std::string name = group->GetGroupName();
    std::lock_guard<std::mutex> insLk(groupMtx);
    group->SetCreateSpecs(groupSpecs_[name]);
    groups_[name] = group;
}

bool GroupManager::IsGroupExist(const std::string &groupName)
{
    std::lock_guard<std::mutex> insLk(groupMtx);
    if (groups_.find(groupName) == groups_.end()) {
        return false;
    } else {
        return true;
    }
}

std::shared_ptr<Group> GroupManager::GetGroup(const std::string &groupName)
{
    std::lock_guard<std::mutex> insLk(groupMtx);
    if (groups_.find(groupName) != groups_.end()) {
        return groups_[groupName];
    }
    return nullptr;
}

void GroupManager::AddSpec(std::shared_ptr<InvokeSpec> spec)
{
    std::lock_guard<std::mutex> insLk(groupMtx);
    if (groupSpecs_.find(spec->opts.groupName) == groupSpecs_.end()) {
        groupSpecs_[spec->opts.groupName] = std::vector<std::shared_ptr<InvokeSpec>>{spec};
    } else {
        groupSpecs_[spec->opts.groupName].push_back(spec);
    }
}

bool GroupManager::IsInsReady(const std::string &groupName)
{
    std::lock_guard<std::mutex> insLk(groupMtx);
    if (groups_.find(groupName) == groups_.end()) {
        return false;
    }
    return groups_[groupName]->IsReady();
}

ErrorInfo GroupManager::GroupCreate(const std::string &groupName)
{
    if (groups_.find(groupName) == groups_.end()) {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                         "group not exist, please select correct group");
    }
    return groups_[groupName]->GroupCreate();
}

ErrorInfo GroupManager::Wait(const std::string &groupName)
{
    if (groups_.find(groupName) == groups_.end()) {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                         "group not exist, please select correct group");
    }
    return groups_[groupName]->Wait();
}

void GroupManager::Terminate(const std::string &groupName)
{
    if (groups_.find(groupName) == groups_.end()) {
        YRLOG_WARN("there is no group named {}, please check param", groupName);
        return;
    }
    groups_[groupName]->Terminate();
    groups_.erase(groupName);
    groupSpecs_.erase(groupName);
}

ErrorInfo GroupManager::Suspend(const std::string &groupName)
{
    if (groups_.find(groupName) == groups_.end()) {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                         "group not exist, please select correct group");
    }
    return groups_[groupName]->Suspend();
}

ErrorInfo GroupManager::Resume(const std::string &groupName)
{
    if (groups_.find(groupName) == groups_.end()) {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                         "group not exist, please select correct group");
    }
    return groups_[groupName]->Resume();
}

void GroupManager::Stop()
{
    for (auto &pair : groups_) {
        pair.second->SetRunFlag();
        pair.second.reset();
    }
}

ErrorInfo GroupManager::Accelerate(const std::string &groupName, const AccelerateMsgQueueHandle &handle,
                                   HandleReturnObjectCallback callback)
{
    if (groups_.find(groupName) == groups_.end()) {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                         "group not exist, please select correct group");
    }
    return groups_[groupName]->Accelerate(handle, callback);
}
}  // namespace Libruntime
}  // namespace YR
