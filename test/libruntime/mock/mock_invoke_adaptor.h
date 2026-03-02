/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/libruntime/invokeadaptor/invoke_adaptor.h"

using namespace testing;
namespace YR {
namespace Libruntime {
class MockInvokeAdaptor : public InvokeAdaptor {
public:
    MockInvokeAdaptor() = default;
    MOCK_METHOD0(Exit, void(void));

    MOCK_METHOD0(ReceiveRequestLoop, void(void));

    MOCK_METHOD1(Finalize, void(bool isDriver));

    MOCK_METHOD3(Kill, ErrorInfo(const std::string &instanceId, const std::string &payload, int sigNo));

    MOCK_METHOD3(Cancel, ErrorInfo(const std::vector<std::string> &objids, bool isForce, bool isRecursive));

    MOCK_METHOD3(KillAsync, void(const std::string &instanceId, const std::string &payload, int sigNo));

    MOCK_METHOD(void, KillAsyncCB, ((const std::string &instanceId), (const std::string &payload), (int signal),
                (std::function<void(const ErrorInfo &err)> cb), (int timeoutSec)), (override));

    MOCK_METHOD2(GroupCreate, ErrorInfo(const std::string &groupName, GroupOpts &opts));

    MOCK_METHOD2(RangeCreate, ErrorInfo(const std::string &groupName, InstanceRange &range));

    MOCK_METHOD1(GroupWait, ErrorInfo(const std::string &groupName));

    MOCK_METHOD1(GroupTerminate, void(const std::string &groupName));

    MOCK_METHOD2(GetInstanceIds, std::pair<std::vector<std::string>, ErrorInfo>(const std::string &objId,
                                                                                const std::string &groupName));

    MOCK_METHOD2(SaveState, ErrorInfo(const std::shared_ptr<Buffer> data, const int &timeout));

    MOCK_METHOD2(LoadState, ErrorInfo(std::shared_ptr<Buffer> &data, const int &timeout));

    MOCK_METHOD1(ExecShutdownCallback, ErrorInfo(uint64_t gracePeriodSec));

    MOCK_METHOD3(AcquireInstance,
                 std::pair<InstanceAllocation, ErrorInfo>(const std::string &stateId, const FunctionMeta &functionMeta,
                                                          InvokeOptions &opts));

    MOCK_METHOD4(ReleaseInstance,
                 ErrorInfo(const std::string &leaseId, const std::string &stateId, bool abnormal, InvokeOptions &opts));

    MOCK_METHOD3(GetInstance,
                 std::pair<YR::Libruntime::FunctionMeta, ErrorInfo>(const std::string &name,
                                                                    const std::string &nameSpace, int timeoutSec));
    MOCK_METHOD3(UpdateSchdulerInfo,
                 void(const std::string &scheduleName, const std::string &schedulerId, const std::string &option));

    MOCK_METHOD1(EraseFsIntf, void(const std::string &id));
    MOCK_METHOD0(IsHealth, bool());
    MOCK_METHOD3(StreamWriteEvent,
                 ErrorInfo(const std::string &streamMessage, const std::string &requestId,
                           const std::string &instanceId));
};
}  // namespace Libruntime
}  // namespace YR
