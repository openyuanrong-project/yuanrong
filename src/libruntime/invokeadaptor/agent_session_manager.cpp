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

#include <array>
#include <chrono>
#include <json.hpp>
#include <openssl/evp.h>
#include <openssl/sha.h>
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
    auto sessionCtx = GetOrCreateSessionContext(sessionId, sessionKey);
    sessionCtx->mutex.lock();
    auto err = EnsureLoaded(sessionCtx, sessionId);
    if (!err.OK()) {
        ReleaseSessionContextReference(sessionCtx);
        sessionCtx->mutex.unlock();
        return err;
    }
    ResetSessionInterrupted(sessionCtx);
    return ErrorInfo();
}

ErrorInfo AgentSessionManager::PersistAndReleaseInvokeSession(const std::string &sessionId)
{
    if (sessionId.empty()) {
        return ErrorInfo();
    }
    auto sessionCtx = GetSessionContext(sessionId);
    if (sessionCtx == nullptr) {
        return ErrorInfo();
    }
    auto saveErr = Persist(sessionCtx);
    sessionCtx->mutex.unlock();
    ReleaseSessionContextReference(sessionCtx);
    return saveErr;
}

std::pair<std::string, ErrorInfo> AgentSessionManager::LoadCurrentSession(const std::string &sessionId)
{
    if (sessionId.empty()) {
        return {"", ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session")};
    }
    auto current = GetSessionContext(sessionId);
    if (current == nullptr) {
        return {"", ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session")};
    }
    std::lock_guard<std::mutex> lock(current->dataMutex);
    return {current->value.sessionData, ErrorInfo()};
}

ErrorInfo AgentSessionManager::UpdateCurrentSession(const std::string &sessionId, const std::string &sessionData)
{
    if (sessionId.empty()) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session");
    }
    json incomingJson;
    try {
        incomingJson = json::parse(sessionData);
    } catch (const json::parse_error &) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "failed to parse sessionJson");
    }
    if (!incomingJson.contains("histories") || !incomingJson["histories"].is_array()) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "session histories is invalid");
    }

    auto current = GetSessionContext(sessionId);
    if (current == nullptr) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session");
    }

    std::lock_guard<std::mutex> lock(current->dataMutex);
    try {
        json currentJson = json::parse(current->value.sessionData);
        currentJson["histories"] = incomingJson["histories"];
        current->value.sessionData = currentJson.dump();
    } catch (const json::parse_error &) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "failed to parse current sessionJson");
    }
    return ErrorInfo();
}

ErrorInfo AgentSessionManager::SetSessionInterrupted(const std::string &sessionId)
{
    if (sessionId.empty()) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "session id is empty");
    }

    auto sessionCtx = GetSessionContext(sessionId);
    if (sessionCtx == nullptr) {
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "current invoke has no active agent session");
    }

    sessionCtx->interrupted.store(true);

    std::shared_ptr<WaitNotifyContext> waitNotifyCtx;
    {
        std::lock_guard<std::mutex> mapLock(waitNotifyMtx_);
        auto iter = waitNotifyMap_.find(sessionId);
        if (iter != waitNotifyMap_.end()) {
            waitNotifyCtx = iter->second;
        }
    }
    bool shouldNotify = false;
    if (waitNotifyCtx != nullptr) {
        std::lock_guard<std::mutex> lock(waitNotifyCtx->mutex);
        if (waitNotifyCtx->state == WaitState::WAITING) {
            waitNotifyCtx->state = WaitState::INTERRUPTED;
            waitNotifyCtx->notifyData = nullptr;
            shouldNotify = true;
        }
    }
    if (shouldNotify) {
        waitNotifyCtx->cv.notify_one();
    }

    return ErrorInfo();
}

bool AgentSessionManager::IsSessionInterrupted(const std::string &sessionId)
{
    if (sessionId.empty()) {
        return false;
    }
    auto sessionCtx = GetSessionContext(sessionId);
    if (sessionCtx == nullptr) {
        return false;
    }

    return sessionCtx->interrupted.load();
}

