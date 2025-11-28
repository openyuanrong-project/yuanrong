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

BASE_DIR=$(
    cd "$(dirname "$0")"
    pwd
)
FUNCTION_SYSTEM_PATH=$(readlink -m "${BASE_DIR}/../../functionsystem")
DATA_SYSTEM_PATH=$(readlink -m "${BASE_DIR}/../../datasystem")
THIRD_PARTY_PATH=$(readlink -m "${BASE_DIR}/../../third_party")

. "${BASE_DIR}"/utils.sh

function stop_monitor() {
  local pid=$(ps -ef | grep "${BASE_DIR}/deploy.sh" | grep -v  grep | awk '{print $2}')
  [ -z "$pid" ] && return 0
  ps -ef | grep "${BASE_DIR}/deploy.sh" | grep -v  grep |awk '{print $2}' | xargs kill -15
  log_info "stop monitor process success"
}

function stop_function_system() {
   local pid=$(ps -ef | grep "${FUNCTION_SYSTEM_PATH}/bin" | grep -v  grep | awk '{print $2}')
   [ -z "$pid" ] && return 0
   ps -ef | grep "${FUNCTION_SYSTEM_PATH}/bin" | grep -v  grep | awk '{print $2}' | xargs kill -15
   log_info "stop function system component success"
}

function stop_data_system() {
   local pid=$(ps -ef | grep "${DATA_SYSTEM_PATH}/service" | grep -v  grep | awk '{print $2}')
   [ -z "$pid" ] && return 0
   ps -ef | grep "${DATA_SYSTEM_PATH}" | grep -v  grep | awk '{print $2}' | xargs kill -15
   log_info "stop data system component success"
}

function stop_third_party() {
  local pid=$(ps -ef | grep "${THIRD_PARTY_PATH}" | grep -v  grep | awk '{print $2}')
   [ -z "$pid" ] && return 0
   ps -ef | grep "${THIRD_PARTY_PATH}" | grep -v  grep | awk '{print $2}' | xargs kill -15
   log_info "stop third party component success"
}

function main() {
    stop_monitor
    stop_function_system
    stop_data_system
    stop_third_party
}

main