/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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
#include "src/libruntime/invokeadaptor/instance_manager.h"

namespace YR {
namespace Libruntime {

class NormalInsManager : public std::enable_shared_from_this<NormalInsManager>, public InsManager {
public:
    NormalInsManager() = default;
    NormalInsManager(ScheduleInsCallback cb, std::shared_ptr<FSClient> client, std::shared_ptr<MemoryStore> store,
                     std::shared_ptr<RequestManager> reqMgr, std::shared_ptr<LibruntimeConfig> config)
        : InsManager(cb, client, store, reqMgr, config)
    {
    }
    void UpdateConfig(int recycleTimeMs) override;
    bool ScaleUp(std::shared_ptr<InvokeSpec> spec, size_t reqNum) override;
    void ScaleDown(const std::shared_ptr<InvokeSpec> spec, bool isInstanceNormal = false) override;
    void ScaleCancel(const RequestResource &resource, size_t reqNum, bool cleanAll = false) override;
    void StartBatchRenewTimer() override;
    void StartNormalInsScaleDownTimer(const RequestResource &resource, const std::string &id);
    void HandleCreateResponse(std::shared_ptr<InvokeSpec> spec, const CreateResponse &resp,
                              std::shared_ptr<CreatingInsInfo> insInfo);
    void HandleCreateNotify(const NotifyRequest &req);

private:
    std::shared_ptr<InvokeSpec> BuildCreateSpec(std::shared_ptr<InvokeSpec> spec);
    void SendKillReq(const std::string &insId);
    void AddInsInfo(const std::shared_ptr<InvokeSpec> spec, const RequestResource &resource);
    // add instance info without locking; the caller must ensure locking.
    void AddInsInfoBare(const std::shared_ptr<InvokeSpec> spec, std::shared_ptr<RequestResourceInfo> info);
    bool CreateInstance(std::shared_ptr<InvokeSpec> spec, size_t reqNum);
    void SendCreateReq(std::shared_ptr<InvokeSpec> spec, size_t delayTime);
    void HandleFailCreateNotify(const std::shared_ptr<InvokeSpec> createSpec, const RequestResource &resource);
    void HandleSuccessCreateNotify(const std::shared_ptr<InvokeSpec> createSpec, const RequestResource &resource,
                                   const NotifyRequest &req);
    void ScaleDownHandler(const RequestResource &resource, const std::string &id);
};
}  // namespace Libruntime
}  // namespace YR