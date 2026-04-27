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

"""ObjectRefGenerator"""
import asyncio
import logging
from typing import Optional
from yr.object_ref import ObjectRef
from yr.runtime_holder import global_runtime
from yr.fnruntime import GeneratorEndError
from yr.exception import GeneratorFinished

_logger = logging.getLogger(__name__)


class ObjectRefGenerator:
    """ ObjectRefGenerator streamming  return"""

    def __init__(self, object_ref: ObjectRef):
        self._obj_ref = object_ref
        self._generator_id = object_ref.id
        self._generator_task_exception = None
        self._generator_ref = None
        self._runtime = global_runtime.get_runtime()
        self._stop = False

    def __iter__(self) -> "ObjectRefGenerator":
        return self

    def __next__(self) -> ObjectRef:
        return self._next_sync()

    def __aiter__(self) -> "ObjectRefGenerator":
        return self

    async def __anext__(self):
        return await self._next_async()

    def get_generator_id(self):
        """get generator id """
        return self._generator_id

    def _next_sync(
        self,
        timeout_s: Optional[float] = None
    ) -> ObjectRef:
        """Waits for timeout_s and returns the object ref if available.

        If an object is not available within the given timeout, it
        returns a nil object reference.

        If -1 timeout is provided, it means it waits infinitely.

        Waiting is implemented as busy waiting.

        Raises StopIteration if there's no more objects
        to generate.

        The object ref will contain an exception if the task fails.
        When the generator task returns N objects, it can return
        up to N + 1 objects (if there's a system failure, the
        last object will contain a system level exception).

        Args:
            timeout_s: If the next object is not ready within
                this timeout, it returns the nil object ref.
        """
        if self._stop:
            raise StopIteration

        try:
            object_id = self._runtime.peek_object_ref_stream(self._generator_id, True)
        except GeneratorEndError as e:
            # generator stop error
            self._stop = True
            raise StopIteration from e

        except Exception as e:
            self._generator_task_exception = e
            self._stop = True
            self._generator_ref = ObjectRef(object_id="", need_incre=False, exception=e)
            return self._generator_ref

        if self._stop:
            raise StopIteration
        self._generator_ref = ObjectRef(object_id, need_incre=False)
        return self._generator_ref

    async def _suppress_exceptions(self, ref: ObjectRef) -> None:
        # Wrap a streamed ref to avoid asyncio warnings about not retrieving
        # the exception when we are just waiting for the ref to become ready.
        # The exception will get returned (or warned) to the user once they
        # actually await the ref.
        try:
            data = await ref
            self._generator_ref.set_data(data)
        except GeneratorFinished as e:
            self._stop = True
            self._generator_task_exception = e
            _logger.debug("generation finished, objectRef %s ", ref)
        except Exception as e:
            self._stop = True
            self._generator_task_exception = e
            _logger.debug("failed to await objectRef %s : %s", ref, str(e))

    async def _next_async(
            self,
            timeout_s: Optional[float] = None
    ):
        """Same API as _next_sync, but it is for async context."""
        if self._stop:
            raise StopAsyncIteration

        try:
            object_id = self._runtime.peek_object_ref_stream(self._generator_id, False)
            self._generator_ref = ObjectRef(object_id=object_id, need_incre=False)
            _, unready = await asyncio.wait(
                [asyncio.create_task(self._suppress_exceptions(self._generator_ref))],
                timeout=timeout_s
            )
            self._generator_ref.set_exception(self._generator_task_exception)
            if len(unready) > 0:
                return ObjectRef.nil()
        except GeneratorEndError as e:
            # generator stop error
            self._stop = True
            raise StopAsyncIteration from e

        except Exception as e:
            self._generator_task_exception = e
            self._stop = True
            self._generator_ref = ObjectRef(object_id="", need_incre=False, exception=e)
            return self._generator_ref

        if self._stop:
            raise StopAsyncIteration
        return self._generator_ref
