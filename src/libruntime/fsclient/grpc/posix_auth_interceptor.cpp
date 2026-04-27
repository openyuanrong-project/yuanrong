/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

#include "src/libruntime/fsclient/grpc/posix_auth_interceptor.h"
#include "src/libruntime/utils/grpc_utils.h"
#include "src/utility/logger/logger.h"
namespace YR {
namespace Libruntime {

void PosixAuthInterceptor::InterceptCommon(::grpc::experimental::InterceptorBatchMethods *methods)
{
    if (stopped) {
        return;
    }
    if (methods->QueryInterceptionHookPoint(::grpc::experimental::InterceptionHookPoints::PRE_SEND_MESSAGE)) {
        const auto *message = dynamic_cast<const StreamingMessage *>(
            static_cast<const ::google::protobuf::Message *>(methods->GetSendMessage()));
        if (message == nullptr) {
            methods->Proceed();
            return;
        }
        StreamingMessage signedMessage;
        if (message->has_heartbeatrsp()) {
            methods->Proceed();
            return;
        }
        signedMessage.CopyFrom(*message);
        if (SignWithAKSK(signedMessage)) {
            methods->ModifySendMessage(&signedMessage);
            methods->Proceed();
        } else {
            YRLOG_ERROR("failed to sign message: {}, message id is {}", message->DebugString(), message->messageid());
            methods->FailHijackedSendMessage();
        }
        return;
    }

    if (methods->QueryInterceptionHookPoint(::grpc::experimental::InterceptionHookPoints::POST_RECV_MESSAGE)) {
        auto *message =
            dynamic_cast<StreamingMessage *>(static_cast<::google::protobuf::Message *>(methods->GetRecvMessage()));
        if (message == nullptr) {
            methods->Proceed();
            return;
        }
        if (message->has_heartbeatreq()) {
            methods->Proceed();
            return;
        }
        if (!VerifyAKSK(*message)) {
            YRLOG_ERROR("failed to verify message: {}, message id is {}", message->DebugString(), message->messageid());
            // clear message, if return directly, the server will be blocked
            message->Clear();
        }
    }
    methods->Proceed();
}

void PosixClientAuthInterceptor::Intercept(::grpc::experimental::InterceptorBatchMethods *clientMethods)
{
    InterceptCommon(clientMethods);
}

void PosixServerAuthInterceptor::Intercept(::grpc::experimental::InterceptorBatchMethods *methods)
{
    InterceptCommon(methods);
}

bool PosixAuthInterceptor::VerifyAKSK(const StreamingMessage &message)
{
    if (security_ != nullptr && !security_->IsFsAuthEnable()) {
        return true;
    }

    auto tenantAccessKey = message.metadata().find(ACCESS_KEY);
    if (tenantAccessKey == message.metadata().end() || tenantAccessKey->second.empty()) {
        YRLOG_ERROR("failed to verify message: {}, failed to find access_key in meta-data, message id is {}",
                    message.DebugString(), message.messageid());
        return false;
    }

    std::string ak;
    SensitiveValue sk;
    security_->GetAKSK(ak, sk);

    if (ak.empty() || sk.Empty()) {
        YRLOG_ERROR(
            "failed to verify message {}, ak or sk is emptgy, failed to get cred from security, message id is {}",
            message.DebugString(), message.messageid());
        return false;
    }

    return VerifyStreamingMessage(ak, sk, message);
}

bool PosixAuthInterceptor::SignWithAKSK(StreamingMessage &message)
{
    if (security_ != nullptr && !security_->IsFsAuthEnable()) {
        YRLOG_WARN("is fs auth enable: {}, no need sign with ak sk", security_->IsFsAuthEnable());
        return true;
    }

    std::string ak;
    SensitiveValue sk;
    security_->GetAKSK(ak, sk);

    if (ak.empty() || sk.Empty()) {
        YRLOG_ERROR("failed to sign message: {}, failed to get cred from iam, message id is {}",
                    message.DebugString(), message.messageid());
        return false;
    }

    if (!SignStreamingMessage(ak, sk, message)) {
        YRLOG_ERROR("failed to sign message: {}, message id is {}", message.DebugString(), message.messageid());
        return false;
    }
    return true;
}
}  // namespace Libruntime
}  // namespace YR