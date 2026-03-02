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

CUR_DIR=$(dirname "$(readlink -f "$0")")
PROJECT_DIR=$(cd "${CUR_DIR}/.."; pwd)
ROOT_PATH=${PROJECT_DIR}
PROJECT_OUTPUT_DIR="${PROJECT_DIR}/output"
POSIX_DIR="${PROJECT_DIR}/proto/posix"
CLIENT_DIR="${PROJECT_DIR}/pkg/dashboard/client"

# go module prepare
export GO111MODULE=on
export GONOSUMDB=*
export CGO_ENABLED=1
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/axw/gocov/gocov@latest
go install github.com/matm/gocov-html/cmd/gocov-html@latest

# resolve missing go.sum entry
go env -w "GOFLAGS"="-mod=mod"

# coverage mode
# set: 每个语句是否执行？
# count: 每个语句执行了几次？
# atomic: 类似于count, 但表示的是并行程序中的精确计数
export GOCOVER_MODE="set"

# protoc
echo "generating fs proto pb objects"
mkdir -p "${PROJECT_OUTPUT_DIR}"
protoc --proto_path=${POSIX_DIR} --go_out=${PROJECT_OUTPUT_DIR} --go-grpc_out=${PROJECT_OUTPUT_DIR} ${POSIX_DIR}/*.proto
cp -ar ${PROJECT_OUTPUT_DIR}/yuanrong.org/kernel/pkg/ ${PROJECT_DIR}
rm -rf "${PROJECT_OUTPUT_DIR}/yuanrong.org"

# dashboard test
sh "${CUR_DIR}/dashboard/test.sh"

# dashboard client test
cd "${CLIENT_DIR}"
npm i
CLIENT_CONFIG_DIR="${CLIENT_DIR}/src/config"
rm -f "${CLIENT_CONFIG_DIR}/config.json"
cp "${CLIENT_CONFIG_DIR}/config.json.bak" "${CLIENT_CONFIG_DIR}/config.json"
npm run coverage

# collector test
sh "${CUR_DIR}/collector/test.sh"

# common test
sh "${CUR_DIR}/common/test.sh"