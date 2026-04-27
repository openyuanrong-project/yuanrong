/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.
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
#include <utility>
#include <vector>

#include "src/dto/accelerate.h"
#include "src/dto/internal_wait_result.h"
#include "src/dto/invoke_arg.h"
#include "src/dto/invoke_options.h"
#include "src/dto/resource_unit.h"
#include "src/dto/stream_conf.h"
#include "src/libruntime/clientsmanager/clients_manager.h"
#include "src/libruntime/connect/domain_socket_client.h"
#include "src/libruntime/connect/message_coder.h"
#include "src/libruntime/dependency_resolver.h"
#include "src/libruntime/driverlog/driverlog_receiver.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/event_notify.h"
#include "src/libruntime/fiber.h"
#include "src/libruntime/fmclient/fm_client.h"
#include "src/libruntime/fsclient/fs_client.h"
#include "src/libruntime/generator/generator_notifier.h"
#include "src/libruntime/generator/generator_receiver.h"
#include "src/libruntime/invokeadaptor/invoke_adaptor.h"
#include "src/libruntime/libruntime_config.h"
#include "src/libruntime/metricsadaptor/metrics_adaptor.h"
#include "src/libruntime/objectstore/object_id_pool.h"
#include "src/libruntime/streamstore/stream_producer_consumer.h"
#include "src/libruntime/utils/constants.h"
#include "src/libruntime/utils/security.h"
#include "src/libruntime/waiting_object_manager.h"
#include "src/utility/thread_pool.h"

namespace YR {
namespace Libruntime {
using WaitAsyncCallback = std::function<void(const std::string &, const ErrorInfo &, void *userData)>;
using GetAsyncCallback = std::function<void(std::shared_ptr<DataObject>, const ErrorInfo &, void *userData)>;
using GetEventCallback = std::function<void(std::shared_ptr<DataObject>, const ErrorInfo &, void *userData)>;
using FiberEventNotify = EventNotify<boost::fibers::mutex, boost::fibers::condition_variable>;
using FunctionLog = ::libruntime::FunctionLog;
const int DEFAULT_TIMEOUT = 60;       // second
const int DEFAULT_TIMEOUT_SEC = 300;  // second
const char *const RGROUP_NAME = "rgroup";
const char *const RGROUP_BUNDLE_PREFIX = "rg_";
const char *const RGROUP_BUNDLE_SUFFIX = "_bundle";
typedef void *StateStorePtr;
class Libruntime {
public:
    /*!
        @brief Libruntime 构造函数
        @param config Libruntime的参数配置，详细参考：@ref LibruntimeConfig
        @param clientsMgr 连接池，用于管理对函数系统及数据系统的连接相关连接
        @param metricsAdaptor 对接metrics后端，上报metrics信息
        @param security 设置安全相关信息
        @param socketClient 创建一个domain socket客户端，用于上报用户日志到RuntimeManager
    */
    Libruntime(std::shared_ptr<LibruntimeConfig> config, std::shared_ptr<ClientsManager> clientsMgr,
               std::shared_ptr<MetricsAdaptor> metricsAdaptor, std::shared_ptr<Security> security,
               std::shared_ptr<DomainSocketClient> socketClient);
    /*!
      @brief 初始化 Libruntime
      @param fsClient 函数系统客户端，支持GRPC Server、GRPC client、Http client这三种模式
      @param dsClients 数据系统客户端的集合，需要传入对象缓存、KV缓存、流缓存客户端
      @param cb 优雅退出的回调函数，在runtime中收到Shutdown请求时触发
      @return Error information
     */
    ErrorInfo Init(std::shared_ptr<FSClient> fsClient, DatasystemClients &dsClients, FinalizeCallback cb = nullptr);

    /*!
      @brief 创建一个 openYuanRong 函数
             该接口为异步接口，返回的 Instance ID 并非真实 ID
             可以通过调用 @ref GetRealInstanceId 获取函数创建结果
      @param functionMeta openYuanRong 函数元信息
      @param invokeArgs 函数创建参数，将会透传到函数实例
      @param opts 函数创建配置
      @return ErrorInfo： 函数创建的错误信息， string：Instance ID（非函数实例真实ID）
     */
    virtual std::pair<ErrorInfo, std::string> CreateInstance(const YR::Libruntime::FunctionMeta &functionMeta,
                                                             std::vector<YR::Libruntime::InvokeArg> &invokeArgs,
                                                             YR::Libruntime::InvokeOptions &opts);

    /*!
      @brief 通过 Instance ID 调用某一个函数实例
             该接口为异步接口，需要通过 returnObjs 获取函数调用结果
             可以通过调用 @ref Wait 获取函数调用状态
             也可以通过调用 @ref Get 获取函数调用结果
      @param funcMeta openYuanRong 函数元信息
      @param instanceId openYuanRong 函数实例 ID
      @param args 函数调用参数，将会透传到函数实例
      @param opts 函数调用配置
      @param returnObjs 函数返回值关联的 ID，支持多返回值
      @return 错误信息
     */
    virtual ErrorInfo InvokeByInstanceId(const YR::Libruntime::FunctionMeta &funcMeta, const std::string &instanceId,
                                         std::vector<YR::Libruntime::InvokeArg> &args,
                                         YR::Libruntime::InvokeOptions &opts, std::vector<DataObject> &returnObjs);

    /*!
      @brief 通过函数名调用函数
             该方法会为用户的调用自动调度一个函数实例，目前支持的调度策略为 Bin-Packing、Priority，详见 @ref
      InvokeOptions 该方法也支持调度FaaS函数请求，需要在 @ref InvokeOptions 中配置 schedulerInstanceIds
      @param funcMeta openYuanRong 函数元信息
      @param args 函数调用参数，将会透传到函数实例
      @param opt 函数调用配置
      @param returnObjs 函数返回值关联的 ID，支持多返回值
      @return 错误信息
     */
    virtual ErrorInfo InvokeByFunctionName(const YR::Libruntime::FunctionMeta &funcMeta,
                                           std::vector<YR::Libruntime::InvokeArg> &args,
                                           YR::Libruntime::InvokeOptions &opt, std::vector<DataObject> &returnObjs);

