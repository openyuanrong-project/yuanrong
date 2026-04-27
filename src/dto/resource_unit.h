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

#include "src/libruntime/fsclient/protobuf/resource.grpc.pb.h"

namespace YR {
namespace Libruntime {
struct ResourceUnit {
    std::string id;
    std::unordered_map<std::string, float> capacity;
    std::unordered_map<std::string, float> allocatable;
    std::unordered_map<std::string, std::vector<std::string>> nodeLabels;
    uint32_t status;
};

struct CommonSatus {
    int code;
    std::string message;
};

struct Resource {
    std::string name;
    enum class Type : int {
        SCALER = 0,
    };
    Type type;
    struct Scalar {
        double value;
        double limit;
    };
    Scalar scalar;
};

struct Resources {
    std::unordered_map<std::string, Resource> resources;
};

struct BundleInfo {
    std::string bundleId;
    std::string rGroupName;
    std::string parentRGroupName;
    std::string functionProxyId;
    std::string functionAgentId;
    std::string tenantId;
    Resources resources;
    std::vector<std::string> labels;
    CommonSatus status;
    std::string parentId;
    std::unordered_map<std::string, std::string> kvLabels;
};

struct Option {
    int priority;
    int groupPolicy;
    std::unordered_map<std::string, std::string> extension;
};

struct RgInfo {
    std::string name;
    std::string owner;
    std::string appId;
    std::string tenantId;
    std::vector<BundleInfo> bundles;
    CommonSatus status;
    std::string parentId;
    std::string requestId;
    std::string traceId;
    Option opt;
};

struct ResourceGroupUnit {
    std::unordered_map<std::string, RgInfo> resourceGroups;
};
}
}  // namespace YR