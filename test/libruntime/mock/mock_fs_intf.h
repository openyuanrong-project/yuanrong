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

#include "src/libruntime/fsclient/fs_intf.h"

namespace YR {
namespace test {
class MockGwClient : public YR::Libruntime::FSIntf {
public:
    MOCK_METHOD4(CreateAsync, void(const YR::Libruntime::CreateRequest &, YR::Libruntime::CreateRespCallback,
                                   YR::Libruntime::CreateCallBack, int));
    MOCK_METHOD4(GroupCreateAsync, void(const YR::Libruntime::CreateRequests &, YR::Libruntime::CreateRespsCallback,
                                        YR::Libruntime::CreateCallBack, int));
    MOCK_METHOD5(Start, YR::Libruntime::ErrorInfo(const std::string &, const std::string &, const std::string &,
                                                  const std::string &, const YR::Libruntime::SubscribeFunc &));
    MOCK_METHOD0(Stop, void(void));
    MOCK_METHOD0(IsHealth, bool(void));
    MOCK_METHOD3(InvokeAsync,
                 void(const std::shared_ptr<YR::Libruntime::InvokeMessageSpec> &, YR::Libruntime::InvokeCallBack, int));
    MOCK_METHOD2(CallResultAsync, void(const std::shared_ptr<YR::Libruntime::CallResultMessageSpec>,
                                       YR::Libruntime::CallResultCallBack));
    MOCK_METHOD3(KillAsync, void(const YR::Libruntime::KillRequest &, YR::Libruntime::KillCallBack, int));
    MOCK_METHOD2(ExitAsync, void(const YR::Libruntime::ExitRequest &, YR::Libruntime::ExitCallBack));
    MOCK_METHOD2(StateSaveAsync, void(const YR::Libruntime::StateSaveRequest &, YR::Libruntime::StateSaveCallBack));
    MOCK_METHOD2(StateLoadAsync, void(const YR::Libruntime::StateLoadRequest &, YR::Libruntime::StateLoadCallBack));
    MOCK_METHOD3(CreateRGroupAsync, void(const YR::Libruntime::CreateResourceGroupRequest &,
                                           YR::Libruntime::CreateResourceGroupCallBack, int));
};
}  // namespace test
}  // namespace YR