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

"""api"""

import atexit
import functools
import logging
import os
from typing import List, Dict, Optional, Tuple, Union

from yr.libruntime_pb2 import LanguageType

from yr import log, runtime_holder
from yr.code_manager import CodeManager
from yr.common import constants, utils
from yr.config import ClientInfo, Config, InvokeOptions
from yr.config_manager import ConfigManager
from yr.decorator import function_proxy, instance_proxy
from yr.executor.executor import Executor
from yr.fnruntime import Consumer, Producer, auto_get_cluster_access_info
from yr.npu_object import NpuObject
from yr.object_ref import ObjectRef
from yr.resource_group_ref import RgObjectRef
from yr.base_runtime import ExistenceOpt, WriteMode, CacheType, SetParam, MSetParam, CreateParam, GetParams
from yr.stream import ProducerConfig, SubscriptionConfig
from yr.decorator.function_proxy import FunctionProxy
from yr.decorator.instance_proxy import InstanceCreator, InstanceProxy
from yr.common.utils import CrossLanguageInfo
from yr.resource_group import ResourceGroup

from yr.serialization import Serialization
from yr.ds_tensor_client_manager import get_tensor_client

__g_is_init = False
_MAX_INT = 0x7FFFFFFF
_logger = logging.getLogger(__name__)
_DEFAULT_ALLOW_PARTIAL = False
_DEFAULT_ALLOW_FORCE = True
_DEFAULT_ALLOW_RECURSIVE = False
_DEFAULT_SAVE_LOAD_STATE_TIMEOUT = 30
_DEFAULT_FUNC_ID = "sn:cn:yrk:12345678901234561234567890123456:function:0-defaultservice-py:$latest"


def is_initialized():
    """
    Check initialize state.
    """
    return __g_is_init


def check_initialized(func):
    """
    Check initialize state.
    """

    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        if not is_initialized():
            raise RuntimeError("system not initialized")
        return func(*args, **kwargs)

    return wrapper


def set_initialized():
    """
    Set initialize state.
    """
    global __g_is_init
    __g_is_init = True


def _get_from_env(conf):
    if conf.function_id == "":
        conf.function_id = os.environ.get("YRFUNCID", _DEFAULT_FUNC_ID)
    if conf.cpp_function_id == "":
        conf.cpp_function_id = os.environ.get("YR_CPP_FUNCID", "")
    if conf.server_address == "":
        conf.server_address = os.environ.get("YR_SERVER_ADDRESS", "")
    if conf.ds_address == "":
        conf.ds_address = os.environ.get("YR_DS_ADDRESS", "")
    if conf.working_dir == "":
        conf.working_dir = os.environ.get("YR_WORKING_DIR", "")
    app_mode_env_var = os.environ.get("YR_APP_MODE", "false").lower()
    if app_mode_env_var == "true":
        conf.is_driver = False
    if conf.rt_server_address == "":
        conf.rt_server_address = os.environ.get("POSIX_LISTEN_ADDR", "")
    return conf


def _auto_get_cluster_access_info(conf):
    conf = _get_from_env(conf)
    if not conf.is_driver:
        return conf

    args = []
    if conf.num_cpus:
        args.append("--cpu_num")
        args.append(str(conf.num_cpus * 1000))

    cluster_access_info = auto_get_cluster_access_info({
        "serverAddr": conf.server_address,
        "dsAddr": conf.ds_address,
        "inCluster": conf.in_cluster
    }, args)

    conf.server_address = cluster_access_info["serverAddr"]
    conf.ds_address = cluster_access_info["dsAddr"]
    conf.in_cluster = cluster_access_info["inCluster"]
    return conf


def init(conf: Config = None) -> ClientInfo:
    """
    The openYuanRong client initialization interface is used to configure the running mode and other parameters.

    Args:
        conf (Config, optional): The initialization parameter configuration for Yuanrong.
            For detailed descriptions of the parameters, see the Config data structure.
            It is optional and will be imported from environment variables if left empty.

    Returns:
        The context information of this invocation. Data type is ClientInfo.

    Raises:
        RuntimeError: If yr.init is called more than once.
        TypeError: If the parameter type is incorrect.
        ValueError: If the parameter value is incorrect.

    Example:

        >>> import yr
        >>>
        >>> conf = yr.Config()
        >>> yr.init(conf)
    """
    if is_initialized() and ConfigManager().is_driver:
        raise RuntimeError("yr.init cannot be called twice")

    conf = Config() if conf is None else conf

    conf = _auto_get_cluster_access_info(conf)
    ConfigManager().init(conf, is_initialized())
    runtime_holder.init()

    if not is_initialized():
        log.init_logger(ConfigManager().is_driver,
                        ConfigManager().runtime_id, ConfigManager().log_level)
        if not ConfigManager().is_driver:
            Executor.load_handler()
        atexit.register(finalize)
        set_initialized()

    msg = f"Succeeded to init YR, jobID is {ConfigManager().job_id}"
    if not conf.is_driver:
        log.get_logger().info(msg)
    else:
        _logger.info(msg)

    return ClientInfo(ConfigManager().job_id)


def finalize() -> None:
    """
    Terminate the system and release the relevant resources.

    Note:
        1. This interface is used to terminate the Unisound system, releasing created function instances,
           unreleased data objects and other resources.
        2. ``finalize`` must be called after ``init``.


    Raises:
        RuntimeError: This exception will be thrown if ``finalize`` is called without initializing ``yr``.

    Returns:
            None.

    Examples:
        >>> import yr
        >>> conf = yr.Config()
        >>> yr.init(conf)
        >>> yr.finalize()
    """
    global __g_is_init
    if not __g_is_init:
        return
    runtime_holder.global_runtime.get_runtime().finalize()
    CodeManager().clear()
    __g_is_init = False
    _logger.info("Succeeded to finalize, jobID is %s", ConfigManager().job_id)


def receive_request_loop() -> None:
    """
    main loop
    :return: None
    """
    runtime_holder.global_runtime.get_runtime().receive_request_loop()


@check_initialized
def put(obj: object, create_param: CreateParam = CreateParam()) -> ObjectRef:
    """
    Put an object to datasystem.

    Note:
        1. this method should be used after `yr.init()`.
        2. If the type of put is memoryview, bytearray or bytes, serialization is omitted at this time.
        3. If the object passed to put() is of type memoryview, bytearray, or bytes, its length must not be ``0``.

    Args:
        obj (object): This is a python object, will be pickled and saved to datasystem.
        create_param (CreateParam): This is create param for datasystem.

    Returns:
        the object ref to the object. Data type is ObjectRef.

    Raises:
        ValueError: If the input `obj` is `None` or a zero-length `bytes`, `bytearray`, or `memoryview` object.
        TypeError: If the input obj is already an `ObjectRef`.
        TypeError: If the input obj is not serializable, e.g. `thread.RLock`.
        RuntimeError: Call `yr.put()` before `yr.init()`.
        RuntimeError: Failed to put to datasystem.

    Examples:
        >>> import yr
        >>> yr.init()
        >>> param = yr.CreateParam()
        >>> param.cache_type = yr.CacheType.DISK
        >>> bs = bytes(0)
        >>> mem = memoryview(bytes(100))
        >>> obj_ref2 = yr.put(mem)
        >>> print(yr.get(obj_ref2))
        >>> # The final print output is a memoryview pointer.
        >>> byte_array = bytearray(20)
        >>> obj_ref3 = yr.put(byte_array)
        >>> print(yr.get(obj_ref3))
        >>> # The final print output is a memoryview pointer.
        >>> obj_ref4 = yr.put(100)
        >>> print(yr.get(obj_ref4))
    """
    if obj is None or (isinstance(obj, (bytes, bytearray, memoryview)) and len(obj) == 0):
        raise ValueError("value is None or has zero length")
    # Make sure that the value is not an object ref.
    if isinstance(obj, ObjectRef):
        raise TypeError(
            "Calling 'put' on an ObjectRef is not allowed. If you really want to do this, "
            "you can wrap the ObjectRef in a list and call 'put' on it.")
    return ObjectRef(runtime_holder.global_runtime.get_runtime().put(obj, create_param), need_incre=False)


