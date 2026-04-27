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

#include "clients_manager.h"
#include "src/libruntime/fsclient/grpc/posix_auth_interceptor.h"

namespace YR {
namespace Libruntime {

std::pair<std::shared_ptr<grpc::Channel>, ErrorInfo> ClientsManager::GetFsConn(const std::string &ip, int port,
                                                                               const std::string &dstInstance)
{
    auto addr = GetIpAddr(ip, port);
    auto connKey = dstInstance + ":" + addr;
    YRLOG_DEBUG("grpc client target is {}", connKey);
    if (!RE2::FullMatch(addr, re2::RE2(IP_PORT_REGEX))) {
        YRLOG_ERROR("failed to get valid runtime-rpc server address({})", addr);
        return std::make_pair(nullptr, ErrorInfo(ErrorCode::ERR_CONNECTION_FAILED, "The server address is invalid."));
    }
    std::lock_guard<std::mutex> fsConnsLock(fsConnsMtx);
    auto iter = fsConns.find(connKey);
    if (iter != fsConns.end()) {
        fsConnsReferCounter[connKey]++;
        return std::make_pair(iter->second, ErrorInfo());
    }
    return std::make_pair(nullptr, ErrorInfo());
}

std::pair<std::shared_ptr<grpc::Channel>, ErrorInfo> ClientsManager::NewFsConn(const std::string &ip, int port,
                                                                               std::shared_ptr<Security> security,
                                                                               const std::string &dstInstance,
                                                                               bool isKeepAlive)
{
    auto addr = GetIpAddr(ip, port);
    bool uds = false;
    if (dstInstance == "function-proxy" && YR::Libruntime::Config::Instance().YR_DPOSIX_UDS() != "") {
        uds = true;
        addr = YR::utility::Join({YR::Libruntime::Config::Instance().YR_DPOSIX_UDS(), "fs.sock"}, "/");
        YRLOG_INFO("Use dposix uds path {} to connect function proxy", addr);
    }
    auto [res, error] = InitFunctionSystemConn(uds, addr, security, isKeepAlive);
    if (!error.OK()) {
        return std::make_pair(nullptr, error);
    }
    std::lock_guard<std::mutex> fsConnsLock(fsConnsMtx);
    auto connKey = dstInstance + ":" + addr;
    fsConns[connKey] = res;
    fsConnsReferCounter[connKey]++;
    return std::make_pair(res, ErrorInfo());
}

ErrorInfo ClientsManager::ReleaseFsConn(const std::string &ip, int port, const std::string &dstInstance)
{
    auto addr = GetIpAddr(ip, port);
    std::lock_guard<std::mutex> fsConnsLock(fsConnsMtx);
    auto connKey = dstInstance + ":" + addr;
    auto iter = fsConnsReferCounter.find(connKey);
    if (iter == fsConnsReferCounter.end()) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "Cannot find function system conn's ref count info.");
    }
    fsConnsReferCounter[connKey]--;
    if (fsConnsReferCounter[connKey] == 0) {
        fsConnsReferCounter.erase(connKey);
        fsConns.erase(connKey);
    }
    return ErrorInfo();
}

std::pair<DatasystemClients, ErrorInfo> ClientsManager::GetOrNewDsClient(
    const std::shared_ptr<LibruntimeConfig> librtCfg, const std::string &ak, const datasystem::SensitiveValue &sk,
    std::int32_t connectTimeout)
{
    auto key = GetIpAddr(librtCfg->dataSystemIpAddr, librtCfg->dataSystemPort);
    std::lock_guard<std::mutex> dsClientsLock(dsClientsMtx);
    auto iter = dsClients.find(key);
    if (iter != dsClients.end()) {
        dsClientsReferCounter[key]++;
        return std::make_pair(iter->second, ErrorInfo());
    }
    auto res = InitDatasystemClient(librtCfg->dataSystemIpAddr, librtCfg->dataSystemPort, librtCfg->enableAuth,
                                    librtCfg->encryptEnable, librtCfg->runtimePublicKey, librtCfg->runtimePrivateKey,
                                    librtCfg->dsPublicKey, librtCfg->token, ak, sk, connectTimeout);
    if (res.second.OK()) {
        dsClients[key] = res.first;
        dsClientsReferCounter[key]++;
    }
    return res;
}

