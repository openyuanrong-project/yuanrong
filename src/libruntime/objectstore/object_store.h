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

#include <memory>
#include <string>

#include <msgpack.hpp>
#include "datasystem/object_client.h"
#include "src/dto/buffer.h"
#include "src/dto/tensor.h"
#include "src/dto/types.h"
#include "src/libruntime/err_type.h"

namespace YR {
namespace Libruntime {
typedef std::pair<ErrorInfo, std::shared_ptr<Buffer>> SingleResult;
typedef std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> MultipleResult;

enum class RetryType : int {
    UNLIMITED_RETRY,
    LIMITED_RETRY,
    NO_RETRY,
};

enum class ConsistencyType : int {
    PRAM = 0,
    CAUSAL = 1,
};

struct RetryInfo {
    ErrorInfo errorInfo;
    RetryType retryType;
};

struct CreateParam {
    WriteMode writeMode = WriteMode::NONE_L2_CACHE;
    ConsistencyType consistencyType = ConsistencyType::PRAM;
    CacheType cacheType = CacheType::MEMORY;
};

class ObjectStore {
public:
    virtual ~ObjectStore() = default;
    virtual ErrorInfo Init(const std::string &addr, int port, std::int32_t connectTimeout) = 0;
    virtual ErrorInfo Init(const std::string &addr, int port, bool enableDsAuth, bool encryptEnable,
                           const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                           const std::string &dsPublicKey, const datasystem::SensitiveValue &token,
                           const std::string &ak, const datasystem::SensitiveValue &sk,
                           std::int32_t connectTimeout) = 0;
    virtual ErrorInfo Init(datasystem::ConnectOptions &inputConnOpt) = 0;
    virtual ErrorInfo CreateBuffer(const std::string &objectId, size_t dataSize, std::shared_ptr<Buffer> &dataBuf,
                                   const CreateParam &createParam) = 0;
    virtual std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffers(const std::vector<std::string> &ids,
                                                                                  int timeoutMS) = 0;
    virtual std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>> GetBuffersWithoutRetry(
        const std::vector<std::string> &ids, int timeoutMS) = 0;

    virtual ErrorInfo Put(std::shared_ptr<Buffer> data, const std::string &objID,
                          const std::unordered_set<std::string> &nestedID, const CreateParam &createParam) = 0;
    virtual SingleResult Get(const std::string &objID, int timeoutMS) = 0;
    virtual MultipleResult Get(const std::vector<std::string> &ids, int timeoutMS) = 0;
    virtual ErrorInfo IncreGlobalReference(const std::vector<std::string> &objectIds) = 0;
    virtual std::pair<ErrorInfo, std::vector<std::string>> IncreGlobalReference(
        const std::vector<std::string> &objectIds, const std::string &remoteId)
    {
        return std::make_pair(ErrorInfo(), std::vector<std::string>());
    }
    virtual ErrorInfo DecreGlobalReference(const std::vector<std::string> &objectIds) = 0;
    virtual std::pair<ErrorInfo, std::vector<std::string>> DecreGlobalReference(
        const std::vector<std::string> &objectIds, const std::string &remoteId)
    {
        return std::make_pair(ErrorInfo(), std::vector<std::string>());
    }
    virtual std::vector<int> QueryGlobalReference(const std::vector<std::string> &objectIds) = 0;
    virtual ErrorInfo ReleaseGRefs(const std::string &remoteId) = 0;
    virtual ErrorInfo GenerateKey(std::string &key, const std::string &prefix, bool isPut) = 0;
    virtual ErrorInfo GetPrefix(const std::string &key, std::string &prefix) = 0;
    virtual void SetTenantId(const std::string &tenantId) = 0;
    virtual void Clear() = 0;
    virtual void Shutdown() = 0;
    virtual ErrorInfo UpdateToken(datasystem::SensitiveValue token) = 0;
    virtual ErrorInfo UpdateAkSk(std::string ak, datasystem::SensitiveValue sk) = 0;
};

class MsgpackBuffer : public NativeBuffer {
public:
    MsgpackBuffer(const std::shared_ptr<msgpack::sbuffer> &mpBuf)
        : NativeBuffer(mpBuf ? mpBuf->data() : nullptr, mpBuf ? mpBuf->size() : 0), msgpackBuf(mpBuf)
    {
    }
    virtual ~MsgpackBuffer() = default;

    virtual ErrorInfo MemoryCopy(const void *data, uint64_t length) override
    {
        msgpackBuf->write(static_cast<const char *>(data), length);
        return ErrorInfo();
    }

private:
    std::shared_ptr<msgpack::sbuffer> msgpackBuf;
};
}  // namespace Libruntime
}  // namespace YR