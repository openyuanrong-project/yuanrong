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

readonly READY_AGENT_COUNT_KEY="/yr/readyAgentCount"
THIRD_PARTY_PATH=${BASE_DIR}/../../third_party
ETCDCTL_PATH=${THIRD_PARTY_PATH}/etcd/etcdctl
ETCD_LD_PATH=${THIRD_PARTY_PATH}/etcd/lib:${LD_LIBRARY_PATH}
QUERY_ADDRESS=""
ENABLE_TLS="false"
ETCD_SSL_BASE_PATH=""
TIMEOUT_SECOND=10

function isValidAddr() {
    addr=$1
    hasColon=`echo $addr | grep :`
    if [[ $hasColon == "" ]]; then
      echo "Invalid address of $addr"
      exit 1
    fi

    ip=${1%:*}
    port=${1#*:}
    ret=1
    if [[ $ip =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
      ip=(${ip//\./ })
      [[ ${ip[0]} -le 255 && ${ip[1]} -le 255 && ${ip[2]} -le 255 && ${ip[3]} -le 255 ]]
      ret=$?
    fi
    if [ $ret -ne 0 ]; then
        echo "Invalid ip of $addr"
        exit 1
    fi

    if [[ $port -gt 65534 || $port -lt 1025 ]]; then
        echo "Invalid port of $addr"
        exit 1
    fi
}

function usage() {
  echo -e "Usage: bash yr_ready.sh [-e address] [-h] [-s]"
}

function query_agent_count_from_etcd() {
  local ready_agent_count="0"
  if [ "X${ENABLE_TLS}" == "Xtrue" ]; then
    local tls_params="--cacert=${ETCD_SSL_BASE_PATH}/ca.crt --cert=${ETCD_SSL_BASE_PATH}/client.crt --key=${ETCD_SSL_BASE_PATH}/client.key"
    ready_agent_count=$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${TIMEOUT_SECOND}s ${ETCDCTL_PATH} \
                                  ${tls_params} --endpoints "https://${QUERY_ADDRESS}" get ${READY_AGENT_COUNT_KEY} --print-value-only)
  else
    ready_agent_count=$(LD_LIBRARY_PATH=${ETCD_LD_PATH} ETCDCTL_API=3 timeout ${TIMEOUT_SECOND}s ${ETCDCTL_PATH} \
                                  --endpoints "${QUERY_ADDRESS}" get ${READY_AGENT_COUNT_KEY} --print-value-only)
  fi
  if  [ "x${ready_agent_count}" == "x" ] || [ ${ready_agent_count} -lt 0 ]; then
    echo "0"
  else
    echo ${ready_agent_count}
  fi
}

function main() {
    while getopts "c:e:t:hs" opt; do
        case "$opt" in
        e)
            QUERY_ADDRESS=$OPTARG
            ;;
        h)
            usage
            exit 0
            ;;
        s)
            ENABLE_TLS="true"
            ;;
        t)   
            ETCD_SSL_BASE_PATH=$OPTARG
            ;;
        *)
            echo -e "Unknown parameter"
            usage
            exit 1
            ;;
        esac
    done


    if [ "x${QUERY_ADDRESS}" == "x" ]; then
      printf "**ERROR** empty address is provided\n"
      exit 1
    fi

    isValidAddr $QUERY_ADDRESS

    query_agent_count_from_etcd
}

main "$@"