ErrorInfo ClientsManager::ReleaseDsClient(const std::string &ip, int port)
{
    auto key = GetIpAddr(ip, port);
    std::lock_guard<std::mutex> dsClientsLock(dsClientsMtx);
    auto iter = dsClientsReferCounter.find(key);
    if (iter == dsClientsReferCounter.end()) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "Cannot find datasystem client's ref count info.");
    }
    dsClientsReferCounter[key]--;
    if (dsClientsReferCounter[key] == 0) {
        dsClientsReferCounter.erase(key);
        if (dsClients[key].dsObjectStore != nullptr) {
            dsClients[key].dsObjectStore->Clear();
            dsClients[key].dsObjectStore->Shutdown();
            YRLOG_DEBUG("Shutdown object store clients");
        }
        if (dsClients[key].dsStateStore != nullptr) {
            dsClients[key].dsStateStore->Shutdown();
            YRLOG_DEBUG("Shutdown state store clients");
        }
        if (dsClients[key].dsStreamStore != nullptr) {
            dsClients[key].dsStreamStore->Shutdown();
            YRLOG_DEBUG("Shutdown stream store clients");
        }
        if (dsClients[key].dsHeteroStore != nullptr) {
            dsClients[key].dsHeteroStore->Shutdown();
            YRLOG_DEBUG("Shutdown hetero store clients");
        }
        dsClients.erase(key);
    }
    return ErrorInfo();
}

std::pair<std::shared_ptr<ClientManager>, ErrorInfo> ClientsManager::GetOrNewHttpClient(
    const std::string &ip, int port, const std::shared_ptr<LibruntimeConfig> &librtCfg)
{
    auto addr = GetIpAddr(ip, port);
    std::lock_guard<std::mutex> httpClientsLock(httpClientsMtx);
    auto iter = httpClients.find(addr);
    if (iter != httpClients.end()) {
        httpClientsReferCounter[addr]++;
        return std::make_pair(iter->second, ErrorInfo());
    }
    auto res = InitHttpClient(ip, port, librtCfg);
    if (res.second.OK()) {
        httpClients[addr] = res.first;
        httpClientsReferCounter[addr]++;
    }
    return res;
}

ErrorInfo ClientsManager::ReleaseHttpClient(const std::string &ip, int port)
{
    auto addr = GetIpAddr(ip, port);
    std::lock_guard<std::mutex> httpClientsLock(httpClientsMtx);
    auto iter = httpClientsReferCounter.find(addr);
    if (iter == httpClientsReferCounter.end()) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "Cannot find http client's ref count info.");
    }
    httpClientsReferCounter[addr]--;
    if (httpClientsReferCounter[addr] == 0) {
        httpClientsReferCounter.erase(addr);
        if (auto it = httpClients.find(addr); it != httpClients.end()) {
            it->second->Stop();
            httpClients.erase(addr);
        }
    }
    return ErrorInfo();
}

std::pair<std::shared_ptr<grpc::Channel>, ErrorInfo> ClientsManager::InitFunctionSystemConn(
    bool uds, std::string target, std::shared_ptr<Security> security, bool isKeepAlive)
{
    grpc::ChannelArguments args;
    std::shared_ptr<grpc::Channel> channel;
    ErrorInfo err;
    uint32_t maxGrpcSize = Config::Instance().YR_MAX_GRPC_SIZE() * SIZE_MEGA_BYTES;
    args.SetInt(GRPC_ARG_INITIAL_RECONNECT_BACKOFF_MS, RECONNECT_BACKOFF_INTERVAL);
    args.SetInt(GRPC_ARG_MIN_RECONNECT_BACKOFF_MS, RECONNECT_BACKOFF_INTERVAL);
    args.SetInt(GRPC_ARG_MAX_RECONNECT_BACKOFF_MS, MAX_RECONNECT_BACKOFF_INTERVAL);
    args.SetInt(GRPC_ARG_MAX_RECEIVE_MESSAGE_LENGTH, maxGrpcSize);
    args.SetInt(GRPC_ARG_MAX_SEND_MESSAGE_LENGTH, maxGrpcSize);
    args.SetInt(GRPC_ARG_ENABLE_HTTP_PROXY, Config::Instance().YR_ENABLE_HTTP_PROXY() ? 1 : 0);
    if (!isKeepAlive) {
        args.SetInt(GRPC_ARG_KEEPALIVE_TIME_MS, GRPC_KEEPALIVE_TIME_MS);
        args.SetInt(GRPC_ARG_KEEPALIVE_TIMEOUT_MS, GRPC_KEEPALIVE_TIMEOUT_MS);
    }
    if (security != nullptr) {
        std::string serverNameOverride;
        (void)security->GetFunctionSystemConnectionMode(serverNameOverride);
        if (!serverNameOverride.empty()) {
            args.SetSslTargetNameOverride(serverNameOverride);
        }
    }
    try {
        std::string prefix = uds ? "unix://" : "ipv4:///";
        if (security != nullptr && security->IsFsAuthEnable()) {
            std::vector<std::unique_ptr<grpc::experimental::ClientInterceptorFactoryInterface>> interCeptorCreators;
            auto interCeptorCreator = new PosixClientAuthInterceptorFactory();
            interCeptorCreator->RegisterSecurity(security);
            interCeptorCreators.push_back(std::unique_ptr<PosixClientAuthInterceptorFactory>(interCeptorCreator));
            channel = grpc::experimental::CreateCustomChannelWithInterceptors(
                prefix + target, YR::GetChannelCreds(security), args, std::move(interCeptorCreators));
        } else {
            channel = grpc::CreateCustomChannel(prefix + target, YR::GetChannelCreds(security), args);
        }
        auto tmout = gpr_time_add(gpr_now(GPR_CLOCK_MONOTONIC), {WAIT_FOR_STAGE_CHANGE_TIMEOUT_SEC, 0, GPR_TIMESPAN});
        auto isConnect = channel->WaitForConnected(tmout);
        auto state = channel->GetState(true);
        if (!isConnect) {
            YRLOG_ERROR("failed to connect to grpc server {}, channel state: {}", target, fmt::underlying(state));
            return std::make_pair(nullptr,
                                  ErrorInfo(ErrorCode::ERR_CONNECTION_FAILED, "failed to connect to grpc server"));
        }
        return std::make_pair(channel, ErrorInfo());
    } catch (std::exception &e) {
        err.SetErrorCode(ErrorCode::ERR_CONNECTION_FAILED);
        err.SetErrorMsg(e.what());
    }
    return std::make_pair(nullptr, err);
}

