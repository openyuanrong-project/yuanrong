/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <future>
#include <thread>

#define private public
#include "src/libruntime/invokeadaptor/agent_session_manager.h"

using namespace YR::Libruntime;

namespace YR {
namespace test {

class AgentSessionManagerTest : public testing::Test {
public:
    void SetUp() override
    {
        cfg_ = std::make_shared<LibruntimeConfig>();
        cfg_->tenantId = "tenant-id";
        rtCtx_ = std::make_shared<RuntimeContext>();
        manager_ = std::make_shared<AgentSessionManager>(cfg_, rtCtx_);
    }

    std::shared_ptr<AgentSessionContext> CreateSessionContext(const std::string &sessionId)
    {
        auto sessionCtx = manager_->GetOrCreateSessionContext(sessionId, "mock-key-" + sessionId);
        sessionCtx->value.sessionData = manager_->BuildDefaultSession(sessionId);
        sessionCtx->loaded = true;
        return sessionCtx;
    }

    std::shared_ptr<LibruntimeConfig> cfg_;
    std::shared_ptr<RuntimeContext> rtCtx_;
    std::shared_ptr<AgentSessionManager> manager_;
};

TEST_F(AgentSessionManagerTest, WaitWithEmptySessionIdTest)
{
    auto [err, buf] = manager_->Wait("", 100);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID);
    ASSERT_EQ(buf, nullptr);
}

TEST_F(AgentSessionManagerTest, NotifyWithoutWaitingContextTest)
{
    auto err = manager_->Notify("not-exist-session", nullptr);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_OK);
}

TEST_F(AgentSessionManagerTest, WaitNotifySuccessTest)
{
    const std::string sessionId = "session-success";
    auto sessionCtx = CreateSessionContext(sessionId);

    std::promise<std::pair<ErrorCode, std::string>> p;
    auto f = p.get_future();
    std::thread waiter([&]() {
        sessionCtx->mutex.lock();
        auto [err, buf] = manager_->Wait(sessionId, 3000);
        std::string got;
        if (buf != nullptr && buf->GetSize() > 0) {
            got.assign(static_cast<const char *>(buf->ImmutableData()), buf->GetSize());
        }
        p.set_value({err.Code(), got});
        sessionCtx->mutex.unlock();
        manager_->ReleaseSessionContextReference(sessionCtx);
    });

    std::this_thread::sleep_for(std::chrono::milliseconds(50));
    const std::string payload = "payload";
    auto data = std::make_shared<StringNativeBuffer>(payload.size());
    ASSERT_EQ(data->MemoryCopy(payload.data(), payload.size()).Code(), ErrorCode::ERR_OK);
    ASSERT_EQ(manager_->Notify(sessionId, data).Code(), ErrorCode::ERR_OK);

    auto [code, got] = f.get();
    ASSERT_EQ(code, ErrorCode::ERR_OK);
    ASSERT_EQ(got, payload);
    waiter.join();
}

TEST_F(AgentSessionManagerTest, WaitWithNegativeTimeoutTest)
{
    const std::string sessionId = "session-negative-timeout";
    auto sessionCtx = CreateSessionContext(sessionId);

    sessionCtx->mutex.lock();
    auto [err, buf] = manager_->Wait(sessionId, -1);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_PARAM_INVALID);
    ASSERT_EQ(buf, nullptr);
    sessionCtx->mutex.unlock();
    manager_->ReleaseSessionContextReference(sessionCtx);
}

TEST_F(AgentSessionManagerTest, WaitTimeoutTest)
{
    const std::string sessionId = "session-timeout";
    auto sessionCtx = CreateSessionContext(sessionId);

    sessionCtx->mutex.lock();
    auto [err, buf] = manager_->Wait(sessionId, 20);
    ASSERT_EQ(err.Code(), ErrorCode::ERR_SESSION_TIMEOUT);
    ASSERT_EQ(buf, nullptr);
    sessionCtx->mutex.unlock();
    manager_->ReleaseSessionContextReference(sessionCtx);
}

TEST_F(AgentSessionManagerTest, ResetSessionInterruptedClearsInMemoryInterruptedState)
{
    const std::string sessionId = "session-reset-memory";
    auto sessionCtx = CreateSessionContext(sessionId);
    sessionCtx->interrupted.store(true);

    auto waitNotifyCtx = manager_->GetOrCreateWaitNotifyContext(sessionId);
    waitNotifyCtx->state = WaitState::INTERRUPTED;
    waitNotifyCtx->notifyData = std::make_shared<StringNativeBuffer>(0);

    manager_->ResetSessionInterrupted(sessionCtx);

    ASSERT_FALSE(manager_->IsSessionInterrupted(sessionId));
    ASSERT_EQ(waitNotifyCtx->state, WaitState::IDLE);
    ASSERT_EQ(waitNotifyCtx->notifyData, nullptr);

    manager_->ReleaseSessionContextReference(sessionCtx);
}

TEST_F(AgentSessionManagerTest, SetSessionInterruptedDoesNotLeaveInterruptedWaitStateWhenNotWaiting)
{
    const std::string sessionId = "session-interrupt-no-wait";
    auto sessionCtx = CreateSessionContext(sessionId);

    auto waitNotifyCtx = manager_->GetOrCreateWaitNotifyContext(sessionId);
    waitNotifyCtx->state = WaitState::IDLE;

    manager_->SetSessionInterrupted(sessionId);

    ASSERT_TRUE(manager_->IsSessionInterrupted(sessionId));
    ASSERT_EQ(waitNotifyCtx->state, WaitState::IDLE);

    manager_->ReleaseSessionContextReference(sessionCtx);
}

}  // namespace test
}  // namespace YR
