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
Port forwarding configuration parser.

Converts ``InvokeOptions.port_forwardings`` into a ``createOptions`` dict entry
with key ``"network"`` and a JSON-serialized value describing the ports to forward
into the sandbox.

JSON format::

    {"portForwardings": [{"port": 8080, "protocol": "TCP"}, ...]}

This module follows the same pattern as ``runtime_env.py``: a pure-Python parser
that returns a dict to be merged into ``CInvokeOptions.createOptions`` by ``fnruntime.pyx``.
"""

import json
from typing import Dict

from yr.config import InvokeOptions

NETWORK_KEY = "network"
_SUPPORTED_PROTOCOLS = ("TCP", "UDP")
_MIN_PORT = 1
_MAX_PORT = 65535


def parse_port_forwardings(opt: InvokeOptions) -> Dict:
    """Parse port_forwardings from InvokeOptions into a createOptions dict entry.

    Args:
        opt: The InvokeOptions instance containing ``port_forwardings``.

    Returns:
        An empty dict if no port forwardings are configured, otherwise
        ``{"network": <json_string>}`` where ``<json_string>`` encodes
        ``{"portForwardings": [{"port": <int>, "protocol": <str>}, ...]}``.

    Raises:
        TypeError: If ``port`` is not ``int`` or ``protocol`` is not ``str``.
        ValueError: If ``port`` is outside [1, 65535] or ``protocol`` is not
            ``"TCP"`` or ``"UDP"``.
    """
    if not opt.port_forwardings:
        return {}

    port_forwarding_list = []
    for pf in opt.port_forwardings:
        if not isinstance(pf.port, int):
            raise TypeError(
                f"PortForwarding.port must be int, got {type(pf.port)}"
            )
        if not isinstance(pf.protocol, str):
            raise TypeError(
                f"PortForwarding.protocol must be str, got {type(pf.protocol)}"
            )
        if not _MIN_PORT <= pf.port <= _MAX_PORT:
            raise ValueError(
                f"PortForwarding.port must be in [{_MIN_PORT}, {_MAX_PORT}], got {pf.port}"
            )
        if pf.protocol not in _SUPPORTED_PROTOCOLS:
            raise ValueError(
                f"PortForwarding.protocol must be one of {_SUPPORTED_PROTOCOLS}, got '{pf.protocol}'"
            )
        port_forwarding_list.append({"port": pf.port, "protocol": pf.protocol})

    return {NETWORK_KEY: json.dumps({"portForwardings": port_forwarding_list})}
