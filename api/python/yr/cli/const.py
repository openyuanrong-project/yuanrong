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

from enum import Enum

DEFAULT_CONFIG_PATH = "/etc/yuanrong/config.toml"
HTTP_PROTOCOL = "http://"
SESSIONS_DIR = "/tmp/yr_sessions"
SESSION_LATEST_PATH = f"{SESSIONS_DIR}/latest"
DEFAULT_LOG_DIR = f"{SESSION_LATEST_PATH}/logs"
DEFAULT_DEPLOY_DIR = f"{SESSION_LATEST_PATH}/deploy"
SESSION_JSON_PATH = f"{SESSION_LATEST_PATH}/session.json"
DEFAULT_DEPLOY_DIR = f"{SESSION_LATEST_PATH}/deploy"
DEFAULT_VALUES_TOML = "cli/values.toml"
DEFAULT_CONFIG_TEMPLATE_PATH = "cli/config.toml.jinja"
DEFAULT_MASTER_INFO_PATH = f"{SESSIONS_DIR}/yr_current_master_info"

FS_COMPONENTS = [
    "domain-scheduler",
    "function_agent",
    "function_master",
    "function_proxy",
    "function-accessor",
    "iam_server",
    "runtime_manager",
]


class StartMode(Enum):
    MASTER = "master"
    AGENT = "agent"
