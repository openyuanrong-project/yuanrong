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

#include "src/libruntime/gwclient/gw_client.h"
#include "datasystem/object_client.h"
#include "json.hpp"
#include "src/libruntime/gwclient/gw_datasystem_client_wrapper.h"
#include "src/libruntime/utils/http_utils.h"
#include "src/utility/logger/logger.h"
#include "src/utility/notification_utility.h"
namespace YR {
namespace Libruntime {
const std::string REMOTE_CLIENT_ID_KEY = "remoteClientId";
const std::string REMOTE_CLIENT_ID_KEY_NEW = "X-Remote-Client-Id";
const std::string TRACE_ID_KEY = "traceId";
const std::string TENANT_ID_KEY = "tenantId";
const std::string TENANT_ID_KEY_NEW = "X-Tenant-Id";
const std::string X_AUTH_TOKEN = "X-Auth";

using json = nlohmann::json;
using YR::utility::NotificationUtility;
ErrorInfo ClientBuffer::Seal(const std::unordered_set<std::string> &nestedIds)
{
    std::shared_ptr<GwClient> gwClient = gwClientWeak_.lock();
    if (gwClient == nullptr) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "gwClient not existed");
    }
    PutRequest req;
    req.set_objectid(this->objectId_);
    req.set_objectdata(reinterpret_cast<const char *>(this->ImmutableData()), this->GetSize());
    for (const auto &id : nestedIds) {
        req.add_nestedobjectids(id);
    }
    req.set_writemode(static_cast<int32_t>(this->createParam_.writeMode));
    req.set_consistencytype(static_cast<int32_t>(this->createParam_.consistencyType));
    req.set_cachetype(static_cast<int32_t>(this->createParam_.cacheType));
    return gwClient->PosixObjPut(req);
}

ErrorInfo GwClient::Init(std::shared_ptr<HttpClient> httpClient, std::int32_t connectTimeout,
                         const std::string &authToken)
{
    if (init_) {
        return ErrorInfo();
    }
    this->httpClient_ = std::move(httpClient);
    this->authToken_ = authToken;
    return Init("", 0, connectTimeout);
}

ErrorInfo GwClient::Init(const std::string &ip, int port)
{
    return Init(ip, port, DS_CONNECT_TIMEOUT);
}

void GwClient::Init(std::shared_ptr<HttpClient> httpClient)
{
    if (init_) {
        return;
    }
    this->httpClient_ = std::move(httpClient);
    init_ = true;
}

ErrorInfo GwClient::Init(const std::string &addr, int port, std::int32_t connectTimeout)
{
    return Init(addr, port, false, false, "", datasystem::SensitiveValue{}, "", datasystem::SensitiveValue{}, "",
                datasystem::SensitiveValue{}, connectTimeout);
}

ErrorInfo GwClient::Init(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                         const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                         const std::string &dsPublicKey, const datasystem::SensitiveValue &token, const std::string &ak,
                         const datasystem::SensitiveValue &sk, std::int32_t connectTimeout)
{
    std::shared_ptr<DatasystemClientWrapper> dsClientWrapper =
        std::make_shared<GwDatasystemClientWrapper>(shared_from_this());
    asyncDecreRef_.Init(dsClientWrapper);
    init_ = true;
    connectTimeout_ = connectTimeout;
    return ErrorInfo();
}

ErrorInfo GwClient::Start(const std::string &jobID, const std::string &instanceID, const std::string &runtimeID,
                          const std::string &functionName, const SubscribeFunc &reSubscribeCb)
{
    if (!init_) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "should init before client start");
    }
    this->jobId_ = jobID;
    if (start_) {
        return ErrorInfo();
    }
    auto error = Lease();
    if (!error.OK()) {
        return error;
    }
    timer_ = timerWorker_->CreateTimer(KEEPALIVE_INTERVAL, KEEPALIVE_TIMES, [this](void) {
        auto error = this->KeepLease();
        if (!error.OK()) {
            YRLOG_WARN("Keepalive fails for {} consecutive times.", ++lostLeaseTimes_);
        } else {
            lostLeaseTimes_ = 0;
        }
    });
    start_ = true;
    return error;
}

void GwClient::Clear()
{
    std::vector<std::string> objectIds = refCountMap_.ToArray();
    refCountMap_.Clear();
    if (asyncDecreRef_.Push(objectIds, this->tenantId_)) {
        asyncDecreRef_.Stop();
    }
}

void GwClient::Stop(void)
{
    YRLOG_DEBUG("GwClient Stop");
    Clear();
    if (timer_ != nullptr) {
        timer_->cancel();
    }
    Release();
    init_ = false;
    start_ = false;
}

ErrorInfo GwClient::Lease()
{
    return HandleLease(POSIX_LEASE, PUT);
}

ErrorInfo GwClient::KeepLease()
{
    return HandleLease(POSIX_LEASE_KEEPALIVE, POST);
}

std::string VerbToString(const boost::beast::http::verb &v)
{
    switch (v) {
        case boost::beast::http::verb::get:
            return "GET";
        case boost::beast::http::verb::post:
            return "POST";
        case boost::beast::http::verb::put:
            return "PUT";
        case boost::beast::http::verb::delete_:
            return "DELETE";
        default:
            return "UNKNOWN";
    }
}

boost::beast::http::verb StringToVerb(const std::string &method)
{
    if (method == "GET") {
        return boost::beast::http::verb::get;
    } else if (method == "POST") {
        return boost::beast::http::verb::post;
    } else if (method == "PUT") {
        return boost::beast::http::verb::put;
    } else if (method == "DELETE") {
        return boost::beast::http::verb::delete_;
    } else {
        return boost::beast::http::verb::unknown;
    }
}

