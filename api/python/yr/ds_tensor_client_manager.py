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

import os

from yr import log
from yr.config_manager import ConfigManager

data_system_import = True
_import_error = None

try:
    import torch
    import torch_npu
    from datasystem import DsTensorClient
    from torch.npu import current_device
except ImportError as e:
    data_system_import = False
    _import_error = e
_global_tensor_client = None


def _inspect_data_system_address():
    address = os.getenv("YR_DS_ADDRESS")
    if address and len(address) != 0:
        return address
    address = ConfigManager().ds_address
    if len(address) != 0:
        return address
    raise RuntimeError("cannot inspect data system address")


def get_tensor_client():
    if not data_system_import:
        raise RuntimeError(f"import err: {_import_error}")

    global _global_tensor_client
    if _global_tensor_client:
        return _global_tensor_client

    ds_worker_addr = _inspect_data_system_address()
    log.get_logger().info(
        f"start get ds tensor client, ds address is {ds_worker_addr}, current device id is {current_device()}")
    if ds_worker_addr.find(":") == -1:
        raise ValueError(
            "Invalid address of data system worker: {}, expect 'ip:port'".format(ds_worker_addr))
    host, port = ds_worker_addr.split(":")
    port = int(port)
    _global_tensor_client = DsTensorClient(host, port, current_device())
    _global_tensor_client.init()

    return _global_tensor_client
