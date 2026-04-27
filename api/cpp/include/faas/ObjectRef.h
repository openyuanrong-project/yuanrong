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

#pragma once

#include <string>

namespace Function {
/*!
 * @class ObjectRef ObjectRef.h "include/faas/ObjectRef.h"
 *  @brief The ObjectRef class.
 */
class ObjectRef {
public:
    // not exposed
    ObjectRef(std::string &futureId, std::string &instanceId)
        : objectRefId_(futureId), instanceId_(instanceId), isResultExist_(false)
    {
    }

    // not exposed
    virtual ~ObjectRef() = default;

    // not exposed
    const std::string GetObjectRefId() const;
    // not exposed
    const std::string GetResult() const;
    
    /*!
     * @brief Get the result of the objectRef.
     * @return The result of the objectRef.
     */
    const std::string Get();

    // not exposed
    bool GetResultFlag() const;

private:
    std::string objectRefId_;
    std::string instanceId_;
    std::string result_;
    bool isResultExist_;
};
}  // namespace Function
