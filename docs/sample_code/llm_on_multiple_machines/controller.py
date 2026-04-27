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

import asyncio
import time
from typing import List, Dict
import logging
import yr
from yr.decorator.instance_proxy import InstanceProxy

from entities import ActorInstanceInfo, Role, VllmInstanceInfo
from constants import BALANCER_ACTOR_NAME, VLLM_NAMESPACE
from config import (
    ControllerConfig,
    VllmInstanceConfig,
    VllmInstancePDConfig,
    VllmInstanceDPConfig,
    VllmInstanceEPConfig,
    InferenceInstanceConfig,
)
from vllm_instance import start_vllm_instance
from balancer import Balancer

logger = logging.getLogger(__name__)

# wait extra `grace_exit_check_timeout_s` to make sure the network transfer
# is overlapped
grace_exit_check_timeout_s = 5


def _get_npu_num_per_yr_node():
    npu_nums = []
    for node in yr.resources():
        nums = 0
        for key, value in node["capacity"].items():
            if "NPU" in key:
                nums = nums + value
        if nums != 0:
            npu_nums.append(int(nums))
    return max(npu_nums)


def split_dp_resources(tp_size: int, dp_size: int, npu_pack_max_size: int = 8) -> List[int]:
    """
    split dp resources into some packed groups, like

    | DP   | TP   | total | 910C  | 910B    |
    | ---- | ---- | ----- | ----- | ------- |
    | 4    | 2    | 8     | 8     | 8       |
    | 3    | 3    | 9     | 9     | 6+3     |
    | 4    | 4    | 16    | 16    | 8+8     |
    | 32   | 1    | 32    | 16+16 | 8+8+8+8 |
    | 64   | 1    | 64    | 16*4  | 8*8     |

    right now, we don't care about resource fragments

    Returns:
        list of npu nums
    """
    if tp_size <= npu_pack_max_size:
        # TP size is within the allowed limit; no action needed
        pass
    else:
        raise ValueError(f"When enabling DP, the TP size ({tp_size}) should not exceed the "
                         f"maximum allowed size ({npu_pack_max_size}) on a single machine.")
    total_npu = dp_size * tp_size
    group_size = (
        npu_pack_max_size - (npu_pack_max_size % tp_size) if npu_pack_max_size % tp_size != 0 else npu_pack_max_size
    )
    num_groups = total_npu // group_size
    remainder = total_npu % group_size
    packs = [group_size] * num_groups
    if remainder > 0:
        packs.append(remainder)
    return packs


