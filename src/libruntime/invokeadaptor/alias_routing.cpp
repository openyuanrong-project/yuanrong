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

#include "alias_routing.h"
#include "re2/re2.h"
#include "load_balancer.h"
#include "src/utility/logger/logger.h"
#include "src/utility/string_utility.h"

namespace YR {
namespace Libruntime {
using namespace YR::utility;

const std::string RoutingTypeRule = "rule";
const size_t SplitedRuleNum = 3;
const int WEIGHT_MULTIPLIER = 100;

const std::string ALIAS_PATTERN_STRING = "^[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$";
const re2::RE2 ALIAS_PATTERN(ALIAS_PATTERN_STRING);
const int ALIAS_LENGTH_LIMIT = 32;

void AliasRouting::UpdateAliasInfo(const std::vector<AliasElement> &aliasInfo)
{
    YRLOG_INFO("recv aliases info");
    std::unordered_map<std::string, std::shared_ptr<AliasEntry>> newAliasInfo;
    for (auto &alias : aliasInfo) {
        if (newAliasInfo.find(alias.aliasUrn) == newAliasInfo.end()) {
            std::unique_ptr<LoadBalancer> lb(LoadBalancer::Factory(LoadBalancerType::WeightedRoundRobin));
            auto entry = std::make_shared<AliasEntry>(alias, std::move(lb));
            newAliasInfo.emplace(alias.aliasUrn, entry);
        } else {
            auto entry = newAliasInfo[alias.aliasUrn];
            entry->Update(alias);
        }
    }

    absl::WriterMutexLock lock(&mu_);
    aliasInfo_ = newAliasInfo;
}

std::string AliasRouting::GetFuncVersionUrnWithParams(const std::string &aliasUrn,
                                                      const std::unordered_map<std::string, std::string> &params)
{
    std::shared_ptr<AliasEntry> aliasEntry;
    {
        absl::ReaderMutexLock lock(&mu_);
        if (aliasInfo_.find(aliasUrn) == aliasInfo_.end()) {
            return aliasUrn;
        }
        aliasEntry = aliasInfo_[aliasUrn];
    }

    if (aliasEntry == nullptr) {
        YRLOG_ERROR("empty alias entry for alias URN: {}", aliasUrn);
        return aliasUrn;
    }

    if (aliasEntry->IsRoutingByRules()) {
        return aliasEntry->GetFuncVersionUrnByRule(params);
    }

    return aliasEntry->GetFuncVersionUrn();
}

static std::vector<AliasRule> ParseRules(const AliasElement &alias)
{
    std::vector<AliasRule> parsed(alias.routingRules.rules.size());
    for (size_t i = 0; i < alias.routingRules.rules.size(); ++i) {
        auto &rule = alias.routingRules.rules[i];
        std::vector<std::string> result;
        Split(rule, result, ':');
        if (result.size() != SplitedRuleNum) {
            YRLOG_ERROR("rule is not splited to size: {}, rule: {}", SplitedRuleNum, rule);
            return std::vector<AliasRule>();
        }

        parsed[i].left = result[0];   // left
        parsed[i].op = result[1];     // operation
        parsed[i].right = result[2];  // right
    }
    return parsed;
}

std::tuple<bool, std::string, std::string, std::string> parseFunctionId(const std::string &functionId)
{
    std::vector<std::string> result;
    Split(functionId, result, '/');
    if (result.size() != 3) {  // {tenantId}/0@default@{functionName}/{aliasOrVersion}
        return {false, "", "", ""};
    }

    return {true, result.at(0), result.at(1), result.at(2)};
};

bool AliasRouting::CheckAlias(const std::string &functionId)
{
    auto [parseOk, tenantId, functionName, aliasOrVersion] = parseFunctionId(functionId);
    if (!parseOk) {
        return false;
    }
    YRLOG_DEBUG("functionId {}, tenantId {}, functionName {}, aliasOrVersion {}",
                functionId, tenantId, functionName, aliasOrVersion);
    if (aliasOrVersion == "latest" || aliasOrVersion == "$latest") {
        return false;
    }
    return RE2::FullMatch(aliasOrVersion, ALIAS_PATTERN);
}

std::tuple<bool, std::string> functionIdToAliasUrn(const std::string &functionId)
{
    auto [parseOk, tenantId, functionFullName, alias] = parseFunctionId(functionId);
    if (!parseOk) {
        YRLOG_ERROR("functionId format error {}", functionId);
        return {false, ""};
    }

    std::ostringstream oss;
    oss << "sn:cn:yrk:" << tenantId << ":function:" << functionFullName << ":" << alias;
    return {true, oss.str()};
}

std::tuple<bool, std::string> functionVersionUrnToFunctionId(const std::string &functionVersionUrn)
{
    std::vector<std::string> result;

    // example: sn:cn:yrk:default:function:helloworld:$latest
    Split(functionVersionUrn, result, ':');
    if (result.size() != 7) {  //
        return {false, ""};
    }
    std::ostringstream oss;
    oss << result.at(3) << "/" << result.at(5) << "/" << result.at(6);   // {tenantId}/{functionName}/{version}
    return {true, oss.str()};
}

std::string AliasRouting::ParseAlias(const std::string &functionId,
                                     const std::unordered_map<std::string, std::string> &params)
{
    auto [ok, aliasUrn] = functionIdToAliasUrn(functionId);
    if (!ok) {
        YRLOG_ERROR("empty alias entry for alias URN: {}", aliasUrn);
        return functionId;
    }
    std::string functionVersionUrn = this->GetFuncVersionUrnWithParams(aliasUrn, params);

    auto [ok1, parsedFunctionId] = functionVersionUrnToFunctionId(functionVersionUrn);
    if (!ok1) {
        YRLOG_WARN("functionVersionUrnToFunctionId failed, urn: {}", functionVersionUrn); // just for clean check
    }

    YRLOG_INFO("parse alias {} to {}", functionId, parsedFunctionId);
    return parsedFunctionId;
}

static bool MatchRuleString(const std::string &val, const std::string &inStrings)
{
    std::vector<std::string> result;
    Split(inStrings, result, ',');
    for (auto &inStr : result) {
        if (!inStr.empty() && TrimSpace(val) == TrimSpace(inStr)) {
            return true;
        }
    }
    return false;
}

static bool MatchOneRule(const AliasRule &rule, const std::unordered_map<std::string, std::string> &params)
{
    if (params.find(rule.left) == params.end()) {
        YRLOG_ERROR("cannot find rule left {} in params", rule.left);
        return false;
    }

    const auto &val = params.at(rule.left);
    if (rule.op == "=") {
        return TrimSpace(val) == TrimSpace(rule.right);
    } else if (rule.op == "!=") {
        return TrimSpace(val) != TrimSpace(rule.right);
    } else if (rule.op == ">") {
        auto [result, err] = CompareIntFromString(val, rule.right);
        return err.empty() && result == 1;
    } else if (rule.op == "<") {
        auto [result, err] = CompareIntFromString(val, rule.right);
        return err.empty() && result == -1;
    } else if (rule.op == ">=") {
        auto [result, err] = CompareIntFromString(val, rule.right);
        return err.empty() && (result == 1 || result == 0);
    } else if (rule.op == "<=") {
        auto [result, err] = CompareIntFromString(val, rule.right);
        return err.empty() && (result == -1 || result == 0);
    } else if (rule.op == "in") {
        return MatchRuleString(val, rule.right);
    } else {
        YRLOG_ERROR("unknown operator in alias rule, op: {}", rule.op);
        return false;
    }
}

static bool MergeMatchResults(const std::list<bool> &results, const std::string &ruleLogic)
{
    if (results.empty()) {
        return false;
    }

    if (ruleLogic == "or") {
        auto it = std::find(results.begin(), results.end(), true);
        if (it != results.end()) {
            return true;
        }
        return false;
    } else if (ruleLogic == "and") {
        auto it = std::find(results.begin(), results.end(), false);
        if (it != results.end()) {
            return false;
        }
        return true;
    }
    YRLOG_ERROR("unknown alias rule logic: {}", ruleLogic);
    return false;
}

static bool MatchRules(const std::unordered_map<std::string, std::string> &params, const std::vector<AliasRule> &rules,
                       const std::string &ruleLogic)
{
    std::list<bool> results;
    for (auto &rule : rules) {
        results.push_back(MatchOneRule(rule, params));
    }
    return MergeMatchResults(results, ruleLogic);
}

AliasEntry::AliasEntry(const AliasElement &alias, std::unique_ptr<LoadBalancer> &&lb) : lb_(std::move(lb))
{
    Update(alias);
}

void AliasEntry::Update(const AliasElement &alias)
{
    absl::WriterMutexLock lock(&mu_);
    aliasElement_ = alias;
    aliasParsedRules_ = ParseRules(alias);
    lb_->RemoveAll();
    for (auto &conf : alias.routingConfig) {
        // rate
        lb_->Add(conf.functionVersionUrn, int(conf.weight * WEIGHT_MULTIPLIER));
    }
}

bool AliasEntry::IsRoutingByRules(void)
{
    return aliasElement_.routingType == RoutingTypeRule;
}

std::string AliasEntry::GetFuncVersionUrnByRule(const std::unordered_map<std::string, std::string> &params)
{
    absl::ReaderMutexLock lock(&mu_);
    if (params.empty() || aliasElement_.routingRules.rules.empty() || aliasParsedRules_.empty()) {
        YRLOG_ERROR("params or alias rules is empty, params size: {}, alias rules size: {}, parsed rule size: {}",
                    params.size(), aliasElement_.routingRules.rules.size(), aliasParsedRules_.size());
        return aliasElement_.functionVersionUrn;
    }

    if (MatchRules(params, aliasParsedRules_, aliasElement_.routingRules.ruleLogic)) {
        return aliasElement_.routingRules.grayVersion;
    }

    return aliasElement_.functionVersionUrn;
}

std::string AliasEntry::GetFuncVersionUrn()
{
    absl::ReaderMutexLock lock(&mu_);
    return lb_->Next("", false);
}

}  // namespace Libruntime
}  // namespace YR