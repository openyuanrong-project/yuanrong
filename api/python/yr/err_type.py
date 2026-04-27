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

"""err type"""

from dataclasses import dataclass
import enum


class ErrorCode(enum.Enum):
    """error code"""
    ERR_OK = 0

    ERR_PARAM_INVALID = 1001
    ERR_RESOURCE_NOT_ENOUGH = 1002
    ERR_INSTANCE_NOT_FOUND = 1003
    ERR_INSTANCE_DUPLICATED = 1004
    ERR_INVOKE_RATE_LIMITED = 1005
    ERR_RESOURCE_CONFIG_ERROR = 1006
    ERR_INSTANCE_EXITED = 1007
    ERR_EXTENSION_META_ERROR = 1008

    ERR_USER_CODE_LOAD = 2001
    ERR_USER_FUNCTION_EXCEPTION = 2002

    ERR_REQUEST_BETWEEN_RUNTIME_BUS = 3001
    ERR_INNER_COMMUNICATION = 3002
    ERR_INNER_SYSTEM_ERROR = 3003
    ERR_DISCONNECT_FRONTEND_BUS = 3004
    ERR_ETCD_OPERATION_ERROR = 3005
    ERR_BUS_DISCONNECTION = 3006
    ERR_REDIS_OPERATION_ERROR = 3007

    ERR_INCORRECT_INIT_USAGE = 4001
    ERR_INIT_CONNECTION_FAILED = 4002
    ERR_DESERIALIZATION_FAILED = 4003
    ERR_INSTANCE_ID_EMPTY = 4004
    ERR_GET_OPERATION_FAILED = 4005
    ERR_INCORRECT_FUNCTION_USAGE = 4006
    ERR_INCORRECT_CREATE_USAGE = 4007
    ERR_INCORRECT_INVOKE_USAGE = 4008
    ERR_INCORRECT_KILL_USAGE = 4009

    ERR_ROCKSDB_FAILED = 4201
    ERR_SHARED_MEMORY_LIMITED = 4202
    ERR_OPERATE_DISK_FAILED = 4203
    ERR_INSUFFICIENT_DISK_SPACE = 4204
    ERR_CONNECTION_FAILED = 4205
    ERR_KEY_ALREADY_EXIST = 4206
    ERR_DEPENDENCY_FAILED = 4306
    ERR_DATASYSTEM_FAILED = 4299

    ERR_FINALIZED = 9000
    ERR_CREATE_RETURN_BUFFER = 9001

    ERR_GENERATOR_FINISHED = 9005


class ModuleCode(enum.Enum):
    """module code"""
    CORE = 10
    RUNTIME = 20
    RUNTIME_CREATE = 21
    RUNTIME_INVOKE = 22
    RUNTIME_KILL = 23
    DATASYSTEM = 30


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class ErrorInfo:
    """
    error info
    Stores runtime error information.

    Attributes:
        error_code: error code, default is ErrorCode.ERR_OK
        module_code: module code, default is ModuleCode.RUNTIME
        msg: error message, default is empty string
    """
    error_code: ErrorCode = ErrorCode.ERR_OK
    module_code: ModuleCode = ModuleCode.RUNTIME
    msg: str = ""
