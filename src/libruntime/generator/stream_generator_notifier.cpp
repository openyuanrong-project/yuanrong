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

#include "stream_generator_notifier.h"

namespace YR {
namespace Libruntime {

void StreamGeneratorNotifier::Initialize(void)
{
    std::call_once(flag_, std::bind(&StreamGeneratorNotifier::DoInitializeOnce, this));
}

StreamGeneratorNotifier::~StreamGeneratorNotifier(void)
{
    Stop();
    if (timer_ != nullptr) {
        timer_->cancel();
    }
    if (notifyThread_.joinable()) {
        notifyThread_.join();
    }
}

void StreamGeneratorNotifier::Stop(void)
{
    stopped = true;
}

void StreamGeneratorNotifier::DoInitializeOnce(void)
{
    notifyThread_ = std::thread(&StreamGeneratorNotifier::Notify, this);
    timer_ = timerWorker_->CreateTimer(10000, -1, [this]() { NotifyHeartbeat(); });  // 10s interval
}

void StreamGeneratorNotifier::Notify()
{
    ErrorInfo err;
    while (!stopped) {
        {
            std::unique_lock<std::mutex> lock(mux_);
            cv.wait(lock, [this] { return !genQueue_.empty(); });
        }
        YRLOG_DEBUG("start process notify info");
        std::vector<std::shared_ptr<GeneratorNotifyData>> datas;
        PopBatch(datas);
        for (size_t i = 0; i < datas.size(); i++) {
            if (datas[i]->isHeartbeat) {
                YRLOG_DEBUG("start send heartbeat, generator id: {}", datas[i]->generatorId);
                err = NotifyHeartbeatByStream(datas[i]->generatorId);
            } else if (datas[i]->finish) {
                YRLOG_DEBUG("start send finish {}-{})", datas[i]->generatorId, datas[i]->numberResult);
                err = NotifyFinishedByStream(datas[i]->generatorId, datas[i]->numberResult);
            } else {
                YRLOG_DEBUG("start send result {}-{})", datas[i]->generatorId, datas[i]->index);
                err = NotifyResultByStream(datas[i]->generatorId, datas[i]->index, datas[i]->data, datas[i]->err);
            }
            if (!err.OK()) {
                YRLOG_ERROR("failed to notify {} of {}-{}, err code: {}, err message: {}",
                            datas[i]->isHeartbeat ? "heartbeat" : "result", datas[i]->generatorId, datas[i]->index,
                            fmt::underlying(err.Code()), err.Msg());
            }
        }
    }
}

void StreamGeneratorNotifier::PopBatch(std::vector<std::shared_ptr<GeneratorNotifyData>> &datas)
{
    datas.clear();
    std::lock_guard<std::mutex> lk(mux_);
    datas.swap(genQueue_);
}

ErrorInfo StreamGeneratorNotifier::NotifyResult(const std::string &generatorId, int index,
                                                std::shared_ptr<DataObject> resultObj, const ErrorInfo &resultErr)
{
    auto notifyData = std::make_shared<GeneratorNotifyData>();
    notifyData->generatorId = generatorId;
    notifyData->index = index;
    notifyData->data = resultObj;
    notifyData->err = resultErr;
    std::lock_guard<std::mutex> lk(mux_);
    genQueue_.push_back(notifyData);
    cv.notify_one();
    return ErrorInfo();
}

ErrorInfo StreamGeneratorNotifier::NotifyFinished(const std::string &generatorId, int numResults)
{
    auto notifyData = std::make_shared<GeneratorNotifyData>();
    notifyData->generatorId = generatorId;
    notifyData->numberResult = numResults;
    notifyData->finish = true;
    std::lock_guard<std::mutex> lk(mux_);
    genQueue_.push_back(notifyData);
    cv.notify_one();
    return ErrorInfo();
}

ErrorInfo StreamGeneratorNotifier::NotifyHeartbeat()
{
    if (!map_) {
        YRLOG_ERROR("null map pointer");
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "null pointer");
    }

