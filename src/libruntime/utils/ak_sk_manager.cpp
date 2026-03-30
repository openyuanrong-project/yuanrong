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

#include "ak_sk_manager.h"

#include <json.hpp>

#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
void TenantAKSKManager::Initialize(std::shared_ptr<ClientsManager> clientsMgr,
                                   std::shared_ptr<Security> security,
                                   std::shared_ptr<LibruntimeConfig> librtConfig)
{
    std::lock_guard<std::mutex> lock(mtx);
    if (initFlag_) {
        YRLOG_DEBUG("tenantAKSKManager has initizalize");
        return;
    }
    iamClient_ = std::make_shared<IamClient>(clientsMgr, security, librtConfig);
    initFlag_ = true;
    YRLOG_DEBUG("tenantAKSKManager Initialize finish");
}

std::shared_ptr<AkSkInfo> TenantAKSKManager::GetValue(const std::string &tenantId)
{
    if (tenantId.empty()) {
        YRLOG_ERROR("get ak/sk failed, tenantId is empty");
        return {};
    }

    // 首次尝试从缓存中获取
    {
        std::lock_guard<std::mutex> lock(mtx);
        auto it = cacheMap.find(tenantId);
        if (it != cacheMap.end()) {
            // 检查是否过期
            auto akskInfo = it->second;
            auto now = static_cast<uint64_t>(std::time(nullptr));
            if (akskInfo->expiredTimeStamp > 0 && akskInfo->expiredTimeStamp < now) {
                YRLOG_DEBUG("the ak/sk in cache has expired, tenantId: {}", tenantId);
            } else {
                YRLOG_DEBUG("get ak/sk from cache success, tenantId: {}, akskInfo: {}", tenantId, akskInfo->toString());
                return akskInfo;
            }
        }
        YRLOG_DEBUG("there is no ak/sk in cache, tenantId: {}", tenantId);
    }

    // 从Iam获取新的 AK/SK
    auto newAkSk = iamClient_->GetTenantAkSkInfo(tenantId);
    if (!newAkSk || newAkSk->accessKey.empty() || newAkSk->secretKey.Empty()) {
        YRLOG_ERROR("get ak/sk from iam failed, tenantId: {}", tenantId);
        return {};
    }

    // 更新缓存
    {
        std::lock_guard<std::mutex> lock(mtx);
        cacheMap[tenantId] = newAkSk;
    }
    return newAkSk;
}

IamClient::IamClient(std::shared_ptr<ClientsManager> clientsMgr,
                     std::shared_ptr<Security> security,
                     std::shared_ptr<LibruntimeConfig> librtConfig)
    : clientsMgr_(std::move(clientsMgr)),
      security_(std::move(security)),
      librtConfig_(std::move(librtConfig))
{
}

YR::Libruntime::ErrorInfo IamClient::Init()
{
    std::lock_guard<std::mutex> lock(gwClientMutex_);
    if (init_) {
        YRLOG_DEBUG("iamClient has init");
        return {};
    }
    auto address = librtConfig_->iamAddress;
    YRLOG_DEBUG("request to iamAddress: {}", address);
    if (address.empty()) {
        YRLOG_ERROR("iamAddress is empty");
        return YR::Libruntime::ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, "iamAddress is empty");
    }
    if (auto err = ValidateAndParseAddress(address); !err.OK()) {
        return err;
    }
    if (auto err = CreateHttpClient(); !err.OK()) {
        return err;
    }
    init_ = true;
    YRLOG_DEBUG("iamClient init finish");
    return {};
}

YR::Libruntime::ErrorInfo IamClient::ValidateAndParseAddress(const std::string& address)
{
    std::string addr = address;
    if (addr.find(HTTP_PROTOCOL_PREFIX) == 0) {
        addr = addr.substr(HTTP_PROTOCOL_PREFIX.size());
        enableTLS_ = false;
    } else if (addr.find(HTTPS_PROTOCOL_PREFIX) == 0) {
        addr = addr.substr(HTTPS_PROTOCOL_PREFIX.size());
        enableTLS_ = true;
    } else {
        enableTLS_ = false;
    }

    YR::ParseIpAddr(addr, ip_, port_);
    if (ip_.empty()) {
        port_ = enableTLS_ ? HTTPS_DEFAULT_PORT : HTTP_DEFAULT_PORT;
        host_ = addr;
        ip_ = host_;
        YRLOG_DEBUG("ip host: {}, port: {}", host_, port_);
    }
    return {};
}

