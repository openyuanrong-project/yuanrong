/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include "auto_init.h"
#if __has_include(<filesystem>)
#include <filesystem>
#else
#include <experimental/filesystem>
#endif
#include "src/dto/config.h"
#include "src/utility/string_utility.h"
namespace YR {
namespace Libruntime {

#ifdef __cpp_lib_filesystem
namespace filesystem = std::filesystem;
#elif __cpp_lib_experimental_filesystem
namespace filesystem = std::experimental::filesystem;
#endif

std::string YR::Libruntime::ClusterAccessInfo::masterAddr;
std::vector<std::string> YR::Libruntime::ClusterAccessInfo::masterAddrList;
bool YR::Libruntime::ClusterAccessInfo::isMasterCluster;
const int PROTOCOL_SUFFIX_LEN = 3;

void ClusterAccessInfo::AutoParse()
{
    // 1. parse from user provided info (include environment)
    ParseFromEnv();
    ParseServerAddrProtocol();
    ParseDsAddr();

    if (!serverAddr.empty() && !dsAddr.empty()) {
        return;
    }

    // 2. if not filled, try parse from master info
    ParseFromMasterInfo();
}

void ClusterAccessInfo::ParseFromMasterInfo()
{
    masterAddrList.clear();
    isMasterCluster = false;
    std::string masterInfo;
    if (!Config::Instance().YR_MASTER_INFO().empty()) {
        masterInfo = Config::Instance().YR_MASTER_INFO();
    } else {
        std::ifstream masterInfoFile(Config::Instance().YR_MASTER_INFO_PATH());
        if (!masterInfoFile.is_open()) {
            return;
        }

        if (!std::getline(masterInfoFile, masterInfo)) {
            return;
        }
    }

    std::map<std::string, std::string> kvMap;
    std::map<std::string, std::vector<std::string>> kvsMap;
    std::string pattern = "[,:]";
    auto result = YR::utility::SplitToStr(masterInfo, pattern);
    for (size_t i = 0; i < result.size();) {
        std::string key = result[i];
        ++i;
        if (i >= result.size()) {
            break;
        }
        std::string value = result[i];
        ++i;
        if (kvMap.find(key) != kvMap.end()) {
            if (kvsMap.find(key) == kvsMap.end()) {
                kvsMap[key].push_back(kvMap[key]);
            }
            kvsMap[key].push_back(value);
        } else {
            kvMap[key] = value;
        }
    }

    std::string masterIP;
    if (auto masterIPIt = kvMap.find("master_ip"); masterIPIt != kvMap.end()) {
        masterIP = masterIPIt->second;
    }

    std::string agentIP;
    if (auto agentIPIt = kvMap.find("local_ip"); agentIPIt != kvMap.end()) {
        agentIP = agentIPIt->second;
    }
    std::string etcdIP;
    if (auto etcdIPIt = kvMap.find("etcd_ip"); etcdIPIt != kvMap.end()) {
        etcdIP = etcdIPIt->second;
    }

    std::string etcdPort;
    if (auto etcdPortIt = kvMap.find("etcd_port"); etcdPortIt != kvMap.end()) {
        etcdPort = etcdPortIt->second;
    }

    std::string busPort;
    if (auto busPortIt = kvMap.find("bus"); busPortIt != kvMap.end()) {
        busPort = busPortIt->second;
    }

    std::string dsPort;
    if (auto dsPortIt = kvMap.find("ds-worker"); dsPortIt != kvMap.end()) {
        dsPort = dsPortIt->second;
    }

    serverAddr = agentIP + ":" + busPort;
    dsAddr = agentIP + ":" + dsPort;
    inCluster = true;

    std::string globalSchedPort;
    if (auto gsPortIt = kvMap.find("global_scheduler_port"); gsPortIt != kvMap.end()) {
        globalSchedPort = gsPortIt->second;
    }

    masterAddr = masterIP + ":" + globalSchedPort;

    auto mastersIt = kvsMap.find("etcd_addr_list");
    if (mastersIt == kvsMap.end()) {
        if (auto it = kvMap.find("etcd_addr_list"); it != kvMap.end()) {
            masterAddrList.push_back(it->second + ":" + globalSchedPort);
            isMasterCluster = true;
        }
    } else {
        for (auto &m : mastersIt->second) {
            masterAddrList.push_back(m + ":" + globalSchedPort);
        }
        isMasterCluster = true;
    }
}

void ClusterAccessInfo::ParseFromEnv()
{
    // if not empty, just use specified value
    if (serverAddr.empty() && !Config::Instance().YR_SERVER_ADDRESS().empty()) {
        serverAddr = Config::Instance().YR_SERVER_ADDRESS();
    }
    if (dsAddr.empty() && !Config::Instance().YR_DS_ADDRESS().empty()) {
        dsAddr = Config::Instance().YR_DS_ADDRESS();
    }
}

void ClusterAccessInfo::ParseServerAddrProtocol()
{
    std::unordered_map<std::string, bool> protocol2Cluster{
        {"http", false},
        {"https", false},
        {"grpc", true},
    };

    auto [sProtocol, sAddr] = ParseURLWithProtocol(serverAddr);
    if (auto it = protocol2Cluster.find(sProtocol); it != protocol2Cluster.end()) {
        inCluster = it->second;
        serverAddr = sAddr;
    }
}

void ClusterAccessInfo::ParseDsAddr()
{
    if (!dsAddr.empty()) {
        // if user specified ds addr, use it
        // datasystem protocol is not concerned
        auto [dProtocol, dAddr] = ParseURLWithProtocol(dsAddr);
        (void)dProtocol;
        dsAddr = dAddr;
    } else if (!inCluster) {
        // if not in cluster, the datasystem addr should be the same as server addr
        // of course, only if user doesn't specified it
        dsAddr = serverAddr;
    } else {
        // empty, and inCluster, do nothing, left later
    }
}

std::pair<std::string, std::string> ClusterAccessInfo::ParseURLWithProtocol(const std::string &url)
{
    re2::RE2 urlPattern(R"((^[a-zA-Z]+://)?(.*))");
    std::string protocol;
    std::string remainder;
    if (RE2::PartialMatch(url, urlPattern, &protocol, &remainder)) {
        if (protocol.empty()) {
            return {"", remainder};
        } else {
            protocol = protocol.substr(0, protocol.size() - PROTOCOL_SUFFIX_LEN);  // get protocol
            return {protocol, remainder};
        }
    }

    return {"", ""};
}

static bool CheckCommandExists(const std::string &command)
{
    const char *path = getenv("PATH");
    if (!path) {
        return false;
    }
    std::stringstream ss(path);
    std::string dir;
    while (getline(ss, dir, ':')) {
        std::string fullPath = dir + "/" + command;
        if (access(fullPath.c_str(), X_OK) == 0) {
            return true;
        }
    }
    return false;
}

void CommandRunner::RunCommandUntil(std::vector<std::string> &args)
{
    pid_t pid = fork();
    if (pid == -1) {
        perror("fork");
        return;
    } else if (pid == 0) {
        // Child process
        if (freopen("/dev/null", "w", stdout) == nullptr) {
            perror("freopen failed");
            return;
        }
        prctl(PR_SET_PDEATHSIG, SIGTERM);  // Send SIGTERM to child if parent dies

        std::vector<std::string> argv = {"yr", "start", "--master", "--block", "true"};
        argv.reserve(argv.size() + args.size());
        argv.insert(argv.end(), args.begin(), args.end());

        char **argValues = new (std::nothrow) char *[argv.size() + 1];
        for (size_t i = 0; i < argv.size(); i++) {
            argValues[i] = const_cast<char *>(argv[i].c_str());
        }
        // execvpe arg last param must be NULL
        argValues[argv.size()] = nullptr;

        execvp("yr", argValues);
        delete[] argValues;
        perror("execl");
        exit(EXIT_FAILURE);
    } else {
        // Parent process
        int retries = 0;
        bool fileExists = false;

        while (!fileExists && retries < MAX_RETRIES) {
            fileExists = filesystem::exists(Config::Instance().YR_MASTER_INFO_PATH());
            if (fileExists) {
                break;
            }
            std::this_thread::sleep_for(std::chrono::milliseconds(SLEEP_MS));
            retries++;
        }

        int status;
        waitpid(pid, &status, WNOHANG);
        if (WIFEXITED(status) && WEXITSTATUS(status) != 0) {
            std::cerr << "Child process exited abnormally. Output:\n" << output_ << std::endl;
        }
    }
}

std::string CommandRunner::GetOutput()
{
    return output_;
}

/**
  There are a lot of std output to explicit let user understand that we are trying to start a local yuanrong cluster.
*/
ClusterAccessInfo AutoCreateYuanRongCluster(std::vector<std::string> &args)
{
    if (!CheckCommandExists("yr")) {
        std::cout << "failed to detect `yr` command in PATH, check if you have install yuanrong core packages."
                  << std::endl;
        return ClusterAccessInfo{};
    }

    std::cout << "There is no existing Yuanrong cluster. Trying to start a temporary one..." << std::endl;
    auto startTime = std::chrono::steady_clock::now();
    auto commandRunner = CommandRunner();
    commandRunner.RunCommandUntil(args);

    ClusterAccessInfo info;
    info.ParseFromMasterInfo();

    if (info.serverAddr != "") {
        std::cout
            << "A temporary Yuanrong cluster has been started, taking "
            << std::chrono::duration_cast<std::chrono::seconds>(std::chrono::steady_clock::now() - startTime).count()
            << " seconds. This cluster will be automatically destroyed when the driver program exits." << std::endl
            << "the address(" << info.serverAddr << ")"
            << ", datasystem address(" << info.dsAddr << ")." << std::endl;
    } else {
        std::cout
            << "Temporary yuanrong cluster started failed, try run `yr start --master` before running your program."
            << std::endl
            << "error msg: " << commandRunner.GetOutput() << std::endl;
    }
    return info;
}

bool IsValidIPPort(const std::string &input)
{
    re2::RE2 pattern(R"((\d{1,3}\.){3}\d{1,3}:\d{1,5})");
    return RE2::FullMatch(input, pattern);
}

bool IsURLHasProtocalPrefix(const std::string &input)
{
    re2::RE2 pattern(R"((http|https|grpc)://([a-zA-Z0-9.-]+|\d{1,3}(\.\d{1,3}){3}):\d{1,5})");
    return RE2::FullMatch(input, pattern);
}

bool NeedToBeParsed(ClusterAccessInfo info)
{
    // Parse the access info if
    // * serverAddress is not specified
    if (info.serverAddr.empty()) {
        return true;
    }
    // * serverAddress contains protocal prefix
    if (IsURLHasProtocalPrefix(info.serverAddr)) {
        return true;
    }
    return false;
}

ClusterAccessInfo AutoGetClusterAccessInfo(ClusterAccessInfo info, std::vector<std::string> args)
{
    if (!NeedToBeParsed(info)) {
        return info;
    }

    info.AutoParse();
    if (!info.serverAddr.empty() && !info.dsAddr.empty()) {
        return info;
    }
    return AutoCreateYuanRongCluster(args);
}

}  // namespace Libruntime
}  // namespace YR
