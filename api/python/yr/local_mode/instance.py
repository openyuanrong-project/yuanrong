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

"""Instance class for scheduler"""
import time
from concurrent.futures import Future
from dataclasses import dataclass, field
from typing import Optional, Dict


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class Resource:
    """instance resource"""
    comparable_res: tuple = ('cpu', 'memory', 'concurrency', 'resources')
    cpu: int = 500
    memory: int = 500
    concurrency: int = 1
    resources: Dict[str, float] = field(default_factory=dict)

    def __hash__(self):
        return hash(str(self))

    def __eq__(self, other: Optional['Resource']):
        if not other:
            return False
        for res in self.comparable_res:
            if getattr(self, res) != getattr(other, res):
                return False
        return True

    def __str__(self):
        res = "".join(map(lambda x: f"({x[0]},{x[1]:.4f})", sorted(self.resources.items(), key=lambda x: x[0])))
        return f"{self.cpu}-{self.memory}-{self.concurrency}-{res}"


class Instance:
    """
    Instance used by Scheduler
    """
    __slots__ = ["__resource", "__instance_id", "__tasks", "__last_activate_time", "__is_recycled", "__future"]

    def __init__(self, instance_id: str, resource: Resource, future: Future):
        self.__future = future
        self.__instance_id = instance_id
        self.__resource = resource
        self.__tasks = set()
        self.__last_activate_time = time.time()
        self.__is_recycled = False

    def __str__(self):
        return f"instance id: {self.__instance_id if self.is_activate else 'not activate'} " \
               f"task count: {self.task_count} " \
               f"resource: {self.__resource} " \
               f"last activate time: {time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(self.__last_activate_time))}"

    @property
    def last_activate_time(self) -> float:
        """
        last activate time. used to recycle
        """
        return self.__last_activate_time

    @property
    def is_activate(self) -> bool:
        """
        is instance activate
        """
        if self.__is_recycled:
            return False
        return self.__future.done() and not self.__future.exception()

    @property
    def resource(self) -> Resource:
        """
        get instance resource
        """
        return self.__resource

    @property
    def task_count(self) -> int:
        """
        get task count
        """
        return len(self.__tasks)

    @property
    def instance_id(self):
        """
        get instance id
        """
        return self.__instance_id

    def set_recycled(self):
        """
        set recycled status
        """
        self.__is_recycled = True

    def add_task(self, task_id: str):
        """
        add a task to instance
        """
        self.__tasks.add(task_id)

    def delete_task(self, task_id: str):
        """
        delete task
        """
        self.__tasks.remove(task_id)
        self.refresh()

    def refresh(self):
        """
        refresh last activate time
        """
        self.__last_activate_time = time.time()
