/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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
#include "src/libruntime/clientsmanager/clients_manager.h"
#include "src/libruntime/gwclient/gw_client.h"
#include "src/libruntime/gwclient/http/async_http_client.h"
#include "src/libruntime/gwclient/http/client_manager.h"
#include "src/libruntime/invoke_spec.h"
#include "src/utility/file_watcher.h"
namespace YR {
namespace scene {
const std::string FAAS_FRONTEND_FUNCTION_NAME_PREFIX = "0/0-system-faasfrontend/";
const std::string DOWNGRADE_FILE_ENV = "YR_DOWNGRADE_FILE";
// 1、http://IP:PORT; 2、https://IP:PORT; 3、IP:PORT; 4、https://www.xxx.com; 5、http://www.xxx.com
const std::string FRONTEND_ADDRESS_ENV = "YR_FRONTEND_ADDRESS";
// "/serverless/v2/functions/wisefunction:cn"
const std::string INVOCATION_URL_PREFIX_ENV = "YR_INVOCATION_URL_PREFIX";
const std::string URN_SEPARATOR = ":";
const std::string HTTP_PROTOCOL_PREFIX = "http://";
const std::string HTTPS_PROTOCOL_PREFIX = "https://";
const std::string CERTIFICATE_FILE_ENV = "YR_CERTIFICATE_FILE";
const std::string YR_BUSINESS_ID_ENV = "YR_BUSINESS_ID";
const int32_t HTTPS_DEFAULT_PORT = 443;
const int32_t HTTP_DEFAULT_PORT = 80;
class ApiClient {
public:
    ApiClient(const std::string &functionId, std::shared_ptr<YR::Libruntime::ClientsManager> clientsMgr,
              std::shared_ptr<YR::Libruntime::Security> security);

    ~ApiClient() = default;

    YR::Libruntime::ErrorInfo Init();

    void InvocationAsync(const std::shared_ptr<YR::Libruntime::InvokeSpec> spec,
                         const YR::Libruntime::InvocationCallback &callback);

private:
    bool init_ = false;
    std::shared_ptr<YR::Libruntime::GwClient> gwClient_;
    std::string functionId_;
    std::shared_ptr<YR::Libruntime::ClientsManager> clientsMgr_;
    std::string ip_;
    int32_t port_;
    bool enableTLS_{false};
    std::string urlPrefix_;
    std::shared_ptr<YR::Libruntime::Security> security_;
    std::string businessId_;
};

class DowngradeController {
public:
    explicit DowngradeController(const std::string &functionId,
                                 std::shared_ptr<YR::Libruntime::ClientsManager> clientsMgr,
                                 std::shared_ptr<YR::Libruntime::Security> security);

    ~DowngradeController();

    YR::Libruntime::ErrorInfo Init();

    YR::Libruntime::ErrorInfo InitApiClient();

    // content: {"downgrade": true}
    void ParseDowngrade(const std::string &fileName);

    bool ShouldDowngrade(const std::shared_ptr<YR::Libruntime::InvokeSpec> spec) const;

    void Downgrade(const std::shared_ptr<YR::Libruntime::InvokeSpec> spec,
                   const YR::Libruntime::InvocationCallback &callback);

    bool ShouldFaultDowngrade() const;

    static bool IsFrontendFunction(const std::string &function);

    void Stop();

private:
    bool init_{false};
    std::shared_ptr<YR::utility::FileWatcher> watcher_;
    bool isFrontendFunction_{false};
    bool existIngress_{false};
    bool isDowngradeEnabled_{false};
    std::string functionId_;
    std::shared_ptr<YR::Libruntime::ClientsManager> clientsMgr_;
    std::string ip_;
    std::string port_;
    std::shared_ptr<ApiClient> apiClient_;
    std::once_flag flag_;
    std::shared_ptr<YR::Libruntime::Security> security_;
    YR::Libruntime::ErrorInfo err_;
};

}  // namespace scene
}  // namespace YR
