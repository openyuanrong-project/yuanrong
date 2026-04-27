/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.
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
class GeneralExecutionManager : public ExecutionManager {
public:
    GeneralExecutionManager(size_t concurrency, std::function<void(std::function<void(void)> &&)> submitHook)
        : ExecutionManager(concurrency, submitHook)
    {
    }
    virtual ~GeneralExecutionManager() = default;

    void Handle(const libruntime::InvocationMeta &meta, std::function<void()> &&hdlr, std::string reqId = "") override;
    ErrorInfo CancelInsFunction(const CancelReqInfo &cancalReqInfo) override;
};
}  // namespace Libruntime
}  // namespace YR