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

#include "fm_client.h"

#include <boost/beast/ssl.hpp>
#include <string>
#include <unordered_map>
#include "json.hpp"
#include "src/utility/id_generator.h"
#include "src/utility/notification_utility.h"
#include "src/utility/string_utility.h"

namespace ssl = boost::asio::ssl;
namespace YR {
namespace Libruntime {
const int HTTP_REQUEST_TIMEOUT = 5;
const int IP_ADDR_SIZE = 2;

QueryResourcesInfoRequest BuildGetResourcesReq(const std::string &reqId)
{
    QueryResourcesInfoRequest req;
    req.set_requestid(reqId);
    return req;
}

std::unordered_map<std::string, std::string> BuildGetResourcesHeaders()
{
    std::unordered_map<std::string, std::string> headers;
    headers[std::string("Type")] = std::string("protobuf");
    return headers;
}

ErrorInfo CheckResponseCode(const boost::beast::error_code &errorCode, const uint statusCode, const std::string &result,
                            const std::string &requestId)
{
    std::stringstream ss;
    if (errorCode) {
        ss << "network error between runtime and function master, error_code: " << errorCode.message()
           << ", requestId: " << requestId;
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_COMMUNICATION, ss.str());
    }
    if (!IsResponseSuccessful(statusCode)) {
        ss << "response is error, status_code: " << std::to_string(statusCode) << ", result: " << result
           << ", requestId: " << requestId;
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, ss.str());
    }
    return ErrorInfo();
}

std::unordered_map<std::string, float> ProcessResources(
    const ::google::protobuf::Map<std::string, resources::Resource> &resources)
{
    std::unordered_map<std::string, float> result;
    for (auto &resource : resources) {
        if (resource.second.type() == ResourceType::Value_Type_SCALAR) {
            result[resource.first] = resource.second.scalar().value();
        } else if (resource.second.type() == ResourceType::Value_Type_VECTORS) {
            auto values = resource.second.vectors().values();
            auto value = 0;
            if (values.find("ids") != values.end()) {
                for (const auto &pair : values["ids"].vectors()) {
                    value = pair.second.values_size();
                    break;
                }
            }
            result[resource.first] = value;
        } else {
            YRLOG_DEBUG("unknow type {}: of {}", fmt::underlying(resource.second.type()), resource.first);
            continue;
        }
    }
    return result;
}

ErrorInfo ParseQueryResponseToRgUnit(const std::string &result, std::shared_ptr<ResourceGroupUnit> rgUnit)
{
    QueryResourceGroupResponse resp;
    auto success = resp.ParseFromString(result);
    YRLOG_DEBUG("query resource group resp is {}", resp.DebugString());
    if (!success) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "failed to parse resource group unit info");
    }
    for (const auto &rGroup : resp.rgroup()) {
        RgInfo rgInfo;
        rgInfo.name = rGroup.name();
        rgInfo.owner = rGroup.owner();
        rgInfo.appId = rGroup.appid();
        rgInfo.tenantId = rGroup.tenantid();
        rgInfo.parentId = rGroup.parentid();
        rgInfo.requestId = rGroup.requestid();
        rgInfo.traceId = rGroup.traceid();
        rgInfo.status.code = rGroup.status().code();
        rgInfo.status.message = rGroup.status().message();

        Option opt;
        opt.priority = rGroup.opt().priority();
        opt.groupPolicy = static_cast<int>(rGroup.opt().grouppolicy());
        for (const auto &pair : rGroup.opt().extension()) {
            opt.extension[pair.first] = pair.second;
        }
        rgInfo.opt = opt;

        for (const auto &rgBundle : rGroup.bundles()) {
            BundleInfo bdInfo;
            bdInfo.bundleId = rgBundle.bundleid();
            bdInfo.rGroupName = rgBundle.rgroupname();
            bdInfo.parentRGroupName = rgBundle.parentrgroupname();
            bdInfo.functionProxyId = rgBundle.functionproxyid();
            bdInfo.functionAgentId = rgBundle.functionagentid();
            bdInfo.tenantId = rgBundle.tenantid();
            bdInfo.parentId = rgBundle.parentid();
            bdInfo.status.code = rgBundle.status().code();
            bdInfo.status.message = rgBundle.status().message();

            for (const auto &label : rgBundle.labels()) {
                bdInfo.labels.push_back(label);
            }

            for (const auto &pair : rgBundle.kvlabels()) {
                bdInfo.kvLabels[pair.first] = pair.second;
            }

            for (const auto &[key, value] : rgBundle.resources().resources()) {
                Resource resource;
                resource.name = value.name();
                resource.type = Resource::Type::SCALER;
                resource.scalar.value = value.scalar().value();
                resource.scalar.limit = value.scalar().limit();
                bdInfo.resources.resources[key] = std::move(resource);
            }
            rgInfo.bundles.push_back(std::move(bdInfo));
        }
        rgUnit->resourceGroups[rgInfo.name] = std::move(rgInfo);
    }
    return ErrorInfo();
}

