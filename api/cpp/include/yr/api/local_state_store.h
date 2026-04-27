/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 * Description: the KV interface provided by yuanrong
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

#include <future>
#include <mutex>
#include <string>
#include <unordered_map>
#include "msgpack.hpp"
#include "yr/api/runtime.h"

namespace YR {
namespace Libruntime {
class ErrorInfo;
}
namespace internal {
typedef std::pair<std::shared_ptr<msgpack::sbuffer>, Libruntime::ErrorInfo> SingleReadResult;
typedef std::pair<std::vector<std::shared_ptr<msgpack::sbuffer>>, Libruntime::ErrorInfo> MultipleReadResult;

class LocalStateStore {
public:
    LocalStateStore();
    ~LocalStateStore();

    void Write(const std::string &key, std::shared_ptr<msgpack::sbuffer> value, ExistenceOpt existence);

    SingleReadResult Read(const std::string &key, int timeoutMS);

    MultipleReadResult Read(const std::vector<std::string> &keys, int timeoutMS);

    void Del(const std::string &key);

    std::vector<std::string> Del(const std::vector<std::string> &keys);

    std::vector<bool> Exist(const std::vector<std::string> &keys);

    // clear all key-values in kv_map
    void Clear() noexcept;

    void MSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<msgpack::sbuffer>> &vals,
                                 ExistenceOpt existence);

private:
    bool IsEmpty();
    int GetRetryInterval(int timeoutMS) const;
    Libruntime::ErrorInfo GetValueWithTimeout(const std::vector<std::string> &keys,
                                              std::vector<std::shared_ptr<msgpack::sbuffer>> &res, bool &isExist,
                                              int timeoutMS);

    std::unordered_map<std::string, std::shared_ptr<msgpack::sbuffer>> kv_map;
    std::mutex mtx;
};
}  // namespace internal
}  // namespace YR