def _recurse(obj, ref_obj):
    if isinstance(obj, NpuObject):
        try:
            import torch
            from torch.npu import current_device
            data_system_import = True
        except ImportError as e:
            data_system_import = False
            import_error = e
        if not data_system_import:
            raise RuntimeError(f"import err: {import_error}")

        out_tensor = torch.zeros(size=obj.size, dtype=obj.dtype, device=current_device())
        ds_tensor_client = get_tensor_client()
        failed_keys = ds_tensor_client.dev_mget([obj.id], [out_tensor])
        if failed_keys:
            raise RuntimeError(f"dev_mget failed, failed keys: {failed_keys}")
        ref_obj.npu_obj_ids.append(obj.id)
        ref_obj.instance_id = obj.instance_id
        return out_tensor
    if isinstance(obj, list):
        return [_recurse(item, ref_obj) for item in obj]
    if isinstance(obj, dict):
        return {k: _recurse(v, ref_obj) for k, v in obj.items()}
    if isinstance(obj, set):
        new_set = set()
        for item in obj:
            new_set.add(_recurse(item, ref_obj))
        return new_set
    if isinstance(obj, tuple):
        return tuple(_recurse(item, ref_obj) for item in obj)
    return obj


def _restore(obj, object_refs):
    for i, _ in enumerate(obj):
        obj[i] = _recurse(obj[i], object_refs[i])


@check_initialized
def get(obj_refs: Union["ObjectRef", List, "RgObjectRef"], timeout: int = constants.DEFAULT_GET_TIMEOUT,
        allow_partial: bool = _DEFAULT_ALLOW_PARTIAL) -> object:
    """
    Retrieve the value of an object stored in the backend based on the object's key.
    The interface call will block until the object's value is obtained or a timeout occurs.

    Note:
        yr.get() uniformly returns a memoryview pointer for bytes, bytearray, and memoryview types.

    Args:
        obj_refs (ObjectRef, List[ObjectRef]): The object_ref of the object in the data system.
        timeout (int, optional): The timeout value. A value of -1 means wait indefinitely. Limit: -1, (0, ∞).
            Defaults to ``300``.
        allow_partial (bool, optional): If set to False, the get interface will throw an exception
            when the data system returns partial results within the timeout period.
            If set to True, the get interface will return a list of objects and fill failed objects with None.
            Default is ``_DEFAULT_ALLOW_PARTIAL``.

    Returns:
        A Python object or a list of Python objects.

    Raises:
        ValueError: If the input parameter type is incorrect.
        RuntimeError: If obtaining the object from the data system fails.
        YRInvokeError: If the function execution fails.
        TimeoutError: If the results of all object references cannot be obtained within the specified timeout period.

    Examples:
        >>> import yr
        >>> yr.init()
        >>>
        >>> @yr.invoke()
        >>> def add(a, b):
        ...     return a + b
        >>> obj_ref_1 = add.invoke(1, 2)
        >>> obj_ref_2 = add.invoke(3, 4)
        >>> result = yr.get([obj_ref_1, obj_ref_2], timeout=-1)
        >>> print(result)
        >>> yr.finalize()

    """
    if timeout <= constants.MIN_TIMEOUT_LIMIT and timeout != constants.NO_LIMIT:
        raise ValueError(
            "Parameter 'timeout' should be greater than 0 or equal to -1 (no timeout)")
    if isinstance(obj_refs, RgObjectRef):
        try:
            obj_refs.resource_group.wait(timeout)
        except Exception as e:
            raise e
        return Serialization().deserialize(obj_refs.data)
    if isinstance(obj_refs, (list, tuple)):
        if len(obj_refs) == 0:
            return []
        if all(isinstance(r, RgObjectRef) for r in obj_refs):
            results = []
            for ref in obj_refs:
                try:
                    ref.resource_group.wait(timeout)
                except Exception as e:
                    raise e
                results.append(Serialization().deserialize(ref.data))
            return results
    is_single_obj = isinstance(obj_refs, ObjectRef)
    if is_single_obj:
        obj_refs = [obj_refs]
    elif isinstance(obj_refs, list) and len(obj_refs) == 0:
        return []
    _check_object_ref(obj_refs)
    objects = runtime_holder.global_runtime.get_runtime().get(
        [ref.id for ref in obj_refs], timeout, allow_partial)
    _restore(objects, obj_refs)
    if is_single_obj:
        return objects[0]
    return objects


@check_initialized
def wait(obj_refs: Union[ObjectRef, List[ObjectRef]], wait_num: Optional[int] = None,
         timeout: Optional[int] = None) -> Tuple[List[ObjectRef], List[ObjectRef]]:
    """
    Wait for the value of the object in the data system to be ready based on the object's key.
    The interface call will block until the value of the object is computed.

    Note:
        The results returned each time may differ because the order of completion of invoke
        is not guaranteed.

    Args:
        object_refs (list): Data saved to the data system.
        wait_num (int, optional): The minimum number of objects to wait for. It defaults to ``1``.
            The value should not exceed the length of `obj_refs`.
        timeout (int, optional): The timeout in seconds. Note that if the default value ``None`` is used,
            it will wait indefinitely, with the actual maximum wait time limited by the wait factors in `get`.

    Returns:
        Returns two lists: the first list contains the completed requests;
        the second list contains the uncompleted requests.
        Data type is tuple[list, list].

    Raises:
        TypeError: If the input parameter type is incorrect.
        ValueError: If the input parameter is incorrect.

    Examples:
        >>> import yr
        >>> import time
        >>>
        >>> yr.init()
        >>>
        >>> @yr.invoke
        ... def demo(a):
        ...     time.sleep(a)
        ...     return "sleep:", a
        ...
        >>> res = [demo.invoke(i) for i in range(4)]
        >>>
        >>> wait_num = 3
        >>> timeout = 10
        >>> result = yr.wait(res, wait_num, timeout)
        >>> print("ready_list  = ", result[0], "unready_list = ", result[1])
        >>> print(yr.get(result[0]))
        [('sleep:', 0), ('sleep:', 1)]
        >>>
        >>> yr.finalize()
    """
    if timeout is None:
        timeout = -1
    if isinstance(obj_refs, ObjectRef):
        obj_refs = [obj_refs]
    _check_object_ref(obj_refs)
    if len(obj_refs) != len(set(obj_refs)):
        raise ValueError(
            "obj_refs value error: duplicate obj_ref exists in the list")

    if wait_num is None:
        wait_num = len(obj_refs)
    if wait_num == 0 or timeout == 0:
        return [], obj_refs

    if not isinstance(wait_num, int):
        raise TypeError(
            f"'invalid wait_num type, actual: {type(wait_num)}, expect: <class 'int'>")
    if wait_num < 0 or wait_num > len(obj_refs):
        raise ValueError(
            f"invalid wait_num value, actual: {wait_num}, expect: [0, {len(obj_refs)}]")

    if timeout is not None:
        if not isinstance(timeout, int):
            raise TypeError(
                f"invalid timeout type, actual: {type(timeout)}, expect: <class 'int'>")
        if timeout != -1 and timeout < 0 or timeout > _MAX_INT:
            raise ValueError(f"invalid timeout value, actual: {timeout}, expect:None, -1, [0, {_MAX_INT}]")
    # deal with YRError
    ready_ids, _ = runtime_holder.global_runtime.get_runtime().wait(
        [obj_ref.id for obj_ref in obj_refs], wait_num, timeout)
    ready_ids_set = set(ready_ids)
    ready_ref = []
    unready_ref = []
    for i in obj_refs:
        if i.id in ready_ids_set:
            ready_ref.append(i)
        else:
            unready_ref.append(i)
    return ready_ref, unready_ref


