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
from unittest.mock import Mock, patch
import yr.executor.faas_executor as faas
from yr.code_manager import CodeManager

from yr.libruntime_pb2 import FunctionMeta, LanguageType, InvokeType, ApiType
from yr.functionsdk.context import load_context_meta
from yr.err_type import ErrorCode

logger = logging.getLogger(__name__)


class TestFaasExecutor(TestCase):

    @patch("yr.log.get_logger")
    @patch("yr.executor.faas_executor.parse_faas_param")
    @patch.object(CodeManager(), 'get_code_path')
    @patch.object(CodeManager(), 'load')
    def test_parse_faas_param(self, mock_load, mock_get_code_path, parse_faas_param, mock_logger):
        mock_logger.return_value = logger
        mock_get_code_path.return_value = "code_path"

        def user_init_func(context):
            return {"body": "hello world"}

        mock_load.return_value = user_init_func
        parse_faas_param.return_value = {"funcMetaData": {"handler": "user_init_func"},
                                         "extendedMetaData": {"initializer": {"initializer_timeout": "10"}}}

        # case1 init success

        res = faas.faas_init_handler(["arg0"])
        self.assertTrue("success" in res)

        # case2 user function exception
        def err_user_init_func(context):
            raise RuntimeError("Something is wrong with the world")

        mock_load.return_value = err_user_init_func
        try:
            res = faas.faas_init_handler(["arg0"])
        except Exception as e:
            self.assertTrue("Fail to run user init handler" in str(e), str(e))

        # case3 args exception
        parse_faas_param.side_effect = TypeError("mock type error")
        try:
            res = faas.faas_init_handler(["arg0"])
        except Exception as e:
            self.assertTrue("faas init request args undefined" in str(e), str(e))

        parse_faas_param.side_effect = json.decoder.JSONDecodeError("mock json error", "", 0)
        try:
            res = faas.faas_init_handler(["arg0"])
        except Exception as e:
            self.assertTrue("faas init request args json decode error" in str(e), str(e))
        parse_faas_param.reset_mock()

    @patch("yr.log.get_logger")
    @patch("yr.executor.faas_executor.parse_faas_param")
    @patch("yr.executor.faas_executor.get_trace_id_from_params")
    @patch.object(CodeManager(), 'load')
    def test_faas_call_handler(self, mock_load, get_trace_id_from_params, mock_parse_faas_param, mock_logger):
        mock_logger.return_value = logger
        mock_parse_faas_param.side_effect = None
        get_trace_id_from_params.side_effect = None

        def user_call_func(event, context):
            return {"body": "hello world"}

        mock_load.return_value = user_call_func

        mock_parse_faas_param.return_value = {"header": "error_header"}
        get_trace_id_from_params.return_value = "traceid"
        res = faas.faas_call_handler(["arg0", "arg1"])
        self.assertTrue("header type is not dict" in res)

        mock_parse_faas_param.return_value = {"header": {}, "body": "error_body"}
        res = faas.faas_call_handler(["arg0", "arg1"])
        self.assertTrue("failed to loads event body err" in res, res)

        load_context_meta({"funcMetaData": {"timeout": "3"}})
        body = {"name": "world"}
        mock_parse_faas_param.return_value = {"header": {}, "body": json.dumps(body)}
        res = faas.faas_call_handler(["arg0", "arg1"])
        self.assertTrue("hello world" in res, res)

        def user_call_err_func(event, context):
            raise RuntimeError("Something is wrong with the world")

        mock_load.return_value = user_call_err_func
        res = faas.faas_call_handler(["arg0", "arg1"])
        self.assertTrue("Fail to run user call handler" in res, res)

        def user_call_err_return_func(event, context):
            return {1, 2, 3}

        mock_load.return_value = user_call_err_return_func
        res = faas.faas_call_handler(["arg0", "arg1"])
        self.assertTrue("failed to convert the result to a JSON string" in res, res)

    @patch("yr.log.get_logger")
    @patch.object(CodeManager(), 'load')
    def test_faas_shutdown_handler(self, mock_load, mock_logger):
        mock_logger.return_value = logger
        load_context_meta({"extendedMetaData": {"pre_stop": {"pre_stop_timeout": 10}}})

        def user_shutdown_func():
            return

        mock_load.return_value = user_shutdown_func
        err = faas.faas_shutdown_handler(0)
        self.assertTrue(err.error_code == ErrorCode.ERR_OK, err.error_code)

        def user_shutdown_err_func():
            raise RuntimeError("shutdown exception")

        mock_load.return_value = user_shutdown_err_func
        err = faas.faas_shutdown_handler(0)
        self.assertTrue(err.error_code == ErrorCode.ERR_USER_FUNCTION_EXCEPTION, err.error_code)


if __name__ == "__main__":
    main()
