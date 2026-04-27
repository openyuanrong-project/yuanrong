#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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
from typing import Any, List, Optional

import yr
from yr.exception import (
    GetTimeoutError,
    YRChannelError,
    YRChannelTimeoutError,
    YRInvokeError,
)


def _process_return_vals(return_vals: List[Any], return_single_output: bool):
    """
    Process return values for return to the DAG caller. Any exceptions found in
    return_vals will be raised. If return_single_output=True, it indicates that
    the original DAG did not have a MultiOutputNode, so the DAG caller expects
    a single return value instead of a list.
    """
    # Check for exceptions.
    if isinstance(return_vals, Exception):
        raise return_vals

    for val in return_vals:
        if isinstance(val, YRInvokeError):
            raise val

    if return_single_output:
        if len(return_vals) != 1:
            raise ValueError("The DAG caller expected a single return value, but got multiple.")
        return return_vals[0]

    return return_vals


class CompiledDAGRef:
    """
    A reference to a compiled DAG execution result.

    This is a subclass of ObjectRef and resembles ObjectRef.  For example,
    similar to ObjectRef, yr.get() can be called on it to retrieve the result.
    However, there are several major differences:
    1. yr.get() can only be called once per CompiledDAGRef.
    2. yr.wait() is not supported.
    3. CompiledDAGRef cannot be copied, deep copied, or pickled.
    4. CompiledDAGRef cannot be passed as an argument to another task.
    """

    def __init__(
        self,
        dag: "yr.dag.CompiledDAG",
        execution_index: int,
        channel_index: Optional[int] = None,
    ):
        """
        Args:
            dag: The compiled DAG that generated this CompiledDAGRef.
            execution_index: The index of the execution for the DAG.
                A DAG can be executed multiple times, and execution index
                indicates which execution this CompiledDAGRef corresponds to.
            actor_execution_loop_refs: The actor execution loop refs that
                are used to execute the DAG. This can be used internally to
                check the task execution errors in case of exceptions.
            channel_index: The index of the DAG's output channel to fetch
                the result from. A DAG can have multiple output channels, and
                channel index indicates which channel this CompiledDAGRef
                corresponds to. If channel index is not provided, this CompiledDAGRef
                wraps the results from all output channels.

        """
        self._dag = dag
        self._execution_index = execution_index
        self._channel_index = channel_index
        # Whether yr.get() was called on this CompiledDAGRef.
        self._yr_get_called = False
        self._dag_output_channels = dag.dag_output_channels

    def __str__(self):
        return (
            f"CompiledDAGRef({self._dag.get_id()}, "
            f"execution_index={self._execution_index}, "
            f"channel_index={self._channel_index})"
        )

    def __del__(self):
        # If the dag is already teardown, it should do nothing.
        if self._dag.is_teardown:
            return

        if self._yr_get_called:
            # get() was already called, no further cleanup is needed.
            return

        self._dag._delete_execution_results(self._execution_index, self._channel_index)

    def get(self, timeout: Optional[float] = None):
        """
        Args:
            timeout (float, optional): The timeout value. A value of -1 means wait indefinitely. Limit: -1, (0, ∞).
        """
        if self._yr_get_called:
            raise ValueError(
                "yr.get() can only be called once "
                "on a CompiledDAGRef, and it was already called."
            )

        self._yr_get_called = True
        try:
            self._dag.execute_until(
                self._execution_index, self._channel_index, timeout
            )
            return_vals = self._dag.get_execution_results(
                self._execution_index, self._channel_index
            )
        except YRChannelTimeoutError:
            raise
        except YRChannelError:
            # If we get a channel error, we'd like to call yr.get() on
            # the actor execution loop refs to check if this is a result
            # of task execution error which could not be passed down
            # (e.g., when a pure NCCL channel is used, it is only
            # able to send tensors, but not the wrapped exceptions).
            # In this case, we'd like to raise the task execution error
            # (which is the actual cause of the channel error) instead
            # of the channel error itself.
            # actor task refs have errors.
            actor_execution_loop_refs = list(self._dag.worker_task_refs.values())
            try:
                yr.get(actor_execution_loop_refs, timeout=10)
            except GetTimeoutError as timeout_error:
                raise Exception(
                    "Timed out when getting the actor execution loop exception. "
                    "This should not happen, please file a GitHub issue."
                ) from timeout_error
            except Exception as execution_error:
                # Use 'from None' to suppress the context of the original
                # channel error, which is not useful to the user.
                raise execution_error from None
        except Exception as e:
            raise e
        return _process_return_vals(return_vals, True)


class CompiledDAGFuture:
    """
    A reference to a compiled DAG execution result, when executed with asyncio.
    This differs from CompiledDAGRef in that `await` must be called on the
    future to get the result, instead of `yr.get()`.

    This resembles async usage of ObjectRefs. For example, similar to
    ObjectRef, `await` can be called directly on the CompiledDAGFuture to
    retrieve the result.  However, there are several major differences:
    1. `await` can only be called once per CompiledDAGFuture.
    2. yr.wait() is not supported.
    3. CompiledDAGFuture cannot be copied, deep copied, or pickled.
    4. CompiledDAGFuture cannot be passed as an argument to another task.
    """

    def __init__(
        self,
        dag: "yr.dag.CompiledDAG",
        execution_index: int,
        fut: "asyncio.Future",
        channel_index: Optional[int] = None,
    ):
        self._dag = dag
        self._execution_index = execution_index
        self._fut = fut
        self._channel_index = channel_index

    def __str__(self):
        return (
            f"CompiledDAGFuture({self._dag.get_id()}, "
            f"execution_index={self._execution_index}, "
            f"channel_index={self._channel_index})"
        )

    def __await__(self):
        if self._fut is None:
            raise ValueError(
                "CompiledDAGFuture can only be awaited upon once, and it has "
                "already been awaited upon."
            )

        # NOTE: If the object is zero-copy deserialized, then it will
        # stay in scope as long as this future is in scope. Therefore, we
        # delete self._fut here before we return the result to the user.
        fut = self._fut
        self._fut = None

        if not self._dag.has_execution_results(self._execution_index):
            result = yield from fut.__await__()
            self._dag._max_finished_execution_index += 1
            self._dag._cache_execution_results(self._execution_index, result)

        return_vals = self._dag.get_execution_results(
            self._execution_index, self._channel_index
        )
        return _process_return_vals(return_vals, True)

    def __del__(self):
        if self._dag.is_teardown:
            return

        if self._fut is None:
            # await() was already called, no further cleanup is needed.
            return

        self._dag.delete_execution_results(self._execution_index, self._channel_index)
