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

#include "api/cpp/src/faas/context_env.h"

namespace Function {

int ContextEnv::GetRunningTimeInSeconds() const
{
    return runningTimeInSeconds_;
}

int ContextEnv::GetCPUNumber() const
{
    return cpuNumber_;
}
int ContextEnv::GetMemorySize() const
{
    return memorySize_;
}

const std::string ContextEnv::GetInstanceLabel() const
{
    return instanceLabel_;
}

const std::string ContextEnv::GetUserData(std::string key) const
{
    auto userDataIt = userData_.find(key);
    if (userDataIt != userData_.end()) {
        return userDataIt->second;
    }
    return "";
}

const std::string ContextEnv::GetFuncPackage() const
{
    return funcPackage_;
}

const std::string ContextEnv::GetFunctionName() const
{
    return functionName_;
}

const std::string ContextEnv::GetVersion() const
{
    return version_;
}

const std::string ContextEnv::GetProjectID() const
{
    return projectID_;
}

void ContextEnv::SetUserData(std::unordered_map<std::string, std::string> &funcKey)
{
    std::swap(funcKey, userData_);
}

void ContextEnv::SetRunningTimeInSeconds(int runningTime)
{
    runningTimeInSeconds_ = runningTime;
}
void ContextEnv::SetCPUNumber(int cpuNum)
{
    cpuNumber_ = cpuNum;
}
void ContextEnv::SetMemorySize(int memorySz)
{
    memorySize_ = memorySz;
}

void ContextEnv::SetInstanceLabel(const std::string &instanceLabel)
{
    instanceLabel_ = instanceLabel;
}

void ContextEnv::SetFuncPackage(const std::string &package)
{
    funcPackage_ = package;
}
void ContextEnv::SetFunctionName(const std::string &funcName)
{
    functionName_ = funcName;
}
void ContextEnv::SetVersion(const std::string &funcVersion)
{
    version_ = funcVersion;
}
void ContextEnv::SetProjectID(const std::string &proId)
{
    projectID_ = proId;
}

}  // namespace Function