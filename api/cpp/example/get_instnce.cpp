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
//! [get instance]
class Counter {
public:
    int count;

    Counter() {}

    explicit Counter(int init) : count(init) {}

    static Counter *FactoryCreate(int init)
    {
        return new Counter(init);
    }

    int Counter::Add(int x)
    {
        count += x;
        return count;
    }
};

YR_INVOKE(Counter::FactoryCreate, &Counter::Add);

int main(void)
{
    YR::Config conf;
    YR::Init(conf);

    std::string name = "test-get-instance";
    auto ins = YR::Instance(Counter::FactoryCreate, name).Invoke(1);
    auto res = ins.Function(&Counter::Add).Invoke(1);

    std::string ns = "";
    auto insGet = YR::GetInstance<Counter>(name, ns, 60);
    auto resGet = insGet.Function(&Counter::Add).Invoke(1);
    return 0;
}
//! [get instance]