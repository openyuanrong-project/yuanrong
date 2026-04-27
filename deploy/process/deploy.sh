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
FUNCTION_SYSTEM_PATH=${BASE_DIR}/../../functionsystem
DATA_SYSTEM_PATH=${BASE_DIR}/../../datasystem
THIRD_PARTY_PATH=${BASE_DIR}/../../third_party
ETCDCTL_PATH=${THIRD_PARTY_PATH}/etcd/etcdctl
ETCD_LD_PATH=${THIRD_PARTY_PATH}/etcd/lib:${LD_LIBRARY_PATH}
readonly EXIT_MISUSE=2
readonly CONTROL_PLANE_RETRY_TIMES=5
readonly RETRY_TIMES=3
readonly CHECK_DIR_INTERVAL=300
ETCD_CONNECT_FAILED_TIMES=6 # 5*6=30 min
export YR_BARE_MENTAL=1
# limit the fd consumption
export GOMAXPROCS=16
# limit litebus std log
export LOG_LEVEL_ENABLE="INFO"

[[ ! -f "${BASE_DIR}/utils.sh" ]] && echo "${BASE_DIR}/utils.sh is not exist" && exit 1
. ${BASE_DIR}/utils.sh
[[ ! -f "${BASE_DIR}/port_utils.sh" ]] && echo "${BASE_DIR}/port_utils.sh does not exist" && exit 1
. ${BASE_DIR}/port_utils.sh
[[ ! -f "${BASE_DIR}/health_check.sh" ]] && echo "${BASE_DIR}/health_check.sh is not exist" && exit 1
. ${BASE_DIR}/health_check.sh
[[ ! -f "${FUNCTION_SYSTEM_PATH}/deploy/install.sh" ]] && echo "${FUNCTION_SYSTEM_PATH}/deploy/install.sh is not exist" && exit 1
. ${FUNCTION_SYSTEM_PATH}/deploy/install.sh
[[ ! -f "${DATA_SYSTEM_PATH}/deploy/install.sh" ]] && echo "${DATA_SYSTEM_PATH}/deploy/install.sh is not exist" && exit 1
. ${DATA_SYSTEM_PATH}/deploy/install.sh
[[ ! -f "${THIRD_PARTY_PATH}/install.sh" ]] && echo "${THIRD_PARTY_PATH}/install.sh is not exist" && exit 1
. ${THIRD_PARTY_PATH}/install.sh
[[ ! -f "${THIRD_PARTY_PATH}/utils.sh" ]] && echo "${THIRD_PARTY_PATH}/utils.sh is not exist" && exit 1
. ${THIRD_PARTY_PATH}/utils.sh

command -v curl >/dev/null
[[ $? != 0 ]] && echo "not found curl for health check. please install it." && exit 1

readonly FUNCTION_PROXY_PORT_INFO="function_proxy_info_tmp.txt"
DEPLOY_DS_WORKER=true
DEPLOY_FUNCTION_PROXY=true
MASTER_INFO_STRING=""

CURRENT_PID=$$
NODE_ID="$(trim_hostname)-${CURRENT_PID}"
UNIQUE_NODE_ID=""

declare -A pid_table
COMPONENT_CHECKLIST=()
DEPLOY_STEP="READY"

# Permission control: remove others' permissions of directories and files
umask -p 027


