#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

"""SSE stream writer for FaaS mode."""

from __future__ import annotations

_DEFAULT_MAX_WRITE_SIZE = 4 * 1024 * 1024


class Stream:
    """
    SSE stream writer.
    """
    def __init__(self, request_id: str, instance_id: str, max_write_size: int = _DEFAULT_MAX_WRITE_SIZE):
        self._request_id = request_id or ""
        self._instance_id = instance_id or ""
        self._max_write_size = max_write_size


    def write(self, serialized_str: str) -> None:
        """
        Write a serialized string message, the input must be a serialized `str`.
        """
        if serialized_str is None:
            serialized_str = ""
        if not isinstance(serialized_str, str):
            raise TypeError(f"stream.write() expects `str`, got {type(serialized_str)}")

        write_size_in_bytes = len(serialized_str.encode("utf-8"))
        if write_size_in_bytes > self._max_write_size:
            raise ValueError(
                f"Write data size is larger than {self._max_write_size} bytes: {write_size_in_bytes}"
            )

        if not self._request_id or not self._instance_id:
            raise RuntimeError("SSE stream not initialized: missing request_id/instance_id")

        from yr.fnruntime import stream_write

        stream_write(serialized_str, self._request_id, self._instance_id)


    def set_max_write_size(self, max_write_size: int) -> None:
        """
        Set the maximum limit for a single write operation.
        """
        self._max_write_size = max_write_size