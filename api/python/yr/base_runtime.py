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

"""runtime interface"""

from abc import ABCMeta, abstractmethod
from dataclasses import dataclass, field
from enum import Enum
from typing import List, Tuple, Union, Any, Callable, Dict

from yr.common.types import GroupInfo
from yr.config import InvokeOptions, GroupOptions
from yr.libruntime_pb2 import FunctionMeta
from yr.common.utils import GaugeData, UInt64CounterData, DoubleCounterData
from yr.common import constants
from yr.accelerate.shm_broadcast import Handle


class ExistenceOpt(Enum):
    """Kv option. Whether key repeat writing is supported."""
    #: Supports repeated writing.
    NONE = 0
    #: Do not support repeated writing.
    NX = 1


class WriteMode(Enum):
    """Object write mode."""
    #: Not written to the second-level cache.
    NONE_L2_CACHE = 0
    #: Write data to the secondary cache synchronously to ensure data reliability.
    WRITE_THROUGH_L2_CACHE = 1
    #: Asynchronously write data to the secondary cache to ensure data reliability.
    WRITE_BACK_L2_CACHE = 2
    #: Not written to the second-level cache, and will be actively evicted by the system when system resources
    #: are insufficient.
    NONE_L2_CACHE_EVICT = 3


class CacheType(Enum):
    """Cache type.

    Identifies whether the allocated memory is a memory medium or a disk medium.
    """
    #: Memory media.
    MEMORY = 0

    #: Disk media.
    DISK = 1


class ConsistencyType(Enum):
    """Consistency type.

    In distributed scenarios, different levels of consistency semantics can be configured.
    """
    #: Asynchronous.
    PRAM = 0

    #: Causal consistency.
    CAUSAL = 1


@dataclass
class SetParam:
    """Set parameters."""
    #: Whether duplicate key writing is supported.
    #: Optional values: ``ExistenceOpt.NONE`` (supported, default) and ``ExistenceOpt.NX`` (not supported, optional).
    existence: ExistenceOpt = ExistenceOpt.NONE

    #: Configures data reliability. When a second-level cache (e.g., Redis) is enabled on the server,
    #: this setting ensures data integrity.
    #: Default value is ``WriteMode.NONE_L2_CACHE``, that is ``0``.
    write_mode: WriteMode = WriteMode.NONE_L2_CACHE

    #: Data lifetime, keys exceeding this time will be deleted. Default is ``0``,
    ttl_second: int = constants.DEFAULT_NO_TTL_LIMIT

    #: Used to indicate whether the assigned storage medium is memory or disk.
    #: Optional values are ``CacheType.Memory`` (memory medium) and ``CacheType.Disk`` (disk medium).
    #: The default value is ``CacheType.Memory``.
    cache_type: CacheType = CacheType.MEMORY


@dataclass
class MSetParam:
    """Represents the parameter configuration class for the mset operation."""
    #: Whether duplicate key writing is supported.
    #: Optional values: ``ExistenceOpt.NONE`` (supported, default) and ``ExistenceOpt.NX`` (not supported, optional).
    existence: ExistenceOpt = ExistenceOpt.NX

    #: Configures data reliability. When a second-level cache (e.g., Redis) is enabled on the server,
    #: this setting ensures data integrity.
    #: Default value is ``WriteMode.NONE_L2_CACHE``, that is ``0``.
    write_mode: WriteMode = WriteMode.NONE_L2_CACHE

    #: Data lifetime; keys exceeding this time will be deleted. Default is ``0``,
    #: meaning the key persists until explicitly deleted via the del interface.
    ttl_second: int = constants.DEFAULT_NO_TTL_LIMIT

    #: Used to indicate whether the assigned storage medium is memory or disk.
    #: Optional values are ``CacheType.Memory`` (memory medium) and ``CacheType.Disk`` (disk medium).
    #: The default value is ``CacheType.Memory``.
    cache_type: CacheType = CacheType.MEMORY


