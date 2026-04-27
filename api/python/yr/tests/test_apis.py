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
from unittest.mock import patch, Mock

from concurrent.futures import Future
import cloudpickle
import pytest

import yr
from yr import exception
from yr import fcc
from yr.config_manager import ConfigManager
from yr.config import FunctionGroupOptions
from yr.decorator import function_proxy, instance_proxy
from yr.base_runtime import AlarmInfo
from yr.object_ref import ObjectRef
from yr.config import InvokeOptions


@yr.invoke(return_nums=0)
def get():
    """get"""
    return 1


@yr.instance
class Counter:
    """Counter"""
    @yr.method(return_nums=0)
    def get(self):
        """get"""
        return 1


class TestApi(unittest.TestCase):
    def setUp(self) -> None:
        pass

    def test_yr_init_failed_when_input_invalid_address(self):
        conf = yr.Config()
        conf.function_id = "sn:cn:yrk:default:function:0-yr-test-config-init:$latest"
        conf.server_address = "127.0.0.1:11111"
        conf.in_cluster = False
        with self.assertRaises(ValueError):
            yr.init(conf)

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    @patch("yr.apis.is_initialized")
    def test_wait_when_input_list_return_in_order(self, is_initialized,  get_runtime):
        mock_runtime = Mock()
        mock_runtime.wait.return_value = (["3", "2", "1"], [])
        mock_runtime.increase_global_reference.return_value = None
        mock_runtime.decrease_global_reference.return_value = None
        get_runtime.return_value = mock_runtime
        is_initialized.return_value = True
        refs = [yr.object_ref.ObjectRef("1"), yr.object_ref.ObjectRef("2"), yr.object_ref.ObjectRef("3")]
        ready_refs, _ = yr.wait(refs)
        assert ready_refs == refs

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    @patch("yr.apis.is_initialized")
    def test_invoke_when_return_nums_is_0(self, is_initialized, get_runtime):
        mock_runtime = Mock()
        mock_runtime.invoke_by_name.return_value = ["1"]
        mock_runtime.create_instance.return_value = "1"
        mock_runtime.invoke_instance.return_value = ["1"]
        mock_runtime.put.return_value = "1"
        mock_runtime.put_serialized.return_value = "serialized-code-id"
        mock_runtime.increase_global_reference.return_value = None
        mock_runtime.decrease_global_reference.return_value = None
        get_runtime.return_value = mock_runtime
        is_initialized.return_value = True

        assert get.invoke() is None

        c = Counter.invoke()
        assert c.get.invoke() is None

    def test_yr_init_failed_when_input_invaild_function_id(self):
        conf = yr.Config()
        conf.function_id = "111"
        conf.in_cluster = False
        with pytest.raises(ValueError):
            yr.init(conf)

    def test_affinity(self):
        """
        Test the init method of Affinity class.
        """
        affinity_kind = yr.AffinityKind.INSTANCE
        affinity_type = yr.AffinityType.PREFERRED
        label_operators = [
            yr.LabelOperator(yr.OperatorType.LABEL_IN, "key", ["value1", "value2"])
        ]
        affinity = yr.Affinity(affinity_kind, affinity_type, label_operators)
        assert affinity_kind == affinity.affinity_kind
        assert affinity_type == affinity.affinity_type
        assert label_operators == affinity.label_operators

        affinity_scope = yr.AffinityScope.NODE
        affinity2 = yr.Affinity(affinity_kind, affinity_type, label_operators, affinity_scope)
        assert affinity_scope == affinity2.affinity_scope

    @patch("yr.apis.is_initialized")
    def test_cancel_with_invalid_value(self, is_initialized):
        is_initialized.return_value = True

        with pytest.raises(TypeError) as e:
            yr.cancel("aaa")
        print(e)
        assert e.value.__str__() == "obj_refs type error, actual: [<class 'str'>], element expect: <class 'ObjectRef'>"

        with pytest.raises(TypeError) as e:
            yr.cancel(["aaa"])
        assert e.value.__str__() == "obj_refs type error, actual: [<class 'str'>], element expect: <class 'ObjectRef'>"

    def test_double_counter(self):
        with self.assertRaises(ValueError):
            yr.init(yr.Config(is_driver=True))
        labels = {"key1": "value1", "key2": "value2"}
        d = yr.DoubleCounter(name="test", description='', unit="ms", labels=labels)
        d.add_labels({"key3": "value3"})
        with pytest.raises(ValueError) as e:
            d.set(100)
        assert e.value.__str__() == "double counter metrics set not support in driver"
        with self.assertRaises(ValueError):
            yr.init(yr.Config(is_driver=False))

        with pytest.raises(ValueError) as e:
            name = "*test"
            d = yr.DoubleCounter(name=name, description='', unit="ms", labels=labels)
            d.set(100)
        assert e.value.__str__() == "invalid metric name: *test"

        with pytest.raises(ValueError) as e:
            d = yr.DoubleCounter(name="test", description='', unit="ms", labels=labels)
            d.set("abc")
        assert e.value.__str__() == "could not convert string to float: 'abc'"

    def test_uint64_counter(self):
        with self.assertRaises(ValueError):
            yr.init(yr.Config(is_driver=True))
        labels = {"key1": "value1", "key2": "value2"}
        u = yr.UInt64Counter(name="test", description='', unit="ms", labels=labels)
        u.add_labels({"key3": "value3"})
        with pytest.raises(ValueError) as e:
            u.set(100)
        assert e.value.__str__() == "uint64 counter metrics set not support in driver"
        with self.assertRaises(ValueError):
            yr.init(yr.Config(is_driver=False))

        with pytest.raises(ValueError) as e:
            name = "*test"
            u = yr.UInt64Counter(name=name, description='', unit="ms", labels=labels)
            u.set(100)
        assert e.value.__str__() == "invalid metric name: *test"

        with pytest.raises(ValueError) as e:
            u = yr.UInt64Counter(name="test", description='', unit="ms", labels=labels)
            u.set("abc")
        assert e.value.__str__() == "invalid literal for int() with base 10: 'abc'"

    def test_gauge(self):
        with self.assertRaises(ValueError):
            yr.init(yr.Config(is_driver=True))
        labels = {"key1": "value1", "key2": "value2"}
        g = yr.Gauge(name="test", description='', unit="ms", labels=labels)
        g.add_labels({"key3": "value3"})
        with pytest.raises(ValueError) as e:
            g.set(100)
        assert e.value.__str__() == "gauge metrics report not support in driver"
        with self.assertRaises(ValueError):
            yr.init(yr.Config(is_driver=False))

        with pytest.raises(ValueError) as e:
            name = "*test"
            g = yr.Gauge(name=name, description='', unit="ms", labels=labels)
            g.set(100)
        assert e.value.__str__() == "invalid metric name: *test"

        with pytest.raises(ValueError) as e:
            g = yr.Gauge(name="test", description='', unit="ms", labels=labels)
            g.set("abc")
        assert e.value.__str__() == "could not convert string to float: 'abc'"

    def test_alarm(self):
        with self.assertRaises(ValueError):
            yr.init(yr.Config(is_driver=True))
        alarm_info = AlarmInfo()
        a = yr.Alarm(name="test", description='')
        with pytest.raises(ValueError) as e:
            a.set(alarm_info)
        assert e.value.__str__() == "alarm metrics set not support in driver"
        with self.assertRaises(ValueError):
            yr.init(yr.Config(is_driver=False))

        with pytest.raises(ValueError) as e:
            name = "*test"
            a = yr.Alarm(name=name, description='')
        assert e.value.__str__() == "invalid metric name: *test"

    def test_serialize_instance_ok(self):
        @yr.instance
        class Counter1:
            pass

        a = cloudpickle.dumps(Counter1)
        b = cloudpickle.loads(a)
        assert type(b) == type(Counter1)
        assert type(b.__user_class__) == type(Counter1.__user_class__)

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_finalize(self, get_runtime):
        mock_runtime = Mock()
        mock_runtime.finalize.return_value = None
        mock_runtime.receive_request_loop.return_value = None
        mock_runtime.exit().return_value = None
        get_runtime.return_value = mock_runtime
        yr.apis.set_initialized()
        self.assertTrue(yr.apis.is_initialized())
        yr.finalize()
        self.assertFalse(yr.apis.is_initialized())
        yr.apis.set_initialized()
        yr.exit()
        yr.finalize()
        yr.apis.receive_request_loop()
        self.assertFalse(yr.apis.is_initialized())

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    @patch("yr.apis.is_initialized")
    def test_get_put_wait(self, is_initialized, get_runtime):
        mock_runtime = Mock()
        mock_runtime.get.return_value = [1, 2]
        mock_runtime.wait.return_value = []
        get_runtime.return_value = mock_runtime
        is_initialized.return_value = True
        v = ObjectRef("test")
        with self.assertRaises(TypeError):
            yr.put(v)

        with self.assertRaises(ValueError):
            yr.get(v, -2, True)
        self.assertFalse(yr.get([]))

        res = yr.get([v, v])
        self.assertEqual(len(res), 2, len(res))

        with self.assertRaises(ValueError):
            yr.wait([v, v], 2)

        with self.assertRaises(ValueError):
            yr.wait([v], -1)

        with self.assertRaises(TypeError):
            yr.wait([v], "")

        with self.assertRaises(TypeError):
            yr.wait([v], 1, "")

        with self.assertRaises(ValueError):
            yr.wait([v], 1, -2)

        with self.assertRaises(ValueError):
            yr.wait([v], 1)

    def test_instance(self):

        with self.assertRaises(RuntimeError):
            @yr.instance(invoke_options=InvokeOptions())
            def hello():
                return ""
            hello()

        with self.assertRaises(TypeError):
            @yr.method(return_nums="")
            def hi():
                return ""
            hi()

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    @patch("yr.apis.is_initialized")
    def test_stream(self, is_initialized, get_runtime):
        mock_runtime = Mock()
        mock_runtime.create_stream_producer.return_value = "producer"
        mock_runtime.create_stream_consumer.return_value = "consummer"
        mock_runtime.query_global_producers_num.return_value = 10
        mock_runtime.query_global_consumers_num.return_value = 10
        mock_runtime.delete_stream.side_effect = RuntimeError("mock exception")
        mock_runtime.kv_write.side_effect = RuntimeError("mock exception")
        mock_runtime.kv_read.return_value = ["value"]
        mock_runtime.kv_del.side_effect = RuntimeError("mock exception")
        get_runtime.return_value = mock_runtime
        is_initialized.return_value = True

        self.assertEqual(yr.create_stream_producer("", None), "producer")

        self.assertEqual(yr.create_stream_consumer("", None), "consummer")
        self.assertEqual(yr.query_global_producers_num(""), 10)
        self.assertEqual(yr.query_global_consumers_num(""), 10)

        with self.assertRaises(RuntimeError):
            yr.delete_stream("")

        with self.assertRaises(RuntimeError):
            yr.kv_write("key", b"abc")

        self.assertEqual(yr.kv_read(""), "value")

        with self.assertRaises(RuntimeError):
            yr.kv_del("key")

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    @patch("yr.apis.is_initialized")
    def test_state(self, is_initialized, get_runtime):
        mock_runtime = Mock()
        mock_runtime.save_state.side_effect = RuntimeError("mock exception")
        mock_runtime.load_state.side_effect = RuntimeError("mock exception")
        get_runtime.return_value = mock_runtime
        is_initialized.return_value = True

        with self.assertRaises(RuntimeError):
            ConfigManager().local_mode = True
            yr.save_state()

        with self.assertRaises(RuntimeError):
            ConfigManager().local_mode = False
            yr.save_state()

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    @patch("yr.apis.is_initialized")
    def test_save_and_load_state(self, is_initialized, get_runtime):
        mock_runtime = Mock()
        mock_runtime.save_state.return_value = None
        mock_runtime.load_state.return_value = None
        get_runtime.return_value = mock_runtime
        is_initialized.return_value = True
        ConfigManager()._ConfigManager__in_cluster = True
        ConfigManager()._ConfigManager__is_driver = False
        ConfigManager().local_mode = False
        yr.save_state()
        self.assertEqual(yr.load_state(), None)

    @patch("yr.decorator.instance_proxy.get_instance_by_name")
    def test_get_instance(self, get_instance_by_name):
        get_instance_by_name.side_effect = RuntimeError("mock exception")

        with self.assertRaises(TypeError):
            yr.get_instance(1)

        with self.assertRaises(TypeError):
            yr.get_instance("instance1", 1, 1)

        with self.assertRaises(TypeError):
            yr.get_instance("instance1", "namespace", -2)

        with self.assertRaises(Exception):
            yr.get_instance("instance1", "namespace", 1)

    def tets_resources(self):
        with self.assertRaises(RuntimeError):
            ConfigManager().local_mode = True
            yr.resources()

    @patch("yr.decorator.instance_proxy.make_cpp_instance_creator")
    def test_class_cross_instance(self, make_cpp_instance_creator):
        make_cpp_instance_creator.return_value = ""
        urn = "sn:cn:yrk:default:function:0-yr-test-config-init:$latest"
        cpp_class = yr.apis.cpp_instance_class(
            "class", "factory", urn)

        with self.assertRaises(Exception):
            cpp_class.invoke()

        with self.assertRaises(Exception):
            cpp_class.Options()

        self.assertEqual(cpp_class.get_class_name(), "class")
        self.assertEqual(cpp_class.get_factory_name(), "factory")
        self.assertTrue("test" in cpp_class.get_function_key())

        cp = yr.apis.cpp_function("class", urn)
        self.assertEqual(cp.cross_language_info.function_name, "class", cp.cross_language_info)

        with self.assertRaises(AttributeError):
            yr.apis.go_function("class", urn)

        jp = yr.apis.java_function("class", "factory", urn)
        self.assertEqual(jp.cross_language_info.function_name, "factory", jp.cross_language_info)

        cp = yr.apis.cpp_instance_class_new(
            "class", "factory", urn)
        self.assertTrue(cp)

    @patch.object(yr.decorator.instance_proxy.InstanceCreator, "options")
    @patch.object(yr.decorator.function_proxy.FunctionProxy, "options")
    def test_create_function_group(self, mock_function_proxy_options, mock_instance_proxy_options):
        opt = FunctionGroupOptions()
        with self.assertRaises(ValueError):
            opt.scheduling_affinity_each_bundle_size = -1
            fcc.create_function_group(None, None, 1, opt)

        with self.assertRaises(ValueError):
            opt.scheduling_affinity_each_bundle_size = 10
            fcc.create_function_group(None, None, 1, opt)

        with self.assertRaises(ValueError):
            opt.scheduling_affinity_each_bundle_size = 3
            fcc.create_function_group(None, None, 10, opt)

        with self.assertRaises(ValueError):
            opt.timeout = -2
            opt.scheduling_affinity_each_bundle_size = 2
            fcc.create_function_group(None, None, 10, opt)

        mock_instance_options = Mock()
        mock_instance_proxy = Mock()
        mock_instance_proxy.get_function_group_handler.return_value = "instance"
        mock_instance_options.invoke.return_value = mock_instance_proxy
        mock_instance_proxy_options.return_value = mock_instance_options

        ip = instance_proxy.InstanceCreator()

        opt.timeout = None
        opt.scheduling_affinity_each_bundle_size = 2
        res = fcc.create_function_group(ip, (1, 2), 10, opt)
        self.assertEqual(res, "instance", res)

        def mock_invoke(a, b):
            return "function"

        class MockWrapperFuncProxy:
            def __init__(self):
                self.invoke = mock_invoke

        class MockFuncProxy:
            self.invoke = mock_invoke

            def set_function_group_size(a):
                return MockWrapperFuncProxy()

        mock_function_proxy_options.return_value = MockFuncProxy
        fp = function_proxy.FunctionProxy(None)
        res = fcc.create_function_group(fp, (1, 2), 10, opt)
        self.assertEqual(res, "function", res)

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    @patch("yr.apis.is_initialized")
    def test_resource_group(self, is_initialized, get_runtime):
        mock_runtime = Mock()
        mock_runtime.create_resource_group.return_value = None
        mock_runtime.remove_resource_group.return_value = None
        mock_runtime.wait_resource_group.return_value = None
        get_runtime.return_value = mock_runtime
        is_initialized.return_value = True
        rg = yr.create_resource_group([{"NPU": 1}, {"CPU": 2000, "Memory": 2000}], "rgname")
        self.assertTrue(rg)
        rg.wait()
        bundles = rg.bundle_specs
        assert bundles == [{"NPU": 1}, {"CPU": 2000, "Memory": 2000}]
        count = rg.bundle_count
        assert count == 2
        yr.remove_resource_group("rgname")
        with self.assertRaises(TypeError):
            rg = yr.create_resource_group(None, "rgname")
        with self.assertRaises(ValueError):
            rg = yr.create_resource_group([], "rgname")
        with self.assertRaises(TypeError):
            rg = yr.create_resource_group([{"NPU": 1}, {"CPU": 2000, "Memory": 2000}], [])


