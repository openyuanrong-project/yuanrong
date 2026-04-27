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

"""faas error code"""

import enum


class FaasErrorCode(enum.Enum):
    """faas error code"""
    # NoneError -
    NONE_ERROR = 0
    ENTRY_EXCEPTION = 4001
    FUNCTION_INVOCATION_EXCEPTION = 4002
    # StateContentTooLarge state content is too large
    STATE_CONTENT_TOO_LARGE = 4003
    # ResponseExceedLimit response of user function exceeds the platform limit
    RESPONSE_EXCEED_LIMIT = 4004
    # UndefinedState state is undefined
    UNDEFINED_STATE = 4005
    # HeartBeatFunction Invalid heart beat function of user invalid
    HEARTBEAT_FUNCTION_INVALID = 4006
    # FunctionResultInvalid user function result is invalid
    FUNCTION_RESULT_INVALID = 4007
    # InitializeFunctionError user initialize function error
    INITIALIZE_FUNCTION_ERROR = 4009
    # HeartBeatInvokeError failed to invoke heart beat function
    HEARTBEAT_INVOKE_ERROR = 4010
    # InvokeFunctionTimeout user function invoke timeout
    INVOKE_FUNCTION_TIMEOUT = 4010
    # InitFunctionTimeout user function init timeout
    INIT_FUNCTION_TIMEOUT = 4211
    # RequestBodyExceedLimit request body exceeds limit
    REQUEST_BODY_EXCEED_LIMIT = 4140
    # InitFunctionFail function initialization failed
    INIT_FUNCTION_FAIL = 4201
    # ShutDownFunctionTimeout user function shutdown timeout
    SHUTDOWN_FUNCTION_TIMEOUT = 4202
    # FunctionShutDownError user function failed to shut down
    FUNCTION_SHUTDOWN_ERROR = 4203
