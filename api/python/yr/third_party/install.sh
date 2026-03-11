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

THIRD_PARTY_BASE_DIR=$(dirname "$(readlink -f "$0")")
if [ -n "${BASE_DIR}" ]; then
  THIRD_PARTY_BASE_DIR=$(readlink -m "${BASE_DIR}/../../third_party")
fi

readonly ETCD_MAX_RETRY_NUM=10
readonly ETCD_MAX_FAIL_CNT=3
PORT_MIN=25500
PORT_MAX=25800
unset http_proxy
unset https_proxy

# before use functions in this script, utils.sh should be imported first

function install_etcd_proxy() {
  local proxy_name="etcd_proxy_$1"
  local etcd_proxy_port=$2
  local etcd_install_dir="${INSTALL_DIR_PARENT}/third_party/etcd"
  local log_file="${ETCD_LOG_PATH}/${proxy_name}.log"
  log_info "start ${proxy_name}, port=${etcd_proxy_port}..."
  local j
  if ! check_port "${ETCD_IP}" "${etcd_proxy_port}"; then
    log_error "${etcd_proxy_port} bind: address already in use"
    return 1
  fi
  if [ "${ETCD_AUTH_TYPE}" == "Noauth" ]; then
    log_info "start ${proxy_name}" >>$log_file 2>&1
    LD_LIBRARY_PATH=${ETCD_LD_PATH} "${THIRD_PARTY_BASE_DIR}"/etcd/etcd grpc-proxy start --endpoints="${ETCD_ENDPOINTS_ADDR}" \
      --listen-addr="${ETCD_IP}:${etcd_proxy_port}" \
      --advertise-client-url="${ETCD_IP}:${etcd_proxy_port}" \
      --resolver-prefix="___grpc_proxy_endpoint" \
      --resolver-ttl=60 >$log_file 2>&1 &
  elif [ "${ETCD_AUTH_TYPE}" == "TLS" ]; then
    log_info "start ${proxy_name}" >>$log_file 2>&1
    LD_LIBRARY_PATH=${ETCD_LD_PATH} "${THIRD_PARTY_BASE_DIR}"/etcd/etcd grpc-proxy start --endpoints="${ETCD_ENDPOINTS_ADDR}" \
      --listen-addr="${ETCD_IP}:${etcd_proxy_port}" \
      --advertise-client-url="${ETCD_IP}:${etcd_proxy_port}" \
      --cert="$ETCD_CERT_FILE" --key="$ETCD_KEY_FILE" --cacert="$ETCD_CA_FILE" \
      --cert-file="$CLIENT_CERT_FILE" --key-file="$CLIENT_KEY_FILE" --trusted-ca-file="$ETCD_CA_FILE" \
      --resolver-prefix="___grpc_proxy_endpoint" \
      --resolver-ttl=60 >$log_file 2>&1 &
  else
    log_error "proxy not support current auth type: ${ETCD_AUTH_TYPE}"
    return 1
  fi
  ETCD_PROXY_PID=$!
  count=0
  log_info "check etcd proxy status" $log_file 2>&1
  while [ "${count}" -lt "${ETCD_MAX_RETRY_NUM}" ]; do
    local check_param=--endpoints="${ETCD_IP}:${etcd_proxy_port}"
    if [ "${ETCD_AUTH_TYPE}" == "TLS" ]; then
      check_param="--cacert=${ETCD_CA_FILE} --cert=${CLIENT_CERT_FILE} --key=${CLIENT_KEY_FILE} --endpoints=${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${etcd_proxy_port}"
    fi
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} "${THIRD_PARTY_BASE_DIR}"/etcd/etcdctl ${check_param} endpoint health)"
    ret=$?
    log_info "ret is ${ret}"
    if [ ${ret} -eq 0 ]; then
      log_info "${proxy_name} is ready, pid=${ETCD_PROXY_PID}"
      return 0
    else
      if ! kill -0 "${ETCD_PROXY_PID}" &>/dev/null; then
        log_error "etcd proxy process is not exist"
        # process not exist
        return 1
      fi
      ((count+=1))
      log_info "failed to start ${proxy_name}, number of retries: ${count}" >>$log_file 2>&1
      sleep 2
    fi
  done
  if [ ${ETCD_PROXY_PID} -gt 0 ]; then
    log_warning "health check failed, killing ${proxy_name} process pid: ${ETCD_PROXY_PID}"
    kill -9 ${ETCD_PROXY_PID}
  fi
  return 1
}

