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

#pragma once

#include <algorithm>
#include <functional>
#include <string>

#include <grpcpp/grpcpp.h>
#include "absl/synchronization/mutex.h"

#include "src/dto/config.h"
#include "src/dto/constant.h"
#include "src/dto/status.h"
#include "src/libruntime/clientsmanager/clients_manager.h"
#include "src/libruntime/fsclient/fs_intf.h"
#include "src/libruntime/fsclient/fs_intf_manager.h"
#include "src/libruntime/fsclient/fs_intf_reader_writer.h"
#include "src/libruntime/fsclient/protobuf/bus_service.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/common.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/core_service.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/runtime_service.grpc.pb.h"
// should be abstracted in the future
#include "src/libruntime/fsclient/grpc/grpc_posix_service.h"

#include "src/libruntime/utils/security.h"
#include "src/utility/id_generator.h"
#include "src/utility/logger/logger.h"
#include "src/utility/timer_worker.h"

namespace YR {
namespace Libruntime {
using StreamingMessage = ::runtime_rpc::StreamingMessage;
using YR::utility::TimerWorker;
using BodyCase = ::runtime_rpc::StreamingMessage::BodyCase;

template <typename Type>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const Type &msg)
{
    STDERR_AND_THROW_EXCEPTION(ERR_INNER_SYSTEM_ERROR, RUNTIME,
                               "GenStreamMsg method with std::string and Type not implemented");
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const CreateResponses &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_creatersps()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const CreateResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_creatersp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const NotifyRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_notifyreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const InvokeResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_invokersp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const CallResultAck &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_callresultack()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const KillResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_killrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const ExitResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_exitrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const StateSaveResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_saversp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const StateLoadResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_loadrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const CallRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_callreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const CheckpointRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_checkpointreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const RecoverRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_recoverreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const ShutdownRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_shutdownreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const SignalRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_signalreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const HeartbeatRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_heartbeatreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const CallResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_callrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const NotifyResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_notifyrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const CheckpointResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_checkpointrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const RecoverResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_recoverrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const ShutdownResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_shutdownrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const SignalResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_signalrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const HeartbeatResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_heartbeatrsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId,
                                                      const CreateResourceGroupResponse &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_rgrouprsp()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const CreateRequests &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_createreqs()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const CreateRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_createreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const KillRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_killreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const ExitRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_exitreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const StateSaveRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_savereq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const StateLoadRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_loadreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId,
                                                      const CreateResourceGroupRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_rgroupreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

template <>
inline std::shared_ptr<StreamingMessage> GenStreamMsg(const std::string &messageId, const EventRequest &msg)
{
    auto streamMsg = std::make_shared<StreamingMessage>();
    streamMsg->mutable_eventreq()->CopyFrom(msg);
    streamMsg->set_messageid(messageId);
    return streamMsg;
}

struct WiredRequest : public std::enable_shared_from_this<WiredRequest> {
    WiredRequest() = default;
    WiredRequest(std::function<void(StreamingMessage, ErrorInfo, std::function<void(bool)>)> cb,
                 std::shared_ptr<TimerWorker> tw)
        : callback(cb), notifyCallback(nullptr), retryCount(0), timer_(nullptr), timerWorkerWeak(tw)
    {
    }

    WiredRequest(std::function<void(StreamingMessage, ErrorInfo, std::function<void(bool)>)> cb,
                 std::function<void(const NotifyRequest &, const ErrorInfo &)> cbNotify,
                 std::shared_ptr<TimerWorker> tw, const std::string &dstInstance = "function-proxy")
        : callback(cb),
          notifyCallback(cbNotify),
          retryCount(0),
          timer_(nullptr),
          timerWorkerWeak(tw),
          dstInstanceID(dstInstance)
    {
    }

    WiredRequest(const WiredRequest &) = delete;
    WiredRequest(WiredRequest &&) = delete;

    WiredRequest &operator=(const WiredRequest &) = delete;
    WiredRequest &operator=(WiredRequest &&) = delete;

    ~WiredRequest()
    {
        CancelAllTimer();
        if (timer_ != nullptr) {
            timer_.reset();
        }
        if (timerForTimeout != nullptr) {
            timerForTimeout.reset();
        }
    }

    void SetupRetry(std::function<void()> &&retry, std::function<bool()> &&needRetry, bool exponentialBackoff = false)
    {
        std::shared_ptr<TimerWorker> timerWorker = timerWorkerWeak.lock();
        if (timerWorker == nullptr) {
            return;
        }

        this->retryHdlr = std::move(retry);
        this->needRetryHdlr = std::move(needRetry);
        this->exponentialBackoff = exponentialBackoff;
        int currentRetryInterval = std::min(requestACKTimeout, Config::Instance().REQUEST_ACK_ACC_MAX_SEC());
        this->retryIntervalSec = currentRetryInterval;
        this->remainTimeoutSec = Config::Instance().REQUEST_ACK_ACC_MAX_SEC();
        auto weakThis = weak_from_this();
        this->timer_ = timerWorker->CreateTimer(currentRetryInterval * YR::Libruntime::MILLISECOND_UNIT, 1, [weakThis] {
            if (auto thisPtr = weakThis.lock(); thisPtr) {
                thisPtr->RetryWrapper();
            }
        });
    }

    void RetryWrapper()
    {
        if (!timer_ || !needRetryHdlr || !retryHdlr) {
            return;
        }
        // needRetry will update new intervalSec.
        if (!needRetryHdlr()) {
            if (timer_ != nullptr) {
                timer_->cancel();
                timer_ = nullptr;
            }
            return;
        }

        // if need retry, get new retry intervalSec , set timer.
        std::shared_ptr<TimerWorker> timerWorker = this->timerWorkerWeak.lock();
        if (timerWorker != nullptr) {
            auto weakThis = weak_from_this();
            timerWorker->ExecuteByTimer(timer_, this->retryIntervalSec * YR::Libruntime::MILLISECOND_UNIT, [weakThis] {
                if (auto thisPtr = weakThis.lock(); thisPtr) {
                    thisPtr->RetryWrapper();
                }
            });
        }
        YRLOG_INFO("Req {} will retry without ack, count: {}", reqId_, retryCount);
        retryHdlr();
    }

    void ResendRequest(void)
    {
        if (retryHdlr) {
            retryHdlr();
        }
    }

    void ResendRequestWithRetry(void)
    {
        YRLOG_DEBUG("RPC resend request with retry");
        if (retryHdlr) {
            retryHdlr();
        }
        std::shared_ptr<TimerWorker> timerWorker = timerWorkerWeak.lock();
        if (timerWorker == nullptr) {
            return;
        }
        int currentRetryInterval = std::min(requestACKTimeout, Config::Instance().REQUEST_ACK_ACC_MAX_SEC());
        auto weakThis = weak_from_this();
        if (timer_ != nullptr) {
            return;
        }
        this->timer_ = timerWorker->CreateTimer(currentRetryInterval * YR::Libruntime::MILLISECOND_UNIT, 1, [weakThis] {
            if (auto thisPtr = weakThis.lock(); thisPtr) {
                thisPtr->RetryWrapper();
            }
        });
    }

    void SetupTimeout(int timeoutSec, std::function<void()> cb)
    {
        if (timeoutSec < 0) {
            return;
        }
        std::shared_ptr<TimerWorker> timerWorker = timerWorkerWeak.lock();
        if (timerWorker == nullptr) {
            return;
        }

        timerForTimeout = timerWorker->CreateTimer(timeoutSec * YR::Libruntime::MILLISECOND_UNIT, 1, cb);
    }

    void CancelAllTimer()
    {
        if (timer_ != nullptr) {
            timer_->cancel();
        }
        if (timerForTimeout != nullptr) {
            timerForTimeout->cancel();
        }
    }

    void SetRequestID(const std::string &reqId)
    {
        reqId_ = reqId;
    }

    std::string GetRequestID()
    {
        return reqId_;
    }

    std::function<void(StreamingMessage, ErrorInfo, std::function<void(bool)>)> callback;
    std::function<void(const NotifyRequest &, const ErrorInfo &)> notifyCallback;
    size_t retryCount;
    std::atomic<bool> ackReceived = {false};  // true if create or invoke response has been received
    std::shared_ptr<YR::utility::Timer> timer_;
    std::shared_ptr<YR::utility::Timer> timerForTimeout;
    std::weak_ptr<TimerWorker> timerWorkerWeak;
    bool exponentialBackoff;
    int remainTimeoutSec;
    int retryIntervalSec;
    std::string dstInstanceID;
    int returnObjectsSize{0};

private:
    std::function<void()> retryHdlr;
    std::function<bool()> needRetryHdlr;
    std::string reqId_ = "";
    inline static size_t requestACKTimeout = REQUEST_ACK_TIMEOUT_SEC;
};

class FSIntfImpl : public FSIntf, public std::enable_shared_from_this<FSIntfImpl> {
public:
    // ipAddr is listening Ip while server mode
    // ipAddr is fs Ip while client mode
    // port is
    FSIntfImpl(const std::string &ipAddr, int port, FSIntfHandlers handlers, bool isDriver,
               std::shared_ptr<Security> sec, std::shared_ptr<ClientsManager> clientsMgr, bool enableClientMode,
               bool enableEvent = false);
    ~FSIntfImpl();

    ErrorInfo Start(const std::string &jobID, const std::string &instanceID = "", const std::string &runtimeID = "",
                    const std::string &functionName = "", const SubscribeFunc &subScribeCb = nullptr) override;
    void Stop(void) override;
    void GroupCreateAsync(const CreateRequests &reqs, CreateRespsCallback respCallback, CreateCallBack callback,
                          int timeoutSec = -1);
    void CreateAsync(const CreateRequest &req, CreateRespCallback respCallback, CreateCallBack callback,
                     int timeoutSec = -1);
    void InvokeAsync(const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback, int timeoutSec = -1);
    void CallResultAsync(const std::shared_ptr<CallResultMessageSpec> req, CallResultCallBack callback);
    void KillAsync(const KillRequest &req, KillCallBack callback, int timeoutSec = -1);
    void ExitAsync(const ExitRequest &req, ExitCallBack callback);
    void StateSaveAsync(const StateSaveRequest &req, StateSaveCallBack callback);
    void StateLoadAsync(const StateLoadRequest &req, StateLoadCallBack callback);
    void CreateRGroupAsync(const CreateResourceGroupRequest &req, CreateResourceGroupCallBack callback,
                           int timeoutSec = -1);
    void EventAsync(const std::shared_ptr<EventMessageSpec> &req, int timeoutSec = -1);
    void UpdateEventServerInfo(const std::string &ip, int port, const std::string &instaceId);
    void ResendRequests(const std::string &dstInstanceID);
    void NotifyDisconnected(const std::string &dstInstanceID);
    bool NeedResendReq(const std::shared_ptr<StreamingMessage> &message);
    inline void EnableClientMode()
    {
        enableClientMode = true;
    }
    void EnableDirectCall()
    {
        enableDirectCall = true;
    }
    void RemoveInsRtIntf(const std::string &instanceId) override;
    bool IsHealth() override;
    int GetSelfPort() const;

protected:
    void Write(const std::shared_ptr<StreamingMessage> &msg, std::function<void(ErrorInfo)> callback = nullptr);
    void TryDirectWrite(const std::string &dstInstanceID, const std::shared_ptr<StreamingMessage> &msg,
                        std::function<void(bool, ErrorInfo)> callback, std::function<void(bool)> preWrite = nullptr);
    void TryDirectWriteEvent(const std::string &dstInstanceID, const std::shared_ptr<StreamingMessage> &msg,
                                        std::function<void(bool, ErrorInfo)> callback);
    std::pair<bus_service::DiscoverDriverResponse, ErrorInfo> NotifyDriverDiscovery(const std::string &jobId,
                                                                                    const std::string &instanceId,
                                                                                    const std::string &functionName,
                                                                                    int listeningPort);
    ErrorInfo StartService(const std::string &jobID, const std::string &instanceID, const std::string &runtimeID,
                           bool enableDirectCall, bool enableEvent = false);

    std::string fsIp;
    std::string listeningIpAddr;
    int selfPort;
    int fsPort;
    bool isDriver;
    bool enableClientMode = false;
    bool enableDirectCall = false;
    bool enableEvent = false;
    mutable absl::Mutex mu;
    std::unordered_map<std::string, std::shared_ptr<WiredRequest>> wiredRequests ABSL_GUARDED_BY(mu);
    std::shared_ptr<TimerWorker> timerWorker;
    std::shared_ptr<absl::Notification> notification;
    std::shared_ptr<Security> security;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::shared_ptr<FSIntfManager> fsInrfMgr;
    std::shared_ptr<GrpcPosixService> service;
    std::string instanceID;
    std::string runtimeID;
    std::atomic<bool> stopped = {false};

private:
    std::unordered_map<std::string, std::shared_ptr<WiredRequest>> GetAllWiredRequests(void);
    void ClearAllWiredRequests(void);
    std::shared_ptr<WiredRequest> SaveWiredRequest(const std::string &reqId, std::shared_ptr<WiredRequest> wr);
    std::shared_ptr<WiredRequest> EraseWiredRequest(const std::string &reqId);
    std::shared_ptr<WiredRequest> GetWiredRequest(const std::string &reqId, bool ackReceived);
    void UpdateWiredRequestRemote(const std::string &reqId, const std::string &remote);
    std::pair<std::shared_ptr<WiredRequest>, bool> UpdateRetryInterval(const std::string &reqId);
    bool IsCommunicationError(const ::common::ErrorCode &code);
    bool NeedRepeat(const std::string requestId);
    void WriteCallback(const std::string requestId, const ErrorInfo &status);
    void NewRTIntfClient(const std::string &remote, const NotifyRequest &req);
    std::shared_ptr<FSIntfReaderWriter> NewOrGetEventIntfClient(const std::string &dstInstanceID);

    template <typename RespType>
    void WriteResponse(const std::string messageId, const RespType &resp);

    template <typename RespType>
    void TryDirectWriteResponse(const std::string messageId, const std::string remote, const RespType &resp,
                                const bool existObjInDs = false);

    void RecvCallRequest(const std::string &, const std::shared_ptr<StreamingMessage> &);
    void RecvNotifyRequest(const std::string &, const std::shared_ptr<StreamingMessage> &);
    void RecvCheckpointRequest(const std::string &, const std::shared_ptr<StreamingMessage> &);
    void RecvRecoverRequest(const std::string &, const std::shared_ptr<StreamingMessage> &);
    void RecvShutdownRequest(const std::string &, const std::shared_ptr<StreamingMessage> &);
    void RecvSignalRequest(const std::string &, const std::shared_ptr<StreamingMessage> &);
    void RecvHeartbeatRequest(const std::string &, const std::shared_ptr<StreamingMessage> &);
    void RecvCreateOrInvokeResponse(const std::string &, const std::shared_ptr<StreamingMessage> &);
    void RecvEventRequest(const std::string &, const std::shared_ptr<StreamingMessage> &);
    void RecvResponse(const std::string &, const std::shared_ptr<StreamingMessage> &);
    std::unordered_map<BodyCase, MsgHdlr> fsMsgHdlrs;
    std::unordered_map<BodyCase, MsgHdlr> rtMsgHdlrs;
    std::unordered_map<BodyCase, MsgHdlr> eventMsgHdlrs;
    SubscribeFunc reSubscribeCb;
    struct EventServerInfo {
        std::string eventServerIp;
        int eventServerPort;
        std::mutex dstMu;
    };
    std::unordered_map<std::string, std::shared_ptr<EventServerInfo>> eventServerInfos;
    std::mutex eventMu_;
    std::mutex eventInfosMu_;
};
}  // namespace Libruntime
}  // namespace YR
