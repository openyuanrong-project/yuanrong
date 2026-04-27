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

#include "datasystem/utils/connection.h"
#include "datasystem/utils/sensitive_value.h"
#include "msgpack.hpp"
#include "src/dto/buffer.h"
#include "src/dto/types.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/utils/exception.h"

namespace YR {
namespace Libruntime {
typedef std::pair<std::vector<std::string>, ErrorInfo> MultipleDelResult;
typedef std::pair<std::vector<bool>, ErrorInfo> MultipleExistResult;
typedef std::pair<std::shared_ptr<Buffer>, ErrorInfo> SingleReadResult;
typedef std::pair<std::vector<std::shared_ptr<Buffer>>, ErrorInfo> MultipleReadResult;

enum class ExistenceOpt : int {
    NONE = 0,  // Does not check for existence.
    NX = 1,    // Only set the key if it does not already exist.
};

struct SetParam {
    WriteMode writeMode = WriteMode::NONE_L2_CACHE;
    // The default value 0 means the key will keep alive until you call Del api to delete the key explicitly.
    uint32_t ttlSecond = 0;
    ExistenceOpt existence = ExistenceOpt::NONE;
    CacheType cacheType = CacheType::MEMORY;
    std::unordered_map<std::string, std::string> extendParams;
};

struct MSetParam {
    WriteMode writeMode = WriteMode::NONE_L2_CACHE;
    uint32_t ttlSecond = 0;
    ExistenceOpt existence = ExistenceOpt::NX;  // MSetNx only support NX mode
    CacheType cacheType = CacheType::MEMORY;
    std::unordered_map<std::string, std::string> extendParams;
};

struct GetParam {
    uint64_t offset;
    uint64_t size;
};

struct GetParams {
    std::vector<GetParam> getParams;
};

struct DsConnectOptions {
    std::string host;
    int32_t port;
    int32_t connectTimeoutMs = 60 * 1000;  // 60s
    std::string token = "";
    std::string clientPublicKey = "";
    std::string clientPrivateKey = "";
    std::string serverPublicKey = "";
    std::string accessKey = "";
    std::string secretKey = "";
    std::string oAuthClientId = "";
    std::string oAuthClientSecret = "";
    std::string oAuthUrl = "";
    std::string tenantId = "";
    bool enableCrossNodeConnection = false;
};

class StateStore {
public:
    virtual ~StateStore() = default;

    /**
     * @brief Init StateStore object without ds auth.
     * @param[in] ip The ip of datasystem worker.
     * @param[in] port The port of datasystem worker.
     */
    virtual ErrorInfo Init(const std::string &ip, int port, std::int32_t connectTimeout) = 0;

    /**
     * @brief Init StateStore object.
     * @param[in] ip The ip of datasystem worker.
     * @param[in] port The port of datasystem worker.
     * @param[in] enableDsAuth enable auth.
     * @param[in] runtimePublicKey The runtime public key.
     * @param[in] runtimePrivateKey The runtime private key.
     * @param[in] dsPublicKey The datasystem public key.
     */
    virtual ErrorInfo Init(const std::string &ip, int port, bool enableDsAuth, bool encryptEnable,
                           const std::string &runtimePublicKey, const datasystem::SensitiveValue &runtimePrivateKey,
                           const std::string &dsPublicKey, const datasystem::SensitiveValue &token,
                           const std::string &ak, const datasystem::SensitiveValue &sk,
                           std::int32_t connectTimeout) = 0;

    virtual ErrorInfo Init(const DsConnectOptions &options) = 0;

    virtual ErrorInfo Init(datasystem::ConnectOptions &inputConnOpt) = 0;

    /**
     * @brief Set the value of a key by datasystem StateClient.
     * @param[in] key The key to set.
     * @param[in] value The value for the key.
     * @param[in] setParam Set option of this key.
     * @return if set fail, throw YR::Exception.
     */
    virtual ErrorInfo Write(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam) = 0;

    /**
     * @brief Write Multiple KV pair in a transaction
     * @param[in] key The key to set.
     * @param[in] value The value for the key.
     * @param[in] mSetParam MSet option of this key.
     * @return ErrorInfo
     */
    virtual ErrorInfo MSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<Buffer>> &vals,
                             const MSetParam &mSetParam) = 0;

    /**
     * @brief Read the value of a keyby datasystem StateClient.
     * @param[in] key The key to get.
     * @return return a shared_ptr<Buffer>,
     * return {} if the key not found.
     */
    virtual SingleReadResult Read(const std::string &key, int timeoutMS) = 0;

    /**
     * @brief Read the values of all the given keys by datasystem StateClient.
     * @param[in] keys The vector of the keys.
     * @return return a vector of shared_ptr<Buffer>>,
     * return std::vector<>{} if any key not found.
     */
    virtual MultipleReadResult GetWithParam(const std::vector<std::string> &keys, const GetParams &params,
                                            int timeoutMs) = 0;

    /**
     * @brief Read the values of all the given keys by datasystem StateClient.
     * @param[in] keys The vector of the keys.
     * @return return a vector of shared_ptr<Buffer>>,
     * return std::vector<>{} if any key not found.
     */
    virtual MultipleReadResult Read(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial) = 0;

    /**
     * @brief Query the value sizes of all the given keys by datasystem StateClient.
     * @param[in] keys The vector of the keys.
     * @param[in] outSizes A vector of value sizes that query from datasystem
     * @return return ErrorInfo
     */
    virtual ErrorInfo QuerySize(const std::vector<std::string> &keys, std::vector<uint64_t> &outSizes) = 0;

    /**
     * @brief Delete a key by datasystem StateClient.
     * @param[in] key The key to delete.
     * @return The count of success delete key.
     */
    virtual ErrorInfo Del(const std::string &key) = 0;

    /**
     * @brief Delete all the given keys by datasystem StateClient.
     * @param[in] keys The vector of the keys.
     * @return The failedKeys and ErrorInfo.
     */
    virtual MultipleDelResult Del(const std::vector<std::string> &keys) = 0;

    /**
     * @brief Query all the given keys by datasystem StateClient.
     * @param[in] keys The vector of the keys.
     * @return MultipleExistResult The exists result and ErrorInfo.
     */
    virtual MultipleExistResult Exist(const std::vector<std::string> &keys) = 0;

    /**
     * @brief Shutdown notify datasystem to release resource.
     */
    virtual void Shutdown() = 0;

    /**
     * @brief update token
     */
    virtual ErrorInfo UpdateToken(datasystem::SensitiveValue token) = 0;

    virtual ErrorInfo GenerateKey(std::string &returnKey) = 0;

    virtual ErrorInfo StartHealthCheck() = 0;

    virtual ErrorInfo UpdateAkSk(std::string ak, datasystem::SensitiveValue sk) = 0;

    virtual ErrorInfo Write(std::shared_ptr<Buffer> value, SetParam setParam, std::string &returnKey) = 0;

    virtual ErrorInfo HealthCheck() = 0;
};

}  // namespace Libruntime
}  // namespace YR