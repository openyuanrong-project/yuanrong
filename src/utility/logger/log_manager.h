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

#include <condition_variable>
#include <functional>
#include <mutex>
#include <string>
#include <thread>
#include <unordered_map>
#include "common.h"
#include "logger.h"

namespace YR {
namespace utility {

const uint32_t DEFAULT_LOG_HANDLER_INTERVAL = 30;  // seconds

class LogManager {
public:
    LogManager() : interval_(DEFAULT_LOG_HANDLER_INTERVAL) {}
    ~LogManager()
    {
        StopRollingCompress();
    }
    void StartRollingCompress(const std::function<void(LogParam &)> &func);

    void StopRollingCompress() noexcept;
    bool AddLogParam(const std::string &rtCtx, LogParam logParam);

private:
    // Allows multiple tasks to run in parallel
    void CronTask(const std::function<void(LogParam &)> &func);
    std::unordered_map<std::string, LogParam> logParams_;
    uint32_t interval_;

    std::thread rollingCompressThread_;
    std::condition_variable rcCond_;
    std::mutex mtx_;

    enum State { INITED = 0, RUNNING, STOPPED };
    State state_ = INITED;
};

}  // namespace utility
}  // namespace YR
