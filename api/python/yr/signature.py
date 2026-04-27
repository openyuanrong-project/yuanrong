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

"""signature"""
import inspect

PLACEHOLDER = b"__YR_PLACEHOLDER__"


def get_signature(func, ignore_first=False):
    """get func signature"""
    if not ignore_first:
        return inspect.signature(func)
    return inspect.Signature(parameters=list(inspect.signature(func).parameters.values())[1:])


def package_args(signature, args, kwargs):
    """check and package args to a list"""
    if signature is not None:
        try:
            signature.bind(*args, **kwargs)
        except TypeError as e:
            raise TypeError(str(e)) from None

    args_list = []
    for arg in args:
        args_list += [PLACEHOLDER, arg]

    for key, value in kwargs.items():
        args_list += [key, value]
    return args_list


def recover_args(args_list):
    """recover args from list"""
    args = []
    kwargs = {}
    for i in range(0, len(args_list), 2):
        key, value = args_list[i], args_list[i + 1]
        if key in (PLACEHOLDER, bytes.decode(PLACEHOLDER)):
            args.append(value)
        else:
            kwargs[key] = value

    return args, kwargs
