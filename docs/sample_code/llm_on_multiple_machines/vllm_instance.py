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
import json
import uuid

import subprocess
import sys
import os
import signal
import logging
from asyncio import Task
import yr
import aiohttp

from yr.decorator.instance_proxy import InstanceProxy

from constants import (VLLM_NAMESPACE,
                       VLLM_INSTANCE_HEALTH_CHECK_INTERVAL_S,
                       BALANCER_ACTOR_NAME,
                       MIN_KV_PORT,
                       MAX_KV_PORT,
                       DEFAULT_KV_PORT)
from entities import Role, VllmInstanceInfo, VllmInstanceStatus
from config import VllmInstanceConfig, VllmInstanceDPConfig
from utils import find_node_ip, find_free_port, find_interface_by_ip, find_ip_by_interface

logger = logging.getLogger(__name__)


def select_distributed_torch_interface():
    """
    Determines the preferred network interface for distributed PyTorch communication.

    Args:
        [Function takes no explicit arguments but inspects environment]

        GLOO_SOCKET_IFNAME (str): Environment variable specifying the network interface
                                for GLOO-based communication. Takes precedence over NCCL.
        NCCL_SOCKET_IFNAME (str): Environment variable specifying the interface for
                                NCCL-based communication.

    Returns:
        Optional[str]: Returns either:
            - Value of GLOO_SOCKET_IFNAME if set
            - Value of NCCL_SOCKET_IFNAME if set
            - None if neither is specified
    """
    for env in ["GLOO_SOCKET_IFNAME", "NCCL_SOCKET_IFNAME"]:
        if env in os.environ:
            return os.environ[env]
    return None


