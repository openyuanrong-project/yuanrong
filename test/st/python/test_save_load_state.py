#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

import pytest
import yr
from common import Add

@pytest.mark.smoke
def test_save_load_state(init_yr):
    """
    A function to test saving and loading the state of a stateful function.
    """
    udf_class = Add.invoke()
    print(f"member value befor save state: {yr.get(udf_class.get.invoke())}")
    udf_class.save.invoke()

    udf_class.add.invoke()
    value = yr.get(udf_class.get.invoke())
    print(f"member value after add one: {value}")
    assert value == 1

    udf_class.load.invoke()
    value = yr.get(udf_class.get.invoke())
    print(f"member value after load state(back to 0): {value}")
    assert value == 0

@pytest.mark.smoke
def test_save_load_state_fail(init_yr):
    """
    A function to test saving and loading the state of an stateful function.
    """
    udf_class = Add.invoke()
    exception = None
    try:
        yr.get(udf_class.load.invoke())
    except RuntimeError as e:
        exception = e
    assert exception is not None
    assert "Failed to load state" in str(exception)

    try:
        yr.get(udf_class.save.invoke("invalid timeout"))
    except TypeError as e:
        exception = e
    assert exception is not None
    assert "an integer is required" in str(exception)

