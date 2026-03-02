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
BUILD_VERSION=v0.0.1
BASE_DIR=$(
  cd "$(dirname "$0")"
  pwd
)
. ${BASE_DIR}/package/utils.sh
OUTPUT_DIR="${BASE_DIR}/../output"
function parse_args () {
    getopt_cmd=$(getopt -o v:h -l version:,python_bin_path:,help -- "$@")
    [ $? -ne 0 ] && exit 1
    eval set -- "$getopt_cmd"
    while true; do
        case "$1" in
        -h|--help) SHOW_HELP="true" && shift ;;
        --python_bin_path) PYTHON_BIN_PATH=$2 && shift 2 ;;
        -v|--version) BUILD_VERSION=$2 && shift 2 ;;
        --) shift && break ;;
        *) die "Invalid option: $1" && exit 1 ;;
        esac
    done

    if [ "$SHOW_HELP" != "" ]; then
        cat <<EOF
Usage:
  packaging rpm packages, args and default values:
    -v|--version             the version (=${BUILD_VERSION})
    -h|--help            show this help info
EOF
        exit 1
    fi
}

function get_all(){
  if [ -n "${FUNCTION_SYSTEM_CACHE}" ]; then
      echo "download functionsystem"
      fs_filename=$(ls *functionsystem*.tar.gz)
      if [ ! -n "${fs_filename}" ]; then
        curl -SO ${FUNCTION_SYSTEM_CACHE}
      fi
  fi
  if [ -n "${DATA_SYSTEM_CACHE}" ]; then
      echo "download datasystem"
      ds_filename=$(ls *datasystem*.tar.gz)
      if [ ! -n "${ds_filename}" ]; then
        curl -SO ${DATA_SYSTEM_CACHE}
      fi
  fi
  if [ -n "${FRONTEND_CACHE}" ]; then
      echo "download frontend"
      frontend_filename=$(ls *frontend*.tar.gz)
      if [ ! -n "${frontend_filename}" ]; then
        curl -SO ${FRONTEND_CACHE}
      fi
  fi
  if [ -n "${DASHBOARD_CACHE}" ]; then
      echo "download dashboard"
      dashboard_filename=$(ls *dashboard*.tar.gz)
      if [ ! -n "${dashboard_filename}" ]; then
        curl -SO ${DASHBOARD_CACHE}
      fi
  fi
}

function main () {
    parse_args "$@"
}


main $@
rm -rf ${OUTPUT_DIR}/openyuanrong
mkdir -p ${OUTPUT_DIR}/openyuanrong
cd ${OUTPUT_DIR}

echo "start to download all packages from cache url if set"

get_all

echo "start to package openyuanrong-${BUILD_VERSION}.tar.gz"

tar -zxf yr-runtime-*.tar.gz -C ${OUTPUT_DIR}/openyuanrong
tar -zxf *functionsystem*.tar.gz -C ${OUTPUT_DIR}/openyuanrong

