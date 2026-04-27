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

#pragma once
#include <list>
#include <memory>
#include <unordered_map>
#include <vector>

#include <boost/type_traits/is_same.hpp>

#include "yr/api/buffer.h"
#include "yr/api/err_type.h"
#include "yr/api/object_ref.h"
#include "yr/api/serdes.h"

namespace YR {
namespace internal {

struct RetryInfo {
    bool needRetry;
    ErrorInfo errorInfo;
};

enum class GetStatus : int {
    ALL_SUCCESS,
    PARTIAL_SUCCESS,
    PARTIAL_SUCCESS_AND_TIMEOUT,
    ALL_FAILED,
    ALL_FAILED_AND_TIMEOUT,
};

template <typename T>
void CheckIfObjectRefsHomogeneous(const std::vector<ObjectRef<T>> &objs)
{
    if (objs.empty()) {
        return;
    }
    bool isLocal = objs[0].IsLocal();
    bool result =
        std::all_of(objs.begin(), objs.end(), [isLocal](const ObjectRef<T> &obj) { return obj.IsLocal() == isLocal; });
    if (!result) {
        throw YR::Exception::InvalidParamException("cannot mix local and cluster object refs");
    }
}

template <typename T>
void CheckObjs(const std::vector<ObjectRef<T>> &objs)
{
    CheckInitialized();
    if (objs.empty()) {
        throw YR::Exception::InvalidParamException("Get does not accept empty object list");
    }
    CheckIfObjectRefsHomogeneous(objs);
}

inline void CheckTimeout(int timeoutSec)
{
    if (timeoutSec < NO_TIMEOUT) {
        std::string msg = "get config timeout (" + std::to_string(timeoutSec) + " s) is invalid";
        throw YR::Exception::InvalidParamException(msg);
    }
}

template <typename T>
void ExtractSuccessObjects(std::vector<std::string> &remainIds,
                           const std::vector<std::shared_ptr<Buffer>> &remainBuffers,
                           std::vector<std::shared_ptr<T>> &returnObjects,
                           std::unordered_map<std::string, std::list<size_t>> &idToIndices)
{
    std::vector<std::string> newRemainIds;
    newRemainIds.reserve(remainIds.size());
    for (size_t i = 0; i < remainIds.size(); i++) {
        if ((i < remainBuffers.size()) && remainBuffers[i]) {
            auto &indices = idToIndices[remainIds[i]];
            if constexpr (boost::is_same<T, Buffer>::value) {
                returnObjects[indices.front()] = remainBuffers[i];
            } else {
                auto obj = YR::internal::Deserialize<std::shared_ptr<T>>(remainBuffers[i]);
                returnObjects[indices.front()] = obj;
            }
            indices.pop_front();
        } else {
            newRemainIds.emplace_back(std::move(remainIds[i]));
        }
    }
    remainIds.swap(newRemainIds);
}

void ThrowExceptionBasedOnStatus(const GetStatus status, const ErrorInfo &err,
                                 const std::vector<std::string> &remainIds, int timeoutMS, bool allowPartial);
}  // namespace internal
}  // namespace YR