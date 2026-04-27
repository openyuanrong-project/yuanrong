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

from __future__ import annotations

import logging
import os
import socket
import ssl
import subprocess
import time
import urllib.error
import urllib.request
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Optional

from yr.cli.config import ConfigResolver

logger = logging.getLogger(__name__)


class ComponentStatus(Enum):
    STOPPED = "stopped"
    STARTING = "starting"
    RUNNING = "running"
    HEALTHY = "healthy"
    ERROR = "error"


@dataclass
class ComponentConfig:
    name: str
    enabled: bool = True
    args: list[str] = field(default_factory=list)
    env_vars: dict[str, str] = field(default_factory=dict)
    working_dir: Optional[str] = None
    depends_on: list[str] = field(default_factory=list)
    prepend_char: str = "--"
    env: dict[str, str] = field(default_factory=dict)
    restart_count: int = 0

    # 运行时状态
    status: ComponentStatus = ComponentStatus.STOPPED
    process: Optional[subprocess.Popen] = None
    pid: Optional[int] = None
    start_time: Optional[float] = None
    exit_code: Optional[int] = None


class ComponentLauncher:
    def __init__(self, name: str, resolver: ConfigResolver, config: Optional[ComponentConfig] = None) -> None:
        self.name = name
        self.resolver = resolver
        self.component_config = config or ComponentConfig(name=name)

    def prepare_command(self) -> list[str]:
        executable = self.resolver.rendered_config[self.name]["bin_path"]
        args = [executable]
        self._get_args_from_config(self.name, args, self.component_config.prepend_char)
        return args

    def health_check(self) -> bool:
        """Component-specific health check.

        The base implementation performs no additional checks beyond the
        mandatory process liveness check done in `wait_until_healthy()`.
        Override this method for component-specific health logic (HTTP/file/etc).
        """
        return True

    def prestart_hook(self) -> None:
        pass

    def poststart_hook(self) -> None:
        pass

    def prepare_environment(
        self,
        env: dict[str, str],
    ) -> dict[str, str]:
        config_env = dict(self.resolver.rendered_config[self.component_config.name].get("env", {}))
        if len(env) == 0:
            return config_env

        # Priority: config_env > env.
        # Special case: if both define LD_LIBRARY_PATH/LD_PRELOAD, append config paths to incoming value.
        merged_env: dict[str, str] = env
        for k, v in config_env.items():
            config_value = v
            if k in ["LD_LIBRARY_PATH", "LD_PRELOAD"] and k in merged_env and merged_env[k]:
                merged_env[k] = f"{merged_env[k]}:{config_value}" if config_value else merged_env[k]
            else:
                merged_env[k] = config_value

        return merged_env

    def wait_until_healthy(self) -> bool:
        """
        Block until the component becomes healthy or definitively fails.

        Args:
            overall_deadline: System-level deadline for overall timeout control.

        Returns:
            True if component is healthy, False otherwise.
        """
        timeout = self.resolver.rendered_config[self.component_config.name]["health_check"].get("timeout", 30)

        start = time.time()
        while True:
            # Always gate on process health. If the process is gone, fail fast
            # and do not keep retrying component-specific health checks.
            if not self._check_process_health():
                logger.error(f"{self.name}: process health check failed")
                return False

            if self.health_check():
                logger.info(f"{self.name}: health check passed")
                self.component_config.status = ComponentStatus.HEALTHY
                self.poststart_hook()
                return True

            if time.time() - start >= timeout:
                logger.info(f"{self.name}: health check timeout ({timeout}s)")
                return False

            time.sleep(1)

    def launch(self) -> subprocess.Popen:
        self.prestart_hook()

        cmd = self.prepare_command()
        full_env = os.environ.copy()
        env = self.prepare_environment(full_env)
        self.component_config.args = cmd
        cwd = self.resolver.runtime_context["deploy_path"]
        self.component_config.working_dir = cwd
        self.component_config.env_vars = env

        logger.info(f"Starting {self.name}: {' '.join(cmd)}")
        logger.info(f"Environment: {full_env}")

        log_file = Path(cwd) / "logs" / f"{self.name}_stdout.log"
        with log_file.open("a") as log:
            log.write(f"\n=== Started at {time.ctime()} ===\n")
            log.write(f"Command: {' '.join(cmd)}\n\n")
        stdout_file = log_file.open("a")
        stderr_file = log_file.open("a")

        process = subprocess.Popen(
            cmd,
            env=env,
            cwd=cwd,
            stdout=stdout_file,
            stderr=stderr_file,
        )

        self.component_config.process = process
        self.component_config.pid = process.pid
        self.component_config.start_time = time.time()
        self.component_config.status = ComponentStatus.RUNNING

        return process

    def terminate(self, force: bool = False):
        if not self.component_config.process:
            return

        process = self.component_config.process
        if process.poll() is None:  # process is still running
            if force:
                process.kill()
                logger.info(f"Killed {self.name} (PID: {process.pid})")
            else:
                process.terminate()
                logger.info(f"Terminated {self.name} (PID: {process.pid})")
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                logger.info(f"Force killed {self.name} after timeout")

        self.component_config.status = ComponentStatus.STOPPED
        self.component_config.process = None

    # exec component in current process for entrypoint use case
    # any other specific preparation should be done before call this method
    # component start parameters should be set in config(default /etc/yuanrong/config.toml)
    # such as mkdir -p -m 750 /home/sn/logs && yr exec function_agent
    def exec(self, inherit_env: bool) -> None:
        cmd = self.prepare_command()
        full_env = os.environ.copy() if inherit_env else {}
        env = self.prepare_environment(full_env)

        logger.info(f"Executing {self.name}: {' '.join(cmd)}")
        logger.info(f"Environment: {env}")
        os.execve(
            cmd[0],
            cmd,
            env,
        )

    def restart(self) -> subprocess.Popen:
        """
        Restart the component if monitoring finds it has exited. Component will be re-launched with the
        same command line args and environment as before.
        """
        logger.info(f"Restarting {self.name}")

        cmd = self.component_config.args
        full_env = self.component_config.env_vars
        cwd = self.component_config.working_dir

        log_file = Path(self.resolver.runtime_context["deploy_path"]) / "logs" / f"{self.name}_stdout.log"
        with log_file.open("a") as log:
            log.write(f"\n=== Started at {time.ctime()} ===\n")
            log.write(f"Command: {' '.join(cmd)}\n\n")
        stdout_file = log_file.open("a")
        stderr_file = log_file.open("a")

        process = subprocess.Popen(
            cmd,
            env=full_env,
            cwd=cwd,
            stdout=stdout_file,
            stderr=stderr_file,
        )

        self.component_config.process = process
        self.component_config.pid = process.pid
        self.component_config.start_time = time.time()
        self.component_config.status = ComponentStatus.RUNNING
        return process

    def _check_port_health(self) -> bool:
        if not self.component_config.health_check_port:
            return True

        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(1)
            result = sock.connect_ex(
                ("127.0.0.1", self.component_config.health_check_port),
            )
            sock.close()
        except Exception as e:
            logger.error(f"Exception during port health check: {e}")
            return False
        else:
            return result == 0

    def _check_http_or_https_health(
        self,
        timeout: int = 1,
        node_id: Optional[str] = None,
        pid: Optional[str] = None,
        tls_cacert: Optional[str] = None,
        tls_cert: Optional[str] = None,
        tls_key: Optional[str] = None,
    ) -> bool:
        url = self.resolver.rendered_config[self.component_config.name]["health_check"]["endpoint"]
        headers = {}
        if node_id:
            headers["Node-ID"] = node_id
        if pid:
            headers["PID"] = pid

        if not url.startswith(("http://", "https://")):
            logger.error(f"{self.name}: invalid health check URL: {url}")
            return False

        request = urllib.request.Request(url, headers=headers)

        ssl_context: Optional[ssl.SSLContext] = None
        if url.startswith("https://") and tls_cacert and tls_cert and tls_key:
            ssl_context = ssl.create_default_context(cafile=tls_cacert)
            ssl_context.load_cert_chain(certfile=tls_cert, keyfile=tls_key)
        logger.info(f"{self.name}: performing HTTP health check on {url}")
        try:
            with urllib.request.urlopen(request, timeout=timeout, context=ssl_context) as response:
                status_code = response.getcode()
            return status_code == 200
        except urllib.error.URLError as e:
            if isinstance(e, urllib.error.HTTPError):
                logger.error(f"HTTP error during health check: {e.code} - {e.reason}")
            else:
                logger.error(f"URL error during health check: {e.errno} - {e.reason}")
            return False
        except Exception as e:
            logger.error(f"Exception during HTTP health check: {e}")
            return False

    def _check_process_health(self) -> bool:
        if not self.component_config.process:
            logger.error(f"{self.name}: process not started")
            self.component_config.status = ComponentStatus.ERROR
            return False
        if self.component_config.process.poll() is not None:
            logger.error(f"{self.name}: process exited with code {self.component_config.process.returncode}")
            self.component_config.status = ComponentStatus.STOPPED
            return False
        return True

    def _check_file_health(self) -> bool:
        endpoint = self.resolver.rendered_config[self.component_config.name]["health_check"]["endpoint"]
        if Path(endpoint).exists():
            return True
        logger.error(f"{self.name}: health check file not found at {endpoint}")
        return False

    def _check_custom_health(self) -> bool:
        return True

    def _get_args_from_config(self, component: str, args: list[str], prepend_char: str = "--") -> list[str]:
        config_args = self.resolver.rendered_config[component].get("args", {})
        for k, v in config_args.items():
            if v is None or v == "":
                continue
            value_str = str(v).lower() if isinstance(v, bool) else str(v)
            args.append(f"{prepend_char}{k}={value_str}")

    def _update_args(self, new_config: dict[str, any]) -> None:
        args_config = self.resolver.rendered_config[self.name]["args"]
        for key, value in new_config.items():
            args_config[key] = value
            logger.debug(f"Set {self.name} arg {key} to {value}")
