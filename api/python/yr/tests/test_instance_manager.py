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

import gc
import unittest
from unittest.mock import Mock, patch
from yr.executor.instance_manager import InstanceManager, InstancePackage

class TestInstanceManager(unittest.TestCase):
    def setUp(self) -> None:
        self.runtime_patcher = patch('yr.runtime_holder.global_runtime.get_runtime')
        self.mock_get_runtime = self.runtime_patcher.start()
        self.mock_get_runtime.return_value = Mock()

    def tearDown(self) -> None:
        manager = InstanceManager()
        manager.init(None)
        manager.class_code = None
        manager.is_async = False
        gc.collect()
        self.runtime_patcher.stop()

    def test_init_instance_manager_code_ref(self):
        InstanceManager().set_code_ref("code_id", True)
        assert len(InstanceManager().code_id) != 0

    def test_get_instance_manager_from_package(self):
        ins_package = InstancePackage(None, None, True, "code_id")
        assert ins_package.get_class_code() is None
        assert ins_package.get_instance() is None
        assert ins_package.get_is_async()
        assert len(ins_package.get_code_id()) != 0
        InstanceManager().init_from_inspackage(ins_package)
        assert InstanceManager().is_async
        assert len(InstanceManager().code_id) != 0

        ins_package_from_manager = InstanceManager().get_instance_package()
        assert ins_package_from_manager.get_is_async()

if __name__ == "__main__":
    unittest.main()
