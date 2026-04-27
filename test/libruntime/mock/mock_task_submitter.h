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
#include "src/libruntime/invokeadaptor/task_submitter.h"
namespace YR {
namespace test {
class MockTaskSubmitter : public YR::Libruntime::TaskSubmitter {
public:
    MOCK_METHOD0(Init, void(void));
    MOCK_METHOD2(AcquireInstance, std::pair<InstanceAllocation, ErrorInfo>(const std::string &stateId,
                                                                           std::shared_ptr<InvokeSpec> spec));
    MOCK_METHOD4(ReleaseInstance, ErrorInfo(const std::string &leaseId, const std::string &stateId, bool abnormal,
                                            std::shared_ptr<InvokeSpec> spec));
    MOCK_METHOD3(UpdateSchdulerInfo,
                 void(const std::string &scheduleName, const std::string &schedulerId, const std::string &option));
    MOCK_METHOD2(UpdateFaaSSchedulerInfo, void(std::string schedulerFuncKey,
                                 const std::vector<SchedulerInstance> &schedulerInstanceList));
    MOCK_METHOD5(CancelStatelessRequest, ErrorInfo(std::shared_ptr<InvokeSpec> spec, const KillFunc &killCallBack,
                                                bool isForce, bool isRecursive, const std::string &objId));
};

}  // namespace test
}  // namespace YR