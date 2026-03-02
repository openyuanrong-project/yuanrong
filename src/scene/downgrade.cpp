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

#include "downgrade.h"
namespace YR {
namespace scene {
using namespace YR::utility;
using namespace YR::Libruntime;
DowngradeController::DowngradeController(const std::string &functionId, std::shared_ptr<ClientsManager> clientsMgr,
                                         std::shared_ptr<Security> security)
    : functionId_(functionId), clientsMgr_(clientsMgr), security_(security)
{
    YRLOG_DEBUG("functionId_: {}", functionId);  // delete
    isFrontendFunction_ = IsFrontendFunction(functionId);
    existIngress_ =
        !YR::GetEnvValue(FRONTEND_ADDRESS_ENV).empty() && !YR::GetEnvValue(INVOCATION_URL_PREFIX_ENV).empty();
}

DowngradeController::~DowngradeController()
{
    Stop();
}

YR::Libruntime::ErrorInfo DowngradeController::Init()
{
    if (init_) {
        YR::Libruntime::ErrorInfo();
    }
    auto watchFileName = YR::GetEnvValue(DOWNGRADE_FILE_ENV);
    if (isFrontendFunction_ || watchFileName.empty()) {
        YRLOG_DEBUG("is frontend function: {}, or DOWNGRADE_FILE env empty: {}", isFrontendFunction_, watchFileName);
        return YR::Libruntime::ErrorInfo();
    }
    ParseDowngrade(watchFileName);
    watcher_ = std::make_shared<FileWatcher>(
        watchFileName, std::bind(&DowngradeController::ParseDowngrade, this, std::placeholders::_1));
    watcher_->Start();
    init_ = true;
    return YR::Libruntime::ErrorInfo();
}

YR::Libruntime::ErrorInfo DowngradeController::InitApiClient()
{
    if (apiClient_ != nullptr) {
        return YR::Libruntime::ErrorInfo();
    }
    apiClient_ = std::make_shared<ApiClient>(functionId_, clientsMgr_, security_);
    return apiClient_->Init();
}

// content: {"downgrade": true}
void DowngradeController::ParseDowngrade(const std::string &fileName)
{
    std::ifstream file(fileName);
    if (!file.is_open()) {
        YRLOG_DEBUG("{} not exit", fileName);
        return;
    }
    nlohmann::json j;
    try {
        file >> j;
        isDowngradeEnabled_ = j.value("downgrade", false);
        YRLOG_DEBUG("update downgrade enable {}", isDowngradeEnabled_);
    } catch (const nlohmann::json::parse_error &e) {
        YRLOG_WARN("json parse error: {}", e.what());
        isDowngradeEnabled_ = false;
    }
}

bool DowngradeController::ShouldDowngrade(const std::shared_ptr<InvokeSpec> spec) const
{
    if (!existIngress_ || isFrontendFunction_) {
        return false;
    }
    if (spec->downgradeFlag_ || isDowngradeEnabled_) {
        return true;
    }
    return false;
}

void DowngradeController::Downgrade(const std::shared_ptr<InvokeSpec> spec, const InvocationCallback &callback)
{
    if (spec == nullptr) {
        return;
    }
    std::call_once(flag_, [&, this]() { err_ = InitApiClient(); });
    if (!err_.OK()) {
        YRLOG_DEBUG("init api client err {}", err_.CodeAndMsg());
        callback(spec->requestId, err_.Code(), err_.Msg());
        return;
    }
    YRLOG_DEBUG("downgrade request id: {}", spec->requestId);
    apiClient_->InvocationAsync(spec, callback);
}

bool DowngradeController::ShouldFaultDowngrade() const
{
    return !isFrontendFunction_ && existIngress_;
}

bool DowngradeController::IsFrontendFunction(const std::string &function)
{
    if (function.size() < FAAS_FRONTEND_FUNCTION_NAME_PREFIX.size()) {
        return false;
    }
    return function.substr(0, FAAS_FRONTEND_FUNCTION_NAME_PREFIX.size()) == FAAS_FRONTEND_FUNCTION_NAME_PREFIX;
}

void DowngradeController::Stop()
{
    if (!init_) {
        return;
    }
    if (watcher_) {
        watcher_->Stop();
    }
    init_ = false;
}

ApiClient::ApiClient(const std::string &functionId, std::shared_ptr<ClientsManager> clientsMgr,
                     std::shared_ptr<Security> security)
    : functionId_(functionId), clientsMgr_(clientsMgr), security_(security)
{
}

YR::Libruntime::ErrorInfo ApiClient::Init()
{
    if (init_) {
        return YR::Libruntime::ErrorInfo();
    }
    businessId_ = YR::GetEnvValue(YR_BUSINESS_ID_ENV);
    auto address = YR::GetEnvValue(FRONTEND_ADDRESS_ENV);
    if (address.empty()) {
        YRLOG_ERROR("YR_FRONTEND_ADDRESS env unset");
        return Libruntime::ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, "YR_FRONTEND_ADDRESS env unset");
    }
    if (address.find(HTTP_PROTOCOL_PREFIX) == 0) {
        address = address.substr(HTTP_PROTOCOL_PREFIX.size());
        enableTLS_ = false;
    } else if (address.find(HTTPS_PROTOCOL_PREFIX) == 0) {
        address = address.substr(HTTPS_PROTOCOL_PREFIX.size());
        enableTLS_ = true;
    } else {
        enableTLS_ = false;
    }
    YR::ParseIpAddr(address, ip_, port_);
    std::string host;
    if (ip_.empty()) {
        port_ = enableTLS_ ? HTTPS_DEFAULT_PORT : HTTP_DEFAULT_PORT;
        host = address;
        ip_ = host;
        YRLOG_DEBUG("ip host {}, port: {}", host, port_);
    }
    urlPrefix_ = YR::GetEnvValue(INVOCATION_URL_PREFIX_ENV);
    if (urlPrefix_.empty()) {
        YRLOG_ERROR("YR_INVOCATION_URL_PREFIX env unset");
        return Libruntime::ErrorInfo(Libruntime::ErrorCode::ERR_PARAM_INVALID, "YR_INVOCATION_URL_PREFIX env unset");
    }
    auto config = std::make_shared<LibruntimeConfig>();
    const uint32_t threadNum = 10;
    config->httpIocThreadsNum = threadNum;
    config->maxConnSize = config->httpIocThreadsNum;
    config->enableTLS = enableTLS_;
    config->verifyFilePath = YR::GetEnvValue(CERTIFICATE_FILE_ENV);
    config->serverName = host;
    config->httpIdleTime = YR::Libruntime::Config::Instance().YR_HTTP_IDLE_TIME();
    auto [httpClient, err] = clientsMgr_->GetOrNewHttpClient(ip_, port_, config);
    if (!err.OK()) {
        YRLOG_ERROR("get or new http client failed, code is {}, msg is {}", fmt::underlying(err.Code()), err.Msg());
        return err;
    }
    FSIntfHandlers handlers;
    gwClient_ = std::make_shared<GwClient>(functionId_, handlers, security_);
    gwClient_->Init(httpClient);
    init_ = true;
    return YR::Libruntime::ErrorInfo();
}

void ApiClient::InvocationAsync(const std::shared_ptr<InvokeSpec> spec, const InvocationCallback &callback)
{
    // 8d86c63b22e24d9ab650878b75408ea6/test-jiuwen-session-004-bj/$latest
    auto functionId = spec->functionMeta.functionId;
    std::vector<std::string> result;
    YR::utility::Split(functionId, result, '/');
    size_t tenantIndex = 0;
    size_t functionNameIndex = 1;
    size_t functionVersionIndex = 2;
    if (result.size() <= functionVersionIndex) {
        callback(spec->requestId, ErrorCode::ERR_PARAM_INVALID, "function id invalid " + functionId);
        return;
    }
    // 8d86c63b22e24d9ab650878b75408ea6
    auto tenantId = result[tenantIndex];
    // test-jiuwen-session-004-bj
    auto functionName = result[functionNameIndex];
    std::vector<std::string> results;
    YR::utility::Split(functionName, results, '@');
    if (!results.empty()) {
        functionName = results.back();
    }
    // $latest
    auto functionVersion = result[functionVersionIndex];
    if (functionVersion == "latest") {
        functionVersion = "$latest";
    }
    // /serverless/v2/functions/wisefunction:cn:iot:8d86c63b22e24d9ab65:function:0@faas@cpp:latest/invocations
    std::string url = urlPrefix_ + URN_SEPARATOR + businessId_ + URN_SEPARATOR + tenantId +
                      ":function:" + functionName + URN_SEPARATOR + functionVersion + "/invocations";
    gwClient_->InvocationAsync(url, spec, callback);
}

}  // namespace scene
}  // namespace YR