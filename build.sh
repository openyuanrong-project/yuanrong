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

readonly USAGE="
Usage: bash build.sh [-thdDcCrvPSbEm:]

Options:
    -t run test.
    -d the directory that contains datasystem dir. default: ../
    -D build with debug mode
    -c coverage
    -C clean build environment
    -v the version of yuanrong
    -p the specified version of python runtime
            if it is "multi", will compile python3.9 and python3.11.
            If it is not specified, python3.9 is compiled by default.
            for example bash build.sh -p multi; bash build.sh -p python3.11
    -P download third party packages, default: true
    -S Use Google Sanitizers tools to detect bugs. Choose from off/address/thread,
           if set the value to 'address' enable AddressSanitizer,
           if set the value to 'thread' enable ThreadSanitizer,
           default off.
    -b Use bep
    -E adapter CI java codecheck
    -m mem limit(MB)
    -h usage.
    -j concurrency limit
"

BASE_DIR=$(
    cd "$(dirname "$0")"
    pwd
)
BUILD_BASE="${BASE_DIR}/build"
OUTPUT_DIR="${BASE_DIR}/output"
OUTPUT_BASE="${BASE_DIR}/build/output"
BAZEL_COMMAND="build"
BUILD_VERSION="v0.0.1"
BAZEL_OPTIONS="--experimental_cc_shared_library=true --verbose_failures --strategy=CcStrip=standalone"
BAZEL_OPTIONS_CONFIG=" --config=release "
BAZEL_TARGETS="//api/cpp:yr_cpp_pkg //api/java:yr_java_pkg //api/python:yr_python_pkg"
BAZEL_PRE_OPTIONS="--output_user_root=${BUILD_BASE} --output_base=${OUTPUT_BASE}"
THIRD_PARTY_DIR="$(dirname "$BASE_DIR")/thirdparty"
PYTHON3_BIN_PATH="python3.9"
PYTHON3_SDK_BIN_PATH=$PYTHON3_BIN_PATH
PYTHON_BAZEL_TARGETS="//api/python:yr_python_pkg"
MULTI_PYTHON_VERSION="true"
SANITIZER="off"
BAZEL_OPTIONS_ENV=""
SECBRELLA_CCE="OFF"
PACKAGE_ALL="false"
LD_LIBRARY_PATH=/opt/buildtools/python3.7/lib:/opt/buildtools/python3.9/lib:/opt/buildtools/python3.11/lib:${LD_LIBRARY_PATH}
if [ ! -d "${THIRD_PARTY_DIR}" ]; then
  mkdir -p "${THIRD_PARTY_DIR}"
fi
function usage() {
    echo -e "$USAGE"
}

log_info() {
    echo "[BUILD_INFO][$(date +%b\ %d\ %H:%M:%S)]$*"
}

log_warning() {
    echo "[BUILD_WARN][$(date +%b\ %d\ %H:%M:%S)]$*"
}

log_error() {
    echo "[BUILD_ERROR][$(date +%b\ %d\ %H:%M:%S)]$*"
}

log_fatal() {
    echo "[BUILD_FATAL][$(date +%b\ %d\ %H:%M:%S)]$*"
    exit 1
}

PYTHON_VERSION_LIST=(\
"python3.11" \
"python3.10"
)

