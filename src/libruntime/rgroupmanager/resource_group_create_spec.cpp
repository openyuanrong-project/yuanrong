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

namespace YR {
namespace Libruntime {

GroupPolicy ConvertStrategyToPolicy(const std::string &stategy)
{
    static const std::unordered_map<std::string, GroupPolicy> strategyMap = {
        {"PACK", GroupPolicy::Pack},
        {"STRICT_PACK", GroupPolicy::StrictPack},
        {"SPREAD", GroupPolicy::Spread},
        {"STRICT_SPREAD", GroupPolicy::StrictSpread}};
    auto it = strategyMap.find(stategy);
    return (it != strategyMap.end()) ? it->second : GroupPolicy::None;
}

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