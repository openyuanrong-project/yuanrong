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

import errno
import json
import logging
import os
import random
import socket
import ssl
import string
import sys
import time
import urllib.request
from contextlib import closing
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)


def get_ip() -> str:
    hostname = socket.gethostname()
    ip_list = socket.getaddrinfo(hostname, None)

    for ip_info in ip_list:
        ip = ip_info[4][0]
        if ip != "127.0.0.1" and not ip.startswith("::") and "." in ip:
            return ip

    logger.error("Failed to get IP address")
    sys.exit(1)


def trim_hostname() -> str:
    hostname = get_hostname()
    length = len(hostname)
    if length == 0:
        random_hostname = "".join(random.choices(string.ascii_letters, k=6))
        return f"node-{random_hostname}"
    return hostname


def get_hostname() -> str:
    try:
        return os.uname().nodename
    except AttributeError:
        return socket.gethostname()
    except Exception:
        logger.exception("get host name failed")
        return ""


def is_valid_port(port: int) -> bool:
    return 1 <= port <= 65535


def is_port_available(port: int, host: str) -> bool:
    if not is_valid_port(port):
        return False
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(0.5)
            s.bind((host, port))
    except OSError:
        return False
    else:
        return True


def port_or_free(port: str) -> int:
    if is_port_available(int(port), ""):
        return int(port)
    with closing(socket.socket(socket.AF_INET, socket.SOCK_STREAM)) as s:
        s.bind(("", 0))
        new_port = s.getsockname()[1]
        logger.info(f"port {port} is occupied, use {new_port} instead")
        return new_port


def is_valid_file_path(path: str, allow_empty: bool = True) -> bool:
    if not path:
        return allow_empty
    return Path(path).exists()


def read_int(path: Path) -> int:
    try:
        with path.open() as f:
            value = f.read().strip()
        if value.isdigit():
            return int(value)
    except Exception:
        return -1
    return -1


def get_cgroup_paths() -> tuple[str, str]:
    mem_v1 = ""
    unified = ""
    try:
        with Path("/proc/self/cgroup").open() as f:
            for line in f:
                parts = line.strip().split(":")
                if len(parts) != 3:
                    continue
                _, controllers, path = parts
                if controllers == "":
                    unified = path
                else:
                    ctrls = controllers.split(",")
                    if "memory" in ctrls:
                        mem_v1 = path
    except Exception:
        return "", ""
    return mem_v1, unified


def get_cgroup_limit_bytes() -> int:
    mem_v1_path, unified_path = get_cgroup_paths()

    # cgroup v2: unified hierarchy
    if unified_path:
        limit_path = Path("/sys/fs/cgroup").resolve() / unified_path.lstrip("/") / "memory.max"
        try:
            with limit_path.open() as f:
                value = f.read().strip()
            if value == "max":
                return 0
            if value.isdigit():
                logger.debug(f"cgroup v2 memory limit: {value} bytes")
                return int(value)
        except Exception as e:
            logger.warning(f"Failed to read cgroup v2 memory limit: {e}")

    # cgroup v1: memory controller
    if mem_v1_path:
        limit_path = Path("/sys/fs/cgroup/memory").resolve() / mem_v1_path.lstrip("/") / "memory.limit_in_bytes"
        limit = read_int(limit_path)
        logger.debug(f"cgroup v1 memory limit: {limit} bytes")
        if limit <= 0:
            return 0
        if limit >= (1 << 60):
            return 0
        return limit

    return 0


def get_memtotal_mb_from_proc() -> int:
    with Path("/proc/meminfo").open() as f:
        meminfo = f.read()
    total_kb = 0
    for line in meminfo.split("\n"):
        if line.startswith("MemTotal:"):
            total_kb = int(line.split(":")[1].strip().replace(" kB", ""))
            logger.debug(f"MemTotal from /proc/meminfo: {total_kb} kB")
    return total_kb // 1024


def get_total_memory_mb() -> int:
    limit_bytes = get_cgroup_limit_bytes()
    if limit_bytes <= 0:
        return get_memtotal_mb_from_proc()
    return limit_bytes // 1024 // 1024


def wait_pid_exit(pid: int, deadline: float, interval: int = 2) -> bool:
    """Poll until PID exits or timeout. Returns True if exited."""
    logger.info(f"Waiting for PID {pid} to exit...")
    while time.time() < deadline:
        try:
            os.kill(pid, 0)
        except OSError as e:
            if e.errno == errno.ESRCH:
                return True
            logger.exception(f"Error checking PID {pid}")
            return True
        time.sleep(interval)
    return False


def fetch_url(
    url: str,
    timeout: float = 2.0,
    ssl_context: Optional[ssl.SSLContext] = None,
    as_json: bool = False,
) -> Optional[any]:
    try:
        req = urllib.request.Request(url, headers={"Accept": "application/json"})  # noqa: S310
        with urllib.request.urlopen(req, timeout=timeout, context=ssl_context) as resp:  # noqa: S310
            content = resp.read().decode("utf-8", errors="replace")
            return json.loads(content) if as_json else content.strip()
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as exc:
        logger.warning("  Failed to fetch %s: %s", url, exc)
        return None


def service_discovery_from_function_master(
    function_master_addr: str,
    ssl_context: Optional[ssl.SSLContext] = None,
) -> dict[str, any]:
    """Get service discovery info from function_master's /masterinfo endpoint."""

    url = f"{function_master_addr}/global-scheduler/masterinfo"
    data = fetch_url(url, timeout=3.0, as_json=True, ssl_context=ssl_context)
    if not isinstance(data, dict):
        raise ValueError(f"Invalid service discovery response from {url}")

    # Example response:
    # {
    #     "master_address": "10.88.0.4:22770",
    #     "meta_store_address": "10.88.0.4:32379",
    #     "schedule_topo": {
    #         "members": [
    #         {
    #             "address": "10.88.0.4:22772",
    #             "name": "56de329657e6-21"
    #         }
    #         ]
    #     }
    # }
    services = {
        "function_master": data["master_address"],
        "meta_store": data["meta_store_address"],
        "function_proxy": [child["address"] for child in data["schedule_topo"]["members"]],
    }
    logger.info(f"Discovered services: {services}")
    return services


def services_to_overrides(services: dict[str, str]) -> list[str]:
    """Convert service discovery info to a list of command-line overrides."""
    overrides = []
    if "function_master" in services:
        ip, port = services["function_master"].split(":")
        overrides.append(f"values.function_master.ip='{ip}'")
        overrides.append(f"values.function_master.global_scheduler_port='{port}'")
    if "meta_store" in services:
        ip, port = services["meta_store"].split(":")
        overrides.append(
            f'values.etcd.address=[{{ip="{ip}",peer_port="{port}",port="{port}"}}]'
        )  # peer_port is unnecessary so we just set it to the same as port
    overrides.append("ds_worker.args.master_address=''")
    return overrides
