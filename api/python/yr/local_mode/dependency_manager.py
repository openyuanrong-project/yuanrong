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

"""
Dependency Manager
"""
import concurrent.futures
import functools
import logging
from dataclasses import dataclass
from typing import Union, List, Callable, Optional

from yr.local_mode.task_spec import TaskSpec
from yr.object_ref import ObjectRef
from yr.runtime_holder import global_runtime

_logger = logging.getLogger(__name__)


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class TaskState:
    """TaskState"""
    task: TaskSpec
    object_dependencies: int
    instance_dependencies: int
    instance_error: Optional[BaseException]
    error: List[BaseException]


def resolve_dependency(task: TaskSpec, on_complete: Callable[[Optional[BaseException], List], None]):
    """resolve dependency"""
    object_dependencies = []
    instance_dependencies = []

    if task.instance_id is not None:
        instance_dependencies.append(task.instance_id)

    for arg in task.args:
        for ref in arg.refs:
            if isinstance(ref, ObjectRef):
                object_dependencies.append(ref)

    if len(object_dependencies) == 0 and len(instance_dependencies) == 0:
        on_complete(None, [])
        return
    state = TaskState(task, len(object_dependencies), len(instance_dependencies), None, [])

    def callback(future: concurrent.futures.Future, dependency: Union[ObjectRef, str]):
        err = future.exception() if future else None

        if isinstance(dependency, ObjectRef):
            state.object_dependencies -= 1
            if err:
                state.error.append(future.exception())
        else:
            state.instance_dependencies -= 1
            if err:
                state.instance_error = err

        if state.object_dependencies == 0 and state.instance_dependencies == 0:
            on_complete(state.instance_error, state.error)

    for ref in object_dependencies:
        ref.on_complete(functools.partial(callback, dependency=ref))
    for ref in instance_dependencies:
        global_runtime.get_runtime().set_get_callback(ref, functools.partial(callback, dependency=ref))