ErrorInfo GwClient::HandleLease(const std::string &url, const http::verb &verb)
{
    auto req = this->BuildLeaseRequest();
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, "");
    std::string body;
    req.SerializeToString(&body);
    auto asyncNotify = std::make_shared<NotificationUtility>();
    YRLOG_DEBUG("{} lease request, requestId :{}", VerbToString(verb), *requestId);
    auto logError = [verb, requestId](const ErrorInfo &err) -> ErrorInfo {
        YRLOG_DEBUG("{} lease Code: {}, MCode: {}, Msg: {}, requestId: {}", VerbToString(verb),
                    fmt::underlying(err.Code()), fmt::underlying(err.MCode()), err.Msg(), *requestId);
        return err;
    };
    httpClient_->SubmitInvokeRequest(
        verb, url, headers, body, requestId,
        [this, asyncNotify, requestId](const std::string &result, const boost::beast::error_code &errorCode,
                                       const uint statusCode) {
            auto err = GenRspError(errorCode, statusCode, result, requestId);
            if (err.OK()) {
                err = ParseLeaseResponse(result);
            }
            asyncNotify->Notify(err);
        });
    int32_t leaseTimeout = url == POSIX_LEASE ? connectTimeout_ : KEEPALIVE_INTERVAL / S_TO_MS;
    std::stringstream ss;
    ss << "lease http request timeout s: " << leaseTimeout << ", requestId: " << *requestId;
    return logError(asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(leaseTimeout), ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, ModuleCode::RUNTIME, ss.str())));
}

ErrorInfo GwClient::Release()
{
    return HandleLease(POSIX_LEASE, DELETE);
}

void GwClient::CreateAsync(const CreateRequest &req, CreateRespCallback createRespCallback, CreateCallBack callback,
                           int timeoutSec)
{
    auto requestId = std::make_shared<std::string>(req.requestid());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    YRLOG_DEBUG("create request, requestId :{}", *requestId);
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_CREATE, headers, body, requestId,
        [requestId, createRespCallback, callback](const std::string &result, const boost::beast::error_code &errorCode,
                                                  const uint statusCode) {
            CreateResponse createRsp;
            NotifyRequest notifyReq;
            std::stringstream ss;
            if (errorCode) {
                ss << "network error between client and frontend, error_code: " << errorCode.message()
                   << ", requestId: " << *requestId;
                createRsp.set_code(common::ERR_INNER_COMMUNICATION);
                createRsp.set_message(ss.str());
            } else if (!IsResponseSuccessful(statusCode)) {
                ss << "failed response status_code: " << std::to_string(statusCode) << ", result: " << result
                   << ", requestId: " << *requestId;
                createRsp.set_code(common::ERR_PARAM_INVALID);
                createRsp.set_message(ss.str());
            } else {
                notifyReq.ParseFromString(result);
                createRsp.set_code(notifyReq.code());
                createRsp.set_message(notifyReq.message());
                createRsp.set_instanceid(notifyReq.instanceid());
            }
            YRLOG_DEBUG("create response, code: {}, requestId :{}, instanceId : {}, msg: {} ",
                        fmt::underlying(createRsp.code()), *requestId, createRsp.instanceid(), createRsp.message());
            createRespCallback(createRsp);
            if (createRsp.code() != common::ERR_NONE) {
                return;
            }
            notifyReq.set_requestid(*requestId);
            callback(notifyReq);
        });
}

void GwClient::InvokeAsync(const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback, int timeoutSec)
{
    auto requestId = std::make_shared<std::string>(req->Immutable().requestid());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req->Immutable().SerializeToString(&body);
    YRLOG_DEBUG("invoke request, requestId :{}, instanceId: {}", *requestId, req->Immutable().instanceid());
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_INVOKE, headers, body, requestId,
        [requestId, callback](const std::string &result, const boost::beast::error_code &errorCode,
                              const uint statusCode) {
            NotifyRequest notifyReq;
            std::stringstream ss;
            if (errorCode) {
                ss << "network error between client and frontend, error_code: " << errorCode.message()
                   << ", requestId: " << *requestId;
                notifyReq.set_requestid(*requestId);
                notifyReq.set_code(common::ERR_INNER_COMMUNICATION);
                notifyReq.set_message(ss.str());
            } else if (!IsResponseSuccessful(statusCode)) {
                ss << "failed response status_code: " << std::to_string(statusCode) << ", result: " << result
                   << ", requestId: " << *requestId;
                notifyReq.set_requestid(*requestId);
                notifyReq.set_code(common::ERR_PARAM_INVALID);
                notifyReq.set_message(ss.str());
            } else {
                notifyReq.ParseFromString(result);
            }
            YRLOG_DEBUG("invoke response, code: {}, requestId: {}, msg: {}, small objects size: {}",
                        fmt::underlying(notifyReq.code()), *requestId, notifyReq.message(),
                        notifyReq.smallobjects_size());
            callback(notifyReq, ErrorInfo());
        });
}

void GwClient::KillAsync(const KillRequest &req, KillCallBack callback, int timeoutSec)
{
    auto requestId = std::make_shared<std::string>(req.requestid());
    auto instanceId = req.instanceid();
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    YRLOG_DEBUG("kill request, requestId: {}, instanceId: {}", *requestId, instanceId);
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_KILL, headers, body, requestId,
        [requestId, callback](const std::string &result, const boost::beast::error_code &errorCode,
                              const uint statusCode) {
            KillResponse killRsp;
            std::stringstream ss;
            if (errorCode) {
                ss << "network error between client and frontend, error_code: " << errorCode.message()
                   << ", requestId: " << *requestId;
                killRsp.set_code(common::ERR_INNER_COMMUNICATION);
                killRsp.set_message(ss.str());
            } else if (!IsResponseSuccessful(statusCode)) {
                ss << "failed response status_code: " << std::to_string(statusCode) << ", result: " << result
                   << ", requestId: " << *requestId;
                killRsp.set_code(common::ERR_PARAM_INVALID);
                killRsp.set_message(ss.str());
            } else {
                killRsp.ParseFromString(result);
            }
            YRLOG_DEBUG("kill response, code: {}, requestId :{}, msg: {}", fmt::underlying(killRsp.code()), *requestId,
                        killRsp.message());
            callback(killRsp, ErrorInfo());
        });
}

