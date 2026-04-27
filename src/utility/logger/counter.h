/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#pragma once

namespace YR {
namespace utility {
const uint32_t DEFAULT_FREQUENCY = 60;

class Counter {
public:
    explicit Counter(uint32_t frequency) : counter_(0), frequent_(frequency == 0 ? DEFAULT_FREQUENCY : frequency) {}
    ~Counter() = default;
    bool Proc()
    {
        bool ret = false;
        if (counter_ % frequent_ == 0) {
            counter_ = 0;
            ret = true;
        }
        counter_++;
        return ret;
    }

private:
    uint32_t counter_;
    uint32_t frequent_;
};
}  // namespace utility
}  // namespace YR