function print_info() {
  if [ "X${ETCD_ADDR_LIST_FOR_MASTER_INFO}" != "X" ]; then
    MASTER_INFO_STRING="${ETCD_ADDR_LIST_FOR_MASTER_INFO},${MASTER_INFO_STRING}"
  fi
  if [ "X${ENABLE_MASTER}" != "Xtrue" ]; then
    MASTER_INFO_STRING="${MASTER_INFO_STRING}master_ip:${FUNCTION_MASTER_IP},"
    MASTER_INFO_STRING="${MASTER_INFO_STRING}etcd_port:${ETCD_PORT},etcd_peer_port:${ETCD_PEER_PORT},global_scheduler_port:${GLOBAL_SCHEDULER_PORT},ds_master_port:${DS_MASTER_PORT},"
  fi
  MASTER_INFO_STRING="${MASTER_INFO_STRING}bus-proxy:${data_port_table['function_proxy_port']},"
  MASTER_INFO_STRING="${MASTER_INFO_STRING}bus:${data_port_table['function_proxy_grpc_port']},"
  MASTER_INFO_STRING="${MASTER_INFO_STRING}ds-worker:${data_port_table['ds_worker_port']},"
  if [ "X${ENABLE_DASHBOARD}" == "Xtrue" ] || [ "X${ENABLE_DASHBOARD}" == "XTRUE" ]; then
    MASTER_INFO_STRING="${MASTER_INFO_STRING}dashboard_port:${data_port_table['dashboard_port']},"
  fi
  if [ "X${ENABLE_FAAS_FRONTEND}" == "Xtrue" ] || [ "X${ENABLE_FAAS_FRONTEND}" == "XTRUE" ]; then
    MASTER_INFO_STRING="${MASTER_INFO_STRING}frontend_port:${data_port_table['faas_frontend_http_port']},"
  fi
  if [ "X${ENABLE_MASTER}" = "Xtrue" ]; then
    log_info "${MASTER_INFO_STRING}"
  fi
  [ "X${MASTER_INFO_OUT_FILE}" != "X" ] && echo "${MASTER_INFO_STRING}" > "${MASTER_INFO_OUT_FILE}"
  printf "OpenYuanrong Deployment Info:\n"
  if [ "X${ENABLE_MASTER}" = "Xtrue" ] ; then
    printf "%30s %10s\n" "CONTROL PLANE INSTALL_DIR:" "${INSTALL_DIR_PARENT}"
  fi
  printf "%30s %10s\n" "DATA PLANE INSTALL_DIR:" "${DATA_PLANE_INSTALL_DIR}"
  printf "%30s %10s\n" "ROOT_LOG_DIR:" "${LOG_ROOT}"
  printf "%30s %10s\n" "FS_LOG_DIR:" "${FS_LOG_PATH}"
  printf "%30s %10s\n" "DS_LOG_DIR:" "${DS_LOG_PATH}"
  printf "%30s %10s\n" "RUNTIME_LOG_DIR:" "${RUNTIME_LOG_PATH}"
  printf "%30s %10s\n" "FS_LOG_LEVEL:" "${FS_LOG_LEVEL}"
  printf "%30s %10s\n" "DS_LOG_LEVEL:" "${DS_LOG_LEVEL}"
  printf "%30s %10s\n" "RUNTIME_LOG_LEVEL:" "${RUNTIME_LOG_LEVEL}"
  printf "\n"
  printf "%30s %10s\n" "HOST_IP:" "${IP_ADDRESS}"
  printf "%30s %10s\n" "NODE_ID:" "${NODE_ID}"
  printf "%30s %10s\n" "FUNCTION_AGENT_ALIAS:" "${FUNCTION_AGENT_ALIAS}"
  if [ "X${ENABLE_MULTI_MASTER}" = "Xtrue" ]; then
    printf "%30s %10s\n" "ETCD_CLUSTER_ADDRESS:" "${ETCD_CLUSTER_ADDRESS}"
  else
    printf "%30s %10s\n" "FUNCTION_MASTER_IP:" "${FUNCTION_MASTER_IP}"
  fi
  printf "\n"
  if [ "X${DRIVER_GATEWAY_ENABLE}" = "XTRUE" ] || [ "X${DRIVER_GATEWAY_ENABLE}" = "Xtrue" ] ; then
    printf "%30s %10s\n" "FUNCTION_PROXY_GRPC_PORT:" "${data_port_table['function_proxy_grpc_port']}"
  fi

  printf "%30s %10s\n" "FUNCTION_PROXY_PORT:" "${data_port_table['function_proxy_port']}"
  if [ -n "${pid_table["faas_frontend"]}" ]; then
    printf "%30s %10s\n" "FAAS_FRONTEND_HTTP_PORT:" "${data_port_table['faas_frontend_http_port']}"
    printf "%30s %10s\n" "FAAS_FRONTEND_GRPC_PORT:" "${data_port_table['faas_frontend_grpc_port']}"
  fi
  if [ -n "${pid_table["dashboard"]}" ]; then
      printf "%30s %10s\n" "DASHBOARD_PORT:" "${data_port_table['dashboard_port']}"
  fi
  printf "%30s %10s\n" "FUNCTION_AGENT_PORT:" "${data_port_table['function_agent_port']}"
  if [ "X${MERGE_PROCESS_ENABLE}" = "Xtrue" ] || [ "X${MERGE_PROCESS_ENABLE}" = "XTRUE" ]; then
    printf "%30s %10s\n" "RUNTIME_MGR_PORT:" "disabled"
  fi
  printf "%30s %10s\n" "RUNTIME_INIT_PORT:" "${data_port_table['runtime_init_port']}"
  printf "\n"
  printf "%30s %10s\n" "OVERALL CPU:" "${CPU_ALL}"
  printf "%30s %10s\n" "OVERALL MEM:" "${MEM_ALL}"
  printf "%30s %10s\n" "CPU FOR FUNCTIONS:" "${CPU4COMP}"
  printf "%30s %10s\n" "MEM FOR FUNCTIONS:" "${MEM4COMP}"
  printf "%30s %10s\n" "SHARED MEM FOR DS:" "${MEM4DATA}"
  printf "\n"
  printf "%30s %10s\n" "DS_WORKER_PORT:" "${data_port_table['ds_worker_port']}"
  if [ "X${ENABLE_DISTRIBUTED_MASTER}" = "Xfalse" ]; then
    printf "%30s %10s\n" "DS_MASTER_ADDRESS:" "${DS_MASTER_IP}:${control_port_table['ds_master_port']}"
  fi
  printf "\n"
  printf "%30s %10s\n" "SERVICES_PATH:" "${SERVICES_PATH}"
  printf "\n\n"
}

function write_master_pid_port_temp_file() {
  local temp_file="${INSTALL_DIR_PARENT}/master_pid_port.txt"
  echo -e "$1" >${temp_file}
}

function start_proxy_for_etcd() {
  log_info "Deploying etcd proxy..."
  local i
  for ((i = 1; i <= ${ETCD_PROXY_NUMS}; i++)); do
    local j ret_code
    for ((j = 1; j <= $CONTROL_PLANE_RETRY_TIMES; j++)); do
      update_control_plane_port "etcd_proxy_port_${i}"
      [ "X${PORT_POLICY}" = "XFIX" ] && control_port_table["etcd_proxy_port_${i}"]=${ETCD_PROXY_PORT}
      install_third_party "etcd_proxy" "$i" "${control_port_table["etcd_proxy_port_${i}"]}"
      ret_code=$?
      if [ ${ret_code} -eq 0 ]; then
        break
      fi
      log_warning "Deploy etcd_proxy_${i} failed, try times: ${j}."
    done
    if [ ${ret_code} -eq 0 ]; then
      check_and_set_component_checklist "etcd_proxy_${i}" ${ETCD_PROXY_PID}
    else
      log_error "Deploy etcd_proxy_${i} failed."
      return 98
    fi
  done
  return 0
}

function remove_unique_info_from_etcd() {
  log_info "removing info text"
  if [ "X${DS_WORKER_UNIQUE_ENABLE}" = "Xtrue" ] && [ "X${DEPLOY_DS_WORKER}" = "Xtrue" ]; then
    log_info "start to remove ds_worker info from etcd"
    del_from_etcd "${UNIQUE_KEY_PREFIX}/unique/ds_worker/${IP_ADDRESS}" "${ETCD_ENDPOINTS_ADDR}"
  fi

  if [ "X${FUNCTION_PROXY_UNIQUE_ENABLE}" = "Xtrue" ] && [ "X${DEPLOY_FUNCTION_PROXY}" = "Xtrue" ]; then
    log_info "start to remove function_proxy info from etcd"
    del_from_etcd "${UNIQUE_KEY_PREFIX}/unique/function_proxy/${IP_ADDRESS}" "${ETCD_ENDPOINTS_ADDR}"
  fi
}