    /*!
      @brief Create an instance using raw data
      @param reqRaw Raw request data
      @param cb Callback for raw response
     */
    virtual void CreateInstanceRaw(std::shared_ptr<Buffer> reqRaw, RawCallback cb);

    /*!
      @brief Invoke a function by instance ID using raw data
      @param reqRaw Raw request data
      @param cb Callback for raw response
     */
    virtual void InvokeByInstanceIdRaw(std::shared_ptr<Buffer> reqRaw, RawCallback cb);

    /*!
      @brief Kill an instance using raw data
      @param reqRaw Raw request data
      @param cb Callback for raw response
     */
    virtual void KillRaw(std::shared_ptr<Buffer> reqRaw, RawCallback cb);

    /*!
      @brief Wait for a specified number of objects to be ready
      @param objs the list of object IDs to wait for
      @param waitNum the number of objects that need to be ready
      @param timeoutSec the timeout in seconds
      @return a pair to the wait result, ready ids and unready ids
      @throw Exception if the wait operation fails or times out
     */
    virtual std::shared_ptr<YR::InternalWaitResult> Wait(const std::vector<std::string> &objs, std::size_t waitNum,
                                                         int timeoutSec);

    /*!
      @brief Stores a data object with nested IDs
      @param dataObj A shared pointer to the `DataObject` to be stored
      @param nestedIds A set of strings containing the IDs of nested objects
      @param createParam Additional creation parameters (default is empty)
      @return A pair containing an `ErrorInfo` object and a string representing the object ID
     */
    virtual std::pair<ErrorInfo, std::string> Put(std::shared_ptr<DataObject> dataObj,
                                                  const std::unordered_set<std::string> &nestedIds,
                                                  const CreateParam &createParam = {});

    /*!
      @brief Stores a object with a specified ID and nested IDs
      @param objId The ID of the object to be stored
      @param dataObj A shared pointer to the `DataObject` to be stored
      @param nestedId A set of strings containing the IDs of nested objects
      @param createParam Additional creation parameters (default is empty)
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo Put(const std::string &objId, std::shared_ptr<DataObject> dataObj,
                          const std::unordered_set<std::string> &nestedId, const CreateParam &createParam = {});

    virtual ErrorInfo Put(std::shared_ptr<Buffer> data, const std::string &objID,
                          const std::unordered_set<std::string> &nestedID, bool toDataSystem,
                          const CreateParam &createParam = {});
    /*!
      @brief Stores a object with with a specified ID and nested IDs Access DataSystem without MemoryStore
      @details Access DataSystem without MemoryStore.
      @param objId The ID of the object to be stored
      @param data A shared pointer to the `Buffer` to be stored
      @param nestedId A set of strings containing the IDs of nested objects
      @param createParam Additional creation parameters
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo PutRaw(const std::string &objId, std::shared_ptr<Buffer> data,
                             const std::unordered_set<std::string> &nestedId, const CreateParam &createParam);

    /*!
      @brief Retrieves multiple data objects by their IDs
      @param ids A vector of strings containing the IDs of the data objects to retrieve
      @param timeoutMs The timeout in milliseconds for the operation
      @param allowPartial If true, partial retrieval will not return an error
      @return A pair containing an `ErrorInfo` object and a vector of shared pointers to the retrieved data objects
     */
    virtual std::pair<ErrorInfo, std::vector<std::shared_ptr<DataObject>>> Get(const std::vector<std::string> &ids,
                                                                               int timeoutMs, bool allowPartial);

    /*!
      @brief Retrieves multiple raw buffers by their IDs
      @details Access DataSystem without MemoryStore.
      @param ids A vector of strings containing the IDs of the buffers to retrieve
      @param timeoutMs The timeout in milliseconds for the operation
      @param allowPartial If true, partial retrieval will not throw an error
      @return A pair containing an `ErrorInfo` object and a vector of shared pointers to the retrieved buffers
     */
    virtual std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> GetRaw(const std::vector<std::string> &ids,
                                                                              int timeoutMs, bool allowPartial);

    /*!
      @brief Increases the reference count of the specified objects
      @param objIds A vector of strings containing the IDs of the objects
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo IncreaseReference(const std::vector<std::string> &objIds, bool toDatasystem = true);

    /*!
      @brief Increases the reference count of the specified objects with a remote ID
      @param objIds A vector of strings containing the IDs of the objects
      @param remoteId The ID of the remote context
      @return A pair containing an `ErrorInfo` object and a vector of strings representing the failed object IDs
     */
    virtual std::pair<ErrorInfo, std::vector<std::string>> IncreaseReference(const std::vector<std::string> &objIds,
                                                                             const std::string &remoteId);

    /*!
      @brief Increases the reference count of the specified raw objects
      @details Access DataSystem without MemoryStore.
      @param objIds A vector of strings containing the IDs of the objects
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo IncreaseReferenceRaw(const std::vector<std::string> &objIds);

    /*!
      @brief Increases the reference count of the specified objects with a remote ID
      @details Access DataSystem without MemoryStore.
      @param objIds A vector of strings containing the IDs of the objects
      @param remoteId The ID of the remote context
      @return A pair containing an `ErrorInfo` object and a vector of strings representing the failed object IDs
     */
    virtual std::pair<ErrorInfo, std::vector<std::string>> IncreaseReferenceRaw(const std::vector<std::string> &objIds,
                                                                                const std::string &remoteId);

    /*!
      @brief Decreases the reference count of the specified objects
      @param objIds A vector of strings containing the IDs of the objects
     */
    virtual void DecreaseReference(const std::vector<std::string> &objIds);

    /*!
      @brief Decreases the reference count of the specified objects with a remote ID
      @param objIds A vector of strings containing the IDs of the objects
      @param remoteId The ID of the remote context
      @return A pair containing an `ErrorInfo` object and a vector of strings representing the failed object IDs
     */
    virtual std::pair<ErrorInfo, std::vector<std::string>> DecreaseReference(const std::vector<std::string> &objIds,
                                                                             const std::string &remoteId);

    /*!
      @brief Decreases the reference count of the specified objects
      @details Access DataSystem without MemoryStore.
      @param objIds A vector of strings containing the IDs of the objects
     */
    virtual void DecreaseReferenceRaw(const std::vector<std::string> &objIds);

