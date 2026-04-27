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
from unittest.mock import Mock, patch


from yr.config_manager import ConfigManager
from yr.config import Config
from yr import Gauge, UInt64Counter, DoubleCounter


class TestMetrics(TestCase):
    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_gauge(self,  get_runtime):
        mock_runtime = Mock()
        mock_runtime.report_gauge.side_effect = RuntimeError("mock exception")
        mock_runtime.increase_global_reference.return_value = None
        get_runtime.return_value = mock_runtime

        data = Gauge("userFuncTime", "user function cost time", "ms", {"runtime": "runtime1"})
        with self.assertRaises(ValueError):
            data.add_labels([])

        with self.assertRaises(ValueError):
            data.add_labels({"*key": "value"})

        with self.assertRaises(ValueError):
            data.add_labels({"1key": "value"})

        with self.assertRaises(ValueError):
            data.add_labels({"key": 1})

        with self.assertRaises(ValueError):
            data.add_labels({})

        data.add_labels({"instance": "instance2"})

        with self.assertRaises(ValueError):
            ConfigManager().init(Config(is_driver=True))
            data.set(1)

        with self.assertRaises(RuntimeError):
            ConfigManager().init(Config(is_driver=False, ds_address="127.0.0.1:31222"))
            data.set(1)

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_uint64_counter(self,  get_runtime):
        mock_runtime = Mock()
        mock_runtime.set_uint64_counter.side_effect = RuntimeError("mock exception")
        mock_runtime.reset_uint64_counter.side_effect = RuntimeError("mock exception")
        mock_runtime.increase_uint64_counter.side_effect = RuntimeError("mock exception")
        mock_runtime.get_value_uint64_counter.return_value = 10
        get_runtime.return_value = mock_runtime

        data = UInt64Counter("userFuncTime", "user function cost time", "ms", {"runtime": "runtime1"})

        with self.assertRaises(ValueError):
            data.add_labels({})

        with self.assertRaises(RuntimeError):
            ConfigManager().init(Config(is_driver=False, ds_address="127.0.0.1:31222"))
            data.set(1)

        with self.assertRaises(ValueError):
            ConfigManager().init(Config(is_driver=True))
            data.reset()

        with self.assertRaises(RuntimeError):
            ConfigManager().init(Config(is_driver=False, ds_address="127.0.0.1:31222"))
            data.reset()

        with self.assertRaises(ValueError):
            ConfigManager().init(Config(is_driver=True))
            data.increase(1)

        with self.assertRaises(RuntimeError):
            ConfigManager().init(Config(is_driver=False, ds_address="127.0.0.1:31222"))
            data.increase(1)

        with self.assertRaises(ValueError):
            ConfigManager().init(Config(is_driver=True))
            data.get_value()

        ConfigManager().init(Config(is_driver=False, ds_address="127.0.0.1:31222"))
        self.assertEqual(data.get_value(), 10)

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_double_counter(self,  get_runtime):
        mock_runtime = Mock()
        mock_runtime.set_double_counter.side_effect = RuntimeError("mock exception")
        mock_runtime.reset_double_counter.side_effect = RuntimeError("mock exception")
        mock_runtime.increase_double_counter.side_effect = RuntimeError("mock exception")
        mock_runtime.get_value_double_counter.return_value = 10.0
        get_runtime.return_value = mock_runtime

        data = DoubleCounter("userFuncTime", "user function cost time", "ms", {"runtime": "runtime1"})

        with self.assertRaises(ValueError):
            data.add_labels({})

        with self.assertRaises(RuntimeError):
            ConfigManager().init(Config(is_driver=False, ds_address="127.0.0.1:31222"))
            data.set(1)

        with self.assertRaises(ValueError):
            ConfigManager().init(Config(is_driver=True))
            data.reset()

        with self.assertRaises(RuntimeError):
            ConfigManager().init(Config(is_driver=False, ds_address="127.0.0.1:31222"))
            data.reset()

        with self.assertRaises(ValueError):
            ConfigManager().init(Config(is_driver=True))
            data.increase(1)

        with self.assertRaises(RuntimeError):
            ConfigManager().init(Config(is_driver=False, ds_address="127.0.0.1:31222"))
            data.increase(1)

        with self.assertRaises(ValueError):
            ConfigManager().init(Config(is_driver=True))
            data.get_value()

        ConfigManager().init(Config(is_driver=False, ds_address="127.0.0.1:31222"))
        self.assertEqual(data.get_value(), 10)


if __name__ == '__main__':
    main()
