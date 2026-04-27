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

#include "absl/synchronization/mutex.h"
#include "absl/synchronization/notification.h"
#include "json.hpp"
#include "src/dto/status.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/fsclient/protobuf/bus_service.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/common.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/core_service.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/data_service.pb.h"
#include "src/libruntime/fsclient/protobuf/lease_service.pb.h"
#include "src/libruntime/fsclient/protobuf/runtime_rpc.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/runtime_service.grpc.pb.h"
#include "src/libruntime/invokeadaptor/agent_session_manager.h"
#include "src/utility/id_generator.h"
#include "src/utility/thread_pool.h"

namespace YR {
namespace Libruntime {
using YR::utility::ThreadPool;
using SubscribeFunc = std::function<void()>;

using PutRequest = ::data_service::PutRequest;
using PutResponse = ::data_service::PutResponse;

using GetRequest = ::data_service::GetRequest;
using GetResponse = ::data_service::GetResponse;

using IncreaseRefRequest = ::data_service::IncreaseRefRequest;
using IncreaseRefResponse = ::data_service::IncreaseRefResponse;

using DecreaseRefRequest = ::data_service::DecreaseRefRequest;
using DecreaseRefResponse = ::data_service::DecreaseRefResponse;

using KvSetRequest = ::data_service::KvSetRequest;
using KvSetResponse = ::data_service::KvSetResponse;

using KvMSetTxRequest = ::data_service::KvMSetTxRequest;
using KvMSetTxResponse = ::data_service::KvMSetTxResponse;

using KvGetRequest = ::data_service::KvGetRequest;
using KvGetResponse = ::data_service::KvGetResponse;

using KvDelRequest = ::data_service::KvDelRequest;
using KvDelResponse = ::data_service::KvDelResponse;

using LeaseRequest = ::lease_service::LeaseRequest;
using LeaseResponse = ::lease_service::LeaseResponse;

using CreateRequests = ::core_service::CreateRequests;
using GroupOptions = ::core_service::GroupOptions;
using BindOptions = ::core_service::BindOptions;
using CreateResponses = ::core_service::CreateResponses;

using CreateRequest = ::core_service::CreateRequest;
using CreateResponse = ::core_service::CreateResponse;
using SchedulingOptions = ::core_service::SchedulingOptions;

using InvokeRequest = ::core_service::InvokeRequest;
using InvokeResponse = ::core_service::InvokeResponse;

using CallResult = ::core_service::CallResult;
using CallResultAck = ::core_service::CallResultAck;

using KillRequest = ::core_service::KillRequest;
using KillResponse = ::core_service::KillResponse;
using SubscriptionPayload = ::core_service::SubscriptionPayload;
using NotificationPayload = ::core_service::NotificationPayload;
using EventPayload = ::core_service::EventPayload;
using InstanceTermination = ::core_service::InstanceTermination;
using FunctionMasterObserve = ::core_service::FunctionMasterObserve;
using KillCallBack = std::function<void(const KillResponse &, const ErrorInfo &err)>;

using ExitRequest = ::core_service::ExitRequest;
using ExitResponse = ::core_service::ExitResponse;

using StateSaveRequest = ::core_service::StateSaveRequest;
using StateSaveResponse = ::core_service::StateSaveResponse;

using StateLoadRequest = ::core_service::StateLoadRequest;
using StateLoadResponse = ::core_service::StateLoadResponse;

using CreateResourceGroupRequest = ::core_service::CreateResourceGroupRequest;
using CreateResourceGroupResponse = ::core_service::CreateResourceGroupResponse;
using CreateResourceGroupCallBack = std::function<void(const CreateResourceGroupResponse &)>;

using EventRequest = ::core_service::EventRequest;

using CallRequest = ::runtime_service::CallRequest;
using CallResponse = ::runtime_service::CallResponse;
using CallCallBack = std::function<void(const CallResponse &)>;

using NotifyRequest = ::runtime_service::NotifyRequest;
using NotifyResponse = ::runtime_service::NotifyResponse;

using CheckpointRequest = ::runtime_service::CheckpointRequest;
using CheckpointResponse = ::runtime_service::CheckpointResponse;

using RecoverRequest = ::runtime_service::RecoverRequest;
using RecoverResponse = ::runtime_service::RecoverResponse;

using ShutdownRequest = ::runtime_service::ShutdownRequest;
using ShutdownResponse = ::runtime_service::ShutdownResponse;

using PrepareSnapRequest = ::runtime_service::PrepareSnapRequest;
using PrepareSnapResponse = ::runtime_service::PrepareSnapResponse;

using SnapStartedRequest = ::runtime_service::SnapStartedRequest;
using SnapStartedResponse = ::runtime_service::SnapStartedResponse;

using SignalRequest = ::runtime_service::SignalRequest;
using SignalResponse = ::runtime_service::SignalResponse;

using HeartbeatRequest = ::runtime_service::HeartbeatRequest;
using HeartbeatResponse = ::runtime_service::HeartbeatResponse;

using Arg = common::Arg;
using Arg_ArgType = common::Arg_ArgType;

const int CKPT_RCVR_THREAD_POOL_SIZE = 1;
const int SHUTDOWN_THREAD_POOL_SIZE = 1;
const int SIGNAL_THREAD_POOL_SIZE = 10;
const int HEARTBEAT_THREAD_POOL_SIZE = 1;
const int RESP_RECV_THREAD_POOL_SIZE = 1;
const int EVENT_THREAD_POOL_SIZE = 1;
const int SLEEP_INTERVAL_BEFORE_TRACEPOINT_MS = 1000;

class MessgeSpec {
public:
    MessgeSpec() : msg(std::make_shared<::runtime_rpc::StreamingMessage>()) {}
    explicit MessgeSpec(const std::shared_ptr<::runtime_rpc::StreamingMessage> &msg) : msg(msg) {}
    virtual ~MessgeSpec() = default;
    void SetMessageID(const std::string &messageID)
    {
        msg->set_messageid(messageID);
    }
    [[nodiscard]] std::shared_ptr<::runtime_rpc::StreamingMessage> Get() const
    {
        return msg;
    }

protected:
    std::shared_ptr<::runtime_rpc::StreamingMessage> msg;
};

class InvokeMessageSpec : public MessgeSpec {
public:
    InvokeMessageSpec() : MessgeSpec()
    {
        msg->mutable_invokereq();
    }
    explicit InvokeMessageSpec(InvokeRequest &&req) : MessgeSpec()
    {
        *msg->mutable_invokereq() = std::move(req);
    }
    ~InvokeMessageSpec() override = default;
    InvokeRequest &Mutable()
    {
        return *msg->mutable_invokereq();
    }
    const InvokeRequest &Immutable() const
    {
        return msg->invokereq();
    }
};

class CallResultMessageSpec : public MessgeSpec {
public:
    CallResultMessageSpec() : MessgeSpec()
    {
        msg->mutable_callresultreq();
    }
    ~CallResultMessageSpec() override = default;
    CallResult &Mutable()
    {
        return *msg->mutable_callresultreq();
    }
    const CallResult &Immutable() const
    {
        return msg->callresultreq();
    }
    void UpdateRuntimeInfo(const std::string &serverIp, const int &serverPort)
    {
        msg->mutable_callresultreq()->mutable_runtimeinfo()->set_serveripaddr(serverIp);
        msg->mutable_callresultreq()->mutable_runtimeinfo()->set_serverport(serverPort);
    }
    bool existObjInDs{false};
};

class CallMessageSpec : public MessgeSpec {
public:
    CallMessageSpec() : MessgeSpec() {};
    explicit CallMessageSpec(const std::shared_ptr<::runtime_rpc::StreamingMessage> &msg) : MessgeSpec(msg) {}
    ~CallMessageSpec() override = default;
    CallRequest &Mutable()
    {
        return *msg->mutable_callreq();
    }
    const CallRequest &Immutable() const
    {
        return msg->callreq();
    }
};

class EventMessageSpec : public MessgeSpec {
public:
    EventMessageSpec() : MessgeSpec() {};
    explicit EventMessageSpec(const std::shared_ptr<::runtime_rpc::StreamingMessage> &msg) : MessgeSpec(msg) {}
    ~EventMessageSpec() override = default;
    EventRequest &Mutable()
    {
        return *msg->mutable_eventreq();
    }
    const EventRequest &Immutable() const
    {
        return msg->eventreq();
    }
};

using CallHandler = std::function<void(const std::shared_ptr<CallMessageSpec> &)>;
using CallCallBack = std::function<void(const CallResponse &)>;
using NotifyHandler = std::function<NotifyResponse(const NotifyRequest &)>;
using NotifyCallBack = std::function<void(const NotifyResponse &)>;
using CheckpointHandler = std::function<CheckpointResponse(const CheckpointRequest &)>;
using CheckpointCallBack = std::function<void(const CheckpointResponse &)>;
using RecoverHandler = std::function<RecoverResponse(const RecoverRequest &)>;
using RecoverCallBack = std::function<void(const RecoverResponse &)>;
using ShutdownHandler = std::function<ShutdownResponse(const ShutdownRequest &)>;
using ShutdownCallBack = std::function<void(const ShutdownResponse &)>;
using PrepareSnapHandler = std::function<PrepareSnapResponse(const PrepareSnapRequest &)>;
using PrepareSnapCallBack = std::function<void(const PrepareSnapResponse &)>;
using SnapStartedHandler = std::function<SnapStartedResponse(const SnapStartedRequest &)>;
using SnapStartedCallBack = std::function<void(const SnapStartedResponse &)>;
using RefreshEnvCallback = std::function<void(void)>;
using SignalHandler = std::function<SignalResponse(const SignalRequest &)>;
using SignalCallBack = std::function<void(const SignalResponse &)>;
using HeartbeatHandler = std::function<HeartbeatResponse(const HeartbeatRequest &)>;
using HeartbeatCallBack = std::function<void(const HeartbeatResponse &)>;
using EventHandler = std::function<void(const std::shared_ptr<EventMessageSpec> &)>;
using EventCallBack = std::function<void(const std::shared_ptr<EventMessageSpec> &)>;

using CreateCallBack = std::function<void(const NotifyRequest &)>;
using CreateRespCallback = std::function<void(const CreateResponse &)>;
using CreateRespsCallback = std::function<void(const CreateResponses &)>;
using InvokeCallBack = std::function<void(const NotifyRequest &, const ErrorInfo &)>;
using CallResultCallBack = std::function<void(const CallResultAck &)>;
using ExitCallBack = std::function<void(const ExitResponse &)>;
using StateSaveCallBack = std::function<void(const StateSaveResponse &)>;
using StateLoadCallBack = std::function<void(const StateLoadResponse &)>;

struct FSIntfHandlers {
    CallHandler init = nullptr;
    CallHandler call = nullptr;
    CheckpointHandler checkpoint = nullptr;
    RecoverHandler recover = nullptr;
    ShutdownHandler shutdown = nullptr;
    PrepareSnapHandler prepareSnap = nullptr;
    SnapStartedHandler snapStarted = nullptr;
    RefreshEnvCallback refreshEnv = nullptr;
    SignalHandler signal = nullptr;
    HeartbeatHandler heartbeat = nullptr;
    EventHandler event = nullptr;
};

struct InterruptResponse {
    std::map<std::string, std::string> headers{{"Content-Type", "application/json"}, {"X-Log-Type", "base64"}};
    std::string billingDuration;
    std::string innerCode = "0";
    std::string invokeSummary;
    std::string logResult;
    int userFuncTime = 0;
    int executorTime = 0;
    std::map<std::string, std::string> body;