std::unordered_map<std::string, std::vector<std::string>> ProcessNodeLabels(
    const ::google::protobuf::Map<std::string, ::resources::Value::Counter> &nodeLabels)
{
    std::unordered_map<std::string, std::vector<std::string>> result;
    for (auto &counter : nodeLabels) {
        for (auto &labels : counter.second.items()) {
            result[counter.first].push_back(labels.first);
        }
    }
    return result;
}

ErrorInfo ParseQueryResponse(const std::string &result, std::vector<ResourceUnit> &res)
{
    QueryResourcesInfoResponse resp;
    auto success = resp.ParseFromString(result);
    if (!success) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "failed to parse response");
    }
    for (auto &resPair : resp.resource().fragment()) {
        auto unit = ResourceUnit();
        unit.id = resPair.second.id();
        unit.status = resPair.second.status();
        unit.capacity = ProcessResources(resPair.second.capacity().resources());
        unit.allocatable = ProcessResources(resPair.second.allocatable().resources());
        unit.nodeLabels = ProcessNodeLabels(resPair.second.nodelabels());
        res.push_back(unit);
    }
    return ErrorInfo();
}

std::pair<ErrorInfo, ResourceGroupUnit> GetResourceGroupTableByHttpClient(std::shared_ptr<HttpClient> c,
                                                                          const std::string &resourceGroupId)
{
    auto reqId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    QueryResourceGroupRequest req;
    req.set_requestid(*reqId);
    req.set_rgroupname(resourceGroupId);
    std::unordered_map<std::string, std::string> headers;
    headers[std::string("Type")] = std::string("protobuf");
    std::string body;
    req.SerializeToString(&body);
    auto isExit = std::make_shared<bool>(false);
    auto asyncNotify = std::make_shared<YR::utility::NotificationUtility>();
    auto res = std::make_shared<ResourceGroupUnit>();
    c->SubmitInvokeRequest(
        POST, GLOBAL_QUERY_RESOURCE_GROUP_TABLE, headers, body, reqId,
        [res, reqId, asyncNotify, isExit](const std::string &result, const boost::beast::error_code &errorCode,
                                          const uint statusCode) {
            if (*isExit) {
                return;
            }
            auto err = CheckResponseCode(errorCode, statusCode, result, *reqId);
            if (err.OK()) {
                err = ParseQueryResponseToRgUnit(result, res);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "get request timeout: " << HTTP_REQUEST_TIMEOUT << ", requestId: " << *reqId;
    auto notifyErr = asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(HTTP_REQUEST_TIMEOUT), ErrorInfo(ErrorCode::ERR_FUNCTION_MASTER_TIMEOUT, ss.str()));
    *isExit = true;
    return std::make_pair(notifyErr, *res);
}

ErrorInfo ParseQueryNamedInstancesResponse(const std::string &result, std::shared_ptr<QueryNamedInsResponse> resp)
{
    if (!resp->ParseFromString(result)) {
        YRLOG_WARN("Failed to parse QueryNamedInstances response: {}", result);
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "failed to parse QueryNamedInstances response");
    }
    return ErrorInfo();
}

std::pair<ErrorInfo, QueryNamedInsResponse> GetNamedInstancesByHttpClient(std::shared_ptr<HttpClient> c)
{
    auto reqId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    QueryNamedInsRequest req;
    req.set_requestid(*reqId);
    std::string body;
    req.SerializeToString(&body);
    auto isExit = std::make_shared<bool>(false);
    std::unordered_map<std::string, std::string> headers = {{"Content-Type", "application/protobuf"}};

    auto resp = std::make_shared<QueryNamedInsResponse>();
    auto asyncNotify = std::make_shared<YR::utility::NotificationUtility>();

    c->SubmitInvokeRequest(
        GET, INSTANCE_MANAGER_QUERY_NAMED_INSTANCES, headers, body, reqId,
        [resp, asyncNotify, reqId, isExit](const std::string &result, const boost::beast::error_code &errorCode,
                                           const uint statusCode) {
            if (*isExit) {
                return;
            }
            auto err = CheckResponseCode(errorCode, statusCode, result, *reqId);
            if (err.OK()) {
                err = ParseQueryNamedInstancesResponse(result, resp);
            }
            asyncNotify->Notify(err);
        });

    std::stringstream ss;
    ss << "get named instances request timeout: " << HTTP_REQUEST_TIMEOUT << ", requestId: " << *reqId;
    auto notifyErr = asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(HTTP_REQUEST_TIMEOUT), ErrorInfo(ErrorCode::ERR_FUNCTION_MASTER_TIMEOUT, ss.str()));
    *isExit = true;
    return {notifyErr, *resp};
}

