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

#include "metrics_context.h"

#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
std::string MetricsContext::GetAttr(const std::string &attr) const
{
    auto it = attribute_.find(attr);
    if (it != attribute_.end()) {
        return it->second;
    }
    return "";
}
void MetricsContext::SetAttr(const std::string &attr, const std::string &value)
{
    attribute_.insert_or_assign(attr, value);
}
} // namespace metrics
} // namespace functionsystem
