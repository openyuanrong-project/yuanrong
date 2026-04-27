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

#include "alias_element.h"

namespace YR {
namespace Libruntime {
using nlohmann::json;
using namespace YR::utility;
void from_json(const json &j, std::vector<AliasElement> &schedulerInstanceList)
{
    for (auto &ele : j) {
        AliasElement aliasEle;
        JsonGetTo(ele, "aliasUrn", aliasEle.aliasUrn);
        JsonGetTo(ele, "functionUrn", aliasEle.functionUrn);
        JsonGetTo(ele, "functionVersionUrn", aliasEle.functionVersionUrn);
        JsonGetTo(ele, "name", aliasEle.name);
        JsonGetTo(ele, "functionVersion", aliasEle.functionVersion);
        JsonGetTo(ele, "revisionId", aliasEle.revisionId);
        JsonGetTo(ele, "description", aliasEle.description);
        JsonGetTo(ele, "routingType", aliasEle.routingType);
        if (aliasEle.routingType == "rule") {
            JsonGetTo(ele, "routingRules", aliasEle.routingRules);
        } else {
            JsonGetTo(ele, "routingconfig", aliasEle.routingConfig);
        }
        schedulerInstanceList.push_back(aliasEle);
    }
}

void from_json(const json &j, RoutingRules &routingRules)
{
    JsonGetTo(j, "ruleLogic", routingRules.ruleLogic);
    JsonGetTo(j, "rules", routingRules.rules);
    JsonGetTo(j, "grayVersion", routingRules.grayVersion);
}

void from_json(const json &j, std::vector<RoutingConfig> &routingConfig)
{
    for (auto &ele : j) {
        RoutingConfig conf;
        JsonGetTo(ele, "functionVersionUrn", conf.functionVersionUrn);
        JsonGetTo(ele, "weight", conf.weight);
        routingConfig.push_back(conf);
    }
}

void to_json(nlohmann::json &j, const RoutingRules &routingRules)
{
    j = json{{"ruleLogic", routingRules.ruleLogic},
             {"rules", routingRules.rules},
             {"grayVersion", routingRules.grayVersion}};
}

void to_json(nlohmann::json &j, const RoutingConfig &routingConfig)
{
    j = json{{"functionVersionUrn", routingConfig.functionVersionUrn}, {"weight", routingConfig.weight}};
}

void to_json(nlohmann::json &j, const AliasElement &aliasElement)
{
    j = json{{"aliasUrn", aliasElement.aliasUrn},
             {"functionUrn", aliasElement.functionUrn},
             {"functionVersionUrn", aliasElement.functionVersionUrn},
             {"name", aliasElement.name},
             {"functionVersion", aliasElement.functionVersion},
             {"revisionId", aliasElement.revisionId},
             {"description", aliasElement.description},
             {"routingType", aliasElement.routingType},
             {"routingRules", aliasElement.routingRules},
             {"routingConfig", aliasElement.routingConfig}};
}

}  // namespace Libruntime
}  // namespace YR