/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

#ifndef AGENT_SESSION_MANAGER_H
#define AGENT_SESSION_MANAGER_H

#include <condition_variable>
#include <memory>
#include <mutex>
#include <string>
#include <unordered_map>

#include "src/dto/buffer.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_config.h"
#include "src/libruntime/runtime_context.h"
#include "src/proto/libruntime.pb.h"

namespace YR {
namespace Libruntime {

struct AgentSessionValue {
    std::string sessionID;
    std::string sessionKey;
    std::string sessionData;
};

struct AgentSessionContext {
    std::mutex mutex;
    AgentSessionValue value;
    bool loaded = false;
};

enum class WaitState {
    IDLE,
    WAITING,
    NOTIFIED,
    TIMEOUT,
    INTERRUPTED
};

struct WaitNotifyContext {
    std::mutex mutex;
    std::condition_variable cv;
    WaitState state = WaitState::IDLE;
    std::shared_ptr<Buffer> notifyData;

    void Reset()
    {
        std::lock_guard<std::mutex> lock(mutex);
        state = WaitState::IDLE;
        notifyData = nullptr;
    }
};

class AgentSessionManager {
public:
    AgentSessionManager(std::shared_ptr<LibruntimeConfig> config, std::shared_ptr<RuntimeContext> runtimeContext);

    ErrorInfo AcquireInvokeSession(const std::string &sessionId, const libruntime::MetaData &meta);

    ErrorInfo PersistAndReleaseInvokeSession(const std::string &sessionId);

    std::pair<std::string, ErrorInfo> LoadCurrentSession(const std::string &sessionId);

    ErrorInfo UpdateCurrentSession(const std::string &sessionId, const std::string &sessionData);

    ErrorInfo SetSessionInterrupted(const std::string &sessionId);

    bool IsSessionInterrupted(const std::string &sessionId);

    std::pair<ErrorInfo, std::shared_ptr<Buffer>> Wait(const std::string &sessionId, int64_t timeoutMs);

    ErrorInfo Notify(const std::string &sessionId, std::shared_ptr<Buffer> data);

private:
    std::shared_ptr<AgentSessionContext> GetOrCreateSessionContext(const std::string &sessionKey);

    std::shared_ptr<AgentSessionContext> GetActiveSessionContext(const std::string &sessionId);

    void BindActiveSessionContext(const std::string &sessionId, const std::shared_ptr<AgentSessionContext> &sessionCtx);

    std::shared_ptr<AgentSessionContext> UnbindActiveSessionContext(const std::string &sessionId);

    ErrorInfo EnsureLoaded(const std::shared_ptr<AgentSessionContext> &sessionCtx, const std::string &sessionId,
                           const libruntime::MetaData &meta);

    ErrorInfo Persist(const std::shared_ptr<AgentSessionContext> &sessionCtx);

    std::string BuildSessionKey(const libruntime::MetaData &meta, const std::string &sessionId) const;

    std::string BuildDefaultSession(const std::string &sessionId) const;

    std::shared_ptr<LibruntimeConfig> librtConfig_;
    std::shared_ptr<RuntimeContext> runtimeContext_;
    std::mutex sessionMapMtx_;
    std::unordered_map<std::string, std::shared_ptr<AgentSessionContext>> sessionMap_;
    std::mutex activeSessionMapMtx_;
    std::unordered_map<std::string, std::shared_ptr<AgentSessionContext>> activeSessionMap_;

    std::mutex waitNotifyMtx_;
    std::unordered_map<std::string, std::shared_ptr<WaitNotifyContext>> waitNotifyMap_;

    std::shared_ptr<WaitNotifyContext> GetOrCreateWaitNotifyContext(const std::string &sessionId);
};

}  // namespace Libruntime
}  // namespace YR

#endif  // AGENT_SESSION_MANAGER_H
