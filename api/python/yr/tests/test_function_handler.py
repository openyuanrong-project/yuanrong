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
import logging
import inspect
from unittest import TestCase, main
from unittest.mock import Mock, patch
from yr.executor.function_handler import FunctionHandler
from yr.executor.instance_manager import InstanceManager
from yr.code_manager import CodeManager
from yr.serialization import Serialization
from yr.err_type import ErrorCode
from yr.libruntime_pb2 import FunctionMeta, LanguageType, InvokeType, ApiType

logger = logging.getLogger(__name__)


class TestFunctionExecutor(TestCase):
    def setUp(self):
        self.runtime_patcher = patch('yr.runtime_holder.global_runtime.get_runtime')
        self.mock_get_runtime = self.runtime_patcher.start()
        self.mock_get_runtime.return_value = Mock()
        self.handler = FunctionHandler()
        self.function_meta = FunctionMeta(applicationName="mytest",
                                          moduleName="test.py",
                                          className="counter",
                                          functionName="add",
                                          language=LanguageType.Python,
                                          codeID="123456",
                                          apiType=ApiType.Function,
                                          signature="",
                                          name="",
                                          ns="",
                                          initializerCodeID="",
                                          isGenerator=False,
                                          isAsync=False)

    def tearDown(self):
        manager = InstanceManager()
        manager.init(None)
        manager.class_code = None
        manager.is_async = False
        gc.collect()
        self.runtime_patcher.stop()

    @patch("yr.log.get_logger")
    @patch.object(CodeManager(), 'load_code')
    def test_execute_function(self, mock_load_code, mock_logger):
        mock_logger.return_value = logger

        class Counter:
            def __init__(self):
                self.value = 0

            def add(self, number):
                self.value = + number
                return self.value

            def get(self):
                return self.value

            def get_exception(self):
                raise RuntimeError("test user exception")

            def __yr_shutdown__(self, timeout):
                return

        mock_load_code.return_value = Counter

        Serialization._instance[Serialization._cls] = Mock()
        Serialization().deserialize.return_value = []

        # case1 CreateInstance
        _, err = self.handler.execute_function(self.function_meta, [], InvokeType.CreateInstance, 0, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)

        # case2 InvokeFunction
        # case2.1 InvokeFunction with args
        self.function_meta.functionName = 'add'
        Serialization().deserialize.return_value = [b"__YR_PLACEHOLDER__", 10]
        res, err = self.handler.execute_function(self.function_meta, [], InvokeType.InvokeFunction, 1, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)
        self.assertEqual(res[0], 10, res[0])

        # case2.2 InvokeFunction without args
        self.function_meta.functionName = 'get'
        Serialization().deserialize.return_value = []
        res, err = self.handler.execute_function(self.function_meta, [], InvokeType.InvokeFunction, 1, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)
        self.assertEqual(res[0], 10, res[0])

        # case2.3 InvokeFunction get exception
        self.function_meta.functionName = 'nothisfunction'
        Serialization().deserialize.return_value = []
        _, err = self.handler.execute_function(self.function_meta, [], InvokeType.InvokeFunction, 1, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_USER_FUNCTION_EXCEPTION, err.error_code)

        # case2.4 InvokeFunction without function
        self.function_meta.functionName = 'get_exception'
        Serialization().deserialize.return_value = []
        _, err = self.handler.execute_function(self.function_meta, [], InvokeType.InvokeFunction, 1, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_USER_FUNCTION_EXCEPTION, err.error_code)
        self.assertTrue("test user exception" in err.msg, err.error_code)

        # case3 CreateInstanceStateless
        res, err = self.handler.execute_function(self.function_meta, [], InvokeType.CreateInstanceStateless, 0, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)

        # case4 InvokeFunctionStateless
        def get():
            return "hello world"

        mock_load_code.return_value = get
        res, err = self.handler.execute_function(self.function_meta, [], InvokeType.InvokeFunctionStateless, 1, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)
        self.assertEqual(res[0], "hello world", res[0])

        # case5 GetNamedInstanceMeta
        _, err = self.handler.execute_function(self.function_meta, [], InvokeType.GetNamedInstanceMeta, 1, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)

        # case6 other invokeType
        _, err = self.handler.execute_function(self.function_meta, [], 111, 1, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_EXTENSION_META_ERROR, err.error_code)

        err = self.handler.shutdown(10)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)

    @patch("yr.log.get_logger")
    @patch.object(CodeManager(), 'load_code')
    def test_execute_async_function(self, mock_load_code, mock_logger):
        mock_logger.return_value = logger

        class Counter:
            def __init__(self):
                self.value = 0

            async def async_get(self):
                await asyncio.sleep(0)
                return self.value

            def sync_get(self):
                return self.value

        mock_load_code.return_value = Counter

        Serialization._instance[Serialization._cls] = Mock()
        Serialization().deserialize.return_value = []

        # case1 CreateInstance
        _, err = self.handler.execute_function(self.function_meta, [], InvokeType.CreateInstance, 0, False)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)

        # case2 InvokeFunction
        # case2.1 InvokeFunction sync function
        self.function_meta.functionName = 'sync_get'
        self.function_meta.isAsync = False
        res, err = self.handler.execute_function(self.function_meta, [], InvokeType.InvokeFunction, 1, True)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)
        self.assertTrue(inspect.iscoroutine(res[0]), type(res[0]))

        # case2.2 InvokeFunction async function
        self.function_meta.functionName = 'async_get'
        self.function_meta.isAsync = True
        res, err = self.handler.execute_function(self.function_meta, [], InvokeType.InvokeFunction, 1, True)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)
        self.assertTrue(inspect.iscoroutine(res[0]), type(res[0]))

if __name__ == "__main__":
    main()
