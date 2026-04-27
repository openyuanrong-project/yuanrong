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

readonly USAGE="
Usage: bash build.sh [-thdDcCrvPSbEm:j:GU]

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
    -G enable gloo collective operations (default: disabled)
    -U enable UCC collective operations (default: disabled)
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
BAZEL_OPTIONS=""
BAZEL_OPTIONS_CONFIG=" --config=release "
BAZEL_TARGETS="//api/cpp:yr_cpp_pkg //api/java:yr_java_pkg //api/python:yr_python_pkg //api/go:yr_go_pkg"
BAZEL_PRE_OPTIONS="--output_user_root=${BUILD_BASE} --output_base=${OUTPUT_BASE}"
THIRD_PARTY_DIR="${BASE_DIR}/thirdparty"
PYTHON3_BIN_PATH="python3"
PYTHON3_SDK_BIN_PATH=$PYTHON3_BIN_PATH
PYTHON_BAZEL_TARGETS="//api/python:yr_python_pkg"
MULTI_PYTHON_VERSION="true"
SANITIZER="off"
BAZEL_OPTIONS_ENV=""
SECBRELLA_CCE="OFF"
PACKAGE_ALL="false"
ENABLE_GLOO="false"
ENABLE_UCC="false"
LD_LIBRARY_PATH=/opt/buildtools/python3.7/lib:/opt/buildtools/python3.9/lib:/opt/buildtools/python3.11/lib:/opt/buildtools/python3.12/lib:/opt/buildtools/python3.13/lib:${LD_LIBRARY_PATH}
BOOST_VERSION="1.87.0"
export BUILD_ALL="false"
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

MODULE_LIST=(\
"runtime_go"
)

PYTHON_VERSION_LIST=(\
"python3.13" \
"python3.12" \
"python3.11" \
"python3.10" \
"python3.9"
)

function go_module_coverage_report() {
    COVERAGE_SUFFIX="_coverage.out"
    [ "${SANITIZER}" != "off" ] && COVERAGE_SUFFIX="_coverage_${SANITIZER}.out"
    MODULE=$1
    MODULE_SOURCE=$(echo "$MODULE" | cut -d '-' -f 1)
    pushd ${GO_SRC_BASE}
    sed -i "/clibruntime.go/d" ${BASE_DIR}/bazel-bin/go/${MODULE_SOURCE}${COVERAGE_SUFFIX}
    gocov convert ${BASE_DIR}/bazel-bin/go/${MODULE_SOURCE}${COVERAGE_SUFFIX} > ${BASE_DIR}/bazel-bin/go/${MODULE_SOURCE}_coverage.json
    gocov report ${BASE_DIR}/bazel-bin/go/${MODULE_SOURCE}_coverage.json > ${BASE_DIR}/bazel-bin/go/${MODULE_SOURCE}_coverage.txt
    gocov-html ${BASE_DIR}/bazel-bin/go/${MODULE_SOURCE}_coverage.json > ${BASE_DIR}/bazel-bin/go/${MODULE_SOURCE}_coverage.html
    popd
}

