/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.
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

#include <unordered_set>

#include "src/libruntime/groupmanager/group.h"
#include "src/libruntime/invoke_order_manager.h"
namespace YR {
namespace Libruntime {
class RangeGroup : public Group {
public:
    RangeGroup(const std::string &name, const std::string &inputTenantId, InstanceRange &inputRange,
               std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
               std::shared_ptr<MemoryStore> memStore, std::shared_ptr<InvokeOrderManager> invokeOrderMgr);

protected:
    RangeGroup(const std::string &name, const std::string &inputTenantId, FunctionGroupOptions &inputOpts,
               std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
               std::shared_ptr<MemoryStore> memStore, std::shared_ptr<InvokeOrderManager> invokeOrderMgr);

protected:
    CreateRequests BuildCreateReqs() override;
    void CreateRespHandler(const CreateResponses &resps) override;
    void HandleCreateResp(const CreateResponses &resps);
    void CreateNotifyHandler(const NotifyRequest &req) override;
    void HandleCreateNotify(const NotifyRequest &req);
    void SetTerminateError() override;
    void SetInstancesError(ErrorInfo err);
    void SetInstancesReady();
    void NotifyInstances();
    void RemoveInstances();
    InstanceRange range;
    std::shared_ptr<InvokeOrderManager> invokeOrderMgr_;
    std::unordered_set<std::string> instanceIds_;
};
}  // namespace Libruntime
}  // namespace YR
