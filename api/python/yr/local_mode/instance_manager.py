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

"""instance manager"""
import concurrent.futures
import logging
import threading
import time
from collections import OrderedDict, defaultdict
from concurrent.futures import Future
from typing import Optional, List, Tuple

from yr.common.utils import generate_task_id, generate_random_id

from yr.exception import YRInvokeError

from yr.libruntime_pb2 import InvokeType
from yr.local_mode.instance import Resource, Instance
from yr.local_mode.local_object_store import LocalObjectStore
from yr.local_mode.scheduler import Scheduler
from yr.local_mode.task_spec import TaskSpec
from yr.local_mode.timer import Timer

_logger = logging.getLogger(__name__)

_PRINT_ERROR_PERIOD = 60


def print_instance_create_error(future: Future):
    """print instance create error"""
    if future.exception():
        _logger.exception(future.exception())


def warning_if_failed(future: concurrent.futures.Future, describe: str):
    """warning for catching future exception"""
    try:
        future.result()
    except RuntimeError as e:
        _logger.warning("%s, err:%s", describe, e)


class InstanceManager:
    """manager stateless instance"""
    __slots__ = ["__instances", "__scheduler", "__invoke_client", "__recycle_scheduler", "__recycle_time",
                 "__recycle_period", "__lock", "__failed_count", "__last_failed_reason", "__local_store",
                 "__default_timeout"]

    def __init__(self, scheduler: Scheduler, invoke_client, recycle_time=2):
        self.__instances = dict()
        self.__scheduler = scheduler
        self.__invoke_client = invoke_client
        self.__local_store = LocalObjectStore()
        self.__recycle_time = recycle_time
        self.__recycle_period = float(recycle_time) / 10
        self.__recycle_scheduler = Timer()
        self.__recycle_scheduler.after(self.__recycle_period, self.__recycle)
        self.__lock = threading.RLock()
        self.__failed_count = defaultdict(int)
        self.__last_failed_reason = {}
        self.__default_timeout = 30

    def get_instances(self, resource: Resource) -> List[Instance]:
        """get instance by resource"""
        with self.__lock:
            instances = self.__instances.get(resource)
            if instances:
                return list(instances.values())
        return list()

    def get_failed_count(self, resource: Resource):
        """get instance create failed count by resource"""
        return self.__failed_count[resource]

    def get_last_failed_reason(self, resource: Resource):
        """get instance create failed count by resource"""
        return self.__failed_count.get(resource, None)

    def check_last_failed_reason(self, resource: Resource) -> Tuple[bool, Optional[Exception]]:
        """
        check whether an exception occurs during instance creation.
        will not return true when instance can be create before but resource not enough now.
        """
        err = self.__last_failed_reason.get(resource)
        if not err:
            return True, err
        if isinstance(err, YRInvokeError):
            return False, err.origin_error()
        return True, err

    def schedule(self, task: TaskSpec, resource: Resource) -> Optional[Instance]:
        """select a instance for stateless function"""

        instance = None
        if resource not in self.__instances:
            return instance
        instances = self.get_instances(resource)
        if len(instances) != 0:
            instance = self.__scheduler.schedule(task, instances[::-1])
            with self.__lock:
                if instance:
                    instance.add_task(task_id=task.task_id)
                    return instance
        return instance

    def scale_out(self, task: TaskSpec, resource: Resource) -> Tuple[str, Future]:
        """scale out a instance"""
        future = concurrent.futures.Future()
        task_id = generate_task_id()
        task_spec = TaskSpec(
            future=future,
            task_id=task_id,
            object_ids=[generate_random_id()],
            function_meta=task.function_meta,
            trace_id=task_id,
            invoke_type=InvokeType.CreateInstanceStateless,
            args=[],
            invoke_options=task.invoke_options,
        )

        instance_id = self.__invoke_client.create(task_spec)
        instance = Instance(instance_id=instance_id, resource=resource, future=future)
        self.__add_instance(resource, instance)
        self.__local_store.set_return_obj(instance_id, future)
        future.add_done_callback(self.__scale_out_result_process(resource, instance))
        return instance_id, future

    def clear(self):
        """clear instances"""
        with self.__lock:
            # send kill(2) to clear stateless and stateful instances from job-id when finalize
            self.__instances.clear()

    def kill_instance(self, instance_id: str):
        """kill stateless instance"""
        need_kill = False
        with self.__lock:
            for instances in self.__instances.values():
                if instance_id in instances:
                    ins = instances.pop(instance_id)
                    _logger.debug("instance need kill: %s", ins)
                    need_kill = True
        if need_kill:
            self.__delete_instance(instance_id)

    def __add_instance(self, resource: Resource, instance: Instance):
        with self.__lock:
            self.__instances.setdefault(resource, OrderedDict())[instance.instance_id] = instance

    def __recycle(self):
        self.__recycle_scheduler.after(self.__recycle_period, self.__recycle)
        with self.__lock:
            need_recycle = self.__clear_instance()
        [self.__delete_instance(ins.instance_id) for ins in need_recycle]

    def __clear_instance(self) -> List[Instance]:
        now = time.time()
        need_recycle = []

        for res, instances in self.__instances.items():
            activate_cnt = 0
            for instance_id in list(instances.keys()):
                ins = instances.get(instance_id)
                if not ins or not ins.instance_id.done():
                    continue
                if now - ins.last_activate_time > self.__recycle_time and ins.task_count == 0:
                    ins = instances.pop(instance_id)
                    ins.set_recycled()
                    _logger.debug("instance need recycle: %s", ins)
                    need_recycle.append(ins)
                activate_cnt += 1
            _logger.debug("activate instance count: %s res: %s", activate_cnt, res)
        return need_recycle

    def __delete_instance(self, instance_id: str):
        self.__invoke_client.kill(instance_id)

    def __scale_out_result_process(self, resource: Resource, instance: Instance):

        def callback(future: Future):
            """print instance create error"""
            if future.exception():
                self.__failed_count[resource] += 1
                self.__last_failed_reason[resource] = future.exception()
                with self.__lock:
                    instances = self.__instances.get(resource)
                    if instances:
                        instances.pop(instance.instance_id, None)
            else:
                self.__failed_count[resource] = 0
                self.__last_failed_reason.pop(resource, None)
                instance.refresh()

        return callback
