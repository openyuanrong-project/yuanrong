/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#pragma once
#include <cstdlib>
#include <cstring>
#include <memory>
#include <unordered_set>

#include "securec.h"
#include "src/libruntime/err_type.h"

namespace YR {
namespace Libruntime {

const unsigned long SECUREC_MAX_LEN = 0x7ffffff0UL;

class Buffer {
public:
    virtual ErrorInfo MemoryCopy(const void *data, uint64_t length) = 0;
    virtual bool IsNative() = 0;
    virtual bool IsString()
    {
        return false;
    }

    virtual uint64_t GetSize() const
    {
        return size_;
    }
    virtual const void *ImmutableData() const
    {
        return data_;
    }

    virtual void *MutableData()
    {
        return data_;
    }

    virtual ErrorInfo Seal(const std::unordered_set<std::string> &nestedIds)
    {
        return ErrorInfo();
    }

    virtual ErrorInfo WriterLatch()
    {
        return ErrorInfo();
    }

    virtual ErrorInfo WriterUnlatch()
    {
        return ErrorInfo();
    }

    virtual ErrorInfo ReaderLatch()
    {
        return ErrorInfo();
    }

    virtual ErrorInfo ReaderUnlatch()
    {
        return ErrorInfo();
    }

    static ErrorInfo DoMemoryCopy(void *dst, uint64_t sizeDst, const void *src, uint64_t lengthSrc)
    {
        if (lengthSrc > sizeDst) {
            std::stringstream ss;
            ss << "memory copy length error, except > " << sizeDst << ", actual: " << lengthSrc;
            return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ss.str());
        }

        uint64_t remainLength = lengthSrc;
        uint64_t index = 0;
        while (remainLength > 0) {
            // limited in securec
            uint64_t copyLength = remainLength > SECUREC_MAX_LEN ? SECUREC_MAX_LEN : remainLength;
            uint64_t maxLength = sizeDst - index;
            maxLength = maxLength > SECUREC_MAX_LEN ? SECUREC_MAX_LEN : maxLength;
            auto ret = memcpy_s(static_cast<char *>(dst) + index, maxLength, static_cast<const char *>(src) + index,
                                copyLength);
            if (ret != 0) {
                std::stringstream ss;
                ss << "memory copy failed, ret: " << ret << ", length: " << lengthSrc;
                return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ss.str());
            }
            index += copyLength;
            remainLength -= copyLength;
        }
        return ErrorInfo();
    }

    virtual ErrorInfo Publish()
    {
        return ErrorInfo();
    }

protected:
    void *data_ = nullptr;
    uint64_t size_ = 0;
};

class NativeBuffer : public Buffer {
public:
    NativeBuffer(void *data, uint64_t size)
    {
        data_ = data;
        size_ = size;
    }
    NativeBuffer(void *data, uint64_t size, bool manageMemory) : selfMalloc_(manageMemory)
    {
        data_ = data;
        size_ = size;
    }
    NativeBuffer(uint64_t size) : selfMalloc_(true)
    {
        size_ = size;
        data_ = std::malloc(size_);
        if (data_ == nullptr) {
            errStr_ = strerror(errno);
        }
    }

    NativeBuffer(std::shared_ptr<Buffer> buffer, uint64_t offset, uint64_t size) : buffer_(buffer)
    {
        size_ = size;
        data_ = static_cast<char *>(buffer->MutableData()) + offset;
    }

    virtual ~NativeBuffer()
    {
        if (selfMalloc_ && data_ != nullptr) {
            std::free(data_);
        }
    }

    virtual bool IsNative() override
    {
        return true;
    }

    virtual ErrorInfo MemoryCopy(const void *data, uint64_t length) override
    {
        return DoMemoryCopy(data_, size_, data, length);
    }

    virtual ErrorInfo Seal(const std::unordered_set<std::string> &nestedIds)
    {
        if (buffer_) {
            return buffer_->Seal(nestedIds);
        }
        return ErrorInfo();
    }

private:
    bool selfMalloc_ = false;
    std::string errStr_ = "";
    std::shared_ptr<Buffer> buffer_;
};

class ReadOnlyNativeBuffer : public NativeBuffer {
public:
    ReadOnlyNativeBuffer(const void *data, uint64_t size) : NativeBuffer(const_cast<void *>(data), size) {}

    virtual ErrorInfo MemoryCopy(const void *data, uint64_t length) override
    {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "Memory copy not supported");
    }
};

class StringNativeBuffer : public Buffer {
public:
    StringNativeBuffer(uint64_t size) : data_(size, '\0') {}

    virtual bool IsNative()
    {
        return true;
    }

    virtual bool IsString() override
    {
        return true;
    }
    std::string &&StringData()
    {
        return std::move(data_);
    }

    virtual ErrorInfo MemoryCopy(const void *data, uint64_t length) override
    {
        return DoMemoryCopy(data_.data(), data_.size(), data, length);
    }

    virtual uint64_t GetSize() const override
    {
        return data_.size();
    }

    virtual const void *ImmutableData() const override
    {
        return data_.data();
    }

    virtual void *MutableData() override
    {
        return static_cast<void *>(data_.data());
    }

private:
    std::string data_;
};

class SharedBuffer : public Buffer {
public:
    SharedBuffer(void *data, uint64_t size) : data_(data), size_(size) {}
    SharedBuffer(std::shared_ptr<Buffer> buffer, uint64_t offset, uint64_t size) : size_(size), buffer_(buffer)
    {
        data_ = static_cast<char *>(buffer->MutableData()) + offset;
    }
    virtual ErrorInfo MemoryCopy(const void *data, uint64_t length) override
    {
        return DoMemoryCopy(data_, size_, data, length);
    }

    virtual bool IsNative()
    {
        return false;
    }

    virtual uint64_t GetSize() const override
    {
        return size_;
    }

    virtual const void *ImmutableData() const override
    {
        return data_;
    }

    virtual void *MutableData() override
    {
        return data_;
    }

    virtual ErrorInfo Seal(const std::unordered_set<std::string> &nestedIds) override
    {
        if (buffer_) {
            return buffer_->Seal(nestedIds);
        }
        return ErrorInfo();
    }

private:
    void *data_;
    uint64_t size_;
    std::shared_ptr<Buffer> buffer_;
};

class ReadOnlySharedBuffer : public SharedBuffer {
public:
    ReadOnlySharedBuffer(const void *data, uint64_t size) : SharedBuffer(const_cast<void *>(data), size) {}
    virtual ErrorInfo MemoryCopy(const void *data, uint64_t length) override
    {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "Memory copy not supported");
    }
};

}  // namespace Libruntime
}  // namespace YR
