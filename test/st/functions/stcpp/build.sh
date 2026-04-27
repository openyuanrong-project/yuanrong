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

BASE_DIR=$(cd "$(dirname "$0")"; pwd)
# TODO Change the value to the dependency location of yrlib.
RUNTIME_LIB=""
export LD_LIBRARY_PATH="${LD_LIBRARY_PATH}":"${RUNTIME_LIB}"
PKG_DIR=${BASE_DIR}/../pkg
ZIP_FILE=stcpp.zip

[ ! -d "${PKG_DIR}" ] && mkdir -p ${PKG_DIR}

cp "$BASE_DIR"/../../cpp/build/libuser_common_func.so "$BASE_DIR"
cd "$BASE_DIR"
zip ${ZIP_FILE} libuser_common_func.so
cp ${ZIP_FILE} "${PKG_DIR}"
cd -