class TestException(unittest.TestCase):
    def test_cancelErr(self):
        err = exception.CancelError("task_1234")
        self.assertTrue("task_1234" in str(err))

    def test_invokeErr(self):
        origin_err = exception.YRInvokeError("origin_err", "some origin exception")
        self.assertTrue("some origin exception" in str(origin_err))
        err = exception.YRInvokeError(origin_err, "some new exception")
        self.assertTrue("some new exception" in str(err))

    def test_requestErr(self):
        err = exception.YRequestError(1001, "some exception", "request1")
        self.assertEqual(err.code, 1001)
        self.assertEqual(err.message, "some exception")
        self.assertTrue("request1" in str(err))

    def test_generator_finished(self):
        e = exception.GeneratorFinished(RuntimeError("some exception"))
        with self.assertRaises(Exception):
            raise e

    def test_deal_with_yr_error(self):
        origin_err = exception.YRInvokeError("origin_err", "some origin exception")
        err = exception.YRInvokeError(origin_err, "some new exception")
        f = Future()
        exception.deal_with_yr_error(f, err)
        self.assertTrue("some new exception" in str(f.exception()))
        f = Future()
        exception.deal_with_yr_error(f, RuntimeError("some exception"))
        with self.assertRaises(RuntimeError):
            f.result()

    def test_deal_with_error(self):
        f = Future()
        with self.assertRaises(Exception):
            exception.deal_with_error(f, 1001, b"some err")
            self.assertTrue("some err" in f.exception())
        class CustomError:
            def __init__(self):
                self.err = None

            def __str__(self):
                return f"msg:{err}"

            def set_msg(self, err):
                self.err = err

            def get_msg(self):
                return self.err

        err = CustomError()
        yr_error = exception.YRInvokeError(err, "custom error")
        self.assertIsNotNone(yr_error.origin_error())


if __name__ == "__main__":
    unittest.main()
