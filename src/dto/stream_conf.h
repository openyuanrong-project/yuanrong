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

#include <string>

#include "src/proto/libruntime.pb.h"

namespace YR {
namespace Libruntime {
struct Element {
    Element(uint8_t *ptr = nullptr, uint64_t size = 0, uint64_t id = ULONG_MAX) : ptr(ptr), size(size), id(id) {}

    ~Element() = default;

    uint8_t *ptr;

    uint64_t size;

    uint64_t id;
};

struct ProducerConf {
    int64_t delayFlushTime = 5;

    int64_t pageSize = 1024 * 1024ul;

    uint64_t maxStreamSize = 100 * 1024 * 1024ul;

    bool autoCleanup = false;

    bool encryptStream = false;

    uint64_t retainForNumConsumers = 0;

    uint64_t reserveSize = 0;

    std::unordered_map<std::string, std::string> extendConfig;

    std::string traceId;
};

struct SubscriptionConfig {
    std::string subscriptionName;

    libruntime::SubscriptionType subscriptionType = libruntime::SubscriptionType::STREAM;

    std::string traceId;

    std::unordered_map<std::string, std::string> extendConfig;

    SubscriptionConfig(std::string subName, const libruntime::SubscriptionType subType)
        : subscriptionName(std::move(subName)), subscriptionType(subType)
    {
    }

    SubscriptionConfig() = default;
};

}  // namespace Libruntime
}  // namespace YR
