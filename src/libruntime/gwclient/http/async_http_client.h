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
#include <memory>
#include <mutex>

#include <boost/asio/strand.hpp>
#include <boost/beast/core.hpp>
#include <boost/beast/http.hpp>
#include <boost/beast/version.hpp>

#include "src/libruntime/gwclient/http/http_client.h"

namespace beast = boost::beast;
namespace http = boost::beast::http;
namespace asio = boost::asio;

namespace YR {
namespace Libruntime {
class AsyncHttpClient : public HttpClient, public std::enable_shared_from_this<AsyncHttpClient> {
public:
    explicit AsyncHttpClient(std::shared_ptr<asio::io_context> ctx);

    ~AsyncHttpClient() override;

    ErrorInfo Init(const ConnectionParam &param) override;
    void SubmitInvokeRequest(const http::verb &method, const std::string &target,
                             const std::unordered_map<std::string, std::string> &headers, const std::string &body,
                             const std::shared_ptr<std::string> requestId,
                             const HttpCallbackFunction &receiver) override;

    void OnRead(const std::shared_ptr<std::string> requestId, const beast::error_code &ec,
                std::size_t bytesTransferred);
    void OnWrite(const std::shared_ptr<std::string> requestId, beast::error_code &ec, std::size_t bytesTransferred);

    void Stop() override;

    void GracefulExit() noexcept override;

private:
    asio::ip::tcp::resolver resolver_;
    beast::tcp_stream stream_;
};
}  // namespace Libruntime
}  // namespace YR
