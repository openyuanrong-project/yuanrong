/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include "datasystem_object_store.h"

#include <unistd.h>
#include <iterator>
#include <memory>
#include <sstream>
#include <string>
#include <unordered_map>

#include "datasystem_object_client_wrapper.h"
#include "object_store_impl.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/utils/constants.h"
#include "src/libruntime/utils/utils.h"
#include "src/utility/id_generator.h"

namespace YR {
using namespace datasystem;
namespace Libruntime {
inline void AssignDSBufferToResult(ds::Optional<ds::Buffer> &&buffer, size_t index,
                                   std::vector<std::shared_ptr<Buffer>> &results)
{
    results[index] = std::make_shared<DataSystemBuffer>(std::move(buffer));
}

template <typename T>
size_t ExtractSuccessObjects(std::vector<std::string> &remainIds, std::vector<ds::Optional<ds::Buffer>> &remainList,
                             std::vector<std::shared_ptr<T>> &bufferList,
                             std::unordered_map<std::string, std::list<size_t>> &idToIndices)
{
    size_t newSuccessCount = 0;
    std::vector<std::string> newRemainIds;
    newRemainIds.reserve(remainIds.size());
    for (size_t i = 0; i < remainIds.size(); ++i) {
        if ((i < remainList.size()) && remainList[i]) {  // success obj
            auto &indices = idToIndices[remainIds[i]];
            if (UNLIKELY(indices.empty())) {
                YRLOG_ERROR("Indices should not be empty. key: {}", remainIds[i]);
                continue;
            }
            AssignDSBufferToResult(std::move(remainList[i]), indices.front(), bufferList);
            indices.pop_front();
            newSuccessCount++;
        } else {
            newRemainIds.emplace_back(std::move(remainIds[i]));
        }
    }
    if (newRemainIds.size() > 0) {
        YRLOG_INFO("Datasystem get partial objects; success objects: ({}/{}); retrying [{}, ...]", newSuccessCount,
                   remainIds.size(), newRemainIds[0]);
    }
    remainIds.swap(newRemainIds);  // remainIds = newRemainIds
    return newSuccessCount;
}

template <typename T>
ErrorInfo ObjectStoreCommonGetImpl(const std::vector<std::string> &ids, int timeoutMS,
                                   std::vector<std::shared_ptr<T>> &bufferList,
                                   std::shared_ptr<ds::ObjectClient> client)
{
    ds::Status status;
    ErrorInfo err;
    std::unordered_map<std::string, std::list<size_t>> idToIndex;
    std::vector<std::string> remainIds(ids);
    size_t successCount = 0;
    for (size_t i = 0; i < ids.size(); ++i) {
        idToIndex[ids[i]].push_back(i);  // convert id to index
    }
    int limitedRetryTime = 0;
    auto start = high_resolution_clock::now();
    auto getElapsedTime = [start]() {
        return duration_cast<milliseconds>(high_resolution_clock::now() - start).count();
    };

    while ((getElapsedTime() <= timeoutMS) || (timeoutMS == NO_TIMEOUT) || (timeoutMS == 0)) {
        std::vector<ds::Optional<ds::Buffer>> remainList;
        remainList.reserve(remainIds.size());
        int to = (timeoutMS == NO_TIMEOUT) ? (DEFAULT_TIMEOUT_MS) : (timeoutMS - static_cast<int>(getElapsedTime()));
        to = to < 0 ? 0 : to;
        status = client->Get(remainIds, to, remainList);
        // 1. ok, put remainIds to completedIds
        if (!IsRetryableStatus(status)) {
            break;
        }
        if (IsLimitedRetryEnd(status, limitedRetryTime)) {
            break;
        }
        size_t newSuccessCount = ExtractSuccessObjects(remainIds, remainList, bufferList, idToIndex);
        successCount += newSuccessCount;
        if (successCount == ids.size()) {
            return err;
        }
        if ((timeoutMS != NO_TIMEOUT && getElapsedTime() > timeoutMS) || (timeoutMS == 0)) {
            break;
        }
        YRLOG_INFO("Datasystem retry to get objects: {}. Elapsed: {}s", status.ToString(), getElapsedTime() / S_TO_MS);
        // 2. other failure reason, retry until timeout
        std::this_thread::sleep_for(std::chrono::seconds(GET_RETRY_INTERVAL));
    }
    return GenerateErrorInfo(successCount, status, timeoutMS, remainIds, ids);
}

template <typename T>
RetryInfo ObjectStoreGetImplWithoutRetry(const std::vector<std::string> &ids, int timeoutMS,
                                         std::vector<std::shared_ptr<T>> &bufferList,
                                         std::shared_ptr<ds::ObjectClient> client)
{
    RetryInfo retryInfo;
    std::vector<ds::Optional<ds::Buffer>> remainList;
    remainList.reserve(ids.size());
    auto start = high_resolution_clock::now();
    auto getElapsedTime = [start]() {
        return duration_cast<milliseconds>(high_resolution_clock::now() - start).count();
    };
    ds::Status status = client->Get(ids, timeoutMS, remainList);
    for (size_t i = 0; i < remainList.size(); i++) {
        if (remainList[i]) {
            AssignDSBufferToResult(std::move(remainList[i]), i, bufferList);
        }
    }
    if (!status.IsOk()) {
        retryInfo.errorInfo.SetErrCodeAndMsg(YR::Libruntime::ConvertDatasystemErrorToCore(status.GetCode()),
                                             YR::Libruntime::ModuleCode::DATASYSTEM, status.ToString(),
                                             status.GetCode());
    }
    if (IsUnlimitedRetryableStatus(status)) {
        retryInfo.retryType = RetryType::UNLIMITED_RETRY;
    } else if (IsLimitedRetryableStatus(status)) {
        retryInfo.retryType = RetryType::LIMITED_RETRY;
    } else {
        retryInfo.retryType = RetryType::NO_RETRY;
    }
    YRLOG_INFO("Datasystem retry get objects: {}. Elapsed: {}s", status.ToString(), getElapsedTime() / S_TO_MS);
    return retryInfo;
}

ErrorInfo DSCacheObjectStore::Init(const std::string &addr, int port, std::int32_t connectTimeout)
{
    return DSCacheObjectStore::Init(addr, port, false, false, "", SensitiveValue{}, "", "", "", SensitiveValue{},
                                    connectTimeout);
}

ErrorInfo DSCacheObjectStore::Init(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                                   const std::string &runtimePublicKey, const SensitiveValue &runtimePrivateKey,
                                   const std::string &dsPublicKey, const SensitiveValue &token, const std::string &ak,
                                   const SensitiveValue &sk, const std::int32_t connectTimeout)
{
    ErrorInfo err;
    YRLOG_DEBUG("Datasystem object store init, ip = {}, port = {}, connectTimeout is {}", ip, port, connectTimeout);
    connectOpts.host = ip;
    connectOpts.port = port;
    connectOpts.connectTimeoutMs = connectTimeout * S_TO_MS;
    if (encryptEnable) {
        connectOpts.clientPublicKey = runtimePublicKey;
        connectOpts.clientPrivateKey = runtimePrivateKey;
        connectOpts.serverPublicKey = dsPublicKey;
    }
    if (enableDsAuth) {
        GetAuthConnectOpts(connectOpts, ak, sk, token);
    }
    return err;
}

ErrorInfo DSCacheObjectStore::Init(datasystem::ConnectOptions &inputConnOpt)
{
    this->connectOpts.host = inputConnOpt.host;
    this->connectOpts.port = inputConnOpt.port;
    this->connectOpts.clientPublicKey = inputConnOpt.clientPublicKey;
    this->connectOpts.clientPrivateKey = inputConnOpt.clientPrivateKey;
    this->connectOpts.serverPublicKey = inputConnOpt.serverPublicKey;
    this->connectOpts.accessKey = inputConnOpt.accessKey;
    this->connectOpts.secretKey = inputConnOpt.secretKey;
    this->connectOpts.connectTimeoutMs = inputConnOpt.connectTimeoutMs;
    this->connectOpts.tenantId = inputConnOpt.tenantId;
    return ErrorInfo();
}

ErrorInfo DSCacheObjectStore::DoInitOnce(void)
{
    YRLOG_DEBUG("begin to init ds object client");
    dsClient = std::make_shared<ds::ObjectClient>(this->connectOpts);
    std::shared_ptr<DatasystemClientWrapper> dsClientWrapper =
        std::make_shared<DatasystemObjectClientWrapper>(dsClient);
    asyncDecreRef.Init(dsClientWrapper);
    ds::Status status = dsClient->Init();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_CONNECTION_FAILED,
                      status.ToString());
    isInit = true;
    YRLOG_INFO("success to init ds object client");
    return ErrorInfo();
}

void DSCacheObjectStore::InitOnce(void)
{
    std::call_once(this->initFlag, [this]() { this->initErr = this->DoInitOnce(); });
}

ErrorInfo DSCacheObjectStore::CreateBuffer(const std::string &objectId, size_t dataSize,
                                           std::shared_ptr<Buffer> &dataBuf, const CreateParam &createParam)
{
    OBJ_STORE_INIT_ONCE();
    std::shared_ptr<ds::Buffer> dataBuffer;
    ds::CreateParam param;
    param.consistencyType = static_cast<datasystem::ConsistencyType>(createParam.consistencyType);
    param.cacheType = static_cast<datasystem::CacheType>(createParam.cacheType);
    ds::Status status = dsClient->Create(objectId, dataSize, param, dataBuffer);
    if (!status.IsOk()) {
        if (status.GetCode() != ds::StatusCode::K_OC_ALREADY_SEALED) {
            std::string msg = "failed to create buffer, objId: " + objectId + ", errMsg:" + status.ToString();
            RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
        }
        YRLOG_WARN("Status code: K_OC_ALREADY_SEALED, objId: {}, Repeated put should directly return", objectId);
        return ErrorInfo();
    }
    dataBuf = std::make_shared<DataSystemBuffer>(dataBuffer);
    return ErrorInfo();
}

std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> DSCacheObjectStore::GetBuffers(
    const std::vector<std::string> &ids, int timeoutMS)
{
    std::vector<std::shared_ptr<Buffer>> results;
    OBJ_STORE_INIT_ONCE_RETURN_PAIR(results);
    results.resize(ids.size());
    ErrorInfo err = GetBuffersImpl(ids, timeoutMS, results);
    return std::make_pair(err, results);
}

std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>> DSCacheObjectStore::GetBuffersWithoutRetry(
    const std::vector<std::string> &ids, int timeoutMS)
{
    std::vector<std::shared_ptr<Buffer>> results;
    results.resize(ids.size());
    return std::make_pair(GetBuffersWithoutRetryImpl(ids, timeoutMS, results), results);
}

ErrorInfo DSCacheObjectStore::GetBuffersImpl(const std::vector<std::string> &ids, int timeoutMS,
                                             std::vector<std::shared_ptr<Buffer>> &bufferList)
{
    OBJ_STORE_INIT_ONCE();
    return ObjectStoreCommonGetImpl(ids, timeoutMS, bufferList, dsClient);
}

RetryInfo DSCacheObjectStore::GetBuffersWithoutRetryImpl(const std::vector<std::string> &ids, int timeoutMS,
                                                         std::vector<std::shared_ptr<Buffer>> &bufferList)
{
    OBJ_STORE_INIT_ONCE_THROW();
    return ObjectStoreGetImplWithoutRetry(ids, timeoutMS, bufferList, dsClient);
}

ErrorInfo DSCacheObjectStore::Put(std::shared_ptr<Buffer> data, const std::string &objId,
                                  const std::unordered_set<std::string> &nestedId, const CreateParam &createParam)
{
    OBJ_STORE_INIT_ONCE();
    std::string msg;
    std::shared_ptr<ds::Buffer> dataBuffer;
    ds::CreateParam param;
    param.consistencyType = static_cast<datasystem::ConsistencyType>(createParam.consistencyType);
    param.cacheType = static_cast<datasystem::CacheType>(createParam.cacheType);
    ds::Status status = dsClient->Create(objId, static_cast<uint64_t>(data->GetSize()), param, dataBuffer);
    if (!status.IsOk()) {
        if (status.GetCode() != ds::StatusCode::K_OC_ALREADY_SEALED) {
            msg = "failed to create buffer, objId: " + objId + ", errMsg:" + status.ToString();
            RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
        }
        YRLOG_WARN("Status code: K_OC_ALREADY_SEALED, objId: {}, Repeated put should directly return", objId);
        return ErrorInfo();
    }
    status = dataBuffer->WLatch();
    msg = "failed to WLatch buffer, objId: " + objId + ", errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    status = dataBuffer->MemoryCopy(data->ImmutableData(), static_cast<uint64_t>(data->GetSize()));
    msg = "failed to memorycopy buffer, objId: " + objId + ", errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    status = dataBuffer->Seal(nestedId);
    if (!status.IsOk()) {
        dataBuffer->UnWLatch();
        msg = "failed to seal objId, objId: " + objId + ", errMsg:" + status.ToString();
        RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    }
    status = dataBuffer->UnWLatch();
    msg = "failed to UnWLatch buffer, objId: " + objId + ", errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo DSCacheObjectStore::UpdateToken(datasystem::SensitiveValue token)
{
    if (!isInit) {
        return ErrorInfo();
    }
    return ErrorInfo();
}

ErrorInfo DSCacheObjectStore::UpdateAkSk(std::string ak, datasystem::SensitiveValue sk)
{
    return ErrorInfo();
}

ErrorInfo DSCacheObjectStore::GetImpl(const std::vector<std::string> &ids, int timeoutMS,
                                      std::vector<std::shared_ptr<Buffer>> &sbufferList)
{
    OBJ_STORE_INIT_ONCE();
    return ObjectStoreCommonGetImpl(ids, timeoutMS, sbufferList, dsClient);
}

SingleResult DSCacheObjectStore::Get(const std::string &objId, int timeoutMS)
{
    OBJ_STORE_INIT_ONCE_RETURN_PAIR(nullptr);
    std::vector<std::string> ids = {objId};
    MultipleResult multiRes = Get(ids, timeoutMS);
    if (multiRes.first.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        auto result = std::shared_ptr<NativeBuffer>();
        return std::make_pair(multiRes.first, result);
    }
    return std::make_pair(ErrorInfo(), multiRes.second[0]);
}

MultipleResult DSCacheObjectStore::Get(const std::vector<std::string> &ids, int timeoutMS)
{
    std::vector<std::shared_ptr<Buffer>> result;
    OBJ_STORE_INIT_ONCE_RETURN_PAIR(result);
    result.resize(ids.size());
    ErrorInfo err = GetImpl(ids, timeoutMS, result);
    return std::make_pair(err, result);
}

ErrorInfo DSCacheObjectStore::IncreGlobalReference(const std::vector<std::string> &objectIds)
{
    OBJ_STORE_INIT_ONCE();
    std::vector<std::string> failedObjectIds;
    ds::Status status = dsClient->GIncreaseRef(objectIds, failedObjectIds);
    ErrorInfo err = IncreaseRefReturnCheck(status, failedObjectIds);
    if (err.Code() != YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_ERROR(err.Msg());
        return err;
    }
    refCountMap.IncreRefCount(objectIds);
    if (!failedObjectIds.empty()) {
        YRLOG_WARN("Datasystem failed to increase all objectRefs, fail count: {}", failedObjectIds.size());
        refCountMap.DecreRefCount(failedObjectIds);
    }
    return err;
}

std::pair<ErrorInfo, std::vector<std::string>> DSCacheObjectStore::IncreGlobalReference(
    const std::vector<std::string> &objectIds, const std::string &remoteId)
{
    std::vector<std::string> failedObjectIds;
    OBJ_STORE_INIT_ONCE_RETURN_PAIR(failedObjectIds);
    ds::Status status = dsClient->GIncreaseRef(objectIds, failedObjectIds);
    auto code =
        YR::Libruntime::ConvertDatasystemErrorToCore(status.GetCode(), static_cast<ErrorCode>(status.GetCode()));
    auto msg = status.GetMsg();
    auto err = ErrorInfo(code, ModuleCode::DATASYSTEM, msg);
    return std::make_pair(err, failedObjectIds);
}

ErrorInfo DSCacheObjectStore::DecreGlobalReference(const std::vector<std::string> &objectIds)
{
    OBJ_STORE_INIT_ONCE();
    ErrorInfo err;
    // if the objectId is not in the map, not decrease
    std::vector<std::string> needDecreObjectIds = refCountMap.DecreRefCount(objectIds);
    if (!needDecreObjectIds.empty()) {
        bool success = asyncDecreRef.Push(needDecreObjectIds, tenantId);
        if (!success) {
            err.SetErrCodeAndMsg(YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                                 YR::Libruntime::ModuleCode::DATASYSTEM, "async decrease thread has exited");
        }
    }
    return err;
}

std::pair<ErrorInfo, std::vector<std::string>> DSCacheObjectStore::DecreGlobalReference(
    const std::vector<std::string> &objectIds, const std::string &remoteId)
{
    std::vector<std::string> failedObjectIds;
    OBJ_STORE_INIT_ONCE_RETURN_PAIR(failedObjectIds);
    ds::Status status = dsClient->GDecreaseRef(objectIds, failedObjectIds);
    auto code =
        YR::Libruntime::ConvertDatasystemErrorToCore(status.GetCode(), static_cast<ErrorCode>(status.GetCode()));
    auto msg = status.GetMsg();
    auto err = ErrorInfo(code, ModuleCode::DATASYSTEM, msg);
    return std::make_pair(err, failedObjectIds);
}

std::vector<int> DSCacheObjectStore::QueryGlobalReference(const std::vector<std::string> &objectIds)
{
    std::vector<int> counting;
    OBJ_STORE_INIT_ONCE_RETURN(counting);
    for (auto &id : objectIds) {
        (void)counting.emplace_back(dsClient->QueryGlobalRefNum(id));
    }
    return counting;
}

ErrorInfo DSCacheObjectStore::ReleaseGRefs(const std::string &remoteId)
{
    OBJ_STORE_INIT_ONCE();
    ds::Status status = dsClient->ReleaseGRefs(YR::utility::ParseRealJobId(remoteId));
    auto code =
        YR::Libruntime::ConvertDatasystemErrorToCore(status.GetCode(), static_cast<ErrorCode>(status.GetCode()));
    auto msg = status.GetMsg();
    return ErrorInfo(code, ModuleCode::DATASYSTEM, msg);
}

ErrorInfo DSCacheObjectStore::GenerateKey(std::string &key, const std::string &prefix, bool isPut)
{
    // if DS-client is not initialized, do not init here, because it may cause memory occupation
    // just return prefix as key although it does not utilize distributed-master-feature of DS
    OBJ_STORE_INIT_ONCE();
    if (!isInit || !isPut) {
        key = prefix;
        return ErrorInfo();
    }
    std::string msg;
    ds::Status status = dsClient->GenerateKey(prefix, key);
    msg = "failed to GenerateKey, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

ErrorInfo DSCacheObjectStore::GetPrefix(const std::string &key, std::string &prefix)
{
    OBJ_STORE_INIT_ONCE();
    std::string msg;
    ds::Status status = dsClient->GetPrefix(key, prefix);
    msg = "failed to GetPrefix, errMsg:" + status.ToString();
    RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED, msg);
    return ErrorInfo();
}

void DSCacheObjectStore::SetTenantId(const std::string &tenantId)
{
    (void)datasystem::Context::SetTenantId(tenantId);
    this->tenantId = tenantId;
}

void DSCacheObjectStore::Clear()
{
    if (dsClient == nullptr) {
        return;
    }
    std::vector<std::string> objectIds = refCountMap.ToArray();
    refCountMap.Clear();
    if (asyncDecreRef.Push(objectIds, tenantId)) {
        asyncDecreRef.Stop();
    }
}

void DSCacheObjectStore::Shutdown()
{
    if (dsClient == nullptr) {
        return;
    }
    ds::Status status = dsClient->ShutDown();
    if (!status.IsOk()) {
        YRLOG_WARN("DSCacheObjectStore Shutdown fail. Status code: {}, Msg: {}", fmt::underlying(status.GetCode()),
                   status.ToString());
    }
    isInit = false;
}
}  // namespace Libruntime
}  // namespace YR