@check_initialized
def cancel(obj_refs: Union[ObjectRef, List[ObjectRef]], allow_force: bool = _DEFAULT_ALLOW_FORCE,
           allow_recursive: bool = _DEFAULT_ALLOW_RECURSIVE) -> None:
    """
    Cancel a specified stateless request.

    Note:
        Currently, passing an instance request will not trigger the corresponding cancel
        operation and will not result in an error.

    Args:
        obj_refs (Union[ObjectRef, List[ObjectRef]]): The ObjectRef(s) of the stateless request to be stopped,
            either a single stateless request or multiple stateless requests.
        allow_force (bool, optional): Force stopping. Default is ``_DEFAULT_ALLOW_FORCE``.
        allow_recursive (bool, optional): Allow recursion. Default is ``_DEFAULT_ALLOW_RECURSIVE``.

    Returns:
        None.

    Raises:
        TypeError: If the parameter type is incorrect.

    Examples:
        >>> import time
        >>> import yr
        >>> yr.init()
        >>>
        >>> @yr.invoke
        >>> def func():
        >>>     time.sleep(100)
        >>>
        >>> ret = func.invoke()
        >>> yr.cancel(ret)
        >>> yr.finalize()
    """
    if isinstance(obj_refs, ObjectRef):
        obj_refs = [obj_refs]
    _check_object_ref(obj_refs)
    runtime_holder.global_runtime.get_runtime().cancel(
        [ref.id for ref in obj_refs], allow_force, allow_recursive)


def invoke(*args, **kwargs) -> function_proxy.FunctionProxy:
    """
    Used to decorate functions that need to be remotely invoked on the openYuanRong system.

    This decorator is used to mark functions that need to be remotely invoked
    on the Yuanrong system and returns a proxy object of the function.

    Note:
        Due to the performance limitations of the HTTP client,
        only 100,000 concurrent connections are currently supported per client.

    Args:
        func (FunctionType): The function that needs to be remotely invoked.
        invoke_options (InvokeOptions): Invocation options, see `InvokeOptions`.
        return_nums (int): The number of return values of the function,
            restrictions: greater than 0, this parameter is not recommended to be set too large.

    Returns:
        Returns the proxy object of the decorated function. Data type is FunctionProxy.

    Raises:
        RuntimeError: If `invoke` cannot decorate objects other than `FunctionType`, this exception will be thrown.

    Examples:
        Simple invocation example:
            >>> import yr
            >>> yr.init()
            >>> @yr.invoke
            ... def add(a, b):
            ...     return a + b
            >>> ret = add.invoke(1, 2)
            >>> print(yr.get(ret))
            >>> yr.finalize()


        Function invocation example:
            >>> import yr
            >>> yr.init()
            >>> @yr.invoke
            ... def func1(a):
            ...     return a + " func1"
            >>> @yr.invoke
            ... def func2(a):
            ...     return yr.get(func1.invoke(a)) + " func2"
            >>> ret = func2.invoke("hello")
            >>> print(yr.get(ret))
            >>> yr.finalize()
    """
    if len(args) == 1 and len(kwargs) == 0 and callable(args[0]):
        return function_proxy.make_decorator()(args[0])
    invoke_options = kwargs.get("invoke_options", None)
    return_nums = kwargs.get("return_nums", None)
    initializer = kwargs.get("initializer", None)
    return function_proxy.make_decorator(invoke_options, return_nums, initializer)


def instance(*args, **kwargs) -> instance_proxy.InstanceCreator:
    """
    Used to decorate classes that need to be remotely invoked on the openYuanRong system.

    Args:
        class (class): The class that needs to be remotely invoked.
        invoke_options (yr.InvokeOptions): Invocation parameters.

    Returns:
        The creator of the decorated class.
        Data type is InstanceCreator.

    Raises:
        RuntimeError: If the object decorated by `instance` is not a class.

    Example:
        Simple invocation example:
            >>> import yr
            >>> yr.init()
            >>>
            >>> @yr.instance
            ... class Instance:
            ...     sum = 0
            ...     def add(self, a):
            ...         self.sum += a
            ...     def get(self):
            ...         return self.sum
            >>>
            >>> ins = Instance.invoke()
            >>> yr.get(ins.add.invoke(1))
            >>> print(yr.get(ins.get.invoke()))
            1
            >>> ins.terminate()
            >>>
            >>> yr.finalize()

        Function invocation example:
            >>> import yr
            >>> yr.init()
            >>>
            >>> @yr.instance
            ... class Instance:
            ...     def __init__(self):
            ...         self.sum = 0
            ...     def add(self, a):
            ...         self.sum += a
            ...     def get(self):
            ...         return self.sum

            >>> @yr.instance
            ... class Instance2:
            ...    def __init__(self):
            ...        self.ins = Instance.invoke()
            ...    def add(self, a):
            ...        return self.ins.add.invoke(a)
            ...    def get(self):
            ...        return yr.get(self.ins.get.invoke())
            >>>
            >>> ins = Instance2.invoke()
            >>> yr.get(ins.add.invoke(2))
            >>> print(yr.get(ins.get.invoke()))
            2
            >>> ins.terminate()
            >>>
            >>> yr.finalize()

    """
    if len(args) == 1 and len(kwargs) == 0 and callable(args[0]):
        return instance_proxy.make_decorator()(args[0])
    invoke_options = kwargs.get("invoke_options", None)
    return instance_proxy.make_decorator(invoke_options)