function signal_handle() {
  log_info "handle signal"
  remove_unique_info_from_etcd
  exit_all_processes
}

function wait_ds_worker_up() {
  local retry_time=150
  local i=0 j=0
  local ds_worker_pid ds_worker_info
  while [ $i -lt $retry_time ]; do
    j=$((RANDOM % 10 + 1))
    i=$((i + j))
    sleep ${j} # sleeps randomly for 1-10 seconds
    ds_worker_info=$(get_from_etcd "${UNIQUE_KEY_PREFIX}/unique/ds_worker/${IP_ADDRESS}" "${ETCD_ENDPOINTS_ADDR}")
    local ret=$?
    log_info "get single ds_worker port info: \"${ds_worker_info}\""
    if [ ${ret} -ne 0 ]; then
      log_warning "get ds_worker port from etcd failed, return code: ${ret}, retry later"
      continue
    fi
    DS_WORKER_PORT=$(echo "${ds_worker_info}" | awk -F ',' '{print $1}')
    if [[ ! "${DS_WORKER_PORT}" =~ ^[1-9][0-9]*$ ]]; then
      continue
    fi
    ds_worker_pid=$(ps -ef | grep -E "worker_address=[^ ]+:$DS_WORKER_PORT" | grep "data_system" | grep -w "worker" | grep -v grep | awk -F " " '{print $2}')
    if [ -z "$ds_worker_pid" ]; then
      continue
    fi
    check_and_set_component_checklist "ds_worker" $ds_worker_pid
    DS_WORKER_HEALTH_CHECK_PATH=$(ps -ef | grep -E "worker_address=[^ ]+:$DS_WORKER_PORT" | grep "data_system" | grep -w "worker" | grep -v grep | sed -rn 's/.*-health_check_path=([^ ]*)[ ].*/\1/p')
    if [ -z "$DS_WORKER_HEALTH_CHECK_PATH" ]; then
      continue
    fi
    if health_check "ds_worker" $ds_worker_pid; then
      data_port_table["ds_worker_port"]=${DS_WORKER_PORT}
      UNIQUE_NODE_ID="$(echo "${ds_worker_info}" | awk -F ',' '{print $2}')"
      return 0
    fi
  done
  return 98
}

function wait_function_proxy_up() {
  local retry_time=150
  local i=0 j=0
  local function_proxy_info function_proxy_pid
  while [ $i -lt $retry_time ]; do
    j=$((RANDOM % 10 + 1))
    i=$((i + j))
    sleep ${j} # sleeps randomly for 1-10 seconds
    function_proxy_info=$(get_from_etcd "${UNIQUE_KEY_PREFIX}/unique/function_proxy/${IP_ADDRESS}" "${ETCD_ENDPOINTS_ADDR}")
    local ret=$?
    log_info "get single function_proxy port info: \"${function_proxy_info}\""
    if [ ${ret} -ne 0 ]; then
      log_warning "get function_proxy port from etcd failed, return code: ${ret}, retry later"
      continue
    fi
    FUNCTION_PROXY_PORT=$(echo "${function_proxy_info}" | awk -F ',' '{print $1}')
    FUNCTION_PROXY_GRPC_PORT=$(echo "${function_proxy_info}" | awk -F ',' '{print $2}')
    UNIQUE_NODE_ID=$(echo "${function_proxy_info}" | awk -F ',' '{print $3}')
    if [[ ! "${FUNCTION_PROXY_PORT}" =~ ^[1-9][0-9]*$ ]] || [[ ! "${FUNCTION_PROXY_GRPC_PORT}" =~ ^[1-9][0-9]*$ ]]; then
      continue
    fi
    function_proxy_pid=$(ps -ef | grep -E "\--address=[^ ]+:${FUNCTION_PROXY_PORT}" | grep "grpc_listen_port=${FUNCTION_PROXY_GRPC_PORT}" | grep "function_proxy" | grep -v grep | awk -F " " '{print $2}')
    if [ -z "$function_proxy_pid" ]; then
      continue
    fi
    data_port_table["function_proxy_port"]=${FUNCTION_PROXY_PORT}
    data_port_table["function_proxy_grpc_port"]=${FUNCTION_PROXY_GRPC_PORT}
    return 0
  done
  return 98
}

function wait_etcd_up() {
 # used for kube-ray, start master, agent at the same time, agent need to wait etcd ready
 if [ "X$YR_WAIT_ETCD_READY" != "Xtrue" ]; then
   return 0
 fi
 local retry_time=45
 local i=0 j=0
 while [ $i -lt $retry_time ]; do
    j=$((RANDOM % 5 + 1))
    i=$((i + j))
    check_etcd_ready "${ETCD_ENDPOINTS_ADDR}"
    local ret=$?
    if [ ${ret} -ne 0 ]; then
      log_warning "etcd is not ready, return code: ${ret}, retry later"
      sleep ${j}
      continue
    fi
    return 0
  done
  return 98
}

function reserve_cpu_for_ds_worker() {
    local cpu4comp_new=$((CPU_ALL - CPU_RESERVED_FOR_DS_WORKER * 1000))
    if [ "$cpu4comp_new" -gt 0 ]; then
        CPU4COMP=$cpu4comp_new
    else
      log_warning "(cpu_num - cpu_reserved*1000) less than or equal 0, ignore cpu_reserved"
      CPU4COMP=$CPU_ALL
    fi
}