std::pair<ErrorInfo, std::vector<ResourceUnit>> GetResourcesByHttpClient(std::shared_ptr<HttpClient> c)
{
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto req = BuildGetResourcesReq(*requestId);
    auto headers = BuildGetResourcesHeaders();
    auto isExit = std::make_shared<bool>(false);
    std::string body;
    req.SerializeToString(&body);
    std::vector<ResourceUnit> res;
    auto asyncNotify = std::make_shared<YR::utility::NotificationUtility>();
    YRLOG_DEBUG("start to get resources by http client, request id: {}.", *requestId);
    c->SubmitInvokeRequest(
        GET, GLOBAL_SCHEDULER_QUERY_RESOURCES, headers, body, requestId,
        [&res, requestId, asyncNotify, isExit](const std::string &result, const boost::beast::error_code &errorCode,
                                               const uint statusCode) {
            if (*isExit) {
                return;
            }
            auto err = CheckResponseCode(errorCode, statusCode, result, *requestId);
            if (err.OK()) {
                err = ParseQueryResponse(result, res);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "get request timeout: " << HTTP_REQUEST_TIMEOUT << ", requestId: " << *requestId;
    auto notifyErr = asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(HTTP_REQUEST_TIMEOUT), ErrorInfo(ErrorCode::ERR_FUNCTION_MASTER_TIMEOUT, ss.str()));
    *isExit = true;
    return std::make_pair(notifyErr, res);
}

std::pair<ErrorInfo, ResourceGroupUnit> FMClient::GetResourceGroupTable(const std::string &resourceGroupId)
{
    YRLOG_DEBUG("start to get resource group table.");
    auto beginTime = std::chrono::steady_clock::now();
    for (int i = 0;
         std::chrono::duration_cast<std::chrono::seconds>(std::chrono::steady_clock::now() - beginTime).count() <
         maxWaitTimeSec;
         ++i) {
        auto [errInfo, res] = GetResourcesGroupWithRetry(resourceGroupId);
        if (errInfo.OK()) {
            return std::make_pair(errInfo, res);
        }
        YRLOG_WARN("retry get resources group table, current times: {}, err: {}", i, errInfo.Msg());
        if (cb_ != nullptr) {
            cb_();
        }
        std::this_thread::sleep_for(std::chrono::milliseconds(interval));
    }
    return std::make_pair(ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION,
                                    "failed to get resources group table, err: connect to function master timeout"),
                          ResourceGroupUnit{});
}

void FMClient::Stop()
{
    if (stopped) {
        return;
    }
    work_->reset();
    ioc_->stop();
    if (iocThread_) {
        if (iocThread_->joinable()) {
            iocThread_->join();
        }
    }
    stopped.store(true);
}

ErrorInfo FMClient::ActivateMasterClientIfNeed()
{
    const int maxWaitTime = 30;
    std::unique_lock<std::mutex> activeMasterLock(activeMasterMu_);
    if (!condVar_.wait_until(activeMasterLock, std::chrono::steady_clock::now() + std::chrono::seconds(maxWaitTime),
                             [this] { return !activeMasterAddr_.empty(); })) {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "failed to get valid active function master http client.");
    }
    if (!iocThread_) {
        iocThread_ = std::make_unique<std::thread>([&] { ioc_->run(); });
    }
    if (activeMasterHttpClient_ == nullptr || !activeMasterHttpClient_->Available() ||
        !activeMasterHttpClient_->IsConnActive()) {
        activeMasterHttpClient_ = InitCtxAndHttpClient();
        std::vector<std::string> result;
        YR::utility::Split(activeMasterAddr_, result, ':');
        if (result.size() != IP_ADDR_SIZE) {
            YRLOG_ERROR("invalid ip addr {}", activeMasterAddr_);
            CleanActiveMaster();
            return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "invalid ip addr");
        }
        ConnectionParam param;
        param.ip = result[0];
        param.port = result[1];
        auto initErr = activeMasterHttpClient_->Init(param);
        if (!initErr.OK()) {
            CleanActiveMaster();
            return initErr;
        } else {
            activeMasterHttpClient_->SetAvailable();
        }
    }
    return ErrorInfo();
}

