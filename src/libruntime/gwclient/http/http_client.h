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

#include <functional>
#include <string>
#include <unordered_map>

#include <boost/asio/ssl.hpp>
#include <boost/beast/core.hpp>
#include <boost/beast/http.hpp>
#include "absl/synchronization/mutex.h"
#include "src/dto/config.h"
#include "src/libruntime/err_type.h"
namespace http = boost::beast::http;
namespace ssl = boost::asio::ssl;
namespace beast = boost::beast;
namespace asio = boost::asio;
namespace YR {
namespace Libruntime {
using HttpCallbackFunction = std::function<void(const std::string &, const boost::beast::error_code &, const uint)>;
const http::verb POST = http::verb::post;
const http::verb DELETE = http::verb::delete_;
const http::verb GET = http::verb::get;
const http::verb PUT = http::verb::put;
extern const int DEFAULT_HTTP_VERSION;
extern const uint HTTP_CONNECTION_ERROR_CODE;
const int CONNECTION_NO_TIMEOUT = -1;
extern const char *HTTP_CONNECTION_ERROR_MSG;
struct ConnectionParam {
    std::string ip;
    std::string port;
    int idleTime{120};
    int timeoutSec = CONNECTION_NO_TIMEOUT;
};

class HttpClient {
public:
    virtual ~HttpClient() = default;
    virtual ErrorInfo Init(const ConnectionParam &param) = 0;
    virtual void SubmitInvokeRequest(const http::verb &method, const std::string &target,
                                     const std::unordered_map<std::string, std::string> &headers,
                                     const std::string &body, const std::shared_ptr<std::string> requestId,
                                     const HttpCallbackFunction &receiver) = 0;

    virtual ErrorInfo ReInit(const std::shared_ptr<std::string> requestId)
    {
        GracefulExit();
        const int totalRetryCount = YR::Libruntime::Config::Instance().MAX_HTTP_RETRY_TIME();
        const int maxTimeoutSec = YR::Libruntime::Config::Instance().MAX_HTTP_TIMEOUT_SEC();
        int timeoutSec = YR::Libruntime::Config::Instance().INITIAL_HTTP_CONNECT_SEC();
        YRLOG_DEBUG("client start reinit. initTimeoutSec: {} maxTimeoutSec: {} totalRetryCount: {} requestId: {}",
            timeoutSec, maxTimeoutSec, totalRetryCount, *requestId);
        int retryCount = 0;
        int backoffFactor = 2;
        ErrorInfo err;
        while (retryCount < totalRetryCount) {
            connParam_.timeoutSec = timeoutSec;
            err = Init(connParam_);
            if (err.OK()) {
                YRLOG_DEBUG("client reinit success, requestId: {}", *requestId);
                connParam_.timeoutSec = CONNECTION_NO_TIMEOUT;
                return err;
            }
            retryCount++;
            if (timeoutSec != CONNECTION_NO_TIMEOUT) {
                timeoutSec = std::min(timeoutSec * backoffFactor, maxTimeoutSec);
            }
            YRLOG_DEBUG("retry count {}, requestId: {} init err: {}", retryCount, *requestId, err.Msg());
        }
        connParam_.timeoutSec = CONNECTION_NO_TIMEOUT;
        return err;
    }

    virtual void Stop() {}

    virtual void GracefulExit() noexcept {}

    bool SetUnavailable()
    {
        absl::WriterMutexLock l(&mu_);
        if (isUsed_) {
            return false;
        }
        isUsed_ = true;
        return true;
    }

    void SetAvailable()
    {
        absl::WriterMutexLock l(&mu_);
        isUsed_ = false;
    }

    void ResetConnActive()
    {
        absl::WriterMutexLock l(&mu_);
        lastActiveTime_ = std::chrono::high_resolution_clock::now();
        isConnectionAlive_ = true;
    }

    void ResetConnActiveTime()
    {
        absl::WriterMutexLock l(&mu_);
        lastActiveTime_ = std::chrono::high_resolution_clock::now();
    }

    void SetConnInActive()
    {
        absl::WriterMutexLock l(&mu_);
        isConnectionAlive_ = false;
    }

    bool Available() const
    {
        absl::ReaderMutexLock l(&mu_);
        return !this->isUsed_;
    };

    bool IsConnActive() const
    {
        absl::ReaderMutexLock l(&mu_);
        auto current = std::chrono::high_resolution_clock::now();
        auto idle = std::chrono::duration_cast<std::chrono::seconds>(current - this->lastActiveTime_).count();
        return isConnectionAlive_ && idle < idleTime_;
    };

    void CheckResponseHeaderAndReset()
    {
        auto headers = resParser_->get().base();
        if (auto it = headers.find("Connection"); it != headers.end() && it->value() == "close") {
            SetConnInActive();
        }
        resParser_.reset();
        buf_.clear();
        req_.clear();
        ResetConnActiveTime();
        SetAvailable();
    };

protected:
    ConnectionParam connParam_;
    HttpCallbackFunction callback_;
    beast::flat_buffer buf_;
    std::shared_ptr<http::response_parser<http::string_body>> resParser_;
    http::request<http::string_body> req_;
    bool isUsed_{true} ABSL_GUARDED_BY(mu_);
    bool isConnectionAlive_{false} ABSL_GUARDED_BY(mu_);
    std::chrono::time_point<std::chrono::high_resolution_clock> lastActiveTime_ ABSL_GUARDED_BY(mu_);
    bool retried_{false};
    int idleTime_{120};
    mutable absl::Mutex mu_;
};

inline bool IsResponseSuccessful(const uint statusCode)
{
    const uint SUCCESS_CODE_MIN = 200;
    const uint SUCCESS_CODE_MAX = 299;
    return (statusCode >= SUCCESS_CODE_MIN && statusCode <= SUCCESS_CODE_MAX);
}

inline bool IsResponseServerError(const uint statusCode)
{
    const uint SUCCESS_CODE_MIN = 500;
    const uint SUCCESS_CODE_MAX = 599;
    return (statusCode >= SUCCESS_CODE_MIN && statusCode <= SUCCESS_CODE_MAX);
}

inline bool IsResponseClientError(const uint statusCode)
{
    const uint SUCCESS_CODE_MIN = 400;
    const uint SUCCESS_CODE_MAX = 499;
    return (statusCode >= SUCCESS_CODE_MIN && statusCode <= SUCCESS_CODE_MAX);
}
}  // namespace Libruntime
}  // namespace YR