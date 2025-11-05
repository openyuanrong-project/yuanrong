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
#include <memory>
#include <optional>
#include <string>
#include "datasystem/object_client.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/utils/datasystem_utils.h"

namespace YR {
namespace Libruntime {
class DeviceBuffer {
};

enum class DeviceBufferLifetimeType : uint8_t {
    REFERENCE = 0,
    MOVE = 1,
};

struct DeviceBufferParam {
    DeviceBufferLifetimeType lifetime = DeviceBufferLifetimeType::REFERENCE;
    bool cacheLocation = true;
};
}  // namespace Libruntime
}  // namespace YR
