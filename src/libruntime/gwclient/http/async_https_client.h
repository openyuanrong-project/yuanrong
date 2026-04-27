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

#pragma once

#include <atomic>
#include <chrono>
#include <iostream>
#include <memory>
#include <mutex>

#include <boost/asio/ssl/error.hpp>
#include <boost/asio/ssl/stream.hpp>
#include <boost/asio/strand.hpp>
#include <boost/beast/core.hpp>
#include <boost/beast/http.hpp>
#include <boost/beast/ssl.hpp>
#include <boost/beast/version.hpp>

#include "src/libruntime/gwclient/http/http_client.h"

namespace beast = boost::beast;
namespace http = boost::beast::http;
namespace asio = boost::asio;
namespace ssl = boost::asio::ssl;

namespace YR {
namespace Libruntime {
class AsyncHttpsClient : public HttpClient, public std::enable_shared_from_this<AsyncHttpsClient> {
public:
    explicit AsyncHttpsClient(const std::shared_ptr<asio::io_context> &ioc,
                              const std::shared_ptr<asio::ssl::context> &ctx,
                              std::string serverName = "");

    ~AsyncHttpsClient() override;

    ErrorInfo Init(const ConnectionParam &param) override;

    void SubmitInvokeRequest(const http::verb &method, const std::string &target,
                             const std::unordered_map<std::string, std::string> &headers, const std::string &body,
                             const std::shared_ptr<std::string> requestId,
                             const HttpCallbackFunction &receiver) override;

    void OnRead(const std::shared_ptr<std::string> requestId, const beast::error_code &ec,
                std::size_t bytesTransferred);

    void OnWrite(const std::shared_ptr<std::string> requestId, const beast::error_code &ec,
                 std::size_t bytesTransferred);

    void Stop() override;

    void GracefulExit() noexcept override;

private:
    std::shared_ptr<asio::io_context> ioc_;
    std::shared_ptr<asio::ssl::context> ctx_;
    std::string serverName_;
    asio::ip::tcp::resolver resolver_;
    std::shared_ptr<beast::ssl_stream<beast::tcp_stream>> stream_;
};
}  // namespace Libruntime
}  // namespace YR
