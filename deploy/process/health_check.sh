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

BASE_DIR=$(
  cd "$(dirname "$0")"
  pwd
)
FUNCTION_SYSTEM_PATH=${BASE_DIR}/../../functionsystem
DATA_SYSTEM_PATH=${BASE_DIR}/../../datasystem
THIRD_PARTY_PATH=${BASE_DIR}/../../third_party

[[ ! -f "${FUNCTION_SYSTEM_PATH}/deploy/health_check.sh" ]] && echo "${FUNCTION_SYSTEM_PATH}/deploy/health_check.sh is not exist" && exit 1
. ${FUNCTION_SYSTEM_PATH}/deploy/health_check.sh
[[ ! -f "${DATA_SYSTEM_PATH}/deploy/health_check.sh" ]] && echo "${DATA_SYSTEM_PATH}/deploy/health_check.sh is not exist" && exit 1
. ${DATA_SYSTEM_PATH}/deploy/health_check.sh
[[ ! -f "${THIRD_PARTY_PATH}/health_check.sh" ]] && echo "${THIRD_PARTY_PATH}/health_check.sh is not exist" && exit 1
. ${THIRD_PARTY_PATH}/health_check.sh


function health_check() {
    case "$1" in
    function_proxy)
        function_system_health_check "$2" "${FUNCTION_PROXY_PORT}" "local-scheduler"
        ;;
    function_agent)
        function_system_health_check "$2" "${FUNCTION_AGENT_PORT}" "function-agent"
        ;;
    function_master)
        function_system_health_check "$2" "${GLOBAL_SCHEDULER_PORT}" "global-scheduler"
        ;;
    iam_server)
        function_system_health_check "$2" "${IAM_SERVER_PORT}" "iam-server"
        ;;
    ds_worker|ds_master)
        data_system_health_check "$1" "$2"
        ;;
    dashboard|collector)
        dashboard_health_check "$2"
        ;;
    faas_frontend|function_scheduler)
        faas_frontend_health_check "$2"
        ;;
    metaservice)
        metaservice_health_check "$2"
        ;;
    etcd)
        third_party_health_check "$2"
        ;;
    *)
        if [[ "$1" == etcd_proxy_* ]]; then
          third_party_health_check "$2"
        else
          log_warning >&2 "Unknown component: $1"
          return 1
        fi
        ;;
    esac
    return $?
}