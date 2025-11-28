#!/bin/bash
set -e
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
# Requirements.

BASE_DIR=$(
    cd "$(dirname "$0")"
    pwd
)
OUTPUT_DIR="${BASE_DIR}/../../../output"
. ${BASE_DIR}/../utils.sh

PYPI_BUILD_DIR=${BASE_DIR}/../.build/pypibuild

# pypi src dir
PYPI_SRC_DIR=${BASE_DIR}/src

# yuanrong binary path, default under OUTPUT_DIR
PREBUILD_BIN_PATH_YUANRONG=${OUTPUT_DIR}/openyuanrong.tar.gz

# whether show help info
SHOW_HELP=""
PYTHON_ONLY="false"
CURRENT_ARCH=$(uname -m)

# default tag
BUILD_VERSION="v0.0.1"
YR_REQUIREMENTS=""
PYTHON_BIN_PATH="python3"

function setup_workspace () {
    log_info "setting up pypi workspace"

    [[ -n "${PYPI_BUILD_DIR}" ]] && rm -rf "${PYPI_BUILD_DIR}"

    make_combined_yuanrong_package \
        ${PREBUILD_BIN_PATH_YUANRONG} \
        ${PYPI_BUILD_DIR}/src/yr/inner \
        /tmp/pypi_temp

    if [ "$PYTHON_ONLY" == "true" ]; then
        rm -rf ${PYPI_BUILD_DIR}/src/yr/inner/runtime/service/cpp
        rm -rf ${PYPI_BUILD_DIR}/src/yr/inner/runtime/service/java
        rm -rf ${PYPI_BUILD_DIR}/src/yr/inner/runtime/sdk/
        rm -rf ${PYPI_BUILD_DIR}/src/yr/inner/alias
        rm -rf ${PYPI_BUILD_DIR}/src/yr/inner/initcontainer
        # keep cpp/include directory
        rm -rf ${PYPI_BUILD_DIR}/src/yr/inner/datasystem/sdk/*.jar
        rm -rf ${PYPI_BUILD_DIR}/src/yr/inner/datasystem/sdk/*.whl
        rm -rf ${PYPI_BUILD_DIR}/src/yr/inner/datasystem/sdk/cpp/lib
    fi

    log_info "copy whl skeleton code to pypi workspace"
    cp -rf ${PYPI_SRC_DIR} ${PYPI_BUILD_DIR}/

    log_info "unpack yr whl and copy into our build dir"
    whl_name="yr_sdk-*.whl"
    PYTHON_VERSION=$(${PYTHON_BIN_PATH} --version 2>&1)
    if [[ "${PYTHON_VERSION}" =~ ([0-9]+\.[0-9]+) ]]; then
        # Extract major.minor version and remove dot (e.g., 3.11 → 311)
        ver="${BASH_REMATCH[1]//./}"
        whl_name="yr_sdk*-cp${ver}-cp${ver}*.whl"
        short_ver=$(echo "$PYTHON_VERSION" | grep -Eo '[0-9]+\.[0-9]+')
        sed -i "s/[Pp]ython[0-9.]\+/python$short_ver/" \
        ${PYPI_BUILD_DIR}/src/yr/inner/deploy/process/services.yaml
    fi

    if [[ ${PYTHON_VERSION} =~ "Python 3.11" ]]; then
        find ${PREBUILD_BIN_PATH_YUANRONG} -name ${whl_name} | xargs -I {} unzip {} -d /tmp/pypi_temp/yr/
        cp -rf /tmp/pypi_temp/yr/yr/* ${PYPI_BUILD_DIR}/src/yr/
    else
        find ${PREBUILD_BIN_PATH_YUANRONG} -name ${whl_name} | xargs -I {} unzip {} -d /tmp/pypi_temp/yr/
        cp -rf /tmp/pypi_temp/yr/yr_sdk*.data/purelib/yr/* ${PYPI_BUILD_DIR}/src/yr/
    fi

    YR_REQUIREMENTS=$(find /tmp/pypi_temp/yr/yr*.dist-info -name "METADATA" | xargs -I {} grep "Requires-Dist" {} \
                      | grep -v "extra ==" | awk -F ': ' '{print $2}')
}

function build_pypi () {
    log_info "building pypi package..."

    log_info "find the glibc requirements"
    glibc_version=$(ldd --version | grep -Po "\d+\.\d+" | awk -F. '{print $1 "_" $2}' | head -n1)
    log_info "USE the glibc requirements == $glibc_version"

    # build
    pushd ${PYPI_BUILD_DIR}/src/
    YR_BUILD_VERSION="${BUILD_VERSION}" \
    YR_REQUIREMENTS="${YR_REQUIREMENTS}" \
        ${PYTHON_BIN_PATH} setup.py bdist_wheel --dist-dir ${OUTPUT_DIR} \
        --plat-name "manylinux_${glibc_version}_${CURRENT_ARCH}"
    popd
}

function parse_args () {
    getopt_cmd=$(getopt -o o:v:h -l output_dir:,version:,pypi_build_dir:,prebuild_yuanrong:,arch:,python_only:,python_bin_path:,help -- "$@")
    [ $? -ne 0 ] && exit 1
    eval set -- "$getopt_cmd"
    while true; do
        case "$1" in
        -o|--output_dir) OUTPUT_DIR=$2 && shift 2 ;;
        -v|--version) BUILD_VERSION=$2 && shift 2 ;;
        --pypi_build_dir) PYPI_BUILD_DIR=$2 && shift 2 ;;
        --prebuild_yuanrong) PREBUILD_BIN_PATH_YUANRONG=$2 && shift 2 ;;
        --arch) CURRENT_ARCH=$2 && shift 2 ;;
        --python_only) PYTHON_ONLY=$2 && shift 2 ;;
        --python_bin_path) PYTHON_BIN_PATH=$2 && shift 2 ;;
        -h|--help) SHOW_HELP="true" && shift ;;
        --) shift && break ;;
        *) die "Invalid option: $1" && exit 1 ;;
        esac
    done

    if [ "$SHOW_HELP" != "" ]; then
        cat <<EOF
Usage:
  packaging rpm packages, args and default values:
    -o|--output_dir      the output dir (=${OUTPUT_DIR})
    -b|--pypi_build_dir  build tree dir (=${PYPI_BUILD_DIR})
    -v|--version             the version and release (=${BUILD_VERSION})
    --prebuild_yuanrong  the prebuild yuanrong.tar.gz path (=${PREBUILD_BIN_PATH_YUANRONG})
    --python_only        the trim pacakge (=${PYTHON_ONLY})
    --python_bin_path    the python path (=${PYTHON_BIN_PATH})
    -h|--help            show this help info
EOF
        exit 1
    fi
}

function main () {
    parse_args "$@"
    setup_workspace
    build_pypi

    [[ -n "${PYPI_BUILD_DIR}" ]] && rm -rf "${PYPI_BUILD_DIR}"
}


main $@
