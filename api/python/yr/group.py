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

from dataclasses import dataclass
from yr.config import GroupOptions
from yr import log, runtime_holder


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class Group:
    """
    A class for managing the lifecycle of grouped instances.

    The `Group` class is responsible for managing the lifecycle of grouped instances, including their creation and
    destruction. It follows the fate-sharing principle, where all instances in the group are created or destroyed
    together.

    The `Group` class provides methods to create, terminate, suspend, resume, and manage grouped instances. 
    It ensures that all instances in the group are treated as a single unit, and any failure during group 
    creation will roll back the entire group.

    Examples:
        >>> import yr
        >>>
        >>> yr.init()
        >>>
        >>> @yr.instance
        ... class Counter:
        ...     sum = 0
        ...
        ...     def add(self, a):
        ...         self.sum += a
        ...
        >>> group_opts = yr.GroupOptions()
        >>> group_name = "test"
        >>> g = yr.Group(group_name, group_opts)
        >>> opts = yr.InvokeOptions()
        >>> opts.group_name = group_name
        >>> ins = Counter.options(opts).invoke()
        >>> g.invoke()
        >>> res = ins.add.invoke()
        >>> print(yr.get(res))
        >>> g.terminate()
        >>>
        >>> yr.finalize()
    """
    group_name: str = ""
    group_opts: GroupOptions = None

    def invoke(self):
        """
        Execute the creation of a group of instances following the fate-sharing principle.

        This function is used to create a group of instances that share the same fate. All instances in the group
        will be created together, and if one instance fails, the entire group will be rolled back.
        The group configuration and options are defined in the `GroupOptions`.
        """
        runtime_holder.global_runtime.get_runtime().create_group(
            self.group_name, self.group_opts)

    def terminate(self):
        """
        Terminate a group of instances.

        This function is used to delete or terminate a group of instances that were created together.
        All instances in the group will be cleaned up or removed as a single unit, following the fate-sharing principle.
        """
        runtime_holder.global_runtime.get_runtime().terminate_group(self.group_name)

    def wait(self):
        runtime_holder.global_runtime.get_runtime().wait_group(self.group_name)

    def suspend(self):
        """
        Suspend all instances managed by this Group.

        This method terminates all instances created and managed by the
        Group, while first checkpointing their state and metadata. The checkpoint
        allows the Group to preserve necessary instance information before the
        instances are killed and their resources are released.

        The saved checkpoint can later be used by `resume` to recover and
        restart the instances.
        Examples:
            >>> import yr
            >>>
            >>> yr.init()
            >>>
            >>> @yr.instance
            ... class Counter:
            ...     sum = 0
            ...
            ...     def add(self, a):
            ...         self.sum += a
            ...
            >>> group_opts = yr.GroupOptions()
            >>> group_name = "test"
            >>> g = yr.Group(group_name, group_opts)
            >>> opts = yr.InvokeOptions()
            >>> opts.group_name = group_name
            >>> ins = Counter.options(opts).invoke()
            >>> g.invoke()
            >>> res = ins.add.invoke()
            >>> print(yr.get(res))
            >>>
            >>> g.suspend()
            >>> g.resume()
            >>> res = ins.add.invoke()
            >>> print(yr.get(res))
            >>> yr.finalize()
        """
        runtime_holder.global_runtime.get_runtime().suspend_group(self.group_name)

    def resume(self):
        """
        Resume all instances managed by this Group.

        This method recovers previously checkpointed instance information created
        during `suspend`, and then recreates and restarts the instances
        based on that checkpoint.

        After recovery, the instances continue execution using the
        restored state, rather than starting entirely from scratch.
        """
        runtime_holder.global_runtime.get_runtime().resume_group(self.group_name)
