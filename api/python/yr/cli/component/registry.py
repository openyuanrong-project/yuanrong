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

from yr.cli.component.base import ComponentLauncher
from yr.cli.component.collector import CollectorLauncher
from yr.cli.component.dashboard import DashboardLauncher
from yr.cli.component.ds_worker import DsWorkerLauncher
from yr.cli.component.etcd import ETCDLauncher
from yr.cli.component.frontend import FrontendLauncher
from yr.cli.component.function_agent import FunctionAgentLauncher
from yr.cli.component.function_master import FunctionMasterLauncher
from yr.cli.component.function_proxy import FunctionProxyLauncher
from yr.cli.component.function_scheduler import FaaSSchedulerLauncher
from yr.cli.component.meta_service import MetaServiceLauncher
from yr.cli.const import StartMode

LAUNCHER_CLASSES: dict[str, type[ComponentLauncher]] = {
    "etcd": ETCDLauncher,
    "ds_master": DsWorkerLauncher,
    "ds_worker": DsWorkerLauncher,
    "function_master": FunctionMasterLauncher,
    "function_proxy": FunctionProxyLauncher,
    "function_agent": FunctionAgentLauncher,
    "frontend": FrontendLauncher,
    "function_scheduler": FaaSSchedulerLauncher,
    "dashboard": DashboardLauncher,
    "collector": CollectorLauncher,
    "meta_service": MetaServiceLauncher,
    # used for yr k8s deployment
    "iam_server": ComponentLauncher,
    "function_manager": ComponentLauncher,
    "runtime_manager": ComponentLauncher,
}

# Per-component overrides for how config args are rendered on the command line.
# By default, ComponentConfig.prepend_char is "--"; components listed here
# will use a different prefix when SystemLauncher constructs them.
PREPEND_CHAR_OVERRIDES: dict[str, str] = {
    "ds_master": "-",
    "ds_worker": "-",
    "frontend": "-",
}


DEPENDS_ON_OVERRIDES_BY_MODE: dict[StartMode, dict[str, list[str]]] = {
    StartMode.MASTER: {
        "function_master": ["etcd"],
        "function_proxy": ["ds_worker"],
        "function_agent": ["ds_worker", "function_proxy"],
        "frontend": ["function_proxy"],
        "function_scheduler": ["function_proxy"],
        "dashboard": ["function_master"],
        "ds_master": ["etcd"],
        "ds_worker": ["etcd"],
        "collector": ["dashboard"],
        "meta_service": ["etcd"],
        "etcd": [],
    },
    StartMode.AGENT: {
        "function_agent": ["ds_worker", "function_proxy"],
        "function_proxy": ["ds_worker"],
    },
}


def get_depends_on_overrides(mode: StartMode) -> dict[str, list[str]]:
    return DEPENDS_ON_OVERRIDES_BY_MODE.get(mode, {})