std::pair<ErrorInfo, QueryNamedInsResponse> FMClient::QueryNamedInstancesWithRetry()
{
    if (auto errInfo = ActivateMasterClientIfNeed(); !errInfo.OK())
        return std::make_pair(errInfo, QueryNamedInsResponse{});
    auto [err, res] = GetNamedInstancesByHttpClient(activeMasterHttpClient_);
    if (err.OK()) {
        return std::make_pair(err, res);
    }
    CleanActiveMaster();
    return std::make_pair(ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, err.Msg()), QueryNamedInsResponse{});
}

std::pair<ErrorInfo, ResourceGroupUnit> FMClient::GetResourcesGroupWithRetry(const std::string &resourceGroupId)
{
    if (auto errInfo = ActivateMasterClientIfNeed(); !errInfo.OK())
        return std::make_pair(errInfo, ResourceGroupUnit{});
    auto [err, res] = GetResourceGroupTableByHttpClient(activeMasterHttpClient_, resourceGroupId);
    if (err.OK()) {
        return std::make_pair(err, res);
    }
    CleanActiveMaster();
    return std::make_pair(ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, err.Msg()), ResourceGroupUnit{});
}

std::pair<ErrorInfo, std::vector<ResourceUnit>> FMClient::GetResourcesWithRetry()
{
    if (auto err = ActivateMasterClientIfNeed(); !err.OK())
        return std::make_pair(err, std::vector<ResourceUnit>());
    auto [err, res] = GetResourcesByHttpClient(activeMasterHttpClient_);
    if (err.OK()) {
        return std::make_pair(err, res);
    }
    CleanActiveMaster();
    return std::make_pair(ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, err.Msg()), std::vector<ResourceUnit>());
}

std::pair<ErrorInfo, std::vector<ResourceUnit>> FMClient::GetResources()
{
    YRLOG_DEBUG("start to get resources.");
    auto startTime = std::chrono::steady_clock::now();
    for (int i = 0;
         std::chrono::duration_cast<std::chrono::seconds>(std::chrono::steady_clock::now() - startTime).count() <
         maxWaitTimeSec;
         ++i) {
        auto [err, res] = GetResourcesWithRetry();
        if (err.OK()) {
            return std::make_pair(err, res);
        }
        YRLOG_WARN("retry get resources, current times: {}, err: {}", i, err.Msg());
        if (cb_ != nullptr) {
            cb_();
        }
        std::this_thread::sleep_for(std::chrono::milliseconds(interval));
    }
    return std::make_pair(ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION,
                                    "failed to get resources, err: connect to function master timeout"),
                          std::vector<ResourceUnit>());
}

