/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
#include <boost/asio/dispatch.hpp>
#include <boost/asio/strand.hpp>
#include <boost/beast/core.hpp>
#include <boost/beast/http.hpp>
#include <boost/beast/ssl.hpp>
#include <boost/beast/version.hpp>
#include <boost/config.hpp>
#include <memory>
#include <string>
#include <thread>
namespace beast = boost::beast;
namespace http = beast::http;
namespace net = boost::asio;
namespace ssl = boost::asio::ssl;
using tcp = boost::asio::ip::tcp;

class Session;
class SslSession;
// Accepts incoming connections and launches the sessions
class Listener : public std::enable_shared_from_this<Listener> {
    net::io_context &ioc_;
    tcp::acceptor acceptor_;

public:
    Listener(net::io_context &ioc, const tcp::endpoint &endpoint);
    void Run();
    void Close();
    void SetSslContext(std::shared_ptr<ssl::context> ctx);

private:
    void DoAccept();

    void OnAccept(beast::error_code ec, tcp::socket socket);
    std::mutex mtx;
    std::vector<std::shared_ptr<Session>> sessions;
    std::shared_ptr<ssl::context> ctx_;
};

class CommonServer {
public:
    CommonServer() = default;
    ~CommonServer()
    {
        StopServer();
    };
    virtual bool Start(const std::string &ip, unsigned short port, int threadNum,
                             std::shared_ptr<ssl::context> ctx = nullptr);
    virtual bool StopServer();

private:
    std::shared_ptr<Listener> listener_;
    std::shared_ptr<net::io_context> ioc_;
    std::vector<std::thread> threads_;
    bool stopped_{true};
};