    std::vector<std::string> generatorIds;
    map_->GetRecordKeys(generatorIds);

    std::lock_guard<std::mutex> lk(mux_);
    for (auto &generatorId : generatorIds) {
        auto notifyData = std::make_shared<GeneratorNotifyData>();
        notifyData->generatorId = generatorId;
        notifyData->isHeartbeat = true;
        genQueue_.push_back(notifyData);
    }
    cv.notify_one();
    return ErrorInfo();
}

ErrorInfo StreamGeneratorNotifier::NotifyResultByStream(const std::string &generatorId, int index,
                                                        std::shared_ptr<DataObject> resultObj,
                                                        const ErrorInfo &resultErr)
{
    std::string topic;
    auto err = CheckAndGetTopic(generatorId, topic);
    if (!err.OK()) {
        return err;
    }
    std::shared_ptr<StreamProducer> producer;
    err = GetOrCreateProducer(topic, producer);
    if (!err.OK()) {
        YRLOG_ERROR("failed to get producer when notify result, err code: {}, err message: {}",
                    fmt::underlying(err.Code()), err.Msg());
        return err;
    }

    if (index == 0) {
        IncreaseProducerReference(topic);
    }

    auto res = BuildGeneratorResult(generatorId, index, resultObj, resultErr);
    Element ele(reinterpret_cast<uint8_t *>(res.data()), res.size());
    err = producer->Send(ele);
    if (!err.OK()) {
        YRLOG_ERROR("failed to send notify result to stream, err code: {}, err message: {}",
                    fmt::underlying(err.Code()), err.Msg());
    }

    if (!resultErr.OK()) {
        if (!DecreaseProducerReference(topic)) {
            RemoveProducer(topic);
        }
        map_->RemoveRecord(generatorId);
    }

    return err;
}

ErrorInfo StreamGeneratorNotifier::NotifyFinishedByStream(const std::string &generatorId, int numResults)
{
    std::string topic;
    auto err = CheckAndGetTopic(generatorId, topic);
    if (!err.OK()) {
        return err;
    }
    std::shared_ptr<StreamProducer> producer;
    err = GetOrCreateProducer(topic, producer);
    if (!err.OK()) {
        YRLOG_ERROR("failed to get producer when notify finished, err code: {}, err message: {}",
                    fmt::underlying(err.Code()), err.Msg());
        return err;
    }

    auto res = BuildGeneratorFinished(generatorId, numResults);
    Element ele(reinterpret_cast<uint8_t *>(res.data()), res.size());
    err = producer->Send(ele);
    if (!err.OK()) {
        YRLOG_ERROR("failed to send notify finished to stream, err code: {}, err message: {}",
                    fmt::underlying(err.Code()), err.Msg());
    }

    if (!DecreaseProducerReference(topic)) {
        RemoveProducer(topic);
    }
    map_->RemoveRecord(generatorId);
    return err;
}

ErrorInfo StreamGeneratorNotifier::NotifyHeartbeatByStream(const std::string &generatorId)
{
    std::string topic;
    auto err = CheckAndGetTopic(generatorId, topic);
    if (!err.OK()) {
        return err;
    }
    std::shared_ptr<StreamProducer> producer;
    err = GetOrCreateProducer(topic, producer);
    if (!err.OK()) {
        YRLOG_ERROR("failed to get producer when notify heartbeat, err code: {}, err message: {}",
                    fmt::underlying(err.Code()), err.Msg());
        return err;
    }
    IncreaseProducerReference(topic);

    auto res = BuildGeneratorHeartbeat(generatorId);
    Element ele(reinterpret_cast<uint8_t *>(res.data()), res.size());
    err = producer->Send(ele);
    if (!err.OK()) {
        YRLOG_ERROR("failed to send notify heartbeat to stream, err code: {}, err message: {}",
                    fmt::underlying(err.Code()), err.Msg());
    }

    if (!DecreaseProducerReference(topic)) {
        RemoveProducer(topic);
    }
    return err;
}

std::string StreamGeneratorNotifier::BuildGeneratorResult(const std::string &generatorId, int index,
                                                          std::shared_ptr<DataObject> resultObj,
                                                          const ErrorInfo &resultErr)
{
    libruntime::NotifyGeneratorResult result;
    result.set_genid(generatorId);
    result.set_index(index);
    result.set_objectid(resultObj->id);
    if (resultObj->buffer->IsNative()) {
        result.set_data(resultObj->buffer->ImmutableData(), resultObj->buffer->GetSize());
    }
    result.set_errorcode(resultErr.Code());
    result.set_errormessage(resultErr.Msg());
    return result.SerializeAsString();
}

std::string StreamGeneratorNotifier::BuildGeneratorFinished(const std::string &generatorId, int numResults)
{
    libruntime::NotifyGeneratorResult result;
    result.set_genid(generatorId);
    result.set_finished(true);
    result.set_numresults(numResults);
    return result.SerializeAsString();
}

std::string StreamGeneratorNotifier::BuildGeneratorHeartbeat(const std::string &generatorId)
{
    libruntime::NotifyGeneratorResult result;
    result.set_genid(generatorId);
    result.set_isheartbeat(true);
    return result.SerializeAsString();
}

ErrorInfo StreamGeneratorNotifier::GetOrCreateProducer(const std::string &topic,
                                                       std::shared_ptr<StreamProducer> &producer)
{
    absl::MutexLock lock(&mu_);
    if (producers_.find(topic) != producers_.end()) {
        producer = producers_[topic];
        return ErrorInfo();
    }

    if (!dsStreamStore_) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "invalid stream store");
    }

    producer = std::make_shared<StreamProducer>();
    ProducerConf producerConf;
    producerConf.retainForNumConsumers = 1;
    // for improving stream speed
    producerConf.delayFlushTime = YR::Libruntime::Config::Instance().DS_DELAY_FLUSH_TIME();
    auto err = dsStreamStore_->CreateStreamProducer(topic, producer, producerConf);
    if (err.OK()) {
        producers_[topic] = producer;
    }
    YRLOG_DEBUG("success create producer of {}", topic);
    return err;
}

