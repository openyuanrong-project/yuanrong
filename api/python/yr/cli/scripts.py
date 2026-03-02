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

from datetime import datetime, timezone
import os
import ctypes
import shutil
import sys
import subprocess
import click

import yr.cli

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
        os.path.join(
            yr.cli.yuanrong_installation_dir, "functionsystem", "bin", cli_name
        )
    ] + sys.argv[1:]
    try:
        process = subprocess.Popen(
            command,
            stdin=sys.stdin,
            stdout=sys.stdout,
            stderr=sys.stderr,
            universal_newlines=True,
            env={
                "YR_INSTALLATION_DIR": yr.cli.yuanrong_installation_dir,
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


__server_address = None
__ds_address = None


@click.group()
@click.option(
    "--server-address",
    required=False,
    type=str,
)
@click.option(
    "--ds-address",
    required=False,
    type=str,
)
@click.version_option(package_name="yr_sdk")
def cli(server_address, ds_address):
    """
    run command
    """
    if server_address:
        global __server_address
        __server_address = server_address
    if ds_address:
        global __ds_address
        __ds_address = ds_address


class YRContext:
    def __init__(self, server_address, ds_address):
        self.__server_address = server_address
        self.__ds_address = ds_address

    def __enter__(self):
        cfg = yr.Config()
        cfg.log_dir = "/tmp/yr_sessions/driver"
        if self.__server_address:
            cfg.server_address = self.__server_address
            cfg.in_cluster = False
            return yr.init(cfg)
        if self.__server_address and self.__ds_address:
            cfg.server_address = self.__server_address
            cfg.ds_address = self.__ds_address
            return yr.init(cfg)
        return yr.init(cfg)

    def __exit__(self, exc_type, exc_value, exc_traceback):
        yr.finalize()
        return False


@cli.command()
@click.option("--backend", required=False, type=str, default="ds")
@click.option("--code-path", default=".")
@click.option("--format", "archive_format", default="zip")
def deploy(backend, code_path, archive_format):
    real_code_path = os.path.realpath(code_path)
    file_name = f"code-{datetime.now(tz=timezone.utc).strftime('%Y%m%d%H%M')}"
    archive_file = os.path.join("/tmp", file_name)
    if archive_format == "zip":
        shutil.make_archive(archive_file, archive_format, real_code_path)
    elif archive_format == "img":
        result = subprocess.run(
            [
                "mkfs.erofs",
                "-E",
                "noinline_data",
                f"{archive_file}.{archive_format}",
                real_code_path,
            ]
        )
        if result.returncode != 0:
            sys.exit(result.returncode)
    else:
        print(f"unkown format: {archive_format}")
        sys.exit(1)
    if backend == "ds":
        with YRContext(__server_address, __ds_address):
            with open(f"{archive_file}.{archive_format}", "rb") as f:
                yr.kv_write(file_name, f.read())

        package_key = f"ds://{file_name}.{archive_format}"
        print(
            f"""already upload {real_code_path} to {backend}.
use export YR_WORKING_DIR={package_key} to set this package.
use yrcli clear {package_key} to delete this package."""
        )
    else:
        print("not support backend: %s" % backend)
        sys.exit(1)


@cli.command
@click.argument("package", type=str)
def clear(package):
    if package.startswith("ds://"):
        key = package.removeprefix("ds://").split(".")[0]
        with YRContext(__server_address, __ds_address):
            yr.kv_del(key)
        print(f"succeed to del {package}")


@cli.command
@click.argument("package", type=str)
def download(package):
    if package.startswith("ds://"):
        key = package.removeprefix("ds://").split(".")[0]
        file_name = package.removeprefix("ds://")
        with YRContext(__server_address, __ds_address):
            with open(file_name, "wb") as f:
                value = yr.kv_get(key)
                f.write(value)
        print(f"save {package} to {file_name}")
