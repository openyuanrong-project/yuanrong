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

BASE_DIR=$(cd "$(dirname "$0")"; pwd)
PROJECT_DIR=$(cd "$(dirname "$0")"/../..; pwd)
OUTPUT_DIR="${BASE_DIR}/../../output/pattern/pattern_faas"
TAR_OUT_DIR="${PROJECT_DIR}/build/_output"

# cleanup and rebuild folders
cd "${PROJECT_DIR}" || die "${PROJECT_DIR} not exist"
[ -d "${OUTPUT_DIR}" ] && rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"
[ -d "${TAR_OUT_DIR}" ] && rm -rf "${TAR_OUT_DIR}"
mkdir -p "${TAR_OUT_DIR}"