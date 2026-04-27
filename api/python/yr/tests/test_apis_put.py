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

import unittest
import yr
from yr.object_ref import ObjectRef
from unittest.mock import patch, Mock


class TestPut(unittest.TestCase):
    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_put_with_memoryview(self, mock_get_runtime):
        obj_refs = memoryview(bytearray(10 * 1024 * 1024))
        mock_runtime = Mock()
        mock_runtime.put.return_value = 1
        mock_get_runtime.return_value = mock_runtime
        yr.apis.set_initialized()
        result = yr.apis.put(obj_refs)
        self.assertIsInstance(result, ObjectRef)

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_put_with_bytes(self, mock_get_runtime):
        obj_refs = bytes(10 * 1024 * 1024)
        mock_runtime = Mock()
        mock_runtime.put.return_value = 10
        mock_get_runtime.return_value = mock_runtime
        yr.apis.set_initialized()
        result = yr.apis.put(obj_refs)
        self.assertIsInstance(result, ObjectRef)

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_put_with_bytearray(self, mock_get_runtime):
        obj_refs = bytearray(10 * 1024 * 1024)
        mock_runtime = Mock()
        mock_runtime.put.return_value = 1
        mock_get_runtime.return_value = mock_runtime
        yr.apis.set_initialized()
        result = yr.apis.put(obj_refs)
        self.assertIsInstance(result, ObjectRef)

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_put_with_object_ref(self, mock_rt):
        mock_rt.return_value.put.side_effect = TypeError
        obj = ObjectRef(1)
        yr.apis.set_initialized()
        with self.assertRaises(TypeError):
            yr.put(obj)

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_put_fail(self, mock_rt):
        mock_rt.return_value.put.side_effect = RuntimeError("mock error")
        with self.assertRaises(RuntimeError):
            yr.put(1)

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_put_null_ptr(self,mock_rt):
        mock_rt.return_value.put.side_effect = ValueError("value is None or has zero length")
        obj = None
        yr.apis.set_initialized()
        with self.assertRaises(ValueError):
            yr.put(obj)


    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_put_len_zero_bytes(self,mock_rt):
        mock_rt.return_value.put.side_effect = ValueError
        obj = bytes(0)
        yr.apis.set_initialized()
        with self.assertRaises(ValueError):
            yr.put(obj)

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_put_len_zero_bytearray(self,mock_rt):
        mock_rt.return_value.put.side_effect = ValueError
        obj = bytearray(0)
        yr.apis.set_initialized()
        with self.assertRaises(ValueError):
            yr.put(obj)

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_put_len_zero_memoryview(self, mock_rt):
        mock_rt.return_value.put.side_effect = ValueError
        o = bytes(0)
        obj = memoryview(o)
        yr.apis.set_initialized()
        with self.assertRaises(ValueError):
            yr.put(obj)




if __name__ == '__main__':
    unittest.main()