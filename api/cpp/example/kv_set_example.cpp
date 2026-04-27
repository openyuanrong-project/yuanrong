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

#include "yr/yr.h"

int main()
{
    YR::Config conf;
    conf.mode = YR::Config::Mode::CLUSTER_MODE;
    YR::Init(conf);
    {
        //! [set char*]
        std::string result{"result"};
        YR::KV().Set("key", result.c_str());  // setVal
        //! [set char*]
    }
    {
        //! [set char* with len]
        std::string result{"result"};
        auto size = result.size() + 1;
        YR::KV().Set("key", result.c_str(), size);  // setValLen
        //! [set char* with len]
    }
    {
        //! [set string]
        std::string result{"result"};
        YR::KV().Set("key", result);
        //! [set string]
    }
    {
        //! [set char* with param]
        YR::SetParam setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        std::string result{"result"};
        YR::KV().Set("key", result.c_str(), setParam);
        //! [set char* with param]
    }
    {
        //! [set char* with len and param]
        YR::SetParam setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        std::string result{"result"};
        auto size = result.size() + 1;
        YR::KV().Set("key", result.c_str(), size, setParam);
        //! [set char* with len and param]
    }
    {
        //! [set string with param]
        YR::SetParam setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        YR::KV().Set("kv-key", "kv-value", setParam);
        //! [set string with param]
    }
    {
        //! [set char* with param v2]
        YR::SetParamV2 setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        setParam.cacheType = YR::CacheType::MEMORY;
        std::string result{"result"};
        YR::KV().Set("key", result.c_str(), setParam);
        //! [set char* with param v2]
    }
    {
        //! [set char* with len and param v2]
        YR::SetParamV2 setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        setParam.cacheType = YR::CacheType::MEMORY;
        std::string result{"result"};
        auto size = result.size() + 1;
        YR::KV().Set("key", result.c_str(), size, setParam);
        //! [set char* with len and param v2]
    }
    {
        //! [set string with param v2]
        YR::SetParamV2 setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        setParam.cacheType = YR::CacheType::MEMORY;
        YR::KV().Set("kv-key", "kv-value", setParam);
        //! [set string with param v2]
    }
    return 0;
}