function start_or_wait_ds_worker() {
  if [ "X$DS_WORKER_UNIQUE_ENABLE" = "Xfalse" ]; then
    start_ds_worker
    reserve_cpu_for_ds_worker
  else
    local key="${UNIQUE_KEY_PREFIX}/unique/ds_worker/${IP_ADDRESS}"
    local val="-1,${NODE_ID}" # DS_WORKER_PORT is invalid at this moment, it will be updated before start
    local ret_msg=""
    ret_msg=$(setnx_or_get "${key}" "${val}" "${ETCD_ENDPOINTS_ADDR}" 150)
    local ret_code=$?
    if [ ${ret_code} != 0 ]; then  # etcdctl failed
      log_error "setnx_or_get etcd failed, error code: ${ret_code}, please check it."
      exit 1
    fi

    if [ -z "${ret_msg}" ]; then
      UNIQUE_NODE_ID="${NODE_ID}"
      start_ds_worker
      reserve_cpu_for_ds_worker
    else
      DEPLOY_DS_WORKER=false
      if ! wait_ds_worker_up; then
        log_error "wait ds_worker up timeout or old ds_worker file lock isn't deleted, port=$DS_WORKER_PORT, pid=${pid_table["ds_worker"]}"
        exit 1
      else
        log_info "single data system worker port=$DS_WORKER_PORT, pid=${pid_table["ds_worker"]}"
      fi
    fi
  fi
}

function start_or_wait_function_proxy() {
  if [ "X${FUNCTION_PROXY_UNIQUE_ENABLE}" = "Xtrue" ]; then
    FUNCTION_PROXY_UNREGISTER_WHILE_STOP=false
  fi
  if [ "X${FUNCTION_PROXY_UNIQUE_ENABLE}" = "Xfalse" ] || ([ "X${DS_WORKER_UNIQUE_ENABLE}" = "Xtrue" ] && [ "X${DEPLOY_DS_WORKER}" = "Xtrue" ]); then
    start_function_proxy
    return
  elif [ "X${DS_WORKER_UNIQUE_ENABLE}" = "Xfalse" ]; then
    local key="${UNIQUE_KEY_PREFIX}/unique/function_proxy/${IP_ADDRESS}"
    local val="-1,-1,${NODE_ID}" # FUNCTION_PROXY_PORT is invalid at this moment, it will be updated before start
    local ret_msg=""
    ret_msg=$(setnx_or_get "${key}" "${val}" "${ETCD_ENDPOINTS_ADDR}" 150)
    local ret_code=$?
    if [ ${ret_code} != 0 ]; then  # etcdctl failed
      log_error "setnx_or_get etcd failed, error code: ${ret_code}, please check it."
      exit 1
    fi
    if [ -z "${ret_msg}" ]; then
      UNIQUE_NODE_ID="${NODE_ID}"
      start_function_proxy
      return
    fi
  fi
  DEPLOY_FUNCTION_PROXY=false
  if ! wait_function_proxy_up; then
    log_error "wait function_proxy up timeout or old function_proxy file lock isn't deleted, port=$FUNCTION_PROXY_PORT"
    exit 1
  else
    log_info "single function proxy port=$FUNCTION_PROXY_PORT"
  fi
  return
}

function start_etcd() {
  if [ "X${ENABLE_MULTI_MASTER}" != "Xtrue" ]; then
    update_control_plane_port "etcd_port etcd_peer_port"
    ETCD_PORT=${control_port_table["etcd_port"]}
    ETCD_PEER_PORT=${control_port_table["etcd_peer_port"]}
    ETCD_CLUSTER_ADDRESS="${ETCD_IP}:${ETCD_PORT}"
    ETCD_ENDPOINTS_ADDR="${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PORT}"
  fi
  install_third_party "etcd"
  ret_code=$?
  if [ ${ret_code} -eq 0 ]; then
    check_and_set_component_checklist "etcd" $ETCD_PID
  fi
  return ${ret_code}
}

function start_etcd_proxy() {
  local num=$1
  install_third_party "etcd_proxy" "$num" ${control_port_table["etcd_proxy_port_${num}"]}
  if [ ${ret_code} -eq 0 ]; then
    check_and_set_component_checklist "etcd_proxy_${num}" $ETCD_PROXY_PID
  fi
  return ${ret_code}
}

function start_function_master() {
  update_control_plane_port "global_scheduler_port"
  GLOBAL_SCHEDULER_PORT=${control_port_table["global_scheduler_port"]}
  install_function_system "function_master"
  ret_code=$?
  if [ ${ret_code} -eq 0 ]; then
    check_and_set_component_checklist "function_master" $FUNCTION_MASTER_PID
  fi
  return ${ret_code}
}

function start_ds_master() {
  local data_system_install_dir=${INSTALL_DIR_PARENT}/data_system
  rm -rf $data_system_install_dir/master/health
  update_control_plane_port "ds_master_port"
  DS_MASTER_PORT=${control_port_table["ds_master_port"]}
  install_data_system "ds_master"
  ret_code=$?
  if [ ${ret_code} -eq 0 ]; then
    check_and_set_component_checklist "ds_master" $DS_MASTER_PID
  fi
  return ${ret_code}
}

function restart_control_plane_component() {
  local module=$1
  local retry_time=$2
  local i
  for ((i = 1; i <= ${retry_time}; i++)); do
    start_${module}
    ret_code=$?
    if [ ${ret_code} -eq 0 ]; then
      break
    fi
    log_warning "Deploy ${module} failed, try times: ${i}."
  done
  if [ -z "${pid_table["${module}"]}" ]; then
    log_error "Deploy ${module} failed."
    write_master_pid_port_temp_file "error: Deploy ${module} failed, ret_code: ${ret_code}"
    return ${ret_code}
  fi
}

function start_control_plane() {
  local ret_code
  if [ "X${ETCD_MODE}" != "Xoutter" ]; then
    # etcd
    log_info "Deploying etcd..."
    if ! restart_control_plane_component "etcd" $CONTROL_PLANE_RETRY_TIMES; then
      return 98
    fi
    # etcd_proxy
    if [ "X${ETCD_PROXY_ENABLE}" = "XTRUE" ] && [ "$ETCD_PROXY_NUMS" -gt 0 ]; then
      start_proxy_for_etcd
      ret_code=$?
      if [ ${ret_code} -ne 0 ]; then
        write_master_pid_port_temp_file "error: Deploy etcd proxy failed.\n${ret_code}"
        return ${ret_code}
      fi
    fi
  fi

  # ds_master
  if [ "X${ENABLE_DISTRIBUTED_MASTER}" = "Xfalse" ]; then
    log_info "Deploying ds_master..."
    if ! restart_control_plane_component "ds_master" $CONTROL_PLANE_RETRY_TIMES; then
        return 98
    fi
  fi
  # function master
  log_info "Deploying function master..."
  if ! restart_control_plane_component "function_master" $CONTROL_PLANE_RETRY_TIMES; then
    return 98
  fi
  # iam_server
  if [ "X${ENABLE_IAM_SERVER}" = "Xtrue" ] || [ "X${ENABLE_IAM_SERVER}" = "XTRUE" ]; then
    log_info "Deploying iam_server..."
    if ! restart_control_plane_component "iam_server" $CONTROL_PLANE_RETRY_TIMES; then
      return 98
    fi
  fi
  return 0
}

