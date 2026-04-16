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

"""ObjectRef"""
import asyncio
import functools
import json
from concurrent.futures import Future
import logging
from typing import Any, Union

from yr.exception import YRInvokeError, YRError, GeneratorFinished
from yr.err_type import ErrorInfo, ErrorCode
from yr.libruntime_pb2 import FunctionMeta

import yr
from yr import runtime_holder
from yr.common import constants
from yr.ds_tensor_client_manager import get_tensor_client

_logger = logging.getLogger(__name__)


def _set_future_helper(
        result: Any,
        *,
        f: Union[asyncio.Future, Future],
):
    if f.done():
        return

    if isinstance(result, ErrorInfo):
        if result.error_code == ErrorCode.ERR_GENERATOR_FINISHED.value:
            f.set_exception(GeneratorFinished(""))
            return
        if result.error_code != ErrorCode.ERR_OK.value:
            f.set_exception(RuntimeError(
                f"code: {result.error_code}, module code {result.module_code}, msg: {result.msg}"))
    elif isinstance(result, YRInvokeError):
        f.set_exception(result.origin_error())
    elif isinstance(result, YRError):
        f.set_exception(result)
    elif isinstance(result, RuntimeError):
        f.set_exception(result)
    else:
        f.set_result(result)


class ObjectRef:
    """ObjectRef"""
    _id = None
    _task_id = None

    def __init__(self, object_id: str, task_id=None, need_incre=True, need_decre=True, exception=None,
                 enable_tensor_transport=False):
        """Initialize the ObjectRef."""
        self._id = object_id
        self._task_id = task_id
        self._need_decre = need_decre
        self._exception = exception
        self._data = None
        self._enable_tensor_transport = enable_tensor_transport
        self.npu_obj_ids = []
        self.instance_id = ""
        global_runtime = runtime_holder.global_runtime.get_runtime()
        if need_incre and global_runtime and exception is None:
            global_runtime.increase_global_reference([self._id])

    def __del__(self):
        try:
            if runtime_holder is None:
                return
            global_runtime = runtime_holder.global_runtime.get_runtime()
        except RuntimeError:
            return
        try:
            if self._need_decre and global_runtime:
                global_runtime.decrease_global_reference([self._id])
            if self._enable_tensor_transport and len(self.npu_obj_ids) != 0:
                ds_tensor_client = get_tensor_client()
                ds_tensor_client.dev_delete(self.npu_obj_ids)
                # In future versions, this step will no longer be required
                if len(self.instance_id) != 0 and global_runtime:
                    opts = yr.InvokeOptions()
                    opts.is_delete_remote_tensor = True
                    global_runtime.invoke_instance(FunctionMeta(), self.instance_id, self.npu_obj_ids, opts, 1)
        except (AttributeError, TypeError, NameError):
            # 忽略模块清理时的属性访问错误
            pass

    def __copy__(self):
        return self

    def __deepcopy__(self, memo):
        return self

    def __str__(self):
        return self.id

    def __eq__(self, other):
        return self.id == other.id

    def __hash__(self):
        return hash(self.id)

    def __repr__(self):
        return self.id

    def __await__(self):
        return self.as_future().__await__()

    @property
    def task_id(self):
        """Task id."""
        return self._task_id

    @task_id.setter
    def task_id(self, value):
        """Task id."""
        if value is not None:
            self._task_id = value

    @property
    def id(self):
        """ObjectRef id."""
        return self._id

    def as_future(self) -> asyncio.Future:
        """
        Wrap `ObjectRef` with an `asyncio.Future`.

        Note that the future cancellation will not cancel the corresponding
        task when the ObjectRef representing return object of a task.

        Returns:
            An `asyncio.Future` that wraps `ObjectRef`. Data type is `asyncio.Future`.
        """
        return asyncio.wrap_future(self.get_future())

    def get_future(self):
        """
        Get future.

        Returns:
            The future of ObjectRef. Data type is Future.
        """
        f = Future()
        if self._exception is not None:
            _set_future_helper(self._exception, f=f)
            return f
        if self._data is not None:
            f.set_result(self._data)
            return f
        self.on_complete(functools.partial(_set_future_helper, f=f))
        f.object_ref = self
        return f

    def wait(self, timeout=None):
        """
        Wait stateless function done.

        Args:
            timeout (int, optional): The number of seconds to wait for the exception if the future isn't done.
                The default value ``None`` indicates that there is no limit on the wait time.
        """
        future = self.get_future()
        if future is not None:
            future.result(timeout=timeout)

    def is_exception(self) -> bool:
        """
        Whether future exception.

        Returns:
            Whether future exception. Data type is bool.
        """
        future = self.get_future()
        if future is None:
            return False
        return future.exception() is not None

    def done(self):
        """
        Return ``True`` if the obj future was cancelled or finished executing.

        Returns:
            Whether the obj future was cancelled or finished executing. Data type is bool.
        """
        future = self.get_future()
        if future:
            return future.done()
        return True

    def cancel(self):
        """
        Cancel the obj future

        Returns:
            Returns ``True`` if the future was cancelled, ``False`` otherwise. A future cannot be cancelled if it is
            running or has already completed. Data type is bool.
        """
        future = self.get_future()
        if future:
            future.cancel()

    def on_complete(self, callback):
        """
        Register callback.

        Args:
            callback (Callable): User callback.
        """
        runtime_holder.global_runtime.get_runtime().set_get_callback(self.id, callback)

    def get(self, timeout: int = constants.DEFAULT_GET_TIMEOUT) -> Any:
        """This function is used to retrieve an object.

        Args:
            timeout (int, optional): The maximum time in seconds to wait for the
                                     interface object to be retrieved. Defaults to ``300``.

        Returns:
            The retrieved object. Data type is Any.

        Raises:
            ValueError: If the timeout parameter is less than or equal to the 0 and is not -1.
            YRInvokeError: If the retrieved result is an instance of YRInvokeError.
        """
        self.exception()
        if timeout <= constants.MIN_TIMEOUT_LIMIT and timeout != constants.NO_LIMIT:
            raise ValueError("Parameter 'timeout' should be greater than 0 or equal to -1 (no timeout)")

        objects = runtime_holder.global_runtime.get_runtime().get([self.id], timeout, False)
        result_str = objects[0]

        try:
            obj = json.loads(result_str)
        except json.decoder.JSONDecodeError:
            _logger.warning(f"Failed to decode the result with object ID [{self.id}] using 'json.loads'."
                                     f"result string: {result_str}")
            obj = result_str
        return obj

    def exception(self):
        """Raise exception if exception is not none."""
        if self._exception is not None:
            raise self._exception

    def set_data(self, data):
        """
        Set data.

        Args:
            data (ObjectRef): Data to be set.
        """
        self._data = data

    def set_exception(self, e):
        """Set exception."""
        self._exception = e
