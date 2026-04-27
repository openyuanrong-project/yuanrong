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

#include "src/libruntime/err_type.h"
#include "src/libruntime/utils/utils.h"
#include "src/libruntime/objectstore/memory_store.h"
#include "src/libruntime/streamstore/stream_store.h"

namespace YR {
namespace Libruntime {
struct DedupState {
    int64_t timestamp;
    int count;
    std::string line;
    std::unordered_set<std::string> sources;
};

class DriverLogReceiver {
public:
    DriverLogReceiver();
    ~DriverLogReceiver();
    void Init(std::shared_ptr<StreamStore> store, std::string &jobID, bool dedup);
    void Stop();

private:
    void Receive();
    void Deduplicate(std::shared_ptr<std::vector<std::string>> lines);
    std::pair<std::string, std::string> ParseLine(const std::string& input);
    std::string Format(const DedupState &state);
    void Flush(int64_t now);
    std::shared_ptr<StreamStore> dsStreamStore_;
    std::shared_ptr<StreamConsumer> consumer_;
    std::thread receiverThread_;
    std::atomic<bool> stopped_{false};
    std::string jobId_;
    bool dedup_ = false;
    mutable std::mutex mtx;
    std::unordered_map<std::string, DedupState> recent_;
};
}  // namespace Libruntime
}  // namespace YR
