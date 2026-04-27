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

#include "sensitive_data.h"

#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
SensitiveData::SensitiveData(const char *str)
{
    SetData(str, str == nullptr ? 0 : std::strlen(str));
}

SensitiveData::SensitiveData(const std::string &str)
{
    SetData(str.data(), str.length());
}

SensitiveData::SensitiveData(const char *str, size_t size)
{
    SetData(str, size);
}

SensitiveData::SensitiveData(std::unique_ptr<char[]> data, size_t size) : data_(std::move(data)), size_(size) {}

SensitiveData::SensitiveData(SensitiveData &&other) noexcept : data_(std::move(other.data_)), size_(other.size_)
{
    other.size_ = 0;
}

SensitiveData::SensitiveData(const SensitiveData &other)
{
    if (!other.Empty()) {
        SetData(other.data_.get(), other.size_);
    }
}

SensitiveData::~SensitiveData()
{
    Clear();
}

SensitiveData &SensitiveData::operator=(const SensitiveData &other)
{
    if (this == &other) {
        return *this;
    }
    Clear();
    if (!other.Empty()) {
        SetData(other.data_.get(), other.size_);
    }
    return *this;
}

SensitiveData &SensitiveData::operator=(SensitiveData &&other) noexcept
{
    if (this == &other) {
        return *this;
    }
    Clear();
    data_ = std::move(other.data_);
    size_ = other.size_;
    other.size_ = 0;
    return *this;
}

SensitiveData &SensitiveData::operator=(const char *str)
{
    Clear();
    SetData(str, std::strlen(str));
    return *this;
}

SensitiveData &SensitiveData::operator=(const std::string &str)
{
    Clear();
    SetData(str.data(), str.length());
    return *this;
}

bool SensitiveData::Empty() const
{
    return data_ == nullptr || size_ == 0;
}

const char *SensitiveData::GetData() const
{
    return Empty() ? "" : data_.get();
}

size_t SensitiveData::GetSize() const
{
    return size_;
}

void SensitiveData::SetData(const char *str, size_t size)
{
    if (str != nullptr && size > 0) {
        data_ = std::make_unique<char[]>(size + 1);
        size_ = size;
        int ret = memcpy_s(data_.get(), size_, str, size_);
        if (ret != EOK) {
            YRLOG_WARN("memcpy failed, ret = {}", ret);
            Clear();
        }
    }
}
void SensitiveData::Clear() noexcept
{
    if (data_ != nullptr && size_ > 0) {
        int ret = memset_s(data_.get(), size_, 0, size_);
        if (ret != EOK) {
            YRLOG_WARN("memset failed, ret = {}", ret);
        }
    }
    size_ = 0;
    data_ = nullptr;
}

bool SensitiveData::MoveTo(std::unique_ptr<char[]> &outData, size_t &outSize)
{
    if (Empty()) {
        return false;
    }
    outData = std::move(data_);
    outSize = size_;
    size_ = 0;
    return true;
}

bool SensitiveData::operator==(const SensitiveData &other) const
{
    if (size_ != other.size_) {
        return false;
    }

    if (size_ == 0) {
        return true;
    }

    if (data_ == nullptr || other.data_ == nullptr) {
        return false;
    }

    return strcmp(data_.get(), other.data_.get()) == 0;
}
}  // namespace Libruntime
}  // namespace YR
