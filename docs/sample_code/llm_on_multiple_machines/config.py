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

from dataclasses import dataclass
from typing import List, Optional

from entities import SchedulerPolicy, Role

# [ (long options) , (short options) ]
vllm_options_should_be_filled = [
    # host and port
    ("host",),
    ("port",),
    # tp config, which should use --p-tp/--d-tp
    ("tensor-parallel-size", "tp"),
    # dp config, which should use --p-dp/--d-dp
    ("data-parallel-size", "dp"),
    ("data-parallel-size-local", "dpl"),
    ("data-parallel-start-rank", "dpr"),
    ("data-parallel-address", "dpa"),
    ("data-parallel-rpc-port", "dpp"),
    ("headless",),
    # ep config
    ("enable-expert-parallel",),
    ("no-enable-expert-parallel",),
    # pd disagg config, auto used if enable pd disagg
    ("kv-transfer-config",),
]


class AutoValidator:
    def __post_init__(self):
        for name, _ in self.__dataclass_fields__.items():
            method = getattr(self, f"_validate_{name}", None)
            if method:
                method()


@dataclass
class InferenceInstanceConfig(AutoValidator):
    startup_params: List[str]
    startup_env: Optional[str]
    tp: Optional[int]
    dp: Optional[int]
    ep: Optional[int]

    def _validate_startup_params(self):
        def __contain_long_options(opname, params):
            underline_op = opname.replace('-', '_')
            return any(p == f"--{opname}" or p.startswith(f"--{opname}=") or
                       p == f"--{underline_op}" or p.startswith(f"--{underline_op}=")
                       for p in params)

        def __contain_short_options(opname, params):
            underline_op = opname.replace('-', '_')
            return any(p == f"-{opname}" or p.startswith(f"-{opname}=") or
                       p == f"--{underline_op}" or p.startswith(f"--{underline_op}=")
                       for p in params)

        not_acceptable_options = []
        for opt in vllm_options_should_be_filled:
            if len(opt) > 0 and __contain_long_options(opt[0], self.startup_params):
                not_acceptable_options.append(opt[0])
            if len(opt) > 1 and __contain_short_options(opt[1], self.startup_params):
                not_acceptable_options.append(opt[1])

        if len(not_acceptable_options) > 0:
            raise ValueError(
                f"Options {not_acceptable_options} are reserved and should not be specified in start up command; "
            )

    def _validate_ep(self):
        if self.ep < 0:
            raise ValueError("expert parallel size should be 0(disable) or >1(enable)")


@dataclass
class ControllerConfig(AutoValidator):
    """
    prefill_instances_num: Number of P (Prefill) instances to start
    prefill_startup_params: Common startup parameters for P instances
    decode_instances_num: Number of D (Decode) instances to start
    decode_startup_params: Common startup parameters for D instances
    scheduler_policy: Scheduling policy enum
    """

    scheduler_policy: SchedulerPolicy

    prefill_instances_num: int
    p_inference_instance_config: InferenceInstanceConfig
    decode_instances_num: int
    d_inference_instance_config: InferenceInstanceConfig
    timeout: int

    def _validate_prefill_instances_num(self):
        if self.prefill_instances_num < 0:
            raise ValueError("prefill instance num should be equal to or greater than 0")

    def _validate_decode_instances_num(self):
        if self.decode_instances_num < 0:
            raise ValueError("decode instance num should be equal to or greater than 0")


@dataclass
class VllmInstancePDConfig(AutoValidator):
    role: Role
    pd_rank: Optional[int] = 0
    pd_size: Optional[int] = 0

    def is_disaggregated_p_d(self):
        """judge if in the pd disaggregated mode"""
        return self.role != Role.MIXED


@dataclass
class VllmInstanceDPConfig(AutoValidator):
    dp_rank: Optional[int] = 0
    dp_size: Optional[int] = 1
    dp_local_size: Optional[int] = 1
    dp_master_ip: Optional[str] = ""
    dp_master_port: Optional[int] = 0

    def is_dp_enabled(self):
        """judge if dp is enabled"""
        return self.dp_size > 1


@dataclass
class VllmInstanceEPConfig(AutoValidator):
    ep_size: Optional[int] = 0

    def is_ep_enabled(self):
        """judge if ep is enabled"""
        return self.ep_size > 0


@dataclass
class VllmInstanceConfig(AutoValidator):
    exec_cmd: List[str]
    env: Optional[str] = None
    tp: Optional[int] = 0
    pd_config: Optional[VllmInstancePDConfig] = None
    dp_config: Optional[VllmInstanceDPConfig] = None
    ep_config: Optional[VllmInstanceEPConfig] = None
