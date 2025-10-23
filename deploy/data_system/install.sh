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

DATA_SYSTEM_DEPLOY_DIR=$(dirname "$(readlink -f "$0")")
if [ -n "${BASE_DIR}" ]; then
  DATA_SYSTEM_DEPLOY_DIR=${BASE_DIR}/../../data_system/deploy
fi
DATA_SYSTEM_DIR=$(readlink -m "${DATA_SYSTEM_DEPLOY_DIR}/..")
MAX_PROCESS_EXIT_TIMES=20

# before use functions in this script, ${BASE_DIR}/utils.sh should be imported first

function install_ds_master() {
  log_info "start data system master, port=${DS_MASTER_PORT}..."
  local data_system_install_dir=${INSTALL_DIR_PARENT}/"ds_master_$$" # use ds_master_+pid as path, can't same as worker

  [[ ! -d "${data_system_install_dir}/rocksdb" ]] && mkdir -p "${data_system_install_dir}/rocksdb"
  [[ ! -d "${data_system_install_dir}/socket" ]] && mkdir -p "${data_system_install_dir}/socket"
  local ret_code=1
  DS_MASTER_HEALTH_CHECK_PATH="${data_system_install_dir}"/master/health
  mkdir -p "${DS_LOG_PATH}/master"
  if check_port "${DS_MASTER_IP}" "${DS_MASTER_PORT}"; then
    OPENSSL_CONF="" LD_LIBRARY_PATH="${DATA_SYSTEM_DIR}/service/lib:${LD_LIBRARY_PATH}" "${DATA_SYSTEM_DIR}"/service/datasystem_worker -master_address="${DS_MASTER_IP}:${DS_MASTER_PORT}" \
      -log_dir="${DS_LOG_PATH}/master" -worker_address="${DS_MASTER_IP}:${DS_MASTER_PORT}" \
      -unix_domain_socket_dir="${data_system_install_dir}/socket" -v="${DS_DEBUG_LOG_LEVEL}" \
      -minloglevel="${DS_LOG_LEVEL}" \
      -health_check_path="${data_system_install_dir}"/master/health \
      -ready_check_path="${data_system_install_dir}"/master/ready \
      -enable_reconciliation=false \
      -rpc_thread_num=${DS_RPC_THREAD_NUM} \
      -heartbeat_interval_ms=${DS_HEARTBEAT_INTERVAL_MS} \
      -enable_multi_stubs=true \
      -max_log_size=${DS_LOG_ROLLING_MAX_SIZE} -max_log_file_num=${DS_LOG_ROLLING_MAX_FILES} \
      -node_dead_timeout_s=${DS_NODE_DEAD_TIMEOUT_S} \
      -node_timeout_s=${DS_NODE_TIMEOUT_S} \
      -backend_store_dir="${data_system_install_dir}/rocksdb" \
      -etcd_address="${ETCD_CLUSTER_ADDRESS}" \
      -other_az_names="${ETCD_TABLE_PREFIX}" \
      -etcd_target_name_override="${ETCD_TARGET_NAME_OVERRIDE}" \
      -enable_etcd_auth=${ENABLE_ETCD_AUTH} \
      -arena_per_tenant=${DS_ARENA_PER_TENANT} \
      -enable_fallocate=${DS_ENABLE_FALLOCATE} \
      -enable_huge_tlb=${DS_ENABLE_HUGE_TLB} \
      -enable_thp=${DS_ENABLE_THP} \
      -etcd_ca=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} \
      -etcd_cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} \
      -etcd_key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE} \
      -enable_component_auth=${DS_COMPONENT_AUTH_ENABLE} \
      -curve_key_dir=${CURVE_KEY_PATH} \
      -logfile_mode=416 \
      -l2_cache_type=${DS_L2_CACHE_TYPE} \
      -sfs_path=${DS_SFS_PATH} \
      -log_monitor=${DS_LOG_MONITOR_ENABLE} \
      -zmq_chunk_sz=${ZMQ_CHUNK_SZ} \
      -enable_lossless_data_exit_mode=${ENABLE_LOSSLESS_DATA_EXIT_MODE} \
      -enable_distributed_master=false -stderrthreshold=3 >> "${DS_LOG_PATH}"/ds_master${STD_LOG_SUFFIX} 2>&1 &
    DS_MASTER_PID="$!"
    if data_system_health_check "ds_master" "${DS_MASTER_PID}"; then
      log_info "succeed to start  data system master, port=${DS_MASTER_PORT}, pid=${DS_MASTER_PID}"
      return 0
    else
      if [ "$?" == 2 ];then
        ret_code=2
        log_warning "data system master start timeout, killing data system master, pid=${DS_MASTER_PID}."
      else
        ret_code=98
        log_warning "data system master health_check failed, killing data system master, pid=${DS_MASTER_PID}."
      fi
    fi
    kill -15 ${DS_MASTER_PID}
    local j
    for ((j = 0; j <= "$MAX_PROCESS_EXIT_TIMES"; j++)); do
      if ! ps -ns ${DS_MASTER_PID} | grep ${DS_MASTER_PID} &>/dev/null; then
        break
      elif [ "$j" == "$MAX_PROCESS_EXIT_TIMES" ]; then
        log_warning "data system master failed to exit, will be forced to exit."
        kill -9 ${DS_MASTER_PID}
        rm -rf "${DATASYSTEM_INSTALL_DIR}"/rocksdb
      fi
      sleep 0.5
    done
  fi
  return ${ret_code}
}

