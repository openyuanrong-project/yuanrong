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

from libc.string cimport memcpy
from libc.stdint cimport uint8_t, uint64_t, INT32_MAX, int32_t, int64_t
import cython
import msgpack
import pickle
cimport cpython
from libcpp cimport bool as c_bool
from libcpp.string cimport string as c_string
from libcpp.vector cimport vector as c_vector

from yr.includes.libruntime cimport CBuffer

DEF MSGPACK_HEADER_OFFSET = 8
DEF METADATA_HEADER_OFFSET = 8

cdef extern from "src/utility/memory.h" namespace "YR::utility" nogil:
    void CopyInParallel(uint8_t *dst, const uint8_t *src, int64_t totalBytes, size_t blockSize)

cdef extern from "google/protobuf/repeated_field.h" nogil:
    cdef cppclass RepeatedField[Element]:
        const Element* data() const


cdef int64_t padded_length(int64_t offsets, int64_t alignment):
    return ((offsets + alignment - 1) // alignment) * alignment


def get_uint8_memoryview(const uint8_t[:] bufferview):
    return bufferview


cdef void memcpy_bytes(uint8_t * dst, const char * src, uint64_t size) noexcept nogil:
    memcpy(dst, src, size)


class BufferView:
    def __init__(self):
        self.address = 0
        self.length = 0
        self.itemsize = 0
        self.ndim = 0
        self.readonly = True
        self.format = "B"
        self.shape = []
        self.strides = []


class BufferObjects:
    def __init__(self):
        self.inband_data_size = 0
        self.raw_buffers_size = 0
        self.buffers = []


# default alignment when len(buffer) < 2048.
DEF MinorBufferAlign = 8
# default alignment when len(buffer) >= 2048.
DEF MajorBufferAlign = 64
DEF MajorBufferSize = 2048
DEF MemcopyDefaultBlocksize = 64
DEF MemcopyDefaultThreshold = 1024 * 1024


cdef class SubBuffer:
    cdef:
        void *buf
        void *internal
        int readonly
        int ndim
        c_string _format
        c_vector[Py_ssize_t] _shape
        c_vector[Py_ssize_t] _strides
        object buffer
        Py_ssize_t itemsize
        Py_ssize_t len
        Py_ssize_t *suboffsets

    def __cinit__(self, object buffer):
        self.buffer = buffer
        self.internal = NULL
        self.suboffsets = NULL

    def __len__(self):
        return self.len // self.itemsize

    def __getbuffer__(self, Py_buffer* buffer, int flag):
        if flag & cpython.PyBUF_WRITABLE:
            raise BufferError
        buffer.internal = self.internal
        buffer.readonly = self.readonly
        buffer.ndim = self.ndim
        buffer.buf = self.buf
        buffer.format = <char *>self._format.c_str()
        buffer.shape = self._shape.data()
        buffer.strides = self._strides.data()
        buffer.obj = self  # This is important for GC.
        buffer.itemsize = self.itemsize
        buffer.len = self.len
        buffer.suboffsets = self.suboffsets

    def __getsegcount__(self, Py_ssize_t *lenth_out):
        if lenth_out != NULL:
            lenth_out[0] = <Py_ssize_t> self.size
        return 1

    def __getreadbuffer__(self, Py_ssize_t index, void ** p):
        if index != 0:
            raise SystemError("buffer segment does not exist")
        if p != NULL:
            p[0] = self.buf
        return self.size

    def __getwritebuffer__(self, Py_ssize_t index, void ** p):
        if index != 0:
            raise SystemError("buffer segment does not exist")
        if p != NULL:
            p[0] = self.buf
        return self.size


@cython.boundscheck(False)
@cython.wraparound(False)
def unpack_pickle5_buffers(const uint8_t[:] bufferview):
    cdef:
        const uint8_t *buf_data = &bufferview[0]
        int inband_offset = sizeof(int64_t) * 2
        int64_t inband_size
        int64_t buffer_view_size
        int32_t i
        const uint8_t *buffers_segment
        uint64_t buffer_address

    inband_size = (<int64_t*>buf_data)[0]
    if inband_size < 0:
        raise ValueError("The inband data size should be positive."
                         "Got negative instead. "
                         "Maybe the buffer has been corrupted.")
    buffer_view_size = (<int64_t*>buf_data)[1]
    if buffer_view_size > INT32_MAX or buffer_view_size < 0:
        raise ValueError("Incorrect protobuf size. "
                         "Maybe the buffer has been corrupted.")
    inband_data = bufferview[inband_offset:inband_offset + inband_size]
    buffers_segment = <uint8_t*>buf_data + inband_offset + inband_size + buffer_view_size
    py_buffers = pickle.loads(get_uint8_memoryview(bufferview[inband_offset + inband_size : inband_offset + inband_size + buffer_view_size]))
    buffers_read = []
    for py_buffer in py_buffers.buffers:
        buffer = SubBuffer(bufferview)
        buffer_address = py_buffer.address
        buffer.buf = <void*>(buffers_segment + buffer_address)
        buffer.len = py_buffer.length
        buffer.itemsize = py_buffer.itemsize
        buffer.readonly = py_buffer.readonly
        buffer.ndim = py_buffer.ndim
        buffer._format = py_buffer.format
        for shape in py_buffer.shape:
            buffer._shape.push_back(shape)
        for stride in py_buffer.strides:
            buffer._strides.push_back(stride)
        buffer.internal = NULL
        buffer.suboffsets = NULL
        buffers_read.append(buffer)
    return inband_data, buffers_read


cdef class Pickle5Writer:
    cdef:
        c_vector[Py_buffer] buffers
        uint64_t _curr_buffer_addr
        uint64_t _protobuf_offset
        int64_t _total_length
        object py_buffers
        object _pickle_data

    def __cinit__(self):
        self._total_length = -1
        self._curr_buffer_addr = 0

    def __dealloc__(self):
        for i in range(self.buffers.size()):
            cpython.PyBuffer_Release(&self.buffers[i])

    def __init__(self):
        self.py_buffers = BufferObjects()

    def buffer_callback(self, pickle_buffer):
        cdef:
            int32_t i
            Py_buffer view
        cpython.PyObject_GetBuffer(pickle_buffer, &view, cpython.PyBUF_FULL_RO)
        buffer = BufferView()
        self.py_buffers.buffers.append(buffer)
        buffer.readonly = True
        buffer.length = view.len
        buffer.ndim = view.ndim
        buffer.itemsize = view.itemsize
        if view.format:
            buffer.format = view.format
        if view.strides:
            for i in range(view.ndim):
                buffer.strides.append(view.strides[i])
        if view.shape:
            for i in range(view.ndim):
                buffer.shape.append(view.shape[i])

        buffer_size = MinorBufferAlign if view.len < MajorBufferSize else MajorBufferAlign
        self._curr_buffer_addr = padded_length(self._curr_buffer_addr, buffer_size)
        buffer.address = self._curr_buffer_addr
        self.buffers.push_back(view)
        self._curr_buffer_addr += view.len

    def get_length(self, const uint8_t[:] inband):
        cdef:
            uint64_t inband_data_offset = sizeof(int64_t) * 2
            uint64_t buffer_view_offset
        self.py_buffers.inband_data_size = len(inband)
        self.py_buffers.raw_buffers_size = self._curr_buffer_addr
        # Since calculating the output size is expensive, we will reuse the cached size.
        # However, protobuf could change the output size according to
        # different values, so we MUST NOT change 'py_obj' afterwards.
        if self._pickle_data is None:
            self._pickle_data = pickle.dumps(self.py_buffers)
        buffer_view_size = len(self._pickle_data)
        if buffer_view_size > INT32_MAX:
            raise ValueError("Total buffer metadata size is bigger than %d. "
                             "Consider reduce the number of buffers "
                             "(number of numpy arrays, etc)." % INT32_MAX)
        buffer_view_offset = inband_data_offset + len(inband)
        self._total_length = buffer_view_offset + buffer_view_size
        if self._curr_buffer_addr > 0:
            self._total_length += MajorBufferAlign + self._curr_buffer_addr
        return self._total_length


    @cython.boundscheck(False)
    @cython.wraparound(False)
    def write_to(self, const uint8_t[:] inband, uint8_t[:] data):
        cdef:
            int i
            uint64_t buffer_addr
            uint64_t buffer_len
            uint8_t *data_ptr = &data[0]
            uint64_t inband_size
        if self._total_length < 0:
            raise ValueError("'get_length()' should be called prior to 'write_to'")

        buffer_view_data = self._pickle_data
        buffer_view_size = len(buffer_view_data)
        inband_size = len(inband)
        (<int64_t*>data_ptr)[0] = inband_size
        (<int64_t*>data_ptr)[1] = buffer_view_size
        data_ptr += sizeof(int64_t) * 2
        with nogil:
            memcpy(data_ptr, &inband[0], inband_size)
        data_ptr += inband_size
        memcpy_bytes(data_ptr, buffer_view_data, buffer_view_size)

        data_ptr += buffer_view_size
        if self._curr_buffer_addr <= 0:
            return
        for i in range(len(self.py_buffers.buffers)):
            buffer = self.py_buffers.buffers[i]
            buffer_addr = buffer.address
            buffer_len = buffer.length
            with nogil:
                if buffer_len > MemcopyDefaultThreshold:
                    CopyInParallel(data_ptr + buffer_addr, <const uint8_t*>(self.buffers[i].buf), buffer_len, MemcopyDefaultBlocksize)
                else:
                    memcpy(data_ptr + buffer_addr, self.buffers[i].buf, buffer_len)

cdef class SerializedInterface:
    cdef:
        object _nested_refs

    def __init__(self, nested_refs=None):
        if nested_refs is None:
            self._nested_refs = set()
        else:
            self._nested_refs = nested_refs

    def __len__(self):
        raise NotImplementedError("{}.__len__ not implemented.".format(type(self).__name__))

    @property
    def nested_refs(self):
        return self._nested_refs

    @cython.boundscheck(False)
    @cython.wraparound(False)
    cdef void write_to(self, uint8_t[:] buffer):
        raise NotImplementedError("%s.write_to not implemented.", type(self).__name__)


cdef class Pickle5SerializedObject(SerializedInterface):
    cdef:
        const uint8_t[:] inband
        Pickle5Writer writer
        object _total_length

    def __init__(self, inband, Pickle5Writer writer, nested_refs):
        self.inband = inband
        self.writer = writer
        self._total_length = None
        super(Pickle5SerializedObject, self).__init__(nested_refs)


    def __len__(self):
        if self._total_length is None:
            self._total_length = self.writer.get_length(self.inband)
        return self._total_length

    @cython.boundscheck(False)
    @cython.wraparound(False)
    cdef void write_to(self, uint8_t[:] buffer):
        self.writer.write_to(self.inband, buffer)


cdef class RawSerializedObject(SerializedInterface):
    cdef:
        object data
        const uint8_t *value_ptr
        int64_t _total_length

    def __init__(self, value, nested_refs):
        super(RawSerializedObject, self).__init__()
        self.data = value
        self.value_ptr = <const uint8_t*> value
        self._total_length = len(self.data)
        super(RawSerializedObject, self).__init__(nested_refs)


    def __len__(self):
        return self._total_length

    @cython.boundscheck(False)
    @cython.wraparound(False)
    cdef void write_to(self, uint8_t[:] buffer):
        with nogil:
            memcpy(&buffer[0], self.value_ptr, self._total_length)


cdef class SerializedObject:
    cdef:
        SerializedInterface py_serialized_object
        object _nested_refs
        object metadata
        object msgpack_data
        object msgpack_header
        int64_t msgpack_data_size
        int64_t msgpack_header_size
        int64_t _total_length
        const uint8_t *metadata_ptr
        const uint8_t *msgpack_header_ptr
        const uint8_t *msgpack_data_ptr

    def __init__(self, metadata: int, msgpack_data: bytes, SerializedInterface py_serialized_object=None):
        _total_length = 0
        if py_serialized_object:
            _total_length = len(py_serialized_object)
            self._nested_refs = py_serialized_object.nested_refs
        else:
            self._nested_refs = []

        # the struct of buffer would be
        # |metadata_header|metadata|msgpack_header|msgpack_data|py_serialized_object(optional)|
        self.metadata = msgpack.dumps(metadata)
        self.msgpack_data = msgpack_data
        self.msgpack_data_size = len(msgpack_data)
        self.msgpack_header = msgpack.dumps(self.msgpack_data_size)
        self.msgpack_header_size = len(self.msgpack_header)
        self.py_serialized_object = py_serialized_object

        self._total_length = (METADATA_HEADER_OFFSET + MSGPACK_HEADER_OFFSET +
                             self.msgpack_data_size + _total_length)

        self.metadata_ptr = <const uint8_t*>self.metadata

        self.msgpack_header_ptr = <const uint8_t*>self.msgpack_header
        self.msgpack_data_ptr = <const uint8_t*>self.msgpack_data

        assert self.msgpack_header_size <= MSGPACK_HEADER_OFFSET


    def __len__(self):
        return self._total_length

    @property
    def total_bytes(self):
        """Total serialized byte length (same as __len__); explicit name for clarity in size math."""
        return self._total_length

    @property
    def nested_refs(self):
        return self._nested_refs

    def to_bytes(self) -> bytes:
        cdef shared_ptr[CBuffer] data = \
          dynamic_pointer_cast[CBuffer, NativeBuffer](
            make_shared[NativeBuffer](self._total_length))
        buffer = Buffer.make(data)
        self.write_to(buffer)
        return buffer.to_pybytes()

    @cython.boundscheck(False)
    @cython.wraparound(False)
    cdef void write_to(self, uint8_t[:] buffer):
        cdef uint8_t *ptr = &buffer[0]

        # Write msgpack data first.
        with nogil:
            # copy metadata
            memcpy(ptr, self.metadata_ptr, METADATA_HEADER_OFFSET)
            # copy the msgpack data size
            ptr = ptr + METADATA_HEADER_OFFSET
            memcpy(ptr, self.msgpack_header_ptr, self.msgpack_header_size)
            # copy the data size
            ptr = ptr + MSGPACK_HEADER_OFFSET
            memcpy(ptr, self.msgpack_data_ptr, self.msgpack_data_size)

        if self.py_serialized_object is not None:
            self.py_serialized_object.write_to(
                buffer[METADATA_HEADER_OFFSET +
                       MSGPACK_HEADER_OFFSET + self.msgpack_data_size:])

@cython.boundscheck(False)
@cython.wraparound(False)
def split_buffer(Buffer buf):
    cdef:
        size_t size = buf.size
        uint8_t[:] bufferview = buf
        int64_t metadata
        int64_t msgpack_data_size
        int64_t msgpack_data_start
    assert METADATA_HEADER_OFFSET + MSGPACK_HEADER_OFFSET <= size

    # metadata size
    header_unpacker = msgpack.Unpacker()
    header_unpacker.feed(bufferview[:METADATA_HEADER_OFFSET])
    metadata = header_unpacker.unpack()

    # msgpack data size
    header_unpacker2 = msgpack.Unpacker()
    header_unpacker2.feed(bufferview[METADATA_HEADER_OFFSET:
                                    METADATA_HEADER_OFFSET + MSGPACK_HEADER_OFFSET])

    msgpack_data_size = header_unpacker2.unpack()
    if chr(metadata) == '0':
        metadata = int(constants.Metadata.BYTES)
        msgpack_data_size = size - METADATA_HEADER_OFFSET - MSGPACK_HEADER_OFFSET
    
    if metadata == 0 and msgpack_data_size == 0:
        metadata = int(constants.Metadata.CROSS_LANGUAGE)
        msgpack_data_size = size - METADATA_HEADER_OFFSET - MSGPACK_HEADER_OFFSET

    msgpack_data_start = METADATA_HEADER_OFFSET + MSGPACK_HEADER_OFFSET
    assert (METADATA_HEADER_OFFSET + MSGPACK_HEADER_OFFSET + msgpack_data_size) <= <int64_t>size
    return (metadata, # metadata
            bufferview[msgpack_data_start:msgpack_data_start + msgpack_data_size],  # msgpack_data
            bufferview[msgpack_data_start + msgpack_data_size:])  # py_serialized_object
