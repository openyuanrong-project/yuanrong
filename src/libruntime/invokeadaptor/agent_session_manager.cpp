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

#include "agent_session_manager.h"

#include <nlohmann/json.hpp>

#include "src/dto/buffer.h"
#include "src/libruntime/libruntime.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/libruntime/utils/constants.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
using json = nlohmann::json;

AgentSessionManager::AgentSessionManager(std::shared_ptr<LibruntimeConfig> config,
                                         std::shared_ptr<RuntimeContext> runtimeContext)
    : librtConfig_(std::move(config)), runtimeContext_(std::move(runtimeContext))
{
}

ErrorInfo AgentSessionManager::AcquireInvokeSession(const std::string &sessionId, const libruntime::MetaData &meta)
{
    if (sessionId.empty()) {
        return ErrorInfo();
    }
    const std::string sessionKey = BuildSessionKey(meta, sessionId);
    auto sessionCtx = GetOrCreateSessionContext(sessionKey);
    sessionCtx->mutex.lock();
    auto err = EnsureLoaded(sessionCtx, sessionId, meta);
    if (!err.OK()) {
        sessionCtx->mutex.unlock();
        return err;
    }
    BindActiveSessionContext(sessionId, sessionCtx);
    return ErrorInfo();
}

ErrorInfo AgentSessionManager::PersistAndReleaseInvokeSession(const std::string &sessionId)
{
    if (sessionId.empty()) {
        return ErrorInfo();
    }
    auto sessionCtx = UnbindActiveSessionContext(sessionId);
    if (sessionCtx == nullptr) {
        return ErrorInfo();
    }
    auto saveErr = Persist(sessionCtx);
    sessionCtx->mutex.unlock();
    return saveErr;
}

std::pair<std::string, ErrorInfo> AgentSessionManager::LoadCurrentSession(const std::string &sessionId)
{
    if (sessionId.empty()) {
        return {"", ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session")};
    }
    auto current = GetActiveSessionContext(sessionId);
    if (current == nullptr) {
        return {"", ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session")};
    }
    return {current->value.sessionData, ErrorInfo()};
}

ErrorInfo AgentSessionManager::UpdateCurrentSession(const std::string &sessionId, const std::string &sessionData)
{
    if (sessionId.empty()) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session");
    }
    auto current = GetActiveSessionContext(sessionId);
    if (current == nullptr) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session");
    }
    current->value.sessionData = sessionData;
    return ErrorInfo();
}

std::shared_ptr<AgentSessionContext> AgentSessionManager::GetOrCreateSessionContext(const std::string &sessionKey)
{
    std::lock_guard<std::mutex> lock(sessionMapMtx_);
    auto iter = sessionMap_.find(sessionKey);
    if (iter != sessionMap_.end()) {
        return iter->second;
    }
    auto sessionCtx = std::make_shared<AgentSessionContext>();
    sessionMap_[sessionKey] = sessionCtx;
    return sessionCtx;
}

std::shared_ptr<AgentSessionContext> AgentSessionManager::GetActiveSessionContext(const std::string &sessionId)
{
    std::lock_guard<std::mutex> lock(activeSessionMapMtx_);
    auto iter = activeSessionMap_.find(sessionId);
    if (iter == activeSessionMap_.end()) {
        return nullptr;
    }
    return iter->second;
}

void AgentSessionManager::BindActiveSessionContext(const std::string &sessionId,
                                                   const std::shared_ptr<AgentSessionContext> &sessionCtx)
{
    std::lock_guard<std::mutex> lock(activeSessionMapMtx_);
    activeSessionMap_[sessionId] = sessionCtx;
}

std::shared_ptr<AgentSessionContext> AgentSessionManager::UnbindActiveSessionContext(const std::string &sessionId)
{
    std::lock_guard<std::mutex> lock(activeSessionMapMtx_);
    auto iter = activeSessionMap_.find(sessionId);
    if (iter == activeSessionMap_.end()) {
        return nullptr;
    }
    auto sessionCtx = iter->second;
    activeSessionMap_.erase(iter);
    return sessionCtx;
}

