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

#pragma once

#include <mutex>

#include "async_decre_ref.h"
#include "reference_count_map.h"
#include "src/dto/buffer.h"
#include "src/libruntime/objectstore/object_store.h"
#include "src/utility/logger/logger.h"
namespace YR {
namespace Libruntime {
class DSCacheObjectStore : public ObjectStore {
public:
    DSCacheObjectStore() = default;

    ~DSCacheObjectStore() = default;

    ErrorInfo Init(const std::string &addr, int port, std::int32_t connectTimeout = DS_CONNECT_TIMEOUT) override;

    ErrorInfo Init(const std::string &addr, int port, bool enableDsAuth, bool encryptEnable,
                   const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                   const std::string &dsPublicKey, const datasystem::SensitiveValue &token, const std::string &ak,
                   const datasystem::SensitiveValue &sk, std::int32_t connectTimeout) override;

    ErrorInfo Init(datasystem::ConnectOptions &inputConnOpt) override;

    void InitOnce(void);

    ErrorInfo CreateBuffer(const std::string &objectId, size_t dataSize, std::shared_ptr<Buffer> &dataBuf,
                           const CreateParam &createParam) override;

    std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffers(const std::vector<std::string> &ids,
                                                                          int timeoutMS);

    std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffersWithoutRetry(
        const std::vector<std::string> &ids, int timeoutMS) override;

    // Store an object in the datasystem.
    ErrorInfo Put(std::shared_ptr<Buffer> data, const std::string &objId,
                  const std::unordered_set<std::string> &nestedId, const CreateParam &createParam) override;

    // Get a single object from the datasystem.
    SingleResult Get(const std::string &objId, int timeoutMS) override;

    // Get a list of objects from the datasystem.
    MultipleResult Get(const std::vector<std::string> &ids, int timeoutMS) override;
    ErrorInfo UpdateToken(datasystem::SensitiveValue token) override;
    ErrorInfo UpdateAkSk(std::string ak, datasystem::SensitiveValue sk) override;

    ErrorInfo IncreGlobalReference(const std::vector<std::string> &objectIds) override;

    std::pair<ErrorInfo, std::vector<std::string>> IncreGlobalReference(const std::vector<std::string> &objectIds,
                                                                        const std::string &remoteId) override;

    ErrorInfo DecreGlobalReference(const std::vector<std::string> &objectIds) override;

    std::pair<ErrorInfo, std::vector<std::string>> DecreGlobalReference(const std::vector<std::string> &objectIds,
                                                                        const std::string &remoteId) override;

    std::vector<int> QueryGlobalReference(const std::vector<std::string> &objectIds) override;

    ErrorInfo ReleaseGRefs(const std::string &remoteId) override;
    ErrorInfo GenerateKey(std::string &key, const std::string &prefix, bool isPut) override;

    ErrorInfo GetPrefix(const std::string &key, std::string &prefix) override;

    void SetTenantId(const std::string &tenantId) override;

    void Clear();
    void Shutdown() override;

private:
    ErrorInfo DoInitOnce(void);
    ErrorInfo GetImpl(const std::vector<std::string> &ids, int timeoutMS,
                      std::vector<std::shared_ptr<Buffer>> &sbufferList);
    ErrorInfo GetBuffersImpl(const std::vector<std::string> &ids, int timeoutMS,
                             std::vector<std::shared_ptr<Buffer>> &bufferList);
    RetryInfo GetBuffersWithoutRetryImpl(const std::vector<std::string> &ids, int timeoutMS,
                                         std::vector<std::shared_ptr<Buffer>> &bufferList);
    std::shared_ptr<datasystem::ObjectClient> dsClient;
    AsyncDecreRef asyncDecreRef;
    bool isInit = false;
    RefCountMap refCountMap;
    std::once_flag initFlag;
    ErrorInfo initErr;
    std::string tenantId;
    datasystem::SensitiveValue tokenUpdated;
    ds::ConnectOptions connectOpts;
};

#define OBJ_STORE_INIT_ONCE_THROW()                                          \
    do {                                                                     \
        InitOnce();                                                          \
        if (!initErr.OK()) {                                                 \
            throw Exception(initErr.Code(), initErr.MCode(), initErr.Msg()); \
        }                                                                    \
    } while (0)

#define OBJ_STORE_INIT_ONCE() \
    do {                      \
        InitOnce();           \
        if (!initErr.OK()) {  \
            return initErr;   \
        }                     \
    } while (0)

#define OBJ_STORE_INIT_ONCE_RETURN_PAIR(_second_)       \
    do {                                                \
        InitOnce();                                     \
        if (!initErr.OK()) {                            \
            return std::make_pair(initErr, (_second_)); \
        }                                               \
    } while (0)

#define OBJ_STORE_INIT_ONCE_RETURN(_ret_) \
    do {                                  \
        InitOnce();                       \
        if (!initErr.OK()) {              \
            return (_ret_);               \
        }                                 \
    } while (0)

class DataSystemBuffer : public SharedBuffer {
public:
    DataSystemBuffer(std::shared_ptr<datasystem::Buffer> buf)
        : SharedBuffer(buf->MutableData(), buf->GetSize()), buffer(buf)
    {
    }
    DataSystemBuffer(datasystem::Optional<datasystem::Buffer> &&buf) : SharedBuffer(buf->MutableData(), buf->GetSize())
    {
        buffer = std::make_shared<datasystem::Buffer>(std::move(*buf));
    }

    virtual ErrorInfo MemoryCopy(const void *data, uint64_t length) override
    {
        datasystem::Status status = buffer->MemoryCopy(data, length);
        RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                          status.ToString());
        return ErrorInfo();
    }

    virtual ErrorInfo Seal(const std::unordered_set<std::string> &nestedIds) override
    {
        datasystem::Status status = buffer->Seal(nestedIds);
        RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                          status.ToString());
        return ErrorInfo();
    }

    virtual ErrorInfo WriterLatch() override
    {
        datasystem::Status status = buffer->WLatch();
        RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                          status.ToString());
        return ErrorInfo();
    }

    virtual ErrorInfo WriterUnlatch() override
    {
        datasystem::Status status = buffer->UnWLatch();
        RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                          status.ToString());
        return ErrorInfo();
    }

    virtual ErrorInfo ReaderLatch() override
    {
        datasystem::Status status = buffer->RLatch();
        RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                          status.ToString());
        return ErrorInfo();
    }

    virtual ErrorInfo ReaderUnlatch() override
    {
        datasystem::Status status = buffer->UnRLatch();
        RETURN_ERR_NOT_OK(status.IsOk(), status.GetCode(), YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED,
                          status.ToString());
        return ErrorInfo();
    }

    virtual ErrorInfo Publish() override
    {
        datasystem::Status status = buffer->Publish();
        if (!status.IsOk()) {
            auto tmp = ConvertDatasystemErrorToCore(status.GetCode(), ErrorCode::ERR_DATASYSTEM_FAILED);
            return ErrorInfo(tmp, ModuleCode::DATASYSTEM, "failed to publish, errMsg: " + status.ToString());
        }
        return ErrorInfo();
    }

private:
    std::shared_ptr<datasystem::Buffer> buffer;
};
}  // namespace Libruntime
}  // namespace YR