@dataclass
class CreateParam:
    """Create param."""

    def __init__(self):
        """
        Initialize a CreateParam instance.
        """
        pass

    #: Data consistency configuration.
    #: In a distributed scenario, different levels of consistency semantics can be configured.
    #: The optional parameters are ConsistencyType.PRAM (asynchronous) and ConsistencyType.CAUSAL (causal consistency).
    #: Defaults to ConsistencyType.PRAM.
    consistency_type: ConsistencyType = ConsistencyType.PRAM

    #: Indicates whether the allocated storage medium is memory or disk.
    #: The optional parameters are CacheType.Memory (memory medium) and CacheType.Disk (disk medium).
    #: The default value is CacheType.Memory.
    cache_type: CacheType = CacheType.MEMORY


class AlarmSeverity(Enum):
    """Object alarm severity."""
    #: Alarm is turned off.
    OFF = 0
    #: Informational alarm.
    INFO = 1
    #: MINOR alarm.
    MINOR = 2
    #: MAJOR alarm.
    MAJOR = 3
    #: Critical alarm.
    CRITICAL = 4


@dataclass
class GetParam():
    """Get the parameter configuration class."""
    #: Offset, default is ``0``.
    offset: int = 0
    #: Size, default is ``0``.
    size: int = 0


@dataclass
class GetParams():
    """Interface class for obtaining parameters."""
    #: A group of GetParam, the quantity needs to be greater than 0.
    get_params: List[GetParam] = field(default_factory=list)


@dataclass
class AlarmInfo:
    """Alarm info."""
    #: The name of the alarm. The default value is an empty string.
    alarm_name: str = ""
    #: The severity of the alarm, of type ``AlarmSeverity``, defaults to ``AlarmSeverity.OFF``.
    alarm_severity: AlarmSeverity = AlarmSeverity.OFF
    #: The location information of the alarm. The default value is an empty string.
    location_info: str = ""
    #: The description of the cause of the alarm.The default value is an empty string.
    cause: str = ""
    #: The timestamp when the alarm starts, with a default value of ``-1``.
    starts_at: int = -1
    #: The timestamp when the alarm ends, with a default value of ``-1``.
    ends_at: int = -1
    #: The timeout duration of the alarm, with a default value of ``-1``.
    timeout: int = -1
    #: Custom options, stored as a key-value pair dictionary, with a default value of an empty dictionary.
    custom_options: Dict[str, str] = field(default_factory=dict)


