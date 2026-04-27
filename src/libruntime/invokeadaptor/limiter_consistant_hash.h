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

#include <atomic>

#include "load_balancer.h"
#include "src/libruntime/invokeadaptor/scheduler_instance_info.h"
#include "src/utility/logger/logger.h"

const int DEFAULT_CH_CONCURRENCY = 100;
namespace YR {
namespace Libruntime {
using namespace std::chrono;

struct LimiterNode {
    std::string schedulerId;
    time_point<steady_clock> lastTime;
    std::shared_ptr<LimiterNode> next;
    LimiterNode() = default;
    LimiterNode(std::string &key) : schedulerId(key), lastTime(steady_clock::now()), next(nullptr) {}
};

struct ConcurrentLimiter {
    std::shared_ptr<LimiterNode> head;
};

class LimiterCsHash {
public:
    LimiterCsHash() = default;
    LimiterCsHash(std::shared_ptr<LoadBalancer> loadBalancer) : loadBalancer_(loadBalancer) {}
    ~LimiterCsHash() = default;
    void Add(const std::string &schedulerName, const std::string &schedulerId, int weight = 0);
    std::string Next(const std::string &functionId, std::shared_ptr<AvailableSchedulerInfos> schedulerInfos,
                     bool move = false);
    std::string Next(const std::string &functionId, bool move = false);
    std::string NextRetry(const std::string &functionId, bool move = false);
    std::string NextRetry(const std::string &functionId, std::shared_ptr<AvailableSchedulerInfos> schedulerInfos,
                          bool move = false);
    bool isAllInsUnavailable(const std::string &instanceKey, long long instanceHash,
                             std::shared_ptr<AvailableSchedulerInfos> schedulerInfos);
    void Remove(const std::string &name);
    void RemoveAll();
    void ResetAll(const std::vector<SchedulerInstance> &schedulerInfoList, int weight = 0);
    bool IsSameWithHashPool(const std::vector<SchedulerInstance> &schedulerInfoList);

private:
    void AddWithoutLock(const std::string &schedulerName, const std::string &schedulerId, int weight = 0);
    std::shared_ptr<LoadBalancer> loadBalancer_;
    std::unordered_map<std::string, std::string> schedulerInfoMap;
    std::unordered_map<std::string, std::shared_ptr<ConcurrentLimiter>> limiters_;
    std::mutex limiterMtx_;
    std::atomic<int> nodeCount_{0};
    int limiterTime_ = 1;
};

}  // namespace Libruntime
}  // namespace YR