std::shared_ptr<LibruntimeConfig> IamClient::BuildLibConfig()
{
    auto config = std::make_shared<LibruntimeConfig>();
    const uint32_t threadNum = 10;
    config->httpIocThreadsNum = threadNum;
    config->maxConnSize = config->httpIocThreadsNum;
    config->enableTLS = enableTLS_;
    config->verifyFilePath = librtConfig_->verifyFilePath;
    config->serverName = host_;
    config->httpIdleTime = YR::Libruntime::Config::Instance().YR_HTTP_IDLE_TIME();
    return config;
}

YR::Libruntime::ErrorInfo IamClient::CreateHttpClient()
{
    auto config = BuildLibConfig();
    auto [httpClient, err] = clientsMgr_->GetOrNewHttpClient(ip_, port_, config);
    if (!err.OK()) {
        YRLOG_ERROR("get or new http client failed, code is {}, msg is {}",
            fmt::underlying(err.Code()), err.Msg());
        return YR::Libruntime::ErrorInfo(Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
            "get or new http client failed");
    }
    FSIntfHandlers handlers;
    std::string functionId_;
    gwClient_ = std::make_shared<GwClient>(functionId_, handlers, security_);
    const int32_t connectTimeout = 3;
    gwClient_->Init(httpClient, connectTimeout);
    return {};
}

YR::Libruntime::ErrorInfo IamClient::DecryptKey(SensitiveValue &key)
{
    std::string sysAK;
    SensitiveValue sysSK;
    SensitiveValue sysDK;
    security_->GetAKSKDK(sysAK, sysSK, sysDK);
    if (sysDK.Empty()) {
        YRLOG_ERROR("sysDK is empty");
        return YR::Libruntime::ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, "sysDK is empty");
    }
    auto err = DecryptWithGCMSpecifyKey(key, sysDK);
    if (!err.OK()) {
        YRLOG_ERROR("decrypt gcm with sysDK failed, err: {}", err.Msg());
        return err;
    }
    YRLOG_DEBUG("decrypt tenant sk finish");
    return {};
}

std::shared_ptr<AkSkInfo> IamClient::GetTenantAkSkInfo(const std::string &tenantId)
{
    std::string method = "GET";
    std::unordered_map<std::string, std::string> headers;
    std::string body;
    std::string path = IAM_CREDENTIAL_REQUIRE_PATH;
    SignHttpRequest2(security_, method, path, headers, body);
    auto it = headers.find(X_SIGNATURE);
    if (it == headers.end() || it->second.empty()) {
        YRLOG_ERROR("get ak/sk fron iam failed, sign failed");
        return {};
    }
    headers[X_TENANT_ID] = tenantId;
    if (gwClient_ == nullptr) {
        this->Init();
        if (gwClient_ == nullptr) {
            YRLOG_ERROR("gwClient_ is nullptr, get ak/sk fron iam failed");
            return {};
        }
    }
    auto promise = std::make_shared<std::promise<std::shared_ptr<AkSkInfo>>>();
    auto future = promise->get_future();
    gwClient_->InvocationSync(
        method, path, headers, body,
        [this, promise, tenantId](const std::string &requestId, Libruntime::ErrorCode code, const std::string &result) {
            auto akSkInfo = std::make_shared<AkSkInfo>();
            if (code != ErrorCode::ERR_OK) {
                YRLOG_ERROR("get ak/sk from iam failed, code: {}, err: {}", fmt::underlying(code), result);
                promise->set_value(akSkInfo);
                return;
            }
            YRLOG_DEBUG("get ak/sk fron iam success");
            nlohmann::json resJson = nlohmann::json::parse(result);
            akSkInfo->accessKey = resJson["accessKey"].get<std::string>();
            akSkInfo->secretKey = resJson["secretKey"].get<std::string>();
            auto err = this->DecryptKey(akSkInfo->secretKey);
            if (!err.OK()) {
                YRLOG_ERROR("decrypt tenant secret key failed, err: {}", err.Msg());
                promise->set_value(std::shared_ptr<AkSkInfo>());
                return;
            }
            akSkInfo->expiredTimeStamp = std::stoull(resJson["expiredTimeStamp"].get<std::string>());
            YRLOG_DEBUG("get ak/sk fron iam, akSkInfo: {}, tenantId: {}", akSkInfo->toString(), tenantId);
            promise->set_value(akSkInfo);
    });
    return future.get();
}
}  // namespace Libruntime
}  // namespace YR
