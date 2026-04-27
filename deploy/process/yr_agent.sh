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

set -o pipefail
# Permission control: remove others' permissions of directories and files
umask -p 027

BASE_DIR=$(dirname "$(readlink -f "$0")")
[[ ! -f "${BASE_DIR}/utils.sh" ]] && echo "${BASE_DIR}/utils.sh does not exist" && exit 1
. ${BASE_DIR}/utils.sh

CURRENT_PID=$$

if [ "X$YR_NODE_ID" != "X" ]; then
  NODE_ID=${YR_NODE_ID:0-20}
fi

function main() {
  log_info "start yr data plane..."
  source ${BASE_DIR}/config.sh --block=true "$@" --only_check_param

  if [ "$BLOCK" == "false" ]; then
    nohup bash ${BASE_DIR}/deploy.sh "$@" -a ${IP_ADDRESS} \
      > "${LOG_ROOT}/${NODE_ID}_deploy${STD_LOG_SUFFIX}" 2>&1 &
    local master_info_string
    for ((t = 1; t < 30; t++ )); do
      sleep 2
      if [ -f "${MASTER_INFO_OUT_FILE}" ]; then
        master_info_string=$( head -n 1 $MASTER_INFO_OUT_FILE )
        echo "${master_info_string}"
        return 0
      fi
    done
    log_error "wait start data plane timeout."
    return 98
  else
    bash ${BASE_DIR}/deploy.sh "$@" -a ${IP_ADDRESS}
  fi
}

main "$@"

if [ "$BLOCK" == "false" ]; then
  ret_code=$?
  if [ ${ret_code} -ne 0 ]; then
      log_error "yr-agent deploy failed, code: $ret_code"
      exit ${ret_code}
  fi
  exit 0
fi
