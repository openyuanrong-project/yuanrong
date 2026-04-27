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

#include "resource_group_create_spec.h"
#include "src/libruntime/utils/utils.h"

namespace YR {
namespace Libruntime {

void ResourceGroupCreateSpec::BuildCreateResourceGroupRequest()
{
    requestCreateRGroup.set_requestid(this->requestId);
    requestCreateRGroup.set_traceid(this->traceId);
    CommonResourceGroupSpec *spec = requestCreateRGroup.mutable_rgroupspec();
    spec->set_name(this->rGroupSpec.name);
    spec->set_appid(this->jobId);
    spec->set_tenantid(this->tenantId);
    for (auto &bundle : this->rGroupSpec.bundles) {
        Bundle specBundle;
        auto *resources = specBundle.mutable_resources();
        if (auto it = bundle.find(std::string(CPU_RESOURCE_NAME)); it == bundle.end() || it->second == 0) {
            resources->insert({CPU_RESOURCE_NAME, 0});
        }
        if (auto it = bundle.find(std::string(MEMORY_RESOURCE_NAME)); it == bundle.end() || it->second == 0) {
            resources->insert({MEMORY_RESOURCE_NAME, 0});
        }
        for (auto &pair : bundle) {
            if (pair.second > 0) {
                resources->insert({pair.first, pair.second});
            }
        }
        *spec->add_bundles() = specBundle;
    }
    auto policy = ConvertStrategyToPolicy(this->rGroupSpec.strategy);
    spec->mutable_opt()->set_grouppolicy(policy);
}
}  // namespace Libruntime
}  // namespace YR