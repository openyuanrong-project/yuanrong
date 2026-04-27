/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include <string>
#include <vector>

#include "datasystem_state_store.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/utils/constants.h"
#include "src/libruntime/utils/utils.h"
#include "src/utility/logger/logger.h"

using namespace datasystem;
using namespace std::chrono;

namespace YR {
namespace Libruntime {
using datasystem::Status;

ErrorInfo DSCacheStateStore::Init(const std::string &ip, int port, std::int32_t connectTimeout)
{
    return this->Init(ip, port, false, false, "", datasystem::SensitiveValue{}, "", datasystem::SensitiveValue{}, "",
                      datasystem::SensitiveValue{}, connectTimeout);
}

ErrorInfo DSCacheStateStore::Init(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                                  const std::string &runtimePublicKey,
                                  const datasystem::SensitiveValue &runtimePrivateKey, const std::string &dsPublicKey,
                                  const datasystem::SensitiveValue &token, const std::string &ak,
                                  const datasystem::SensitiveValue &sk, std::int32_t connectTimeout)
{
    ErrorInfo err;
    YRLOG_DEBUG("Datasystem State store init, ip = {}, port = {}", ip, port);
    connectOpts.host = ip;
    connectOpts.port = port;
    connectOpts.connectTimeoutMs = connectTimeout * S_TO_MS;
    if (encryptEnable) {
        connectOpts.clientPublicKey = runtimePublicKey;
        connectOpts.clientPrivateKey = runtimePrivateKey;
        connectOpts.serverPublicKey = dsPublicKey;
    }
    if (enableDsAuth) {
        GetAuthConnectOpts(connectOpts, ak, sk, token);
    }
    return err;
}

ErrorInfo DSCacheStateStore::Init(datasystem::ConnectOptions &inputConnOpt)
{
    connectOpts.host = inputConnOpt.host;
    connectOpts.port = inputConnOpt.port;
    connectOpts.clientPublicKey = inputConnOpt.clientPublicKey;
    connectOpts.clientPrivateKey = inputConnOpt.clientPrivateKey;
    connectOpts.serverPublicKey = inputConnOpt.serverPublicKey;
    connectOpts.accessKey = inputConnOpt.accessKey;
    connectOpts.secretKey = inputConnOpt.secretKey;
    connectOpts.token = inputConnOpt.token;
    connectOpts.connectTimeoutMs = inputConnOpt.connectTimeoutMs;
    connectOpts.tenantId = inputConnOpt.tenantId;
    return ErrorInfo();
}

ErrorInfo DSCacheStateStore::DoInitOnce(void)
{
    dsStateClient = std::make_shared<KVClient>(this->connectOpts);
    Status status = dsStateClient->Init();
    std::string msg = "failed to init state store, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_CONNECTION_FAILED, msg);
    isInit = true;
    if (!tokenUpdated.Empty()) {
        (void)UpdateToken(tokenUpdated);
    }
    return ErrorInfo();
}

void DSCacheStateStore::InitOnce(void)
{
    std::call_once(this->initFlag, [this]() { this->initErr = this->DoInitOnce(); });
}

ErrorInfo DSCacheStateStore::Init(const DsConnectOptions &options)
{
    ErrorInfo err;
    YRLOG_DEBUG("Datasystem State store init, ip = {}, port = {}", options.host, options.port);
    ConnectOptions connectOptsInput{options.host,
                               options.port,
                               options.connectTimeoutMs,
                               options.connectTimeoutMs,
                               options.token,
                               options.clientPublicKey,
                               options.clientPrivateKey,
                               options.serverPublicKey,
                               options.accessKey,
                               options.secretKey,
                               options.tenantId,
                               options.enableCrossNodeConnection};
    this->connectOpts = connectOptsInput;
    InitOnce();
    return this->initErr;
}

ErrorInfo DSCacheStateStore::GenerateKey(std::string &returnKey)
{
    STATE_STORE_INIT_ONCE();
    returnKey = dsStateClient->GenerateKey();
    return ErrorInfo();
}

ErrorInfo DSCacheStateStore::DoHealthCheck()
{
    if (dsStateClient == nullptr) {
        YRLOG_INFO("ds client is nullptr");
        return ErrorInfo();
    }
    Status status = dsStateClient->HealthCheck();
    if (!status.IsOk()) {
        YRLOG_INFO("health check failed, code:{}, msg:{}", fmt::underlying(status.GetCode()), status.ToString());
        return GenerateSetErrorInfo(status);
    }
    return ErrorInfo();
}


ErrorInfo DSCacheStateStore::StartHealthCheck()
{
    STATE_STORE_INIT_ONCE();
    if (timerWorker_ == nullptr) {
        timerWorker_ = std::make_shared<YR::utility::TimerWorker>();
        YRLOG_INFO("start ds client health check");
        timer_ = timerWorker_->CreateTimer(DS_HEALTHCHECK_INTERVAL, DS_HEALTHCHECK_TIMES,
                                           [this]() {
            auto error = this->DoHealthCheck();
            if (!error.OK()) {
                YRLOG_WARN("ds object client health check failed for {} times.", ++lostHealthTimes_);
                if (lostHealthTimes_ >= DS_HEALTHCHECK_FAILED_LIMIT) {
                    YRLOG_ERROR("ds object client health check failed reach max limit 10, start exiting.");
                    timer_->cancel();
                    YR::Libruntime::AlarmInfo dsAlarmInfo;
                    dsAlarmInfo.id = "YuanrongDsWorkerUnhealthy00001";
                    dsAlarmInfo.customOptions["site"] = MetricsAdaptor::GetInstance()->GetContextValue("site");
                    dsAlarmInfo.customOptions["application_id"] =
                        MetricsAdaptor::GetInstance()->GetContextValue("application_id");
                    dsAlarmInfo.customOptions["service_id"] =
                        MetricsAdaptor::GetInstance()->GetContextValue("service_id");
                    dsAlarmInfo.customOptions["tenant_id"] =
                        MetricsAdaptor::GetInstance()->GetContextValue("tenant_id");
                    dsAlarmInfo.customOptions["clear_type"] = "ADAC";
                    dsAlarmInfo.customOptions["op_type"] = "firing";
                    dsAlarmInfo.alarmName = "yr_ds_alarm";
                    dsAlarmInfo.alarmSeverity = AlarmSeverity::MAJOR;
                    dsAlarmInfo.cause = error.Msg();
                    MetricsAdaptor::GetInstance()->SetAlarm("yr_ds_alarm",
                                             "ds client health check failed reach max limit 20",
                                             dsAlarmInfo);
                }
            } else {
                lostHealthTimes_ = 0;
            }
        });
    }
    return ErrorInfo();
}

ErrorInfo DSCacheStateStore::Write(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam)
{
    STATE_STORE_INIT_ONCE();
    StringView stringView(static_cast<const char *>(value->ImmutableData()), value->GetSize());
    datasystem::SetParam p;
    p.existence = static_cast<datasystem::ExistenceOpt>(setParam.existence);
    p.writeMode = static_cast<datasystem::WriteMode>(setParam.writeMode);
    p.ttlSecond = setParam.ttlSecond;
    p.cacheType = static_cast<datasystem::CacheType>(setParam.cacheType);
    Status status = dsStateClient->Set(key, stringView, p);
    if (!status.IsOk()) {
        return GenerateSetErrorInfo(status);
    }
    return ErrorInfo();
}

ErrorInfo DSCacheStateStore::Write(std::shared_ptr<Buffer> value, SetParam setParam, std::string &returnKey)
{
    STATE_STORE_INIT_ONCE();
    StringView stringView(static_cast<const char *>(value->ImmutableData()), value->GetSize());
    datasystem::SetParam p;
    p.existence = static_cast<datasystem::ExistenceOpt>(setParam.existence);
    p.writeMode = static_cast<datasystem::WriteMode>(setParam.writeMode);
    p.ttlSecond = setParam.ttlSecond;
    p.cacheType = static_cast<datasystem::CacheType>(setParam.cacheType);
    returnKey = dsStateClient->Set(stringView, p);
    return ErrorInfo();
}

ErrorInfo DSCacheStateStore::MSetTx(const std::vector<std::string> &keys,
                                    const std::vector<std::shared_ptr<Buffer>> &vals, const MSetParam &mSetParam)
{
    STATE_STORE_INIT_ONCE();
    if (keys.size() != vals.size()) {
        std::string msg = "MSetTx arguments vector length not equal";
        YRLOG_ERROR(msg);
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME, msg);
    }
    std::vector<StringView> valViews;
    valViews.resize(keys.size());
    for (size_t i = 0; i < keys.size(); i++) {
        valViews[i] = StringView(static_cast<const char *>(vals[i]->ImmutableData()), vals[i]->GetSize());
    }
    datasystem::MSetParam p;
    p.existence = static_cast<datasystem::ExistenceOpt>(mSetParam.existence);
    p.writeMode = static_cast<datasystem::WriteMode>(mSetParam.writeMode);
    p.ttlSecond = mSetParam.ttlSecond;
    p.cacheType = static_cast<datasystem::CacheType>(mSetParam.cacheType);
    Status status = dsStateClient->MSetTx(keys, valViews, p);
    if (!status.IsOk()) {
        return GenerateSetErrorInfo(status);
    }
    return ErrorInfo();
}

SingleReadResult DSCacheStateStore::Read(const std::string &key, int timeoutMS)
{
    STATE_STORE_INIT_ONCE_RETURN_PAIR(nullptr);
    std::vector<std::string> keys = {key};
    MultipleReadResult multipleResult = Read(keys, timeoutMS, false);
    return std::make_pair(multipleResult.first[0], multipleResult.second);
}

MultipleReadResult DSCacheStateStore::Read(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial)
{
    std::vector<std::shared_ptr<Buffer>> result{};
    STATE_STORE_INIT_ONCE_RETURN_PAIR(result);
    result.resize(keys.size());

    GetParams params;
    ErrorInfo err = GetValueWithTimeout(keys, result, timeoutMS, params);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("GetValueWithTimeout error: Code:{}, MCode:{}, Msg:{}.", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg());
        return std::make_pair(result, err);
    }
    if (!allowPartial) {
        ErrorInfo partialErrInfo = ProcessKeyPartialResult(keys, result, err, timeoutMS);
        if (partialErrInfo.Code() != ErrorCode::ERR_OK) {
            err = partialErrInfo;
        }
    }
    return std::make_pair(result, err);
}

