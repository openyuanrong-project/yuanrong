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

int Handler(int x)
{
    return x;
}

YR_INVOKE(f);

int main(void)
{
    YR::Config conf;
    YR::Init(conf);

    {
        //! [Wait multiple objects]
        int num = 5;
        std::size_t waitNum = 1;
        std::vector<YR::ObjectRef<int>> vec;
        for (int i = 0; i < num; ++i) {
            auto obj = YR::Function(Handler).Invoke(i);
            vec.emplace_back(std::move(obj));
        }
        int timeout = 30;
        auto waitResult = YR::Wait(vec, waitNum, timeout);
        std::cout << waitResult.first.size() << std::endl;
        std::cout << waitResult.second.size() << std::endl;
        //! [Wait multiple objects]
    }

    {
        //! [Wait a single object]
        int timeout = 30;
        auto obj = YR::Function(Handler).Invoke(1);
        YR::Wait(obj, timeout);
        //! [Wait a single object]
    }

    return 0;
}