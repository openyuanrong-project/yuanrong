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

"""yr runtime main"""

import json
import os
import sys

from yr import init
from yr.apis import receive_request_loop
from yr.config import Config
from yr.common.utils import load_env_from_file, try_install_uvloop

DEFAULT_LOG_DIR = "/home/snuser/log/"
_ENV_KEY_FUNCTION_LIB_PATH = "YR_FUNCTION_LIB_PATH"
_ENV_KEY_ENV_DELEGATE_DOWNLOAD = "ENV_DELEGATE_DOWNLOAD"
_ENV_KEY_LAYER_LIB_PATH = "LAYER_LIB_PATH"

DEFAULT_RUNTIME_CONFIG = {
    "maxRequestBodySize": 6,
    "maxHandlerNum": 1024,
    "maxProcessNum": 1024,
    "dataSystemConnectionTimeout": 0.15,
    "traceEnable": False,
    "tlsEnable": False,
}


def get_runtime_config():
    """
    get runtime configuration
    """
    config_path = "/home/snuser/config/runtime.json"
    if os.getenv("YR_BARE_MENTAL") is not None:
        config_path = sys.path[0] + "/../config/runtime.json"
    with open(config_path, "r") as config_file:
        config_json = json.load(config_file)

    return config_json


def configure():
    """configure"""
    config = Config()
    config.rt_server_address = os.getenv("POSIX_LISTEN_ADDR", "127.0.0.1:22732")
    config.runtime_id = os.getenv("YR_RUNTIME_ID", "driver")
    config.job_id = os.getenv("YR_JOB_ID", "job-ffffffff")
    config.log_level = os.getenv("YR_LOG_LEVEL", "INFO")
    config.env_file = os.getenv("YR_ENV_FILE", "")
    config.ds_address = os.getenv("DATASYSTEM_ADDR")
    config.is_driver = False
    log_dir = os.getenv("GLOG_log_dir")
    if log_dir:
        config.log_dir = log_dir
    else:
        config.log_dir = DEFAULT_LOG_DIR
    config.in_cluster = True
    return config


def insert_sys_path():
    """insert sys path"""
    code_dir = os.getenv(_ENV_KEY_FUNCTION_LIB_PATH, "")
    custom_handler = os.getenv(_ENV_KEY_ENV_DELEGATE_DOWNLOAD, "")
    layer_paths = os.getenv(_ENV_KEY_LAYER_LIB_PATH, "").split(",")
    paths = layer_paths + [code_dir, custom_handler]
    for path in paths:
        if path != "":
            sys.path.insert(1, path)


def main():
    """main"""
    # If YR_SEED_FILE is set, read the specified file to block
    seed_file = os.environ.get("YR_SEED_FILE", "")
    if seed_file:
        with open(seed_file, "rb") as f:
            f.read()

    # Get env_file from environment variable
    # Load environment variables from file BEFORE any code reads from os.environ
    # This must be done before insert_sys_path() and configure() which read env vars
    load_env_from_file(os.environ.get("YR_ENV_FILE", ""))

    insert_sys_path()
    init(configure())
    try_install_uvloop()
    receive_request_loop()


if __name__ == "__main__":
    main()
