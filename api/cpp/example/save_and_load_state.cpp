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

//! [save and load state]
#include "yr/yr.h"

class Counter {
public:
    Counter() {}
    ~Counter() {}
    explicit Counter(int init) : count(init) {}

    static Counter *FactoryCreate(int init)
    {
        return new Counter(init);
    }

    int Save()
    {
        YR::SaveState();
        return count;
    }

    int Load()
    {
        YR::LoadState();
        return count;
    }

    int Add(int x)
    {
        count += x;
        return count;
    }

    YR_STATE(count)

private:
    int count;
};

YR_INVOKE(Counter::FactoryCreate, &Counter::Add, &Counter::Save, &Counter::Load)

int main()
{
    YR::Config conf;
    conf.mode = YR::Config::Mode::CLUSTER_MODE;
    YR::Init(conf);
    auto creator1 = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto member1 = creator1.Function(&Counter::Add).Invoke(3);
    auto res1 = *YR::Get(member1);
    printf("res1 is %d\n", res1);  // 4
    auto member2 = creator1.Function(&Counter::Save).Invoke();
    auto res2 = *YR::Get(member2);
    printf("res2 is %d\n", res2);  // 4
    auto member3 = creator1.Function(&Counter::Add).Invoke(3);
    auto res3 = *YR::Get(member3);
    printf("ref3 is %d\n", res3);  // 7
    auto member4 = creator1.Function(&Counter::Load).Invoke();
    auto res4 = *YR::Get(member4);
    printf("res4 is %d\n", res4);  // 7
    auto member5 = creator1.Function(&Counter::Add).Invoke(3);
    auto res5 = *YR::Get(member5);
    printf("res5 is %d\n", res5);  // 7
    YR::Finalize();
    return 0;
}
//! [save and load state]