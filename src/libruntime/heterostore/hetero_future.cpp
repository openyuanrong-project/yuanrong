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
#include "src/libruntime/utils/datasystem_utils.h"

namespace YR {
namespace Libruntime {

YR::Libruntime::AsyncResult ConverDsAsyncResultToLib(datasystem::AsyncResult dsResult)
{
    YR::Libruntime::AsyncResult result;
    YRLOG_DEBUG("convert async result, code is {}, msg is {}, failedList size is {}",
                fmt::underlying(dsResult.status.GetCode()), dsResult.status.GetMsg(), result.failedList.size());
    if (dsResult.status.IsOk()) {
        result.error = YR::Libruntime::ErrorInfo();
    } else {
        result.error = ErrorInfo(YR::Libruntime::ConvertDatasystemErrorToCore(dsResult.status.GetCode()),
                                 dsResult.status.GetMsg());
    }
    result.failedList = dsResult.failedList;
    return result;
}

YR::Libruntime::AsyncResult ConverDsStatusToAsyncRes(datasystem::Status dsStatus)
{
    YR::Libruntime::AsyncResult result;
    YRLOG_DEBUG("convert async result from status, code is {}, msg is {}", fmt::underlying(dsStatus.GetCode()),
                dsStatus.GetMsg());
    if (dsStatus.IsOk()) {
        result.error = YR::Libruntime::ErrorInfo();
    } else {
        result.error = ErrorInfo(YR::Libruntime::ConvertDatasystemErrorToCore(dsStatus.GetCode()), dsStatus.GetMsg());
    }
    result.failedList = {};
    return result;
}

HeteroFuture::HeteroFuture(std::shared_ptr<std::shared_future<datasystem::AsyncResult>> dsFuture)
{
    this->sharedFuture_ = dsFuture;
}

HeteroFuture::HeteroFuture(std::shared_ptr<datasystem::Future> dsFuture)
{
    this->dsFuture_ = dsFuture;
    this->isDsFuture_ = true;
}

bool HeteroFuture::IsDsFuture()
{
    return this->isDsFuture_;
}

YR::Libruntime::AsyncResult HeteroFuture::Get()
{
    if (!this->isDsFuture_) {
        auto dsResult = this->sharedFuture_->get();
        return ConverDsAsyncResultToLib(dsResult);
    }
    datasystem::Status status = this->dsFuture_->Get();
    return ConverDsStatusToAsyncRes(status);
}

}  // namespace Libruntime
}  // namespace YR