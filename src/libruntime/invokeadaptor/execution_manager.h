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
#include "src/libruntime/fsclient/fs_intf.h"
#include "src/libruntime/invoke_spec.h"
#include "boost/asio.hpp"

namespace YR {
namespace Libruntime {
class ExecutionManager {
public:
    ExecutionManager(size_t concurrency, std::function<void(std::function<void(void)> &&)> submitHook);
    virtual ~ExecutionManager() = default;

    virtual void Handle(const libruntime::InvocationMeta &meta, std::function<void()> &&hdlr,
                        std::string reqId = "") = 0;
    virtual ErrorInfo CancelInsFunction(const CancelReqInfo &cancalReqInfo) = 0;
    void ErasePendingThread(const std::string &reqId);
    bool isMultipleConcurrency();
    ErrorInfo DoInit(size_t concurrency);

protected:
    void DoHandle(std::function<void()> &&hdlr, std::string reqId = "");

private:
    std::shared_ptr<boost::asio::thread_pool> callExecutor_;
    bool useCustomExecutor;
    size_t concurrency_ = 1;
    std::function<void(std::function<void(void)> &&)> customExecutorSubmit;
};
}  // namespace Libruntime
}  // namespace YR