void GwClient::ExitAsync(const ExitRequest &req, ExitCallBack callback)
{
    STDERR_AND_THROW_EXCEPTION(ERR_INNER_SYSTEM_ERROR, RUNTIME,
                               "ExitAsync method not implemented when inCluster is false");
}

void GwClient::StateSaveAsync(const StateSaveRequest &req, StateSaveCallBack callback)
{
    STDERR_AND_THROW_EXCEPTION(ERR_INNER_SYSTEM_ERROR, RUNTIME,
                               "StateSaveAsync method not implemented when inCluster is false");
}

void GwClient::StateLoadAsync(const StateLoadRequest &req, StateLoadCallBack callback)
{
    STDERR_AND_THROW_EXCEPTION(ERR_INNER_SYSTEM_ERROR, RUNTIME, "StateLoadAsync is not supported with gateway client");
}

ErrorInfo GwClient::CreateBuffer(const std::string &objectId, size_t dataSize, std::shared_ptr<Buffer> &dataBuf,
                                 const CreateParam &createParam)
{
    dataBuf = std::make_shared<ClientBuffer>(dataSize, objectId, createParam, weak_from_this());
    return ErrorInfo();
}

std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> GwClient::GetBuffers(const std::vector<std::string> &ids,
                                                                                int timeoutMS)
{
    auto [err, results] = Get(ids, timeoutMS);
    return std::make_pair(err, results);
}

std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>> GwClient::GetBuffersWithoutRetry(
    const std::vector<std::string> &ids, int timeoutMS)
{
    auto [err, results] = Get(ids, timeoutMS);
    RetryInfo retryInfo{err, RetryType::NO_RETRY};
    return std::make_pair(retryInfo, results);
}

ErrorInfo GwClient::Put(std::shared_ptr<Buffer> data, const std::string &objID,
                        const std::unordered_set<std::string> &nestedID, const CreateParam &createParam)
{
    auto req = this->BuildObjPutRequest(data, objID, nestedID, createParam);
    return PosixObjPut(req);
}

SingleResult GwClient::Get(const std::string &objID, int timeoutMS)
{
    ErrorInfo err;
    std::vector<std::string> ids = {objID};
    auto multiRes = Get(ids, timeoutMS);
    if (multiRes.first.Code() != ErrorCode::ERR_OK) {
        return std::make_pair(multiRes.first, std::shared_ptr<Buffer>());
    }
    return std::make_pair(err, multiRes.second[0]);
}

MultipleResult GwClient::Get(const std::vector<std::string> &ids, int timeoutMS)
{
    auto result = std::make_shared<std::vector<std::shared_ptr<Buffer>>>();
    result->resize(ids.size());
    ErrorInfo err = PosixObjGet(ids, result, timeoutMS);
    return std::make_pair(err, *result);
}

ErrorInfo GwClient::UpdateToken(datasystem::SensitiveValue token)
{
    return ErrorInfo();
};

ErrorInfo GwClient::UpdateAkSk(std::string ak, datasystem::SensitiveValue sk)
{
    return ErrorInfo();
};

ErrorInfo GwClient::GenerateKey(std::string &key, const std::string &prefix, bool isPut)
{
    key = prefix;
    return ErrorInfo();
}

ErrorInfo GwClient::GetPrefix(const std::string &key, std::string &prefix)
{
    prefix = key;
    return ErrorInfo();
}

ErrorInfo GwClient::IncreGlobalReference(const std::vector<std::string> &objectIds)
{
    auto failedObjectIds = std::make_shared<std::vector<std::string>>();
    auto err = PosixGInCreaseRef(objectIds, failedObjectIds);
    if (err.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_ERROR(err.Msg());
        return err;
    }
    refCountMap_.IncreRefCount(objectIds);
    if (!failedObjectIds->empty()) {
        YRLOG_WARN("Datasystem failed to increase all objectRefs, fail count: {}", failedObjectIds->size());
        refCountMap_.DecreRefCount(*failedObjectIds);
    }
    return err;
}

ErrorInfo GwClient::DecreGlobalReference(const std::vector<std::string> &objectIds)
{
    ErrorInfo err;
    // if the objectId is not in the map, not decrease
    std::vector<std::string> needDecreObjectIds = refCountMap_.DecreRefCount(objectIds);
    if (!needDecreObjectIds.empty()) {
        bool success = asyncDecreRef_.Push(needDecreObjectIds, this->tenantId_);
        if (!success) {
            err.SetErrCodeAndMsg(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                                 YR::Libruntime::ModuleCode::DATASYSTEM, "async decrease thread has exited");
        }
    }
    return err;
}

ErrorInfo GwClient::Write(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam)
{
    return PosixKvSet(key, value, setParam);
}

ErrorInfo GwClient::MSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<Buffer>> &vals,
                           const MSetParam &mSetParam)
{
    return PosixKvMSetTx(keys, vals, mSetParam);
}

SingleReadResult GwClient::Read(const std::string &key, int timeoutMS)
{
    std::vector<std::string> keys = {key};
    auto result = Read(keys, timeoutMS, false);
    return std::make_pair(result.first[0], result.second);
}

