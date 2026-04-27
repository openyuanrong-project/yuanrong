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
 
#include "ordered_execution_manager.h"

#include "src/libruntime/invoke_spec.h"

namespace YR {
namespace Libruntime {
void OrderedExecutionManager::Handle(const libruntime::InvocationMeta &meta, std::function<void()> &&hdlr,
                                     std::string reqId)
{
    auto invokerId = meta.invokerruntimeid();
    if (invokerId.empty()) {
        YRLOG_ERROR("empty invoker id");
        return;
    }

    absl::MutexLock lock(&mu);
    auto it = this->invokers.find(invokerId);
    if (it == this->invokers.end()) {
        it = this->invokers.insert({invokerId, ConstructInvoker()}).first;
    }

    std::shared_ptr<Invoker> invoker = it->second;
    int64_t unfinishedSeqNo = meta.minunfinishedsequenceno();
    if (unfinishedSeqNo > invoker->invokeUnfinishedSeqNo) {
        YRLOG_DEBUG("update invoker {} invoke unfinished sequence No. from {} to {}", invokerId,
                    invoker->invokeUnfinishedSeqNo, unfinishedSeqNo);
        invoker->invokeUnfinishedSeqNo = unfinishedSeqNo;
    }

    int64_t seqNo = meta.invocationsequenceno();
    if (seqNo >= invoker->invokeUnfinishedSeqNo) {
        invoker->waitingInvokeReqs.insert({seqNo, ConstructInokeReq(std::move(hdlr), reqId)});
    }

    while (!invoker->waitingInvokeReqs.empty() &&
           invoker->invokeUnfinishedSeqNo == invoker->waitingInvokeReqs.begin()->first) {
        bool isCancel = false;
        {
            absl::ReaderMutexLock lock(&cancelMu);
            if (this->cancelReqs.find(invoker->waitingInvokeReqs.begin()->second->reqId) != this->cancelReqs.end()) {
                isCancel = true;
            }
        }
        if (isCancel) {
            YRLOG_WARN("request: {} has been set cancel status, no need exec", reqId);
            {
                absl::MutexLock lock(&cancelMu);
                this->cancelReqs.erase(invoker->waitingInvokeReqs.begin()->second->reqId);
            }
        } else {
            this->DoHandle(std::move(invoker->waitingInvokeReqs.begin()->second->hdlr), reqId);
        }
        invoker->waitingInvokeReqs.erase(invoker->waitingInvokeReqs.begin());
        ++invoker->invokeUnfinishedSeqNo;
    }

    YRLOG_DEBUG("current invoker {} waiting unfinished sequence No.: {}", invokerId, invoker->invokeUnfinishedSeqNo);
}

std::shared_ptr<InvokeReq> OrderedExecutionManager::ConstructInokeReq(std::function<void()> &&hdlr, std::string reqId)
{
    return std::make_shared<InvokeReq>(InvokeReq{std::move(hdlr), reqId});
}

std::shared_ptr<Invoker> OrderedExecutionManager::ConstructInvoker()
{
    return std::make_shared<Invoker>();
}

ErrorInfo OrderedExecutionManager::CancelInsFunction(const CancelReqInfo &cancalReqInfo)
{
    auto invokerId = cancalReqInfo.instanceId;
    std::string reqId = cancalReqInfo.requestId;
    if (invokerId.empty()) {
        YRLOG_ERROR("empty invoker id of cancelReqInfo");
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "empty instance id");
    }
    {
        absl::MutexLock lock(&mu);
        auto it = this->invokers.find(invokerId);
        if (it == this->invokers.end()) {
            YRLOG_ERROR("there is no request queue of invoker id: {}", invokerId);
            return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "invalid instance id");
        }
    }
    {
        absl::MutexLock lock(&cancelMu);
        this->cancelReqs[reqId] = true;
    }
    return ErrorInfo();
}

}  // namespace Libruntime
}  // namespace YR