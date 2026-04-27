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

#pragma once

#include "api/cpp/include/yr/api/future.h"
#include "src/libruntime/heterostore/hetero_future.h"
#include "src/utility/logger/logger.h"

namespace YR {
class HeteroFuture : public Future {
public:
    HeteroFuture() = default;
    explicit HeteroFuture(std::shared_ptr<YR::Libruntime::HeteroFuture> future);
    ~HeteroFuture() = default;
    void Get() override;

private:
    std::shared_ptr<YR::Libruntime::HeteroFuture> future_;
};
}  // namespace YR