class BaseRuntime(metaclass=ABCMeta):
    """
    Base runtime
    """

    @abstractmethod
    def init(self) -> None:
        """
        :return:
        """

    @abstractmethod
    def put(self, obj, create_param: CreateParam) -> str:
        """
        Put data to ds
        :param obj: user object
        :param create_param: create param for datasystem
        :return: object id
        """

    @abstractmethod
    def get(self, ids: List[str], timeout: int, allow_partial: bool) -> List[Any]:
        """
        Get data from ds
        :param ids: ids which need to get
        :param timeout: timeout
        :param allow_partial: allow partial
        :return: data which get from ds
        """

    @abstractmethod
    def wait(self, objs: List[str], wait_num: int, timeout: int) -> Tuple[List[str], List[str]]:
        """
        wait objects
        :param objs: objs which need to wait
        :param wait_num: wait number
        :param timeout: timeout
        :return: ready objects, unready objects, exception of objects
        """

    @abstractmethod
    def set_get_callback(self, object_id: str, callback: Callable):
        """
        set get callback
        :param object_id: object id
        :param callback: user callback
        :return: None
        """

    @abstractmethod
    def kv_write(self, key: str, value: bytes, set_param: SetParam) -> None:
        """
        store value to ds
        :param key: key
        :param value: value
        :param set_param: set param
        :return: None
        """

    @abstractmethod
    def kv_m_write_tx(self, keys: List[str], values: List[bytes], m_set_param: MSetParam) -> None:
        """
        store multiple key-value pairs to ds
        :param keys: the keys to set
        :param values: the values to set. Size of values should equal to size of keys.
        :param m_set_param: MSetParams for datasystem stateClient.
        :return: None
        """

    @abstractmethod
    def kv_read(self, key: Union[str, List[str]], timeout: int) -> List[Union[bytes, memoryview]]:
        """
        read value from ds
        :param key: key
        :param timeout: timeout
        :return: value
        """

    @abstractmethod
    def kv_get_with_param(self, keys: List[str], get_params: GetParams, timeout: int) -> List[Union[bytes, memoryview]]:
        """
        get value with param from ds
        :param keys: keys
        :param get_params: get params for datasystem
        :param timeout: timeout
        :return: values
        """

    @abstractmethod
    def kv_del(self, key: Union[str, List[str]]) -> None:
        """
        del value in ds
        :param key: key
        :return: None
        """

    @abstractmethod
    def increase_global_reference(self, object_ids: List[str]) -> None:
        """
        increase global reference for object
        :param object_ids: objects
        :return: None
        """

    @abstractmethod
    def decrease_global_reference(self, object_ids: List[str]) -> None:
        """
        decrease global reference for object
        :param object_ids: objects
        :return: None
        """

    @abstractmethod
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

    @abstractmethod
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

    @abstractmethod
    def invoke_instance(self, func_meta: FunctionMeta, instance_id: str, args: List[Any],
                        opt: InvokeOptions, return_nums: int) -> List[str]:
        """
        invoke instance by id
        :param func_meta: function meta
        :param instance_id: instance id
        :param args: args
        :param opt: options
        :return: object ids
        """

    @abstractmethod
    def cancel(self, object_ids: List[str], allow_force: bool, allow_recursive: bool) -> None:
        """
        cancel stateless function
        :param object_id: object id which need to cancel
        :return: None
        """

    @abstractmethod
    def terminate_instance(self, instance_id: str) -> None:
        """
        terminate instance
        :param instance_id: instance id which need to terminate
        :return: None
        """

    @abstractmethod
    def terminate_instance_sync(self, instance_id: str) -> None:
        """
        terminate instance sync
        :param instance_id: instance id which need to terminate
        :return: None
        """

    @abstractmethod
    def terminate_group(self, group_name: str) -> None:
        """
        terminate group
        :param group_name: group name which need to terminate
        :return: None
        """

    @abstractmethod
    def snapshot_instance(self, instance_id: str, ttl: int = -1, leave_running: bool = False) -> str:
        """
        Create instance snapshot with signal 18
        :param instance_id: instance id to snapshot
        :param ttl: time-to-live for the snapshot in seconds
        :param leave_running: whether to keep instance running after snapshot
        :return: checkpointID
        """

    @abstractmethod
    def snapstart_instance(self, checkpoint_id: str) -> str:
        """
        Start instance from snapshot with signal 19
        :param checkpoint_id: checkpoint id to restore from
        :return: new instance id
        """

    @abstractmethod
    def exit(self) -> None:
        """
        exit current instance. only support in runtime.
        :return: None
        """

    @abstractmethod
    def receive_request_loop(self) -> None:
        """
        main loop
        :return: None
        """

    @abstractmethod
    def finalize(self) -> None:
        """
        finalize
        :return: None
        """

    @abstractmethod
    def get_real_instance_id(self, instance_id: str) -> str:
        """
        get real instance id
        :param instance_id: instance key for instance id
        :return: real instance id
        """

    @abstractmethod
    def save_real_instance_id(self, instance_id: str, need_order: bool) -> None:
        """
        save real instance id
        :param instance_id: instance id
        :param need_order: need order
        :return: None
        """

    @abstractmethod
    def create_stream_producer(self, stream_name: str, config):
        """
        create stream producer
        :param stream_name: stream name
        :param config: ProducerConfig
        :return: producer
        """

    @abstractmethod
    def create_stream_consumer(self, stream_name: str, config):
        """
        create stream consumer
        :param stream_name: stream name
        :param config: SubscriptionConfig
        :return: consumer
        """

    @abstractmethod
    def delete_stream(self, stream_name: str) -> None:
        """
        delete stream
        :param stream_name: stream name
        :return: None
        """

    @abstractmethod
    def query_global_producers_num(self, stream_name: str) -> int:
        """
        query global producers num
        :param stream_name: stream name
        :return: producers num
        """

    @abstractmethod
    def query_global_consumers_num(self, stream_name: str) -> int:
        """
        query global consumers num
        :param stream_name: stream name
        :return: consumers num
        """

    @abstractmethod
    def is_object_existing_in_local(self, object_id: str) -> bool:
        """
        is object existed in local
        :param object_id: object id
        :return: is existed
        """

    @abstractmethod
    def save_state(self, timeout_sec: int) -> None:
        """
        Saves the state of the current stateful function.
        Args:
            timeout_sec (int): The timeout value in seconds for saving the state.
        Returns:
            None
        """

    @abstractmethod
    def load_state(self, timeout_sec: int) -> None:
        """
        Loads the state of the current stateful function.
        Args:
            timeout_sec (int): The timeout value in seconds for loading the state.
        Returns:
            None
        """

    @abstractmethod
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

    @abstractmethod
    def remove_resource_group(self, name: str) -> None:
        """
        Remove resource group.
        Args:
            name: (str): The name of resource group
        Returns:
            None
        """

    @abstractmethod
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

    @abstractmethod
    def set_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        set uint64 counter metrics
        :param data: UInt64CounterData
        :return: None
        """

    @abstractmethod
    def reset_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        reset uint64 counter metrics
        :param data: UInt64CounterData
        :return: None
        """

    @abstractmethod
    def increase_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        increase uint64 counter metrics
        :param data: UInt64CounterData
        :return: None
        """

    @abstractmethod
    def get_value_uint64_counter(self, data: UInt64CounterData) -> int:
        """
        get value of uint64 counter metrics
        :param data: UInt64CounterData
        :return: value
        """

    @abstractmethod
    def set_double_counter(self, data: DoubleCounterData) -> None:
        """
        set double counter metrics
        :param data: DoubleCounterData
        :return: None
        """

    @abstractmethod
    def reset_double_counter(self, data: DoubleCounterData) -> None:
        """
        reset double counter metrics
        :param data: DoubleCounterData
        :return: None
        """

    @abstractmethod
    def increase_double_counter(self, data: DoubleCounterData) -> None:
        """
        increase double counter metrics
        :param data: DoubleCounterData
        :return: None
        """

    @abstractmethod
    def get_value_double_counter(self, data: DoubleCounterData) -> float:
        """
        get value of double counter metrics
        :param data: DoubleCounterData
        :return: value
        """

    @abstractmethod
    def report_gauge(self, data: GaugeData) -> None:
        """
        report gauge metrics
        :param data: GaugeData
        :return: None
        """

    @abstractmethod
    def set_alarm(self, name: str, description: str, info: AlarmInfo) -> None:
        """
        set alarm
        :param name: The name of alarm metrics
        :param description: The description of alarm metrics
        :param info: The alarm info
        : return: None
        """

    @abstractmethod
    def generate_group_name(self) -> str:
        """
        generate group name.
        Returns:
            group name
        """

    @abstractmethod
    def get_instances(self, obj_id, group_name) -> List[str]:
        """
        get instance id list.
        Returns:
            instance id list
        """

    @abstractmethod
    def get_function_group_context(self):
        """
        get function group context.
        Returns:
            FunctionGroupContext
        """

    @abstractmethod
    def is_local_instances(self, instance_ids: List[str]) -> bool:
        """
        is local instances
        Returns:
            bool
        """

    @abstractmethod
    def accelerate(self, group_name: str, handle: Handle):
        """
        accelerate
        Returns:
            handle list
        """

    @abstractmethod
    def create_buffer(self, buffer_size: int):
        """
        create buffer
        Returns:
            obj id and buffer
        """

    @abstractmethod
    def get_buffer(self, obj_id: str):
        """
        get buffer
        Returns:
            buffer
        """

    @abstractmethod
    def add_return_object(self, obj_ids: List[str]):
        """
        add return object
        Returns:
            bool
        """

    @abstractmethod
    def create_group(self, group_name: str, group_opts: GroupOptions):
        """
        create group
        Returns:
            None
        """

    @abstractmethod
    def terminate_group(self, group_name: str):
        """
        terminate group
        Returns:
            None
        """

    @abstractmethod
    def wait_group(self, group_name: str):
        """
        wait group
        Returns:
            None
        """

    @abstractmethod
    def suspend_group(self, group_name: str):
        """
        suspend group
        Returns:
            None
        """

    @abstractmethod
    def resume_group(self, group_name: str):
        """
        resume group
        Returns:
            None
        """
