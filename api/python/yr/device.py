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

"""device providing device buffer function."""
from enum import Enum
from typing import List
from yr import runtime_holder
from yr.common.utils import Validator as validator


class DataType(Enum):
    """
    Features: HCCL data type
    """

    DATA_TYPE_INT8 = 0
    DATA_TYPE_INT16 = 1
    DATA_TYPE_INT32 = 2
    DATA_TYPE_INT64 = 3

    DATA_TYPE_UINT8 = 4
    DATA_TYPE_UINT16 = 5
    DATA_TYPE_UINT32 = 6
    DATA_TYPE_UINT64 = 7

    DATA_TYPE_FP16 = 8
    DATA_TYPE_FP32 = 9
    DATA_TYPE_FP64 = 10

    DATA_TYPE_BFP16 = 11
    DATA_TYPE_CP64 = 12
    DATA_TYPE_CP128 = 13


class DeviceBufferLifetimeType(Enum):
    """
    DeviceBuffer life time type
    """
    REFERENCE = 0
    MOVE = 1


class DataInfo:
    """
    The DataInfo class is used to store the basic information of data.
    """

    def __init__(self, dev_ptr: int, data_type: DataType, count: int, device_idx: int):
        """
        Initialize the DataInfo object.
        """
        self.dev_ptr = dev_ptr
        self.data_type = data_type
        self.count = count
        self.device_idx = device_idx


class DeviceBufferParam:
    """
    DeviceBufferParam device buffer param
    """

    def __init__(self, lifetime: DeviceBufferLifetimeType, cache_location):
        self.lifetime = lifetime
        self.cache_location = cache_location


def g_increase_ref(object_ids: List[str]) -> None:
    """Increase the global reference of the given objects.

    Args:
        object_ids(list): The ids of the objects to be increased. It cannot be empty.

    Returns:
        failed_object_ids(list): list of failed object id.

    Raises:
        TypeError: Raise a type error if the input parameter is invalid.
        RuntimeError: Raise a runtime error if the increase fails for all objects.
    """
    args = [["object_ids", object_ids, list]]
    validator.check_args_types(args)
    if not object_ids:
        raise RuntimeError(r"The input of object_ids list should not be empty")
    runtime_holder.global_runtime.get_runtime().increase_global_reference(object_ids)


def g_decrease_ref(object_ids: List[str]) -> None:
    """Decrease the global reference of the given objects.

    Args:
        object_ids(list): The ids of the objects to be decreased. It cannot be empty.

    Returns:
        None.

    Raises:
        TypeError: Raise a type error if the input parameter is invalid.
        RuntimeError: Raise a runtime error if the decrease fails for all objects.
    """
    args = [["object_ids", object_ids, list]]
    validator.check_args_types(args)
    if not object_ids:
        raise RuntimeError(r"The input of object_ids list should not be empty")
    runtime_holder.global_runtime.get_runtime().decrease_global_reference(object_ids)
