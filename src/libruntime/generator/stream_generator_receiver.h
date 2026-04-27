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
#include <memory>
#include <mutex>

#include "generator_id_map.h"
#include "generator_receiver.h"
#include "src/libruntime/libruntime_config.h"
#include "src/libruntime/objectstore/memory_store.h"
#include "src/libruntime/streamstore/stream_store.h"

namespace YR {
namespace Libruntime {
using namespace YR::utility;
class StreamGeneratorReceiver : public GeneratorReceiver {
public:
    StreamGeneratorReceiver(std::shared_ptr<LibruntimeConfig> config, std::shared_ptr<StreamStore> store,
                            std::shared_ptr<MemoryStore> mStore)
        : GeneratorReceiver(), libruntimeConfig_(config), dsStreamStore_(store), memoryStore_(mStore)
    {
        this->timerWorker_ = std::make_shared<TimerWorker>();
        this->map_ = std::make_shared<GeneratorIdMap>();
    }
    virtual ~StreamGeneratorReceiver();
    void Initialize(void) override;
    void Stop(void) override;
    void MarkEndOfStream(const std::string &genId, const ErrorInfo &errInfo) override;
    void AddRecord(const std::string &genId) override;

private:
    void DoInitializeOnce(void);
    void Receive(void);
    void HandleGeneratorFinished(const libruntime::NotifyGeneratorResult &result);
    void HandleGeneratorResult(const libruntime::NotifyGeneratorResult &result);
    void HandleGeneratorHeartbeat(const std::string &genId);
    void DetectHeartbeat(void);
    void ClearCountersByGenId(const std::string &genId);
    void SetError(const std::string &genId, const std::string &genObjectID, uint64_t index, const ErrorInfo &errInfo);

    std::once_flag flag_;
    std::shared_ptr<LibruntimeConfig> libruntimeConfig_;
    std::shared_ptr<StreamStore> dsStreamStore_;
    std::shared_ptr<MemoryStore> memoryStore_;
    std::thread receiverThread_;
    std::shared_ptr<StreamConsumer> consumer_;
    std::string topic_;
    std::atomic<bool> stopped{false};
    ErrorInfo initErr_;
    std::unordered_map<std::string, int> generatorResultsCounter_;
    std::unordered_map<std::string, int> numGeneratorResults_;
    absl::Mutex mu_;
    std::shared_ptr<GeneratorIdMap> map_;
    std::shared_ptr<TimerWorker> timerWorker_;
    std::shared_ptr<YR::utility::Timer> timer_;
};
}  // namespace Libruntime
}  // namespace YR