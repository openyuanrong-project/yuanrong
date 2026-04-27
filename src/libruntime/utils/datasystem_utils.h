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

#include <set>

#include <memory>
#include <vector>
#include "datasystem/utils/status.h"
#include "src/dto/buffer.h"
#include "src/dto/tensor.h"
#include "src/libruntime/err_type.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
#define RETURN_ERR_NOT_OK(flag, code, defaultCode, msg)                                       \
    do {                                                                                      \
        if (!(flag)) {                                                                        \
            ErrorInfo errInfo;                                                                \
            auto tmp = YR::Libruntime::ConvertDatasystemErrorToCore(code, defaultCode);       \
            errInfo.SetErrCodeAndMsg(tmp, YR::Libruntime::ModuleCode::DATASYSTEM, msg, code); \
            YRLOG_ERROR("occurs error, code is {}", static_cast<int>(code));                                    \
            return errInfo;                                                                   \
        }                                                                                     \
    } while (0)

#define THROW_EXCEPTION_ERR_NOT_OK(flag, code, defaultCode, msg)                                 \
    do {                                                                                         \
        if (!(flag)) {                                                                           \
            auto tmp = YR::Libruntime::ConvertDatasystemErrorToCore(code, defaultCode);          \
            throw YR::Libruntime::Exception(tmp, YR::Libruntime::ModuleCode::DATASYSTEM, (msg)); \
        }                                                                                        \
    } while (0)

bool IsRetryableStatus(const datasystem::Status &status);
bool IsUnlimitedRetryableStatus(const datasystem::Status &status);
bool IsLimitedRetryableStatus(const datasystem::Status &status);
bool IsLimitedRetryEnd(const datasystem::Status &status, int &limitedRetryTime);

ErrorCode ConvertDatasystemErrorToCore(const datasystem::StatusCode &datasystemCode,
                                       const ErrorCode &defaultCode = ErrorCode::ERR_DATASYSTEM_FAILED);

ErrorInfo GenerateErrorInfo(const int &successCount, const datasystem::Status &status, const int &timeoutMS,
                            const std::vector<std::string> &remainIds, const std::vector<std::string> &ids);
ErrorInfo GenerateSetErrorInfo(const datasystem::Status &status);
ErrorInfo ProcessKeyPartialResult(const std::vector<std::string> &keys,
                                  const std::vector<std::shared_ptr<Buffer>> &result, const ErrorInfo &errInfo,
                                  int timeoutMs);
}  // namespace Libruntime
}  // namespace YR