def method(*args, **kwargs):
    """
    Used to decorate class member methods.

    Args:
        return_nums (int): The number of return values of the member function,
            restrictions: greater than 0, this parameter is not recommended to be set too large.

    Returns:
        The decorated function. Data type is FunctionType.

    Raises:
        ValueError: If the input parameters are incorrect.
        TypeError: If the type of the input parameters is incorrect.

    Example:
        >>> import yr
        >>> yr.init()
        >>>
        >>> @yr.instance
        ... class Instance:
        ...     sum = 0
        ...
        ...     def add(self, a):
        ...         self.sum += a
        ...
        ...     def get(self):
        ...         return self.sum
        ...
        ...     @yr.method(return_nums=2)
        ...     def detail(self, a, b):
        ...         return a, b
        ...
        >>> ins = Instance.invoke()
        >>> res1, res2 = ins.detail.invoke(0, 1)
        >>> print("detail result1:", yr.get(res1))
        detail result1: 0
        >>> print("detail result2:", yr.get(res2))
        detail result2: 1
        >>> ins.terminate()
        >>>
        >>> yr.finalize()
    """
    if "concurrency_group" in kwargs and len(args) == 0:
        kwargs.pop("concurrency_group")

        def identity_decorator(class_method):
            return class_method

        return identity_decorator

    def annotate_method(class_method):
        return_nums = kwargs.get("return_nums", 1)
        if not isinstance(return_nums, int):
            raise TypeError(
                f"invalid return_nums type: {type(return_nums)}, should be an int")
        class_method.__return_nums__ = return_nums

        tensor_transport_target = kwargs.get("tensor_transport_target", None)
        if isinstance(tensor_transport_target, str) and tensor_transport_target != "npu":
            raise TypeError(
                f"invalid tensor_transport_target value: {tensor_transport_target}")
        class_method.__tensor_transport_target__ = "" if tensor_transport_target is None else tensor_transport_target

        enable_tensor_transport = kwargs.get("enable_tensor_transport", None)
        if enable_tensor_transport is not None and not isinstance(enable_tensor_transport, bool):
            raise TypeError(
                f"invalid tensor_transport_target type: {type(enable_tensor_transport)}, should be an bool")
        class_method.__enable_tensor_transport__ = enable_tensor_transport
        return class_method

    return annotate_method


def _check_object_ref(obj_refs: List[ObjectRef]):
    if not isinstance(obj_refs, list):
        raise TypeError(
            f"obj_refs type error, actual: [{type(obj_refs)}], element expect: <class 'ObjectRef'>")
    for ref in obj_refs:
        if not isinstance(ref, ObjectRef):
            raise TypeError(
                f"obj_refs type error, actual: [{type(ref)}], element expect: <class 'ObjectRef'>")
        ref.exception()


@check_initialized
def exit() -> None:
    """
    The function for exiting the program.

    Returns:
        None.
    """
    runtime_holder.global_runtime.get_runtime().exit()


@check_initialized
def create_stream_producer(stream_name: str, config: ProducerConfig) -> Producer:
    """
    Create a producer.

    Args:
        stream_name (str): The name of the stream.
            The length must be less than 256 characters and contain only the following characters
            `(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^&*()\\+\\=;:)` .
        config (ProducerConfig): The configuration of the producer.

    Returns:
        Producer.

    Raises:
        RuntimeError: If creating the Producer fails.

    Examples:
        >>> try:
        ...     producer_config = yr.ProducerConfig(
        ...         delay_flush_time=5,
        ...         page_size=1024 * 1024,
        ...         max_stream_size=1024 * 1024 * 1024,
        ...         auto_clean_up=True,
        ...     )
        ...     stream_producer = yr.create_stream_producer("streamName", producer_config)
        ... except RuntimeError as exp:
        ...     # 处理异常
        ...     pass
    """
    return runtime_holder.global_runtime.get_runtime().create_stream_producer(stream_name, config)


@check_initialized
def create_stream_consumer(stream_name: str, config: SubscriptionConfig) -> Consumer:
    """
    Create a consumer.

    Args:
        stream_name (str): The name of the stream.
            The length must be less than 256 characters and contain only the following characters
            `(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^&*()\\+\\=;:)` .
        config (SubscriptionConfig): The configuration of the consumer.

    Returns:
        Consumer.

    Raises:
        RuntimeError:  If creating the Consumer fails.

    Examples:
        >>> try:
        ...     config = yr.SubscriptionConfig("subName", yr.SubscriptionConfig.subscriptionType.STREAM)
        ...     consumer = yr.create_stream_consumer("streamName", config)
        ... except RuntimeError as exp:
        ...     pass
    """
    return runtime_holder.global_runtime.get_runtime().create_stream_consumer(stream_name, config)


@check_initialized
def query_global_producers_num(stream_name: str) -> int:
    """
    查询流生产者数量

    Args:
        stream_name: 流名称

    Returns:
        数量
    """
    return runtime_holder.global_runtime.get_runtime().query_global_producers_num(stream_name)


@check_initialized
def query_global_consumers_num(stream_name: str) -> int:
    """
    查询流消费者数量

    Args:
        stream_name: 流名称

    Returns:
        数量
    """
    return runtime_holder.global_runtime.get_runtime().query_global_consumers_num(stream_name)


@check_initialized
def delete_stream(stream_name: str) -> None:
    """
    Delete the data stream. When the global count of producers and consumers is 0, the data stream is no longer in use,
    and the metadata related to the data stream on each worker and master is cleaned up.
    This function can be called on any Host node.

    Args:
        stream_name (str): The name of the stream.
            The length must be less than 256 characters and contain only the following characters
            `(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^&*()\\+\\=;:)` .

    Returns:
        None.

    Raises:
        RuntimeError: If the stream fails to be deleted in the data system, an exception will be thrown.

    Examples:
        >>> try:
        ...     yr.delete_stream("streamName")
        ... except RuntimeError as exp:
        ...     pass
    """
    runtime_holder.global_runtime.get_runtime().delete_stream(stream_name)


@check_initialized
def kv_write(key: str, value: bytes, existence: ExistenceOpt = ExistenceOpt.NONE,
             write_mode: WriteMode = WriteMode.NONE_L2_CACHE, ttl_second: int = constants.DEFAULT_NO_TTL_LIMIT,
             cache_type: CacheType = CacheType.MEMORY) -> None:
    """
    Provides the Redis class's set storage interface,
    which supports saving binary data to the data system.

    Args:
        key (str): Sets a key for the data to be saved, which is used to identify the data.
            This key is used for querying the data and cannot be empty.
        value (bytes): The binary data to be stored. The maximum storage limit is 100M outside the cloud.
        existence (ExistenceOpt, optional): Whether to support overwriting of the `Key`.
            This parameter is optional and defaults to ``ExistenceOpt.NONE``,
            which means overwriting is allowed; ``ExistenceOpt.NX`` means overwriting is not allowed.
        write_mode (WriteMode, optional): Sets the reliability of the data.
            When the server configuration supports secondary caching to ensure reliability (e.g., Redis service),
            this configuration can be used to ensure data reliability.
            This parameter is optional and defaults to ``WriteMode.NONE_L2_CACHE``.
        ttl_second (int, optional): The data's time-to-live (TTL) in seconds, after which the data will be deleted.
            This parameter is optional and defaults to ``0``,
            meaning the key will exist indefinitely until the kv_del interface is explicitly called.
        cache_type (CacheType, optional): Sets the medium for data storage.
            This parameter is optional and defaults to ``CacheType.MEMORY``,
            which means storing in memory; ``CacheType.DISK`` means storing on disk.

    Returns:
        None.

    Raises:
        RuntimeError: If ``kv_write`` is called without initialization, an exception will be thrown.
        RuntimeError: If the data fails to be written to the data system.

    Example:
        >>> import yr
        >>> yr.init()
        >>>
        >>> yr.kv_write("kv-key", b"value1", yr.ExistenceOpt.NONE, yr.WriteMode.NONE_L2_CACHE, 0, yr.CacheType.MEMORY)
        >>>
        >>> yr.finalize()
    """
    set_param = SetParam()
    set_param.existence = existence
    set_param.write_mode = write_mode
    set_param.ttl_second = ttl_second
    set_param.cache_type = cache_type
    runtime_holder.global_runtime.get_runtime().kv_write(key, value, set_param)


