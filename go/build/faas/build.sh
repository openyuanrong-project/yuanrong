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

BASE_DIR=$(cd "$(dirname "$0")"; pwd)
PROJECT_DIR=$(cd "$(dirname "$0")"/../..; pwd)
RUNTIME_OUTPUT_DIR="${PROJECT_DIR}/../output"
OUTPUT_DIR="${BASE_DIR}/../../output/pattern/pattern_faas"
TAR_OUT_DIR="${PROJECT_DIR}/build/_output"
TAR_OUT_FILE="faasfunctions.tar.gz"
EXECUTOR_DIR="${PROJECT_DIR}/build/faas/executor-meta"
TEST_CERT_PATH="${GOROOT}/src/net/http/internal/testcert.go"
BUILD_TAG_FUNCTION="function"
echo LD_LIBRARY_PATH=$LD_LIBRARY_PATH
MODULE_NAME_LIST=("faasscheduler" "faasmanager")
BUILD_VERSION=v0.5.0
# go module prepare
export GO111MODULE=on
export GONOSUMDB=*
export CGO_ENABLED=1
mkdir -p ${OUTPUT_DIR}
# resolve missing go.sum entry
go env -w "GOFLAGS"="-mod=mod"
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# remove hard coded cert file in net/http
[ -f "${TEST_CERT_PATH}" ] && rm -f "${TEST_CERT_PATH}"

function parse_args () {
    getopt_cmd=$(getopt -o v:h -l help -- "$@")
    [ $? -ne 0 ] && exit 1
    eval set -- "$getopt_cmd"
    while true; do
        case "$1" in
        -v|--version) BUILD_VERSION=$2 && shift 2 ;;
        -h|--help) SHOW_HELP="true" && shift ;;
        --) shift && break ;;
        *) die "Invalid option: $1" && exit 1 ;;
        esac
    done
    if [ "$SHOW_HELP" != "" ]; then
        cat <<EOF
Usage:
    -v|--version             the version (=${BUILD_VERSION})
    -h|--help                show this help info
EOF
        exit 1
    fi
}

function main () {
    parse_args "$@"
}

main $@

# clean build history
bash "${BASE_DIR}"/clean.sh

cd "${PROJECT_DIR}"
. "${BASE_DIR}"/compile_functions.sh

# zip function file
FLAGS='-extldflags "-fPIC -fstack-protector-strong -Wl,-z,now,-z,relro,-z,noexecstack,-s -Wall -Werror"'
export CGO_CFLAGS="$CGO_CFLAGS -fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2"
MODULE_NAME="faasscheduler"
for MODULE_NAME in "${MODULE_NAME_LIST[@]}"
do
    cd "${TAR_OUT_DIR}"
    mkdir -p "${MODULE_NAME}"
    SO_PATH="${TAR_OUT_DIR}/${MODULE_NAME}/${MODULE_NAME}.so"
    BIN_PATH="${TAR_OUT_DIR}/${MODULE_NAME}/${MODULE_NAME}"
    cd "${PROJECT_DIR}"
    if [ "${MODULE_NAME}" = "faasscheduler" ]; then
        CC='gcc -fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2'
        go build -buildmode=pie -ldflags "-linkmode=external ${FLAGS}" \
        -o ${BIN_PATH} "${PROJECT_DIR}/cmd/faas/${MODULE_NAME}/module_main.go"

        CC='gcc -fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2'
        go build -tags "${BUILD_TAG_FUNCTION}" -buildmode=plugin -ldflags "${FLAGS}" \
        -o ${SO_PATH} "${PROJECT_DIR}/cmd/faas/${MODULE_NAME}/function_main.go"
        chmod -R 500 ${SO_PATH}
        cd "${TAR_OUT_DIR}/${MODULE_NAME}"
        zip -r "${MODULE_NAME}.zip" *
        continue
    fi

    if [ "${MODULE_NAME}" = "faasmanager" ]; then
        CC='gcc -fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2'
        go build -tags "${BUILD_TAG_FUNCTION}" -buildmode=plugin -ldflags "${FLAGS}" \
        -o ${SO_PATH} "${PROJECT_DIR}/cmd/faas/${MODULE_NAME}/main.go"
        chmod -R 500 ${SO_PATH}
        cd "${TAR_OUT_DIR}/${MODULE_NAME}"
        zip -r "${MODULE_NAME}.zip" *
        continue
    fi
done
for MODULE_NAME in "${MODULE_NAME_LIST[@]}"
do      
    cp "${PROJECT_DIR}/build/faas/function_meta.json" "${TAR_OUT_DIR}/${MODULE_NAME}/${MODULE_NAME}_meta.json"
    sed -i "s/moduleName/${MODULE_NAME}/g" "${TAR_OUT_DIR}/${MODULE_NAME}/${MODULE_NAME}_meta.json"
    code_size_line=11
    code_size_old=0
    code_size_new=$(stat --format=%s "${TAR_OUT_DIR}/${MODULE_NAME}/${MODULE_NAME}.zip")
    sed -i "${code_size_line} s@${code_size_old}@${code_size_new}@" "${TAR_OUT_DIR}/${MODULE_NAME}/${MODULE_NAME}_meta.json"
done

# get the final tar package.
chmod -R 700 "${TAR_OUT_DIR}"

cp -ar "${TAR_OUT_DIR}/"* "${OUTPUT_DIR}"
mkdir -p "$OUTPUT_DIR/templates/"
cp -arf "${PROJECT_DIR}/build/faas/faasscheduler-function-config.yaml" "${OUTPUT_DIR}/templates/faasscheduler-function-config.yaml"
cp -arf "${PROJECT_DIR}/build/faas/faasscheduler-function-meta.yaml" "${OUTPUT_DIR}/templates/faasscheduler-function-meta.yaml"
cp -arf "${PROJECT_DIR}/build/faas/faasmanager-function-meta.yaml" "${OUTPUT_DIR}/templates/faasmanager-function-meta.yaml"
cp -arf "${PROJECT_DIR}/build/faas/faasmanager-function-config.yaml" "${OUTPUT_DIR}/templates/faasmanager-function-config.yaml"
cp -arf "${PROJECT_DIR}/build/faas/init_scheduler_args.json" "${OUTPUT_DIR}"
chmod -R 700 "${OUTPUT_DIR}"
cd ${OUTPUT_DIR}/../../
tar -czvf yr-faas-${BUILD_VERSION}.tar.gz pattern
mkdir -p "${RUNTIME_OUTPUT_DIR}"
cp -f yr-faas-${BUILD_VERSION}.tar.gz ${RUNTIME_OUTPUT_DIR}