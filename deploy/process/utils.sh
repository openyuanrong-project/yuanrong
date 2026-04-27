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

# ----------------------------------------------------------------------
# funcname:     log_info.
# description:  Print build info log.
# parameters:   NA
# return value: NA
# ----------------------------------------------------------------------
log_info() {
    echo "[INFO][$(date +%b\ %d\ %H:%M:%S)]$*"
}

# ----------------------------------------------------------------------
# funcname:     log_warning.
# description:  Print build warning log.
# parameters:   NA
# return value: NA
# ----------------------------------------------------------------------
log_warning() {
    echo "[WARNING][$(date +%b\ %d\ %H:%M:%S)]$*"
}

# ----------------------------------------------------------------------
# funcname:     log_error.
# description:  Print build error log.
# parameters:   NA
# return value: NA
# ----------------------------------------------------------------------
log_error() {
    echo "[ERROR][$(date +%b\ %d\ %H:%M:%S)]$*"
}

# ----------------------------------------------------------------------
# funcname:     die.
# description:  Print build error log.
# parameters:   NA
# return value: NA
# ----------------------------------------------------------------------
die() {
    log_error "$*"
    stty echo
    exit 1
}

function trim_hostname() {
  HOST_NAME=$(get_hostname)
  local length=${#HOST_NAME}
  if [ "$length" -eq 0 ]; then
    random_string=$(head /dev/urandom | tr -dc 'a-z' | head -c 6)
    echo "node-$random_string"
  else
    echo "$HOST_NAME"
  fi
}

function get_hostname() {
    if [ "$(command -v hostname)" ]; then
        hostname
    elif [ "$(command -v python3)" ]; then
        python3 -c "import socket; print(socket.gethostname())"
    else
        die "cannot get hostname, need \"hostname\" or \"python3\""
    fi
}

function get_ip_list() {
    if [ "$(command -v python3)" ]; then
        python3 -c "import socket; s=socket.socket(socket.AF_INET, socket.SOCK_DGRAM);s.connect(('8.8.8.8', 80));print(s.getsockname()[0]);"
    elif [ "$(command -v ifconfig)" ]; then
        ifconfig | grep inet | grep -vw "127.0.0.1" | grep -vw "172.17.0.1" | grep -v inet6 | awk '{print $2}'
    elif [ "$(command -v ip)" ]; then
        ip addr show | grep inet | grep -vw "127.0.0.1" | grep -vw "172.17.0.1" | grep -v inet6 | awk '{print $2}' | cut -f1 -d '/'
    else
        die "cannot get host ip, need \"ip\" or \"ifconfig\" or \"python3\""
    fi
}

# get node IP automatically
function get_ip_auto() {
    echo "Try to get IP address of this device"
    ip_list_len=$(get_ip_list | wc -l)
    local_ip=0
    if [ "$ip_list_len" -ge 2 ]; then
        echo "Cannot get IP address of this device. Please choose the appropriate one manually"
        for i in $(seq 1 "$ip_list_len"); do
            ip=$(get_ip_list | head -n "$i" | tail -n 1)
            read -rp "Local IP address is ${ip}. Press y to ensure or press Enter to skip:" conf
            conf_flag='x'$conf
            if [ "$conf_flag" == 'xy' ]; then
                local_ip=$ip
                break
            fi
        done
    fi
    if [ "$ip_list_len" -eq 1 ]; then
        local_ip=$(get_ip_list)
    fi
}

function check_ip_exist() {
    local ip_address=$1
    if [ "$(command -v ip)" ]; then
        ip addr show | grep inet | grep -v inet6 | awk '{print $2}' | grep -w "${ip_address}" &>/dev/null
        return $?
    elif [ "$(command -v ifconfig)" ]; then
        ifconfig | grep inet | grep -v inet6 | awk '{print $2}' | grep -w "${ip_address}" &>/dev/null
        return $?
    elif [ "$(command -v python3)" ]; then
        python3 -c "import socket; sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM);sock.bind((\"${ip_address}\", 0));sock.close();" &>/dev/null
        return $?
    else
        die "cannot get host ip, need \"ip\" or \"ifconfig\" or \"python3\""
    fi
}

# please install libxml2 first
function init_config_var() {
    local config_file=$1
    DEPLOY_PATH="$(xmllint --xpath "string(//config/deploy_path)" "${config_file}")"
    LOCAL_IP="$(xmllint --xpath "string(//config/local_ip)" "${config_file}")"
    REPO_PORT="$(xmllint --xpath "string(//config/function_repo_port)" "${config_file}")"
    ADMIN_PORT="$(xmllint --xpath "string(//config/admin_port)" "${config_file}")"
    export DEPLOY_PATH LOCAL_IP REPO_PORT ADMIN_PORT

    ETCD_AUTH_TYPE="$(xmllint --xpath "string(//config/etcd_auth_type)" "${config_file}")"
    ETCD_PORT="$(xmllint --xpath "string(//config/etcd_port)" "${config_file}")"
    ETCD_IP="$(xmllint --xpath "string(//config/etcd_ip)" "${config_file}")"
    ETCD_PROXY_PORT="$(xmllint --xpath "string(//config/etcd_proxy_ports)" "${config_file}")"
    ETCD_PROXY_ENABLE="$(xmllint --xpath "string(//config/etcd_proxy_enable)" "${config_file}")"
    export ETCD_AUTH_TYPE ETCD_PORT ETCD_IP ETCD_PROXY_PORT ETCD_PROXY_ENABLE

    MINIO_IP="$(xmllint --xpath "string(//config/minio_ip)" "${config_file}")"
    MINIO_PORT="$(xmllint --xpath "string(//config/minio_port)" "${config_file}")"
    export MINIO_IP MINIO_PORT

    REDIS_PORT="$(xmllint --xpath "string(//config/redis_port)" "${config_file}")"
    export REDIS_PORT

    WORKERMGR_LISTEN_PORT="$(xmllint --xpath "string(//config/workermgr_listen_port)" "${config_file}")"
    CODE_DIR="$(xmllint --xpath "string(//config/code_dir)" "${config_file}")"
    DS_MASTER_IP="$(xmllint --xpath "string(//config/ds_master_ip)" "${config_file}")"
    DS_MASTER_PORT="$(xmllint --xpath "string(//config/ds_master_port)" "${config_file}")"
    export WORKERMGR_LISTEN_PORT CODE_DIR DS_MASTER_IP DS_MASTER_PORT

    LOG_LEVEL="$(xmllint --xpath "string(//config/log/function_system/level)" "${config_file}")"
    RUNTIME_LOG_LEVEL="$(xmllint --xpath "string(//config/log/runtime/level)" "${config_file}")"
    export LOG_LEVEL RUNTIME_LOG_LEVEL

    FUNCTIONCORE_LOG_PATH="$(xmllint --xpath "string(//config/log/function_system/path)" "${config_file}")"
    export FUNCTIONCORE_LOG_PATH
    if [ "$FUNCTIONCORE_LOG_PATH" = "default" ]; then
        unset FUNCTIONCORE_LOG_PATH
    fi

    RUNTIME_LOG_PATH="$(xmllint --xpath "string(//config/log/runtime/path)" "${config_file}")"
    export RUNTIME_LOG_PATH
    if [ "$RUNTIME_LOG_PATH" = "default" ]; then
        unset RUNTIME_LOG_PATH
    fi
    LOG_ROLLING_MAXSIZE="$(xmllint --xpath "string(//config/log/function_system/rolling/maxsize)" "${config_file}")"
    LOG_ROLLING_MAXFILES="$(xmllint --xpath "string(//config/log/function_system/rolling/maxfiles)" "${config_file}")"
    LOG_ASYNC_LOGBUFSECS="$(xmllint --xpath "string(//config/log/function_system/async/logBufSecs)" "${config_file}")"
    LOG_ASYNC_MAXQUEUESIZE="$(xmllint --xpath "string(//config/log/function_system/async/maxQueueSize)" "${config_file}")"
    LOG_ASYNC_THREADCOUNT="$(xmllint --xpath "string(//config/log/function_system/async/threadCount)" "${config_file}")"
    LOG_ALSOLOGTOSTDERR="$(xmllint --xpath "string(//config/log/function_system/alsologtostderr)" "${config_file}")"
    export LOG_ROLLING_MAXSIZE LOG_ROLLING_MAXFILES LOG_ASYNC_LOGBUFSECS LOG_ASYNC_MAXQUEUESIZE LOG_ASYNC_THREADCOUNT
    export LOG_ALSOLOGTOSTDERR

    GLOBAL_SCHEDULER_PORT="$(xmllint --xpath "string(//config/global_scheduler_port)" "${config_file}")"
    LOKI_IP="$(xmllint --xpath "string(//config/loki_ip)" "${config_file}")"
    LOKI_PORT="$(xmllint --xpath "string(//config/loki_port)" "${config_file}")"
    export GLOBAL_SCHEDULER_PORT LOKI_IP LOKI_PORT

    export SYS_FUNC_RETRY_PERIOD="$(xmllint --xpath "string(//config/sys_func_retry_period)" "${config_file}")"
}

declare -A PORT_HASH_MAP

function set_master_used_ds_port() {
    for PORT in $MASTER_USED_DS_PORT; do
        PORT_HASH_MAP["${PORT}"]="${PORT}"
    done
}

declare -a ORDERED_PORTS=()
function get_free_port() {
    local bind_ip="$1"
    local port_min="$2"
    local port_max="$3"
    command -v nc &>/dev/null
    local ret=$?
    # if nc is not install or nc disabled, skip port check
    if [ ${ret} != 0 ] || [ "x${DISABLE_NC_CHECK}" = "xtrue" ]; then
        PORT=$(shuf -i "${port_min}"-"${port_max}" -n 1)
        while [[ ${PORT_HASH_MAP["$PORT"]} ]]; do
            PORT=$(shuf -i "${port_min}"-"${port_max}" -n 1)
        done
        PORT_HASH_MAP[$PORT]=$PORT
        ORDERED_PORTS=("$PORT" "${ORDERED_PORTS[@]}")
        echo "$PORT"
        return 0
    fi

    CHECK="port not assigned"
    PORT="port"
    while [[ -n "$CHECK" ]]; do
        PORT=$(shuf -i "${port_min}"-"${port_max}" -n 1)
        if [[ -z "${PORT_HASH_MAP[$PORT]}" ]]; then
            set +e; CHECK=$(LD_LIBRARY_PATH="" timeout 0.2 nc -l "$bind_ip" "$PORT" >/dev/null 2>&1)
        fi
    done
    PORT_HASH_MAP[$PORT]=$PORT
    ORDERED_PORTS=("$PORT" "${ORDERED_PORTS[@]}")
    echo "$PORT"
    return 0
}

function check_port() {
    command -v nc &>/dev/null
    local ret=$?
    # if nc is not install or nc disabled, skip port check
    if [ ${ret} != 0 ] || [ "x${DISABLE_NC_CHECK}" = "xtrue" ]; then
        return 0
    fi
    local bind_ip="$1"
    local port="$2"
    local check=$(LD_LIBRARY_PATH="" timeout 0.2 nc -l "$bind_ip" "$port" >/dev/null 2>&1)
    if [[ -n "$check" ]]; then
        return 1
    fi
    return 0
}

function log_backup()
{
    local log_file_dir="$1"
    local log_file_num="$2"
    local name_suffix="$3"
    local node_id="$4"
    local date_time old_file cur_log_file_num logfile_size
    local maxsize=$((1024*1024*10))
    if [ ! -f "$log_file_dir/${node_id}${name_suffix}.log" ]; then
        return
    fi
    logfile_size=$(stat -c %s "$log_file_dir/${node_id}${name_suffix}.log" | tr -d '\n')

    if [ "$logfile_size" -lt "$maxsize" ]; then
        return
    fi

    date_time=$(date +%Y_%m_%d_%H_%M_%S)
    local compress_file_name="${node_id}${name_suffix}_${date_time}"

    cp "$log_file_dir/${node_id}${name_suffix}.log" "$log_file_dir/$compress_file_name"
    echo '' > "$log_file_dir/${node_id}${name_suffix}.log"
    tar -zcf "$log_file_dir/$compress_file_name".tar.gz -C "$log_file_dir" "${node_id}${name_suffix}_${date_time}"
    rm -rf "$log_file_dir/$compress_file_name"
    while :; do
       cur_log_file_num=$(ls -l $log_file_dir/${node_id}${name_suffix}_*.tar.gz | wc -l)
       if [ "$cur_log_file_num" -gt "$log_file_num" ];then
           old_file=$(ls -tl $log_file_dir/${node_id}${name_suffix}_*.tar.gz | tail -1 | awk -F " " '{print $9}')
           [[ -n "$old_file" ]] && rm -rf "$old_file"
       else
           break;
       fi
    done
}

function set_runtime_pid_list() {
    local ppid
    for ppid in "${runtime_ppid_list[@]}"; do
        child_process_pids=$(pgrep -P "$ppid")
        for child_pid in $child_process_pids; do
            runtime_pid_list+=("${child_pid}")
        done
    done
}

function record_runtime_ppid() {
    local parent_process_name
    local pid=$1
    if [ "X${MERGE_PROCESS_ENABLE^^}" = "XTRUE" ];then
        parent_process_name="function_agent"
    else
        parent_process_name="runtime_manager"
    fi
    process_name=$(ps -p $pid -o comm=)
    if [ "${process_name}" == "${parent_process_name}" ]; then
      runtime_ppid_list+=("$pid")
    fi
}

function record() {
    local fd_threshold=4000
    local fd_num
    local pid=$1
    fd_num=$(ls /proc/"$pid"/fd 2>/dev/null| wc -l)
    if [ "$fd_num" == 0 ]; then
        return
    fi
    log_info "$pid fd num: $fd_num"
    if [ "$fd_num" -gt "$fd_threshold" ]; then
        ls -l /proc/"$pid"/fd
    fi
    log_info "$pid stat: $(($(date '+%s')*1000+10#$(date '+%N')/1000000)) $(cat /proc/"$pid"/stat 2>/dev/null)"
    ps -ef | awk -v pid="$pid" '{if(pid==$2)print}'
    record_runtime_ppid $pid
}

function monitor() {
    local log_dir=$1
    local node_id=$2
    local compress_file_num=5
    local pgid pid
    runtime_ppid_list=()
    log_info "start to monitor" &>>"${log_dir}"/${node_id}-monitor_std.log
    pgid=$(ps -o pgid $$|tail -n 1|sed s/[[:space:]]//g)
    for pid in $(ps axj | awk -v pgid="$pgid" '{if(pgid==$3)print}' | grep -v grep |grep -v PPID | awk '{print $2}');
    do
        record $pid &>>"${log_dir}"/${node_id}-monitor_std.log
    done

    runtime_pid_list=()
    set_runtime_pid_list
    for pid in "${runtime_pid_list[@]}"; do
        record $pid &>>"${log_dir}"/${node_id}-monitor_std.log
    done
    log_backup "${log_dir}" "$compress_file_num" "-monitor_std" "${node_id}"
}

function check_and_set_component_checklist() {
    local new_component=$1
    local pid=$2
    pid_table["${new_component}"]=${pid}
    for c in ${COMPONENT_CHECKLIST[@]}; do
        if [ X"${c}" = X"${new_component}" ]; then
            return
        fi
    done
    COMPONENT_CHECKLIST+=("${new_component}")
}

function show_component_error_log() {
  local log_name=$1
  local key_word=$2
  local note=$3
  local output=$(grep -a -E "^[EW]" ${log_name} | grep "${key_word}")
  if [ -n "$output" ]; then
    log_warning "${note} occurs error, errors as follows:"
    grep -E "^[EW]" ${log_name} | grep "${key_word}" | xargs -I {} echo "{}"
    log_warning "For more details, see ${log_name}."
  fi
}