    /*!
      @brief Decreases the reference count of the specified objects with a remote ID
      @param objIds A vector of strings containing the IDs of the objects
      @param remoteId The ID of the remote context
      @return A pair containing an `ErrorInfo` object and a vector of strings representing the failed object IDs
     */
    virtual std::pair<ErrorInfo, std::vector<std::string>> DecreaseReferenceRaw(const std::vector<std::string> &objIds,
                                                                                const std::string &remoteId);

    /*!
      @brief Releases global references associated with a remote ID
      @param remoteId The ID of the remote context
      @return An `ErrorInfo` object indicating the success or failure of the operation
    */
    virtual ErrorInfo ReleaseGRefs(const std::string &remoteId);

    /*!
      @brief Allocates and initializes a return object with specified metadata and data sizes
      This function allocates a `DataObject` and initializes it with the provided metadata size, data size,
      and a list of nested object IDs. It also calculates and returns the total native buffer size.
      @param returnObj A shared pointer to the `DataObject` to be allocated and initialized
      @param metaSize The size of the metadata to be allocated
      @param dataSize The size of the data to be allocated
      @param nestedObjIds A vector of strings containing the IDs of nested objects
      @param totalNativeBufferSize A reference to a uint64_t variable to store the total native buffer size
      @return An `ErrorInfo` object indicating the success or failure of the allocation
     */
    virtual ErrorInfo AllocReturnObject(std::shared_ptr<DataObject> &returnObj, size_t metaSize, size_t dataSize,
                                        const std::vector<std::string> &nestedObjIds, uint64_t &totalNativeBufferSize);

    /*!
      @brief Allocates and initializes a return object with specified metadata and data sizes
      @param returnObj A pointer to the `DataObject` to be allocated and initialized
      @param metaSize The size of the metadata to be allocated
      @param dataSize The size of the data to be allocated
      @param nestedObjIds A vector of strings containing the IDs of nested objects
      @param totalNativeBufferSize A reference to a uint64_t variable to store the total native buffer size
      @return An `ErrorInfo` object indicating the success or failure of the allocation
     */
    virtual ErrorInfo AllocReturnObject(DataObject *returnObj, size_t metaSize, size_t dataSize,
                                        const std::vector<std::string> &nestedObjIds, uint64_t &totalNativeBufferSize);

    /*!
      @brief Creates a buffer with the specified data size
      @param dataSize The size of the data to be allocated in the buffer
      @param dataBuf A shared pointer to the `Buffer` to be created
      @return A pair containing an `ErrorInfo` object and a string representing the buffer ID
     */
    virtual std::pair<ErrorInfo, std::string> CreateBuffer(size_t dataSize, std::shared_ptr<Buffer> &dataBuf);

    /*!
      @brief Retrieves multiple buffers by their IDs
      @param ids A vector of strings containing the IDs of the buffers to retrieve
      @param timeoutMs The timeout in milliseconds for the operation
      @param allowPartial If true, partial retrieval will not return an error
      @return A pair containing an `ErrorInfo` object and a vector of shared pointers to the retrieved buffers
     */
    virtual std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffers(const std::vector<std::string> &ids,
                                                                                  int timeoutMs, bool allowPartial);

    /*!
      @brief Retrieves multiple buffers by their IDs
      @param ids A vector of strings containing the IDs of the buffers to retrieve
      @param timeoutMs The timeout in milliseconds for the operation
      @param allowPartial If true, partial retrieval will not throw an error
      @return A pair containing an `ErrorInfo` object and a vector of shared pointers to the retrieved buffers
     */
    virtual std::pair<ErrorInfo, std::string> CreateDataObject(size_t metaSize, size_t dataSize,
                                                               std::shared_ptr<DataObject> &dataObj,
                                                               const std::vector<std::string> &nestedObjIds,
                                                               const CreateParam &createParam);

    /*!
      @brief Creates a data object with the specified ID, metadata, and data sizes
      @param objId The ID of the object to be created
      @param metaSize The size of the metadata to be allocated
      @param dataSize The size of the data to be allocated
      @param dataObj A shared pointer to the `DataObject` to be created
      @param nestedObjIds A vector of strings containing the IDs of nested objects
      @param createParam Additional creation parameters (default is empty)
      @return An `ErrorInfo` object indicating the success or failure of the creation
     */
    virtual ErrorInfo CreateDataObject(const std::string &objId, size_t metaSize, size_t dataSize,
                                       std::shared_ptr<DataObject> &dataObj,
                                       const std::vector<std::string> &nestedObjIds,
                                       const CreateParam &createParam = {});

    /*!
      @brief Retrieves multiple data objects by their IDs
      @param ids A vector of strings containing the IDs of the data objects to retrieve
      @param timeoutMs The timeout in milliseconds for the operation
      @param allowPartial If true, partial retrieval will not throw an error
      @return A pair containing an `ErrorInfo` object and a vector of shared pointers to the retrieved data objects
     */
    virtual std::pair<ErrorInfo, std::vector<std::shared_ptr<DataObject>>> GetDataObjects(
        const std::vector<std::string> &ids, int timeoutMs, bool allowPartial);

    /*!
      @brief Retrieves multiple data objects by their IDs without waiting
      @param ids A vector of strings containing the IDs of the data objects to retrieve
      @param timeoutMS The timeout in milliseconds for the operation
      @return A pair containing a `RetryInfo` object and a vector of shared pointers to the retrieved data objects
     */
    virtual std::pair<RetryInfo, std::vector<std::shared_ptr<DataObject>>> GetDataObjectsWithoutWait(
        const std::vector<std::string> &ids, int timeoutMS);

    /*!
      @brief Checks if an object exists in the local storage
      @param objId The ID of the object to check
      @return true if the object exists locally, false otherwise
     */
    virtual bool IsObjectExistingInLocal(const std::string &objId);

    /*!
      @brief 通过指定一组 Object ID 取消其对应的函数调用
      @param objids 需要取消的函数调用对应的 Object ID
      @param isForce 设置为 True 时，如果函数调用正在执行，将会强制杀死函数调用对应的函数实例进程，
                     注意：当前不支持强制杀死并发度配置不为 1 的函数实例
      @param isRecursive 设置为 True 时会取消嵌套的函数调用
      @return 错误信息
     */
    virtual ErrorInfo Cancel(const std::vector<std::string> &objids, bool isForce, bool isRecursive);

