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

"""Serialization"""
import threading
from typing import Any, List, Union

import yr
from yr.common import constants
from yr.common.singleton import Singleton
from yr.fnruntime import SerializedObject, split_buffer, Buffer
from yr.serialization.serializers import MessagePackSerializer, PySerializer, pop_local_object_refs


@Singleton
class Serialization:
    """serialization context"""
    _lock = threading.Lock()
    _msgpack_serialize_hooks = []
    _msgpack_deserialize_hook = []

    def __init__(self):
        PySerializer.init()
        self._protocol = 5

    @staticmethod
    def normalize_input(value: Any):
        """
        Normalize bytes, memoryview, bytearray to memoryview to avoid copying and attach meta-information
        """
        if isinstance(value, bytes):
            return constants.Metadata.BYTES, value
        if isinstance(value, memoryview):
            return constants.Metadata.MEMORYVIEW, value
        if isinstance(value, bytearray):
            return constants.Metadata.BYTEARRAY, value
        raise TypeError(f"Unsupported input type: {type(value)}")

    def serialize(self, value: Any) -> SerializedObject:
        """serialize"""
        metadata = constants.Metadata.CROSS_LANGUAGE
        py_serialized_object = None
        if isinstance(value, (bytes, memoryview, bytearray)):
            metadata, msgpack_data = self.normalize_input(value)
        else:
            msgpack_serialized_object = MessagePackSerializer.serialize(value)
            msgpack_data = msgpack_serialized_object.msgpack_data
            py_objects = msgpack_serialized_object.py_objects
            if py_objects:
                metadata = constants.Metadata.PYTHON
                py_serialized_object = PySerializer.serialize(py_objects)
        if not isinstance(msgpack_data, bytes):
            msgpack_data = bytes(msgpack_data)
        return SerializedObject(
            metadata=metadata.value,
            msgpack_data=msgpack_data,
            py_serialized_object=py_serialized_object
        )

    def deserialize(self, buffers: Union[None, Buffer, List[Buffer]]) -> Union[Any, List]:
        """Deserializes bytes data. The bytes data should be with the
        '|metadata_header|metadata|msgpack_header|msgpack_data|py_serialized_object|'
        structure. Otherwise, cloudpickles is directly used for deserialization.

        Args:
            buffers (Union[None, Buffer, memoryview, List[Buffer]]): Bytes data to  be deserialized.

        Returns:
            Union[Any, List]: The deserialized result.
        """
        is_buffer = isinstance(buffers, Buffer)

        if is_buffer:
            buffers = [buffers]
        pop_local_object_refs()
        result = []
        # Deserializing user code may execute 'import' operation, which is not thread safe.
        # Therefore, a lock needs to be added.
        with self._lock:
            for buf in buffers:
                if buf is None:
                    result.append(None)
                    continue
                metadata, msgpack_data, py_serialized_data = split_buffer(buf)
                if constants.Metadata(metadata) in [
                    constants.Metadata.BYTES,
                    constants.Metadata.BYTEARRAY,
                    constants.Metadata.MEMORYVIEW
                ]:
                    result.append(memoryview(msgpack_data))
                    continue
                python_objects = []
                if constants.Metadata(metadata) == constants.Metadata.PYTHON:
                    python_objects = PySerializer.deserialize(py_serialized_data)
                result.append(MessagePackSerializer.deserialize(msgpack_data, python_objects))
        object_refs = pop_local_object_refs()
        if len(object_refs) != 0:
            object_ref_ids = [ref.id for ref in object_refs]
            yr.runtime_holder.global_runtime.get_runtime().increase_global_reference(object_ref_ids)
        return result[0] if is_buffer else result

    def register_pack_hook(self, hook):
        """register pack hook to Serialization"""
        self._msgpack_serialize_hooks.append(hook)

    def register_unpack_hook(self, hook):
        """register unpack hook to Serialization"""
        self._msgpack_deserialize_hook.append(hook)

    def _pack(self, obj):
        for hook in self._msgpack_serialize_hooks:
            if callable(hook):
                obj = hook(obj)
        return obj

    def _unpack(self, obj):
        for hook in self._msgpack_deserialize_hook:
            if callable(hook):
                obj = hook(obj)
        return obj


def register_pack_hook(hook):
    """register pack hook helper"""
    Serialization().register_pack_hook(hook)


def register_unpack_hook(hook):
    """register unpack hook helper"""
    Serialization().register_unpack_hook(hook)
