/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include <iostream>

#include "yr/yr.h"
//! [yr invoke]
// functions
int AddOne(int x)
{
    return x + 1;
}

YR_INVOKE(AddOne);

class Counter {
public:
    Counter() {}
    Counter(int init)
    {
        count = init;
    }

    static Counter *FactoryCreate(int init)
    {
        return new Counter(init);
    }

    int Add(int x)
    {
        count += x;
        return count;
    }

    int Get(void)
    {
        return count;
    }

    YR_STATE(count);

public:
    int count;
};

YR_INVOKE(Counter::FactoryCreate, &Counter::Add, &Counter::Get);
//! [yr invoke]
int main(void)
{
    YR::Config conf;
    conf.inCluster = true;
    YR::Init(conf);

    {
        //! [terminate instance]
        auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
        auto c = counter.Function(&Counter::Add).Invoke(1);
        std::cout << "counter is " << *YR::Get(c) << std::endl;
        counter.Terminate();
        //! [terminate instance]
    }

    {
        //! [terminate instance sync]
        auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
        auto c = counter.Function(&Counter::Add).Invoke(1);
        std::cout << "counter is " << *YR::Get(c) << std::endl;
        counter.Terminate(true);
        //! [terminate instance sync]
    }

    {
        //! [terminate instance async]
        auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
        auto c = counter.Function(&Counter::Add).Invoke(1);
        std::cout << "counter is " << *YR::Get(c) << std::endl;
        auto f = counter.AsyncTerminate(true);
        f.get();
        //! [terminate instance async]
    }

    {
        auto counter = YR::Instance(Counter::FactoryCreate, "name_1").Invoke(1);
        auto c = counter.Function(&Counter::Add).Invoke(1);
        std::cout << "counter is " << *YR::Get(c) << std::endl;
    }

    {
        //! [instance function]
        auto counter = YR::Instance(Counter::FactoryCreate, "name_1").Invoke(100);
        auto c = counter.Function(&Counter::Add).Invoke(1);
        std::cout << "counter is " << *YR::Get(c) << std::endl;
        //! [instance function]
    }

    {
        //! [Export]
        auto counter = YR::Instance(Counter::FactoryCreate).Invoke(100);
        auto out = counter.Export();
        //! [Export]
    }

    {
        //! [Import]
        NamedInstance<Counter> counter;
        counter.Import(in);
        //! [Import]
    }

    YR::Finalize();
    return 0;
}