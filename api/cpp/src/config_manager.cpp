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

#include "api/cpp/src/config_manager.h"

#include <libgen.h>
#include <limits.h>
#include <unistd.h>
#include <iostream>
#include <thread>

#include "absl/flags/flag.h"
#include "absl/flags/parse.h"

#include "api/cpp/src/utils/utils.h"
#include "api/cpp/src/utils/version.h"
#include "src/dto/config.h"
#include "src/dto/invoke_options.h"
#include "src/libruntime/auto_init.h"
#include "src/libruntime/err_type.h"
#include "src/utility/id_generator.h"
#include "src/utility/string_utility.h"
#include "yr/api/exception.h"

ABSL_FLAG(std::string, logDir, "", "Log directory, default empty");
ABSL_FLAG(std::string, logLevel, "", "Log level, default empty");
ABSL_FLAG(uint32_t, logFlushInterval, 5, "Maximum log file size in MB, default 400");
ABSL_FLAG(std::string, grpcAddress, "", "Grpc address, default empty");
ABSL_FLAG(std::string, runtimeId, "", "Runtime id, default empty");
ABSL_FLAG(std::string, jobId, "", "Job id, default empty");
ABSL_FLAG(std::string, runtimeConfigPath, "/home/snuser/config/runtime.json",
          "Runtime config file path, default /home/snuser/config/runtime.json");
ABSL_FLAG(std::vector<std::string>, codePath, std::vector<std::string>(),
          "Comma-separated paths that Yuanrong will search for code packages, default to empty.");

namespace {
const int MAX_THREADPOOL_SIZE = 64;
const int MIN_THREADPOOL_SIZE = 1;
const int MAX_LOCAL_THREAD_POOL_SIZE = 300;
}  // namespace