    /*!
      @brief 退出当前上下文，调用该方法会执行函数系统客户端及数据系统客户端的优雅退出，清理相应的函数实例及数据对象
      @throw Exception if the exit operation fails
     */
    virtual void Exit(void);

    /*!
      @brief 退出当前上下文，调用该方法会执行函数系统客户端及数据系统客户端的优雅退出，清理相应的函数实例及数据对象
      @throw Exception if the exit operation fails
     */
    virtual void Exit(const int code, const std::string &message);

    /*!
      @brief 向一个函数实例或一组任务发送一个指定的信号
      @param instanceId 指定的函数实例 ID 或者任务 ID，当前仅信号2支持传入任务 ID
      @param sigNo 需要发送的信号，信号1为退出函数实例，信号2为退出任务 ID 对应的所有函数实例，详见 @ref Signal
      @return error information if the operation fails
     */
    virtual ErrorInfo Kill(const std::string &instanceId, int sigNo = libruntime::Signal::KillInstance);

    /*!
      @brief 向一个函数实例或一组任务发送一个指定的信号并携带特定的数据，当前不支持向一组任务发送特定数据
      @param instanceId 指定的函数实例 ID 或者任务 ID
      @param sigNo 需要发送的信号，信号1为退出函数实例，详见 @ref Signal
      @param data 向指定函数实例发送的数据
      @return error information if the operation fails
     */
    virtual ErrorInfo Kill(const std::string &instanceId, int sigNo, std::shared_ptr<Buffer> data);

    /*!
    @brief 向一个函数实例或一组任务发送一个指定的信号并携带特定的数据，并通过callback返回执行结果
    @param instanceId 指定的函数实例 ID 或者任务 ID
    @param sigNo 需要发送的信号，信号1为退出函数实例，详见 @ref Signal
    @param cb 执行结果的回调函数
   */
    virtual void KillAsync(const std::string &instanceId, int sigNo, std::function<void(const ErrorInfo &err)> cb);

    /*!
      @brief 对指定函数实例创建快照
      @param instanceId 指定的函数实例 ID
      @param snapOpts 快照选项，包括是否保持实例运行状态
      @return pair<ErrorInfo, string> 错误信息和 checkpointID
     */
    virtual std::pair<ErrorInfo, std::string> Snapshot(const std::string &instanceId, const SnapOptions &snapOpts);

    /*!
      @brief 从快照恢复函数实例
      @param checkpointId 快照 ID
      @param snapStartOpts 快照启动选项（预留用于未来扩展调度选项等）
      @return pair<ErrorInfo, string> 错误信息和新的实例 ID
     */
    virtual std::pair<ErrorInfo, std::string> Snapstart(const std::string &checkpointId,
                                                         const SnapStartOptions &snapStartOpts);

    /*!
      @brief 结束当前上下文
      @param isDriver 如果设置为 True，将会退出当前任务对应的所有函数实例
      @throw Exception if the finalize operation fails
     */
    virtual void Finalize(bool isDriver = true);

    /*!
      @brief Asynchronously wait for an object to be ready
      @param objectId the ID of the object to wait for
      @param callback the callback function to be called when the object is ready
      @param userData user-defined data to pass to the callback
     */
    virtual void WaitAsync(const std::string &objectId, WaitAsyncCallback callback, void *userData);

    /*!
      @brief Asynchronously get the value of an object
      @param objectId the ID of the object to get
      @param callback the callback function to be called when the value is retrieved
      @param userData user-defined data to pass to the callback
     */
    virtual void GetAsync(const std::string &objectId, GetAsyncCallback callback, void *userData);

    /*!
      @brief Asynchronously get the value of an object
      @param objectId the ID of the object to get
      @param callback the callback function to be called when the value is retrieved
      @param userData user-defined data to pass to the callback
     */
    virtual void GetEvent(const std::string &objectId, GetEventCallback callback, void *userData);

    /*!
      @brief Start a loop to receive and process requests
      @throw Exception if the loop fails to start or run
     */
    virtual void ReceiveRequestLoop(void);

    /*!
      @brief Get the real instance ID of an object
      @param objectId the ID of the object
      @param timeout the timeout in seconds (default is DEFAULT_TIMEOUT)
      @return the real instance ID of the object
      @throw Exception if the operation fails or times out
     */
    virtual std::string GetRealInstanceId(const std::string &objectId, int timeout = DEFAULT_TIMEOUT);

    /*!
      @brief Save the real instance ID of an object (to be deprecated)
      @param objectId the ID of the object
      @param instanceId the real instance ID to save
      @deprecated This function is to be deprecated in future versions
     */
    virtual void SaveRealInstanceId(const std::string &objectId, const std::string &instanceId);  // to be deprecated

    /*!
      @brief Save the real instance ID of an object with additional options
      @param objectId the ID of the object
      @param instanceId the real instance ID to save
      @param opts additional options for saving the instance ID
     */
    virtual void SaveRealInstanceId(const std::string &objectId, const std::string &instanceId,
                                    const InstanceOptions &opts);

    /*!
      @brief Get the instance Ids of an object
      @param objectId the ID of the object
      @param timeout the timeout in seconds (default is DEFAULT_TIMEOUT)
      @return the group instance Ids of the object
      @throw Exception if the operation fails or times out
     */
    virtual std::string GetGroupInstanceIds(const std::string &objectId, int timeout);

    /*!
      @brief Save the group instance Ids of an object with additional options
      @param objectId the ID of the object
      @param groupInsIds the instance Ids in the group
      @param opts additional options for saving the instance ID
     */
    virtual void SaveGroupInstanceIds(const std::string &objectId, const std::string &groupInsIds,
                                      const InstanceOptions &opts);

