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
source /etc/profile.d/*.sh

readonly SCRIPT_NAME=$(basename "$0")

show_help() {
cat << EOF
Usage: $SCRIPT_NAME -v VERSION [-P] [-h]

Configure and build openYuanrong documentation.

Options:
  -v VERSION    (Optional) Specify the version string. Defaults to "latest".
  -P            (Optional) Use the installed package instead of building the runtime from source.
  -h            Display this help message and exit.

Examples:
  ./$SCRIPT_NAME -v 1.0.0
  ./$SCRIPT_NAME -v 1.0.0 -P
EOF
}

BUILD_VERSION="latest"
BUILD_WITH_PACKAGE="false"

while getopts "hv:P" opt; do
  case $opt in
    h)
      show_help
      exit 0
      ;;
    v)
      BUILD_VERSION="$OPTARG"
      ;;
    P)
      BUILD_WITH_PACKAGE="true"
      ;;
    \? | :)
      echo "Try '$SCRIPT_NAME -h' for more information." >&2
      exit 1
      ;;
  esac
done

# Export environment variables
export BUILD_VERSION
export BUILD_WITH_PACKAGE

BASE_DIR=$(dirname "$(readlink -f "$0")")
OUTPUT_DIR=${BASE_DIR}/../output
function doc_build() {
  pip install -r ${BASE_DIR}/../docs/requirements_dev.txt

  pushd ${BASE_DIR}
  make html
  # disable configuration：SPHINXOPTS="-W --keep-going -n", enable it after all alarms are cleared.
  popd

  # modify sphinx(7.3.7) build-in search，open limit on numbers in search.
  # Changes in later versions need to be modified accordingly.
  sed -i '285d' "${BASE_DIR}"/_build/html/_static/searchtools.js
  sed -i '284s/ ||//' "${BASE_DIR}"/_build/html/_static/searchtools.js
  rm -rf "${OUTPUT_DIR}"/docs && mkdir -p "${OUTPUT_DIR}"/docs
  cp -rf "${BASE_DIR}"/_build/html/* "${OUTPUT_DIR}"/docs
}

doc_build