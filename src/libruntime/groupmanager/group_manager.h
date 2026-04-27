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
#include "src/libruntime/groupmanager/group.h"
#include "src/libruntime/groupmanager/named_group.h"
#include "src/libruntime/groupmanager/range_group.h"
namespace YR {
namespace Libruntime {
class GroupManager {
public:
    ErrorInfo GroupCreate(const std::string &groupName);
    ErrorInfo Wait(const std::string &groupName);
    void Terminate(const std::string &groupName);
    ErrorInfo Suspend(const std::string &groupName);
    ErrorInfo Resume(const std::string &groupName);
    void Stop();
    bool IsGroupExist(const std::string &groupName);
    void AddSpec(std::shared_ptr<InvokeSpec> spec);
    void AddGroup(std::shared_ptr<Group> group);
    std::shared_ptr<Group> GetGroup(const std::string &groupName);
    bool IsInsReady(const std::string &groupName);
    ErrorInfo Accelerate(const std::string &groupName, const AccelerateMsgQueueHandle &handle,
                         HandleReturnObjectCallback callback);

private:
    mutable std::mutex groupMtx;
    std::unordered_map<std::string, std::shared_ptr<Group>> groups_;
    std::unordered_map<std::string, std::vector<std::shared_ptr<InvokeSpec>>> groupSpecs_;
};

}  // namespace Libruntime
}  // namespace YR