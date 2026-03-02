/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.
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

#include "fs_intf_impl.h"

#include "src/dto/config.h"
#include "src/dto/status.h"
#include "src/libruntime/traceadaptor/trace_adapter.h"
#include "src/libruntime/utils/utils.h"
#include "src/utility/logger/logger.h"
#include "src/utility/notification_utility.h"
#include "src/utility/string_utility.h"
namespace YR {

namespace Libruntime {
using namespace std::placeholders;
using YR::utility::NotificationUtility;
const std::string FUNCTION_PROXY = "function-proxy";
const int DOUBLE_INTERVAL = 2;

static const StreamingMessage CALL_RESULT_ACK = []() {
    StreamingMessage fake;
    (void)*fake.mutable_callresultack();
    return fake;
}();

static const StreamingMessage INVOKE_RESPONSE = []() {
    StreamingMessage fake;
    (void)*fake.mutable_invokersp();
    return fake;
}();

FSIntfImpl::FSIntfImpl(const std::string &ipAddr, int port, FSIntfHandlers handlers, bool isDriver,
                       std::shared_ptr<Security> sec, std::shared_ptr<ClientsManager> clientsMgr, bool enableClientMode,
                       bool enableEvent)
    : FSIntf(handlers),
      isDriver(isDriver),
      enableClientMode(enableClientMode),
      security(sec),
      clientsMgr(clientsMgr),
      fsInrfMgr(std::make_shared<FSIntfManager>(clientsMgr))
{
    enableDirectCall = Config::Instance().RUNTIME_DIRECT_CONNECTION_ENABLE();
    this->enableEvent = enableEvent;
    if (isDriver) {
        fsPort = port;
        selfPort = 0;
        fsIp = ipAddr;
        listeningIpAddr = ipAddr;
    } else if (enableClientMode) {
        // client mode the input address is runtime address.
        fsIp = ParseIpAddr(Config::Instance().YR_SERVER_ADDRESS()).ip;
        fsPort = port;
        // client mode on cloud no need to start posix server, unless direct call mode enable which
        // POSIX_LISTEN_ADDR is required
        listeningIpAddr = "";
        selfPort = 0;
    } else {
        // server mode on cloud no need to connect functionsystem
        // start service with same ip:port with posix server on direct call mode
        fsIp = "";
        fsPort = 0;
        listeningIpAddr = ipAddr;
        selfPort = port;
    }
    timerWorker = std::make_shared<TimerWorker>();
    fsMsgHdlrs = {{StreamingMessage::kCallReq, std::bind(&FSIntfImpl::RecvCallRequest, this, _1, _2)},
                  {StreamingMessage::kNotifyReq, std::bind(&FSIntfImpl::RecvNotifyRequest, this, _1, _2)},
                  {StreamingMessage::kCheckpointReq, std::bind(&FSIntfImpl::RecvCheckpointRequest, this, _1, _2)},
                  {StreamingMessage::kRecoverReq, std::bind(&FSIntfImpl::RecvRecoverRequest, this, _1, _2)},
                  {StreamingMessage::kShutdownReq, std::bind(&FSIntfImpl::RecvShutdownRequest, this, _1, _2)},
                  {StreamingMessage::kSignalReq, std::bind(&FSIntfImpl::RecvSignalRequest, this, _1, _2)},
                  {StreamingMessage::kHeartbeatReq, std::bind(&FSIntfImpl::RecvHeartbeatRequest, this, _1, _2)},
                  {StreamingMessage::kCreateRsp, std::bind(&FSIntfImpl::RecvCreateOrInvokeResponse, this, _1, _2)},
                  {StreamingMessage::kInvokeRsp, std::bind(&FSIntfImpl::RecvCreateOrInvokeResponse, this, _1, _2)},
                  {StreamingMessage::kCallResultAck, std::bind(&FSIntfImpl::RecvResponse, this, _1, _2)},
                  {StreamingMessage::kKillRsp, std::bind(&FSIntfImpl::RecvResponse, this, _1, _2)},
                  {StreamingMessage::kSaveRsp, std::bind(&FSIntfImpl::RecvResponse, this, _1, _2)},
                  {StreamingMessage::kLoadRsp, std::bind(&FSIntfImpl::RecvResponse, this, _1, _2)},
                  {StreamingMessage::kRGroupRsp, std::bind(&FSIntfImpl::RecvResponse, this, _1, _2)},
                  {StreamingMessage::kExitRsp, std::bind(&FSIntfImpl::RecvResponse, this, _1, _2)},
                  {StreamingMessage::kCreateRsps, std::bind(&FSIntfImpl::RecvCreateOrInvokeResponse, this, _1, _2)}};

    rtMsgHdlrs = {
        {StreamingMessage::kCallReq, std::bind(&FSIntfImpl::RecvCallRequest, this, _1, _2)},
        {StreamingMessage::kInvokeRsp, std::bind(&FSIntfImpl::RecvCreateOrInvokeResponse, this, _1, _2)},
        {StreamingMessage::kNotifyReq, std::bind(&FSIntfImpl::RecvNotifyRequest, this, _1, _2)},
        {StreamingMessage::kCallResultAck, std::bind(&FSIntfImpl::RecvResponse, this, _1, _2)}
    };
    eventMsgHdlrs = {
        {StreamingMessage::kEventReq, std::bind(&FSIntfImpl::RecvEventRequest, this, _1, _2)}
    };
}

std::shared_ptr<WiredRequest> FSIntfImpl::SaveWiredRequest(const std::string &reqId, std::shared_ptr<WiredRequest> wr)
{
    absl::MutexLock lock(&this->mu);
    auto [it, hanppened] = wiredRequests.emplace(reqId, wr);
    YRLOG_DEBUG("saved callback of req id {}, callback is exsited: {}", reqId, hanppened);
    if (!hanppened) {
        it->second->retryCount++;
    }
    return it->second;
}

std::shared_ptr<WiredRequest> FSIntfImpl::EraseWiredRequest(const std::string &reqId)
{
    absl::MutexLock lock(&this->mu);
    auto it = wiredRequests.find(reqId);
    if (it == wiredRequests.end()) {
        YRLOG_DEBUG("there is no wired request belong reqid : {}", reqId);
        return nullptr;
    }
    auto wr = it->second;
    wr->CancelAllTimer();
    wiredRequests.erase(it);
    return wr;
}

std::shared_ptr<WiredRequest> FSIntfImpl::GetWiredRequest(const std::string &reqId, bool ackReceived)
{
    absl::MutexLock lock(&this->mu);
    auto it = wiredRequests.find(reqId);
    if (it == wiredRequests.end()) {
        return nullptr;
    }
    it->second->ackReceived = ackReceived;
    return it->second;
}

void FSIntfImpl::UpdateWiredRequestRemote(const std::string &reqId, const std::string &dstInstanceID)
{
    absl::MutexLock lock(&this->mu);
    auto it = wiredRequests.find(reqId);
    if (it == wiredRequests.end()) {
        return;
    }
    it->second->dstInstanceID = dstInstanceID;
}

std::unordered_map<std::string, std::shared_ptr<WiredRequest>> FSIntfImpl::GetAllWiredRequests()
{
    absl::MutexLock lock(&this->mu);
    return wiredRequests;
}

// return bool:expired, don't retry any more.
std::pair<std::shared_ptr<WiredRequest>, bool> FSIntfImpl::UpdateRetryInterval(const std::string &reqId)
{
    absl::MutexLock lock(&this->mu);
    auto it = wiredRequests.find(reqId);
    if (it == wiredRequests.end()) {
        return std::make_pair<std::shared_ptr<WiredRequest>, bool>(nullptr, true);
    }
    auto wr = it->second;
    ++wr->retryCount;
    wr->remainTimeoutSec -= wr->retryIntervalSec;
    // Current is not need to retry, because there is no time to wait response.
    if (wr->remainTimeoutSec <= 0 && !wr->ackReceived) {
        wiredRequests.erase(it);
        return std::make_pair(wr, true);
    }
    if (wr->exponentialBackoff) {
        wr->retryIntervalSec *= DOUBLE_INTERVAL;  // double interval
    }
    if (wr->retryIntervalSec > wr->remainTimeoutSec) {
        wr->retryIntervalSec = wr->remainTimeoutSec;
    }
    return std::make_pair(wr, false);
}

void FSIntfImpl::ClearAllWiredRequests(void)
{
    absl::MutexLock lock(&this->mu);
    for (auto &e : this->wiredRequests) {
        auto wr = e.second;
        if (wr->timer_ != nullptr) {
            wr->timer_->cancel();
        }
        if (wr->callback != nullptr) {
            StreamingMessage fake;
            wr->callback(fake, ErrorInfo(ErrorCode::ERR_FINALIZED, "Function system client quit"), [](bool) {});
        }
    }
    this->wiredRequests.clear();
}

bool FSIntfImpl::IsCommunicationError(const ::common::ErrorCode &code)
{
    if (code == ::common::ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_BUS ||
        code == ::common::ErrorCode::ERR_INNER_COMMUNICATION) {
        return true;
    }
    return false;
}

bool FSIntfImpl::NeedRepeat(const std::string requestId)
{
    auto [wr, expired] = UpdateRetryInterval(requestId);
    if (expired) {
        if (wr != nullptr && wr->callback != nullptr) {
            YRLOG_ERROR("RPC request retry expired, request ID: {}, is ack received: {}", requestId,
                        wr->ackReceived.load());
            ErrorInfo err(ERR_REQUEST_BETWEEN_RUNTIME_BUS, "Response timeout, request ID is " + requestId);
            err.SetIsAckTimeout(true);
            StreamingMessage fake;
            wr->callback(fake, err, [this, requestId](bool needEraseWiredReq) {
                if (needEraseWiredReq) {
                    this->EraseWiredRequest(requestId);
                }
            });
        }
        return false;
    }
    if (wr && wr->ackReceived) {
        YRLOG_DEBUG(" {} has received ack, no need retry", requestId);
        return false;
    }
    return true;
}

void FSIntfImpl::WriteCallback(const std::string requestId, const ErrorInfo &err)
{
    if (err.OK()) {
        return;
    }

    if (IsCommunicationError(::common::ErrorCode(err.Code()))) {
        YRLOG_ERROR("Communicate fails for request({}) errcode({}), msg({})", requestId, fmt::underlying(err.Code()),
                    err.Msg());
        return;
    }
    YRLOG_DEBUG("send grpc request failed for request: {}, err code is {}, err msg is {}", requestId,
                fmt::underlying(err.Code()), err.Msg());
    auto wr = EraseWiredRequest(requestId);
    if (wr != nullptr && wr->callback != nullptr) {
        StreamingMessage fakeMsg;
        wr->callback(fakeMsg, err, [](bool) {});
    }
}

void FSIntfImpl::GroupCreateAsync(const CreateRequests &reqs, CreateRespsCallback createRespCallback,
                                  CreateCallBack callback, int timeoutSec)
{
    auto reqId = reqs.requestid();
    auto tenantId = reqs.tenantid();
    auto traceId = reqs.traceid();

    auto respCallback = [reqId, tenantId, traceId, createRespCallback](const StreamingMessage &createResps,
                                                                       ErrorInfo status,
                                                                       std::function<void(bool)> needEraseWiredReq) {
        YRLOG_DEBUG("Receive group create responses, request ID:{}, tenant id {}, trace ID:{}", reqId, tenantId,
                    traceId);
        if (status.OK() && createResps.has_creatersps()) {
            if (createResps.creatersps().code() == common::ERR_NONE) {
                createRespCallback(createResps.creatersps());
                needEraseWiredReq(false);
                return;
            } else {
                createRespCallback(createResps.creatersps());
                needEraseWiredReq(true);
                return;
            }
        }
        CreateResponses rsps;
        rsps.set_code(common::ErrorCode(status.Code()));
        rsps.set_message("create group response failed, request id: " + reqId + ", msg: " + status.Msg());
        createRespCallback(rsps);
        needEraseWiredReq(true);
        return;
    };
    auto notifyCb = [callback](const NotifyRequest &req, const ErrorInfo &err) {
        YRLOG_DEBUG("Receive group create notify request, request ID:{}, code: {}, message: {}", req.requestid(),
                    fmt::underlying(req.code()), req.message());
        callback(req);
    };
    auto wr = std::make_shared<WiredRequest>(respCallback, notifyCb, timerWorker);
    wr->SetRequestID(reqId);
    wr = SaveWiredRequest(reqId, wr);
    std::weak_ptr<WiredRequest> weakWr(wr);
    auto sendMsgHandler = [this, reqs, weakWr]() {
        if (auto thisPtr = weakWr.lock(); thisPtr) {
            YRLOG_DEBUG("Begin to send group create instance request, request ID: {}", reqs.requestid());
            auto messageId =
                YR::utility::IDGenerator::GenMessageId(reqs.requestid(), static_cast<uint8_t>(thisPtr->retryCount));
            this->Write(GenStreamMsg(messageId, reqs),
                        std::bind(&FSIntfImpl::WriteCallback, this, reqs.requestid(), _1));
        }
    };

    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, reqId));
}

