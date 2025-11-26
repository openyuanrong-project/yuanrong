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

readonly USAGE="
Usage: bash build.sh [-h] [-v]

Options:
    -v set version.
    -h show usage.
"

PROJECT_DIR=$(cd "$(dirname "$0")"; pwd)
OUTPUT_DIR="${PROJECT_DIR}/output"
RUNTIME_OUTPUT_DIR="${PROJECT_DIR}/../output"
POSIX_DIR="${PROJECT_DIR}/proto/posix"
BUILD_TAGS=""
VERSION="0.5.0"
FLAGS='-extldflags "-fPIC -fstack-protector-strong -Wl,-z,now,-z,relro,-z,noexecstack,-s -Wall -Werror"'

while getopts "v:h" opt; do
    case $opt in
        v)
            VERSION="$OPTARG"
            ;;
        h)
            echo -e "${USAGE}"
            exit 0
            ;;
        *)
            echo "Invalid command"
            exit 1
            ;;
    esac
done

# go module prepare
export GO111MODULE=on
export GONOSUMDB=*
export CGO_ENABLED=1
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# resolve missing go.sum entry
go env -w "GOFLAGS"="-mod=mod"

echo "generating fs proto pb objects"
mkdir -p "${OUTPUT_DIR}"
protoc --proto_path=${POSIX_DIR} --go_out=${OUTPUT_DIR} --go-grpc_out=${OUTPUT_DIR} ${POSIX_DIR}/*.proto
cp -ar ${OUTPUT_DIR}/yuanrong.org/kernel/pkg/ ${PROJECT_DIR}
rm -rf "${OUTPUT_DIR}/yuanrong.org"

echo "start to compile dashboard -s ${SCC_BUILD_ENABLED}"
mkdir -p "${OUTPUT_DIR}/bin/"
rm -rf "${OUTPUT_DIR}/bin/dashboard"

DASHBOARD_FLAGS="${FLAGS} -X 'yuanrong.org/kernel/pkg/dashboard/flags.version=${VERSION}'"
CC='gcc -fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2' go build -tags="${BUILD_TAGS}" -buildmode=pie -ldflags "${DASHBOARD_FLAGS}"  -o \
"${OUTPUT_DIR}"/bin/dashboard "${PROJECT_DIR}"/cmd/dashboard/main.go

mkdir -p "${OUTPUT_DIR}/config/"
rm -rf "${OUTPUT_DIR}/config/dashboard*"
cp -ar "${PROJECT_DIR}/build/dashboard/config/" "${OUTPUT_DIR}/"

npm config set strict-ssl false
echo "start to compile dashboard client"
mkdir -p "${OUTPUT_DIR}/bin/"
rm -rf "${OUTPUT_DIR}/bin/client"
cd "${PROJECT_DIR}/pkg/dashboard/client"
CLIENT_CONFIG_DIR="${PROJECT_DIR}/pkg/dashboard/client/src/config"
rm -f "${CLIENT_CONFIG_DIR}/config.json"
cp "${CLIENT_CONFIG_DIR}/config.json.bak" "${CLIENT_CONFIG_DIR}/config.json"
sed -i "s#{version}#${VERSION}#g" "${CLIENT_CONFIG_DIR}/config.json"
npm install || die "dashboard client install failed"
npm run build || die "dashboard client build failed"
mkdir -p "${OUTPUT_DIR}/bin/client"
cp -ar ./dist "${OUTPUT_DIR}/bin/client/"
cd "${PROJECT_DIR}"

echo "start to compile collector"
mkdir -p "${OUTPUT_DIR}/bin/"
rm -rf "${OUTPUT_DIR}/bin/collector"
echo LD_LIBRARY_PATH=$LD_LIBRARY_PATH
COLLECTOR_FLAGS="${FLAGS} -X 'yuanrong.org/kernel/pkg/collector/common.version=${VERSION}'"
CC='gcc -fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2' go build -tags="${BUILD_TAGS}" -buildmode=pie -ldflags "${COLLECTOR_FLAGS}"  -o \
"${OUTPUT_DIR}"/bin/collector "${PROJECT_DIR}"/cmd/collector/main.go

cd "${OUTPUT_DIR}"
DASHBOARD_TAR_NAME="yr-dashboard-${VERSION}.tar.gz"
tar -czvf "${DASHBOARD_TAR_NAME}" ./*
mkdir -p "${RUNTIME_OUTPUT_DIR}"
rm -rf "${RUNTIME_OUTPUT_DIR}/${DASHBOARD_TAR_NAME}"
cp "${DASHBOARD_TAR_NAME}" "${RUNTIME_OUTPUT_DIR}"
cd "${PROJECT_DIR}"