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
#include <optional>
#include <string>
#include "datasystem/utils/sensitive_value.h"
#include "securec.h"
#include "src/dto/config.h"
#include "src/dto/function_group_running_info.h"
#include "src/libruntime/libruntime_options.h"
#include "src/proto/libruntime.pb.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
using SensitiveValue = datasystem::SensitiveValue;

const int DEFAULT_RECYCLETIME = 2;
const int MAX_RECYCLETIME = 3000;
const int MIN_RECYCLETIME = 1;
const int MAX_PASSWD_LENGTH = 100;
const int HTTP_IDLE_TIME = 120;
std::pair<uint32_t, ErrorInfo> GetValidMaxLogFileNum(uint32_t maxLogFileNum);
std::pair<uint32_t, ErrorInfo> GetValidMaxLogSizeMb(uint32_t maxLogSizeMb);
struct LibruntimeConfig {
    void InitConfig(const libruntime::MetaConfig &config)
    {
        this->jobId = config.jobid();
        this->recycleTime = config.recycletime();
        this->maxTaskInstanceNum = config.maxtaskinstancenum();
        this->maxConcurrencyCreateNum = config.maxconcurrencycreatenum();
        this->enableMetrics = config.enablemetrics();
        this->threadPoolSize = config.threadpoolsize();
        this->localThreadPoolSize = config.localthreadpoolsize();
        this->ns = config.ns();
        this->tenantId = config.tenantid();
        for (int i = 0; i < config.functionids_size(); i++) {
            this->functionIds[config.functionids(i).language()] = config.functionids(i).functionid();
        }
        this->loadPaths.resize(config.codepaths_size());
        for (int i = 0; i < config.codepaths_size(); i++) {
            this->loadPaths[i] = config.codepaths(i);
        }
        this->schedulerInstanceIds.resize(config.schedulerinstanceids_size());
        for (int i = 0; i < config.schedulerinstanceids_size(); i++) {
            this->schedulerInstanceIds[i] = config.schedulerinstanceids(i);
        }
        for (const auto &pair : config.customenvs()) {
            this->customEnvs[pair.first] = pair.second;
        }
        this->functionMasters.resize(config.functionmasters_size());
        for (int i = 0; i < config.functionmasters_size(); i++) {
            this->functionMasters[i] = config.functionmasters(i);
        }
        this->isLowReliabilityTask = config.islowreliabilitytask();
        auto tenantId = Config::Instance().YR_TENANT_ID();
        if (!tenantId.empty()) {
            this->tenantId = tenantId;
        }
        this->funcMeta = config.funcmeta();
    }

    void BuildMetaConfig(libruntime::MetaConfig &config) const
    {
        config.set_jobid(jobId);
        config.set_recycletime(recycleTime);
        config.set_maxtaskinstancenum(maxTaskInstanceNum);
        config.set_maxconcurrencycreatenum(maxConcurrencyCreateNum);
        config.set_enablemetrics(enableMetrics);
        config.set_threadpoolsize(threadPoolSize);
        config.set_localthreadpoolsize(localThreadPoolSize);
        config.set_ns(ns);
        config.set_tenantid(tenantId);
        for (const auto &path : loadPaths) {
            config.add_codepaths(path);
        }
        for (const auto &addr : functionMasters) {
            config.add_functionmasters(addr);
        }
        for (const auto &[key, value] : functionIds) {
            auto funcId = config.add_functionids();
            funcId->set_language(key);
            funcId->set_functionid(value);
        }
        for (const auto &pair : customEnvs) {
            config.mutable_customenvs()->insert({pair.first, pair.second});
        }
        config.set_islowreliabilitytask(isLowReliabilityTask);
        config.mutable_funcmeta()->CopyFrom(funcMeta);
    }

    ErrorInfo MergeConfig(const LibruntimeConfig &config)
    {
        this->jobId = config.jobId;
        this->recycleTime = config.recycleTime;
        this->maxTaskInstanceNum = config.maxTaskInstanceNum;
        this->enableMetrics = config.enableMetrics;
        if (!config.functionMasters.empty()) {
            this->functionMasters = config.functionMasters;
        }
        this->threadPoolSize = config.threadPoolSize;
        this->localThreadPoolSize = config.localThreadPoolSize;
        this->ns = config.ns;
        this->functionIds = config.functionIds;
        this->loadPaths = config.loadPaths;
        this->selfLanguage = config.selfLanguage;
        this->isLowReliabilityTask = config.isLowReliabilityTask;
        return ErrorInfo();
    }

    void InitFunctionGroupRunningInfo(const common::FunctionGroupRunningInfo &runningInfo)
    {
        if (runningInfo.serverlist_size() == 0) {
            return;
        }
        this->groupRunningInfo.deviceName = runningInfo.devicename();
        this->groupRunningInfo.worldSize = runningInfo.worldsize();
        this->groupRunningInfo.instanceRankId = runningInfo.instancerankid();
        this->groupRunningInfo.serverList.resize(runningInfo.serverlist_size());
        for (int i = 0; i < runningInfo.serverlist_size(); i++) {
            auto &serverInfo = runningInfo.serverlist(i);
            this->groupRunningInfo.serverList[i].serverId = serverInfo.serverid();
            this->groupRunningInfo.serverList[i].devices.resize(serverInfo.devices_size());
            for (int j = 0; j < serverInfo.devices_size(); j++) {
                auto &device = serverInfo.devices(j);
                this->groupRunningInfo.serverList[i].devices[j].deviceId = device.deviceid();
                this->groupRunningInfo.serverList[i].devices[j].deviceIp = device.deviceip();
                this->groupRunningInfo.serverList[i].devices[j].rankId = device.rankid();
            }
        }
    }

