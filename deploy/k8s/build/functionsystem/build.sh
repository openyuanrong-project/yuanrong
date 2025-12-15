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

cur=$(dirname "$(readlink -f $0)")

set -e
BUILD_LOCAL="${BUILD_LOCAL:-"true"}"

SN_ID=1002
SNUSER_ID=1003

source ${cur}/common.sh $1
project_dir=${cur}../../../
var_pipeline_version=$2
var_image_version=$2
echo "build with image_version='${var_image_version}' pipeline_version='${var_pipeline_version}'"

var_image_prefix=""
var_service_image_prefix=""

var_base_image_full="openeuler-22.03-lts-sp4:latest"
var_service_image_full="${var_service_image_prefix}/${var_service_image_name}:${var_image_version}"


cd ${cur}

docker build --no-cache --build-arg docker_image=${var_base_image_full} \
                        --build-arg app=${app} \
                        --build-arg SNUSER_ID="${SNUSER_ID}" \
                        --build-arg SN_ID="${SN_ID}" \
                        -t ${var_service_image_full} -f ${cur}/dockerfile/Dockerfile-${app} .

echo "show docker images after build"
docker images "${var_service_image_full}" || true
docker inspect ${var_service_image_full} || true
write_buildInfo ${var_pipeline_version} ${var_service_image_full}
write_buildImage ${cur} ${var_base_image_full}
