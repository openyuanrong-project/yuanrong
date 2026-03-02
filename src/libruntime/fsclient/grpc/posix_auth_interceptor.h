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

#pragma once
#include "src/libruntime/fsclient/fs_intf_manager.h"
#include "src/libruntime/fsclient/protobuf/runtime_rpc.grpc.pb.h"
#include "src/libruntime/utils/security.h"

namespace YR {
namespace Libruntime {
const std::string ACCESS_KEY = "access_key";
const std::string TIMESTAMP = "timestamp";
const std::string SIGNATURE = "signature";
const std::string INSTANCE_ID = "instance_id";
const std::string RUNTIME_ID = "runtime_id";
using StreamingMessage = ::runtime_rpc::StreamingMessage;

class PosixAuthInterceptor : public ::grpc::experimental::Interceptor {
public:
    PosixAuthInterceptor() = default;
    virtual ~PosixAuthInterceptor()
    {
        stopped.store(true);
    }

    bool VerifyAKSK(const StreamingMessage &message);

    bool SignWithAKSK(StreamingMessage &message);

    void InterceptCommon(::grpc::experimental::InterceptorBatchMethods *methods);

    void RegisterSecurity(std::shared_ptr<Security> security)
    {
        this->security_ = security;
    }

protected:
    std::shared_ptr<Security> security_;
    std::atomic<bool> stopped{false};
};

class PosixClientAuthInterceptor : public PosixAuthInterceptor {
public:
    explicit PosixClientAuthInterceptor(::grpc::experimental::ClientRpcInfo *info) : info_(info) {}
    ~PosixClientAuthInterceptor() override
    {
        stopped.store(true);
    }
    void Intercept(::grpc::experimental::InterceptorBatchMethods *methods) override;

private:
    ::grpc::experimental::ClientRpcInfo *info_;
};

class PosixServerAuthInterceptor : public PosixAuthInterceptor {
public:
    explicit PosixServerAuthInterceptor(::grpc::experimental::ServerRpcInfo *info) : info_(info) {}
    ~PosixServerAuthInterceptor() override
    {
        stopped.store(true);
    }
    void Intercept(::grpc::experimental::InterceptorBatchMethods *methods) override;

private:
    ::grpc::experimental::ServerRpcInfo *info_;
};

class PosixClientAuthInterceptorFactory : public ::grpc::experimental::ClientInterceptorFactoryInterface {
public:
    PosixClientAuthInterceptorFactory() = default;

    ~PosixClientAuthInterceptorFactory() override = default;

    void RegisterSecurity(std::shared_ptr<Security> security)
    {
        this->security_ = security;
    }

    ::grpc::experimental::Interceptor *CreateClientInterceptor(::grpc::experimental::ClientRpcInfo *info) override
    {
        auto interceptor = new PosixClientAuthInterceptor(info);
        interceptor->RegisterSecurity(security_);
        return interceptor;
    }

private:
    std::shared_ptr<Security> security_;
};

class PosixServerAuthInterceptorFactory : public ::grpc::experimental::ServerInterceptorFactoryInterface {
public:
    PosixServerAuthInterceptorFactory() = default;

    ~PosixServerAuthInterceptorFactory() override = default;

    void RegisterSecurity(std::shared_ptr<Security> security)
    {
        this->security_ = security;
    }

    ::grpc::experimental::Interceptor *CreateServerInterceptor(::grpc::experimental::ServerRpcInfo *info) override
    {
        auto interceptor = new PosixServerAuthInterceptor(info);
        interceptor->RegisterSecurity(security_);
        return interceptor;
    }

private:
    std::shared_ptr<Security> security_;
};

}  // namespace Libruntime
}  // namespace YR
