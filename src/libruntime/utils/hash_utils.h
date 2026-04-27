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

#pragma once
#include <string>

#include "datasystem/utils/sensitive_value.h"

namespace YR {
namespace Libruntime {
using SensitiveValue = datasystem::SensitiveValue;
std::string GetHMACSha256(const SensitiveValue &key, const std::string &data);
void SHA256AndHex(const std::string &input, std::stringstream &output);
}  // namespace Libruntime
}  // namespace YR