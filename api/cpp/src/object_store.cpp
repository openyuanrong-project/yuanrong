/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include "yr/api/object_store.h"
namespace YR {
namespace internal {
void ThrowExceptionBasedOnStatus(const GetStatus status, const ErrorInfo &err,
                                 const std::vector<std::string> &remainIds, int timeoutMS, bool allowPartial)
{
    std::ostringstream oss;
    if (status == GetStatus::ALL_SUCCESS) {
        return;
    }
    if (status == GetStatus::ALL_FAILED || status == GetStatus::PARTIAL_SUCCESS) {
        // If the failure is not due to a timeout, there will be an error message from the data system.
        oss << err.Msg();
    } else if (status == GetStatus::ALL_FAILED_AND_TIMEOUT || status == GetStatus::PARTIAL_SUCCESS_AND_TIMEOUT) {
        oss << "Get timeout " << timeoutMS << "ms.";
    }
    if (status == GetStatus::ALL_FAILED || status == GetStatus::ALL_FAILED_AND_TIMEOUT) {
        oss << " all";
    } else {
        oss << " partial";
    }
    oss << " failed: "
        << "(" << remainIds.size() << "). ";
    oss << "Failed objects: [ ";
    oss << remainIds[0] << " ... "
        << "]";
    if (status == GetStatus::ALL_FAILED || status == GetStatus::ALL_FAILED_AND_TIMEOUT) {
        throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), oss.str());
    }
    if (allowPartial) {
        return;
    }
    throw YR::Exception(static_cast<int>(err.Code()), static_cast<int>(err.MCode()), oss.str());
}
}  // namespace internal
}  // namespace YR
