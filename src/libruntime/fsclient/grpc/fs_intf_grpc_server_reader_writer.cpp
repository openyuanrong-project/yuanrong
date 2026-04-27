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

#include <future>

#include "fs_intf_grpc_server_reader_writer.h"
#include "src/libruntime/utils/constants.h"
#include "src/libruntime/utils/utils.h"

namespace YR {
namespace Libruntime {

bool FSIntfGrpcServerReaderWriter::IsBatched()
{
    return batchStream_ != nullptr;
}

void FSIntfGrpcServerReaderWriter::PreStart()
{
    isConnect_.store(true);
}

ErrorInfo FSIntfGrpcServerReaderWriter::Start()
{
    RecvFunc();
    isConnect_.store(false);
    return {};
}

void FSIntfGrpcServerReaderWriter::Stop()
{
    bool expected = false;
    bool exchanged = abnormal_.compare_exchange_strong(expected, true);
    if (!exchanged) {
        return;
    }
    // trigger to stop, avoid race condition
    FSIntfGrpcReaderWriter::Stop();
    if (!isConnect_.load()) {
        // to avoid context_ early end of life cycle by reader thread
        ClearStream();
        return;
    }
    if (this->context_ != nullptr) {
        this->context_->TryCancel();
    }
    ClearStream();
}

bool FSIntfGrpcServerReaderWriter::GrpcRead(const std::shared_ptr<StreamingMessage> &message)
{
    return stream_->Read(message.get());
}

bool FSIntfGrpcServerReaderWriter::GrpcWrite(const std::shared_ptr<StreamingMessage> &request)
{
    if (!Available() || this->stream_ == nullptr) {
        YRLOG_DEBUG("stream is nullptr while writing message");
        return false;
    }
    return stream_->Write(*request.get());
}

bool FSIntfGrpcServerReaderWriter::GrpcBatchRead(const std::shared_ptr<BatchStreamingMessage> &message)
{
    return batchStream_->Read(message.get());
}

bool FSIntfGrpcServerReaderWriter::GrpcBatchWrite(const std::shared_ptr<BatchStreamingMessage> &request)
{
    if (!FSIntfGrpcReaderWriter::Available()) {
        YRLOG_WARN("client is not connected");
        return false;
    }
    if (batchStream_ == nullptr) {
        YRLOG_WARN("stream has reset nullptr");
        return false;
    }
    return batchStream_->Write(*request.get());
}

bool FSIntfGrpcServerReaderWriter::IsHealth()
{
    return true;
}

void FSIntfGrpcServerReaderWriter::ClearStream()
{
    this->stream_ = nullptr;
    this->batchStream_ = nullptr;
    this->context_ = nullptr;
}

}  // namespace Libruntime
}  // namespace YR
