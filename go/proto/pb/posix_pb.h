/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
#ifndef POSIX_PB_H
#define POSIX_PB_H
#include "posix/core_service.pb.h"
#include "posix/runtime_service.pb.h"
#include "posix/inner_service.pb.h"
#include "posix/runtime_rpc.pb.h"
#include "posix/common.pb.h"
#include "posix/affinity.pb.h"

namespace functionsystem {
    using InvokeRequest = core_service::InvokeRequest;
    using InvokeResponse = core_service::InvokeResponse;
    using CreateRequest = core_service::CreateRequest;
    using CreateResponse = core_service::CreateResponse;
    using CallResult = core_service::CallResult;
    using CallResultAck = core_service::CallResultAck;
    using StateSaveRequest = core_service::StateSaveRequest;
    using StateSaveResponse = core_service::StateSaveResponse;
    using StateLoadRequest = core_service::StateLoadRequest;
    using StateLoadResponse = core_service::StateLoadResponse;
    using KillRequest = core_service::KillRequest;
    using KillResponse = core_service::KillResponse;
    using ExitRequest = core_service::ExitRequest;
    using ExitResponse = core_service::ExitResponse;
    using GroupOptions = core_service::GroupOptions;
    using CreateRequests = core_service::CreateRequests;
    using CreateResponses = core_service::CreateResponses;
    using SharedStreamMsg = std::shared_ptr<runtime_rpc::StreamingMessage>;
    using InstanceRange = core_service::InstanceRange;
    using CreateResourceGroupRequest = core_service::CreateResourceGroupRequest;
    using CreateResourceGroupResponse = core_service::CreateResourceGroupResponse;
}

namespace internal {
    using ForwardCallRequest = inner_service::ForwardCallRequest;
    using ForwardCallResponse = inner_service::ForwardCallResponse;
    using ForwardCallResultRequest = inner_service::ForwardCallResultRequest;
    using ForwardCallResultResponse = inner_service::ForwardCallResultResponse;
    using ForwardKillRequest = inner_service::ForwardKillRequest;
    using ForwardKillResponse = inner_service::ForwardKillResponse;
    using RouteCallRequest = inner_service::RouteCallRequest;
}

namespace runtime {
    using CallRequest = runtime_service::CallRequest;
    using CallResponse = runtime_service::CallResponse;
    using NotifyRequest = runtime_service::NotifyRequest;
    using NotifyResponse = runtime_service::NotifyResponse;
    using SignalRequest = runtime_service::SignalRequest;
    using SignalResponse = runtime_service::SignalResponse;
    using ShutdownRequest = runtime_service::ShutdownRequest;
    using ShutdownResponse = runtime_service::ShutdownResponse;
    using HeartbeatRequest = runtime_service::HeartbeatRequest;
    using HeartbeatResponse = runtime_service::HeartbeatResponse;
    using CheckpointRequest = runtime_service::CheckpointRequest;
    using CheckpointResponse = runtime_service::CheckpointResponse;
    using RecoverRequest = runtime_service::RecoverRequest;
    using RecoverResponse = runtime_service::RecoverResponse;
}

namespace common {
    using ErrorCode = common::ErrorCode;
    using HealthCheckCode = common::HealthCheckCode;
    using Arg = common::Arg;
    using HeteroDeviceInfo = common::HeteroDeviceInfo;
    using ServerInfo = common::ServerInfo;
    using FunctionGroupRunningInfo = common::FunctionGroupRunningInfo;
    using ResourceGroupSpec = common::ResourceGroupSpec;
}

namespace affinity {
    using LabelIn = affinity::LabelIn;
    using LabelNotIn = affinity::LabelNotIn;
    using LabelExists = affinity::LabelExists;
    using LabelDoesNotExist = affinity::LabelDoesNotExist;
    using LabelOperator = affinity::LabelOperator;
    using SubCondition = affinity::SubCondition;
    using Condition = affinity::Condition;
    using Selector = affinity::Selector;
    using AffinityType = affinity::AffinityType;
    using AffinityScope = affinity::AffinityScope;
    using ResourceAffinity = affinity::ResourceAffinity;
    using InstanceAffinity = affinity::InstanceAffinity;
    using Affinity = affinity::Affinity;
}

#endif