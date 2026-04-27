#!/usr/bin/env python3
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

from enum import Enum, auto
from dataclasses import dataclass
from typing import ClassVar, List, Optional

from yr.decorator.instance_proxy import InstanceProxy

import constants


class VllmInstanceStatus(Enum):
    UNREADY = auto()
    RUNNING = auto()
    SUBPROCESS_EXITED = auto()
    HEALTHCHECK_FAILED = auto()


class Role(Enum):
    """定义实例角色类型"""

    PREFILL = 0
    DECODE = 1
    MIXED = 2


class SchedulerPolicy(Enum):
    """负载均衡策略类型"""

    ROUND_ROBIN = 0  # 轮询


@dataclass
class VllmInstanceInfo:
    """VLLM实例的健康状态信息"""

    id: str
    #: the api server's http address, e.g. "http://10.2.3.4:8000"
    uri: str
    #: the instance's role
    role: Role
    #: the instance's status
    status: VllmInstanceStatus = VllmInstanceStatus.UNREADY
    #: dp master ip
    dp_master_ip: str = ""
    #: dp master port
    dp_master_port: int = 0


@dataclass
class DispatchResult:
    prefill_vllm_instance_uri: str
    decode_vllm_instance_uri: str


@dataclass
class MetricsInfo:
    running_num: int = 0
    waiting_num: int = 0
    npu_usage_perc: float = 0.0

    # key:param_name, value:metric_name
    METRIC_NAME_MAPPING: ClassVar[dict] = {
        "running_num": constants.NUM_REQUESTS_RUNNING,
        "waiting_num": constants.NUM_REQUESTS_WAITING,
        "npu_usage_perc": constants.GPU_CACHE_USAGE_PERC,
    }


@dataclass
class InstanceInfo:
    actor: InstanceProxy
    in_p: bool
    in_d: bool


@dataclass
class ActorInstanceInfo:
    actor: InstanceProxy
    config: 'VllmInstanceConfig'
    instance_id: str


@dataclass
class DeployConfig:
    head_ip: str
    prefill_instances_num: int
    prefill_startup_params: List[str]
    prefill_startup_env: Optional[str]
    prefill_data_parallel_size: int
    prefill_tensor_parallel_size: int
    prefill_expert_parallel_size: int
    decode_instances_num: int
    decode_startup_params: List[str]
    decode_startup_env: Optional[str]
    decode_data_parallel_size: int
    decode_tensor_parallel_size: int
    decode_expert_parallel_size: int
    scheduler_policy: SchedulerPolicy
    timeout: int
    proxy_host: str
    proxy_port: int
