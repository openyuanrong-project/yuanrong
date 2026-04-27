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

#include "hetero_future.h"
#include "src/utility/logger/logger.h"
#include "yr/api/hetero_exception.h"

namespace YR {

HeteroFuture::HeteroFuture(std::shared_ptr<YR::Libruntime::HeteroFuture> libFuture)
{
    this->future_ = libFuture;
}

void HeteroFuture::Get()
{
    if (!this->future_) {
        YRLOG_WARN("future_ is null, return directly");
        return;
    }
    YR::Libruntime::AsyncResult result = this->future_->Get();
    if (result.error.OK() && result.failedList.empty()) {
        return;
    }
    throw YR::HeteroException(static_cast<int>(result.error.Code()),
                              static_cast<int>(YR::Libruntime::ModuleCode::DATASYSTEM),
                              "failed to get future, error message: " + result.error.Msg(), result.failedList);
}

}  // namespace YR