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

#include "datasystem/stream_client.h"
#include "src/dto/stream_conf.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/libruntime/utils/constants.h"
#include "src/utility/logger/logger.h"
#include "stream_store.h"

namespace YR {
namespace Libruntime {
class DatasystemStreamStore : public StreamStore {
public:
    ErrorInfo Init(const std::string &ip, int port) override;

    ErrorInfo Init(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                   const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                   const std::string &dsPublicKey, const datasystem::SensitiveValue &token, const std::string &ak,
                   const datasystem::SensitiveValue &sk) override;

    ErrorInfo Init(datasystem::ConnectOptions &inputConnOpt) override;

    ErrorInfo Init(datasystem::ConnectOptions &inputConnOpt, std::shared_ptr<StateStore> dsStateStore) override;

    ErrorInfo CreateStreamProducer(const std::string &streamName, std::shared_ptr<StreamProducer> &producer,
                                   ProducerConf producerConf = {}) override;

    ErrorInfo CreateStreamConsumer(const std::string &streamName, const SubscriptionConfig &config,
                                   std::shared_ptr<StreamConsumer> &consumer, bool autoAck = false) override;

    ErrorInfo DeleteStream(const std::string &streamName) override;

    ErrorInfo QueryGlobalProducersNum(const std::string &streamName, uint64_t &gProducerNum) override;

    ErrorInfo QueryGlobalConsumersNum(const std::string &streamName, uint64_t &gConsumerNum) override;

    void Shutdown() override;

    ErrorInfo UpdateToken(datasystem::SensitiveValue token) override;

    ErrorInfo UpdateAkSk(std::string ak, datasystem::SensitiveValue sk) override;

private:
    void InitOnce(void);
    ErrorInfo DoInitOnce(void);

    std::pair<datasystem::ProducerConf, ErrorInfo> CheckAndBuildProducerConf(const ProducerConf &producerConf);

    bool isInit = false;
    std::mutex initMutex;
    std::once_flag initFlag;
    ErrorInfo initErr;
    bool isReady = false;
    std::shared_ptr<datasystem::StreamClient> streamClient;
    std::string ip;
    int port;
    bool enableDsAuth = false;
    bool encryptEnable = false;
    std::string runtimePublicKey;
    datasystem::SensitiveValue runtimePrivateKey;
    std::string dsPublicKey;
    std::string ak;
    datasystem::SensitiveValue sk;
    datasystem::SensitiveValue token;
    datasystem::ConnectOptions connectOpts;
    std::unordered_map<libruntime::SubscriptionType, datasystem::SubscriptionType> typeMap = {
        {libruntime::SubscriptionType::STREAM, datasystem::SubscriptionType::STREAM},
        {libruntime::SubscriptionType::KEY_PARTITIONS, datasystem::SubscriptionType::KEY_PARTITIONS},
        {libruntime::SubscriptionType::ROUND_ROBIN, datasystem::SubscriptionType::ROUND_ROBIN},
        {libruntime::SubscriptionType::UNKNOWN, datasystem::SubscriptionType::UNKNOWN}};
    std::unordered_map<std::string, datasystem::StreamMode> streamModeMap = {
        {std::string(MPMC), datasystem::StreamMode::MPMC},
        {std::string(MPSC), datasystem::StreamMode::MPSC},
        {std::string(SPSC), datasystem::StreamMode::SPSC}};
    std::shared_ptr<StateStore> dsStateStore;

    ErrorInfo EnsureInit(void);
};

#define STREAM_STORE_INIT_ONCE()  \
    do {                          \
        InitOnce();               \
        if (!initErr.OK()) {      \
            return initErr;       \
        }                         \
    } while (0)

}  // namespace Libruntime
}  // namespace YR
