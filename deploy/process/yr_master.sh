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

BASE_DIR=$(
  cd "$(dirname "$0")"
  pwd
)

# Permission control: remove others' permissions of directories and files
umask -p 027

function main() {
  local ret_code
  source ${BASE_DIR}/config.sh "$@" --master -g ON --only_check_param
  ret_code=$?
  if [ ${ret_code} -ne 0 ]; then
    return ${ret_code}
  fi
  local master_pid_info_file="${INSTALL_DIR_PARENT}/master_pid_port.txt"
  [ -f "${master_pid_info_file}" ] && rm -f ${master_pid_info_file}
  if [ "X${MASTER_INFO_OUT_FILE}" = "X" ]; then
    MASTER_INFO_OUT_FILE="${INSTALL_DIR_PARENT}/master.info"
    [ -f "${MASTER_INFO_OUT_FILE}" ] && rm -f ${MASTER_INFO_OUT_FILE}
  else
    # if master info parent dir is not exist, need create first
    rm -rf ${MASTER_INFO_OUT_FILE}
    mkdir -p ${MASTER_INFO_OUT_FILE}
    rm -rf ${MASTER_INFO_OUT_FILE}
  fi
  if [ "$BLOCK" == "true" ]; then
    bash ${BASE_DIR}/deploy.sh "$@" --master --master_info_output ${MASTER_INFO_OUT_FILE} -g TRUE -a ${IP_ADDRESS}
    return 0
  fi
  nohup bash ${BASE_DIR}/deploy.sh "$@" --master --master_info_output ${MASTER_INFO_OUT_FILE} -g TRUE -a ${IP_ADDRESS} \
   > "${LOG_ROOT}/${NODE_ID}-deploy${STD_LOG_SUFFIX}" 2>&1 &
  log_info "wait start control plane..."
  local t master_pid_info master_port_info
  for ((t = 1; t < 30; t++ )); do
    sleep 2
    if [ -f "${master_pid_info_file}" ]; then
      master_pid_info=$( head -n 1 $master_pid_info_file )
      if [ "X${master_pid_info}" = "X" ]; then
        continue
      fi
      log_info ${master_pid_info}
      if [[ "${master_pid_info}" = error* ]]; then
        return 98
      fi
      master_port_info=$( head -n 2 $master_pid_info_file | tail -1 )
      log_info ${master_port_info}
      log_info "control plane start success"
      break
    fi
  done
  if [ -z "${master_port_info}" ]; then
    log_error "wait start control plane timeout"
    return 98
  fi
  log_info "wait start data plane..."
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
}

main "$@"
ret_code=$?
if [ ${ret_code} -ne 0 ]; then
    log_error "yr-master deploy failed, code: $ret_code"
    exit ${ret_code}
fi
exit 0
