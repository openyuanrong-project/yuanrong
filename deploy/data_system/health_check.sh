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

# must > 60 (waiting for datasystem start timeout)
datasystemTimeOut=65
# max wait time (second) for data_system worker
readonly WAIT_DS_WORKER_MAX_TIME=$datasystemTimeOut
# max wait time (second) for data_system master
readonly WAIT_DS_MASTER_MAX_TIME=$datasystemTimeOut

function health_check_by_pid_and_file() {
  local pid=$1
  local file=$2
  local wait_max_time=$3
  local i=0
  local ret_code=0
  for ((i = 0; i <= wait_max_time; i++)); do
    if ! kill -0 ${pid} &>/dev/null; then
      ret_code=1
      return ${ret_code}
    fi
    if [ ! -f "${file}" ]; then
      ret_code=2
      sleep 1
      continue
    fi
    return 0
  done
  return ${ret_code}
}

function data_system_health_check() {
  case "$1" in
  ds_worker)
    health_check_by_pid_and_file "$2" "${DS_WORKER_HEALTH_CHECK_PATH}" ${WAIT_DS_WORKER_MAX_TIME}
    ;;
  ds_master)
    health_check_by_pid_and_file "$2" "${DS_MASTER_HEALTH_CHECK_PATH}" ${WAIT_DS_MASTER_MAX_TIME}
    ;;
  *)
    log_warning >&2 "Unknown component"
    return 1
    ;;
  esac
  return $?
}
