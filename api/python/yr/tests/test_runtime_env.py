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

import json
import os
import tempfile
import unittest
from unittest.mock import patch, Mock

from yr import runtime_env, InvokeOptions


class TestPut(unittest.TestCase):

    def tearDown(self):
        if "YR_CONDA_HOME" in os.environ:
            os.environ.pop("YR_CONDA_HOME")

    def test_runtime_env_pip_succeed(self):
        opt = InvokeOptions()
        opt.runtime_env["pip"] = ["numpy==1.24", "scipy==1.11"]
        create_opt = runtime_env.parse_runtime_env(opt)
        self.assertEqual(create_opt["POST_START_EXEC"], "pip3 install numpy==1.24 scipy==1.11")

    def test_runtime_env_pip_failed(self):
        opt = InvokeOptions()
        opt.runtime_env["pip"] = ["numpy==1.24"]
        opt.runtime_env["conda"] = "test"
        with self.assertRaises(ValueError) as context:
            _ = runtime_env.parse_runtime_env(opt)

    def test_runtime_env_conda_succeed(self):
        os.environ["YR_CONDA_HOME"] = "/tmp"
        opt = InvokeOptions()
        opt.runtime_env["conda"] = "test"
        create_opt = runtime_env.parse_runtime_env(opt)
        self.assertEqual(create_opt["CONDA_COMMAND"], "conda activate test")
        self.assertEqual(create_opt["CONDA_DEFAULT_ENV"], "test")
        self.assertEqual(create_opt[runtime_env.CONDA_PREFIX], "/tmp")

        with tempfile.NamedTemporaryFile(mode='w+t', delete=True, suffix='.yaml') as temp_file:
            temp_file.write("name: test\n")
            temp_file.seek(0)
            opt.runtime_env["conda"] = temp_file.name
            create_opt = runtime_env.parse_runtime_env(opt)
            self.assertEqual(json.loads(create_opt["CONDA_CONFIG"]), {"name": "test"})
            self.assertEqual(create_opt["CONDA_COMMAND"], "conda env create -f env.yaml")
            self.assertEqual(create_opt["CONDA_DEFAULT_ENV"], "test")

        opt.runtime_env["conda"] = {"name": "test"}
        create_opt = runtime_env.parse_runtime_env(opt)
        self.assertEqual(json.loads(create_opt["CONDA_CONFIG"]), {"name": "test"})
        self.assertEqual(create_opt["CONDA_COMMAND"], "conda env create -f env.yaml")
        self.assertEqual(create_opt["CONDA_DEFAULT_ENV"], "test")

    def test_runtime_env_conda_failed(self):
        opt = InvokeOptions()
        opt.runtime_env["conda"] = "test"
        with self.assertRaises(ValueError) as context:
            _ = runtime_env.parse_runtime_env(opt)

        os.environ["YR_CONDA_HOME"] = "/tmp"

        opt.runtime_env["conda"] = []
        with self.assertRaises(TypeError) as context:
            _ = runtime_env.parse_runtime_env(opt)

        opt.runtime_env["conda"] = "/tmp/aaa.yaml"
        with self.assertRaises(ValueError) as context:
            _ = runtime_env.parse_runtime_env(opt)

        opt.runtime_env["conda"] = {"aaa": 1+2j}
        with self.assertRaises(ValueError) as context:
            _ = runtime_env.parse_runtime_env(opt)

        with tempfile.NamedTemporaryFile(mode='w+t', delete=True, suffix='.yaml') as temp_file:
            temp_file.write("name:x\nbb")
            temp_file.seek(0)
            opt.runtime_env["conda"] = temp_file.name
            with self.assertRaises(ValueError) as context:
                _ = runtime_env.parse_runtime_env(opt)

    def test_runtime_env_working_dir_succeed(self):
        opt = InvokeOptions()
        opt.runtime_env["working_dir"] = "aaa"
        create_opt = runtime_env.parse_runtime_env(opt)
        self.assertEqual(create_opt[runtime_env.WORKING_DIR_KEY], "aaa")

    def test_runtime_env_working_dir_failed(self):
        opt = InvokeOptions()
        opt.runtime_env["working_dir"] = 1
        with self.assertRaises(TypeError) as context:
            _ = runtime_env.parse_runtime_env(opt)

    def test_runtime_env_env_vars_succeed(self):
        opt = InvokeOptions()
        opt.runtime_env["env_vars"] = {"key": "value"}
        _ = runtime_env.parse_runtime_env(opt)
        self.assertEqual(opt.env_vars, {"key": "value"})

    def test_runtime_env_env_vars_failed(self):
        opt = InvokeOptions()
        opt.runtime_env["env_vars"] = 1
        with self.assertRaises(TypeError) as context:
            _ = runtime_env.parse_runtime_env(opt)

        opt.runtime_env["env_vars"] = {"key": 1}
        with self.assertRaises(TypeError) as context:
            _ = runtime_env.parse_runtime_env(opt)

        opt.runtime_env["env_vars"] = {1: "value"}
        with self.assertRaises(TypeError) as context:
            _ = runtime_env.parse_runtime_env(opt)

    def test_runtime_env_shared_dir_succeed(self):
        opt = InvokeOptions()
        opt.runtime_env["shared_dir"] = {"name": "abc", "TTL": 1}
        create_opt = runtime_env.parse_runtime_env(opt)
        self.assertEqual(create_opt["DELEGATE_SHARED_DIRECTORY"], "abc")
        self.assertEqual(create_opt["DELEGATE_SHARED_DIRECTORY_TTL"], "1")

        opt.runtime_env["shared_dir"] = {"name": "abc"}
        create_opt = runtime_env.parse_runtime_env(opt)
        self.assertEqual(create_opt["DELEGATE_SHARED_DIRECTORY"], "abc")
        self.assertEqual(create_opt["DELEGATE_SHARED_DIRECTORY_TTL"], "0")

    def test_runtime_env_shared_dir_failed(self):
        opt = InvokeOptions()
        opt.runtime_env["shared_dir"] = 1
        with self.assertRaises(TypeError) as context:
            _ = runtime_env.parse_runtime_env(opt)

        opt.runtime_env["shared_dir"] = {}
        with self.assertRaises(ValueError) as context:
            _ = runtime_env.parse_runtime_env(opt)

        opt.runtime_env["shared_dir"] = {"name": 1}
        with self.assertRaises(TypeError) as context:
            _ = runtime_env.parse_runtime_env(opt)

        opt.runtime_env["shared_dir"] = {"TTL": "aaa"}
        with self.assertRaises(ValueError) as context:
            _ = runtime_env.parse_runtime_env(opt)


if __name__ == '__main__':
    unittest.main()