ErrorInfo StreamGeneratorNotifier::RemoveProducer(const std::string &topic)
{
    if (!YR::Libruntime::Config::Instance().ENABLE_CLEAN_STREAM_PRODUCER()) {
        YRLOG_DEBUG("Don't remove producer {}", topic);
        return ErrorInfo();
    }
    absl::MutexLock lock(&mu_);
    if (producers_.find(topic) != producers_.end()) {
        producers_.erase(topic);
    }
    return ErrorInfo();
}

ErrorInfo StreamGeneratorNotifier::CheckAndGetTopic(const std::string &generatorId, std::string &topic)
{
    if (!map_) {
        YRLOG_ERROR("null map pointer");
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "null pointer");
    }
    map_->GetRecord(generatorId, topic);
    if (topic.empty()) {
        YRLOG_ERROR("topic not found, generator id: {}", generatorId);
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "invalid topic");
    }
    return ErrorInfo();
}

void StreamGeneratorNotifier::IncreaseProducerReference(const std::string &topic)
{
    if (producerReferences_.find(topic) != producerReferences_.end()) {
        producerReferences_[topic]++;
        return;
    }
    producerReferences_[topic] = 1;
}

bool StreamGeneratorNotifier::DecreaseProducerReference(const std::string &topic)
{
    if (producerReferences_.find(topic) == producerReferences_.end()) {
        return false;
    }
    producerReferences_[topic]--;
    if (producerReferences_[topic] <= 0) {
        producerReferences_.erase(topic);
        return false;
    }
    return true;
}
}  // namespace Libruntime
}  // namespace YR