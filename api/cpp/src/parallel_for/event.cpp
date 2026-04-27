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

#include "event.h"
#include <chrono>
namespace YR {
namespace Parallel {
void Event::EventWait()
{
    std::unique_lock<std::mutex> lock(mutex);
    cond.wait(lock, [&] { return ready; });
}

void Event::EventReady()
{
    std::unique_lock<std::mutex> lock(mutex);
    ready = false;
}

void Event::EventWakeUp()
{
    std::unique_lock<std::mutex> lock(mutex);
    ready = true;
    cond.notify_all();
}
}  // namespace Parallel
}  // namespace YR