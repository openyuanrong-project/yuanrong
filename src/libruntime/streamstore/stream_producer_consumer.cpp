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

#include "stream_producer_consumer.h"
#include "datasystem/stream/element.h"
#include "datasystem/stream/stream_config.h"
#include "src/dto/config.h"
#include "src/libruntime/utils/datasystem_utils.h"

namespace YR {
namespace Libruntime {
ErrorInfo StreamProducer::Send(const Element &element)
{
    datasystem::Status status = dsProducer->Send(datasystem::Element(element.ptr, element.size, element.id));
    auto msg = "failed to send element, errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo StreamProducer::Send(const Element &element, int64_t timeoutMs)
{
    datasystem::Status status = dsProducer->Send(datasystem::Element(element.ptr, element.size, element.id), timeoutMs);
    auto msg = "failed to send element, errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo StreamProducer::Close()
{
    datasystem::Status status = dsProducer->Close();
    auto msg = "failed to Close producer, errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

std::shared_ptr<datasystem::Producer> &StreamProducer::GetProducer()
{
    return dsProducer;
}

ErrorInfo StreamConsumer::Receive(uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements)
{
    std::vector<datasystem::Element> receiver;
    datasystem::Status status = dsConsumer->Receive(expectNum, timeoutMs, receiver);
    uint64_t totalSize = 0;
    for (auto element : receiver) {
        YRLOG_DEBUG("receive stream with expectNum with element id: {}, size, {}", element.id, element.size);
        outElements.push_back(Element(element.ptr, element.size, element.id));
        if (totalSize > std::numeric_limits<uint64_t>::max() - element.size) {
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                             "the total size exceed the uint64_t max value");
        }
        totalSize += element.size;
    }
    auto msg = "failed to Receive element with expectNum, errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    if (Config::Instance().STREAM_RECEIVE_LIMIT() != 0 && totalSize > Config::Instance().STREAM_RECEIVE_LIMIT()) {
        YRLOG_ERROR("receive size: {} exceeded the limit: {}", totalSize, Config::Instance().STREAM_RECEIVE_LIMIT());
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "the total size exceed the limit");
    }
    return ErrorInfo();
}

ErrorInfo StreamConsumer::Receive(uint32_t timeoutMs, std::vector<Element> &outElements)
{
    std::vector<datasystem::Element> receiver;
    datasystem::Status status = dsConsumer->Receive(timeoutMs, receiver);
    uint64_t totalSize = 0;
    for (auto element : receiver) {
        YRLOG_DEBUG("receive stream with element id: {}, size, {}", element.id, element.size);
        outElements.push_back(Element(element.ptr, element.size, element.id));
        if (totalSize > std::numeric_limits<uint64_t>::max() - element.size) {
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                             "the total size exceed the uint64_t max value");
        }
        totalSize += element.size;
    }
    auto msg = "failed to Receive element, errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    if (Config::Instance().STREAM_RECEIVE_LIMIT() != 0 && totalSize > Config::Instance().STREAM_RECEIVE_LIMIT()) {
        YRLOG_ERROR("receive size: {} exceeded the limit: {}", totalSize, Config::Instance().STREAM_RECEIVE_LIMIT());
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "the total size exceed the limit");
    }
    return ErrorInfo();
}

ErrorInfo StreamConsumer::Ack(uint64_t elementId)
{
    datasystem::Status status = dsConsumer->Ack(elementId);
    auto msg = "failed to Ack, errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo StreamConsumer::Close()
{
    datasystem::Status status = dsConsumer->Close();
    auto msg = "failed to Close consumer, errMsg: " + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

std::shared_ptr<datasystem::Consumer> &StreamConsumer::GetConsumer()
{
    return dsConsumer;
}

}  // namespace Libruntime
}  // namespace YR
