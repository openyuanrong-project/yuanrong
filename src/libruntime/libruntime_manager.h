/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#include <memory>

#include "src/libruntime/libruntime.h"
#include "src/libruntime/libruntime_config.h"
#include "src/utility/logger/log_manager.h"
namespace YR {
namespace Libruntime {
using YR::utility::LogManager;
class LibruntimeManager {
public:
    static LibruntimeManager &Instance()
    {
        static LibruntimeManager instance;
        return instance;
    }

    ErrorInfo Init(const LibruntimeConfig &config, const std::string &rtCtx = "");

    void Finalize(const std::string &rtCtx = "");

    std::shared_ptr<Libruntime> GetLibRuntime(const std::string &rtCtx = "");

    void SetLibRuntime(std::shared_ptr<Libruntime> libruntime, const std::string &rtCtx = "");

    void SetRuntimeContext(const std::string &jobId);

    bool IsInitialized(const std::string &rtCtx = "");

    void ReceiveRequestLoop(const std::string &rtCtx = "");

    ErrorInfo HandleInitialized(const LibruntimeConfig &config, const std::string &rtCtx);

    void ExecShutdownCallback(int signum, bool needExit = true);

private:
    LibruntimeManager();

    void InstallSigtermHandler();

    void static SigtermHandler(int signum, siginfo_t *info, void *ptr);

    void ExecShutdownCallbackAsync(int signum);

    ErrorInfo CreateLibruntime(std::shared_ptr<LibruntimeConfig> librtConfig, std::shared_ptr<Libruntime> &librt);

    mutable std::mutex rtMtx;

    mutable std::mutex rtCfgMtx;

    std::unordered_map<std::string, std::shared_ptr<Libruntime>> librts;

    std::unordered_map<std::string, std::shared_ptr<LibruntimeConfig>> librtConfigs;

    std::shared_ptr<ClientsManager> clientsMgr;

    std::shared_ptr<MetricsAdaptor> metricsAdaptor;

    std::shared_ptr<RuntimeContext> runtimeContext;

    std::shared_ptr<DomainSocketClient> socketClient_;

    bool isLogMerge = false;

    std::shared_ptr<LogManager> logManager_;

    std::once_flag flag_;
};
}  // namespace Libruntime
}  // namespace YR