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

#include "src/libruntime/gwclient/http/async_http_client.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
const char *HTTP_CONNECTION_ERROR_MSG = "connection error";
const int DEFAULT_HTTP_VERSION = 11;
const uint HTTP_CONNECTION_ERROR_CODE = 999;
AsyncHttpClient::AsyncHttpClient(std::shared_ptr<asio::io_context> ctx)
    : resolver_(asio::make_strand(*ctx)), stream_(asio::make_strand(*ctx))
{
}

AsyncHttpClient::~AsyncHttpClient()
{
    if (IsConnActive()) {
        GracefulExit();
    }
}

void AsyncHttpClient::GracefulExit() noexcept
{
    SetConnInActive();
    beast::error_code ec;
    stream_.socket().shutdown(asio::ip::tcp::socket::shutdown_both, ec);
    if (ec) {
        YRLOG_WARN("failed to shutdown stream: {}", ec.message().c_str());
    }
    stream_.close();
}

void AsyncHttpClient::SubmitInvokeRequest(const http::verb &method, const std::string &target,
                                          const std::unordered_map<std::string, std::string> &headers,
                                          const std::string &body, const std::shared_ptr<std::string> requestId,
                                          const HttpCallbackFunction &receiver)
{
    callback_ = receiver;
    retried_ = false;
    req_ = {method, target, DEFAULT_HTTP_VERSION};
    for (auto &iter : headers) {
        req_.set(iter.first, iter.second);
    }
    req_.set("HOST", this->connParam_.ip + ":" + this->connParam_.port);
    req_.body() = body;
    // set body size in headers
    req_.prepare_payload();
    resParser_ = std::make_shared<http::response_parser<http::string_body>>();
    resParser_->body_limit(std::numeric_limits<std::uint64_t>::max());
    http::async_write(stream_, req_,
                      beast::bind_front_handler(&AsyncHttpClient::OnWrite, shared_from_this(), requestId));
}

ErrorInfo AsyncHttpClient::Init(const ConnectionParam &param)
{
    YRLOG_DEBUG("Http init, serverAddr = {}:{}", param.ip, param.port);
    connParam_ = param;
    idleTime_ = param.idleTime;
    // sync connection
    try {
        auto const resolveRes = resolver_.resolve(param.ip, param.port);
        if (param.timeoutSec != CONNECTION_NO_TIMEOUT) {
            stream_.expires_after(std::chrono::seconds(param.timeoutSec));
        }
        stream_.connect(resolveRes);
        if (param.timeoutSec != CONNECTION_NO_TIMEOUT) {
            stream_.expires_never();
        }
    } catch (const std::exception &e) {
        std::stringstream ss;
        ss << "failed to connect to cluster, target: ";
        ss << param.ip;
        ss << ":";
        ss << param.port;
        ss << ", exception: ";
        ss << e.what();
        YRLOG_DEBUG(ss.str());
        return ErrorInfo(ErrorCode::ERR_INIT_CONNECTION_FAILED, ModuleCode::RUNTIME, ss.str());
    }
    ResetConnActive();
    return ErrorInfo();
}

void AsyncHttpClient::OnRead(const std::shared_ptr<std::string> requestId, const beast::error_code &ec,
                             std::size_t bytesTransferred)
{
    boost::ignore_unused(bytesTransferred);
    if (ec) {
        YRLOG_ERROR("requestId {} failed to read response , err message: {}, client disconnect", *requestId,
                    ec.message().c_str());
        SetConnInActive();
    }
    if (callback_) {
        callback_(resParser_->get().body(), ec, resParser_->get().result_int());
    }
    CheckResponseHeaderAndReset();
}

void AsyncHttpClient::OnWrite(const std::shared_ptr<std::string> requestId, beast::error_code &ec,
                              std::size_t bytesTransferred)
{
    boost::ignore_unused(bytesTransferred);
    if (!ec) {
        http::async_read(stream_, buf_, *resParser_,
                         beast::bind_front_handler(&AsyncHttpClient::OnRead, shared_from_this(), requestId));
        return;
    }
    YRLOG_ERROR("requestId {} failed to write, err message: {}, this client disconnect", *requestId,
                ec.message().c_str());
    if (!retried_) {
        YRLOG_DEBUG("requestId {} start to retry once", *requestId);
        retried_ = true;
        if (ReInit(requestId).OK()) {
            http::async_write(stream_, req_,
                              beast::bind_front_handler(&AsyncHttpClient::OnWrite, shared_from_this(), requestId));
            return;
        }
    }
    if (callback_) {
        callback_(HTTP_CONNECTION_ERROR_MSG, ec, HTTP_CONNECTION_ERROR_CODE);
    }
    SetConnInActive();
    SetAvailable();
    return;
}

void AsyncHttpClient::Stop()
{
    if (IsConnActive()) {
        GracefulExit();
    }
}
}  // namespace Libruntime
}  // namespace YR