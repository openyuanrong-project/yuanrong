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
import pickle
from unittest.mock import patch, Mock

from yr.common import constants
from yr.serialization import Serialization
from yr.fnruntime import write_to_cbuffer
import numpy as np


class TestApi(unittest.TestCase):
    def setUp(self) -> None:
        pass

    @patch("yr.runtime_holder.global_runtime.get_runtime")
    def test_serialize_base(self, get_runtime):
        mock_runtime = Mock()
        mock_runtime.increase_global_reference.return_value = None
        get_runtime.return_value = mock_runtime

        class Data:
            def __init__(self, v):
                self.data = v

            def __eq__(self, o: object) -> bool:
                return self.data == o.data

        values = ["test", 1, Data("aaa")]
        for value in values:
            obj = Serialization().serialize(value)
            buf = write_to_cbuffer(obj)
            ret = Serialization().deserialize(buf)
            assert ret == value

        values = [np.array([1, 2, 3, 4, 5])]
        for value in values:
            obj = Serialization().serialize(value)
            buf = write_to_cbuffer(obj)
            ret = Serialization().deserialize(buf)
            assert ret.sum() == value.sum()
        if pickle.HIGHEST_PROTOCOL >= 5:
            default_highest_protocol = pickle.HIGHEST_PROTOCOL
            pickle.HIGHEST_PROTOCOL = 4
            values = ["test", 1, Data("aaa")]
            for value in values:
                obj = Serialization().serialize(value)
                buf = write_to_cbuffer(obj)
                ret = Serialization().deserialize(buf)
                assert ret == value

            values = [np.array([1, 2, 3, 4, 5])]
            for value in values:
                obj = Serialization().serialize(value)
                buf = write_to_cbuffer(obj)
                ret = Serialization().deserialize(buf)
                assert ret.sum() == value.sum()
            pickle.HIGHEST_PROTOCOL = default_highest_protocol

    def test_normalize_input(self):
        input_one = b"hello, world"
        res_one, res_two = Serialization().normalize_input(input_one)
        assert res_one == constants.Metadata.BYTES

        input_two = memoryview(input_one)
        res_three, res_four = Serialization().normalize_input(input_two)
        assert res_three == constants.Metadata.MEMORYVIEW

        input_three = bytearray(input_one)
        res_five, res_six = Serialization().normalize_input(input_three)
        assert res_five == constants.Metadata.BYTEARRAY
if __name__ == "__main__":
    unittest.main()