    InterruptResponse() = default;

    std::string toJson() const
    {
        nlohmann::json j;
        j["headers"] = headers;
        j["billingDuration"] = billingDuration;
        j["innerCode"] = innerCode;
        j["invokeSummary"] = invokeSummary;
        j["logResult"] = logResult;
        j["userFuncTime"] = userFuncTime;
        j["executorTime"] = executorTime;
        j["body"] = body;
        return j.dump();
    }
};

class FSIntf {
public:
    FSIntf() = default;
    explicit FSIntf(const FSIntfHandlers &handlers);
    virtual ~FSIntf();

    virtual ErrorInfo Start(const std::string &jobID, const std::string &instanceID = "",
                            const std::string &runtimeID = "", const std::string &functionName = "",
                            const SubscribeFunc &subScribeCb = nullptr) = 0;
    virtual void Stop(void) = 0;
    void ReceiveRequestLoop(void);
    virtual void GroupCreateAsync(const CreateRequests &reqs, CreateRespsCallback respCallback, CreateCallBack callback,
                                  int timeoutSec = -1) = 0;
    virtual void CreateAsync(const CreateRequest &req, CreateRespCallback respCallback, CreateCallBack callback,
                             int timeoutSec = -1) = 0;
    virtual void InvokeAsync(const std::shared_ptr<InvokeMessageSpec> &req, InvokeCallBack callback,
                             int timeoutSec = -1) = 0;
    virtual void CallResultAsync(const std::shared_ptr<CallResultMessageSpec> req, CallResultCallBack callback) = 0;
    virtual void KillAsync(const KillRequest &req, KillCallBack callback, int timeoutSec = -1) = 0;
    virtual void ExitAsync(const ExitRequest &req, ExitCallBack callback) = 0;
    virtual void StateSaveAsync(const StateSaveRequest &req, StateSaveCallBack callback) = 0;
    virtual void StateLoadAsync(const StateLoadRequest &req, StateLoadCallBack callback) = 0;
    virtual void CreateRGroupAsync(const CreateResourceGroupRequest &req, CreateResourceGroupCallBack callback,
                                   int timeoutSec = -1) = 0;
    virtual std::string GetServerVersion();
    virtual std::pair<ErrorInfo, std::string> GetNodeId();
    virtual std::pair<ErrorInfo, std::string> GetNodeIp();
    virtual void ReturnCallResult(const std::shared_ptr<CallResultMessageSpec> result, bool isCreate,
                                  CallResultCallBack callback);
    virtual HeartbeatResponse HandleHeartbeat(const HeartbeatRequest &hb);
    virtual void RemoveInsRtIntf(const std::string &instanceId) {}
    virtual ErrorInfo ReconnectProxyClient(const std::string &fsIp, int fsPort) { return ErrorInfo(); }
    void HandleCallRequest(const std::shared_ptr<CallMessageSpec> &req, CallCallBack callback);
    void ProcessCallRequest(const std::shared_ptr<CallMessageSpec> &req, CallCallBack callback);
    void HandleInterruptRequest(const google::protobuf::Map<std::string, std::string> &createOptions,
                                const std::shared_ptr<CallMessageSpec> &req, CallResponse &resp);
    std::string GetInterruptResponse(const google::protobuf::Map<std::string, std::string> &createOptions,
                                     const std::string &requestId, CallResponse &resp);
    std::shared_ptr<CallResultMessageSpec> BuildInterruptedCallResult(const std::shared_ptr<CallMessageSpec> &req,
                                                                      std::string interruptResponse);
    void HandleNotifyRequest(const NotifyRequest &req, std::function<NotifyResponse(void)> createOrInvokeCallback,
                             NotifyCallBack callback);
    void HandleCheckpointRequest(const CheckpointRequest &req, CheckpointCallBack callback);
    void HandleRecoverRequest(const RecoverRequest &req, RecoverCallBack callback);
    void HandleShutdownRequest(const ShutdownRequest &req, ShutdownCallBack callback);
    void HandlePrepareSnapRequest(const PrepareSnapRequest &req, PrepareSnapCallBack callback);
    void HandleSnapStartedRequest(const SnapStartedRequest &req, SnapStartedCallBack callback);
    void HandleSignalRequest(const SignalRequest &req, SignalCallBack callback);
    void HandleHeartbeatRequest(const HeartbeatRequest &req, HeartbeatCallBack callback);
    void HandleEventRequest(const std::shared_ptr<EventMessageSpec> &req);
    int WaitRequestEmpty(uint64_t gracePeriodSec);
    void SetInitialized();
    virtual void EventAsync(const std::shared_ptr<EventMessageSpec> &req, int timeoutSec = -1) {}
    virtual bool IsHealth() = 0;
    virtual void UpdateEventServerInfo(const std::string &ip, int port, const std::string &instaceId) {}
    virtual int GetSelfPort() const
    {
        return -1;
    }
    virtual std::string GetSelfIP() const
    {
        return "";
    }
    void SetAgentSessionManager(const std::shared_ptr<AgentSessionManager> &agentSessionManager)
    {
        agentSessionManager_ = agentSessionManager;
    }

protected:
    void Clear();
    std::string serverVersion_;
    std::string nodeIp_;
    std::string nodeId_;

private:
    bool AddProcessingRequestId(const std::string &requestId);
    bool DeleteProcessingRequestId(const std::string &requestId);
    bool ExistProcessingRequestId();
    bool cleared_{false};
    FSIntfHandlers handlers;
    bool syncHeartbeat;
    std::shared_ptr<AgentSessionManager> agentSessionManager_;
    ThreadPool callReceiver;
    ThreadPool noitfyExecutor;
    ThreadPool checkpointRecoverExecutor;
    ThreadPool shutdownExecutor;
    ThreadPool signalExecutor;
    ThreadPool heartbeatExecutor;
    ThreadPool responseReceiver;
    ThreadPool eventExecutor;

