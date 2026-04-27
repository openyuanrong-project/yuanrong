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

app=$1

microServiceName=YuanRong

var_service_image_name=${app}
echo "build app is ${app}"
echo "build microServiceName is ${microServiceName}"

arch_x86="x86"
arch_aarch64="aarch64"

result=$(uname -a)

if [[ $result =~ $arch_x86 ]]; then
  architecture="x86"
elif [[ $result =~ $arch_aarch64 ]]; then
  architecture="arm"
else
  result=$(arch)
  if [[ $result =~ $arch_x86 ]]; then
    architecture="x86"
  elif [[ $result =~ $arch_aarch64 ]]; then
    architecture="arm"
  fi
fi
echo "build architecture is $architecture"

read_version_file() {
  local filename="${cur}/../version.txt"
  version=$(cat ${filename})
  version_number=${version#*=}
  echo ${version_number}
}

datetime=$(date +"%Y%m%d%H%M%S")
get_timestamp() {
  if [[ ${pipelineStartTime} == "" ]]; then
    echo "${datetime}"
  else
    echo ${pipelineStartTime}
  fi
}

get_version() {
  local pipeline_version=""
  local image_version=""
  local timestamp=$(get_timestamp)
  local version_timestamp="$(read_version_file).${timestamp}"

  if [[ ${BuildType} == "" ]]; then
    pipeline_version="${version_timestamp}"
    image_version="${version_timestamp}"
  else
    # 这里是 dailybuild
    pipeline_version="${BuildType}.${timestamp}"
    image_version="${BuildType}"
  fi

   echo ${pipeline_version} ${image_version}
}

write_buildInfo() {
  # meta信息写入buildInfo.properties环境变量
  # 细则1: 版本可以不一致
  #       buildVersion 是给流水线用的
  #       imageName 是给推镜像用的
  #       这两个地方写的版本可以不一致
  # 细则2:如果写imageInfos='<json array>'必须要求 build.yml 文件里搭配 xxx_batch_xx.yml的模板使用
  cat <<EOF >${WORKSPACE}/buildInfo.properties
buildVersion=$1
imageName=$2
orgName=yuanrong
scopeName=yuanrong
EOF
  echo "show file ${WORKSPACE}/buildInfo.properties"
  cat ${WORKSPACE}/buildInfo.properties
}

write_buildImage() {
  # 镜像可溯源
  cat <<EOF >${WORKSPACE}/buildImage.properties
dockerfile=$1/dockerfile/Dockerfile-${app}
docker_var=$2
EOF
  echo "show file ${WORKSPACE}/buildImage.properties"
  cat ${WORKSPACE}/buildImage.properties
}
