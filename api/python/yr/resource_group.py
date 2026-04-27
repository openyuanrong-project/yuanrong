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

from dataclasses import dataclass
from typing import Dict, List, Optional
from yr.common import constants
from yr.runtime_holder import global_runtime


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class ResourceGroup:
    """
    The handle returned after creating a ResourceGroup.

    Examples:
        >>> rg = yr.create_resource_group([{"NPU":1},{"CPU":2000,"Memory":2000}], "rgname")
    """

    def __init__(
            self,
            name: '',
            request_id: '',
            bundles: Optional[List[Dict]] = None,
    ):
        """Initialize ResourceGroup."""
        self.name = name
        self.request_id = request_id
        self.bundles = bundles

    @property
    def bundle_specs(self) -> List[Dict]:
        """
        Returns all bundles under the current resource group.

        Returns:
            All bundles under the current resource group.Data type is List[Dict].

        Examples:
            >>> rg = yr.create_resource_group([{"NPU":1},{"CPU":2000,"Memory":2000}], "rgname")
            >>> bundles = rg.bundle_specs
        """
        return self.bundles

    @property
    def bundle_count(self) -> int:
        """
        Returns the number of bundles in the current resource group.

        Returns:
            Number of bundles in the current resource group. Data type is int.

        Examples:
            >>> import yr
            >>> yr.init()
            >>> rg = yr.create_resource_group([{"NPU":1},{"CPU":2000,"Memory":2000}], "rgname")
            >>> count = rg.bundle_count
            >>> print(count)
        """
        return len(self.bundles)

    @property
    def resource_group_name(self) -> str:
        """
        Returns the name of the current resource group.

        Returns:
            Name of the current resource group. Data type is str.

        Examples:
            >>> import yr
            >>> yr.init()
            >>> rg = yr.create_resource_group([{"NPU":1},{"CPU":2000,"Memory":2000}], "rgname")
            >>> name = rg.resource_group_name
            >>> print(name)
        """
        return self.name

    def wait(self, timeout_seconds: Optional[int] = None) -> None:
        """
        Block and wait for the result of creating a ResourceGroup.

        Args:
            timeout_seconds (Optional[int], optional): The timeout time for blocking while waiting for the creation of a
                ResourceGroup result. The default value is None, which is ``-1`` (never timeout). The value range must
                be an integer >= ``-1``. If it equals ``-1``, it will wait indefinitely.

        Raises:
            ValueError: Timeout is less than ``-1``.
    
        Examples:
            >>> rg = yr.create_resource_group([{"NPU":1},{"CPU":2000,"Memory":2000}], "rgname")
            >>> rg.wait(60)
        """
        if timeout_seconds is None:
            timeout_seconds = constants.NO_LIMIT
        if timeout_seconds < constants.NO_LIMIT:
            raise ValueError(
                "Parameter 'timeout_seconds' should be greater than or equal to -1 (no timeout)")
        global_runtime.get_runtime().wait_resource_group(self.name, self.request_id, timeout_seconds)
