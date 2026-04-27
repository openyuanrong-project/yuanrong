# !/usr/bin/env python3
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

import unittest
import json
import logging
import time
import asyncio
from dataclasses import asdict
from yr.decorator.instance_proxy import FunctionGroupHandler
from unittest import TestCase, main
from unittest.mock import patch, Mock, AsyncMock
from yr.fnruntime import SharedBuffer
from yr.accelerate.shm_broadcast import Handle
from yr.accelerate.executor import Worker
from yr.executor.instance_manager import InstanceManager
from yr.accelerate.shm_broadcast import STOP_EVENT, MessageQueue
import yr
import threading
logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.DEBUG)


class FccTestInstance:
    def __init__(self, name: str):
        self.name = name

    def get(self, a, b):
        return self.name


def new_instance():
    return FccTestInstance("fcc")


def get_new_logger():
    return logger


class MockSharedMemory:

    def __init__(self, data):
        self.buf = memoryview(data)

    def get_buf(self):
        return self.buf


class TestFcc(unittest.TestCase):
    def setUp(self) -> None:
        pass

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_accelerate(self, get_runtime):
        data = bytearray(10 * 1024 * 1024)
        shared_buffer = SharedBuffer()
        shared_buffer.set_buf(memoryview(data))

        handle = Handle(1, 1024 * 1024 * 10, 10, "obj", 1, True)
        handle_str = json.dumps(asdict(handle))
        mock_runtime = Mock()
        mock_runtime.is_local_instances.return_value = True
        mock_runtime.create_buffer.return_value = ("obj", shared_buffer)
        mock_runtime.get_buffer.return_value = shared_buffer
        mock_runtime.accelerate.return_value = [handle_str]
        mock_runtime.terminate.return_value = None
        get_runtime.return_value = mock_runtime

        handler = FunctionGroupHandler(["instance1"], None, None, None, None)
        handler.accelerate()
        self.assertIsNone(handler.terminate())

    def test_worker_sync(self):
        yr.log.get_logger = get_new_logger
        InstanceManager().instance = new_instance
        mock_queue = Mock()
        mock_queue.dequeue.return_value = ("obj", "get", (1, 2), {})
        worker = Worker(mock_queue, mock_queue, False)
        self.assertIsNone(worker.start())
        asyncio.run(asyncio.sleep(0.2))
        STOP_EVENT.set()
        asyncio.run(asyncio.sleep(0.2))
        STOP_EVENT.clear()

        def stop():
            asyncio.run(asyncio.sleep(0.5))
            STOP_EVENT.set()
        thread = threading.Thread(target=stop)
        thread.start()
        self.assertIsNone(worker.worker_busy_loop_sync())

    def test_worker_async(self):
        yr.log.get_logger = get_new_logger
        mock_queue = AsyncMock()
        InstanceManager().instance = new_instance
        mock_queue.dequeue_async.return_value = ("obj", "get", (1, 2), {})
        worker = Worker(mock_queue, mock_queue, True)
        self.assertIsNone(worker.start())
        asyncio.run(asyncio.sleep(0.2))
        STOP_EVENT.set()
        asyncio.run(asyncio.sleep(0.2))
        STOP_EVENT.clear()

        def stop():
            asyncio.run(asyncio.sleep(0.5))
            STOP_EVENT.set()
        thread = threading.Thread(target=stop)
        thread.start()
        self.assertIsNone(asyncio.run(worker.worker_busy_loop_async()))

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_message_queue_sync(self, get_runtime):
        STOP_EVENT.clear()
        yr.log.get_logger = get_new_logger
        logger.info("==========================")
        data = bytearray((10 * 1024 * 1024 + 2) * 10)
        shared_buffer = MockSharedMemory(data)
        logger.info("==========================")
        logger.debug("========")
        logger.debug(len(shared_buffer.get_buf()))
        logger.debug(len(data))
        mock_runtime = Mock()
        mock_runtime.create_buffer.return_value = ("obj", shared_buffer)
        mock_runtime.get_buffer.return_value = shared_buffer
        get_runtime.return_value = mock_runtime
        queue = MessageQueue(1)
        queue.enqueue(("obj", "get", (1, 2), {}))
        handle = queue.export_handle()
        handle.rank = 0
        r_queue = MessageQueue.create_from_handle(handle)
        self.assertIsNotNone(r_queue.dequeue())

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_message_queue_async(self, get_runtime):
        STOP_EVENT.clear()
        yr.log.get_logger = get_new_logger
        data = bytearray((10 * 1024 * 1024 + 2) * 10)
        shared_buffer = MockSharedMemory(data)
        mock_runtime = Mock()
        mock_runtime.create_buffer.return_value = ("obj", shared_buffer)
        mock_runtime.get_buffer.return_value = shared_buffer
        get_runtime.return_value = mock_runtime
        queue = MessageQueue(1)
        asyncio.run(queue.enqueue_async(("obj", "get", (1, 2), {})))
        handle = queue.export_handle()
        handle.rank = 0
        r_queue = MessageQueue.create_from_handle(handle)
        self.assertIsNotNone((asyncio.run(r_queue.dequeue_async())))


if __name__ == "__main__":
    unittest.main()