@check_initialized
def kv_write_with_param(key: str, value: bytes, set_param: SetParam) -> None:
    """
    Provide a storage interface that supports custom parameter configuration, 
    enabling binary data to be written to the data system in a more flexible manner.

    Args:
        key (str): Sets a key for the data to be saved, which is used to identify the data.
            This key is used for querying the data and cannot be empty.
        value (bytes): The binary data to be stored. The maximum storage limit is 100M outside the cloud.
        set_param (SetParam): Configuration parameters for kv write in the data system,
            `including existence`, `write_mode`, `ttl_second`, and `cache_type`.

    Returns:
        None.

    Raises:
        RuntimeError: If `kv_write_with_param` is called without initialization, an exception will be thrown.
        RuntimeError: If the data fails to be written to the data system.

    Example:
        >>> import yr
        >>> yr.init()
        >>> # The worker startup parameters need to be configured with shared_disk_directory and shared_disk_size_mb;
        >>> # otherwise, this example will result in an error
        >>> set_param = yr.SetParam()
        >>> set_param.existence = yr.ExistenceOpt.NX
        >>> set_param.write_mode = yr.WriteMode.NONE_L2_CACHE_EVICT
        >>> set_param.ttl_second = 10
        >>> yr.kv_write_with_param("kv-key", b"value1", set_param)
        >>>
        >>> yr.finalize()
    """
    runtime_holder.global_runtime.get_runtime().kv_write(key, value, set_param)


@check_initialized
def kv_m_write_tx(keys: List[str], values: List[bytes], m_set_param: MSetParam = MSetParam()) -> None:
    """
    Provide a Redis-like set storage interface that supports persisting 
    a collection of binary data into the data system.

    Args:
        keys (List[str]): Set a set of keys for the saved data to identify the data. Use this key for querying data.
            It cannot be empty.
        values (List[bytes]): A set of binary data that needs to be stored.
            The maximum storage limit outside the cloud is ``100`` M.
        m_set_param (MSetParam, optional): configuration parameters of mulrit kv write in the data system,
            including ``existence``, ``write_mode``, ``ttl_second` and ``cache_type``.

    Returns:
        None.

    Raises:
        RuntimeError: If `kv_m_write_tx` is not initialized and called, an exception will be thrown.
            If data writing to the data system fails.

    Example:
        >>> import yr
        >>> yr.init()
        >>> mset_param = yr.MSetParam()
        >>> mset_param.existence = yr.ExistenceOpt.NX
        >>> mset_param.write_mode = yr.WriteMode.NONE_L2_CACHE_EVICT
        >>> mset_param.ttl_second = 100
        >>> yr.kv_m_write_tx(["key1", "key2"], [b"value1", b"value2"], mset_param)
        >>>
        >>> yr.finalize()
    """
    if len(keys) != len(values):
        raise ValueError(
            f"arguments list size not equal. keys size is: {len(keys)}, values size is: {len(values)}.")
    if m_set_param.existence != ExistenceOpt.NX:
        raise ValueError("invalid param: existence should be NX.")
    runtime_holder.global_runtime.get_runtime().kv_m_write_tx(keys, values, m_set_param)


@check_initialized
def kv_read(
        key: Union[str, List[str]],
        timeout: int = constants.DEFAULT_GET_TIMEOUT) \
        -> Union[Union[bytes, memoryview], List[Union[bytes, memoryview]]]:
    """
    Provides a Redis-like interface to retrieve data,
    supporting the retrieval of multiple data items simultaneously.

    Args:
        key (Union[str, List[str]]): A single key or a list of keys specifying the data to be retrieved.
        timeout (int, optional): The timeout in seconds. This parameter is optional and defaults to ``300``.
            A value of ``-1`` indicates permanent blocking wait.

    Returns:
        Returns one or a list of retrieved data items. Data type is Union[bytes, List[bytes]].

    Raises:
        RuntimeError: If `kv_read()` is called without initialization, an exception will be thrown.
        RuntimeError: If data retrieval from the data system fails.

    Example:
        >>> v1 = yr.kv_read("kv-key")
    """
    is_single_obj = isinstance(key, str)
    rets = runtime_holder.global_runtime.get_runtime().kv_read(key, timeout)
    if is_single_obj:
        return rets[0]
    return rets


@check_initialized
def kv_set(key: str, value: bytes, set_param: SetParam = SetParam()) -> None:
    """
    Provide a set storage interface similar to Redis, supporting the saving of binary data to the data system.

    Args:
        key (str): Set a key for the saved data to identify it. Use this key to query data. It cannot be empty.
        value (bytes): Binary data to be stored. The maximum storage limit outside the cloud is ``100M``.
        set_param (SetParam): The configuration parameters of kv write in the data system include ``existence``,
            ``write_mode``, ``ttl_second`` and ``cache_type``.

    Returns:
        None.

    Raises:
        RuntimeError: If ``kv_set`` is not initialized and called, an exception will be thrown.
        RuntimeError: If the data writing to the data system fails.

    Example:
        >>> import yr
        >>> yr.init()
        >>> # The worker startup parameters need to be configured with shared_disk_directory and shared_disk_size_mb;
        >>> # otherwise, this example will result in an error
        >>> set_param = yr.SetParam()
        >>> set_param.existence = yr.ExistenceOpt.NX
        >>> set_param.write_mode = yr.WriteMode.NONE_L2_CACHE_EVICT
        >>> set_param.ttl_second = 10
        >>> yr.kv_set("kv-key", b"value1", set_param)
        >>>
        >>> yr.finalize()
    """
    runtime_holder.global_runtime.get_runtime().kv_write(key, value, set_param)


@check_initialized
def kv_get(
        key: Union[str, List[str]],
        timeout: int = constants.DEFAULT_GET_TIMEOUT) \
        -> Union[Union[bytes, memoryview], List[Union[bytes, memoryview]]]:
    """
    Provides a Redis-like interface to retrieve data,
    supporting the retrieval of multiple data items simultaneously.

    Args:
        key (Union[str, List[str]]): A single key or a list of keys specifying the data to be retrieved.
        timeout (int, optional): The timeout in seconds. Defaults to ``300``.
            A value of ``-1`` indicates permanent blocking wait.

    Returns:
        Returns one or a list of retrieved data items. Data type is Union[bytes, List[bytes]].

    Raises:
        RuntimeError: If `kv_get()` is called without initialization, an exception will be thrown.
        RuntimeError: If data retrieval from the data system fails.

    Example:
        >>> v1 = yr.kv_get("kv-key")
    """
    if timeout <= constants.MIN_TIMEOUT_LIMIT and timeout != constants.NO_LIMIT:
        raise ValueError(
            "Parameter 'timeout' should be greater than 0 or equal to -1 (no timeout)")
    is_single_obj = isinstance(key, str)
    rets = runtime_holder.global_runtime.get_runtime().kv_read(key, timeout)
    if is_single_obj:
        return rets[0]
    return rets


