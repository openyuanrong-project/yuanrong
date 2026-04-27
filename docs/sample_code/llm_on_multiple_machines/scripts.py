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

from typing import List, Optional
import logging
import shlex
import yr
import click
from controller import Controller
from endpoint import deploy_endpoint_to_cluster
from log import setup_logging
from constants import VLLM_NAMESPACE, CONTROLLER_ACTOR_NAME, YR_LOG_LEVEL
from entities import SchedulerPolicy, DeployConfig
from config import ControllerConfig, InferenceInstanceConfig

setup_logging()
logger = logging.getLogger(__name__)


@click.group()
def cli():
    """Cluster Management"""
    pass


@cli.command(name="deploy", context_settings={"show_default": True})
@click.option("--head-ip", type=str, help='IP of YR head node (e.g. "10.2.3.4")', default="auto")
@click.option("--prefill-instances-num", type=int, help="the num of Prefill instances", default=1)
@click.option(
    "--prefill-startup-params",
    type=str,
    help="the Prefill instance start up command",
    default="vllm serve /workspace/models/qwen2.5_7B",
    callback=lambda ctx, param, value: shlex.split(value),
)
@click.option(
    "--prefill-startup-env",
    type=str,
    help="the Prefill instance start up env",
    default=None,
)
@click.option("--prefill-data-parallel-size", "-pdp", type=int, help="the dp of Prefill instances", default=1)
@click.option("--prefill-tensor-parallel-size", "-ptp", type=int, help="the tp of Prefill instances", default=1)
@click.option(
    "--prefill-expert-parallel-size",
    "-pep",
    type=int,
    help="the ep of Prefill instances, should be equal to dp*tp, 0 means disable expert parallelism",
    default=0,
)
@click.option("--decode-instances-num", type=int, help="the num of Decode instances", default=1)
@click.option(
    "--decode-startup-params",
    type=str,
    help="the Decode instance start up command",
    default="vllm serve /workspace/models/qwen2.5_7B",
    callback=lambda ctx, param, value: shlex.split(value),
)
@click.option(
    "--decode-startup-env",
    type=str,
    help="the decode instance start up env",
    default=None,
)
@click.option("--decode-data-parallel-size", "-ddp", type=int, help="the dp of Decode instances", default=1)
@click.option("--decode-tensor-parallel-size", "-dtp", type=int, help="the tp of Decode instances", default=1)
@click.option(
    "--decode-expert-parallel-size",
    "-dep",
    type=int,
    help="the ep of Decode instances, should be equal to dp*tp, 0 means disable expert parallelism",
    default=0,
)
@click.option(
    "--scheduler-policy",
    type=click.Choice([e.name for e in SchedulerPolicy], case_sensitive=False),
    help="the scheduling policy, default to RoundRobin",
    default=SchedulerPolicy.ROUND_ROBIN.name,
    callback=lambda ctx, param, value: SchedulerPolicy[value.upper()],
)
@click.option(
    "--timeout",
    type=int,
    help="Timeout in seconds for resource allocation or deployment, default to 3600",
    default=3600,
)
@click.option("--proxy-host", type=str, help="the service listening host", default="127.0.0.1")
@click.option("--proxy-port", type=int, help="the service listening port", default=8000)
def deploy(
        head_ip: str,
        prefill_instances_num: int,
        prefill_startup_params: List[str],
        prefill_startup_env: Optional[str],
        prefill_data_parallel_size: int,
        prefill_tensor_parallel_size: int,
        prefill_expert_parallel_size: int,
        decode_instances_num: int,
        decode_startup_params: List[str],
        decode_startup_env: Optional[str],
        decode_data_parallel_size: int,
        decode_tensor_parallel_size: int,
        decode_expert_parallel_size: int,
        scheduler_policy: SchedulerPolicy,
        timeout: int,
        proxy_host: str,
        proxy_port: int,
):
    """deploy"""
    _inner_deploy(DeployConfig(
        head_ip,
        prefill_instances_num,
        prefill_startup_params,
        prefill_startup_env,
        prefill_data_parallel_size,
        prefill_tensor_parallel_size,
        prefill_expert_parallel_size,
        decode_instances_num,
        decode_startup_params,
        decode_startup_env,
        decode_data_parallel_size,
        decode_tensor_parallel_size,
        decode_expert_parallel_size,
        scheduler_policy,
        timeout,
        proxy_host,
        proxy_port)
    )