function dump_master_info() {
  log_info "all port for control plane:${!control_port_table[*]} ${control_port_table[*]}"
  local master_pid_string=""
  for key in ${!pid_table[*]}; do
    master_pid_string="${master_pid_string}${key}:${pid_table["${key}"]},"
  done
  log_info "all pid for control plane:${master_pid_string}"
  MASTER_INFO_STRING="local_ip:${LOCAL_IP},host_ip:${IP_ADDRESS},master_ip:${IP_ADDRESS},"
  if [ "X${ETCD_MODE}" != "Xoutter" ]; then
    MASTER_INFO_STRING="${MASTER_INFO_STRING}etcd_ip:${ETCD_IP},"
  fi
  for key in ${!control_port_table[*]}; do
    MASTER_INFO_STRING="${MASTER_INFO_STRING}${key}:${control_port_table["${key}"]},"
  done
  write_master_pid_port_temp_file "${master_pid_string}\n${MASTER_INFO_STRING}"
}

function dump_agent_info() {
  # right now the master ip is not really the master ip but the agent ip
  if [ "$MASTER_INFO_STRING"X = ""X ]; then
    MASTER_INFO_STRING="local_ip:${LOCAL_IP},host_ip:${IP_ADDRESS},"
  fi
}

function start_function_proxy() {
  update_data_plane_port "function_proxy_port function_proxy_grpc_port function_proxy_exec_grpc_port"
  FUNCTION_PROXY_PORT=${data_port_table["function_proxy_port"]}
  FUNCTION_PROXY_GRPC_PORT=${data_port_table["function_proxy_grpc_port"]}
  FUNCTION_PROXY_EXEC_GRPC_PORT=${data_port_table["function_proxy_exec_grpc_port"]}
  install_function_system "function_proxy"
  check_and_set_component_checklist "function_proxy" $FUNCTION_PROXY_PID
}

function start_function_agent() {
  if [ "X${FUNCTION_PROXY_MERGE_PROCESS_ENABLE^^}" == "XTRUE" ]; then
    log_info "function_proxy_merge_process_enable is true, function_agent will be started by function_proxy, skip starting function_agent separately"
    return 0
  fi
  update_data_plane_port "function_agent_port runtime_init_port"
  FUNCTION_AGENT_PORT=${data_port_table["function_agent_port"]}
  RUNTIME_INIT_PORT=${data_port_table["runtime_init_port"]}
  install_function_system "function_agent"
  check_and_set_component_checklist "function_agent" $FUNCTION_AGENT_PID
}

function start_faas_frontend() {
  if [ "X${ENABLE_FAAS_FRONTEND}" != "Xtrue" ] && [ "X${ENABLE_FAAS_FRONTEND}" != "XTRUE" ]; then
    return 0
  fi
  if [ -z "${META_SERVICE_ADDRESS}" ]; then
    META_SERVICE_ADDRESS="${IP_ADDRESS}:${META_SERVICE_PORT}"
  fi
  export META_SERVICE_ADDRESS
  update_data_plane_port "faas_frontend_http_port faas_frontend_grpc_port"
  FAAS_FRONTEND_HTTP_PORT=${data_port_table["faas_frontend_http_port"]}
  FAAS_FRONTEND_GRPC_PORT=${data_port_table["faas_frontend_grpc_port"]}
  install_function_system "faas_frontend"
  check_and_set_component_checklist "faas_frontend" $FAAS_FRONTEND_PID
}

function start_function_scheduler() {
  if [ "X${ENABLE_FUNCTION_SCHEDULER}" != "Xtrue" ] && [ "X${ENABLE_FUNCTION_SCHEDULER}" != "XTRUE" ]; then
    return 0
  fi
  install_function_system "function_scheduler"
  check_and_set_component_checklist "function_scheduler" $SCHEDULER_PID
}

function start_dashboard() {
  if [ "X${ENABLE_DASHBOARD}" != "Xtrue" ] && [ "X${ENABLE_DASHBOARD}" != "XTRUE" ]; then
    return 0
  fi
  update_data_plane_port "dashboard_port dashboard_grpc_port"
  DASHBOARD_PORT=${data_port_table["dashboard_port"]}
  DASHBOARD_GRPC_PORT=${data_port_table["dashboard_grpc_port"]}
  install_function_system "dashboard"
  ret_code=$?
  if [ ${ret_code} -eq 0 ]; then
    check_and_set_component_checklist "dashboard" $DASHBOARD_PID
  fi
  return ${ret_code}
}

function start_collector() {
  if [ "X${ENABLE_COLLECTOR}" != "Xtrue" ] && [ "X${ENABLE_COLLECTOR}" != "XTRUE" ]; then
    return 0
  fi
  update_data_plane_port "collector_port"
  COLLECTOR_PORT=${data_port_table["collector_port"]}
  install_function_system "collector"
  ret_code=$?
  if [ ${ret_code} -eq 0 ]; then
    check_and_set_component_checklist "collector" $COLLECTOR_PID
  fi
  return ${ret_code}
}

function start_meta_service() {
  if [ "X${ENABLE_META_SERVICE}" != "Xtrue" ] && [ "X${ENABLE_META_SERVICE}" != "XTRUE" ]; then
    return 0
  fi
  update_data_plane_port "meta_service_port"
  META_SERVICE_PORT=${control_port_table["meta_service_port"]}
  install_function_system "meta_service"
  ret_code=$?
  if [ ${ret_code} -eq 0 ]; then
    check_and_set_component_checklist "meta_service" $META_SERVICE_PID
  fi
  return ${ret_code}
}

