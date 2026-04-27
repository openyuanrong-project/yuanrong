/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

#include "ObjectRef.h"

#include <iostream>
#include <string>

#include "Function.h"
#include "FunctionError.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/utility/logger/logger.h"

namespace Function {
const std::string ObjectRef::GetObjectRefId() const
{
    return objectRefId_;
}

const std::string ObjectRef::GetResult() const
{
    return result_;
}

bool ObjectRef::GetResultFlag() const
{
    return isResultExist_;
}

const std::string ObjectRef::Get()
{
    int defaultTimeout = 300 * 1000;
    auto [err, rets] =
        YR::Libruntime::LibruntimeManager::Instance().GetLibRuntime()->Get({objectRefId_}, defaultTimeout, false);
    if (!err.OK()) {
        YRLOG_WARN("failed to get result {}, err: {}", objectRefId_, err.Msg());
        throw FunctionError(ErrorCode::FUNCTION_EXCEPTION, err.Msg());
    }
    isResultExist_ = true;
    auto result = std::string(static_cast<const char *>(rets[0]->data->ImmutableData()), rets[0]->data->GetSize());
    if (result.empty()) {
        result_ = "";
        return result_;
    }
    std::string innerCode;
    try {
        auto resultJson = nlohmann::json::parse(result);
        result_ = resultJson["body"];
        innerCode = resultJson["innerCode"];
    } catch (nlohmann::detail::exception &e) {
        std::stringstream ss;
        ss << "failed to parse result, err: " << e.what();
        YRLOG_WARN(ss.str());
        throw FunctionError(ErrorCode::FUNCTION_EXCEPTION, ss.str());
    }
    if (innerCode != "0") {
        throw FunctionError(std::stoi(innerCode), result_);
    }
    return result_;
}

}  // namespace Function