    absl::Mutex mu;
    absl::CondVar cv_;
    std::atomic<bool> isShutdownDone{false};
    std::unordered_set<std::string> processingRequestIds ABSL_GUARDED_BY(mu);

    class InstanceStatus {
    public:
        enum InstanceState : int {
            STARTED = 0,
            INITIALIZING,
            INITIALIZING_FAILURE,
            INITIALIZED,
            SHUTTING_DOWN,
            SHUTDOWN,
            CHECKPOINTING,
            INTERRUPT_CHECKPOITING,
        };

        InstanceStatus() : state(STARTED)
        {
            err = std::make_pair(common::ERR_NONE, "");
        }
        ~InstanceStatus() = default;

        bool SetInitializing()
        {
            absl::MutexLock lock(&this->mu);
            if (state == STARTED) {
                state = INITIALIZING;
            }
            return state == INITIALIZING;
        }

        void SetInitialized()
        {
            absl::MutexLock lock(&this->mu);
            if (state == INITIALIZING) {
                state = INITIALIZED;
            }
            n.Notify();
        }

        void SetInitializingFailure(common::ErrorCode code, const std::string &msg)
        {
            absl::MutexLock lock(&this->mu);
            if (state == INITIALIZING) {
                state = INITIALIZING_FAILURE;
                err = std::make_pair(code, msg);
            }
            n.Notify();
        }