function start_iam_server() {
  if [ "X${ENABLE_IAM_SERVER}" != "Xtrue" ] && [ "X${ENABLE_IAM_SERVER}" != "XTRUE" ]; then
    return 0
  fi
  update_control_plane_port "iam_server_port"
  IAM_SERVER_PORT=${control_port_table["iam_server_port"]}
  install_function_system "iam_server"
  ret_code=$?
  if [ ${ret_code} -eq 0 ]; then
    check_and_set_component_checklist "iam_server" $IAM_SERVER_PID
  fi
  IAM_SERVER_ADDRESS="${IP_ADDRESS}:${IAM_SERVER_PORT}"
  export IAM_SERVER_ADDRESS
  return ${ret_code}
}

function start_ds_worker() {
  update_data_plane_port "ds_worker_port"
  DS_WORKER_PORT=${data_port_table["ds_worker_port"]}
  install_data_system "ds_worker"
  check_and_set_component_checklist "ds_worker" $DS_WORKER_PID
}

function need_health_check() {
    local component="$1"
    if [ "${component}X" == "ds_workerX" ] && [ "${DEPLOY_DS_WORKER}X" == "falseX" ]; then
        return 1
    fi
    if [ "${component}X" == "function_proxyX" ] && [ "${DEPLOY_FUNCTION_PROXY}X" == "falseX" ]; then
        return 1
    fi
    return 0
}

function set_unique_info() {
  if [ "X${DS_WORKER_UNIQUE_ENABLE}" = "Xtrue" ] && [ "X${DEPLOY_DS_WORKER}" = "Xtrue" ]; then
    set_to_etcd "${UNIQUE_KEY_PREFIX}/unique/ds_worker/${IP_ADDRESS}" "${DS_WORKER_PORT},${NODE_ID}" "${ETCD_ENDPOINTS_ADDR}"
  fi
  if [ "X${FUNCTION_PROXY_UNIQUE_ENABLE}" = "Xtrue" ] && [ "X${DEPLOY_FUNCTION_PROXY}" = "Xtrue" ]; then
    set_to_etcd "${UNIQUE_KEY_PREFIX}/unique/function_proxy/${IP_ADDRESS}" "${FUNCTION_PROXY_PORT},${FUNCTION_PROXY_GRPC_PORT},${NODE_ID}" "${ETCD_ENDPOINTS_ADDR}"
  fi
}

function is_child_process() {
    local pid="$1"
    ppid=$(ps -o ppid= -p $pid | tr -d ' ')
    if [ "${ppid}X" == "${CURRENT_PID}X" ];then
      return 0
    fi
    return 1
}

function monitor_control_process_exit() {
  local component_name=$1
  local pid=${pid_table[$component_name]}
  if ! kill -0 ${pid}; then
    # if etcd exit during deploying yr-agent, deploy failed
    if [ X"${DEPLOY_STEP}" = XAGENT ]; then
      exit 98
    else
      # if etcd exit during running, ignore it
      log_warning "${component} has exited, pid: ${pid}"
      unset pid_table[${component_name}]
    fi
  fi
}

function terminate_process() {
  local pid=$1
  if [ -z "$pid" ]; then
    return
  fi
  let kill_times=5
  for ((i = 0; i <= kill_times; i++));
  do
    log_info "terminating $pid"
    kill -15 "${pid}" &>/dev/null
    if ! kill -0 "${pid}" &>/dev/null; then
      break
    fi
    sleep 1
  done
  if kill -0 "${pid}" &>/dev/null; then
      kill -9 "$pid" &>/dev/null
  fi
  log_info "kill $pid successfully"
}

function restart_module() {
  local module=$1
  local module_name_prefix=$module
  local extra_param=""
  if [[ $module == etcd_proxy_* ]]; then
    extra_param=${module:11}
    module_name_prefix="etcd_proxy"
  fi
  local pid=${pid_table[${module}]}
  terminate_process $pid
  local i
  for ((i=0; i<${RETRY_TIMES}; i++)); do
    start_${module_name_prefix} ${extra_param}
    ret_code=$?
    if [ ${ret_code} -eq 0 ]; then
       log_info "$module restarted, new pid=${pid_table["${module}"]}"
       return 0
    else
      log_error "$module restart failed, return value ${ret_code}, retry_time:${i}"
      sleep 1
    fi
  done
  return 1
}

function restart_agent_runtime_accessor() {
  [ -n "${pid_table["function_agent"]}" ] && restart_module "function_agent" && sleep 1
}

function restart_component() {
  case "$1" in
  function_proxy)
    restart_module "$1"
    if [ "X${PORT_POLICY}" != "XFIX" ]; then
      restart_agent_runtime_accessor
    fi
    ;;
  function_agent)
    restart_module "$1"
    ;;
  ds_worker)
     restart_module "$1"
     if [ "X${PORT_POLICY}" != "XFIX" ]; then
       [ -n "${pid_table["function_proxy"]}" ] && restart_module "function_proxy" && sleep 1
       restart_agent_runtime_accessor
     fi
    ;;
  function_master|ds_master|collector|faas_frontend|dashboard|function_scheduler|meta_service|iam_server)
    restart_module "$1"
    ;;
  etcd)
    if [ "X${ENABLE_MULTI_MASTER}" = "Xtrue" ]; then
      restart_module "$1"
      return 0
    fi
    # restart control plane if etcd is exited
    if [ "X${ETCD_PROXY_ENABLE}" = "XTRUE" ] && [ "$ETCD_PROXY_NUMS" -gt 0 ]; then
      for ((i = 1; i <= ${ETCD_PROXY_NUMS}; i++)); do
        terminate_process ${pid_table["etcd_proxy_${i}"]}
      done
    fi
    terminate_process ${pid_table["ds_master"]}
    terminate_process ${pid_table["function_master"]}
    terminate_process ${pid_table["collector"]}
    terminate_process ${pid_table["dashboard"]}
    terminate_process ${pid_table["faas_frontend"]}
    terminate_process ${pid_table["function_scheduler"]}
    terminate_process ${pid_table["meta_service"]}
    terminate_process ${pid_table["iam_server"]}
    start_control_plane
    ;;
  *)
    if [[ "$1" == etcd_proxy_* ]]; then
      restart_module "$1"
    else
      log_error "Unknown component: $1"
      return 1
    fi
    ;;
  esac
}

