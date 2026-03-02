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

"""
runtime holder
"""

from yr.base_runtime import BaseRuntime
from yr.config_manager import ConfigManager
from yr.cluster_mode_runtime import ClusterModeRuntime
from yr.local_mode.local_mode_runtime import LocalModeRuntime


class RuntimeHolder:
    """
    runtime holder
    """

    def __init__(self):
        self.yr_runtime = None

    def init(self, runtime: BaseRuntime):
        """init"""
        self.yr_runtime = runtime

    def get_runtime(self) -> BaseRuntime:
        """get runtime"""
        return self.yr_runtime


global_runtime = RuntimeHolder()


def save_real_instance_id(instance_id, need_order):
    """
    save real instance id
    :param instance_id: instance id
    :param need_order: need order
    :return: None
    """
    global_runtime.get_runtime().save_real_instance_id(instance_id, need_order)


def init(runtime=global_runtime) -> None:
    """
    init yr
    :param runtime: RuntimeHolder
    :return: None
    """
    if ConfigManager().local_mode:
        rt = LocalModeRuntime()
        rt.init()
    else:
        rt = ClusterModeRuntime()
        rt.init()
    runtime.init(rt)