class Controller:
    def __init__(self, controller_config: ControllerConfig):
        """
        Initialize the global controller.

        Args:
            controller_config: ControllerConfig
        """
        self.config = controller_config

        self.is_disaggregated_pd = True
        self.p_instances_actors: List[InstanceProxy] = []
        self.d_instances_actors: List[InstanceProxy] = []
        self.vllm_instances_info: Dict[str, VllmInstanceInfo] = {}
        self.actor_instance_info: Dict[str, ActorInstanceInfo] = {}
        self.balancer = None
        self.prefill_instances_cnt = 0
        self.decode_instances_cnt = 0

        self.active_dp_groups: Dict[str, Dict] = {}
        self.dp_groups: Dict[str, Dict] = {}  # group_key (string) -> group info
        self.instance_to_group: Dict[str, str] = {}  # instance_id -> group_key

    async def make_inference_instance(
            self, pd_role: Role, pd_rank: int, inference_instance_config: InferenceInstanceConfig
    ) -> List[InstanceProxy]:
        """make inference instance (PREFILL instance, or DECODE instance)
        1. if dp enabled,     ==> start dp group
        2. if dp not enabled, ==> just start vllm instance

        Returns:
            all vllm instances actors in this inference instance
        """
        if inference_instance_config.dp > 1:
            # enable dp
            return await make_dp_group(
                controller=self,
                pd_role=pd_role,
                pd_idx=pd_rank,
                tp_size=inference_instance_config.tp,
                dp_size=inference_instance_config.dp,
                ep_size=inference_instance_config.ep,
                start_params=inference_instance_config.startup_params,
                env=inference_instance_config.startup_env,
            )

        # no dp
        return [
            self.create_vllm_instance(
                VllmInstanceConfig(
                    exec_cmd=inference_instance_config.startup_params,
                    env=inference_instance_config.startup_env,
                    tp=inference_instance_config.tp,
                    pd_config=VllmInstancePDConfig(role=pd_role, pd_rank=pd_rank),
                    dp_config=VllmInstanceDPConfig(),
                    ep_config=VllmInstanceEPConfig(inference_instance_config.ep),
                )
            )
        ]

    async def make_balancer(self) -> List[InstanceProxy]:
        """make balancer, and send all vllm instance info to the balancer

        Returns:
            balancer handle
        """
        opt = yr.InvokeOptions()
        opt.name = BALANCER_ACTOR_NAME
        opt.namespace = VLLM_NAMESPACE
        opt.concurrency = 5
        balancer = yr.instance(Balancer).options(opt).invoke(
            policy=self.config.scheduler_policy)
        return balancer

    async def initialize(self):
        """initialize all vllm instances, construct pd/dp groups"""
        logger.info(f"initialize with config: {self.config}")
        # Dictionary to track VLLM instances health status
        self.vllm_instances_info: Dict[str, VllmInstanceInfo] = {}  #
        self.is_disaggregated_pd = self.config.prefill_instances_num > 0 and self.config.decode_instances_num > 0
        self.balancer = await self.make_balancer()
        logger.info(f"Finished create Balancer")

        # start VllmInstance
        await self.scale_out_inference_ins(self.config.prefill_instances_num, self.config.decode_instances_num)

    async def scale_out_inference_ins(self, prefill_ins_to_scale, decode_ins_to_scale):
        """
        scale out instance
        """
        scaled_p_ins: List[InstanceProxy] = []
        scaled_d_ins: List[InstanceProxy] = []
        p_pd_rank = self.prefill_instances_cnt
        for p_pd_idx in range(prefill_ins_to_scale):
            p_actors = self.make_inference_instance(
                pd_rank=p_pd_rank + p_pd_idx,
                pd_role=Role.PREFILL if self.is_disaggregated_pd else Role.MIXED,
                inference_instance_config=self.config.p_inference_instance_config,
            )
            scaled_p_ins.extend(await p_actors)
            self.prefill_instances_cnt += 1

        # start Decode Instances
        d_pd_rank = self.decode_instances_cnt
        for d_pd_idx in range(decode_ins_to_scale):
            d_actors = self.make_inference_instance(
                pd_rank=d_pd_rank + d_pd_idx,
                pd_role=Role.DECODE if self.is_disaggregated_pd else Role.MIXED,
                inference_instance_config=self.config.d_inference_instance_config,
            )
            scaled_d_ins.extend(await d_actors)
            self.decode_instances_cnt += 1

        # init all vllm instances
        for vllm_instance_actor in [*scaled_p_ins, *scaled_d_ins]:
            vllm_instance_actor.initialize.invoke()

        self.p_instances_actors.extend(scaled_p_ins)
        self.d_instances_actors.extend(scaled_d_ins)
        # wait for all instances ready
        await self.balancer.check_instances_ready_and_monitor.invoke(timeout=self.config.timeout)
        logger.info(f"All instances ready, VllmInstance num: {len(self.vllm_instances_info)}, updating Balancer")

        # update Balancer
        self.balancer.update_vllm_instance_info.invoke(list(self.vllm_instances_info.values()))

        logger.info(
            f"Controller initialized with {self.config.prefill_instances_num} P instances and "
            f"{self.config.decode_instances_num} D instances"
        )

    async def graceful_shutdown(self, timeout_s=5) -> None:
        """
        TODO: clean all actors started by controller
        """
        try:
            if self.balancer:
                self.balancer.terminate()
            terminate_futures = []
            for actor in [*self.p_instances_actors, *self.d_instances_actors]:
                terminate_futures.append(actor.graceful_shutdown.invoke(timeout_s=timeout_s))

            try:
                # may raise TimeoutError
                async with asyncio.timeout(delay=timeout_s + grace_exit_check_timeout_s):
                    await asyncio.gather(*terminate_futures)
            finally:
                for actor in [*self.p_instances_actors, *self.d_instances_actors]:
                    actor.terminate()
        finally:
            pass

    async def report_failure_from_balancer(self, instance_id: str) -> None:
        """
        Report a failure for an instance from the balancer.

        Args:
            instance_id (str): The ID of the instance that has failed.

        Returns:
            bool: True if the report was handled successfully, False otherwise.
        """
        logger.info(f"Received report from balancer, instance_id is {instance_id} ")
        actor_info = self.actor_instance_info.get(instance_id)

        await self.balancer.remove_failed_instance.invoke(instance_id)
        await self.terminate_instance(actor_info.actor)
        logger.info(f"Actor {instance_id} terminated, will restart the new actor")

        actor_handle = actor_info.actor
        in_p_instances_actors = actor_handle in self.p_instances_actors
        in_d_instances_actors = actor_handle in self.d_instances_actors
        if in_p_instances_actors:
            self.p_instances_actors = [a for a in self.p_instances_actors if a != actor_handle]
        else:
            self.d_instances_actors = [a for a in self.d_instances_actors if a != actor_handle]

        self.actor_instance_info.pop(instance_id, None)

        max_retries = 3
        retry_count = 0
        restart_success = False
        new_actor = None
        while retry_count < max_retries and not restart_success:
            retry_count += 1
            try:
                new_actor = await self.restart_instance(actor_info)
                if new_actor is None:
                    raise Exception
            except Exception as restart_err:
                if new_actor is not None:
                    new_actor.terminate()
                    new_actor = None
                logger.warning(f"Attempt {retry_count} to restart actor {instance_id} failed, \
                               error: {str(restart_err)}")
                if retry_count < max_retries:
                    logger.warning(f"Retrying restart for actor {instance_id}...")
                    time.sleep(2 ** retry_count)
                else:
                    logger.error(f"CRITICAL: FAILED to restart actor {instance_id} after {max_retries} attempts")
            else:
                logger.info(f"New actor created successfully: {new_actor}")
                restart_success = True

        if in_p_instances_actors:
            self.p_instances_actors.append(new_actor)
        elif in_d_instances_actors:
            self.d_instances_actors.append(new_actor)

        if restart_success:
            logger.info("New Actor has been restarted")

    def create_vllm_instance(self, vllm_instance_config: VllmInstanceConfig) -> InstanceProxy:
        """
        Create and start a new VLLM instance.

        Args:
            vllm_instance_config (VllmInstanceConfig): Configuration for the VLLM instance.

        Returns:
            InstanceProxy: The handle of the created VLLM instance.
        """
        role_enum = vllm_instance_config.pd_config.role
        role_name = role_enum.name
        pd_rank = vllm_instance_config.pd_config.pd_rank
        instance_id = f"vllm-instance-{role_name}-{pd_rank}"

        actor = start_vllm_instance(vllm_instance_config=vllm_instance_config, name=instance_id)
        self.actor_instance_info[instance_id] = ActorInstanceInfo(
            actor=actor,
            config=vllm_instance_config,
            instance_id=instance_id
        )
        return actor

    async def terminate_instance(self, actor_to_terminate: InstanceProxy) -> None:
        """
        Terminate a given actor instance.

        Args:
            actor_to_terminate (InstanceProxy): The actor instance to terminate.
        """
        try:
            logger.info("start terminated actor")
            await asyncio.gather(actor_to_terminate.graceful_shutdown.invoke(timeout_s=5))
            actor_to_terminate.terminate()
            logger.info("finished terminated actor")
        except Exception as e:
            logger.warning(f"Cannot terminate instance, it maybe already killed: {str(e)}")

    async def restart_instance(self, actor_info: Dict[str, ActorInstanceInfo]) -> InstanceProxy:
        """
        Restart a VLLM instance based on the provided actor information (Should be same as the before one).

        Args:
            actor_info (dict): Information about the instance to restart.

        Returns:
            InstanceProxy: The newly created actor instance.
        """
        new_actor: InstanceProxy = None
        try:
            new_actor = self.create_vllm_instance(
                vllm_instance_config=actor_info.config,
            )
            yr.get(new_actor.initialize.invoke())
        except Exception as e:
            logger.warning(f"Actor initialization failed: {str(e)}")
            if new_actor is not None:
                new_actor.terminate()
            return None
        logger.info("New actor restarted successfully")
        return new_actor

    def get_expected_vllm_actors_num(self):
        """get expected vllm actors num"""
        return len(self.p_instances_actors) + len(self.d_instances_actors)


