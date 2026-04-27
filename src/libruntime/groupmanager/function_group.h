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

#pragma once
#include "src/libruntime/groupmanager/group.h"
#include "src/libruntime/groupmanager/range_group.h"
#include "src/libruntime/invoke_order_manager.h"
#include "src/libruntime/invokeadaptor/request_manager.h"

namespace YR {
namespace Libruntime {
using ReturnedObjectHandler = std::function<void(const NotifyRequest &req, const std::shared_ptr<InvokeSpec> &spec)>;
class FunctionGroup : public RangeGroup {
public:
    FunctionGroup(const std::string &name, const std::string &inputTenantId, FunctionGroupOptions &inputOpts,
                  std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
                  std::shared_ptr<MemoryStore> memStore, std::shared_ptr<InvokeOrderManager> invokeOrderMgr,
                  std::shared_ptr<RequestManager> requestManager = nullptr, ReturnedObjectHandler handler = nullptr);
    void SetInvokeSpec(std::shared_ptr<InvokeSpec> invokeSpec);
    ErrorInfo Accelerate(const AccelerateMsgQueueHandle &handle, HandleReturnObjectCallback callback);
    void AddInstance(const std::vector<std::string> &insIds);
    void Stop();

protected:
    CreateRequests BuildCreateReqs() override;
    void CreateRespHandler(const CreateResponses &resps) override;
    void CreateNotifyHandler(const NotifyRequest &req) override;
    void SetTerminateError() override;

private:
    void InvokeHandler();
    std::vector<std::shared_ptr<InvokeSpec>> BuildInvokeSpec(const std::shared_ptr<InvokeSpec> &spec,
                                                             const std::vector<std::string> &instanceIds);
    void InvokeByInstanceIds(const std::shared_ptr<InvokeSpec> &spec, const std::vector<std::string> &instanceIds);
    void InvokeNotifyHandler(const NotifyRequest &req, const ErrorInfo &err);
    void AssembleAffinityRequest(CreateRequests &reqs);
    void HandleReturnObjectLoop();

private:
    std::shared_ptr<RequestManager> requestManager_ = nullptr;
    ReturnedObjectHandler returnedObjectHandler_;
    std::mutex finishTaskMtx_;
    std::shared_ptr<InvokeSpec> invokeSpec_ = nullptr;
    std::thread t_;
    std::vector<std::shared_ptr<AccelerateMsgQueue>> queues_;
    HandleReturnObjectCallback handleReturnObjectCallback_;
    std::atomic<bool> stop_{false};
};
}  // namespace Libruntime
}  // namespace YR
