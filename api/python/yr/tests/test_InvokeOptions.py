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
import yr

class TestInvokeOptions(unittest.TestCase):

    def setUp(self):
        @yr.invoke
        def add(x):
            return x + 1
        self.add = add
        self.opts = yr.InvokeOptions()

    def test_option_function_name(self):
        self.opts.name = 999
        with self.assertRaises(TypeError):
            self.add.options(self.opts)

    def test_option_function_preferred_anti_other_labels(self):
        self.opts.preferred_anti_other_labels = ''
        with self.assertRaises(TypeError):
            self.add.options(self.opts)

    def test_option_function_preferred_priority(self):
        self.opts.preferred_priority = ''
        with self.assertRaises(TypeError):
            self.add.options(self.opts)

    def test_option_function_namespace(self):
        self.opts.namespace = 11
        with self.assertRaises(TypeError):
            self.add.options(self.opts)

    def test_option_function_dict(self):
        self.opts.env_vars = []
        with self.assertRaises(TypeError):
            self.add.options(self.opts)

    def test_option_function_namespace_true(self):
        self.opts.namespace = "aaa"
        result = self.add.options(self.opts)
        self.assertIsNotNone(result)

    def test_option_function_custom_resources(self):
        self.opts.custom_resources = []
        with self.assertRaises(TypeError):
            self.add.options(self.opts)

    def test_option_function_custom_extensions(self):
        self.opts.custom_extensions = []
        with self.assertRaises(TypeError):
            self.add.options(self.opts)

    def test_option_function_recover_retry_times_range(self):
        self.opts.recover_retry_times = -1
        with self.assertRaises(ValueError):
            self.add.options(self.opts)

    def test_option_function_max_instances_range(self):
        self.opts.max_instances = 9**20
        with self.assertRaises(ValueError):
            self.add.options(self.opts)

    def test_option_function_max_invoke_latency_range(self):
        self.opts.max_invoke_latency = 9**20
        with self.assertRaises(ValueError):
            self.add.options(self.opts)

    def test_option_function_min_instances_range(self):
        self.opts.min_instances = -1
        with self.assertRaises(ValueError):
            self.add.options(self.opts)


if __name__ == '__main__':
    unittest.main()