function fetch_component_std_logs() {
  local component=$1
  local log_filepath="${FS_LOG_PATH}/${NODE_ID}-${component}${STD_LOG_SUFFIX}"
  if [ -f "${log_filepath}" ]; then
    log_warning "${component} logs at ${log_filepath}"
    log_warning "=========================================================================="
    tail -50 "${log_filepath}"
    log_warning "=========================================================================="
  fi
}

function listen_component() {
  local is_check_ready=$1
  local not_ready=0
  local sleep_interval=3
  [ "${is_check_ready}" = "true" ] && sleep_interval=10
  for component in "${COMPONENT_CHECKLIST[@]}"; do
    # component maybe unset from pid_table
    if [ -z ${pid_table[$component]} ]; then
      continue
    fi
    if ! need_health_check ${component}; then
      continue
    fi
    if [[ "$component" = etcd* ]] && [ X"${DEPLOY_STEP}" != XMASTER ] && [ "X${ENABLE_MULTI_MASTER}" != "Xtrue" ]; then
      monitor_control_process_exit $component
      continue
    fi
    if ! health_check ${component} ${pid_table[$component]}; then
      if  [ "${is_check_ready}" = "false" ]; then
        monitor "${LOG_ROOT}" "${NODE_ID}"
      fi
      log_error "will terminate $component because health check failed. pid: ${pid_table[$component]}"
      terminate_process ${pid_table[$component]}
      log_error "**WARN** health check err: $component exited, pid: ${pid_table[$component]}"
      fetch_component_std_logs "$component"
      wait ${pid_table[$component]}  # get process exit code
      exit_code=$?
      # if process is panic, this code will also be 2
      if [ "${is_check_ready}" = "true" ] && [ ${exit_code} -eq ${EXIT_MISUSE} ]; then  # command misuse, most likely wrong arguments
        log_error "**ERROR** $component exit code ${EXIT_MISUSE}, please check startup arguments"
        exit ${EXIT_MISUSE}
      fi
      log_info "restarting $component..."
      restart_component "$component"
      not_ready=1
      sleep $sleep_interval
      # if not check ready, continue to check next component
      if [ "${is_check_ready}" = "true" ]; then
        return $not_ready
      fi
    else
      [ "${is_check_ready}" = "true" ] && log_info "health check passed: $component, pid: ${pid_table[$component]}"
    fi
  done
  return $not_ready
}

function wait_for_ready_on_checklist() {
  local not_ready=0
  while :; do
    listen_component "true"
    not_ready=$?
    [ ${not_ready} -eq 0 ] && break
    sleep 1
  done
}

function check_ds_worker_alive()
{
  local ds_worker_info=""
  ds_worker_info=$(get_from_etcd "${UNIQUE_KEY_PREFIX}/unique/ds_worker/${IP_ADDRESS}" "${ETCD_ENDPOINTS_ADDR}")
  local ret=$?
  if [ ${ret} -ne 0 ]; then
    log_warning "connect to etcd failed, ret_code: ${ret}, remain times: ${ETCD_CONNECT_FAILED_TIMES}"
    ETCD_CONNECT_FAILED_TIMES=$((ETCD_CONNECT_FAILED_TIMES - 1))
    if [ ${ETCD_CONNECT_FAILED_TIMES} -eq 0 ]; then
      log_error "check ds_worker alive from etcd failed"
      return 1
    fi
    return 0
  fi
  ETCD_CONNECT_FAILED_TIMES=6
  local ds_worker_port=$(echo "${ds_worker_info}" | awk -F ',' '{print $1}')
  local cur_ds_worker_node_id="$(echo "${ds_worker_info}" | awk -F ',' '{print $2}')"
  if [[ ! "${ds_worker_port}" =~ ^[1-9][0-9]*$ ]] || [ "X${cur_ds_worker_node_id}" != "X${UNIQUE_NODE_ID}" ]; then
    log_error "the single ds_worker(port: ${DS_WORKER_PORT}, node_id: ${UNIQUE_NODE_ID}) had exited"
    return 1
  fi
  return 0
}

function check_unique_lock() {
  if [ "X${DS_WORKER_UNIQUE_ENABLE}" = "Xtrue" ] && [ "X${DEPLOY_DS_WORKER}" = "Xtrue" ]; then
    local ds_worker_info=""
    ds_worker_info=$(get_from_etcd "${UNIQUE_KEY_PREFIX}/unique/ds_worker/${IP_ADDRESS}" "${ETCD_ENDPOINTS_ADDR}")
    local ret=$?
    if [ ${ret} -eq 0 ]; then
      local ds_worker_port=$(echo "${ds_worker_info}" | awk -F ',' '{print $1}')
      # if unique info is lost, reset it to etcd
      if [ "X${ds_worker_port}" = "X" ]; then
        set_to_etcd "${UNIQUE_KEY_PREFIX}/unique/ds_worker/${IP_ADDRESS}" "${DS_WORKER_PORT},${NODE_ID}" "${ETCD_ENDPOINTS_ADDR}"
      fi
    fi
  fi
  if [ "X${FUNCTION_PROXY_UNIQUE_ENABLE}" = "Xtrue" ] && [ "X${DEPLOY_FUNCTION_PROXY}" = "Xtrue" ]; then
    local function_proxy_info=""
    function_proxy_info=$(get_from_etcd "${UNIQUE_KEY_PREFIX}/unique/function_proxy/${IP_ADDRESS}" "${ETCD_ENDPOINTS_ADDR}")
    local ret=$?
    if [ ${ret} -eq 0 ]; then
      local proxy_port=$(echo "${function_proxy_info}" | awk -F ',' '{print $1}')
      # if unique info is lost, reset it to etcd
      if [ "X${proxy_port}" = "X" ]; then
        set_to_etcd "${UNIQUE_KEY_PREFIX}/unique/function_proxy/${IP_ADDRESS}" "${FUNCTION_PROXY_PORT},${FUNCTION_PROXY_GRPC_PORT},${NODE_ID}" "${ETCD_ENDPOINTS_ADDR}"
      fi
    fi
  fi
}

