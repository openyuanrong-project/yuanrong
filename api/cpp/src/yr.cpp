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

#include "yr/yr.h"

#include "api/cpp/include/yr/api/kv_manager.h"
#include "api/cpp/src/code_manager.h"
#include "api/cpp/src/config_manager.h"
#include "api/cpp/src/internal.h"
#include "cluster_mode_runtime.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_manager.h"
#include "yr/api/runtime.h"
#include "yr/api/runtime_manager.h"
#include "yr/api/serdes.h"
thread_local std::unordered_set<std::string> localNestedObjList;

namespace YR {

static bool g_isInit = false;

bool IsInitialized()
{
    return g_isInit;
}

void CheckInitialized()
{
    if (!IsInitialized()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_INIT_USAGE, YR::Libruntime::ModuleCode::RUNTIME,
                            "the current yr init status is abnormal, please init YR first");
    }
}

void SetInitialized(bool flag)
{
    g_isInit = flag;
}

void ReentrantFinalize()
{
    if (!IsInitialized()) {
        return;
    }
    YR::internal::RuntimeManager::GetInstance().Stop();
}

struct Defer {
    explicit Defer(std::function<void()> &&f)
    {
        deferFunc = std::move(f);
    }
    ~Defer()
    {
        deferFunc();
    }
    std::function<void()> deferFunc = nullptr;
};

void AtExitHandler()
{
    ReentrantFinalize();
}

ClientInfo Init(const Config &conf, int argc, char **argv)
{
    static bool hasRegisteredAtExit = false;
    Defer registerAtExitHandler([]() {
        if (!hasRegisteredAtExit) {
            atexit(AtExitHandler);
            hasRegisteredAtExit = true;
        }
    });
    std::string serverVersion = "";
    if (!IsInitialized()) {
        ConfigManager::Singleton().Init(conf, argc, argv);
        YR::internal::RuntimeManager::GetInstance().Initialize(conf.mode);
        if (!ConfigManager::Singleton().isDriver) {
            auto err = internal::CodeManager::LoadFunctions(ConfigManager::Singleton().loadPaths);
            if (!err.OK()) {
                YRLOG_INFO("load function error: Code:{}, MCode:{}, Msg:{}", fmt::underlying(err.Code()),
                           fmt::underlying(err.MCode()), err.Msg());
            }
        }
        SetInitialized();
    }
    auto clientInfo = ConfigManager::Singleton().GetClientInfo();
    if (conf.mode == Config::Mode::CLUSTER_MODE) {
        clientInfo.serverVersion = YR::internal::GetRuntime()->GetServerVersion();
    }
    YRLOG_INFO("Current SDK Version: {}, Server Version: {}", clientInfo.version, clientInfo.serverVersion);
    return clientInfo;
}

ClientInfo Init(const Config &conf)
{
    return Init(conf, 0, nullptr);
}

ClientInfo Init(int argc, char **argv)
{
    Config conf;
    return Init(conf, argc, argv);
}

void Run(int argc, char *argv[])
{
    Config conf;
    conf.isDriver = false;
    conf.launchUserBinary = true;
    Init(conf, argc, argv);
    ReceiveRequestLoop();
}

void Finalize(void)
{
    CheckInitialized();
    if (ConfigManager::Singleton().inCluster && !ConfigManager::Singleton().isDriver) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE,
                            YR::Libruntime::ModuleCode::RUNTIME, "Finalize is not allowed to use on cloud");
    }
    ReentrantFinalize();
    SetInitialized(false);
}

void Exit(void)
{
    CheckInitialized();
    bool isRemoteRuntime = ConfigManager::Singleton().inCluster && !ConfigManager::Singleton().isDriver;
    if (!isRemoteRuntime) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE,
                            YR::Libruntime::ModuleCode::RUNTIME, "Not support exit out of cluster");
    } else if (!YR::internal::IsLocalMode()) {
        YR::internal::GetRuntime()->Exit();
    }
}

void ReceiveRequestLoop(void)
{
    YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->ReceiveRequestLoop();
}

bool IsOnCloud()
{
    return !ConfigManager::Singleton().isDriver;
}

bool IsLocalMode()
{
    if (!g_isInit) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_INIT_USAGE, YR::Libruntime::ModuleCode::RUNTIME,
                            "Please init YR first");
    }
    return ConfigManager::Singleton().IsLocalMode();
}

void WriteBytes(void *data, size_t size, int tag, int index, int src, int dst)
{
    char *data_char = reinterpret_cast<char *>(data);
    std::string key =
        std::to_string(tag) + "-" + std::to_string(index) + "-" + std::to_string(src) + "-" + std::to_string(dst);
    return KVManager::WriteRaw(key, data_char, size);
}

