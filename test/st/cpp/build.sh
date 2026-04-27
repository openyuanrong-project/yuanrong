#!/usr/bin/env bash
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

SANITIZER="off"
BASE_DIR=$(
    cd "$(dirname "$0")"
    pwd
)
. "${BASE_DIR}/../utils.sh"
YUANRONG_DIR="${BASE_DIR}/../../../output/yuanrong"
BUILD_DIR="${BASE_DIR}/build"
readonly USAGE="
Usage: bash build.sh [-h] [-y path-to-yuanrong] [-S thread/address/off]

Options:
    -h Output this help and exit.
    -y YuanRong dir.
    -S Use Google Sanitizers tools to detect bugs. Choose from off/address/thread,
                   if set the value to 'address' enable AddressSanitizer,
                   if set the value to 'thread' enable ThreadSanitizer,
                   default off.
Example:
  $ bash build.sh
"

usage() {
    echo -e "$USAGE"
}

get_opt() {
    while getopts 'y:h:S:' opt; do
    case "$opt" in
    y)
        YUANRONG_DIR="$(readlink -f ${OPTARG})"
        ;;
    h)
        usage
        exit 0
        ;;
    S)
        check_sanitizers "${OPTARG}"
        SANITIZER="${OPTARG}"
        ;;
    *)
        echo "invalid command"
        usage
        exit 1
        ;;
    esac
done
}

main() {
    get_opt "$@"
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR" && cd "$BUILD_DIR"
    cmake "$BASE_DIR" -DSDK_DIR="$YUANRONG_DIR/runtime/sdk/cpp/" -DSANITIZER="${SANITIZER}"
    make -j
}

main "$@"