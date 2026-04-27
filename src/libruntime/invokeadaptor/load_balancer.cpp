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

#include "load_balancer.h"

#include <algorithm>
#include <map>
#include <vector>

#include "src/libruntime/utils/utils.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
std::hash<std::string> g_hashFunc;
struct WeightedNode {
    std::string name;
    int weight;
    int currentWeight;
};

struct AnchorInfo {
    size_t instanceHash;
    std::string instanceKey;
};

class WeightedRoundRobin : public LoadBalancer {
public:
    WeightedRoundRobin() = default;
    virtual ~WeightedRoundRobin() = default;

    std::string Next(const std::string &name, bool move = false) override
    {
        if (nodes_.size() == 0) {
            return "";
        }
        return NextNode().name;
    }

    void Add(const std::string &node, int weight) override
    {
        nodes_.emplace_back(WeightedNode{
            .name = node,
            .weight = weight,
            .currentWeight = 0,
        });
    }

    void RemoveAll(void) override
    {
        nodes_.clear();
    }

    void Remove(const std::string &name) override {}

private:
    WeightedNode &NextNode(void)
    {
        int total = 0;
        WeightedNode *best = &nodes_[0];
        for (auto &node : nodes_) {
            node.currentWeight += node.weight;
            total += node.weight;
            if (node.currentWeight > best->currentWeight) {
                best = &node;
            }
        }
        best->currentWeight -= total;
        return *best;
    }

    std::vector<WeightedNode> nodes_;
};

class CsHashRoundRobin : public LoadBalancer {
public:
    CsHashRoundRobin() = default;
    virtual ~CsHashRoundRobin() = default;

    std::string Next(const std::string &name, bool move = false) override
    {
        if (hashPool_.size() == 0) {
            YRLOG_WARN("current hash pool size is empty, return empty res directly");
            return ":0";
        }
        std::unique_lock<std::mutex> lk(insMtx_);
        if (auto it = anchorPoint_.find(name); it == anchorPoint_.end()) {
            YRLOG_DEBUG("current anchor point map has no value of id: {}", name);
            auto anchor = AddAnchorPoint(name);
            lk.unlock();
            if (hashPool_.find(anchor.instanceHash) != hashPool_.end()) {
                return anchor.instanceKey + ":" + std::to_string(hashPool_[anchor.instanceHash]);
            }
            return anchor.instanceKey + ":" + std::to_string(0);
        }
        auto anchor = anchorPoint_[name];
        if (move) {
            YRLOG_DEBUG("id: {} need remove anchor point, current anchor ins hash is {}, ins key is {}", name,
                        anchor.instanceHash, anchor.instanceKey);
            MoveAnchorPoint(name, anchor.instanceHash);
        }
        if (instanceMap_.find(anchor.instanceHash) == instanceMap_.end()) {
            YRLOG_DEBUG(
                "scheduler: {} not exist in instance map, need move anchor point, current ins hash is {}, "
                "ins key is {}",
                name, anchor.instanceHash, anchor.instanceKey);
            MoveAnchorPoint(name, anchor.instanceHash);
        }
        lk.unlock();
        if (hashPool_.find(anchorPoint_[name].instanceHash) != hashPool_.end()) {
            return anchorPoint_[name].instanceKey + ":" + std::to_string(hashPool_[anchorPoint_[name].instanceHash]);
        }
        return anchorPoint_[name].instanceKey + ":" + std::to_string(0);
    }

    void Add(const std::string &name, int weight) override
    {
        std::unique_lock<std::mutex> lk(insMtx_);
        auto hashKey = g_hashFunc(name);
        if (instanceMap_.find(hashKey) != instanceMap_.end()) {
            hashPool_[hashKey] = YR::GetCurrentTimestampNs();
            lk.unlock();
            YRLOG_DEBUG("scheduler: {} has already exist in instance map", name);
            return;
        }

        instanceMap_[hashKey] = name;

        hashPool_.emplace(hashKey, YR::GetCurrentTimestampNs());
        YRLOG_DEBUG("start add scheduler: {}, hashKey {} to hash ring, total nodes is {}", name, hashKey, totalNodes);
        totalNodes++;
        lk.unlock();
    }

    void RemoveAll(void) override
    {
        std::unique_lock<std::mutex> lk(insMtx_);
        anchorPoint_.clear();
        instanceMap_.clear();
        hashPool_.clear();
        totalNodes = 0;
        lk.unlock();
    }

    void Remove(const std::string &name) override
    {
        auto hashKey = g_hashFunc(name);
        std::unique_lock<std::mutex> lk(insMtx_);
        if (instanceMap_.find(hashKey) != instanceMap_.end()) {
            instanceMap_.erase(hashKey);
        }
        auto it = hashPool_.find(hashKey);
        if (it != hashPool_.end()) {
            YRLOG_DEBUG("remove scheduler : {}, total noedes is {}", name, totalNodes);
            hashPool_.erase(it);
            totalNodes--;
        }
        lk.unlock();
    }

private:
    std::unordered_map<std::string, AnchorInfo> anchorPoint_ = std::unordered_map<std::string, AnchorInfo>();
    std::unordered_map<size_t, std::string> instanceMap_ = std::unordered_map<size_t, std::string>();
    std::map<size_t, long long> hashPool_ = std::map<size_t, long long>();
    std::mutex insMtx_;
    size_t totalNodes = 0;

    void MoveAnchorPoint(const std::string &name, size_t currentHash)
    {
        auto instanceHash = GetNextHashKey(currentHash);
        if (auto it = anchorPoint_.find(name); it != anchorPoint_.end()) {
            if (instanceMap_.find(instanceHash) != instanceMap_.end()) {
                it->second.instanceKey = instanceMap_[instanceHash];
            }
            it->second.instanceHash = instanceHash;
            YRLOG_DEBUG("after move anchor point, instance hash of {} is {}, instance key is {}", name,
                        it->second.instanceHash, it->second.instanceKey);
        }
    }

    size_t GetNextHashKey(const size_t hashKey)
    {
        if (hashPool_.empty()) {
            return 0;
        }

        auto nextHashKey = hashPool_.begin()->first;
        for (auto v : hashPool_) {
            if (v.first > hashKey) {
                nextHashKey = v.first;
                break;
            }
        }
        return nextHashKey;
    }

    AnchorInfo AddAnchorPoint(const std::string &key)
    {
        size_t hashKey = g_hashFunc(key);

        size_t nextHashKey = GetNextHashKey(hashKey);
        AnchorInfo anchorInfo{.instanceHash = nextHashKey};
        if (instanceMap_.find(nextHashKey) != instanceMap_.end()) {
            anchorInfo.instanceKey = instanceMap_[nextHashKey];
        }
        anchorPoint_.emplace(key, anchorInfo);
        YRLOG_DEBUG("end add name: {}, instance hash key is {}, instance name is {}", key, anchorInfo.instanceHash,
                    anchorInfo.instanceKey);
        return anchorInfo;
    }
};

LoadBalancer *LoadBalancer::Factory(LoadBalancerType type)
{
    switch (type) {
        case LoadBalancerType::WeightedRoundRobin: {
            return new WeightedRoundRobin();
        }
        case LoadBalancerType::ConsistantRoundRobin: {
            return new CsHashRoundRobin();
        }
        default: {
            YRLOG_ERROR("unknown load balancer type {}", static_cast<int>(type));
            return new WeightedRoundRobin();
        }
    }
}
}  // namespace Libruntime
}  // namespace YR