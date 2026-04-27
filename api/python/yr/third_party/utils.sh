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

function setnx_or_get() {
    local key="$1"
    local val="$2"
    local addr="$3"
    local timeout_s=$4
    [[ -z "${timeout_s}" ]] && timeout_s=10
    # empty line is necessaries
    local result=""
    if [ "X${ETCD_AUTH_TYPE}" == "XTLS" ]; then
      local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} --cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} --key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}"
      result=$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} ${tls_params} --endpoints "${addr}" txn <<EOF
create("${key}") = "0"

put -- "${key}" "${val}"

get -- "${key}"

EOF
    )
    else
      result=$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} --endpoints "${addr}" txn <<EOF
create("${key}") = "0"

put -- "${key}" "${val}"

get -- "${key}"

EOF
    )
    fi
    local ret=$?
    if [ "${ret}" != 0 ]; then
        return ${ret}
    fi
    if [ $(echo ${result} | awk '{print $1}') == "SUCCESS" ]; then
        return 0
    fi
    echo $(echo ${result} | awk '{print $NF}')
    return 0
}

function get_from_etcd() {
    local key="$1"
    local addr="$2"
    local timeout_s="$3"
    [[ -z "${timeout_s}" ]] && timeout_s=10
    local result=""
    if [ "X${ETCD_AUTH_TYPE}" == "XTLS" ]; then
      local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} --cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} --key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}"
      result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} ${tls_params} --endpoints "${addr}" get --print-value-only -- "${key}")"
    else
      result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} --endpoints "${addr}" get --print-value-only -- "${key}")"
    fi
    local ret=$?
    echo "${result}"
    return ${ret}
}

function set_to_etcd() {
    local key="$1"
    local val="$2"
    local addr="$3"
    local timeout_s=$4
    [[ -z "${timeout_s}" ]] && timeout_s=10
    local result=""
    if [ "X${ETCD_AUTH_TYPE}" == "XTLS" ]; then
      local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} --cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} --key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}"
      result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} ${tls_params} --endpoints "${addr}" put -- "${key}" "${val}")"
    else
      result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} --endpoints "${addr}" put -- "${key}" "${val}")"
    fi
    local ret=$?
    return ${ret}
}

function del_from_etcd() {
    local key="$1"
    local addr="$2"
    local timeout_s=$3
    [[ -z "${timeout_s}" ]] && timeout_s=10
    local result=""
    if [ "X${ETCD_AUTH_TYPE}" == "XTLS" ]; then
      local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} --cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} --key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}"
      result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} ${tls_params} --endpoints "${addr}" del -- "${key}")"
    else
      result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} --endpoints "${addr}" del -- "${key}")"
    fi
    local ret=$?
    return ${ret}
}

function remove_cluster_node() {
  local addr="$1"
  local etcd_node_id="$2"
  local timeout_s=3
  local result=""
  # check etcd member list exist; if member not exist indicates this is a new cluster, else this is a old cluster
  if [ "X${ETCD_AUTH_TYPE}" == "XTLS" ]; then
    local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} --cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} --key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}"
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} ${tls_params} --endpoints "${addr}" member remove ${etcd_node_id})"
  else
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} --endpoints "${addr}" member remove ${etcd_node_id})"
  fi
}

function add_cluster_node() {
  local addr="$1"
  local etcd_node_name="$2"
  local timeout_s=3
  local result=""
  # check etcd member list exist; if member not exist indicates this is a new cluster, else this is a old cluster
  if [ "X${ETCD_AUTH_TYPE}" == "XTLS" ]; then
    local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} --cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} --key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}"
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} ${tls_params} --endpoints "${addr}" member add ${etcd_node_name} --peer-urls=${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PEER_PORT})"
  else
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} --endpoints "${addr}" member add ${etcd_node_name} --peer-urls=${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PEER_PORT})"
  fi
}

function recover_etcd_node() {
  local addr="$1"
  local timeout_s=3
  local result=""
  # check etcd member list exist; if member not exist indicates this is a new cluster, else this is a old cluster
  if [ "X${ETCD_AUTH_TYPE}" == "XTLS" ]; then
    local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} --cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} --key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}"
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} ${tls_params} --endpoints "${addr}" member list)"
  else
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} --endpoints "${addr}" member list)"
  fi
  # result example as follows
  # id1, started, etcd2, http://X.X.X.X:23280, http://X.X.X.X:23279, false
  # id2, started, etcd0, http://X.X.X.X:23280, http://X.X.X.X:23279, false
  # id3, started, etcd1, http://X.X.X.X:23280, http://X.X.X.X:23279, false
  local ret=$?
  if [ ${ret} -ne 0 ]; then
    log_info "etcd cluster is not ready"
    return ${ret}
  fi
  local current_node_info=$(echo "$result" | grep "${ETCD_NAME}, ${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PEER_PORT}, ${ETCD_HTTP_PROTOCOL}${ETCD_IP}:${ETCD_PORT},")
  # need remove old member info
  if [ "X${current_node_info}" != "X" ]; then
    local current_node_info_arr=(${current_node_info//,/ })
    local current_node_info_id=${current_node_info_arr[0]}
    if [ "X${current_node_info}" != "X" ]; then
      remove_cluster_node $addr $current_node_info_id
    fi
  fi
  # add new member
  add_cluster_node $addr ${ETCD_NAME}
  ETCD_INITIAL_CLUSTER_STATE="existing"
  return 0
}

function check_etcd_ready() {
  local addr="$1"
  local timeout_s=3
  local result=""
  if [ "X${ETCD_AUTH_TYPE}" == "XTLS" ]; then
    local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/${ETCD_CA_FILE} --cert=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_CERT_FILE} --key=${ETCD_SSL_BASE_PATH}/${ETCD_CLIENT_KEY_FILE}"
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} ${tls_params} --endpoints "${addr}" member list)"
  else
    result="$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${timeout_s}s ${ETCDCTL_PATH} --endpoints "${addr}" member list)"
  fi
  return $?
}