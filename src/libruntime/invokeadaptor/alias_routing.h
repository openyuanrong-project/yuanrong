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
#include <memory>
#include <unordered_map>

#include "absl/synchronization/mutex.h"

#include "alias_element.h"
#include "load_balancer.h"

namespace YR {
namespace Libruntime {

class AliasEntry;
class AliasRouting {
public:
    AliasRouting() = default;
    virtual ~AliasRouting() = default;
    void UpdateAliasInfo(const std::vector<AliasElement> &aliasInfo);
    std::string GetFuncVersionUrnWithParams(const std::string &aliasUrn,
                                            const std::unordered_map<std::string, std::string> &params);
    std::string ParseAlias(const std::string &functionId, const std::unordered_map<std::string, std::string> &params);
    bool CheckAlias(const std::string &functionId);

private:
    absl::Mutex mu_;
    std::unordered_map<std::string, std::shared_ptr<AliasEntry>> aliasInfo_ ABSL_GUARDED_BY(mu_);
};

struct AliasRule {
    std::string left;
    std::string op;
    std::string right;
};

class AliasEntry {
public:
    AliasEntry(const AliasElement &alias, std::unique_ptr<LoadBalancer> &&lb);
    virtual ~AliasEntry() = default;

    void Update(const AliasElement &alias);
    bool IsRoutingByRules(void);
    std::string GetFuncVersionUrnByRule(const std::unordered_map<std::string, std::string> &params);
    std::string GetFuncVersionUrn();

private:
    absl::Mutex mu_;
    AliasElement aliasElement_;
    std::vector<AliasRule> aliasParsedRules_;
    std::unique_ptr<LoadBalancer> lb_;
};
}  // namespace Libruntime
}  // namespace YR