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

"""cluster mode runtime"""

import logging
from typing import Any, Dict, List, Tuple, Union, Callable

from yr.exception import YRInvokeError
from yr.err_type import ErrorCode, ErrorInfo, ModuleCode
from yr.common.types import InvokeArg, GroupInfo
from yr.config import InvokeOptions, GroupOptions
from yr.config_manager import ConfigManager
from yr.fnruntime import Consumer, Fnruntime, Producer, SharedBuffer
from yr.libruntime_pb2 import ApiType, FunctionMeta
from yr.common.utils import GaugeData, UInt64CounterData, DoubleCounterData
from yr.object_ref import ObjectRef
from yr.base_runtime import BaseRuntime, AlarmInfo, SetParam, MSetParam, CreateParam, GetParams
from yr.serialization import Serialization
from yr.stream import ProducerConfig, SubscriptionConfig
from yr.accelerate.shm_broadcast import Handle

_logger = logging.getLogger(__name__)


class ClusterModeRuntime(BaseRuntime):
    """
    Cluster mode runtime
    """

    def __init__(self):
        """
        init
        """
        self.libruntime: Fnruntime = None
        self.__enable_flag = False

    def init(self) -> None:
        """Initialize ClusterModeRuntime.
        """
        function_system_ip_addr = ""
        function_system_port = 0
        function_system_rt_server_ip_addr = ""
        function_system_rt_server_port = 0
        if ConfigManager().server_address != "":
            function_system_ip_addr = ConfigManager().server_address.split(":")[
                0]
            function_system_port = int(
                ConfigManager().server_address.split(":")[1])
        _logger.debug(
            "Initialize libruntime with functionsystem address: %s:%d",
            function_system_ip_addr,
            function_system_port,
        )

        if ConfigManager().rt_server_address != "":
            function_system_rt_server_ip_addr = ConfigManager(
            ).rt_server_address.split(":")[0]
            function_system_rt_server_port = int(
                ConfigManager().rt_server_address.split(":")[1])
        _logger.debug(
            "Initialize libruntime with functionsystem rt-server address: %s:%d",
            function_system_rt_server_ip_addr,
            function_system_rt_server_port,
        )

        self.libruntime = Fnruntime(function_system_ip_addr,
                                    function_system_port,
                                    function_system_rt_server_ip_addr,
                                    function_system_rt_server_port)
        _logger.debug(
            "Initialize libruntime with datasystem address: %s", ConfigManager().ds_address)
        self.libruntime.init(Serialization())
        self.__enable_flag = True

    def set_initialized(self):
        """set initialized """
        self.__enable_flag = True

    def put(self, obj, create_param: CreateParam) -> str:
        """
        Put data to ds
        :param obj: user object
        :param create_param: create param for datasystem
        :return: object id
        """
        serialized_object = Serialization().serialize(obj)
        return self.libruntime.put(serialized_object, create_param)

    def get(self, ids: List[str], timeout: int, allow_partial: bool) -> List[Any]:
        """
        Get data from ds
        :param ids: ids which need to get
        :param timeout: timeout
        :param allow_partial: allow partial
        :return: data which get from ds
        """
        if timeout < 0 and timeout != -1:
            raise RuntimeError(
                f"Invalid parameter, timeout: {timeout}, expect -1 or > 0")

        timeout_ms = -1 if timeout == -1 else timeout * 1000
        results = self.libruntime.get(ids, timeout_ms, allow_partial)
        objects = Serialization().deserialize(results)
        for obj in objects:
            if isinstance(obj, YRInvokeError):
                raise obj.origin_error()
        return objects

    def wait(self, objs: List[str], wait_num: int, timeout: int) -> Tuple[List[str], List[str]]:
        """
        wait objects
        :param objs: objs which need to wait
        :param wait_num: wait number
        :param timeout: timeout
        :return: ready objects, unready objects, exception to objects
        """
        ready_ids, unready_ids, exception_ids = self.libruntime.wait(
            objs, wait_num, timeout)
        if exception_ids:
            ready_ids.extend(exception_ids)
        if len(ready_ids) > wait_num:
            unready_ids = ready_ids[wait_num:] + unready_ids
            ready_ids = ready_ids[:wait_num]
        return ready_ids, unready_ids

    def set_get_callback(self, object_id: str, callback: Callable):
        """
        set get callback
        :param object_id: object id
        :param callback: user callback
        :return: None
        """
        self._check_init()

        def callback_wrapper(object_id, err, buf):
            if err.error_code != ErrorCode.ERR_OK:
                callback(err)
            else:
                try:
                    res = Serialization().deserialize(buf)
                    if isinstance(res, YRInvokeError):
                        callback(res.origin_error())
                    else:
                        callback(res)
                except Exception as e:
                    callback(ErrorInfo(ErrorCode.ERR_INNER_SYSTEM_ERROR,
                             ModuleCode.RUNTIME, str(e)))

        self.libruntime.get_async(object_id, callback_wrapper)

    def kv_write(self, key: str, value: bytes, set_param: SetParam = SetParam()) -> None:
        """
        store value to ds
        :param key: key
        :param value: value
        :param set_param: set param
        :return: None
        """
        self.libruntime.kv_write(key, value, set_param)

    def kv_m_write_tx(self, keys: List[str], values: List[bytes], m_set_param: MSetParam = MSetParam()) -> None:
        """
        store multiple key-value pairs to ds
        :param keys: the keys to set
        :param values: the values to set. Size of values should equal to size of keys.
        :param m_set_param: MSetParams for datasystem stateClient.
        :return: None
        """
        self.libruntime.kv_m_write_tx(keys, values, m_set_param)

    def kv_read(self, key: Union[str, List[str]], timeout: int) -> List[Union[bytes, memoryview]]:
        """
        read value from ds
        :param key: key
        :param timeout: timeout
        :return: value
        """
        timeout_ms = -1 if timeout == -1 else timeout * 1000
        return self.libruntime.kv_read(key, timeout_ms)

    def kv_get_with_param(self, keys: List[str], get_params: GetParams, timeout: int) \
            -> List[Union[bytes, memoryview]]:
        """
        get value with param from ds
        :param keys: keys
        :param get_params: get params for datasystem
        :param timeout: timeout
        :return: values
        """
        timeout_ms = -1 if timeout == -1 else timeout * 1000
        return self.libruntime.kv_get_with_param(keys, get_params, timeout_ms)

    def kv_del(self, key: Union[str, List[str]]) -> None:
        """
        del value in ds
        :param key: key
        :return: None
        """
        self.libruntime.kv_del(key)

    def increase_global_reference(self, object_ids: List[str]) -> None:
        """
        increase global reference for object
        :param object_ids: objects
        :return: None
        """
        if not self.__enable_flag:
            return
        self.libruntime.increase_global_reference(object_ids)

    def decrease_global_reference(self, object_ids: List[str]) -> None:
        """
        decrease global reference for object
        :param object_ids: objects
        :return: None
        """
        if not self.__enable_flag:
            return
        self.libruntime.decrease_global_reference(object_ids)

    def invoke_by_name(self, func_meta: FunctionMeta, args: List[Any], opt: InvokeOptions,
                       return_nums: int, group_info: GroupInfo = None) -> List[str]:
        """
        invoke instance by name
        :param func_meta: function meta
        :param args: args
        :param opt: options
        :param return_nums: return values amount
        :param group_info: group info
        :return: None
        """
        self._check_init()
        if func_meta.apiType == ApiType.Faas:
            return self.libruntime.invoke_by_name(func_meta, self._package_python_args(args, True), opt, return_nums)
        return self.libruntime.invoke_by_name(func_meta, self._package_python_args(args), opt, return_nums, group_info)

    def create_instance(self, func_meta: FunctionMeta, args: List[Any], opt: InvokeOptions,
                        group_info: GroupInfo = None) -> str:
        """
        create instance
        :param func_meta: function_meta
        :param args: args
        :param opt: options
        :param group_info: group info
        :return: instance id
        """
        self._check_init()
        return self.libruntime.create_instance(func_meta, self._package_python_args(args), opt, group_info)

    def invoke_instance(self, func_meta: FunctionMeta, instance_id: str, args: List[Any],
                        opt: InvokeOptions, return_nums: int) -> List[str]:
        """
        invoke instance by id
        :param func_meta: function meta
        :param instance_id: instance id
        :param args: args
        :param opt: options
        :param return_nums: return values amount
        :return: object ids
        """
        self._check_init()
        return self.libruntime.invoke_instance(func_meta, instance_id, self._package_python_args(args), opt,
                                               return_nums)

    def cancel(self, object_ids: List[str], allow_force: bool, allow_recursive: bool) -> None:
        """
        cancel stateless function
        :param object_id: object id which need to cancel
        :return: None
        """
        self.libruntime.cancel(object_ids, allow_force, allow_recursive)

    def terminate_instance(self, instance_id: str) -> None:
        """
        terminate instance
        :param instance_id: instance id which need to terminate
        :return: None
        """
        self.libruntime.terminate_instance(instance_id)

    def terminate_instance_sync(self, instance_id: str) -> None:
        """
        terminate instance sync
        :param instance_id: instance id which need to terminate
        :return: None
        """
        self.libruntime.terminate_instance_sync(instance_id)

    def terminate_group(self, group_name: str) -> None:
        """
        terminate group
        :param group_name: group name which need to terminate
        :return: None
        """
        self.libruntime.terminate_group(group_name)

    def exit(self) -> None:
        """
        exit current instance. only support in runtime.
        :return: None
        """
        self.libruntime.exit()

    def receive_request_loop(self) -> None:
        """
        main loop
        :return: None
        """
        self.libruntime.receive_request_loop()

    def finalize(self) -> None:
        """
        finalize
        :return: None
        """
        self.libruntime.finalize()
        self.__enable_flag = False

    def create_stream_producer(self, stream_name: str, config: ProducerConfig) -> Producer:
        """
        create stream producer
        :param stream_name: stream name
        :param config: ProducerConfig
        :return: producer
        """
        if config.max_stream_size < 0:
            raise RuntimeError(
                f"Invalid parameter, max_stream_size: {config.max_stream_size}, expect >= 0")
        if config.retain_for_num_consumers < 0:
            raise RuntimeError(
                f"Invalid parameter, retain_for_num_consumers: {config.retain_for_num_consumers}, expect >= 0")
        if config.reserve_size < 0:
            raise RuntimeError(
                f"Invalid parameter, reserve_size: {config.reserve_size}, expect >= 0")
        return self.libruntime.create_stream_producer(stream_name, config)

    def create_stream_consumer(self, stream_name: str, config: SubscriptionConfig) -> Consumer:
        """
        create stream consumer
        :param stream_name: stream name
        :param config: SubscriptionConfig
        :return: consumer
        """
        return self.libruntime.create_stream_consumer(stream_name, config)

    def delete_stream(self, stream_name: str) -> None:
        """
        delete stream
        :param stream_name: stream name
        :return: None
        """
        self.libruntime.delete_stream(stream_name)

    def query_global_producers_num(self, stream_name: str) -> int:
        """
        query global producers num
        :param stream_name: stream name
        :return: producers num
        """
        return self.libruntime.query_global_producers_num(stream_name)

    def query_global_consumers_num(self, stream_name: str) -> int:
        """
        query global consumers num
        :param stream_name: stream name
        :return: consumers num
        """
        return self.libruntime.query_global_consumers_num(stream_name)

    def get_real_instance_id(self, instance_id: str) -> str:
        """
        get real instance id
        :param instance_id: instance key for instance id
        :return: real instance id
        """
        return self.libruntime.get_real_instance_id(instance_id)

    def save_real_instance_id(self, instance_id: str, need_order: bool):
        """
        save real instance id
        :param instance_id: instance id
        :param need_order: need order
        :return: None
        """
        return self.libruntime.save_real_instance_id(instance_id, need_order)

    def is_object_existing_in_local(self, object_id: str) -> bool:
        """
        is object existed in local
        :param object_id: object id
        :return: is existed
        """
        return self.libruntime.is_object_existing_in_local(object_id)

    def save_state(self, timeout_sec: int) -> None:
        """
        Saves the state of the current stateful function.
        Args:
            timeout_sec (int): The timeout value in seconds for saving the state.
        Returns:
            None
        """
        timeout_ms = -1 if timeout_sec == -1 else timeout_sec * 1000
        self.libruntime.save_state(timeout_ms)

    def load_state(self, timeout_sec: int) -> None:
        """
        Loads the state of the current stateful function.
        Args:
            timeout_sec (int): The timeout value in seconds for loading the state.
        Returns:
            None
        """
        timeout_ms = -1 if timeout_sec == -1 else timeout_sec * 1000
        self.libruntime.load_state(timeout_ms)

    def create_resource_group(self, bundles: List[Dict[str, float]], name: str, strategy: str) -> str:
        """
        Create resource group.
        Args:
            bundles (List[Dict[str, float]]): The bundles in resource group.
            name: (str): The name of resource group
            strategy: (str): The created strategy of resource group
        Returns:
            request id
        """
        return self.libruntime.create_resource_group(bundles, name, strategy)

    def remove_resource_group(self, name: str) -> None:
        """
        Remove resource group.
        Args:
            name: (str): The name of resource group
        Returns:
            None
        """
        self.libruntime.remove_resource_group(name)

    def wait_resource_group(self, name: str, request_id: str, timeout_seconds: int) -> None:
        """
        Wait resource group.
        Args:
            name: (str): The name of resource group
            request_id: (str): The request id of resource group
            timeout_seconds: (int): timeout for wait create resource group response.
        Returns:
            None
        """
        self.libruntime.wait_resource_group(name, request_id, timeout_seconds)

    def set_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        Set uint64 counter metrics
        Args:
            data: UInt64CounterData
        Returns:
            None
        """
        self.libruntime.set_uint64_counter(data)

    def reset_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        Reset uint64 counter metrics
        Args:
            data: UInt64CounterData
        Returns:
            None
        """
        self.libruntime.reset_uint64_counter(data)

    def increase_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        Increase uint64 counter metrics
        Args:
            data: UInt64CounterData
        Returns:
            None
        """
        self.libruntime.increase_uint64_counter(data)

    def get_value_uint64_counter(self, data: UInt64CounterData) -> int:
        """
        Get value of uint64 counter metrics
        Args:
            data: UInt64CounterData
        Returns:
            value
        """
        return self.libruntime.get_value_uint64_counter(data)

    def set_double_counter(self, data: DoubleCounterData) -> None:
        """
        Set double counter metrics
        Args:
            data: DoubleCounterData
        Returns:
            None
        """
        self.libruntime.set_double_counter(data)

    def reset_double_counter(self, data: DoubleCounterData) -> None:
        """
        Reset double counter metrics
        Args:
            data: DoubleCounterData
        Returns:
            None
        """
        self.libruntime.reset_double_counter(data)

    def increase_double_counter(self, data: DoubleCounterData) -> None:
        """
        Increase double counter metrics
        Args:
            data: DoubleCounterData
        Returns:
            None
        """
        self.libruntime.increase_double_counter(data)

    def get_value_double_counter(self, data: DoubleCounterData) -> float:
        """
        Get value of double counter metrics
        Args:
            data: DoubleCounterData
        Returns:
            value
        """
        return self.libruntime.get_value_double_counter(data)

    def report_gauge(self, data: GaugeData) -> None:
        """
        Report gauge metric.
        Args:
            data GaugeData.
        Returns:
            None
        """
        self.libruntime.report_gauge(data)

    def set_alarm(self, name: str, description: str, info: AlarmInfo) -> None:
        """
        Set alarm.
        Args:
            name: The name of alarm metrics
            description: The description of alarm metrics
            info: The alarm info
        Returns:
            None
        """
        self.libruntime.set_alarm(name, description, info)

    def peek_object_ref_stream(self, generator_id, blocking=True, timeout_ms=-1):
        """
        peek object reference stream.
        Args:
            timeout_ms int.
            generator_id str.
        Returns:
            object_id
        """
        self._check_init()
        result = self.libruntime.peek_object_ref_stream(
            generator_id, blocking, timeout_ms)
        if not isinstance(result, str):
            objects = Serialization().deserialize(result)
            for obj in objects:
                if isinstance(obj, YRInvokeError):
                    raise obj.origin_error()
        return result

    def generate_group_name(self) -> str:
        """
        generate group name.
        Returns:
            group name
        """
        return self.libruntime.generate_group_name()

    def get_instances(self, obj_id, group_name) -> List[str]:
        """
        get instance id list.
        Returns:
            instance id list
        """
        return self.libruntime.get_instances(obj_id, group_name)

    def resource_group_table(self, resource_group_id: str):
        """query resource group info"""
        return self.libruntime.resource_group_table(resource_group_id)

    def resources(self):
        """
        get node resources from functionsystem master
        """
        return self.libruntime.resources()

    def query_named_instances(self):
        """
        get actor name from functionsystem master
        """
        return self.libruntime.query_named_instances()

    def get_node_ip_address(self):
        """
        get node resources from functionsystem master
        """
        return self.libruntime.get_node_ip_address()

    def get_node_id(self):
        """
        """
        return self.libruntime.get_node_id()

    def get_namespace(self):
        """
        """
        return self.libruntime.get_namespace()

    def get_function_group_context(self):
        """
        get function group context.
        Returns:
            FunctionGroupContext
        """
        return self.libruntime.get_function_group_context()

    def get_instance_by_name(self, name, namespace, timeout) -> FunctionMeta:
        """
        get instance proxy by name.
        Returns:
            instance proxy
        """
        return self.libruntime.get_instance_by_name(name, namespace, timeout)

    def is_local_instances(self, instance_ids: List[str]) -> bool:
        """
        determine whether it is local mode
        """
        return self.libruntime.is_local_instances(instance_ids)

    def create_buffer(self, buffer_size: int) -> Tuple[str, SharedBuffer]:
        """
        create buffer
        """
        return self.libruntime.create_buffer(buffer_size)

    def get_buffer(self, obj_id: str) -> SharedBuffer:
        """
        get buffer
        """
        return self.libruntime.get_buffer(obj_id)

    def accelerate(self, group_name: str, handle: Handle):
        """
        accelerate
        """
        return self.libruntime.accelerate(group_name, handle)

    def add_return_object(self, obj_ids: List[str]):
        """
        add return objects
        """
        return self.libruntime.add_return_object(obj_ids)

    def _package_python_args(self, args_list, is_faas=False):
        """package python args"""
        args_list_new = []
        for arg in args_list:
            if isinstance(arg, ObjectRef):
                invoke_arg = InvokeArg(
                    buf=bytes(), is_ref=True, obj_id=arg.id, nested_objects=set())
            else:
                if is_faas:
                    invoke_arg = InvokeArg(buf=arg, is_ref=False, obj_id="",
                                           nested_objects=set())
                else:
                    serialized_arg = Serialization().serialize(arg)
                    invoke_arg = InvokeArg(buf=None, is_ref=False, obj_id="",
                                           nested_objects=set(
                                               [ref.id for ref in serialized_arg.nested_refs]),
                                           serialized_obj=serialized_arg)
            args_list_new.append(invoke_arg)
        return args_list_new

    def _check_init(self):
        """
        check init
        """
        if not self.__enable_flag:
            raise RuntimeError("runtime not enable")

    def create_group(self, group_name: str, group_opts: GroupOptions):
        if group_opts.timeout < 0 and group_opts.timeout != -1:
            raise RuntimeError(
                f"Invalid parameter, timeout: {group_opts.timeout}, expect -1 or > 0")
        self.libruntime.create_group(group_name, group_opts)

    def terminate_group(self, group_name: str):
        self.libruntime.terminate_group(group_name)

    def wait_group(self, group_name: str):
        self.libruntime.wait_group(group_name)

    def suspend_group(self, group_name: str):
        self.libruntime.suspend_group(group_name)

    def resume_group(self, group_name: str):
        self.libruntime.resume_group(group_name)
