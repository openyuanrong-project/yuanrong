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

import yr
import asyncio
import os

SHUTDOWN_KEY = "key"
SHUTDOWN_VALUE = b"value_shutdown"
CUSTOM_SHUTDOWN_OPT = yr.InvokeOptions()
CUSTOM_SHUTDOWN_OPT.custom_extensions = {"GRACEFUL_SHUTDOWN_TIME": "1"}


@yr.instance
class Add:
    """Add class"""

    def __init__(self):
        self.cnt = 0

    def add(self):
        """Adds one to cnt"""
        self.cnt += 1
        return self.cnt

    def get(self):
        """Gets cnt"""
        return self.cnt

    def return_custom_envs(self, key):
        return os.getenv(key)

    def save(self, timeout=30):
        """ "Save stateful function state"""
        yr.save_state(timeout)

    def load(self, timeout=30):
        """Load stateful function state"""
        yr.load_state(timeout)
        return "ok"

    def __yr_shutdown__(self, period):
        yr.kv_write(SHUTDOWN_KEY, SHUTDOWN_VALUE)
        print("Class Add shutdown finished")


@yr.instance
class AsyncAdd:
    """AsyncAdd class"""

    def __init__(self):
        self.cnt = 1

    async def get(self):
        await asyncio.sleep(2)
        return self.cnt


@yr.invoke
def get():
    """get"""
    return 1


@yr.invoke
def return_custom_envs(key):
    return os.getenv(key)


@yr.invoke
def may_throw_exception(num: int) -> int:
    if num % 2 == 0:
        raise RuntimeError("num cannot be divisible by 2.")
    return num


KB = 1024
MB = KB * KB
GB = MB * KB
TB = GB * KB


def format_byte_repr(byte_num):
    """
    size转换
    :param byte_num: 单位Byte
    :return:
    """
    try:
        if isinstance(byte_num, str):
            byte_num = int(byte_num)
        if byte_num > TB:
            result = '%s TB' % round(byte_num / TB, 2)
        elif byte_num > GB:
            result = '%s GB' % round(byte_num / GB, 2)
        elif byte_num > MB:
            result = '%s MB' % round(byte_num / MB, 2)
        elif byte_num > KB:
            result = '%s KB' % round(byte_num / KB, 2)
        else:
            result = '%s B' % byte_num
        return result
    except Exception as e:
        print(e.args)
        return byte_num


def add_one(num):
    return num + 1


class SimpleInstance:
    def add_one(self, num):
        return num + 1


@yr.instance
class Counter:
    def __init__(self):
        self.number = 0

    def add(self):
        self.number = self.number + 1
        return self.number


class FccData:
    a: int
    b: str

    def __init__(self, a: int, b: str):
        self.a = a
        self.b = b


@yr.instance
class FccDemo:
    def __init__(self, name):
        self.name = name

    def exception(self):
        raise RuntimeError("raise exception")

    def get(self, data: FccData) -> FccData:
        return data

    async def get_async(self, data: FccData) -> FccData:
        return data


def init_in_process(function_id, server_address, ds_address):
    try:
        conf = yr.Config(
            function_id=function_id,
            server_address=server_address,
            ds_address=ds_address,
            log_level="DEBUG",
            recycle_time=300,
        )

        yr.init(conf)
        counter = yr.get_instance('counter-actor')
        res = yr.get(counter.add.invoke())
        print("res of counter-actor:", res)
        return "init yr succeed"
    except Exception as e:
        raise RuntimeError(f"exec initand invoke failed {str(e)}")
