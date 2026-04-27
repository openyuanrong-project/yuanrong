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

#include <mutex>
#include <unordered_map>
#include "src/libruntime/invoke_spec.h"

namespace YR {
namespace Libruntime {
class RequestManager {
public:
    void PushFaasRequest(const std::string &leaseId, const std::shared_ptr<InvokeSpec> spec);
    
    void PushRequest(const std::shared_ptr<InvokeSpec> spec);

    bool PopRequest(const std::string &requestId, std::shared_ptr<InvokeSpec> &spec);

    std::shared_ptr<InvokeSpec> GetRequest(const std::string &requestId);

    bool RemoveRequest(const std::string &requestId);

    std::vector<std::string> GetObjIds();

private:
    mutable absl::Mutex reqMtx_;
    std::unordered_map<std::string, std::shared_ptr<InvokeSpec>> requestMap ABSL_GUARDED_BY(reqMtx_);
};
} // namespace Libruntime
} // namespace YR