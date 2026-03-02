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

#include "fs_intf_grpc_reader_writer.h"
#include "src/utility/timer_worker.h"

namespace YR {
namespace Libruntime {
using StreamingMessage = ::runtime_rpc::StreamingMessage;
using YR::utility::TimerWorker;
using BodyCase = ::runtime_rpc::StreamingMessage::BodyCase;

struct SteamRW {
    grpc::ServerReaderWriter<StreamingMessage, StreamingMessage> *stream;
    grpc::ServerReaderWriter<BatchStreamingMessage, BatchStreamingMessage> *batchStream;
};

class FSIntfGrpcServerReaderWriter : public FSIntfGrpcReaderWriter {
public:
    FSIntfGrpcServerReaderWriter(const std::string &srcInstance, const std::string &dstInstance,
                                 const std::string &runtimeID, grpc::ServerContext *context, const SteamRW &rw)
        : FSIntfGrpcReaderWriter(srcInstance, dstInstance, runtimeID),
          stream_(rw.stream),
          batchStream_(rw.batchStream),
          context_(context)
    {
    }

    ~FSIntfGrpcServerReaderWriter()
    {
        this->Stop();
    }
    void PreStart() override;
    ErrorInfo Start() override;
    void Stop() override;
    bool GrpcRead(const std::shared_ptr<StreamingMessage> &message) override;
    bool GrpcWrite(const std::shared_ptr<StreamingMessage> &request) override;
    bool GrpcBatchRead(const std::shared_ptr<BatchStreamingMessage> &message) override;
    bool GrpcBatchWrite(const std::shared_ptr<BatchStreamingMessage> &request) override;
    bool IsBatched() override;
    bool IsHealth() override;
    bool IsSameDstAddr(const std::string &dstIp, const int &dstPort) override { return true; }

private:
    void ClearStream();
    grpc::ServerReaderWriter<StreamingMessage, StreamingMessage> *stream_ = nullptr;
    grpc::ServerReaderWriter<BatchStreamingMessage, BatchStreamingMessage> *batchStream_ = nullptr;
    grpc::ServerContext *context_ = nullptr;
};

}  // namespace Libruntime
}  // namespace YR
