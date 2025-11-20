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

declare -A control_port_table
declare -A data_port_table

function init_control_plane_port() {
  # etcd port
  control_port_table["etcd_port"]=${ETCD_PORT}
  control_port_table["etcd_peer_port"]=${ETCD_PEER_PORT}
  if [ "X${ENABLE_MULTI_MASTER}" != "Xtrue" ]; then
      update_control_plane_port "etcd_port etcd_peer_port"
  fi
  # ds master
  if [ "X${ENABLE_DISTRIBUTED_MASTER}" = "Xfalse" ]; then
    control_port_table["ds_master_port"]=${DS_MASTER_PORT}
    update_control_plane_port "ds_master_port"
  fi
  # function master
  control_port_table["global_scheduler_port"]=${GLOBAL_SCHEDULER_PORT}
  update_control_plane_port "global_scheduler_port"
}

function update_control_plane_port() {
  if [ "X${PORT_POLICY}" = "XFIX" ] || [ "X${ENABLE_MASTER}" != "Xtrue" ] || [ "X${DEPLOY_STEP}" = "XAGENT" ]; then
    return 0
  fi
  local tag_array=($(echo "$@"))
  local len=${#tag_array[@]}
  local j
  declare -a ORDERED_PORTS=()
  for ((j = 0; j < len; j++)); do
    get_free_port "$IP_ADDRESS" "$CONTROL_PLANE_PORT_MIN" "$CONTROL_PLANE_PORT_MAX" >/dev/null
  done
  for ((j = 0; j < len; j++)); do
    if [ "X${port_policy_table[${tag_array[j]}]}" == "XFIX" ]; then
          continue
    fi
    control_port_table[${tag_array[j]}]=${ORDERED_PORTS[j]}
  done
}

function init_data_plane_port() {
  data_port_table["function_proxy_port"]=${FUNCTION_PROXY_PORT}
  data_port_table["function_proxy_grpc_port"]=${FUNCTION_PROXY_GRPC_PORT}
  data_port_table["function_agent_port"]=${FUNCTION_AGENT_PORT}
  data_port_table["runtime_init_port"]=${RUNTIME_INIT_PORT}
  data_port_table["ds_worker_port"]=${DS_WORKER_PORT}
  data_port_table["dashboard_port"]=${DASHBOARD_PORT}
  data_port_table["dashboard_grpc_port"]=${DASHBOARD_GRPC_PORT}
  data_port_table["collector_port"]=${COLLECTOR_PORT}
  data_port_table["faas_frontend_http_port"]=${FAAS_FRONTEND_HTTP_PORT}
  data_port_table["faas_frontend_grpc_port"]=${FAAS_FRONTEND_GRPC_PORT}
}

function update_data_plane_port() {
  if [ "X${PORT_POLICY}" = "XFIX" ]; then
    return 0
  fi
  local tag_array=($(echo "$@"))
  local len=${#tag_array[@]}
  local j
  declare -a ORDERED_PORTS=()
  for ((j = 0; j < len; j++)); do
    get_free_port "$IP_ADDRESS" "$DATA_PLANE_PORT_MIN" "$DATA_PLANE_PORT_MAX" >/dev/null
  done
  for ((j = 0; j < len; j++)); do
    if [ "X${port_policy_table[${tag_array[j]}]}" == "XFIX" ]; then
          continue
    fi
    data_port_table[${tag_array[j]}]=${ORDERED_PORTS[j]}
  done
}
