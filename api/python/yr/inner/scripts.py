#!/usr/bin/env python3
# coding=UTF-8
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

import ctypes
import os
import sys
import subprocess

import yr.inner


PR_SET_PDEATHSIG = 1
SIGTERM = 15


def set_pdeathsig():
    """
    Send a SIGTERM signal to the child process when its parent process terminates.
    """
    libc = ctypes.CDLL("libc.so.6")
    result = libc.prctl(PR_SET_PDEATHSIG, SIGTERM, 0, 0, 0)
    if result != 0:
        raise OSError(f"prctl failed with error code {result}")


def run_cli_prog(cli_name: str = "yr"):
    """
    run yuanrong cli tool
    :param cli_name: the name of yuanrong cli tool
    """
    set_pdeathsig()
    command = [
        os.path.join(yr.inner.yuanrong_installation_dir, "functionsystem", "bin", cli_name)
    ] + sys.argv[1:]
    try:
        process = subprocess.Popen(
            command,
            stdin=sys.stdin,
            stdout=sys.stdout,
            stderr=sys.stderr,
            universal_newlines=True,
            env={
                "YR_INSTALLATION_DIR": yr.inner.yuanrong_installation_dir,
                **os.environ,
            },
            preexec_fn=set_pdeathsig,
        )
        process.communicate()
        sys.exit(process.returncode)
    except subprocess.CalledProcessError:
        pass
        sys.exit(1)


def run_yr():
    """
    run yr command
    """
    run_cli_prog("yr")


if __name__ == "__main__":
    run_yr()