        bool WaitInitialized(void)
        {
            n.WaitForNotification();
            absl::ReaderMutexLock lock(&this->mu);
            return state == INITIALIZED;
        }

        bool IsInitialized()
        {
            absl::ReaderMutexLock lock(&this->mu);
            return state == INITIALIZED;
        }

        std::pair<common::ErrorCode, std::string> GetErrorInfo(void) const
        {
            return err;
        }

        bool SetShuttingDown()
        {
            absl::MutexLock lock(&this->mu);
            if (err.first == common::ErrorCode::ERR_NONE) {
                err = std::make_pair(common::ErrorCode::ERR_INSTANCE_EXITED, "instance has already exited");
            }
            if (state != SHUTDOWN) {
                state = SHUTTING_DOWN;
            }
            return state == SHUTTING_DOWN;
        }

        bool SetCheckpointing()
        {
            absl::MutexLock lock(&this->mu);
            if (state != CHECKPOINTING) {
                state = CHECKPOINTING;
                err = std::make_pair(common::ERR_INNER_SYSTEM_ERROR, "instance is checkpointing");
            }
            return state == CHECKPOINTING;
        }

        void SetCheckpointed()
        {
            absl::MutexLock lock(&this->mu);
            if (state == CHECKPOINTING || state == INTERRUPT_CHECKPOITING) {
                state = INITIALIZED;
            }
            err = std::make_pair(common::ERR_NONE, "");
        }

        void InterruptCheckpointing()
        {
            absl::MutexLock lock(&this->mu);
            if (state == CHECKPOINTING) {
                state = INTERRUPT_CHECKPOITING;
                err = std::make_pair(common::ERR_INNER_SYSTEM_ERROR, "checkpoint is interrupted");
            }
        }

        bool IsCheckpointingInterrupted()
        {
            absl::MutexLock lock(&this->mu);
            return state == INTERRUPT_CHECKPOITING;
        }

        void SetShutdown()
        {
            absl::MutexLock lock(&this->mu);
            if (state == SHUTTING_DOWN) {
                state = SHUTDOWN;
            }
            shutdownNotify_.Notify();
        }

        bool WaitShutdown()
        {
            shutdownNotify_.WaitForNotification();
            absl::ReaderMutexLock lock(&this->mu);
            return state == SHUTDOWN;
        }

    private:
        absl::Mutex mu;
        InstanceState state;
        std::pair<common::ErrorCode, std::string> err;
        absl::Notification n;
        absl::Notification shutdownNotify_;
    };

    InstanceStatus status;
};
}  // namespace Libruntime
}  // namespace YR