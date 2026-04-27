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

#include "src/dto/buffer.h"
#include "src/dto/data_object.h"
#include "src/dto/invoke_options.h"
#include "src/dto/status.h"
#include "src/libruntime/err_type.h"
#include "src/dto/accelerate.h"

namespace YR {
namespace Libruntime {

struct LibruntimeOptions {
    // load function handler: called at the beginning of initialization, before 'init call'
    using LoadFunctionCallback = std::function<ErrorInfo(const std::vector<std::string> &codePaths)>;
    LoadFunctionCallback loadFunctionCallback;

    // call handler: for both 'init-call' and 'normal-call'.
    // 'init-call': invokeType could either be CREATE_INSTANCE or CREATE_NORMAL_FUNCTION_INSTANCE corresponding to
    // Libruntime API CreateInstance and InvokeByFunctionName
    // 'call': invokeType could either be INVOKE_MEMBER_FUNCTION or INVOKE_NORMAL_FUNCTION corresponding to Libruntime
    // API InvokeByInstanceId and InvokeByFunctionName
    // API layer must record FunctionMeta->ApiType when 'init-call' handler executes, so that it can handle Checkpoint,
    // Recover, Shutdown, Signal properly.
    using FunctionExecuteCallback =
        std::function<ErrorInfo(const FunctionMeta &function, const libruntime::InvokeType invokeType,
                                const std::vector<std::shared_ptr<DataObject>> &rawArgs,
                                std::vector<std::shared_ptr<DataObject>> &returnValues)>;
    FunctionExecuteCallback functionExecuteCallback;

    // checkpoint handler: checkpoint API layer's internal state, so that it can be recover laterly in different runtime
    // process.
    using CheckpointCallback = std::function<ErrorInfo(const std::string &checkpointId, std::shared_ptr<Buffer> &data)>;
    CheckpointCallback checkpointCallback;

    // recovery handler: restore state that previously checkpointed
    using RecoverCallback = std::function<ErrorInfo(std::shared_ptr<Buffer> data)>;
    RecoverCallback recoverCallback;

    // shutdown handler: gracefully shutdown
    using ShutdownCallback = std::function<ErrorInfo(uint64_t gracePeriodSeconds)>;
    ShutdownCallback shutdownCallback;

    // prepare snapshot handler: called before taking snapshot
    using PrepareSnapCallback = std::function<ErrorInfo(void)>;
    PrepareSnapCallback prepareSnapCallback;

    // snap started handler: called after restoring from snapshot
    using SnapStartedCallback = std::function<ErrorInfo(void)>;
    SnapStartedCallback snapStartedCallback;

    // refresh env handler: called after snapStarted handler completes to refresh environment
    using RefreshEnvCallback = std::function<void(void)>;
    RefreshEnvCallback refreshEnvCallback;

    // signal handler: handle user defined signal
    using SignalCallback = std::function<ErrorInfo(int sigNo, std::shared_ptr<Buffer> payload)>;
    SignalCallback signalCallback;

    // health check handler: check health of function
    using HealthCheckCallback = std::function<ErrorInfo(void)>;
    HealthCheckCallback healthCheckCallback;
    using AccelerateCallback = std::function<ErrorInfo(const AccelerateMsgQueueHandle &, AccelerateMsgQueueHandle &)>;
    AccelerateCallback accelerateCallback;
};

}  // namespace Libruntime
}  // namespace YR