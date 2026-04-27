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

#include "datasystem/stream/consumer.h"
#include "datasystem/stream/producer.h"
#include "src/dto/stream_conf.h"
#include "src/libruntime/err_type.h"

namespace YR {
namespace Libruntime {

class StreamProducer {
public:
    virtual ErrorInfo Send(const Element &element);

    virtual ErrorInfo Send(const Element &element, int64_t timeoutMs);

    virtual ErrorInfo Close();

    std::shared_ptr<datasystem::Producer> &GetProducer();

private:
    std::shared_ptr<datasystem::Producer> dsProducer;
};

class StreamConsumer {
public:
    virtual ErrorInfo Receive(uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements);

    virtual ErrorInfo Receive(uint32_t timeoutMs, std::vector<Element> &outElements);

    virtual ErrorInfo Ack(uint64_t elementId);

    virtual ErrorInfo Close();

    std::shared_ptr<datasystem::Consumer> &GetConsumer();

private:
    std::shared_ptr<datasystem::Consumer> dsConsumer;
};
}  // namespace Libruntime
}  // namespace YR