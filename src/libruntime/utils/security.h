/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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
#include <functional>
#include <list>

#include "boost/asio.hpp"
#include "boost/iostreams/stream.hpp"

#include "certs_utils.h"
#include "datasystem/utils/sensitive_value.h"
#include "sensitive_data.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/libruntime_config.h"
#include "src/utility/notification_utility.h"

namespace YR {
namespace Libruntime {
using SensitiveValue = datasystem::SensitiveValue;
using SensitiveData = YR::Libruntime::SensitiveData;
const size_t DEFAULT_STDIN_PIPE_TIMEOUT_MS = 30000;
class Security {
public:
    /**
     * @brief Get the Data System Auth Config For Object-client, State-client and Stream-client
     *
     * @param clientPublicKey client public key
     * @param clientPrivateKey client private key
     * @param serverPublicKey server public key
     * @return true data system auth enabled
     * @return false data system auth disabled
     */
    std::pair<bool, bool> GetDataSystemConfig(std::string &clientPublicKey, SensitiveValue &clientPrivateKey,
                                              std::string &serverPublicKey);

    /**
     * @brief Get the Function System Connection Mode
     *
     * @param serverNameOverride function system server name override
     * @return true function system is server, runtime is client
     * @return false function system is client, runtime is server
     */
    bool GetFunctionSystemConnectionMode(std::string &serverNameOverride);

    /**
     * @brief Get the Function System Config
     *
     * @param rootCACert root CA's certificates
     * @param certChain cert chain
     * @param privateKey private key
     * @return true function system auth enabled
     * @return false function system auth disabled
     */
    bool GetFunctionSystemConfig(std::string &rootCACert, std::string &certChain, std::string &privateKey);

    /**
     * @brief Get the Token
     *
     * @param token runtime's session token for build connections with function system and data system
     */
    void GetToken(SensitiveValue &token);

    /**
     * @brief Get the AK, SK and DK
     *
     * @param ak system function runtime's access key for build connections with data system
     * @param sk system function runtime's security key for build connections with data system
     * @param dk system function runtime's data key for build connections with data system
     */
    virtual void GetAKSKDK(std::string &ak, SensitiveValue &sk, SensitiveValue &dk);

    /**
     * @brief Get the AK and SK
     *
     * @param ak system function runtime's access key for build connections with data system
     * @param sk system function runtime's security key for build connections with data system
     */
    virtual void GetAKSK(std::string &ak, SensitiveValue &sk);

    /**
     * @brief register handler when token updated
     *
     * @param updateTokenHandler registered token updated event handler, input parameter is the token updated
     */
    void WhenTokenUpdated(std::function<void(const SensitiveValue &)> updateTokenHandler);

    /**
     * @brief register handler when token updated
     *
     * @param updateAkSkHandler registered aksk updated event handler, input parameter is the aksk updated
     */
    void WhenAkSkUpdated(std::function<void(const std::string &, const SensitiveValue &)> updateAkSkHandler);

    /**
     * @brief clear private key
     */
    void ClearPrivateKey();

    int GetUpdateHandersSize();

    virtual bool IsFsAuthEnable();

    Credential GetCredential();

    Security(int confFileNo = STDIN_FILENO, size_t stdinPipeTimeoutMs = DEFAULT_STDIN_PIPE_TIMEOUT_MS);

    virtual ~Security();

    virtual ErrorInfo Init();

    virtual ErrorInfo InitWithDriver(std::shared_ptr<LibruntimeConfig> librtConfig);

    virtual void SetAKSKAndCredential(const std::string &ak, const SensitiveValue &sk);

    std::string GetValueFromFile(const std::string &path);

private:
    void StreamReaderWaitHandler(const boost::system::error_code &error,
                                 std::shared_ptr<YR::utility::NotificationUtility> notify);

    void StreamReaderWaitHandler(const boost::system::error_code &error);

    size_t GetReadableSize();

    void Stop(void);

    bool ReadOnce();

    boost::asio::io_context streamReaderIoContext_;
    boost::asio::posix::stream_descriptor confStreamDesc_;
    std::unique_ptr<std::thread> streamReaderThread_;

    struct DataSystemSecurityConfig {
        bool authEnable = false;
        bool encryptEnable = false;
        std::string clientPublicKey = "";
        SensitiveValue clientPrivateKey = "";
        std::string serverPublicKey = "";
    };
    DataSystemSecurityConfig dsConf_;

    struct FunctionSystemSecurityConfig {
        bool authEnable = false;
        std::string rootCertData = "";
        std::string certChainData = "";
        SensitiveData privateKeyData;
    };
    FunctionSystemSecurityConfig fsConf_;
    SensitiveValue token_ = "";
    std::string ak_ = "";
    SensitiveValue sk_ = "";
    SensitiveValue dk_ = "";
    bool isCredential_ = false;  // true means runtime auth with ds and fs use ak、sk
    bool fsConnMode_ = false;    // false means runtime is server, function system is client
    std::string serverNameoverride_ = "";

    std::list<std::function<void(const SensitiveValue &)>> updateTokenHandlers;
    std::list<std::function<void(const std::string &, const SensitiveValue &)>> updateAkSkHandlers;
    size_t stdinPipeTimeoutMs_;
};
}  // namespace Libruntime
}  // namespace YR