    /*!
      @brief Writes a key-value pair to the datasystem
      @param key The key to write
      @param value A shared pointer to the `Buffer` containing the value
      @param setParam Additional parameters for the write operation
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo KVWrite(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam);

    /*!
      @brief Writes multiple key-value pairs to datasystem in a transaction
      @details Operations on all key-value pairs must either succeed or fail completely.
      @param keys A vector of strings containing the keys to write
      @param vals A vector of shared pointers to `Buffer` objects containing the values
      @param mSetParam Additional parameters for the multi-set operation
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo KVMSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<Buffer>> &vals,
                               const MSetParam &mSetParam);

    /*!
      @brief Reads a value from the datasystem by its key
      @param key The key to read
      @param timeoutMS The timeout in milliseconds for the operation
      @return A `SingleReadResult` object containing the read value and error information
     */
    virtual SingleReadResult KVRead(const std::string &key, int timeoutMS);

    /*!
      @brief Reads multiple values from the key-value store by their keys
      @param keys A vector of strings containing the keys to read
      @param timeoutMS The timeout in milliseconds for the operation
      @param allowPartial If true, partial retrieval will not return an error (default is false)
      @return A `MultipleReadResult` object containing the read values and error information
     */
    virtual MultipleReadResult KVRead(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial = false);

    /*!
      @brief Reads multiple values from the datasystem with additional parameters
      @param keys A vector of strings containing the keys to read
      @param params Additional parameters for the read operation
      @param timeoutMs The timeout in milliseconds for the operation
      @return A `MultipleReadResult` object containing the read values and error information
     */
    virtual MultipleReadResult KVGetWithParam(const std::vector<std::string> &keys, const GetParams &params,
                                              int timeoutMs);

    /*!
      @brief Deletes a key-value pair from the datasystem
      @param key The key to delete
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo KVDel(const std::string &key);

    /*!
      @brief Deletes multiple key-value pairs from the datasystem
      @param keys A vector of strings containing the keys to delete
      @return A `MultipleDelResult` object containing the results of the failed objs and error information
     */
    virtual MultipleDelResult KVDel(const std::vector<std::string> &keys);

    /*!
      @brief Deletes multiple key-value pairs from the datasystem
      @param keys A vector of strings containing the keys to query
      @return A `MultipleExistResult` object containing the results of the exists and error information
     */
    virtual MultipleExistResult KVExist(const std::vector<std::string> &keys);

    /*!
      @brief Creates a stream producer with the specified configuration
      @param streamName The name of the stream
      @param producerConf Configuration for the stream producer
      @param producer A shared pointer to the `StreamProducer` to be created
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo CreateStreamProducer(const std::string &streamName, ProducerConf producerConf,
                                           std::shared_ptr<StreamProducer> &producer);

    /*!
      @brief Creates a stream consumer with the specified configuration
      @param streamName The name of the stream
      @param config Configuration for the stream consumer
      @param consumer A shared pointer to the `StreamConsumer` to be created
      @param autoAck If true, automatically acknowledges consumed messages (default is false)
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo CreateStreamConsumer(const std::string &streamName, const SubscriptionConfig &config,
                                           std::shared_ptr<StreamConsumer> &consumer, bool autoAck = false);

    /*!
      @brief Deletes a stream by its name
      @param streamName The name of the stream to delete
      @return An `ErrorInfo` object indicating the success or failure of the operation
     */
    virtual ErrorInfo DeleteStream(const std::string &streamName);

    /*!
      @brief Query the number of global producers for a specific stream
      @param streamName the name of the stream
      @param gProducerNum the output parameter to store the number of global producers
      @return ErrorInfo indicating the success or failure of the operation
     */
    virtual ErrorInfo QueryGlobalProducersNum(const std::string &streamName, uint64_t &gProducerNum);

    /*!
      @brief Query the number of global consumers for a specific stream
      @param streamName the name of the stream
      @param gConsumerNum the output parameter to store the number of global consumers
      @return ErrorInfo indicating the success or failure of the operation
     */
    virtual ErrorInfo QueryGlobalConsumersNum(const std::string &streamName, uint64_t &gConsumerNum);
    virtual ErrorInfo SaveState(const std::shared_ptr<Buffer> data, const int &timeout);
    virtual ErrorInfo LoadState(std::shared_ptr<Buffer> &data, const int &timeout);

    /*!
      @brief Get the ID of the currently invoking request
      @return the ID of the currently invoking request
     */
    virtual std::string GetInvokingRequestId(void);

    virtual uint32_t GetThreadPoolSize() const
    {
        return config->threadPoolSize;
    }

    virtual uint32_t GetLocalThreadPoolSize() const
    {
        return config->localThreadPoolSize;
    }

    /*!
      @brief Create a new group with the specified name and options
      @param groupName the name of the group to create
      @param opts the options for the group
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo GroupCreate(const std::string &groupName, GroupOpts &opts);

    /*!
      @brief Wait for the group to complete its operations
      @param groupName the name of the group to wait for
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo GroupWait(const std::string &groupName);

    /*!
      @brief Terminate the group with the specified name
      @param groupName the name of the group to terminate
    */
    virtual void GroupTerminate(const std::string &groupName);

    virtual ErrorInfo GroupSuspend(const std::string &groupName);

    virtual ErrorInfo GroupResume(const std::string &groupName);

    /*!
      @brief Get instances associated with the specified object ID within a timeout
      @param objId the ID of the object
      @param timeoutSec the timeout in seconds
      @return a pair containing a vector of instance names and an ErrorInfo indicating success or failure
    */
    virtual std::pair<std::vector<std::string>, ErrorInfo> GetInstances(const std::string &objId, int timeoutSec);

    /*!
      @brief Get instances associated with the specified object ID and group name
      @param objId the ID of the object
      @param groupName the name of the group
      @return a pair containing a vector of instance names and an ErrorInfo indicating success or failure
    */
    virtual std::pair<std::vector<std::string>, ErrorInfo> GetInstances(const std::string &objId,
                                                                        const std::string &groupName);

    /*!
      @brief Generate a unique group name
      @return a unique group name as a string
    */
    virtual std::string GenerateGroupName();

