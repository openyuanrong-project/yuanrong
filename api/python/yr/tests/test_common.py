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

import logging
import json
from unittest import TestCase, main

from yr.common import utils

logger = logging.getLogger(__name__)


class TestDecorator(TestCase):

    def test_utils(self):
        self.assertFalse(utils.validate_ip("12345"))
        self.assertTrue(utils.validate_ip("127.0.0.1"))

        with self.assertRaises(ValueError):
            utils.validate_address("12345")

        with self.assertRaises(ValueError):
            utils.validate_address("127.0.0.1:port")

        with self.assertRaises(ValueError):
            utils.validate_address("127.0.0.1:65537")

        ip, port = utils.validate_address("127.0.0.1:65531", True)
        self.assertEqual(ip, "127.0.0.1", ip)
        self.assertEqual(port, 65531, port)

        with self.assertRaises(ValueError):
            utils.validate_address("a127.0.0.1:65531")

        id = utils.generate_runtime_id()
        self.assertEqual(len(id), 8, len(id))

        id = utils.generate_task_id_with_serial_num("runtime_id", 1)
        self.assertTrue("runtime_id" in id, id)

        request_id = "job-724755d1-runtime-11000000-0000-4000-b483-a87b3d562767-00000040aa6c-x"

        parts = request_id.split("-")
        print("!!!!!!!!!!!!!!", parts[-1])
        self.assertFalse(utils.check_request_id_in_order(request_id))

        request_id = "job-724755d1-runtime-11000000-0000-4000-b483-a87b3d562767-00000040aa6c-1"
        self.assertTrue(utils.check_request_id_in_order(request_id))
        

        self.assertEqual(utils.extract_serial_num("task-1234"), 1234)
        self.assertEqual(utils.extract_runtime_id("task-runtime_id-1234-x"), "1234")
        self.assertTrue("trace" in utils.generate_trace_id("1234"))

        import math
        self.assertEqual(utils.get_module_name(math.sin), "math")

        class TestClass:
            pass

        TestClass.__module__ = "__main__"
        obj = TestClass()
        self.assertEqual(utils.get_module_name(obj), "__main__")

        test_cases = [
            (b"Hello, World!", "48656c6c6f2c20576f726c6421"),
            (b"", ""),
            (b"\x00\x01\x02", "000102"),
        ]
        for value, expected in test_cases:
            self.assertEqual(utils.binary_to_hex(value), expected)

        with self.assertRaises(TypeError):
            utils.binary_to_hex("Hello, World!")

        test_cases = [
            ("48656c6c6f2c20576f726c6421", b"Hello, World!"),
            ("", b""),
            ("000102", b"\x00\x01\x02"),
        ]
        for hex_id, expected in test_cases:
            self.assertEqual(utils.hex_to_binary(hex_id), expected)

        res = json.loads(utils.package_args("hello", "world"))
        self.assertTrue("args" in res)

        res = utils.make_cross_language_args([1], {"name":"test"})
        self.assertEqual(len(res), 3, len(res))

        obj_des = utils.ObjectDescriptor()
        data = {"moduleName": "test_module_name"}
        data_str = json.dumps(data)
        data_bytes = data_str.encode()
        obj_des = obj_des.parse(data)
        obj_dict = obj_des.to_dict()
        self.assertEqual(obj_dict.get("moduleName"), "test_module_name", obj_dict)

        obj_des = obj_des.parse(data_str)
        obj_dict = obj_des.to_dict()
        self.assertEqual(obj_dict.get("moduleName"), "test_module_name", obj_dict)

        obj_des = obj_des.parse(data_bytes)
        obj_dict = obj_des.to_dict()
        self.assertEqual(obj_dict.get("moduleName"), "test_module_name", obj_dict)

        class MyClass:
            function_name: str = "test"

        obj = MyClass()

        with self.assertRaises(Exception):
            utils.to_json_string(obj, indent=4, sort_keys=True)

        with self.assertRaises(RuntimeError):
            utils.to_json_string(1)

        env_value = utils.get_environment_variable("TEST_ENV_KEY", "default_value")

        self.assertEqual(env_value, "default_value", env_value)

        with self.assertRaises(RuntimeError):
            utils.get_environment_variable("TEST_ENV_KEY")

        with self.assertRaises(TypeError):
            utils.Validator.check_args_types({})

        with self.assertRaises(TypeError):
            utils.Validator.check_args_types([1])

        with self.assertRaises(TypeError):
            utils.Validator.check_args_types([[1]])

        with self.assertRaises(TypeError):
            utils.Validator.check_args_types([["time_out", 1, str]])

        utils.Validator.check_args_types([["time_out", 1, int]])

        with self.assertRaises(TypeError):
            utils.Validator.check_key_exists([], [])

        with self.assertRaises(TypeError):
            utils.Validator.check_key_exists({}, {})

        with self.assertRaises(TypeError):
            utils.Validator.check_key_exists({"name": "test"}, ["no_key"])

        res = utils.Validator.check_key_exists({"name": "test"}, ["name"])

        self.assertEqual(res[0], "test", res)

        with self.assertRaises(TypeError):
            utils.Validator.check_timeout([])


if __name__ == "__main__":
    main()
