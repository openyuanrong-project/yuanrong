/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: used to block and wake threads
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

#include <condition_variable>
#include <mutex>
namespace YR {
namespace Parallel {
class Event {
public:
    Event(){};
    ~Event(){};

    void EventReady();
    void EventWait();
    void EventWakeUp();

private:
    bool ready = false;
    std::mutex mutex;
    std::condition_variable cond;
};
}  // namespace Parallel
}  // namespace YR