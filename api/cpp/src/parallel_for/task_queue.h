/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: implementation of lock-free queues
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

#include <cstdint>
#include <vector>
#include "complier.h"

namespace YR {
namespace Parallel {
constexpr int H_SUCCESS = 0;
constexpr int H_FAIL = -1;
const uint32_t DEFAULT_QUEUE_SIZE = 1024;

struct RingHeadTail {
    volatile uint32_t head;
    volatile uint32_t tail;
};

class TaskQueue {
public:
    explicit TaskQueue(uint32_t queueSize = DEFAULT_QUEUE_SIZE)
    {
        ring.resize(queueSize);
        size = static_cast<uint32_t>(queueSize);
        mask = static_cast<uint32_t>(queueSize) - 1;
        prod.head = cons.head = 0;
        prod.tail = cons.tail = 0;
    };
    ~TaskQueue(){};

    inline int EnqueueParallel(intptr_t data);
    inline int DequeueParallel(intptr_t *data);

private:
    RingHeadTail prod;
    RingHeadTail cons;
    uint32_t size;
    uint32_t mask;
    std::vector<intptr_t> ring;
};

inline int TaskQueue::EnqueueParallel(intptr_t data)
{
    uint32_t prodHead;
    uint32_t prodNext;
    uint32_t consTail;
    uint32_t freeEntries;
    bool success;
    do {
        prodHead = prod.head;
        consTail = cons.tail;
        freeEntries = size + consTail - prodHead;
        if (UNLIKELY(freeEntries == 0)) {
            return H_FAIL;
        }
        prodNext = prodHead + 1;
        success = AtomicCmpset32(&prod.head, prodHead, prodNext);
    } while (UNLIKELY(!success));
    ring[prodHead & this->mask] = data;
    do {
        success = AtomicCmpset32(&prod.tail, prodHead, prodNext);
    } while (UNLIKELY(!success));
    return H_SUCCESS;
}

inline int TaskQueue::DequeueParallel(intptr_t *data)
{
    uint32_t consHead;
    uint32_t prodTail;
    uint32_t entries;
    uint32_t consNext;
    bool success;
    intptr_t res;

    do {
        consHead = cons.head;
        prodTail = prod.tail;
        entries = (prodTail - consHead);
        if (UNLIKELY(entries == 0)) {
            return H_FAIL;
        }
        consNext = consHead + 1;
        success = AtomicCmpset32(&cons.head, consHead, consNext);
    } while (UNLIKELY(!success));
    res = ring[consHead & this->mask];
    do {
        success = AtomicCmpset32(&cons.tail, consHead, consNext);
    } while (UNLIKELY(!success));
    *data = res;
    return H_SUCCESS;
}
}  // namespace Parallel
}  // namespace YR