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

import logging
import asyncio
from typing import List, Dict, Optional
import time

import yr
from yr.decorator.instance_proxy import InstanceProxy

from entities import Role, SchedulerPolicy, VllmInstanceInfo, DispatchResult, MetricsInfo, VllmInstanceStatus
from constants import (
    CONTROLLER_ACTOR_NAME,
    VLLM_NAMESPACE,
    INSTANCE_STATUS_CHECK_INTERVAL
)

logger = logging.getLogger(__name__)


class Balancer:
    def __init__(
            self,
            policy: SchedulerPolicy = SchedulerPolicy.ROUND_ROBIN,
    ):
        self.policy = policy
        self.role_2_instances: Dict[Role, VllmInstanceInfo] = {}  # prefill/decode/mixed => VllmInstanceInfo
        self.instance_infos: Dict[str, VllmInstanceInfo] = {}  # id -> VllmInstanceInfo
        self.active_instances: Dict[str, VllmInstanceInfo] = {}  # id -> VllmInstanceInfo
        self.instance_metrics: Dict[str, MetricsInfo] = {}  # id -> MetricsInfo
        self._round_robin_index_p = 0
        self._round_robin_index_d = 0
        self._round_robin_index_m = 0
        self.last_heartbeat: Dict[str, float] = {}
        self._controller_handle: InstanceProxy = None
        self.reported_failures = set()  # 用于跟踪已报告的失败实例
        self.lock = asyncio.Lock()  # 用于并发控制

    def dispatch_request(self) -> DispatchResult:
        """
        分发请求给相应的处理模块。

        根据请求内容或类型，选择合适的处理逻辑进行处理。

        Args:

        Returns:
            处理结果， 包含prefill_vllm_instance_uri和decode_vllm_instance_uri
        """
        if self.policy == SchedulerPolicy.ROUND_ROBIN:
            return self._round_robin_pair()
        raise ValueError(f"Unsupported policy: {self.policy}")

    def get_all_instance(self) -> Dict[str, VllmInstanceInfo]:
        '''Return all vllm instance.'''
        return {key: item for key, item in self.instance_infos.items() if item.uri is not None}

    def update_vllm_instance_info(self, infos: List[VllmInstanceInfo]):
        """
        更新 vLLM 实例的相关信息。

        该方法负责获取并刷新 vLLM 实例的状态或配置信息，
        以支持负载均衡和监控逻辑。

        Args:
            vLLM 实例的相关信息类

        Returns:

        """
        for item in infos:
            self.instance_infos[item.id] = item
            self.instance_metrics[item.id] = MetricsInfo()

        # reconstruct the role map
        self.role_2_instances.clear()
        for _, instance_info in self.instance_infos.items():
            if instance_info.role not in self.role_2_instances:
                self.role_2_instances[instance_info.role] = []
            self.role_2_instances[instance_info.role].append(instance_info)

    async def add_vllm_instance(self, vllm_instance_info: VllmInstanceInfo):
        """
        Add a VLLM instance information to the balancer.

        Args:
            instance_info (VllmInstanceInfo): The information of the VLLM instance to add.

        Updates the internal state by adding the instance information,
        metrics, and last heartbeat time.
        """
        self.update_vllm_instance_info([vllm_instance_info])
        # Once the instance is ready, add it to monitoring_instance and begin monitoring its health
        asyncio.create_task(self.monitor_instance_status(vllm_instance_info.id))

    async def monitor_instance_status(self, instance_id):
        """
        Monitor the status of the VLLM instance until it becomes RUNNING.
        """
        # Check if instance is ready
        while self.instance_infos[instance_id].status != VllmInstanceStatus.RUNNING:
            logger.debug(f'Balancer waiting for new instance: %s to become RUNNING status', instance_id)
            await asyncio.sleep(INSTANCE_STATUS_CHECK_INTERVAL)

        if instance_id in self.reported_failures:
            self.reported_failures.remove(instance_id)
        self.active_instances[instance_id] = self.instance_infos[instance_id]
        logger.info(f"Instance {instance_id} is now RUNNING and added to monitoring")

    async def update_vllm_instance_health(self, vllm_instance_info: List[VllmInstanceInfo]) -> bool:
        """
        Update health status of VLLM instances.

        Args:
            vllm_instance_info: List of VllmInstanceInfo objects containing information

        Returns:
            bool: True if update was successful
        """
        current_time = time.time()
        for item in vllm_instance_info:
            self.instance_infos[item.id] = item
            self.last_heartbeat[item.id] = current_time
            if item.id in self.active_instances:
                self.active_instances[item.id] = item

    async def check_instances_ready_and_monitor(self, timeout: Optional[float] = None):
        """
        Wait until all VLLM actor running status

        Returns:
            No return value. End of function when all actor ready,.
        """
        logger.info(f"Start checking if all instances are ready.")
        if not self._controller_handle:
            try:
                self._controller_handle = yr.get_instance(name=CONTROLLER_ACTOR_NAME, namespace=VLLM_NAMESPACE)
            except BaseException:
                logger.warning('get _controller_handle fail')
        get_expected_vllm_actors_num = await self._controller_handle.get_expected_vllm_actors_num.invoke()

        start_time = time.time()
        while self._get_ready_vllm_actors_num() < get_expected_vllm_actors_num:
            logger.debug(f"expect %s waiting vllm actor, %s", self._get_ready_vllm_actors_num(), self.instance_infos)
            for s in self.instance_infos.values():
                if s.status == VllmInstanceStatus.SUBPROCESS_EXITED:
                    raise RuntimeError(f"vllm instance: {s} exited unexpectedly")
            if timeout is not None and time.time() - start_time > timeout:
                raise TimeoutError(f"Timeout waiting for instance to become RUNNING")
            await asyncio.sleep(1)
        logger.info(f"All instances are already")

    async def remove_failed_instance(self, instance_id: str):
        """
        Controller调用此方法通知Balancer删除实例
        """
        if instance_id not in self.reported_failures:
            self.reported_failures.add(instance_id)

        if instance_id in self.last_heartbeat:
            self.last_heartbeat.pop(instance_id, None)

        if instance_id in self.instance_metrics:
            self.instance_metrics.pop(instance_id, None)

        if instance_id in self.active_instances:
            self.active_instances.pop(instance_id, None)
        logger.info(f"Removed failed instance {instance_id} from balancer")

    async def report_failures_to_controller(self, failed_instances: List[VllmInstanceInfo]) -> None:
        """批量报告失败的实例"""
        for instance in failed_instances:
            instance_id = instance.id
            try:
                await self._controller_handle.report_failure_from_balancer.invoke(instance_id)
                logger.info(f"Sccuessful rebuild instance.")
            except Exception as e:
                logger.warning(f"Failed to report failure for instance {instance_id}: {str(e)}")

    def _round_robin_selection(self, role: Role) -> str:
        instances = [
            item.uri
            for _, item in self.active_instances.items()
            if item.role == role and item.uri is not None
        ]
        if role == Role.PREFILL:
            instance = instances[self._round_robin_index_p]
            self._round_robin_index_p = (self._round_robin_index_p + 1) % len(instances)
        if role == Role.DECODE:
            instance = instances[self._round_robin_index_d]
            self._round_robin_index_d = (self._round_robin_index_d + 1) % len(instances)
        if role == Role.MIXED:
            instance = instances[self._round_robin_index_m]
            self._round_robin_index_m = (self._round_robin_index_m + 1) % len(instances)
        return instance

    def _round_robin_pair(self) -> DispatchResult:
        # current policy: if has mixed, use mixed
        is_pd_disagged = (Role.MIXED not in self.role_2_instances
                          or len(self.role_2_instances.get(Role.MIXED, None)) == 0)
        if not is_pd_disagged:
            mixed_uri = self._round_robin_selection(Role.MIXED)
            return DispatchResult(prefill_vllm_instance_uri=None, decode_vllm_instance_uri=mixed_uri)

        prefill_uri = self._round_robin_selection(Role.PREFILL)
        decode_uri = self._round_robin_selection(Role.DECODE)
        return DispatchResult(prefill_vllm_instance_uri=prefill_uri, decode_vllm_instance_uri=decode_uri)

    async def _monitor_instance_health(self):
        """
        Monitor instance health, report to controller if >20s no response / failed status
        """
        while True:
            self.active_instances = dict(self.instance_infos)
            current_time = time.time()
            instances_to_report = []
            for instance_id, instance_info in self.active_instances.items():
                if instance_id in self.reported_failures:
                    continue
                logger.debug(f"Monitoring ID: %s, Status: %s", instance_id, instance_info.status)

                if instance_info.status == VllmInstanceStatus.HEALTHCHECK_FAILED:
                    logger.warning(f"Instance {instance_id} has failed health check.")
                    instances_to_report.append(instance_info)
                    self.reported_failures.add(instance_id)

                # Consider instance unhealthy if no heartbeat
                elif current_time - self.last_heartbeat.get(instance_id, 0) > 20:
                    logger.warning(f"Instance {instance_id} is unhealthy (no heartbeat).")
                    instances_to_report.append(instance_info)
                    self.reported_failures.add(instance_id)

            if instances_to_report:
                logger.warning(f'instances_to_report is {instances_to_report}')
                await self.report_failures_to_controller(instances_to_report)
                instances_to_report = []

            await asyncio.sleep(5)

    def _get_ready_vllm_actors_num(self):
        """
        Get the number of ready VLLM instances.

        Returns:
            Number of ready VLLM instances.
        """
        return sum(info.status == VllmInstanceStatus.RUNNING for info in self.instance_infos.values())

    def _get_unready_vllm_actors_num(self):
        """
        Get the number of unready VLLM instances.

        Returns:
            Number of unready VLLM instances.
        """
        return sum(info.status != VllmInstanceStatus.RUNNING for info in self.instance_infos.values())
