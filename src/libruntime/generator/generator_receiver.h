/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

#include "src/libruntime/err_type.h"
#include "src/libruntime/utils/utils.h"
#include "src/utility/timer_worker.h"

namespace YR {
namespace Libruntime {
class GeneratorReceiver {
public:
    virtual void Initialize(void) = 0;
    virtual void Stop(void) = 0;
    virtual void MarkEndOfStream(const std::string &genId, const ErrorInfo &errInfo) = 0;
    virtual void AddRecord(const std::string &genId) = 0;
};
}  // namespace Libruntime
}  // namespace YR