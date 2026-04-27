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

#include "yr/api/runtime_env.h"

namespace YR {

void RuntimeEnv::SetJsonStr(const std::string &name, const std::string &jsonStr)
{
    try {
        nlohmann::json valueJ = nlohmann::json::parse(jsonStr);
        jsons_[name] = valueJ;
    } catch (std::exception &e) {
        throw YR::Exception::InvalidParamException("Failed to set the field " + name + " by json string: " + e.what());
    }
}

std::string RuntimeEnv::GetJsonStr(const std::string &name) const
{
    if (!Contains(name)) {
        throw YR::Exception::InvalidParamException("The field " + name + " not found.");
    }
    auto j = jsons_[name].get<nlohmann::json>();
    return j.dump();
}

bool RuntimeEnv::Contains(const std::string &name) const
{
    return jsons_.contains(name);
}

bool RuntimeEnv::Remove(const std::string &name)
{
    if (Contains(name)) {
        jsons_.erase(name);
        return true;
    }
    return false;
}

}  // namespace YR