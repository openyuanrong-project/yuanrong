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
from unittest.mock import patch, Mock
from yr.object_ref import ObjectRef

class TestGet(unittest.TestCase):
    @patch('yr.runtime_holder.global_runtime.get_runtime')
    def test_get_invalid_timeout(self, mock_get_runtime):
        obj_ref = yr.object_ref.ObjectRef(123)
        time = -10
        mock_get_runtime =Mock()
        mock_get_runtime.get.return_value="Parameter 'timeout' should be greater than 0 or equal to -1 (no timeout)"
        mock_get_runtime.return_value=mock_get_runtime
        yr.apis.set_initialized()
        with self.assertRaises(ValueError):
            yr.apis.get(obj_ref,time)


if __name__ == '__main__':
    unittest.main()