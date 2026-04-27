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

"""task manager"""
import functools
import logging
from collections import deque
from dataclasses import dataclass
from enum import Enum
from threading import RLock, Lock, BoundedSemaphore, Thread
from typing import List
from typing import Optional

import yr.apis
from yr.exception import CancelError

from yr.config import InvokeOptions
from yr.config_manager import ConfigManager
from yr.libruntime_pb2 import InvokeType
from yr.local_mode.instance import Resource, Instance
from yr.local_mode.instance_manager import InstanceManager
from yr.local_mode.scheduler import NormalScheduler, ConcurrencyScorer
from yr.local_mode.task_spec import TaskSpec
from yr.local_mode.timer import Timer


class TaskState(Enum):
    """task state"""
    PENDING = 1
    RUNNING = 2
    CANCELED = 3


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class TaskRecord:
    """record for task"""
    state: TaskState
    resource: Resource = None
    task: TaskSpec = None


_logger = logging.getLogger(__name__)


class TaskManager:
    """Task manager"""
    __slots__ = ["__invoke_client", "__queue", "__ins_mgr", "__pending_tasks",
                 "__timer", "__sleep_time_sequence", "__queue_lock", "__task_lock", "__schedule_semaphore",
                 "__schedule_thread"]

    def __init__(self, invoke_client):
        # task_id/future
        self.__invoke_client = invoke_client
        self.__queue = deque()
        self.__pending_tasks = dict()
        scheduler = NormalScheduler(scorers=[ConcurrencyScorer()])
        self.__ins_mgr = InstanceManager(scheduler, invoke_client, ConfigManager().meta_config.recycleTime)
        self.__timer = Timer()
        # delay to schedule 0,1,2,5,10,30,60,60
        self.__sleep_time_sequence = [0, 1, 2, 5, 10, 30, 60]
        self.__queue_lock = RLock()
        self.__task_lock = Lock()
        self.__schedule_semaphore = BoundedSemaphore(1)
        self.__schedule_thread = Thread(target=self.__schedule_task, name="streaming", daemon=True)
        self.__schedule_thread.start()

    def submit_task(self, task: TaskSpec):
        """submit task"""
        if not yr.apis.is_initialized():
            _logger.warning("Can not submit task %s before yr.init", task.task_id)
            return
        if task.invoke_options:
            invoke_options = task.invoke_options
        else:
            invoke_options = InvokeOptions()
        resource = Resource(cpu=invoke_options.cpu,
                            memory=invoke_options.memory,
                            concurrency=invoke_options.concurrency,
                            resources=invoke_options.custom_resources)

        if task.invoke_type == InvokeType.InvokeFunctionStateless:
            task_record = TaskRecord(state=TaskState.PENDING, resource=resource, task=task)
            if self.__add_task(task_record):
                self.__add_task_id(task.task_id)
                self.__schedule()
            elif self.__is_canceled(task.task_id):
                task.future.set_exception(CancelError(task.task_id))
            else:
                _logger.warning("task already in schedule, %s", task.task_id)
        else:
            task.future.set_exception(TypeError(f"unexpect invoke type {task.invoke_type}"))

    def cancel(self, task_id: str) -> None:
        """cancel a task"""
        _logger.info("task canceled: %s", task_id)
        state = self.__set_canceled(task_id)
        if state == TaskState.CANCELED:
            return
        try:
            self.__queue.remove(task_id)
        except ValueError:
            _logger.debug("task (%s) has been popped out from queue.", task_id)
        task_record = self.__get_task(task_id)
        task_record.task.future.set_exception(CancelError(task_id))
        if state == TaskState.RUNNING:
            self.__cancel_task(task_record.task.instance_id)

    def clear(self):
        """clear tasks and instances"""
        self.__queue.clear()
        self.__pending_tasks.clear()
        self.__ins_mgr.clear()

    def __schedule_task(self):
        while self.__schedule_semaphore.acquire():
            task_id = self.__pop_task_id()
            if task_id == "":
                continue
            task_record = self.__get_task(task_id)
            if not task_record:
                continue
            ret, err = self.__ins_mgr.check_last_failed_reason(resource=task_record.resource)
            if not ret:
                task_record.task.future.set_exception(err)
                continue
            _logger.debug("schedule task %s", task_record.task.task_id)
            ins = self.__ins_mgr.schedule(task_record.task, task_record.resource)
            if not ins:
                if task_record.state == TaskState.PENDING:
                    self.__add_task_id(task_id, True)
                tasks_of_resource = list(filter(lambda record: record.resource == task_record.resource,
                                                self.__get_all_tasks()))
                instances_of_resource = list(self.__ins_mgr.get_instances(task_record.resource))
                if scale_out(tasks_of_resource, instances_of_resource):
                    self.__scale_out_after(self.__get_scale_out_delay(task_record.resource), task_record)
            else:
                _logger.debug("schedule successfully %s %s", task_record.task.task_id, ins.instance_id)
                task_record.state = TaskState.RUNNING
                task_record.task.instance_id = ins.instance_id
                self.__invoke_client.invoke(task_record.task)

                def callback(task_id, ins, _):
                    _logger.debug("Invoke successfully. task_id: %s", task_record.task.task_id)
                    ins.delete_task(task_id)
                    self.__schedule()

                # Callback param should be captured and fixed. Otherwise when callback is executed
                # after new 'task_id' is popped by '__pop_task_id()', 'task_id' in callback would be
                # replaced by the new 'task_id' just popped.
                callback_param_captured = functools.partial(callback, task_id, ins)
                task_record.task.future.add_done_callback(callback_param_captured)
                self.__schedule()

    def __schedule(self) -> None:
        try:
            self.__schedule_semaphore.release()
        except ValueError:
            return

    def __scale_out_after(self, sleep_time: int, task_record: TaskRecord) -> None:
        def scale_out_inner():
            _, future = self.__ins_mgr.scale_out(task_record.task, task_record.resource)
            future.add_done_callback(lambda x: self.__schedule())

        if sleep_time == 0:
            scale_out_inner()
        else:
            self.__timer.after(sleep_time, scale_out_inner)

    def __get_scale_out_delay(self, resource: Resource) -> int:
        count = self.__ins_mgr.get_failed_count(resource)
        if count >= len(self.__sleep_time_sequence):
            return self.__sleep_time_sequence[-1]
        return self.__sleep_time_sequence[count]

    def __cancel_task(self, instance_id: str):
        self.__invoke_client.kill(instance_id)

    def __add_task(self, task_record: TaskRecord) -> bool:
        with self.__task_lock:
            if task_record.task.task_id in self.__pending_tasks:
                return False
            self.__pending_tasks[task_record.task.task_id] = task_record
            return True

    def __get_task(self, task_id: str) -> Optional[TaskRecord]:
        with self.__task_lock:
            return self.__pending_tasks.get(task_id, None)

    def __pop_task(self, task_id: str):
        with self.__task_lock:
            return self.__pending_tasks.pop(task_id, None)

    def __get_all_tasks(self) -> List[TaskRecord]:
        with self.__task_lock:
            return list(self.__pending_tasks.values())

    def __add_task_id(self, task_id: str, left: bool = False):
        with self.__queue_lock:
            if left:
                self.__queue.appendleft(task_id)
            else:
                self.__queue.append(task_id)

    def __pop_task_id(self) -> str:
        with self.__queue_lock:
            if len(self.__queue) != 0:
                return self.__queue.pop()
        return ""

    def __is_canceled(self, task_id: str):
        with self.__task_lock:
            task_record = self.__pending_tasks.get(task_id, None)
            if not task_record:
                return False
            if task_record.state == TaskState.CANCELED:
                return True
            return False

    def __set_canceled(self, task_id: str) -> TaskState:
        with self.__task_lock:
            task_record = self.__pending_tasks.get(task_id, None)
            if not task_record:
                self.__pending_tasks[task_id] = TaskRecord(TaskState.CANCELED)
                return TaskState.CANCELED
            state = task_record.state
            task_record.state = TaskState.CANCELED
            return state


def scale_out(tasks: List[TaskRecord], instances: List[Instance]):
    """judge whether scaling out"""
    task_sum = len(list(tasks))
    concurrency_sum = sum([ins.resource.concurrency for ins in instances])
    _logger.debug("start to judge scale out, invoke task count: %s concurrency sum: %s",
                  task_sum, concurrency_sum)
    return task_sum > concurrency_sum