    /*!
      @brief Create a state store with the specified connection options
      @param opts the connection options
      @param stateStore a shared pointer to the created state store
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo CreateStateStore(const DsConnectOptions &opts, std::shared_ptr<StateStore> &stateStore);

    /*!
      @brief Set the trace ID for the current context
      @param traceId the trace ID to set
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo SetTraceId(const std::string &traceId = "");

    /*!
      @brief Set the tenant ID for the current context
      @param tenantId the tenant ID to set
      @param isReturnErrWhenTenantIDEmpty whether to return an error if the tenant ID is empty
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo SetTenantId(const std::string &tenantId, bool isReturnErrWhenTenantIDEmpty = false);

    /*!
      @brief Set the tenant ID with priority
    */
    virtual void SetTenantIdWithPriority();

    /*!
      @brief Get the current tenant ID
      @return the current tenant ID as a string
    */
    virtual std::string GetTenantId();

    /*!
      @brief Get the server version
      @return the server version as a string
    */
    virtual std::string GetServerVersion();

    virtual ErrorInfo GenerateKeyByStateStore(std::shared_ptr<StateStore> stateStore, std::string &returnKey);
    virtual ErrorInfo SetByStateStore(std::shared_ptr<StateStore> stateStore, const std::string &key,
                                      std::shared_ptr<ReadOnlyNativeBuffer> value, SetParam setParam);
    virtual ErrorInfo SetValueByStateStore(std::shared_ptr<StateStore> stateStore,
                                           std::shared_ptr<ReadOnlyNativeBuffer> value, SetParam setParam,
                                           std::string &returnKey);
    virtual SingleReadResult GetByStateStore(std::shared_ptr<StateStore> stateStore, const std::string &key,
                                             int timeoutMs);
    virtual MultipleReadResult GetArrayByStateStore(std::shared_ptr<StateStore> stateStore,
                                                    const std::vector<std::string> &keys, int timeoutMs,
                                                    bool allowPartial = false);
    virtual ErrorInfo QuerySizeByStateStore(std::shared_ptr<StateStore> stateStore,
                                            const std::vector<std::string> &keys, std::vector<uint64_t> &outSizes);
    virtual ErrorInfo DelByStateStore(std::shared_ptr<StateStore> stateStore, const std::string &key);
    virtual MultipleDelResult DelArrayByStateStore(std::shared_ptr<StateStore> stateStore,
                                                   const std::vector<std::string> &keys);

    /*!
      @brief Execute the shutdown callback with a grace period
      @param gracePeriodSec the grace period in seconds
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo ExecShutdownCallback(uint64_t gracePeriodSec);

    /*!
      @brief Set the value of a UInt64 counter
      @param data the data containing the counter information
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo SetUInt64Counter(const YR::Libruntime::UInt64CounterData &data);

    /*!
      @brief Reset the value of a UInt64 counter
      @param data the data containing the counter information
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo ResetUInt64Counter(const YR::Libruntime::UInt64CounterData &data);

    /*!
      @brief Increase the value of a UInt64 counter
      @param data the data containing the counter information
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo IncreaseUInt64Counter(const YR::Libruntime::UInt64CounterData &data);

    /*!
      @brief Get the value of a UInt64 counter
      @param data the data containing the counter information
      @return a pair containing the value of the counter and an ErrorInfo indicating success or failure
    */
    virtual std::pair<ErrorInfo, uint64_t> GetValueUInt64Counter(const YR::Libruntime::UInt64CounterData &data);

    /*!
      @brief Set the value of a Double counter
      @param data the data containing the counter information
      @return ErrorInfo indicating success or failure
    */
    virtual ErrorInfo SetDoubleCounter(const YR::Libruntime::DoubleCounterData &data);

    /*!
      @brief Reset the value of a double counter
      @param data the data containing the double counter information
      @return ErrorInfo indicating the success or failure of the operation
    */
    virtual ErrorInfo ResetDoubleCounter(const YR::Libruntime::DoubleCounterData &data);

    /*!
      @brief Increase the value of a double counter
      @param data the data containing the double counter information
      @return ErrorInfo indicating the success or failure of the operation
    */
    virtual ErrorInfo IncreaseDoubleCounter(const YR::Libruntime::DoubleCounterData &data);

    /*!
      @brief Get the value of a double counter
      @param data the data containing the double counter information
      @return a pair containing ErrorInfo and the current value of the double counter
    */
    virtual std::pair<ErrorInfo, double> GetValueDoubleCounter(const YR::Libruntime::DoubleCounterData &data);

    /*!
      @brief Report the value of a gauge
      @param gauge the data containing the gauge information
      @return ErrorInfo indicating the success or failure of the operation
    */
    virtual ErrorInfo ReportGauge(const YR::Libruntime::GaugeData &gauge);

    /*!
      @brief Set an alarm with the specified name and description
      @param name the name of the alarm
      @param description the description of the alarm
      @param alarmInfo the information about the alarm
      @return ErrorInfo indicating the success or failure of the operation
    */
    virtual ErrorInfo SetAlarm(const std::string &name, const std::string &description,
                               const YR::Libruntime::AlarmInfo &alarmInfo);

    /*!
      @brief Acquire an instance based on the state ID and function metadata
      @param stateId the state ID of the instance
      @param functionMeta the metadata of the function
      @param opts the invoke options
      @return a pair containing the instance allocation and ErrorInfo
    */
    virtual std::pair<InstanceAllocation, ErrorInfo> AcquireInstance(const std::string &stateId,
                                                                     const FunctionMeta &functionMeta,
                                                                     InvokeOptions &opts);

    /*!
      @brief Release an instance based on the lease ID and state ID
      @param leaseId the lease ID of the instance
      @param stateId the state ID of the instance
      @param abnormal whether the release is due to an abnormal condition
      @param opts the invoke options
      @return ErrorInfo indicating the success or failure of the operation
    */
    virtual ErrorInfo ReleaseInstance(const std::string &leaseId, const std::string &stateId, bool abnormal,
                                      InvokeOptions &opts);

    /*!
      @brief Process a function log
      @param functionLog the function log to process
      @return ErrorInfo indicating the success or failure of the operation
    */
    virtual ErrorInfo ProcessLog(FunctionLog &functionLog);

    /*!
      @brief Wait for a specified timeout before getting the values of the given IDs
      @param ids the list of IDs to wait for
      @param timeoutMs the timeout in milliseconds
      @param allowPartial whether to allow partial results
      @return a pair containing ErrorInfo and the value of the IDs
    */
    virtual std::pair<ErrorInfo, int64_t> WaitBeforeGet(const std::vector<std::string> &ids, int timeoutMs,
                                                        bool allowPartial);

