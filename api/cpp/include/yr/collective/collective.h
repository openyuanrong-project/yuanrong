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
#include <mutex>
#include <string>
#include <vector>
#include <unordered_map>
#include <utility>

#include "reduce_op.h"

namespace YR::Collective {
enum Backend : uint8_t {
    GLOO = 0,

    INVALID,
};

const int DEFAULT_COLLECTIVE_TIMEOUT = 60 * 1000;  // ms
const std::string DEFAULT_GROUP_NAME = "default";  // ms

class CollectiveGroup {
public:
    CollectiveGroup(std::string groupName, int worldSize, int rank, Backend backend, int timeout,
                    std::string storePrefix = "")
        : groupName_(std::move(groupName)),
          rank_(rank),
          backend_(backend),
          worldSize_(worldSize),
          timeout_(timeout),
          storePrefix_(std::move(storePrefix))
    {
    }

    virtual ~CollectiveGroup() = default;

    virtual void AllReduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op) = 0;

    virtual void Reduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op,
                        int dstRank) = 0;

    virtual void AllGather(const void *sendbuf, void *recvbuf, int count, DataType dtype) = 0;

    virtual void Barrier() = 0;

    virtual void Scatter(const std::vector<void *> sendbuf, void *recvbuf, int count, DataType dtype, int srcRank) = 0;

    virtual void Broadcast(const void *sendbuf, void *recvbuf, int count, DataType dtype, int srcRank) = 0;

    virtual void Recv(void *recvbuf, int count, DataType dtype, int srcRank, int tag) = 0;

    virtual void Send(const void *sendbuf, int count, DataType dtype, int dstRank, int tag) = 0;

    int GetRank() const;
    std::string GetGroupName();
    Backend GetBackend();
    int GetWorldSize() const;

protected:
    std::string groupName_;

    int rank_;

    Backend backend_;

    int worldSize_;

    int timeout_;

    std::string storePrefix_;
};

/**
 * init collective group in actor
 *
 * @param worldSize
 * @param rank
 * @param groupName
 */
void InitCollectiveGroup(int worldSize, int rank, const std::string &groupName = DEFAULT_GROUP_NAME,
                         Backend backend = Backend::GLOO, int timeout = DEFAULT_COLLECTIVE_TIMEOUT,
                         const std::string &storePrefix = "");

/**
 * create collective group with actor ids in driver
 *
 * @param instanceIDs
 * @param worldSize
 * @param ranks
 * @param groupName
 */
void CreateCollectiveGroup(const std::vector<std::string> &instanceIDs, int worldSize, const std::vector<int> &ranks,
                           const std::string &groupName = DEFAULT_GROUP_NAME, Backend backend = Backend::GLOO,
                           int timeout = DEFAULT_COLLECTIVE_TIMEOUT);
/**
 *
 *
 * @param groupName
 */
void DestroyCollectiveGroup(const std::string &groupName);

void AllReduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op,
               const std::string &groupName = DEFAULT_GROUP_NAME);

void Reduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op, int dstRank,
            const std::string &groupName = DEFAULT_GROUP_NAME);

void AllGather(const void *sendbuf, void *recvbuf, int count, DataType dtype,
               const std::string &groupName = DEFAULT_GROUP_NAME);

void Barrier(const std::string &groupName = DEFAULT_GROUP_NAME);

void Scatter(const std::vector<void *> sendbuf, void *recvbuf, int count, DataType dtype, int srcRank,
             const std::string &groupName = DEFAULT_GROUP_NAME);

void Broadcast(const void *sendbuf, void *recvbuf, int count, DataType dtype, int srcRank,
               const std::string &groupName = DEFAULT_GROUP_NAME);

void Recv(void *recvbuf, int count, DataType dtype, int srcRank, int tag = 0,
          const std::string &groupName = DEFAULT_GROUP_NAME);

void Send(const void *sendbuf, int count, DataType dtype, int dstRank, int tag = 0,
          const std::string &groupName = DEFAULT_GROUP_NAME);

int GetWorldSize(const std::string &groupName = DEFAULT_GROUP_NAME);

int GetRank(const std::string &groupName = DEFAULT_GROUP_NAME);

class CollectiveGroupMgr {
public:
    static CollectiveGroupMgr &GetInstance()
    {
        static CollectiveGroupMgr instance;
        return instance;
    }

    std::shared_ptr<CollectiveGroup> CheckAndCreateGroup(const std::string &groupName);

    void InitCollectiveGroup(int worldSize, int rank, const std::string &groupName, Backend backend, int timeout,
                             const std::string &storePrefix);

    void DestroyCollectiveGroup(const std::string &groupName);

private:
    CollectiveGroupMgr() = default;

    ~CollectiveGroupMgr();

    std::recursive_mutex mtx_{};

    std::unordered_map<std::string, std::shared_ptr<CollectiveGroup>> groups_{};
};

}  // namespace YR::collective
