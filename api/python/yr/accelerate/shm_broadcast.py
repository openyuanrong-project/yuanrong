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
import asyncio
import logging
import threading
import time
import os
import sys
import pickle
from contextlib import contextmanager, asynccontextmanager
from typing import Optional
from enum import IntEnum
from dataclasses import dataclass
import yr

_logger = logging.getLogger(__name__)

STOP_EVENT = threading.Event()
USE_SCHED_YIELD = ((sys.version_info[:3] >= (3, 11, 1))
                   or (sys.version_info[:2] == (3, 10)
                       and sys.version_info[2] >= 8))

FCC_RINGBUFFER_WARNING_INTERVAL = int(os.environ.get("FCC_RINGBUFFER_WARNING_INTERVAL", "60"))


class ResponseStatus(IntEnum):
    SUCCESS = 0
    FAILURE = 1


def sched_yield():
    """yield"""
    if USE_SCHED_YIELD:
        os.sched_yield()
    else:
        time.sleep(0)


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class Handle:
    n_reader: int
    max_chunk_bytes: int
    max_chunks: int
    name: str
    rank: int
    use_async_loop: bool


class ShmRingBuffer:
    def __init__(self, n_reader: int, max_chunk_bytes: int, max_chunks: int, name: Optional[str] = None):
        self.n_reader = n_reader
        self.metadata_size = 1 + n_reader
        self.max_chunk_bytes = max_chunk_bytes
        self.max_chunks = max_chunks
        self.total_bytes_of_buffer = (self.max_chunk_bytes + self.metadata_size) * self.max_chunks
        self.data_offset = 0
        self.metadata_offset = self.max_chunk_bytes * self.max_chunks
        if name is None:
            obj_id, shared_memory = yr.runtime_holder.global_runtime.get_runtime().create_buffer(
                self.total_bytes_of_buffer)
            self.name = obj_id
            self.shared_memory = shared_memory
            self.shared_memory.get_buf()[:] = bytes(len(self.shared_memory.get_buf()))
        else:
            self.name = name
            self.shared_memory = yr.runtime_holder.global_runtime.get_runtime().get_buffer(name)

    def handle(self):
        """message queue handle"""
        return self.n_reader, self.max_chunk_bytes, self.max_chunks, self.name

    @contextmanager
    def get_metadata(self, current_idx: int):
        """get metadata"""
        start = self.metadata_offset + current_idx * self.metadata_size
        end = start + self.metadata_size
        with memoryview(self.shared_memory.get_buf()[start:end]) as buf:
            yield buf

    @contextmanager
    def get_data(self, current_idx: int):
        """get data"""
        start = self.data_offset + current_idx * self.max_chunk_bytes
        end = start + self.max_chunk_bytes
        with memoryview(self.shared_memory.get_buf()[start:end]) as buf:
            yield buf


class DataType(IntEnum):
    PICKLE = 255  # max value of uint8


def get_datatype(obj):
    """get datatype"""
    return DataType.PICKLE


def encode(obj, data_type: DataType = DataType.PICKLE):
    """encode obj"""
    if data_type == DataType.PICKLE:
        return pickle.dumps(obj, protocol=pickle.HIGHEST_PROTOCOL)
    raise NotImplementedError


def decode(obj, data_type: DataType = DataType.PICKLE):
    """decode obj"""
    if data_type == DataType.PICKLE:
        return pickle.loads(obj)
    raise NotImplementedError


