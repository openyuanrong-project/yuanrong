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

#include "invoke_spec.h"
namespace YR {
namespace Libruntime {
struct InstanceOrdering;
class InvokeOrderManager {
public:
    InvokeOrderManager() = default;
    explicit InvokeOrderManager(std::shared_ptr<LibruntimeConfig> config) : libConfig(config) {}
    virtual ~InvokeOrderManager() = default;
    void CreateInstance(std::shared_ptr<InvokeSpec> spec);
    void RegisterInstance(const std::string &instanceId);
    void RegisterInstanceAndUpdateOrder(const std::string &instanceId);
    void RemoveInstance(std::shared_ptr<InvokeSpec> spec);
    void Invoke(std::shared_ptr<InvokeSpec> spec);
    void UpdateUnfinishedSeq(std::shared_ptr<InvokeSpec> spec);
    void NotifyInvokeSuccess(std::shared_ptr<InvokeSpec> spec);
    void UpdateFinishReqSeqNo(const std::string &instanceId, int64_t invokeSeqNo);
    void ClearInsOrderMsg(const std::string &insId, int signal);
    void RemoveGroupInstance(const std::string &instanceId);
    void CreateGroupInstance(const std::string &instanceId);
    void NotifyGroupInstance(const std::string &instanceId);

private:
    std::shared_ptr<InstanceOrdering> ConstuctInstOrder();

    absl::Mutex mu;
    std::unordered_map<std::string, std::shared_ptr<InstanceOrdering>> instances ABSL_GUARDED_BY(mu);
    std::shared_ptr<LibruntimeConfig> libConfig;
};

struct InstanceOrdering {
    InstanceOrdering() = default;
    ~InstanceOrdering() = default;
    std::atomic<int64_t> orderingCounter{0};
    int64_t unfinishedSeqNo = 0;
    std::map<int64_t, std::shared_ptr<InvokeSpec>> finishedUnorderedInvokeSpecs;
    std::set<int64_t> finishedOrderedReqSeqNo;
};
}  // namespace Libruntime
}  // namespace YR