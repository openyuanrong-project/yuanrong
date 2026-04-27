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
#include <unordered_map>
#include <utility>
#include <vector>

#include "reduce_op.h"

namespace YR::Collective {
/*!
 * @enum Backend
 * @brief Backend type for collective communication.
 */
enum Backend : uint8_t {
    GLOO = 0,  ///< Use GLOO backend for collective communication.
    INVALID,   ///< Invalid backend type.
};

/*!
 * @var const int DEFAULT_COLLECTIVE_TIMEOUT
 * @brief Default timeout for collective communication operations in milliseconds.
 */
const int DEFAULT_COLLECTIVE_TIMEOUT = 60 * 1000;  // ms

/*!
 * @var const std::string DEFAULT_GROUP_NAME
 * @brief Default name for collective communication groups.
 */
const std::string DEFAULT_GROUP_NAME = "default";

/*!
 * @struct CollectiveGroupSpec
 * @brief Configuration specification for collective communication groups.
 */
struct CollectiveGroupSpec {
    int worldSize;  ///< Total number of processes in the group (world size).
    std::string groupName =
        "default";  ///< Name of the group, default is "default". Must match regex: ^[a-zA-Z0-9\-_!#%\^\*\(\)\+\=\:;]*$
    Backend backend = Backend::GLOO;           ///< Backend type to use, default is GLOO.
    int timeout = DEFAULT_COLLECTIVE_TIMEOUT;  ///< Operation timeout in milliseconds, default is 60000 ms (60 seconds).
};

/*!
 * @class CollectiveGroup
 * @brief Abstract base class for collective communication groups, used to manage collective communication operations in
 * distributed environments.
 *
 * The CollectiveGroup class provides a unified interface for collective communication operations, including AllReduce,
 * Reduce, AllGather, Broadcast, etc. Different backend implementations (such as GLOO) can inherit from this class and
 * provide specific implementations.
 */
class CollectiveGroup {
public:
    CollectiveGroup(const CollectiveGroupSpec &groupSpec, int rank, std::string storePrefix)
        : groupName_(groupSpec.groupName),
          rank_(rank),
          backend_(groupSpec.backend),
          worldSize_(groupSpec.worldSize),
          timeout_(groupSpec.timeout),
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

/*!
 * @brief Initializes a collective communication group in an actor instance.
 *
 * This function is used to initialize a collective communication group in the current actor instance.
 * Typically in distributed training or parallel computing scenarios, each process needs to call this function to join a
 * collective communication group.
 *
 * @param groupSpec Configuration specification for the collective communication group, including worldSize, groupName,
 * backend, and timeout.
 * @param rank Rank of the current process in the group, should be in the range [0, worldSize-1].
 * @param prefix Storage prefix for key-value storage used by backend communication. Default is an empty string.
 * @note Must be called after YR::Init(). The same groupName cannot be initialized repeatedly, otherwise an exception
 * will be thrown. groupName must match the regex: ^[a-zA-Z0-9\-_!#%\^\*\(\)\+\=\:;]*$
 *       Mixing CreateCollectiveGroup (in driver) and InitCollectiveGroup (in actor) for the same group is not
 * supported. All members of a group must use either CreateCollectiveGroup or InitCollectiveGroup exclusively. Dynamic
 * addition or removal of group members is not supported. Once a group is created, the member count is fixed.
 * @throws Exception Thrown if called before initialization, if groupName is invalid, or if the collective group already
 * exists.
 *
 * @snippet{trimleft} collective_example.cpp init collective group
 */
void InitCollectiveGroup(const CollectiveGroupSpec &groupSpec, int rank, const std::string &prefix = "");

/*!
 * @brief Creates a collective communication group in the driver using actor instance IDs.
 *
 * This function is used to create a collective communication group in the driver process, specifying the actor
 * instances participating in collective communication and their corresponding ranks. Typically in distributed training
 * scenarios, the driver process calls this function to create the group, and then each actor instance joins the group
 * through InitCollectiveGroup.
 *
 * @param groupSpec Configuration specification for the collective communication group.
 * @param instanceIDs List of actor instance IDs, size must equal worldSize.
 * @param ranks List of ranks corresponding to each instance, size must equal worldSize, and rank values should be in
 * the range [0, worldSize-1].
 * @note The size of instanceIDs must equal worldSize. The size of ranks must equal worldSize.
 *       If groupName already exists, an exception will be thrown. You need to call DestroyCollectiveGroup first to
 * destroy the existing group.
 *       Mixing CreateCollectiveGroup (in driver) and InitCollectiveGroup (in actor) for the same group is not
 * supported. All members of a group must use either CreateCollectiveGroup or InitCollectiveGroup exclusively. Dynamic
 * addition or removal of group members is not supported. Once a group is created, the member count is fixed.
 * @throws Exception Thrown if instanceIDs, ranks, and worldSize don't match, if groupName already exists, or if
 * groupName is invalid.
 *
 * @snippet{trimleft} collective_example.cpp create collective group
 */
void CreateCollectiveGroup(const CollectiveGroupSpec &groupSpec, const std::vector<std::string> &instanceIDs,
                           const std::vector<int> &ranks);

/*!
 * @brief Destroys the specified collective communication group.
 *
 * This function is used to destroy a created collective communication group and release related resources.
 *
 * @param groupName Name of the group to destroy.
 * @note If the group doesn't exist, this function won't throw an exception and will handle it silently.
 */
void DestroyCollectiveGroup(const std::string &groupName);

/*!
 * @brief Performs a reduction operation across all processes and broadcasts the result to all processes.
 *
 * This function performs a reduction operation (such as sum, max, etc.) across all processes and writes the result back
 * to all processes' recvbuf. All processes' recvbuf will eventually contain the same result.
 *
 * @param sendbuf Send buffer containing local input data. All processes' sendbuf should have the same size.
 * @param recvbuf Receive buffer for storing the reduction result. Size should be the same as sendbuf.
 * @param count Number of data elements.
 * @param dtype Data type (DataType::INT, DataType::FLOAT, DataType::DOUBLE, DataType::LONG).
 * @param op Reduction operator (ReduceOp::SUM, ReduceOp::PRODUCT, ReduceOp::MIN, ReduceOp::MAX).
 * @param groupName Name of the group, default is "default".
 * @note sendbuf and recvbuf can point to the same memory (in-place operation). All processes must call this function
 * with consistent parameters. Must be called after the group is created and initialized.
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp allreduce example
 */
void AllReduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op,
               const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @brief Performs a reduction operation across all processes and sends the result to the specified destination process.
 *
 * This function performs a reduction operation across all processes, but only the dstRank process's recvbuf will
 * contain the reduction result. The recvbuf content of other processes is undefined.
 *
 * @param sendbuf Send buffer containing local input data.
 * @param recvbuf Receive buffer, only used in the dstRank process, for storing the reduction result. Other processes'
 * recvbuf content is undefined.
 * @param count Number of data elements.
 * @param dtype Data type.
 * @param op Reduction operator.
 * @param dstRank Rank of the destination process where the reduction result will be sent.
 * @param groupName Name of the group, default is "default".
 * @note The recvbuf output data of non-root ranks (non-dstRank) is unreliable and should not be used. Only the dstRank
 * process's recvbuf contains valid reduction results.
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp reduce example
 */
void Reduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op, int dstRank,
            const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @brief Gathers data from all processes and broadcasts the result to all processes.
 *
 * This function gathers data from all processes and writes the gathered data back to all processes' recvbuf in rank
 * order. The size of recvbuf should be count * worldSize.
 *
 * @param sendbuf Send buffer containing local data to send.
 * @param recvbuf Receive buffer for storing data gathered from all processes. Size should be count * worldSize.
 * @param count Number of data elements sent by each process.
 * @param dtype Data type.
 * @param groupName Name of the group, default is "default".
 * @note The size of recvbuf must be at least count * worldSize. Gathered data is arranged in rank order in recvbuf.
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp allgather example
 */
void AllGather(const void *sendbuf, void *recvbuf, int count, DataType dtype,
               const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @brief Synchronization barrier that blocks until all processes in the group reach this point.
 *
 * This function is used to synchronize all processes, ensuring that all processes reach the Barrier call point before
 * executing subsequent code.
 *
 * @param groupName Name of the group, default is "default".
 * @note All processes must call this function, otherwise it will cause a deadlock. This function blocks until all
 * processes reach the Barrier call point.
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp barrier example
 */
void Barrier(const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @brief Scatters data from the source process to all processes.
 *
 * This function scatters data from the srcRank process's sendbuf vector to all processes.
 * The sendbuf vector of the srcRank process should contain worldSize buffers, each buffer corresponding to a target
 * rank.
 *
 * @param sendbuf Send buffer vector, only used in the srcRank process. Size should be worldSize, each element points to
 * data to be sent to the corresponding rank.
 * @param recvbuf Receive buffer for storing data received from srcRank.
 * @param count Number of data elements received by each process.
 * @param dtype Data type.
 * @param srcRank Rank of the source process.
 * @param groupName Name of the group, default is "default".
 * @note The sendbuf vector is only used in the srcRank process, other processes can pass an empty vector. The size of
 * the sendbuf vector must equal worldSize.
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp scatter example
 */
void Scatter(const std::vector<void *> sendbuf, void *recvbuf, int count, DataType dtype, int srcRank,
             const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @brief Broadcasts data from the source process to all processes.
 *
 * This function broadcasts data from the srcRank process to all processes in the group.
 * All processes' recvbuf will eventually contain the same data (from srcRank's sendbuf).
 *
 * @param sendbuf Send buffer, only used in the srcRank process, containing data to broadcast.
 * @param recvbuf Receive buffer for storing the broadcast data. All processes' recvbuf will eventually contain the same
 * data.
 * @param count Number of data elements.
 * @param dtype Data type.
 * @param srcRank Rank of the source process.
 * @param groupName Name of the group, default is "default".
 * @note sendbuf and recvbuf can point to the same memory (in-place operation). All processes must call this function
 * with consistent parameters.
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp broadcast example
 */
void Broadcast(const void *sendbuf, void *recvbuf, int count, DataType dtype, int srcRank,
               const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @brief Receives data from the specified process.
 *
 * This function is used for point-to-point communication, receiving data from the specified source process.
 * Must be paired with the Send function, matched through the tag parameter.
 *
 * @param recvbuf Receive buffer for storing received data.
 * @param count Number of data elements to receive.
 * @param dtype Data type.
 * @param srcRank Rank of the source process.
 * @param tag Message tag for matching the corresponding Send operation, default is 0.
 * @param groupName Name of the group, default is "default".
 * @note Recv and Send must be paired, and tags must match. Recv operation blocks until the corresponding Send operation
 * is called. count and dtype must match the corresponding Send operation.
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp send recv example
 */
void Recv(void *recvbuf, int count, DataType dtype, int srcRank, int tag = 0,
          const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @brief Sends data to the specified process.
 *
 * This function is used for point-to-point communication, sending data to the specified destination process.
 * Must be paired with the Recv function, matched through the tag parameter.
 *
 * @param sendbuf Send buffer containing data to send.
 * @param count Number of data elements to send.
 * @param dtype Data type.
 * @param dstRank Rank of the destination process.
 * @param tag Message tag for matching the corresponding Recv operation, default is 0.
 * @param groupName Name of the group, default is "default".
 * @note Send and Recv must be paired, and tags must match. Send operation blocks until the corresponding Recv operation
 * is called.
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp send recv example
 */
void Send(const void *sendbuf, int count, DataType dtype, int dstRank, int tag = 0,
          const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @brief Gets the total number of processes in the specified group.
 *
 * This function returns the total number of processes (world size) in the collective communication group.
 *
 * @param groupName Name of the group, default is "default".
 * @return The total number of processes (world size) in the group.
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp get world size
 */
int GetWorldSize(const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @brief Gets the rank of the current process in the specified group.
 *
 * This function returns the rank of the current process in the collective communication group.
 *
 * @param groupName Name of the group, default is "default".
 * @return The rank of the current process in the group, in the range [0, worldSize-1].
 * @throws Exception Thrown if the group doesn't exist or hasn't been created yet.
 *
 * @snippet{trimleft} collective_example.cpp get rank
 */
int GetRank(const std::string &groupName = DEFAULT_GROUP_NAME);

/*!
 * @class CollectiveGroupMgr
 * @brief Manager class for collective communication groups (internal use).
 *
 * This class manages the lifecycle of collective communication groups using a singleton pattern.
 * It is used internally by the collective communication API functions.
 */
class CollectiveGroupMgr {
public:
    /*!
     * @brief Gets the singleton instance of CollectiveGroupMgr.
     * @return Reference to the singleton instance.
     */
    static CollectiveGroupMgr &GetInstance()
    {
        static CollectiveGroupMgr instance;
        return instance;
    }

    /*!
     * @brief Checks if a group exists, and creates it if it doesn't exist.
     * @param groupName Name of the group.
     * @return Shared pointer to the CollectiveGroup instance.
     */
    std::shared_ptr<CollectiveGroup> CheckAndCreateGroup(const std::string &groupName);

    /*!
     * @brief Initializes a collective communication group (internal use).
     * @param groupSpec Configuration specification for the collective communication group.
     * @param rank Rank of the current process.
     * @param prefix Storage prefix for backend communication.
     */
    void InitCollectiveGroup(const CollectiveGroupSpec &groupSpec, int rank, const std::string &prefix);

    /*!
     * @brief Destroys a collective communication group (internal use).
     * @param groupName Name of the group to destroy.
     */
    void DestroyCollectiveGroup(const std::string &groupName);

private:
    CollectiveGroupMgr() = default;

    ~CollectiveGroupMgr();

    std::recursive_mutex mtx_{};

    std::unordered_map<std::string, std::shared_ptr<CollectiveGroup>> groups_{};
};

}  // namespace YR::Collective