MultipleReadResult DSCacheStateStore::GetWithParam(const std::vector<std::string> &keys, 
                                                   const GetParams &params, int timeoutMs)
{
    std::vector<std::shared_ptr<Buffer>> results{};
    STATE_STORE_INIT_ONCE_RETURN_PAIR(results);
    results.resize(keys.size());
    ErrorInfo err = GetValueWithTimeout(keys, results, timeoutMs, params);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("GetValueWithTimeout error: Code:{}, MCode:{}, Msg:{}.", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg());
        return std::make_pair(results, err);
    }
    return std::make_pair(results, ErrorInfo());
}

ErrorInfo DSCacheStateStore::QuerySize(const std::vector<std::string> &keys, std::vector<uint64_t> &outSizes)
{
    STATE_STORE_INIT_ONCE();
    ErrorInfo errInfo;
    Status status = dsStateClient->QuerySize(keys, outSizes);
    if (status.IsError()) {
        YRLOG_ERROR("failed to query the value sizes from state store, errMsg:{}", status.ToString());
        ErrorCode errCode = ConvertDatasystemErrorToCore(status.GetCode(), ErrorCode::ERR_DATASYSTEM_FAILED);
        errInfo.SetErrCodeAndMsg(errCode, YR::Libruntime::ModuleCode::DATASYSTEM, status.ToString(), status.GetCode());
    }
    return errInfo;
}

