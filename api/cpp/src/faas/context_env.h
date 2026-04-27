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

#include <string>
#include <unordered_map>
namespace Function {
class ContextEnv {
public:
    int GetRunningTimeInSeconds() const;
    int GetCPUNumber() const;
    int GetMemorySize() const;
    const std::string GetInstanceLabel() const;
    const std::string GetUserData(std::string key) const;
    const std::string GetFuncPackage() const;
    const std::string GetFunctionName() const;
    const std::string GetVersion() const;
    const std::string GetProjectID() const;

    void SetRunningTimeInSeconds(int runningTimeInSeconds);
    void SetCPUNumber(int cpuNumber);
    void SetMemorySize(int memorySize);
    void SetInstanceLabel(const std::string &instanceLabel);
    void SetUserData(std::unordered_map<std::string, std::string> &userData);
    void SetFuncPackage(const std::string &package);
    void SetFunctionName(const std::string &funcName);
    void SetVersion(const std::string &version);
    void SetProjectID(const std::string &projectID);

private:
    int runningTimeInSeconds_ = 0;
    int cpuNumber_ = 0;
    int memorySize_ = 0;
    std::string instanceLabel_;
    std::unordered_map<std::string, std::string> userData_;
    std::string funcPackage_;
    std::string functionName_;
    std::string version_;
    std::string projectID_;
};
}  // namespace Function
