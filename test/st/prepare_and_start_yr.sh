#!/usr/bin/env bash
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

set -e
BASE_DIR=$(cd "$(dirname "$0")"; pwd)
. "${BASE_DIR}/utils.sh"
SRC_DIR="${BASE_DIR}/../"
COMPILE_OUTPUT="${SRC_DIR}/output"
OUTPUT_DIR="${SRC_DIR}/../output/"
YUANRONG_DIR="${OUTPUT_DIR}/openyuanrong"
DEPLOY_PATH="/tmp/deploy/$(date -u "+%d%H%M%S")"
SANITIZER="off"
AGENT_NUM=1
ENABLE_FAAS="false"
PORT_POLICY="RANDOM"


LOCAL_IP=$(ifconfig | grep inet | grep -vw "127.0.0.1" | grep -vw "172.17.0.1"| grep -vw "10.42.0.1" | grep -v inet6 | awk '{print $2}'|awk '{print $1}')

while getopts 'y:d:t:a:S:n:fws' opt; do
    case "${opt}" in
    y)
        YUANRONG_DIR="${OPTARG}"
        ;;
    d)
        DEPLOY_PATH="${OPTARG}"
        ;;
    a)
        LOCAL_IP="${OPTARG}"
        ;;
    S)
        check_sanitizers "${OPTARG}"
        SANITIZER="${OPTARG}"
        ;;
    n)
        AGENT_NUM="${OPTARG}"
        ;;
    f)
        ENABLE_FAAS="true"
        ;;
    s)
        PORT_POLICY="FIX"
        ;;
    *)
        echo "invalid command"
        exit 1
        ;;
    esac
done

function wait_cluster_readiness() {
    echo "start check yuanrong cluster readiness..."
    mkdir -p ${DEPLOY_PATH}/yr_master/
    local etcd_ip=$(grep "etcd" ${DEPLOY_PATH}/yr_master/master.info | sed -nr 's/.*etcd_ip:([^,]*),.*/\1/p')
    local etcd_port=$(grep "etcd" ${DEPLOY_PATH}/yr_master/master.info | sed -nr 's/.*etcd_port:([^,]*),.*/\1/p')
    local addr="${etcd_ip}:${etcd_port}"
    local expect_local_scheduler_count=$((AGENT_NUM+1))

    local ready_local_scheduler_count=$(bash "$YUANRONG_DIR"/deploy/process/yr_ready.sh -e "${addr}")
    local retry_time=0
    while [[ "X${ready_local_scheduler_count}" == "X" || ${ready_local_scheduler_count} < ${expect_local_scheduler_count} ]]; do
        if [ "X${ready_local_scheduler_count}" == "X" ]; then
            echo "no local scheduler are running; expecting ${expect_local_scheduler_count}"
        else
            echo "ready local scheduler: [${ready_local_scheduler_count}/${expect_local_scheduler_count}]"
        fi
        sleep 1
        ready_local_scheduler_count=$(bash "$YUANRONG_DIR"/deploy/process/yr_ready.sh -e "${addr}")
        retry_time=$(($retry_time+1))
        if [ $retry_time -gt 20 ]; then
            echo "retry wait cluster readiness for 100 times, has reached max retry time and stop retry"
            exit 1
        fi
    done
    echo "ready local scheduler: [${ready_local_scheduler_count}/${expect_local_scheduler_count}]"
}

function start_yr() {
    cd "${YUANRONG_DIR}/"
    local services_file="${BASE_DIR}/services.yaml"
    local absolute_cpp_code_path=$(readlink -f "${DEPLOY_PATH}/pkg")
    local absolute_python_code_path=$(readlink -f "${BASE_DIR}/python")
    local absolute_java_code_path=$(readlink -f "${DEPLOY_PATH}/pkg")
    cp -f "$services_file" ${DEPLOY_PATH}
    sed -i "s#codePath: /cpp#codePath: ${absolute_cpp_code_path}/#g" "${DEPLOY_PATH}/services.yaml"
    sed -i "s#codePath: /python#codePath: ${absolute_python_code_path}/#g" "${DEPLOY_PATH}/services.yaml"
    sed -i "s#codePath: /java#codePath: ${absolute_java_code_path}/#g" "${DEPLOY_PATH}/services.yaml"
    if [ "${SANITIZER}" == "address" ]; then
      sed -i "s#FIRST_PYTHON_ENV: FIRST#LD_PRELOAD: /usr/lib64/libasan.so.6#g" "${DEPLOY_PATH}/services.yaml"
      sed -i "s#FIRST_JAVA_ENV: FIRST#LD_PRELOAD: /usr/lib64/libasan.so.6#g" "${DEPLOY_PATH}/services.yaml"
    fi
    local direct_call="false"
    if [ "X${RUNTIME_DIRECT_CONNECTION_ENABLE}" == "Xtrue" ]; then
        direct_call="true"
    fi
    bash deploy/process/yr_master.sh -d ${DEPLOY_PATH}/yr_master \
        -a ${LOCAL_IP} -l DEBUG -s 2048 -c 10000 -m 22048 \
        -o ${DEPLOY_PATH}/yr_master/master.info \
        --services_path "${DEPLOY_PATH}/services.yaml" \
        --runtime_heartbeat_timeout_ms 2000 \
        --runtime_max_heartbeat_timeout_times 3 \
        --runtime_direct_connection_enable "${direct_call}" \
        --is_protomsg_to_runtime  true \
        --deploy_path "$DEPLOY_PATH" \
        --state_storage_type "datasystem" \
        --port_policy "$PORT_POLICY" \
        --disable_nc_check
    for((k=0;k<"${AGENT_NUM}";k++))
    do
        nohup bash deploy/process/yr_agent.sh --master_info "$(cat ${DEPLOY_PATH}/yr_master/master.info)" \
            -a ${LOCAL_IP} -l DEBUG \
            --deploy_path ${DEPLOY_PATH}/yr_agent \
            --services_path "${DEPLOY_PATH}/services.yaml" \
            --runtime_heartbeat_timeout_ms 2000 \
            --runtime_max_heartbeat_timeout_times 3 \
            --is_protomsg_to_runtime  true \
            --runtime_direct_connection_enable "${direct_call}" \
            --state_storage_type "datasystem" \
            --disable_nc_check \
            -c 10000 -m 20000 -s 2048 > ${DEPLOY_PATH}/nohup_start.log 2>&1 &
    done
    wait_cluster_readiness
    echo "Succeed to start yuanrong"
}

mkdir -p "$DEPLOY_PATH"
rm -rf ./deploy
ln -s "$DEPLOY_PATH" ./deploy
echo "deploy to $DEPLOY_PATH"
start_yr
