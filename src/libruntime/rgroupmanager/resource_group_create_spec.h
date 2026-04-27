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

#include "src/dto/resource_group_spec.h"
#include "src/libruntime/fsclient/protobuf/common.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/core_service.grpc.pb.h"
#include "src/libruntime/utils/constants.h"

namespace YR {
namespace Libruntime {
using CreateResourceGroupRequest = ::core_service::CreateResourceGroupRequest;
using Bundle = ::common::Bundle;
using CommonResourceGroupSpec = ::common::ResourceGroupSpec;
using GroupPolicy = ::common::GroupPolicy;

const int DEFAULT_RG_CPU = 300;
const int DEFAULT_RG_MEMORY = 128;

struct ResourceGroupCreateSpec {
    ResourceGroupCreateSpec(const ResourceGroupSpec &rGroupSpec, const std::string &requestId,
                             const std::string &traceId, const std::string &jobId, const std::string &tenantId)
        : rGroupSpec(rGroupSpec), requestId(requestId), traceId(traceId), jobId(jobId), tenantId(tenantId)
    {
    }
    ~ResourceGroupCreateSpec() = default;

    ResourceGroupSpec rGroupSpec;

    std::string requestId;

    std::string traceId;

    std::string jobId;

    std::string tenantId;

    CreateResourceGroupRequest requestCreateRGroup;

    void BuildCreateResourceGroupRequest();
};
}  // namespace Libruntime
}  // namespace YR