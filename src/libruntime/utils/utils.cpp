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
#include "utils.h"

#include <cctype>
#include <chrono>
#include <fstream>
#include <sstream>

#include "src/utility/logger/logger.h"

namespace YR {
const int IP_INDEX = 0;
const int PORT_INDEX = 1;
const int HEX_OFFSET = 10;
const int HEX_BIT_OFFSET = 4;
const size_t FUNCNAME = 1;
const size_t FUNCVERSION = 2;
const size_t UID_INDEX = 0;
const size_t SERVICE_INDEX = 0;
void Split(const std::string &source, std::vector<std::string> &result, const char sep)
{
    result.clear();
    std::istringstream iss(source);
    std::string temp;
    while (std::getline(iss, temp, sep)) {
        result.emplace_back(std::move(temp));
    }
    return;
}

IpAddrInfo ParseIpAddr(const std::string &addr)
{
    std::vector<std::string> result;
    Split(addr, result, ':');
    if (result.size() < PORT_INDEX + 1) {
        return {"", 0};
    }
    return {result[IP_INDEX], std::stoi(result[PORT_INDEX])};
}

void ParseIpAddr(const std::string &addr, std::string &ip, int32_t &port)
{
    auto info = ParseIpAddr(addr);
    ip = info.ip;
    port = info.port;
}

std::string GetIpAddr(const std::string &ip, int port)
{
    return ip + ":" + std::to_string(port);
}

long long GetCurrentTimestampMs()
{
    auto now = std::chrono::steady_clock::now();
    auto duration = now.time_since_epoch();
    return std::chrono::duration_cast<std::chrono::milliseconds>(duration).count();
}

long long GetCurrentTimestampNs()
{
    auto now = std::chrono::system_clock::now();
    auto duration = now.time_since_epoch();
    return std::chrono::duration_cast<std::chrono::nanoseconds>(duration).count();
}

std::string GetCurrentUTCTime()
{
    std::ostringstream oss;
    std::time_t now = std::time(nullptr);
    // Converts the current time to the UTC time.
    std::tm *utcTime = std::gmtime(&now);
    oss << std::put_time(utcTime, "%Y%m%dT%H%M%SZ");
    return oss.str();
}

std::tm ParseTimestamp(const std::string &timestamp)
{
    std::tm tm = {};
    std::istringstream ss(timestamp);
    ss >> std::get_time(&tm, "%Y%m%dT%H%M%SZ");
    if (ss.fail()) {
        return {};
    }
    return tm;
}

bool IsLaterThan(const std::string &timestamp1, const std::string &timestamp2, double seconds)
{
    std::tm tm1 = ParseTimestamp(timestamp1);
    std::tm tm2 = ParseTimestamp(timestamp2);
    std::time_t time1 = std::mktime(&tm1);
    std::time_t time2 = std::mktime(&tm2);
    return std::difftime(time1, time2) > seconds;
}

void SetCallResultWithStackTraceInfo(std::vector<YR::Libruntime::StackTraceInfo> &infos, CallResult &callResult)
{
    for (size_t i = 0; i < infos.size(); i++) {
        auto setInfo = callResult.add_stacktraceinfos();
        auto eles = infos[i].StackTraceElements();
        for (size_t j = 0; j < eles.size(); j++) {
            auto stackTraceEle = setInfo->add_stacktraceelements();
            stackTraceEle->set_classname(eles[j].className);
            stackTraceEle->set_methodname(eles[j].methodName);
            stackTraceEle->set_filename(eles[j].fileName);
            stackTraceEle->set_linenumber(eles[j].lineNumber);
            for (auto &it : eles[j].extensions) {
                stackTraceEle->mutable_extensions()->insert({it.first, it.second});
            }
        }
        setInfo->set_type(infos[i].Type());
        setInfo->set_message(infos[i].Message());
    }
}

std::vector<YR::Libruntime::StackTraceInfo> GetStackTraceInfos(const NotifyRequest &req)
{
    std::vector<YR::Libruntime::StackTraceInfo> infos;
    int stackTraceInfosSize = req.stacktraceinfos_size();
    for (int i = 0; i < stackTraceInfosSize; i++) {
        auto &info = req.stacktraceinfos(i);
        std::vector<YR::Libruntime::StackTraceElement> elements;
        int elementSize = info.stacktraceelements_size();
        for (int j = 0; j < elementSize; j++) {
            auto &ele = info.stacktraceelements(j);
            std::unordered_map<std::string, std::string> extensions;
            for (auto it = ele.extensions().cbegin(); it != ele.extensions().cend(); ++it) {
                extensions[it->first] = it->second;
            }
            YR::Libruntime::StackTraceElement stackTraceElement = {ele.classname(), ele.methodname(), ele.filename(),
                                                                   (int)ele.linenumber(), extensions};
            elements.emplace_back(stackTraceElement);
        }
        YR::Libruntime::StackTraceInfo stackTraceInfo(info.type(), info.message(), elements);
        infos.emplace_back(stackTraceInfo);
    }
    if (req.code() == common::ErrorCode::ERR_USER_FUNCTION_EXCEPTION && req.smallobjects_size() == 1) {
        auto &smallObj = req.smallobjects(0);
        const std::string &bufStr = smallObj.value();
        auto st = YR::Libruntime::StackTraceInfo("YRInvokeError", bufStr, "python");
        infos.push_back(st);
    }
    return infos;
}

void GetServerName(std::shared_ptr<YR::Libruntime::Security> security, std::string &serverName)
{
    if (security != nullptr) {
        (void)security->GetFunctionSystemConnectionMode(serverName);
    }
}

std::shared_ptr<grpc::ChannelCredentials> GetChannelCreds(std::shared_ptr<YR::Libruntime::Security> security)
{
    bool enable = false;
    std::string rootCACert;
    std::string certChain;
    std::string privateKey;
    if (security != nullptr) {
        enable = security->GetFunctionSystemConfig(rootCACert, certChain, privateKey);
    }
    if (!enable) {
        return grpc::InsecureChannelCredentials();
    }

    auto credsOpts = grpc::SslCredentialsOptions();
    credsOpts.pem_root_certs = rootCACert;
    credsOpts.pem_cert_chain = certChain;
    credsOpts.pem_private_key = privateKey;
    return grpc::SslCredentials(credsOpts);
}

void GetAuthConnectOpts(ConnectOptions &connnections, const std::string &ak, const datasystem::SensitiveValue &sk,
                        const datasystem::SensitiveValue &token)
{
    if (!ak.empty() && !sk.Empty()) {
        // ak、sk used for auth of system function instance
        connnections.accessKey = ak;
        connnections.secretKey = sk;
    } else {
        YRLOG_DEBUG("GetAuthConnectOpts token not empty");
        // token used for auth of normal function instance
    }
}

std::shared_ptr<grpc::ServerCredentials> GetServerCreds(std::shared_ptr<YR::Libruntime::Security> security)
{
    bool enable = false;
    std::string rootCACert;
    std::string certChain;
    std::string privateKey;
    if (security != nullptr) {
        enable = security->GetFunctionSystemConfig(rootCACert, certChain, privateKey);
    }
    if (!enable) {
        return grpc::InsecureServerCredentials();
    }

    ::grpc::SslServerCredentialsOptions::PemKeyCertPair pemKeyCertPair;
    pemKeyCertPair.private_key = privateKey;
    pemKeyCertPair.cert_chain = certChain;
    ::grpc::SslServerCredentialsOptions sslServerCredentialsOptions(GRPC_SSL_REQUEST_CLIENT_CERTIFICATE_AND_VERIFY);
    sslServerCredentialsOptions.pem_key_cert_pairs.push_back(std::move(pemKeyCertPair));
    sslServerCredentialsOptions.pem_root_certs = rootCACert;
    return grpc::SslServerCredentials(sslServerCredentialsOptions);
}

bool InstanceRangeEnabled(YR::Libruntime::InstanceRange instanceRange)
{
    if (instanceRange.min == YR::Libruntime::DEFAULT_INSTANCE_RANGE_NUM &&
        instanceRange.max == YR::Libruntime::DEFAULT_INSTANCE_RANGE_NUM) {
        return false;
    }
    return true;
}

bool ResourceGroupEnabled(YR::Libruntime::ResourceGroupOptions resourceGroupOpts)
{
    return !resourceGroupOpts.resourceGroupName.empty();
}

bool FunctionGroupEnabled(YR::Libruntime::FunctionGroupOptions options)
{
    return (options.functionGroupSize > 0) && (options.bundleSize > 0);
}

int to_int(char c)
{
    if (not isxdigit(c))
        return -1;  // error: non-hexadecimal digit found
    if (isdigit(c))
        return c - '0';
    if (isupper(c))
        c = tolower(c);
    return c - 'a' + HEX_OFFSET;
}

int32_t ToMs(int32_t timeoutS)
{
    int64_t tmp = static_cast<int64_t>(timeoutS) * 1000;
    if (tmp > INT32_MAX) {
        return INT32_MAX;
    }
    return static_cast<int32_t>(tmp);
}

bool WillSizeOverFlow(size_t a, size_t b)
{
    return b > (std::numeric_limits<size_t>::max() - a);
}

GroupPolicy ConvertStrategyToPolicy(const std::string &stategy)
{
    static const std::unordered_map<std::string, GroupPolicy> strategyMap = {
        {"PACK", GroupPolicy::Pack},
        {"STRICT_PACK", GroupPolicy::StrictPack},
        {"SPREAD", GroupPolicy::Spread},
        {"STRICT_SPREAD", GroupPolicy::StrictSpread}};
    auto it = strategyMap.find(stategy);
    return (it != strategyMap.end()) ? it->second : GroupPolicy::None;
}

std::string ConvertBindResource(const std::string &resource)
{
    static const std::unordered_map<std::string, std::string> resourceMap = {
        {"NONE", "NONE"},
        {"NUMA", "NUMA"}};
    auto resourceIt = resourceMap.find(resource);
    if (resourceIt != resourceMap.end()) {
        return resourceIt->second;
    }
    YRLOG_WARN("Invalid bind resource, {}", resource);
    return "NONE";
}

BindStrategy ConvertBindStrategy(const std::string &strategy)
{
    static const std::unordered_map<std::string, BindStrategy> strategyMap = {
        {"NONE", BindStrategy::BIND_None},
        {"SPREAD", BindStrategy::BIND_Spread},
        {"PACK", BindStrategy::BIND_Pack}};
    auto strategyIt = strategyMap.find(strategy);
    if (strategyIt != strategyMap.end()) {
        return strategyIt->second;
    }
    YRLOG_WARN("Invalid bind strategy, {}", strategy);
    return BindStrategy::BIND_None;
}

int unhexlify(std::string input, char *ascii)
{
    auto first = input.c_str();
    auto last = first + input.size();
    while (first != last) {
        int top = to_int(*first++);
        int bot = to_int(*first++);
        if (top == -1 or bot == -1)
            return -1;  // error
        *ascii++ = (top << 4) + bot;
    }
    return 0;
}

std::string GetEnvValue(const std::string &key)
{
    if (const char *env = std::getenv(key.c_str())) {
        return std::string(env);
    }
    return std::string("");
}

std::string StripWhitespace(const std::string &str)
{
    size_t start = str.find_first_not_of(" \t\r\n");
    if (start == std::string::npos) {
        return "";
    }
    size_t end = str.find_last_not_of(" \t\r\n");
    return str.substr(start, end - start + 1);
}

bool ShouldSkipLine(const std::string &line)
{
    return line.empty() || line[0] == '#';
}

bool ParseKeyValue(const std::string &line, std::string &key, std::string &value)
{
    size_t eqPos = line.find('=');
    if (eqPos == std::string::npos) {
        return false;
    }

    key = line.substr(0, eqPos);
    value = line.substr(eqPos + 1);
    return true;
}

std::string StripKeyValueWhitespace(const std::string &str)
{
    size_t start = str.find_first_not_of(" \t");
    if (start == std::string::npos) {
        return "";
    }
    size_t end = str.find_last_not_of(" \t");
    return str.substr(start, end - start + 1);
}

void RemoveQuotes(std::string &value)
{
    const uint32_t minSize = 2;
    if (value.length() >= minSize) {
        if ((value[0] == '"' && value[value.length() - 1] == '"') ||
            (value[0] == '\'' && value[value.length() - 1] == '\'')) {
            value = value.substr(1, value.length() - minSize);
        }
    }
}

void LoadEnvFromFile(const std::string &envFile)
{
    if (envFile.empty()) {
        YRLOG_WARN("Environment variable file is empty");
        return;
    }

    std::ifstream file(envFile);
    if (!file.is_open()) {
        YRLOG_WARN("Environment variable file not found: {}", envFile);
        return;
    }

    try {
        std::string line;
        size_t lineNum = 0;
        size_t loadedCount = 0;

        while (std::getline(file, line)) {
            lineNum++;

            line = StripWhitespace(line);
            if (ShouldSkipLine(line)) {
                continue;
            }

            std::string key = "";
            std::string value = "";
            if (!ParseKeyValue(line, key, value)) {
                YRLOG_WARN("Invalid format in {} at line {}: expected KEY=VALUE format, got: {}", envFile, lineNum,
                           line);
                continue;
            }

            key = StripKeyValueWhitespace(key);
            if (key.empty()) {
                YRLOG_WARN("Empty key in {} at line {}", envFile, lineNum);
                continue;
            }

            value = StripKeyValueWhitespace(value);
            RemoveQuotes(value);

            if (setenv(key.c_str(), value.c_str(), 1) != 0) {
                YRLOG_ERROR("Failed to set environment variable: {}", key);
                continue;
            }
            loadedCount++;
        }

        if (loadedCount > 0) {
            YRLOG_DEBUG("Loaded {} environment variables from {}", loadedCount, envFile);
            Libruntime::Config::Reset();
        }
    } catch (const std::exception &e) {
        YRLOG_ERROR("Failed to load environment variables from {}: {}", envFile, e.what());
    }
}
}  // namespace YR
