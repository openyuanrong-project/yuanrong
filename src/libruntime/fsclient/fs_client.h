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

#include <memory>

#include "fs_intf.h"
#include "src/libruntime/clientsmanager/clients_manager.h"
#include "src/libruntime/utils/security.h"

namespace YR {
namespace Libruntime {
class FSClient {
public:
    enum class ClientType { GRPC_SERVER, GRPC_CLIENT, LITEBUS, URPC, GW_CLIENT };
    FSClient() = default;
    explicit FSClient(std::shared_ptr<FSIntf> fsIntface);
    ErrorInfo Start(const std::string &ipAddr, int port, FSIntfHandlers handlers, ClientType type, bool isDriver,
                    std::shared_ptr<Security> security, std::shared_ptr<ClientsManager> clientsMgr,
                    const std::string &jobID, const std::string &instanceID, const std::string &runtimeID,
                    const std::string &functionName, const SubscribeFunc &reSubscribeCb = nullptr,
                    bool enableEvent = false);
    void Stop();
    void ReceiveRequestLoop(void);
    void GroupCreateAsync(const CreateRequests &reqs, CreateRespsCallback respCallback, CreateCallBack callback,
                          int timeoutSec = -1);
    void CreateAsync(const CreateRequest &req, CreateRespCallback respCallback, CreateCallBack callback,
                     int timeoutSec = -1);
    void InvokeAsync(const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback, int timeoutSec = -1);
    void CallResultAsync(const std::shared_ptr<CallResultMessageSpec> req, CallResultCallBack callback);
    void ReturnCallResult(const std::shared_ptr<CallResultMessageSpec> result, bool isCreate,
                          CallResultCallBack callback);
    void KillAsync(const KillRequest &req, KillCallBack callback, int timeoutSec = -1);
    void ExitAsync(const ExitRequest &req, ExitCallBack callback);
    void StateSaveAsync(const StateSaveRequest &req, StateSaveCallBack callback);
    void StateLoadAsync(const StateLoadRequest &req, StateLoadCallBack callback);
    int WaitRequestEmpty(uint64_t gracePeriodSec);
    std::string GetServerVersion();
    std::pair<ErrorInfo, std::string> GetNodeId();
    std::pair<ErrorInfo, std::string> GetNodeIp();
    void RemoveInsRtIntf(const std::string &instanceId);
    void CreateRGroupAsync(const CreateResourceGroupRequest &req, CreateResourceGroupCallBack callback,
                             int timeoutSec = -1);
    void EraseIntf(const std::string &id);
    bool IsHealth();
    void UpdateEventServerInfo(const std::string &ip, int port, const std::string &instaceId);
    void EventAsync(const std::shared_ptr<EventMessageSpec> &req, int timeoutSec = -1);
    std::string GetEventServerIP();
    int GetEventServerPort();

private:
    std::shared_ptr<FSIntf> fsIntf;
    std::string ipAddr;
    int port{};
    int eventServerPort{};
    ClientType type;
    bool isDriver{};
};
}  // namespace Libruntime
}  // namespace YR