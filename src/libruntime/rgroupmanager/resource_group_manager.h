/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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
#include <unordered_map>

#include "absl/synchronization/mutex.h"

#include "src/libruntime/err_type.h"
#include "src/libruntime/utils/constants.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
struct ResourceGroupDetail {
    int bundleSize_ = 0;
    std::unordered_map<std::string, std::promise<ErrorInfo>> rgCreateErrInfo_;
    std::unordered_map<std::string, std::shared_future<ErrorInfo>> rgCreateErrInfoFuture_;
    absl::Mutex mu_;

    explicit ResourceGroupDetail(int bundleSize)
    {
        bundleSize_ = bundleSize;
    }
};

class ResourceGroupManager {
public:
    void StoreRGDetail(const std::string &rgName, const std::string &requestId, int bundleSize);
    void RemoveRGDetail(const std::string &rgName);
    bool IsRGDetailExist(const std::string &rgName);
    int GetRGroupBundleSize(const std::string &rgName);
    void SetRGCreateErrInfo(const std::string &rgName, const std::string &requestId, const ErrorInfo &err);
    ErrorInfo GetRGCreateErrInfo(const std::string &rgName, const std::string &requestId, int timeoutSec);

private:
    absl::Mutex mu_;
    std::unordered_map<std::string, std::shared_ptr<ResourceGroupDetail>> storeMap_ ABSL_GUARDED_BY(mu_);
};
}  // namespace Libruntime
}  // namespace YR