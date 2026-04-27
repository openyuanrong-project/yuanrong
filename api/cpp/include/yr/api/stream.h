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

#pragma once

#include <climits>
#include <string>
#include <unordered_map>
#include <vector>

namespace YR {
enum SubscriptionType { STREAM, ROUND_ROBIN, KEY_PARTITIONS };

/**
 * @struct Element
 */
struct Element {
    Element(uint8_t *ptr = nullptr, uint64_t size = 0, uint64_t id = ULONG_MAX) : ptr(ptr), size(size), id(id) {}

    ~Element() = default;
    /**
     * @brief Pointer to the data
     */
    uint8_t *ptr;

    /**
     * @brief Size of the data
     */
    uint64_t size;

    /**
     * @brief ID of the Element
     */
    uint64_t id;
};

/**
 * @struct ProducerConf
 */
struct ProducerConf {
    /**
     * @brief After sending, the producer will delay for the specified duration before triggering a flush.
     * <0: Do not automatically flush; 0: Flush immediately; otherwise, the delay duration in milliseconds. Default: 5.
     */
    int64_t delayFlushTime = 5;

    /**
     * @brief Specifies the buffer page size for the producer, in bytes (B). When a page is full, it will trigger a
     * flush. Default: 1 MB. Must be greater than 0 and a multiple of 4 KB.
     */
    int64_t pageSize = 1024 * 1024ul;

    /**
     * @brief Specifies the maximum shared memory size that the stream can use on a worker, in bytes (B).
     * Default: 100 MB. Range: [64 KB, size of worker shared memory].
     */
    uint64_t maxStreamSize = 100 * 1024 * 1024ul;

    /**
     * @brief Specifies whether to enable automatic cleanup for the stream. Default: false (disabled).
     * When the last producer/consumer exits, the stream will be automatically cleaned up.
     */
    bool autoCleanup = false;

    /**
     * @brief Specifies whether to enable content encryption for the stream. Default: false (disabled).
     */
    bool encryptStream = false;

    /**
     * @brief Specifies how many consumers should retain the producer's data. Default: 0.
     * If set to 0, data will not be retained if there are no consumers.
     * This parameter is only effective for the first consumer created, and the current valid range is
     * [0, 1]. Multiple consumers are not supported. A consumer created after the producer may not receive data.
     */
    uint64_t retainForNumConsumers = 0;

    /**
     * @brief Specifies the reserved memory size, in bytes (B). When creating a producer, it will attempt to reserve
     * reserveSize bytes of memory. If reservation fails, an exception will be thrown. reserveSize must be an integer
     * multiple of pageSize and within the range [0, maxStreamSize]. If reserveSize is 0, it will be set to pageSize by
     * default. Default: 0.
     */
    uint64_t reserveSize = 0;

    /**
     * @brief Extended configuration for the producer. Common configuration items include:
     * "STREAM_MODE": The stream mode, which can be "MPMC", "MPSC", or "SPSC". Default: "MPMC". If an unsupported mode
     * is specified, an exception will be thrown. MPMC represents multi-producer multi-consumer, MPSC represents
     * multi-producer single-consumer, and SPSC represents single-producer single-consumer. If MPSC or SPSC is selected,
     * the data system will enable multi-stream shared page functionality internally.
     */
    std::unordered_map<std::string, std::string> extendConfig;

    /**
     * @brief Custom trace ID for troubleshooting and performance optimization. Only supported in the cloud; settings
     * outside the cloud will not take effect. Maximum length: 36. Valid characters must match the regular expression:
     * ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     */
    std::string traceId;
};

/**
 * @struct SubscriptionConfig
 */
struct SubscriptionConfig {
    /**
     * @brief Subscription name
     */
    std::string subscriptionName;

    /**
     * @brief Subscription type, including three types: STREAM, ROUND_ROBIN, and KEY_PARTITIONS.
     * STREAM indicates that a single consumer in a subscription group consumes the stream.
     * ROUND_ROBIN indicates that multiple consumers in a subscription group consume the stream in a round-robin
     * load-balancing manner. KEY_PARTITIONS indicates that multiple consumers in a subscription group consume the
     * stream in a key-partitioned load-balancing manner. Currently, only the STREAM type is supported; other types are
     * temporarily unsupported. Default subscription type: STREAM.
     */
    SubscriptionType subscriptionType = SubscriptionType::STREAM;

    /**
     * @brief Extended configuration for SubscriptionConfig.
     */
    std::unordered_map<std::string, std::string> extendConfig;

    /**
     * @brief Custom trace ID for troubleshooting and performance optimization. Only supported in the cloud; settings
     * outside the cloud will not take effect. Maximum length: 36. Valid characters must match the regular expression:
     * ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     */
    std::string traceId;

    /**
     * @brief Constructor for SubscriptionConfig.
     * @param subName Subscription name
     * @param subType Subscription type
     */
    SubscriptionConfig(std::string subName, const SubscriptionType subType)
        : subscriptionName(std::move(subName)), subscriptionType(subType)
    {
    }

    SubscriptionConfig() = default;

    SubscriptionConfig(const SubscriptionConfig &other) = default;

    SubscriptionConfig &operator=(const SubscriptionConfig &other) = default;

    SubscriptionConfig(SubscriptionConfig &&other) noexcept
    {
        subscriptionName = std::move(other.subscriptionName);
        subscriptionType = other.subscriptionType;
    }

    SubscriptionConfig &operator=(SubscriptionConfig &&other) noexcept
    {
        subscriptionName = std::move(other.subscriptionName);
        subscriptionType = other.subscriptionType;
        return *this;
    }

    bool operator==(const SubscriptionConfig &config) const
    {
        return subscriptionName == config.subscriptionName && subscriptionType == config.subscriptionType;
    }

    bool operator!=(const SubscriptionConfig &config) const
    {
        return !(*this == config);
    }
};

class Producer {
public:
    virtual void Send(const Element &element) = 0;

    virtual void Send(const Element &element, int64_t timeoutMs) = 0;

    virtual void Close() = 0;
};

class Consumer {
public:
    virtual void Receive(uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements) = 0;

    virtual void Receive(uint32_t timeoutMs, std::vector<Element> &outElements) = 0;

    virtual void Ack(uint64_t elementId) = 0;

    virtual void Close() = 0;
};

}  // namespace YR