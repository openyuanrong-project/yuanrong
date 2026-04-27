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
#include "datasystem_utils.h"

#include <sstream>

#include "src/libruntime/utils/constants.h"

namespace ds = datasystem;
namespace YR {
namespace Libruntime {
bool IsRetryableStatus(const ds::Status &status)
{
    static std::set<ds::StatusCode> retryableStatus(
        {ds::StatusCode::K_OK, ds::StatusCode::K_NOT_FOUND, ds::StatusCode::K_OUT_OF_MEMORY,
         ds::StatusCode::K_TRY_AGAIN, ds::StatusCode::K_RPC_CANCELLED, ds::StatusCode::K_RPC_DEADLINE_EXCEEDED,
         ds::StatusCode::K_RPC_UNAVAILABLE});
    return status.IsOk() || retryableStatus.find(status.GetCode()) != retryableStatus.end();
}

bool IsUnlimitedRetryableStatus(const ds::Status &status)
{
    static std::set<ds::StatusCode> unlimitedRetryableStatus({ds::StatusCode::K_OK, ds::StatusCode::K_NOT_FOUND,
                                                              ds::StatusCode::K_OUT_OF_MEMORY,
                                                              ds::StatusCode::K_UNKNOWN_ERROR});
    return status.IsOk() || unlimitedRetryableStatus.find(status.GetCode()) != unlimitedRetryableStatus.end();
}

bool IsLimitedRetryableStatus(const ds::Status &status)
{
    static std::set<ds::StatusCode> limitedRetryableStatus(
        {ds::StatusCode::K_TRY_AGAIN, ds::StatusCode::K_RPC_CANCELLED, ds::StatusCode::K_RPC_DEADLINE_EXCEEDED,
         ds::StatusCode::K_RPC_UNAVAILABLE});
    return limitedRetryableStatus.find(status.GetCode()) != limitedRetryableStatus.end();
}

bool IsLimitedRetryEnd(const ds::Status &status, int &limitedRetryTime)
{
    if (!IsLimitedRetryableStatus(status)) {
        limitedRetryTime = 0;
        return false;
    }
    limitedRetryTime++;
    return limitedRetryTime >= LIMITED_RETRY_TIME;
}

ErrorCode ConvertDatasystemErrorToCore(const ds::StatusCode &datasystemCode, const ErrorCode &defaultCode)
{
    auto iter = datasystemErrCodeMap.find(datasystemCode);
    if (iter != datasystemErrCodeMap.end()) {
        return datasystemErrCodeMap.at((uint32_t)datasystemCode);
    }
    return defaultCode;
}

ErrorInfo GenerateErrorInfo(const int &successCount, const ds::Status &status, const int &timeoutMS,
                            const std::vector<std::string> &remainIds, const std::vector<std::string> &ids)
{
    // 3. timeout, all ids fails then set err message
    ErrorInfo err;
    err.SetDsStatusCode(status.GetCode());
    std::ostringstream oss;
    if (successCount != 0) {
        if (IsUnlimitedRetryableStatus(status)) {
            return err;
        }
        err.SetErrorMsg(status.ToString());
        return err;
    }
    if (!IsUnlimitedRetryableStatus(status)) {
        oss << status.ToString();
    } else {
        oss << "Get timeout " << timeoutMS << "ms from datasystem,";
    }
    oss << " all failed: "
        << "(" << ids.size() << "). ";
    oss << "Failed objects: [ ";
    oss << remainIds[0] << " ... "
        << "]";
    err.SetErrCodeAndMsg(YR::Libruntime::ConvertDatasystemErrorToCore(status.GetCode()),
                         YR::Libruntime::ModuleCode::DATASYSTEM, oss.str(), status.GetCode());
    return err;
}

ErrorInfo GenerateSetErrorInfo(const ds::Status &status)
{
    ErrorInfo err;
    std::string msg = std::string("set KV error, errCode: ") + status.ToString();
    err.SetErrCodeAndMsg(YR::Libruntime::ConvertDatasystemErrorToCore(status.GetCode()),
                         YR::Libruntime::ModuleCode::DATASYSTEM, msg, status.GetCode());
    return err;
}

ErrorInfo ProcessKeyPartialResult(const std::vector<std::string> &keys,
                                  const std::vector<std::shared_ptr<Buffer>> &result, const ErrorInfo &errInfo,
                                  int timeoutMs)
{
    ErrorInfo err;
    std::vector<std::string> failKeys;
    bool isPartialResult = false;
    for (unsigned int i = 0; i < result.size(); i++) {
        if (!(result[i])) {  // result[i] is nullptr
            isPartialResult = true;
            failKeys.push_back(keys[i]);
        }
    }
    if (isPartialResult) {
        err.SetErrCodeAndMsg(YR::Libruntime::ErrorCode::ERR_GET_OPERATION_FAILED,
                             YR::Libruntime::ModuleCode::DATASYSTEM, errInfo.GetExceptionMsg(failKeys, timeoutMs),
                             errInfo.GetDsStatusCode());
    }
    return err;
}
}  // namespace Libruntime
}  // namespace YR