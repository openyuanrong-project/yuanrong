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

#include <climits>
#include <string>

namespace YR {
const int DEFAULT_GET_TIMEOUT_SEC = 300;
const std::string CONCURRENCY_KEY = "Concurrency";
const int NO_TIMEOUT = -1;
const int S_TO_MS = 1000;
const int TIMEOUT_MAX = INT_MAX / S_TO_MS;
/*! @property std::string RESOURCE
 *  @brief Affinity kind, indicating the predefined resource label affinity.
 */
const std::string RESOURCE = "Resource";
/*! @property std::string INSTANCE
 *  @brief Affinity kind, indicating the dynamic instance label affinity.
 */
const std::string INSTANCE = "Instance";
/*! @property std::string PREFERRED
 *  @brief Affinity type, indicating weak affinity.
 */
const std::string PREFERRED = "PreferredAffinity";
/*! @property std::string PREFERRED_ANTI
 *  @brief Affinity type, indicating weak anti-affinity.
 */
const std::string PREFERRED_ANTI = "PreferredAntiAffinity";
/*! @property std::string REQUIRED
 *  @brief Affinity type, indicating strong affinity.
 */
const std::string REQUIRED = "RequiredAffinity";
/*! @property std::string REQUIRED_ANTI
 *  @brief Affinity type, indicating strong anti-affinity.
 */
const std::string REQUIRED_ANTI = "RequiredAntiAffinity";
/*! @property std::string LABEL_IN
 *  @brief Label operation type, indicating that the corresponding label has a corresponding value.
 */
const std::string LABEL_IN = "LabelIn";
/*! @var LABEL_NOT_IN
 *  @brief Label operation type, indicating that the corresponding label does not have a corresponding value.
 */
const std::string LABEL_NOT_IN = "LabelNotIn";
/*! @var LABEL_EXISTS
 *  @brief Label operation type, indicating the presence of corresponding tags.
 *
 */
const std::string LABEL_EXISTS = "LabelExists";
/*! @var LABEL_DOES_NOT_EXIST
 *  @brief Label operation type, indicating that there is no corresponding label.
 */
const std::string LABEL_DOES_NOT_EXIST = "LabelDoesNotExist";
/*! @var AFFINITYSCOPE_POD
 *  @brief Instance affinity scope, indicating Pod-level affinity.
 */
const std::string AFFINITYSCOPE_POD = "POD";
/*! @var AFFINITYSCOPE_NODE
 *  @brief Instance affinity scope, indicating Node-level affinity.
 */
const std::string AFFINITYSCOPE_NODE = "NODE";
const uint32_t DEFAULT_HTTP_IOC_THREADS_NUM = 200;
const size_t MAX_OPTIONS_RETRY_TIME = 10;
const int DEFAULT_RECYCLETIME = 2;
const std::string FUNCTION_NOT_REGISTERED_ERROR_MSG = "Function to be invoked is not registered by using YR_INVOKE";
const int DEFAULT_SAVE_LOAD_STATE_TIMEOUT = 30;
const int DEFAULT_INSTANCE_RANGE_NUM = -1;
const int DEFAULT_INSTANCE_RANGE_STEP = 2;
const int DEFAULT_TIMEOUT_MS = 600 * 1000;  // 10min
const int GET_RETRY_INTERVAL = 1;           // seconds
const int LIMITED_RETRY_TIME = 2;
}  // namespace YR