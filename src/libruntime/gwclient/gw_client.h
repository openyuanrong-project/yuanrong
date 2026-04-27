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

#include "src/libruntime/fsclient/fs_intf.h"
#include "src/libruntime/gwclient/http/client_manager.h"
#include "src/libruntime/heterostore/hetero_store.h"
#include "src/libruntime/invoke_spec.h"
#include "src/libruntime/objectstore/async_decre_ref.h"
#include "src/libruntime/objectstore/object_store.h"
#include "src/libruntime/statestore/state_store.h"
#include "src/libruntime/streamstore/stream_store.h"
#include "src/utility/timer_worker.h"
namespace YR {
namespace Libruntime {
using YR::utility::TimerWorker;
extern const std::string REMOTE_CLIENT_ID_KEY;
extern const std::string REMOTE_CLIENT_ID_KEY_NEW;
extern const std::string TRACE_ID_KEY;
extern const std::string TENANT_ID_KEY;
extern const std::string TENANT_ID_KEY_NEW;

const std::string POSIX_CREATE = "/serverless/v1/posix/instance/create";
const std::string POSIX_INVOKE = "/serverless/v1/posix/instance/invoke";
const std::string POSIX_KILL = "/serverless/v1/posix/instance/kill";

const std::string POSIX_LEASE = "/client/v1/lease";
const std::string POSIX_LEASE_KEEPALIVE = "/client/v1/lease/keepalive";
const std::string POSIX_OBJ_PUT = "/datasystem/v1/obj/put";
const std::string POSIX_OBJ_GET = "/datasystem/v1/obj/get";
const std::string POSIX_OBJ_INCREASE = "/datasystem/v1/obj/increaseref";
const std::string POSIX_OBJ_DECREASE = "/datasystem/v1/obj/decreaseref";
const std::string POSIX_KV_SET = "/datasystem/v1/kv/set";
const std::string POSIX_KV_MSET_TX = "/datasystem/v1/kv/msettx";
const std::string POSIX_KV_GET = "/datasystem/v1/kv/get";
const std::string POSIX_KV_DEL = "/datasystem/v1/kv/del";
const int KEEPALIVE_INTERVAL = 60 * 1000;  // 1 min
const int KEEPALIVE_TIMES = -1;            // unlimited retry

using InvocationCallback =
    std::function<void(const std::string &requestId, YR::Libruntime::ErrorCode code, const std::string &result)>;

thread_local static std::string threadLocalTenantId;

boost::beast::http::verb StringToVerb(const std::string &method);

class GwClient : public FSIntf,
                 public ObjectStore,
                 public StateStore,
                 public StreamStore,
                 public HeteroStore,
                 public std::enable_shared_from_this<GwClient> {
    friend class ClientBuffer;

public:
    GwClient(const std::string &funcId, FSIntfHandlers handlers, std::shared_ptr<Security> security = nullptr)
        : FSIntf(handlers), funcId_(funcId), security_(security)
    {
        this->funcName_ = funcId.substr(funcId.find('/') + 1, funcId.find_last_of('/'));
        this->funcVersion_ = funcId.substr(funcId.find_last_of('/') + 1, funcId.size());
        this->timerWorker_ = std::make_shared<TimerWorker>();
    }
    ~GwClient() = default;
    ErrorInfo Start(const std::string &jobID, const std::string &instanceID = "", const std::string &runtimeID = "",
                    const std::string &functionName = "", const SubscribeFunc &reSubscribeCb = nullptr) override;
    void Stop(void) override;
    void Clear() override;
    void GroupCreateAsync(const CreateRequests &reqs, CreateRespsCallback respCallback, CreateCallBack callback,
                          int timeoutSec = -1) override;
    void CreateAsync(const CreateRequest &req, CreateRespCallback respCallback, CreateCallBack callback,
                     int timeoutSec = -1) override;
    void InvokeAsync(const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback,
                     int timeoutSec = -1) override;
    void InvocationAsync(const std::string &url, const std::shared_ptr<InvokeSpec> spec,
                         const InvocationCallback &callback);
    virtual void InvocationSync(const std::string &method, const std::string &path,
                                const std::unordered_map<std::string, std::string> &headers,
                                const std::string &body, const InvocationCallback &callback);
    void CallResultAsync(const std::shared_ptr<CallResultMessageSpec> req, CallResultCallBack callback)
    {
        STDERR_AND_THROW_EXCEPTION(ERR_INNER_SYSTEM_ERROR, RUNTIME,
                                   "CallResultAsync method not implemented when inCluster is false");
    }
    void KillAsync(const KillRequest &req, KillCallBack callback, int timeoutSec = -1) override;
    void ExitAsync(const ExitRequest &req, ExitCallBack callback) override;
    void StateSaveAsync(const StateSaveRequest &req, StateSaveCallBack callback) override;
    void StateLoadAsync(const StateLoadRequest &req, StateLoadCallBack callback) override;
    void CreateRGroupAsync(const CreateResourceGroupRequest &req, CreateResourceGroupCallBack callback,
                             int timeoutSec = -1)
    {
        STDERR_AND_THROW_EXCEPTION(ERR_INNER_SYSTEM_ERROR, RUNTIME,
                                   "CreateRGroupAsync method not implemented when inCluster is false");
    }
    ErrorInfo Init(std::shared_ptr<HttpClient> httpClient, std::int32_t connectTimeout,
                   const std::string &authToken = "");
    void Init(std::shared_ptr<HttpClient> httpClient);
    ErrorInfo Init(const std::string &ip, int port) override;
    ErrorInfo Init(const std::string &addr, int port, std::int32_t connectTimeout);
    ErrorInfo Init(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                   const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                   const std::string &dsPublicKey, const datasystem::SensitiveValue &token, const std::string &ak,
                   const datasystem::SensitiveValue &sk) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR,
                         "Init method with the nine params is not supported when inCluster is false");
    }

