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

#include "object_store_impl.h"

#include "src/libruntime/utils/datasystem_utils.h"

namespace YR {
namespace Libruntime {
ErrorInfo IncreaseRefReturnCheck(const ds::Status &status, const std::vector<std::string> &failedObjectIds)
{
    ErrorInfo err;
    if (!status.IsOk()) {
        std::ostringstream oss;
        oss << "failed to increase ref count, errMsg:" << status.ToString() << ". Failed Objects :[ ";
        std::ostringstream failedListStr;
        int printCount = 0;
        std::copy_if(failedObjectIds.begin(), failedObjectIds.end(),
                     std::ostream_iterator<std::string>(failedListStr, " "),
                     [&printCount](const std::string &id) { return ++printCount < 10; });
        oss << failedListStr.str() << "]";
        err.SetErrCodeAndMsg(YR::Libruntime::ConvertDatasystemErrorToCore(status.GetCode()),
                             YR::Libruntime::ModuleCode::DATASYSTEM, oss.str());
    }
    return err;
}

ErrorInfo DecreaseRefReturnCheck(const ds::Status &status, const std::vector<std::string> &failedObjectIds)
{
    ErrorInfo err;
    if (!status.IsOk()) {
        std::ostringstream oss;
        oss << "DataSystem failed to decrease ref count, errMsg:" << status.ToString() << ". Failed Objects :[ ";
        std::ostringstream failedListStr;
        std::copy(failedObjectIds.begin(), failedObjectIds.end(),
                  std::ostream_iterator<std::string>(failedListStr, " "));
        oss << failedListStr.str() << "]";
        err.SetErrCodeAndMsg(YR::Libruntime::ConvertDatasystemErrorToCore(status.GetCode()),
                             YR::Libruntime::ModuleCode::DATASYSTEM, oss.str());
    }
    return err;
}
}  // namespace Libruntime
}  // namespace YR