function run_java_coverage_report() {
    rm -rf bazel-testlogs/api/java/liblib_yr_api_sdk/
    rm -rf bazel-testlogs/api/java/libfunction_common/
    rm -rf bazel-testlogs/api/java/yr_runtime/
    bazel $BAZEL_PRE_OPTIONS test $BAZEL_OPTIONS //api/java:java_tests
    execs=$(find bazel-out/k8-opt/bin/api/java/ -name "jacoco.exec")
    java -jar build/output/external/jacoco/lib/jacococli.jar merge $execs --destfile bazel-testlogs/api/java/merged.exec
    java -jar build/output/external/jacoco/lib/jacococli.jar report bazel-testlogs/api/java/merged.exec \
    --classfiles bazel-out/k8-opt/bin/api/java/liblib_yr_api_sdk.jar \
    --classfiles bazel-out/k8-opt/bin/api/java/libfunction_common.jar \
    --classfiles bazel-out/k8-opt/bin/api/java/yr_runtime.jar \
    --sourcefiles api/java/yr-api-sdk/src/main/java/ \
    --sourcefiles api/java/function-common/src/main/java/ \
    --sourcefiles api/java/yr-runtime/src/main/java/ \
    --html bazel-testlogs/api/java/html \
    --xml bazel-testlogs/api/java/jacoco-report.xml

    java_total_lines_covered=$(xmllint --xpath 'string(/report/counter[@type="LINE"]/@covered)' bazel-testlogs/api/java/jacoco-report.xml)
    java_total_lines_missed=$(xmllint --xpath 'string(/report/counter[@type="LINE"]/@missed)' bazel-testlogs/api/java/jacoco-report.xml)
    java_total_lines=$((java_total_lines_covered + java_total_lines_missed))
    java_coverage=$(awk "BEGIN {printf \"%.2f\", ($java_total_lines_covered / $java_total_lines) * 100}")
    mkdir -p bazel-testlogs/api/java/liblib_yr_api_sdk/ && unzip bazel-out/k8-opt/bin/api/java/liblib_yr_api_sdk.jar -d bazel-testlogs/api/java/liblib_yr_api_sdk/classes
    mkdir -p bazel-testlogs/api/java/libfunction_common/ && unzip bazel-out/k8-opt/bin/api/java/libfunction_common.jar -d bazel-testlogs/api/java/libfunction_common/classes
    mkdir -p bazel-testlogs/api/java/yr_runtime/ && unzip bazel-out/k8-opt/bin/api/java/yr_runtime.jar -d bazel-testlogs/api/java/yr_runtime/classes
}

function run_python_coverage_report() {
    coverage_files=$(find bazel-out/k8-opt/testlogs/api/python/yr/tests/ -name coverage.dat)
    for i in $coverage_files; do
        MODULE_NAME=$(basename $(dirname $i))
        lcov -q -r $i 'src/*' '*pb2.py' '*tests*' 'apis*' 'cluster_mode*' 'code_manager*' -o ${BASE_DIR}/bazel-bin/api/python/${MODULE_NAME}_coverage.txt
    done

    mkdir -p ${BASE_DIR}/bazel-bin/api/python/coverage_html
    genhtml -q --ignore-errors source ${BASE_DIR}/bazel-bin/api/python/*_coverage.txt -o ${BASE_DIR}/bazel-bin/api/python/coverage_html

    python_coverage_func=$(grep headerCovTableEntryLo bazel-bin/api/python/coverage_html/index.html | awk -F '>' '{print $2}'| awk -F '<' '{print $1}')
    python_coverage=$(grep headerCovTableEntryMed bazel-bin/api/python/coverage_html/index.html | awk -F '>' '{print $2}'| awk -F '<' '{print $1}')
}

function build_python_sdk() {
    if  [ "${BAZEL_COMMAND}" != "build" ];then
        return
    fi
    API_DIR="$BASE_DIR/api"
    cd $API_DIR/python
    $PYTHON3_SDK_BIN_PATH setup.py bdist_wheel
    mkdir -p $OUTPUT_BASE/runtime/sdk/python/
    if [ -e "${OUTPUT_BASE}"/runtime/service/python/yr ]; then
        cp -arf $API_DIR/python/yr/* $OUTPUT_BASE/runtime/service/python/yr
        cp -ar $API_DIR/python/dist/*whl $OUTPUT_BASE/runtime/sdk/python/
        rm -rf $OUTPUT_BASE/runtime/service/python/yr/tests
    else
        mkdir -p $OUTPUT_BASE/runtime/service/python/yr
        mkdir -p $OUTPUT_BASE/runtime/service/python/fnruntime
        cp -ar $API_DIR/python/server.py $OUTPUT_BASE/runtime/service/python/fnruntime
        cp -ar $API_DIR/python/yr/* $OUTPUT_BASE/runtime/service/python/yr
        cp -ar $API_DIR/python/dist/*whl $OUTPUT_BASE/runtime/sdk/python/
        rm -rf $OUTPUT_BASE/runtime/service/python/yr/tests
    fi
    rm -f $OUTPUT_BASE/runtime/service/python/yr/fnruntime.pyx
}

function build_multi_python_version() {
    if [[ "${MULTI_PYTHON_VERSION}" != "true" || "${BAZEL_COMMAND}" != "build" ]];then
        return
    fi
    export PYTHONWARNINGS="ignore::DeprecationWarning"
    for item in "${PYTHON_VERSION_LIST[@]}";
        do
            PYTHON_BIN_PATH=$item
            if command -v $PYTHON_BIN_PATH;then
                PYTHON_INCLUDE=$($PYTHON_BIN_PATH -c "from __future__ import print_function;import sysconfig;print(sysconfig.get_path('include'))")
                if [ ! -e "${PYTHON_INCLUDE}/Python.h" ]; then
                  log_warning "${PYTHON_INCLUDE}/Python.h not exit"
                  continue
                fi
                log_info "start build $PYTHON_BIN_PATH "
                BAZEL_PYTHON_OPTIONS_ENV="${BAZEL_OPTIONS_ENV} --action_env=BUILD_VERSION=${BUILD_VERSION} --action_env=PYTHON3_BIN_PATH=$PYTHON_BIN_PATH"
                BAZEL_PYTHON_OPTIONS="${BAZEL_OPTIONS} ${BAZEL_OPTIONS_CONFIG} ${BAZEL_PYTHON_OPTIONS_ENV}"
                cd $BASE_DIR
                rm -rf ${BASE_DIR}/api/python/yr/fnruntime.cpython-*.so
                rm -rf ${BASE_DIR}/api/python/build/lib*/yr/fnruntime.cpython-*.so
                bazel ${BAZEL_PRE_OPTIONS} ${BAZEL_COMMAND} ${BAZEL_PYTHON_OPTIONS} -- ${PYTHON_BAZEL_TARGETS}
                PYTHON3_SDK_BIN_PATH=$PYTHON_BIN_PATH
                build_python_sdk
            else
                log_warning "there is no $PYTHON_BIN_PATH "
            fi
    done
}

