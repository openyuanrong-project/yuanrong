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

import time
import pytest
import yr

@yr.invoke
def get_num(x):
    # 模拟函数执行延迟，确保在调用时，obj 不是立刻就 ready。
    time.sleep(1)
    return x

@yr.invoke
def dis_sum(args):
    # 获取列表中的所有对象并求和
    return sum(yr.get(args))

@pytest.mark.smoke
def test_yr_put(init_yr):
    """
    测试 yr.put 和 yr.get 的端到端流程：
    1. 调用 get_num 生成多个远程对象（含延迟）。
    2. 用 yr.put 写入数据。
    3. 调用 dis_sum 对结果进行求和。
    4. 最终通过 yr.get 获取结果并断言正确性。
    """
    try:
        objref1 = get_num.invoke(1)
        objref2 = get_num.invoke(2)
        objref3 = get_num.invoke(3)

        ref = yr.put([objref1, objref2, objref3])
        ref = dis_sum.invoke(ref)

        # 获取结果，并进行断言检查
        num = yr.get(ref, 30)
        assert num == 6
    except Exception as e:
        return "failed:" + str(e)
