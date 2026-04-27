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
#include <atomic>
#include <string>
#include <thread>
#include "src/dto/buffer.h"
namespace YR {
namespace Libruntime {
struct AccelerateMsgQueueHandle {
    int worldSize;
    int rank;
    int maxChunkBytes;
    int maxChunks;
    std::string name;
    bool isAsync;
    std::string ToJson() const;
    static AccelerateMsgQueueHandle FromJson(const std::string &data);
};

class ShmRingBuffer {
public:
    ShmRingBuffer(int worldSize, int maxChunks, int maxChunkBytes, const std::shared_ptr<Buffer> buffer)
        : worldSize_(worldSize), maxChunks_(maxChunks), maxChunkBytes_(maxChunkBytes)
    {
        metadataSize_ = worldSize + 1;
        metadataOffset_ = maxChunks * maxChunkBytes;
        buffer_ = buffer;
        data_ = static_cast<char *>(buffer->MutableData());
    };
    ~ShmRingBuffer() = default;
    char *GetMetadata(int currentId);
    std::shared_ptr<Buffer> GetData(int currentId);

private:
    std::shared_ptr<Buffer> buffer_;
    char *data_{nullptr};
    int worldSize_{0};
    int maxChunks_{0};
    int maxChunkBytes_{0};
    int metadataOffset_{0};
    int metadataSize_{0};
    int dataOffset{0};
};

class AccelerateMsgQueue {
public:
    AccelerateMsgQueue(const AccelerateMsgQueueHandle &handle,
                       const std::shared_ptr<Buffer> buffer);
    ~AccelerateMsgQueue() = default;
    static AccelerateMsgQueue CreateFromHandle(const AccelerateMsgQueueHandle &handle,
                                               const std::shared_ptr<Buffer> buffer);
    std::shared_ptr<Buffer> Dequeue();
    std::shared_ptr<Buffer> AcquireRead();
    void SetReadFlag();
    void Stop();

private:
    int currentId_{0};
    bool rank_{0};
    int maxChunks_{0};
    std::shared_ptr<ShmRingBuffer> buffer_;
    std::atomic<bool> stop_{false};
};
};  // namespace Libruntime
};  // namespace YR
