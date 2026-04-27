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

#include "yr/parallel/parallel_for.h"

#include "api/cpp/src/config_manager.h"
#include "src/libruntime/libruntime_manager.h"
namespace YR {
namespace Parallel {

int GetThreadPoolSize()
{
    if (internal::IsLocalMode()) {
        return ConfigManager::Singleton().threadPoolSize;
    } else {
        return YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->GetThreadPoolSize();
    }
}

}  // namespace Parallel
}  // namespace YR