@check_initialized
def kv_get_with_param(keys: List[str], get_params: GetParams, timeout: int = constants.DEFAULT_GET_TIMEOUT) \
        -> List[Union[bytes, memoryview]]:
    """
    Provides the kv_get_with_param interface, which supports offset reading of data.

    Args:
        keys (List[str]): A list of keys specifying the data to be retrieved.
        get_params (GetParams): A set of attributes configured for retrieving the value,
            including parameters such as `offset` and `size`.
        timeout (int): The timeout in seconds. Defaults to ``300``.
            A value of ``-1`` indicates permanent blocking wait. The parameter must be greater than or equal to ``-1``.

    Returns:
        Returns one or a list of retrieved data items. Data type is Union[bytes, List[bytes]].

    Raises:
        ValueError: If the `get_params` list of `GetParam` type is an empty list.
        ValueError: If the number of items in the `get_params`
            list does not match the number of items in the `keys` list.
        RuntimeError: If `kv_get_with_param()` is called without initialization, an exception will be thrown.
        RuntimeError: If data retrieval from the data system fails.

    Example:
        >>> get_param = yr.GetParam()
        >>> get_param.offset = 0
        >>> get_param.size = 0
        >>> params = yr.GetParams()
        >>> params.get_params = [get_param]
        >>> v1 = yr.kv_get_with_param(["kv-key"], params, 10)
    """
    if timeout < constants.NO_LIMIT:
        raise ValueError(
            "Parameter 'timeout' should be greater than or equal to -1 (no timeout).")
    if len(get_params.get_params) == 0:
        raise ValueError("Get params does not accept empty list")
    if len(get_params.get_params) != len(keys):
        raise ValueError("Get params size is not equal to keys size")
    return runtime_holder.global_runtime.get_runtime().kv_get_with_param(keys, get_params, timeout)


@check_initialized
def kv_del(key: Union[str, List[str]]) -> None:
    """
    Provides a Redis-like interface to delete data,
    supporting the deletion of multiple data items simultaneously.

    Args:
        key (Union[str, List[str]]): A single key or a list of keys specifying the data to be deleted.

    Returns:
        None.

    Raises:
        RuntimeError: A single key or a list of keys specifying the data to be deleted.
        RuntimeError: If data deletion from the data system fails.

    Example:
        >>> yr.kv_write("kv-key", b"value1", yr.ExistenceOpt.NONE, yr.WriteMode.NONE_L2_CACHE, 0) # doctest: +SKIP
        >>> yr.kv_del("kv-key")
    """
    runtime_holder.global_runtime.get_runtime().kv_del(key)


@check_initialized
def save_state(timeout_sec: int = _DEFAULT_SAVE_LOAD_STATE_TIMEOUT) -> None:
    """
    Used to save the instance status.

    Args:
        timeout_sec (int, optional): The timeout in seconds. Defaults to ``30``.

    Returns:
        None.

    Raises:
        RuntimeError: If `save_state` is called in local mode,
            an exception is raised with the message: `save_state is not called in local mode`.
        RuntimeError: If `save_state` is called in cloud mode without using the POSIX API,
            an exception is raised with the message: `save_state is only called on cloud with POSIX API`.
        RuntimeError: If `save_state` fails to retrieve the saved instance state,
            an exception is raised with the message: `Failed to save state`.


    Example:
        >>> @yr.instance
        ... class Counter:
        ...     def __init__(self):
        ...         self.cnt = 0
        ...
        ...     def add(self):
        ...         self.cnt += 1
        ...         return self.cnt
        ...
        ...     def get(self):
        ...         return self.cnt
        ...
        ...     def save(self, timeout=30):
        ...         yr.save_state(timeout)
        ...
        ...     def load(self, timeout=30):
        ...         yr.load_state(timeout)
        ...
        >>> counter = Counter.invoke()
        >>> print(f"member value before save state: {yr.get(counter.get.invoke())}")
        >>> counter.save.invoke()
        >>>
        >>> counter.add.invoke()
        >>> print(f"member value after add one: {yr.get(counter.get.invoke())}")
        >>>
        >>> counter.load.invoke()
        >>> print(f"member value after load state(back to 0): {yr.get(counter.get.invoke())}")
    """
    remote_runtime = ConfigManager().in_cluster and not ConfigManager().is_driver
    if not remote_runtime:
        raise RuntimeError(
            "save_state is only supported on cloud with posix api")

    if ConfigManager().local_mode:
        raise RuntimeError("save_state is not supported in local mode")
    utils.Validator.check_timeout(timeout_sec)
    _logger.debug("Start to save instance state, timeout: %d seconds", timeout_sec)
    runtime_holder.global_runtime.get_runtime().save_state(timeout_sec)
    _logger.info("Succeeded to save instance state")


@check_initialized
def load_state(timeout_sec: int = _DEFAULT_SAVE_LOAD_STATE_TIMEOUT) -> None:
    """
    Used to load the saved instance state.

    Args:
        timeout_sec (int, optional): The timeout in seconds. Defaults to ``30``.

    Returns:
        None.

    Raises:
        RuntimeError: If `load_state` is not called in local mode,
            an exception is raised with the message: `load_state is not called in local mode`.
        RuntimeError: If `load_state` is called in cloud mode without using the POSIX API,
            an exception is raised with the message: `load_state is only supported on cloud with POSIX API`.
        RuntimeError: If `load_state` fails to retrieve the saved instance state,
            an exception is raised with the message: `Failed to load state`.


    Example:
        >>> @yr.instance
        ... class Counter:
        ...     def __init__(self):
        ...         self.cnt = 0
        ...
        ...     def add(self):
        ...         self.cnt += 1
        ...         return self.cnt
        ...
        ...     def get(self):
        ...         return self.cnt
        ...
        ...     def save(self, timeout=30):
        ...         yr.save_state(timeout)
        ...
        ...     def load(self, timeout=30):
        ...         yr.load_state(timeout)
        ...
        >>> counter = Counter.invoke()
        >>> print(f"member value before save state: {yr.get(counter.get.invoke())}")
        member value before save state: 0
        >>> counter.save.invoke()
        >>>
        >>> counter.add.invoke()
        >>> print(f"member value after add one: {yr.get(counter.get.invoke())}")
        member value after add one: 1
        >>>
        >>> counter.load.invoke()
        >>> print(f"member value after load state(back to 0): {yr.get(counter.get.invoke())}")
        member value after load state(back to 0): 0
    """
    remote_runtime = ConfigManager().in_cluster and not ConfigManager().is_driver
    if not remote_runtime:
        raise RuntimeError(
            "load_state is only supported on cloud with posix api")

    if ConfigManager().local_mode:
        raise RuntimeError("load_state is not supported in local mode")
    utils.Validator.check_timeout(timeout_sec)
    _logger.debug("Start to load instance state, timeout: %d seconds", timeout_sec)
    runtime_holder.global_runtime.get_runtime().load_state(timeout_sec)
    _logger.info("Succeeded to load instance state")


def get_instance(name: str, namespace: str = "", timeout: int = 60) -> instance_proxy.InstanceProxy:
    """
    Get an instance handle based on the name and namespace of the named instance.
    The interface call will block until the instance handle is obtained or a timeout occurs.

    Args:
        name (str): The name of the instance (instance ID).
        namespace (str): The namespace.
        timeout (int): The timeout in seconds.

    Returns:
        A Python named instance handle. Data type is InstanceProxy.

    Raises:
        TypeError: If the parameter types are incorrect.
        RuntimeError: If the instance does not exist.
        TimeoutError: If a timeout occurs.

    Examples:
        >>> yr.get_instance("name")
    """

    if not isinstance(name, str):
        raise TypeError(f"invalid name type: {type(name)}, should be a str")
    if namespace and not isinstance(namespace, str):
        raise TypeError(f"invalid namespace type: {type(namespace)}, "
                        f"should be a str")
    if timeout < 0:
        raise TypeError("The timeout value needs to be greater than 0")
    try:
        ins_proxy = instance_proxy.get_instance_by_name(name, namespace, timeout)
    except TimeoutError as e:
        raise TimeoutError(
            f"failed to get instance: {namespace + '-' + name if namespace else name} by name, {str(e)}") from e
    except Exception as e:
        raise RuntimeError(
            f"failed to get instance: {namespace + '-' + name if namespace else name} by name, {str(e)}") from e
    return ins_proxy


