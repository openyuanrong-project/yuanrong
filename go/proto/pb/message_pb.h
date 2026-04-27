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
#ifndef MESSAGE_PB_H
#define MESSAGE_PB_H

#include "posix/message.pb.h"

namespace functionsystem::messages {
using ResourceView = ::messages::ResourceView;
using ScheduleRequest = ::messages::ScheduleRequest;
using PluginContext = ::messages::PluginContext;
using ScheduleResponse = ::messages::ScheduleResponse;
using ScheduleResult = ::messages::ScheduleResult;

using StaticFunctionChangeRequest = ::messages::StaticFunctionChangeRequest;
using StaticFunctionChangeResponse= ::messages::StaticFunctionChangeResponse;

using Register = ::messages::Register;
using Registered = ::messages::Registered;
using ScheduleTopology = ::messages::ScheduleTopology;
using NotifySchedAbnormalRequest = ::messages::NotifySchedAbnormalRequest;
using NotifySchedAbnormalResponse = ::messages::NotifySchedAbnormalResponse;

using Register = ::messages::Register;
using Registered = ::messages::Registered;
using ScheduleTopology = ::messages::ScheduleTopology;
using NotifySchedAbnormalRequest = ::messages::NotifySchedAbnormalRequest;

using NotifyWorkerStatusRequest = ::messages::NotifyWorkerStatusRequest;
using NotifyWorkerStatusResponse = ::messages::NotifyWorkerStatusResponse;
using UpdateNodeTaintRequest = ::messages::UpdateNodeTaintRequest;
using UpdateNodeTaintResponse = ::messages::UpdateNodeTaintResponse;

using CreateAgentRequest = ::messages::CreateAgentRequest;
using CreateAgentResponse = ::messages::CreateAgentResponse;

using DeployInstanceRequest = ::messages::DeployInstanceRequest;
using DeployInstanceResponse = ::messages::DeployInstanceResponse;
using StartInstanceRequest = ::messages::StartInstanceRequest;
using StartInstanceResponse = ::messages::StartInstanceResponse;
using Layer = ::messages::Layer;
using FuncDeploySpec = ::messages::FuncDeploySpec;
using FuncMount = ::messages::FuncMount;
using FuncMountUser = ::messages::FuncMountUser;
using FuncMountConfig = ::messages::FuncMountConfig;
using DeployRequest = ::messages::DeployRequest;
using KillInstanceRequest = ::messages::KillInstanceRequest;
using KillInstanceResponse = ::messages::KillInstanceResponse;
using UpdateCredRequest = ::messages::UpdateCredRequest;
using UpdateCredResponse = ::messages::UpdateCredResponse;

using RegisterRuntimeManagerRequest = ::messages::RegisterRuntimeManagerRequest;
using RegisterRuntimeManagerResponse = ::messages::RegisterRuntimeManagerResponse;

using RuntimeInstanceInfo = ::messages::RuntimeInstanceInfo;
using RuntimeConfig = ::messages::RuntimeConfig;
using DeploymentConfig = ::messages::DeploymentConfig;
using CodePackageThresholds = ::messages::CodePackageThresholds;

using DeployRequest = ::messages::DeployRequest;
using DeployResult = ::messages::DeployResult;
using DeployDuration = ::messages::DeployDuration;

using StopInstanceRequest = ::messages::StopInstanceRequest;
using StopInstanceResponse = ::messages::StopInstanceResponse;

using UpdateResourcesRequest = ::messages::UpdateResourcesRequest;
using UpdateInstanceStatusRequest = ::messages::UpdateInstanceStatusRequest;
using UpdateInstanceStatusResponse = ::messages::UpdateInstanceStatusResponse;

using QueryInstanceStatusRequest = ::messages::QueryInstanceStatusRequest;
using QueryInstanceStatusResponse = ::messages::QueryInstanceStatusResponse;

using InstanceStatusInfo = ::messages::InstanceStatusInfo;

using UpdateAgentStatusRequest = ::messages::UpdateAgentStatusRequest;
using UpdateAgentStatusResponse = ::messages::UpdateAgentStatusResponse;

using UpdateRuntimeStatusRequest = ::messages::UpdateRuntimeStatusRequest;
using UpdateRuntimeStatusResponse = ::messages::UpdateRuntimeStatusResponse;

using SchedulerNode = ::messages::SchedulerNode;
using FuncAgentRegisInfo = ::messages::FuncAgentRegisInfo;
using FuncAgentRegisInfoCollection = ::messages::FuncAgentRegisInfoCollection;

using CreateAgentRequest = ::messages::CreateAgentRequest;
using CreateAgentResponse = ::messages::CreateAgentResponse;

using CleanStatusRequest = ::messages::CleanStatusRequest;
using CleanStatusResponse = ::messages::CleanStatusResponse;

using ForwardKillRequest = ::messages::ForwardKillRequest;
using ForwardKillResponse = ::messages::ForwardKillResponse;

using EvictAgentRequest = ::messages::EvictAgentRequest;
using EvictAgentAck = ::messages::EvictAgentAck;
using EvictAgentResult = ::messages::EvictAgentResult;
using EvictAgentResultAck = ::messages::EvictAgentResultAck;
using UpdateLocalStatusRequest = ::messages::UpdateLocalStatusRequest;
using UpdateLocalStatusResponse = ::messages::UpdateLocalStatusResponse;

using QueryAgentInfoRequest = ::messages::QueryAgentInfoRequest;
using QueryAgentInfoResponse = ::messages::QueryAgentInfoResponse;
using ExternalAgentInfo = ::messages::ExternalAgentInfo;
using ExternalQueryAgentInfoResponse = ::messages::ExternalQueryAgentInfoResponse;
using FunctionSystemStatus = ::messages::FunctionSystemStatus;
using RuleType = ::messages::NetworkIsolationRuleType;
using SetNetworkIsolationRequest = ::messages::SetNetworkIsolationRequest;
using SetNetworkIsolationResponse = ::messages::SetNetworkIsolationResponse;
using QueryResourcesInfoRequest = ::messages::QueryResourcesInfoRequest;
using QueryResourcesInfoResponse = ::messages::QueryResourcesInfoResponse;

using MetaStoreRequest = ::messages::MetaStoreRequest;
using MetaStoreResponse = ::messages::MetaStoreResponse;

using GetAndWatchResponse = ::messages::GetAndWatchResponse;

using GroupInfo = ::messages::GroupInfo;
using GroupResponse = ::messages::GroupResponse;
using KillGroup = ::messages::KillGroup;
using KillGroupResponse = ::messages::KillGroupResponse;
using DeletePodRequest = ::messages::DeletePodRequest;
using DeletePodResponse = ::messages::DeletePodResponse;

using ResourceGroupInfo = ::messages::ResourceGroupInfo;
using BundleInfo = ::messages::BundleInfo;
using BundleCollection = ::messages::BundleCollection;
using RemoveBundleRequest = ::messages::RemoveBundleRequest;
using RemoveBundleResponse = ::messages::RemoveBundleResponse;

using GetTokenRequest = ::messages::GetTokenRequest;
using GetTokenResponse = ::messages::GetTokenResponse;

using GetAKSKByTenantIDRequest = ::messages::GetAKSKByTenantIDRequest;
using GetAKSKByAKRequest = ::messages::GetAKSKByAKRequest;
using GetAKSKResponse = ::messages::GetAKSKResponse;

using CancelType = ::messages::CancelType;
using CancelSchedule = ::messages::CancelSchedule;
using CancelScheduleResponse = ::messages::CancelScheduleResponse;

using QueryInstancesInfoRequest = ::messages::QueryInstancesInfoRequest;
using QueryInstancesInfoResponse = ::messages::QueryInstancesInfoResponse;

using DebugInstanceInfo = ::messages::DebugInstanceInfo;
using QueryDebugInstanceInfosRequest = ::messages::QueryDebugInstanceInfosRequest;
using QueryDebugInstanceInfosResponse = ::messages::QueryDebugInstanceInfosResponse;

using ReportAgentAbnormalRequest = ::messages::ReportAgentAbnormalRequest;
using ReportAgentAbnormalResponse = ::messages::ReportAgentAbnormalResponse;

using CheckInstanceStateRequest = ::messages::CheckInstanceStateRequest;
using CheckInstanceStateResponse = ::messages::CheckInstanceStateResponse;

namespace MetaStore {
using PutRequest = ::messages::PutRequest;
using PutResponse = ::messages::PutResponse;
using Lease = ::messages::Lease;
using ObserveResponse = ::messages::ObserveResponse;
using ObserveCancelRequest = ::messages::ObserveCancelRequest;
using ForwardWatchRequest = ::messages::ForwardWatchRequest;
}  // namespace MetaStore
}  // namespace functionsystem::messages
#endif