class VllmInstance:
    """
    VllmInstance is a vllm engine wrapped by a yr instance, responsibilities:
    1. start vllm api server (and pass some args)
    2. do the health check job (report to Controller if any failure)
    """

    _vllm_instance_config: VllmInstanceConfig
    _vllm_instance_info: VllmInstanceInfo
    #: the actor handle of balancer
    _balancer_handle: InstanceProxy
    _vllm_api_server_process: subprocess.Popen
    _vllm_api_server_health_monitor_task: Task[None]

    def __init__(self, name: str, vllm_config: VllmInstanceConfig):
        """
        Args:
            env: the environment variables pass to subprocess
            exec_cmd: the vllm api server startup command, e.g. ["vllm", "serve", "--a=1", "--b=2"]
        """
        self._vllm_instance_config = vllm_config
        self._vllm_instance_info = VllmInstanceInfo(id=name, uri=None, role=vllm_config.pd_config.role)
        self._balancer_handle = None
        self._vllm_api_server_process = None
        self._vllm_api_server_health_monitor_task = None
        self._env = dict(os.environ)
        self._env["HCCL_IF_BASE_PORT"] = os.environ.get('HCCL_IF_BASE_PORT', "50000")

        self.__has_process_started = False

    async def init_dp_master_ip_port(self):
        """
        if dp config is None, init dp master
        """
        intf = select_distributed_torch_interface()
        if intf:
            ip = find_ip_by_interface(intf)
        else:
            ip = find_node_ip()
            intf = find_interface_by_ip(ip)
        self._env["GLOO_SOCKET_IFNAME"] = intf
        self._env["NCCL_SOCKET_IFNAME"] = intf
        master_port = find_free_port(ip)
        return ip, master_port

    async def init_dp_config(self, dp_config: VllmInstanceDPConfig = None):
        """
        if dp config is None, init dp master
        """
        self._vllm_instance_info.dp_master_ip = dp_config.dp_master_ip
        self._vllm_instance_info.dp_master_port = dp_config.dp_master_port
        self._vllm_instance_config.dp_config = dp_config

    async def initialize(self) -> None:
        """launch subprocess"""

        # normalize and set some env vars
        self._resort_ascend_rt_visible_devices_env()
        self._env["VLLM_USE_V1"] = "1"

        # api server options
        # dp slaves have no http api server
        if self._vllm_instance_config.dp_config.dp_size == 0 or self._vllm_instance_config.dp_config.dp_rank == 0:
            protocal = "http"
            target_iface = os.environ.get("NET_IFACE", None)
            if target_iface is None:
                ip = find_node_ip()
            else:
                ip = find_ip_by_interface(target_iface)
            port = find_free_port()
            self._vllm_instance_info.uri = f"{protocal}://{ip}:{port}"
            self._vllm_instance_config.exec_cmd.extend(["--host", ip, "--port", str(port)])

        # tp, pd, and dp options
        self._vllm_instance_config.exec_cmd.extend(["--tensor-parallel-size", str(self._vllm_instance_config.tp)])
        self._add_pd_command_options()
        self._add_dp_command_options()
        self._add_ep_command_options()
        self._add_env()
        self._env["DS_WORKER_ADDR"] = os.getenv("DATASYSTEM_ADDR", "172.17.0.4:9000")  # for YuanRongConnector

        logger.info(f"initialize with command: {self._vllm_instance_config.exec_cmd}, env:{self._env}")

        self._vllm_api_server_process = subprocess.Popen(
            self._vllm_instance_config.exec_cmd,
            stdout=sys.stdout,
            stdin=sys.stdin,
            stderr=sys.stderr,
            text=True,
            preexec_fn=os.setpgrp,
            env=self._env,
        )

        self._vllm_api_server_health_monitor_task = asyncio.create_task(self._monitor_health())

    async def graceful_shutdown(self, timeout_s=5):
        """terminate"""
        if self._vllm_api_server_process is None:
            return

        try:
            pgid = os.getpgid(self._vllm_api_server_process.pid)
            os.killpg(pgid, signal.SIGTERM)
        except ProcessLookupError:
            logger.warning("process already exited")
            return

        # Another way is "self._vllm_api_server_process.terminate()"
        try:
            self._vllm_api_server_process.wait(timeout_s)
            logger.info("graceful_shutdown successfully!")

        except (TimeoutError, subprocess.TimeoutExpired):
            logger.warning("try to graceful_shutdown wait timeout!")
            pass
        finally:
            if self._vllm_api_server_process.poll() is None:
                # Another way is "self._vllm_api_server_process.kill()"
                os.killpg(pgid, signal.SIGKILL)

    def _resort_ascend_rt_visible_devices_env(self):
        if "ASCEND_RT_VISIBLE_DEVICES" not in os.environ:
            return
        try:
            device_ids = [int(id.strip()) for id in os.environ["ASCEND_RT_VISIBLE_DEVICES"].split(",")]
        except ValueError:
            return
        os.environ["ASCEND_RT_VISIBLE_DEVICES"] = ",".join(map(str, sorted(device_ids)))
        self._env["ASCEND_RT_VISIBLE_DEVICES"] = ",".join(map(str, sorted(device_ids)))

    def _add_pd_command_options(self):
        connector_type = int(os.environ.get("USING_CONNECTOR_TYPE", 0))
        if self._vllm_instance_config.pd_config.is_disaggregated_p_d() and connector_type == 1:
            # P instance kv_port need to be unique
            if self._vllm_instance_config.pd_config.role is Role.PREFILL:
                port_offset = (self._vllm_instance_config.pd_config.pd_rank *
                               self._vllm_instance_config.dp_config.dp_size * self._vllm_instance_config.tp)
                kv_base_port = int(os.getenv("D2D_KV_BASE_PORT", DEFAULT_KV_PORT))
                kv_port = kv_base_port + port_offset
                if kv_port < MIN_KV_PORT or kv_port + (
                        self._vllm_instance_config.dp_config.dp_size * self._vllm_instance_config.tp - 1) > MAX_KV_PORT:
                    raise ValueError("kv_port is not within the port range[1024,65535]")
                logger.info(f"kv_base_port={kv_base_port}, port_offset={port_offset}, "
                            f"pd_rank={self._vllm_instance_config.pd_config.pd_rank}, "
                            f"dp_size={self._vllm_instance_config.dp_config.dp_size}, "
                            f"tp={self._vllm_instance_config.tp}, role={self._vllm_instance_config.pd_config.role}")
            else:
                kv_port = DEFAULT_KV_PORT
            self._vllm_instance_config.exec_cmd.extend(
                [
                    "--kv-transfer-config",
                    json.dumps(
                        {
                            "kv_connector": "YuanRongConnector",
                            "engine_id": str(uuid.uuid4()),
                            "kv_port": kv_port,
                            "kv_role": (
                                "kv_producer"
                                if self._vllm_instance_config.pd_config.role is Role.PREFILL
                                else "kv_consumer"
                            ),
                        }
                    ),
                ]
            )
        else:
            enable_prefix_connector = int(os.environ.get("USING_PREFIX_CONNECTOR", 0))
            if (not self._vllm_instance_config.pd_config.is_disaggregated_p_d()) and enable_prefix_connector:
                self._vllm_instance_config.exec_cmd.extend(
                    [
                        "--kv-transfer-config",
                        json.dumps(
                            {
                                "kv_connector": "YuanRongConnector",
                                "kv_role": (
                                    "kv_both"
                                ),
                            }
                        ),
                    ]
                )
            elif self._vllm_instance_config.pd_config.is_disaggregated_p_d():
                self._vllm_instance_config.exec_cmd.extend(
                    [
                        "--kv-transfer-config",
                        json.dumps(
                            {
                                "kv_connector": "YuanRongConnector",
                                "kv_role": (
                                    "kv_producer"
                                    if self._vllm_instance_config.pd_config.role is Role.PREFILL
                                    else "kv_consumer"
                                ),
                            }
                        ),
                    ]
                )
            else:
                return

    def _add_dp_command_options(self):
        if not self._vllm_instance_config.dp_config.is_dp_enabled():
            return

        self._vllm_instance_config.exec_cmd.extend(
            [
                "--data-parallel-size",
                str(self._vllm_instance_config.dp_config.dp_size),
                "--data-parallel-size-local",
                str(self._vllm_instance_config.dp_config.dp_local_size),
                "--data-parallel-start-rank",
                str(self._vllm_instance_config.dp_config.dp_rank),
                "--data-parallel-address",
                str(self._vllm_instance_config.dp_config.dp_master_ip),
                "--data-parallel-rpc-port",
                str(self._vllm_instance_config.dp_config.dp_master_port),
            ]
        )

        if self._vllm_instance_config.dp_config.dp_rank > 0:
            self._vllm_instance_config.exec_cmd.extend(["--headless"])

    def _add_ep_command_options(self):
        if not self._vllm_instance_config.ep_config.is_ep_enabled():
            return

        self._vllm_instance_config.exec_cmd.extend(
            [
                "--enable-expert-parallel",
            ]
        )

    def _add_env(self):
        if self._vllm_instance_config.env is None:
            return

        env_dict = dict(item.split('=') for item in self._vllm_instance_config.env.split())
        for env_key, env_value in env_dict.items():
            self._env[env_key] = env_value

    async def _monitor_health(self):
        """Asynchronously monitor subprocess health and report to controller"""
        while not self._balancer_handle:
            try:
                self._balancer_handle = yr.get_instance(name=BALANCER_ACTOR_NAME, namespace=VLLM_NAMESPACE)
            except Exception as e:
                logger.warning(f'Instance get _balancer_handle failed, wait for 1 second and retry. e={e}')
                await asyncio.sleep(1)

        await self._balancer_handle.add_vllm_instance.invoke(self._vllm_instance_info)
        async with aiohttp.ClientSession() as session:
            last_report_time = asyncio.get_event_loop().time()
            last_status = self._vllm_instance_info.status
            while True:
                self._vllm_instance_info.status = VllmInstanceStatus.RUNNING
                if self._vllm_api_server_process.poll() is not None:
                    self._vllm_instance_info.status = VllmInstanceStatus.SUBPROCESS_EXITED
                elif self._vllm_instance_info.uri is not None:  # only check DP master's healthy
                    try:
                        async with session.get(
                                f"{self._vllm_instance_info.uri}/health", timeout=aiohttp.ClientTimeout(total=5)
                        ) as response:
                            self._vllm_instance_info.status = (
                                VllmInstanceStatus.HEALTHCHECK_FAILED
                                if response.status != 200
                                else VllmInstanceStatus.RUNNING
                            )
                    except (aiohttp.ClientError, asyncio.TimeoutError):
                        self._vllm_instance_info.status = VllmInstanceStatus.HEALTHCHECK_FAILED
                        logger.warning(f"VllmInstance check timeout!")

                if (
                        # not healthy
                        self._vllm_instance_info.status != VllmInstanceStatus.RUNNING
                        # or changed
                        or self._vllm_instance_info.status != last_status
                        # or past quite long time, we should let controller know that we are still alive
                        or asyncio.get_event_loop().time() - last_report_time > VLLM_INSTANCE_HEALTH_CHECK_INTERVAL_S
                ):
                    await self._balancer_handle.update_vllm_instance_health.invoke([self._vllm_instance_info])
                    last_report_time = asyncio.get_event_loop().time()
                    last_status = self._vllm_instance_info.status

                if self._vllm_instance_info.status == VllmInstanceStatus.SUBPROCESS_EXITED:
                    # terminate self
                    logger.warning(f"vllm subprocess exited unexpectedly, VllmInstance exit with vllm together")
                await asyncio.sleep(5)


def start_vllm_instance(vllm_instance_config: VllmInstanceConfig, name: str = None
                        ) -> InstanceProxy:
    """
    Start a VLLM instance.

    Args:
        vllm_instance_config (VllmInstanceConfig): Configuration for the VLLM instance.
        name (str, optional): Name of the actor. Defaults to None.

    Returns:
        InstanceProxy: Handle to the newly created VLLM instance actor.
    """
    opt = yr.InvokeOptions()
    opt.custom_resources = {"NPU/.+/count": vllm_instance_config.dp_config.dp_local_size * vllm_instance_config.tp}
    opt.name = name

    vllm_instance_actor = yr.instance(VllmInstance).options(opt).invoke(name, vllm_instance_config)
    return vllm_instance_actor