std::shared_ptr<AgentSessionContext> AgentSessionManager::GetOrCreateSessionContext(const std::string &sessionId,
                                                                                    const std::string &sessionKey)
{
    std::lock_guard<std::mutex> lock(sessionMapMtx_);
    auto iter = sessionMap_.find(sessionId);
    if (iter != sessionMap_.end()) {
        iter->second->refCount++;
        return iter->second;
    }
    auto sessionCtx = std::make_shared<AgentSessionContext>();
    sessionCtx->value.sessionID = sessionId;
    sessionCtx->value.sessionKey = sessionKey;
    sessionCtx->refCount = 1;
    sessionMap_[sessionId] = sessionCtx;
    return sessionCtx;
}

void AgentSessionManager::ReleaseSessionContextReference(const std::shared_ptr<AgentSessionContext> &sessionCtx)
{
    if (sessionCtx == nullptr) {
        return;
    }

    bool shouldCleanupWaitNotify = false;
    std::string sessionId;
    {
        std::lock_guard<std::mutex> lock(sessionMapMtx_);
        if (sessionCtx->refCount == 0) {
            return;
        }

        sessionCtx->refCount--;
        if (sessionCtx->refCount != 0) {
            return;
        }

        auto iter = sessionMap_.find(sessionCtx->value.sessionID);
        if (iter != sessionMap_.end() && iter->second == sessionCtx) {
            sessionId = iter->first;
            sessionMap_.erase(iter);
            shouldCleanupWaitNotify = true;
        }
    }
    if (shouldCleanupWaitNotify) {
        RemoveWaitNotifyContext(sessionId);
    }
}

std::shared_ptr<AgentSessionContext> AgentSessionManager::GetSessionContext(const std::string &sessionId)
{
    std::lock_guard<std::mutex> lock(sessionMapMtx_);
    auto iter = sessionMap_.find(sessionId);
    if (iter == sessionMap_.end()) {
        return nullptr;
    }
    return iter->second;
}

void AgentSessionManager::RemoveWaitNotifyContext(const std::string &sessionId)
{
    std::lock_guard<std::mutex> lock(waitNotifyMtx_);
    waitNotifyMap_.erase(sessionId);
}

ErrorInfo AgentSessionManager::EnsureLoaded(const std::shared_ptr<AgentSessionContext> &sessionCtx,
                                            const std::string &sessionId)
{
    {
        std::lock_guard<std::mutex> lock(sessionCtx->dataMutex);
        if (sessionCtx->loaded) {
            return ErrorInfo();
        }
    }
    auto libRuntime = GetLibRuntime();
    if (libRuntime == nullptr) {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "failed to get libruntime for agent session");
    }

    auto [buffer, readErr] = libRuntime->KVRead(sessionCtx->value.sessionKey, ZERO_TIMEOUT);
    if (!readErr.OK() && readErr.Code() != ErrorCode::ERR_GET_OPERATION_FAILED) {
        return readErr;
    }
    if (buffer != nullptr && buffer->GetSize() > 0) {
        std::lock_guard<std::mutex> lock(sessionCtx->dataMutex);
        sessionCtx->value.sessionData =
            std::string(static_cast<const char *>(buffer->ImmutableData()), buffer->GetSize());
        sessionCtx->loaded = true;
        return ErrorInfo();
    }

    std::string defaultSession = BuildDefaultSession(sessionId);
    {
        std::lock_guard<std::mutex> lock(sessionCtx->dataMutex);
        sessionCtx->value.sessionData = defaultSession;
        sessionCtx->loaded = true;
    }
    auto nativeBuffer = std::make_shared<StringNativeBuffer>(defaultSession.size());
    auto err = nativeBuffer->MemoryCopy(defaultSession.data(), defaultSession.size());
    if (!err.OK()) {
        std::lock_guard<std::mutex> lock(sessionCtx->dataMutex);
        sessionCtx->loaded = false;
        return err;
    }
    SetParam setParam;
    setParam.ttlSecond = 0;
    err = libRuntime->KVWrite(sessionCtx->value.sessionKey, nativeBuffer, setParam);
    if (!err.OK()) {
        std::lock_guard<std::mutex> lock(sessionCtx->dataMutex);
        sessionCtx->loaded = false;
        return err;
    }
    return ErrorInfo();
}