ErrorInfo DSCacheStateStore::Del(const std::string &key)
{
    STATE_STORE_INIT_ONCE();
    ErrorInfo errInfo;
    Status status = dsStateClient->Del(key);
    if (status.IsError()) {
        YRLOG_ERROR("failed to del the value of state store, errMsg:{}", status.ToString());
        ErrorCode errCode = ConvertDatasystemErrorToCore(status.GetCode(), ErrorCode::ERR_DATASYSTEM_FAILED);
        errInfo.SetErrCodeAndMsg(errCode, YR::Libruntime::ModuleCode::DATASYSTEM, status.ToString(), status.GetCode());
    }
    return errInfo;
}

// return failedKeys and ErrorInfo
MultipleDelResult DSCacheStateStore::Del(const std::vector<std::string> &keys)
{
    std::vector<std::string> failedKeys;
    STATE_STORE_INIT_ONCE_RETURN_PAIR(failedKeys);
    ErrorInfo errInfo;
    Status status = dsStateClient->Del(keys, failedKeys);
    if (status.IsError()) {
        YRLOG_ERROR("failed to del all values of state store, errMsg:{}", status.ToString());
        ErrorCode errCode = ConvertDatasystemErrorToCore(status.GetCode(), ErrorCode::ERR_DATASYSTEM_FAILED);
        errInfo.SetErrCodeAndMsg(errCode, YR::Libruntime::ModuleCode::DATASYSTEM, status.ToString(), status.GetCode());
    }
    return std::make_pair(failedKeys, errInfo);
}

