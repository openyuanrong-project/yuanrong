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
from typing import Optional
from yr.resource_group import ResourceGroup
from yr.fnruntime import write_to_cbuffer


class RgObjectRef:
    """RgObjectRef"""

    def __init__(self, resource_group: ResourceGroup, data):
        self.__resource_group = resource_group
        self._data = write_to_cbuffer(data)

    @property
    def resource_group(self):
        """Returns the resource group of the current resource object ref."""
        return self.__resource_group

    @property
    def data(self):
        """Returns the data of the current resource object ref."""
        return self._data

    def wait(self, timeout_seconds: Optional[int] = 60):
        """wait the group"""
        return self.__resource_group.wait(timeout_seconds)
