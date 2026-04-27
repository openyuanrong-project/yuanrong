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

#include "yr/yr.h"

int main(void)
{
    YR::Init(YR::Config{});

    {
        //! [Get and put a single object]
        auto objRef = YR::Put(100);
        auto value = *(YR::Get(objRef));
        assert(value == 100);  // should be 100
        //! [Get and put a single object]
    }

    {
        //! [Get and put a single object with param]
        YR::CreateParam param;
        param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        param.consistencyType = YR::ConsistencyType::PRAM;
        param.cacheType = YR::CacheType::DISK;
        auto objRef = YR::Put(100, param);
        auto value = *(YR::Get(objRef));
        assert(value == 100);  // should be 100
        //! [Get and put a single object with param]
    }

    {
        //! [Get and put multiple objects]
        int originValue = 100;
        auto objRefs = std::vector<YR::ObjectRef<int>>{YR::Put(100), YR::Put(101)};
        auto values = *(YR::Get(objRefs));
        assert(values.size() == 2);  // should be [100, 101]
        assert(*values[0] == 100);
        assert(*values[1] == 101);
        //! [Get and put multiple objects]
    }

    return 0;
}