function install_python_requirements() {
    pip3.9 install pytest coverage
    pip3.9 install -r api/python/requirements.txt
    pip3.9 install numpy
    pip3.9 install fastapi
    pip3.9 install aiohttp # only for test
    pip3.9 install requests
}

function check_sanitizers() {
  local name
  name="$1"
  if [[ "X$name" != "Xaddress" && "X$name" != "Xthread" && "X$name" != "Xoff" ]]; then
    log_error "Invalid value $1 for option -S"
    usage
    exit 1
  fi
  if [[ "X$name" == "Xthread" ]]; then
    BAZEL_OPTIONS_ENV="--action_env=TSAN_OPTIONS=suppressions=${BASE_DIR}/test/cpp_tsan.supp"
  fi
}

while getopts 'thr:v:S:DcCgPET:p:bm:j:g' opt; do
    case "$opt" in
    t)
        BAZEL_COMMAND="test"
        BAZEL_TARGETS="//test/... //api/python/yr/tests/... //api/java:java_tests"
        install_python_requirements
        ;;
    T)
        BAZEL_OPTIONS="$BAZEL_OPTIONS --test_arg=${OPTARG}"
        ;;
    h)
        usage
        exit 0
        ;;
    r)
        if curl --connect-timeout 3 --max-time 5 "${OPTARG}" &>/dev/null;then
          log_info "use remote cache server: ${OPTARG}"
          BAZEL_OPTIONS="$BAZEL_OPTIONS --remote_cache=${OPTARG}"
        else
          log_warning "no remote cache server available"
        fi
        ;;
    v)
        BUILD_VERSION="${OPTARG}"
        export BUILD_VERSION="${OPTARG}"
        ;;
    D)
        BAZEL_OPTIONS_CONFIG=" --config=debug "
        ;;
    m)
        BAZEL_OPTIONS="$BAZEL_OPTIONS --local_ram_resources=${OPTARG} "
        ;;
    j)
        BAZEL_OPTIONS="$BAZEL_OPTIONS --jobs=${OPTARG} "
        ;;
    c)
        BAZEL_COMMAND="coverage"
        BAZEL_TARGETS="//test/... //api/python/yr/tests/... //api/java:java_tests"
        BAZEL_OPTIONS="$BAZEL_OPTIONS --combined_report=lcov --nocache_test_results --instrumentation_filter=^//.*[/:]"
        install_python_requirements
        ;;
    C)
        BAZEL_COMMAND="clean"
        BAZEL_OPTIONS="$BAZEL_OPTIONS --expunge"
        BAZEL_TARGETS=""
        ;;
    S)
        check_sanitizers "${OPTARG}"
        SANITIZER="${OPTARG}"
        if [[ "${OPTARG}" != "off" ]]; then
          BAZEL_OPTIONS_CONFIG=" --config=sanitize_${SANITIZER} "
        fi
        ;;
    P)
        PACKAGE_ALL="true"
        ;;
    p)
        if [[ "${OPTARG}" == "multi" ]]; then
            MULTI_PYTHON_VERSION="true"
        else
            MULTI_PYTHON_VERSION="false"
            PYTHON3_BIN_PATH="${OPTARG}"
        fi
        ;;
    E)
        SECBRELLA_CCE="ON"
        BAZEL_OPTIONS_ENV="${BAZEL_OPTIONS_ENV} --action_env=SECBRELLA_CCE_LD_PRELOAD=${SECBRELLA_CCE_LD_PRELOAD}"
        BAZEL_OPTIONS="${BAZEL_OPTIONS} --sandbox_debug"
        ;;
    b)
        BAZEL_OPTIONS_ENV="${BAZEL_OPTIONS_ENV} --action_env=LD_PRELOAD=libbep_env.so"
        ;;
    g)
        BAZEL_TARGETS="//api/python:cp_yr_proto //src/proto:libruntime_cc_proto //src/proto:libruntime_java_proto //src/proto:socket_cc_proto //src/proto:socket_java_proto"
        ;;
    *)
        log_error "invalid command: $opt"
        usage
        exit 1
        ;;
    esac
