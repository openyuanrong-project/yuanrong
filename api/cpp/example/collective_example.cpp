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

#include <iostream>
#include <vector>
#include "yr/api/exception.h"
#include "yr/collective/collective.h"
#include "yr/yr.h"

class CollectiveActor {
public:
    int count;

    CollectiveActor() = default;

    ~CollectiveActor() = default;

    static CollectiveActor *FactoryCreate()
    {
        return new CollectiveActor();
    }

    int Compute(std::vector<int> in, std::string &groupName, uint8_t op)
    {
        //! [allreduce example]
        std::vector<int> output(in.size());

        // AllReduce: perform reduction across all processes and broadcast result to all
        YR::Collective::AllReduce(in.data(), output.data(), in.size(), YR::DataType::INT, YR::ReduceOp(op), groupName);

        // Barrier: synchronize all processes
        YR::Collective::Barrier(groupName);
        YR::Collective::DestroyCollectiveGroup(groupName);
        int result = 0;
        for (int i = 0; i < in.size(); ++i) {
            result += output[i];
        }
        return result;
        //! [allreduce example]
    }

    // Example: Initialize collective group in actor
    void InitCollectiveGroupExample()
    {
        //! [init collective group]
        YR::Collective::CollectiveGroupSpec spec;
        spec.worldSize = 4;
        spec.groupName = "my_group";
        spec.backend = YR::Collective::Backend::GLOO;
        spec.timeout = 60000;

        int rank = 0;  // Current process rank
        YR::Collective::InitCollectiveGroup(spec, rank);

        // Use collective communication operations...
        //! [init collective group]
    }

    // Example: Get world size
    void GetWorldSizeExample()
    {
        //! [get world size]
        // Initialize collective communication group...
        std::string groupName = "my_group";
        int worldSize = YR::Collective::GetWorldSize(groupName);
        std::cout << "World size: " << worldSize << std::endl;
        //! [get world size]
    }

    // Example: Get rank
    void GetRankExample()
    {
        //! [get rank]
        // Initialize collective communication group...
        std::string groupName = "my_group";
        int rank = YR::Collective::GetRank(groupName);
        std::cout << "My rank: " << rank << std::endl;
        //! [get rank]
    }

    // Example: Reduce operation
    void ReduceExample()
    {
        //! [reduce example]
        std::string groupName = "my_group";
        std::vector<int> localData(100, 1);
        std::vector<int> result(100);

        int rootRank = 0;  // Result sent to rank 0
        YR::Collective::Reduce(localData.data(), result.data(), localData.size(), YR::DataType::INT, YR::ReduceOp::SUM,
                               rootRank, groupName);

        int rank = YR::Collective::GetRank(groupName);
        if (rank == rootRank) {
            // Only rootRank's result contains valid result
            std::cout << "Reduced result: " << result[0] << std::endl;
        }
        //! [reduce example]
    }

    // Example: AllGather operation
    void AllGatherExample()
    {
        //! [allgather example]
        std::string groupName = "my_group";
        int rank = YR::Collective::GetRank(groupName);
        int worldSize = YR::Collective::GetWorldSize(groupName);

        // Each process sends different data
        std::vector<float> localData(10, static_cast<float>(rank));
        std::vector<float> gatheredData(10 * worldSize);

        YR::Collective::AllGather(localData.data(), gatheredData.data(), localData.size(), YR::DataType::FLOAT,
                                  groupName);
        //! [allgather example]
    }

    // Example: Broadcast operation
    void BroadcastExample()
    {
        //! [broadcast example]
        std::string groupName = "my_group";
        int rank = YR::Collective::GetRank(groupName);
        int srcRank = 0;

        std::vector<int> data(100);
        if (rank == srcRank) {
            // Source process initializes data
            for (int i = 0; i < 100; i++) {
                data[i] = i;
            }
        }

        YR::Collective::Broadcast(data.data(), data.data(), data.size(), YR::DataType::INT, srcRank, groupName);
        // All processes' data now contains the same content (from srcRank)
        //! [broadcast example]
    }

