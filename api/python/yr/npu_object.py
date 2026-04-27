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

"""ObjectRef"""
from yr.exception import YRInvokeError, YRError, GeneratorFinished
from yr.err_type import ErrorInfo, ErrorCode

import yr
from yr import log
from yr.common import constants


class NpuObject:
    """ObjectRef"""

    def __init__(self, npu_id: str, size, dtype, device, ds_address, instance_id):
        """Initialize the ObjectRef."""
        self._id = npu_id
        self._size = size
        self._dtype = dtype
        self._device = device
        self._ds_address = ds_address
        self._instance_id = instance_id

    def __del__(self):
        pass

    def __copy__(self):
        return self

    def __deepcopy__(self, memo):
        return self

    def __str__(self):
        return self.id

    def __eq__(self, other):
        return self.id == other.id

    def __hash__(self):
        return hash(self.id)

    def __repr__(self):
        return f"NpuObject(id={self._id}, size={self._size}, dtype={self._dtype}, deviece={self._device})"

    @property
    def id(self):
        """npu_obj id."""
        return self._id

    @property
    def size(self):
        """npu_obj size."""
        return self._size

    @property
    def dtype(self):
        """npu_obj size."""
        return self._dtype

    @property
    def instance_id(self):
        """npu_obj size."""
        return self._instance_id
