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

#include "object_id_pool.h"
#include "src/libruntime/utils/exception.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
ObjectIdPool::ObjectIdPool(std::shared_ptr<MemoryStore> memStore, size_t size) : memoryStore(memStore), poolSize(size)
{
    idPool.reserve(size);
}

std::pair<ErrorInfo, std::string> ObjectIdPool::Pop()
{
    std::lock_guard<std::mutex> lk(mu);
    if (idPool.empty()) {
        auto errorInfo = Scale();
        if (!errorInfo.OK()) {
            return std::make_pair<>(errorInfo, "");
        }
    }
    std::string objId = idPool.back();
    idPool.pop_back();
    return std::make_pair(ErrorInfo(), objId);
}

void ObjectIdPool::Clear()
{
    std::vector<std::string> toDecrease;
    {
        std::lock_guard<std::mutex> lk(mu);
        if (idPool.empty()) {
            return;
        }
        toDecrease.swap(idPool);
        idPool.clear();
    }
    memoryStore->DecreGlobalReference(toDecrease);
}

ErrorInfo ObjectIdPool::Scale()
{
    ErrorInfo err;
    auto generateKey = [this, &err](std::string &dsObjId, const std::string &objId) {
        err = this->memoryStore->GenerateKey(dsObjId, objId);
    };
    for (size_t i = 0; i < poolSize; ++i) {
        auto dsObjectId = YR::utility::IDGenerator::GenObjectId(generateKey);
        if (err.Code() != ErrorCode::ERR_OK) {
            idPool.clear();
            YRLOG_ERROR(err.Msg());
            return err;
        }
        idPool.emplace_back(dsObjectId);
    }
    err = memoryStore->IncreGlobalReference(idPool);
    if (err.Code() != ErrorCode::ERR_OK) {
        // clear pool.
        idPool.clear();
        YRLOG_ERROR(err.Msg());
    }
    return err;
}
}  // namespace Libruntime
}  // namespace YR