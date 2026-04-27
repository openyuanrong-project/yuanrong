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

import glob
import logging
import socket
import psutil
import yr
from yr.affinity import LabelOperator, OperatorType, AffinityType, AffinityKind, Affinity

logger = logging.getLogger(__name__)


def yr_run_on_every_nodes(func, *args, **kwargs):
    """configuring affinity modes"""
    labels = set([node.get("labels", {})['NODE_ID'][0] for node in yr.resources() if node.get("status", -1) == 0])
    futures = []
    for label in labels:
        operators = [LabelOperator(OperatorType.LABEL_EXISTS, label)]
        affinity = Affinity(AffinityKind.INSTANCE, AffinityType.REQUIRED, operators)
        invoke_opt = yr.InvokeOptions()
        invoke_opt.preferred_priority = True
        invoke_opt.labels = [label]
        invoke_opt.schedule_affinities = [affinity]
        futures.append(yr.instance(func).options(invoke_opt).invoke(*args, **kwargs))
    return yr.get(futures)


def get_num_npus() -> int:
    """get npu number from `/dev/davinci?`"""
    try:
        return len(glob.glob("/dev/davinci[0-9]*"))
    except Exception as e:
        logger.error(f"Failed to get npu number! exception: %s.", str(e))
        pass
    return 0


def find_node_ip() -> str:
    """
    IP address by which the local node can be reached.

    Returns:
        The IP address by which the local node can be reached from the address.
    """
    node_ip_address = "127.0.0.1"

    try:
        # try to get node ip address from host name
        host_name = socket.getfqdn(socket.gethostname())
        node_ip_address = socket.gethostbyname(host_name)
    except Exception as e:
        logger.error(f"find node ip error, host_name: %s.", str(e))
        pass

    return node_ip_address


def find_free_port(address: str = "") -> str:
    """
    find one free port

    Returns:
        port
    """
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind((address, 0))
        s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        return str(s.getsockname()[1])


def find_interface_by_ip(ip_address):
    """
    Find the network interface name associated with the given IP address.

    Args:
        ip_address (str): The IP address to look up (e.g., "192.168.1.100").

    Returns:
        str: The name of the matching network interface (e.g., "eth0" or "wlan0"), or None if not found.
    """
    interfaces = psutil.net_if_addrs()

    for interface_name, addresses in interfaces.items():
        for address in addresses:
            if address.family == socket.AF_INET and address.address == ip_address:
                return interface_name

    # Return None if no match is found
    return None


def find_ip_by_interface(interface_name: str):
    """
    Find the IP address associated with the given network interface name.

    Args:
        interface_name (str): The name of the network interface (e.g., "eth0", "wlan0").

    Returns:
        str: The IP address associated with the interface, or None if not found.
    """
    # Get all network interfaces and their addresses
    interfaces = psutil.net_if_addrs()

    # Check if the interface exists
    if interface_name not in interfaces:
        return None

    # Determine the address family (IPv4 or IPv6)
    family = socket.AF_INET  # IPv6: 10 (AF_INET6), IPv4: 2 (AF_INET)

    # Iterate through the addresses of the specified interface
    for address in interfaces[interface_name]:
        if address.family == family:
            return address.address

    # Return None if no matching IP address is found
    return None
