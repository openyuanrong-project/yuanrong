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

"""faas handler"""

import logging
from typing import List

from yr.common.utils import err_to_str
from yr.err_type import ErrorCode, ErrorInfo, ModuleCode
from yr.executor.faas_executor import faas_shutdown_handler
from yr.executor.handler_intf import HandlerIntf

_logger = logging.getLogger(__name__)


class FaasHandler(HandlerIntf):
    """
    FaaS handler
    """
    def __init__(self):
        pass

    def execute_function(self, func_meta, args: List, invoke_type, return_num: int, is_actor_async: bool):
        """execute function"""
        pass

    def shutdown(self, grace_period_second: int) -> ErrorInfo:
        """shutdown"""
        try:
            return faas_shutdown_handler(grace_period_second)
        except Exception as e:
            _logger.exception(e)
            return ErrorInfo(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME, err_to_str(e))

    def before_snapshot(self) -> ErrorInfo:
        """FaaS path has no user instance hooks; snapshot is handled outside user code."""
        return ErrorInfo()

    def after_snapstart(self) -> ErrorInfo:
        """FaaS path has no user instance hooks after snapstart."""
        return ErrorInfo()
