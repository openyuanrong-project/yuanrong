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

#include <mutex>
#include <string>
#include <unordered_map>

#include "src/libruntime/clientsmanager/clients_manager.h"
#include "src/libruntime/gwclient/gw_client.h"
#include "src/libruntime/gwclient/http/async_http_client.h"
#include "src/libruntime/gwclient/http/client_manager.h"
#include "src/libruntime/libruntime_config.h"
#include "src/libruntime/utils/crypto/aes_crypto.h"
#include "src/libruntime/utils/http_utils.h"

namespace YR {
namespace Libruntime {
// 1、http://IP:PORT; 2、https://IP:PORT; 3、IP:PORT; 4、https://www.xxx.com; 5、http://www.xxx.com
const std::string HTTP_PROTOCOL_PREFIX = "http://";
const std::string HTTPS_PROTOCOL_PREFIX = "https://";
const int32_t HTTPS_DEFAULT_PORT = 443;
const int32_t HTTP_DEFAULT_PORT = 80;
const std::string X_TENANT_ID = "X-Tenant-ID";
const std::string IAM_CREDENTIAL_REQUIRE_PATH = "/iam-server/v1/credential/require";
struct AkSkInfo {
    std::string accessKey;
    SensitiveValue secretKey;
    uint64_t expiredTimeStamp{ 0 };

    std::string toString()
    {
        std::ostringstream oss;
        oss << "ak: " << accessKey << ", expiredTimeStamp: " << std::to_string(expiredTimeStamp);
        return oss.str();
    }
};

class IamClient {
public:
    IamClient(std::shared_ptr<ClientsManager> clientsMgr,
              std::shared_ptr<Security> security,
              std::shared_ptr<LibruntimeConfig> librtConfig);

    ~IamClient() = default;

    virtual YR::Libruntime::ErrorInfo Init();

    virtual YR::Libruntime::ErrorInfo DecryptKey(SensitiveValue &key);

    virtual std::shared_ptr<AkSkInfo> GetTenantAkSkInfo(const std::string &tenantId);

private:
    ErrorInfo ValidateAndParseAddress(const std::string& address);
    
    std::shared_ptr<LibruntimeConfig> BuildLibConfig();

    ErrorInfo CreateHttpClient();

    bool init_ = false;
    std::shared_ptr<YR::Libruntime::GwClient> gwClient_;
    std::shared_ptr<YR::Libruntime::ClientsManager> clientsMgr_;
    std::shared_ptr<YR::Libruntime::Security> security_;
    std::shared_ptr<YR::Libruntime::LibruntimeConfig> librtConfig_;
    std::string host_;
    std::string ip_;
    int32_t port_{};
    bool enableTLS_{false};
    std::string urlPrefix_;
    std::mutex gwClientMutex_; // 用于保护 gwClient_的互斥锁
};

class TenantAKSKManager {
public:
    // 获取单例实例
    static TenantAKSKManager &GetInstance()
    {
        static TenantAKSKManager instance;
        return instance;
    }

    // 删除拷贝构造函数和赋值运算符
    TenantAKSKManager(const TenantAKSKManager &) = delete;
    TenantAKSKManager &operator=(const TenantAKSKManager &) = delete;

    void Initialize(std::shared_ptr<ClientsManager> clientsMgr,
                    std::shared_ptr<Security> security,
                    std::shared_ptr<LibruntimeConfig> librtConfig);

    // 从缓存中获取指定 key 的 value
    std::shared_ptr<AkSkInfo> GetValue(const std::string& tenantId);

private:
    TenantAKSKManager() = default;
    ~TenantAKSKManager() = default;
    bool initFlag_{false};
    std::shared_ptr<IamClient> iamClient_;
    std::unordered_map<std::string, std::shared_ptr<AkSkInfo>> cacheMap;
    std::mutex mtx;
};
}  // namespace Libruntime
}  // namespace YR