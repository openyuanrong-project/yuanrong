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
#include <memory>
#include <unordered_map>
#include <vector>

#include "absl/synchronization/mutex.h"

namespace YR {
namespace Libruntime {
class GeneratorIdMap {
public:
    void AddRecord(const std::string &key, const std::string &value)
    {
        absl::MutexLock lock(&mu_);
        records_.emplace(key, value);
    }

    void RemoveRecord(const std::string &key)
    {
        absl::MutexLock lock(&mu_);
        records_.erase(key);
    }

    void GetRecord(const std::string &key, std::string &value)
    {
        absl::MutexLock lock(&mu_);
        if (records_.find(key) != records_.end()) {
            value = records_[key];
        }
    }

    void UpdateRecord(const std::string &key, const std::string &value)
    {
        absl::MutexLock lock(&mu_);
        records_[key] = value;
    }

    void GetRecordKeys(std::vector<std::string> &keys)
    {
        absl::MutexLock lock(&mu_);
        for (auto &record : records_) {
            keys.push_back(record.first);
        }
    }

private:
    absl::Mutex mu_;
    std::unordered_map<std::string, std::string> records_ ABSL_GUARDED_BY(mu_);
};

class GeneratorIdRecorder {
public:
    GeneratorIdRecorder(const std::string &genId, const std::string &srcRuntimeId, std::shared_ptr<GeneratorIdMap> map)
        : genId_(genId), map_(map)
    {
        if (map_ && !genId_.empty()) {
            map_->AddRecord(genId_, srcRuntimeId);
        }
    }

    virtual ~GeneratorIdRecorder(void) {}

private:
    const std::string genId_;
    std::shared_ptr<GeneratorIdMap> map_;
};
}  // namespace Libruntime
}  // namespace YR
