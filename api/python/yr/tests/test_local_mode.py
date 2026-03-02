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
from unittest import TestCase, main
import time
import concurrent.futures

import yr
from yr.object_ref import ObjectRef
from yr.exception import YRInvokeError
from yr.local_mode.local_mode_runtime import LocalModeRuntime
from yr.local_mode import local_client, instance_manager
from yr.local_mode.instance import Resource, Instance
from yr.local_mode.task_spec import TaskSpec
from yr.local_mode.local_object_store import LocalObjectStore
from yr.base_runtime import SetParam


class Mock(object):
    pass


class TestApi(TestCase):
    def setUp(self) -> None:
        pass

    def test_local_mode_base(self):
        @yr.invoke
        def func(x):
            return x

        yr.init(yr.Config(local_mode=True, log_level="DEBUG"))
        assert yr.get(func.invoke(1)) == 1
        assert yr.get(func.invoke(func.invoke(1))) == 1

        @yr.instance
        class Counter:
            cnt = 0

            def add(self):
                self.cnt += 1
                return self.cnt

            def get(self):
                return self.cnt

            def delay(self):
                time.sleep(1)
                return self.cnt

        c = Counter.invoke()
        assert yr.get(c.get.invoke()) == 0
        assert yr.get(c.add.invoke()) == 1

        obj_id = c.add.invoke()
        ready, _ = yr.wait(obj_id, 1, 2)
        self.assertTrue(len(ready) == 1, len(ready))
        assert yr.get(ready[0]) == 2

        obj_id = c.delay.invoke()
        yr.cancel(obj_id)

        obj_id = ObjectRef(object_id="noid")

        with self.assertRaises(RuntimeError):
            yr.cancel(obj_id)

        obj = func.invoke(1)

        def cb():
            return
        obj.on_complete(cb)
        self.assertEqual(yr.get(obj), 1)

    def test_local_mode_runtime(self):
        lr = LocalModeRuntime()
        with self.assertRaises(RuntimeError):
            lr.get([1], -2, True)

        with self.assertRaises(Exception):
            lr.get([YRInvokeError(RuntimeError("mock"), "exception")], -2, True)

        with self.assertRaises(RuntimeError):
            lr.set_get_callback("no", None)

        key = "key"
        value = b"value"
        param = SetParam()
        lr.kv_write(key, value, param)
        get_value = lr.kv_read(key)
        self.assertEqual(get_value, value, get_value)

        lr.kv_del(key)
        get_value = lr.kv_read(key)
        self.assertEqual(get_value, None, get_value)

        lr.increase_global_reference([])
        lr.decrease_global_reference([])

        with self.assertRaises(RuntimeError):
            lr.terminate_group("group")

        with self.assertRaises(RuntimeError):
            lr.exit()

        lr.finalize()

        with self.assertRaises(RuntimeError):
            lr.create_stream_producer("stream", None)

        with self.assertRaises(RuntimeError):
            lr.create_stream_consumer("stream", None)

        with self.assertRaises(RuntimeError):
            lr.delete_stream("stream")

        with self.assertRaises(RuntimeError):
            lr.query_global_producers_num("stream")

        with self.assertRaises(RuntimeError):
            lr.query_global_consumers_num("stream")

        with self.assertRaises(RuntimeError):
            lr.save_state(1)

        with self.assertRaises(RuntimeError):
            lr.load_state(1)

        real_ins = lr.get_real_instance_id("real_ins")
        self.assertEqual(real_ins, "real_ins", real_ins)

        self.assertFalse(lr.is_object_existing_in_local("noins"))

        with self.assertRaises(RuntimeError):
            lr.set_uint64_counter(None)

        with self.assertRaises(RuntimeError):
            lr.reset_uint64_counter(None)

        with self.assertRaises(RuntimeError):
            lr.get_value_uint64_counter(None)

        with self.assertRaises(RuntimeError):
            lr.set_double_counter(None)

        with self.assertRaises(RuntimeError):
            lr.reset_double_counter(None)

        with self.assertRaises(RuntimeError):
            lr.increase_double_counter(None)

        with self.assertRaises(RuntimeError):
            lr.get_value_double_counter(None)

        with self.assertRaises(RuntimeError):
            lr.report_gauge(None)

        with self.assertRaises(RuntimeError):
            lr.set_alarm("", "", None)

        with self.assertRaises(RuntimeError):
            lr.generate_group_name()

        with self.assertRaises(RuntimeError):
            lr.get_instances("", "")

        with self.assertRaises(RuntimeError):
            lr.get_function_group_context()

    def test_local_client(self):
        client = local_client.LocalClient()
        key = client.save_state("state")
        value = client.load_state(key)
        self.assertEqual(value, "state")

        spec = TaskSpec(task_id="task1234",
                        invoke_type=1,
                        future=None,
                        object_ids=["instance1234"],
                        trace_id="trace1234",
                        )
        instance_id = client.create(spec)
        self.assertEqual(instance_id, "instance1234", instance_id)

        with self.assertRaises(RuntimeError):
            spec.instance_id = "noinstanceid"
            client.invoke(spec)

        client.create(spec)
        client.exit()
        client.clear()
        instance_id = client.create(spec)
        self.assertEqual(instance_id, "", instance_id)

    def test_local_object_store(self):
        key1 = "key1"
        key2 = "key2"
        value = "value"
        LocalObjectStore().put(key1, value)
        LocalObjectStore().put(key2, value)
        get_value1 = LocalObjectStore().get(key1)
        self.assertEqual(get_value1, value, get_value1)

        values = LocalObjectStore().get([key1, key2])
        self.assertEqual(len(values), 2, len(values))
        self.assertEqual(values[1], value, values[1])

        future = concurrent.futures.Future()
        LocalObjectStore().set_return_obj(key1, future)
        LocalObjectStore().set_result(key1, "future result")
        get_future = LocalObjectStore().get_future(key1)

        result = get_future.result()
        self.assertEqual(result, "future result", result)

        future = concurrent.futures.Future()
        LocalObjectStore().set_return_obj(key2, future)
        LocalObjectStore().set_exception(key2, "future exception")
        get_future = LocalObjectStore().get_future(key2)

        with self.assertRaises(Exception):
            get_future.result()

        key3 = "key3"
        with self.assertRaises(RuntimeError):
            LocalObjectStore().add_done_callback(key3, None)

        LocalObjectStore().put(key3, value)

        self.assertTrue(LocalObjectStore().is_object_existing_in_local(key3))

        def cb(a):
            raise TypeError("test callback exception")

        with self.assertRaises(TypeError):
            LocalObjectStore().add_done_callback(key3, cb)

        key4 = "key4"
        new_future = concurrent.futures.Future()
        LocalObjectStore().set_return_obj(key4, new_future)
        LocalObjectStore().add_done_callback(key4, cb)

        LocalObjectStore().set_result(key4, "future result")
        new_future.result()

        LocalObjectStore().increase_global_reference(["obj1"])
        LocalObjectStore().decrease_global_reference(["obj1"])
        LocalObjectStore().clear()

    def test_instance_manager(self):
        future = concurrent.futures.Future()
        exception = RuntimeError("test exception")
        future.set_exception(exception)
        instance_manager.print_instance_create_error(future)

        instance_manager.warning_if_failed(future, "failed")

        client = local_client.LocalClient()
        ins_mgr = instance_manager.InstanceManager(None, client, 0.1)
        res = Resource()
        self.assertEqual(ins_mgr.get_last_failed_reason(res), None)
        success, _ = ins_mgr.check_last_failed_reason(res)
        self.assertTrue(success)

        task = TaskSpec(task_id="task1234",
                        invoke_type=1,
                        future=None,
                        object_ids=["instance1234"],
                        trace_id="trace1234",
                        )
        instance_id, _ = ins_mgr.scale_out(task, res)
        self.assertIn("yr-api-obj", instance_id)
        get_ins = ins_mgr.get_instances(res)
        self.assertEqual(len(get_ins), 1, len(get_ins))

        ins_mgr.kill_instance(instance_id)
        get_ins = ins_mgr.get_instances(res)
        self.assertEqual(len(get_ins), 0, len(get_ins))


if __name__ == "__main__":
    main()
