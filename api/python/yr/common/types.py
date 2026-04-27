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

"""types"""

from dataclasses import dataclass
from typing import Set, Any, Dict, List, Optional, Union
import enum


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class InvokeArg:
    """
    Invoke arg
    """
    buf: bytes
    is_ref: bool
    obj_id: str
    nested_objects: Set[str]
    serialized_obj: Any = None


class SupportLanguage(enum.Enum):
    """
    Function support language
    """
    PYTHON = "python"
    CPP = "cpp"
    Java = "java"
    Go = "go"


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class GroupInfo:
    """
    group info
    """
    group_size: int = 0
    group_name: str = ""


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class CommonStatus:
    code: int = 0
    message: str = ""


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class Resource:
    class Type(int, enum.Enum):
        SCALER = 0

    name: str
    type: Type

    @dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
    class Scalar:
        value: float
        limit: float

    scalar: Scalar


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class Resources:
    resources: Dict[str, Resource]


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class BundleInfo:
    bundleId: str
    rGroupName: str
    parentRGroupName: str
    functionProxyId: str
    functionAgentId: str
    tenantId: str
    resources: Resources
    labels: List[str]
    status: CommonStatus
    parentId: str
    kvLabels: Dict[str, str]


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class Option:
    priority: int
    groupPolicy: int
    extension: Dict[str, str]


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class RgInfo:
    name: str
    owner: str
    appId: str
    tenantId: str
    bundles: List[BundleInfo]
    status: CommonStatus
    parentId: str
    requestId: str
    traceId: str
    opt: Option


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class ResourceGroupUnit:
    resourceGroups: Dict[str, RgInfo]