function install_ds_worker() {
  log_info "start data system worker, port=${DS_WORKER_PORT}..."
  local data_system_install_path=${INSTALL_DIR_PARENT}/"ds_worker_$$" # use ds_master_+pid as path, can't same as master

  [[ ! -d "${data_system_install_path}/socket" ]] && mkdir -p "${data_system_install_path}/socket"
  [[ ! -d "${data_system_install_path}/rocksdb" ]] && mkdir -p "${data_system_install_path}/rocksdb"
  local spill_args=""
  if [ "X$DS_SPILL_ENABLE" == "Xtrue" ]; then
    [[ ! -d "$DS_SPILL_DIRECTORY" ]] && mkdir -p "$DS_SPILL_DIRECTORY"
    DS_WORKER_SPILL_SIZE_LIMIT=$((DS_SPILL_SIZE_LIMIT * 1024 * 1024))
    spill_args="-spill_directory=$DS_SPILL_DIRECTORY/prefix -spill_size_limit=$DS_WORKER_SPILL_SIZE_LIMIT"
  fi
  DS_WORKER_HEALTH_CHECK_PATH="${data_system_install_path}"/worker/health
  DS_MASTER_ADDRESS="${DS_MASTER_IP}:${DS_MASTER_PORT}"
  if [ "X${ENABLE_DISTRIBUTED_MASTER}" = "Xtrue" ]; then
    DS_MASTER_ADDRESS=""
  fi
  OPENSSL_CONF="" LD_LIBRARY_PATH="${DATA_SYSTEM_DIR}/service/lib:${LD_LIBRARY_PATH}" "${DATA_SYSTEM_DIR}"/service/datasystem_worker \
    -master_address=${DS_MASTER_ADDRESS} \
    -log_dir="${DS_LOG_PATH}" -shared_memory_size_mb=${MEM4DATA} -worker_address="${IP_ADDRESS}:${DS_WORKER_PORT}" \
    -unix_domain_socket_dir="${data_system_install_path}/socket" -v="${DS_DEBUG_LOG_LEVEL}" \
    -minloglevel="${DS_LOG_LEVEL}" \
    -health_check_path="${DS_WORKER_HEALTH_CHECK_PATH}" \
    -ready_check_path="${data_system_install_path}"/worker/ready \
    -enable_reconciliation=false \
    -rpc_thread_num=${DS_RPC_THREAD_NUM} \
    -client_dead_timeout_s=${DS_CLIENT_DEAD_TIMEOUT_S} \
    -heartbeat_interval_ms=${DS_HEARTBEAT_INTERVAL_MS} \
    -enable_multi_stubs=true \
    -max_log_size=${DS_LOG_ROLLING_MAX_SIZE} -max_log_file_num=${DS_LOG_ROLLING_MAX_FILES} \
    -max_client_num=${DS_MAX_CLIENT_NUM} $spill_args \
    -memory_reclamation_time_second=${DS_MEMORY_RECLAMATION_TIME_SECOND} \
    -node_dead_timeout_s=${DS_NODE_DEAD_TIMEOUT_S} \
    -node_timeout_s=${DS_NODE_TIMEOUT_S} \
    -backend_store_dir="${data_system_install_path}/rocksdb" \
    -arena_per_tenant=${DS_ARENA_PER_TENANT} \
    -enable_fallocate=${DS_ENABLE_FALLOCATE} \
    -enable_huge_tlb=${DS_ENABLE_HUGE_TLB} \
    -enable_thp=${DS_ENABLE_THP} \
    -etcd_address="${ETCD_CLUSTER_ADDRESS}" \
    -other_az_names="${ETCD_TABLE_PREFIX}" \
    -etcd_target_name_override="${ETCD_TARGET_NAME_OVERRIDE}" \
    -enable_etcd_auth=${ENABLE_ETCD_AUTH} \
    -etcd_ca=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} \
    -etcd_cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} \
    -etcd_key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE} \
    -enable_component_auth=${DS_COMPONENT_AUTH_ENABLE} \
    -curve_key_dir=${CURVE_KEY_PATH} -logfile_mode=416 \
    -l2_cache_type=${DS_L2_CACHE_TYPE} \
    -sfs_path=${DS_SFS_PATH} \
    -log_monitor=${DS_LOG_MONITOR_ENABLE} \
    -zmq_chunk_sz=${ZMQ_CHUNK_SZ} \
    -enable_lossless_data_exit_mode=${ENABLE_LOSSLESS_DATA_EXIT_MODE} \
    -enable_distributed_master=${ENABLE_DISTRIBUTED_MASTER} -stderrthreshold=3 >> "${DS_LOG_PATH}"/ds_worker${STD_LOG_SUFFIX} 2>&1 &
  DS_WORKER_PID="$!"
  log_info "succeed to start data system worker, port=${DS_WORKER_PORT}, pid=${DS_WORKER_PID}"
}

function install_data_system() {
  case "$1" in
  ds_master)
    install_ds_master
    return $?
    ;;
  ds_worker)
    install_ds_worker
    return $?
    ;;
  *)
    log_warning >&2 "Unknown component $1"
    return 1
    ;;
  esac
}
