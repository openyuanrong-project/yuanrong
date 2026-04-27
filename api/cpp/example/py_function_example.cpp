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
//! [py function]
int main(void)
{
    // def add_one(a):
    //     return a + 1
    YR::Config conf;
    YR::Init(conf);

    auto r1 = YR::PyFunction<int>("pycallee", "add_one").Invoke(x);  // moduleName, functionName
    auto res = *YR::Get(r1);

    std::cout << "PlusOneWithPyFunc with result=" << res << std::endl;
    return res;

    return 0;
}
//! [py function]