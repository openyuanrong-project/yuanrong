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
import logging
import uuid
import inspect
import traceback
import threading
import asyncio
from typing import Dict, Tuple, Optional
from yr import log
from yr.accelerate.shm_broadcast import MessageQueue, ResponseStatus, STOP_EVENT
from yr.executor.instance_manager import InstanceManager

_logger = logging.getLogger(__name__)

ACCELERATE_WORKER = None


class Worker:
    def __init__(self, rpc_broadcast_mq: MessageQueue, worker_response_mq: MessageQueue, use_async_loop: bool):
        self.rpc_broadcast_mq = rpc_broadcast_mq
        self.worker_response_mq = worker_response_mq
        self.use_async_loop = use_async_loop
        self.with_unfinished_tasks_event = None
        self.tasks = None
        self._input_loop_unshielded = None
        self.input_loop = None
        self.thread = None

    @staticmethod
    def execute_method(method: str, *args, **kwargs):
        """execute member function"""
        try:
            instance = InstanceManager().instance()
            executor = getattr(instance, method)
            return executor(*args, **kwargs)
        except Exception as e:
            _logger.debug(f"Error executing method {method}.")
            raise e

    @staticmethod
    async def execute_method_wrapper(
            task_id: str,
            method: str,
            args: Tuple = (),
            kwargs: Optional[Dict] = None):
        """execute async member function wrapper"""
        kwargs = kwargs or {}
        try:
            output = await Worker.execute_method_async(method, *args, **kwargs)
        except Exception:
            exc_output = traceback.format_exc()
            return ResponseStatus.FAILURE, (task_id, exc_output)
        return ResponseStatus.SUCCESS, (task_id, output)

    @staticmethod
    async def execute_method_async(method: str, *args, **kwargs):
        """execute async member function"""
        try:
            instance = InstanceManager().instance()
            executor = getattr(instance, method)
            awaitable = inspect.iscoroutinefunction(executor)
            if awaitable:
                return await executor(*args, **kwargs)
            return executor(*args, **kwargs)
        except Exception as e:
            msg = (f"Error executing method {method}. "
                   f"This might cause deadlock in distributed execution. e: {e}")
            _logger.debug(msg)
            raise e

    def worker_busy_loop_sync(self):
        """Sync busy loop for Worker"""
        _logger.debug(f"start to run loop sync")
        while True:
            if STOP_EVENT.is_set():
                break
            task_id, method, args, kwargs = self.rpc_broadcast_mq.dequeue()
            _logger.debug(f"start to execute member function: {method}, task id: {task_id}")
            try:
                output = self.execute_method(method, *args, **kwargs)
            except Exception as e:
                exc_output = traceback.format_exc()
                _logger.debug(
                    f"failed to execute member function: {method}, task id: {task_id}, err: {str(e)}")
                self.worker_response_mq.enqueue((ResponseStatus.FAILURE, (task_id, exc_output)))
                continue
            _logger.debug(f"success to execute member function: {method}, task id: {task_id}")
            self.worker_response_mq.enqueue((ResponseStatus.SUCCESS, (task_id, output)))

    def start(self):
        """start worker_busy_loop"""
        self.thread = threading.Thread(target=self.worker_busy_loop)
        self.thread.start()

    def worker_busy_loop(self):
        """Main busy loop for Worker"""
        if not self.use_async_loop:
            self.worker_busy_loop_sync()
        else:
            asyncio.run(self.worker_busy_loop_async())

    async def create_task_from_input_mq(self):
        """create task from input queue"""
        while True:
            task_id, method, args, kwargs = await (self.rpc_broadcast_mq.
                                                   dequeue_async())
            _logger.debug(f"start to execute member function: {method}, task id: {task_id}")
            self.tasks[task_id] = asyncio.create_task(
                self.execute_method_wrapper(task_id, method,
                                            args, kwargs))
            self.with_unfinished_tasks_event.set()
            await asyncio.sleep(0)

    async def worker_busy_loop_async(self):
        """Async busy loop for Worker"""
        _logger.debug(f"start to run loop async")
        self.with_unfinished_tasks_event = asyncio.Event()
        self.tasks: Dict[uuid.UUID, asyncio.Task] = {}
        self._input_loop_unshielded = asyncio.create_task(self.create_task_from_input_mq())
        self.input_loop = asyncio.shield(self._input_loop_unshielded)
        while True:
            if not self.tasks:
                self.with_unfinished_tasks_event.clear()
                await self.with_unfinished_tasks_event.wait()
            if STOP_EVENT.is_set():
                break
            done, _ = await asyncio.wait(self.tasks.values(), return_when=asyncio.FIRST_COMPLETED)
            for done_task in done:
                _, (task_id, _) = done_task.result()
                _logger.debug(f"finish to execute task, task id: {task_id}")
                del self.tasks[task_id]
            for done_task in done:
                await self.worker_response_mq.enqueue_async(done_task.result())
            await asyncio.sleep(0)
