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

import time
import os

import pytest

import yr
from common import FccData, FccDemo


@pytest.mark.smoke
def test_fcc_create_function_group_by_function(init_yr):
    @yr.invoke
    def demo_func_wrapper(name):
        return name

    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
    )
    name = "function_demo"
    objs = yr.fcc.create_function_group(
        demo_func_wrapper, args=(name,), group_size=4, options=opts)
    rets = yr.get(objs)
    assert rets == [name for i in range(1, 5)]


@pytest.mark.smoke
def test_fcc_multi_return_value_function(init_yr):
    @yr.invoke(return_nums=2)
    def demo_func_wrapper(name1, name2):
        return name1, name2

    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
    )
    name1 = "function_demo1"
    name2 = "function_demo2"
    obj1 = yr.fcc.create_function_group(demo_func_wrapper, args=(
        name1, name2,), group_size=4, options=opts)
    rets = yr.get(obj1)
    assert rets == [name1 if i % 2 == 0 else name2 for i in range(8)]


@pytest.mark.smoke
def test_fcc_create_function_group_bundle_size_equal_group_size(init_yr):
    @yr.invoke
    def demo_func_wrapper(name):
        return name

    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=1,
    )
    name = "function_demo"
    objs = yr.fcc.create_function_group(
        demo_func_wrapper, args=(name,), group_size=1, options=opts)
    rets = yr.get(objs)
    assert rets == name


@pytest.mark.smoke
def test_create_function_group_by_class(init_yr):
    @yr.instance
    class Demo:
        def __init__(self, name):
            self.name = name

        def add_name(self):
            return self.name

    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
    )
    name = "class_demo"
    function_group_handler = yr.fcc.create_function_group(
        Demo, args=(name,), group_size=4, options=opts)
    objs = function_group_handler.add_name.invoke()
    results = yr.get(objs)
    assert results == [name for i in range(1, 5)]
    function_group_handler.terminate()


@pytest.mark.smoke
def test_fcc_multi_return_value_class_function(init_yr):
    @yr.instance
    class Demo:
        def __init__(self, name1, name2):
            self.name1 = name1
            self.name2 = name2

        @yr.method(return_nums=2)
        def add_name(self):
            return self.name1, self.name2

    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
    )
    name1 = "class_demo1"
    name2 = "class_demo2"
    function_group_handler = yr.fcc.create_function_group(
        Demo, args=(name1, name2), group_size=4, options=opts)
    objs = function_group_handler.add_name.invoke()
    results = yr.get(objs)
    print(results)
    assert results == [name1 if i % 2 == 0 else name2 for i in range(8)]
    function_group_handler.terminate()


@pytest.mark.smoke
def test_create_function_group_by_function_with_big_bytes(init_yr):
    @yr.invoke
    def get(x):
        return x

    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
    )
    arg = b"1" * 1024 * 200
    objs = yr.fcc.create_function_group(
        get, args=(arg,), group_size=4, options=opts)
    rets = yr.get(objs)
    for ret in rets:
        assert ret == arg


@pytest.mark.smoke
def test_create_function_group_by_class_with_big_bytes(init_yr):
    @yr.instance
    class Demo:
        def __init__(self, name):
            self.name = name

        def get(self, x):
            return x

    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
    )
    name = "class_demo"
    function_group_handler = yr.fcc.create_function_group(
        Demo, args=(name,), group_size=4, options=opts)
    arg = b"1" * 1024 * 200
    objs = function_group_handler.get.invoke(arg)
    results = yr.get(objs)
    for ret in results:
        assert ret == arg
    function_group_handler.terminate()


@pytest.mark.smoke
def test_accelerate(init_yr):
    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
    )
    name = "class_demo"
    function_group_handler = yr.fcc.create_function_group(
        FccDemo, args=(name,), group_size=4, options=opts)
    function_group_handler.accelerate()
    arg = FccData(1, "2")
    objs = function_group_handler.get.invoke(arg)
    results = yr.get(objs)
    for ret in results:
        assert ret.a == arg.a
        assert ret.b == arg.b
    function_group_handler.terminate()


@pytest.mark.smoke
def test_accelerate_return_exception(init_yr):
    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
    )
    name = "class_demo"
    function_group_handler = yr.fcc.create_function_group(
        FccDemo, args=(name,), group_size=4, options=opts)
    function_group_handler.accelerate()
    objs = function_group_handler.exception.invoke()
    with pytest.raises(RuntimeError):
        results = yr.get(objs)
    function_group_handler.terminate()


@pytest.mark.smoke
def test_accelerate_with_async(init_yr):
    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
    )
    name = "class_demo"
    function_group_handler = yr.fcc.create_function_group(
        FccDemo, args=(name,), group_size=4, options=opts)
    function_group_handler.accelerate()
    arg = FccData(1, "2")
    objs = function_group_handler.get_async.invoke(arg)
    results = yr.get(objs)
    for ret in results:
        assert ret.a == arg.a
        assert ret.b == arg.b
    function_group_handler.terminate()


@pytest.mark.smoke
def test_fcc_get_function_group_context(init_yr):
    context = yr.fcc.get_function_group_context()
    assert context.world_size == 0


@pytest.mark.smoke
def test_fcc_create_function_group_exception(init_yr):
    @yr.invoke
    def demo_func_wrapper(name):
        return name

    opts = yr.FunctionGroupOptions(
        cpu=1000,
        memory=1000,
        scheduling_affinity_each_bundle_size=2,
        timeout=-2
    )
    name = "function_demo"

    try:
        yr.fcc.create_function_group(
            demo_func_wrapper, args=(name,), group_size=4, options=opts)
    except ValueError as e:
        assert "invalid timeout" in str(e)

    opts.timeout = -0.1
    try:
        yr.fcc.create_function_group(
            demo_func_wrapper, args=(name,), group_size=4, options=opts)
    except ValueError as e:
        assert "invalid timeout" in str(e)

    opts.scheduling_affinity_each_bundle_size = -1
    try:
        yr.fcc.create_function_group(
            demo_func_wrapper, args=(name,), group_size=4, options=opts)
    except ValueError as e:
        assert "invalid bundle size" in str(e)

    opts.scheduling_affinity_each_bundle_size = 10
    try:
        yr.fcc.create_function_group(
            demo_func_wrapper, args=(name,), group_size=4, options=opts)
    except ValueError as e:
        assert "invalid bundle size" in str(e)

    opts.scheduling_affinity_each_bundle_size = None
    try:
        yr.fcc.create_function_group(
            demo_func_wrapper, args=(name,), group_size=4, options=opts)
    except ValueError as e:
        assert "invalid bundle size" in str(e)
