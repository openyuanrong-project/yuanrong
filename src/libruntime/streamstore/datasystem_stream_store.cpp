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

#include "datasystem_stream_store.h"

#include <utility>
#include "src/libruntime/err_type.h"
#include "src/libruntime/statestore/datasystem_state_store.h"
#include "src/libruntime/utils/datasystem_utils.h"
#include "src/libruntime/utils/exception.h"
#include "src/libruntime/utils/utils.h"

namespace YR {
namespace Libruntime {
using namespace datasystem;
using datasystem::Status;
using YR::Libruntime::ErrorCode;
using YR::Libruntime::ModuleCode;
ErrorInfo DatasystemStreamStore::Init(const std::string &ip, int port)
{
    return this->Init(ip, port, false, false, "", datasystem::SensitiveValue{}, "", datasystem::SensitiveValue{}, "",
                      datasystem::SensitiveValue{});
}

ErrorInfo DatasystemStreamStore::Init(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                                      const std::string &runtimePublicKey,
                                      const datasystem::SensitiveValue &runtimePrivateKey,
                                      const std::string &dsPublicKey, const datasystem::SensitiveValue &token,
                                      const std::string &ak, const datasystem::SensitiveValue &sk)
{
    this->ip = ip;
    this->port = port;
    this->enableDsAuth = enableDsAuth;
    this->encryptEnable = encryptEnable;
    this->runtimePublicKey = runtimePublicKey;
    this->runtimePrivateKey = runtimePrivateKey;
    this->dsPublicKey = dsPublicKey;
    this->ak = ak;
    this->sk = sk;
    this->token = token;
    return ErrorInfo();
}

ErrorInfo DatasystemStreamStore::Init(datasystem::ConnectOptions &inputConnOpt, std::shared_ptr<StateStore> stateStore)
{
    this->dsStateStore = stateStore;
    return Init(inputConnOpt);
}

ErrorInfo DatasystemStreamStore::Init(datasystem::ConnectOptions &inputConnOpt)
{
    this->connectOpts.host = inputConnOpt.host;
    this->connectOpts.port = inputConnOpt.port;
    this->connectOpts.clientPublicKey = inputConnOpt.clientPublicKey;
    this->connectOpts.clientPrivateKey = inputConnOpt.clientPrivateKey;
    this->connectOpts.serverPublicKey = inputConnOpt.serverPublicKey;
    this->connectOpts.accessKey = inputConnOpt.accessKey;
    this->connectOpts.secretKey = inputConnOpt.secretKey;
    this->connectOpts.token = inputConnOpt.token;
    this->connectOpts.connectTimeoutMs = inputConnOpt.connectTimeoutMs;
    this->connectOpts.tenantId = inputConnOpt.tenantId;
    return ErrorInfo();
}

ErrorInfo DatasystemStreamStore::DoInitOnce()
{
    streamClient = std::make_shared<StreamClient>(this->connectOpts);
    Status status = streamClient->Init();
    std::string msg = "failed to init stream client, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    isInit = true;
    if (Config::Instance().ENABLE_DS_HEALTH_CHECK()) {
        YRLOG_INFO("ds client heath check is enabled");
        if (dsStateStore == nullptr) {
            YRLOG_ERROR("ds state client for stream client healthy check is null, run without healthy check");
            return ErrorInfo();
        }
        return dsStateStore->StartHealthCheck();
    }
    return ErrorInfo();
}

void DatasystemStreamStore::InitOnce(void)
{
    std::call_once(this->initFlag, [this]() { this->initErr = this->DoInitOnce(); });
}

std::pair<datasystem::ProducerConf, ErrorInfo> DatasystemStreamStore::CheckAndBuildProducerConf(
    const ProducerConf &producerConf)
{
    datasystem::ProducerConf dsConf;
    ErrorInfo errInfo;
    dsConf.delayFlushTime = producerConf.delayFlushTime;
    dsConf.pageSize = producerConf.pageSize;
    dsConf.maxStreamSize = producerConf.maxStreamSize;
    dsConf.autoCleanup = producerConf.autoCleanup;
    dsConf.encryptStream = producerConf.encryptStream;
    dsConf.retainForNumConsumers = producerConf.retainForNumConsumers;
    dsConf.reserveSize = producerConf.reserveSize;
    auto it = producerConf.extendConfig.find(std::string(STREAM_MODE));
    if (it != producerConf.extendConfig.end()) {
        auto mode = it->second;
        auto modeIt = streamModeMap.find(mode);
        if (modeIt != streamModeMap.end()) {
            dsConf.streamMode = modeIt->second;
        } else {
            errInfo = ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                                "unsupported stream mode: " + mode + ", only support MPMC、MPSC and SPSC.");
        }
    }
    return {dsConf, errInfo};
}

ErrorInfo DatasystemStreamStore::CreateStreamProducer(const std::string &streamName,
                                                      std::shared_ptr<StreamProducer> &producer,
                                                      ProducerConf producerConf)
{
    if (producer == nullptr) {
        return ErrorInfo(ERR_PARAM_INVALID, RUNTIME,
                         "check the second param of YR::CreateProducer interface, nullptr is ont supported.");
    }
    auto [dsConf, err] = CheckAndBuildProducerConf(producerConf);
    if (!err.OK()) {
        return err;
    }
    err = EnsureInit();
    if (!err.OK()) {
        return err;
    }
    Status status = streamClient->CreateProducer(streamName, producer->GetProducer(), dsConf);
    auto msg = "failed to CreateProducer, errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo DatasystemStreamStore::CreateStreamConsumer(const std::string &streamName, const SubscriptionConfig &config,
                                                      std::shared_ptr<StreamConsumer> &consumer, bool autoAck)
{
    if (consumer == nullptr) {
        return ErrorInfo(ERR_PARAM_INVALID, RUNTIME,
                         "check the third param of YR::Subscribe interface, nullptr is ont supported.");
    }
    auto err = EnsureInit();
    if (!err.OK()) {
        return err;
    }
    datasystem::SubscriptionConfig dsConfig(config.subscriptionName, typeMap[config.subscriptionType]);
    Status status = streamClient->Subscribe(streamName, dsConfig, consumer->GetConsumer(), autoAck);
    auto msg = "failed to Subscribe, streamName: " + streamName + " , errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo DatasystemStreamStore::DeleteStream(const std::string &streamName)
{
    if (auto err = EnsureInit(); !err.OK()) {
        return err;
    }
    Status status = streamClient->DeleteStream(streamName);
    auto msg = "failed to DeleteStream, streamName: " + streamName + " , errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo DatasystemStreamStore::QueryGlobalProducersNum(const std::string &streamName, uint64_t &gProducerNum)
{
    if (auto err = EnsureInit(); !err.OK()) {
        return err;
    }
    Status status = streamClient->QueryGlobalProducersNum(streamName, gProducerNum);
    auto msg = "failed to QueryGlobalProducersNum, streamName: " + streamName + " , errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo DatasystemStreamStore::QueryGlobalConsumersNum(const std::string &streamName, uint64_t &gConsumerNum)
{
    if (auto err = EnsureInit(); !err.OK()) {
        return err;
    }
    Status status = streamClient->QueryGlobalConsumersNum(streamName, gConsumerNum);
    auto msg = "failed to QueryGlobalConsumersNum, streamName: " + streamName + " , errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

void DatasystemStreamStore::Shutdown()
{
    if (streamClient == nullptr) {
        return;
    }
    Status status = streamClient->ShutDown();
    if (!status.IsOk()) {
        YRLOG_WARN("DatasystemStreamStore Shutdown fail. Status code: {}, Msg: {}", fmt::underlying(status.GetCode()),
                   status.ToString());
    }
}

ErrorInfo DatasystemStreamStore::UpdateToken(datasystem::SensitiveValue token)
{
    ErrorInfo err;
    YRLOG_DEBUG("UpdateToken, token {}", !token.Empty());
    if (streamClient == nullptr) {
        this->connectOpts.token = token;
        return err;
    }
    Status status = streamClient->UpdateToken(token);
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                      "stream client update token failed");
    return err;
}

ErrorInfo DatasystemStreamStore::UpdateAkSk(std::string ak, datasystem::SensitiveValue sk)
{
    ErrorInfo err;
    return err;
}
ErrorInfo DatasystemStreamStore::EnsureInit()
{
    std::unique_lock<std::mutex> lock(initMutex);
    if (isInit && initErr.OK()) {
        return initErr;
    }
    initErr = DoInitOnce();
    isInit = true;
    return initErr;
}
}  // namespace Libruntime
}  // namespace YR
