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
//! [py instance function]
int main(void)
{
    // class Instance:
    //     sum = 0
    //
    //     def __init__(self, init):
    //       self.sum = init
    //
    //     def add(self, a):
    //       self.sum += a
    //
    //     def get(self):
    //       return self.sum
    YR::Config conf;
    YR::Init(conf);

    auto pyInstance = YR::PyInstanceClass::FactoryCreate("pycallee", "Instance");  // moduleName, className
    auto r1 = YR::Instance(pyInstance).Invoke(x);
    r1.PyFunction<void>("add").Invoke(1);  // returnType, memberFunctionName

    auto r2 = r1.PyFunction<int>("get").Invoke();
    auto res = *YR::Get(r2);

    std::cout << "PlusOneWithPyClass with result=" << res << std::endl;
    return res;

    return 0;
}
//! [py instance function]