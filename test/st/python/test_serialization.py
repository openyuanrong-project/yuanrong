#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

import pytest
import yr
import string
import numpy as np
import collections
import time


@pytest.mark.smoke
def test_serialization_with_many_types(init_yr):
    primitive_objects = [
        # Various primitive types.
        0,
        0.0,
        b"1",
        string.printable,
        "\u262F",
        "hello world",
        "\xff\xfe\x9c\x001\x000\x00",
        True,
        False,
        [],
        (),
        {},
        type,
        int,
        set(),
    ]

    composite_objects = (
        [[obj] for obj in primitive_objects]
        + [(obj,) for obj in primitive_objects]
        + [{(): obj} for obj in primitive_objects]
    )

    numpy_objects = [
        np.random.normal(size=(10, 20, 30)),
    ]

    complex_numpy_objects = [
        [np.random.normal(size=i) for i in range(10)],
        [np.random.normal(size=i * (1,)) for i in range(20)],
        [np.random.normal(size=i * (5,)) for i in range(1, 8)],
    ]

    @yr.invoke
    def f(x):
        return x

    for obj in primitive_objects + composite_objects:
        new_obj_1 = yr.get(f.invoke(obj))
        new_obj_2 = yr.get(yr.put(obj))
        assert obj == new_obj_1
        assert obj == new_obj_2
        if type(obj).__module__ != "numpy" and type(obj) is not bytes:
            assert type(obj) is type(new_obj_1)
            assert type(obj) is type(new_obj_2)

    for obj in numpy_objects:
        new_obj_1 = yr.get(f.invoke(obj))
        new_obj_2 = yr.get(yr.put(obj))
        assert (obj == new_obj_1).all()
        assert (obj == new_obj_2).all()
        if type(obj).__module__ != "numpy":
            assert type(obj) is type(new_obj_1)
            assert type(obj) is type(new_obj_2)

    for obj in complex_numpy_objects:
        new_obj_1 = yr.get(f.invoke(obj))
        new_obj_2 = yr.get(yr.put(obj))
        for o1, o2 in zip(obj, new_obj_1):
            assert (o1 == o2).all()
        for o1, o2 in zip(obj, new_obj_2):
            assert (o1 == o2).all()
        if type(obj).__module__ != "numpy":
            assert type(obj) is type(new_obj_1)
            assert type(obj) is type(new_obj_2)


@pytest.mark.benchmark
def test_serialization_with_big_numpy(init_yr):
    obj = np.random.normal(size=(100, 200, 300))

    start = time.perf_counter_ns()

    for i in range(100):
        new_obj = yr.get(yr.put(obj))

    end = time.perf_counter_ns()
    elapsed_ns = end - start
    print(f"cost: {elapsed_ns/1000/1000} ms")
