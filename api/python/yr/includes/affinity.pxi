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

from libcpp cimport bool
from libcpp.list cimport list
from libcpp.memory cimport dynamic_pointer_cast, make_shared, shared_ptr, nullptr
from libcpp.string cimport string
from yr.affinity import (
    Affinity,
    AffinityKind,
    AffinityType,
    LabelOperator,
    OperatorType,
    AffinityScope,
)
from yr.includes.affinity cimport (
    CAffinity,
    CInstancePreferredAffinity,
    CInstancePreferredAntiAffinity,
    CInstanceRequiredAffinity,
    CInstanceRequiredAntiAffinity,
    CLabelDoesNotExistOperator,
    CLabelExistsOperator,
    CLabelInOperator,
    CLabelNotInOperator,
    CLabelOperator,
    CResourcePreferredAffinity,
    CResourcePreferredAntiAffinity,
    CResourceRequiredAffinity,
    CResourceRequiredAntiAffinity,
)

cdef shared_ptr[CLabelOperator] label_operator_from_py_to_cpp(label_operator: LabelOperator):
    cdef:
        shared_ptr[CLabelOperator] c_label_opt
        list[string] c_values
        string key = label_operator.key.encode()
    operator_type = label_operator.operator_type
    if operator_type == OperatorType.LABEL_IN:
        c_label_opt = dynamic_pointer_cast[CLabelOperator, CLabelInOperator](make_shared[CLabelInOperator]())
    elif operator_type == OperatorType.LABEL_NOT_IN:
        c_label_opt = dynamic_pointer_cast[CLabelOperator, CLabelNotInOperator](make_shared[CLabelNotInOperator]())
    elif operator_type == OperatorType.LABEL_EXISTS:
        c_label_opt = dynamic_pointer_cast[CLabelOperator, CLabelExistsOperator](make_shared[CLabelExistsOperator]())
    elif operator_type == OperatorType.LABEL_NOT_EXISTS:
        c_label_opt = dynamic_pointer_cast[CLabelOperator, CLabelDoesNotExistOperator](
            make_shared[CLabelDoesNotExistOperator]())
    else:
        raise ValueError(f"Invalid operator type: {operator_type}")
    c_label_opt.get().SetKey(key)
    for v in label_operator.values:
        c_values.push_back(v.encode())
    c_label_opt.get().SetValues(c_values)
    return c_label_opt


cdef shared_ptr[CAffinity] affinity_from_py_to_cpp(affinity: Affinity, bool preferredPriority, bool requiredPriority, bool preferredAntiOtherLabels):
    cdef:
        shared_ptr[CAffinity] c_affinity
        list[shared_ptr[CLabelOperator]] c_operators
        shared_ptr[CLabelOperator] c_operator

    if affinity.affinity_kind == AffinityKind.RESOURCE:
        if affinity.affinity_type == AffinityType.PREFERRED:
            c_affinity = dynamic_pointer_cast[CAffinity, CResourcePreferredAffinity](
                make_shared[CResourcePreferredAffinity]())
        elif affinity.affinity_type == AffinityType.PREFERRED_ANTI:
            c_affinity = dynamic_pointer_cast[CAffinity, CResourcePreferredAntiAffinity](
                make_shared[CResourcePreferredAntiAffinity]())
        elif affinity.affinity_type == AffinityType.REQUIRED:
            c_affinity = dynamic_pointer_cast[CAffinity, CResourceRequiredAffinity](
                make_shared[CResourceRequiredAffinity]())
        elif affinity.affinity_type == AffinityType.REQUIRED_ANTI:
            c_affinity = dynamic_pointer_cast[CAffinity, CResourceRequiredAntiAffinity](
                make_shared[CResourceRequiredAntiAffinity]())
    elif affinity.affinity_kind == AffinityKind.INSTANCE:
        if affinity.affinity_type == AffinityType.PREFERRED:
            c_affinity = dynamic_pointer_cast[CAffinity, CInstancePreferredAffinity](
                make_shared[CInstancePreferredAffinity]())
        elif affinity.affinity_type == AffinityType.PREFERRED_ANTI:
            c_affinity = dynamic_pointer_cast[CAffinity, CInstancePreferredAntiAffinity](
                make_shared[CInstancePreferredAntiAffinity]())
        elif affinity.affinity_type == AffinityType.REQUIRED:
            c_affinity = dynamic_pointer_cast[CAffinity, CInstanceRequiredAffinity](
                make_shared[CInstanceRequiredAffinity]())
        elif affinity.affinity_type == AffinityType.REQUIRED_ANTI:
            c_affinity = dynamic_pointer_cast[CAffinity, CInstanceRequiredAntiAffinity](
                make_shared[CInstanceRequiredAntiAffinity]())

    if c_affinity == nullptr:
        raise ValueError(f"c_affinity is nullptr. Please check: affinity_kind: {affinity.affinity_kind},"
                         f"affinity_type: {affinity.affinity_type}")

    for operator in affinity.label_operators:
        c_operator = label_operator_from_py_to_cpp(operator)
        if c_operator == nullptr:
            raise ValueError("Failed to convert LabelOperator to cpp LabelOperator.")
        c_operators.push_back(c_operator)

    if affinity.affinity_scope == AffinityScope.POD:
        c_affinity.get().SetAffinityScope(AffinityScope.POD.name.encode())
    elif affinity.affinity_scope == AffinityScope.NODE:
        c_affinity.get().SetAffinityScope(AffinityScope.NODE.name.encode())

    c_affinity.get().SetLabelOperators(c_operators);
    c_affinity.get().SetPreferredPriority(preferredPriority);
    c_affinity.get().SetRequiredPriority(requiredPriority);
    c_affinity.get().SetPreferredAntiOtherLabels(preferredAntiOtherLabels);
    return c_affinity