MultipleReadResult GwClient::Read(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial)
{
    auto result = std::make_shared<std::vector<std::shared_ptr<Buffer>>>();
    result->resize(keys.size());
    auto err = PosixKvGet(keys, result, timeoutMS);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("GetValueWithTimeout error: Code:{}, MCode:{}, Msg:{}.", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg());
    }
    if (!allowPartial) {
        auto errInfo = ProcessKeyPartialResult(keys, *result, err, timeoutMS);
        if (errInfo.Code() != ErrorCode::ERR_OK) {
            err = errInfo;
        }
    }
    return std::make_pair(*result, err);
}

ErrorInfo GwClient::Del(const std::string &key)
{
    std::vector<std::string> keys = {key};
    auto result = Del(keys);
    return result.second;
}

MultipleDelResult GwClient::Del(const std::vector<std::string> &keys)
{
    auto failedKeys = std::make_shared<std::vector<std::string>>();
    auto err = PosixKvDel(keys, failedKeys);
    return std::make_pair(*failedKeys, err);
}

ErrorInfo GwClient::GenRspError(const boost::beast::error_code &errorCode, const uint statusCode,
                                const std::string &result, const std::shared_ptr<std::string> requestId)
{
    ErrorInfo err;
    std::stringstream ss;
    if (errorCode) {
        ss << "network error between client and frontend, error_code: " << errorCode.message()
           << ", requestId: " << *requestId;
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_COMMUNICATION, ss.str());
    }
    if (!IsResponseSuccessful(statusCode)) {
        ss << "failed response status_code: " << std::to_string(statusCode) << ", result: " << result
           << ", requestId: " << *requestId;
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, ss.str());
    }
    return err;
}

ErrorInfo GwClient::PosixKvGet(const std::vector<std::string> &keys,
                               std::shared_ptr<std::vector<std::shared_ptr<Buffer>>> results, int32_t timeoutMs)
{
    auto req = this->BuildKvGetRequest(keys, timeoutMs);
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    auto asyncNotify = std::make_shared<NotificationUtility>();
    YRLOG_DEBUG("kv get request, requestId :{}", *requestId);
    auto logError = [requestId](const ErrorInfo &err) -> ErrorInfo {
        YRLOG_DEBUG("kv get Code: {}, MCode: {}, Msg: {}, requestId: {}", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg(), *requestId);
        return err;
    };
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_KV_GET, headers, body, requestId,
        [this, asyncNotify, results, requestId](const std::string &result, const boost::beast::error_code &errorCode,
                                                const uint statusCode) {
            auto err = GenRspError(errorCode, statusCode, result, requestId);
            if (err.OK()) {
                err = ParseKvGetResponse(result, results);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "kv get http request timeout s: " << connectTimeout_ << ", requestId: " << *requestId;
    return logError(asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(connectTimeout_), ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, ModuleCode::RUNTIME, ss.str())));
}

ErrorInfo GwClient::PosixKvSet(const std::string &key, const std::shared_ptr<Buffer> value, const SetParam &setParam)
{
    auto req = this->BuildKvSetRequest(key, value, setParam);
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    auto asyncNotify = std::make_shared<NotificationUtility>();
    YRLOG_DEBUG("kv set request, requestId :{}", *requestId);
    auto logError = [requestId](const ErrorInfo &err) -> ErrorInfo {
        YRLOG_DEBUG("kv set Code: {}, MCode: {}, Msg: {}", fmt::underlying(err.Code()), fmt::underlying(err.MCode()),
                    err.Msg(), *requestId);
        return err;
    };
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_KV_SET, headers, body, requestId,
        [this, asyncNotify, requestId](const std::string &result, const boost::beast::error_code &errorCode,
                                       const uint statusCode) {
            auto err = GenRspError(errorCode, statusCode, result, requestId);
            if (err.OK()) {
                err = ParseKvSetResponse(result);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "kv set http request timeout s: " << connectTimeout_ << ", requestId: " << *requestId;
    return logError(asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(connectTimeout_), ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, ModuleCode::RUNTIME, ss.str())));
}

ErrorInfo GwClient::PosixKvMSetTx(const std::vector<std::string> &keys,
                                  const std::vector<std::shared_ptr<Buffer>> &vals, const MSetParam &mSetParam)
{
    auto req = this->BuildKvMSetTxRequest(keys, vals, mSetParam);
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    auto asyncNotify = std::make_shared<NotificationUtility>();
    YRLOG_DEBUG("kv multi set tx request, requestId :{}", *requestId);
    auto logError = [requestId](const ErrorInfo &err) -> ErrorInfo {
        YRLOG_DEBUG("kv multi set tx Code: {}, MCode: {}, Msg: {}", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg(), *requestId);
        return err;
    };
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_KV_MSET_TX, headers, body, requestId,
        [this, asyncNotify, requestId](const std::string &result, const boost::beast::error_code &errorCode,
                                       const uint statusCode) {
            auto err = GenRspError(errorCode, statusCode, result, requestId);
            if (err.OK()) {
                err = ParseKvMSetTxResponse(result);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "kv multi set tx http request timeout: " << connectTimeout_ << "s, requestId: " << *requestId;
    return logError(asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(connectTimeout_), ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, ModuleCode::RUNTIME, ss.str())));
}