/**
 * @brief create a function instance asynchronous, but synchronous return a instance ID.
 *
 * @param req create request.
 * @param callback create callback, called when notify request received. if @return not sucess, it will not be called at
 * all.
 * @param instanceId instance ID, generated by function system.
 * @return ErrorInfo error code and message.
 */

void FSIntfImpl::CreateAsync(const CreateRequest &req, CreateRespCallback createRespCallback, CreateCallBack callback,
                             int timeoutSec)
{
    auto reqId = std::make_shared<std::string>(req.requestid());
    auto funcName = req.function();
    auto traceId = std::make_shared<std::string>(req.traceid());
    auto designatedInstanceID = req.designatedinstanceid();
    auto span = TraceAdapter::GetInstance().StartSpan(
        "Create", {{"requestID", *reqId}, {"funcName", funcName}, {"designatedInstanceID", designatedInstanceID}});
    auto respCallback = [this, reqId, funcName, traceId, createRespCallback, span](
                            const StreamingMessage &createResp, ErrorInfo status,
                            std::function<void(bool)> needEraseWiredReq) {
        YRLOG_DEBUG("Receive create response, function: {}, request ID:{}, trace ID:{}", funcName, *reqId, *traceId);
        if (createResp.has_creatersp()) {
            span->SetAttribute("respCode", createResp.creatersp().code());
            span->SetAttribute("respMessage", createResp.creatersp().message());
            span->SetAttribute("respInstanceID", createResp.creatersp().instanceid());
        }
        if (status.OK() && createResp.has_creatersp()) {
            if (createResp.creatersp().code() == common::ERR_NONE) {
                createRespCallback(createResp.creatersp());
                UpdateWiredRequestRemote(*reqId, createResp.creatersp().instanceid());
                needEraseWiredReq(false);
                return;
            } else {
                createRespCallback(createResp.creatersp());
                needEraseWiredReq(true);
                return;
            }
        }
        CreateResponse rsp;
        rsp.set_code(common::ErrorCode(status.Code()));
        rsp.set_message("create response failed, request id: " + (*reqId) + ", msg: " + status.Msg());
        createRespCallback(rsp);
        needEraseWiredReq(true);
        return;
    };

    auto notifyCallback = [callback, span](const NotifyRequest &req, const ErrorInfo &err) {
        YRLOG_DEBUG("Receive create notify request, request ID:{}, code: {}, message: {}", req.requestid(),
                    fmt::underlying(req.code()), req.message());
        span->SetAttribute("notifyCode", req.code());
        span->SetAttribute("notifyMessage", req.message());
        span->End();
        callback(req);
    };

    auto wr = std::make_shared<WiredRequest>(respCallback, notifyCallback, timerWorker);
    wr->SetRequestID(*reqId);
    wr = SaveWiredRequest(*reqId, wr);
    std::weak_ptr<WiredRequest> weakWr(wr);
    auto sendMsgHandler = [this, req, weakWr]() {
        if (auto thisPtr = weakWr.lock(); thisPtr) {
            YRLOG_DEBUG("Begin to send create instance request, request ID: {}", req.requestid());
            auto messageId =
                YR::utility::IDGenerator::GenMessageId(req.requestid(), static_cast<uint8_t>(thisPtr->retryCount));
            this->Write(GenStreamMsg(messageId, req), std::bind(&FSIntfImpl::WriteCallback, this, req.requestid(), _1));
        }
    };

    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, *reqId));
    if (timeoutSec > 0) {
        wr->SetupTimeout(timeoutSec, [this, reqId, traceId]() {
            NotifyRequest notifyRequest;
            notifyRequest.set_code(common::ErrorCode(ERR_INNER_SYSTEM_ERROR));
            notifyRequest.set_message("create request timeout, requestId: " + (*reqId));
            notifyRequest.set_requestid(*reqId);
            auto wiredReq = GetWiredRequest(*reqId, false);
            if (wiredReq != nullptr) {
                YRLOG_ERROR("Request timeout, start exec notify callback, request ID:{}, trace ID:{}", *reqId,
                            *traceId);
                if (wiredReq->notifyCallback) {
                    wiredReq->notifyCallback(notifyRequest, ErrorInfo());
                }
                EraseWiredRequest(*reqId);
            }
        });
    }
}

