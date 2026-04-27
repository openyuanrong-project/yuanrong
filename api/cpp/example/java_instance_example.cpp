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

//! [java instance function]
int main(void)
{
    // package io.yuanrong.demo;
    //
    // // A regular Java class.
    // public class Counter {
    //
    //     private int value = 0;
    //
    //     public int increment() {
    //         this.value += 1;
    //         return this.value;
    //     }
    // }
    YR::Config conf;
    YR::Init(conf);
    auto javaInstance = YR::JavaInstanceClass::FactoryCreate("io.yuanrong.demo.Counter");
    auto r1 = YR::Instance(javaInstance).Invoke();
    auto r2 = r1.JavaFunction<int>("increment").Invoke(1);
    auto res = *YR::Get(r2);
    std::cout << "PlusOneWithJavaClass with result=" << res << std::endl;
    return res;
}
//! [java instance function]