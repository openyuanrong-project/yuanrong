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
//! [cpp instance function]
class Counter {
public:
    int count;

    Counter() {}

    explicit Counter(int init) : count(init) {}

    static Counter *FactoryCreate()
    {
        int c = 10;
        return new Counter(c);
    }

    std::string RemoteVersion()
    {
        return "RemoteActor v0";
    }
};

YR_INVOKE(Counter::FactoryCreate, &Counter::RemoteVersion)

int main(void)
{
    YR::Config conf;
    YR::Init(conf);

    auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
    auto cppIns = YR::Instance(cppCls).Invoke();
    auto obj = cppIns.CppFunction<std::string>("&Counter::RemoteVersion").Invoke();
    auto res = *YR::Get(obj);
    std::cout << "add one result is " << res << std::endl;
    return 0;
}
//! [cpp instance function]