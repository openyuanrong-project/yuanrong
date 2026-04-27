/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
#include <ctime>
#include <iomanip>
#include <string>
#include <vector>

#include "datasystem/utils/connection.h"
#include "datasystem/utils/sensitive_value.h"
#include "src/dto/config.h"
#include "src/dto/invoke_options.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/fsclient/protobuf/core_service.grpc.pb.h"
#include "src/libruntime/fsclient/protobuf/runtime_service.grpc.pb.h"
#include "src/libruntime/stacktrace/stack_trace_info.h"
#include "src/libruntime/utils/security.h"

namespace YR {
using CallResult = ::core_service::CallResult;
using NotifyRequest = ::runtime_service::NotifyRequest;
using GroupPolicy = ::common::GroupPolicy;
using BindStrategy = ::common::BindStrategy;

using datasystem::ConnectOptions;
struct IpAddrInfo {
    std::string ip;
    int32_t port;
};

void Split(const std::string &source, std::vector<std::string> &result, const char sep);
IpAddrInfo ParseIpAddr(const std::string &addr);
void ParseIpAddr(const std::string &addr, std::string &ip, int32_t &port);
std::string GetIpAddr(const std::string &ip, int port);
long long GetCurrentTimestampMs();
long long GetCurrentTimestampNs();
void SetCallResultWithStackTraceInfo(std::vector<YR::Libruntime::StackTraceInfo> &infos, CallResult &callResult);
std::vector<YR::Libruntime::StackTraceInfo> GetStackTraceInfos(const NotifyRequest &req);
void GetServerName(std::shared_ptr<YR::Libruntime::Security> security, std::string &serverName);
std::shared_ptr<grpc::ChannelCredentials> GetChannelCreds(std::shared_ptr<YR::Libruntime::Security> security);
std::shared_ptr<grpc::ServerCredentials> GetServerCreds(std::shared_ptr<YR::Libruntime::Security> security);
bool InstanceRangeEnabled(YR::Libruntime::InstanceRange instanceRange);
bool ResourceGroupEnabled(YR::Libruntime::ResourceGroupOptions resourceGroupOpts);
bool FunctionGroupEnabled(YR::Libruntime::FunctionGroupOptions options);
int unhexlify(std::string input, char *ascii);
void GetAuthConnectOpts(ConnectOptions &connnections, const std::string &ak, const datasystem::SensitiveValue &sk,
                        const datasystem::SensitiveValue &token);
std::string GetCurrentUTCTime();
bool IsLaterThan(const std::string &timestamp1, const std::string &timestamp2, double seconds);
std::tm ParseTimestamp(const std::string &timestamp);
std::string GetEnvValue(const std::string &key);
int32_t ToMs(int32_t timeoutS);
bool WillSizeOverFlow(size_t a, size_t b);
GroupPolicy ConvertStrategyToPolicy(const std::string &stategy);

// Load environment variables from a .env format file
// This must be called before any code reads from environment variables
// The file should contain environment variables in KEY=VALUE format, one per line
// Lines starting with # are treated as comments and ignored
// Empty lines are ignored
void LoadEnvFromFile(const std::string &envFile);
std::string ConvertBindResource(const std::string &resource);
BindStrategy ConvertBindStrategy(const std::string &strategy);
}  // namespace YR
