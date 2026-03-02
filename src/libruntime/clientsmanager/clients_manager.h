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

#include <grpcpp/grpcpp.h>

#include "src/libruntime/err_type.h"
#include "src/libruntime/gwclient/http/client_manager.h"
#include "src/libruntime/heterostore/datasystem_hetero_store.h"
#include "src/libruntime/heterostore/hetero_store.h"
#include "src/libruntime/objectstore/datasystem_object_store.h"
#include "src/libruntime/objectstore/object_store.h"
#include "src/libruntime/statestore/datasystem_state_store.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/libruntime/streamstore/datasystem_stream_store.h"
#include "src/libruntime/streamstore/stream_store.h"
#include "src/libruntime/utils/security.h"
#include "src/libruntime/utils/utils.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
const uint32_t WAIT_FOR_STAGE_CHANGE_TIMEOUT_SEC = 5;
const int32_t RECONNECT_BACKOFF_INTERVAL = 100;
const int32_t MAX_RECONNECT_BACKOFF_INTERVAL = 5 * 1000;
const uint32_t SIZE_MEGA_BYTES = 1024 * 1024;
const int32_t DEFAULT_MAX_GRPC_SIZE = 10;  // MB
const int32_t GRPC_KEEPALIVE_TIME_MS = 600000;
const int32_t GRPC_KEEPALIVE_TIMEOUT_MS = 5000;
const std::string IP_PORT_REGEX = R"(((\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5})))";

struct DatasystemClients {
    std::shared_ptr<ObjectStore> dsObjectStore;
    std::shared_ptr<StateStore> dsStateStore;
    std::shared_ptr<StreamStore> dsStreamStore;
    std::shared_ptr<HeteroStore> dsHeteroStore;
};

class ClientsManager {
public:
    ClientsManager() = default;

    std::pair<std::shared_ptr<grpc::Channel>, ErrorInfo> NewFsConn(const std::string &ip, int port,
                                                                   std::shared_ptr<Security> security,
                                                                   const std::string &dstInstance,
                                                                   bool isKeepAlive = true);

    std::pair<std::shared_ptr<grpc::Channel>, ErrorInfo> GetFsConn(const std::string &ip, int port,
                                                                   const std::string &dstInstance);

    ErrorInfo ReleaseFsConn(const std::string &ip, int port, const std::string &dstInstance);

    std::pair<DatasystemClients, ErrorInfo> GetOrNewDsClient(const std::shared_ptr<LibruntimeConfig> librtCfg,
                                                             const std::string &ak,
                                                             const datasystem::SensitiveValue &sk,
                                                             std::int32_t connectTimeout);

    ErrorInfo ReleaseDsClient(const std::string &ip, int port);

    std::pair<std::shared_ptr<ClientManager>, ErrorInfo> GetOrNewHttpClient(
        const std::string &ip, int port, const std::shared_ptr<LibruntimeConfig> &librtCfg);

    ErrorInfo ReleaseHttpClient(const std::string &ip, int port);
    std::pair<std::shared_ptr<grpc::Channel>, ErrorInfo> InitFunctionSystemConn(std::string target,
                                                                                std::shared_ptr<Security> security,
                                                                                bool isKeepAlive);

private:
    std::pair<DatasystemClients, ErrorInfo> InitDatasystemClient(
        const std::string &ip, int port, bool enableDsAuth, bool encryptEnable, const std::string &runtimePublicKey,
        const datasystem::SensitiveValue &runtimePrivateKey, const std::string &dsPublicKey,
        const datasystem::SensitiveValue &token, const std::string &ak, const datasystem::SensitiveValue &sk,
        std::int32_t connectTimeout);

    std::pair<std::shared_ptr<ClientManager>, ErrorInfo> InitHttpClient(
        const std::string &ip, int port, const std::shared_ptr<LibruntimeConfig> &librtCfg);

    std::unordered_map<std::string, std::shared_ptr<grpc::Channel>> fsConns;

    std::unordered_map<std::string, DatasystemClients> dsClients;

    std::unordered_map<std::string, std::shared_ptr<ClientManager>> httpClients;

    std::unordered_map<std::string, int> fsConnsReferCounter;

    std::unordered_map<std::string, int> dsClientsReferCounter;

    std::unordered_map<std::string, int> httpClientsReferCounter;

    mutable std::mutex fsConnsMtx;  // Protects the following data:
                                    // - `fsConns` the map of grpc channel
                                    // - `fsConnsReferCounter`

    mutable std::mutex dsClientsMtx;

    mutable std::mutex httpClientsMtx;
};
}  // namespace Libruntime
}  // namespace YR
