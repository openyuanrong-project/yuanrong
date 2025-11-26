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
BASE_DIR=$(cd "$(dirname "$0")"; pwd)
. "${BASE_DIR}/utils.sh"
YUANRONG_DIR="${BASE_DIR}/../../output/yuanrong"
BUILD_DIR="${BASE_DIR}/../../build"
RESERVED_CLUSTER="off"
GTEST_FILTER="*.*"
DEPLOY_PATH="/tmp/deploy/$(date -u "+%d%H%M%S")"
LANGUAGE="all"
COMPILE="off"
DEPLOY_MODE="process"
LOCAL_IP=$(ifconfig | grep inet | grep -vw "127.0.0.1" | grep -vw "172.17.0.1" | grep -vw "10.42.0.1" | grep -v inet6 | awk '{print $2}')
SANITIZER="off"
PARALLELISM="off"
AGENT_NUM=1
START_ONLY="off"

readonly USAGE="
Usage: bash test.sh [-h] [-f *.*] [-y path-to-yuanrong] [-r][-l language][-m deploy mode][-S thread/address/off]

Options:
    -h Output this help and exit.
    -f gtest filter
    -y yuanrong dir
    -r do not stop yuanrong
    -m Deployment mode. The value can be 'k8s' or 'process'. If value is 'k8s', Kubernetes cluster is required.
       If value is 'process', you need to download yuanrong.tar.gz and decompress it.
    -l Case language. the value can be 'cpp'/'java'/'python'/'go'/'all'/'manual', the corresponding language cases
       would be executed.
    -S Use Google Sanitizers tools to detect bugs. Choose from off/address/thread,
               if set the value to 'address' enable AddressSanitizer,
               if set the value to 'thread' enable ThreadSanitizer,
               default off.

Example:
  $ bash test.sh
"

function usage() {
    echo -e "$USAGE"
}

function clear_env() {
    set +e
    ps -ef | grep "yuanrong" | grep -v "grep" | grep -v "test.sh" | awk -F " " '{print $2}' | xargs kill -9 &>/dev/null
    set -e
}

function compile_st() {
    if [ "$COMPILE" == "on" ]; then
        bash ${BASE_DIR}/cpp/build.sh -S ${SANITIZER} -y ${YUANRONG_DIR}
        bash ${BASE_DIR}/java/build.sh -y ${YUANRONG_DIR}
    fi
}

function generate_test_dir() {
    set +e
    cd ${BASE_DIR}
    mkdir -p "${DEPLOY_PATH}/pkg"
    if [ ! -f ${BASE_DIR}/cpp/build/libuser_common_func.so ] || [ ! -d ${BASE_DIR}/java/target ]; then
        if [ "$COMPILE" == "off" ];then
            compile_st
        fi
    fi
    cp -r ${BASE_DIR}/cpp/build/libuser_common_func.so "${DEPLOY_PATH}/pkg/"
    cp -r ${BASE_DIR}/java/target/java-st-test-*.jar "${DEPLOY_PATH}/pkg/"
    set -e
}

function install_python_pkg() {
    pip3.9 install pytest
    pip3.9 install requests
    pip3.9 install numpy
    pip3.9 uninstall -y yr
    pip3.9 install $YUANRONG_DIR/runtime/sdk/python/yr_sdk-*cp39-cp39-linux*.whl
}

function common_check_st_result() {
    ret_code=$1
    failed_case_num=$2
    language=$3
    if [ $failed_case_num -eq 0 -a $ret_code -eq 0 ]; then
        echo "----------------------Success to run ${language} st----------------------"
    elif [ $failed_case_num -eq 0 -a $ret_code -ne 0 ]; then
        echo "Last one ${language} case failed, may timeout or happened core dump!"
        echo "----------------------Failed to run ${language} st----------------------"
        cat "$GLOG_log_dir/${language}_output.txt"
        exit $ret_code
    else
        echo "Failed ${language} Count is: ${failed_case_num}"
        echo "Failed ${language} Cases List: "
        cat failed_case_list
        echo "----------------------Failed to run ${language} st----------------------"
        cat "$GLOG_log_dir/${language}_output.txt"
        exit 1
    fi
}

function check_python_run_st_result() {
    set +e
    python_ret_code=$1
    grep FAILED "$GLOG_log_dir/python_output.txt" |grep test_ > failed_case_list
    python_failed_case_num=$(cat failed_case_list | wc -l)
    set -e
    common_check_st_result $python_ret_code $python_failed_case_num python
}