@check_initialized
def resource_group_table(resource_group_id: Union[str, None]):
    """get resource group table"""
    if ConfigManager().local_mode:
        raise RuntimeError("resources is not supported in local mode")
    return runtime_holder.global_runtime.get_runtime().resource_group_table(
        resource_group_id if resource_group_id is not None else "")


@check_initialized
def resources() -> List[dict]:
    """
    Get the resource information of nodes in the cluster.

    When requesting resource information, you need to configure `master_addr_list` in `yr.Config`.

    Returns:
        list[dict], The resource information of nodes in the cluster. The dict contains the following keys,

        - id: The name of the node.
        - capacity: The total resources of the node.
        - allocatable: The available resources of the node.
        - status: The status of the node. 0 indicates normal.

    Raises:
        RuntimeError: If the information retrieval from the functionsystem master fails.

    Examples:
        >>> import yr
        >>> yr.init()
        >>>
        >>> res = yr.resources()
        >>> print(res)
        [{'id': 'function-agent-172.17.0.2-25742','status': 0, 'capacity': {'CPU': 1000.0, 'Memory': 8192.0},
        'allocatable': {'CPU': 500.0, 'Memory': 4096.0}}]
        >>>
        >>> yr.finalize()
    """
    if ConfigManager().local_mode:
        raise RuntimeError("resources is not supported in local mode")
    return runtime_holder.global_runtime.get_runtime().resources()


@check_initialized
def get_node_ip_address():
    """
    Obtain the node ip.
    Examples:
        >>> import yr
        >>> yr.init()
        >>> node_ip = yr.get_node_ip_address()
        >>> print(node_ip)
        >>> yr.finalize()
    """
    return runtime_holder.global_runtime.get_runtime().get_node_ip_address()


@check_initialized
def get_node_id():
    """"""
    return runtime_holder.global_runtime.get_runtime().get_node_id()


@check_initialized
def get_namespace():
    """get namespace"""
    return runtime_holder.global_runtime.get_runtime().get_namespace()


@check_initialized
def create_resource_group(bundles: List[Dict[str, float]], name: Optional[str] = None,
                          strategy: Optional[str] = "PACK") -> ResourceGroup:
    """
    Asynchronously create a ResourceGroup.

    Note:
        - `bundle_index` is the index of each `bundle` in the list `bundles`.
        - Valid resource names include: "CPU", "Memory", and custom resources.
          Custom resources currently support "GPU/XX/YY" and "NPU/XX/YY", where "XX" is the card model,
          such as "Ascend910B4", and "YY" can be "count", "latency", or "stream".
        - For each `bundle`, if "CPU" or "Memory" is not set, or is set to ``0``,
          they will be assigned ``300`` and ``128`` respectively.
        - The amount of resources in each `bundle` must be greater than ``0``;
          if less than ``0``, an exception will be thrown. Resource names must not be empty;
          if empty, an exception will be thrown.
        - Other resources besides "CPU" and "Memory" will be filtered out if set to ``0``.

    Args:
        bundles (List[Dict[str, float]]): A set of `bundles` representing resource requests, must not be empty.
        name (Optional[str], optional): The name of the ResourceGroup to be created, which must be unique and
            cannot be 'primary' or an empty string. This parameter is optional, with a default value of ``None``,
            meaning a rgroup-{uuid} type string will be randomly generated as the `resource group name`.
        strategy (OptionsOptional[str], optional): The strategy to create the resource group, defalut strategy is ``PACK``
            
            - ``'None'``: No strategy.

            - ``'PACK'``: Pack multiple bundles into the same node as much as possible.
            
            - ``'SPREAD`'': Distribute multiple bundles across different nodes as much as possible.
            
            - ''`STRICT_PACK`'': All bundles must be placed on the same node, otherwise creation fails.
           
            - ''`STRICT_SPREAD`'': All bundles must be placed on different nodes, otherwise creation fails.

    Returns:
        A ResourceGroup handle.

    Raises:
        TypeError: If the parameter type is incorrect.
        ValueError: If the parameter is empty.
        RuntimeError: If there are insufficient resources to create the resource group.
        RuntimeError: If the resource group is created repeatedly.
        RuntimeError: If the resource group name is invalid.

    Examples:
        >>> rg1 = yr.create_resource_group([{"NPU/Ascend910B4/count":1},{"CPU":2000,"Memory":2000}])
        >>>
        >>> rg2 = yr.create_resource_group([{"NPU/Ascend910B4/count":1},{"CPU":2000,"Memory":2000}], "rgname")
    """
    if not isinstance(bundles, list):
        raise TypeError(f"invalid bundles type, actual: {type(bundles)}, expect: list.")
    if not bundles:
        raise ValueError(
            "invalid bundles number, actual: 0, expect: > 0."
        )
    if name is None:
        name = utils.generate_resource_group_name()
    if not isinstance(name, str):
        raise TypeError(f"invalid name type: {type(name)}, should be a str")
    request_id = runtime_holder.global_runtime.get_runtime().create_resource_group(bundles, name, strategy)
    return ResourceGroup(name, request_id, bundles)


@check_initialized
def remove_resource_group(resource_group: Union[str, ResourceGroup]):
    """
    Asynchronously delete a ResourceGroup.

    Args:
        resource_group (Union[str, ResourceGroup]): The name of the ResourceGroup to be deleted,
            or the handle returned by creating a ResourceGroup.

    Raises:
        TypeError: The input parameter is not of type str or ResourceGroup.
        RuntimeError: The ResourceGroup name is invalid.

    Examples:
        >>> yr.remove_resource_group("rgname")
        >>>
        >>> rg = yr.create_resource_group([{"NPU":1},{"CPU":2000,"Memory":2000}], "rgname")
        >>> yr.remove_resource_group(rg)
    """
    name = resource_group
    if isinstance(resource_group, ResourceGroup):
        name = resource_group.resource_group_name
    if not isinstance(name, str):
        raise TypeError(f"invalid name type: {type(name)}, should be a str")
    runtime_holder.global_runtime.get_runtime().remove_resource_group(name)