done

if [ "$BAZEL_COMMAND" != "clean" ]; then
   bash -x ${BASE_DIR}/tools/download_dependency.sh
fi

API_DIR="${BASE_DIR}/api"
sed -i "s/<version>1.0.0<\/version>/<version>${BUILD_VERSION}<\/version>/" $API_DIR/java/pom.xml
sed -i "s/<version>1.0.0<\/version>/<version>${BUILD_VERSION}<\/version>/" $API_DIR/java/yr-api-sdk/pom.xml
sed -i "s/<version>1.0.0<\/version>/<version>${BUILD_VERSION}<\/version>/" $API_DIR/java/function-common/pom.xml
sed -i "s/<version>1.0.0<\/version>/<version>${BUILD_VERSION}<\/version>/" $API_DIR/java/yr-runtime/pom.xml

build_multi_python_version

BAZEL_OPTIONS_ENV="${BAZEL_OPTIONS_ENV} --action_env=BUILD_VERSION=${BUILD_VERSION} --action_env=PYTHON3_BIN_PATH=$PYTHON3_BIN_PATH"
BAZEL_OPTIONS="${BAZEL_OPTIONS} ${BAZEL_OPTIONS_CONFIG} ${BAZEL_OPTIONS_ENV}"

cd $BASE_DIR
rm -rf ${BASE_DIR}/api/python/yr/fnruntime.cpython-*.so
rm -rf ${BASE_DIR}/api/python/build/lib*/yr/fnruntime.cpython-*.so
bazel ${BAZEL_PRE_OPTIONS} ${BAZEL_COMMAND} ${BAZEL_OPTIONS} -- ${BAZEL_TARGETS}

PYTHON3_SDK_BIN_PATH=$PYTHON3_BIN_PATH
build_python_sdk

if [ "$BAZEL_COMMAND" == "coverage" ]; then
    lcov -q -r ${BASE_DIR}/bazel-out/_coverage/_coverage_report.dat '*python*' '*.pb.*' '*test*'  -o ${BASE_DIR}/bazel-out/_coverage/_coverage_report.info
    genhtml -q --ignore-errors source --output genhtml ${BASE_DIR}/bazel-out/_coverage/_coverage_report.info
    cpp_coverage=$(grep headerCovTableEntryMed genhtml/index.html | head -n 1 | awk -F '>' '{print $2}'| awk -F '<' '{print $1}')
    run_java_coverage_report
    run_python_coverage_report
    echo "cpp_covearge: $cpp_coverage" >> genhtml/coverage.txt
    echo "python_covearge: $python_coverage" >> genhtml/coverage.txt
    echo "java_coverage: $java_coverage%" >> genhtml/coverage.txt
    cat genhtml/coverage.txt
fi

if [ "$BAZEL_COMMAND" == "clean" ]; then
    [[ -n "${BUILD_BASE}" ]] && rm -rf "${BUILD_BASE}"
    [[ -n "${BASE_DIR}/api/python/dist/" ]] && rm -rf "${BASE_DIR}/api/python/dist/"
fi

if [ "$BAZEL_COMMAND" == "build" ]; then
    mkdir -p ${OUTPUT_DIR}
    tar -czf ${OUTPUT_DIR}/yr-runtime-${BUILD_VERSION}.tar.gz -C ${OUTPUT_BASE} runtime
    tar -czf ${OUTPUT_DIR}/symbols_libruntime.tar.gz -C ${OUTPUT_BASE} symbols
fi

if [ "$PACKAGE_ALL" == "true" ]; then
    bash ${BASE_DIR}/scripts/package.sh -t ${BUILD_VERSION}
fi
cd -