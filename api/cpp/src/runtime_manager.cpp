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
#include "yr/api/runtime_manager.h"

#include "api/cpp/src/cluster_mode_runtime.h"
#include "api/cpp/src/config_manager.h"
#include "yr/api/local_mode_runtime.h"
namespace YR {
namespace internal {

RuntimeManager &RuntimeManager::GetInstance()
{
    static RuntimeManager instance;
    return instance;
}

void RuntimeManager::Initialize(Config::Mode mode)
{
    this->mode_ = mode;
    if (mode == Config::Mode::CLUSTER_MODE) {
        yrRuntime = std::make_shared<ClusterModeRuntime>();
        yrRuntime->Init();
    } else {
        YR::utility::LogParam logParam;
        logParam.logLevel = ConfigManager::Singleton().logLevel;
        logParam.logDir = ConfigManager::Singleton().logDir;
        logParam.logBufSecs = ConfigManager::Singleton().logFlushInterval;
        logParam.maxSize = ConfigManager::Singleton().maxLogFileSize;
        logParam.maxFiles = ConfigManager::Singleton().maxLogFileNum;
        logParam.nodeName = ConfigManager::Singleton().jobId;
        logParam.modelName = ConfigManager::Singleton().runtimeId;
        InitLog(logParam);
    }
    // cluster mode also inits local mode runtime
    // local after cluster because cluster mode runtime inits libruntime
    localModeRuntime_ = std::make_shared<LocalModeRuntime>();
    localModeRuntime_->Init();
}

void RuntimeManager::Initialize(std::shared_ptr<YR::Runtime> runtime)
{
    this->mode_ = Config::Mode::CLUSTER_MODE;
    // cluster mode also init local mode runtime
    localModeRuntime_ = std::make_shared<LocalModeRuntime>();
    localModeRuntime_->Init();
    yrRuntime = runtime;
}

void RuntimeManager::Stop()
{
    if (localModeRuntime_) {
        localModeRuntime_->Stop();
    }
    if (!IsLocalMode()) {
        ClusterModeRuntime::StopRuntime();
    }
}

std::shared_ptr<LocalModeRuntime> GetLocalModeRuntime()
{
    return RuntimeManager::GetInstance().GetLocalModeRuntime();
}

bool IsLocalMode()
{
    return RuntimeManager::GetInstance().IsLocalMode();
}

std::shared_ptr<YR::Runtime> GetRuntime()
{
    return RuntimeManager::GetInstance().GetRuntime();
}
}  // namespace internal
}  // namespace YR