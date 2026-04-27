/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#include <memory>
#include "datasystem_client_wrapper.h"
#include "object_store_impl.h"

namespace YR {
namespace Libruntime {
constexpr unsigned int DECRE_RETRY_INTERVAL = 1;
constexpr unsigned int DECRE_REF_BATCH_SIZE = 1000;

class AsyncDecreRef {
public:
    AsyncDecreRef() : running(false), nonEmpty(false), client(nullptr) {}
    ~AsyncDecreRef();

    void Init(std::shared_ptr<DatasystemClientWrapper> clientWrapper);

    void Stop() noexcept;

    bool Push(const std::vector<std::string> &objs, const std::string tenantId);

    bool IsEmpty();

private:
    bool PopBatch(std::vector<std::string> &objs, std::string &tenantId, std::size_t size);

    void Process();

    std::thread bgThread;
    mutable std::mutex mu;
    std::condition_variable cv;
    std::unordered_map<std::string, std::vector<std::string>> objQueue;
    std::atomic<bool> running;
    bool nonEmpty;
    std::shared_ptr<DatasystemClientWrapper> client;
};
}  // namespace Libruntime
}  // namespace YR