MultipleExistResult DSCacheStateStore::Exist(const std::vector<std::string> &keys)
{
    std::vector<bool> exists;
    STATE_STORE_INIT_ONCE_RETURN_PAIR(exists);
    exists.resize(keys.size());
    ErrorInfo errInfo;
    Status status = dsStateClient->Exist(keys, exists);
    if (status.IsError()) {
        YRLOG_ERROR("failed to query keys from state store, errMsg:{}", status.ToString());
        ErrorCode errCode = ConvertDatasystemErrorToCore(status.GetCode(), ErrorCode::ERR_DATASYSTEM_FAILED);
        errInfo.SetErrCodeAndMsg(errCode, YR::Libruntime::ModuleCode::DATASYSTEM, status.ToString(), status.GetCode());
    }
    return std::make_pair(exists, errInfo);
}

void DSCacheStateStore::Shutdown()
{
    if (dsStateClient == nullptr) {
        return;
    }
    Status status = dsStateClient->ShutDown();
    if (!status.IsOk()) {
        YRLOG_WARN("DSCacheStateStore Shutdown fail. Status code: {}, Msg: {}", fmt::underlying(status.GetCode()),
                   status.ToString());
    }
    isInit = false;
}

ErrorInfo DSCacheStateStore::UpdateToken(datasystem::SensitiveValue token)
{
    if (!isInit) {
        return ErrorInfo();
    }
    ErrorInfo err;
    return err;
}

ErrorInfo DSCacheStateStore::UpdateAkSk(std::string ak, datasystem::SensitiveValue sk)
{
    ErrorInfo err;
    return err;
}

ErrorInfo DSCacheStateStore::HealthCheck()
{
    STATE_STORE_INIT_ONCE();
    auto status = dsStateClient->HealthCheck();
    if (!status.OK()) {
        return ErrorInfo(ErrorCode::ERR_CONNECTION_FAILED, "datasystem client is not healthy");
    }
    return ErrorInfo();
}
}  // namespace Libruntime
}  // namespace YR