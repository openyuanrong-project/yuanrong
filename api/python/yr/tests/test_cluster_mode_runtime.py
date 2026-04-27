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
from unittest.mock import Mock

import yr
from yr.config_manager import ConfigManager
from yr.err_type import ErrorCode, ErrorInfo, ModuleCode
from yr.serialization import Serialization
from yr.libruntime_pb2 import ApiType, FunctionMeta
from yr.object_ref import ObjectRef
from yr.base_runtime import CreateParam


class TestClusterModeRuntime(unittest.TestCase):
    num = 1

    def setUp(self):
        ConfigManager().rt_server_address = "127.0.0.1:1122"
        self.runtime = yr.cluster_mode_runtime.ClusterModeRuntime()
        self.runtime.set_initialized()

        def mock_get_async(obj_id, callback_wrapper):
            global num
            num = 2
            buf = b'{"key": "value"}'
            callback_wrapper(obj_id, ErrorInfo(), buf)

        self.call_back = Mock()

        mock_fnruntime = Mock()
        mock_fnruntime.get_async.side_effect = mock_get_async
        mock_fnruntime.get.return_value = [b"data1", b"data2"]
        mock_fnruntime.put.return_value = "putobj"
        mock_fnruntime.wait.return_value = tuple([["ready1", "ready2"], ["unready"], []])
        mock_fnruntime.kv_write.side_effect = RuntimeError("mock exception")
        mock_fnruntime.kv_read.return_value = ["value"]
        mock_fnruntime.kv_del.side_effect = RuntimeError("mock exception")

        mock_fnruntime.increase_global_reference.side_effect = RuntimeError("mock exception")
        mock_fnruntime.decrease_global_reference.side_effect = RuntimeError("mock exception")
        mock_fnruntime.invoke_by_name.side_effect = RuntimeError("mock exception")
        mock_fnruntime.create_instance.return_value = "instance"
        mock_fnruntime.invoke_instance.return_value = ["invoke_instance"]

        mock_fnruntime.cancel.side_effect = RuntimeError("mock exception")
        mock_fnruntime.terminate_instance.side_effect = RuntimeError("mock exception")
        mock_fnruntime.terminate_instance_sync.side_effect = RuntimeError("mock exception")
        mock_fnruntime.terminate_group.side_effect = RuntimeError("mock exception")
        mock_fnruntime.exit.side_effect = RuntimeError("mock exception")
        mock_fnruntime.receive_request_loop.side_effect = RuntimeError("mock exception")
        mock_fnruntime.create_stream_producer.return_value = "producer"
        mock_fnruntime.create_stream_consumer.return_value = "consumer"
        mock_fnruntime.query_global_producers_num.return_value = 10
        mock_fnruntime.query_global_consumers_num.return_value = 10
        mock_fnruntime.delete_stream.side_effect = RuntimeError("mock exception")

        mock_fnruntime.get_real_instance_id.return_value = "get_real_instance_id"
        mock_fnruntime.save_real_instance_id.side_effect = RuntimeError("mock exception")
        mock_fnruntime.is_object_existing_in_local.return_value = True

        mock_fnruntime.set_uint64_counter.side_effect = RuntimeError("mock exception")
        mock_fnruntime.increase_uint64_counter.side_effect = RuntimeError("mock exception")
        mock_fnruntime.set_double_counter.side_effect = RuntimeError("mock exception")
        mock_fnruntime.reset_uint64_counter.side_effect = RuntimeError("mock exception")
        mock_fnruntime.reset_double_counter.side_effect = RuntimeError("mock exception")
        mock_fnruntime.increase_double_counter.side_effect = RuntimeError("mock exception")
        mock_fnruntime.report_gauge.side_effect = RuntimeError("mock exception")
        mock_fnruntime.set_alarm.side_effect = RuntimeError("mock exception")
        mock_fnruntime.get_value_double_counter.return_value = 0.1
        mock_fnruntime.get_value_uint64_counter.return_value = 1

        mock_fnruntime.peek_object_ref_stream.return_value = "peek_object_ref_stream"

        mock_fnruntime.generate_group_name.return_value = "generate_group_name"
        mock_fnruntime.get_instances.return_value = ["instance1"]
        mock_fnruntime.resources.return_value = "resources"
        mock_fnruntime.get_function_group_context.return_value = "get_function_group_context"
        mock_fnruntime.get_instance_by_name.return_value = "get_instance_by_name"
        mock_fnruntime.resources.return_value = "resources"
        mock_fnruntime.is_local_instances.return_value = True
        mock_fnruntime.create_buffer.return_value = ("buffer_id_1", None)
        mock_fnruntime.get_buffer.return_value = None
        mock_fnruntime.accelerate.return_value = ["handle"]
        mock_fnruntime.add_return_object.return_value = None
        self.runtime.libruntime = mock_fnruntime

    def test_runtime_error(self):
        with self.assertRaises(RuntimeError):
            self.runtime.get(["id1", "id2"], -2, False)

    def test_timeout_ms_set_to_minus_one(self):
        Serialization._instance[Serialization._cls] = Mock()
        Serialization().deserialize.return_value = ["data1", "data2"]
        self.runtime.get(["id1", "id2"], -1, False)
        self.assertEqual(self.runtime.libruntime.get.call_args[0][1], -1)

    def test_get_when_allow_partial_is_true(self):
        Serialization._instance[Serialization._cls] = Mock()
        Serialization().deserialize.return_value = ["data1", "data2"]
        result = self.runtime.get(["id1", "id2", "id3"], -1, True)
        self.assertEqual(result, ["data1", "data2"])

    def test_save_load_state_with_timeout_minus_one(self):
        self.runtime.save_state(-1)
        self.runtime.load_state(-1)
        self.runtime.libruntime.save_state.assert_called_once_with(-1)
        self.runtime.libruntime.load_state.assert_called_once_with(-1)

    def test_save_load_state_with_timeout(self):
        self.runtime.save_state(10)
        self.runtime.load_state(10)
        self.runtime.libruntime.save_state.assert_called_once_with(10 * 1000)
        self.runtime.libruntime.load_state.assert_called_once_with(10 * 1000)

    def test_stream(self):
        with self.assertRaises(RuntimeError):
            cfg = yr.ProducerConfig()
            cfg.max_stream_size = -1
            self.runtime.create_stream_producer("stream", cfg)
        with self.assertRaises(RuntimeError):
            cfg = yr.ProducerConfig()
            cfg.retain_for_num_consumers = -1
            self.runtime.create_stream_producer("stream", cfg)
        with self.assertRaises(RuntimeError):
            cfg = yr.ProducerConfig()
            cfg.reserve_size = -1
            self.runtime.create_stream_producer("stream", cfg)

        self.assertEqual(self.runtime.create_stream_producer("", yr.ProducerConfig()), "producer")
        self.assertEqual(self.runtime.create_stream_consumer("", yr.ProducerConfig()), "consumer")

        self.assertEqual(self.runtime.query_global_producers_num(""), 10)
        self.assertEqual(self.runtime.query_global_consumers_num(""), 10)

        with self.assertRaises(RuntimeError):
            self.runtime.delete_stream("")

    def test_instance(self):
        self.assertEqual(self.runtime.get_real_instance_id(""), "get_real_instance_id")
        self.assertEqual(self.runtime.get_instances("", "")[0], "instance1")
        with self.assertRaises(RuntimeError):
            self.runtime.save_real_instance_id("", True)

        with self.assertRaises(RuntimeError):
            self.runtime.terminate_instance("")

        with self.assertRaises(RuntimeError):
            self.runtime.terminate_instance_sync("")

        with self.assertRaises(RuntimeError):
            self.runtime.terminate_group("")

        with self.assertRaises(RuntimeError):
            self.runtime.exit()

        with self.assertRaises(RuntimeError):
            self.runtime.receive_request_loop()

    def test_get_put_wait(self):
        param = CreateParam()
        self.assertEqual(self.runtime.put("abc", param), "putobj")
        ready, _ = self.runtime.wait(["obj1"], 1, 10)
        self.assertEqual(len(ready), 1, ready)

    def test_kv(self):
        with self.assertRaises(RuntimeError):
            self.runtime.kv_write("key", b"abc")

        self.assertEqual(self.runtime.kv_read("", 1)[0], "value")

        with self.assertRaises(RuntimeError):
            self.runtime.kv_del("key")

    def test_reference(self):
        with self.assertRaises(RuntimeError):
            self.runtime.increase_global_reference([])
            self.runtime.decrease_global_reference([])

    def test_invoke(self):
        with self.assertRaises(RuntimeError):
            meta = FunctionMeta()
            meta.apiType = ApiType.Faas
            args = [ObjectRef(""), "abc"]
            self.runtime.invoke_by_name(meta, args, None, 1, None)

        self.assertEqual(self.runtime.create_instance(None, [], None), "instance")
        self.assertEqual(self.runtime.invoke_instance(None, "", [], None, 1)[0], "invoke_instance")

        with self.assertRaises(RuntimeError):
            self.runtime.cancel([], False, False)
        self.assertTrue(self.runtime.is_object_existing_in_local(""))

        self.assertEqual(self.runtime.resources(), "resources")
        self.assertEqual(self.runtime.get_function_group_context(), "get_function_group_context")
        self.assertEqual(self.runtime.get_instance_by_name(
            "", "", 1), "get_instance_by_name")

    def test_metrics(self):
        with self.assertRaises(RuntimeError):
            self.runtime.set_uint64_counter(None)

        with self.assertRaises(RuntimeError):
            self.runtime.reset_uint64_counter(None)

        with self.assertRaises(RuntimeError):
            self.runtime.increase_uint64_counter(None)

        with self.assertRaises(RuntimeError):
            self.runtime.set_double_counter(None)

        with self.assertRaises(RuntimeError):
            self.runtime.reset_double_counter(None)

        with self.assertRaises(RuntimeError):
            self.runtime.increase_double_counter(None)

        with self.assertRaises(RuntimeError):
            self.runtime.report_gauge(None)

        with self.assertRaises(RuntimeError):
            self.runtime.set_alarm("", "", None)

        self.assertEqual(self.runtime.get_value_uint64_counter(None), 1)
        self.assertEqual(self.runtime.get_value_double_counter(None), 0.1)

    def test_generator(self):
        self.assertEqual(self.runtime.peek_object_ref_stream(""), "peek_object_ref_stream")
        self.assertEqual(self.runtime.generate_group_name(), "generate_group_name")

    def test_private(self):
        with self.assertRaises(RuntimeError):
            self.runtime.finalize()
            self.runtime._check_init()

    def test_is_local_instances(self):
        self.assertEqual(self.runtime.is_local_instances([""]), True)

    def test_create_buffer(self):
        self.assertEqual(self.runtime.create_buffer(0), ("buffer_id_1", None))

    def test_get_buffer(self):
        self.assertEqual(self.runtime.get_buffer("obj"), None)

    def test_accelerate(self):
        self.assertEqual(self.runtime.accelerate(["instance"], None), ["handle"])

    def test_add_return_object(self):
        self.assertEqual(self.runtime.add_return_object("obj"), None)

    def test_set_get_callback(self):
        Serialization().deserialize.return_value = "data"

        self.runtime.set_get_callback("object_id", self.call_back)
        self.call_back.assert_called_once_with('data')


if __name__ == "__main__":
    unittest.main()
