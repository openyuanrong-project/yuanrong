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
#include "src/libruntime/invoke_order_manager.h"

namespace YR {
namespace Libruntime {
class NamedGroup : public Group {
public:
    NamedGroup(const std::string &name) : Group(name) {};
    NamedGroup(const std::string &name, const std::string &inputTenantId, GroupOpts &inputOpts,
               std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
               std::shared_ptr<MemoryStore> memStore, std::shared_ptr<InvokeOrderManager> invokeOrderMgr);

private:
    CreateRequests BuildCreateReqs() override;
    void SetTerminateError() override;
    void CreateRespHandler(const CreateResponses &resps) override;
    void CreateNotifyHandler(const NotifyRequest &req) override;
    void NotifyInstances();
    std::shared_ptr<InvokeOrderManager> invokeOrderMgr_;
};
}  // namespace Libruntime
}  // namespace YR