ErrorInfo GwClient::PosixObjPut(const PutRequest &req)
{
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    auto asyncNotify = std::make_shared<NotificationUtility>();
    YRLOG_DEBUG("obj put request, requestId :{}", *requestId);
    auto logError = [requestId](const ErrorInfo &err) -> ErrorInfo {
        YRLOG_DEBUG("obj put Code: {}, MCode: {}, Msg: {}, requestId: {}", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg(), *requestId);
        return err;
    };
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_OBJ_PUT, headers, body, requestId,
        [this, asyncNotify, requestId](const std::string &result, const boost::beast::error_code &errorCode,
                                       const uint statusCode) {
            auto err = GenRspError(errorCode, statusCode, result, requestId);
            if (err.OK()) {
                err = ParseObjPutResponse(result);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "obj put http request timeout s: " << connectTimeout_ << ", requestId: " << *requestId;
    return logError(asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(connectTimeout_), ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, ModuleCode::RUNTIME, ss.str())));
}

ErrorInfo GwClient::PosixObjGet(const std::vector<std::string> &keys,
                                std::shared_ptr<std::vector<std::shared_ptr<Buffer>>> results, int32_t timeoutMs)
{
    auto req = this->BuildObjGetRequest(keys, timeoutMs);
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    auto asyncNotify = std::make_shared<NotificationUtility>();
    YRLOG_DEBUG("obj get request, requestId :{}", *requestId);
    auto logError = [requestId](const ErrorInfo &err) -> ErrorInfo {
        YRLOG_DEBUG("obj get Code: {}, MCode: {}, Msg: {}, requestId: {}", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg(), *requestId);
        return err;
    };
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_OBJ_GET, headers, body, requestId,
        [this, asyncNotify, results, requestId](const std::string &result, const boost::beast::error_code &errorCode,
                                                const uint statusCode) {
            auto err = GenRspError(errorCode, statusCode, result, requestId);
            if (err.OK()) {
                err = ParseObjGetResponse(result, results);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "obj get http request timeout s: " << connectTimeout_ << ", requestId: " << *requestId;
    return logError(asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(connectTimeout_), ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, ModuleCode::RUNTIME, ss.str())));
}

ErrorInfo GwClient::PosixKvDel(const std::vector<std::string> &keys,
                               std::shared_ptr<std::vector<std::string>> failedKeys)
{
    auto req = this->BuildKvDelRequest(keys);
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    auto asyncNotify = std::make_shared<NotificationUtility>();
    YRLOG_DEBUG("kv del request, requestId :{}", *requestId);
    auto logError = [requestId](const ErrorInfo &err) -> ErrorInfo {
        YRLOG_DEBUG("kv del Code: {}, MCode: {}, Msg: {}, requestId: {}", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg(), *requestId);
        return err;
    };
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_KV_DEL, headers, body, requestId,
        [this, asyncNotify, failedKeys, requestId](const std::string &result, const boost::beast::error_code &errorCode,
                                                   const uint statusCode) {
            auto err = GenRspError(errorCode, statusCode, result, requestId);
            if (err.OK()) {
                err = ParseKvDelResponse(result, failedKeys);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "kv del http request timeout s: " << connectTimeout_ << ", requestId: " << *requestId;
    return logError(asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(connectTimeout_), ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, ModuleCode::RUNTIME, ss.str())));
}

ErrorInfo GwClient::PosixGDecreaseRef(const std::vector<std::string> &objectIds,
                                      std::shared_ptr<std::vector<std::string>> failedObjectIds)
{
    auto req = this->BuildDecreaseRefRequest(objectIds);
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    auto asyncNotify = std::make_shared<NotificationUtility>();
    YRLOG_DEBUG("decrease ref request, requestId :{}", *requestId);
    auto logError = [requestId](const ErrorInfo &err) -> ErrorInfo {
        YRLOG_DEBUG("decrease ref Code: {}, MCode: {}, Msg: {}, requestId: {}", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg(), *requestId);
        return err;
    };
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_OBJ_DECREASE, headers, body, requestId,
        [this, asyncNotify, failedObjectIds, requestId](
            const std::string &result, const boost::beast::error_code &errorCode, const uint statusCode) {
            auto err = GenRspError(errorCode, statusCode, result, requestId);
            if (err.OK()) {
                err = ParseDecreaseRefResponse(result, failedObjectIds);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "decrease ref http request timeout s: " << connectTimeout_ << ", requestId: " << *requestId;
    return logError(asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(connectTimeout_), ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, ModuleCode::RUNTIME, ss.str())));
}

ErrorInfo GwClient::PosixGInCreaseRef(const std::vector<std::string> &objectIds,
                                      std::shared_ptr<std::vector<std::string>> failedObjectIds)
{
    auto req = this->BuildIncreaseRefRequest(objectIds);
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenRequestId());
    auto headers = this->BuildHeaders(this->jobId_, *requestId, this->tenantId_);
    std::string body;
    req.SerializeToString(&body);
    auto asyncNotify = std::make_shared<NotificationUtility>();
    YRLOG_DEBUG("increase ref request, requestId :{}", *requestId);
    auto logError = [requestId](const ErrorInfo &err) -> ErrorInfo {
        YRLOG_DEBUG("increase ref Code: {}, MCode: {}, Msg: {}, requestId: {}", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg(), *requestId);
        return err;
    };
    httpClient_->SubmitInvokeRequest(
        POST, POSIX_OBJ_INCREASE, headers, body, requestId,
        [this, asyncNotify, failedObjectIds, requestId](
            const std::string &result, const boost::beast::error_code &errorCode, const uint statusCode) {
            auto err = GenRspError(errorCode, statusCode, result, requestId);
            if (err.OK()) {
                err = ParseIncreaseRefResponse(result, failedObjectIds);
            }
            asyncNotify->Notify(err);
        });
    std::stringstream ss;
    ss << "increase ref request by http has timed out: " << connectTimeout_ << ", requestId: " << *requestId;
    return logError(asyncNotify->WaitForNotificationWithTimeout(
        absl::Seconds(connectTimeout_), ErrorInfo(ErrorCode::ERR_INNER_COMMUNICATION, ModuleCode::RUNTIME, ss.str())));
}