void AgentSessionManager::ResetSessionInterrupted(const std::shared_ptr<AgentSessionContext> &sessionCtx)
{
    if (sessionCtx == nullptr) {
        return;
    }

    bool shouldResetWaitState = sessionCtx->interrupted.exchange(false);
    if (!shouldResetWaitState) {
        return;
    }

    std::shared_ptr<WaitNotifyContext> waitNotifyCtx;
    {
        std::lock_guard<std::mutex> mapLock(waitNotifyMtx_);
        auto iter = waitNotifyMap_.find(sessionCtx->value.sessionID);
        if (iter != waitNotifyMap_.end()) {
            waitNotifyCtx = iter->second;
        }
    }
    if (waitNotifyCtx == nullptr) {
        return;
    }

    std::lock_guard<std::mutex> waitLock(waitNotifyCtx->mutex);
    if (waitNotifyCtx->state == WaitState::INTERRUPTED) {
        waitNotifyCtx->state = WaitState::IDLE;
        waitNotifyCtx->notifyData = nullptr;
    }
}

ErrorInfo AgentSessionManager::Persist(const std::shared_ptr<AgentSessionContext> &sessionCtx)
{
    if (sessionCtx == nullptr) {
        return ErrorInfo();
    }
    auto libRuntime = GetLibRuntime();
    if (libRuntime == nullptr) {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "failed to get libruntime for agent session");
    }
    std::string sessionKey;
    std::string sessionData;
    {
        std::lock_guard<std::mutex> lock(sessionCtx->dataMutex);
        if (!sessionCtx->loaded) {
            return ErrorInfo();
        }
        sessionKey = sessionCtx->value.sessionKey;
        sessionData = sessionCtx->value.sessionData;
    }
    auto nativeBuffer = std::make_shared<StringNativeBuffer>(sessionData.size());
    auto err = nativeBuffer->MemoryCopy(sessionData.data(), sessionData.size());
    if (!err.OK()) {
        return err;
    }
    SetParam setParam;
    setParam.ttlSecond = 0;
    return libRuntime->KVWrite(sessionKey, nativeBuffer, setParam);
}