tar -zxf *datasystem*.tar.gz -C ${OUTPUT_DIR}/openyuanrong/
rm -rf ${OUTPUT_DIR}/openyuanrong/datasystem/sdk/DATASYSTEM_SYM
rm -rf ${OUTPUT_DIR}/openyuanrong/datasystem/service/DATASYSTEM_SYM
mkdir -p ${OUTPUT_DIR}/openyuanrong/datasystem/deploy
cp -fr ${BASE_DIR}/../deploy/data_system/* ${OUTPUT_DIR}/openyuanrong/datasystem/deploy/
unzip ${OUTPUT_DIR}/openyuanrong/datasystem/sdk/openyuanrong_datasystem_sdk*.whl -d ${OUTPUT_DIR}/openyuanrong/runtime/service/python/

cp -fr ${BASE_DIR}/../deploy ${OUTPUT_DIR}/openyuanrong
rm -rf ${OUTPUT_DIR}/openyuanrong/deploy/data_system

if [ -d "${BASE_DIR}/../../yuanrong-datasystem" ];then
  mkdir -p ${OUTPUT_DIR}/openyuanrong/deploy/k8s/build/datasystem
  cp -fr ${BASE_DIR}/../../yuanrong-datasystem/k8s/helm_chart/datasystem ${OUTPUT_DIR}/openyuanrong/deploy/k8s/charts/
  cp -fr ${BASE_DIR}/../../yuanrong-datasystem/k8s/docker/* ${OUTPUT_DIR}/openyuanrong/deploy/k8s/build/datasystem/
fi

frontend_filename=$(ls *frontend*.tar.gz)
if [ -n "${frontend_filename}" ]; then
    tar -zxf ${frontend_filename} -C ${OUTPUT_DIR}/openyuanrong
    cp -fr ${OUTPUT_DIR}/openyuanrong/pattern/pattern_faas/init_frontend_args.json ${OUTPUT_DIR}/openyuanrong/functionsystem/config/
fi

faas_filename=$(ls *faas*.tar.gz)
if [ -n "${faas_filename}" ]; then
    tar -zxf ${faas_filename} -C ${OUTPUT_DIR}/openyuanrong
    cp -fr ${OUTPUT_DIR}/openyuanrong/pattern/pattern_faas/init_scheduler_args.json ${OUTPUT_DIR}/openyuanrong/functionsystem/config/
fi

dashboard_filename=$(ls *dashboard*.tar.gz)
if [ -n "${dashboard_filename}" ]; then
    mkdir -p ${OUTPUT_DIR}/openyuanrong/dashboard/
    tar -zxf ${dashboard_filename} -C ${OUTPUT_DIR}/openyuanrong/dashboard/
fi

find . -type d -exec chmod 750 {} \;
find . -type l -exec chmod 777 {} \;
find . -type f -exec chmod 640 {} \;
find . -type d -name bin -exec chmod -R 755 {} \;
find . -type f -name datasystem_worker -exec chmod 755 {} \;
find . -type f -name "etcd*" -exec chmod 550 {} \;
if [ -d ${OUTPUT_DIR}/openyuanrong/deploy/process/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/deploy/process/ -type f -exec chmod 550 {} \;
  find ${OUTPUT_DIR}/openyuanrong/deploy/process/ -type f -name "*.yaml" -exec chmod 640 {} \;
fi

if [ -d ${OUTPUT_DIR}/openyuanrong/datasystem/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/datasystem/ -type f -exec chmod 550 {} \;
fi

mv ${OUTPUT_DIR}/openyuanrong/functionsystem/deploy/third_party ${OUTPUT_DIR}/openyuanrong/
mv ${OUTPUT_DIR}/openyuanrong/functionsystem/deploy/function_system/* ${OUTPUT_DIR}/openyuanrong/functionsystem/deploy/
rm -rf ${OUTPUT_DIR}/openyuanrong/functionsystem/deploy/function_system/
mv ${OUTPUT_DIR}/openyuanrong/functionsystem/deploy/vendor/etcd ${OUTPUT_DIR}/openyuanrong/third_party/
rm -rf ${OUTPUT_DIR}/openyuanrong/functionsystem/deploy/vendor
if [ -d ${OUTPUT_DIR}/openyuanrong/third_party/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/third_party/ -type f -exec chmod 550 {} \;
fi

if [ -d ${OUTPUT_DIR}/openyuanrong/functionsystem/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/functionsystem/ -type f -exec chmod 550 {} \;
fi
if [ -d ${OUTPUT_DIR}/openyuanrong/functionsystem/config/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/functionsystem/config/ -type f -exec chmod 640 {} \;
fi

if [ -d ${OUTPUT_DIR}/openyuanrong/runtime/deploy/process/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/runtime/deploy/process/ -type f -exec chmod 550 {} \;
fi
if [ -d ${OUTPUT_DIR}/openyuanrong/runtime/sdk/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/runtime/sdk/ -type f -exec chmod 550 {} \;
  find ${OUTPUT_DIR}/openyuanrong/runtime/sdk/ -type f -name "*.xml" -exec chmod 640 {} \;
fi
if [ -d ${OUTPUT_DIR}/openyuanrong/runtime/service/java/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/runtime/service/java/ -type f -exec chmod 550 {} \;
  find ${OUTPUT_DIR}/openyuanrong/runtime/service/java/ -type f -name "*.xml" -exec chmod 640 {} \;
fi
if [ -d ${OUTPUT_DIR}/openyuanrong/runtime/service/cpp/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/runtime/service/cpp/ -type f -exec chmod 550 {} \;
fi
if [ -d ${OUTPUT_DIR}/openyuanrong/runtime/service/cpp/config/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/runtime/service/cpp/config/ -type f -exec chmod 640 {} \;
fi
if [ -d ${OUTPUT_DIR}/openyuanrong/runtime/service/python/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/runtime/service/python/ -type f -exec chmod 550 {} \;
fi
if [ -d ${OUTPUT_DIR}/openyuanrong/runtime/service/python/config/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/runtime/service/python/config/ -type f -exec chmod 640 {} \;
fi
if [ -d ${OUTPUT_DIR}/openyuanrong/runtime/service/python/yr/config/ ]; then
  find ${OUTPUT_DIR}/openyuanrong/runtime/service/python/yr/config/ -type f -exec chmod 640 {} \;
fi

cat >${OUTPUT_DIR}/openyuanrong/VERSION <<EOF
"${BUILD_VERSION}"
EOF
[ -d "${OUTPUT_DIR}/openyuanrong/runtime/sdk/cpp" ] && cp -ar ${OUTPUT_DIR}/openyuanrong/VERSION ${OUTPUT_DIR}/openyuanrong/runtime/sdk/cpp/VERSION

tar -zcf openyuanrong-${BUILD_VERSION}.tar.gz openyuanrong
