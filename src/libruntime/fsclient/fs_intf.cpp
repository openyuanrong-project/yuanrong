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

#include "fs_intf.h"

#include <algorithm>

#include "src/dto/config.h"
#include "src/utility/logger/logger.h"
namespace YR {
namespace Libruntime {
FSIntf::FSIntf(const FSIntfHandlers &handlers) : handlers(handlers)
{
    if (handlers.call == nullptr || handlers.checkpoint == nullptr || handlers.recover == nullptr ||
        handlers.shutdown == nullptr || handlers.signal == nullptr) {
        YRLOG_WARN("One or more function system handlers is empty!");
        return;
    }

    this->syncHeartbeat = this->handlers.heartbeat == nullptr ? true : false;
    if (this->syncHeartbeat) {
        this->handlers.heartbeat = std::bind(&FSIntf::HandleHeartbeat, this, std::placeholders::_1);
    }

    this->noitfyExecutor.Init(YR::Libruntime::Config::Instance().YR_NOTIFY_THREAD_POOL_SIZE(), "fs.notify");
    this->shutdownExecutor.Init(SHUTDOWN_THREAD_POOL_SIZE, "fs.shutdown");
    this->signalExecutor.Init(SIGNAL_THREAD_POOL_SIZE, "fs.signal");
    if (!this->syncHeartbeat) {
        this->heartbeatExecutor.Init(HEARTBEAT_THREAD_POOL_SIZE, "fs.heartbeat");
    }
    this->responseReceiver.Init(RESP_RECV_THREAD_POOL_SIZE, "fs.resp_recv");
    this->eventExecutor.Init(EVENT_THREAD_POOL_SIZE, "fs.event");
}

FSIntf::~FSIntf()
{
    Clear();
}

void FSIntf::Clear()
{
    if (cleared_) {
        return;
    }
    noitfyExecutor.Shutdown();
    checkpointRecoverExecutor.Shutdown();
    shutdownExecutor.Shutdown();
    signalExecutor.Shutdown();
    if (!this->syncHeartbeat) {
        heartbeatExecutor.Shutdown();
    }
    responseReceiver.Shutdown();
    callReceiver.Shutdown();
    eventExecutor.Shutdown();
    cleared_ = true;
}

void FSIntf::ReceiveRequestLoop(void)
{
    this->checkpointRecoverExecutor.Init(CKPT_RCVR_THREAD_POOL_SIZE, "fs.ckpt_rcvr");
    this->callReceiver.InitAndRun();
}

void FSIntf::ReturnCallResult(const std::shared_ptr<CallResultMessageSpec> result, bool isCreate,
                              CallResultCallBack callback)
{
    if (isCreate) {
        if (result->Immutable().code() == common::ERR_NONE) {
            status.SetInitialized();
        } else {
            status.SetInitializingFailure(result->Immutable().code(), result->Immutable().message());
        }
    }
    auto reqId = result->Immutable().requestid();
    this->CallResultAsync(result, [this, reqId, callback](const CallResultAck &ack) {
        if (!DeleteProcessingRequestId(reqId)) {
            YRLOG_ERROR("Call request has already finished, request ID: {}", reqId);
        }
        if (callback) {
            callback(ack);
        }
    });
}

HeartbeatResponse FSIntf::HandleHeartbeat(const HeartbeatRequest &hb)
{
    // heartbeat counter
    return HeartbeatResponse();
}

bool FSIntf::AddProcessingRequestId(const std::string &requestId)
{
    absl::MutexLock lock(&this->mu);
    auto [it, happened] = this->processingRequestIds.emplace(requestId);
    (void)it;
    return happened;
}

bool FSIntf::DeleteProcessingRequestId(const std::string &requestId)
{
    absl::MutexLock lock(&this->mu);
    auto num = this->processingRequestIds.erase(requestId);
    if (this->processingRequestIds.empty()) {
        cv_.SignalAll();
    }
    return num == 1;
}

bool FSIntf::ExistProcessingRequestId()
{
    absl::MutexLock lock(&this->mu);
    return !this->processingRequestIds.empty();
}

void FSIntf::HandleCallRequest(const std::shared_ptr<CallMessageSpec> &req, CallCallBack callback)
{
    if (!AddProcessingRequestId(req->Immutable().requestid())) {
        YRLOG_DEBUG("Duplicated call request, request ID: {}", req->Immutable().requestid());
        CallResponse resp;
        resp.set_code(common::ERR_NONE);
        callback(resp);
        return;
    }

    YRLOG_DEBUG("Receive call request, request ID: {}", req->Immutable().requestid());
    this->callReceiver.Handle(
        [this, req, callback]() {
            CallResponse resp;
            if (req->Immutable().iscreate()) {
                if (!status.SetInitializing()) {
                    status.WaitInitialized();
                    auto [code, msg] = status.GetErrorInfo();
                    resp.set_code(code);
                    resp.set_message(msg);
                    YRLOG_DEBUG("send init call response, request ID: {}, code {}, message {}",
                                req->Immutable().requestid(), fmt::underlying(resp.code()), resp.message());
                    callback(resp);
                } else {
                    callback(resp);
                    this->handlers.init(req);
                    YRLOG_DEBUG("send init call response , request ID: {}, code {}, message {}",
                                req->Immutable().requestid(), fmt::underlying(resp.code()), resp.message());
                }
            } else {
                if (!status.WaitInitialized() && req->Immutable().createoptions().find("ENABLE_FORCE_INVOKE") ==
                                                     req->Immutable().createoptions().end()) {
                    auto [code, msg] = status.GetErrorInfo();
                    resp.set_code(code);
                    resp.set_message(msg);
                    YRLOG_DEBUG("after wait initialized, send call response, request ID: {}, code {}, message {}",
                                req->Immutable().requestid(), fmt::underlying(resp.code()), resp.message());
                    status.InterruptCheckpointing();
                    callback(resp);
                } else {
                    callback(resp);
                    this->handlers.call(req);
                    YRLOG_DEBUG("send call response , request ID: {}, code {}, message {}",
                                req->Immutable().requestid(), fmt::underlying(resp.code()), resp.message());
                }
            }
            if (resp.code() != common::ERR_NONE) {
                DeleteProcessingRequestId(req->Immutable().requestid());
            }
        },
        "");
}

void FSIntf::HandleNotifyRequest(const NotifyRequest &req, std::function<NotifyResponse(void)> createOrInvokeCallback,
                                 NotifyCallBack callback)
{
    this->noitfyExecutor.Handle(
        [createOrInvokeCallback, callback]() {
            auto resp = createOrInvokeCallback();
            callback(resp);
        },
        "");
}

void FSIntf::HandleCheckpointRequest(const CheckpointRequest &req, CheckpointCallBack callback)
{
    this->checkpointRecoverExecutor.Handle(
        [this, req, callback]() {
            if (ExistProcessingRequestId()) {
                CheckpointResponse resp;
                resp.set_code(common::ERR_INSTANCE_BUSY);
                resp.set_message("Instance is busy handling requests, checkpoint cannot be performed now.");
                callback(resp);
                return;
            }
            if (status.SetCheckpointing()) {
                auto resp = this->handlers.checkpoint(req);
                if (status.IsCheckpointingInterrupted()) {
                    CheckpointResponse resp;
                    auto [code, msg] = status.GetErrorInfo();
                    resp.set_code(code);
                    resp.set_message(msg);
                    callback(resp);
                } else {
                    callback(resp);
                }
                status.SetCheckpointed();
                return;
            }
            CheckpointResponse resp;
            auto [code, msg] = status.GetErrorInfo();
            resp.set_code(code);
            resp.set_message(msg);
            callback(resp);
        },
        "");
}

void FSIntf::HandleRecoverRequest(const RecoverRequest &req, RecoverCallBack callback)
{
    this->checkpointRecoverExecutor.Handle(
        [this, req, callback]() {
            auto resp = this->handlers.recover(req);
            if (resp.code() == common::ERR_NONE) {
                YRLOG_DEBUG("Set initialized status for recover");
                status.SetInitializing();
                status.SetInitialized();
            }
            callback(resp);
        },
        "");
}

void FSIntf::HandleShutdownRequest(const ShutdownRequest &req, ShutdownCallBack callback)
{
    this->shutdownExecutor.Handle(
        [this, req, callback]() {
            if (!status.SetShuttingDown()) {
                status.WaitShutdown();
                ShutdownResponse resp;
                resp.set_code(common::ERR_NONE);
                callback(resp);
            } else {
                YRLOG_DEBUG("will exec handlers Shutdown");
                auto resp = this->handlers.shutdown(req);
                callback(resp);
                status.SetShutdown();
            }
        },
        "");
}

int FSIntf::WaitRequestEmpty(uint64_t gracePeriodSec)
{
    absl::MutexLock lock(&this->mu);
    const static size_t RESERVE_SECOND = 1;
    absl::Time deadline = absl::Now() + absl::Seconds(std::max(0, static_cast<int>(gracePeriodSec - RESERVE_SECOND)));

    if (!processingRequestIds.empty() && !isShutdownDone) {
        if (cv_.WaitWithDeadline(&this->mu, deadline)) {
            std::stringstream ss;
            std::for_each(processingRequestIds.begin(), processingRequestIds.end(),
                          [&ss](const std::string &element) { ss << element << " "; });
            YRLOG_DEBUG("shutdown wait timeout, There are still unfinished requests: {}", ss.str());
        }
    }

    isShutdownDone = true;
    cv_.SignalAll();

    absl::Duration remainingTime = deadline - absl::Now() + absl::Seconds(RESERVE_SECOND);
    return std::max(0, static_cast<int>(remainingTime / absl::Seconds(RESERVE_SECOND)));
}

void FSIntf::HandleSignalRequest(const SignalRequest &req, SignalCallBack callback)
{
    this->signalExecutor.Handle(
        [this, req, callback]() {
            YRLOG_DEBUG("receive signal req, signal is {}, payload is {}", req.signal(), req.payload());
            auto resp = this->handlers.signal(req);
            callback(resp);
        },
        "");
}

void FSIntf::HandleHeartbeatRequest(const HeartbeatRequest &req, HeartbeatCallBack callback)
{
    if (this->syncHeartbeat) {
        auto resp = this->handlers.heartbeat(req);
        callback(resp);
        return;
    }

    this->heartbeatExecutor.Handle(
        [this, req, callback]() {
            auto resp = this->handlers.heartbeat(req);
            callback(resp);
        },
        "");
}

void FSIntf::HandleEventRequest(const std::shared_ptr<EventMessageSpec> &req)
{
    this->eventExecutor.Handle(
        [this, req]() {
            YRLOG_DEBUG("receive event req, request ID: {}", req->Immutable().requestid());
            this->handlers.event(req);
        },
        "");
}

std::string FSIntf::GetServerVersion()
{
    return this->serverVersion_;
}

std::pair<ErrorInfo, std::string> FSIntf::GetNodeId()
{
    return std::make_pair(ErrorInfo(), this->nodeId_);
}

std::pair<ErrorInfo, std::string> FSIntf::GetNodeIp()
{
    return std::make_pair(ErrorInfo(), this->nodeIp_);
}

void FSIntf::SetInitialized()
{
    if (status.SetInitializing()) {
        status.SetInitialized();
    }
}
}  // namespace Libruntime
}  // namespace YR
