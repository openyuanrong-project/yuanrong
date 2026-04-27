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
#include "execution_manager.h"

namespace YR {
namespace Libruntime {
struct InvokeReq;
struct Invoker;
class OrderedExecutionManager : public ExecutionManager {
public:
    OrderedExecutionManager(size_t concurrency, std::function<void(std::function<void(void)> &&)> submitHook)
        : ExecutionManager(concurrency, submitHook)
    {
    }
    virtual ~OrderedExecutionManager() = default;

    void Handle(const libruntime::InvocationMeta &meta, std::function<void()> &&hdlr, std::string reqId = "") override;
    ErrorInfo CancelInsFunction(const CancelReqInfo &cancalReqInfo) override;
private:
    std::shared_ptr<InvokeReq> ConstructInokeReq(std::function<void()> &&hdlr, std::string reqId);
    std::shared_ptr<Invoker> ConstructInvoker();
    absl::Mutex mu;
    std::unordered_map<std::string, std::shared_ptr<Invoker>> invokers ABSL_GUARDED_BY(mu);
    absl::Mutex cancelMu;
    std::map<std::string, bool> cancelReqs ABSL_GUARDED_BY(cancelMu);
};

struct InvokeReq {
    std::function<void()> hdlr;
    std::string reqId;
    InvokeReq(std::function<void()> hdlr_, std::string reqId_) : hdlr(std::move(hdlr_)), reqId(std::move(reqId_)) {}
};

struct Invoker {
    int64_t invokeUnfinishedSeqNo = 0;
    std::map<int64_t, std::shared_ptr<InvokeReq>> waitingInvokeReqs;
};
}  // namespace Libruntime
}  // namespace YR