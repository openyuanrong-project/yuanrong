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

class RangeCounter {
public:
    RangeCounter() {}
    explicit RangeCounter(int init)
    {
        count = init;
    }

    static RangeCounter *FactoryCreate(int init)
    {
        return new RangeCounter(init);
    }

    int Add(int x)
    {
        count += x;
        return count;
    }

    YR_STATE(count);

public:
    int count;
};
//! [GetInstances]
int main(void)
{
    YR::Config conf;
    YR::Init(conf);

    YR::InstanceRange range;
    int rangeMax = 10;
    int rangeMin = 1;
    int rangeStep = 2;
    int rangeTimeout = 10;
    range.max = rangeMax;
    range.min = rangeMin;
    range.step = rangeStep;
    range.sameLifecycle = true;
    YR::RangeOptions rangeOpts;
    rangeOpts.timeout = rangeTimeout;
    range.rangeOpts = rangeOpts;
    YR::InvokeOptions opt;
    opt.instanceRange = range;
    auto instances = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    auto insList = instances.GetInstances(5);

    for (auto ins : insList) {
        auto res = ins.Function(&Counter::Add).Invoke(1);
        std::cout << "res is " << *YR::Get(res) << std::endl;
    }
    instances.Terminate();
    YR::Finalize();

    return 0;
}
//! [GetInstances]