std::string AgentSessionManager::BuildSessionKey(const libruntime::MetaData &meta, const std::string &sessionId) const
{
    // Hash "functionID:sessionId" so invoke and delete paths share the same stable key source.
    std::string sessionKey = std::string(AGENT_SESSION_KEY_PREFIX) + ":" +
                             EncodeSha256Base64Url(meta.functionmeta().functionid() + ":" + sessionId);
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

std::shared_ptr<WaitNotifyContext> AgentSessionManager::GetOrCreateWaitNotifyContext(const std::string &sessionId)
{
    std::lock_guard<std::mutex> lock(waitNotifyMtx_);
    auto iter = waitNotifyMap_.find(sessionId);
    if (iter != waitNotifyMap_.end()) {
        return iter->second;
    }
    auto ctx = std::make_shared<WaitNotifyContext>();
    waitNotifyMap_[sessionId] = ctx;
    return ctx;
}

std::pair<ErrorInfo, std::shared_ptr<Buffer>> AgentSessionManager::Wait(const std::string &sessionId, int64_t timeoutMs)
{
    if (sessionId.empty()) {
        YRLOG_WARN("current session id is empty, return directly");
        return {ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "session id is empty"), nullptr};
    }
    if (timeoutMs < 0) {
        YRLOG_WARN("timeoutMs is negative, return directly");
        return {ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "timeoutMs must not be negative"), nullptr};
    }
    auto sessionCtx = GetSessionContext(sessionId);
    if (sessionCtx == nullptr) {
        YRLOG_WARN("session ctx of session id: {} is empty, return derectly", sessionId);
        return {ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "no active session for wait"), nullptr};
    }
    sessionCtx->mutex.unlock();
    auto waitNotifyCtx = GetOrCreateWaitNotifyContext(sessionId);
    std::unique_lock<std::mutex> lock(waitNotifyCtx->mutex);
    auto relockSession = [&]() { sessionCtx->mutex.lock(); };
    if (sessionCtx->interrupted.load()) {
        lock.unlock();
        relockSession();
        return {ErrorInfo(ERR_SESSION_INTERRUPTED, ModuleCode::RUNTIME, "session has been interrupted"), nullptr};
    }
    if (waitNotifyCtx->state == WaitState::INTERRUPTED) {
        lock.unlock();
        relockSession();
        return {ErrorInfo(ERR_SESSION_INTERRUPTED, ModuleCode::RUNTIME, "session has been interrupted"), nullptr};
    }
    if (waitNotifyCtx->state == WaitState::WAITING) {
        lock.unlock();
        relockSession();
        return {ErrorInfo(ERR_SESSION_NOT_WAITING, ModuleCode::RUNTIME, "session is already waiting"), nullptr};
    }
    waitNotifyCtx->state = WaitState::WAITING;
    waitNotifyCtx->notifyData = nullptr;
    YRLOG_DEBUG("current notify ctx state of session id: {} is WAITING, timeoutMs is {}", sessionId, timeoutMs);
    auto deadline = std::chrono::steady_clock::now() + std::chrono::milliseconds(timeoutMs);
    bool notified = waitNotifyCtx->cv.wait_until(lock, deadline, [waitNotifyCtx]() {
        return waitNotifyCtx->state == WaitState::NOTIFIED || waitNotifyCtx->state == WaitState::INTERRUPTED;
    });
    if (!notified) {
        waitNotifyCtx->state = WaitState::TIMEOUT;
    }
    auto result = waitNotifyCtx->notifyData;
    auto finalState = waitNotifyCtx->state;
    waitNotifyCtx->state = WaitState::IDLE;
    waitNotifyCtx->notifyData = nullptr;
    lock.unlock();
    relockSession();
    if (finalState == WaitState::INTERRUPTED) {
        return {ErrorInfo(ERR_SESSION_INTERRUPTED, ModuleCode::RUNTIME, "session wait was interrupted"), nullptr};
    }
    if (finalState == WaitState::TIMEOUT) {
        return {ErrorInfo(ERR_SESSION_TIMEOUT, ModuleCode::RUNTIME, "session wait timeout"), nullptr};
    }
    return {ErrorInfo(), result};
}

ErrorInfo AgentSessionManager::Notify(const std::string &sessionId, std::shared_ptr<Buffer> data)
{
    if (sessionId.empty()) {
        YRLOG_WARN("session id is empty, no need notify, return directly");
        return ErrorInfo(ERR_PARAM_INVALID, ModuleCode::RUNTIME, "session id is empty");
    }

    std::shared_ptr<WaitNotifyContext> waitNotifyCtx;
    {
        std::lock_guard<std::mutex> mapLock(waitNotifyMtx_);
        auto iter = waitNotifyMap_.find(sessionId);
        if (iter == waitNotifyMap_.end()) {
            YRLOG_DEBUG("No waiting thread for session {}, notify request discarded", sessionId);
            return ErrorInfo();
        }
        waitNotifyCtx = iter->second;
    }

    bool shouldNotify = false;
    {
        std::lock_guard<std::mutex> ctxLock(waitNotifyCtx->mutex);
        if (waitNotifyCtx->state != WaitState::WAITING) {
            YRLOG_DEBUG("Session {} is not in waiting state, current state: {}", sessionId,
                        static_cast<int>(waitNotifyCtx->state));
            return ErrorInfo();
        }
        waitNotifyCtx->notifyData = std::move(data);
        waitNotifyCtx->state = WaitState::NOTIFIED;
        shouldNotify = true;
    }
    if (shouldNotify) {
        waitNotifyCtx->cv.notify_one();
    }
    return ErrorInfo();
}

}  // namespace Libruntime
}  // namespace YR
