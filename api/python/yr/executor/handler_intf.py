#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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

"""handler intf"""

from abc import ABC, abstractmethod
from typing import List

from yr.err_type import ErrorInfo


class HandlerIntf(ABC):
    """
    Handler interface
    """

    @abstractmethod
    def execute_function(self, func_meta, args: List, invoke_type, return_num: int, is_actor_async: bool):
        """
        Execute the function.
        """

    @abstractmethod
    def shutdown(self, grace_period_second: int) -> ErrorInfo:
        """
        Shutdown the instance gracefully.
        Args:
            grace_period_second (int): The time to wait for the instance to shutdown gracefully.
        Returns:
            The ErrorInfo of the execution of shutdown function.

        Raises:
            RuntimeError: If the instance has not been initialized.
        """

    @abstractmethod
    def before_snapshot(self) -> ErrorInfo:
        """
        Trigger snapshot preparation hook before taking snapshot.
        
        This method is called by libruntime before creating a snapshot.
        It invokes the user-defined __yr_before_snapshot__ method if present.
        
        Returns:
            ErrorInfo: Error information if hook execution failed.
        """

    @abstractmethod
    def after_snapstart(self) -> ErrorInfo:
        """
        Trigger snapshot recovery hook after restoring from snapshot.
        
        This method is called by libruntime after restoring from a snapshot.
        It invokes the user-defined __yr_after_snapstart__ method if present.
        
        Returns:
            ErrorInfo: Error information if hook execution failed.
        """