    ErrorInfo Init(datasystem::ConnectOptions &options) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR,
                         "Init method with the ConnectOptions is not supported when inCluster is false");
    }
    ErrorInfo Init(datasystem::ConnectOptions &options, std::shared_ptr<StateStore> dsStateStore) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR,
                         "Init method with the ConnectOptions is not supported when inCluster is false");
    }
    ErrorInfo Init(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                   const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                   const std::string &dsPublicKey, const datasystem::SensitiveValue &token, const std::string &ak,
                   const datasystem::SensitiveValue &sk, std::int32_t connectTimeout) override;

    ErrorInfo CreateBuffer(const std::string &objectId, size_t dataSize, std::shared_ptr<Buffer> &dataBuf,
                           const CreateParam &createParam) override;
    std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffers(const std::vector<std::string> &ids,
                                                                          int timeoutMS) override;
    std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffersWithoutRetry(
        const std::vector<std::string> &ids, int timeoutMS) override;
    ErrorInfo Put(std::shared_ptr<Buffer> data, const std::string &objID,
                  const std::unordered_set<std::string> &nestedID, const CreateParam &createParam) override;
    SingleResult Get(const std::string &objID, int timeoutMS) override;
    MultipleResult Get(const std::vector<std::string> &ids, int timeoutMS) override;
    ErrorInfo IncreGlobalReference(const std::vector<std::string> &objectIds) override;
    ErrorInfo DecreGlobalReference(const std::vector<std::string> &objectIds) override;
    ErrorInfo UpdateToken(datasystem::SensitiveValue token) override;
    ErrorInfo UpdateAkSk(std::string ak, datasystem::SensitiveValue sk) override;
    std::vector<int> QueryGlobalReference(const std::vector<std::string> &objectIds) override
    {
        STDERR_AND_THROW_EXCEPTION(ERR_INNER_SYSTEM_ERROR, RUNTIME,
                                   "QueryGlobalReference method is not supported when inCluster is false");
    }
    ErrorInfo ReleaseGRefs(const std::string &remoteId) override;
    ErrorInfo GenerateKey(std::string &key, const std::string &prefix, bool isPut) override;
    ErrorInfo GetPrefix(const std::string &key, std::string &prefix) override;
    ErrorInfo Write(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam) override;
    ErrorInfo MSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<Buffer>> &vals,
                     const MSetParam &mSetParam) override;
    SingleReadResult Read(const std::string &key, int timeoutMS) override;
    MultipleReadResult Read(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial) override;
    ErrorInfo QuerySize(const std::vector<std::string> &keys, std::vector<uint64_t> &outSizes) override
    {
        STDERR_AND_THROW_EXCEPTION(ERR_INNER_SYSTEM_ERROR, RUNTIME,
                                   "QuerySize method is not supported when inCluster is false");
    }
    ErrorInfo Del(const std::string &key) override;
    MultipleDelResult Del(const std::vector<std::string> &keys) override;
    MultipleExistResult Exist(const std::vector<std::string> &keys) override
    {
        auto result = std::make_shared<std::vector<bool>>();
        result->resize(keys.size());
        ErrorInfo err(ERR_INNER_SYSTEM_ERROR, "Exist method is not supported when inCluster is false");
        return std::make_pair(*result, err);
    }
    ErrorInfo PosixGDecreaseRef(const std::vector<std::string> &objectIds,
                                std::shared_ptr<std::vector<std::string>> failedObjectIds);
    void Shutdown() override {}
    void SetTenantId(const std::string &tenantId) override;
    ErrorInfo Init(const DsConnectOptions &options) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR,
                         "Init method with DsConnectOptions not implemented when inCluster is false");
    }
    ErrorInfo GenerateKey(std::string &returnKey) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "GenerateKey method is not supported when inCluster is false");
    }
    ErrorInfo StartHealthCheck() override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "HealthCheck method is not supported when inCluster is false");
    }
    ErrorInfo Write(std::shared_ptr<Buffer> value, SetParam setParam, std::string &returnKey)
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "Write method is not supported when inCluster is false");
    }
    ErrorInfo CreateStreamProducer(const std::string &streamName, std::shared_ptr<StreamProducer> &producer,
                                   ProducerConf producerConf = {}) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR,
                         "CreateStreamProducer method is not supported when inCluster is false");
    }
    ErrorInfo CreateStreamConsumer(const std::string &streamName, const SubscriptionConfig &config,
                                   std::shared_ptr<StreamConsumer> &consumer, bool autoAck = false) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR,
                         "CreateStreamConsumer method is not supported when inCluster is false");
    }
    ErrorInfo DeleteStream(const std::string &streamName) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "DeleteStream method is not supported when inCluster is false");
    }
    ErrorInfo QueryGlobalProducersNum(const std::string &streamName, uint64_t &gProducerNum) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR,
                         "QueryGlobalProducersNum method is not supported when inCluster is false");
    }
    ErrorInfo QueryGlobalConsumersNum(const std::string &streamName, uint64_t &gConsumerNum) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR,
                         "QueryGlobalConsumersNum method is not supported when inCluster is false");
    }

    ErrorInfo DevDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "Delete method is not supported when inCluster is false");
    }

    ErrorInfo DevLocalDelete(const std::vector<std::string> &objectIds,
                             std::vector<std::string> &failedObjectIds) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "DevLocalDelete method is not supported when inCluster is false");
    }

    ErrorInfo DevSubscribe(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                           std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "DevSubscribe method is not supported when inCluster is false");
    }

    ErrorInfo DevPublish(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                         std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "DevPublish method is not supported when inCluster is false");
    }

    ErrorInfo DevMSet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                      std::vector<std::string> &failedKeys) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "DevMSet method is not supported when inCluster is false");
    }

    ErrorInfo DevMGet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                      std::vector<std::string> &failedKeys, int32_t timeoutMs) override
    {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "DevMGet method is not supported when inCluster is false");
    }
    MultipleReadResult GetWithParam(const std::vector<std::string> &keys, const GetParams &params,
                                    int timeoutMs) override
    {
        auto result = std::make_shared<std::vector<std::shared_ptr<Buffer>>>();
        result->resize(keys.size());
        ErrorInfo err(ERR_INNER_SYSTEM_ERROR, "GetWithParam method is not supported when inCluster is false");
        return std::make_pair(*result, err);
    }

    bool IsHealth() override
    {
        return true;
    };

    ErrorInfo HealthCheck() override
    {
        return ErrorInfo();
    }