void GwClient::SetTenantId(const std::string &tenantId)
{
    tenantId_ = tenantId;
}

ErrorInfo GwClient::ParseLeaseResponse(const std::string &result)
{
    LeaseResponse leaseRsp;
    // An empty result is valid: protobuf serializes default values (code=0=ERR_NONE) as empty bytes.
    if (!result.empty() && !leaseRsp.ParseFromString(result)) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "failed to parse lease response");
    }
    return leaseRsp.code() == common::ERR_NONE
               ? ErrorInfo()
               : ErrorInfo(static_cast<ErrorCode>(leaseRsp.code()), ModuleCode::RUNTIME, leaseRsp.message());
}

ErrorInfo GwClient::ParseObjGetResponse(const std::string &result,
                                        std::shared_ptr<std::vector<std::shared_ptr<Buffer>>> results)
{
    GetResponse getRsp;
    getRsp.ParseFromString(result);
    for (int i = 0; i < getRsp.buffers_size(); i++) {
        if (getRsp.buffers(i).size() == 0) {
            continue;
        }
        auto buf = std::make_shared<NativeBuffer>(getRsp.buffers(i).size());
        buf->MemoryCopy(getRsp.buffers(i).data(), getRsp.buffers(i).size());
        (*results)[i] = std::move(buf);
    }
    return getRsp.code() == common::ERR_NONE
               ? ErrorInfo()
               : ErrorInfo(static_cast<ErrorCode>(getRsp.code()), ModuleCode::DATASYSTEM, getRsp.message());
}

ErrorInfo GwClient::ParseObjPutResponse(const std::string &result)
{
    PutResponse putRsp;
    putRsp.ParseFromString(result);
    return putRsp.code() == common::ERR_NONE
               ? ErrorInfo()
               : ErrorInfo(static_cast<ErrorCode>(putRsp.code()), ModuleCode::DATASYSTEM, putRsp.message());
}

ErrorInfo GwClient::ParseKvGetResponse(const std::string &result,
                                       std::shared_ptr<std::vector<std::shared_ptr<Buffer>>> results)
{
    KvGetResponse kvRsp;
    kvRsp.ParseFromString(result);
    for (int i = 0; i < kvRsp.values_size(); i++) {
        if (kvRsp.values(i).size() == 0) {
            continue;
        }
        auto buf = std::make_shared<NativeBuffer>(kvRsp.values(i).size());
        buf->MemoryCopy(kvRsp.values(i).data(), kvRsp.values(i).size());
        (*results)[i] = std::move(buf);
    }
    return kvRsp.code() == common::ERR_NONE
               ? ErrorInfo()
               : ErrorInfo(static_cast<ErrorCode>(kvRsp.code()), ModuleCode::DATASYSTEM, kvRsp.message());
}

ErrorInfo GwClient::ParseKvSetResponse(const std::string &result)
{
    KvSetResponse kvRsp;
    kvRsp.ParseFromString(result);
    return kvRsp.code() == common::ERR_NONE
               ? ErrorInfo()
               : ErrorInfo(static_cast<ErrorCode>(kvRsp.code()), ModuleCode::DATASYSTEM, kvRsp.message());
}

ErrorInfo GwClient::ParseKvMSetTxResponse(const std::string &result)
{
    KvMSetTxResponse mSetRsp;
    mSetRsp.ParseFromString(result);
    return mSetRsp.code() == common::ERR_NONE
               ? ErrorInfo()
               : ErrorInfo(static_cast<ErrorCode>(mSetRsp.code()), ModuleCode::DATASYSTEM, mSetRsp.message());
}

ErrorInfo GwClient::ParseKvDelResponse(const std::string &result, std::shared_ptr<std::vector<std::string>> failedKeys)
{
    KvDelResponse kvRsp;
    kvRsp.ParseFromString(result);
    failedKeys->resize(kvRsp.failedkeys_size());
    for (int i = 0; i < kvRsp.failedkeys_size(); i++) {
        (*failedKeys)[i] = kvRsp.failedkeys(i);
    }
    return kvRsp.code() == common::ERR_NONE
               ? ErrorInfo()
               : ErrorInfo(static_cast<ErrorCode>(kvRsp.code()), ModuleCode::DATASYSTEM, kvRsp.message());
}

ErrorInfo GwClient::ParseIncreaseRefResponse(const std::string &result,
                                             std::shared_ptr<std::vector<std::string>> failedObjectIds)
{
    IncreaseRefResponse increaseRefRsp;
    increaseRefRsp.ParseFromString(result);
    failedObjectIds->resize(increaseRefRsp.failedobjectids_size());
    for (int i = 0; i < increaseRefRsp.failedobjectids_size(); i++) {
        (*failedObjectIds)[i] = increaseRefRsp.failedobjectids(i);
    }
    return increaseRefRsp.code() == common::ERR_NONE ? ErrorInfo()
                                                     : ErrorInfo(static_cast<ErrorCode>(increaseRefRsp.code()),
                                                                 ModuleCode::DATASYSTEM, increaseRefRsp.message());
}

ErrorInfo GwClient::ParseDecreaseRefResponse(const std::string &result,
                                             std::shared_ptr<std::vector<std::string>> failedObjectIds)
{
    DecreaseRefResponse decreaseRefRsp;
    decreaseRefRsp.ParseFromString(result);
    failedObjectIds->resize(decreaseRefRsp.failedobjectids_size());
    for (int i = 0; i < decreaseRefRsp.failedobjectids_size(); i++) {
        (*failedObjectIds)[i] = decreaseRefRsp.failedobjectids(i);
    }
    return decreaseRefRsp.code() == common::ERR_NONE ? ErrorInfo()
                                                     : ErrorInfo(static_cast<ErrorCode>(decreaseRefRsp.code()),
                                                                 ModuleCode::DATASYSTEM, decreaseRefRsp.message());
}

