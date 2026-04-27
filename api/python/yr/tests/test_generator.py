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

import asyncio
import gc
import unittest
import logging
from unittest.mock import Mock, patch
from concurrent.futures import Future

import yr
from yr.fnruntime import GeneratorEndError
from yr.generator import ObjectRefGenerator
from yr.object_ref import ObjectRef


logger = logging.getLogger(__name__)

class testObjectRef(unittest.TestCase):
    def setUp(self):
        self.runtime_patcher = patch('yr.runtime_holder.global_runtime.get_runtime')
        self.mock_get_runtime = self.runtime_patcher.start()
        self.mock_get_runtime.return_value = Mock()

    def tearDown(self):
        gc.collect()
        self.runtime_patcher.stop()

    def test_obj(self):
        obj = ObjectRef(object_id="object_1234", task_id="task_abcd")
        self.assertEqual(obj.task_id, "task_abcd")
        self.assertEqual(obj.id, "object_1234")
        with patch.object(obj, 'on_complete') as mock_on_complete:
            mock_on_complete.return_value = "on_complete"
            f = obj.get_future()
            self.assertTrue(isinstance(f, Future))

        with patch.object(obj, 'get_future') as mock_get_future:
            f = Future()
            f.set_result("result")
            mock_get_future.return_value = f
            obj.wait()
            self.assertTrue(obj.done())

        with patch.object(obj, 'get_future') as mock_get_future:
            f = Future()
            f.set_exception(RuntimeError("mock exception"))
            mock_get_future.return_value = f
            res = obj.is_exception()
            self.assertTrue(res, res)

        with patch.object(obj, 'get_future') as mock_get_future:
            f = Future()
            mock_get_future.return_value = f
            obj.cancel()
            with self.assertRaises(Exception):
                f.result()


    def test_exception(self):
        obj = ObjectRef("object_1234")
        obj.set_exception(RuntimeError("mock exception"))
        with self.assertRaises(RuntimeError):
            obj.exception()

        with self.assertRaises(RuntimeError):
            obj.get()

        f = obj.get_future()
        self.assertTrue(f.exception() is not None)

    def test_data(self):
        obj = ObjectRef("object_1234")
        obj.set_data("data")
        self.assertEqual(obj.get_future().result(), "data")

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    @patch("yr.log.get_logger")
    def test_get(self, mock_logger, get_runtime):
        mock_logger.return_value = logger
        mock_runtime = Mock()
        mock_runtime.get.return_value = ["data1"]
        get_runtime.return_value = mock_runtime

        obj = ObjectRef("object_1234")
        with self.assertRaises(ValueError):
            obj.get(-2)

        self.assertEqual(obj.get(), "data1")


class TestObjectRefGenerator(unittest.TestCase):

    def setUp(self):
        pass

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_iter(self, get_runtime):
        mock_runtime = Mock()
        mock_runtime.peek_object_ref_stream.return_value = 'test_object_id'
        get_runtime.return_value = mock_runtime
        generator_id = 'test_id'
        generator = ObjectRefGenerator(
            ObjectRef(generator_id, need_incre=False))
        self.assertEqual(iter(generator), generator)

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_next_sync(self, get_runtime):
        mock_runtime = Mock()
        mock_runtime.peek_object_ref_stream.return_value = 'test_object_id'
        get_runtime.return_value = mock_runtime
        generator_id = 'test_id'
        generator = ObjectRefGenerator(
            ObjectRef(generator_id, need_incre=False))
        self.assertEqual(generator._next_sync().id, 'test_object_id')
        print("finished test_next_sync")

    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_next_sync_with_exception(self, get_runtime):
        mock_runtime = Mock()
        mock_runtime.peek_object_ref_stream.side_effect = GeneratorEndError("failed")
        get_runtime.return_value = mock_runtime
        generator_id = 'test_id'
        generator = ObjectRefGenerator(
            ObjectRef(generator_id, need_incre=False))
        self.assertRaises(StopIteration, generator._next_sync)
        print("finished test_next_sync")


if __name__ == '__main__':
    unittest.main()