void FSIntfImpl::InvokeAsync(const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback, int timeoutSec)
{
    auto reqId = std::make_shared<std::string>(req->Immutable().requestid());
    auto instanceId = std::make_shared<std::string>(req->Immutable().instanceid());
    auto traceId = std::make_shared<std::string>(req->Immutable().traceid());
    auto funcName = std::make_shared<std::string>(req->Immutable().function());
    auto span = TraceAdapter::GetInstance().StartSpan(
        "Invoke", {{"requestID", *reqId}, {"funcName", *funcName}, {"instanceId", *instanceId}});
    auto respCallback = [this, callback, reqId, instanceId, traceId, span](
                            const StreamingMessage &invokeResp, ErrorInfo status,
                            std::function<void(bool)> needEraseWiredReq) {
        YRLOG_DEBUG("Receive invoke response, instance: {}, request ID:{}, trace ID:{}", *instanceId, *reqId, *traceId);
        if (invokeResp.has_invokersp()) {
            span->SetAttribute("respCode", invokeResp.invokersp().code());
            span->SetAttribute("respMessage", invokeResp.invokersp().message());
        }
        if (status.OK() && invokeResp.has_invokersp()) {
            if (invokeResp.invokersp().code() == common::ERR_NONE) {
                needEraseWiredReq(false);
                return;
            } else {
                status.SetErrCodeAndMsg(static_cast<ErrorCode>(invokeResp.invokersp().code()), ModuleCode::CORE,
                                        invokeResp.invokersp().message());
            }
        }

        NotifyRequest notifyRequest;
        notifyRequest.set_code(common::ErrorCode(status.Code()));
        notifyRequest.set_message("invoke response failed, request id: " + (*reqId) + ", msg: " + status.Msg());
        notifyRequest.set_requestid(*reqId);
        YRLOG_ERROR("Receive invoke response, instance: {}, request ID:{}, trace ID:{}, code: {}, message: {}",
                    *instanceId, *reqId, *traceId, fmt::underlying(status.Code()), status.Msg());
        needEraseWiredReq(true);

        this->HandleNotifyRequest(
            notifyRequest,
            [callback, status, notifyRequest]() -> NotifyResponse {
                callback(notifyRequest, status);
                return NotifyResponse();
            },
            [](const NotifyResponse &resp) -> void { return; });
        return;
    };
    auto notifyCallback = [callback, span](const NotifyRequest &req, const ErrorInfo &err) {
        YRLOG_DEBUG("Receive invoke notify request, request ID:{}, code: {}", req.requestid(),
                    fmt::underlying(req.code()));
        span->SetAttribute("notifyCode", req.code());
        span->SetAttribute("notifyMessage", req.message());
        span->End();
        callback(req, err);
    };
    auto wr = std::make_shared<WiredRequest>(respCallback, notifyCallback, timerWorker, req->Immutable().instanceid());
    wr->returnObjectsSize = req->Immutable().returnobjectids_size();
    wr->SetRequestID(*reqId);
    wr = SaveWiredRequest(req->Immutable().requestid(), wr);
    std::weak_ptr<WiredRequest> weakWr(wr);
    std::shared_ptr<FSIntfImpl> self;
    try {
        self = shared_from_this();
    } catch (const std::exception &e) {
        YRLOG_ERROR("FSIntfImpl has been destructed");
    }
    auto sendMsgHandler = [self, reqId, req, weakWr]() {
        if (auto wr = weakWr.lock(); wr && self) {
            auto messageId = YR::utility::IDGenerator::GenMessageId(req->Immutable().requestid(),
                                                                    static_cast<uint8_t>(wr->retryCount));
            YRLOG_DEBUG("Send invoke message, message id {}", messageId);
            req->SetMessageID(messageId);
            self->TryDirectWrite(req->Immutable().instanceid(), req->Get(),
                                 [self, reqId, weakWr](bool isDirect, ErrorInfo status) {
                                     if (!isDirect || !status.OK()) {
                                         return self->WriteCallback(*reqId, status);
                                     }
                                     if (auto wr = weakWr.lock(); wr) {
                                         if (wr->callback) {
                                             wr->ackReceived = true;
                                             wr->callback(INVOKE_RESPONSE, status, [](bool) {});
                                         }
                                     }
                                 });
        }
    };
    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, *reqId), true);
    if (timeoutSec > 0) {
        wr->SetupTimeout(timeoutSec, [this, reqId, instanceId, traceId, timeoutSec]() {
            NotifyRequest notifyRequest;
            notifyRequest.set_code(common::ErrorCode(ERR_INNER_SYSTEM_ERROR));
            notifyRequest.set_message("invoke request timeout with " + std::to_string(timeoutSec) +
                                      " s, requestId: " + (*reqId));
            notifyRequest.set_requestid(*reqId);
            auto wiredReq = GetWiredRequest(*reqId, false);
            if (wiredReq != nullptr) {
                YRLOG_ERROR("Request timeout with {} s, instance: {}, request ID:{}, trace ID:{}", timeoutSec,
                            *instanceId, *reqId, *traceId);
                ErrorInfo err;
                err.SetIsTimeout(true);
                if (wiredReq->notifyCallback) {
                    wiredReq->notifyCallback(notifyRequest, err);
                }
                EraseWiredRequest(*reqId);
            }
        });
    }
}

