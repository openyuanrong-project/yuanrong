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

#include <algorithm>

#include <cstdlib>
#include <functional>
#include <iostream>
#include <memory>
#include <mutex>
#include <string>
#include <thread>
#include <vector>
#include <future>
#include "common_server.h"
const int SESSION_KEEP_ALIVE_S = 30;
template <class Body, class Allocator, class Send>
void HandleRequest(http::request<Body, http::basic_fields<Allocator>> &&req, Send &&send)
{
    std::cout << "session start to handle one request" << std::endl;

    std::string body = "ok";

    http::response<http::string_body> res{http::status::ok, req.version()};
    res.set(http::field::server, BOOST_BEAST_VERSION_STRING);
    res.set(http::field::content_type, "application/json");
    res.body() = body;
    res.prepare_payload();
    res.keep_alive(req.keep_alive());
    return send(std::move(res));
}

void Fail(beast::error_code ec, char const *what)
{
    std::cerr << what << ": " << ec.message() << "\n";
}
class Session {
public:
    Session() = default;
    virtual ~Session() = default;
    virtual void Run() = 0;
    virtual void DoClose() {}
    beast::flat_buffer buffer_;
    http::request<http::string_body> req_;
    std::shared_ptr<void> res_;
    std::atomic<bool> isClosed_{false};
};

class HttpsSession : public Session, public std::enable_shared_from_this<HttpsSession> {
    struct SendLambda {
        HttpsSession &self_;

        explicit SendLambda(HttpsSession &self) : self_(self) {}

        template <bool isRequest, class Body, class Fields>
        void operator()(http::message<isRequest, Body, Fields> &&msg) const
        {
            auto sp = std::make_shared<http::message<isRequest, Body, Fields>>(std::move(msg));
            self_.res_ = sp;
            http::async_write(
                self_.sslStream_, *sp,
                beast::bind_front_handler(&HttpsSession::OnWrite, self_.shared_from_this(), sp->need_eof()));
        }
    };
    SendLambda lambda_;
    beast::ssl_stream<beast::tcp_stream> sslStream_;

public:
    HttpsSession(tcp::socket &&socket, std::shared_ptr<ssl::context> ctx)
        : lambda_(*this), sslStream_(std::move(socket), *ctx)
    {
    }
    ~HttpsSession() override
    {
        if (!isClosed_) {
            DoClose();
        }
    }
    void Run() override
    {
        net::dispatch(sslStream_.get_executor(),
                      beast::bind_front_handler(&HttpsSession::OnSslRun, shared_from_this()));
    }

    void OnSslRun()
    {
        beast::get_lowest_layer(sslStream_).expires_after(std::chrono::seconds(SESSION_KEEP_ALIVE_S));
        sslStream_.async_handshake(ssl::stream_base::server,
                                   beast::bind_front_handler(&HttpsSession::OnHandshake, shared_from_this()));
    }

    void OnHandshake(beast::error_code ec)
    {
        if (ec) {
            return Fail(ec, "handshake");
        }
        DoRead();
    }

    void DoRead()
    {
        std::cout << "https session waiting to read request from connection" << std::endl;
        req_ = {};
        beast::get_lowest_layer(sslStream_).expires_after(std::chrono::seconds(SESSION_KEEP_ALIVE_S));
        http::async_read(sslStream_, buffer_, req_,
                         beast::bind_front_handler(&HttpsSession::OnRead, shared_from_this()));
    }

    void OnRead(beast::error_code ec, std::size_t bytes_transferred)
    {
        std::cout << "https session has read one request from connection" << std::endl;
        boost::ignore_unused(bytes_transferred);

        if (ec == http::error::end_of_stream) {
            return DoClose();
        }

        if (ec) {
            return Fail(ec, "read");
        }

        HandleRequest(std::move(req_), lambda_);
    }

    void OnWrite(bool close, beast::error_code ec, std::size_t bytes_transferred)
    {
        std::cout << "https session is writing one response" << std::endl;
        boost::ignore_unused(bytes_transferred);

        if (ec) {
            return Fail(ec, "write");
        }

        if (close || isClosed_) {
            std::cout << "close:" << close << ", isClosed: " << isClosed_ << std::endl;
            return DoClose();
        }

        res_ = nullptr;

        DoRead();
    }

    void DoClose() override
    {
        std::cout << "https session start to close the connection" << std::endl;
        sslStream_.async_shutdown(beast::bind_front_handler([](beast::error_code ec) {
            if (ec) {
                return Fail(ec, "shutdown");
            }
        }));
        isClosed_ = true;
    }
};

class HttpSession : public Session, public std::enable_shared_from_this<HttpSession> {
    struct SendLambda {
        HttpSession &self_;

        explicit SendLambda(HttpSession &self) : self_(self) {}

        template <bool isRequest, class Body, class Fields>
        void operator()(http::message<isRequest, Body, Fields> &&msg) const
        {
            auto sp = std::make_shared<http::message<isRequest, Body, Fields>>(std::move(msg));
            self_.res_ = sp;
            http::async_write(
                self_.stream_, *sp,
                beast::bind_front_handler(&HttpSession::OnWrite, self_.shared_from_this(), sp->need_eof()));
        }
    };

    beast::tcp_stream stream_;
    SendLambda lambda_;

public:
    HttpSession(tcp::socket &&socket) : lambda_(*this), stream_(std::move(socket)) {}
    ~HttpSession() override
    {
        if (!isClosed_) {
            DoClose();
        }
    }
    void Run() override
    {
        net::dispatch(stream_.get_executor(), beast::bind_front_handler(&HttpSession::DoRead, shared_from_this()));
    }

