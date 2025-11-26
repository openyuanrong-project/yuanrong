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

#include "yr/collective/collective.h"

#include <json.hpp>
#include <ostream>
#include <regex>
#include <vector>

#include "api/cpp/src/utils/utils.h"
#include "gloo_collective_group.h"
#include "src/dto/config.h"
#include "src/libruntime/err_type.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"
#include "yr/api/exception.h"
#include "yr/api/kv_manager.h"

namespace YR::Collective {

const std::string COLLECTIVE_GROUP_INFO_PREFIX = "collective-group-";

struct CollectiveGroupInfo {
    std::vector<std::string> instances;
    int worldSize;
    std::vector<int> ranks;
    const std::string groupName;
    Backend backend;
    int timeout;
    std::string prefix;

    std::string ToString()
    {
        nlohmann::json j = nlohmann::json{{"instances", instances}, {"worldSize", worldSize}, {"ranks", ranks},
                                          {"backend", backend},     {"timeout", timeout},     {"prefix", prefix}};
        return j.dump();
    }

    void FromJson(const std::string &str)
    {
        nlohmann::json j;
        try {
            j = nlohmann::json::parse(str);
        } catch (const std::exception &e) {
            YRLOG_WARN("json parse error: {}", e.what());
            return;
        }

        j.at("instances").get_to(instances);
        j.at("worldSize").get_to(worldSize);
        j.at("ranks").get_to(ranks);
        j.at("backend").get_to(backend);
        j.at("timeout").get_to(timeout);
        j.at("prefix").get_to(prefix);
    }
};

int CollectiveGroup::GetRank() const
{
    return rank_;
}

std::string CollectiveGroup::GetGroupName()
{
    return groupName_;
}

Backend CollectiveGroup::GetBackend()
{
    return backend_;
}

int CollectiveGroup::GetWorldSize() const
{
    return worldSize_;
}

void CheckGroupNameValid(const std::string &groupName)
{
    static const std::regex GROUP_NAME_REGEX(R"(^[a-zA-Z0-9\-_!#%\^\*\(\)\+\=\:;]*$)");
    THROW_IF_TRUE(!std::regex_match(groupName, GROUP_NAME_REGEX), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  R"(groupName is invalid. It should match the regex: ^[a-zA-Z0-9\-_!#%^\*\(\)\+\=\:;]*$)");
}

void InitCollectiveGroup(int worldSize, int rank, const std::string &groupName, Backend backend, int timeout,
                         const std::string &storePrefix)
{
    CheckInitialized();
    CollectiveGroupMgr::GetInstance().InitCollectiveGroup(worldSize, rank, groupName, backend, timeout, storePrefix);
}

void CreateCollectiveGroup(const std::vector<std::string> &instanceIDs, int worldSize, const std::vector<int> &ranks,
                           const std::string &groupName, Backend backend, int timeout)
{
    CheckInitialized();
    CheckGroupNameValid(groupName);

    THROW_IF_TRUE(
        instanceIDs.size() != static_cast<size_t>(worldSize) || static_cast<size_t>(worldSize) != ranks.size(),
        YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
        "failed to create collective group, unequal actor, rank or world size, please check");

    CollectiveGroupInfo info{.instances = instanceIDs,
                             .worldSize = worldSize,
                             .ranks = ranks,
                             .groupName = groupName,
                             .backend = backend,
                             .timeout = timeout,
                             .prefix = YR::utility::IDGenerator::GenObjectId()};

    try {
        YR::KVManager::Set(COLLECTIVE_GROUP_INFO_PREFIX + groupName, info.ToString(), YR::ExistenceOpt::NX);
    } catch (YR::Exception &e) {
        throw YR::Exception("collective group " + groupName + " already existed, please destroy it first");
    }
}

void DestroyCollectiveGroup(const std::string &groupName)
{
    CheckInitialized();
    CollectiveGroupMgr::GetInstance().DestroyCollectiveGroup(groupName);
}

void AllReduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op,
               const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    group->AllReduce(sendbuf, recvbuf, count, dtype, op);
}

void Reduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op, int dstRank,
            const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    group->Reduce(sendbuf, recvbuf, count, dtype, op, dstRank);
}

void AllGather(const void *sendbuf, void *recvbuf, int count, DataType dtype, const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    group->AllGather(sendbuf, recvbuf, count, dtype);
}

void Barrier(const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    group->Barrier();
}

void Scatter(const std::vector<void *> sendbuf, void *recvbuf, int count, DataType dtype, int srcRank,
             const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    group->Scatter(sendbuf, recvbuf, count, dtype, srcRank);
}

void Broadcast(const void *sendbuf, void *recvbuf, int count, DataType dtype, int srcRank, const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    group->Broadcast(sendbuf, recvbuf, count, dtype, srcRank);
}

void Recv(void *recvbuf, int count, DataType dtype, int srcRank, int tag, const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    group->Recv(recvbuf, count, dtype, srcRank, tag);
}

void Send(const void *sendbuf, int count, DataType dtype, int dstRank, int tag, const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    group->Send(sendbuf, count, dtype, dstRank, tag);
}

int GetWorldSize(const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    return group->GetWorldSize();
}

int GetRank(const std::string &groupName)
{
    CheckInitialized();
    auto group = CollectiveGroupMgr::GetInstance().CheckAndCreateGroup(groupName);
    THROW_IF_TRUE(group == nullptr, YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");
    return group->GetRank();
}

std::shared_ptr<CollectiveGroup> CollectiveGroupMgr::CheckAndCreateGroup(const std::string &groupName)
{
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    if (groups_.find(groupName) != groups_.end()) {
        return groups_[groupName];
    }
    auto str = YR::KVManager::Get(COLLECTIVE_GROUP_INFO_PREFIX + groupName);
    THROW_IF_TRUE(str.empty(), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                  "please create group " + groupName + "first");

    CollectiveGroupInfo info;
    info.FromJson(str);

    size_t index = 0;
    for (; index < info.instances.size(); ++index) {
        if (info.instances.at(index) == YR::Libruntime::Config::Instance().INSTANCE_ID()) {
            break;
        }
    }

    THROW_IF_TRUE(index == info.instances.size(), YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "invalid instance id");
    InitCollectiveGroup(info.worldSize, info.ranks[index], groupName, info.backend, info.timeout, info.prefix);
    return groups_[groupName];
}

void CollectiveGroupMgr::InitCollectiveGroup(int worldSize, int rank, const std::string &groupName, Backend backend,
                                             int timeout, const std::string &storePrefix)
{
    CheckGroupNameValid(groupName);
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    if (groups_.find(groupName) != groups_.end()) {
        YRLOG_DEBUG("collective group({}) already existed", groupName);
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "collective group already existed");
    }

    std::shared_ptr<CollectiveGroup> group;
    switch (backend) {
        case Backend::GLOO:
            group = std::make_shared<GlooCollectiveGroup>(groupName, worldSize, rank, timeout, storePrefix);
            break;

        case Backend::INVALID:
            // fall-through
        default:
            YRLOG_ERROR("failed to init collective group({}), invalid backend: {}", groupName, backend);
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "invalid collective group backend");
    }

    groups_[groupName] = group;
}

void CollectiveGroupMgr::DestroyCollectiveGroup(const std::string &groupName)
{
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    YR::KVManager::Del(COLLECTIVE_GROUP_INFO_PREFIX + groupName);
    groups_.erase(groupName);
}

CollectiveGroupMgr::~CollectiveGroupMgr()
{
    std::lock_guard<std::recursive_mutex> lock(mtx_);
    groups_.clear();
}

}  // namespace YR::collective