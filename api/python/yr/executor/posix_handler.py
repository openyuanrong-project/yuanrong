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

"""posix handler"""

from typing import List

from yr.err_type import ErrorCode, ErrorInfo, ModuleCode


class PosixHandler:
    """
    Posix handler class
    """

    def __init__(self):
        pass

    def execute_function(self, func_meta, args: List, invoke_type, return_num: int, is_actor_async: bool):
        """
        Execute the function
        """
        return [], ErrorInfo(
            ErrorCode.ERR_EXTENSION_META_ERROR,
            ModuleCode.RUNTIME,
            "Api type 'Posix' is not supported currently",
        )

    def shutdown(self, grace_period_second: int) -> ErrorInfo:
        """
        Shutdown handler.
        """
