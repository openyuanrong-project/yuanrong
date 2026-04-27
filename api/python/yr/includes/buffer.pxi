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
from cpython cimport Py_buffer, PyBytes_FromStringAndSize
from libcpp.memory cimport shared_ptr

from yr.includes.libruntime cimport CBuffer

cdef class Buffer:
    cdef shared_ptr[CBuffer] buffer
    cdef Py_ssize_t shape, strides

    @staticmethod
    cdef make(const shared_ptr[CBuffer]& buffer):
        cdef Buffer self = Buffer.__new__(Buffer)
        self.buffer = buffer
        self.shape = <Py_ssize_t>self.size
        self.strides = <Py_ssize_t>(1)
        return self

    def __len__(self):
        return self.size

    def __getbuffer__(self, Py_buffer* buffer, int flags):
        buffer.readonly = 0
        buffer.buf = <char *>self.buffer.get().MutableData()
        buffer.format = 'B'
        buffer.internal = NULL
        buffer.itemsize = 1
        buffer.len = self.size
        buffer.ndim = 1
        buffer.obj = self
        buffer.shape = &self.shape
        buffer.strides = &self.strides
        buffer.suboffsets = NULL

    def __getsegcount__(self, Py_ssize_t *len_out):
        if len_out != NULL:
            len_out[0] = <Py_ssize_t>self.size
        return 1

    def __getreadbuffer__(self, Py_ssize_t idx, void **p):
        if idx != 0:
            raise SystemError("accessing non-existent buffer segment")
        if p != NULL:
            p[0] = <void*> self.buffer.get().MutableData()
        return self.size

    def __getwritebuffer__(self, Py_ssize_t idx, void **p):
        if idx != 0:
            raise SystemError("accessing non-existent buffer segment")
        if p != NULL:
            p[0] = <void*> self.buffer.get().MutableData()
        return self.size

    @property
    def size(self):
        return self.buffer.get().GetSize()

    def to_pybytes(self):
        return PyBytes_FromStringAndSize(
            <const char*>self.buffer.get().MutableData(),
            self.buffer.get().GetSize())
