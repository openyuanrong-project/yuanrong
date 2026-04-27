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
#include <thread>
#include <utility>

#include "src/dto/resource_unit.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/fsclient/protobuf/message.pb.h"
#include "src/libruntime/fsclient/protobuf/resource.pb.h"
#include "src/libruntime/gwclient/http/async_http_client.h"
#include "src/libruntime/gwclient/http/async_https_client.h"
#include "src/libruntime/libruntime_config.h"

namespace YR {
namespace Libruntime {
using QueryResourcesInfoRequest = ::messages::QueryResourcesInfoRequest;
using QueryResourcesInfoResponse = ::messages::QueryResourcesInfoResponse;
using ResourceType = ::resources::Value_Type;
using QueryNamedInsRequest = ::messages::QueryNamedInsRequest;
using QueryNamedInsResponse = ::messages::QueryNamedInsResponse;
using QueryResourceGroupRequest = ::messages::QueryResourceGroupRequest;
using QueryResourceGroupResponse = ::messages::QueryResourceGroupResponse;
using ResourceGroupInfo = ::messages::ResourceGroupInfo;
using SubscribeActiveMasterCb = std::function<void()>;
ErrorInfo CheckResponseCode(const boost::beast::error_code &errorCode, const uint statusCode, const std::string &result,
                            const std::string &requestId);
const std::string GLOBAL_SCHEDULER_QUERY_RESOURCES = "/global-scheduler/resources";
const std::string INSTANCE_MANAGER_QUERY_NAMED_INSTANCES = "/instance-manager/named-ins";
const std::string GLOBAL_QUERY_RESOURCE_GROUP_TABLE = "/resource-group/rgroup";

class FMClient {
public:
    FMClient() : ioc_(std::make_shared<boost::asio::io_context>())
    {
        work_ = std::make_unique<boost::asio::executor_work_guard<boost::asio::io_context::executor_type>>(
            boost::asio::make_work_guard(*ioc_));
    }
    FMClient(const std::shared_ptr<LibruntimeConfig> config)
        : libConfig_(config), ioc_(std::make_shared<boost::asio::io_context>())
    {
        work_ = std::make_unique<boost::asio::executor_work_guard<boost::asio::io_context::executor_type>>(
            boost::asio::make_work_guard(*ioc_));
        enableMTLS_ = config->enableMTLS;
    }
    ~FMClient()
    {
        Stop();
    }
    void Stop();
    ErrorInfo ActivateMasterClientIfNeed();
    std::pair<ErrorInfo, ResourceGroupUnit> GetResourceGroupTable(const std::string &resourceGroupId);
    std::pair<ErrorInfo, std::vector<ResourceUnit>> GetResources();
    std::pair<ErrorInfo, std::vector<ResourceUnit>> GetResourcesWithRetry();
    std::pair<ErrorInfo, ResourceGroupUnit> GetResourcesGroupWithRetry(const std::string &resourceGroupId);
    std::pair<ErrorInfo, QueryNamedInsResponse> QueryNamedInstancesWithRetry();
    std::pair<ErrorInfo, QueryNamedInsResponse> QueryNamedInstances();
    void SetSubscribeActiveMasterCb(SubscribeActiveMasterCb cb);
    void UpdateActiveMaster(const std::string activeMasterAddr);
    void CleanActiveMaster();
    void RemoveActiveMaster();
    std::string GetActiveMasterAddr();

private:
    std::shared_ptr<HttpClient> InitCtxAndHttpClient(void);
    void MockResp();

    std::shared_ptr<LibruntimeConfig> libConfig_;
    std::map<std::string, std::shared_ptr<HttpClient>> httpClients_;
    std::string currentMaster_;
    std::shared_ptr<asio::io_context> ioc_;
    std::unique_ptr<std::thread> iocThread_;
    std::unique_ptr<boost::asio::executor_work_guard<boost::asio::io_context::executor_type>> work_;
    bool enableMTLS_{false};
    std::mutex activeMasterMu_;
    std::condition_variable condVar_;
    std::string activeMasterAddr_;
    std::shared_ptr<HttpClient> activeMasterHttpClient_;
    SubscribeActiveMasterCb cb_;
    std::atomic<bool> stopped{false};

    int maxWaitTimeSec = 90;
    int interval = 1000;
};

}  // namespace Libruntime
}  // namespace YR
