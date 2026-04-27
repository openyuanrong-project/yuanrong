/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

// Wapper of datasystem::ObjectClient / datasystem::AgentClient.

#pragma once

#include <string>
#include <vector>

#include "datasystem/utils/status.h"

namespace YR {
namespace Libruntime {
class DatasystemClientWrapper {
public:
    virtual datasystem::Status GDecreaseRef(const std::vector<std::string> &objectIds,
                                            std::vector<std::string> &failedObjectIds) = 0;

    virtual void SetTenantId(const std::string &tenantId) = 0;
};

}  // namespace Libruntime
}  // namespace YR