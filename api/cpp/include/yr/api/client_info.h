/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

namespace YR {
/**
 * @struct ClientInfo
 * @brief Object returned by `Init`, containing runtime context metadata.
 */
struct ClientInfo {
    /**
     * @brief The job id of the current context. Each Init generates a random job id.
     */
    std::string jobId;
    /**
     * @brief The version number of the current openYuanRong SDK.
     */
    std::string version;
    /**
     * @brief The version number of the server obtained after successful connection initialization.
     */
    std::string serverVersion;
};
}  // namespace YR