void FSIntfImpl::CallResultAsync(const std::shared_ptr<CallResultMessageSpec> req, CallResultCallBack callback)
{
    auto reqId = std::make_shared<std::string>(
        std::move(YR::utility::IDGenerator::GetRealRequestId(req->Immutable().requestid())));
    YRLOG_DEBUG("Start Call Result Request, requestid: {}, source instanceid: {}", req->Immutable().requestid(),
                req->Immutable().instanceid());
    auto respCallback = [callback, req](const StreamingMessage &callResultResp, ErrorInfo status,
                                        std::function<void(bool)> needEraseWiredReq) {
        YRLOG_DEBUG("Receive call result ack, instance: {}, request ID:{}", req->Immutable().instanceid(),
                    req->Immutable().requestid());
        if (status.OK() && callResultResp.has_callresultack()) {
            needEraseWiredReq(true);
            callback(callResultResp.callresultack());
            return;
        }
        CallResultAck resp;
        resp.set_code(common::ErrorCode(status.Code()));
        resp.set_message(status.Msg());
        YRLOG_DEBUG("Receive call result ack, instance: {}, request ID:{}, code: {}, message: {}",
                    req->Immutable().instanceid(), req->Immutable().requestid(), fmt::underlying(status.Code()),
                    status.Msg());
        needEraseWiredReq(true);
        callback(resp);
        return;
    };

    auto wr = std::make_shared<WiredRequest>(respCallback, timerWorker);
    wr->SetRequestID(*reqId);
    bool existObjInDs = req->existObjInDs;
    std::weak_ptr<FSIntfImpl> weakSelf(shared_from_this());
    std::weak_ptr<WiredRequest> weakWr(wr);
    auto flag = std::make_shared<std::once_flag>();
    auto preWrite = [weakSelf, weakWr, reqId, existObjInDs, flag](bool isDirect) {
        if (isDirect && !existObjInDs) {
            return;
        }
        auto self = weakSelf.lock();
        auto wr = weakWr.lock();
        if (self == nullptr || wr == nullptr) {
            return;
        }
        std::call_once(*flag, [reqId, wr, self]() {
            (void)self->SaveWiredRequest(*reqId, wr);
        });
    };
    auto sendMsgHandler = [weakSelf, req, weakWr, reqId, preWrite, existObjInDs]() {
        auto self = weakSelf.lock();
        auto wr = weakWr.lock();
        if (self == nullptr || wr == nullptr) {
            return;
        }
        req->SetMessageID(YR::utility::IDGenerator::GenMessageId(*reqId, static_cast<uint8_t>(wr->retryCount)));
        if (self->enableDirectCall) {
            req->UpdateRuntimeInfo(self->listeningIpAddr, self->selfPort);
        }
        self->TryDirectWrite(
            req->Immutable().instanceid(), req->Get(),
            [self, reqId, wr, existObjInDs](bool isDirect, ErrorInfo status) {
                if (!isDirect || existObjInDs) {
                    return self->WriteCallback(*reqId, status);
                }
                if (self->IsCommunicationError(::common::ErrorCode(status.Code()))) {
                    (void)self->SaveWiredRequest(*reqId, wr);
                    YRLOG_ERROR("Communicate fails for request({}) errcode({}), msg({})", *reqId,
                                fmt::underlying(status.Code()), status.Msg());
                    return;
                }
                YRLOG_DEBUG_IF(!status.OK(), "send grpc call result failed for {}, err code is {}, err msg is {}",
                               *reqId, fmt::underlying(status.Code()), status.Msg());
                (void)self->EraseWiredRequest(*reqId);
                if (wr->callback != nullptr) {
                    wr->callback(CALL_RESULT_ACK, status, [](bool) {});
                }
            },
            preWrite);
    };
    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, *reqId));
}

void FSIntfImpl::KillAsync(const KillRequest &req, KillCallBack callback, int timeoutSec)
{
    std::string reqId = req.requestid();
    if (reqId.empty()) {
        reqId = YR::utility::IDGenerator::GenRequestId();
    }
    auto span = TraceAdapter::GetInstance().StartSpan(
        "Kill", {{"requestID", reqId}, {"instanceID", req.instanceid()}, {"signal", req.signal()}});
    auto respCallback = [callback, reqId, span](const StreamingMessage &killResp, ErrorInfo status,
                                                std::function<void(bool)> needEraseWiredReq) {
        YRLOG_DEBUG("Receive kill response, request ID:{}", reqId);
        if (killResp.has_killrsp()) {
            span->SetAttribute("respCode", killResp.killrsp().code());
            span->SetAttribute("respMessage", killResp.killrsp().message());
        }
        span->End();
        if (status.OK() && killResp.has_killrsp()) {
            callback(killResp.killrsp(), status);
            needEraseWiredReq(true);
            return;
        }

        KillResponse resp;
        resp.set_code(common::ErrorCode(status.Code()));
        resp.set_message(status.Msg());
        YRLOG_DEBUG("Receive kill response, request ID:{}, code: {}, msg: {}", reqId, fmt::underlying(status.Code()),
                    status.Msg());
        callback(resp, status);
        needEraseWiredReq(true);
        return;
    };

    auto wr = std::make_shared<WiredRequest>(respCallback, timerWorker);
    wr->SetRequestID(reqId);
    wr = SaveWiredRequest(reqId, wr);
    std::weak_ptr<WiredRequest> weak(wr);
    auto sendMsgHandler = [this, req, reqId, weak]() {
        if (auto thisPtr = weak.lock(); thisPtr) {
            auto msgId = YR::utility::IDGenerator::GenMessageId(reqId, static_cast<uint8_t>(thisPtr->retryCount));
            this->Write(GenStreamMsg(msgId, req), std::bind(&FSIntfImpl::WriteCallback, this, reqId, _1));
        }
    };

    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, reqId));
    if (timeoutSec > 0) {
        wr->SetupTimeout(timeoutSec, [this, reqId]() {
            auto wiredReq = GetWiredRequest(reqId, false);
            if (wiredReq != nullptr) {
                YRLOG_ERROR("Request timeout, start exec notify callback, request ID : {}", reqId);
                StreamingMessage fake;
                ErrorInfo errorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::CORE,
                                    "kill request timeout, requestId: " + reqId);
                errorInfo.SetIsTimeout(true);
                if (wiredReq->callback != nullptr) {
                    wiredReq->callback(fake, errorInfo, [](bool needEraseWiredReq) {});
                }
                EraseWiredRequest(reqId);
            }
        });
    }
}

void FSIntfImpl::ExitAsync(const ExitRequest &req, ExitCallBack callback)
{
    auto reqId = YR::utility::IDGenerator::GenRequestId();
    auto respCallback = [callback, reqId](const StreamingMessage &exitResp, ErrorInfo status,
                                          std::function<void(bool)> needEraseWiredReq) {
        YRLOG_DEBUG("Receive exit response, request ID:{}", reqId);
        if (status.OK() && exitResp.has_exitrsp()) {
            needEraseWiredReq(true);
            callback(exitResp.exitrsp());
            return;
        }

        ExitResponse resp;
        YRLOG_DEBUG("Receive exit response, request ID:{}, code: {}, message: {}", reqId,
                    fmt::underlying(status.Code()), status.Msg());
        needEraseWiredReq(true);
        callback(resp);
        return;
    };

    auto wr = std::make_shared<WiredRequest>(respCallback, timerWorker);
    wr->SetRequestID(reqId);
    wr = SaveWiredRequest(reqId, wr);
    std::weak_ptr<WiredRequest> weakReq(wr);
    auto sendMsgHandler = [this, req, reqId, weakReq]() {
        if (auto thisPtr = weakReq.lock(); thisPtr) {
            auto messageId = YR::utility::IDGenerator::GenMessageId(reqId, static_cast<uint8_t>(thisPtr->retryCount));
            this->Write(GenStreamMsg(messageId, req), std::bind(&FSIntfImpl::WriteCallback, this, reqId, _1));
        }
    };

    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, reqId));
}