private:
    ErrorInfo Lease();
    ErrorInfo Release();
    ErrorInfo KeepLease();

    std::unordered_map<std::string, std::string> BuildHeaders(const std::string &remoteClientId,
                                                              const std::string &traceId, const std::string &tenantId);
    std::pair<std::unordered_map<std::string, std::string>, std::string> BuildRequestWithAkSk(
        const std::shared_ptr<InvokeSpec> spec, const std::string &url);
    LeaseRequest BuildLeaseRequest();
    PutRequest BuildObjPutRequest(std::shared_ptr<Buffer> data, const std::string &objID,
                                  const std::unordered_set<std::string> &nestedID, const CreateParam &createParam);
    GetRequest BuildObjGetRequest(const std::vector<std::string> &keys, int32_t timeoutMs);
    KvGetRequest BuildKvGetRequest(const std::vector<std::string> &keys, int32_t timeoutMs);
    KvSetRequest BuildKvSetRequest(const std::string &key, const std::shared_ptr<Buffer> value,
                                   const SetParam &setParam);
    KvMSetTxRequest BuildKvMSetTxRequest(const std::vector<std::string> &keys,
                                         const std::vector<std::shared_ptr<Buffer>> &vals, const MSetParam &mSetParam);
    KvDelRequest BuildKvDelRequest(const std::vector<std::string> &keys);
    IncreaseRefRequest BuildIncreaseRefRequest(const std::vector<std::string> &objectIds);
    DecreaseRefRequest BuildDecreaseRefRequest(const std::vector<std::string> &objectIds);

    ErrorInfo ParseLeaseResponse(const std::string &result);
    ErrorInfo ParseObjGetResponse(const std::string &result,
                                  std::shared_ptr<std::vector<std::shared_ptr<Buffer>>> results);
    ErrorInfo ParseObjPutResponse(const std::string &result);
    ErrorInfo ParseKvGetResponse(const std::string &result,
                                 std::shared_ptr<std::vector<std::shared_ptr<Buffer>>> results);
    ErrorInfo ParseKvSetResponse(const std::string &result);
    ErrorInfo ParseKvMSetTxResponse(const std::string &result);
    ErrorInfo ParseKvDelResponse(const std::string &result, std::shared_ptr<std::vector<std::string>> failedKeys);

    ErrorInfo ParseIncreaseRefResponse(const std::string &result,
                                       std::shared_ptr<std::vector<std::string>> failedObjectIds);

    ErrorInfo ParseDecreaseRefResponse(const std::string &result,
                                       std::shared_ptr<std::vector<std::string>> failedObjectIds);

    ErrorInfo PosixObjPut(const PutRequest &req);
    ErrorInfo PosixObjGet(const std::vector<std::string> &keys,
                          std::shared_ptr<std::vector<std::shared_ptr<Buffer>>> results, int32_t timeoutMs = 0);
    ErrorInfo PosixKvSet(const std::string &key, const std::shared_ptr<Buffer> value, const SetParam &setParam);
    ErrorInfo PosixKvMSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<Buffer>> &vals,
                            const MSetParam &mSetParam);
    ErrorInfo PosixKvGet(const std::vector<std::string> &keys,
                         std::shared_ptr<std::vector<std::shared_ptr<Buffer>>> results, int32_t timeoutMs = 0);
    ErrorInfo PosixKvDel(const std::vector<std::string> &keys, std::shared_ptr<std::vector<std::string>> failedKeys);

    ErrorInfo PosixGInCreaseRef(const std::vector<std::string> &objectIds,
                                std::shared_ptr<std::vector<std::string>> failedObjectIds);

    ErrorInfo GenRspError(const boost::beast::error_code &errorCode, const uint statusCode, const std::string &result,
                          const std::shared_ptr<std::string> requestId);

    ErrorInfo HandleLease(const std::string &url, const http::verb &verb);

