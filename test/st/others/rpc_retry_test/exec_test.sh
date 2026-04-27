#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

export BASE_DIR=$(
    cd "$(dirname "$0")"
    pwd
)




# 调小一点测得快
export REQUEST_ACK_ACC_MAX_SEC=180

# (1) MultiLanguageRuntime 向 proxy 发出一个 Create 请求。期望：立即收到bus回复。没有重试。
 ${BASE_DIR}/build/main CREATE 0

# (2) 生成持续时间190秒的 runtime --> bus 网络故障。MultiLanguageRuntime 向 proxy 发出一个 Create 请求。期望：一直没有收到回复，170秒是最后一次重试，之后不再重试。180秒生成错误信息。
 ${BASE_DIR}/build/main CREATE 190

# (3) MultiLanguageRuntime 向 proxy 发出一个 Invoke 请求，立即收到bus回复。没有重试。
 ${BASE_DIR}/build/main INVOKE 0

# (4) 生成持续时间190秒的 runtime --> bus 网络故障。MultiLanguageRuntime 向 proxy 发出一个 Invoke 请求，第一次经过10s重试，之后每次重试时间翻倍，第180秒后不再重试。
${BASE_DIR}/build/main INVOKE 190