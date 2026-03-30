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

#include "security.h"
#include <thread>

#include <boost/bind/bind.hpp>
#include "src/dto/buffer.h"
#include "src/dto/config.h"
#include "src/libruntime/fsclient/protobuf/common.grpc.pb.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
using TLSConfig = ::common::TLSConfig;
const int LOG_FREQUENT = 10000;
Security::Security(int confFileNo, size_t stdinPipeTimeoutMs)
    : streamReaderIoContext_(),
      confStreamDesc_(streamReaderIoContext_, dup(confFileNo)),
      stdinPipeTimeoutMs_(stdinPipeTimeoutMs)
{
}

ErrorInfo Security::Init()
{
    if (this->GetReadableSize() == 0) {
        if (!Config::Instance().ENABLE_DS_AUTH() && !Config::Instance().ENABLE_SERVER_AUTH()) {
            YRLOG_INFO("Skip init security because zero readable size");
            return ErrorInfo();
        }
    }

    this->streamReaderThread_ = std::make_unique<std::thread>([this]() {
        auto work = boost::asio::make_work_guard(this->streamReaderIoContext_);
        this->streamReaderIoContext_.run();
    });

    if (this->GetReadableSize() == 0) {
        YRLOG_INFO("readable size is 0, wait until having data from stdin or timeout");
        auto notify = std::make_shared<YR::utility::NotificationUtility>();
        confStreamDesc_.async_wait(
            boost::asio::posix::stream_descriptor::wait_read,
            boost::bind(&Security::StreamReaderWaitHandler, this, boost::asio::placeholders::error, notify));
        auto err = notify->WaitForNotificationWithTimeout(
            absl::Milliseconds(stdinPipeTimeoutMs_),
            YR::Libruntime::ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                      YR::Libruntime::ModuleCode::RUNTIME, "read stdin pipe timeout"));
        if (!err.OK()) {
            return err;
        }
    }
    if (!this->ReadOnce()) {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, "Failed to read config from stream");
    }

    confStreamDesc_.async_wait(boost::asio::posix::stream_descriptor::wait_read,
                               boost::bind(&Security::StreamReaderWaitHandler, this, boost::asio::placeholders::error));
    return ErrorInfo();
}

ErrorInfo Security::InitWithDriver(std::shared_ptr<LibruntimeConfig> librtConfig)
{
    YRLOG_DEBUG("When init security as driver, enableMTLS is {}, enableAuth is {}, ak is empty: {}, sk is empty: {}",
                librtConfig->enableMTLS, librtConfig->enableAuth, librtConfig->ak_.empty(), librtConfig->sk_.Empty());
    if (librtConfig->enableMTLS) {
        this->fsConf_.authEnable = librtConfig->enableMTLS;
        STACK_OF(X509) *ca = GetCAFromFile(librtConfig->verifyFilePath);
        X509 *cert = GetCertFromFile(librtConfig->certificateFilePath);
        EVP_PKEY *pkey = GetPrivateKeyFromFile(librtConfig->privateKeyPath, librtConfig->privateKeyPaaswd);
        this->fsConf_.rootCertData = GetCa(ca);
        this->fsConf_.certChainData = GetCert(cert);
        this->fsConf_.privateKeyData = GetPrivateKey(pkey);
        this->serverNameoverride_ = librtConfig->serverName;
        ClearPemCerts(pkey, cert, ca);
    }
    if (librtConfig->encryptEnable) {
        this->dsConf_.encryptEnable = librtConfig->encryptEnable;
        this->dsConf_.clientPublicKey = GetValueFromFile(librtConfig->runtimePublicKeyPath);
        this->dsConf_.clientPrivateKey = GetValueFromFile(librtConfig->runtimePrivateKeyPath);
        this->dsConf_.serverPublicKey = GetValueFromFile(librtConfig->dsPublicKeyPath);
    }
    if (!librtConfig->ak_.empty() && !librtConfig->sk_.Empty()) {
        this->ak_ = librtConfig->ak_;
        this->sk_ = SensitiveValue(librtConfig->sk_);
        this->isCredential_ = true;
        this->dsConf_.authEnable = true;
        if (!librtConfig->dk_.Empty()) {
            this->dk_ = SensitiveValue(librtConfig->dk_);
        }
    }
    return ErrorInfo();
}

std::string Security::GetValueFromFile(const std::string &path)
{
    if (path.empty()) {
        YRLOG_ERROR("when encryptEnable param is true and path is empty, return empty res directly");
        return "";
    }
    std::ifstream file(path);
    if (!file.is_open()) {
        YRLOG_ERROR("wrong file path and return empty res directly");
        return "";
    }
    std::string content;
    std::string line;
    while (std::getline(file, line)) {
        content += line;
    }
    file.close();
    return content;
}

