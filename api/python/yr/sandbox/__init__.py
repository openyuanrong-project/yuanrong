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

"""Sandbox module for isolated code execution."""

__all__ = ["SandBox", "SandBoxInstance", "create"]

import importlib


def __getattr__(name):
    if name in __all__:
        module = importlib.import_module("yr.sandbox.sandbox")
        return getattr(module, name)
    raise AttributeError(f"module 'yr.sandbox' has no attribute '{name}'")


def __dir__():
    return sorted(list(globals().keys()) + __all__)
