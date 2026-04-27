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

//! [Function]
#include "yr/yr.h"

int PlusOne(int x)
{
    return x + 1;
}

YR_INVOKE(PlusOne)

int main()
{
    YR::Config conf;
    YR::Init(conf);

    std::vector<YR::ObjectRef<int>> n2;
    int k = 1;
    for (int i = 0; i < k; i++) {
        auto r2 = YR::Function(PlusOne).Invoke(2);
        n2.emplace_back(r2);
    }
    for (int i = 0; i < k; i++) {
        auto integer = *YR::Get(n2[i]);
        printf("%d :%d\n", i, integer);
    }

    YR::Finalize();
    return 0;
}
//! [Function]