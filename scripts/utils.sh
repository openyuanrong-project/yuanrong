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
set -e

log_info() {
    echo "[INFO][$(date +%b\ %d\ %H:%M:%S)] $*"
}

log_warning() {
    echo "[WARNING][$(date +%b\ %d\ %H:%M:%S)] $*"
}

log_error() {
    echo "[ERROR][$(date +%b\ %d\ %H:%M:%S)] $*"
}

die() {
    log_error "$*"
    stty echo
    exit 1
}

function make_combined_yuanrong_package() {
    # path to openyuanrong.tar.gz
    PREBUILD_BIN_PATH_YUANRONG=$1
    # path of the output
    COMBINE_OUTPUT_DIR=$2
    # path of the temperaray building dir, normally should be /..../tmpbuild/
    # in case there are some middle product, default to /tmp/
    TEMP_BUILD_CACHE_DIR=${3:-/tmp/buildcache}

    log_info "clear output(${COMBINE_OUTPUT_DIR}) and cache(${TEMP_BUILD_CACHE_DIR}) before start make combined yuanrong package"
    [[ -n "${COMBINE_OUTPUT_DIR}" ]] && rm -rf "${COMBINE_OUTPUT_DIR}"
    [[ -n "${TEMP_BUILD_CACHE_DIR}" ]] && rm -rf "${TEMP_BUILD_CACHE_DIR}"

    mkdir -p ${COMBINE_OUTPUT_DIR}
    if [[ $(basename "$PREBUILD_BIN_PATH_YUANRONG") == *.tar.gz ]]; then
        log_info "extract prebuild openyuanrong.tar.gz to ${COMBINE_OUTPUT_DIR}"
        tar --strip-components=1 -zxf ${PREBUILD_BIN_PATH_YUANRONG} -C ${COMBINE_OUTPUT_DIR}
    else
        log_info "copy prebuild yuanrong to ${COMBINE_OUTPUT_DIR}"
        cp -rf ${PREBUILD_BIN_PATH_YUANRONG}/* ${COMBINE_OUTPUT_DIR}/
    fi

    log_info "extract prebuild tools to ${TEMP_BUILD_CACHE_DIR}"
    mkdir -p ${TEMP_BUILD_CACHE_DIR}/cli/bin/
    cp -rf ${PREBUILD_BIN_PATH_YUANRONG}/functionsystem/bin/yr ${TEMP_BUILD_CACHE_DIR}/cli/bin/
    log_info "copy prebuild tools to ${COMBINE_OUTPUT_DIR}"
    cp -rf ${TEMP_BUILD_CACHE_DIR}/cli ${COMBINE_OUTPUT_DIR}

    log_info "make combined yuanrong package succeed at ${COMBINE_OUTPUT_DIR}"
}

function make_valid_semver() {
    local tag=$1
    # lower
    local tag_lower=$(echo "$tag" | awk '{print tolower($0)}')
    if [[ "$tag_lower" == *spc* ]]; then
      VERSION=$(echo $tag_lower | awk -F\. '{print $1 "." $2 "." $3 "+" $4 $5}')
    else
      VERSION=$(echo $tag_lower | awk -F\. '{print $1 "." $2 "." $3 "+" $4}')
    fi

    # default arch is always x86_64
    local arch=${2:-x86_64}
    if [[ "${arch}" != "x86_64" ]]; then
        echo "${VERSION}-${arch}"
    else
        echo "${VERSION}"
    fi
}

