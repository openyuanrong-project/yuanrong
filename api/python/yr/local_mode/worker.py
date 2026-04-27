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

"""local mode worker"""
import concurrent.futures
import logging
import queue
import threading
import traceback
from typing import Iterable

import yr

from yr import signature
from yr.exception import YRInvokeError, YRError
from yr.libruntime_pb2 import InvokeType
from yr.local_mode.local_object_store import LocalObjectStore
from yr.local_mode.task_spec import TaskSpec
from yr.object_ref import ObjectRef

_logger = logging.getLogger(__name__)


class Worker(threading.Thread):
    """Local mode worker"""
    _instance = None
    _queue = None
    _running = False
    _instance_id = ''
    pool = None

    def init(self, instance_id: str, concurrency):
        """init worker and create instance"""
        self.pool = concurrent.futures.ThreadPoolExecutor(max_workers=concurrency, thread_name_prefix=instance_id)
        self._instance_id = instance_id
        self._queue = queue.Queue()
        self._running = True
        self.setDaemon(True)
        self.start()
        _logger.debug("init %s finished.", self._instance_id)

    def run(self) -> None:
        """main loop"""
        while self._running:
            task = self._queue.get()
            if task is None:
                break
            future = self.pool.submit(self._execute, task)
            future.add_done_callback(_error_logging)

    def submit(self, task: TaskSpec):
        """submit a invoke task"""
        self._queue.put(task)

    def stop(self, timeout=None):
        """stop worker"""
        self._running = False
        self._queue.put(None)
        self.join(timeout)
        while not self._queue.empty():
            item = self._queue.get_nowait()
            if item is None:
                break
            item.callback(None, RuntimeError("worked is stopped."))
        self.pool.shutdown(wait=False)

    def _execute(self, task: TaskSpec):
        try:
            result = self._invoke_function(task)
        except Exception as err:
            if isinstance(err, YRInvokeError):
                result = YRInvokeError(err.cause, traceback.format_exc())
            else:
                result = YRInvokeError(err, traceback.format_exc())
        try:
            if isinstance(result, YRError):
                for i in task.object_ids:
                    LocalObjectStore().set_exception(i, result)
            else:
                if len(task.object_ids) == 1:
                    result = [result]
                if isinstance(result, Iterable):
                    for k, v in zip(task.object_ids, result):
                        LocalObjectStore().set_result(k, v)
                else:
                    raise RuntimeError("except return nums %s, actual 1", len(task.object_ids))
        except Exception as e:
            for i in task.object_ids:
                LocalObjectStore().set_exception(i, YRInvokeError(e, traceback.format_exc()))
        _logger.debug("succeed to call, traceID: %s", task.task_id)

    def _invoke_function(self, task: TaskSpec):
        task.args = _process_args(task.args)
        args, kwargs = signature.recover_args(task.args)
        if task.invoke_type == InvokeType.CreateInstance:
            return self._create_instance(task, *args, **kwargs)
        if task.invoke_type == InvokeType.InvokeFunction:
            return self._instance_function(task, *args, **kwargs)
        if task.invoke_type == InvokeType.CreateInstanceStateless:
            return None
        return _normal_function(task, *args, **kwargs)

    def _instance_function(self, task: TaskSpec, *args, **kwargs):
        instance_function_name = task.function_meta.functionName
        return getattr(self._instance, instance_function_name)(*args, **kwargs)

    def _create_instance(self, task: TaskSpec, *args, **kwargs):
        code = LocalObjectStore().get(task.function_meta.codeID)
        self._instance = code(*args, **kwargs)


def _error_logging(future: concurrent.futures.Future):
    if future.exception() is not None:
        _logger.exception(future.exception())


def _normal_function(task: TaskSpec, *args, **kwargs):
    code = LocalObjectStore().get(task.function_meta.codeID)
    return code(*args, **kwargs)


def _process_args(args_list):
    def func(arg):
        if isinstance(arg, ObjectRef):
            return yr.get(arg)
        return arg

    return list(map(func, args_list))
