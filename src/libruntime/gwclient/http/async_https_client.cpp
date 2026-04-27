/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#include "src/libruntime/gwclient/http/async_https_client.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
AsyncHttpsClient::AsyncHttpsClient(const std::shared_ptr<asio::io_context> &ioc,
                                   const std::shared_ptr<asio::ssl::context> &ctx, std::string serverName)
    : ioc_(ioc), ctx_(ctx), serverName_(std::move(serverName)), resolver_(asio::make_strand(*ioc))
{
}

AsyncHttpsClient::~AsyncHttpsClient()
{
    if (IsConnActive()) {
        GracefulExit();
    }
}

void AsyncHttpsClient::SubmitInvokeRequest(const http::verb &method, const std::string &target,
                                           const std::unordered_map<std::string, std::string> &headers,
                                           const std::string &body, const std::shared_ptr<std::string> requestId,
                                           const HttpCallbackFunction &receiver)
{
    retried_ = false;
    callback_ = receiver;
    req_ = {method, target, DEFAULT_HTTP_VERSION};
    for (auto &iter : headers) {
        req_.set(iter.first, iter.second);
    }
    req_.set("HOST", connParam_.ip + ":" + connParam_.port);
    req_.body() = body;
    req_.prepare_payload();
    resParser_ = std::make_shared<http::response_parser<http::string_body>>();
    resParser_->body_limit(std::numeric_limits<std::uint64_t>::max());
    http::async_write(*stream_, req_,
                      beast::bind_front_handler(&AsyncHttpsClient::OnWrite, shared_from_this(), requestId));
}

ErrorInfo AsyncHttpsClient::Init(const ConnectionParam &param)
{
    // A new stream_ must be generated for the reconnection. Otherwise, an error is reported:
    // protocol is shutdown(SSL routines, ssl write internal)
    stream_ = std::make_shared<beast::ssl_stream<beast::tcp_stream>>(asio::make_strand(*ioc_), *ctx_);
    // Set SNI Hostname (hosts need this to handshake successfully)
    YRLOG_INFO("Https init, serverAddr = {}:{}", param.ip, param.port);
    connParam_ = param;
    idleTime_ = connParam_.idleTime;
    std::string msg;
    const auto &tlsServerName = serverName_.empty() ? param.ip : serverName_;
    if (!tlsServerName.empty()) {
        if (!SSL_set_tlsext_host_name(stream_->native_handle(), tlsServerName.c_str())) {
            YRLOG_ERROR("failed to set servername: {}", tlsServerName);
            msg = "failed to set servername during initing invoke client, serverName:" + tlsServerName;
            return ErrorInfo(ErrorCode::ERR_INIT_CONNECTION_FAILED, ModuleCode::RUNTIME, msg);
        }
    }
    try {
        // sync connect
        auto const results = resolver_.resolve(param.ip, param.port);
        auto &lowgest = beast::get_lowest_layer(*stream_);
        if (param.timeoutSec != CONNECTION_NO_TIMEOUT) {
            lowgest.expires_after(std::chrono::seconds(param.timeoutSec));
        }
        (void)lowgest.connect(results);
        YRLOG_DEBUG("Https init successfully, serverAddr: {}:{} connectionTimeout = {}", param.ip, param.port,
                    param.timeoutSec);
        if (param.timeoutSec != CONNECTION_NO_TIMEOUT) {
            lowgest.expires_never();
        }
    } catch (const std::exception &e) {
        std::stringstream ss;
        ss << "failed to connect to cluster, target: " << param.ip << ":" << param.port;
        ss << ", exception: " << e.what();
        YRLOG_DEBUG(ss.str());
        return ErrorInfo(ErrorCode::ERR_INIT_CONNECTION_FAILED, ModuleCode::RUNTIME, ss.str());
    }

    boost::system::error_code ec;
    stream_->handshake(ssl::stream_base::client, ec);
    if (ec) {
        YRLOG_ERROR("handshake error: {}", ec.message().c_str());
        msg = "failed to handshake during initing invoke client, host:" + param.ip;
        return ErrorInfo(ErrorCode::ERR_INIT_CONNECTION_FAILED, ModuleCode::RUNTIME, msg);
    }
    ResetConnActive();
    return ErrorInfo();
}

void AsyncHttpsClient::OnWrite(const std::shared_ptr<std::string> requestId, const beast::error_code &ec,
                               std::size_t bytesTransferred)
{
    boost::ignore_unused(bytesTransferred);
    if (!ec) {
        http::async_read(*stream_, this->buf_, *resParser_,
                         beast::bind_front_handler(&AsyncHttpsClient::OnRead, shared_from_this(), requestId));
        return;
    }
    YRLOG_ERROR("requestId {} failed to write, err message: {}, this client disconnect", *requestId,
                ec.message().c_str());
    if (!retried_) {
        YRLOG_DEBUG("requestId {} start to retry once", *requestId);
        retried_ = true;
        if (ReInit(requestId).OK()) {
            http::async_write(*stream_, req_,
                              beast::bind_front_handler(&AsyncHttpsClient::OnWrite, shared_from_this(), requestId));
            return;
        }
    }
    if (callback_) {
        callback_(HTTP_CONNECTION_ERROR_MSG, ec, HTTP_CONNECTION_ERROR_CODE);
    }
    SetConnInActive();
    SetAvailable();
}

void AsyncHttpsClient::OnRead(const std::shared_ptr<std::string> requestId, const beast::error_code &ec,
                              std::size_t bytesTransferred)
{
    boost::ignore_unused(bytesTransferred);
    if (ec) {
        YRLOG_ERROR("requestId {} failed to read response , err message: {}, this client disconnect", *requestId,
                    ec.message().c_str());
        SetConnInActive();
    }

    if (callback_) {
        callback_(resParser_->get().body(), ec, resParser_->get().result_int());
    }
    CheckResponseHeaderAndReset();
}

void AsyncHttpsClient::GracefulExit() noexcept
{
    SetConnInActive();
    boost::system::error_code ec = {};
    if (stream_) {
        YRLOG_DEBUG("start shutdown ssl stream.");
        stream_->shutdown(ec);
        if (ec) {
            YRLOG_WARN("SSL shutdown failed: {}", ec.message().c_str());
            return;
        }
        auto &sock = stream_->next_layer().socket();
        sock.shutdown(boost::asio::ip::tcp::socket::shutdown_both, ec);
        if (ec) {
            YRLOG_WARN("Socket shutdown failed: {}", ec.message().c_str());
            return;
        }
        sock.close(ec);
        if (ec) {
            YRLOG_WARN("Socket close failed: {}", ec.message().c_str());
            return;
        }
    }
}

void AsyncHttpsClient::Stop()
{
    if (stream_) {
        stream_.reset();
    }
}
}  // namespace Libruntime
}  // namespace YR
