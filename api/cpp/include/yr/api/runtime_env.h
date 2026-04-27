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

#include <list>

#include "json.hpp"

#include "yr/api/constant.h"
#include "yr/api/exception.h"

namespace YR {
class RuntimeEnv {
public:
    void SetJsonStr(const std::string &name, const std::string &jsonStr);

    template <typename T>
    void Set(const std::string &name, const T &value);

    bool Empty() const
    {
        return jsons_.empty();
    }

    template <typename T>
    T Get(const std::string &name) const;

    std::string GetJsonStr(const std::string &name) const;

    bool Contains(const std::string &name) const;

    bool Remove(const std::string &name);

private:
    nlohmann::json jsons_;
};

template <typename T>
inline T RuntimeEnv::Get(const std::string &name) const
{
    if (!Contains(name)) {
        throw YR::Exception::InvalidParamException("The field " + name + " not found.");
    }
    try {
        return jsons_[name].get<T>();
    } catch (std::exception &e) {
        throw YR::Exception::InvalidParamException("Failed to get the field " + name + ": " + e.what());
    }
}

template <typename T>
inline void RuntimeEnv::Set(const std::string &name, const T &value)
{
    try {
        nlohmann::json valueJ = value;
        jsons_[name] = valueJ;
    } catch (std::exception &e) {
        throw YR::Exception::InvalidParamException("Failed to set the field " + name + ": " + e.what());
    }
}
}  // namespace YR