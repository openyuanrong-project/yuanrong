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

#include "utils.h"

std::string GetOutput(std::string cmd)
{
    FILE *fp;
    char buffer[1024];
    fp = popen(cmd.c_str(), "r");
    fgets(buffer, sizeof(buffer), fp);
    return std::string(buffer);
}

std::string GetPid(std::string keyword)
{
    std::string cmd = "ps -ef | grep " + keyword + " | grep -v grep | awk '{print $2}'";
    auto pid = GetOutput(cmd);
    auto pos = pid.find("\n");
    if (pos > -1) {
        pid.erase(pos, 1);
    }
    return pid;
}

std::string GetPpid(std::string keyword)
{
    std::string cmd = "ps -ef | grep " + keyword + " | grep -v grep | awk '{print $3}'";
    auto pid = GetOutput(cmd);
    auto pos = pid.find("\n");
    if (pos > -1) {
        pid.erase(pos, 1);
    }
    return pid;
}

void Kill(std::string pid)
{
    std::string cmd = "kill -9 " + pid;
    system(cmd.c_str());
}