ErrorInfo GwClient::ReleaseGRefs(const std::string &remoteId)
{
    return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, "not support out of cluster");
}

LeaseRequest GwClient::BuildLeaseRequest()
{
    LeaseRequest leaseReq;
    leaseReq.set_remoteclientid(this->jobId_);
    return leaseReq;
}

GetRequest GwClient::BuildObjGetRequest(const std::vector<std::string> &keys, int32_t timeoutMs)
{
    GetRequest req;
    for (const auto &key : keys) {
        req.add_objectids(key);
        req.set_timeoutms(timeoutMs);
    }
    return req;
}

PutRequest GwClient::BuildObjPutRequest(std::shared_ptr<Buffer> data, const std::string &objID,
                                        const std::unordered_set<std::string> &nestedID, const CreateParam &createParam)
{
    PutRequest req;
    req.set_objectid(objID);
    req.set_objectdata(data->ImmutableData(), data->GetSize());
    for (const auto &id : nestedID) {
        req.add_nestedobjectids(id);
    }
    req.set_writemode(static_cast<int32_t>(createParam.writeMode));
    req.set_consistencytype(static_cast<int32_t>(createParam.consistencyType));
    req.set_cachetype(static_cast<int32_t>(createParam.cacheType));
    return req;
}

KvSetRequest GwClient::BuildKvSetRequest(const std::string &key, const std::shared_ptr<Buffer> value,
                                         const SetParam &setParam)
{
    KvSetRequest req;
    req.set_key(key);
    req.set_value(value->ImmutableData(), value->GetSize());
    req.set_existence(static_cast<int32_t>(setParam.existence));
    req.set_writemode(static_cast<int32_t>(setParam.writeMode));
    req.set_ttlsecond(setParam.ttlSecond);
    req.set_cachetype(static_cast<int32_t>(setParam.cacheType));
    return req;
}

KvMSetTxRequest GwClient::BuildKvMSetTxRequest(const std::vector<std::string> &keys,
                                               const std::vector<std::shared_ptr<Buffer>> &vals,
                                               const MSetParam &mSetParam)
{
    KvMSetTxRequest req;
    for (const auto &key : keys) {
        req.add_keys(key);
    }
    for (const auto &val : vals) {
        req.add_values(val->ImmutableData(), val->GetSize());
    }
    req.set_existence(static_cast<int32_t>(mSetParam.existence));
    req.set_writemode(static_cast<int32_t>(mSetParam.writeMode));
    req.set_ttlsecond(mSetParam.ttlSecond);
    req.set_cachetype(static_cast<int32_t>(mSetParam.cacheType));
    return req;
}

KvGetRequest GwClient::BuildKvGetRequest(const std::vector<std::string> &keys, int32_t timeoutMs)
{
    KvGetRequest req;
    for (const auto &key : keys) {
        req.add_keys(key);
    }
    req.set_timeoutms(timeoutMs);
    return req;
}

KvDelRequest GwClient::BuildKvDelRequest(const std::vector<std::string> &keys)
{
    KvDelRequest req;
    for (const auto &key : keys) {
        req.add_keys(key);
    }
    return req;
}

IncreaseRefRequest GwClient::BuildIncreaseRefRequest(const std::vector<std::string> &objectIds)
{
    IncreaseRefRequest req;
    for (const auto &objId : objectIds) {
        req.add_objectids(objId);
    }
    req.set_remoteclientid(YR::utility::ParseRealJobId(this->jobId_));
    return req;
}

DecreaseRefRequest GwClient::BuildDecreaseRefRequest(const std::vector<std::string> &objectIds)
{
    DecreaseRefRequest req;
    for (size_t i = 0; i < objectIds.size(); i++) {
        req.add_objectids(objectIds[i]);
    }
    req.set_remoteclientid(YR::utility::ParseRealJobId(this->jobId_));
    return req;
}

std::unordered_map<std::string, std::string> GwClient::BuildHeaders(const std::string &remoteClientId,
                                                                    const std::string &traceId,
                                                                    const std::string &tenantId)
{
    std::unordered_map<std::string, std::string> headers;
    if (!remoteClientId.empty()) {
        headers.emplace(REMOTE_CLIENT_ID_KEY, remoteClientId);
        headers.emplace(REMOTE_CLIENT_ID_KEY_NEW, remoteClientId);
    }
    if (!traceId.empty()) {
        headers.emplace(TRACE_ID_KEY, traceId);
        headers.emplace(TRACE_ID_KEY_NEW, traceId);
    }
    if (!tenantId.empty()) {
        headers.emplace(TENANT_ID_KEY, tenantId);
        headers.emplace(TENANT_ID_KEY_NEW, tenantId);
    }
    if (!authToken_.empty()) {
        headers.emplace(X_AUTH_TOKEN, authToken_);
    }
    return headers;
}