std::pair<DatasystemClients, ErrorInfo> ClientsManager::InitDatasystemClient(
    const std::string &ip, int port, bool enableDsAuth, bool encryptEnable, const std::string &runtimePublicKey,
    const datasystem::SensitiveValue &runtimePrivateKey, const std::string &dsPublicKey,
    const datasystem::SensitiveValue &token, const std::string &ak, const datasystem::SensitiveValue &sk,
    std::int32_t connectTimeout)
{
    datasystem::ConnectOptions connectOptions;
    connectOptions.host = ip;
    connectOptions.port = port;
    connectOptions.connectTimeoutMs = connectTimeout * S_TO_MS;
    if (encryptEnable) {
        connectOptions.clientPublicKey = runtimePublicKey;
        connectOptions.clientPrivateKey = runtimePrivateKey;
        connectOptions.serverPublicKey = dsPublicKey;
    }
    if (enableDsAuth) {
        if (!ak.empty() && !sk.Empty()) {
            connectOptions.accessKey = ak;
            connectOptions.secretKey = sk;
        } else {
                connectOptions.token = token;
        }
    }
    std::string tenantId = Config::Instance().YR_TENANT_ID();
    if (!tenantId.empty()) {
        connectOptions.tenantId = tenantId;
    }
    YRLOG_DEBUG(
        "start init datasystem client connect param, tenant id is {}, ip is {}, port is {}, enableDsAuth is {}, "
        "encryptEnableis {}, runtimePublicKey is empty {}, ak is empty {}, token is empty {}, timeout is {}",
        tenantId, ip, port, enableDsAuth, encryptEnable, runtimePublicKey.empty(), ak.empty(), token.Empty(),
        connectTimeout);
    DatasystemClients clients;
    clients.dsObjectStore = std::make_shared<DSCacheObjectStore>();
    ErrorInfo infoObjectStore = clients.dsObjectStore->Init(connectOptions);
    if (!infoObjectStore.OK()) {
        return std::make_pair(clients, infoObjectStore);
    }

    clients.dsStateStore = std::make_shared<DSCacheStateStore>();
    ErrorInfo infoStateStore = clients.dsStateStore->Init(connectOptions);
    if (!infoStateStore.OK()) {
        return std::make_pair(clients, infoObjectStore);
    }

    clients.dsStreamStore = std::make_shared<DatasystemStreamStore>();
    ErrorInfo infoStreamStore = clients.dsStreamStore->Init(connectOptions, clients.dsStateStore);
    if (!infoStreamStore.OK()) {
        return std::make_pair(clients, infoStreamStore);
    }

    clients.dsHeteroStore = std::make_shared<DatasystemHeteroStore>();
    auto infoHeteroStore = clients.dsHeteroStore->Init(connectOptions);
    if (!infoHeteroStore.OK()) {
        return std::make_pair(clients, infoHeteroStore);
    }

    return std::make_pair(clients, ErrorInfo());
}

std::pair<std::shared_ptr<ClientManager>, ErrorInfo> ClientsManager::InitHttpClient(
    const std::string &ip, int port, const std::shared_ptr<LibruntimeConfig> &config)
{
    auto httpClient = std::make_shared<ClientManager>(config);
    ErrorInfo error = httpClient->Init(ConnectionParam{ip, std::to_string(port), config->httpIdleTime});
    if (!error.OK()) {
        return std::make_pair(nullptr, error);
    }
    return std::make_pair(httpClient, error);
}
}  // namespace Libruntime
}  // namespace YR
