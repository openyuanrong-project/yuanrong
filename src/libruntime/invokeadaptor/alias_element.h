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
#include <string>

#include "src/utility/json_utility.h"

namespace YR {
namespace Libruntime {
struct RoutingRules {
    std::string ruleLogic;
    std::vector<std::string> rules;
    std::string grayVersion;
};

struct RoutingConfig {
    std::string functionVersionUrn;
    double weight;
};

struct AliasElement {
    std::string aliasUrn;
    std::string functionUrn;
    std::string functionVersionUrn;
    std::string name;
    std::string functionVersion;
    std::string revisionId;
    std::string description;
    std::string routingType;
    RoutingRules routingRules;
    std::vector<RoutingConfig> routingConfig;
};

void from_json(const nlohmann::json &j, std::vector<AliasElement> &schedulerInstanceList);

void from_json(const nlohmann::json &j, RoutingRules &routingRules);

void from_json(const nlohmann::json &j, std::vector<RoutingConfig> &schedulerInstanceList);

void to_json(nlohmann::json &j, const RoutingRules &routingRules);

void to_json(nlohmann::json &j, const RoutingConfig &routingConfig);

void to_json(nlohmann::json &j, const AliasElement &aliasElement);

}  // namespace Libruntime
}  // namespace YR