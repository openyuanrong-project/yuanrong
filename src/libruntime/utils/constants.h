/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 * Description: compiling and optimization for ARM and x86
 *
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

namespace YR {
namespace Libruntime {
#define UNLIKELY(x) __builtin_expect(!!(x), 0)
const int32_t DS_CONNECT_TIMEOUT = 30 * 60;
const int NO_TIMEOUT = -1;
const int ZERO_TIMEOUT = 0;
const int DEFAULT_TIMEOUT_MS = 600 * 1000;  // 10min
const int S_TO_MS = 1000;
const int GET_RETRY_INTERVAL = 1;  // seconds
const int LIMITED_RETRY_TIME = 2;
const size_t IN_MEM_STORE_SIZE_THRESHOLD = 100 * 1024;  // 100KB
const int DEFAULT_INVOKE_TIMEOUT = 900;                 // second
const int DISCONNECT_TIMEOUT_MS = 900 * 1000;           // 900 s
const int PROXY_DISCONNECT_TIMEOUT_MS = 60 * 1000;           // 60 s
const int RT_DISCONNECT_TIMEOUT_MS = 10 * 1000;         // 10 s
const char* const DEFAULT_SOCKET_PATH = "/home/snuser/socket/runtime.sock";
const char* const UNSUPPORTED_RGROUP_NAME = "primary";
const char* const CPU_RESOURCE_NAME = "CPU";
const char* const MEMORY_RESOURCE_NAME = "Memory";
const char* const STREAM_MODE = "STREAM_MODE";
const char* const MPMC = "MPMC";
const char* const MPSC = "MPSC";
const char* const SPSC = "SPSC";
}  // namespace Libruntime
}  // namespace YR