private:
    std::shared_ptr<HttpClient> httpClient_;
    bool init_ = false;
    bool start_ = false;
    std::string funcName_;
    std::string funcVersion_;
    std::string funcId_;
    std::string jobId_;
    AsyncDecreRef asyncDecreRef_;
    mutable std::mutex refCountMu_;
    RefCountMap refCountMap_;
    std::shared_ptr<TimerWorker> timerWorker_;
    std::shared_ptr<YR::utility::Timer> timer_;
    int lostLeaseTimes_ = 0;
    std::int32_t connectTimeout_ = DS_CONNECT_TIMEOUT;
    std::shared_ptr<Security> security_;
    std::string authToken_;
    std::string tenantId_;
};

class ClientBuffer : public NativeBuffer {
public:
    ClientBuffer(uint64_t size, const std::string &objectId, const CreateParam &createParam, std::weak_ptr<GwClient> c)
        : NativeBuffer(size), gwClientWeak_(c), objectId_(objectId), createParam_(createParam)
    {
    }
    virtual ~ClientBuffer() = default;

    virtual ErrorInfo Seal(const std::unordered_set<std::string> &nestedIds) override;

private:
    std::weak_ptr<GwClient> gwClientWeak_;
    std::string objectId_;
    CreateParam createParam_;
};

}  // namespace Libruntime
}  // namespace YR