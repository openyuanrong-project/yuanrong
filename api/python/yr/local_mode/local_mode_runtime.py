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

"""local mode runtime"""
import logging
import time
from abc import ABC
from concurrent import futures
from concurrent.futures import Future
from typing import Union, Dict, List, Any, Tuple, Callable
from yr.accelerate.shm_broadcast import Handle
from yr.common.types import GroupInfo
from yr.config import InvokeOptions, GroupOptions
from yr.exception import YRInvokeError
from yr.stream import ProducerConfig, SubscriptionConfig
from yr.common.utils import (
    generate_random_id,
    generate_task_id,
    GaugeData,
    UInt64CounterData,
    DoubleCounterData,
)
from yr.fnruntime import Producer, Consumer
from yr.local_mode.local_client import LocalClient
from yr.local_mode.local_object_store import LocalObjectStore
from yr.local_mode.task_manager import TaskManager
from yr.libruntime_pb2 import FunctionMeta, InvokeType
from yr.local_mode.task_spec import TaskSpec
from yr.base_runtime import (
    BaseRuntime,
    AlarmInfo,
    SetParam,
    MSetParam,
    CreateParam,
    GetParams,
)
from yr.fnruntime import SharedBuffer

_logger = logging.getLogger(__name__)


class LocalModeRuntime(BaseRuntime, ABC):
    """local mode runtime"""

    def __init__(self):
        self.__invoke_client = LocalClient()
        self.__enable_flag = False
        self.__local_store = LocalObjectStore()
        self.__task_mgs = TaskManager(self.__invoke_client)
        self.__object_ids = {}

    def init(self) -> None:
        """Initialize ClusterModeRuntime.
        """
        self.__enable_flag = True

    def put(self, obj, create_param: CreateParam) -> str:
        """
        Put data to ds
        :param obj: user object
        :param create_param: create param for datasystem
        :return: object id
        """
        key = generate_random_id()
        self.__local_store.put(key, obj)
        return key

    def put_serialized(self, serialized_object) -> str:
        """Deserialize and store a SerializedObject so worker.py can retrieve it as a callable."""
        key = generate_random_id()
        obj = self._deserialize_code_bytes(serialized_object.to_bytes())
        self.__local_store.put(key, obj)
        return key

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
        self.wait(ids, len(ids), timeout)
        objects = [self.__local_store.get(i) for i in ids]
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
        :return: ready objects, unready objects, exception of objects
        """
        start_time = time.time()
        ready_objs, unready_map, exist_exception = self._check_objs(objs)
        unready_futures = set(unready_map.keys())
        if len(unready_futures) == 0:
            return objs[:wait_num], objs[wait_num:]

        while len(ready_objs) < wait_num and not exist_exception:
            if timeout:
                iteration_timeout = timeout - time.time() + start_time
                if iteration_timeout <= 0:
                    break
            else:
                iteration_timeout = None
            ready, _ = futures.wait(
                unready_futures, iteration_timeout, futures.FIRST_COMPLETED)
            if not ready:
                break
            unready_futures -= ready
            for r in ready:
                tmp = unready_map.pop(r)
                ready_objs.extend(tmp)

        ready_objs.sort(key=objs.index)
        ready_objs = ready_objs[:wait_num]
        unready_objs = [obj for obj in objs if obj not in ready_objs]
        return ready_objs, unready_objs

    def set_get_callback(self, object_id: str, callback: Callable):
        """
        set get callback
        :param object_id: object id
        :param callback: user callback
        :return: None
        """
        self.__local_store.add_done_callback(object_id, callback)

    def kv_write(self, key: str, value: bytes, set_param: SetParam) -> None:
        """
        store value to ds
        :param key: key
        :param value: value
        :param set_param: set param
        :return: None
        """
        self.__local_store.put(key, value)

    def kv_m_write_tx(self, keys: List[str], values: List[bytes], m_set_param: MSetParam) -> None:
        """
        store multiple key-value pairs to ds
        :param keys: the keys to set
        :param values: the values to set. Size of values should equal to size of keys.
        :param m_set_param: MSetParams for datasystem stateClient.
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def kv_read(self, key: Union[str, List[str]], timeout: int = 0) -> List[Union[bytes, memoryview]]:
        """
        read value from ds
        :param key: key
        :param timeout: timeout
        :return: value
        """
        return self.__local_store.get(key)

    def kv_get_with_param(self, keys: List[str], get_params: GetParams, timeout: int) \
            -> List[Union[bytes, memoryview]]:
        """
        get value with param from ds
        :param keys: keys
        :param get_params: get params for datasystem
        :param timeout: timeout
        :return: values
        """
        raise RuntimeError("not support in local mode")

    def kv_del(self, key: Union[str, List[str]]) -> None:
        """
        del value in ds
        :param key: key
        :return: None
        """
        self.__local_store.release(key)

    def increase_global_reference(self, object_ids: List[str]) -> None:
        """
        increase global reference for object
        :param object_ids: objects
        :return: None
        """
        if not self.__enable_flag:
            return
        self.__local_store.increase_global_reference(object_ids)

    def decrease_global_reference(self, object_ids: List[str]) -> None:
        """
        decrease global reference for object
        :param object_ids: objects
        :return: None
        """
        if not self.__enable_flag:
            return
        self.__local_store.decrease_global_reference(object_ids)

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
        func_meta = self._ensure_code_in_store(func_meta)
        task_id = generate_task_id()
        f = Future()

        task_spec = TaskSpec(
            future=f,
            task_id=task_id,
            object_ids=[generate_random_id() for _ in range(return_nums)],
            function_meta=func_meta,
            trace_id=task_id,
            invoke_type=InvokeType.InvokeFunctionStateless,
            args=args,
            invoke_options=opt,
        )
        for i in task_spec.object_ids:
            self.__local_store.set_return_obj(i, f)
            self.__object_ids[i] = task_id
        self.__task_mgs.submit_task(task_spec)
        return task_spec.object_ids

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
        func_meta = self._ensure_code_in_store(func_meta)
        task_id = generate_task_id()
        f = Future()
        task_spec = TaskSpec(
            future=f,
            task_id=task_id,
            object_ids=[generate_random_id()],
            function_meta=func_meta,
            trace_id=task_id,
            invoke_type=InvokeType.CreateInstance,
            args=args,
            invoke_options=opt,
        )
        for i in task_spec.object_ids:
            self.__local_store.set_return_obj(i, f)
            self.__object_ids[i] = task_id
        return self.__invoke_client.create(task_spec)

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
        task_id = generate_task_id()
        f = Future()

        task_spec = TaskSpec(
            future=f,
            task_id=task_id,
            object_ids=[generate_random_id() for _ in range(return_nums)],
            function_meta=func_meta,
            trace_id=task_id,
            invoke_type=InvokeType.InvokeFunction,
            args=args,
            invoke_options=opt,
            instance_id=instance_id,
        )
        for i in task_spec.object_ids:
            self.__local_store.set_return_obj(i, f)
            self.__object_ids[i] = task_id
        self.__invoke_client.invoke(task_spec)
        return task_spec.object_ids

    def cancel(self, object_ids: List[str], allow_force: bool, allow_recursive: bool) -> None:
        """
        cancel stateless function
        :param object_ids: object id which need to cancel
        :return: None
        """
        for i in object_ids:
            if i not in self.__object_ids:
                raise RuntimeError("object without stateless function")
            self.__task_mgs.cancel(self.__object_ids.get(i))

    def terminate_instance(self, instance_id: str) -> None:
        """
        terminate instance
        :param instance_id: instance id which need to terminate
        :return: None
        """
        self.__invoke_client.kill(instance_id)

    def terminate_instance_sync(self, instance_id: str) -> None:
        """
        terminate instance sync
        :param instance_id: instance id which need to terminate
        :return: None
        """
        self.__invoke_client.kill(instance_id)

    def snapshot_instance(self, instance_id: str, ttl: int = -1, leave_running: bool = False) -> str:
        raise RuntimeError("not support in local mode")

    def snapstart_instance(self, checkpoint_id: str) -> str:
        raise RuntimeError("not support in local mode")

    def terminate_group(self, group_name: str) -> None:
        """
        terminate group
        :param group_name: group name which need to terminate
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def exit(self) -> None:
        """
        exit current instance. only support in runtime.
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def receive_request_loop(self) -> None:
        """
        main loop
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def finalize(self) -> None:
        """
        finalize
        :return: None
        """
        self.__enable_flag = False
        self.__local_store.clear()

    def create_stream_producer(self, stream_name: str, config: ProducerConfig) -> Producer:
        """
        create stream producer
        :param stream_name: stream name
        :param config: ProducerConfig
        :return: producer
        """
        raise RuntimeError("not support in local mode")

    def create_stream_consumer(self, stream_name: str, config: SubscriptionConfig) -> Consumer:
        """
        create stream consumer
        :param stream_name: stream name
        :param config: SubscriptionConfig
        :return: consumer
        """
        raise RuntimeError("not support in local mode")

    def delete_stream(self, stream_name: str) -> None:
        """
        delete stream
        :param stream_name: stream name
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def query_global_producers_num(self, stream_name: str) -> int:
        """
        query global producers num
        :param stream_name: stream name
        :return: producers num
        """
        raise RuntimeError("not support in local mode")

    def query_global_consumers_num(self, stream_name: str) -> int:
        """
        query global consumers num
        :param stream_name: stream name
        :return: consumers num
        """
        raise RuntimeError("not support in local mode")

    def get_real_instance_id(self, instance_id: str) -> str:
        """
        get real instance id
        :param instance_id: instance key for instance id
        :return: real instance id
        """
        return instance_id

    def save_real_instance_id(self, instance_id: str, need_order: bool) -> None:
        """
        save real instance id
        :param instance_id: instance id
        :param need_order: need order
        :return: None
        """

    def is_object_existing_in_local(self, object_id: str) -> bool:
        """
        is object existed in local
        :param object_id: object id
        :return: is existed
        """
        return self.__local_store.is_object_existing_in_local(object_id)

    def save_state(self, timeout_sec: int) -> None:
        """
        Saves the state of the current stateful function.
        Args:
            timeout_sec (int): The timeout value in seconds for saving the state.
        Returns:
            None
        """
        raise RuntimeError("not support in local mode")

    def load_state(self, timeout_sec: int) -> None:
        """
        Loads the state of the current stateful function.
        Args:
            timeout_sec (int): The timeout value in seconds for loading the state.
        Returns:
            None
        """
        raise RuntimeError("not support in local mode")

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
        raise RuntimeError("not support in local mode")

    def remove_resource_group(self, name: str) -> None:
        """
        Remove resource group.

        Args:
            name (str): The name of resource group.

        Raises:
            TypeError: The input parameter is neither of type str nor ResourceGroup.
            RuntimeError: The name of ResourceGroup is invalid.

        Examples:
            >>> yr.remove_resource_group("rgname")
            >>>
            >>> rg = yr.create_resource_group([{"NPU":1},{"CPU":2000,"Memory":2000}], "rgname")
            >>> yr.remove_resource_group(rg)

        """
        raise RuntimeError("not support in local mode")

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
        raise RuntimeError("not support in local mode")

    def set_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        set uint64 counter metrics
        :param data: UInt64CounterData
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def reset_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        reset uint64 counter metrics
        :param data: UInt64CounterData
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def increase_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        increase uint64 counter metrics
        :param data: UInt64CounterData
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def get_value_uint64_counter(self, data: UInt64CounterData) -> int:
        """
        get value of uint64 counter metrics
        :param data: UInt64CounterData
        :return: value
        """
        raise RuntimeError("not support in local mode")

    def set_double_counter(self, data: DoubleCounterData) -> None:
        """
        set double counter metrics
        :param data: DoubleCounterData
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def reset_double_counter(self, data: DoubleCounterData) -> None:
        """
        reset double counter metrics
        :param data: DoubleCounterData
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def increase_double_counter(self, data: DoubleCounterData) -> None:
        """
        increase double counter metrics
        :param data: DoubleCounterData
        :return: None
        """
        raise RuntimeError("not support in local mode")

    def get_value_double_counter(self, data: DoubleCounterData) -> float:
        """
        get value of double counter metrics
        :param data: DoubleCounterData
        :return: value
        """
        raise RuntimeError("not support in local mode")

    def report_gauge(self, data: GaugeData) -> None:
        """
        report gauge metrics
        :param data: GaugeData
        :return: None
        """
        raise RuntimeError("not support in local mode")

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
        raise RuntimeError("not support in local mode")

    def generate_group_name(self) -> str:
        """
        generate group name.
        Returns:
            group name
        """
        raise RuntimeError("not support in local mode")

    def get_instances(self, obj_id, group_name) -> List[str]:
        """
        get instance id list.
        Returns:
            instance id list
        """
        raise RuntimeError("not support in local mode")

    def resources(self):
        """
        get node resources from functionsystem master
        """
        return self.libruntime.resources()

    def query_named_instances(self):
        """
        get actor name from functionsystem master
        """
        raise RuntimeError("not support in local mode")

    def get_node_ip_address(self):
        """
        get node ip address from functionsystem master
        """
        raise RuntimeError("not support in local mode")

    def get_node_id(self):
        """
        """
        return RuntimeError("not support in local mode")

    def get_function_group_context(self):
        """
        get function group context.
        Returns:
            FunctionGroupContext
        """
        raise RuntimeError("not support in local mode")

    def is_local_instances(self, instance_ids: List[str]) -> bool:
        """
        is local instances
        Returns:
            bool
        """
        raise RuntimeError("not support in local mode")

    def accelerate(self, group_name: str, handle: Handle):
        """
        accelerate
        Returns:
            handle list
        """
        raise RuntimeError("not support in local mode")

    def create_buffer(self, buffer_size: int) -> Tuple[str, SharedBuffer]:
        """
        create buffer
        Returns:
            obj id and buffer
        """
        raise RuntimeError("not support in local mode")

    def get_buffer(self, obj_id: str) -> SharedBuffer:
        """
        get buffer
        Returns:
            buffer
        """
        raise RuntimeError("not support in local mode")

    def add_return_object(self, obj_ids: List[str]):
        """
        add return object
        Returns:
            bool
        """
        raise RuntimeError("not support in local mode")

    def create_group(self, group_name: str, group_opts: GroupOptions):
        raise RuntimeError("not support in local mode")

    def terminate_group(self, group_name: str):
        raise RuntimeError("not support in local mode")

    def wait_group(self, group_name: str):
        raise RuntimeError("not support in local mode")

    def suspend_group(self, group_name: str):
        raise RuntimeError("not support in local mode")

    def resume_group(self, group_name: str):
        raise RuntimeError("not support in local mode")

    def _check_objs(self, objs):
        exist_exception = False
        ready_objs = []
        unready_map = {}
        for object_id in objs:
            future = self.__local_store.get_future(object_id)
            if future is None:
                ready_objs.append(object_id)
                continue
            if future.done():
                ready_objs.append(object_id)
                if future.exception():
                    exist_exception = True
                continue
            # multiple return value:one future contain multiple obj
            if future in unready_map:
                unready_map[future].append(object_id)
            else:
                unready_map[future] = [object_id]
        return ready_objs, unready_map, exist_exception

    def _deserialize_code_bytes(self, raw: bytes) -> object:
        """Deserialize code bytes (SerializedObject binary format) into a callable."""
        import msgpack as _msgpack
        from yr.common import constants
        from yr.serialization.serializers import PySerializer, MessagePackSerializer

        # Buffer format: |metadata(8 bytes)|msgpack_size(8 bytes)|msgpack_data(N bytes)|py_data|
        meta_unpacker = _msgpack.Unpacker()
        meta_unpacker.feed(raw[:8])
        metadata = meta_unpacker.unpack()

        size_unpacker = _msgpack.Unpacker()
        size_unpacker.feed(raw[8:16])
        msgpack_data_size = size_unpacker.unpack()

        msgpack_data = raw[16:16 + msgpack_data_size]
        py_data = raw[16 + msgpack_data_size:]

        python_objects = []
        if py_data and constants.Metadata(metadata) == constants.Metadata.PYTHON:
            python_objects = PySerializer.deserialize(py_data)

        return MessagePackSerializer.deserialize(msgpack_data, python_objects)

    def _ensure_code_in_store(self, func_meta: FunctionMeta) -> FunctionMeta:
        """If code is passed inline (codeID empty, code bytes present), store it and return updated meta."""
        if func_meta.codeID or not func_meta.code:
            return func_meta
        obj = self._deserialize_code_bytes(bytes(func_meta.code))
        key = generate_random_id()
        self.__local_store.put(key, obj)
        new_meta = FunctionMeta()
        new_meta.CopyFrom(func_meta)
        new_meta.codeID = key
        return new_meta