ErrorInfo AgentSessionManager::EnsureLoaded(const std::shared_ptr<AgentSessionContext> &sessionCtx,
                                            const std::string &sessionId, const libruntime::MetaData &meta)
{
    if (sessionCtx->loaded) {
        return ErrorInfo();
    }
    auto libRuntime = GetLibRuntime();
    if (libRuntime == nullptr) {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "failed to get libruntime for agent session");
    }

    sessionCtx->value.sessionID = sessionId;
    sessionCtx->value.sessionKey = BuildSessionKey(meta, sessionId);

    auto [buffer, readErr] = libRuntime->KVRead(sessionCtx->value.sessionKey, DEFAULT_TIMEOUT_MS);
    if (!readErr.OK()) {
        return readErr;
    }
    if (buffer != nullptr && buffer->GetSize() > 0) {
        sessionCtx->value.sessionData = std::string(static_cast<const char *>(buffer->ImmutableData()), buffer->GetSize());
        sessionCtx->loaded = true;
        return ErrorInfo();
    }

    sessionCtx->value.sessionData = BuildDefaultSession(sessionId);
    auto nativeBuffer = std::make_shared<StringNativeBuffer>(sessionCtx->value.sessionData.size());
    auto err = nativeBuffer->MemoryCopy(sessionCtx->value.sessionData.data(), sessionCtx->value.sessionData.size());
    if (!err.OK()) {
        return err;
    }
    SetParam setParam;
    setParam.ttlSecond = 0;
    err = libRuntime->KVWrite(sessionCtx->value.sessionKey, nativeBuffer, setParam);
    if (!err.OK()) {
        return err;
    }
    sessionCtx->loaded = true;
    return ErrorInfo();
}

ErrorInfo AgentSessionManager::Persist(const std::shared_ptr<AgentSessionContext> &sessionCtx)
{
    if (sessionCtx == nullptr || !sessionCtx->loaded) {
        return ErrorInfo();
    }
    auto libRuntime = GetLibRuntime();
    if (libRuntime == nullptr) {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "failed to get libruntime for agent session");
    }
    auto nativeBuffer = std::make_shared<StringNativeBuffer>(sessionCtx->value.sessionData.size());
    auto err = nativeBuffer->MemoryCopy(sessionCtx->value.sessionData.data(), sessionCtx->value.sessionData.size());
    if (!err.OK()) {
        return err;
    }
    SetParam setParam;
    setParam.ttlSecond = 0;
    return libRuntime->KVWrite(sessionCtx->value.sessionKey, nativeBuffer, setParam);
}

std::string AgentSessionManager::BuildSessionKey(const libruntime::MetaData &meta, const std::string &sessionId) const
{
    const auto &funcMeta = meta.functionmeta();
    return std::string(AGENT_SESSION_KEY_PREFIX) + ":" + librtConfig_->tenantId + ":" + funcMeta.modulename() + ":" +
           funcMeta.functionname() + ":" + sessionId;
}

std::string AgentSessionManager::BuildDefaultSession(const std::string &sessionId) const
{
    json sessionJson;
    sessionJson["sessionID"] = sessionId;
    sessionJson["histories"] = json::array();
    return sessionJson.dump();
}

std::shared_ptr<Libruntime> AgentSessionManager::GetLibRuntime() const
{
    if (runtimeContext_ != nullptr) {
        const std::string rtCtx = runtimeContext_->GetJobIdThreadlocal();
        if (!rtCtx.empty()) {
            return LibruntimeManager::Instance().GetLibRuntime(rtCtx);
        }
    }
    return LibruntimeManager::Instance().GetLibRuntime();
}

}  // namespace Libruntime
}  // namespace YR