void Security::StreamReaderWaitHandler(const boost::system::error_code &error)
{
    if (error) {
        YRLOG_ERROR("Reader waiting error: {}", error.message());
        confStreamDesc_.async_wait(
            boost::asio::posix::stream_descriptor::wait_read,
            boost::bind(&Security::StreamReaderWaitHandler, this, boost::asio::placeholders::error));
        return;
    }

    if (this->GetReadableSize() == 0) {
        // async_wait is triggered, but there is no data, indicating that the write end has been closed (EOF). In fact,
        // there is no need to continue listening, but we still add a sleep here to prevent the write end from being
        // reopened. After the graceful exit time (default 5s) expires, the runtime will be killed with kill -9.
        static const int sleepTime = 1;
        std::this_thread::sleep_for(std::chrono::seconds(sleepTime));
    }
    if (!this->ReadOnce()) {
        YRLOG_DEBUG_COUNT(LOG_FREQUENT, "Reader read once failed");
    } else {
        // in wait handler, token is expected to be updated, so call all token handlers
        for (auto &h : this->updateTokenHandlers) {
            h(this->token_);
        }
        YRLOG_INFO("token is updated.");
        for (auto &h : this->updateAkSkHandlers) {
            h(this->ak_, this->sk_);
        }
        YRLOG_INFO("ak,sk is updated.");
    }
    confStreamDesc_.async_wait(boost::asio::posix::stream_descriptor::wait_read,
                               boost::bind(&Security::StreamReaderWaitHandler, this, boost::asio::placeholders::error));
}

void Security::StreamReaderWaitHandler(const boost::system::error_code &error,
                                       std::shared_ptr<YR::utility::NotificationUtility> notify)
{
    if (error) {
        notify->Notify(YR::Libruntime::ErrorInfo(ERR_INNER_SYSTEM_ERROR, "Reader waiting error: " + error.message()));
        return;
    }
    size_t sleepMs = 100;
    int retryTimes = stdinPipeTimeoutMs_ / sleepMs;
    while (retryTimes-- >= 0) {
        size_t readableSize = this->GetReadableSize();
        if (readableSize == 0) {
            YRLOG_INFO("stdin readable data size is 0");
        } else {
            YRLOG_INFO("stdin have readable data size: {}", readableSize);
            notify->Notify();
            return;
        }
        std::this_thread::sleep_for(std::chrono::milliseconds(sleepMs));
    }
    notify->Notify(YR::Libruntime::ErrorInfo(ERR_INNER_SYSTEM_ERROR, "wait stdin timeout"));
}

size_t Security::GetReadableSize()
{
    boost::asio::posix::descriptor_base::bytes_readable command;
    boost::system::error_code ec;
    confStreamDesc_.io_control(command, ec);
    if (ec) {
        YRLOG_ERROR("Get readable size error: {}", ec.message());
        return 0;
    }
    return command.get();
}

bool Security::ReadOnce()
{
    size_t readableSize = this->GetReadableSize();
    if (readableSize == 0) {
        YRLOG_DEBUG_COUNT(LOG_FREQUENT, "Not readable");
        return false;
    }

    auto nativeBuf = std::make_shared<NativeBuffer>(readableSize);
    boost::system::error_code ec;
    auto readSize = confStreamDesc_.read_some(
        boost::asio::buffer(static_cast<char *>(nativeBuf->MutableData()), static_cast<size_t>(nativeBuf->GetSize())),
        ec);
    if (ec || readSize == 0) {
        YRLOG_ERROR("Read from stream, error code: {}, read size: {}", ec.message(), readSize);
        return false;
    }

    TLSConfig tlsConf;
    if (!tlsConf.ParseFromArray(nativeBuf->MutableData(), nativeBuf->GetSize())) {
        YRLOG_ERROR("Parse tls config failed, read size: {}", nativeBuf->GetSize());
        return false;
    }

    this->dsConf_.authEnable = tlsConf.dsauthenable();
    this->dsConf_.encryptEnable = tlsConf.dsencryptenable();
    this->dsConf_.clientPublicKey = tlsConf.dsclientpublickey();
    this->dsConf_.clientPrivateKey = SensitiveValue(tlsConf.dsclientprivatekey());
    this->dsConf_.serverPublicKey = tlsConf.dsserverpublickey();

    this->fsConf_.authEnable = tlsConf.serverauthenable();
    this->fsConf_.rootCertData = tlsConf.rootcertdata();

    if (tlsConf.has_tenantcredentials() && !tlsConf.tenantcredentials().accesskey().empty()) {
        this->ak_ = tlsConf.tenantcredentials().accesskey();
    } else {
        this->ak_ = tlsConf.accesskey();
    }

    if (tlsConf.has_tenantcredentials() && !tlsConf.tenantcredentials().secretkey().empty()) {
        this->sk_ = SensitiveValue(tlsConf.tenantcredentials().secretkey());
    } else {
        this->sk_ = SensitiveValue(tlsConf.securitykey());
    }

    if (tlsConf.has_tenantcredentials() && !tlsConf.tenantcredentials().datakey().empty()) {
        this->dk_ = SensitiveValue(tlsConf.tenantcredentials().datakey());
    }

    if (tlsConf.has_tenantcredentials()) {
        this->isCredential_ = tlsConf.tenantcredentials().iscredential();
    }

    this->token_ = SensitiveValue(tlsConf.token());

    this->fsConnMode_ = tlsConf.enableservermode();

    this->serverNameoverride_ = tlsConf.servernameoverride();
    YRLOG_INFO("Read tls config finished, fs auth: {}, ds auth: {}, is credential {}, ak {}, sk {}, token {}",
               this->fsConf_.authEnable, this->dsConf_.authEnable, isCredential_, !ak_.empty(), !sk_.Empty(),
               !token_.Empty());
    return true;
}

