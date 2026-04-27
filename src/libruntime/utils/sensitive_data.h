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

#include <securec.h>
#include <cstring>
#include <memory>
#include <string>

namespace YR {
namespace Libruntime {
class SensitiveData {
public:
    SensitiveData() = default;
    explicit SensitiveData(const char *str);
    explicit SensitiveData(const std::string &str);
    SensitiveData(const char *str, size_t size);
    SensitiveData(std::unique_ptr<char[]> data, size_t size);

    SensitiveData(SensitiveData &&other) noexcept;
    SensitiveData(const SensitiveData &other);
    ~SensitiveData();

    SensitiveData &operator=(const SensitiveData &other);
    SensitiveData &operator=(SensitiveData &&other) noexcept;
    SensitiveData &operator=(const char *str);
    SensitiveData &operator=(const std::string &str);
    bool operator==(const SensitiveData &other) const;

    bool Empty() const;
    const char *GetData() const;
    size_t GetSize() const;
    bool MoveTo(std::unique_ptr<char[]> &outData, size_t &outSize);
    void Clear() noexcept;

private:
    void SetData(const char *str, size_t size);

    std::unique_ptr<char[]> data_ = nullptr;
    size_t size_ = 0;
};
}  // namespace Libruntime
}  // namespace YR