function install_etcd() {
  log_info "start etcd, ip=${ETCD_IP},port=${ETCD_PORT},peer_port=${ETCD_PEER_PORT},cluster=${ETCD_INITIAL_CLUSTER}..."
  if ! check_port "${ETCD_IP}" "${ETCD_PORT}" || ! check_port "${ETCD_IP}" "${ETCD_PEER_PORT}"; then
    log_error "start etcd port is already bind"
    return 1
  fi
  local etcd_install_dir="${INSTALL_DIR_PARENT}/third_party/etcd"
  [ -d "${etcd_install_dir}" ] || mkdir -p "${etcd_install_dir}"
  if [ "X${ENABLE_MULTI_MASTER}" = "Xtrue" ] && [[ $ETCD_NODE_NUM -gt 1 ]]; then
    # if cluster node is re-deploy or restart etcd exceed max times, will try to recover
    if [ "X${ETCD_IS_FIRST_DEPLOY}" = "Xtrue" ] || [ ${ETCD_DEPLOY_FAIL_CNT} -gt ${ETCD_MAX_FAIL_CNT} ]; then
      recover_etcd_node "${ETCD_ENDPOINTS_ADDR}"
    fi
  fi
  if [ "X${ETCD_IS_FIRST_DEPLOY}" = "Xtrue" ]; then
    # clear old deploy data
    rm -rf "${etcd_install_dir}"/*
    ETCD_IS_FIRST_DEPLOY="false"
  fi
  chmod 700 "${etcd_install_dir}"
  local log_file=${ETCD_LOG_PATH}/etcd-run.log
  local auto_compaction_mode="revision"
  local auto_compaction_retention=${ETCD_COMPACT_RETENTION}
  local quota_backend=8589934592
  local etcd_cluster_params=""
  if [ "X${ENABLE_MULTI_MASTER}" = "Xtrue" ]; then
    etcd_cluster_params="--initial-advertise-peer-urls ${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PEER_PORT} --initial-cluster-token etcd-cluster-1 --initial-cluster ${ETCD_INITIAL_CLUSTER} --initial-cluster-state ${ETCD_INITIAL_CLUSTER_STATE}"
  fi
  if [ "${ETCD_AUTH_TYPE}" == "TLS" ]; then
    LD_LIBRARY_PATH=${ETCD_LD_PATH} "${THIRD_PARTY_BASE_DIR}"/etcd/etcd \
      --name=${ETCD_NAME} \
      --unsafe-no-fsync="${ETCD_NO_FSYNC}" \
      --data-dir="${etcd_install_dir}" \
      --initial-advertise-peer-urls="${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PEER_PORT}" \
      --listen-client-urls="${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PORT}" \
      --advertise-client-urls="${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PORT}" \
      --listen-peer-urls="${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PEER_PORT}" \
      --client-cert-auth ${etcd_cluster_params} \
      --trusted-ca-file="${ETCD_SSL_BASE_PATH}/$ETCD_CA_FILE" \
      --cert-file="${ETCD_SSL_BASE_PATH}/$ETCD_CERT_FILE" \
      --key-file="${ETCD_SSL_BASE_PATH}/$ETCD_KEY_FILE" \
      --auto-compaction-mode=${auto_compaction_mode} \
      --auto-compaction-retention=${auto_compaction_retention} \
      --quota-backend-bytes=${quota_backend} \
      --cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384 \
      --peer-client-cert-auth \
      --peer-trusted-ca-file="${ETCD_SSL_BASE_PATH}/$ETCD_CA_FILE" \
      --peer-cert-file="${ETCD_SSL_BASE_PATH}/$ETCD_CERT_FILE" \
      --peer-key-file="${ETCD_SSL_BASE_PATH}/$ETCD_KEY_FILE" >>$log_file 2>&1 &
  else
    LD_LIBRARY_PATH=${ETCD_LD_PATH} "${THIRD_PARTY_BASE_DIR}"/etcd/etcd \
      --name=${ETCD_NAME} \
      --unsafe-no-fsync="${ETCD_NO_FSYNC}" \
      --data-dir="${etcd_install_dir}" \
      --auto-compaction-mode=${auto_compaction_mode} \
      --auto-compaction-retention=${auto_compaction_retention} \
      --quota-backend-bytes=${quota_backend} \
      --listen-client-urls="${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PORT}" \
      --advertise-client-urls="${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PORT}" ${etcd_cluster_params} \
      --listen-peer-urls="${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PEER_PORT}" >>$log_file 2>&1 &
  fi
  ETCD_PID=$!

  local count=0
  log_info "check etcd status" >>$log_file
  while [ "${count}" -lt "${ETCD_MAX_RETRY_NUM}" ]; do
    local check_param=--endpoints="${ETCD_IP}:${ETCD_PORT}"
    if [ "${ETCD_AUTH_TYPE}" == "TLS" ]; then
      local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} --cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} --key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}"
      check_param="${tls_params} --endpoints=${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PORT}"
    fi
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} "${THIRD_PARTY_BASE_DIR}"/etcd/etcdctl ${check_param} endpoint health)"
    ret=$?
    log_info "ret is ${ret}"
    if [ ${ret} -eq 0 ]; then
      log_info "etcd is ready"
      break
    else
      if ! kill -0 "${ETCD_PID}" &>/dev/null; then
        log_error "etcd process is not exist"
        ((ETCD_DEPLOY_FAIL_CNT+=1))
        # process not exist
        return 1
      fi
      ((count+=1))
      log_info "number of retries: ${count}"
      sleep 2
    fi
  done

  if [ "${count}" -ge "${ETCD_MAX_RETRY_NUM}" ]; then
    log_error "etcd is not ready"
    if [ ${ETCD_PID} -gt 0 ]; then
      ((ETCD_DEPLOY_FAIL_CNT+=1))
      log_error "killing etcd, pid=${ETCD_PID}"
      kill -9 ${ETCD_PID}
    fi
    return 1
  fi

  # init etcd
  if [ "${ETCD_AUTH_TYPE}" == "Noauth" ]; then
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} "${THIRD_PARTY_BASE_DIR}"/etcd/etcdctl --endpoints="${ETCD_IP}:${ETCD_PEER_PORT}" auth disable)"
  fi
  if [[ $ETCD_NODE_NUM -gt 1 ]]; then
    ETCD_INITIAL_CLUSTER_STATE="existing"
  fi
  ETCD_DEPLOY_FAIL_CNT=0
  log_info "succeed to start etcd, port=${ETCD_PORT},peer_port=${ETCD_PEER_PORT}, pid=${ETCD_PID}"
  return 0
}

function install_third_party() {
  case "$1" in
  etcd)
    install_etcd
    return $?
    ;;
  etcd_proxy)
    install_etcd_proxy $2 $3
    return $?
    ;;
  *)
    log_warning >&2 "Unknown component $1"
    return 1
    ;;
  esac
}
