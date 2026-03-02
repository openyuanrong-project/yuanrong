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

function generate_pb() {
    # generate pb files
    if [ -z "${GOPATH}" ] || [ ! -d "${GOPATH}" ]; then
        log_error "GOPATH ${GOPATH} not exist!"
        return 1
    fi
    cd "${PROJECT_DIR}"/pkg
    [ -d "${GOPATH}/src/faas" ] && rm -rf "${GOPATH}/src/faas"
    mkdir -p "${GOPATH}"/src/
    ln -s "${PROJECT_DIR}" "${GOPATH}"/src/faas
    if ! bash "${PROJECT_DIR}"/build/faas/gen_grpc_pb.sh; then
        log_error "Failed to generate pb files!"
        return 1
    fi
    cp -afr ${GOPATH}/src/yuanrong.org/kernel/pkg/ ${GOPATH}/src/faas/
}

generate_pb
go mod tidy