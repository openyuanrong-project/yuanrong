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
#include "src/dto/accelerate.h"
#include "src/libruntime/fsclient/fs_client.h"
#include "src/libruntime/invoke_spec.h"
#include "src/libruntime/utils/constants.h"
#include "src/libruntime/waiting_object_manager.h"
namespace YR {
namespace Libruntime {
using HandleReturnObjectCallback = std::function<std::pair<ErrorInfo, std::shared_ptr<Buffer>>(
    std::shared_ptr<Buffer> buffer, int rank, std::string &objId)>;
class Group : public std::enable_shared_from_this<Group> {
public:
    Group() = default;
    Group(const std::string &name) : groupName(name) {};
    ~Group() = default;
    ErrorInfo Wait();
    ErrorInfo GroupCreate();
    void Terminate();
    ErrorInfo Suspend();
    ErrorInfo Resume();
    void SetRunFlag();
    bool IsReady();
    std::string GetGroupName();
    void SetCreateSpecs(std::vector<std::shared_ptr<InvokeSpec>> &specs);
    InstanceRange GetInstanceRange();
    FunctionGroupOptions GetFunctionGroupOptions();
    virtual ErrorInfo Accelerate(const AccelerateMsgQueueHandle &handle, HandleReturnObjectCallback callback)
    {
        return ErrorInfo();
    }

protected:
    Group(const std::string &name, const std::string &inputTenantId, GroupOpts &inputOpts,
          std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
          std::shared_ptr<MemoryStore> memStore);
    Group(const std::string &name, const std::string &inputTenantId, InstanceRange &inputRange,
          std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
          std::shared_ptr<MemoryStore> memStore);
    Group(const std::string &name, const std::string &inputTenantId, FunctionGroupOptions &inputOpts,
          std::shared_ptr<FSClient> client, std::shared_ptr<WaitingObjectManager> waitManager,
          std::shared_ptr<MemoryStore> memStore);
    virtual CreateRequests BuildCreateReqs() = 0;
    virtual void CreateRespHandler(const CreateResponses &resps) = 0;
    virtual void CreateNotifyHandler(const NotifyRequest &req) = 0;
    virtual void SetTerminateError() = 0;
    std::atomic<bool> runFlag{true};
    std::atomic<bool> isSendReq{false};
    std::atomic<bool> isReady{false};
    std::string groupName = "";
    std::string traceId;
    std::string tenantId;
    std::string groupId;
    std::mutex groupIdMtx;
    GroupOpts opts;
    InstanceRange range;
    FunctionGroupOptions functionGroupOpts;
    std::shared_ptr<FSClient> fsClient;
    std::vector<std::shared_ptr<InvokeSpec>> createSpecs;
    std::shared_ptr<WaitingObjectManager> waitManager_;
    std::shared_ptr<MemoryStore> memStore_;

private:
    ErrorInfo Signal(libruntime::Signal signal);
};

}  // namespace Libruntime
}  // namespace YR