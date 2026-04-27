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

class Counter {
public:
    Counter() = default;
    Counter(std::string name, int limit) : name(name), limit(limit) {}
    std::string name;
    int limit;
    MSGPACK_DEFINE(name, limit)
};

int main()
{
    YR::Config conf;
    conf.mode = YR::Config::Mode::CLUSTER_MODE;
    YR::Init(conf);
    {
        //! [multi writeTx objects]
        int count = 100;
        Counter c("Counter1-", count);
        std::vector<std::string> keys = {c.name};
        std::vector<Counter> vals = {c};
        try {
            YR::KV().MWriteTx<Counter>(keys, vals, YR::ExistenceOpt::NX);
        } catch (YR::Exception &e) {
            // handle exception...
        }
        //! [multi writeTx objects]
    }

    {
        //! [multi writeTx objects with param]
        int count = 100;
        Counter c("Counter1-", count);
        std::vector<std::string> keys = {c.name};
        std::vector<Counter> vals = {c};
        YR::MSetParam param;
        param.ttlSecond = 0;
        param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        param.existence = YR::ExistenceOpt::NONE;
        param.cacheType = YR::CacheType::MEMORY;
        try {
            YR::KV().MWriteTx<Counter>(keys, vals, param);
        } catch (YR::Exception &e) {
            // handle exception...
        }
        //! [multi writeTx objects with param]
    }

    {
        //! [read objects]
        int count = 100;
        Counter c1("Counter1-", count);
        auto result = YR::KV().Write(c1.name, c1);
        auto v1 = *(YR::KV().Read<Counter>(c1.name));  // get Counter
        Counter c2("Counter2-", count);
        result = YR::KV().Write(c2.name, c2);
        std::vector<std::string> keys{c1.name, c2.name};
        auto returnVal = YR::KV().Read<Counter>(keys);  // get std::vector<shared_ptr<Counter>>
        //! [read objects]
    }

    {
        //! [write objects]
        int count = 100;
        Counter c("Counter1-", count);
        try {
            YR::KV().Write<Counter>(c.name, c);
        } catch (YR::Exception &e) {
            // handle exception...
        }
        //! [write objects]
    }

    {
        //! [write objects with param]
        int count = 100;
        Counter c("Counter1-", count);
        YR::SetParam setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        try {
            YR::KV().Write<Counter>(c.name, c, setParam);
        } catch (YR::Exception &e) {
            // handle exception...
        }
        //! [write objects with param]
    }

    {
        //! [write objects with param v2]
        int count = 100;
        Counter c("Counter1-", count);
        YR::SetParamV2 setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        setParam.cacheType = YR::CacheType::MEMORY;
        try {
            YR::KV().Write<Counter>(c.name, c, setParam);
        } catch (YR::Exception &e) {
            // handle exception...
        }
        //! [write objects with param v2]
    }
    return 0;
}