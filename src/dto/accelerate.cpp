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

#include "accelerate.h"
#include <string>
#include "json.hpp"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
using json = nlohmann::json;
std::string AccelerateMsgQueueHandle::ToJson() const
{
    json handle;
    handle["world_size"] = worldSize;
    handle["rank"] = rank;
    handle["max_chunk_bytes"] = maxChunkBytes;
    handle["max_chunks"] = maxChunks;
    handle["name"] = name;
    handle["is_async"] = isAsync;
    try {
        return handle.dump();
    } catch (std::exception &e) {
        YRLOG_ERROR("dump handle json failed, error: {}", e.what());
        return std::string();
    }
}

AccelerateMsgQueueHandle AccelerateMsgQueueHandle::FromJson(const std::string &data)
{
    AccelerateMsgQueueHandle handle;
    try {
        json j = json::parse(data);
        handle.worldSize = j["world_size"].get<int>();
        handle.rank = j["rank"].get<int>();
        handle.maxChunkBytes = j["max_chunk_bytes"].get<int>();
        handle.maxChunks = j["max_chunks"].get<int>();
        handle.name = j["name"].get<std::string>();
        handle.isAsync = j["is_async"].get<bool>();
    } catch (std::exception &e) {
        YRLOG_ERROR("parse payload json failed, error: {}", e.what());
    }
    return handle;
}

char *ShmRingBuffer::GetMetadata(int currentId)
{
    auto start = metadataOffset_ + currentId * metadataSize_;
    return data_ + start;
}

std::shared_ptr<Buffer> ShmRingBuffer::GetData(int currentId)
{
    auto start = dataOffset + currentId * maxChunkBytes_;
    return std::make_shared<NativeBuffer>(data_ + start, maxChunkBytes_);
}

AccelerateMsgQueue::AccelerateMsgQueue(const AccelerateMsgQueueHandle &handle, const std::shared_ptr<Buffer> buffer)
{
    rank_ = handle.rank;
    maxChunks_ = handle.maxChunks;
    buffer_ = std::make_shared<ShmRingBuffer>(handle.worldSize, handle.maxChunks, handle.maxChunkBytes, buffer);
}

std::shared_ptr<Buffer> AccelerateMsgQueue::Dequeue()
{
    return AcquireRead();
}

std::shared_ptr<Buffer> AccelerateMsgQueue::AcquireRead()
{
    while (!stop_) {
        auto metadata = buffer_->GetMetadata(currentId_);
        auto writeFlag = static_cast<bool>(metadata[0]);
        auto readFlag = static_cast<bool>(metadata[rank_ + 1]);
        if (writeFlag && !readFlag) {
            return buffer_->GetData(currentId_);
        }
        std::this_thread::yield();
    }
    return nullptr;
}

void AccelerateMsgQueue::SetReadFlag()
{
    auto metadata = buffer_->GetMetadata(currentId_);
    metadata[rank_ + 1] = 1;
    currentId_ = (currentId_ + 1) % maxChunks_;
}

void AccelerateMsgQueue::Stop()
{
    stop_ = true;
}
};  // namespace Libruntime
};  // namespace YR
