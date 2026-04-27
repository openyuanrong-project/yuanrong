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

#pragma once

#include <memory>

#include "datasystem/object_client.h"

#include "src/libruntime/objectstore/datasystem_client_wrapper.h"

#include "src/libruntime/gwclient/gw_client.h"

namespace YR {
namespace Libruntime {
datasystem::StatusCode oneOfNoRetryCode = datasystem::StatusCode::K_RUNTIME_ERROR;
class GwDatasystemClientWrapper : public DatasystemClientWrapper {
public:
    GwDatasystemClientWrapper(std::shared_ptr<GwClient> client)
    {
        gwClient = client;
    }

    datasystem::Status GDecreaseRef(const std::vector<std::string> &objectIds,
                                    std::vector<std::string> &failedObjectIds)
    {
        if (auto locked = gwClient.lock()) {
            auto failedIdsPtr = std::make_shared<std::vector<std::string>>();
            auto err = locked->PosixGDecreaseRef(objectIds, failedIdsPtr);
            if (err.OK()) {
                failedObjectIds.assign(failedIdsPtr->begin(), failedIdsPtr->end());
                return datasystem::Status(datasystem::StatusCode::K_OK, err.Msg());
            } else {
                return datasystem::Status(oneOfNoRetryCode, err.Msg());
            }
        } else {
            YRLOG_DEBUG("gw client pointer is expired.");
        }
        return {};
    }

    void SetTenantId(const std::string &tenantId)
    {
        if (auto locked = gwClient.lock()) {
            locked->SetTenantId(tenantId);
        }
    }

private:
    std::weak_ptr<GwClient> gwClient;
};

}  // namespace Libruntime
}  // namespace YR