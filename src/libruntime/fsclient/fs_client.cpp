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

#include "src/libruntime/fsclient/fs_client.h"
#include "src/libruntime/fsclient/fs_intf_impl.h"
#include "src/libruntime/gwclient/gw_client.h"

namespace YR {
namespace Libruntime {

FSClient::FSClient(std::shared_ptr<FSIntf> fsIntface) : fsIntf(fsIntface) {}

ErrorInfo FSClient::Start(const std::string &ipAddr, int port, FSIntfHandlers handlers, ClientType type, bool isDriver,
                          std::shared_ptr<Security> security, std::shared_ptr<ClientsManager> clientsMgr,
                          const std::string &jobID, const std::string &instanceID, const std::string &runtimeID,
                          const std::string &functionName, const SubscribeFunc &reSubscribeCb, bool enableEvent)
{
    this->ipAddr = ipAddr;
    this->port = port;
    this->type = type;
    this->isDriver = isDriver;
    if (!fsIntf) {
        switch (type) {
            case ClientType::GRPC_SERVER:
            case ClientType::GRPC_CLIENT: {
                fsIntf = std::make_shared<FSIntfImpl>(ipAddr, port, handlers, isDriver, security, clientsMgr,
                                                      type == ClientType::GRPC_CLIENT, enableEvent);
                break;
            }
            case ClientType::GW_CLIENT: {
                break;
            }
            default: {
                return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "ClientType only support GRPC_SERVER, GRPC_CLIENT, GW_CLIENT");
            }
        }
    }
    auto err = fsIntf->Start(jobID, instanceID, runtimeID, functionName, reSubscribeCb);
    if (isDriver && err.OK()) {
        fsIntf->SetInitialized();
    }
    return err;
}

void FSClient::RemoveInsRtIntf(const std::string &instanceId)
{
    if (fsIntf) {
        fsIntf->RemoveInsRtIntf(instanceId);
    }
}

void FSClient::ReceiveRequestLoop(void)
{
    fsIntf->ReceiveRequestLoop();
}

void FSClient::Stop()
{
    fsIntf->Stop();
}

void FSClient::GroupCreateAsync(const CreateRequests &reqs, CreateRespsCallback respCallback, CreateCallBack callback,
                                int timeoutSec)
{
    this->fsIntf->GroupCreateAsync(reqs, respCallback, callback, timeoutSec);
}

void FSClient::CreateAsync(const CreateRequest &req, CreateRespCallback respCallback, CreateCallBack callback,
                           int timeoutSec)
{
    this->fsIntf->CreateAsync(req, respCallback, callback, timeoutSec);
}

void FSClient::InvokeAsync(const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback, int timeoutSec)
{
    this->fsIntf->InvokeAsync(req, callback, timeoutSec);
}

void FSClient::CallResultAsync(const std::shared_ptr<CallResultMessageSpec> req, CallResultCallBack callback)
{
    this->fsIntf->CallResultAsync(req, callback);
}

void FSClient::ReturnCallResult(const std::shared_ptr<CallResultMessageSpec> res, bool isCreate,
                                CallResultCallBack callback)
{
    this->fsIntf->ReturnCallResult(res, isCreate, callback);
}

void FSClient::KillAsync(const KillRequest &req, KillCallBack callback, int timeoutSec)
{
    this->fsIntf->KillAsync(req, callback, timeoutSec);
}
void FSClient::ExitAsync(const ExitRequest &req, ExitCallBack callback)
{
    this->fsIntf->ExitAsync(req, callback);
}

void FSClient::StateSaveAsync(const StateSaveRequest &req, StateSaveCallBack callback)
{
    this->fsIntf->StateSaveAsync(req, callback);
}

void FSClient::StateLoadAsync(const StateLoadRequest &req, StateLoadCallBack callback)
{
    this->fsIntf->StateLoadAsync(req, callback);
}

int FSClient::WaitRequestEmpty(uint64_t gracePeriodSec)
{
    return this->fsIntf->WaitRequestEmpty(gracePeriodSec);
}

std::string FSClient::GetServerVersion()
{
    return this->fsIntf->GetServerVersion();
}

std::pair<ErrorInfo, std::string> FSClient::GetNodeId()
{
    return this->fsIntf->GetNodeId();
}

std::pair<ErrorInfo, std::string>  FSClient::GetNodeIp()
{
    return this->fsIntf->GetNodeIp();
}

void FSClient::CreateRGroupAsync(const CreateResourceGroupRequest &req, CreateResourceGroupCallBack callback,
                                   int timeoutSec)
{
    return this->fsIntf->CreateRGroupAsync(req, callback, timeoutSec);
}

void FSClient::EraseIntf(const std::string &id)
{
    fsIntf->RemoveInsRtIntf(id);
}

bool FSClient::IsHealth()
{
    if (!fsIntf) {
        return false;
    }
    return fsIntf->IsHealth();
}

void FSClient::UpdateEventServerInfo(const std::string &ip, int port, const std::string &instanceId)
{
    this->fsIntf->UpdateEventServerInfo(ip, port, instanceId);
}

void FSClient::EventAsync(const std::shared_ptr<EventMessageSpec> &req, int timeoutSec)
{
    this->fsIntf->EventAsync(req, timeoutSec);
}

std::string FSClient::GetEventServerIP()
{
    return this->ipAddr;
}

int FSClient::GetEventServerPort()
{
    return this->fsIntf->GetSelfPort();
}

}  // namespace Libruntime
}  // namespace YR