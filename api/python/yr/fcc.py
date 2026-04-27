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
from typing import Union, List
from yr.config import FunctionGroupOptions, FunctionGroupContext
from yr.decorator import function_proxy, instance_proxy
from yr.runtime_holder import global_runtime
import yr


def create_function_group(
    wrapper: Union[function_proxy.FunctionProxy, instance_proxy.InstanceCreator],
    args: tuple,
    group_size: int,
    options: FunctionGroupOptions,
) -> Union[List["yr.ObjectRef"], instance_proxy.FunctionGroupHandler]:
    """
    Create function group instance.

    This function is used to create function group instances, and supports the invocation of
    function proxies or class instance creators.

    Note:
        1. A single group can create up to 256 instances.
        2. The number of instances of a single group must be divisible by the number of instances of a single bundle.
        3. During the process deployment, the heterogeneous resource collection function needs to be enabled (enabled
           by default). Different modes need to be selected to collect parameters of different ranges.
           The parameter `npu_collection_mode` can be set.

    Args:
        wrapper (Union[function_proxy.FunctionProxy, instance_proxy.InstanceCreator]):
            The decorated function proxy or the creator of the decorated class.
        args (tuple): Function parameters or positional parameters of a class constructor.
        group_size (int): Number of instances in the function group.
        options (FunctionGroupOptions): Options for creating function groups.

    Returns:
        Returns a list of references to data objects or a function group handle.
            Data type is Union[List[ObjectRef], FunctionGroupHandler].

    Raises:
        ValueError: If the `FunctionGroupOptions` or `group_size` parameter is incorrectly filled,
            this exception will be thrown.
        RuntimeError: If function is not wrapped by `@yr.invoke` or `@yr.instance`, this exception will be thrown.

    Examples:
        stateless function invoke example:
            >>> import yr
            >>>
            >>> yr.init()
            >>>
            >>> @yr.invoke
            ... def demo_func(name):
            ...     return name
            >>>
            >>> opts = yr.FunctionGroupOptions(
            ...     cpu=1000,
            ...     memory=1000,
            ...     resources={
            ...         "NPU/Ascend910B4/count": 1,
            ...     },
            ...     scheduling_affinity_each_bundle_size=2,
            ... )
            >>> name = "function_demo"
            >>> objs = yr.fcc.create_function_group(demo_func, args=(name,), group_size=8, options=opts)
            >>> rets = yr.get(objs)
            >>> assert rets == [name for i in range(1, 9)]
            >>> yr.finalize()

        class invoke example:
            >>> import yr
            >>>
            >>> @yr.instance
            ... class Demo(object):
            ...     name = ""
            >>>
            >>>     def __init__(self, name):
            ...         self.name = name
            >>>
            >>>     def return_name(self):
            ...         return self.name
            >>>
            >>> opts = yr.FunctionGroupOptions(
            ...         cpu=1000,
            ...         memory=1000,
            ...         resources={
            ...             "NPU/Ascend910B4/count": 1,
            ...         },
            ...         scheduling_affinity_each_bundle_size=2,
            ... )
            >>> name = "class_demo"
            >>> function_group_handler = yr.fcc.create_function_group(Demo, args=(name, ), group_size=8, options=opts)
            >>> objs = function_group_handler.return_name.invoke()
            >>> results = yr.get(objs)
            >>> assert results == [name for i in range(1, 9)]
            >>> function_group_handler.terminate()
            >>>
            >>> yr.finalize()
    """
    bundle_size = options.scheduling_affinity_each_bundle_size
    if bundle_size is None or bundle_size <= 0:
        raise ValueError(
            f"invalid bundle size: {bundle_size}, bundle size should bigger than 0")
    if bundle_size > group_size:
        raise ValueError(
            f"invalid bundle size or group size, group size {group_size} should be "
            f"bigger or equal to bundle size {bundle_size}")
    if group_size % bundle_size != 0:
        raise ValueError(f"group size: {group_size} is not divisible by bundle size: {bundle_size}")
    if options.timeout is None:
        options.timeout = -1
    if options.timeout != -1 and options.timeout < 0 or options.timeout > 0x7FFFFFFF:
        raise ValueError(f"invalid timeout {options.timeout} in FunctionGroupOptions, expect:-1, [0, 0x7FFFFFFF]")
    opts = yr.InvokeOptions()
    opts.function_group_options = options
    opts.cpu = opts.cpu if options.cpu is None else options.cpu
    opts.memory = opts.memory if options.memory is None else options.memory
    opts.concurrency = opts.concurrency if options.concurrency is None else options.concurrency
    opts.recover_retry_times = options.recover_retry_times
    for key, value in options.resources.items():
        opts.custom_resources[key] = value
    if isinstance(wrapper, instance_proxy.InstanceCreator):
        proxy = wrapper.set_function_group_size(group_size).options(opts).invoke(*args)
        return proxy.get_function_group_handler()
    if isinstance(wrapper, function_proxy.FunctionProxy):
        return wrapper.options(opts).set_function_group_size(group_size).invoke(*args)
    raise RuntimeError("function is not wrapped by @yr.invoke or @yr.instance")


def get_function_group_context() -> FunctionGroupContext:
    """
    Get the function group context of the current function instance.

    Note:
        - A valid function group context can only be obtained when using heterogeneous resources.
        - During the process deployment, the heterogeneous resource collection feature must be
          enabled (enabled by default). You need to choose different modes to collect different 
          ranges of parameters, and you can set the parameter `npu_collection_mode`.

    Returns:
        The function group context, i.e., FunctionGroupContext.

    Examples:
        >>> import yr
        >>> yr.init()
        >>>
        >>> @yr.invoke
        ... def demo_func(name):
        ...     context = yr.fcc.get_function_group_context()
        ...     print(context)
        ...     return name
        >>>
        >>> opts = yr.FunctionGroupOptions(
        ...     cpu=1000,
        ...     memory=1000,
        ...     resources={
        ...         "NPU/Ascend910B4/count": 1,
        ...     },
        ...     scheduling_affinity_each_bundle_size=2,
        ... )
        >>> name = "function_demo"
        >>> objs = yr.fcc.create_function_group(demo_func, args=(name,), group_size=8, options=opts)
        >>> rets = yr.get(objs)
        >>> assert rets == [name for i in range(1, 9)]
        >>> yr.finalize()
    """
    return global_runtime.get_runtime().get_function_group_context()
