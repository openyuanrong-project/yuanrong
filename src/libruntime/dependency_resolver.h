/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
#include <functional>
#include <future>
#include <memory>
#include <string>
#include <unordered_map>
#include <vector>

#include "absl/synchronization/mutex.h"

#include "src/dto/invoke_options.h"
#include "src/dto/status.h"
#include "src/libruntime/invoke_spec.h"
#include "src/libruntime/objectstore/memory_store.h"

namespace YR {
namespace Libruntime {

struct DependencyState {
    DependencyState(size_t idNums, std::function<void(const ErrorInfo &)> callback)
        : dependenciesRemaining(idNums), onComplete(callback){};
    size_t dependenciesRemaining;
    std::function<void(const ErrorInfo &)> onComplete;
};

class DependencyResolver {
public:
    DependencyResolver(std::shared_ptr<MemoryStore> memStore) : memoryStore(memStore) {}

    ~DependencyResolver() = default;

    void ResolveDependencies(std::shared_ptr<InvokeSpec> spec, std::function<void(const ErrorInfo &)> onComplete);

private:
    std::shared_ptr<MemoryStore> memoryStore;
    absl::Mutex mu;
    std::unordered_map<std::string, std::unique_ptr<DependencyState>> dependencyState ABSL_GUARDED_BY(mu);
};
}  // namespace Libruntime
}  // namespace YR