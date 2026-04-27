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

import os
import json
import logging
from logging import handlers
from yr.functionsdk import context
from yr.functionsdk import function
from yr.functionsdk import utils
from yr.functionsdk import logger as sdklogger
from yr.functionsdk import logger_manager
from unittest import TestCase, main
from unittest.mock import Mock, patch
import os
import logging
import queue



logger = logging.getLogger(__name__)


class TestFunctionSdk(TestCase):

    @patch("yr.log.get_logger")
    def test_context(self, mock_logger):
        mock_logger.return_value = logger

        delegate = {"environment": {"LD_LIBRARY_PATH": "default_path", "TEST_ENV_KEY": "test_env_value"},
                    "encrypted_user_data": {"LD_LIBRARY_PATH": "user_path", "TEST_USER_KEY": "test_user_value"}}
        os.environ["ENV_DELEGATE_DECRYPT"] = json.dumps(delegate)
        context_meta = {"funcMetaData": {"timeout": "3"}, "extendedMetaData": {"pre_stop": {"pre_stop_timeout": 10}}}
        context.load_context_meta(context_meta)
        env_user_data = os.environ.get("RUNTIME_USERDATA")
        user_data = json.loads(env_user_data)
        path = user_data.get("LD_LIBRARY_PATH", "")
        self.assertTrue("user_path" in path, path)
        self.assertTrue("test_env_value" in user_data.get("TEST_ENV_KEY", ""), user_data)
        self.assertTrue("test_user_value" in user_data.get("TEST_USER_KEY", ""), user_data)

        header = {"X-Request-Id": "12345"}
        invoke_context = context.init_context_invoke("invoke", header)
        self.assertEqual(invoke_context.get_trace_id(), "12345")
        self.assertEqual(invoke_context.getUserData("TEST_ENV_KEY"), "test_env_value")

    @patch("yr.log.get_logger")
    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_invoke(self, get_runtime, mock_logger):
        mock_logger.return_value = logger

        mock_runtime = Mock()
        mock_runtime.invoke_by_name.return_value = ["obj_abcd"]
        get_runtime.return_value = mock_runtime

        os.environ["RUNTIME_SERVICE_FUNC_VERSION"] = "0@faashello@hello"

        f = function.Function("hello")
        opt = function.InvokeOptions(cpu=200, memory=200)
        args = {"body": "test"}
        obj = f.options(opt).invoke(args)
        self.assertEqual(obj.id, "obj_abcd", obj.id)

        args = [1, 2, 3]
        with self.assertRaises(TypeError):
            f.options(opt).invoke(args)

        args = 'null'
        with self.assertRaises(ValueError):
            f.options(opt).invoke(args)

        args = 'abcd'
        with self.assertRaises(TypeError):
            f.options(opt).invoke(args)

        large_body = "a" * (1024 * 1024 * 7)
        args = {"body": large_body}
        args_str = json.dumps(args)
        with self.assertRaises(ValueError):
            f.options(opt).invoke(args_str)

        func_name = {}
        with self.assertRaises(TypeError):
            function.Function(func_name)

    @patch("yr.log.get_logger")
    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_invoke_alias(self, get_runtime, mock_logger):
        mock_logger.return_value = logger

        mock_runtime = Mock()
        mock_runtime.invoke_by_name.return_value = ["obj_abcd"]
        get_runtime.return_value = mock_runtime

        os.environ["RUNTIME_SERVICE_FUNC_VERSION"] = "0@faashello@hello"

        f = function.Function("hello:alias")
        opt = function.InvokeOptions(cpu=200, memory=200, alias_params={"a": "b"})
        args = {"body": "test"}
        obj = f.options(opt).invoke(args)
        self.assertEqual(obj.id, "obj_abcd", obj.id)

        args = [1, 2, 3]
        with self.assertRaises(TypeError):
            f.options(opt).invoke(args)

        args = 'null'
        with self.assertRaises(ValueError):
            f.options(opt).invoke(args)

        args = 'abcd'
        with self.assertRaises(TypeError):
            f.options(opt).invoke(args)

        large_body = "a" * (1024 * 1024 * 7)
        args = {"body": large_body}
        args_str = json.dumps(args)
        with self.assertRaises(ValueError):
            f.options(opt).invoke(args_str)

        func_name = {}
        with self.assertRaises(TypeError):
            function.Function(func_name)

    @patch("yr.log.get_logger")
    def test_private_function(self, mock_logger):
        mock_logger.return_value = logger

        func_name = "hello:latest"
        name, version = function._check_function_name(func_name)
        self.assertEqual(name, "hello", name)
        self.assertEqual(version, "latest", version)

        func_name = "hello:!alise"
        name, version = function._check_function_name(func_name)
        self.assertEqual(name, "hello", name)
        self.assertEqual(version, "!alise", version)

        large_body = "1" * (1024)
        func_name = f"hello:!{large_body}"
        with self.assertRaises(ValueError):
            function._check_function_name(func_name)
        
        del os.environ["RUNTIME_SERVICE_FUNC_VERSION"]
        with self.assertRaises(RuntimeError):
            function._get_service_name_from_env()

        os.environ["RUNTIME_SERVICE_FUNC_VERSION"] = "faashello"
        with self.assertRaises(ValueError):
            function._get_service_name_from_env()

        os.environ["RUNTIME_SERVICE_FUNC_VERSION"] = "0@faashello@hello"
        service_name = function._get_service_name_from_env()
        self.assertEqual(service_name, "faashello", service_name)

        reg = r'^[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$'
        with self.assertRaises(ValueError):
            function._check_reg_length("aaa", reg, 1)

        with self.assertRaises(ValueError):
            function._check_reg_length("_*aaa", reg, 10)

    def test_utils(self):
        class SampleClass:
            def __init__(self, value):
                self.value = value

        obj = SampleClass(10)
        expected = {"value": 10}
        self.assertEqual(utils.convert_obj_to_json(obj), expected)

        obj = {"key": "value"}
        expected = json.dumps(obj)
        self.assertEqual(utils.to_json_string(obj), expected)

        obj = {"b": 2, "a": 1}
        expected = json.dumps(obj, indent=4, sort_keys=True)
        self.assertEqual(utils.to_json_string(obj, indent=4, sort_keys=True), expected)

        request_id = utils.generate_request_id()
        self.assertTrue(request_id.startswith("task-"))
        self.assertEqual(len(request_id), len("task-") + 36)

        trace_id = utils.generate_trace_id()
        self.assertTrue(trace_id.startswith("trace-"))
        self.assertEqual(len(trace_id), len("trace-") + 36)

        trace_id = "test-id"
        utils.set_trace_id(trace_id)
        self.assertTrue("trace" in utils.get_trace_id())

        json_str = '{"key": "value"}'
        expected = {"key": "value"}
        self.assertEqual(utils.parse_json_data_to_dict(json_str), expected)

        json_bytes = b'{"key": "value"}'
        self.assertEqual(utils.parse_json_data_to_dict(json_bytes), expected)

        json_bytearray = bytearray(b'{"key": "value"}')
        self.assertEqual(utils.parse_json_data_to_dict(json_bytearray), expected)

        self.assertEqual(utils.parse_json_data_to_dict(""), {})
        self.assertEqual(utils.parse_json_data_to_dict({}), {})
        self.assertEqual(utils.parse_json_data_to_dict(None), {})

        with self.assertRaises(RuntimeError):
            utils.parse_json_data_to_dict("invalid json")

        obj = {"value": 10}
        expected = json.dumps(obj)
        self.assertEqual(utils.dump_data_to_json_str(obj), expected)

        expected = json.dumps({"value": 10})
        self.assertEqual(utils.dump_data_to_json_str(obj), expected)

        with self.assertRaises(RuntimeError):
            utils.dump_data_to_json_str(object())
    
    def test_utils_timeout(self):
        import time

        @utils.timeout(1, 1 )
        def test_timeout(t):
            time.sleep(t)
            print("finished sleep", t)
            return "hello"

        res = test_timeout(0)
        self.assertEqual(res, "hello", res)
        with self.assertRaises(TimeoutError):
            test_timeout(2)

    def test_sdklogger(self):
        print("start test get_user_function_logger 1")
        log = sdklogger.get_user_function_logger(logging.DEBUG)
        self.assertEqual(log.name, "user-function", log.name)

    def test_sdklogger_mgr(self):
        print("start test faaslog 1")
        log1 = logger_manager.Log(logging.DEBUG, "test")
        self.assertEqual(log1.msg, "test", log1.msg)

        print("start test faaslog")
        log2 = logger_manager.FaasLogger()
        q = queue.Queue(maxsize=10)
        log2.set_queue(q)
        log2.log(logging.DEBUG, "testlog")
        get_msg = q.get(block=False).msg
        self.assertEqual(get_msg, "testlog", get_msg)

        log2.debug("debugtest")
        get_msg = q.get(block=False).msg
        self.assertEqual(get_msg, "debugtest", get_msg)

        log2.error("errlog")
        get_msg = q.get(block=False).msg
        self.assertEqual(get_msg, "errlog", get_msg)

        log2.warning("warning")
        get_msg = q.get(block=False).msg
        self.assertEqual(get_msg, "warning", get_msg)

        log2.info("infolog")
        get_msg = q.get(block=False).msg
        self.assertEqual(get_msg, "infolog", get_msg)
    
    def test_UserLogManager(self):
        user_logger = logger_manager.UserLogManager()
        cfg = {"log_level":logging.DEBUG, "tenant_id":"test_tenant", "function_name":"", "version":"", "package":"", "stream":"", "instance_id":""}
        user_logger.load_logger_config(cfg)
        user_logger.set_stage("test_stage")
        #user_logger.start_user_log()
        user_logger.insert_start_log()
        user_logger.insert_end_log()
        self.assertEqual(user_logger.tenant_id, "test_tenant", user_logger.tenant_id)
        user_logger.shutdown()


if __name__ == "__main__":
    main()