function run_go_coverage_report() {
   for i in "${!MODULE_LIST[@]}"
   do
        MODULE_SOURCE=$(echo "${MODULE_LIST[$i]}" | cut -d '-' -f 1)
        go_module_coverage_report ${MODULE_LIST[$i]}
        coverage_info=$(tail -1 ${BASE_DIR}/bazel-bin/go/${MODULE_SOURCE}_coverage.txt)
        ((cov_go_total+=$(echo $coverage_info | awk -F '[()/]' '{print $(NF-1)}')))
        ((cov_go+=$(echo $coverage_info | awk -F '[()/]' '{print $(NF-2)}')))
   done
   go_coverage=$(echo "scale=4; $cov_go / $cov_go_total * 100" | bc)
}

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
    rm -rf build/ dist/ *.egg-info
    # Determine python runtime version for services.yaml
    if [ "$MULTI_PYTHON_VERSION" == "true" ]; then
        PYTHON_RUNTIME_VERSION=python3.11
    else
        PYTHON_RUNTIME_VERSION=$PYTHON3_BIN_PATH
    fi
    # Main wheel declares install_requires on openyuanrong_sdk; build and ship both.
    SETUP_TYPE=sdk PYTHON_RUNTIME_VERSION=$PYTHON_RUNTIME_VERSION $PYTHON3_SDK_BIN_PATH setup.py bdist_wheel
    mkdir -p ${OUTPUT_DIR}
    mkdir -p $OUTPUT_BASE/runtime/sdk/python/
    cp -ar $API_DIR/python/dist/*whl $BASE_DIR/output/
    cp -ar $API_DIR/python/dist/*whl $OUTPUT_BASE/runtime/sdk/python/
    rm -rf build/ dist/ *.egg-info
    SETUP_TYPE= PYTHON_RUNTIME_VERSION=$PYTHON_RUNTIME_VERSION $PYTHON3_SDK_BIN_PATH setup.py bdist_wheel
    cp -ar $API_DIR/python/dist/*whl $BASE_DIR/output/
    chmod 750 $BASE_DIR/output/*.whl
    cp -ar $API_DIR/python/dist/*whl $OUTPUT_BASE/runtime/sdk/python/
    if [ -e "${OUTPUT_BASE}"/runtime/service/python/yr ]; then
        cp -arf $API_DIR/python/yr/* $OUTPUT_BASE/runtime/service/python/yr
        cp -ar $API_DIR/python/dist/*whl $OUTPUT_BASE/runtime/sdk/python/
        rm -rf $OUTPUT_BASE/runtime/service/python/yr/tests
    else
        mkdir -p $OUTPUT_BASE/runtime/service/python/yr
        cp -ar $API_DIR/python/yr/* $OUTPUT_BASE/runtime/service/python/yr
        cp -ar $API_DIR/python/dist/*whl $OUTPUT_BASE/runtime/sdk/python/
        rm -rf $OUTPUT_BASE/runtime/service/python/yr/tests
    fi
    rm -f $OUTPUT_BASE/runtime/service/python/yr/fnruntime.pyx
}

function install_python_requirements() {
    "${PYTHON3_BIN_PATH}" -m pip install -U pip setuptools wheel
    "${PYTHON3_BIN_PATH}" -m pip install pytest coverage
    "${PYTHON3_BIN_PATH}" -m pip install -r api/python/requirements.txt
    "${PYTHON3_BIN_PATH}" -m pip install numpy
    "${PYTHON3_BIN_PATH}" -m pip install fastapi
    "${PYTHON3_BIN_PATH}" -m pip install aiohttp # only for test
    "${PYTHON3_BIN_PATH}" -m pip install requests
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

while getopts 'athr:l:v:S:DcCgPET:p:B:m:j:gGU' opt; do
    case "$opt" in
    a)
        BUILD_ALL="true"
        ;;
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
        REMOTE_CACHE="${OPTARG}"
        ;;
    l)
        if [ ! -d "${OPTARG}" ] ;then
          mkdir -p ${OPTARG}
        fi
        BAZEL_OPTIONS="$BAZEL_OPTIONS --disk_cache=${OPTARG} "
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
        BAZEL_TARGETS="//api/go:yr_go_test //test/... //api/python/yr/tests/..."
        BAZEL_OPTIONS="$BAZEL_OPTIONS --combined_report=lcov --nocache_test_results --instrumentation_filter=^//.*[/:] --test_tag_filters=-cgo"
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
    B)
        BOOST_VERSION="${OPTARG}"
        ;;
    g)
        BAZEL_TARGETS="//api/python:cp_yr_proto //src/proto:libruntime_cc_proto //src/proto:libruntime_java_proto //src/proto:socket_cc_proto //src/proto:socket_java_proto"
        ;;
    G)
        ENABLE_GLOO="true"
        ;;
    U)
        ENABLE_UCC="true"
        ;;
    *)
        log_error "invalid command: $opt"
        usage
        exit 1
        ;;
    esac
done

# Support remote cache via environment variable REMOTE_CACHE
# Command-line -r takes precedence; falls back to env var
if [ -n "${REMOTE_CACHE}" ]; then
    cache_addr="${REMOTE_CACHE}"
    # Extract host and port from various formats like:
    #   http://host:port, https://host:port, grpc://host:port, host:port
    host=$(echo "${cache_addr}" | sed -E 's|grpc://||; s|http://||; s|https://||; s|/.*||' | cut -d: -f1)
    port=$(echo "${cache_addr}" | sed -E 's|grpc://||; s|http://||; s|https://||; s|/.*||' | cut -s -d: -f2)

    # Set default port based on protocol
    if [ -z "$port" ]; then
        if echo "${cache_addr}" | grep -q "^https://"; then
            port=443
        elif echo "${cache_addr}" | grep -q "^grpc://"; then
            port=443
        else
            port=80
        fi
    fi

    # Check if port is reachable using multiple methods
    port_reachable=false
    if timeout 3 bash -c "exec 3<>/dev/tcp/${host}/${port}" 2>/dev/null; then
        port_reachable=true
    elif command -v nc &>/dev/null && nc -z -w 3 "${host}" "${port}" 2>/dev/null; then
        port_reachable=true
    elif curl --connect-timeout 3 --max-time 5 "${cache_addr}" &>/dev/null; then
        port_reachable=true
    fi

    if [ "$port_reachable" = true ]; then
        log_info "use remote cache server: ${cache_addr} (host: ${host}, port: ${port})"
        BAZEL_OPTIONS="$BAZEL_OPTIONS --remote_cache=${cache_addr}"
    else
        log_warning "no remote cache server available at ${host}:${port}"
    fi
fi

if [ "$BAZEL_COMMAND" != "clean" ]; then
   bash ${BASE_DIR}/tools/download_dependency.sh
fi

API_DIR="${BASE_DIR}/api"
sed -i "s/<version>1.0.0<\/version>/<version>${BUILD_VERSION}<\/version>/" $API_DIR/java/pom.xml
sed -i "s/<version>1.0.0<\/version>/<version>${BUILD_VERSION}<\/version>/" $API_DIR/java/yr-api-sdk/pom.xml
sed -i "s/<version>1.0.0<\/version>/<version>${BUILD_VERSION}<\/version>/" $API_DIR/java/function-common/pom.xml
sed -i "s/<version>1.0.0<\/version>/<version>${BUILD_VERSION}<\/version>/" $API_DIR/java/yr-runtime/pom.xml

if command -v "${PYTHON3_BIN_PATH}" >/dev/null 2>&1; then
    "${PYTHON3_BIN_PATH}" -m pip install -U wheel
else
    pip3 install -U wheel
fi

PYTHON_BIN_FULL_PATH="$(command -v "${PYTHON3_BIN_PATH}" 2>/dev/null || true)"
if [[ -z "${PYTHON_BIN_FULL_PATH}" ]]; then
    log_fatal "python not found: ${PYTHON3_BIN_PATH}"
fi

# - action_env: for genrules (e.g. api/python/BUILD.bazel suffix rename)
# - repo_env: for @local_config_python (python headers + libs) to match the selected interpreter
BAZEL_OPTIONS_ENV="${BAZEL_OPTIONS_ENV} --action_env=BOOST_VERSION=$BOOST_VERSION --action_env=GOPATH=$(go env GOPATH) --action_env=GOEXPERIMENT=$(go env GOEXPERIMENT) --action_env=GOCACHE=$(go env GOCACHE) --action_env=BUILD_VERSION=${BUILD_VERSION} --action_env=PYTHON3_BIN_PATH=${PYTHON3_BIN_PATH} --define ENABLE_GLOO=${ENABLE_GLOO}"
BAZEL_OPTIONS="${BAZEL_OPTIONS} ${BAZEL_OPTIONS_CONFIG} ${BAZEL_OPTIONS_ENV}"

cd $BASE_DIR
bazel ${BAZEL_PRE_OPTIONS} ${BAZEL_COMMAND} ${BAZEL_OPTIONS} -- ${BAZEL_TARGETS}

PYTHON3_SDK_BIN_PATH=$PYTHON3_BIN_PATH
build_python_sdk

if [ "$BAZEL_COMMAND" == "coverage" ]; then
    lcov -q -r ${BASE_DIR}/bazel-out/_coverage/_coverage_report.dat '*python*' '*.pb.*' '*test*'  -o ${BASE_DIR}/bazel-out/_coverage/_coverage_report.info
    genhtml -q --ignore-errors source --output genhtml ${BASE_DIR}/bazel-out/_coverage/_coverage_report.info
    cpp_coverage=$(grep headerCovTableEntryMed genhtml/index.html | head -n 1 | awk -F '>' '{print $2}'| awk -F '<' '{print $1}')
    run_go_coverage_report
    run_java_coverage_report
    run_python_coverage_report
    echo "cpp_covearge: $cpp_coverage" >> genhtml/coverage.txt
    echo "python_covearge: $python_coverage" >> genhtml/coverage.txt
    echo "java_coverage: $java_coverage%" >> genhtml/coverage.txt
    echo "go_coverage: $go_coverage%" >> genhtml/coverage.txt
    cat genhtml/coverage.txt
fi

if [ "$BAZEL_COMMAND" == "clean" ]; then
    [[ -n "${BUILD_BASE}" ]] && rm -rf "${BUILD_BASE}"
    [[ -n "${BASE_DIR}/api/python/dist/" ]] && rm -rf "${BASE_DIR}/api/python/dist/"
fi

if [ "$BAZEL_COMMAND" == "build" ]; then
    mkdir -p ${OUTPUT_DIR}
    tar -czf ${OUTPUT_DIR}/yr-runtime-${BUILD_VERSION}.tar.gz -C ${OUTPUT_BASE} runtime
    # tar -czf ${OUTPUT_DIR}/symbols_libruntime.tar.gz -C ${OUTPUT_BASE} symbols
fi

if [ "$PACKAGE_ALL" == "true" ]; then
    # Determine the python version to use in package
    if [ "$MULTI_PYTHON_VERSION" == "true" ]; then
        # Default to python3.11 when building multiple versions
        PACKAGE_PYTHON_VERSION=python3.11
    else
        # Use the specified python version
        PACKAGE_PYTHON_VERSION=$PYTHON3_BIN_PATH
    fi

    start=$(date +%s)
    bash ${BASE_DIR}/scripts/package_yuanrong.sh -v ${BUILD_VERSION}
    end1=$(date +%s)
    echo "Package openyuanrong.tar.gz elapsed: $((end1 - start)) seconds"

    cd "$BASE_DIR"/api/python
    rm -rf build/ dist/ *.egg-info
    PYTHON_RUNTIME_VERSION=${PACKAGE_PYTHON_VERSION} $PYTHON3_SDK_BIN_PATH setup.py bdist_wheel
    cp -ar $API_DIR/python/dist/*whl $BASE_DIR/output/
    rm -rf build/ dist/ *.egg-info
    SETUP_TYPE=dashboard PYTHON_RUNTIME_VERSION=${PACKAGE_PYTHON_VERSION} $PYTHON3_SDK_BIN_PATH setup.py bdist_wheel
    cp -ar $API_DIR/python/dist/*whl $BASE_DIR/output/
    rm -rf build/ dist/ *.egg-info
    SETUP_TYPE=faas PYTHON_RUNTIME_VERSION=${PACKAGE_PYTHON_VERSION} $PYTHON3_SDK_BIN_PATH setup.py bdist_wheel
    cp -ar $API_DIR/python/dist/*whl $BASE_DIR/output/
    rm -rf build/ dist/ *.egg-info
    SETUP_TYPE=sdk_cpp PYTHON_RUNTIME_VERSION=${PACKAGE_PYTHON_VERSION} $PYTHON3_SDK_BIN_PATH setup.py bdist_wheel
    cp -ar $API_DIR/python/dist/*whl $BASE_DIR/output/
    rm -rf build/ dist/ *.egg-info
    SETUP_TYPE=runtime PYTHON_RUNTIME_VERSION=${PACKAGE_PYTHON_VERSION} $PYTHON3_SDK_BIN_PATH setup.py bdist_wheel
    cp -ar $API_DIR/python/dist/*whl $BASE_DIR/output/
    rm -rf build/ dist/ *.egg-info
    SETUP_TYPE=full PYTHON_RUNTIME_VERSION=${PACKAGE_PYTHON_VERSION} $PYTHON3_SDK_BIN_PATH setup.py bdist_wheel
    cp -ar $API_DIR/python/dist/*whl $BASE_DIR/output/
    end2=$(date +%s)
    echo "Package openyuanrong.whl elapsed: $((end2 - end1)) seconds"
    chmod 750 $BASE_DIR/output/*.whl
fi
cd -
