# distutils: language = c++
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
from libcpp.memory cimport shared_ptr
from libcpp.string cimport string
from libcpp.vector cimport vector

cdef extern from "src/libruntime/fsclient/protobuf/common.pb.h" nogil:
    cdef cppclass PBAffinity "::common::Affinity"
    cdef cppclass PBInstanceAffinity "::common::InstanceAffinity"
    cdef cppclass CLabelMatchExpression "::common::LabelMatchExpression"

cdef extern from "src/dto/affinity.h" nogil:
    cdef cppclass CLabelOperator "YR::Libruntime::LabelOperator":
        CLabelOperator(const string &type)
        string GetOperatorType() const
        string GetKey() const
        list[string] GetValues() const
        size_t GetLabelOperatorHash() const
        void SetKey(string key)
        void SetValues(list[string] values)
        CLabelMatchExpression GetLabelMatchExpression()

    cdef cppclass CAffinity "YR::Libruntime::Affinity":
        CAffinity(const string kind, string & type)
        string GetAffinityKind() const
        string GetAffinityType() const
        list[shared_ptr[CLabelOperator]] GetLabelOperators() const
        void SetLabelOperators(list[shared_ptr[CLabelOperator]] operators)
        void SetPreferredPriority(bool preferredPriority)
        void SetRequiredPriority(bool requiredPriority)
        void SetPreferredAntiOtherLabels(bool preferredAntiOtherLabels)
        bool GetPreferredAntiOtherLabels()
        string GetAffinityScope() const
        void SetAffinityScope(const string &affinityScope)
        void UpdateAffinityScope(PBInstanceAffinity *pbInstanceAffinity)
        void UpdatePbAffinity(PBAffinity *pbAffinity)
        size_t GetAffinityHash()
        vector[CLabelMatchExpression] GetLabels()

    cdef cppclass CResourcePreferredAffinity "YR::Libruntime::ResourcePreferredAffinity":
        ResourcePreferredAffinity()
        void UpdatePbAffinity(PBAffinity *pbAffinity)

    cdef cppclass CInstancePreferredAffinity "YR::Libruntime::InstancePreferredAffinity":
        InstancePreferredAffinity()
        void UpdatePbAffinity(PBAffinity *pbAffinity)

    cdef cppclass CResourceRequiredAffinity "YR::Libruntime::ResourceRequiredAffinity":
        ResourceRequiredAffinity()
        void UpdatePbAffinity(PBAffinity *pbAffinity)

    cdef cppclass CInstanceRequiredAffinity "YR::Libruntime::InstanceRequiredAffinity":
        InstanceRequiredAffinity()
        void UpdatePbAffinity(PBAffinity *pbAffinity)

    cdef cppclass CResourcePreferredAntiAffinity "YR::Libruntime::ResourcePreferredAntiAffinity":
        ResourcePreferredAntiAffinity()
        void UpdatePbAffinity(PBAffinity *pbAffinity)

    cdef cppclass CInstancePreferredAntiAffinity "YR::Libruntime::InstancePreferredAntiAffinity":
        InstancePreferredAntiAffinity()
        void UpdatePbAffinity(PBAffinity *pbAffinity)

    cdef cppclass CResourceRequiredAntiAffinity "YR::Libruntime::ResourceRequiredAntiAffinity":
        ResourceRequiredAntiAffinity()
        void UpdatePbAffinity(PBAffinity *pbAffinity)

    cdef cppclass CInstanceRequiredAntiAffinity "YR::Libruntime::InstanceRequiredAntiAffinity":
        InstanceRequiredAntiAffinity()
        void UpdatePbAffinity(PBAffinity *pbAffinity)

    cdef cppclass CLabelInOperator "YR::Libruntime::LabelInOperator":
        CLabelInOperator()
        CLabelMatchExpression GetLabelMatchExpression()

    cdef cppclass CLabelNotInOperator "YR::Libruntime::LabelNotInOperator":
        CLabelNotInOperator()
        CLabelMatchExpression GetLabelMatchExpression()

    cdef cppclass CLabelExistsOperator "YR::Libruntime::LabelExistsOperator":
        CLabelExistsOperator()
        CLabelMatchExpression GetLabelMatchExpression()

    cdef cppclass CLabelDoesNotExistOperator "YR::Libruntime::LabelDoesNotExistOperator":
        CLabelDoesNotExistOperator()
        CLabelMatchExpression GetLabelMatchExpression()