    /*!
      @brief Wait for a fiber event to be signaled
      @param event the fiber event to wait for
    */
    virtual void WaitEvent(FiberEventNotify &event);

    /*!
      @brief Notify a fiber event to wake up waiting fibers
      @param event the fiber event to notify
    */
    virtual void NotifyEvent(FiberEventNotify &event);

    /*!
      @brief Peek into an ObjectRef stream to get the next value
      @param generatorId the ID of the generator
      @param blocking whether to block until a value is available
      @return a pair containing ErrorInfo and the next value in the stream
    */
    virtual std::pair<ErrorInfo, std::string> PeekObjectRefStream(const std::string &generatorId, bool blocking);

    /*!
      @brief Notify the result of a generator
      @param generatorId the ID of the generator
      @param index the index of the result
      @param resultObj the result object
      @param resultErr the error information associated with the result
      @return ErrorInfo indicating the success or failure of the operation
    */
    virtual ErrorInfo NotifyGeneratorResult(const std::string &generatorId, int index,
                                            std::shared_ptr<DataObject> resultObj, const ErrorInfo &resultErr);

    /*!
      @brief Notify that a generator has finished producing results
      @param generatorId the ID of the generator
      @param numResults the number of results produced by the generator
      @return ErrorInfo indicating the success or failure of the operation
      @throw Exception if the generator does not exist or the operation fails
    */
    virtual ErrorInfo NotifyGeneratorFinished(const std::string &generatorId, int numResults);

    /*!
      @brief Get the running information of a function group
      @return the running information of the function group
      @throw Exception if the function group information cannot be retrieved
    */
    virtual FunctionGroupRunningInfo GetFunctionGroupRunningInfo();

    virtual std::pair<ErrorInfo, ResourceGroupUnit> GetResourceGroupTable(const std::string &resourceGroupId);

    /*!
      @brief Get the list of available resources
      @return a pair containing ErrorInfo and a vector of ResourceUnit objects
      @throw Exception if the resources cannot be retrieved
    */
    virtual std::pair<ErrorInfo, std::vector<ResourceUnit>> GetResources(void);

    /**
     * @brief Get node ip address
     * @param[out] a pair containing ErrorInfo and node ip address
     * @return ERR_OK on any key success; the error code otherwise.
     */
    virtual std::pair<ErrorInfo, std::string> GetNodeIpAddress(void);

    /*!
      @brief Get the credential information
      @return the credential information
      @throw Exception if the credential cannot be retrieved
    */
    virtual Credential GetCredential();

    /**
     * @brief For device object, to async get multiple objects
     * @param[in] objIds multiple keys support
     * @param[out] devBlobList vector of blobs, only modify the data pointed to by the pointer.
     * @param[in] timeoutMs max waiting time of getting data
     * @return future of AsyncResult, describe get ErrorInfo and failed list.
     */

    /**
     * @brief For device object Async set multiple objects, and return before publish rpc called.
     * @param[in] objectIds multiple keys support
     * @param[in] devBlobList vector of blobs
     * @return future of AsyncResult, describe set ErrorInfo and failed list.
     */

    /**
     * @brief Invoke worker client to delete all the given objectId.
     * @param[in] objectIds The vector of the objId.
     * @param[out] failedObjectIds The failed delete objIds.
     * @return ERR_OK on any key success; the error code otherwise.
     */
    virtual ErrorInfo DevDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds);

    /**
     * @brief DevLocalDelete interface. After calling this interface, the data replica stored in the data system by the
     * current client connection will be deleted.
     * @param[in] objectIds The objectIds of the data expected to be deleted.
     * @param[out] failedObjectIds Partial failures will be returned through this parameter.
     * @return ERR_OK on when return success; the error code otherwise.
     */
    virtual ErrorInfo DevLocalDelete(const std::vector<std::string> &objectIds,
                                     std::vector<std::string> &failedObjectIds);

    /**
     * @brief Initialize multipath transfer for a given target device (current context)
     * @param[in] devices The device ids participating in the multipath transfer
     * @return ErrorInfo of the call.
     */

    /**
     * @brief Destroys multipath transfer for the current target device (context)
     * @return ErrorInfo of the call.
     */

    /**
     * @brief Subscribe data from device.
     * @param[in] keys A list of keys corresponding to the blob2dList.
     * @param[in] blob2dList A list of structures describing the Device memory.
     * @param[out] futureVec A list of futures to track the  operation.
     * @return ERR_OK on when return all futures success; the error code otherwise.
     */
    virtual ErrorInfo DevSubscribe(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                   std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec);

    /**
     * @brief Publish data to device.
     * @param[in] keys A list of keys corresponding to the blob2dList.
     * @param[in] blob2dList A list of structures describing the Device memory.
     * @param[out] futureVec A list of futures to track the  operation.
     * @return ERR_OK on when return all futures success; the error code otherwise.
     */
    virtual ErrorInfo DevPublish(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                 std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec);

    /**
     * @brief Store Device cache through the data system, caching corresponding keys-blob2dList metadata to the data
     * system
     * @param[in] keys Keys corresponding to blob2dList
     * @param[in] blob2dList List describing the structure of Device memory
     * @param[out] failedKeys Returns failed keys if caching fails
     * @return ERR_OK on when return successfully; the error code otherwise.
     */
    virtual ErrorInfo DevMSet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                              std::vector<std::string> &failedKeys);