std::shared_ptr<HttpClient> FMClient::InitCtxAndHttpClient(void)
{
    ErrorInfo err;
    if (enableMTLS_) {
        try {
            auto ctx = std::make_shared<ssl::context>(ssl::context::tlsv12_client);
            ctx->set_options(ssl::context::default_workarounds | ssl::context::no_sslv2 | ssl::context::no_sslv3 |
                             ssl::context::no_tlsv1 | ssl::context::no_tlsv1_1);
            ctx->set_verify_mode(ssl::verify_peer);
            ctx->load_verify_file(libConfig_->verifyFilePath);
            ctx->use_certificate_chain_file(libConfig_->certificateFilePath);
            ctx->use_private_key_file(libConfig_->privateKeyPath, ssl::context::pem);
            return std::make_shared<AsyncHttpsClient>(this->ioc_, ctx);
        } catch (const std::exception &e) {
            YRLOG_ERROR("caught exception when init ssl context : {}", e.what());
            return {};
        } catch (...) {
            YRLOG_ERROR("caught unknown exception when init ssl context");
            return {};
        }
    }
    return std::make_shared<AsyncHttpClient>(this->ioc_);
}

std::pair<ErrorInfo, QueryNamedInsResponse> FMClient::QueryNamedInstances()
{
    YRLOG_DEBUG("start to get resource group table.");
    auto beginTime = std::chrono::steady_clock::now();
    for (int i = 0;
         std::chrono::duration_cast<std::chrono::seconds>(std::chrono::steady_clock::now() - beginTime).count() <
         maxWaitTimeSec;
         ++i) {
        auto [errInfo, res] = QueryNamedInstancesWithRetry();
        if (errInfo.OK()) {
            return std::make_pair(errInfo, res);
        }
        YRLOG_WARN("retry query named instance, current times: {}, err: {}", i, errInfo.Msg());
        if (cb_ != nullptr) {
            cb_();
        }
        std::this_thread::sleep_for(std::chrono::milliseconds(interval));
    }
    return std::make_pair(ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION,
                                    "failed to query named instance, err: connect to function master timeout"),
                          QueryNamedInsResponse{});
}

void FMClient::SetSubscribeActiveMasterCb(SubscribeActiveMasterCb cb)
{
    cb_ = cb;
}

void FMClient::UpdateActiveMaster(const std::string activeMasterAddr)
{
    std::lock_guard<std::mutex> activeMasterLock(activeMasterMu_);
    YRLOG_DEBUG("update active master, address is: {}", activeMasterAddr);
    activeMasterAddr_ = activeMasterAddr;
    if (activeMasterHttpClient_ != nullptr) {
        activeMasterHttpClient_->Stop();
    }
    activeMasterHttpClient_ = nullptr;
    condVar_.notify_all();
}

void FMClient::RemoveActiveMaster()
{
    std::unique_lock<std::mutex> activeMasterLock(activeMasterMu_);
    YRLOG_DEBUG("remove active master");
    CleanActiveMaster();
}

void FMClient::CleanActiveMaster()
{
    activeMasterAddr_ = "";
    if (activeMasterHttpClient_ != nullptr) {
        activeMasterHttpClient_->Stop();
    }
    activeMasterHttpClient_ = nullptr;
}

std::string FMClient::GetActiveMasterAddr()
{
    if (auto errInfo = ActivateMasterClientIfNeed(); !errInfo.OK()) {
        YRLOG_WARN("failed to get active master address, err: {}", errInfo.Msg());
        return "";
    }
    std::lock_guard<std::mutex> activeMasterLock(activeMasterMu_);
    return activeMasterAddr_;
}

}  // namespace Libruntime
}  // namespace YR