namespace YR {
const int MIN_TASK_INS_NUM_LIMIT = 1;
const int NO_TASK_INS_NUM_LIMIT = -1;
const int MAX_LOAD_PATH_NUM_LIMIT = 1024;
const std::string DEFAULT_CPP_URN = "sn:cn:yrk:default:function:0-defaultservice-cpp:$latest";

int GetValidMaxTaskInstanceNum(int maxTaskInstanceNum)
{
    if (maxTaskInstanceNum < MIN_TASK_INS_NUM_LIMIT) {
        if (maxTaskInstanceNum != NO_TASK_INS_NUM_LIMIT) {
            std::cerr << "Config maxTaskInstanceNum is invalid; should be equal or greater than "
                      << MIN_TASK_INS_NUM_LIMIT;
        }
        return NO_TASK_INS_NUM_LIMIT;
    }
    return maxTaskInstanceNum;
}

std::vector<std::string> GetValidLoadPaths(const std::vector<std::string> &loadPaths)
{
    if (loadPaths.size() > MAX_LOAD_PATH_NUM_LIMIT) {
        std::cerr << "Config loadPaths is invalid; The number of loading paths should be <= " << MAX_LOAD_PATH_NUM_LIMIT
                  << std::endl;
        return std::vector<std::string>(loadPaths.cbegin(), loadPaths.cbegin() + MAX_LOAD_PATH_NUM_LIMIT);
    }
    return loadPaths;
}

int GetValidMaxConcurrencyCreateNum(int maxConcurrencyCreateNum)
{
    if (maxConcurrencyCreateNum <= 0) {
        std::cerr << "Config maxConcurrencyCreateNum is invalid; should be positive" << std::endl;
        throw YR::Exception(static_cast<int>(YR::Libruntime::ErrorCode::ERR_INCORRECT_INIT_USAGE),
                            static_cast<int>(YR::Libruntime::ModuleCode::RUNTIME),
                            "maxConcurrencyCreateNum is required to be > 0");
    }
    return maxConcurrencyCreateNum;
}

bool GetValidLogCompress(bool logCompress)
{
    if (logCompress != true) {  // higher priority than env
        return logCompress;
    }
    return (logCompress = YR::Libruntime::Config::Instance().YR_LOG_COMPRESS());
}

int GetValidLocalThreadPoolSize(int threadPoolSize)
{
    if (threadPoolSize > MAX_LOCAL_THREAD_POOL_SIZE || threadPoolSize < MIN_THREADPOOL_SIZE) {
        // default is the number of CPUs
        std::cerr << "Config localThreadPoolSize is invalid; the valid range is " << MIN_THREADPOOL_SIZE << " to "
                  << MAX_LOCAL_THREAD_POOL_SIZE << "; set to core number "
                  << static_cast<int>(std::thread::hardware_concurrency()) << " by default" << std::endl;
        return static_cast<int>(std::thread::hardware_concurrency());
    }
    return threadPoolSize;
}

int GetValidThreadPoolSize(int threadPoolSize)
{
    if (threadPoolSize > MAX_THREADPOOL_SIZE || threadPoolSize < MIN_THREADPOOL_SIZE) {
        // default is the number of CPUs
        std::cerr << "Config threadPoolSize is invalid; the valid range is " << MIN_THREADPOOL_SIZE << " to "
                  << MAX_THREADPOOL_SIZE << "; set to core number "
                  << static_cast<int>(std::thread::hardware_concurrency()) << " by default" << std::endl;
        return static_cast<int>(std::thread::hardware_concurrency());
    }
    return threadPoolSize;
}

std::string GetExecutablePath()
{
    char buffer[PATH_MAX];
    ssize_t len = readlink("/proc/self/exe", buffer, sizeof(buffer) - 1);
    if (len != -1) {
        buffer[len] = '\0';
        return std::string(dirname(buffer));
    }
    return "";
}

ConfigManager &ConfigManager::Singleton() noexcept
{
    static ConfigManager confMgr;
    return confMgr;
}

ClientInfo ConfigManager::GetClientInfo()
{
    return {this->jobId, this->version};
}

void ConfigManager::ClearPasswd()
{
    memset_s(privateKeyPaaswd, MAX_PASSWD_LENGTH, 0, MAX_PASSWD_LENGTH);
}

void ConfigManager::Init(const Config &conf, int argc, char **argv)
{
    this->version = BUILD_VERSION;
    this->jobId = YR::utility::IDGenerator::GenApplicationId();
    if (conf.mode != Config::Mode::INVALID) {
        this->mode = conf.mode;
    }

    if (!conf.loadPaths.empty()) {
        this->loadPaths = GetValidLoadPaths(conf.loadPaths);
    }
    this->enableMTLS = conf.enableMTLS;
    if (conf.enableMTLS) {
        this->privateKeyPath = conf.privateKeyPath;
        this->certificateFilePath = conf.certificateFilePath;
        this->verifyFilePath = conf.verifyFilePath;
        int len = sizeof(conf.privateKeyPaaswd);
        memcpy_s(this->privateKeyPaaswd, len, conf.privateKeyPaaswd, len);
    }
    this->primaryKeyStoreFile = conf.primaryKeyStoreFile;
    this->standbyKeyStoreFile = conf.standbyKeyStoreFile;
    this->tlsContext = conf.tlsContext;
    this->inCluster = conf.inCluster;
    this->httpIocThreadsNum = conf.httpIocThreadsNum;
    this->serverName = conf.serverName;
    this->isDriver = conf.isDriver;
    this->isLowReliabilityTask = conf.isLowReliabilityTask;
    this->attach = conf.attach;
    if (!conf.serverAddr.empty()) {
        this->functionSystemAddr = conf.serverAddr;
    } else if (this->isDriver && !YR::Libruntime::Config::Instance().YR_SERVER_ADDRESS().empty()) {
        this->functionSystemAddr = YR::Libruntime::Config::Instance().YR_SERVER_ADDRESS();
    } else {
        this->functionSystemAddr = YR::Libruntime::Config::Instance().POSIX_LISTEN_ADDR();
        this->grpcAddress = YR::Libruntime::Config::Instance().POSIX_LISTEN_ADDR();
    }
    this->enableServerMode = conf.enableServerMode;
    if (!conf.dataSystemAddr.empty()) {
        this->dataSystemAddr = conf.dataSystemAddr;
    } else if (this->isDriver && !YR::Libruntime::Config::Instance().YR_DS_ADDRESS().empty()) {
        this->dataSystemAddr = YR::Libruntime::Config::Instance().YR_DS_ADDRESS();
    } else {
        this->dataSystemAddr = YR::Libruntime::Config::Instance().DATASYSTEM_ADDR();
    }
    this->enableDsEncrypt = conf.enableDsEncrypt;
    if (conf.enableDsEncrypt) {
        this->dsPublicKeyContextPath = conf.dsPublicKeyContextPath;
        this->runtimePublicKeyContextPath = conf.runtimePublicKeyContextPath;
        this->runtimePrivateKeyContextPath = conf.runtimePrivateKeyContextPath;
    }

    if (conf.threadPoolSize > 0) {
        this->threadPoolSize = static_cast<uint32_t>(GetValidThreadPoolSize(conf.threadPoolSize));
    }

    this->localThreadPoolSize = conf.localThreadPoolSize;
    if (conf.localThreadPoolSize > 0) {
        this->localThreadPoolSize = static_cast<uint32_t>(GetValidLocalThreadPoolSize(conf.localThreadPoolSize));
    }

    this->defaultGetTimeoutSec = conf.defaultGetTimeoutSec;
    if (this->isDriver) {
        this->runtimeId = "driver";
    }
    if (!YR::Libruntime::Config::Instance().YR_RUNTIME_ID().empty()) {
        this->runtimeId = YR::Libruntime::Config::Instance().YR_RUNTIME_ID();
    }
    this->recycleTime = conf.recycleTime;
    this->maxTaskInstanceNum = GetValidMaxTaskInstanceNum(conf.maxTaskInstanceNum);

    if (conf.mode == Config::CLUSTER_MODE) {
        std::string urn = DEFAULT_CPP_URN;
        if (!YR::Libruntime::Config::Instance().YRFUNCID().empty() || !conf.functionUrn.empty()) {
            urn = conf.functionUrn;
        }
        this->functionId = ConvertFunctionUrnToId(urn);
    }

    if (!conf.pythonFunctionUrn.empty()) {
        this->functionIdPython = ConvertFunctionUrnToId(conf.pythonFunctionUrn);
    } else if (!YR::Libruntime::Config::Instance().YR_PYTHON_FUNCID().empty()) {
        this->functionIdPython = ConvertFunctionUrnToId(YR::Libruntime::Config::Instance().YR_PYTHON_FUNCID());
    }

    if (!conf.javaFunctionUrn.empty()) {
        this->functionIdJava = ConvertFunctionUrnToId(conf.javaFunctionUrn);
    } else if (!YR::Libruntime::Config::Instance().YR_JAVA_FUNCID().empty()) {
        this->functionIdJava = ConvertFunctionUrnToId(YR::Libruntime::Config::Instance().YR_JAVA_FUNCID());
    }

    this->enableMetrics = conf.enableMetrics;
    this->maxConcurrencyCreateNum = GetValidMaxConcurrencyCreateNum(conf.maxConcurrencyCreateNum);
    this->ns = conf.ns;

    if (!conf.logDir.empty()) {
        this->logDir = conf.logDir;
    } else if (!conf.logPath.empty()) {
        this->logDir = conf.logPath;
    } else if (conf.isDriver) {
        this->logDir = YR::Libruntime::Config::Instance().YR_LOG_PATH();
    } else {
        this->logDir = YR::Libruntime::Config::Instance().GLOG_log_dir();
    }

    if (!conf.logLevel.empty()) {
        this->logLevel = conf.logLevel;
    } else if (!YR::Libruntime::Config::Instance().YR_LOG_LEVEL().empty()) {
        this->logLevel = YR::Libruntime::Config::Instance().YR_LOG_LEVEL();
    }
    this->logCompress = GetValidLogCompress(conf.logCompress);
    this->maxLogFileNum = conf.maxLogFileNum;
    this->maxLogFileSize = conf.maxLogSizeMb;
    this->tenantId = conf.tenantId;
    if (argc != 0 && argv != nullptr) {
        absl::ParseCommandLine(argc, argv);
        this->logFlushInterval = absl::GetFlag(FLAGS_logFlushInterval);

        if (!FLAGS_logDir.CurrentValue().empty()) {
            this->logDir = FLAGS_logDir.CurrentValue();
        }

        if (!FLAGS_logLevel.CurrentValue().empty()) {
            this->logLevel = FLAGS_logLevel.CurrentValue();
        }

        if (!FLAGS_grpcAddress.CurrentValue().empty()) {
            this->functionSystemAddr = FLAGS_grpcAddress.CurrentValue();
            this->grpcAddress = FLAGS_grpcAddress.CurrentValue();
        }

        if (!FLAGS_runtimeId.CurrentValue().empty()) {
            this->runtimeId = FLAGS_runtimeId.CurrentValue();
        }

        if (!FLAGS_jobId.CurrentValue().empty()) {
            this->jobId = FLAGS_jobId.CurrentValue();
        }

        const auto &codePathsValue = absl::GetFlag(FLAGS_codePath);
        this->loadPaths.insert(this->loadPaths.end(), codePathsValue.begin(), codePathsValue.end());
    }
    if (conf.mode == Config::CLUSTER_MODE) {
        // parse the info with auto init
        YR::Libruntime::ClusterAccessInfo info{
            .serverAddr = this->functionSystemAddr,
            .dsAddr = this->dataSystemAddr,
            .inCluster = this->inCluster,
        };
        info = YR::Libruntime::AutoGetClusterAccessInfo(info);
        this->functionSystemAddr = info.serverAddr;  // leading protocol will be trimmed, the value would never change
        this->dataSystemAddr = info.dsAddr;          // changes when this is empty
        this->inCluster = info.inCluster;            // changes only when read from masterinfo,
                                                     // or a protocol is specified in functionSystemAddr
    }

    this->customEnvs = conf.customEnvs;
    this->launchUserBinary = conf.launchUserBinary;
    this->workingDir = conf.workingDir;
}

bool ConfigManager::IsLocalMode() const
{
    if (this->mode == Config::LOCAL_MODE) {
        return true;
    } else {
        return false;
    }
}
}  // namespace YR