void ReadBytes(void *data, size_t size, int tag, int index, int src, int dst)
{
    std::string key =
        std::to_string(tag) + "-" + std::to_string(index) + "-" + std::to_string(src) + "-" + std::to_string(dst);
    auto buffer = KVManager::ReadBuffer(key);
    auto read_size = buffer->GetSize() < size ? buffer->GetSize() : size;
    memcpy_s(data, size, buffer->ImmutableData(), read_size);
}

std::shared_ptr<Producer> CreateProducer(const std::string &streamName, ProducerConf producerConf)
{
    CheckInitialized();
    if (ConfigManager::Singleton().IsLocalMode()) {  // local mode
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE,
                            YR::Libruntime::ModuleCode::RUNTIME, "local mode does not support CreateProducer\n");
    }
    std::shared_ptr<Producer> producer = YR::internal::GetRuntime()->CreateStreamProducer(streamName, producerConf);
    return producer;
}

std::shared_ptr<Consumer> Subscribe(const std::string &streamName, const SubscriptionConfig &config, bool autoAck)
{
    CheckInitialized();
    if (ConfigManager::Singleton().IsLocalMode()) {  // local mode
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE,
                            YR::Libruntime::ModuleCode::RUNTIME, "local mode does not support Subscribe\n");
    }
    std::shared_ptr<Consumer> consumer = YR::internal::GetRuntime()->CreateStreamConsumer(streamName, config, autoAck);
    return consumer;
}

void DeleteStream(const std::string &streamName)
{
    CheckInitialized();
    if (ConfigManager::Singleton().IsLocalMode()) {  // local mode
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE,
                            YR::Libruntime::ModuleCode::RUNTIME, "local mode does not support DeleteStream\n");
    }
    YR::internal::GetRuntime()->DeleteStream(streamName);
}

void SaveState(const int &timeout)
{
    CheckInitialized();
    bool isRemoteRuntime = ConfigManager::Singleton().inCluster && !ConfigManager::Singleton().isDriver;
    if (!isRemoteRuntime) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE,
                            YR::Libruntime::ModuleCode::RUNTIME, "SaveState is only supported on cloud with posix api");
    }
    if (YR::internal::IsLocalMode()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE,
                            YR::Libruntime::ModuleCode::RUNTIME, "SaveState is not supported in local mode");
    }
    YR::internal::GetRuntime()->SaveState(timeout);
}

void LoadState(const int &timeout)
{
    CheckInitialized();
    bool isRemoteRuntime = ConfigManager::Singleton().inCluster && !ConfigManager::Singleton().isDriver;
    if (!isRemoteRuntime) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE,
                            YR::Libruntime::ModuleCode::RUNTIME, "LoadState is only supported on cloud with posix api");
    }

    if (YR::internal::IsLocalMode()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE,
                            YR::Libruntime::ModuleCode::RUNTIME, "LoadState is not supported in local mode");
    }
    YR::internal::GetRuntime()->LoadState(timeout);
}

std::vector<Node> Nodes()
{
    CheckInitialized();
    std::vector<Node> nodes = YR::internal::GetRuntime()->Nodes();
    return nodes;
}

std::shared_ptr<MutableBuffer> CreateBuffer(uint64_t size)
{
    CheckInitialized();
    if (ConfigManager::Singleton().IsLocalMode()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE, YR::ModuleCode::RUNTIME_,
                            "local mode does not support CreateBuffer\n");
    } else {
        return YR::internal::GetRuntime()->CreateMutableBuffer(size);
    }
}

std::vector<std::shared_ptr<MutableBuffer>> Get(const std::vector<ObjectRef<MutableBuffer>> &objs, int timeoutSec)
{
    CheckInitialized();
    if (ConfigManager::Singleton().IsLocalMode()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE, YR::ModuleCode::RUNTIME_,
                            "local mode does not support Get\n");
    } else {
        std::vector<std::string> ids;
        for (size_t i = 0; i < objs.size(); i++) {
            ids.push_back(objs[i].ID());
        }
        return YR::internal::GetRuntime()->GetMutableBuffer(ids, timeoutSec);
    }
}

std::string Serialize(ObjectRef<MutableBuffer> &obj)
{
    msgpack::sbuffer buffer = YR::internal::Serialize(std::move(obj));
    std::string str(buffer.data(), buffer.size());
    return str;
}

ObjectRef<MutableBuffer> Deserialize(const void *value, int size)
{
    msgpack::sbuffer buffer;
    const char *valueChar = static_cast<const char *>(value);
    buffer.write(valueChar, size);
    return YR::internal::Deserialize<ObjectRef<MutableBuffer>>(buffer);
}

}  // namespace YR