    // Example: Scatter operation
    void ScatterExample()
    {
        //! [scatter example]
        std::string groupName = "my_group";
        int rank = YR::Collective::GetRank(groupName);
        int worldSize = YR::Collective::GetWorldSize(groupName);
        int srcRank = 0;

        std::vector<int> recvData(10);

        if (rank == srcRank) {
            // Source process prepares data to scatter
            std::vector<std::vector<int>> sendData(worldSize, std::vector<int>(10));
            for (int i = 0; i < worldSize; i++) {
                for (int j = 0; j < 10; j++) {
                    sendData[i][j] = i * 10 + j;
                }
            }

            std::vector<void *> sendbuf;
            for (auto &vec : sendData) {
                sendbuf.push_back(vec.data());
            }

            YR::Collective::Scatter(sendbuf, recvData.data(), 10, YR::DataType::INT, srcRank, groupName);
        } else {
            // Non-source processes
            std::vector<void *> sendbuf;  // Can be empty
            YR::Collective::Scatter(sendbuf, recvData.data(), 10, YR::DataType::INT, srcRank, groupName);
        }

        // Each process's recvData now contains corresponding data from srcRank
        //! [scatter example]
    }

    // Example: Barrier synchronization
    void BarrierExample()
    {
        //! [barrier example]
        std::string groupName = "my_group";
        int rank = YR::Collective::GetRank(groupName);

        // Perform some asynchronous operations...
        std::cout << "Rank " << rank << " before barrier" << std::endl;

        YR::Collective::Barrier(groupName);

        // All processes will wait here until all processes arrive
        std::cout << "Rank " << rank << " after barrier" << std::endl;
        //! [barrier example]
    }

    // Example: Send Recv operation
    void SendRecvExample()
    {
        //! [send recv example]
        std::string groupName = "my_group";
        int rank = YR::Collective::GetRank(groupName);
        int worldSize = YR::Collective::GetWorldSize(groupName);

        if (rank == 0) {
            // Rank 0 sends data to Rank 1
            std::vector<float> data(100, 1.0f);
            YR::Collective::Send(data.data(), data.size(), YR::DataType::FLOAT, 1, 0, groupName);
        } else if (rank == 1) {
            // Rank 1 receives data from Rank 0
            std::vector<float> recvData(100);
            YR::Collective::Recv(recvData.data(), recvData.size(), YR::DataType::FLOAT, 0, 0, groupName);
        }
        //! [send recv example]
    }
};

YR_INVOKE(CollectiveActor::FactoryCreate, &CollectiveActor::Compute, &CollectiveActor::InitCollectiveGroupExample,
          &CollectiveActor::GetWorldSizeExample, &CollectiveActor::GetRankExample, &CollectiveActor::ReduceExample,
          &CollectiveActor::AllGatherExample, &CollectiveActor::BroadcastExample, &CollectiveActor::ScatterExample,
          &CollectiveActor::BarrierExample, &CollectiveActor::SendRecvExample)

int main(void)
{
    //! [create collective group]
    YR::Config conf;
    YR::Init(conf);

    // Create collective communication group with 4 instances
    std::vector<YR::NamedInstance<CollectiveActor>> instances;
    std::vector<std::string> instanceIDs;
    for (int i = 0; i < 4; ++i) {
        auto ins = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
        instances.push_back(ins);
        instanceIDs.push_back(ins.GetInstanceId());
    }

    std::string groupName = "test-group";
    YR::Collective::CollectiveGroupSpec spec{
        .worldSize = 4,
        .groupName = groupName,
        .backend = YR::Collective::Backend::GLOO,
        .timeout = YR::Collective::DEFAULT_COLLECTIVE_TIMEOUT,
    };
    YR::Collective::CreateCollectiveGroup(spec, instanceIDs, {0, 1, 2, 3});

    // Invoke collective operations on all instances
    std::vector<int> input = {1, 2, 3, 4};
    std::vector<YR::ObjectRef<int>> res;
    for (int i = 0; i < 4; ++i) {
        res.push_back(instances[i]
                          .Function(&CollectiveActor::Compute)
                          .Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::SUM)));
    }

    // Get results
    auto res0 = *YR::Get(res[0]);  // allreduce result
    auto res1 = *YR::Get(res[1]);  // send recv result(input)
    std::cout << "AllReduce result: " << res0 << ", Recv result: " << res1 << std::endl;

    // Destroy the collective group
    YR::Collective::DestroyCollectiveGroup(groupName);
    YR::Finalize();
    return 0;
    //! [create collective group]
}
