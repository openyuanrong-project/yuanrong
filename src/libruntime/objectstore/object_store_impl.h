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

#include <iterator>
#include <sstream>

#include "datasystem/object_client.h"
#include "reference_count_map.h"
#include "src/libruntime/objectstore/object_store.h"
#include "src/libruntime/utils/constants.h"
#include "src/libruntime/utils/datasystem_utils.h"
#include "src/libruntime/utils/exception.h"
#include "src/libruntime/utils/utils.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
namespace ds = datasystem;
using namespace std::chrono;
ErrorInfo IncreaseRefReturnCheck(const ds::Status &status, const std::vector<std::string> &failedObjectIds);
ErrorInfo DecreaseRefReturnCheck(const ds::Status &status, const std::vector<std::string> &failedObjectIds);
}  // namespace Libruntime
}  // namespace YR