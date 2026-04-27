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

#include <string>
#include <vector>

#include "src/dto/data_object.h"

namespace YR {
void ThrowIfTrue(bool condition, Libruntime::ErrorCode code, const std::string &msg);

std::string ConvertFunctionUrnToId(const std::string &functionUrn);
YR::Libruntime::ErrorInfo WriteDataObject(const void *data, std::shared_ptr<YR::Libruntime::DataObject> &dataObj,
                                          uint64_t size, const std::unordered_set<std::string> &nestedIds);
std::string GetEnv(const std::string &key);
bool SetEnv(const std::string &k, const std::string &v);
}  // namespace YR