class MessageQueue:
    def __init__(self, n_reader: int, max_chunk_bytes: int = 1024 * 1024 * 10, max_chunks: int = 10,
                 use_async_loop: bool = False):
        self.n_reader = n_reader
        self.buffer = ShmRingBuffer(n_reader, max_chunk_bytes, max_chunks)
        self.use_async_loop = use_async_loop
        self.is_writer = True
        self.is_local_reader = False
        self.local_reader_rank = -1
        self.handle = Handle(*self.buffer.handle(), self.local_reader_rank, self.use_async_loop)
        self.current_idx = 0

    @staticmethod
    def create_from_handle(handle: Handle) -> 'MessageQueue':
        """create message queue from handle"""
        self = MessageQueue.__new__(MessageQueue)
        self.handle = handle
        self.is_writer = False
        self.buffer = ShmRingBuffer(handle.n_reader, handle.max_chunk_bytes, handle.max_chunks, handle.name)
        self.current_idx = 0
        self.is_local_reader = True
        self.local_reader_rank = handle.rank
        self.use_async_loop = handle.use_async_loop
        return self

    def export_handle(self) -> Handle:
        """export handle"""
        return self.handle

    def dequeue(self, timeout: Optional[float] = None):
        """ Read from message queue with optional timeout (in seconds) """
        if self.is_local_reader:
            with self.acquire_read(timeout) as buf:
                overflow = buf[0] == 1
                data_type = DataType(buf[1])
                if not overflow:
                    obj = decode(buf[2:], data_type)
        else:
            raise RuntimeError("Only readers can dequeue")
        return obj

    def enqueue(self, obj, timeout: Optional[float] = None):
        """ Write to message queue with optional timeout (in seconds) """
        if not self.is_writer:
            raise RuntimeError("Only writers can enqueue")

        data_type = get_datatype(obj)
        serialized_obj = encode(obj, data_type)
        with self.acquire_write(timeout) as buf:
            buf[0] = 0  # not overflow
            buf[1] = data_type
            buf[2:len(serialized_obj) + 2] = serialized_obj

    @contextmanager
    def acquire_write(self, timeout: Optional[float] = None):
        """acquire write"""
        if not self.is_writer:
            raise RuntimeError("Only writers can acquire write")

        start_time = time.monotonic()
        n_warning = 1
        while True:
            with self.buffer.get_metadata(self.current_idx) as metadata_buffer:
                read_count = sum(metadata_buffer[1:])
                written_flag = metadata_buffer[0]
                if written_flag and read_count != self.buffer.n_reader:
                    sched_yield()
                    if (time.monotonic() - start_time >
                            FCC_RINGBUFFER_WARNING_INTERVAL * n_warning):
                        _logger.debug(f"wait for a long time, log a message, n_warning: {n_warning}")
                        n_warning += 1
                    if (timeout is not None
                            and time.monotonic() - start_time > timeout):
                        raise TimeoutError

                    continue
                metadata_buffer[0] = 0
                with self.buffer.get_data(self.current_idx) as buf:
                    yield buf

                for i in range(1, self.buffer.n_reader + 1):
                    metadata_buffer[i] = 0
                metadata_buffer[0] = 1
                self.current_idx = (self.current_idx +
                                    1) % self.buffer.max_chunks
                break

    @contextmanager
    def acquire_read(self, timeout: Optional[float] = None):
        """acquire read"""
        if not self.is_local_reader:
            raise RuntimeError("Only readers can acquire read")

        start_time = time.monotonic()
        n_warning = 1
        while True:
            with self.buffer.get_metadata(self.current_idx) as metadata_buffer:
                read_flag = metadata_buffer[self.local_reader_rank + 1]
                written_flag = metadata_buffer[0]
                if not written_flag or read_flag:
                    if STOP_EVENT.is_set():
                        raise RuntimeError("event is stop")
                    sched_yield()

                    if (timeout is not None
                            and time.monotonic() - start_time > timeout):
                        raise TimeoutError

                    if (time.monotonic() - start_time >
                            FCC_RINGBUFFER_WARNING_INTERVAL * n_warning):
                        _logger.debug(f"wait for a long time, log a message, n_warning: {n_warning}")
                        n_warning += 1

                    continue

                with self.buffer.get_data(self.current_idx) as buf:
                    yield buf

                metadata_buffer[self.local_reader_rank + 1] = 1
                self.current_idx = (self.current_idx +
                                    1) % self.buffer.max_chunks
                break

    async def dequeue_async(self, timeout: Optional[float] = None):
        """dequeue async"""
        if self.is_local_reader:
            async with self.acquire_read_async(timeout) as buf:
                overflow = buf[0] == 1
                data_type = DataType(buf[1])
                if not overflow:
                    obj = decode(buf[2:], data_type)
        return obj

    @asynccontextmanager
    async def acquire_read_async(self, timeout: Optional[float] = None):
        """acquire read async"""
        if not self.is_local_reader:
            raise RuntimeError("Only readers can acquire read async")

        start_time = time.monotonic()
        n_warning = 1
        while True:
            with self.buffer.get_metadata(self.current_idx) as metadata_buffer:
                read_flag = metadata_buffer[self.local_reader_rank + 1]
                written_flag = metadata_buffer[0]
                if not written_flag or read_flag:
                    if STOP_EVENT.is_set():
                        raise RuntimeError("acquire read event is stop")
                    await asyncio.sleep(0.0001)

                    if (time.monotonic() - start_time >
                            FCC_RINGBUFFER_WARNING_INTERVAL * n_warning):
                        _logger.debug(f"wait for a long time, log a message, n_warning: {n_warning}")
                        n_warning += 1

                    if (timeout is not None
                            and time.monotonic() - start_time > timeout):
                        raise TimeoutError

                    continue
                with self.buffer.get_data(self.current_idx) as buf:
                    yield buf
                metadata_buffer[self.local_reader_rank + 1] = 1
                self.current_idx = (self.current_idx +
                                    1) % self.buffer.max_chunks
                break

    async def enqueue_async(self, obj, timeout: Optional[float] = None):
        """enqueue async"""
        data_type = get_datatype(obj)
        serialized_obj = encode(obj, data_type)
        async with self.acquire_write_async(timeout) as buf:
            buf[0] = 0  # not overflow
            buf[1] = data_type
            buf[2:len(serialized_obj) + 2] = serialized_obj

    @asynccontextmanager
    async def acquire_write_async(self, timeout: Optional[float] = None):
        """acquire write async"""
        if not self.is_writer:
            raise RuntimeError("Only writers can acquire write async")

        start_time = time.monotonic()
        n_warning = 1
        while True:
            with self.buffer.get_metadata(self.current_idx) as metadata_buf:
                read_count = sum(metadata_buf[1:])
                written_flag = metadata_buf[0]
                if written_flag and read_count != self.buffer.n_reader:
                    if (time.monotonic() - start_time >
                            FCC_RINGBUFFER_WARNING_INTERVAL * n_warning):
                        _logger.debug(f"wait for a long time, log a message, n_warning: {n_warning}")
                        n_warning += 1
                    if (timeout is not None
                            and time.monotonic() - start_time > timeout):
                        raise TimeoutError
                    await asyncio.sleep(0.0001)
                    continue

                metadata_buf[0] = 0
                with self.buffer.get_data(self.current_idx) as buf:
                    yield buf

                for i in range(1, self.buffer.n_reader + 1):
                    metadata_buf[i] = 0
                metadata_buf[0] = 1
                self.current_idx = (self.current_idx +
                                    1) % self.buffer.max_chunks
                break