function check_cpp_run_st_result() {
    set +e
    cpp_ret_code=$1
    grep -B 3 "FAILED" "$GLOG_log_dir/cpp_output.txt" | grep ms | awk -F "]" '{print $2}'| grep Test. > failed_case_list
    cpp_failed_case_num=$(cat failed_case_list | wc -l)
    set -e
    common_check_st_result $cpp_ret_code $cpp_failed_case_num cpp
}

function check_java_run_st_result() {
    set +e
    java_ret_code=$1
    # 3 rules to select lines:
    #   1. after '[ERROR] Failures:';
    #   2. before '[ERROR] Tests run:'
    #   3. lines start with '[ERROR]   '.
    awk '
    /\[ERROR\] (Failures|Errors):/ { in_failures = 1; next }
    in_failures && /^\[ERROR\]/ {
        if (/Tests run/) {
            exit;
        }
        print;
    }
    ' "$GLOG_log_dir/java_output.txt" > failed_case_list
    java_failed_case_num=$(cat failed_case_list | wc -l)
    set -e
    common_check_st_result $java_ret_code $java_failed_case_num java
}

function check_run_faas_case_result() {
    set +e
    faas_ret_code=$1
    grep FAILED "$GLOG_log_dir/faas_output.txt" |grep test_ > failed_case_list
    faas_failed_case_num=$(cat failed_case_list | wc -l)
    set -e
    common_check_st_result $faas_ret_code $faas_failed_case_num faas
}

function timeout_run_case_wrapper() {
    set +e
    run_cmd=$@
    echo "start to run ${run_cmd}"
    start_time=$(date +%s)
    timeout 300 ${run_cmd} 2>&1
    ret=$?
    end_time=$(date +%s)
    execute_time=$((end_time - start_time))
    echo "The cmd ${run_cmd} executed cost ${execute_time} seconds"
    return $ret
}

function run_cpp_case() {
    echo "----------------------Start to run cpp st----------------------"
    cd "$GLOG_log_dir"
    export DEPLOY_PATH
    timeout_run_case_wrapper ${BASE_DIR}/cpp/build/yrapicpp --gtest_filter=${GTEST_FILTER} 2>&1 > "$GLOG_log_dir/cpp_output.txt"
    check_cpp_run_st_result $?
}

function run_python_case() {
    echo "----------------------Start to run python st----------------------"
    cd ${BASE_DIR}/python
    [ "${SANITIZER}" == "address" ] && export LD_PRELOAD="/usr/lib64/libasan.so.6" && export ASAN_OPTIONS=detect_leaks=0
    [ "${SANITIZER}" == "thread" ] && export LD_PRELOAD="/usr/local/gcc-10.3.0/lib64/libtsan.so"
    timeout_run_case_wrapper python3.9 -m pytest -s -vv -m smoke . 2>&1 > "$GLOG_log_dir/python_output.txt"
    check_python_run_st_result $?
    export LD_PRELOAD="" && export ASAN_OPTIONS=""
}

function run_manual_case() {
    # You can modify manual case here.
    cp -f ${BASE_DIR}/others/rpc_retry_test/build/libuser_common_func.so "${DEPLOY_PATH}/pkg/"
    export DEPLOY_PATH
    bash ${BASE_DIR}/others/rpc_retry_test/exec_test.sh
}

function run_java_case() {
    echo "----------------------Start to run java st----------------------"
    cd ${BASE_DIR}/java
    [ "${SANITIZER}" == "address" ] && export LD_PRELOAD="/usr/lib64/libasan.so.6" && export ASAN_OPTIONS=detect_leaks=0
    timeout_run_case_wrapper mvn test 2>&1 > "$GLOG_log_dir/java_output.txt"
    check_java_run_st_result $?
    export LD_PRELOAD="" && export ASAN_OPTIONS=""
}

function run_faas_case() {
    echo "----------------------Start to run faas st----------------------"
    cd ${BASE_DIR}/faas
    timeout_run_case_wrapper pytest -s -vv test_python.py 2>&1 > "$GLOG_log_dir/faas_output.txt"
    check_run_faas_case_result $?
}

function clean_hook() {
    if [ "$RESERVED_CLUSTER" == "off" ]; then
        trap clear_env SIGTERM SIGINT EXIT
    fi
}

