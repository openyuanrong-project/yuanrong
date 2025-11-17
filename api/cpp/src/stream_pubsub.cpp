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

#include "api/cpp/src/stream_pubsub.h"
#include "api/cpp/include/yr/api/exception.h"
#include "src/dto/stream_conf.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_manager.h"

namespace YR {
void StreamProducer::Send(const Element &element)
{
    YR::Libruntime::Element streamElement(element.ptr, element.size, element.id);
    YR::Libruntime::ErrorInfo err = producer_->Send(streamElement);
    if (err.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_ERROR("Send err: Code:{}, MCode:{}, Msg:{}", fmt::underlying(err.Code()), fmt::underlying(err.MCode()),
                    err.Msg());
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), err.Msg());
    }
}

void StreamProducer::Send(const Element &element, int64_t timeoutMs)
{
    YR::Libruntime::Element streamElement(element.ptr, element.size, element.id);
    YR::Libruntime::ErrorInfo err = producer_->Send(streamElement, timeoutMs);
    if (err.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_ERROR("Send err: Code:{}, MCode:{}, Msg:{}", fmt::underlying(err.Code()), fmt::underlying(err.MCode()),
                    err.Msg());
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), err.Msg());
    }
}

void StreamProducer::Close()
{
    auto err = YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->SetTraceId(traceId_);
    if (!err.OK()) {
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), err.Msg());
    }
    err = producer_->Close();
    if (err.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_ERROR("Close err: Code:{}, MCode:{}, Msg:{}", fmt::underlying(err.Code()), fmt::underlying(err.MCode()),
                    err.Msg());
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), err.Msg());
    }
}

void StreamConsumer::Receive(uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements)
{
    std::vector<YR::Libruntime::Element> elements;
    YR::Libruntime::ErrorInfo err = consumer_->Receive(expectNum, timeoutMs, elements);
    if (err.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_ERROR("Receive err: Code:{}, MCode:{}, Msg:{}", fmt::underlying(err.Code()), fmt::underlying(err.MCode()),
                    err.Msg());
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), err.Msg());
    }
    for (auto &element : elements) {
        outElements.push_back(Element(element.ptr, element.size, element.id));
    }
}

void StreamConsumer::Receive(uint32_t timeoutMs, std::vector<Element> &outElements)
{
    std::vector<YR::Libruntime::Element> elements;
    YR::Libruntime::ErrorInfo err = consumer_->Receive(timeoutMs, elements);
    if (err.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_ERROR("Receive err: Code:{}, MCode:{}, Msg:{}", fmt::underlying(err.Code()), fmt::underlying(err.MCode()),
                    err.Msg());
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), err.Msg());
    }
    for (auto &element : elements) {
        outElements.push_back(Element(element.ptr, element.size, element.id));
    }
}

void StreamConsumer::Ack(uint64_t elementId)
{
    YR::Libruntime::ErrorInfo err = consumer_->Ack(elementId);
    if (err.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_ERROR("Ack err: Code:{}, MCode:{}, Msg:{}", fmt::underlying(err.Code()), fmt::underlying(err.MCode()),
                    err.Msg());
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), err.Msg());
    }
}

void StreamConsumer::Close()
{
    auto err = YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->SetTraceId(traceId_);
    if (!err.OK()) {
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), err.Msg());
    }
    err = consumer_->Close();
    if (err.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_ERROR("Close err: Code:{}, MCode:{}, Msg:{}", fmt::underlying(err.Code()), fmt::underlying(err.MCode()),
                    err.Msg());
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), err.Msg());
    }
}
}  // namespace YR
