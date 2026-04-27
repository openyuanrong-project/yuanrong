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

"""local handler"""
import logging
import threading

from yr.common.utils import generate_random_id
from yr.local_mode.local_object_store import LocalObjectStore
from yr.local_mode.task_spec import TaskSpec
from yr.local_mode.worker import Worker

_logger = logging.getLogger(__name__)


class LocalClient:
    """local handler"""

    def __init__(self):
        self.workers = {}
        self._running = True
        self._lock = threading.Lock()

    def create(self, task: TaskSpec) -> str:
        """create worker"""
        if not self._running:
            return ""
        instance_id = task.object_ids[0]
        concurrency = 100 if task.invoke_options is None else task.invoke_options.concurrency
        worker = Worker()
        worker.init(instance_id, concurrency)
        with self._lock:
            self.workers[instance_id] = worker
        worker.submit(task)
        return instance_id

    def invoke(self, task: TaskSpec):
        """instance function invoke"""
        if not self._running:
            return
        worker = self.workers.get(task.instance_id)
        if worker is None:
            raise RuntimeError(f"No such instance: {task.instance_id}")
        worker.submit(task)

    def kill(self, instance_id: str):
        """kill worker"""
        with self._lock:
            if instance_id in self.workers:
                worker = self.workers.pop(instance_id)
                worker.stop()

    def exit(self) -> None:
        """exit instance"""
        _logger.warning("local mode not support exit")

    def save_state(self, state) -> str:
        """save instance state"""
        key = generate_random_id()
        LocalObjectStore().put(key, state)
        return key

    def load_state(self, checkpoint_id) -> bytes:
        """load instance state"""
        return LocalObjectStore().get(checkpoint_id)

    def clear(self):
        """stop all workers"""
        self._running = False
        with self._lock:
            for worker in self.workers.values():
                worker.stop()
            self.workers.clear()
