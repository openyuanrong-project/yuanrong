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
import sys

from unittest import mock, TestCase, main

from yr.code_manager import CodeManager
from yr.err_type import ErrorCode
from yr.libruntime_pb2 import FunctionMeta


logger = logging.getLogger(__name__)


class TestCodeManager(TestCase):
    def setUp(self):
        self.cm = CodeManager()

    def tearDown(self):
        self.cm = None

    @mock.patch.object(sys, 'path', new_callable=mock.Mock())
    def test_load_functions_when_input_one(self, mock_sys_path):
        path = "test"
        mock_sys_path.insert.return_value = None
        self.cm.load_functions([path])
        mock_sys_path.insert.assert_called_once_with(0, path)

    @mock.patch("yr.log.get_logger")
    def test_load_functions_when_input_invalid_faas_entry(self, mock_logger):
        mock_logger.return_value = logger
        self.cm.custom_handler = "/tmp"
        err = self.cm.load_functions(["test.init", "test.handler"])
        assert err.error_code == ErrorCode.ERR_USER_CODE_LOAD

    @mock.patch.object(CodeManager(), 'load_code_from_local')
    @mock.patch("yr.log.get_logger")
    def test_load_functions_when_user_code_syntax_err(self, mock_logger, mock_load_code_from_local):
        mock_logger.return_value = logger
        mock_load_code_from_local.side_effect = SyntaxError("a syntax error in user code")
        err = CodeManager().load_functions(["test.init", "test.handler"])
        assert err.error_code == ErrorCode.ERR_USER_CODE_LOAD

    @mock.patch("yr.log.get_logger")
    def test_entry_load(self, mock_logger):
        mock_logger.return_value = logger
        self.assertFalse(self.cm.get_code_path("none"))

        def functiona():
            return "functiona"
        self.cm.register("functiona", functiona)
        self.assertFalse(self.cm.load("functionb"))
        self.assertTrue(self.cm.load("functiona"))
        self.assertEqual(self.cm.load("functiona")(), "functiona")

        meta = FunctionMeta()
        meta.codeID = "functiona"
        self.assertTrue(self.cm.load_code(meta))
        self.assertEqual(self.cm.load_code(meta)(), "functiona")

        meta.codeID = "functionb"
        with self.assertRaises(RuntimeError):
            self.cm.load_code(meta)

        def functionb():
            return "functionb"

        def download_from_ds(codeid):
            return functionb
        self.cm.register_load_code_from_datasystem_func(download_from_ds)
        self.assertTrue(self.cm.load_code(meta))
        self.assertEqual(self.cm.load_code(meta)(), "functionb")

        def download_from_ds_None(codeid):
            return None
        self.cm.register_load_code_from_datasystem_func(download_from_ds_None)
        meta.codeID = "functionc"
        with self.assertRaises(ImportError):
            self.cm.load_code(meta)

        meta.codeID = ""
        meta.functionName = "functiona"
        with self.assertRaises(ValueError):
            self.cm.load_code(meta)
        

    @mock.patch("os.path.exists")
    @mock.patch("importlib.util.spec_from_file_location")
    @mock.patch("importlib.util.module_from_spec")
    @mock.patch("yr.log.get_logger")
    def test_load_module(self, mock_logger, mock_module_from_spec, mock_spec_from_file_location, mock_exists):
        mock_logger.return_value = logger
        mock_exists.return_value = True
        mock_spec = mock.Mock()
        mock_spec.loader = mock.Mock()
        mock_spec.loader.exec_module.return_value = "exec_module"
        mock_spec_from_file_location.return_value = mock_spec
        mock_module_from_spec.return_value = "module"
        # Use the public testing interface for load_module_impl
        res = self.cm.load_module_impl_for_testing("code_dir", "module_name")
        self.assertEqual(res, "module", res)


class TestCodeManagerClassLookupAlias(TestCase):
    """
    Covers get_instance / load_code path: alias key moduleName%%className so slim FunctionMeta
    (no code / codeID) still resolves nested-class qualnames. Related to instance inline-code flow.
    """

    def setUp(self):
        self.cm = CodeManager()
        self.cm.clear()

    def tearDown(self):
        self.cm.clear()

    def test_register_alias_skips_non_type(self):
        meta = FunctionMeta()
        meta.moduleName = "m"
        meta.className = "C"
        self.cm.register_class_lookup_alias(meta, object())
        self.assertIsNone(self.cm.load("m%%C"))

    def test_register_alias_skips_empty_module_or_class(self):
        class DummyType:
            pass

        meta = FunctionMeta()
        meta.className = "DummyType"
        self.cm.register_class_lookup_alias(meta, DummyType)
        self.assertIsNone(self.cm.load("%%DummyType"))

        meta2 = FunctionMeta()
        meta2.moduleName = "m"
        self.cm.register_class_lookup_alias(meta2, DummyType)
        self.assertIsNone(self.cm.load("m%%"))

    def test_load_code_class_from_bytes_registers_alias(self):
        class UTNested:
            pass

        def from_bytes(_b):
            return UTNested

        self.cm.register_load_code_from_bytes_func(from_bytes)
        meta = FunctionMeta()
        meta.code = b"x"
        meta.moduleName = "ut_fake_mod"
        meta.className = "outer.<locals>.UTNested"
        got = self.cm.load_code(meta, is_class=True)
        self.assertIs(got, UTNested)

        slim = FunctionMeta()
        slim.moduleName = meta.moduleName
        slim.className = meta.className
        again = self.cm.load_code(slim, is_class=True)
        self.assertIs(again, UTNested)

    def test_load_code_class_from_codeid_registers_alias(self):
        class FromDS:
            pass

        self.cm.register_load_code_from_datasystem_func(lambda _cid: FromDS)
        meta = FunctionMeta()
        meta.codeID = "ut-code-id-1"
        meta.moduleName = "m_ds"
        meta.className = "FromDS"
        got = self.cm.load_code(meta, is_class=True)
        self.assertIs(got, FromDS)

        slim = FunctionMeta()
        slim.moduleName = "m_ds"
        slim.className = "FromDS"
        again = self.cm.load_code(slim, is_class=True)
        self.assertIs(again, FromDS)


if __name__ == '__main__':
    main()