void Security::Stop(void)
{
    boost::system::error_code ec;
    confStreamDesc_.cancel(ec);
    this->streamReaderIoContext_.stop();
    if (this->streamReaderThread_) {
        this->streamReaderThread_->join();
    }
    confStreamDesc_.close();
}

Security::~Security()
{
    this->Stop();
}

std::pair<bool, bool> Security::GetDataSystemConfig(std::string &clientPublicKey, SensitiveValue &clientPrivateKey,
                                                    std::string &serverPublicKey)
{
    clientPublicKey = this->dsConf_.clientPublicKey;
    clientPrivateKey = this->dsConf_.clientPrivateKey;
    serverPublicKey = this->dsConf_.serverPublicKey;
    return std::make_pair<>(this->dsConf_.authEnable, this->dsConf_.encryptEnable);
}

bool Security::GetFunctionSystemConnectionMode(std::string &serverNameOverride)
{
    serverNameOverride = this->serverNameoverride_;
    return this->fsConnMode_;
}

bool Security::GetFunctionSystemConfig(std::string &rootCACert, std::string &certChain, std::string &privateKey)
{
    rootCACert = this->fsConf_.rootCertData;
    certChain = this->fsConf_.certChainData;
    privateKey = std::string(this->fsConf_.privateKeyData.GetData(), this->fsConf_.privateKeyData.GetSize());
    return this->fsConf_.authEnable;
}

void Security::GetToken(SensitiveValue &token)
{
    token = this->token_;
}

void Security::GetAKSK(std::string &ak, SensitiveValue &sk)
{
    ak = this->ak_;
    sk = this->sk_;
}

void Security::GetAKSKDK(std::string &ak, SensitiveValue &sk, SensitiveValue &dk)
{
    ak = this->ak_;
    sk = this->sk_;
    dk = this->dk_;
}

void Security::WhenTokenUpdated(std::function<void(const SensitiveValue &)> updateTokenHandler)
{
    if (ak_.empty() && sk_.Empty()) {
        this->updateTokenHandlers.push_back(updateTokenHandler);
    }
}

void Security::WhenAkSkUpdated(std::function<void(const std::string &, const SensitiveValue &)> updateAkSkHandler)
{
    if (!ak_.empty() && !sk_.Empty()) {
        this->updateAkSkHandlers.push_back(updateAkSkHandler);
    }
}

void Security::ClearPrivateKey()
{
    this->fsConf_.privateKeyData.Clear();
}

int Security::GetUpdateHandersSize()
{
    return this->updateTokenHandlers.size();
}

bool Security::IsFsAuthEnable()
{
    return this->isCredential_ && !this->ak_.empty() && !this->sk_.Empty();
}

Credential Security::GetCredential()
{
    return Credential{ak : this->ak_,
                      sk : std::string(this->sk_.GetData(), this->sk_.GetSize()),
                      dk : std::string(this->dk_.GetData(), this->dk_.GetSize())};
}

void Security::SetAKSKAndCredential(const std::string &ak, const SensitiveValue &sk)
{
    this->ak_ = ak;
    this->sk_ = sk;
    this->isCredential_ = true;
}

}  // namespace Libruntime
}  // namespace YR