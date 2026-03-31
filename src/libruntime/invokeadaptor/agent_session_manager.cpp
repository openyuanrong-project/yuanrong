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

#include <openssl/evp.h>
#include <openssl/sha.h>
#include <array>
#include <json.hpp>

#include "src/dto/buffer.h"
#include "src/libruntime/libruntime.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/libruntime/utils/constants.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
using json = nlohmann::json;

namespace {
constexpr size_t MAX_SESSION_KEY_LENGTH = 64;
constexpr size_t SHA256_BASE64_LENGTH = 44;
constexpr char BASE64_PADDING = '=';

std::shared_ptr<Libruntime> GetLibRuntime()
{
    return LibruntimeManager::Instance().GetLibRuntime();
}

std::string EncodeSha256Base64Url(const std::string &input)
{
    std::array<unsigned char, SHA256_DIGEST_LENGTH> digest{};
    SHA256(reinterpret_cast<const unsigned char *>(input.data()), input.size(), digest.data());

    std::array<unsigned char, SHA256_BASE64_LENGTH + 1> encodedBytes{};
    const int encodedLen = EVP_EncodeBlock(encodedBytes.data(), digest.data(), static_cast<int>(digest.size()));
    std::string encoded(reinterpret_cast<const char *>(encodedBytes.data()), encodedLen);
    for (auto &ch : encoded) {
        if (ch == '+') {
            ch = '-';
        } else if (ch == '/') {
            ch = '_';
        }
    }
    while (!encoded.empty() && encoded.back() == BASE64_PADDING) {
        encoded.pop_back();
    }
    return encoded;
}
}  // namespace

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

ErrorInfo AgentSessionManager::SetSessionInterrupted(const std::string &sessionId)
{
    if (sessionId.empty()) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "session id is empty");
    }

    auto sessionCtx = GetActiveSessionContext(sessionId);
    if (sessionCtx == nullptr) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session");
    }

    try {
        json sessionJson = json::parse(sessionCtx->value.sessionData);
        sessionJson["isInterrupted"] = true;
        sessionCtx->value.sessionData = sessionJson.dump();
    } catch (const json::parse_error &e) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "failed to parse sessionJson");
    }

    auto saveErr = Persist(sessionCtx);
    return saveErr;
}

bool AgentSessionManager::IsSessionInterrupted(const std::string &sessionId)
{
    if (sessionId.empty()) {
        return false;
    }
    auto sessionCtx = GetActiveSessionContext(sessionId);
    if (sessionCtx == nullptr) {
        return false;
    }

    std::string sessionData;
    {
        std::lock_guard<std::mutex> lock(activeSessionMapMtx_);
        auto iter = activeSessionMap_.find(sessionId);
        if (iter == activeSessionMap_.end()) {
            YRLOG_INFO("sessionId:{} is not found in activeSessionMap_", sessionId);
            return false;
        }
        sessionData = iter->second->value.sessionData;
    }

    try {
        json sessionJson = json::parse(sessionData);
        return sessionJson.value("isInterrupted", false);
    } catch (const json::parse_error &e) {
        return false;
    }
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

    auto [buffer, readErr] = libRuntime->KVRead(sessionCtx->value.sessionKey, ZERO_TIMEOUT);
    if (!readErr.OK() && readErr.Code() != ErrorCode::ERR_GET_OPERATION_FAILED) {
        return readErr;
    }
    if (buffer != nullptr && buffer->GetSize() > 0) {
        sessionCtx->value.sessionData =
            std::string(static_cast<const char *>(buffer->ImmutableData()), buffer->GetSize());
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
    // Hash "name:sessionId" so the session key has an unambiguous boundary and a fixed length.
    std::string sessionKey = std::string(AGENT_SESSION_KEY_PREFIX) + ":" +
                             EncodeSha256Base64Url(meta.functionmeta().name() + ":" + sessionId);
    if (sessionKey.size() > MAX_SESSION_KEY_LENGTH) {
        sessionKey.resize(MAX_SESSION_KEY_LENGTH);
    }
    return sessionKey;
}

std::string AgentSessionManager::BuildDefaultSession(const std::string &sessionId) const
{
    json sessionJson;
    sessionJson["sessionID"] = sessionId;
    sessionJson["histories"] = json::array();
    return sessionJson.dump();
}

}  // namespace Libruntime
}  // namespace YR