    /**
     * @brief Retrieves data from the Device through the data system, storing it in the corresponding DeviceBlobList
     * @param[in] keys Keys corresponding to blob2dList
     * @param[in] blob2dList List describing the structure of Device memory
     * @param[out] failedKeys Returns failed keys if retrieval fails
     * @param[in] timeoutMs Provides a timeout time, defaulting to 0
     * @return ERR_OK on when return successfully; the error code otherwise.
     */
    virtual ErrorInfo DevMGet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                              std::vector<std::string> &failedKeys, int32_t timeoutMs);

    /**
     * @brief Create resource group
     * @param[in] resourceGroupSpec Resource group spec, include name and bundles
     * @param[out] requestId The request id of resource group
     * @return ErrorInfo indicating the success or failure of the operation
     */
    virtual ErrorInfo CreateResourceGroup(const ResourceGroupSpec &resourceGroupSpec, std::string &requestId);

    /**
     * @brief Remove resource group
     * @param[in] resourceGroupName Resource group name
     * @return ErrorInfo indicating the success or failure of the operation
     */
    virtual ErrorInfo RemoveResourceGroup(const std::string &resourceGroupName);

    /**
     * @brief Wait resource group
     * @param[in] resourceGroupName Resource group name
     * @param[in] requestId The request id of resource group
     * @param[in] timeoutMs Provides a timeout time
     * @return ErrorInfo indicating the success or failure of the operation
     */
    virtual ErrorInfo WaitResourceGroup(const std::string &resourceGroupName, const std::string &requestId,
                                        int timeoutSec);
    virtual std::pair<YR::Libruntime::FunctionMeta, ErrorInfo> GetInstance(const std::string &name,
                                                                           const std::string &nameSpace,
                                                                           int timeoutSec);

    virtual bool IsLocalInstances(const std::vector<std::string> &instanceIds);

    virtual ErrorInfo Accelerate(const std::string &groupName, const AccelerateMsgQueueHandle &handle,
                                 HandleReturnObjectCallback callback);

    virtual bool AddReturnObject(const std::vector<std::string> &objIds);

    virtual bool SetError(const std::string &objId, const ErrorInfo &err);

    virtual void UpdateSchdulerInfo(const std::string &schedulerName, const std::string &schedulerId,
                                    const std::string &option);

    /*!
      @brief Get the instance route of an object
      @param objectId the ID of the object
      @return the instance route of the object
      @throw Exception if the operation fails or times out
     */
    virtual std::string GetInstanceRoute(const std::string &objectId);

    /*!
      @brief Save the instance route of an object
      @param objectId the ID of the object
      @param instanceRoute the instance route to save
     */
    virtual void SaveInstanceRoute(const std::string &objectId, const std::string &instanceRoute);

    virtual bool IsHealth();

    virtual bool IsDsHealth();

    virtual ErrorInfo StreamWrite(const std::string &streamMessage, const std::string &requestId,
                                  const std::string &instanceId);

    virtual std::pair<std::string, ErrorInfo> LoadCurrentSession(const std::string &sessionId);

    virtual ErrorInfo UpdateCurrentSession(const std::string &sessionId, const std::string &sessionData);

    virtual bool IsSessionInterrupted(const std::string &sessionId);

    virtual std::pair<ErrorInfo, std::shared_ptr<Buffer>> SessionWait(const std::string &sessionId, int64_t timeoutMs);

    virtual ErrorInfo SessionNotify(const std::string &sessionId, std::shared_ptr<Buffer> data);

    virtual std::pair<std::string, std::string> GetRequestAndInstanceID();

    std::pair<ErrorInfo, std::string> GetNodeId(void);

    std::string GetNameSpace(void);

    std::pair<ErrorInfo, QueryNamedInsResponse> QueryNamedInstances();

    std::string GetActiveMasterAddr();

private:
    std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffersWithoutWait(
        const std::vector<std::string> &ids, int timeoutMS);
    ErrorInfo CheckSpec(std::shared_ptr<InvokeSpec> spec);
    ErrorInfo CheckInstanceRange(std::shared_ptr<InvokeSpec> spec);
    ErrorInfo CheckRGroupOpts(std::shared_ptr<InvokeSpec> spec);
    ErrorInfo PreProcessArgs(const std::shared_ptr<InvokeSpec> &spec);
    bool PutRefArgToDs(std::shared_ptr<InvokeSpec> spec);
    void ProcessErr(const std::shared_ptr<InvokeSpec> &spec, const ErrorInfo &errInfo);
    ErrorInfo CheckRGroupName(const std::string &vGroupName);
    ErrorInfo CheckRGroupSpec(const ResourceGroupSpec &resourceGroupSpec);
    void AddGeneratorReceiver(std::shared_ptr<InvokeSpec> spec);
    ErrorInfo GenerateReturnObjectIds(const std::string &requestId,
                                      std::vector<YR::Libruntime::DataObject> &returnObjs);
    ErrorInfo CreateBuffer(const std::string &objectId, size_t dataSize, std::shared_ptr<Buffer> &dataBuf);
    void FinalizeHandler();
    std::string ConstructTraceId(const YR::Libruntime::InvokeOptions &opts);
    void SetResourceGroupAffinity(std::shared_ptr<InvokeSpec> spec, const std::string &key, const std::string &value);

    std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> MakeGetResult(MultipleResult &res,
                                                                             const std::vector<std::string> &ids,
                                                                             int timeoutMs, bool allowPartial);

    std::shared_ptr<LibruntimeConfig> config;
    std::shared_ptr<DependencyResolver> dependencyResolver;
    // Pre +1 objid Pool
    std::shared_ptr<ObjectIdPool> objectIdPool;
    std::shared_ptr<WaitingObjectManager> waitingObjectManager;

    std::shared_ptr<InvokeAdaptor> invokeAdaptor;
    std::shared_ptr<MemoryStore> memStore;
    DatasystemClients dsClients;
    std::shared_ptr<InvokeOrderManager> invokeOrderMgr;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::shared_ptr<MetricsAdaptor> metricsAdaptor;
    std::shared_ptr<Security> security_;
    std::shared_ptr<RuntimeContext> runtimeContext;
    std::shared_ptr<DomainSocketClient> socketClient_;
    std::shared_ptr<GeneratorNotifier> generatorNotifier_;
    std::shared_ptr<GeneratorReceiver> generatorReceiver_;
    std::shared_ptr<DriverLogReceiver> driverLogReceiver_;
    std::function<ErrorInfo()> checkSignals_ = nullptr;
    std::shared_ptr<MessageCoder> messageCoder_;
    std::shared_ptr<ResourceGroupManager> rGroupManager_;
    std::shared_ptr<ThreadPool> dispatcherThread_;
    std::shared_ptr<YR::scene::DowngradeController> downgrade_;
};

}  // namespace Libruntime
}  // namespace YR