void FSIntfImpl::StateSaveAsync(const StateSaveRequest &req, StateSaveCallBack callback)
{
    auto reqId = YR::utility::IDGenerator::GenRequestId();
    auto respCallback = [callback, reqId](const StreamingMessage &saveResp, ErrorInfo status,
                                          std::function<void(bool)> needEraseWiredReq) {
        YRLOG_DEBUG("Receive save response, request ID:{}", reqId);
        if (status.OK() && saveResp.has_saversp()) {
            callback(saveResp.saversp());
            needEraseWiredReq(true);
            return;
        }

        StateSaveResponse resp;
        resp.set_code(common::ErrorCode(status.Code()));
        resp.set_message(status.Msg());
        YRLOG_DEBUG("Receive save response, request ID:{}, code: {}, message: {}", reqId,
                    fmt::underlying(status.Code()), status.Msg());
        callback(resp);
        needEraseWiredReq(true);
        return;
    };

    auto wr = std::make_shared<WiredRequest>(respCallback, timerWorker);
    wr->SetRequestID(reqId);
    wr = SaveWiredRequest(reqId, wr);
    std::weak_ptr<WiredRequest> weakWr(wr);
    auto sendMsgHandler = [this, req, reqId, weakWr]() {
        if (auto thisPtr = weakWr.lock(); thisPtr) {
            auto messageId = YR::utility::IDGenerator::GenMessageId(reqId, static_cast<uint8_t>(thisPtr->retryCount));
            this->Write(GenStreamMsg(messageId, req), std::bind(&FSIntfImpl::WriteCallback, this, reqId, _1));
        }
    };

    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, reqId));
}

void FSIntfImpl::StateLoadAsync(const StateLoadRequest &req, StateLoadCallBack callback)
{
    auto reqId = YR::utility::IDGenerator::GenRequestId();
    auto respCallback = [callback, reqId](const StreamingMessage &loadResp, ErrorInfo status,
                                          std::function<void(bool)> needEraseWiredReq) {
        YRLOG_DEBUG("Receive load response, request ID:{}", reqId);
        if (status.OK() && loadResp.has_loadrsp()) {
            callback(loadResp.loadrsp());
            needEraseWiredReq(true);
            return;
        }

        StateLoadResponse resp;
        resp.set_code(common::ErrorCode(status.Code()));
        resp.set_message(status.Msg());
        YRLOG_DEBUG("Receive load response, request ID:{}, code: {}, message: {}", reqId,
                    fmt::underlying(status.Code()), status.Msg());
        callback(resp);
        needEraseWiredReq(true);
        return;
    };

    auto wr = std::make_shared<WiredRequest>(respCallback, timerWorker);
    wr->SetRequestID(reqId);
    wr = SaveWiredRequest(reqId, wr);
    std::weak_ptr<WiredRequest> weakPtr(wr);
    auto sendMsgHandler = [this, req, reqId, weakPtr]() {
        if (auto thisPtr = weakPtr.lock(); thisPtr) {
            auto messageId = YR::utility::IDGenerator::GenMessageId(reqId, static_cast<uint8_t>(thisPtr->retryCount));
            this->Write(GenStreamMsg(messageId, req), std::bind(&FSIntfImpl::WriteCallback, this, reqId, _1));
        }
    };

    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, reqId));
}

void FSIntfImpl::CreateRGroupAsync(const CreateResourceGroupRequest &req, CreateResourceGroupCallBack callback,
                                   int timeoutSec)
{
    auto reqId = req.requestid();
    auto respCallback = [callback, reqId](const StreamingMessage &createRGroupResp, ErrorInfo status,
                                          std::function<void(bool)> needEraseWiredReq) {
        YRLOG_DEBUG("Receive create resource group response, request ID:{}", reqId);
        if (status.OK() && createRGroupResp.has_rgrouprsp()) {
            callback(createRGroupResp.rgrouprsp());
            needEraseWiredReq(true);
            return;
        }

        CreateResourceGroupResponse resp;
        resp.set_code(common::ErrorCode(status.Code()));
        resp.set_message(status.Msg());
        YRLOG_DEBUG("Receive create resource group response, request ID:{}, error code: {}, error message: {}", reqId,
                    fmt::underlying(status.Code()), status.Msg());
        callback(resp);
        needEraseWiredReq(true);
        return;
    };

    auto wr = std::make_shared<WiredRequest>(respCallback, timerWorker);
    wr->SetRequestID(reqId);
    wr = SaveWiredRequest(reqId, wr);
    std::weak_ptr<WiredRequest> weakWr(wr);
    auto sendMsgHandler = [this, req, reqId, weakWr]() {
        if (auto thisPtr = weakWr.lock(); thisPtr) {
            auto messageId = YR::utility::IDGenerator::GenMessageId(reqId, static_cast<uint8_t>(thisPtr->retryCount));
            this->Write(GenStreamMsg(messageId, req), std::bind(&FSIntfImpl::WriteCallback, this, reqId, _1));
        }
    };

    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, reqId));
    if (timeoutSec > 0) {
        wr->SetupTimeout(timeoutSec, [this, reqId]() {
            auto wiredReq = GetWiredRequest(reqId, false);
            if (wiredReq != nullptr) {
                YRLOG_ERROR("Request timeout, start exec create resource group callback, request ID : {}", reqId);
                StreamingMessage fake;
                if (wiredReq->callback != nullptr) {
                    wiredReq->callback(fake,
                                       ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::CORE,
                                                 "create resource group request timeout, requestId: " + reqId),
                                       [](bool needEraseWiredReq) {});
                }
                EraseWiredRequest(reqId);
            }
        });
    }
}

void FSIntfImpl::EventAsync(const std::shared_ptr<EventMessageSpec> &req, int timeoutSec)
{
    YRLOG_DEBUG("Begin to event async, instanceid is {}, reqId is {}", req->Immutable().instanceid(),
                req->Immutable().requestid());
    std::string eventReqId = req->Immutable().requestid() + "_event_" + YR::utility::IDGenerator::GenRequestId();
    auto reqId = std::make_shared<std::string>(eventReqId);
    auto wr = std::make_shared<WiredRequest>(nullptr, nullptr, timerWorker, req->Immutable().instanceid());
    wr->SetRequestID(*reqId);
    wr = SaveWiredRequest(eventReqId, wr);
    std::weak_ptr<WiredRequest> weakWr(wr);
    std::shared_ptr<FSIntfImpl> self;
    try {
        self = shared_from_this();
    } catch (const std::exception &e) {
        YRLOG_ERROR("FSIntfImpl has been destructed");
    }
    auto sendMsgHandler = [self, reqId, req, weakWr]() {
        if (auto wr = weakWr.lock(); wr && self) {
            auto messageId = YR::utility::IDGenerator::GenMessageId(req->Immutable().requestid(),
                                                                    static_cast<uint8_t>(wr->retryCount));
            YRLOG_DEBUG("Send event message, message id {}", messageId);
            req->SetMessageID(messageId);
            self->TryDirectWriteEvent(req->Immutable().instanceid(), req->Get(),
                                 [self, reqId, wr](bool isDirect, ErrorInfo status) {
                                     if (status.OK()) {
                                         YRLOG_DEBUG("TryDirectWriteEvent status is ok, reqid : {}", *reqId);
                                         // If the message is sent successfully, the request is considered successful.
                                         self->GetWiredRequest(*reqId, true);
                                         self->EraseWiredRequest(*reqId);
                                         return;
                                     }
                                     if (self->IsCommunicationError(::common::ErrorCode(status.Code()))) {
                                         (void)self->SaveWiredRequest(*reqId, wr);
                                         YRLOG_ERROR("Communicate fails for request({}) errcode({}), msg({})", *reqId,
                                         fmt::underlying(status.Code()), status.Msg());
                                         return;
                                     }
                                     self->WriteCallback(*reqId, status);
                                 });
        }
    };
    sendMsgHandler();
    wr->SetupRetry(sendMsgHandler, std::bind(&FSIntfImpl::NeedRepeat, this, *reqId), true);
}

void FSIntfImpl::UpdateEventServerInfo(const std::string &ip, int port, const std::string &instanceId)
{
    std::unique_lock<std::mutex> lock(eventInfosMu_);
    auto &eventServerInfo = eventServerInfos[instanceId];
    if (!eventServerInfo) {
        eventServerInfo = std::make_shared<EventServerInfo>();
    }
    eventServerInfo->eventServerIp = ip;
    eventServerInfo->eventServerPort = port;
}

