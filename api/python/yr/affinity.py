#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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

"""affinity"""

import enum
from dataclasses import dataclass
from typing import List, Optional


class AffinityType(enum.Enum):
    """
    Enum for affinity type.
    """
    PREFERRED = 1
    """
    Preferred affinity type.
    """
    PREFERRED_ANTI = 2
    """
    Preferred anti-affinity type.
    """
    REQUIRED = 3
    """
    Required affinity type.
    """
    REQUIRED_ANTI = 4
    """
    Required anti-affinity type.
    """


class AffinityKind(enum.Enum):
    """
    Enum for affinity kind.
    """
    RESOURCE = 10
    """
    Resource affinity kind. 
    
    To use resource affinity, you need to ensure that the label is applied before the Pod is started. 
    """
    INSTANCE = 20
    """
    Instance affinity kind
    
    1. The instance label only contains key. For instance affinity, the label selection condition can 
       only be set to ``LABEL_EXISTS`` or ``LABEL_NOT_EXISTS``. 
    2. Weak affinity/anti-affinity does not guarantee that an instance will be scheduled to a specified POD. 
       If no POD that meets the conditions is found, other PODs with resources may be selected for scheduling. 
    """


class OperatorType(enum.Enum):
    """
    Enum for the types of label operators.
    This enum defines the different types of label operators that can be used for affinity.
    """
    LABEL_IN = 1
    """
    Represents the "label in" operator.
    
    The resource has a specified key and value in a specific list tag. 
    """
    LABEL_NOT_IN = 2
    """
    Represents the "label not in" operator.
    
    The resource has a specified key, but the value is not in the specified list. 
    """
    LABEL_EXISTS = 3
    """
    Represents the "label exists" operator.
    
    The resource has a tag with the specified key. 
    """
    LABEL_NOT_EXISTS = 4
    """
    Represents the "label not exists" operator.
    
    The resource does not have a tag with the specified key. 
    """


class LabelOperator:
    """
    Label operator for affinity.
    """

    def __init__(self, operator_type: OperatorType, key: str, values: List[str] = None):
        self.operator_type = operator_type
        self.key = key
        self.values = values if values else []


class AffinityScope(enum.Enum):
    """
    Enum for Affinity scope of instances.
    """
    POD = 1
    """
    POD level instance affinity
    POD 级别实例亲和
    """
    NODE = 2
    """
    NODE level instance affinity
    NODE 级别实例亲和
    """


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class Affinity:
    """
    Represents an affinity.
    """
    #: The kind of affinity.
    affinity_kind: AffinityKind
    #: The type of affinity.
    affinity_type: AffinityType
    #: The label operators in the affinity.
    label_operators: List[LabelOperator]
    #: The affinity scope of instances.
    affinity_scope: Optional[AffinityScope] = None

    def __init__(self,
                 affinity_kind: AffinityKind,
                 affinity_type: AffinityType,
                 label_operators: List[LabelOperator],
                 affinity_scope: Optional[AffinityScope] = None):
        self.affinity_kind = affinity_kind
        self.affinity_type = affinity_type
        self.label_operators = label_operators
        self.affinity_scope = affinity_scope
