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
//! [yr shutdown]

class Counter {
public:
    int count;

    Counter() {}
    explicit Counter(int init) : count(init) {}

    static Counter *FactoryCreate(int init)
    {
        return new Counter(init);
    }

    void Counter::MyShutdown(uint64_t gracePeriodSecond)
    {
        YR::KV().Set("myKey", "myValue");
    }

    int Add(int x)
    {
        count += x;
        return count;
    }
};

YR_INVOKE(Counter::FactoryCreate, &Counter::Add)
YR_SHUTDOWN(&Counter::MyShutdown);

int main(int argc, char **argv)
{
    YR::KV().Del("myKey");
    auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
    auto ret = counter.Function(&Counter::Add).Invoke(1);
    std::cout << *YR::Get(ret) << std::endl; // 2
    // Using the KV interface in the cloud to get the value mapped by `my_key` will return `'my_value'`.
    counter.Terminate();
    std::string result = YR::KV().Get("myKey", 30);
    EXPECT_EQ(result, "myValue");

    YR::KV().Del("myKey");
    YR::InvokeOptions opt;
    opt.customExtensions.insert({"GRACEFUL_SHUTDOWN_TIME", "10"});
    auto counter2 = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
    auto ret2 = counter2.Function(&Counter::Add).Invoke(1);
    std::cout << *YR::Get(ret2) << std::endl; // 2

    counter2.Terminate();
    std::string result2 = YR::KV().Get("myKey", 30);
    EXPECT_EQ(result2, "myValue");
}
//! [yr shutdown]