    void DoRead()
    {
        std::cout << "http session waiting to read request from connection" << std::endl;
        req_ = {};
        // 30s后服务端断链
        stream_.expires_after(std::chrono::seconds(SESSION_KEEP_ALIVE_S));
        http::async_read(stream_, buffer_, req_, beast::bind_front_handler(&HttpSession::OnRead, shared_from_this()));
    }

    void OnRead(beast::error_code ec, std::size_t bytes_transferred)
    {
        std::cout << "http session has read one request from connection" << std::endl;
        boost::ignore_unused(bytes_transferred);

        if (ec == http::error::end_of_stream) {
            return DoClose();
        }

        if (ec) {
            return Fail(ec, "read");
        }

        HandleRequest(std::move(req_), lambda_);
    }

    void OnWrite(bool close, beast::error_code ec, std::size_t bytes_transferred)
    {
        std::cout << "http session is writing one response" << std::endl;
        boost::ignore_unused(bytes_transferred);

        if (ec) {
            return Fail(ec, "write");
        }

        if (close) {
            return DoClose();
        }

        res_ = nullptr;

        DoRead();
    }

    void DoClose() override
    {
        std::cout << "http session start to close the connection" << std::endl;
        beast::error_code ec;
        stream_.socket().shutdown(tcp::socket::shutdown_send, ec);
        if (ec) {
            std::cerr << "send shutdown: " << ec.message() << std::endl;
        } else {
            std::cout << "send shutdown successfully." << std::endl;
        }
        stream_.socket().close(ec);
        if (ec) {
            std::cerr << "Error closing stream: " << ec.message() << std::endl;
        } else {
            std::cout << "Stream closed successfully." << std::endl;
        }
        isClosed_ = true;
    }
};

Listener::Listener(net::io_context &ioc, const tcp::endpoint &endpoint) : ioc_(ioc), acceptor_(net::make_strand(ioc))
{
    beast::error_code ec;
    acceptor_.open(endpoint.protocol(), ec);
    if (ec) {
        Fail(ec, "open");
        return;
    }

    acceptor_.set_option(net::socket_base::reuse_address(true), ec);
    if (ec) {
        Fail(ec, "set_option");
        return;
    }

    acceptor_.bind(endpoint, ec);
    if (ec) {
        Fail(ec, "bind");
        return;
    }

    acceptor_.listen(net::socket_base::max_listen_connections, ec);
    if (ec) {
        Fail(ec, "listen");
        return;
    }
}

void Listener::Run()
{
    std::cout << "listener start to run" << std::endl;
    DoAccept();
}

void Listener::Close()
{
    std::cout << "listener start to close" << std::endl;
    acceptor_.cancel();
    acceptor_.close();
    {
        std::lock_guard<std::mutex> lockGuard(mtx);
        for (auto &session : sessions) {
            session->DoClose();
        }
        sessions.clear();
    }
}

void Listener::SetSslContext(std::shared_ptr<ssl::context> ctx)
{
    ctx_ = ctx;
}

void Listener::DoAccept()
{
    acceptor_.async_accept(net::make_strand(ioc_), beast::bind_front_handler(&Listener::OnAccept, shared_from_this()));
}

void Listener::OnAccept(beast::error_code ec, tcp::socket socket)
{
    std::cout << "http server has accepted one connection" << std::endl;
    if (ec) {
        Fail(ec, "accept");
    } else {
        std::shared_ptr<Session> session;
        if (ctx_) {
            session = std::make_shared<HttpsSession>(std::move(socket), ctx_);
        } else {
            session = std::make_shared<HttpSession>(std::move(socket));
        }
        {
            std::lock_guard<std::mutex> lockGuard(mtx);
            sessions.emplace_back(session);
        }
        session->Run();
    }
    // Accept another connection
    DoAccept();
}

bool CommonServer::Start(const std::string &ip, unsigned short port, int threadNum, std::shared_ptr<ssl::context> ctx)
{
    const auto address = net::ip::make_address(ip);
    ioc_ = std::make_shared<net::io_context>(threadNum);
    listener_ = std::make_shared<Listener>(*ioc_, tcp::endpoint{address, port});
    listener_->SetSslContext(ctx);
    listener_->Run();
    threads_.reserve(threadNum);
    for (auto i = 0; i < threadNum; i++) {
        threads_.emplace_back([this] { this->ioc_->run(); });
        std::string name = "test_server_io_" + std::to_string(i);
        pthread_setname_np(threads_[i].native_handle(), name.c_str());
    }
    std::cout << "start to start http server" << std::endl;
    stopped_ = false;
    return true;
}

bool CommonServer::StopServer()
{
    std::promise<bool> waitPromise;
    auto waitFuture = waitPromise.get_future();
    waitFuture.wait_for(std::chrono::milliseconds(100));
    std::cout << "start to stop http server" << std::endl;
    if (stopped_) {
        return stopped_;
    }
    if (listener_) {
        listener_->Close();
        listener_.reset();
    }
    if (ioc_ && !ioc_->stopped()) {
        std::cout << "start to stop ioc" << std::endl;
        ioc_->stop();
    }
    std::cout << "start to wait thread join" << std::endl;
    for (auto &thread : threads_) {
        if (thread.joinable()) {
            thread.join();
        }
    }
    threads_.clear();
    stopped_ = true;
    return stopped_;
}