template <typename RespType>
void FSIntfImpl::WriteResponse(const std::string messageId, const RespType &resp)
{
    this->Write(GenStreamMsg<RespType>(messageId, resp));
}

template <typename RespType>
void FSIntfImpl::TryDirectWriteResponse(const std::string messageId, const std::string dstInstanceID,
                                        const RespType &resp, const bool existObjInDs)
{
    if (dstInstanceID != FUNCTION_PROXY && !existObjInDs) {
        // directly call does not need to response ack
        return;
    }
    this->TryDirectWrite(
        dstInstanceID, GenStreamMsg<RespType>(messageId, resp), [messageId, dstInstanceID](bool, ErrorInfo err) {
            if (err.OK()) {
                return;
            }
            YRLOG_WARN("failed to send resp {} to {}, err: {}", messageId, dstInstanceID, err.CodeAndMsg());
        });
}

void FSIntfImpl::RecvCallRequest(const std::string &from, const std::shared_ptr<StreamingMessage> &message)
{
    auto req = std::make_shared<CallMessageSpec>(message);
    this->HandleCallRequest(
        req, std::bind(&FSIntfImpl::TryDirectWriteResponse<CallResponse>, this, message->messageid(), from, _1, false));
}

void FSIntfImpl::NewRTIntfClient(const std::string &dstInstanceID, const NotifyRequest &req)
{
    auto rtIntf = fsInrfMgr->TryGet(dstInstanceID);
    if (rtIntf != nullptr && rtIntf->Available()) {
        return;
    }
    rtIntf = fsInrfMgr->NewFsIntfClient(
        instanceID, dstInstanceID, runtimeID,
        ReaderWriterClientOption{.ip = req.runtimeinfo().serveripaddr(),
                                 .port = req.runtimeinfo().serverport(),
                                 .disconnectedTimeout = RT_DISCONNECT_TIMEOUT_MS,
                                 .security = security,
                                 .resendCb = std::bind(&FSIntfImpl::ResendRequests, this, _1),
                                 .disconnectedCb = std::bind(&FSIntfImpl::NotifyDisconnected, this, _1),
                                 .isKeepAlive = false},
        ProtocolType::GRPC);
    rtIntf->RegisterMessageHandler(rtMsgHdlrs);
    (void)rtIntf->Start();
    (void)fsInrfMgr->Emplace(dstInstanceID, rtIntf);
}

std::shared_ptr<FSIntfReaderWriter> FSIntfImpl::NewOrGetEventIntfClient(const std::string &dstInstanceID)
{
    auto eventIntf = fsInrfMgr->TryGetEventIntfs(dstInstanceID);
    if (eventIntf != nullptr && eventIntf->Available()) {
        return eventIntf;
    }
    std::unique_lock<std::mutex> lock(eventInfosMu_);
    auto it = eventServerInfos.find(dstInstanceID);
    if (it == eventServerInfos.end()) {
        YRLOG_ERROR("No event server infomation of dstInstance: {}", dstInstanceID);
        return nullptr;
    }
    auto eventServerInfo = it->second;
    lock.unlock();
    if (eventIntf != nullptr && eventIntf->IsSameDstAddr(eventServerInfo->eventServerIp,
                                                         eventServerInfo->eventServerPort)) {
        YRLOG_DEBUG("EventIntfClient is not avaliable but no need to create new client, dstInstance ip: {}, port:{}",
                    eventServerInfo->eventServerIp,
                    eventServerInfo->eventServerPort);
        return eventIntf;
    }
    std::unique_lock<std::mutex> dstInstanceLock(eventServerInfo->dstMu);
    YRLOG_DEBUG("Begin to NewEventIntfClient, ip: {}, port:{}", eventServerInfo->eventServerIp,
                eventServerInfo->eventServerPort);

    eventIntf = fsInrfMgr->NewEventIntfClient(
        instanceID, dstInstanceID, runtimeID,
        ReaderWriterClientOption{.ip = eventServerInfo->eventServerIp,
                                 .port = eventServerInfo->eventServerPort,
                                 .disconnectedTimeout = RT_DISCONNECT_TIMEOUT_MS,
                                 .security = security,
                                 .resendCb = std::bind(&FSIntfImpl::ResendRequests, this, _1),
                                 .disconnectedCb = std::bind(&FSIntfImpl::NotifyDisconnected, this, _1),
                                 .isKeepAlive = false},
        ProtocolType::GRPC);
    eventIntf->RegisterMessageHandler(eventMsgHdlrs);
    auto error = eventIntf->Start();
    // If NewEventIntfClient return an existing available client, then start() will return ERR_INIT_CONNECTION_FAILED.
    // If start() fails, it will return ERR_CONNECTION_FAILED. Both cases don't need to execute EmplaceEventIntfs.
    if (error.OK()) {
        fsInrfMgr->EmplaceEventIntfs(dstInstanceID, eventIntf);
    }
    return eventIntf;
}

void FSIntfImpl::RecvNotifyRequest(const std::string &from, const std::shared_ptr<StreamingMessage> &message)
{
    auto reqId = message->notifyreq().requestid();
    YRLOG_DEBUG("recv notify request, req id:{}", reqId);
    auto wr = EraseWiredRequest(reqId);
    auto dstInstanceID = FUNCTION_PROXY;
    if (wr != nullptr) {
        dstInstanceID = wr->dstInstanceID;
    }
    if (dstInstanceID != FUNCTION_PROXY && message->notifyreq().has_runtimeinfo() &&
        !message->notifyreq().runtimeinfo().serveripaddr().empty() && wr != nullptr) {
        NewRTIntfClient(wr->dstInstanceID, message->notifyreq());
    }
    bool existObjInDs = false;
    if (wr != nullptr) {
        existObjInDs = message->notifyreq().smallobjects_size() != wr->returnObjectsSize;
    }
    this->HandleNotifyRequest(
        message->notifyreq(),
        [wr, message]() -> NotifyResponse {
            if (wr != nullptr && wr->notifyCallback) {
                wr->notifyCallback(message->notifyreq(), ErrorInfo());
            }
            return NotifyResponse();
        },
        std::bind(&FSIntfImpl::TryDirectWriteResponse<NotifyResponse>, this, message->messageid(), from, _1,
                  existObjInDs));
}

void FSIntfImpl::RecvCheckpointRequest(const std::string &, const std::shared_ptr<StreamingMessage> &message)
{
    this->HandleCheckpointRequest(message->checkpointreq(), std::bind(&FSIntfImpl::WriteResponse<CheckpointResponse>,
                                                                      this, message->messageid(), _1));
}

void FSIntfImpl::RecvRecoverRequest(const std::string &, const std::shared_ptr<StreamingMessage> &message)
{
    this->HandleRecoverRequest(message->recoverreq(),
                               std::bind(&FSIntfImpl::WriteResponse<RecoverResponse>, this, message->messageid(), _1));
}

void FSIntfImpl::RecvShutdownRequest(const std::string &, const std::shared_ptr<StreamingMessage> &message)
{
    YRLOG_DEBUG("grpc shutdown request, message id: {}, timeout exit period second: {}", message->messageid(),
                message->shutdownreq().graceperiodsecond());
    this->HandleShutdownRequest(message->shutdownreq(), std::bind(&FSIntfImpl::WriteResponse<ShutdownResponse>, this,
                                                                  message->messageid(), _1));
}

void FSIntfImpl::RecvSignalRequest(const std::string &, const std::shared_ptr<StreamingMessage> &message)
{
    this->HandleSignalRequest(message->signalreq(),
                              std::bind(&FSIntfImpl::WriteResponse<SignalResponse>, this, message->messageid(), _1));
}

void FSIntfImpl::RecvHeartbeatRequest(const std::string &, const std::shared_ptr<StreamingMessage> &message)
{
    this->HandleHeartbeatRequest(message->heartbeatreq(), std::bind(&FSIntfImpl::WriteResponse<HeartbeatResponse>, this,
                                                                    message->messageid(), _1));
}