std::pair<std::unordered_map<std::string, std::string>, std::string> GwClient::BuildRequestWithAkSk(
    const std::shared_ptr<InvokeSpec> spec, const std::string &url)
{
    std::string callReq;
    const size_t userEventIndex = 1;
    if (spec->invokeArgs.size() > userEventIndex) {
        auto &invokeArg = spec->invokeArgs[userEventIndex];
        callReq = std::string(static_cast<const char *>(invokeArg.dataObj->data->ImmutableData()),
                              invokeArg.dataObj->data->GetSize());
    } else {
        YRLOG_ERROR("invoke args size not valid");
    }
    std::string event;
    try {
        json j = json::parse(callReq);
        if (j.contains("body")) {
            if (j["body"].is_string()) {
                event = j["body"].get<std::string>();
            } else {
                event = j["body"].dump();
            }
        } else {
            YRLOG_ERROR("event is empty, callReq {}", callReq);
        }
    } catch (const std::exception &e) {
        YRLOG_ERROR("{} JSON parse error: {}", callReq, e.what());
    }
    std::unordered_map<std::string, std::string> headers;
    if (!spec->traceId.empty()) {
        headers.emplace(TRACE_ID_KEY_NEW, spec->traceId);
    }
    headers.emplace(REMOTE_CLIENT_ID_KEY, jobId_);       // for llt test
    headers.emplace(REMOTE_CLIENT_ID_KEY_NEW, jobId_);   // keep in sync with BuildHeaders
    headers.emplace(INSTANCE_CPU_KEY, std::to_string(spec->opts.cpu));
    headers.emplace(INSTANCE_MEMORY_KEY, std::to_string(spec->opts.memory));
    std::string ak;
    datasystem::SensitiveValue sk;
    security_->GetAKSK(ak, sk);
    if (ak.empty() || sk.Empty()) {
        YRLOG_WARN("ak or sk is empty");
        return std::make_pair(headers, event);
    }
    SignHttpRequest(ak, sk, headers, event, url);
    return std::make_pair(headers, event);
}

void GwClient::GroupCreateAsync(const CreateRequests &reqs, CreateRespsCallback respCallback, CreateCallBack callback,
                                int timeoutSec)
{
    // wait inplement not in cluster
}

std::string TransformJson(uint statusCode, const std::string &input)
{
    if (IsResponseSuccessful(statusCode)) {
        json out;
        out["innerCode"] = "0";
        try {
            json j = json::parse(input);
            out["body"] = j;
        } catch (const std::exception &e) {
            YRLOG_WARN("json parse error {}", e.what());
            out["body"] = input;
        }
        return out.dump();
    }
    if (IsResponseServerError(statusCode)) {
        try {
            json j = json::parse(input);
            json out;
            if (j.contains("code")) {
                out["innerCode"] = std::to_string(j["code"].get<int>());
            }
            if (j.contains("message")) {
                out["body"] = j["message"];
            }
            return out.dump();
        } catch (const std::exception &e) {
            YRLOG_WARN("{} JSON parse error: {}", input, e.what());
            return "{}";
        }
    }
    return input;
}

void GwClient::InvocationAsync(const std::string &url, const std::shared_ptr<InvokeSpec> spec,
                               const InvocationCallback &callback)
{
    auto requestId = std::make_shared<std::string>(spec->requestId);
    auto [headers, body] = BuildRequestWithAkSk(spec, url);
    YRLOG_DEBUG("invocation request, url: {}, requestId: {}", url, *requestId);
    httpClient_->SubmitInvokeRequest(
        POST, url, headers, body, requestId,
        [requestId, callback](const std::string &result, const boost::beast::error_code &errorCode,
                              const uint statusCode) {
            if (errorCode) {
                std::stringstream ss;
                ss << "invocation network error between client and frontend, error_code: " << errorCode.message()
                   << ", requestId: " << *requestId;
                YRLOG_ERROR(ss.str());
                callback(*requestId, ErrorCode::ERR_INNER_COMMUNICATION, ss.str());
            } else if (!IsResponseSuccessful(statusCode)) {
                YRLOG_ERROR("invocation response, status_code: {}, result: {}, requestId: {}", statusCode, result,
                            *requestId);
                if (IsResponseServerError(statusCode)) {
                    callback(*requestId, ErrorCode::ERR_OK, TransformJson(statusCode, result));
                } else {
                    callback(*requestId, ErrorCode::ERR_INNER_SYSTEM_ERROR, result);
                }
            } else {
                YRLOG_DEBUG("invocation response, http status code: {}, requestId: {}", statusCode, *requestId);
                callback(*requestId, ErrorCode::ERR_OK, TransformJson(statusCode, result));
            }
        });
}

void GwClient::InvocationSync(const std::string &method, const std::string &path,
                              const std::unordered_map<std::string, std::string> &headers,
                              const std::string &body, const InvocationCallback &callback)
{
    YRLOG_DEBUG("start request to method: {}, path: {}", method, path);
    auto verb = StringToVerb(method);
    auto requestId = std::make_shared<std::string>(YR::utility::IDGenerator::GenTraceId());
    httpClient_->SubmitInvokeRequest(
        verb, path, headers, body, requestId,
        [requestId, callback](const std::string &result, const boost::beast::error_code &errorCode,
                              const uint statusCode) {
            if (errorCode) {
                std::stringstream ss;
                ss << "invocation network error between client and frontend, error_code: " << errorCode.message()
                   << ", requestId: " << *requestId;
                YRLOG_ERROR(ss.str());
                callback(*requestId, ErrorCode::ERR_INNER_COMMUNICATION, ss.str());
            } else if (!IsResponseSuccessful(statusCode)) {
                YRLOG_ERROR("invocation response, status_code: {}, result: {}, requestId: {}", statusCode, result,
                            *requestId);
                if (IsResponseServerError(statusCode)) {
                    callback(*requestId, ErrorCode::ERR_INNER_COMMUNICATION, result);
                } else {
                    callback(*requestId, ErrorCode::ERR_INNER_SYSTEM_ERROR, result);
                }
            } else {
                YRLOG_DEBUG("invocation response, http status code: {}, requestId: {}", statusCode, *requestId);
                callback(*requestId, ErrorCode::ERR_OK, result);
            }
    });
}
}  // namespace Libruntime
}  // namespace YR