    ErrorInfo Check() const
    {
        if (this->recycleTime < MIN_RECYCLETIME || this->recycleTime > MAX_RECYCLETIME) {
            std::stringstream ss;
            ss << "invalid recycle_time value, expect " << MIN_RECYCLETIME << " <= time <= " << MAX_RECYCLETIME
               << ", actual " << this->recycleTime;
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                             ss.str());
        }
        return ErrorInfo();
    }

    const std::string GetInstanceId()
    {
        if (funcMeta.ns().empty()) {
            return std::string(DEFAULT_YR_NAMESPACE + "-" + funcMeta.name());
        }
        return std::string(funcMeta.ns() + "-" + funcMeta.name());
    }

    void ClearPaaswd()
    {
        memset_s(privateKeyPaaswd, MAX_PASSWD_LENGTH, 0, MAX_PASSWD_LENGTH);
    }
    ErrorInfo Decrypt();
    // functionSystemIpAddr is an IP address of function system server that used to discover driver.
    // functionSystemIpAddr and functionSystemPort is used by driver process
    std::string functionSystemIpAddr = "";

    // functionSystemPort is the coresponding port of functionSystemIpAddr of function system server
    int functionSystemPort = 0;

    // functionSystemRtServerIpAddr is IP address of runtime server listening to function system's connection
    // functionSystemRtServerIpAddr and functionSystemRtServerPort is used by runtime process that pulled up by function
    // system
    std::string functionSystemRtServerIpAddr = "";

    // functionSystemRtServerPort is the coresponding port of functionSystemRtServerIpAddr of runtime server
    int functionSystemRtServerPort = 0;

    std::string dataSystemIpAddr = "";
    int dataSystemPort = 0;

    // iamAdreess is address to request iamAdapter
    std::string iamAddress = "";

    // functionMasters is IP address list of function masters for get resources or other info from master
    std::vector<std::string> functionMasters;

    bool isDriver = false;
    std::string jobId = "";
    std::string runtimeId = "";
    std::string instanceId = "";
    std::string functionName = "";
    bool enableServerMode = false;

    libruntime::LanguageType selfLanguage;
    std::unordered_map<libruntime::LanguageType, std::string> functionIds;
    libruntime::ApiType selfApiType = libruntime::ApiType::Function;
    std::string logLevel = "";
    std::string logDir = ".";
    uint32_t logFileSizeMax = 0;
    uint32_t logFileNumMax = 0;
    int logFlushInterval = 1;
    bool isLogMerge = false;
    LibruntimeOptions libruntimeOptions;
    int recycleTime = 2;
    int maxTaskInstanceNum = -1;
    int maxConcurrencyCreateNum = 100;
    bool enableMetrics = true;
    uint32_t threadPoolSize = 0;
    uint32_t localThreadPoolSize = 0;
    std::vector<std::string> loadPaths;
    std::string tenantId = "";

    std::string metaConfig = "";  // deprecated? pyx use it.
    bool enableMTLS = false;
    bool enableTLS = false;
    std::string privateKeyPath = "";
    std::string certificateFilePath = "";
    std::string verifyFilePath = "";
    char privateKeyPaaswd[MAX_PASSWD_LENGTH] = {0};
    std::shared_ptr<void> tlsContext = nullptr;
    uint32_t httpIocThreadsNum = 200;
    int httpIdleTime = HTTP_IDLE_TIME;
    std::string serverName = "";
    bool skipServerVerify = true;  // Whether to skip server certificate verification
    bool inCluster = true;
    std::string ns = "";
    bool isLowReliabilityTask = false;
    bool attach = false;
    int32_t rpcTimeout = 30 * 60;  // 30min
    std::vector<std::string> schedulerInstanceIds;
    bool enableAuth = false;
    bool encryptEnable = false;
    std::string runtimePublicKeyPath = "";
    std::string runtimePrivateKeyPath = "";
    std::string dsPublicKeyPath = "";
    std::string runtimePublicKey = "";
    SensitiveValue runtimePrivateKey = "";
    std::string dsPublicKey = "";
    SensitiveValue token = "";
    std::string ak_ = "";
    SensitiveValue sk_ = "";
    SensitiveValue dk_ = "";

    using SubmitHook = std::function<void(std::function<void(void)> &&)>;
    SubmitHook funcExecSubmitHook = nullptr;
    std::unordered_map<std::string, std::string> customEnvs;
    std::string serverVersion;
    FunctionGroupRunningInfo groupRunningInfo;
    std::function<ErrorInfo()> checkSignals_ = nullptr;
    std::string workingDir;
    std::string rtCtx;
    bool logToDriver = false;
    bool dedupLogs = false;
    std::string primaryKeyStoreFile;
    std::string standbyKeyStoreFile;
    libruntime::FunctionMeta funcMeta;
    bool needOrder = false;
    bool enableSigaction = true;
    int64_t invokeTimeoutSec = 0;
    uint32_t maxConnSize = 10000;
    std::string nodeId;
    std::string nodeIp;
    std::string envFile = "";  // Path to environment variable file (.env format: KEY=VALUE, one per line)
    bool enableEvent = false;
    std::string authToken = "";

private:
    ErrorInfo ValidateKeyParams();
};
}  // namespace Libruntime
}  // namespace YR