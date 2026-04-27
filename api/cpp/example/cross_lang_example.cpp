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

int main(int argc, char **argv)
{
    YR::Config conf;
    YR::Init(conf);

    {
        auto pyCls = YR::PyInstanceClass::FactoryCreate("pycallee", "SimpleInstance");
        auto pyIns = YR::Instance(pyCls).SetUrn("").Invoke();
        auto b = "def";
        auto obj = YR::Put(b);
        auto ret = pyIns.PyFunction<std::string>("show").Invoke(obj);
        auto res = *YR::Get(ret);
        std::cout << "SimpleInstance show result is " << res << std::endl;
    }

    {
        auto obj = YR::Put(10);
        auto r1 = YR::PyFunction<int>("pycallee", "add_one").SetUrn("").Invoke(obj);  // moduleName, functionName
        auto res = *YR::Get(r1);
        std::cout << "add one result is " << res << std::endl;
    }

    YR::Finalize();
    return 0;
}
