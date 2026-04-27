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

"""
Faas Runtime utils
"""
import base64
import ctypes
import functools
import json
import socket
import threading
import uuid
from typing import Union

SUPPORTED_RESOURCE = {
    128: 300,
    256: 400,
    512: 600,
    768: 800,
    1024: 1000,
    1280: 1200,
    1536: 1400,
    1792: 1600,
    2048: 1800,
    2560: 2200,
    3072: 2600,
    3584: 3000,
    4096: 3400,
    8192: 6600,
    10240: 8200
}

HOST_NAME = socket.gethostname()
POD_NAME = {'podname': HOST_NAME}
G_TENANT_ID = ""


def encode_base64(data: bytes):
    """Method encode string to base64"""
    base64_result = base64.b64encode(data).decode('utf-8')
    return base64_result


def convert_obj_to_json(obj):
    """Method convert_obj_to_json"""
    return obj.__dict__


def is_instance_type(obj):
    """Method is_instance_type"""
    return hasattr(obj, '__dict__')


def to_json_string(obj, indent=None, sort_keys=False):
    """Method to_json_string"""
    if isinstance(obj, dict):
        return json.dumps(obj, indent=indent, sort_keys=sort_keys)
    return json.dumps(obj, indent=indent, default=convert_obj_to_json, sort_keys=sort_keys)


def generate_request_id() -> str:
    """
    Format: task-188ca8cc-35ea-429c-8a07-68163b47c914
    """
    random_id = str(uuid.uuid4())
    return f"task-{random_id}"


def generate_trace_id() -> str:
    """
    Format: trace-188ca8cc-35ea-429c-8a07-68163b47c914
    """
    random_id = str(uuid.uuid4())
    return f"trace-{random_id}"


def set_trace_id(trace_id: str) -> None:
    """
    Set the trace id, this is unique in one processing thread
    """
    thread_local = threading.local()
    thread_local.trace_id = trace_id


def get_trace_id() -> str:
    """
    Return trace id in this thread, if not set, generate one before return
    """
    thread_local = threading.local()
    if hasattr(thread_local, "trace_id"):
        return thread_local.trace_id
    return generate_trace_id()


def set_tenant_id(tenant_id) -> None:
    """
    Set the tenant id of this instance, unique in this process
    """
    global G_TENANT_ID
    G_TENANT_ID = tenant_id


def get_tenant_id() -> str:
    """
    Return the process unique tenant id
    """
    return G_TENANT_ID


def parse_json_data_to_dict(user_data: Union[str, bytes, bytearray, dict]) -> dict:
    '''This function parses user data in JSON format and returns a dictionary.

    Args:
        user_data (Union[str, bytes, bytearray]): The user data to be parsed,
        which can be a string, bytes, or bytearray.

    Raises:
        RuntimeError: If there is an error during JSON parsing.

    Returns:
        dict: The parsed user data in dictionary format.
    '''
    if isinstance(user_data, dict):
        return user_data
    result = {}
    if user_data in ("", {}, None):
        return result
    try:
        result = json.loads(user_data)
    except Exception as e:
        raise RuntimeError(f"parse user_data error, err: {e}") from e
    return result


def dump_data_to_json_str(user_data: object):
    """This function converts an object to a JSON string.

    Args:
        user_data (object): The object to be converted.

    Raises:
        RuntimeError: If there is an error converting the object to a JSON string.
    """
    try:
        result = json.dumps(user_data)
    except Exception as e:
        raise RuntimeError(f"dump user_data error, err: {e}") from e
    return result


def timeout(sec, check_sec=1):
    """
    timeout decorator
    :param sec: raise TimeoutError after %s seconds
    :param check_sec: retry kill thread per %s seconds
        default: 1 second
    """

    def decorator(func):
        @functools.wraps(func)
        def wrapped_func(*args, **kwargs):

            result, exception = [], []

            def run_user_func():
                try:
                    response = func(*args, **kwargs)
                except TimeoutError:
                    pass
                except SystemExit as e:
                    exception.append(e)
                except BaseException as err:
                    exception.append(err)
                else:
                    if response:
                        result.append(response)
                    else:
                        result.append("")

            thread = TerminableThread(target=run_user_func, daemon=True)
            thread.start()
            thread.join(timeout=sec)

            if thread.is_alive():
                exc = type('TimeoutError', TimeoutError.__bases__, dict(TimeoutError.__dict__))
                thread.terminated(exception_cls=exc, check_sec=check_sec)
                err_msg = f'invoke timed out after {sec} seconds'
                raise TimeoutError(err_msg)
            if exception:
                raise exception[0]
            return result[0]

        return wrapped_func

    return decorator


class UserThreadKiller(threading.Thread):
    """separate thread to kill TerminableThread"""

    def __init__(self, target_thread, exception_cls, check_sec=2.0):
        super().__init__()
        self.user_thread = target_thread
        self.exception_cls = exception_cls
        self.check_sec = check_sec
        self.daemon = True

    def run(self):
        """loop util user code finished or raise exception"""
        while self.user_thread.is_alive():
            ctypes.pythonapi.PyThreadState_SetAsyncExc(ctypes.c_long(self.user_thread.ident),
                                                       ctypes.py_object(self.exception_cls))
            self.user_thread.join(self.check_sec)


class TerminableThread(threading.Thread):
    """a thread that can be stopped by forcing an exception in the execution context"""

    def terminated(self, exception_cls, check_sec=2.0):
        """Method to terminated user thread"""
        if self.is_alive():
            killer = UserThreadKiller(self, exception_cls, check_sec=check_sec)
            killer.start()
