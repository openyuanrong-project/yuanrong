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

"""constants"""

# http headers
from enum import IntEnum

HEADER_CONTENT_TYPE = "Content-Type"
HEADER_CONTENT_TYPE_APPLICATION_JSON = "application/json"
HEADER_CPU = "X-Instance-CPU"
HEADER_MEMORY = "X-Instance-Memory"
HEADER_TRACE_ID = 'X-Trace-ID'
HEADER_EVENT_SOURCE = "X-Event-Source-Id"
HEADER_INVOKE_URN = "X-Tag-VersionUrn"

FUNC_NAME = "functionName"
MODULE_NAME = "moduleName"
CLASS_NAME = "className"
TARGET_LANGUAGE = "targetLanguage"
SRC_LANGUAGE = "srcLanguage"
INSTANCE_ID = "instanceID"
FUNCTION_KEY = "functionKey"
CLASS_METHOD = "classMethod"
NEED_ORDER = "needOrder"
BASE_CLS = "basecls"
IS_ASYNC = "isAsync"
DEFAULT_INVOKE_TIMEOUT = 900
DEFAULT_GET_TIMEOUT = 5 * 60
DEFAULT_WAIT_RGROUP_TIMEOUT = 30
NO_LIMIT = -1
MIN_TIMEOUT_LIMIT = 0
INDEX_FIRST = 0
INDEX_SECOND = 1

SEPARATOR_SLASH = "/"

ENV_KEY_ENV_DELEGATE_DOWNLOAD = "ENV_DELEGATE_DOWNLOAD"
ENV_KEY_LD_LIBRARY_PATH = "LD_LIBRARY_PATH"
ENV_KEY_FUNCTION_LIBRARY_PATH = "YR_FUNCTION_LIB_PATH"

PATTERN_FAAS_ENTRY = r'^[^/]*\.[^/]*$'  # conatains only one '.' and without '/'
KEY_USER_INIT_ENTRY = "userInitEntry"
KEY_USER_CALL_ENTRY = "userCallEntry"
KEY_USER_SHUT_DOWN_ENTRY = "userShutDownEntry"

STACK_TRACE_INFO_TYPE = "YRInvokeError"
STACK_TRACE_INFO_LANGUAGE = "PYTHON"

DEFAULT_NO_TTL_LIMIT = 0
METALEN = 16
META_PREFIX = "0000000000000000"


class Metadata(IntEnum):
    """
    Metadata type
    """
    CROSS_LANGUAGE = 1
    PYTHON = 2
    BYTES = 3
    MEMORYVIEW = 4
    BYTEARRAY = 5