function deploy_yr() {
    cd "${BASE_DIR}"
    if [[ "$DEPLOY_MODE" == "process" ]]; then
        bash prepare_and_start_yr.sh -d "$DEPLOY_PATH" -y "$YUANRONG_DIR" -a "$LOCAL_IP" -S "$SANITIZER" -n "$AGENT_NUM"
        PROXY_GRPC_PORT=$(grep PROXY_GRPC_PORT ${DEPLOY_PATH}/log/*-deploy_std.log | awk '{print $NF}')
        DS_WORKER_PORT=$(grep DS_WORKER_PORT ${DEPLOY_PATH}/log/*-deploy_std.log | awk '{print $NF}')
        MASTER_PORT=$(grep "all port for control plane:" ${DEPLOY_PATH}/log/*-deploy_std.log | awk '{print $12}')
        export YR_DS_ADDRESS="${LOCAL_IP}:$DS_WORKER_PORT"
        export YR_SERVER_ADDRESS="${LOCAL_IP}:$PROXY_GRPC_PORT"
        export YR_MASTER_ADDRESS="${LOCAL_IP}:$MASTER_PORT"
        export YR_IS_CLUSTER="true"
    elif [[ "$DEPLOY_MODE" == "k8s" ]]; then
        cd "$BASE_DIR"/functions
        yrctl deploy -p ./pkg || exit 0
        cd -
        cluster_ip=$(grep serverlessnative /etc/hosts|awk '{print $1}')
        export YR_SERVER_ADDRESS="$cluster_ip:31222"
        export YR_DS_ADDRESS="$cluster_ip:31222"
        export YR_IS_CLUSTER="false"
    else
        echo "-m parameter ${DEPLOY_MODE} is invalid, see -h for help"
        usage
        exit 1
    fi
    export YRFUNCID='sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:$latest'
    export YR_PYTHON_FUNC_ID='sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stpython:$latest'
    export YR_JAVA_FUNC_ID='sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stjava:$latest'
    export YR_LOG_LEVEL=DEBUG
    export GLOG_log_dir="$DEPLOY_PATH/driver"
    export DEPLOY_PATH

    mkdir -p "$GLOG_log_dir"

    echo "export YR_DS_ADDRESS=$YR_DS_ADDRESS"
    echo "export YR_SERVER_ADDRESS=$YR_SERVER_ADDRESS"
    echo "export YR_FRONTEND_ADDRESS=$YR_FRONTEND_ADDRESS"
    echo "export YRFUNCID='sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:\$latest'"
    echo "export YR_PYTHON_FUNC_ID='sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stpython:\$latest'"
    echo "export YR_JAVA_FUNC_ID='sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stjava:\$latest'"
    echo "export YR_IS_CLUSTER=$YR_IS_CLUSTER"
    echo "export YR_LOG_LEVEL=$YR_LOG_LEVEL"
    echo "export DEPLOY_PATH=$DEPLOY_PATH"
    echo "export GLOG_log_dir=$DEPLOY_PATH/driver"
    echo "export YR_MASTER_ADDRESS=$YR_MASTER_ADDRESS"
}

function run_st() {

    install_python_pkg

    if [[ "$LANGUAGE" == "cpp" ]]; then
        run_cpp_case
    elif [[ "$LANGUAGE" == "python" ]]; then
        run_python_case
    elif [[ "$LANGUAGE" == "java" ]]; then
        run_java_case
    elif [[ "$LANGUAGE" == "faas" ]]; then
        run_faas_case
    elif [[ "$LANGUAGE" == "all" ]]; then
        if [[ "$PARALLELISM" == "off" ]]; then
            run_cpp_case
            run_python_case
            run_java_case
        else
            run_cpp_case &
            pid_cpp=$!
            run_python_case &
            pid_py=$!
            run_java_case &
            pid_java=$!
            wait $pid_cpp && wait $pid_py && wait $pid_java
            if [ $? -ne 0 ]; then
                exit 1
            fi
        fi
    elif [[ "$LANGUAGE" == "manual" ]]; then
        run_manual_case
    fi
}

while getopts 'rbf:y:l:m:a:h:S:n:sp' opt; do
    case "${opt}" in
        r)
            RESERVED_CLUSTER="on"
            ;;
        f)
            GTEST_FILTER="${OPTARG}"
            ;;
        y)
            YUANRONG_DIR="${OPTARG}"
            YUANRONG_DIR=$(realpath "$YUANRONG_DIR")
            ;;
        l)
            LANGUAGE="${OPTARG}"
            ;;
        b)
            COMPILE="on"
            ;;
        m)
            DEPLOY_MODE="${OPTARG}"
            ;;
        a)
            LOCAL_IP="${OPTARG}"
            ;;
        h)
            usage
            exit 0
            ;;
        S)
            check_sanitizers "${OPTARG}"
            SANITIZER="${OPTARG}"
            ;;
        p)
            PARALLELISM="off"
            ;;
        n)
            AGENT_NUM="${OPTARG}"
            ;;
        s)
            START_ONLY="on"
            RESERVED_CLUSTER="on"
            ;;
        *)
            echo "invalid command"
            usage
            exit 1
            ;;
    esac
done
compile_st
clean_hook
deploy_yr
if [ "$START_ONLY" == "off" ]; then
    generate_test_dir
    run_st
fi
