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

#include "stream_generator_receiver.h"

namespace YR {
namespace Libruntime {
const size_t RECEIVE_INTERVAL = 1000;
const uint32_t CONSUMER_TIMEOUT = 5;
const long long HEARTBEAT_TIMEOUT = 30000;
const int HEARTBEAT_TIMER_TIMEOUT = 10000;
void StreamGeneratorReceiver::Initialize(void)
{
    std::call_once(flag_, std::bind(&StreamGeneratorReceiver::DoInitializeOnce, this));
}

StreamGeneratorReceiver::~StreamGeneratorReceiver(void)
{
    Stop();
    if (timer_ != nullptr) {
        timer_->cancel();
    }
    if (receiverThread_.joinable()) {
        receiverThread_.join();
    }
}

void StreamGeneratorReceiver::Stop(void)
{
    stopped = true;
    if (consumer_ != nullptr) {
        auto err = consumer_->Close();
        if (!err.OK()) {
            YRLOG_ERROR("failed to close consumer, err code: {}, err message: {}", fmt::underlying(err.Code()),
                        err.Msg());
        }
        YRLOG_INFO("consummer closed");
        if (!topic_.empty()) {
            err = dsStreamStore_->DeleteStream(topic_);
            if (!err.OK()) {
                YRLOG_ERROR("failed to delete stream {}, err code: {}, err message: {}", topic_,
                            fmt::underlying(err.Code()), err.Msg());
            }
        }
    }
}

void StreamGeneratorReceiver::DoInitializeOnce(void)
{
    receiverThread_ = std::thread(&StreamGeneratorReceiver::Receive, this);
    timer_ = timerWorker_->CreateTimer(HEARTBEAT_TIMER_TIMEOUT, -1, [this]() { DetectHeartbeat(); });
}

void StreamGeneratorReceiver::DetectHeartbeat(void)
{
    std::vector<std::string> generatorIds;
    map_->GetRecordKeys(generatorIds);
    for (auto &genId : generatorIds) {
        std::string timestamp;
        map_->GetRecord(genId, timestamp);
        if (timestamp.empty()) {
            YRLOG_WARN("timestamp not found, generator id: {}", genId);
            continue;
        }
        long long now = GetCurrentTimestampMs();
        long long lastTime = std::stoll(timestamp);
        if (now - lastTime >= HEARTBEAT_TIMEOUT) {
            auto errMsg =
                "stream generator receiver detect heartbeat failed, the connection between runtime and ds worker maybe "
                "disconnected, generator id: " + genId;
            MarkEndOfStream(genId, ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, errMsg));
            map_->RemoveRecord(genId);
        }
    }
}

void StreamGeneratorReceiver::Receive(void)
{
    auto topic = libruntimeConfig_->runtimeId;
    if (topic == "driver") {
        topic = libruntimeConfig_->runtimeId + "_" + libruntimeConfig_->jobId;
    }
    SubscriptionConfig config;
    std::shared_ptr<StreamConsumer> consumer = std::make_shared<StreamConsumer>();
    initErr_ = dsStreamStore_->CreateStreamConsumer(topic, config, consumer, true);
    if (!initErr_.OK()) {
        YRLOG_ERROR("failed to create stream consumer, err code: {}, err message: {}", fmt::underlying(initErr_.Code()),
                    initErr_.Msg());
        return;
    }
    this->consumer_ = consumer;
    this->topic_ = topic;

    YRLOG_INFO("begin to receive message from topic: {}", topic);

    while (!stopped) {
        std::vector<Element> elements;
        auto err = consumer->Receive(1, CONSUMER_TIMEOUT, elements);
        if (!err.OK()) {
            YRLOG_ERROR("failed to receive from consumer, err code: {}, err message: {}", fmt::underlying(err.Code()),
                        err.Msg());
            if (err.Code() == ErrorCode::ERR_CLIENT_ALREADY_CLOSED) {
                std::this_thread::sleep_for(std::chrono::milliseconds(RECEIVE_INTERVAL));
            }
            continue;
        }
        for (auto ele : elements) {
            libruntime::NotifyGeneratorResult result;
            if (!result.ParseFromArray(ele.ptr, ele.size)) {
                YRLOG_ERROR("failed to parse element, topic: {}, id: {}", topic, ele.id);
                continue;
            }

            if (result.finished()) {
                HandleGeneratorFinished(result);
            } else if (result.isheartbeat()) {
                HandleGeneratorHeartbeat(result.genid());
            } else {
                HandleGeneratorResult(result);
            }
        }
    };
    YRLOG_INFO("finishe to receive message from topic: {}", topic);
}

void StreamGeneratorReceiver::HandleGeneratorFinished(const libruntime::NotifyGeneratorResult &result)
{
    absl::MutexLock lock(&mu_);
    auto &genId = result.genid();
    map_->RemoveRecord(genId);
    if (generatorResultsCounter_.find(genId) == generatorResultsCounter_.end()) {
        generatorResultsCounter_[genId] = 0;
    }
    auto genObjectID = memoryStore_->GenerateObjectId(genId, result.numresults());
    YRLOG_DEBUG("received finished {}", genObjectID);
    SetError(genId, genObjectID, result.numresults(),
             ErrorInfo(ErrorCode::ERR_GENERATOR_FINISHED, ModuleCode::RUNTIME, ""));
    if (generatorResultsCounter_[genId] == result.numresults()) {
        memoryStore_->GeneratorFinished(genId);
        ClearCountersByGenId(genId);
    } else {
        YRLOG_INFO("stream message maybe unordered, gen id: {}, num of results: {}, counter of results: {}", genId,
                   result.numresults(), generatorResultsCounter_[genId]);
        numGeneratorResults_[genId] = result.numresults();
    }
}

void StreamGeneratorReceiver::HandleGeneratorResult(const libruntime::NotifyGeneratorResult &result)
{
    absl::MutexLock lock(&mu_);
    auto &genId = result.genid();
    if (generatorResultsCounter_.find(genId) == generatorResultsCounter_.end()) {
        generatorResultsCounter_[genId] = 0;
    }
    generatorResultsCounter_[genId] += 1;
    auto genObjectID = memoryStore_->GenerateObjectId(genId, result.index());
    YRLOG_DEBUG("received result {}", genObjectID);
    if (result.errorcode()) {
        map_->RemoveRecord(genId);
        SetError(genId, genObjectID, result.index(),
                 ErrorInfo(static_cast<ErrorCode>(result.errorcode()), ModuleCode::RUNTIME, result.errormessage()));
        memoryStore_->GeneratorFinished(genId);
        ClearCountersByGenId(genId);
        return;
    }
    if (!result.data().empty()) {
        std::shared_ptr<Buffer> buf = std::make_shared<NativeBuffer>(result.data().size());
        memoryStore_->AddReturnObject(genObjectID);
        buf->MemoryCopy(result.data().data(), result.data().size());
        auto err = memoryStore_->Put(buf, genObjectID, {}, false);
        if (err.OK()) {
            memoryStore_->SetReady(genObjectID);
        } else {
            memoryStore_->AddOutput(genId, genObjectID, result.index(), err);
            YRLOG_ERROR("failed to put stream result, err code: {}, err message: {}", fmt::underlying(err.Code()),
                        err.Msg());
            return;
        }
    } else {
        YRLOG_ERROR("receive empty result data {}", genId);
    }
    memoryStore_->AddOutput(genId, genObjectID, result.index());
    if (numGeneratorResults_.find(genId) != numGeneratorResults_.end() &&
        numGeneratorResults_[genId] == generatorResultsCounter_[genId]) {
        memoryStore_->GeneratorFinished(genId);
        ClearCountersByGenId(genId);
    }
}

void StreamGeneratorReceiver::AddRecord(const std::string &genId)
{
    long long now = GetCurrentTimestampMs();
    map_->AddRecord(genId, std::to_string(now));
}

void StreamGeneratorReceiver::HandleGeneratorHeartbeat(const std::string &genId)
{
    long long now = GetCurrentTimestampMs();
    map_->UpdateRecord(genId, std::to_string(now));
}

void StreamGeneratorReceiver::MarkEndOfStream(const std::string &genId, const ErrorInfo &errInfo)
{
    absl::MutexLock lock(&mu_);
    auto index = 0;
    if (generatorResultsCounter_.find(genId) != generatorResultsCounter_.end()) {
        index = generatorResultsCounter_[genId];
    }
    auto genObjectID = memoryStore_->GenerateObjectId(genId, index);
    YRLOG_DEBUG("mark end of stream {}", genObjectID);
    SetError(genId, genObjectID, index, errInfo);
    memoryStore_->GeneratorFinished(genId);
    ClearCountersByGenId(genId);
}

void StreamGeneratorReceiver::SetError(const std::string &genId, const std::string &genObjectID, uint64_t index,
                                       const ErrorInfo &errInfo)
{
    YRLOG_DEBUG("set error stream result of {}, err code: {}, err message: {}", genObjectID,
                fmt::underlying(errInfo.Code()), errInfo.Msg());
    std::shared_ptr<Buffer> buf = std::make_shared<NativeBuffer>(genId.size());
    memoryStore_->AddReturnObject(genObjectID);
    auto err = memoryStore_->Put(buf, genObjectID, {}, false, errInfo);
    if (err.OK()) {
        memoryStore_->SetError(genObjectID, errInfo);
    } else {
        memoryStore_->AddOutput(genId, genObjectID, index, err);
        YRLOG_ERROR("failed to put stream finished result, err code: {}, err message: {}",
                    fmt::underlying(err.Code()), err.Msg());
        return;
    }
    memoryStore_->AddOutput(genId, genObjectID, index, errInfo);
}

void StreamGeneratorReceiver::ClearCountersByGenId(const std::string &genId)
{
    generatorResultsCounter_.erase(genId);
    numGeneratorResults_.erase(genId);
}
}  // namespace Libruntime
}  // namespace YR