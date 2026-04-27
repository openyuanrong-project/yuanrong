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

#include "src/libruntime/runtime_context.h"

namespace YR {
namespace Libruntime {
class RuntimeThreadContext {
public:
    void SetInvokingRequestId(const std::string &reqId)
    {
        requestId = reqId;
    }

    std::string GetInvokingRequestId(void)
    {
        return requestId;
    }

    void SetJobId(const std::string &jobId)
    {
        this->jobId_ = jobId;
    }

    std::string GetJobId(void)
    {
        absl::ReaderMutexLock l(&mu_);
        return this->jobId_;
    }

private:
    std::string requestId;

    mutable absl::Mutex mu_;

    std::string jobId_ ABSL_GUARDED_BY(mu_);
};

thread_local std::unique_ptr<RuntimeThreadContext> threadContext = nullptr;

static RuntimeThreadContext &GetThreadContext(void)
{
    if (!threadContext) {
        threadContext = std::make_unique<RuntimeThreadContext>();
    }
    return *threadContext;
}

std::string RuntimeContext::GetJobId(void)
{
    absl::ReaderMutexLock l(&mu_);
    return this->jobId_;
}

void RuntimeContext::SetInvokingRequestId(const std::string &reqId)
{
    GetThreadContext().SetInvokingRequestId(reqId);
}

std::string RuntimeContext::GetInvokingRequestId(void)
{
    return GetThreadContext().GetInvokingRequestId();
}

void RuntimeContext::SetJobIdThreadlocal(const std::string &jobId)
{
    GetThreadContext().SetJobId(jobId);
}

std::string RuntimeContext::GetJobIdThreadlocal(void)
{
    return GetThreadContext().GetJobId();
}
}  // namespace Libruntime
}  // namespace YR