async def make_dp_group(
        controller: Controller, pd_role: Role, pd_idx: int, tp_size: int, dp_size: int, ep_size: int,
        start_params: List[str], env: str = None
) -> List[InstanceProxy]:
    """
    prepare one dp group
    1. start dp master vllm instance
        1.1. find dp master ip and a free port as dp master port
        1.2. init dp master vllm instance's dp config
    2. start other dp vllm instances with dp master ip and dp master port
    """
    packs = split_dp_resources(
        tp_size=tp_size, dp_size=dp_size, npu_pack_max_size=_get_npu_num_per_yr_node()
    )

    actors = []
    dp_master_vllm_instance_config = VllmInstanceConfig(
        exec_cmd=start_params,
        env=env,
        tp=tp_size,
        pd_config=VllmInstancePDConfig(role=pd_role, pd_rank=pd_idx),
        dp_config=VllmInstanceDPConfig(dp_rank=0, dp_size=dp_size, dp_local_size=packs[0] // tp_size),
        ep_config=VllmInstanceEPConfig(ep_size=ep_size),
    )
    dp_master_actor = controller.create_vllm_instance(vllm_instance_config=dp_master_vllm_instance_config)
    actors.append(dp_master_actor)

    dp_master_ip, dp_master_port = await dp_master_actor.init_dp_master_ip_port.invoke()
    dp_master_vllm_instance_config.dp_config.dp_master_ip = dp_master_ip
    dp_master_vllm_instance_config.dp_config.dp_master_port = dp_master_port
    await dp_master_actor.init_dp_config.invoke(dp_master_vllm_instance_config.dp_config)

    dp_rank = packs[0] // tp_size
    for idx in range(1, len(packs)):
        dp_vllm_instance_config = VllmInstanceConfig(
            exec_cmd=start_params,
            env=env,
            tp=tp_size,
            pd_config=VllmInstancePDConfig(role=pd_role, pd_rank=pd_idx),
            dp_config=VllmInstanceDPConfig(
                dp_rank=dp_rank, dp_size=dp_size, dp_master_ip=dp_master_ip, dp_master_port=dp_master_port,
                dp_local_size=packs[idx] // tp_size,
            ),
            ep_config=VllmInstanceEPConfig(ep_size=ep_size),
        )
        dp_rank += packs[idx] // tp_size
        actor = controller.create_vllm_instance(vllm_instance_config=dp_vllm_instance_config)
        await actor.init_dp_config.invoke(dp_vllm_instance_config.dp_config)
        actors.append(actor)
    return actors
