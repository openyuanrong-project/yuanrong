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

#include "dependency_resolver.h"

namespace YR {
namespace Libruntime {
void DependencyResolver::ResolveDependencies(std::shared_ptr<InvokeSpec> spec,
                                             std::function<void(const ErrorInfo &)> onComplete)
{
    std::vector<std::string> unfinishObjIds;
    for (auto &arg : spec->invokeArgs) {
        if (arg.isRef) {
            unfinishObjIds.push_back(arg.objId);
        }

        for (auto &id : arg.nestedObjects) {
            unfinishObjIds.push_back(id);
        }
    }

    if (spec->invokeType == libruntime::InvokeType::InvokeFunction) {
        YRLOG_DEBUG("Invoke instance member function, instance ID: {}", spec->instanceId);
        unfinishObjIds.push_back(spec->instanceId);
    }

    YRLOG_DEBUG("unfinished object size: {}", unfinishObjIds.size());
    if (unfinishObjIds.empty()) {
        onComplete(ErrorInfo());
        return;
    }

    {
        absl::MutexLock lock(&mu);
        dependencyState.emplace(spec->requestId, std::make_unique<DependencyState>(unfinishObjIds.size(), onComplete));
    }
    for (auto &id : unfinishObjIds) {
        YRLOG_DEBUG("Register object ID {} for request ID {}", id, spec->requestId);
        memoryStore->AddReadyCallback(id, [this, spec, id](const ErrorInfo &err) {
            YRLOG_DEBUG("Object ID {} ready for request ID {} ", id, spec->requestId);
            std::unique_ptr<DependencyState> resolved_dependency_state = nullptr;
            {
                absl::MutexLock lock(&mu);
                auto it = this->dependencyState.find(spec->requestId);
                if (it == this->dependencyState.end()) {
                    YRLOG_DEBUG("Dependency state not found, request ID {}", spec->requestId);
                    return;
                }
                if (--it->second->dependenciesRemaining == 0 || !err.OK()) {
                    YRLOG_DEBUG("Dependency resolved, request ID {}", spec->requestId);
                    resolved_dependency_state = std::move(it->second);
                    this->dependencyState.erase(it);
                }
            }
            if (resolved_dependency_state) {
                resolved_dependency_state->onComplete(err);
            }
        });
    }
}
}  // namespace Libruntime
}  // namespace YR