class cpp_instance_class:
    """Constructs a proxy for a C++ class to enable remote invocation."""

    def __init__(self, class_name: str, factory_name: str, function_urn: str):
        """
        init.

        Args:
            class_name (str): cpp class name.
            factory_name (str): Name of the static factory function of the cpp class.
            function_urn (str): Function URN, Defaults to
                sn:cn:yrk:12345678901234561234567890123456:function:0-defaultservice-cpp:$latest.

        Examples:
            .. code-block:: cpp

                #include <iostream>
                #include "yr/yr.h"
                class Counter {
                public:
                    int count;

                    Counter() {}
                    Counter(int init) { count = init; }

                    static Counter* FactoryCreate(int init) {
                        std::cout << "Counter FactoryCreate with init=" << init << std::endl;
                        return new Counter(init);
                    }

                    int Add(int x) {
                        count += x;
                        std::cout << "Counter Add with init=" << count << std::endl;
                        return count;
                    }

                    YR_STATE(count);
                };

                YR_INVOKE(Counter::FactoryCreate, &Counter::Add);

            .. code:: python

                >>> import yr
                >>> yr.init()
                >>> cpp_function_urn = (
                ...     "sn:cn:yrk:12345678901234561234567890123456:"
                ...     "function:0-yr-defaultservice-cpp:$latest"
                ... )
                >>> counter_class = yr.cpp_instance_class("Counter", "Counter::FactoryCreate", cpp_function_urn)
                >>> opt = yr.InvokeOptions(cpu=1000, memory=1024)
                >>> ins = counter_class.options(opt).invoke(11)
                >>> result = ins.Add.invoke(9)
                >>> yr.get(result)
                >>> ins.terminate()
                >>> yr.finalize()
        """
        self.__class_name__ = class_name
        self.__factory_name__ = factory_name
        self.__function_urn__ = function_urn

    def invoke(self, *args, **kwargs) -> InstanceProxy:
        """
        Create an instance of a cpp class.

        Returns:
            The corresponding handle to the cpp instance.
        """
        creator = instance_proxy.make_cpp_instance_creator(self)
        return creator.invoke(*args, **kwargs)

    def options(self, invoke_options: InvokeOptions) -> None:
        """
        Set user invoke options

        Args:
            invoke_options (InvokeOptions): invoke options for users to set resources
        """
        creator = instance_proxy.make_cpp_instance_creator(self)
        return creator.options(invoke_options)

    def get_class_name(self) -> str:
        """
        Get class name in this CPP instance

        Returns:
            class name in this CPP instance
        """
        return self.__class_name__

    def get_factory_name(self) -> str:
        """
        Get factory function name in this CPP instance

        Returns:
            factory function name in this CPP instance
        """
        return self.__factory_name__

    def get_function_key(self) -> str:
        """
        Get function key in this CPP instance

        Returns:
            function key in this CPP instance
        """
        return utils.get_function_from_urn(self.__function_urn__)


def cpp_function(function_name: str, function_urn: str) -> FunctionProxy:
    """
    A proxy for constructing cpp functions and remotely calling cpp functions.

    Args:
        function_name (str): cpp function name.
        function_urn (str): The URN (Uniform Resource Name) of cpp function.

    Returns:
        Return a proxy object for the remote C++ function.
        Data type is FunctionProxy.

    Examples:
        .. code-block:: cpp

            #include "yr/yr.h"
            int Square(int x)
            {
                return x * x;
            }

            // Define the stateless function Square
            YR_INVOKE(Square)

        .. code:: python

            >>> import yr
            >>> yr.init()
            >>> cpp_function_urn = (
            ...     "sn:cn:yrk:12345678901234561234567890123456:"
            ...     "function:0-yr-defaultservice-cpp:$latest"
            ... )
            >>> square_func = yr.cpp_function("Square", cpp_function_urn)
            >>> result = square_func.invoke(5)
            >>> print(yr.get(result))
            >>> yr.finalize()

    """
    return function_proxy.make_cross_language_function_proxy(function_name, function_urn, LanguageType.Cpp)


def cpp_instance_class_new(class_name: str, factory_name: str, function_urn: str) -> InstanceCreator:
    """
    cpp instance class
    """
    function_key = utils.get_function_from_urn(function_urn)
    return InstanceCreator.create_cross_user_class(
        CrossLanguageInfo(class_name=class_name, function_key=function_key, target_language=LanguageType.Cpp,
                          function_name=factory_name))


def java_function(class_name: str, function_name: str, function_urn: str) -> FunctionProxy:
    """
    A proxy used to construct java functions and remotely call java functions.

    Args:
        class_name (str): The name of the Java class .
        function_name (str):  The name of the Java function.
        function_urn (str): The URN (Uniform Resource Name) of the Java function.

    Returns:
        The corresponding function proxy.
        Data type is FunctionProxy.

    Examples:
        .. code:: java

            package org.yuanrong.demo;

            public class PlusOne{
                public static int PlusOne(int x) {
                    System.out.println("PlusOne with x=" + x);
                    return x + 1;
                }
            }

        .. code:: python

            >>> import yr
            >>> yr.init()
            >>> java_function_urn = (
            ...     "sn:cn:yrk:12345678901234561234567890123456:"
            ...     "function:0-yr-defaultservice-java:$latest"
            ... )
            >>> java_add = yr.java_function("org.yuanrong.demo.PlusOne", "PlusOne", java_function_urn)
            >>> result = java_add.invoke(1)
            >>> print(yr.get(result))

    """
    function_key = utils.get_function_from_urn(function_urn)
    return FunctionProxy(None, CrossLanguageInfo(function_name=function_name, function_key=function_key,
                                                 target_language=LanguageType.Java, class_name=class_name))


def java_instance_class(class_name: str, function_urn: str) -> InstanceCreator:
    """
    A proxy used to construct Java classes and invoke them remotely. 

    Args:
        class_name (str): The name of java.
        function_urn (str, optional): The urn of java,
            default is sn:cn:yrk:12345678901234561234567890123456:function:0-defaultservice-java:$latest.

    Returns:
        The corresponding instance creator.


    Examples:
        .. code:: java

            package org.yuanrong.demo;

            public class Counter {
                private int count;

                public Counter(int init) {
                    System.out.println("Counter constructor with init=" + init);
                    this.count = init;
                }

                public int Add(int value) {
                    this.count += value;
                    System.out.println("Add called, new count=" + this.count);
                    return this.count;
                }

                public int Get() {
                    System.out.println("Get called, count=" + this.count);
                    return this.count;
                }
            }

        .. code:: python

            >>> import yr
            >>> yr.init()
            >>> java_function_urn = (
            ...     "sn:cn:yrk:12345678901234561234567890123456:"
            ...     "function:0-yr-defaultservice-java:$latest"
            ... )
            >>>
            >>> java_instance = yr.java_instance_class("org.yuanrong.demo.Counter", java_function_urn).invoke(1)
            >>> res = java_instance.Add.invoke(5)
            >>> print(yr.get(res))
            >>>
            >>> res = java_instance.Get.invoke()
            >>> print(yr.get(res))
            >>>
            >>> java_instance.terminate()
            >>> yr.finalize()

    """
    function_key = utils.get_function_from_urn(function_urn)
    return InstanceCreator.create_cross_user_class(
        CrossLanguageInfo(class_name=class_name, function_key=function_key, target_language=LanguageType.Java,
                          function_name=""))


def go_function(function_name: str, function_urn: str) -> FunctionProxy:
    """
    Make proxy for go function
    """
    return function_proxy.make_cross_language_function_proxy(function_name, function_urn, LanguageType.Go)


def go_instance_class(function_name: str, function_urn: str) -> InstanceCreator:
    """
    go instance class
    """
    function_key = utils.get_function_from_urn(function_urn)
    return InstanceCreator.create_cross_user_class(
        CrossLanguageInfo(class_name="", function_key=function_key, target_language=LanguageType.Go,
                          function_name=function_name))


@check_initialized
def list_named_instances(all_namespaces: bool = False):
    """
    List named instances.

    Note:
        This interface does not support querying named instances under a specified namespace.
        If an instance is configured with a namespace, the namespace and instance name will be connected using a `-`.

    Examples:
        >>> import yr
        >>> yr.init()
        >>> named_instances = yr.list_named_instances()
        >>> print(named_instances)
        >>> yr.finalize()
    """
    all_actors = runtime_holder.global_runtime.get_runtime().query_named_instances()
    if all_namespaces:
        return [{"name": part.split("-")[1], "namespace": part.split("-")[0]} for part in all_actors]
    ns = get_namespace()
    return [part.split("-")[1] for part in all_actors if part.split("-")[0] == ns]
