/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
#include "memory.h"
#include <securec.h>
#include <cstddef>
#include <cstdint>
#include <cstring>
#include <thread>
#include <vector>
#include "logger/logger.h"

namespace YR {
namespace utility {
// Round down to the block boundary
static inline const uint8_t *FloorToBlock(const uint8_t *ptr, size_t blockSize)
{
    auto rawValue = reinterpret_cast<uintptr_t>(ptr);
    rawValue -= rawValue % blockSize;  // Modulo alignment
    return reinterpret_cast<const uint8_t *>(rawValue);
}

void CopyInParallel(uint8_t *dst, const uint8_t *src, int64_t totalBytes, size_t blockSize)
{
    const static int threadCount = 6;
    if (blockSize == 0 || totalBytes <= 0) {
        return;
    }

    // Calculate alignment boundary
    const uint8_t *leftAligned = FloorToBlock(src + blockSize - 1, blockSize);
    const uint8_t *rightAligned = FloorToBlock(src + totalBytes, blockSize);

    int64_t totalBlocks = (rightAligned - leftAligned) / blockSize;

    // Align to the right at thread multiples
    size_t remainderBlocks = totalBlocks % threadCount;
    rightAligned -= remainderBlocks * blockSize;

    // Interval size for splitting
    int64_t headSize = leftAligned - src;
    int64_t bodySize = rightAligned - leftAligned;
    int64_t chunkSize = bodySize / threadCount;
    int64_t tailSize = (src + totalBytes) - rightAligned;

    // Create a thread pool
    std::vector<std::thread> threads;
    threads.reserve(threadCount);

    for (int i = 0; i < threadCount; ++i) {
        const uint8_t *segmentSrc = leftAligned + i * chunkSize;
        uint8_t *segmentDst = dst + headSize + i * chunkSize;
        threads.emplace_back([segmentDst, segmentSrc, chunkSize]() {
            if (memcpy_s(segmentDst, chunkSize, segmentSrc, chunkSize) != 0) {
                YRLOG_ERROR("Failed to memcpy_s.");
            }
        });
    }

    // The main thread processes prefixes and suffixes.
    if (headSize > 0) {
        if (memcpy_s(dst, headSize, src, headSize) != 0) {
            YRLOG_ERROR("Failed to memcpy_s.");
        }
    }
    if (tailSize > 0) {
        if (memcpy_s(dst + headSize + bodySize, tailSize, rightAligned, tailSize) != 0) {
            YRLOG_ERROR("Failed to memcpy_s.");
        }
    }

    // Wait for the child thread to complete
    for (auto &t : threads) {
        t.join();
    }
}
}  // namespace utility
}  // namespace YR