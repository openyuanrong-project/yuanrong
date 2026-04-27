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

# ----------------------------------------------------------------------
# funcname:     log_info.
# description:  Print build info log.
# parameters:   NA
# return value: NA
# ----------------------------------------------------------------------
log_info()
{
    echo "[BUILD_INFO][$(date -u +%b\ %d\ %H:%M:%S)]$*"
}

# ----------------------------------------------------------------------
# funcname:     log_warning.
# description:  Print build warning log.
# parameters:   NA
# return value: NA
# ----------------------------------------------------------------------
log_warning()
{
    echo "[BUILD_WARNING][$(date -u +%b\ %d\ %H:%M:%S)]$*"
}

# ----------------------------------------------------------------------
# funcname:     log_error.
# description:  Print build error log.
# parameters:   NA
# return value: NA
# ----------------------------------------------------------------------
log_error()
{
    echo "[BUILD_ERROR][$(date -u +%b\ %d\ %H:%M:%S)]$*"
}

# ----------------------------------------------------------------------
# funcname:     die.
# description:  Print build error log.
# parameters:   NA
# return value: NA
# ----------------------------------------------------------------------
die()
{
    log_error "$*"
    stty echo
    exit 1
}

# ----------------------------------------------------------------------
# funcname: usage
# description:  the build Instructions
# parameters:   void
# return value: void
# ----------------------------------------------------------------------
usage()
{
    echo -e "Usage: ./build.sh [-m subsystem_name] [-h help]"
    echo -e "Usage: ./docker_build.sh [-m subsystem_name -v version] [-h help]"
    echo -e "Options:"
    echo -e "     -m subsystem_name, such as nodemanager workermanager crontrigger frontend worker"
    echo -e "                                functionrepo adminservice cli"
    echo -e "                                functionstate functiontask"
    echo -e "     -m functioncore build functioncore containing all functioncore modules"
    echo -e "     -m all contains runtime and worker-flow in addition to functioncore"
    echo -e "Notes: If the parameter of -m is all, you must download the runtime project first,"
    echo -e "       and keep the the projects' path at the same level as the functioncore."
    echo -e "     -h usage help"
    echo -e "     -l compile with local runtime code"
    echo -e "      "
    echo -e "Example:"
    echo -e "   sh  build.sh -m \"worker\""
    echo -e "   sh  build.sh -m \"functioncore\""
    echo -e "   sh  build.sh -m \"all\""
    echo -e "   sh  build.sh -l -m \"runtimemanager\""
    echo -e "   sh  docker_build.sh -v \"version\" -m \"worker\""
    echo -e "   sh  docker_build.sh -v \"version\" -m \"functioncore\""
    echo -e "   sh  docker_build.sh -v \"version\" -m \"all\""
    echo -e "   sh  docker_build.sh -v \"version\" -m \"runtimemanager\""
    echo -e ""
}

# ----------------------------------------------------------------------
# funcname:     get_base_image.
# description:  check the system
# parameters:   base images path
# return value: NA
# ----------------------------------------------------------------------
function get_base_image() {
    local local_os=$(head -1 /etc/os-release | tail -1 | awk -F "\"" '{print $2}')_$(uname -m)
    log_info "The operating system is ${local_os}."

    case "${local_os}" in
    Ubuntu_x86_64)
        BASE_IMAGE="ubuntu:18.04"
        ;;
    EulerOS_x86_64)
        BASE_IMAGE="euleros:v2r9"
        ;;
    EulerOS_aarch64)
        BASE_IMAGE="euleros:v2r8"
        ;;
    openEuler_x86_64)
        BASE_IMAGE="openeuler:20.03"
        ;;
    "CentOS Linux_x86_64")
        BASE_IMAGE="centos:7.9"
        ;;
    esac
}