function exit_all_processes()
{
  log_info "exiting all processes"
  for component in "${!pid_table[@]}"; do
    if ! need_health_check ${component}; then
      continue
    fi
    if ! is_child_process ${pid_table[$component]}; then
      continue
    fi
    terminate_process ${pid_table[$component]}
  done
  exit 1
}

function prepare_log_rotate()
{
  if [ "X${RUNTIME_STD_ROLLING_ENABLE}" != "Xtrue" ]; then
    return
  fi
  if ! command -v logrotate >/dev/null 2>&1; then
      log_error "logrotate command not found. runtime std log rolling cannot be enabled"
      RUNTIME_STD_ROLLING_ENABLE=false
      return
  fi
  echo "$LOG_ROTATE_CONFIG" > ${INSTALL_DIR_PARENT}/logrotate.conf
}

function run_log_rotate()
{
  if [ "X${RUNTIME_STD_ROLLING_ENABLE}" != "Xtrue" ]; then
    return
  fi
  logrotate ${INSTALL_DIR_PARENT}/logrotate.conf >/dev/null 2>&1
}

function main() {
  log_info "script parameters: $@"
  local ret_code
  source ${BASE_DIR}/config.sh "$@"
  ret_code=$?
  if [ ${ret_code} -ne 0 ]; then
    return ${ret_code}
  fi
  trap 'signal_handle' SIGTERM SIGINT EXIT
  local start time
  init_control_plane_port
  if [ "X${ENABLE_MASTER}" = "Xtrue" ]; then
    DEPLOY_STEP="MASTER"
    start=$(date +%s)
    start_control_plane
    ret_code=$?
    if [ ${ret_code} -ne 0 ]; then
      return ${ret_code}
    fi
    # if enable multi master, skip it to speed up starting
    if [ "X${ENABLE_MULTI_MASTER}" != "Xtrue" ]; then
      wait_for_ready_on_checklist
    fi
    dump_master_info
    time=$(echo $start "$(date +%s)" | awk '{print $2-$1}')
    log_info "yuanrong control plane components installation to ${INSTALL_DIR_PARENT} costs ${time} seconds."
  fi
  if ! wait_etcd_up; then
    log_error "failed to wait etcd ready"
    exit 1
  fi
  DEPLOY_STEP="AGENT"
  start=$(date +%s)
  init_data_plane_port
  start_or_wait_ds_worker
  # if enable multi master, skip it to speed up starting
  if [ "X${ENABLE_MULTI_MASTER}" != "Xtrue" ]; then
    # health check all after starting ds_worker,  just make sure when proxy is start, ds_worker can be ready
    wait_for_ready_on_checklist
  fi
  start_or_wait_function_proxy
  # make sure when agent is start, proxy can be ready
  if [ "x${DEPLOY_FUNCTION_PROXY}" = "xtrue" ]; then
      health_check "function_proxy" ${pid_table["function_proxy"]}
  fi
  if [ ${CPU4COMP} -le 100 ]; then
    log_warning "no cpu available for function agent, skip starting it"
  else
    start_function_agent
  fi
  start_faas_frontend
  start_function_scheduler
  start_dashboard
  start_collector
  start_meta_service
  # check all component is working, if not, restart with change port(for control plane only restart).
  wait_for_ready_on_checklist
  time=$(echo $start "$(date +%s)" | awk '{print $2-$1}')
  log_info "yuanrong data plane components installation to ${DATA_PLANE_INSTALL_DIR} costs ${time} seconds."
  # show error msg of process services.yaml
  show_component_error_log "${FS_LOG_PATH}/${NODE_ID}-function_proxy.log" "service_json.cpp" "LoadYourServicesYaml"
  dump_agent_info
  print_info
  set_unique_info
  prepare_log_rotate

  PORT_POLICY="FIX"
  DEPLOY_STEP="END"
  local collect_cnt=0
  local ds_worker_check_time=$((RANDOM % 300))
  local ds_worker_not_alive=0
  local compress_file_num=5
  local check_unique_lock_time=$((RANDOM % 60))

  local logrotate_interval=0
  # check die
  while :; do
    listen_component "false"
    ds_worker_check_time=$((ds_worker_check_time + 1))
    check_unique_lock_time=$((check_unique_lock_time + 1))
    if [ "X${ENABLE_MULTI_MASTER}" = "Xtrue" ] && [ $((check_unique_lock_time % 60)) -eq 0 ]; then
      check_unique_lock
      check_unique_lock_time=0
    fi
    #  check single ds_worker & kill the processes in pid_table
    if [ "X${DS_WORKER_UNIQUE_ENABLE}" = "Xtrue" ] && [ $((ds_worker_check_time % CHECK_DIR_INTERVAL)) -eq 0 ]; then
      check_ds_worker_alive
      ds_worker_not_alive=$?
      [ ${ds_worker_not_alive} -eq 1 ] && exit_all_processes
      ds_worker_check_time=0
    fi
    if [ "X$STATUS_COLLECT_ENABLE" == "XTRUE" ] && [ $collect_cnt -eq $STATUS_COLLECT_INTERVAL ]; then
        monitor "${LOG_ROOT}" "${NODE_ID}"
        collect_cnt=0
    fi
    ((collect_cnt+=1))
    if [ "${DEPLOY_DS_WORKER}X" == "trueX" ]; then
        log_backup "${LOG_ROOT}" $compress_file_num "ds_worker_std" ""
    fi
    if [ "X${ENABLE_MASTER}" = "Xtrue" ]; then
        log_backup "${LOG_ROOT}" $compress_file_num "ds_master_std" ""
    fi
    ((logrotate_interval+=1))
    if (( logrotate_interval % 5 == 0 )); then
        run_log_rotate
    fi
    sleep 1
  done
}

main "$@"