bool FSIntfImpl::NeedResendReq(const std::shared_ptr<StreamingMessage> &message)
{
    switch (message->body_case()) {
        case StreamingMessage::kCreateRsp:
            return IsCommunicationError(message->creatersp().code());
        case StreamingMessage::kInvokeRsp:
            return IsCommunicationError(message->invokersp().code());
        case StreamingMessage::kCallResultAck:
            return IsCommunicationError(message->callresultack().code());
        case StreamingMessage::kKillRsp:
            return IsCommunicationError(message->killrsp().code());
        case StreamingMessage::kSaveRsp:
            return IsCommunicationError(message->saversp().code());
        case StreamingMessage::kLoadRsp:
            return IsCommunicationError(message->loadrsp().code());
        case StreamingMessage::kExitRsp:
            return false;
        case StreamingMessage::kCreateRsps:
            return IsCommunicationError(message->creatersps().code());
        case StreamingMessage::kRGroupRsp:
            return IsCommunicationError(message->rgrouprsp().code());
        default:
            YRLOG_ERROR("grpc body not match, messageid: {}, body case: {}", message->messageid(),
                        fmt::underlying(message->body_case()));
            return false;
    }
}

void FSIntfImpl::RecvCreateOrInvokeResponse(const std::string &, const std::shared_ptr<StreamingMessage> &message)
{
    auto reqId = YR::utility::IDGenerator::GetRequestIdFromMsg(message->messageid());
    YRLOG_DEBUG("receive create or invoke response, msg id {}, req id {}", message->messageid(), reqId);
    if (NeedResendReq(message)) {
        YRLOG_WARN("create or invoke response has communication error, need resend req, meesage id is {}",
                   message->messageid());
        return;
    }
    auto wr = GetWiredRequest(reqId, true);
    if (wr != nullptr && wr->callback != nullptr) {
        wr->callback(*message, ErrorInfo(), [this, &reqId](bool needEraseWiredReq) {
            if (needEraseWiredReq) {
                EraseWiredRequest(reqId);
            }
        });
    }
}

void FSIntfImpl::RecvEventRequest(const std::string &, const std::shared_ptr<StreamingMessage> &message)
{
    YRLOG_DEBUG("Begin to receive EventRequest");
    auto req = std::make_shared<EventMessageSpec>(message);
    this->HandleEventRequest(req);
}

void FSIntfImpl::RecvResponse(const std::string &, const std::shared_ptr<StreamingMessage> &message)
{
    auto reqId = YR::utility::IDGenerator::GetRequestIdFromMsg(message->messageid());
    YRLOG_DEBUG("req id {}", reqId);
    if (NeedResendReq(message)) {
        YRLOG_WARN("response has communication error, need resend req, meesage id is {}", message->messageid());
        return;
    }
    auto wr = EraseWiredRequest(reqId);
    if (wr != nullptr && wr->callback != nullptr) {
        wr->callback(*message, ErrorInfo(), [](bool) {});
    }
}

void FSIntfImpl::ResendRequests(const std::string &dstInstanceID)
{
    std::vector<std::shared_ptr<WiredRequest>> resendWrs;
    {
        absl::MutexLock lock(&this->mu);
        for (auto &wr : wiredRequests) {
            if (dstInstanceID != FUNCTION_PROXY && wr.second->dstInstanceID != dstInstanceID) {
                // while rt reconnected, only resend dst rt request
                continue;
            }

            if (auto intf = fsInrfMgr->TryGet(wr.second->dstInstanceID);
                dstInstanceID == FUNCTION_PROXY && intf != nullptr && intf->Available()) {
                // while fs reconnected, direct call client is alive, not need to resend
                continue;
            }

            if (dstInstanceID != FUNCTION_PROXY && wr.second->dstInstanceID == dstInstanceID) {
                YRLOG_DEBUG("direct call client {} disconnect, should resend with retry", dstInstanceID);
                resendWrs.emplace_back(wr.second);
                continue;
            }
            resendWrs.emplace_back(wr.second);
        }
    }
    for (auto &wr : resendWrs) {
        wr->ackReceived = false;
        wr->ResendRequestWithRetry();
    }
    YRLOG_INFO("current wired requests size: {}", wiredRequests.size());
    if (reSubscribeCb) {
        reSubscribeCb();
    }
}

void FSIntfImpl::NotifyDisconnected(const std::string &dstInstanceID)
{
    if (stopped) {
        return;
    }
    if (dstInstanceID != FUNCTION_PROXY) {
        YRLOG_WARN("{} disconnected. defer to resend request", dstInstanceID);
        (void)YR::utility::ExecuteByGlobalTimer(
            [weak(weak_from_this()), dstInstanceID]() {
                auto thisPtr = weak.lock();
                if (thisPtr != nullptr) {
                    thisPtr->ResendRequests(dstInstanceID);
                }
            },
            MILLISECOND_UNIT, 1);
        return;
    }
    // check whether the connection is reconnected.
    YRLOG_DEBUG("fs grpc reconnect timeout, pop remained reqs and set error");
    // still disconnected.
    std::unordered_map<std::string, std::shared_ptr<WiredRequest>> reqs = GetAllWiredRequests();
    for (auto &[requestId, req] : reqs) {
        auto rtIntf = fsInrfMgr->TryGet(req->dstInstanceID);
        if (req->notifyCallback && (rtIntf == nullptr || !rtIntf->Available())) {
            NotifyRequest notifyReq;
            notifyReq.set_code(common::ERR_BUS_DISCONNECTION);
            notifyReq.set_message("connected lost from proxy");
            notifyReq.set_requestid(requestId);
            req->notifyCallback(notifyReq, ErrorInfo());  // execute notifyCallback of each WiredRequest.
            EraseWiredRequest(requestId);
        }
    }
}

void CommunicationErrCallback(std::function<void(bool, ErrorInfo)> callback)
{
    ErrorInfo err;
    err.SetErrorCode(ERR_INNER_COMMUNICATION);
    err.SetErrorMsg("Function system client is unavailable.");
    if (callback != nullptr) {
        callback(false, err);
    }
}

void FSIntfImpl::Write(const std::shared_ptr<StreamingMessage> &msg, std::function<void(ErrorInfo)> callback)
{
    auto cb = [callback](bool, ErrorInfo err) {
        if (callback) {
            callback(err);
        }
    };
    auto rw = this->fsInrfMgr->GetSystemIntf();
    if (rw != nullptr) {
        rw->Write(msg, cb);
        return;
    }
    CommunicationErrCallback(cb);
}

void FSIntfImpl::TryDirectWrite(const std::string &dstInstanceID, const std::shared_ptr<StreamingMessage> &msg,
                                std::function<void(bool, ErrorInfo)> callback, std::function<void(bool)> preWrite)
{
    auto rw = this->fsInrfMgr->Get(dstInstanceID);
    if (rw != nullptr) {
        rw->Write(msg, callback, preWrite);
        return;
    }
    CommunicationErrCallback(callback);
}

void FSIntfImpl::TryDirectWriteEvent(const std::string &dstInstanceID, const std::shared_ptr<StreamingMessage> &msg,
                                    std::function<void(bool, ErrorInfo)> callback)
{
    auto rw = NewOrGetEventIntfClient(dstInstanceID);
    if (rw != nullptr) {
        rw->Write(msg, callback);
        return;
    }
    CommunicationErrCallback(callback);
}

FSIntfImpl::~FSIntfImpl()
{
    this->ClearAllWiredRequests();
}

