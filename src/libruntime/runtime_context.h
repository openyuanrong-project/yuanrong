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
#include <string>

#include "absl/synchronization/mutex.h"

#include "src/utility/id_generator.h"

namespace YR {
namespace Libruntime {
class RuntimeContext {
public:
    RuntimeContext() = default;
    RuntimeContext(const std::string &jobId) : jobId_(jobId)
    {
        SetJobIdThreadlocal(jobId);
    }

    std::string GetJobId(void);

    void SetInvokingRequestId(const std::string &reqId);

    std::string GetInvokingRequestId(void);

    void SetJobIdThreadlocal(const std::string &jobId);

    std::string GetJobIdThreadlocal(void);

private:
    mutable absl::Mutex mu_;

    std::string jobId_ ABSL_GUARDED_BY(mu_);
};
}  // namespace Libruntime
}  // namespace YR