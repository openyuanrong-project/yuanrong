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

#include "generator_id_map.h"
#include "generator_notifier.h"
#include "src/dto/config.h"
#include "src/dto/invoke_options.h"
#include "src/libruntime/streamstore/stream_producer_consumer.h"
#include "src/libruntime/streamstore/stream_store.h"

namespace YR {
namespace Libruntime {
using namespace YR::utility;

struct GeneratorNotifyData {
    std::string generatorId;
    int index = 0;
    int numberResult = 0;
    std::shared_ptr<DataObject> data;
    ErrorInfo err = ErrorInfo();
    bool finish = false;
    bool isHeartbeat = false;
};

class StreamGeneratorNotifier : public GeneratorNotifier {
public:
    StreamGeneratorNotifier(std::shared_ptr<StreamStore> store, std::shared_ptr<GeneratorIdMap> map)
        : GeneratorNotifier(), dsStreamStore_(store), map_(map)
    {
        this->timerWorker_ = std::make_shared<TimerWorker>();
    }

    virtual ~StreamGeneratorNotifier();
    void Initialize(void) override;
    void Stop(void) override;

    virtual ErrorInfo NotifyResult(const std::string &generatorId, int index, std::shared_ptr<DataObject> resultObj,
                                   const ErrorInfo &resultErr) override;

    virtual ErrorInfo NotifyFinished(const std::string &generatorId, int numResults) override;

    ErrorInfo NotifyResultByStream(const std::string &generatorId, int index, std::shared_ptr<DataObject> resultObj,
                                   const ErrorInfo &resultErr);
    ErrorInfo NotifyFinishedByStream(const std::string &generatorId, int numResults);
    ErrorInfo NotifyHeartbeatByStream(const std::string &generatorId);

private:
    void Notify();
    void DoInitializeOnce(void);
    ErrorInfo NotifyHeartbeat();
    std::string BuildGeneratorResult(const std::string &generatorId, int index, std::shared_ptr<DataObject> resultObj,
                                     const ErrorInfo &resultErr);
    std::string BuildGeneratorFinished(const std::string &generatorId, int numResults);
    std::string BuildGeneratorHeartbeat(const std::string &generatorId);
    void PopBatch(std::vector<std::shared_ptr<GeneratorNotifyData>> &datas);

    ErrorInfo GetOrCreateProducer(const std::string &topic, std::shared_ptr<StreamProducer> &producer);
    ErrorInfo RemoveProducer(const std::string &topic);
    ErrorInfo CheckAndGetTopic(const std::string &generatorId, std::string &topic);
    void IncreaseProducerReference(const std::string &topic);
    bool DecreaseProducerReference(const std::string &topic);

    std::unordered_map<std::string, std::shared_ptr<StreamProducer>> producers_;
    std::unordered_map<std::string, int> producerReferences_;
    std::shared_ptr<StreamStore> dsStreamStore_;
    std::shared_ptr<GeneratorIdMap> map_;
    absl::Mutex mu_;
    std::atomic<bool> stopped{false};
    std::once_flag flag_;
    std::thread notifyThread_;
    mutable std::mutex mux_;
    std::condition_variable cv;
    std::vector<std::shared_ptr<GeneratorNotifyData>> genQueue_;
    std::shared_ptr<TimerWorker> timerWorker_;
    std::shared_ptr<YR::utility::Timer> timer_;
};
}  // namespace Libruntime
}  // namespace YR