std::pair<bus_service::DiscoverDriverResponse, ErrorInfo> FSIntfImpl::NotifyDriverDiscovery(
    const std::string &jobId, const std::string &instanceId, const std::string &functionName, int listeningPort)
{
    YRLOG_DEBUG("start to notify driver discovery jobId {}, instanceId {}, listeningPort {}", jobId, instanceId,
                listeningPort);
    std::string serverName = "";
    YR::GetServerName(security, serverName);
    std::shared_ptr<grpc::Channel> chan;
    grpc::ChannelArguments args;
    args.SetInt(GRPC_ARG_ENABLE_HTTP_PROXY, Config::Instance().YR_ENABLE_HTTP_PROXY() ? 1 : 0);
    if (!serverName.empty()) {
        args.SetSslTargetNameOverride(serverName);
    }
    chan = grpc::CreateCustomChannel(fsIp + ":" + std::to_string(fsPort), YR::GetChannelCreds(security), args);
    auto stub = bus_service::BusService::NewStub(chan);
    bus_service::DiscoverDriverRequest req;
    req.set_driverip(listeningIpAddr);
    req.set_driverport(std::to_string(listeningPort));
    req.set_jobid(jobId);
    req.set_instanceid(instanceId);
    req.set_functionname(functionName);

    const int maxRetryTime = 3;
    const int retryInternal = 2;
    int i = 0;
    grpc::Status status;
    bus_service::DiscoverDriverResponse resp;
    do {
        grpc::ClientContext ctx;
        status = stub->DiscoverDriver(&ctx, req, &resp);
        if (status.error_code() != grpc::StatusCode::OK) {
            i++;
            YRLOG_DEBUG("Discover driver call grpc status code: {}, retry index: {}",
                        fmt::underlying(status.error_code()), i);
            sleep(retryInternal);
        }
    } while (status.error_code() != grpc::StatusCode::OK && i < maxRetryTime);

    if (status.error_code() != grpc::StatusCode::OK) {
        YRLOG_ERROR("Discover driver call grpc status code: {}", fmt::underlying(status.error_code()));
        return std::make_pair(resp, ErrorInfo(ErrorCode::ERR_INIT_CONNECTION_FAILED, ModuleCode::RUNTIME,
                                              "failed to connect to cluster " + fsIp + ":" + std::to_string(fsPort)));
    }
    return std::make_pair(resp, ErrorInfo());
}

ErrorInfo FSIntfImpl::StartService(const std::string &jobID, const std::string &instanceID,
                                   const std::string &runtimeID, bool enableDirectCall, bool enableEvent)
{
    if (service != nullptr) {
        return {};
    }

    notification = std::make_shared<absl::Notification>();
    service = std::make_shared<GrpcPosixService>(instanceID, runtimeID, listeningIpAddr, selfPort, timerWorker,
                                                 notification, fsInrfMgr, security, enableDirectCall, enableEvent);
    if (enableEvent) {
        service->RegisterEventHdlrs(eventMsgHdlrs);
    }
    if (enableDirectCall || !enableEvent) {
        service->RegisterFSHandler(fsMsgHdlrs);
        service->RegisterRTHandler(rtMsgHdlrs);
        service->RegisterResendCallback(std::bind(&FSIntfImpl::ResendRequests, this, _1));
        service->RegisterDisconnectedCallback(std::bind(&FSIntfImpl::NotifyDisconnected, this, _1));
    }
    auto err = service->Start();
    if (!err.OK()) {
        return err;
    }
    this->selfPort = service->GetListeningPort();
    return {};
}

ErrorInfo FSIntfImpl::Start(const std::string &jobID, const std::string &instanceID, const std::string &runtimeID,
                            const std::string &functionName, const SubscribeFunc &subScribeCb)
{
    if (enableClientMode && enableDirectCall && !isDriver) {
        if (Config::Instance().POD_IP().empty()) {
            return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                             "POD_IP env should be properly set, while client mode & direct call enabled on cloud");
        }
        listeningIpAddr = Config::Instance().POD_IP();
        selfPort = 0;  // service listened port is dynamic set by grpc
    }
    this->instanceID = instanceID.empty() ? "driver-" + jobID : instanceID;
    this->runtimeID = runtimeID;
    if (!enableClientMode && (enableDirectCall || enableEvent)) {
        return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, ModuleCode::RUNTIME,
                         "does not support enable direct call or enableEvent when enable_server_mode is false");
    }
    if (!enableClientMode || enableDirectCall || enableEvent) {
        YRLOG_INFO("start with server mode {} or direct call {} or enableEvent {}, ready to start service",
                   !enableClientMode, enableDirectCall, enableEvent);
        // The service startup should be shielded in the future. The invoking here should not be aware of the protocol
        // grpc type.
        if (auto err = StartService(jobID, this->instanceID, runtimeID, enableDirectCall, enableEvent); !err.OK()) {
            return err;
        }
    }
    auto weakPtr = weak_from_this();
    auto discoverDriverCb = [this, weakPtr, jobID, functionName]() -> ErrorInfo {
        auto thisPtr = weakPtr.lock();
        if (!thisPtr) {
            return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "Has been destructed");
        }
        if (!isDriver) {
            return ErrorInfo();
        }
        if (auto [rsp, error] = this->NotifyDriverDiscovery(jobID, this->instanceID, functionName,
                                                            enableClientMode ? 0 : service->GetListeningPort());
            !error.OK()) {
            return error;
        } else {
            serverVersion_ = rsp.serverversion();
            thisPtr->nodeId_ = rsp.nodeid();
            thisPtr->nodeIp_ = rsp.hostip();
            YRLOG_INFO("serverVersion is {}, node id is {}, node ip is {}", serverVersion_, thisPtr->nodeId_,
                       thisPtr->nodeIp_);
        }
        return ErrorInfo();
    };
    if (auto err = discoverDriverCb(); !err.OK()) {
        return err;
    }
    this->reSubscribeCb = subScribeCb;
    if (enableClientMode) {
        auto fsIntf = fsInrfMgr->NewFsIntfClient(
            this->instanceID, "function-proxy", runtimeID,
            ReaderWriterClientOption{.ip = fsIp,
                                     .port = fsPort,
                                     .disconnectedTimeout = PROXY_DISCONNECT_TIMEOUT_MS,
                                     .security = security,
                                     .resendCb = std::bind(&FSIntfImpl::ResendRequests, this, _1),
                                     .disconnectedCb = std::bind(&FSIntfImpl::NotifyDisconnected, this, _1)},
            ProtocolType::GRPC);
        fsIntf->SetDiscoverDriverCb(discoverDriverCb);
        (void)fsInrfMgr->UpdateSystemIntf(fsIntf);
        fsIntf->RegisterMessageHandler(fsMsgHdlrs);
        auto err = fsIntf->Start();
        if (reSubscribeCb) {
            reSubscribeCb();
        }
        return err;
    }
    // default to wait 30s
    if (!notification->WaitForNotificationWithTimeout(absl::Seconds(30))) {
        return ErrorInfo(ErrorCode::ERR_INIT_CONNECTION_FAILED, ModuleCode::RUNTIME, "wait for connection timeout");
    }
    return {};
}

void FSIntfImpl::Stop(void)
{
    if (stopped) {
        return;
    }
    stopped = true;
    ClearAllWiredRequests();
    fsInrfMgr->Clear();
    if (service != nullptr) {
        service->Stop();
    }
    Clear();
}

void FSIntfImpl::RemoveInsRtIntf(const std::string &instanceId)
{
    fsInrfMgr->Remove(instanceId);
}

bool FSIntfImpl::IsHealth()
{
    if (!fsInrfMgr || stopped) {
        return false;
    }
    auto fsIntf = fsInrfMgr->GetSystemIntf();
    if (!fsIntf) {
        YRLOG_WARN("function system client reader writer is nullptr, return false directly");
        return false;
    }
    return fsIntf->IsHealth();
}

int FSIntfImpl::GetSelfPort() const
{
    return this->selfPort;
}

}  // namespace Libruntime
}  // namespace YR
