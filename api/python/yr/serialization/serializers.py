#!/usr/bin/env python
# coding=utf-8
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
"""Serializers"""

import logging
import threading
from dataclasses import dataclass, field
from typing import Any, List
import pickle
import cloudpickle
import msgpack

import yr
from yr import log
from yr.common.utils import NoGC, err_to_str
from yr.fnruntime import (
    Pickle5Writer,
    unpack_pickle5_buffers,
    Pickle5SerializedObject,
    RawSerializedObject,
    SerializedInterface
)
from yr.object_ref import ObjectRef

_EXT_TYPE_CODE = 101
_DEFAULT_PROTOCAL = 4
global_thread_local = threading.local()
global_thread_local.object_refs = set()
_logger = logging.getLogger(__name__)


def pop_local_object_refs():
    """pop local object refs"""
    if not hasattr(global_thread_local, "object_refs"):
        global_thread_local.object_refs = set()
        return set()

    object_refs = global_thread_local.object_refs
    global_thread_local.object_refs = set()
    return object_refs


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class MessagePackSerializedObject:
    """message pack serialized object"""
    msgpack_data: bytes = None
    py_objects: List = field(default_factory=list)


class MessagePackSerializer:
    """
    This class provides methods to serialize and deserialize objects using the MessagePack format.

    Methods:
    - serialize(object: Any) -> Tuple[bytes, List]:
        Serializes object to be the form in bytes using MessagePack format.

    - deserialize(packed_object: Any, py_objects: List) -> Any:
        Deserialized object. The py_objects list is used to deserialize extension types.
    """

    @staticmethod
    def serialize(obj: Any) -> MessagePackSerializedObject:
        """
        Serialize an object using MessagePack format.

        Args:
            obj (Any): The object to be serialized.

        Returns:
            bytes: The serialized form of the object in bytes using MessagePack format.
        """

        py_objects = []

        def _default(python_object):
            index = len(py_objects)
            py_objects.append(python_object)
            return msgpack.ExtType(_EXT_TYPE_CODE, msgpack.dumps(index))

        try:
            msgpack_data = msgpack.dumps(obj, default=_default, use_bin_type=True, strict_types=True)
        except ValueError as err:
            _logger.warning("Failed to uses msgpack.dumps object %s, dumps its index instead. %s", type(obj),
                                     err_to_str(err))
            msgpack_data = msgpack.dumps(_default(obj), default=_default, use_bin_type=True, strict_types=True)

        return MessagePackSerializedObject(msgpack_data, py_objects)

    @staticmethod
    def deserialize(serialized_data: bytes, py_objects: List) -> Any:
        """
        Deserialize a packed object using MessagePack format.

        Args:
            serialized_data (Any): The serialized data in bytes to be deserialized.
            py_objects (List): A list of python objects used to deserialize extension types.

        Returns:
            Any: The deserialized object.
        """

        def _ext_hook(code, data):
            if code == _EXT_TYPE_CODE:
                length = len(py_objects)
                index = msgpack.loads(data)
                if index >= length:
                    raise RuntimeError("Length (%s) of 'py_objects' provided is too small to hook the "
                                       "extension type of index: %s", length, index)
                return py_objects[msgpack.loads(data)]

        with NoGC():  # optimizes the msgpack performance.
            return msgpack.loads(serialized_data, ext_hook=_ext_hook, raw=False, strict_map_key=False)


def _object_ref_deserializer(object_id):
    object_ref = ObjectRef(object_id=object_id, need_incre=False)
    global_thread_local.object_refs.add(object_ref)
    return object_ref


def object_ref_reducer(object_ref: ObjectRef):
    """object ref reducer"""
    global_thread_local.object_refs.add(object_ref)
    return _object_ref_deserializer, (object_ref.id,)


class PySerializer:
    """py serializer"""

    @staticmethod
    def init():
        """init"""
        cloudpickle.CloudPickler.dispatch[ObjectRef] = object_ref_reducer

    @staticmethod
    def serialize(value: Any) -> SerializedInterface:
        """serialize"""
        global_thread_local.object_refs = set()
        try:
            if pickle.HIGHEST_PROTOCOL >= 5:
                return PySerializer._serialize_to_pickle5(value)
            return PySerializer._serialize_to_pickle4(value)
        except Exception as err:
            _logger.error("Failed to serialize value with PickleSerializer. %s", err_to_str(err))
            pop_local_object_refs()
            raise err

    @staticmethod
    def deserialize(serialized_data: bytes) -> Any:
        """deserialize"""
        global_thread_local.object_refs = set()
        if pickle.HIGHEST_PROTOCOL >= 5:
            obj = PySerializer._deserialize_to_pickle5(serialized_data)
        else:
            obj = PySerializer._deserialize_to_pickle4(serialized_data)
        nested_refs = pop_local_object_refs()
        if len(nested_refs) != 0:
            yr.runtime_holder.global_runtime.get_runtime().increase_global_reference([ref.id for ref in nested_refs])
        return obj

    @staticmethod
    def _serialize_to_pickle5(value: Any):
        """serialize to pickle5"""
        writer = Pickle5Writer()
        inband = cloudpickle.dumps(value, protocol=5, buffer_callback=writer.buffer_callback)
        nested_refs = pop_local_object_refs()
        return Pickle5SerializedObject(inband, writer, nested_refs)

    @staticmethod
    def _serialize_to_pickle4(value: Any):
        result = cloudpickle.dumps(value, protocol=4)
        nested_refs = pop_local_object_refs()
        return RawSerializedObject(result, nested_refs)

    @staticmethod
    def _deserialize_to_pickle5(serialized_data: bytes) -> Any:
        """deserialize"""
        try:
            in_band, buffers = unpack_pickle5_buffers(serialized_data)
        except ValueError as err:
            _logger.error("Failed to unpack pickle5 buffers. %s", err_to_str(err))
            raise err
        try:
            if len(buffers) > 0:
                return cloudpickle.loads(in_band, buffers=buffers)
            return cloudpickle.loads(in_band)
        except cloudpickle.pickle.PicklingError as err:
            _logger.error("Failed to load inband data. %s", err_to_str(err))
            raise err
        except EOFError as e:
            raise ValueError("EOFError") from e

    @staticmethod
    def _deserialize_to_pickle4(serialized_data: bytes) -> Any:
        """deserialize"""
        try:
            return cloudpickle.loads(serialized_data)
        except cloudpickle.pickle.PicklingError as err:
            _logger.error("Failed to load serialized data. %s", err_to_str(err))
            raise err
        except EOFError as e:
            raise ValueError("EOFError") from e
