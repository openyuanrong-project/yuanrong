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

BASE_DIR=$(
    cd "$(dirname "$0")"
    pwd
)
RUNTIME_SRC_DIR="${BASE_DIR}/../"
YR_DATASYSTEM_BIN_DIR="${RUNTIME_SRC_DIR}/.."
YR_METRICS_BIN_DIR="${RUNTIME_SRC_DIR}/../metrics"
THIRD_PARTY_DIR="${RUNTIME_SRC_DIR}/../thirdparty/"
MODULES="runtime"
bash -x ${BASE_DIR}/download_opensource.sh -M $MODULES -T $THIRD_PARTY_DIR
RUNTIME_THIRD_PARTY_CACHE=${RUNTIME_THIRD_PARTY_CACHE:-"https://build-logs.openeuler.openatom.cn:38080/temp-archived/openeuler/openYuanrong/runtime_deps/"}
DATA_SYSTEM_CACHE=${DATA_SYSTEM_CACHE:-"https://build-logs.openeuler.openatom.cn:38080/temp-archived/openeuler/openYuanrong/yr_cache/$(uname -m)/yr-datasystem-v0.5.0.tar.gz"}
FUNCTION_SYSTEM_CACHE=${FUNCTION_SYSTEM_CACHE:-"https://build-logs.openeuler.openatom.cn:38080/temp-archived/openeuler/openYuanrong/yr_cache/$(uname -m)/yr-functionsystem-v0.5.0.tar.gz"}
function check_datasystem() {
    # check whether datasystem exist
    if [ ! -d "${YR_DATASYSTEM_BIN_DIR}"/datasystem/output/sdk/cpp/include ]; then
        echo "datasystem sdk not exist!"
        exit 1
    fi
}

function check_metrics(){
    # check whether metrics exit
    if [ ! -d "${YR_METRICS_BIN_DIR}"/lib ]; then
        echo "metrics sdk not exist!"
        exit 1
    fi
}

function download_datasystem() {
    # check whether datasystem exist
    if [ -d "${YR_DATASYSTEM_BIN_DIR}"/datasystem/output/sdk/cpp/include ]; then
        echo "datasystem sdk exist."
        return
    fi
    DS_OUT_DIR="${YR_DATASYSTEM_BIN_DIR}/datasystem/output"
    mkdir -p "${DS_OUT_DIR}"
    pushd "${DS_OUT_DIR}"
    curl -S -o datasystem.tar.gz ${DATA_SYSTEM_CACHE}
    tar --no-same-owner -zxf datasystem.tar.gz
    popd
}

function download_metrics() {
    # check whether datasystem exist
    if [ -d "${YR_METRICS_BIN_DIR}"/lib ]; then
        echo "datasystem sdk exist."
        return
    fi
    METRICS_OUT_DIR="${YR_METRICS_BIN_DIR}/"
    mkdir -p "${METRICS_OUT_DIR}"
    pushd "${METRICS_OUT_DIR}"
    curl -S -o functionsystem.tar.gz ${FUNCTION_SYSTEM_CACHE}
    tar --no-same-owner -zxf functionsystem.tar.gz
    mv function_system/metrics/* .
    rm -rf function_system
    popd
}

function compile_all(){
  if [ ! -d "${THIRD_PARTY_DIR}/boost/lib" ]; then
    pushd "${THIRD_PARTY_DIR}/boost/"
    chmod -R 700 "${THIRD_PARTY_DIR}/boost/"
    ./bootstrap.sh --without-libraries=python
    ./b2 cxxflags=-fPIC cflags=-fPIC  link=static install --with-fiber --prefix=${THIRD_PARTY_DIR}/boost
    popd
  fi
  if [ ! -d "${THIRD_PARTY_DIR}/openssl/install" ]; then
    pushd "${THIRD_PARTY_DIR}/openssl/"
    chmod -R 700 "${THIRD_PARTY_DIR}/openssl/"
    ./config enable-ssl3 enable-ssl3-method --prefix="${THIRD_PARTY_DIR}/openssl/install"
    make -j build_libs install_dev
    popd
  fi
}

function download_third_party_cache() {
    if [ -n "${RUNTIME_THIRD_PARTY_CACHE}" ]; then
      wget -r -np -nH --no-directories ${RUNTIME_THIRD_PARTY_CACHE}
    fi
}

function download_cache() {
    if [ -d "${RUNTIME_SRC_DIR}/thirdparty/runtime_deps" ]; then
        echo "third party cache exist."
        return
    fi
    CACHE_OUT_DIR="${RUNTIME_SRC_DIR}/thirdparty/runtime_deps"
    mkdir -p "${CACHE_OUT_DIR}"
    pushd "${CACHE_OUT_DIR}"
    download_third_party_cache
    popd
}

download_datasystem
download_metrics
download_cache
check_datasystem
check_metrics
compile_all