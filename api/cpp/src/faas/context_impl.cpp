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

#include "api/cpp/src/faas/context_impl.h"

namespace Function {

ContextImpl::ContextImpl(std::shared_ptr<ContextInvokeParams> &invokeParams, std::shared_ptr<ContextEnv> &contextEnv)
{
    this->contextInvokeParams_ = invokeParams;
    this->contextEnv_ = contextEnv;
}

ContextImpl::ContextImpl(const ContextImpl &contextImpl)
{
    this->stateId_ = contextImpl.stateId_;
    this->state_ = contextImpl.state_;
    this->property_ = contextImpl.property_;
    this->funcStartTime_ = contextImpl.funcStartTime_;
    this->contextInvokeParams_ = contextImpl.contextInvokeParams_;
    this->contextEnv_ = contextImpl.contextEnv_;
}

ContextImpl &ContextImpl::operator=(const ContextImpl &contextImpl)
{
    this->stateId_ = contextImpl.stateId_;
    this->state_ = contextImpl.state_;
    this->property_ = contextImpl.property_;
    this->funcStartTime_ = contextImpl.funcStartTime_;
    this->contextInvokeParams_ = contextImpl.contextInvokeParams_;
    this->contextEnv_ = contextImpl.contextEnv_;
    return *this;
}

const std::string ContextImpl::GetTraceId() const
{
    return contextInvokeParams_->GetTraceId();
}

const std::string ContextImpl::GetInvokeId() const
{
    return contextInvokeParams_->GetInvokeId();
}

const FunctionLogger &ContextImpl::GetLogger()
{
    logger_.SetInvokeID(GetInvokeId());
    logger_.SetTraceID(GetTraceId());
    return logger_;
}

const std::string ContextImpl::GetInstanceId() const
{
    return stateId_;
}

const std::string ContextImpl::GetInstanceLabel() const
{
    return contextEnv_->GetInstanceLabel();
}

void ContextImpl::SetStateId(const std::string &stateId)
{
    std::lock_guard<std::mutex> lock(stateMtx_);
    this->stateId_ = stateId;
}

const std::string ContextImpl::GetState() const
{
    return state_;
}

void ContextImpl::SetState(const std::string &state)
{
    std::lock_guard<std::mutex> lock(stateMtx_);
    this->state_ = state;
}

const std::string ContextImpl::GetInvokeProperty() const
{
    return property_;
}

void ContextImpl::SetInvokeProperty(const std::string &prop)
{
    this->property_ = prop;
}

const std::string ContextImpl::GetRequestID() const
{
    return contextInvokeParams_->GetRequestId();
}

const std::string ContextImpl::GetUserData(std::string key) const
{
    return contextEnv_->GetUserData(key);
}

const std::string ContextImpl::GetFunctionName() const
{
    return contextEnv_->GetFunctionName();
}

int ContextImpl::GetRunningTimeInSeconds() const
{
    return contextEnv_->GetRunningTimeInSeconds();
}

int ContextImpl::GetRemainingTimeInMilliSeconds() const
{
    return 0;
}

const std::string ContextImpl::GetVersion() const
{
    return contextEnv_->GetVersion();
}

int ContextImpl::GetMemorySize() const
{
    return contextEnv_->GetMemorySize();
}

int ContextImpl::GetCPUNumber() const
{
    return contextEnv_->GetCPUNumber();
}

const std::string ContextImpl::GetProjectID() const
{
    return contextEnv_->GetProjectID();
}

const std::string ContextImpl::GetPackage() const
{
    return contextEnv_->GetFuncPackage();
}

const std::string ContextImpl::GetAccessKey() const
{
    return contextInvokeParams_->GetAccessKey();
}

const std::string ContextImpl::GetAlias() const
{
    return contextInvokeParams_->GetAlias();
}

const std::string ContextImpl::GetSecretKey() const
{
    return contextInvokeParams_->GetSecretKey();
}

const std::string ContextImpl::GetSecurityAccessKey() const
{
    return contextInvokeParams_->GetSecurityAccessKey();
}

const std::string ContextImpl::GetSecuritySecretKey() const
{
    return contextInvokeParams_->GetSecuritySecretKey();
}

const std::string ContextImpl::GetToken() const
{
    return contextInvokeParams_->GetToken();
}

void ContextImpl::SetFuncStartTime(const long startTime)
{
    this->funcStartTime_ = startTime;
}
}  // namespace Function
