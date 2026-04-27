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

#pragma once
#include <grpcpp/grpcpp.h>
#include "grpcpp/security/server_credentials.h"

#include "src/libruntime/utils/security.h"
#include "src/dto/config.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/fsclient/fs_intf_manager.h"
#include "src/libruntime/fsclient/protobuf/runtime_rpc.grpc.pb.h"
#include "src/utility/timer_worker.h"

namespace YR {
namespace Libruntime {
using StreamingMessage = ::runtime_rpc::StreamingMessage;
using YR::utility::TimerWorker;
class FSIntfGrpcServerReaderWriter;
class GrpcPosixService : public runtime_rpc::RuntimeRPC::Service,
                         public std::enable_shared_from_this<GrpcPosixService> {
public:
    GrpcPosixService(const std::string &instanceID, const std::string &runtimeID, const std::string &listeningIpAddr,
                     int selfPort, const std::shared_ptr<TimerWorker> &timerWorker,
                     const std::shared_ptr<absl::Notification> &notification,
                     const std::shared_ptr<FSIntfManager> &fsIntfManager, std::shared_ptr<Security> security,
                     bool enableDirectCall = false, bool enableEvent = false)
        : instanceID(instanceID),
          runtimeID(runtimeID),
          listeningIpAddr(listeningIpAddr),
          selfPort(selfPort),
          timerWorker(timerWorker),
          notification(notification),
          fsIntfMgr(fsIntfManager),
          security_(security),
          enableEvent_(enableEvent),
          enableDirectCall_(enableDirectCall)
    {
    }
    ~GrpcPosixService() override
    {
        this->Stop();
    };
    ErrorInfo Start();

    void Stop();

    grpc::Status MessageStream(grpc::ServerContext *context,
                               grpc::ServerReaderWriter<StreamingMessage, StreamingMessage> *stream) override;

    grpc::Status BatchMessageStream(
        grpc::ServerContext *context,
        grpc::ServerReaderWriter<BatchStreamingMessage, BatchStreamingMessage> *stream) override;

    inline void RegisterFSHandler(const std::unordered_map<BodyCase, MsgHdlr> &fsMsgHdlr)
    {
        this->fsMsgHdlrs = fsMsgHdlr;
    }

    inline void RegisterRTHandler(const std::unordered_map<BodyCase, MsgHdlr> &rtMsgHdlr)
    {
        this->rtMsgHdlrs = rtMsgHdlr;
    }

    inline void RegisterEventHdlrs(const std::unordered_map<BodyCase, MsgHdlr> &eventMsgHdlrs)
    {
        this->eventMsgHdlrs = eventMsgHdlrs;
    }

    std::shared_ptr<grpc::ServerCredentials> GetCreds()
    {
        return YR::GetServerCreds(security_);
    }

    inline int GetListeningPort()
    {
        return listeningPort;
    }

    inline void RegisterResendCallback(const std::function<void(const std::string &)> &resendCb)
    {
        this->resendCb = resendCb;
    }

    inline void RegisterDisconnectedCallback(const std::function<void(const std::string &)> &disconnectedCb)
    {
        this->disconnectedCb = disconnectedCb;
    }

private:
    bool CompareInstanceID(grpc::ServerContext *context) const;
    grpc::Status HandleMessageStreamFromFS(grpc::ServerContext *context,
                                           grpc::ServerReaderWriter<StreamingMessage, StreamingMessage> *stream);

    grpc::Status HandleDirectStream(
        grpc::ServerContext *context,
        grpc::ServerReaderWriter<StreamingMessage, StreamingMessage> *stream,
        grpc::ServerReaderWriter<BatchStreamingMessage, BatchStreamingMessage> *batchStream);
    void StartRead(const std::string &remote, const std::shared_ptr<FSIntfGrpcServerReaderWriter> &fsIntf,
                   int disconnectedTimeout);
    void StartDisconnectTimer(const std::string &remote, int disconnectedTimeout);
    void StopDisconnectTimer(const std::string &remote);
    bool VerifySrcWihtAkSk(const std::multimap<grpc::string_ref, grpc::string_ref> &metadata);

    std::string instanceID;
    std::string runtimeID;
    std::string listeningIpAddr;
    int selfPort;
    int listeningPort;
    std::shared_ptr<TimerWorker> timerWorker;
    std::shared_ptr<absl::Notification> notification;
    std::shared_ptr<FSIntfManager> fsIntfMgr;
    std::unique_ptr<grpc::Server> server;
    mutable absl::Mutex disconnectedMu;
    std::unordered_map<std::string, std::shared_ptr<YR::utility::Timer>> disconnectCallbackTimers ABSL_GUARDED_BY(mu);
    std::unordered_map<BodyCase, MsgHdlr> fsMsgHdlrs;
    std::unordered_map<BodyCase, MsgHdlr> rtMsgHdlrs;
    std::unordered_map<BodyCase, MsgHdlr> eventMsgHdlrs;
    std::function<void(const std::string &)> resendCb;
    std::function<void(const std::string &)> disconnectedCb;
    int rtDisconnectedTimeout = RT_DISCONNECT_TIMEOUT_MS;
    int fsDisconnectedTimeout = DISCONNECT_TIMEOUT_MS;
    std::atomic<bool> stopped{false};
    std::atomic<bool> fsConnected{false};
    std::shared_ptr<Security> security_;
    bool enableEvent_{false};
    bool enableDirectCall_{false};
};

}  // namespace Libruntime
}  // namespace YR