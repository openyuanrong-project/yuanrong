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

#include <string>

namespace YR {
namespace Libruntime {
const int DEFAULT_CONCURRENCY = 1;
const size_t MIN_CONCURRENCY = 1;
const size_t MAX_CONCURRENCY = 1000;
const size_t MAX_PODLABELS = 5;
const long DEFAULT_ALARM_TIMESTAMP = -1L;
const long DEFAULT_ALARM_TIMEOUT = -1L;
const int MILLISECOND_UNIT = 1000;  // 1s = 1000ms
const int RETRY_TIME = 3;
const int RECONNECT_TIMES = 450;
const int INTERVAL_TIME = 2;
const int DEFAULT_INSTANCE_RANGE_NUM = -1;
const int DEFAULT_INSTANCE_RANGE_STEP = 2;
const int EVENT_SIGNAL_TIMEOUT_SECOND = 10;
const std::string RESOURCE = "Resource";
const std::string INSTANCE = "Instance";
const std::string PREFERRED = "PreferredAffinity";
const std::string PREFERRED_ANTI = "PreferredAntiAffinity";
const std::string REQUIRED = "RequiredAffinity";
const std::string REQUIRED_ANTI = "RequiredAntiAffinity";
const std::string LABEL_IN = "LabelIn";
const std::string LABEL_NOT_IN = "LabelNotIn";
const std::string LABEL_EXISTS = "LabelExists";
const std::string LABEL_DOES_NOT_EXIST = "LabelDoesNotExist";
const std::string AFFINITYSCOPE_POD = "POD";
const std::string AFFINITYSCOPE_NODE = "NODE";
const std::string DEFAULT_YR_NAMESPACE = "yr_defalut_namespace";
const std::string EVENT_EOF = "yuanrong_event_EOF";
const std::string INSTANCE_REQUIREMENT_ACCEPT = "accept";
const std::string INSTANCE_REQUIREMENT_ACCEPT_EVENT_STREAM = "text/event-stream";
}  // namespace Libruntime
}  // namespace YR
