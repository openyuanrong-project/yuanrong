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
import os
from unittest import TestCase, main
from unittest.mock import Mock, patch

import yr
from yr.executor.posix_handler import PosixHandler
from yr.executor.faas_handler import FaasHandler
from yr.executor.function_handler import FunctionHandler
from yr.executor.executor import INIT_HANDLER, Executor
import  yr.executor.faas_executor as faas
from yr.libruntime_pb2 import FunctionMeta, LanguageType, InvokeType, ApiType

logger = logging.getLogger(__name__)


class TestExecutor(TestCase):
    def setUp(self) -> None:
        self.runtime = yr.cluster_mode_runtime.ClusterModeRuntime()
        mock_fnruntime = Mock()
        mock_fnruntime.get.return_value = [b"data1", b"data2"]
        mock_fnruntime.config.libruntimeOptions.functionExecuteCallback = Mock()

        self.runtime.libruntime = mock_fnruntime
        self.function_meta = FunctionMeta(applicationName="mytest",
                                     moduleName="test.py",
                                     className = "counter",
                                     functionName = "add",
                                     language=LanguageType.Python,
                                     codeID="123456",
                                     apiType=ApiType.Faas,
                                     signature="",
                                     name="",
                                     ns="",
                                     initializerCodeID="",
                                     isGenerator=False,
                                     isAsync=False)
        

    @patch("yr.log.get_logger")
    def test_load_excutor(self, mock_logger):
        mock_logger.return_value = logger
        os.environ[INIT_HANDLER] = "yrlib_handler.init"
        Executor.load_handler()
        from yr.executor.executor import HANDLER
        self.assertTrue(isinstance(HANDLER, FunctionHandler), f"Failed to load executor, HANDLER type is {type(HANDLER)}")

        os.environ[INIT_HANDLER] = "faas_executor.init"
        Executor.load_handler()
        from yr.executor.executor import HANDLER
        self.assertTrue(isinstance(HANDLER, FaasHandler), f"Failed to load executor, HANDLER type is {type(HANDLER)}")

        os.environ[INIT_HANDLER] = "posix.init"
        Executor.load_handler()
        from yr.executor.executor import HANDLER
        self.assertTrue(isinstance(HANDLER, PosixHandler), f"Failed to load executor, HANDLER type is {type(HANDLER)}")

    @patch("yr.log.get_logger")
    def test_executor(self, mock_logger):
        mock_logger.return_value = logger
        e = Executor(self.function_meta, [], InvokeType.CreateInstanceStateless, 1, None, False)
        _, err = e.execute()

        self.assertTrue("faaS failed" in err.msg)

        e = Executor(self.function_meta, [], InvokeType.InvokeFunctionStateless, 1, None, False)
        res, _ = e.execute()
        self.assertTrue("faas executor find empty user call code" in res[0])

        e = Executor(self.function_meta, [], 10, 1, None, False)
        _, err = e.execute()
        self.assertTrue("invalid invoke type" in err.msg)



if __name__ == "__main__":
    main()
