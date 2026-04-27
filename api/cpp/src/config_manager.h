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

#include "securec.h"

#include "yr/api/client_info.h"
#include "yr/api/config.h"

namespace YR {
class ConfigManager {
public:
    static ConfigManager &Singleton() noexcept;

    ClientInfo GetClientInfo();

    void Init(const Config &conf, int argc, char **argv);

    bool IsLocalMode() const;

    void ClearPasswd();

    std::string jobId = "";

    std::string version = "";

    std::string runtimeId = "";

    bool enableServerMode = true;

    Config::Mode mode = Config::CLUSTER_MODE;

    std::vector<std::string> loadPaths;

    bool inCluster = true;

    std::string functionSystemAddr = "";

    std::string dataSystemAddr = "";

    std::string grpcAddress = "";

    std::string logDir = "";

    std::string logLevel = "INFO";

    bool logCompress;

    uint32_t maxLogFileNum;

    uint32_t maxLogFileSize;

    uint32_t logFlushInterval;

    uint32_t threadPoolSize = 0;

    uint32_t localThreadPoolSize = 10;

    uint32_t defaultGetTimeoutSec = 0;

    bool isDriver;

    int recycleTime;

    int maxTaskInstanceNum;

    std::string autoFunctionName;

    std::string functionId;

    std::string functionIdPython;

    std::string functionIdJava;

    bool enableMetrics;

    int maxConcurrencyCreateNum;

    bool enableMTLS = false;

    std::string privateKeyPath = "";

    std::string certificateFilePath = "";

    std::string verifyFilePath = "";

    char privateKeyPaaswd[MAX_PASSWD_LENGTH] = {0};

    bool enableDsAuth = false;

    bool enableDsEncrypt = false;

    std::string dsPublicKeyContextPath = "";

    std::string runtimePublicKeyContextPath = "";

    std::string runtimePrivateKeyContextPath = "";

    std::string primaryKeyStoreFile;

    std::string standbyKeyStoreFile;

    std::shared_ptr<void> tlsContext;

    uint32_t httpIocThreadsNum;

    std::string serverName = "";

    std::string ns = "";

    std::string tenantId = "";

    std::int32_t rpcTimeout = 30 * 60;  // 30min

    std::unordered_map<std::string, std::string> customEnvs;

    bool isLowReliabilityTask = false;

    bool attach = false;

    bool logToDriver = false;

    bool dedupLogs = true;

    bool launchUserBinary = false;

    std::string workingDir = "";
};
}  // namespace YR
