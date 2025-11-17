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

#include "api/cpp/include/yr/api/stream.h"
#include "src/libruntime/streamstore/stream_producer_consumer.h"
#include "src/utility/logger/logger.h"

namespace YR {
/**
 * @class StreamProducer
 */
class StreamProducer : public Producer {
public:
    StreamProducer(std::shared_ptr<YR::Libruntime::StreamProducer> producer, const std::string &traceId = "")
    {
        producer_ = std::move(producer);
        traceId_ = traceId;
    }

    /**
     * @brief Sends data to the producer. The data is first placed in the buffer and then flushed based on the
     * configured automatic flush strategy (either after a certain interval or when the buffer is full), or manually via
     * Flush to make the data available to consumers.
     * @param element The Element data to be sent.
     * @throws Exception
     * - **4299**: failed to send element.
     *
     * @snippet{trimleft} stream_example.cpp producer send
     */
    void Send(const Element &element);

    /**
     * @brief Sends data to the producer. The data is first placed in the buffer and then flushed based on the
     * configured automatic flush strategy (either after a certain interval or when the buffer is full), or manually via
     * Flush to make the data available to consumers.
     * @param element The Element data to be sent.
     * @param timeoutMs Optional timeout in milliseconds.
     * @throws Exception
     * - **4299**: failed to send element.
     */
    void Send(const Element &element, int64_t timeoutMs);

    /**
     * @brief Closes the producer, triggering an automatic flush of the buffer and indicating that the buffer will no
     * longer be used. Once closed, the producer cannot be used again.
     * @throws Exception
     * - **4299**: failed to close producer.
     *
     * @snippet{trimleft} stream_example.cpp close producer
     */
    void Close();

private:
    std::shared_ptr<YR::Libruntime::StreamProducer> producer_;
    std::string traceId_;
};

/**
 * @class StreamConsumer
 */
class StreamConsumer : public Consumer {
public:
    StreamConsumer(std::shared_ptr<YR::Libruntime::StreamConsumer> consumer, const std::string &traceId = "")
    {
        consumer_ = std::move(consumer);
        traceId_ = traceId;
    }

    /**
     * @brief Receives data with subscription functionality. The consumer waits for the expected number of elements
     * (`expectNum`) to be received. The call returns when the timeout (`timeoutMs`) is reached or the expected number
     * of elements is received.
     * @param expectNum The expected number of elements to receive.
     * @param timeoutMs The timeout in milliseconds.
     * @param outElements The actual elements received.
     * @throws YR::Exception Thrown in the following cases:
     * - **3003**: the total size exceed the uint64_t max value or the total size exceed the limit.
     * - **4299**: failed to receive element with expectNum.
     *
     * @snippet{trimleft} stream_example.cpp consumer recv
     */
    void Receive(uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements);

    /**
     * @brief Receives data with subscription functionality. The consumer waits for the expected number of elements
     * (`expectNum`) to be received. The call returns when the timeout (`timeoutMs`) is reached or the expected number
     * of elements is received.
     * @param timeoutMs The timeout in milliseconds.
     * @param outElements The actual elements received.
     * @throws YR::Exception Thrown in the following cases:
     * - **3003**: the total size exceed the uint64_t max value or the total size exceed the limit.
     * - **4299**: failed to receive element.
     */
    void Receive(uint32_t timeoutMs, std::vector<Element> &outElements);

    /**
     * @brief Acknowledges that a specific element (identified by `elementId`) has been consumed. This allows workers to
     * determine if all consumers have finished consuming the element, enabling internal memory回收 mechanisms if all
     * consumers have acknowledged it. If not acknowledged, the element will be automatically acknowledged when the
     * consumer exits.
     * @param elementId The ID of the element to acknowledge.
     * @throws YR::Exception
     * - **4299**: failed to ack.
     *
     * @snippet{trimleft} stream_example.cpp consumer recv
     */
    void Ack(uint64_t elementId);

    /**
     * @brief Closes the consumer. Once closed, the consumer cannot be used again.
     * @throws YR::Exception
     * - **4299**: failed to Close consumer.
     *
     * @snippet{trimleft} stream_example.cpp close consumer
     */
    void Close();

private:
    std::shared_ptr<YR::Libruntime::StreamConsumer> consumer_;
    std::string traceId_;
};
}  // namespace YR