def _inner_deploy(
        deploy_config: DeployConfig
):
    config = ControllerConfig(
        scheduler_policy=deploy_config.scheduler_policy,
        prefill_instances_num=deploy_config.prefill_instances_num,
        p_inference_instance_config=InferenceInstanceConfig(
            startup_params=deploy_config.prefill_startup_params,
            startup_env=deploy_config.prefill_startup_env,
            dp=deploy_config.prefill_data_parallel_size,
            tp=deploy_config.prefill_tensor_parallel_size,
            ep=deploy_config.prefill_expert_parallel_size,
        ),
        decode_instances_num=deploy_config.decode_instances_num,
        d_inference_instance_config=InferenceInstanceConfig(
            startup_params=deploy_config.decode_startup_params,
            startup_env=deploy_config.decode_startup_env,
            dp=deploy_config.decode_data_parallel_size,
            tp=deploy_config.decode_tensor_parallel_size,
            ep=deploy_config.decode_expert_parallel_size,
        ),
        timeout=deploy_config.timeout,
    )
    logger.info(f"deploy_config:{deploy_config}")

    """Deploy to YR cluster"""
    try:
        logger.info(f"Deploying to YR cluster with config: {config}")
        yr.init(yr.Config(log_level=YR_LOG_LEVEL))
    except Exception as e:
        logger.warning(f"Failed to connect YR cluster: {str(e)}", )
        return

    should_start_controller = False
    try:
        controller = yr.get_instance(name=CONTROLLER_ACTOR_NAME, namespace=VLLM_NAMESPACE)
        logger.info(
            "There is already an controller running in the cluster, please clean before deploy again"
        )
    except RuntimeError:
        should_start_controller = True

    if not should_start_controller:
        return

    logger.info("No existing Controller found, creating new instance")
    opt = yr.InvokeOptions()
    opt.name = CONTROLLER_ACTOR_NAME
    opt.namespace = VLLM_NAMESPACE
    opt.custom_extensions["lifecycle"] = "detached"
    controller = yr.instance(Controller).options(opt).invoke(config)
    try:
        yr.get(controller.initialize.invoke(), -1)
    except Exception as e:
        logger.warning(f"Controller actor create failed : {str(e)}")
        return

    logger.info("Controller actor created.")

    try:
        deploy_endpoint_to_cluster(deploy_config.proxy_host, deploy_config.proxy_port)
    except Exception as e:
        logger.warning(f"Deployment failed: {str(e)}")
        terminate_controller()


@cli.command("clean", context_settings={"show_default": True})
def clean():
    """Clean up deployment from YR cluster"""
    _inner_clean()


def _inner_clean():
    try:
        logger.info("Connecting to existing YR cluster")
        yr.init(yr.Config(log_level=YR_LOG_LEVEL))
    except Exception as e:
        logger.warning(f"Failed to connect YR cluster: {str(e)}")
        return
    terminate_controller()


def terminate_controller():
    """
    terminate controller
    """
    controller = None
    try:
        controller = yr.get_instance(name=CONTROLLER_ACTOR_NAME, namespace=VLLM_NAMESPACE)
        logger.info("Found existing Controller actor, attempting to kill it")
        yr.get(controller.graceful_shutdown.invoke(), -1)
    except TimeoutError:
        logger.warning("Timeout when graceful terminate controller, will try to clean by force")
    except RuntimeError:
        logger.warning("No existing Controller actor found, nothing to clean")
    except Exception as e:
        logger.warning(f"Failed to clean up controller {e}")
    finally:
        if controller:
            controller.terminate()


if __name__ == "__main__":
    cli()
