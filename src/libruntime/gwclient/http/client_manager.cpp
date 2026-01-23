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

#include <boost/beast/ssl.hpp>
#include <chrono>

#include "src/dto/config.h"
#include "src/libruntime/gwclient/http/async_http_client.h"
#include "src/libruntime/gwclient/http/async_https_client.h"
#include "src/libruntime/gwclient/http/client_manager.h"
#include "src/utility/logger/logger.h"
#include "src/utility/notification_utility.h"

namespace {
const int RETRY_TIME = 3;
const int INTERVAL_TIME = 2;
}  // namespace

namespace ssl = boost::asio::ssl;
namespace YR {
namespace Libruntime {
using YR::utility::NotificationUtility;
ClientManager::ClientManager(const std::shared_ptr<LibruntimeConfig> &libruntimeConfig) : librtCfg(libruntimeConfig)
{
    this->ioc = std::make_shared<boost::asio::io_context>();
    this->work = std::make_unique<boost::asio::executor_work_guard<boost::asio::io_context::executor_type>>(
        boost::asio::make_work_guard(*ioc));
    this->maxIocThread = libruntimeConfig->httpIocThreadsNum;
    this->enableMTLS = libruntimeConfig->enableMTLS;
    this->maxConnSize_ = libruntimeConfig->maxConnSize;
    this->enableTLS_ = libruntimeConfig->enableTLS;
}

ClientManager::~ClientManager()
{
    Stop();
}

void ClientManager::Stop()
{
    if (stopped_) {
        return;
    }
    stopped_ = true;
    try {
        this->work->reset();
        ioc->stop();
        for (auto client : clients) {
            client->Stop();
        }
        for (uint32_t i = 0; i < asyncRunners.size(); i++) {
            if (this->asyncRunners[i]->get_id() != std::this_thread::get_id()) {
                this->asyncRunners[i]->join();
            }
        }
    } catch (const std::exception &e) {
        YRLOG_ERROR("caught exception when stop client manager: {}", e.what());
    } catch (...) {
        YRLOG_ERROR("caught unknown exception when stop client manager");
    }
}

ErrorInfo ClientManager::InitCtxAndIocThread()
{
    ErrorInfo err;
    if (enableMTLS) {
        try {
            auto ctx = std::make_shared<ssl::context>(ssl::context::tlsv12_client);
            ctx->set_options(ssl::context::default_workarounds | ssl::context::no_sslv2 | ssl::context::no_sslv3 |
                             ssl::context::no_tlsv1 | ssl::context::no_tlsv1_1);
            ctx->set_verify_mode(ssl::verify_peer);
            ctx->load_verify_file(librtCfg->verifyFilePath);
            ctx->use_certificate_chain_file(librtCfg->certificateFilePath);
            ctx->set_password_callback([&](std::size_t max_length, ssl::context_base::password_purpose purpose) {
                return std::string(librtCfg->privateKeyPaaswd);
            });
            ctx->use_private_key_file(librtCfg->privateKeyPath, ssl::context::pem);
            for (uint32_t i = 0; i < maxConnSize_; i++) {
                this->clients.emplace_back(std::make_shared<AsyncHttpsClient>(this->ioc, ctx, librtCfg->serverName));
            }
        } catch (const std::exception &e) {
            YRLOG_ERROR("caught exception when init context : {}", e.what());
            err.SetErrCodeAndMsg(ErrorCode::ERR_INIT_CONNECTION_FAILED, ModuleCode::RUNTIME, e.what());
            return err;
        } catch (...) {
            YRLOG_ERROR("caught unknown exception when init context");
            err.SetErrCodeAndMsg(ErrorCode::ERR_INIT_CONNECTION_FAILED, ModuleCode::RUNTIME,
                                 "caught unknown exception when init context");
            return err;
        }
    } else if (enableTLS_) {
        auto ctx = std::make_shared<ssl::context>(ssl::context::tlsv12_client);
        ctx->set_options(ssl::context::default_workarounds | ssl::context::no_sslv2 | ssl::context::no_sslv3 |
                         ssl::context::no_tlsv1 | ssl::context::no_tlsv1_1);
        if (librtCfg->verifyFilePath.empty()) {
            ctx->set_default_verify_paths();
        } else {
            ctx->load_verify_file(librtCfg->verifyFilePath);
        }
        ctx->set_verify_mode(ssl::verify_peer);
        for (uint32_t i = 0; i < maxConnSize_; i++) {
            this->clients.emplace_back(std::make_shared<AsyncHttpsClient>(this->ioc, ctx, librtCfg->serverName));
        }
    } else {
        for (uint32_t i = 0; i < maxConnSize_; i++) {
            this->clients.emplace_back(std::make_shared<AsyncHttpClient>(this->ioc));
        }
    }

    for (uint32_t i = 0; i < maxIocThread; i++) {
        asyncRunners.push_back(std::make_unique<std::thread>([&] { this->ioc->run(); }));
        std::string name = "yr_client_io_" + std::to_string(i);
        pthread_setname_np(this->asyncRunners[i]->native_handle(), name.c_str());
    }
    return err;
}

ErrorInfo ClientManager::Init(const ConnectionParam &param)
{
    ErrorInfo error = InitCtxAndIocThread();
    if (!error.OK()) {
        return error;
    }
    this->connParam = param;
    connectedClientsCnt_ = YR::Libruntime::Config::Instance().YR_HTTP_CONNECTION_NUM();
    YRLOG_INFO("http initial connection num {}", connectedClientsCnt_);
    for (uint32_t i = 0; i < connectedClientsCnt_; i++) {
        for (int j = 0; j < RETRY_TIME; j++) {
            error = clients[i]->Init(param);
            clients[i]->SetAvailable();
            if (error.OK()) {
                break;
            }
            std::this_thread::sleep_for(std::chrono::seconds(INTERVAL_TIME));
        }
        if (!error.OK()) {
            return ErrorInfo(ErrorCode::ERR_INIT_CONNECTION_FAILED, ModuleCode::RUNTIME, error.Msg());
        }
    }
    return error;
}

void ClientManager::SubmitInvokeRequest(const http::verb &method, const std::string &target,
                                        const std::unordered_map<std::string, std::string> &headers,
                                        const std::string &body, const std::shared_ptr<std::string> requestId,
                                        const HttpCallbackFunction &receiver)
{
    for (;;) {
        if (SubmitRequest(method, target, headers, body, requestId, receiver)) {
            break;
        }
        std::this_thread::yield();
    }
}

bool ClientManager::SubmitRequest(const http::verb &method, const std::string &target,
                                  const std::unordered_map<std::string, std::string> &headers, const std::string &body,
                                  const std::shared_ptr<std::string> requestId, const HttpCallbackFunction &receiver)
{
    for (uint32_t i = 0;; i++) {
        {
            absl::ReaderMutexLock l(&connCntMu_);
            if (i >= connectedClientsCnt_) {
                break;
            }
        }
        if (!this->clients[i]->SetUnavailable()) {
            continue;
        }
        YRLOG_DEBUG("httpclient {} is available, will use this client. requestId: {}", i, *requestId);
        // while the connection idletime exceed setup timeout, the server may close the connection
        // in this situation, client should try to reconnect
        if (!this->clients[i]->IsConnActive()) {
            YRLOG_DEBUG("httpclient {} is not active, reinit now. requestId: {}", i, *requestId);
            auto err = this->clients[i]->ReInit(requestId);
            if (!err.OK()) {
                YRLOG_DEBUG("httpclient {} is reInit failed, requestId:{} err: {}", i, *requestId, err.CodeAndMsg());
                receiver(err.CodeAndMsg(), boost::asio::error::make_error_code(boost::asio::error::connection_reset),
                         HTTP_CONNECTION_ERROR_CODE);
                this->clients[i]->SetAvailable();
                return true;
            }
        }
        this->clients[i]->SubmitInvokeRequest(method, target, headers, body, requestId, receiver);
        return true;
    }
    uint32_t newClientIdx = 0;
    {
        absl::WriterMutexLock l(&connCntMu_);
        if (connectedClientsCnt_ >= maxConnSize_) {
            return false;
        }
        newClientIdx = connectedClientsCnt_++;
    }
    YRLOG_DEBUG("init httpclient {}", newClientIdx);
    this->clients[newClientIdx]->Init(this->connParam);
    this->clients[newClientIdx]->SubmitInvokeRequest(method, target, headers, body, requestId, receiver);
    return true;
}
}  // namespace Libruntime
}  // namespace YR
