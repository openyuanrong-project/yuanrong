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
#include <signal.h>
#include <sys/prctl.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>
#include <chrono>
#include <fstream>
#include <functional>
#include <future>
#include <iostream>
#include <memory>
#include <string>
#include <unordered_map>
#include <vector>

namespace YR {
namespace Libruntime {

/**

Automatically get cluster information for init, right now including 3 parts
   * address for posix invoke, can be frontend or bus
   * datasystem address
   * whether in cluster, which indicate whether frontend or bus

Each lang should read user config in different way, and pass all parts of them into the `AutoGetClusterAccessInfo`
method, `AutoGetClusterAccessInfo` will read the configs and normalize each key

===============================================================================

Will try to get the cluster info in following steps,

1. Try get from user config params, this should be done by API layer (in different languages)

2. Try get from environment variables, this should be united between all languages
     * YR_SERVER_ADDRESS
     * YR_DS_ADDRESS
     * YR_IN_CLUSTER

3. Try get from /tmp/yr_sessions/yr_current_master_info (mainly server addresses)
     * always in cluster right now

4. Try start new temporary env with `yr start` command
     * and then get info the same as step 3.
*/

const std::string kEnvYrServerAddress = "YR_SERVER_ADDRESS";  // address for posix invoke, can be frontend or bus
const std::string kEnvYrDatasystemAddress = "YR_DS_ADDRESS";  // datasystem address

class ClusterAccessInfo {
public:
    // addresses
    std::string serverAddr = "";                     // for posix invocation
    std::string dsAddr = "";                         // for datasystem worker
    static std::string masterAddr;                   // for function master
    static std::vector<std::string> masterAddrList;  // for function master cluster

    // is in cluster
    bool inCluster;
    static bool isMasterCluster;

public:
    // public methods
    void AutoParse();
    void ParseFromMasterInfo();

private:
    void ParseFromEnv();
    void ParseServerAddrProtocol();
    void ParseDsAddr();
    std::pair<std::string, std::string> ParseURLWithProtocol(const std::string &url);
};

class CommandRunner {
public:
    CommandRunner() {}

    void RunCommandUntil(std::vector<std::string> &args);
    std::string GetOutput();

private:
    std::string output_;

    static constexpr int SLEEP_MS = 100;     // Waiting time
    static constexpr int MAX_RETRIES = 100;  // Maximum retry count
};

/**
  There are a lot of std output to explicit let user understand that we are trying to start a local yuanrong cluster.
*/
ClusterAccessInfo AutoCreateYuanRongCluster(std::vector<std::string> &args);
ClusterAccessInfo AutoGetClusterAccessInfo(ClusterAccessInfo info, std::vector<std::string> args = {});

}  // namespace Libruntime
}  // namespace YR
